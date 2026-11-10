package mcp

import (
	"context"
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

// onPage calls a confirm page handler as the owner, or as a delegated token
// when actor is set.
func onPage(f *fixture, handler http.HandlerFunc, id, actor string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/mcp/pending/"+id+"/confirm", nil)
	ctx := authcontext.WithIdentity(req.Context(), authcontext.Identity{UserID: f.owner, Actor: actor})
	routes := chi.NewRouteContext()
	routes.URLParams.Add("id", id)
	recorder := httptest.NewRecorder()
	handler(recorder, req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, routes)))
	return recorder
}

// TestConfirmPageDeletesOnce checks the person confirms in Opus, the task
// goes once, a second confirmation finds nothing, and an agent cannot
// confirm its own request.
func TestConfirmPageDeletesOnce(t *testing.T) {
	f := newFixture(t)
	taskID, id := pendingDelete(t, f)
	if got := onPage(f, f.service.handleConfirm, id, "bulle"); got.Code != http.StatusForbidden || !f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("a delegated token must not confirm, got %d", got.Code)
	}
	list := onPage(f, f.service.handleList, "", "")
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"state":"ready"`) || !strings.Contains(list.Body.String(), `"ref":"web-`) {
		t.Fatalf("want one ready action with its task on the page, got %d %s", list.Code, list.Body)
	}
	if got := onPage(f, f.service.handleConfirm, id, ""); got.Code != http.StatusOK || f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("want the task deleted, got %d %s", got.Code, got.Body)
	}
	if again := onPage(f, f.service.handleConfirm, id, ""); again.Code != http.StatusNotFound {
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
	if got := onPage(f, f.service.handleConfirm, id, ""); got.Code != http.StatusPreconditionFailed || !f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("want 412 and the task kept, got %d", got.Code)
	}
}

// TestDismissKeepsTheTask checks dismissing closes the action and keeps the
// task.
func TestDismissKeepsTheTask(t *testing.T) {
	f := newFixture(t)
	taskID, id := pendingDelete(t, f)
	if got := onPage(f, f.service.handleDismiss, id, ""); got.Code != http.StatusOK || !f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("want the action dismissed and the task kept, got %d", got.Code)
	}
	if got := onPage(f, f.service.handleConfirm, id, ""); got.Code != http.StatusNotFound || !f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("a dismissed action cannot be confirmed, got %d", got.Code)
	}
}
