package task

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
)

type testRouter struct {
	service   *Service
	router    chi.Router
	projectID int64
	colSlug   string
}

func setupTestRouter(t *testing.T, userID int64) testRouter {
	t.Helper()
	service, db := newTestService(t)
	uID, projectID, col := seedProject(t, db)
	r := chi.NewRouter()
	RegisterRoutes(r, service, mockAuth{userID: uID})
	return testRouter{
		service:   service,
		router:    r,
		projectID: projectID,
		colSlug:   col.Slug,
	}
}

func TestHTTPTaskLifecycle(t *testing.T) {
	env := setupTestRouter(t, 1)

	createBody, _ := json.Marshal(CreateTaskRequest{
		Title:  "HTTP Task",
		Status: env.colSlug,
	})
	req := httptest.NewRequest("POST", "/task/"+strconv.FormatInt(env.projectID, 10), bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on create, got %d: %s", w.Code, w.Body.String())
	}

	var created TaskResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created task: %v", err)
	}

	req = httptest.NewRequest("GET", "/task/"+strconv.FormatInt(created.ID, 10), nil)
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on get, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest("GET", "/task/tasks/"+strconv.FormatInt(env.projectID, 10), nil)
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on list, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest("DELETE", "/task/"+strconv.FormatInt(created.ID, 10), nil)
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHTTPBulkAndFieldUpdates(t *testing.T) {
	env := setupTestRouter(t, 1)
	ctx := t.Context()

	t1, err := env.service.controller.create(ctx, 1, env.projectID, &CreateTaskRequest{
		Title:  "Bulk 1",
		Status: env.colSlug,
	})
	if err != nil {
		t.Fatalf("create t1: %v", err)
	}

	statusBody, _ := json.Marshal(UpdateStatusRequest{Status: "planned"})
	req := httptest.NewRequest("PUT", "/task/status/"+strconv.FormatInt(t1.ID, 10), bytes.NewReader(statusBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on status update, got %d: %s", w.Code, w.Body.String())
	}

	val := "urgent"
	bulkBody, _ := json.Marshal(BulkUpdateRequest{
		TaskIDs:   []int64{t1.ID},
		Operation: "updatePriority",
		Value:     &val,
	})
	req = httptest.NewRequest("PATCH", "/task/bulk", bytes.NewReader(bulkBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on bulk patch, got %d: %s", w.Code, w.Body.String())
	}
}
