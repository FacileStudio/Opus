package mcp

import (
	"context"
	"net/http"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/modules/task"
	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// webActor marks, in the audit, a confirmation made on the confirm page.
const webActor = "opus-web"

// pendingUser is the signed-in person of a confirm page request. A delegated
// token is refused: an agent must never confirm its own request.
func pendingUser(r *http.Request) (string, error) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		return "", errors.Unauthorized("missing auth token")
	}
	if identity.Actor != "" {
		return "", mcpkit.ErrDelegated
	}
	return strconv.FormatInt(identity.UserID, 10), nil
}

// describe adds the task or project an action would delete, as it is now,
// to the confirm page entry.
func (t *toolset) describe(ctx context.Context, item mcpkit.PendingItem, a mcpkit.Action) PendingItem {
	return PendingItem{PendingItem: item, Target: t.pendingTarget(ctx, a)}
}

// done is the body answered once a confirmed deletion ran.
func done(context.Context, mcpkit.Outcome) mcpkit.PendingResult {
	return mcpkit.PendingResult{Status: "deleted"}
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
