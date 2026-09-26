package sharelink

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

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
		&schemas.Task{},
		&schemas.ShareLink{},
	)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func seedFixture(t *testing.T, db *gorm.DB) (int64, int64, int64) {
	t.Helper()
	user := schemas.User{Email: "test@facile.studio", Name: "Test"}
	db.Create(&user)
	ws := schemas.Workspace{Name: "W1", Slug: "w1"}
	db.Create(&ws)
	member := schemas.WorkspaceMember{WorkspaceID: ws.ID, UserID: user.ID, Role: "admin"}
	db.Create(&member)
	proj := schemas.Project{WorkspaceID: ws.ID, Name: "P1", Slug: "p1"}
	db.Create(&proj)
	col := schemas.Column{ProjectID: proj.ID, Name: "Todo", Slug: "todo", Position: 0}
	db.Create(&col)
	task := schemas.Task{ProjectID: proj.ID, WorkspaceID: ws.ID, ColumnID: &col.ID, Title: "T1", Status: "todo"}
	db.Create(&task)
	return user.ID, ws.ID, proj.ID
}

func TestShareLinkCreateAndList(t *testing.T) {
	db := newTestDB(t)
	userID, _, projID := seedFixture(t, db)
	service := NewService(db)
	auth := &testAuth{userID: userID, email: "test@facile.studio"}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	body, _ := json.Marshal(CreateShareLinkRequest{ProjectID: projID})
	req := httptest.NewRequest("POST", "/share-link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create expected 200 got %d", w.Code)
	}

	req = httptest.NewRequest("GET", "/share-link?projectId="+strconv.FormatInt(projID, 10), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list expected 200 got %d", w.Code)
	}
}

func TestShareLinkUpdateAndDelete(t *testing.T) {
	db := newTestDB(t)
	userID, _, projID := seedFixture(t, db)
	service := NewService(db)
	auth := &testAuth{userID: userID, email: "test@facile.studio"}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	link, _ := service.Create(context.Background(), userID, &CreateShareLinkRequest{ProjectID: projID})
	newLabel := "Client Review"
	updateBody, _ := json.Marshal(UpdateShareLinkRequest{Label: &newLabel})
	req := httptest.NewRequest("PUT", "/share-link/"+strconv.FormatInt(link.ID, 10), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update expected 200 got %d", w.Code)
	}

	req = httptest.NewRequest("GET", "/shared-project/"+link.Token, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("public access expected 200 got %d", w.Code)
	}

	req = httptest.NewRequest("DELETE", "/share-link/"+strconv.FormatInt(link.ID, 10), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete expected 200 got %d", w.Code)
	}
}

func TestShareLinkAccessDenied(t *testing.T) {
	db := newTestDB(t)
	_, _, projID := seedFixture(t, db)
	service := NewService(db)
	auth := &testAuth{userID: 999, email: "other@facile.studio"}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	body, _ := json.Marshal(CreateShareLinkRequest{ProjectID: projID})
	req := httptest.NewRequest("POST", "/share-link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d", w.Code)
	}
}

func TestShareLinkExpiry(t *testing.T) {
	db := newTestDB(t)
	userID, _, projID := seedFixture(t, db)
	past := time.Now().Add(-1 * time.Hour)
	link := schemas.ShareLink{
		ProjectID: projID,
		Token:     "expired-token",
		ExpiresAt: &past,
		IsActive:  true,
		CreatedBy: userID,
	}
	db.Create(&link)

	service := NewService(db)
	auth := &testAuth{userID: userID, email: "test@facile.studio"}
	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	req := httptest.NewRequest("GET", "/shared-project/expired-token", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for expired share link, got %d", w.Code)
	}
}
