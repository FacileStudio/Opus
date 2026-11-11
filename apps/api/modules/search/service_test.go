package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type testAuth struct {
	userID int64
}

func (a *testAuth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := porte.WithIdentity(r.Context(), porte.Identity{UserID: a.userID, SessionID: 1})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *testAuth) IdentityForUser(_ context.Context, userID int64) (int64, string, error) {
	return userID, "user@facile.studio", nil
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := dbtest.Open(t)
	return db
}

func seedFixture(t *testing.T, db *gorm.DB) (int64, int64) {
	t.Helper()
	user := schemas.User{Email: "searcher@facile.studio", Name: "Searcher"}
	db.Create(&user)
	ws := schemas.Workspace{Name: "Acme Corp", Slug: "acme"}
	db.Create(&ws)
	member := schemas.WorkspaceMember{WorkspaceID: ws.ID, UserID: user.ID, Role: "member"}
	db.Create(&member)

	proj := schemas.Project{WorkspaceID: ws.ID, Name: "Frontend App", Slug: "frontend"}
	db.Create(&proj)

	task := schemas.Task{
		WorkspaceID: ws.ID,
		ProjectID:   proj.ID,
		Title:       "Design authentication modal",
		Description: "Use Base UI dialog component",
		Status:      "in-progress",
		Priority:    "high",
		Number:      101,
	}
	db.Create(&task)

	comment := schemas.Comment{
		TaskID:  task.ID,
		UserID:  user.ID,
		Content: "Remember to check responsive layout",
	}
	db.Create(&comment)

	return user.ID, ws.ID
}

func TestSearchTasksAndProjects(t *testing.T) {
	db := newTestDB(t)
	userID, wsID := seedFixture(t, db)
	service := NewService(db)
	auth := &testAuth{userID: userID}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	req := httptest.NewRequest("GET", "/search?q=authentication", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("search expected 200 got %d: %s", w.Code, w.Body.String())
	}

	var resp SearchResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.TotalCount == 0 {
		t.Fatalf("expected at least 1 task match")
	}
	if resp.Results[0].Type != "task" {
		t.Fatalf("expected type task, got %s", resp.Results[0].Type)
	}

	req = httptest.NewRequest("GET", "/search?q=frontend&type=projects", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("project search expected 200 got %d", w.Code)
	}

	req = httptest.NewRequest("GET", "/search?q=modal&workspaceId="+strconv.FormatInt(wsID, 10), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("scoped search expected 200 got %d", w.Code)
	}
}

func TestSearchWorkspaceIsolation(t *testing.T) {
	db := newTestDB(t)
	seedFixture(t, db)
	service := NewService(db)
	auth := &testAuth{userID: 999}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	req := httptest.NewRequest("GET", "/search?q=authentication", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}

	var resp SearchResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.TotalCount != 0 {
		t.Fatalf("expected 0 results for non-member user, got %d", resp.TotalCount)
	}
}

func TestSearchValidationTable(t *testing.T) {
	db := newTestDB(t)
	userID, _ := seedFixture(t, db)
	service := NewService(db)
	auth := &testAuth{userID: userID}
	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	cases := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "empty query string",
			path:       "/search?q=",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing query parameter",
			path:       "/search",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "search comments type",
			path:       "/search?q=responsive&type=comments",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, w.Code)
			}
		})
	}
}
