package sharelink

import (
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/Opus/apps/api/schemas"
)

type ShareLinkResponse struct {
	ID             int64      `json:"id"`
	ProjectID      int64      `json:"projectId"`
	Token          string     `json:"token"`
	Label          string     `json:"label"`
	ExpiresAt      *time.Time `json:"expiresAt"`
	IsActive       bool       `json:"isActive"`
	LastAccessedAt *time.Time `json:"lastAccessedAt"`
	AccessCount    int        `json:"accessCount"`
	CreatedBy      int64      `json:"createdBy"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// CreateShareLinkRequest carries the payload for creating a share link.
type CreateShareLinkRequest struct {
	ProjectID wireid.ID  `json:"projectId"`
	Label     *string    `json:"label,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// UpdateShareLinkRequest carries the payload for updating a share link.
type UpdateShareLinkRequest struct {
	Label     *string    `json:"label,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	IsActive  *bool      `json:"isActive,omitempty"`
}

// SharedColumnResponse represents a column in a shared project view.
type SharedColumnResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Position int    `json:"position"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
	IsFinal  bool   `json:"isFinal"`
}

// SharedTaskResponse represents a task in a shared project view.
type SharedTaskResponse struct {
	ID          int64      `json:"id"`
	Number      int        `json:"number"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	ColumnID    *int64     `json:"columnId,omitempty"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// SharedProjectData carries project, column, and task details for public viewing.
type SharedProjectData struct {
	ID          int64                  `json:"id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Icon        string                 `json:"icon"`
	Description string                 `json:"description"`
	IsPublic    bool                   `json:"isPublic"`
	WorkspaceID int64                  `json:"workspaceId"`
	Columns     []SharedColumnResponse `json:"columns"`
	Tasks       []SharedTaskResponse   `json:"tasks"`
}

// SharedProjectResult holds the project and its child entities loaded from storage.
type SharedProjectResult struct {
	Project schemas.Project
	Columns []schemas.Column
	Tasks   []schemas.Task
}
