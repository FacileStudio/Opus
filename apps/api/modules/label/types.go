package label

import "time"

// LabelResponse represents the serialized label payload.
type LabelResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"createdAt"`
	TaskID      *int64    `json:"taskId"`
	WorkspaceID int64     `json:"workspaceId"`
}

// CreateLabelRequest is the payload for creating a label.
type CreateLabelRequest struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	WorkspaceID int64  `json:"workspaceId"`
	TaskID      *int64 `json:"taskId,omitempty"`
}

// UpdateLabelRequest is the payload for updating an existing label.
type UpdateLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// AssignTaskRequest is the payload for attaching a label to a task.
type AssignTaskRequest struct {
	TaskID int64 `json:"taskId"`
}
