package timeentry

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
		&schemas.Task{},
		&schemas.TimeEntry{},
	)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func seedFixture(t *testing.T, db *gorm.DB) (int64, int64, int64) {
	t.Helper()
	user := schemas.User{Email: "user@facile.studio", Name: "User One"}
	db.Create(&user)
	ws := schemas.Workspace{Name: "W1", Slug: "w1"}
	db.Create(&ws)
	member := schemas.WorkspaceMember{WorkspaceID: ws.ID, UserID: user.ID, Role: "admin"}
	db.Create(&member)
	proj := schemas.Project{WorkspaceID: ws.ID, Name: "P1", Slug: "p1"}
	db.Create(&proj)
	task := schemas.Task{ProjectID: proj.ID, WorkspaceID: ws.ID, Title: "Task 1", Status: "todo"}
	db.Create(&task)
	return user.ID, ws.ID, task.ID
}

func TestTimeEntryCreateAndList(t *testing.T) {
	db := newTestDB(t)
	userID, _, taskID := seedFixture(t, db)
	service := NewService(db)
	auth := &testAuth{userID: userID, email: "user@facile.studio"}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	body, _ := json.Marshal(CreateTimeEntryRequest{
		TaskID:      taskID,
		StartTime:   start,
		EndTime:     &end,
		Description: "Debugging",
	})
	req := httptest.NewRequest("POST", "/time-entry", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("create expected 200 got %d: %s", w.Code, w.Body.String())
	}
	var created TimeEntryResponse
	json.Unmarshal(w.Body.Bytes(), &created)
	if created.Duration <= 0 {
		t.Fatalf("expected positive duration, got %d", created.Duration)
	}

	req = httptest.NewRequest("GET", "/time-entry/task/"+strconv.FormatInt(taskID, 10), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list expected 200 got %d", w.Code)
	}
}

func TestTimeEntryUpdateAndDelete(t *testing.T) {
	db := newTestDB(t)
	userID, _, taskID := seedFixture(t, db)
	service := NewService(db)
	auth := &testAuth{userID: userID, email: "user@facile.studio"}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	start := time.Now().Add(-2 * time.Hour)
	entry, _ := service.Create(context.Background(), userID, &CreateTimeEntryRequest{
		TaskID:    taskID,
		StartTime: start,
	})

	newDesc := "Updated Work"
	body, _ := json.Marshal(UpdateTimeEntryRequest{Description: &newDesc})
	req := httptest.NewRequest("PUT", "/time-entry/"+strconv.FormatInt(entry.ID, 10), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update expected 200 got %d", w.Code)
	}

	req = httptest.NewRequest("GET", "/time-entry/"+strconv.FormatInt(entry.ID, 10), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get expected 200 got %d", w.Code)
	}

	req = httptest.NewRequest("DELETE", "/time-entry/"+strconv.FormatInt(entry.ID, 10), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete expected 200 got %d", w.Code)
	}
}

func TestTimeEntryUnauthorized(t *testing.T) {
	db := newTestDB(t)
	_, _, taskID := seedFixture(t, db)
	service := NewService(db)
	auth := &testAuth{userID: 999, email: "other@facile.studio"}

	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	body, _ := json.Marshal(CreateTimeEntryRequest{
		TaskID:    taskID,
		StartTime: time.Now(),
	})
	req := httptest.NewRequest("POST", "/time-entry", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d", w.Code)
	}
}

func TestTimeEntryValidationTable(t *testing.T) {
	db := newTestDB(t)
	userID, _, taskID := seedFixture(t, db)
	service := NewService(db)
	auth := &testAuth{userID: userID, email: "user@facile.studio"}
	router := chi.NewRouter()
	RegisterRoutes(router, service, auth)

	now := time.Now()
	before := now.Add(-1 * time.Hour)
	after := now.Add(1 * time.Hour)

	cases := []struct {
		name       string
		req        CreateTimeEntryRequest
		wantStatus int
	}{
		{
			name:       "missing task id",
			req:        CreateTimeEntryRequest{TaskID: 0, StartTime: now},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "end before start",
			req:        CreateTimeEntryRequest{TaskID: taskID, StartTime: now, EndTime: &before},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "valid with duration",
			req:        CreateTimeEntryRequest{TaskID: taskID, StartTime: now, EndTime: &after},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.req)
			req := httptest.NewRequest("POST", "/time-entry", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, w.Code)
			}
		})
	}
}
