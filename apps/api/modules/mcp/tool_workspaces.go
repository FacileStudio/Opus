package mcp

import (
	"context"
	"slices"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/modules/label"
	"github.com/FacileStudio/Opus/apps/api/modules/workspace"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// maxMembers is how many members one workspace lists.
const maxMembers = 50

// noteMembers explains a cut member list.
const noteMembers = "Some workspaces list only their first 50 members, by name. A task's assignee shows in opus_search_tasks."

// listWorkspaces is opus_list_workspaces.
func (t *toolset) listWorkspaces(ctx context.Context, req *sdk.CallToolRequest, _ workspacesInput) (*sdk.CallToolResult, workspacesOutput, error) {
	ctx, userID, err := t.caller(ctx, req)
	if err != nil {
		return nil, workspacesOutput{}, err
	}
	list, err := t.domain.Workspaces.ListForUser(ctx, userID)
	if err != nil {
		return nil, workspacesOutput{}, mcpkit.Failure(err, "")
	}
	out := workspacesOutput{Workspaces: make([]workspaceSummary, 0, len(list))}
	for _, ws := range list {
		summary, cut, err := t.workspaceSummary(ctx, userID, ws)
		if err != nil {
			return nil, workspacesOutput{}, err
		}
		if cut {
			out.Note = noteMembers
		}
		out.Workspaces = append(out.Workspaces, summary)
	}
	return nil, out, nil
}

// workspaceSummary renders one workspace with its members and labels, and
// reports whether the member list was cut.
func (t *toolset) workspaceSummary(ctx context.Context, userID int64, ws workspace.WorkspaceResponse) (workspaceSummary, bool, error) {
	members, err := t.domain.Workspaces.Members(ctx, ws.ID)
	if err != nil {
		return workspaceSummary{}, false, mcpkit.Failure(err, "")
	}
	labels, err := t.domain.Labels.Calls().ListByWorkspace(ctx, userID, ws.ID)
	if err != nil {
		return workspaceSummary{}, false, mcpkit.Failure(err, hintWorkspace)
	}
	s := workspaceSummary{
		WorkspaceID: ws.ID, Name: ws.Name, Personal: ws.Personal, MemberCount: len(members),
		Members: make([]person, 0, min(len(members), maxMembers)), Labels: distinctLabels(labels),
		ManagedByIdP: ws.RegistreSpaceID != nil,
	}
	for _, m := range members {
		if m.ID == userID {
			s.Role = m.Role
		}
		if len(s.Members) < maxMembers {
			s.Members = append(s.Members, person{UserID: m.ID, Name: m.Name, Email: m.Email, Role: m.Role})
		}
	}
	return s, len(members) > maxMembers, nil
}

// distinctLabels lists a workspace's label names once each, sorted, with
// the color of the first label found under that name.
func distinctLabels(labels []label.LabelResponse) []labelSummary {
	seen := map[string]bool{}
	out := []labelSummary{}
	for _, l := range labels {
		key := strings.ToLower(l.Name)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, labelSummary{Name: l.Name, Color: l.Color})
	}
	slices.SortFunc(out, func(a, b labelSummary) int { return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)) })
	return out
}
