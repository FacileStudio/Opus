// Package task declares the task domain types, HTTP routing and service.
package task

import (
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
)

// TaskLabelResponse represents a label attached to a task.
type TaskLabelResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// TaskActor groups assignee metadata in a task response.
type TaskActor struct {
	ActorID    *int64 `json:"actorId,omitempty"`
	UserID     *int64 `json:"userId,omitempty"`
	ActorName  string `json:"actorName,omitempty"`
	ActorImage string `json:"actorImage,omitempty"`
}

// TaskTimestamps groups scheduling and audit timestamps in a task response.
type TaskTimestamps struct {
	StartDate *time.Time `json:"startDate,omitempty"`
	DueDate   *time.Time `json:"dueDate,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// TaskResponse is the JSON wire shape for a task.
type TaskResponse struct {
	ID             int64               `json:"id"`
	ProjectID      int64               `json:"projectId"`
	WorkspaceID    int64               `json:"workspaceId"`
	ColumnID       *int64              `json:"columnId,omitempty"`
	Number         int                 `json:"number"`
	Position       int                 `json:"position"`
	Title          string              `json:"title"`
	Description    string              `json:"description"`
	Status         string              `json:"status"`
	Priority       string              `json:"priority"`
	TaskActor                          `json:",inline"`
	TaskTimestamps                     `json:",inline"`
	Labels         []TaskLabelResponse `json:"labels"`
}

// ProjectColumnTasksResponse represents a column with its assigned tasks.
type ProjectColumnTasksResponse struct {
	ID      string         `json:"id"`
	Slug    string         `json:"slug"`
	Name    string         `json:"name"`
	IsFinal bool           `json:"isFinal"`
	Tasks   []TaskResponse `json:"tasks"`
}

// ProjectTasksData holds the project details and tasks organized by status.
type ProjectTasksData struct {
	ID            int64                        `json:"id"`
	Name          string                       `json:"name"`
	Slug          string                       `json:"slug"`
	Icon          string                       `json:"icon"`
	Description   string                       `json:"description"`
	IsPublic      bool                         `json:"isPublic"`
	WorkspaceID   int64                        `json:"workspaceId"`
	Columns       []ProjectColumnTasksResponse `json:"columns"`
	ArchivedTasks []TaskResponse               `json:"archivedTasks"`
	PlannedTasks  []TaskResponse               `json:"plannedTasks"`
}

// PaginationResponse represents pagination metadata.
type PaginationResponse struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalPages int   `json:"totalPages"`
}

// ProjectTasksResponse is the response shape for listing tasks in a project.
type ProjectTasksResponse struct {
	Data       ProjectTasksData   `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// CreateTaskRequest is the payload for creating a task in a project.
type CreateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status,omitempty"`
	Priority    string     `json:"priority,omitempty"`
	StartDate   *time.Time `json:"startDate,omitempty"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	UserID      *int64     `json:"userId,omitempty"`
	ActorID     *int64     `json:"actorId,omitempty"`
}

// UpdateTaskRequest is the payload for updating all fields of a task.
type UpdateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	ProjectID   int64      `json:"projectId"`
	Position    int        `json:"position"`
	StartDate   *time.Time `json:"startDate,omitempty"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	UserID      *int64     `json:"userId,omitempty"`
	ActorID     *int64     `json:"actorId,omitempty"`
}

// MoveTaskRequest is the payload for moving a task to another project.
type MoveTaskRequest struct {
	DestinationProjectID int64  `json:"destinationProjectId"`
	DestinationStatus    string `json:"destinationStatus,omitempty"`
}

// MoveTaskResponse is the response payload returned after moving a task.
type MoveTaskResponse struct {
	Task                 TaskResponse `json:"task"`
	SourceProjectID      int64        `json:"sourceProjectId"`
	DestinationProjectID int64        `json:"destinationProjectId"`
}

// UpdateStatusRequest is the payload for updating task status.
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// UpdatePriorityRequest is the payload for updating task priority.
type UpdatePriorityRequest struct {
	Priority string `json:"priority"`
}

// UpdateActorRequest is the payload for assigning or unassigning a task.
type UpdateActorRequest struct {
	UserID  *int64 `json:"userId,omitempty"`
	ActorID *int64 `json:"actorId,omitempty"`
}

// UpdateDueDateRequest is the payload for updating task due date.
type UpdateDueDateRequest struct {
	DueDate *time.Time `json:"dueDate"`
}

// UpdateTitleRequest is the payload for updating task title.
type UpdateTitleRequest struct {
	Title string `json:"title"`
}

// UpdateDescriptionRequest is the payload for updating task description.
type UpdateDescriptionRequest struct {
	Description string `json:"description"`
}

// BulkUpdateRequest is the payload for bulk operations on multiple tasks.
type BulkUpdateRequest struct {
	TaskIDs   []int64 `json:"taskIds"`
	Operation string  `json:"operation"`
	Value     *string `json:"value,omitempty"`
}

// BulkUpdateResponse is the response payload returned after a bulk operation.
type BulkUpdateResponse struct {
	Success      bool  `json:"success"`
	UpdatedCount int64 `json:"updatedCount"`
}

// ListTasksFilter carries query parameters for listing project tasks.
type ListTasksFilter struct {
	Status    string
	Priority  string
	ActorID   *int64
	Page      *int
	Limit     *int
	SortBy    string
	SortOrder string
	DueBefore *time.Time
	DueAfter  *time.Time
}

// TaskRecord groups a task entity with its actor and labels.
type TaskRecord struct {
	Task   schemas.Task
	Actor  *schemas.User
	Labels []schemas.Label
}

// ListTasksResult holds the entities returned from a project task query.
type ListTasksResult struct {
	Project schemas.Project
	Columns []schemas.Column
	Tasks   []TaskRecord
	Total   int64
}

// ExportTaskItem represents a task in the export bundle.
type ExportTaskItem struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"dueDate"`
	StartDate   *time.Time `json:"startDate"`
	UserID      *int64     `json:"userId"`
}

// ExportProjectMeta contains project metadata for task export.
type ExportProjectMeta struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ExportedAt  string `json:"exportedAt"`
}

// ExportTasksResponse is the root payload returned when exporting project tasks.
type ExportTasksResponse struct {
	Project ExportProjectMeta `json:"project"`
	Tasks   []ExportTaskItem  `json:"tasks"`
}

// ImportTaskItem represents an individual task to import.
type ImportTaskItem struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"dueDate"`
	StartDate   *time.Time `json:"startDate"`
	UserID      *int64     `json:"userId"`
}

// ImportTasksSummary reports the outcome of the import batch.
type ImportTasksSummary struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Failed     int `json:"failed"`
}

// ImportTasksRequest specifies the payload for importing tasks into a project.
type ImportTasksRequest struct {
	Tasks []ImportTaskItem `json:"tasks"`
}

// ImportTasksResponse is returned after tasks have been imported.
type ImportTasksResponse struct {
	ImportedAt string             `json:"importedAt"`
	Results    ImportTasksSummary `json:"results"`
}

