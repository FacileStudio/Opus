package mcp

import (
	"context"
	"errors"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte/mcpauth"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool names. Every name carries the app prefix, so a client holding many
// servers can tell them apart.
const (
	toolListWorkspaces = "opus_list_workspaces"
	toolListProjects   = "opus_list_projects"
	toolSearchTasks    = "opus_search_tasks"
	toolReadTask       = "opus_read_task"
	toolCreateTask     = "opus_create_task"
	toolUpdateTask     = "opus_update_task"
	toolCommentTask    = "opus_comment_task"
	toolCreateProject  = "opus_create_project"
	toolUpdateProject  = "opus_update_project"
	toolDeleteTask     = "opus_delete_task"
	toolDeleteProject  = "opus_delete_project"
)

// Hints appended to a not_found error, naming the tool that lists valid ids.
const (
	hintWorkspace = "Call opus_list_workspaces for the workspace ids you can use."
	hintProject   = "Call opus_list_projects for the project ids you can use."
	hintTask      = "Call opus_search_tasks for the task ids you can use."
)

// errNoCaller is the refusal for a token that resolves to no Opus user.
var errNoCaller = errors.New("this request carries no valid access token. Reconnect Opus")

// toolset holds what the tool handlers share.
type toolset struct {
	domain  Domain
	confirm *mcpkit.Confirmations
	appURL  string
}

// caller is the person a tool call acts for, with a context carrying their
// identity the way the HTTP auth middleware sets it, so the module calls
// stamp the same audit and sync fields as a web request.
func (t *toolset) caller(ctx context.Context, req *sdk.CallToolRequest) (context.Context, int64, error) {
	identity, ok := mcpauth.Identity(req.Extra.TokenInfo)
	if !ok || identity.UserID == 0 {
		return ctx, 0, errNoCaller
	}
	return t.as(ctx, identity.UserID, identity.Actor), identity.UserID, nil
}

// as returns ctx carrying userID's identity, with their email when the row
// is there.
func (t *toolset) as(ctx context.Context, userID int64, actor string) context.Context {
	var user schemas.User
	t.domain.DB.WithContext(ctx).Select("id", "email").Where("id = ?", userID).Limit(1).Find(&user)
	return authcontext.WithIdentity(ctx, authcontext.Identity{UserID: userID, Email: user.Email, Actor: actor})
}
