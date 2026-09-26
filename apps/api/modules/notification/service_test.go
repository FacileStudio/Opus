package notification

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
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	err = db.AutoMigrate(
		&schemas.User{},
		&schemas.Workspace{},
		&schemas.WorkspaceMember{},
		&schemas.Notification{},
		&schemas.UserNotificationPreference{},
		&schemas.UserNotificationWorkspaceRule{},
	)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestNotificationLifecycle(t *testing.T) {
	db := newTestDB(t)
	user := schemas.User{Email: "user@facile.studio", Name: "User One"}
	db.Create(&user)
	service := NewService(db)
	auth := &testAuth{userID: user.ID}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	msg := "Task assigned"
	body, _ := json.Marshal(CreateNotificationRequest{
		Type:    "task_assigned",
		Message: &msg,
	})
	req := httptest.NewRequest("POST", "/notification", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create expected 200 got %d: %s", w.Code, w.Body.String())
	}

	var created NotificationResponse
	json.Unmarshal(w.Body.Bytes(), &created)
	if created.IsRead {
		t.Fatalf("expected unread notification")
	}

	req = httptest.NewRequest("PATCH", "/notification/"+strconv.FormatInt(created.ID, 10)+"/read", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("mark read expected 200 got %d", w.Code)
	}

	req = httptest.NewRequest("GET", "/notification", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list expected 200 got %d", w.Code)
	}

	req = httptest.NewRequest("PATCH", "/notification/read-all", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("read-all expected 200 got %d", w.Code)
	}

	req = httptest.NewRequest("DELETE", "/notification/clear-all", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("clear-all expected 200 got %d", w.Code)
	}
}

func TestPreferencesLifecycle(t *testing.T) {
	db := newTestDB(t)
	user := schemas.User{Email: "user@facile.studio", Name: "User One"}
	db.Create(&user)
	ws := schemas.Workspace{Name: "W1", Slug: "w1"}
	db.Create(&ws)
	member := schemas.WorkspaceMember{WorkspaceID: ws.ID, UserID: user.ID, Role: "admin"}
	db.Create(&member)

	service := NewService(db)
	auth := &testAuth{userID: user.ID}
	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	req := httptest.NewRequest("GET", "/notification-preferences", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get preferences expected 200 got %d", w.Code)
	}

	enabled := true
	updateBody, _ := json.Marshal(UpdatePreferencesRequest{EmailEnabled: &enabled})
	req = httptest.NewRequest("PUT", "/notification-preferences", bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update preferences expected 200 got %d", w.Code)
	}

	wsRuleBody, _ := json.Marshal(UpsertWorkspaceRuleRequest{
		IsActive:    true,
		ProjectMode: "all",
	})
	req = httptest.NewRequest("PUT", "/notification-preferences/workspaces/"+strconv.FormatInt(ws.ID, 10), bytes.NewReader(wsRuleBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("upsert workspace rule expected 200 got %d", w.Code)
	}

	req = httptest.NewRequest("DELETE", "/notification-preferences/workspaces/"+strconv.FormatInt(ws.ID, 10), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete workspace rule expected 200 got %d", w.Code)
	}
}

func TestNotificationValidationTable(t *testing.T) {
	db := newTestDB(t)
	user := schemas.User{Email: "user@facile.studio", Name: "User One"}
	db.Create(&user)
	service := NewService(db)
	auth := &testAuth{userID: user.ID}
	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	cases := []struct {
		name       string
		method     string
		path       string
		body       any
		wantStatus int
	}{
		{
			name:       "missing notification type",
			method:     "POST",
			path:       "/notification",
			body:       CreateNotificationRequest{Type: ""},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid project mode in workspace rule",
			method:     "PUT",
			path:       "/notification-preferences/workspaces/1",
			body:       UpsertWorkspaceRuleRequest{ProjectMode: "invalid-mode"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, w.Code)
			}
		})
	}
}

