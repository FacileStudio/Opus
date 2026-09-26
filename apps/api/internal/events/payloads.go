package events

import (
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/schemas"
)

// TaskPayload represents the wire payload for task domain events.
type TaskPayload struct {
	ID              int64  `json:"id"`
	FacileID        string `json:"facile_id"`
	ProjectID       int64  `json:"project_id"`
	ProjectFacileID string `json:"project_facile_id"`
	WorkspaceID     int64  `json:"workspace_id"`
	Title           string `json:"title"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	Priority        string `json:"priority,omitempty"`
	ActorID         *int64 `json:"actor_id,omitempty"`
	ActorEmail      string `json:"actor_email,omitempty"`
}

func (t TaskPayload) entityID() string {
	if t.FacileID != "" {
		return t.FacileID
	}
	return strconv.FormatInt(t.ID, 10)
}

// ProjectPayload represents the wire payload for project domain events.
type ProjectPayload struct {
	ID          int64   `json:"id"`
	FacileID    string  `json:"facile_id"`
	WorkspaceID int64   `json:"workspace_id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	ActorEmail  string  `json:"actor_email,omitempty"`
}

func (pr ProjectPayload) entityID() string {
	if pr.FacileID != "" {
		return pr.FacileID
	}
	if pr.Slug != "" {
		return pr.Slug
	}
	return strconv.FormatInt(pr.ID, 10)
}

// CommentPayload represents the wire payload for task comment domain events.
type CommentPayload struct {
	ID         int64  `json:"id"`
	FacileID   string `json:"facile_id"`
	TaskID     int64  `json:"task_id"`
	UserID     int64  `json:"user_id"`
	Content    string `json:"content"`
	ActorEmail string `json:"actor_email,omitempty"`
}

func (c CommentPayload) entityID() string {
	if c.FacileID != "" {
		return c.FacileID
	}
	return strconv.FormatInt(c.ID, 10)
}

// WorkspacePayload represents the wire payload for workspace domain events.
type WorkspacePayload struct {
	ID          int64   `json:"id"`
	FacileID    string  `json:"facile_id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	ActorEmail  string  `json:"actor_email,omitempty"`
}

func (w WorkspacePayload) entityID() string {
	if w.FacileID != "" {
		return w.FacileID
	}
	if w.Slug != "" {
		return w.Slug
	}
	return strconv.FormatInt(w.ID, 10)
}

// NewTaskPayload constructs a TaskPayload from a schema Task and actor email.
func NewTaskPayload(task schemas.Task, actorEmail string) TaskPayload {
	return TaskPayload{
		ID:              task.ID,
		FacileID:        strconv.FormatInt(task.ID, 10),
		ProjectID:       task.ProjectID,
		ProjectFacileID: strconv.FormatInt(task.ProjectID, 10),
		WorkspaceID:     task.WorkspaceID,
		Title:           task.Title,
		Name:            task.Title,
		Status:          task.Status,
		Priority:        task.Priority,
		ActorID:         task.ActorID,
		ActorEmail:      actorEmail,
	}
}

// NewProjectPayload constructs a ProjectPayload from a schema Project and actor email.
func NewProjectPayload(project schemas.Project, actorEmail string) ProjectPayload {
	facileID := project.Slug
	if facileID == "" {
		facileID = strconv.FormatInt(project.ID, 10)
	}
	var desc *string
	if project.Description != "" {
		desc = &project.Description
	}
	var icon *string
	if project.Icon != "" {
		icon = &project.Icon
	}
	return ProjectPayload{
		ID:          project.ID,
		FacileID:    facileID,
		WorkspaceID: project.WorkspaceID,
		Name:        project.Name,
		Slug:        project.Slug,
		Description: desc,
		Icon:        icon,
		ActorEmail:  actorEmail,
	}
}

// NewCommentPayload constructs a CommentPayload from a schema Comment and actor email.
func NewCommentPayload(comment schemas.Comment, actorEmail string) CommentPayload {
	return CommentPayload{
		ID:         comment.ID,
		FacileID:   strconv.FormatInt(comment.ID, 10),
		TaskID:     comment.TaskID,
		UserID:     comment.UserID,
		Content:    comment.Content,
		ActorEmail: actorEmail,
	}
}

// NewWorkspacePayload constructs a WorkspacePayload from a schema Workspace and actor email.
func NewWorkspacePayload(workspace schemas.Workspace, actorEmail string) WorkspacePayload {
	facileID := workspace.Slug
	if facileID == "" {
		facileID = strconv.FormatInt(workspace.ID, 10)
	}
	var desc *string
	if workspace.Description != "" {
		desc = &workspace.Description
	}
	return WorkspacePayload{
		ID:          workspace.ID,
		FacileID:    facileID,
		Name:        workspace.Name,
		Slug:        workspace.Slug,
		Description: desc,
		ActorEmail:  actorEmail,
	}
}
