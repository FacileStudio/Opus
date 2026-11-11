package mcp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/Opus/apps/api/modules/activity"
	"github.com/FacileStudio/Opus/apps/api/modules/column"
	"github.com/FacileStudio/Opus/apps/api/modules/label"
	"github.com/FacileStudio/Opus/apps/api/modules/project"
	"github.com/FacileStudio/Opus/apps/api/modules/task"
	"github.com/FacileStudio/Opus/apps/api/modules/workspace"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/mcpauth"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	"github.com/FacileStudio/tronc/mcp/mcptest"
	"github.com/go-chi/chi/v5"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Subjects of the test people: the owner calls the tools, the teammate
// shares the studio workspace, the stranger owns a workspace the owner is
// not in.
const (
	subject         = "owner-sub"
	strangerSubject = "stranger-sub"
)

// identities maps the test subjects to their users and nothing else.
type identities struct {
	porte.IdentityStore
	users map[string]int64
}

// Find resolves only the test subjects.
func (s identities) Find(_ context.Context, provider, sub string) (porte.StoredIdentity, error) {
	id, ok := s.users[sub]
	if !ok {
		return porte.StoredIdentity{}, porte.ErrNotFound
	}
	return porte.StoredIdentity{UserID: id, Provider: provider, Subject: sub}, nil
}

// fixture is a running MCP endpoint over Opus's real services.
type fixture struct {
	db        *gorm.DB
	srv       *httptest.Server
	issuer    *mcptest.Issuer
	pending   mcpkit.Pending
	service   *Service
	domain    Domain
	owner     int64
	teammate  int64
	stranger  int64
	studio    int64
	project   int64
	foreign   int64
	resource  string
	perMinute int
}

// openDB opens the test database with Opus's migrations, mcp_actions
// included, and the Postgres pending store on it.
func openDB(t *testing.T) (*gorm.DB, mcpkit.Pending) {
	t.Helper()
	db := dbtest.Open(t)
	db.Logger = logger.Discard
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	return db, mcpkit.PostgresPending{DB: sqlDB}
}

// newFixture starts the endpoint.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	return newFixtureOn(t, nil, nil)
}

// newFixtureOn is newFixture on a given listener, nil for any free port,
// with front wrapping the routes when it is set.
func newFixtureOn(t *testing.T, listener net.Listener, front func(*fixture, http.Handler) http.Handler) *fixture {
	t.Helper()
	f := &fixture{issuer: mcptest.NewIssuer(t)}
	f.db, f.pending = openDB(t)
	if listener != nil {
		f.perMinute = 1_000_000
	}
	f.seed(t)
	var handler http.Handler
	f.srv = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handler.ServeHTTP(w, r) }))
	if listener != nil {
		f.srv.Listener = listener
	}
	f.srv.Start()
	t.Cleanup(f.srv.Close)
	f.resource = f.srv.URL + "/mcp"
	f.service = f.build(t)
	router := chi.NewRouter()
	RegisterRoutes(router, f.service)
	handler = router
	if front != nil {
		handler = front(f, router)
	}
	return f
}

// seed creates the people, the shared studio workspace with one project and
// a bug label, and a workspace of the stranger's with one project.
func (f *fixture) seed(t *testing.T) {
	t.Helper()
	f.owner = f.user(t, "owner@example.test", "Owner")
	f.teammate = f.user(t, "mate@example.test", "Mate")
	f.stranger = f.user(t, "stranger@example.test", "Stranger")
	f.studio = f.workspace(t, "Studio", map[int64]string{f.owner: "owner", f.teammate: "member"})
	foreignWorkspace := f.workspace(t, "Elsewhere", map[int64]string{f.stranger: "owner"})
	f.project = f.newProject(t, f.studio, "Web", "web")
	f.foreign = f.newProject(t, foreignWorkspace, "Secret", "secret")
	now := time.Now()
	if err := f.db.Create(&schemas.Label{WorkspaceID: f.studio, Name: "bug", Color: "#e11d48", CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
}

// user creates one person.
func (f *fixture) user(t *testing.T, email, name string) int64 {
	t.Helper()
	u := schemas.User{Email: email, Name: name, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := f.db.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	return u.ID
}

// workspace creates a shared workspace with members and their roles.
func (f *fixture) workspace(t *testing.T, name string, members map[int64]string) int64 {
	t.Helper()
	ws := schemas.Workspace{Name: name, Slug: strings.ToLower(name), CreatedAt: time.Now()}
	if err := f.db.Create(&ws).Error; err != nil {
		t.Fatal(err)
	}
	for userID, role := range members {
		if err := f.db.Create(&schemas.WorkspaceMember{WorkspaceID: ws.ID, UserID: userID, Role: role, JoinedAt: time.Now()}).Error; err != nil {
			t.Fatal(err)
		}
	}
	return ws.ID
}

// newProject creates a project with the default board.
func (f *fixture) newProject(t *testing.T, workspaceID int64, name, slug string) int64 {
	t.Helper()
	now := time.Now()
	p := schemas.Project{WorkspaceID: workspaceID, Name: name, Slug: slug, CreatedAt: now, UpdatedAt: now}
	if err := f.db.Create(&p).Error; err != nil {
		t.Fatal(err)
	}
	if err := project.CreateDefaultColumns(f.db, p.ID, now); err != nil {
		t.Fatal(err)
	}
	return p.ID
}

// build makes the MCP service behind the fake Registre.
func (f *fixture) build(t *testing.T) *Service {
	t.Helper()
	verifier, err := mcpauth.Verifier(context.Background(), mcpauth.Config{
		Issuer: f.issuer.URL, Resource: f.resource, ScopePrefix: "opus",
		Identities: identities{users: map[string]int64{subject: f.owner, strangerSubject: f.stranger}},
	})
	if err != nil {
		t.Fatal(err)
	}
	f.domain = Domain{
		Workspaces: workspace.NewService(f.db, t.TempDir()), Projects: project.NewService(f.db),
		Columns: column.NewService(f.db), Tasks: task.NewService(f.db), Labels: label.NewService(f.db),
		Activities: activity.NewService(f.db), DB: f.db,
	}
	service, err := NewService(f.domain, Config{
		Resource: f.resource, Issuer: f.issuer.URL, Version: "test", StateKey: []byte("state-key"), Verifier: verifier,
		Pending: f.pending, PerMinute: f.perMinute,
	})
	if err != nil {
		t.Fatal(err)
	}
	return service
}

// token is the owner's token for this endpoint with scopes, or both when
// none is given.
func (f *fixture) token(t testing.TB, scopes ...string) string {
	if len(scopes) == 0 {
		scopes = []string{"opus.read", "opus.write"}
	}
	return f.issuer.Token(t, f.resource, subject, scopes...)
}

// session opens the owner's client session with every scope and no
// elicitation.
func (f *fixture) session(t *testing.T) *sdk.ClientSession {
	t.Helper()
	return mcptest.Connect(t, f.resource, f.token(t), mcptest.ClientOptions{})
}

// task creates a task in a project through the task routes.
func (f *fixture) task(t testing.TB, projectID, userID int64, title string) task.TaskResponse {
	t.Helper()
	created, err := f.domain.Tasks.Calls().Create(context.Background(), userID, projectID, &task.CreateTaskRequest{Title: title})
	if err != nil {
		t.Fatal(err)
	}
	return *created
}

// exists reports whether a row of model with id is still there.
func (f *fixture) exists(model any, id int64) bool {
	var count int64
	f.db.Model(model).Where("id = ?", id).Count(&count)
	return count == 1
}

// freshTask adds a task and reports 1 once it is gone.
func (f *fixture) freshTask(t testing.TB) (map[string]any, func() int) {
	tk := f.task(t, f.project, f.owner, fmt.Sprintf("Review %s", t.Name()))
	return map[string]any{"task_id": tk.ID}, func() int {
		if f.exists(&schemas.Task{}, tk.ID) {
			return 0
		}
		return 1
	}
}

// freshProject adds a project with one task and reports 1 once it is gone.
func (f *fixture) freshProject(t testing.TB) (map[string]any, func() int) {
	now := time.Now()
	p := schemas.Project{WorkspaceID: f.studio, Name: "Temp " + t.Name(), Slug: fmt.Sprintf("temp-%d", now.UnixNano()), CreatedAt: now, UpdatedAt: now}
	if err := f.db.Create(&p).Error; err != nil {
		t.Fatal(err)
	}
	if err := project.CreateDefaultColumns(f.db, p.ID, now); err != nil {
		t.Fatal(err)
	}
	f.task(t, p.ID, f.owner, "Inside")
	return map[string]any{"project_id": p.ID}, func() int {
		if f.exists(&schemas.Project{}, p.ID) {
			return 0
		}
		return 1
	}
}
