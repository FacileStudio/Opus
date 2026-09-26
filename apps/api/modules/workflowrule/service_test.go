package workflowrule

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testAuth struct {
	userID int64
	email  string
}

func (a *testAuth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := porte.WithIdentity(r.Context(), porte.Identity{UserID: a.userID, SessionID: 1})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *testAuth) IdentityForUser(_ context.Context, userID int64) (int64, string, error) {
	return userID, a.email, nil
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	err = db.AutoMigrate(
		&schemas.User{},
		&schemas.Workspace{},
		&schemas.WorkspaceMember{},
		&schemas.Project{},
		&schemas.Column{},
		&schemas.WorkflowRule{},
	)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

type testFixture struct {
	userID    int64
	wsID      int64
	projectID int64
	columnID  int64
}

func seedFixture(t *testing.T, db *gorm.DB) testFixture {
	t.Helper()
	user := schemas.User{Email: "user@facile.studio", Name: "User One"}
	db.Create(&user)
	ws := schemas.Workspace{Name: "W1", Slug: "w1"}
	db.Create(&ws)
	member := schemas.WorkspaceMember{WorkspaceID: ws.ID, UserID: user.ID, Role: "admin"}
	db.Create(&member)
	proj := schemas.Project{WorkspaceID: ws.ID, Name: "P1", Slug: "p1"}
	db.Create(&proj)
	col := schemas.Column{ProjectID: proj.ID, Name: "Done", Slug: "done", Position: 1}
	db.Create(&col)
	return testFixture{
		userID:    user.ID,
		wsID:      ws.ID,
		projectID: proj.ID,
		columnID:  col.ID,
	}
}

func TestWorkflowRuleUpsertAndList(t *testing.T) {
	db := newTestDB(t)
	f := seedFixture(t, db)
	service := NewService(db)
	auth := &testAuth{userID: f.userID, email: "user@facile.studio"}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	body, _ := json.Marshal(UpsertWorkflowRuleRequest{
		IntegrationType: "github",
		EventType:       "pull_request.merged",
		ColumnID:        f.columnID,
	})
	req := httptest.NewRequest("PUT", "/workflow-rule/"+strconv.FormatInt(f.projectID, 10), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("upsert expected 200 got %d: %s", w.Code, w.Body.String())
	}
	var created WorkflowRuleResponse
	json.Unmarshal(w.Body.Bytes(), &created)
	if created.ColumnID != f.columnID {
		t.Fatalf("expected columnId %d, got %d", f.columnID, created.ColumnID)
	}

	req = httptest.NewRequest("GET", "/workflow-rule/"+strconv.FormatInt(f.projectID, 10), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list expected 200 got %d", w.Code)
	}
}

func TestWorkflowRuleDelete(t *testing.T) {
	db := newTestDB(t)
	f := seedFixture(t, db)
	service := NewService(db)
	auth := &testAuth{userID: f.userID, email: "user@facile.studio"}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	rule, _ := service.Upsert(context.Background(), f.userID, f.projectID, &UpsertWorkflowRuleRequest{
		IntegrationType: "github",
		EventType:       "issue.closed",
		ColumnID:        f.columnID,
	})

	req := httptest.NewRequest("DELETE", "/workflow-rule/"+strconv.FormatInt(rule.ID, 10), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete expected 200 got %d", w.Code)
	}
}

func TestWorkflowRuleUnauthorized(t *testing.T) {
	db := newTestDB(t)
	f := seedFixture(t, db)
	service := NewService(db)
	auth := &testAuth{userID: 999, email: "other@facile.studio"}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	body, _ := json.Marshal(UpsertWorkflowRuleRequest{
		IntegrationType: "github",
		EventType:       "pr.merged",
		ColumnID:        f.columnID,
	})
	req := httptest.NewRequest("PUT", "/workflow-rule/"+strconv.FormatInt(f.projectID, 10), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d", w.Code)
	}
}
