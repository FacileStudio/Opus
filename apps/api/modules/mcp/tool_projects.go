package mcp

import (
	"context"
	"slices"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/modules/project"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Limits of opus_list_projects.
const (
	defaultProjectLimit = 50
	maxProjectLimit     = 100
)

// listProjects is opus_list_projects.
func (t *toolset) listProjects(ctx context.Context, req *sdk.CallToolRequest, in projectsInput) (*sdk.CallToolResult, projectsOutput, error) {
	ctx, userID, err := t.caller(ctx, req)
	if err != nil {
		return nil, projectsOutput{}, err
	}
	full, err := mcpkit.Detailed(in.ResponseFormat)
	if err != nil {
		return nil, projectsOutput{}, err
	}
	offset, err := mcpkit.Offset(in.Cursor)
	if err != nil {
		return nil, projectsOutput{}, err
	}
	found, err := t.projectsIn(ctx, userID, in.WorkspaceID, in.IncludeArchived)
	if err != nil {
		return nil, projectsOutput{}, err
	}
	found = matchingProjects(found, in.Query)
	if offset > len(found) {
		return nil, projectsOutput{}, mcpkit.ErrCursor
	}
	rest := found[offset:min(offset+clamp(in.Limit, defaultProjectLimit, maxProjectLimit), len(found))]
	summaries := make([]projectSummary, 0, len(rest))
	for _, p := range rest {
		s, err := t.projectSummary(ctx, userID, p, full)
		if err != nil {
			return nil, projectsOutput{}, err
		}
		summaries = append(summaries, s)
	}
	keep := mcpkit.FitCount(len(summaries), func(n int) int {
		return mcpkit.EncodedSize(projectsOutput{Projects: summaries[:n], Total: len(found), NextCursor: "next-cursor-placeholder"})
	})
	return nil, projectsOutput{Projects: summaries[:keep], Total: len(found), NextCursor: mcpkit.Cursor(offset+keep, len(found))}, nil
}

// projectsIn lists the projects of one workspace, or of every workspace of
// the user, through the project route.
func (t *toolset) projectsIn(ctx context.Context, userID, workspaceID int64, archived bool) ([]project.ProjectResponse, error) {
	ids := []int64{workspaceID}
	if workspaceID == 0 {
		list, err := t.domain.Workspaces.ListForUser(ctx, userID)
		if err != nil {
			return nil, mcpkit.Failure(err, "")
		}
		ids = ids[:0]
		for _, ws := range list {
			ids = append(ids, ws.ID)
		}
	}
	var all []project.ProjectResponse
	for _, id := range ids {
		list, err := t.domain.Projects.Calls().List(ctx, userID, id, archived)
		if err != nil {
			return nil, mcpkit.Failure(err, hintWorkspace)
		}
		all = append(all, list...)
	}
	return all, nil
}

// matchingProjects keeps the projects whose name or description hold every
// word of query, ignoring case.
func matchingProjects(list []project.ProjectResponse, query string) []project.ProjectResponse {
	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return list
	}
	return slices.DeleteFunc(list, func(p project.ProjectResponse) bool {
		text := strings.ToLower(p.Name + "\n" + p.Slug + "\n" + p.Description)
		return slices.ContainsFunc(words, func(word string) bool { return !strings.Contains(text, word) })
	})
}

// projectSummary renders one project with its board's statuses.
func (t *toolset) projectSummary(ctx context.Context, userID int64, p project.ProjectResponse, full bool) (projectSummary, error) {
	cols, err := t.domain.Columns.List(ctx, userID, p.ID)
	if err != nil {
		return projectSummary{}, mcpkit.Failure(err, hintProject)
	}
	s := projectSummary{
		ProjectID: p.ID, WorkspaceID: p.WorkspaceID, Name: p.Name, Key: p.Slug, Archived: p.ArchivedAt != nil,
		Statuses: make([]string, 0, len(cols)+2),
	}
	if p.Statistics != nil {
		s.TotalTasks, s.PercentDone = p.Statistics.TotalTasks, p.Statistics.CompletionPercentage
	}
	for _, c := range cols {
		s.Statuses = append(s.Statuses, c.Slug)
	}
	s.Statuses = append(s.Statuses, "planned", "archived")
	if full {
		s.projectDetail = projectDetail{
			Description: p.Description, Icon: p.Icon, CreatedAt: stamp(p.CreatedAt), URL: t.projectURL(p.WorkspaceID, p.ID),
		}
		for _, c := range cols {
			s.Columns = append(s.Columns, boardColumn{Status: c.Slug, Name: c.Name, Final: c.IsFinal})
		}
	}
	return s, nil
}

// clamp is limit within [1, most], or fallback when unset.
func clamp(limit, fallback, most int) int {
	if limit <= 0 {
		return fallback
	}
	return min(limit, most)
}
