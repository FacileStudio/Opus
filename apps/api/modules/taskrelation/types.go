// Package taskrelation declares task relation domain types, HTTP routing and service.
package taskrelation

import (
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/Opus/apps/api/schemas"
)

type TaskSummary struct {
	ID        int64
	Title     string
	Status    string
	Priority  string
	Number    int
	ProjectID int64
	ActorID   *int64
	ActorName *string
}

// RelationRecord pairs a task relation entity with source and target task summaries.
type RelationRecord struct {
	Relation   schemas.TaskRelation
	SourceTask *TaskSummary
	TargetTask *TaskSummary
}

// TaskSummaryResponse represents the serialized task summary attached to a relation.
type TaskSummaryResponse struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	Status    string  `json:"status"`
	Priority  string  `json:"priority"`
	Number    int     `json:"number"`
	ProjectID int64   `json:"projectId"`
	UserID    *int64  `json:"userId"`
	ActorName *string `json:"actorName"`
}

// TaskRelationResponse is the JSON wire shape for task relation endpoints.
type TaskRelationResponse struct {
	ID           int64                `json:"id"`
	SourceTaskID int64                `json:"sourceTaskId"`
	TargetTaskID int64                `json:"targetTaskId"`
	RelationType string               `json:"relationType"`
	CreatedAt    time.Time            `json:"createdAt"`
	SourceTask   *TaskSummaryResponse `json:"sourceTask,omitempty"`
	TargetTask   *TaskSummaryResponse `json:"targetTask,omitempty"`
}

// CreateTaskRelationRequest is the payload for creating a relation between two tasks.
type CreateTaskRelationRequest struct {
	SourceTaskID wireid.ID `json:"sourceTaskId"`
	TargetTaskID wireid.ID `json:"targetTaskId"`
	RelationType string    `json:"relationType"`
}
