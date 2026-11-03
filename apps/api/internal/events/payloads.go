package events

import (
	"strconv"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
)

// TaskPayload represents the wire payload for task domain events.
type TaskPayload struct {
	ID              int64   `json:"id"`
	FacileID        string  `json:"facile_id"`
	ProjectID       int64   `json:"project_id"`
	ProjectFacileID string  `json:"project_facile_id"`
	WorkspaceID     int64   `json:"workspace_id"`
	Title           string  `json:"title"`
	Name            string  `json:"name"`
	Status          string  `json:"status"`
	Description     string  `json:"description"`
	Priority        string  `json:"priority"`
	DueDate         *string `json:"due_date"`
	ActorID         *int64  `json:"actor_id,omitempty"`
	ActorEmail      string  `json:"actor_email,omitempty"`
	ActorSubject    string  `json:"actor_subject,omitempty"`
	OwnerEmail      string  `json:"owner_email,omitempty"`
	UpdatedAt       string  `json:"updated_at,omitempty"`
}

func (t TaskPayload) entityID() string {
	if t.FacileID != "" {
		return t.FacileID
	}
	return strconv.FormatInt(t.ID, 10)
}

// SpacePayload tells the receiver where a project lives: a Registre shared
// space, or its owner's personal space.
type SpacePayload struct {
	RegistreSpaceID string `json:"registre_space_id"`
	Personal        bool   `json:"personal"`
	OwnerSubject    string `json:"owner_subject"`
	OwnerEmail      string `json:"owner_email"`
}

// ProjectPayload represents the wire payload for project domain events.
type ProjectPayload struct {
	ID          int64         `json:"id"`
	FacileID    string        `json:"facile_id"`
	WorkspaceID int64         `json:"workspace_id"`
	Name        string        `json:"name"`
	Slug        string        `json:"slug"`
	Description *string       `json:"description,omitempty"`
	Icon        *string       `json:"icon"`
	ActorEmail  string        `json:"actor_email,omitempty"`
	Space       *SpacePayload `json:"space,omitempty"`
	UpdatedAt   string        `json:"updated_at,omitempty"`
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

func facileIDOf(id *string) string {
	if id != nil {
		return *id
	}
	return ""
}

// WirePriority maps an Opus priority onto the contract's none, low, medium,
// high and urgent.
func WirePriority(priority string) string {
	switch priority {
	case "low", "medium", "high", "urgent":
		return priority
	default:
		return "none"
	}
}

// LocalPriority maps a wire priority onto Opus's values, reading normal as
// medium and critical as urgent. It reports false for any other value.
func LocalPriority(wire string) (string, bool) {
	switch wire {
	case "low", "medium", "high", "urgent":
		return wire, true
	case "normal":
		return "medium", true
	case "critical":
		return "urgent", true
	case "none":
		return "no-priority", true
	default:
		return "", false
	}
}

func stamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

// NewTaskPayload constructs a TaskPayload from a schema Task, its project facile_id and the actor email.
func NewTaskPayload(task schemas.Task, projectFacileID, actorEmail string) TaskPayload {
	var dueDate *string
	if task.DueDate != nil {
		day := task.DueDate.UTC().Format(time.DateOnly)
		dueDate = &day
	}
	return TaskPayload{
		ID:              task.ID,
		FacileID:        facileIDOf(task.FacileID),
		ProjectID:       task.ProjectID,
		ProjectFacileID: projectFacileID,
		WorkspaceID:     task.WorkspaceID,
		Title:           task.Title,
		Name:            task.Title,
		Status:          task.Status,
		Description:     task.Description,
		Priority:        WirePriority(task.Priority),
		DueDate:         dueDate,
		ActorID:         task.ActorID,
		ActorEmail:      actorEmail,
		UpdatedAt:       stamp(task.UpdatedAt),
	}
}

// NewProjectPayload constructs a ProjectPayload from a schema Project and actor email.
func NewProjectPayload(project schemas.Project, actorEmail string) ProjectPayload {
	var icon *string
	if project.Icon != "" {
		icon = &project.Icon
	}
	return ProjectPayload{
		ID:          project.ID,
		FacileID:    facileIDOf(project.FacileID),
		WorkspaceID: project.WorkspaceID,
		Name:        project.Name,
		Slug:        project.Slug,
		Description: &project.Description,
		Icon:        icon,
		ActorEmail:  actorEmail,
		UpdatedAt:   stamp(project.UpdatedAt),
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
