package realtime

// TypeTaskCreated indicates a new task was created.
const TypeTaskCreated = "TASK_CREATED"

// TypeTaskUpdated indicates a task was modified.
const TypeTaskUpdated = "TASK_UPDATED"

// TypeTaskDeleted indicates a task was deleted.
const TypeTaskDeleted = "TASK_DELETED"

// TypeTaskMoved indicates a task was moved between columns or projects.
const TypeTaskMoved = "TASK_MOVED"

// TypeTaskLabelUpdated indicates task labels were modified.
const TypeTaskLabelUpdated = "TASK_LABEL_UPDATED"

// TypeTaskRelationUpdated indicates relations between tasks were modified.
const TypeTaskRelationUpdated = "TASK_RELATION_UPDATED"

// TypeCommentUpdated indicates comments or activities were updated.
const TypeCommentUpdated = "COMMENT_UPDATED"

// BroadcastMessage represents a live broadcast payload dispatched to clients.
type BroadcastMessage struct {
	Type         string `json:"type"`
	ProjectID    string `json:"projectId"`
	WorkspaceID  string `json:"workspaceId,omitempty"`
	TaskID       string `json:"taskId,omitempty"`
	SourceTaskID string `json:"sourceTaskId,omitempty"`
	TargetTaskID string `json:"targetTaskId,omitempty"`
}

// Envelope wraps a broadcast message and routing metadata across the adapter bus.
type Envelope struct {
	ProjectID          string           `json:"projectId,omitempty"`
	WorkspaceID        string           `json:"workspaceId,omitempty"`
	ExcludeInitiatorID string           `json:"excludeInitiatorId,omitempty"`
	Message            BroadcastMessage `json:"message"`
}
