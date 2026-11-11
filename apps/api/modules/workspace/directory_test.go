package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

const testIssuer = "https://sso.test"

func newPostgresService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()
	db := dbtest.Open(t)
	return NewService(db, t.TempDir()), db
}

func addIdentity(t *testing.T, db *gorm.DB, userID int64, subject string) {
	t.Helper()
	if err := db.Exec("INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)",
		userID, testIssuer, subject).Error; err != nil {
		t.Fatalf("add identity: %v", err)
	}
}

type dbState struct {
	Users       int64
	Identities  int64
	Workspaces  []string
	Memberships []string
}

func readState(t *testing.T, db *gorm.DB) dbState {
	t.Helper()
	var state dbState
	db.Model(&schemas.User{}).Count(&state.Users)
	db.Table("porte_identities").Count(&state.Identities)
	db.Raw("SELECT id || ':' || name || ':' || coalesce(registre_space_id, '') FROM workspaces ORDER BY id").
		Scan(&state.Workspaces)
	db.Raw("SELECT workspace_id || ':' || user_id || ':' || role FROM workspace_members ORDER BY 1").
		Scan(&state.Memberships)
	return state
}

func roleOf(t *testing.T, db *gorm.DB, workspaceID, userID int64) string {
	t.Helper()
	var role string
	db.Model(&schemas.WorkspaceMember{}).Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Select("role").Scan(&role)
	return role
}

type applyFixture struct {
	alice, bob, carol, dave, studio, invitation int64
}

func seedApplyFixture(t *testing.T, db *gorm.DB) applyFixture {
	t.Helper()
	alice := createUser(t, db, "alice@facile.studio", "Alice")
	addIdentity(t, db, alice, "sub-alice")
	bob := createUser(t, db, "bob@facile.studio", "Bob")
	carol := createUser(t, db, "carol@facile.studio", "Carol")
	dave := createUser(t, db, "dave@facile.studio", "Dave")
	addIdentity(t, db, dave, "sub-dave")
	studio := createWorkspace(t, db, "Studio")
	addMember(t, db, studio, alice, "admin")
	addMember(t, db, studio, carol, "member")
	addMember(t, db, studio, dave, "member")
	invitation := schemas.Invitation{WorkspaceID: studio, Email: "zoe@facile.studio", Role: "member",
		Status: "pending", InviterID: alice, CreatedAt: time.Now()}
	if err := db.Create(&invitation).Error; err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	return applyFixture{alice: alice, bob: bob, carol: carol, dave: dave, studio: studio, invitation: invitation.ID}
}

func applySnapshot() directory.Snapshot {
	studio := []directory.Member{
		{Subject: "sub-alice", Email: "alice@facile.studio", Role: directory.RoleOwner},
		{Subject: "sub-bob", Email: "BOB@facile.studio", Role: directory.RoleMember},
		{Subject: "sub-eve", Email: "eve@facile.studio", Name: "Eve",
			Picture: testIssuer + "/avatars/sub-eve", Role: directory.RoleAdmin},
	}
	ops := []directory.Member{{Subject: "sub-alice", Email: "alice@facile.studio", Role: directory.RoleMember}}
	return directory.Snapshot{Version: 3, Spaces: []directory.Space{
		{ID: "space-studio", Name: "studio", Members: studio},
		{ID: "space-ops", Name: "Ops", Members: ops},
	}}
}

func TestDirectoryApply(t *testing.T) {
	service, db := newPostgresService(t)
	ctx := context.Background()
	f := seedApplyFixture(t, db)
	dir := NewDirectory(service, directory.Client{Issuer: testIssuer}, slog.New(slog.DiscardHandler))
	if err := dir.Apply(ctx, applySnapshot()); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	checkStudio(t, db, f)
	checkOpsAndInvitation(t, db, f)

	before := readState(t, db)
	if err := dir.Apply(ctx, applySnapshot()); err != nil {
		t.Fatalf("second Apply: %v", err)
	}
	if after := readState(t, db); fmt.Sprint(after) != fmt.Sprint(before) {
		t.Fatalf("second Apply changed state:\nbefore %+v\nafter  %+v", before, after)
	}

	_, err := service.UpdateMemberRole(ctx, f.alice, f.studio, f.carol, "admin")
	if errors.Status(err) != http.StatusConflict || !strings.Contains(err.Error(), testIssuer+"/spaces/space-studio") {
		t.Fatalf("role change on a linked workspace = %v, want a 409 pointing to Registre", err)
	}
}

func checkStudio(t *testing.T, db *gorm.DB, f applyFixture) {
	t.Helper()
	studio, alice, bob, carol, dave := f.studio, f.alice, f.bob, f.carol, f.dave
	var linked schemas.Workspace
	db.First(&linked, studio)
	if linked.RegistreSpaceID == nil || *linked.RegistreSpaceID != "space-studio" || linked.Name != "studio" {
		t.Fatalf("Studio = %+v, want linked to space-studio by name and renamed", linked)
	}
	if got := roleOf(t, db, studio, alice); got != "owner" {
		t.Fatalf("alice role = %q, want owner", got)
	}
	var bobIdentity int64
	db.Table("porte_identities").Where("subject = ? AND user_id = ?", "sub-bob", bob).Count(&bobIdentity)
	if bobIdentity != 1 || roleOf(t, db, studio, bob) != "member" {
		t.Fatalf("bob not linked by email: identity rows %d role %q", bobIdentity, roleOf(t, db, studio, bob))
	}
	var eve schemas.User
	db.Where("email = ?", "eve@facile.studio").First(&eve)
	if eve.ID == 0 || eve.Name != "Eve" || eve.Image != testIssuer+"/avatars/sub-eve" || roleOf(t, db, studio, eve.ID) != "admin" {
		t.Fatalf("eve not pre-provisioned as admin: %+v", eve)
	}
	if roleOf(t, db, studio, carol) != "member" {
		t.Fatal("carol, a local-only member, was removed")
	}
	if roleOf(t, db, studio, dave) != "" {
		t.Fatal("dave has a Registre identity and left the space but kept his membership")
	}
}

func checkOpsAndInvitation(t *testing.T, db *gorm.DB, f applyFixture) {
	t.Helper()
	alice := f.alice
	var ops schemas.Workspace
	db.Where("registre_space_id = ?", "space-ops").First(&ops)
	if ops.ID == 0 || ops.Name != "Ops" || roleOf(t, db, ops.ID, alice) != "member" {
		t.Fatalf("Ops not created with alice: %+v", ops)
	}
	var status string
	db.Model(&schemas.Invitation{}).Where("id = ?", f.invitation).Select("status").Scan(&status)
	if status != "canceled" {
		t.Fatalf("pending invitation status = %q, want canceled", status)
	}
}

type fakeRegistre struct {
	mu       sync.Mutex
	calls    []string
	imported []directory.ImportSpace
	failures int
	synced   chan struct{}
}

func (f *fakeRegistre) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, r.URL.Path)
	if r.URL.Path == "/api/directory/import" {
		if f.failures > 0 {
			f.failures--
			http.NotFound(w, r)
			return
		}
		var body struct {
			Spaces []directory.ImportSpace `json:"spaces"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.imported = body.Spaces
	}
	if r.URL.Path == "/api/directory/snapshot" {
		select {
		case f.synced <- struct{}{}:
		default:
		}
	}
	if err := json.NewEncoder(w).Encode(directory.Snapshot{Version: 1}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func TestDirectoryImportsBeforeSync(t *testing.T) {
	service, db := newPostgresService(t)
	alice := createUser(t, db, "alice@facile.studio", "Alice")
	addIdentity(t, db, alice, "sub-alice")
	studio := createWorkspace(t, db, "Studio")
	addMember(t, db, studio, alice, "owner")

	fake := &fakeRegistre{failures: 2, synced: make(chan struct{}, 1)}
	server := httptest.NewServer(fake)
	defer server.Close()
	dir := NewDirectory(service, directory.Client{Issuer: server.URL}, slog.New(slog.DiscardHandler))
	dir.issuer = testIssuer
	dir.retryDelay = 10 * time.Millisecond
	if service.SpacesURL() != "" {
		t.Fatal("Opus reported Registre management before the import succeeded")
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { dir.Run(ctx); close(done) }()
	select {
	case <-fake.synced:
	case <-time.After(5 * time.Second):
		t.Fatal("no sync after the import succeeded")
	}
	cancel()
	<-done

	fake.mu.Lock()
	defer fake.mu.Unlock()
	want := []string{"/api/directory/import", "/api/directory/import", "/api/directory/import", "/api/directory/snapshot"}
	if fmt.Sprint(fake.calls[:4]) != fmt.Sprint(want) {
		t.Fatalf("calls = %v, want three imports before the first snapshot", fake.calls)
	}
	if len(fake.imported) != 1 || fake.imported[0].Name != "Studio" ||
		fake.imported[0].Members[0] != (directory.ImportMember{Subject: "sub-alice", Role: "owner"}) {
		t.Fatalf("imported = %+v, want Studio with alice as owner by subject", fake.imported)
	}
	if service.SpacesURL() != server.URL+"/spaces" {
		t.Fatalf("SpacesURL = %q after the import, want the Registre spaces page", service.SpacesURL())
	}
}
