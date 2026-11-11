package mcp

import (
	"context"
	"fmt"

	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/Opus/apps/api/modules/project"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// errNoPersonal is the refusal for a create without a workspace when the
// user has no personal workspace.
var errNoPersonal = fmt.Errorf("%s", "the user has no personal workspace. Pass workspace_id; "+hintWorkspace)

// createProject is opus_create_project.
func (t *toolset) createProject(ctx context.Context, req *sdk.CallToolRequest, in createProjectInput) (*sdk.CallToolResult, projectOutput, error) {
	ctx, userID, err := t.caller(ctx, req)
	if err != nil {
		return nil, projectOutput{}, err
	}
	workspaceID, err := t.targetWorkspace(ctx, userID, in.WorkspaceID)
	if err != nil {
		return nil, projectOutput{}, err
	}
	created, err := t.domain.Projects.Calls().Create(ctx, userID, &project.CreateProjectRequest{
		Name: in.Name, WorkspaceID: wireid.ID(workspaceID), Slug: in.Key,
	})
	if err != nil {
		return nil, projectOutput{}, mcpkit.Failure(err, hintWorkspace)
	}
	if in.Description != "" {
		description := in.Description
		if _, err := t.domain.Projects.Calls().Update(ctx, userID, created.ID, &project.UpdateProjectRequest{Description: &description}); err != nil {
			return nil, projectOutput{}, mcpkit.Failure(err, hintProject)
		}
	}
	return t.projectWritten(ctx, userID, created.ID, "created", "The project exists with the default board: to-do, in-progress, in-review and done.")
}

// targetWorkspace is the workspace a create writes to: the one asked for,
// or the user's personal workspace.
func (t *toolset) targetWorkspace(ctx context.Context, userID, asked int64) (int64, error) {
	if asked != 0 {
		return asked, nil
	}
	list, err := t.domain.Workspaces.ListForUser(ctx, userID)
	if err != nil {
		return 0, mcpkit.Failure(err, "")
	}
	for _, ws := range list {
		if ws.Personal {
			return ws.ID, nil
		}
	}
	return 0, errNoPersonal
}

// updateProject is opus_update_project.
func (t *toolset) updateProject(ctx context.Context, req *sdk.CallToolRequest, in updateProjectInput) (*sdk.CallToolResult, projectOutput, error) {
	ctx, userID, err := t.caller(ctx, req)
	if err != nil {
		return nil, projectOutput{}, err
	}
	calls := t.domain.Projects.Calls()
	if in.Name != nil || in.Description != nil {
		if _, err := calls.Update(ctx, userID, in.ProjectID, &project.UpdateProjectRequest{Name: in.Name, Description: in.Description}); err != nil {
			return nil, projectOutput{}, mcpkit.Failure(err, hintProject)
		}
	}
	if in.Archived != nil {
		archive := calls.Unarchive
		if *in.Archived {
			archive = calls.Archive
		}
		if _, err := archive(ctx, userID, in.ProjectID); err != nil {
			return nil, projectOutput{}, mcpkit.Failure(err, hintProject)
		}
	}
	return t.projectWritten(ctx, userID, in.ProjectID, "updated", "The project is updated.")
}

// projectWritten reports a project write with the project as it is now.
func (t *toolset) projectWritten(ctx context.Context, userID, projectID int64, status, message string) (*sdk.CallToolResult, projectOutput, error) {
	p, err := t.domain.Projects.Calls().Get(ctx, userID, projectID)
	if err != nil {
		return nil, projectOutput{}, mcpkit.Failure(err, hintProject)
	}
	summary, err := t.projectSummary(ctx, userID, *p, true)
	if err != nil {
		return nil, projectOutput{}, err
	}
	return nil, projectOutput{Status: status, Message: message, Project: summary}, nil
}
