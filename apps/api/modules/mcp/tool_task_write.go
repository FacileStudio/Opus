package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/Opus/apps/api/modules/label"
	"github.com/FacileStudio/Opus/apps/api/modules/task"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// errNotMember is the refusal for an assignee outside the task's workspace.
var errNotMember = errors.New("the assignee is not a member of this workspace. Call opus_list_workspaces for its members")

// createTask is opus_create_task.
func (t *toolset) createTask(ctx context.Context, req *sdk.CallToolRequest, in createTaskInput) (*sdk.CallToolResult, writeOutput, error) {
	ctx, userID, err := t.caller(ctx, req)
	if err != nil {
		return nil, writeOutput{}, err
	}
	p, err := t.domain.Projects.Calls().Get(ctx, userID, in.ProjectID)
	if err != nil {
		return nil, writeOutput{}, mcpkit.Failure(err, hintProject)
	}
	body, err := t.createBody(ctx, userID, p.WorkspaceID, in)
	if err != nil {
		return nil, writeOutput{}, err
	}
	labels, err := t.resolveLabels(ctx, userID, p.WorkspaceID, in.Labels)
	if err != nil {
		return nil, writeOutput{}, err
	}
	created, err := t.domain.Tasks.Calls().Create(ctx, userID, in.ProjectID, body)
	if err != nil {
		return nil, writeOutput{}, mcpkit.Failure(err, hintProject)
	}
	if err := t.setLabels(ctx, userID, *created, labels); err != nil {
		return nil, writeOutput{}, err
	}
	return t.written(ctx, userID, created.ID, "created", "The task is on the board.")
}

// createBody is the create route's request for in.
func (t *toolset) createBody(ctx context.Context, userID, workspaceID int64, in createTaskInput) (*task.CreateTaskRequest, error) {
	body := &task.CreateTaskRequest{Title: in.Title, Description: in.Description, Status: in.Status, Priority: in.Priority}
	var err error
	if body.DueDate, err = parseDate(in.Due); err != nil {
		return nil, err
	}
	if body.StartDate, err = parseDate(in.Start); err != nil {
		return nil, err
	}
	if in.Assignee != "" {
		id, err := t.member(ctx, userID, workspaceID, in.Assignee)
		if err != nil {
			return nil, err
		}
		body.ActorID = wireid.FromInt64Ptr(&id)
	}
	return body, nil
}

// member reads an assignee and checks they belong to the workspace.
func (t *toolset) member(ctx context.Context, userID, workspaceID int64, value string) (int64, error) {
	id, err := assigneeID(value, userID)
	if err != nil {
		return 0, err
	}
	var count int64
	if err := t.domain.DB.WithContext(ctx).Model(&schemas.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ?", workspaceID, id).Count(&count).Error; err != nil {
		return 0, mcpkit.ErrUnavailable
	}
	if count == 0 {
		return 0, errNotMember
	}
	return id, nil
}

// updateTask is opus_update_task.
func (t *toolset) updateTask(ctx context.Context, req *sdk.CallToolRequest, in updateTaskInput) (*sdk.CallToolResult, writeOutput, error) {
	ctx, userID, err := t.caller(ctx, req)
	if err != nil {
		return nil, writeOutput{}, err
	}
	current, err := t.domain.Tasks.Calls().Get(ctx, userID, in.TaskID)
	if err != nil {
		return nil, writeOutput{}, mcpkit.Failure(err, hintTask)
	}
	var labels []label.LabelResponse
	if in.Labels != nil {
		if labels, err = t.resolveLabels(ctx, userID, current.WorkspaceID, *in.Labels); err != nil {
			return nil, writeOutput{}, err
		}
	}
	if current, err = t.moveIfAsked(ctx, userID, *current, in); err != nil {
		return nil, writeOutput{}, err
	}
	body, changed, err := t.merged(ctx, userID, *current, in)
	if err != nil {
		return nil, writeOutput{}, err
	}
	if changed {
		if _, err := t.domain.Tasks.Calls().Update(ctx, userID, in.TaskID, body); err != nil {
			return nil, writeOutput{}, mcpkit.Failure(err, hintTask)
		}
	}
	if in.Labels != nil {
		if err := t.setLabels(ctx, userID, *current, labels); err != nil {
			return nil, writeOutput{}, err
		}
	}
	return t.written(ctx, userID, in.TaskID, "updated", "The task is updated.")
}

// moveIfAsked moves the task to in.ProjectID when that is another project,
// keeping the status asked for, and returns the task as it is after.
func (t *toolset) moveIfAsked(ctx context.Context, userID int64, current task.TaskResponse, in updateTaskInput) (*task.TaskResponse, error) {
	if in.ProjectID == nil || *in.ProjectID == current.ProjectID {
		return &current, nil
	}
	move := &task.MoveTaskRequest{DestinationProjectID: wireid.ID(*in.ProjectID)}
	if in.Status != nil {
		move.DestinationStatus = *in.Status
	}
	moved, err := t.domain.Tasks.Calls().Move(ctx, userID, current.ID, move)
	if err != nil {
		return nil, mcpkit.Failure(err, hintProject)
	}
	return &moved.Task, nil
}

// merged is the full update route's request: the task as it is, with the
// fields in sets. changed reports whether any field differs.
func (t *toolset) merged(ctx context.Context, userID int64, tk task.TaskResponse, in updateTaskInput) (*task.UpdateTaskRequest, bool, error) {
	body := &task.UpdateTaskRequest{
		Title: or(in.Title, tk.Title), Description: or(in.Description, tk.Description), Status: or(in.Status, tk.Status),
		Priority: or(in.Priority, tk.Priority), ProjectID: wireid.ID(tk.ProjectID), Position: tk.Position,
		StartDate: tk.StartDate, DueDate: tk.DueDate, ActorID: wireid.FromInt64Ptr(tk.ActorID),
	}
	var err error
	if in.Due != nil {
		if body.DueDate, err = parseDate(*in.Due); err != nil {
			return nil, false, err
		}
	}
	if in.Start != nil {
		if body.StartDate, err = parseDate(*in.Start); err != nil {
			return nil, false, err
		}
	}
	if in.Assignee != nil {
		if body.ActorID, err = t.actor(ctx, userID, tk.WorkspaceID, *in.Assignee); err != nil {
			return nil, false, err
		}
	}
	return body, differs(tk, body), nil
}

// actor reads an update's assignee: none clears it.
func (t *toolset) actor(ctx context.Context, userID, workspaceID int64, value string) (*wireid.ID, error) {
	if value == "none" || value == "" {
		return nil, nil
	}
	id, err := t.member(ctx, userID, workspaceID, value)
	if err != nil {
		return nil, err
	}
	return wireid.FromInt64Ptr(&id), nil
}

// differs reports whether body changes anything on tk.
func differs(tk task.TaskResponse, body *task.UpdateTaskRequest) bool {
	return body.Title != tk.Title || body.Description != tk.Description || body.Status != tk.Status ||
		body.Priority != tk.Priority || !sameTime(body.DueDate, tk.DueDate) || !sameTime(body.StartDate, tk.StartDate) ||
		!sameID(wireid.Int64Ptr(body.ActorID), tk.ActorID)
}

// sameTime compares two optional times.
func sameTime(a, b *time.Time) bool {
	return (a == nil) == (b == nil) && (a == nil || a.Equal(*b))
}

// sameID compares two optional ids.
func sameID(a, b *int64) bool {
	return (a == nil) == (b == nil) && (a == nil || *a == *b)
}

// or is *value when set, else fallback.
func or[T any](value *T, fallback T) T {
	if value != nil {
		return *value
	}
	return fallback
}

// resolveLabels finds the workspace labels names refers to, refusing a name
// the workspace does not have.
func (t *toolset) resolveLabels(ctx context.Context, userID, workspaceID int64, names []string) ([]label.LabelResponse, error) {
	if len(names) == 0 {
		return nil, nil
	}
	all, err := t.domain.Labels.Calls().ListByWorkspace(ctx, userID, workspaceID)
	if err != nil {
		return nil, mcpkit.Failure(err, hintWorkspace)
	}
	known := map[string]label.LabelResponse{}
	for _, l := range all {
		if _, seen := known[strings.ToLower(l.Name)]; !seen {
			known[strings.ToLower(l.Name)] = l
		}
	}
	out := make([]label.LabelResponse, 0, len(names))
	for _, name := range names {
		l, ok := known[strings.ToLower(strings.TrimSpace(name))]
		if !ok {
			return nil, fmt.Errorf("the workspace has no label %q. Call opus_list_workspaces for its labels", name)
		}
		out = append(out, l)
	}
	return out, nil
}

// setLabels makes the task carry exactly wanted, through the label routes.
func (t *toolset) setLabels(ctx context.Context, userID int64, tk task.TaskResponse, wanted []label.LabelResponse) error {
	keep := map[string]bool{}
	for _, l := range wanted {
		keep[strings.ToLower(l.Name)] = true
	}
	held := map[string]bool{}
	for _, l := range tk.Labels {
		held[strings.ToLower(l.Name)] = true
		if keep[strings.ToLower(l.Name)] {
			continue
		}
		if _, err := t.domain.Labels.Calls().Delete(ctx, userID, l.ID); err != nil {
			return mcpkit.Failure(err, hintTask)
		}
	}
	for _, l := range wanted {
		if held[strings.ToLower(l.Name)] {
			continue
		}
		held[strings.ToLower(l.Name)] = true
		taskID := wireid.ID(tk.ID)
		if _, err := t.domain.Labels.Calls().Create(ctx, userID, &label.CreateLabelRequest{
			Name: l.Name, Color: l.Color, WorkspaceID: wireid.ID(tk.WorkspaceID), TaskID: &taskID,
		}); err != nil {
			return mcpkit.Failure(err, hintTask)
		}
	}
	return nil
}

// written reports a task write with the task as it is now.
func (t *toolset) written(ctx context.Context, userID, taskID int64, status, message string) (*sdk.CallToolResult, writeOutput, error) {
	tk, err := t.domain.Tasks.Calls().Get(ctx, userID, taskID)
	if err != nil {
		return nil, writeOutput{}, mcpkit.Failure(err, hintTask)
	}
	return nil, writeOutput{Status: status, Message: message, Task: t.fullOf(ctx, *tk)}, nil
}
