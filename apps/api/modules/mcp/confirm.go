package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/modules/project"
	"github.com/FacileStudio/Opus/apps/api/modules/task"
	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// The two sensitive actions.
const (
	kindDeleteTask    = "delete_task"
	kindDeleteProject = "delete_project"
)

// errTarget is the refusal for a pending action whose target is not an id.
var errTarget = errors.NotFound("this action names no task or project")

// kinds are the sensitive actions. Each fingerprints its target, so an edit
// after the question closes the action instead of deleting what nobody saw.
func (t *toolset) kinds() map[string]mcpkit.Kind {
	return map[string]mcpkit.Kind{
		kindDeleteTask:    {Tool: toolDeleteTask, Fingerprint: t.taskPrint, Run: t.removeTask},
		kindDeleteProject: {Tool: toolDeleteProject, Fingerprint: t.projectPrint, Run: t.removeProject},
	}
}

// refOf reads the user and the id an action targets.
func refOf(a mcpkit.Action) (int64, int64, error) {
	userID, errUser := strconv.ParseInt(a.UserID, 10, 64)
	id, errID := strconv.ParseInt(a.Target, 10, 64)
	if errUser != nil || errID != nil {
		return 0, 0, errTarget
	}
	return userID, id, nil
}

// taskPrint is the digest of the targeted task as it is now.
func (t *toolset) taskPrint(ctx context.Context, a mcpkit.Action) (string, error) {
	userID, taskID, err := refOf(a)
	if err != nil {
		return "", err
	}
	ctx = t.as(ctx, userID, "")
	tk, err := t.domain.Tasks.Calls().Get(ctx, userID, taskID)
	if err != nil {
		return "", err
	}
	return digest(tk.UpdatedAt.UTC().Format(time.RFC3339Nano), tk.Title, tk.Status, strconv.FormatInt(tk.ProjectID, 10)), nil
}

// removeTask deletes the targeted task.
func (t *toolset) removeTask(ctx context.Context, a mcpkit.Action) error {
	userID, taskID, err := refOf(a)
	if err != nil {
		return err
	}
	ctx = t.as(ctx, userID, "")
	_, err = t.domain.Tasks.Calls().Delete(ctx, userID, taskID)
	return err
}

// projectPrint is the digest of the targeted project as it is now.
func (t *toolset) projectPrint(ctx context.Context, a mcpkit.Action) (string, error) {
	userID, projectID, err := refOf(a)
	if err != nil {
		return "", err
	}
	ctx = t.as(ctx, userID, "")
	p, err := t.domain.Projects.Calls().Get(ctx, userID, projectID)
	if err != nil {
		return "", err
	}
	return digest(p.Name, p.Slug, p.Description, strconv.Itoa(taskCount(*p))), nil
}

// removeProject deletes the targeted project with its tasks.
func (t *toolset) removeProject(ctx context.Context, a mcpkit.Action) error {
	userID, projectID, err := refOf(a)
	if err != nil {
		return err
	}
	ctx = t.as(ctx, userID, "")
	_, err = t.domain.Projects.Calls().Delete(ctx, userID, projectID)
	return err
}

// digest fingerprints the fields the person was shown.
func digest(fields ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(fields, "\x00")))
	return hex.EncodeToString(sum[:16])
}

// taskCount is how many tasks a project holds.
func taskCount(p project.ProjectResponse) int {
	if p.Statistics == nil {
		return 0
	}
	return p.Statistics.TotalTasks
}

// taskQuestion is what the client asks before deleting a task.
func taskQuestion(tk task.TaskResponse, key string) string {
	return fmt.Sprintf("Delete the task %s %q, with its comments and labels, for good?", ref(key, tk.Number), tk.Title)
}

// projectQuestion is what the client asks before deleting a project.
func projectQuestion(p project.ProjectResponse) string {
	switch n := taskCount(p); n {
	case 0:
		return fmt.Sprintf("Delete the project %q for good?", p.Name)
	case 1:
		return fmt.Sprintf("Delete the project %q and its 1 task for good?", p.Name)
	default:
		return fmt.Sprintf("Delete the project %q and its %d tasks for good?", p.Name, n)
	}
}
