// Package project manages project workspaces, columns, and workflows.
package project

import "time"

// ProjectStatistics carries aggregate progress metrics for a project.
type ProjectStatistics struct {
	CompletionPercentage int        `json:"completionPercentage"`
	TotalTasks           int        `json:"totalTasks"`
	DueDate              *time.Time `json:"dueDate"`
}

// ProjectResponse represents the project wire payload.
type ProjectResponse struct {
	ID          int64              `json:"id"`
	WorkspaceID int64              `json:"workspaceId"`
	Slug        string             `json:"slug"`
	Icon        string             `json:"icon"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	IsPublic    bool               `json:"isPublic"`
	ArchivedAt  *time.Time         `json:"archivedAt"`
	CreatedAt   time.Time          `json:"createdAt"`
	Statistics  *ProjectStatistics `json:"statistics,omitempty"`
}

// CreateProjectRequest is the payload for creating a project.
type CreateProjectRequest struct {
	Name        string `json:"name"`
	WorkspaceID int64  `json:"workspaceId"`
	Icon        string `json:"icon,omitempty"`
	Slug        string `json:"slug,omitempty"`
}

// UpdateProjectRequest is the payload for updating a project.
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	IsPublic    *bool   `json:"isPublic,omitempty"`
}
