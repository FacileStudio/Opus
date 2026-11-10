package mcp

import (
	"context"
	"fmt"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/modules/task"

	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// waiting tells the model that nothing happened yet, and what to do.
const waiting = "Nothing was deleted: your client cannot ask the user to confirm. Give the user confirm_url " +
	"and ask them to confirm there. Do not call this tool again for the same %s; the request stays open for 7 days."

// declined tells the model the person said no.
const declined = "The user declined. Nothing happened, and the %s is kept."

// deleteTask is opus_delete_task.
func (t *toolset) deleteTask(ctx context.Context, req *sdk.CallToolRequest, in taskRef) (*sdk.CallToolResult, actionOutput, error) {
	if mcpkit.Retrying(req) {
		return t.resume(ctx, req, kindDeleteTask, "task")
	}
	ctx, userID, err := t.caller(ctx, req)
	if err != nil {
		return nil, actionOutput{}, err
	}
	tk, err := t.domain.Tasks.Calls().Get(ctx, userID, in.TaskID)
	if err != nil {
		return nil, actionOutput{}, mcpkit.Failure(err, hintTask)
	}
	key := t.projectsOf(ctx, []task.TaskResponse{*tk})[tk.ProjectID].Slug
	result, outcome, err := t.confirm.Request(ctx, req, kindDeleteTask, strconv.FormatInt(in.TaskID, 10), taskQuestion(*tk, key))
	if err != nil {
		return nil, actionOutput{}, mcpkit.Failure(err, hintTask)
	}
	if result != nil {
		return result, actionOutput{}, nil
	}
	return nil, actionOutput{Status: outcome.Status, TaskID: in.TaskID, ConfirmURL: outcome.ConfirmURL, Message: fmt.Sprintf(waiting, "task")}, nil
}

// deleteProject is opus_delete_project.
func (t *toolset) deleteProject(ctx context.Context, req *sdk.CallToolRequest, in projectRef) (*sdk.CallToolResult, actionOutput, error) {
	if mcpkit.Retrying(req) {
		return t.resume(ctx, req, kindDeleteProject, "project")
	}
	ctx, userID, err := t.caller(ctx, req)
	if err != nil {
		return nil, actionOutput{}, err
	}
	p, err := t.domain.Projects.Calls().Get(ctx, userID, in.ProjectID)
	if err != nil {
		return nil, actionOutput{}, mcpkit.Failure(err, hintProject)
	}
	result, outcome, err := t.confirm.Request(ctx, req, kindDeleteProject, strconv.FormatInt(in.ProjectID, 10), projectQuestion(*p))
	if err != nil {
		return nil, actionOutput{}, mcpkit.Failure(err, hintProject)
	}
	if result != nil {
		return result, actionOutput{}, nil
	}
	return nil, actionOutput{Status: outcome.Status, ProjectID: in.ProjectID, ConfirmURL: outcome.ConfirmURL, Message: fmt.Sprintf(waiting, "project")}, nil
}

// resume handles the retry that carries the person's answer.
func (t *toolset) resume(ctx context.Context, req *sdk.CallToolRequest, kind, noun string) (*sdk.CallToolResult, actionOutput, error) {
	outcome, err := t.confirm.Resume(ctx, req, kind)
	if err != nil {
		return nil, actionOutput{}, mcpkit.Failure(err, "Call the tool again to ask the user afresh.")
	}
	if outcome.Status == mcpkit.StatusDeclined {
		return nil, actionOutput{Status: mcpkit.StatusDeclined, Message: fmt.Sprintf(declined, noun)}, nil
	}
	id, _ := strconv.ParseInt(outcome.Action.Target, 10, 64)
	out := actionOutput{Status: "deleted", Message: "The user confirmed and the " + noun + " was deleted."}
	if kind == kindDeleteTask {
		out.TaskID = id
	} else {
		out.ProjectID = id
	}
	return nil, out, nil
}
