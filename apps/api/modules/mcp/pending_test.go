package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
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

// ownSession is the owner's own browser session, the only caller the confirm
// page accepts.
func ownSession(f *fixture) porte.Identity {
	return porte.Identity{UserID: f.owner, SessionID: 1, BrowserSession: true}
}

// onPage calls a confirm page route, such as "/" or "/{id}/confirm", as who.
func onPage(f *fixture, method, path string, who porte.Identity) *httptest.ResponseRecorder {
	router := chi.NewRouter()
	f.service.page.Routes(router)
	req := httptest.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req.WithContext(porte.WithIdentity(req.Context(), who)))
	return recorder
}

// confirmAs confirms the action id on the page as who.
func confirmAs(f *fixture, id string, who porte.Identity) *httptest.ResponseRecorder {
	return onPage(f, http.MethodPost, "/"+id+"/confirm", who)
}

// TestConfirmPageDeletesOnce checks the person confirms in Opus, the task
// goes once, a second confirmation finds nothing, and neither an agent nor
// a CLI session can confirm the request.
func TestConfirmPageDeletesOnce(t *testing.T) {
	f := newFixture(t)
	taskID, id := pendingDelete(t, f)
	delegated := ownSession(f)
	delegated.Actor = "bulle"
	if got := confirmAs(f, id, delegated); got.Code != http.StatusForbidden || !f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("a delegated token must not confirm, got %d", got.Code)
	}
	cli := ownSession(f)
	cli.BrowserSession = false
	if got := confirmAs(f, id, cli); got.Code != http.StatusForbidden || !f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("a CLI session must not confirm, got %d", got.Code)
	}
	list := onPage(f, http.MethodGet, "/", ownSession(f))
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"state":"ready"`) || !strings.Contains(list.Body.String(), `"ref":"web-`) {
		t.Fatalf("want one ready action with its task on the page, got %d %s", list.Code, list.Body)
	}
	if got := confirmAs(f, id, ownSession(f)); got.Code != http.StatusOK || !strings.Contains(got.Body.String(), `"status":"deleted"`) || f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("want the task deleted, got %d %s", got.Code, got.Body)
	}
	if again := confirmAs(f, id, ownSession(f)); again.Code != http.StatusNotFound {
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
	if got := confirmAs(f, id, ownSession(f)); got.Code != http.StatusPreconditionFailed || !f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("want 412 and the task kept, got %d", got.Code)
	}
}

// TestDismissKeepsTheTask checks dismissing closes the action and keeps the
// task.
func TestDismissKeepsTheTask(t *testing.T) {
	f := newFixture(t)
	taskID, id := pendingDelete(t, f)
	if got := onPage(f, http.MethodPost, "/"+id+"/dismiss", ownSession(f)); got.Code != http.StatusOK || !f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("want the action dismissed and the task kept, got %d", got.Code)
	}
	if got := confirmAs(f, id, ownSession(f)); got.Code != http.StatusNotFound || !f.exists(&schemas.Task{}, taskID) {
		t.Fatalf("a dismissed action cannot be confirmed, got %d", got.Code)
	}
}
