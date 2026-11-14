package mcp

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/modules/task"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Limits of opus_search_tasks.
const (
	defaultTaskLimit = 30
	maxTaskLimit     = 100
	searchTruncated  = "Descriptions are shortened. Read one task in full with opus_read_task."
)

// Refusals of the task reads.
var (
	errAssignee = errors.New(`assignee is "me", "none" or a user_id from opus_list_workspaces`)
	errPriority = errors.New("priority is no-priority, low, medium, high or urgent")
	errSort     = errors.New("sort is updated, due, priority, created or number")
	errNoTask   = errors.New("pass task_id, or ref such as opus-42")
)

// searchTasks is opus_search_tasks.
func (t *toolset) searchTasks(ctx context.Context, req *sdk.CallToolRequest, in searchInput) (*sdk.CallToolResult, searchOutput, error) {
	ctx, userID, err := t.caller(ctx, req)
	if err != nil {
		return nil, searchOutput{}, err
	}
	full, err := mcpkit.Detailed(in.ResponseFormat)
	if err != nil {
		return nil, searchOutput{}, err
	}
	offset, err := mcpkit.Offset(in.Cursor)
	if err != nil {
		return nil, searchOutput{}, err
	}
	q, err := searchQuery(in, userID)
	if err != nil {
		return nil, searchOutput{}, err
	}
	q.Offset, q.Limit = offset, clamp(in.Limit, defaultTaskLimit, maxTaskLimit)
	records, total, err := t.domain.Tasks.Search(ctx, userID, q)
	if err != nil {
		return nil, searchOutput{}, mcpkit.Failure(err, hintProject)
	}
	return nil, t.taskPage(ctx, records, total, offset, full), nil
}

// searchQuery turns the tool input into a task search.
func searchQuery(in searchInput, userID int64) (task.SearchQuery, error) {
	q := task.SearchQuery{
		WorkspaceID: in.WorkspaceID, ProjectID: in.ProjectID, Text: in.Query, Statuses: in.Statuses,
		Label: strings.TrimSpace(in.Label), IncludeDone: in.IncludeDone, IncludeArchived: in.IncludeArchived,
		Priority: in.Priority, Sort: in.Sort,
	}
	if q.Priority != "" && !validPriority(q.Priority) {
		return q, errPriority
	}
	if !validSort(q.Sort) {
		return q, errSort
	}
	switch in.Assignee {
	case "":
	case "none":
		q.Unassigned = true
	default:
		id, err := assigneeID(in.Assignee, userID)
		if err != nil {
			return q, err
		}
		q.AssigneeID = &id
	}
	var err error
	if q.DueBefore, err = parseDate(in.DueBefore); err != nil {
		return q, err
	}
	q.DueAfter, err = parseDate(in.DueAfter)
	return q, err
}

// assigneeID reads "me" or a user id.
func assigneeID(value string, userID int64) (int64, error) {
	if value == "me" {
		return userID, nil
	}
	id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || id <= 0 {
		return 0, errAssignee
	}
	return id, nil
}

// validPriority reports whether p is one of Opus's priorities.
func validPriority(p string) bool {
	switch p {
	case "no-priority", "low", "medium", "high", "urgent":
		return true
	}
	return false
}

// validSort reports whether s is a sort opus_search_tasks knows.
func validSort(s string) bool {
	switch s {
	case "", task.SortUpdated, task.SortDue, task.SortPriority, task.SortCreated, task.SortNumber:
		return true
	}
	return false
}

// taskPage renders one page of search results, cut to the size budget.
func (t *toolset) taskPage(ctx context.Context, records []task.TaskRecord, total int64, offset int, full bool) searchOutput {
	responses := make([]task.TaskResponse, len(records))
	for i, r := range records {
		responses[i] = task.Response(r)
	}
	projects := t.projectsOf(ctx, responses)
	summaries := make([]taskSummary, len(responses))
	for i, tk := range responses {
		summaries[i] = t.summarize(tk, projects[tk.ProjectID], full)
	}
	keep := mcpkit.FitCount(len(summaries), func(n int) int {
		return mcpkit.EncodedSize(searchOutput{Tasks: summaries[:n], Total: total, NextCursor: "next-cursor-placeholder"})
	})
	out := searchOutput{Tasks: summaries[:keep], Total: total, NextCursor: mcpkit.Cursor(offset+keep, int(total))}
	if full {
		out.Note = searchTruncated
	}
	return out
}

// readTask is opus_read_task.
func (t *toolset) readTask(ctx context.Context, req *sdk.CallToolRequest, in readInput) (*sdk.CallToolResult, taskOutput, error) {
	ctx, userID, err := t.caller(ctx, req)
	if err != nil {
		return nil, taskOutput{}, err
	}
	offset, err := mcpkit.Offset(in.CommentsCursor)
	if err != nil {
		return nil, taskOutput{}, err
	}
	taskID := in.TaskID
	if taskID == 0 {
		if taskID, err = t.resolveRef(ctx, userID, in.Ref); err != nil {
			return nil, taskOutput{}, err
		}
	}
	tk, err := t.domain.Tasks.Calls().Get(ctx, userID, taskID)
	if err != nil {
		return nil, taskOutput{}, mcpkit.Failure(err, hintTask)
	}
	out := t.fullOf(ctx, *tk)
	if err := t.attachComments(ctx, userID, &out, offset); err != nil {
		return nil, taskOutput{}, err
	}
	return nil, out, nil
}

// resolveRef finds the task a reference such as opus-42 names among the
// user's workspaces.
func (t *toolset) resolveRef(ctx context.Context, userID int64, value string) (int64, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	cut := strings.LastIndex(value, "-")
	if cut <= 0 {
		return 0, errNoTask
	}
	number, err := strconv.Atoi(value[cut+1:])
	if err != nil || number <= 0 {
		return 0, errNoTask
	}
	var ids []int64
	err = t.domain.DB.WithContext(ctx).Table("tasks").Select("tasks.id").
		Joins("JOIN projects ON projects.id = tasks.project_id").
		Where("LOWER(projects.slug) = ? AND tasks.number = ?", value[:cut], number).
		Where("tasks.workspace_id IN (SELECT workspace_id FROM workspace_members WHERE user_id = ?)", userID).
		Order("tasks.id").Limit(5).Pluck("tasks.id", &ids).Error
	switch {
	case err != nil:
		return 0, mcpkit.ErrUnavailable
	case len(ids) == 0:
		return 0, errors.New("no task has the ref " + value + ". " + hintTask)
	case len(ids) > 1:
		return 0, errors.New("several workspaces have a task " + value + ". Pass task_id instead; " + hintTask)
	}
	return ids[0], nil
}
