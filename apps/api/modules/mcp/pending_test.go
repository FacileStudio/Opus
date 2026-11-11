package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/go-chi/chi/v5"
)

// pendingDelete asks to delete a fresh task from a client that cannot
// confirm, and returns the task and the action id.
func pendingDelete(t *testing.T, f *fixture) (int64, string) {
	t.Helper()
	args, _ := f.freshTask(t)
	var out actionOutput
	call(t, f.session(t), toolDeleteTask, args, &out)
	return args["task_id"].(int64), out.ConfirmURL[strings.LastIndex(out.ConfirmURL, "=")+1:]
}

// onPage calls a confirm page route, such as "/" or "/{id}/confirm", as the
// owner, or as a delegated token when actor is set.
func onPage(f *fixture, method, path, actor string) *httptest.ResponseRecorder {
	router := chi.NewRouter()
	f.service.page.Routes(router)
	req := httptest.NewRequest(method, path, nil)
	ctx := authcontext.WithIdentity(req.Context(), authcontext.Identity{UserID: f.owner, Actor: actor})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req.WithContext(ctx))
	return recorder
}

// confirmAs confirms the action id on the page.
func confirmAs(f *fixture, id, actor string) *httptest.ResponseRecorder {
	return onPage(f, http.MethodPost, "/"+id+"/confirm", actor)
}

// TestConfirmPageDeletesOnce checks the person confirms in Opus, the task
// goes once, a second confirmation finds nothing, and an agent cannot
// confirm its own request.
func TestConfirmPageDeletesOnce(t *testing.T) {
	f := newFixture(t)
	taskID, id := pendingDelete(t, f)
	if got := confirmAs(f, id, "bulle"); got.Code != http.StatusForbidden || !f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("a delegated token must not confirm, got %d", got.Code)
	}
	list := onPage(f, http.MethodGet, "/", "")
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"state":"ready"`) || !strings.Contains(list.Body.String(), `"ref":"web-`) {
		t.Fatalf("want one ready action with its task on the page, got %d %s", list.Code, list.Body)
	}
	if got := confirmAs(f, id, ""); got.Code != http.StatusOK || !strings.Contains(got.Body.String(), `"status":"deleted"`) || f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("want the task deleted, got %d %s", got.Code, got.Body)
	}
	if again := confirmAs(f, id, ""); again.Code != http.StatusNotFound {
		t.Fatalf("a confirmation is single use, got %d", again.Code)
	}
}

// TestEditedTaskIsNotDeleted checks an edit after the request closes it
// instead of deleting a version the person never saw.
func TestEditedTaskIsNotDeleted(t *testing.T) {
	f := newFixture(t)
	taskID, id := pendingDelete(t, f)
	var out writeOutput
	call(t, f.session(t), toolUpdateTask, map[string]any{"task_id": taskID, "title": "Renamed"}, &out)
	if got := confirmAs(f, id, ""); got.Code != http.StatusPreconditionFailed || !f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("want 412 and the task kept, got %d", got.Code)
	}
}

// TestDismissKeepsTheTask checks dismissing closes the action and keeps the
// task.
func TestDismissKeepsTheTask(t *testing.T) {
	f := newFixture(t)
	taskID, id := pendingDelete(t, f)
	if got := onPage(f, http.MethodPost, "/"+id+"/dismiss", ""); got.Code != http.StatusOK || !f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("want the action dismissed and the task kept, got %d", got.Code)
	}
	if got := confirmAs(f, id, ""); got.Code != http.StatusNotFound || !f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("a dismissed action cannot be confirmed, got %d", got.Code)
	}
}
