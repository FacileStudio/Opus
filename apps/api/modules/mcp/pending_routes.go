package mcp

import (
	"context"
	"net/http"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/modules/task"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	"github.com/go-chi/chi/v5"
)

// webActor marks, in the audit, a confirmation made on the confirm page.
const webActor = "opus-web"

// pendingUser is the signed-in person of a confirm page request. A delegated
// token is refused: an agent must never confirm its own request.
func pendingUser(r *http.Request) (mcpkit.Caller, error) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		return mcpkit.Caller{}, errors.Unauthorized("missing auth token")
	}
	if identity.Actor != "" {
		return mcpkit.Caller{}, errors.Forbidden("a delegated token cannot confirm actions")
	}
	return mcpkit.Caller{UserID: strconv.FormatInt(identity.UserID, 10), Actor: webActor}, nil
}

// handleList answers the person's open actions with their targets.
func (s *Service) handleList(w http.ResponseWriter, r *http.Request) {
	who, err := pendingUser(r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	list, err := s.tools.pendingList(r.Context(), who.UserID)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, list)
}

// handleConfirm runs one action.
func (s *Service) handleConfirm(w http.ResponseWriter, r *http.Request) {
	who, err := pendingUser(r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if _, err := s.tools.confirm.Confirm(r.Context(), who, chi.URLParam(r, "id")); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, PendingResult{Status: "deleted"})
}

// handleDismiss closes one action without running it. The target stays.
func (s *Service) handleDismiss(w http.ResponseWriter, r *http.Request) {
	who, err := pendingUser(r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if err := s.tools.confirm.Dismiss(r.Context(), who, chi.URLParam(r, "id")); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, PendingResult{Status: mcpkit.OutcomeDismissed})
}

// pendingList renders the person's open actions with their targets as they
// are now.
func (t *toolset) pendingList(ctx context.Context, userID string) (PendingList, error) {
	actions, err := t.confirm.Open(ctx, userID)
	if err != nil {
		return PendingList{}, err
	}
	list := PendingList{Actions: make([]PendingItem, 0, len(actions))}
	for _, action := range actions {
		item := PendingItem{
			ID: action.ID, Kind: action.Kind, ClientID: action.ClientID, Summary: action.Summary,
			CreatedAt: action.CreatedAt, ExpiresAt: action.ExpiresAt, State: t.confirm.State(ctx, action),
		}
		item.Target = t.pendingTarget(ctx, action)
		list.Actions = append(list.Actions, item)
	}
	return list, nil
}

// pendingTarget describes what an action would delete, or nil when it is
// gone.
func (t *toolset) pendingTarget(ctx context.Context, action mcpkit.Action) *PendingTarget {
	userID, id, err := refOf(action)
	if err != nil {
		return nil
	}
	ctx = t.as(ctx, userID, "")
	if action.Kind == kindDeleteProject {
		p, err := t.domain.Projects.Calls().Get(ctx, userID, id)
		if err != nil {
			return nil
		}
		return &PendingTarget{Type: "project", ID: p.ID, Title: p.Name, TaskCount: taskCount(*p), URL: t.projectURL(p.WorkspaceID, p.ID)}
	}
	tk, err := t.domain.Tasks.Calls().Get(ctx, userID, id)
	if err != nil {
		return nil
	}
	p := t.projectsOf(ctx, []task.TaskResponse{*tk})[tk.ProjectID]
	return &PendingTarget{Type: "task", ID: tk.ID, Title: tk.Title, Ref: ref(p.Slug, tk.Number), ProjectName: p.Name, URL: t.taskURL(*tk)}
}
