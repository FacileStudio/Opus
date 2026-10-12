// Package activity provides task activity tracking and comment functionality.
package activity

import (
	"encoding/json"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
)

// ActivityResponse is the wire shape for an activity or comment record.
type ActivityResponse struct {
	ID        int64           `json:"id"`
	TaskID    int64           `json:"taskId"`
	UserID    *int64          `json:"userId"`
	Type      string          `json:"type"`
	Content   *string         `json:"content"`
	EventData json.RawMessage `json:"eventData"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

// CreateActivityRequest is the payload for creating a general activity entry.
type CreateActivityRequest struct {
	TaskID    wireid.ID       `json:"taskId"`
	UserID    *wireid.ID      `json:"userId,omitempty"`
	Message   *string         `json:"message,omitempty"`
	Content   *string         `json:"content,omitempty"`
	Type      string          `json:"type"`
	EventData json.RawMessage `json:"eventData,omitempty"`
}

// CommentActivityRequest is the payload for creating a comment on a task.
type CommentActivityRequest struct {
	TaskID  wireid.ID `json:"taskId"`
	Comment string    `json:"comment"`
	Content string    `json:"content,omitempty"`
}

// UpdateCommentActivityRequest is the payload for updating an existing comment.
type UpdateCommentActivityRequest struct {
	ActivityID wireid.ID `json:"activityId"`
	Comment    string    `json:"comment"`
	Content    string    `json:"content,omitempty"`
}

// DeleteCommentActivityRequest is the payload for deleting a comment.
type DeleteCommentActivityRequest struct {
	ActivityID wireid.ID `json:"activityId"`
}

// ActivityParams carries database parameters for inserting an activity record.
type ActivityParams struct {
	TaskID    int64
	UserID    *int64
	Type      string
	Content   string
	EventData string
}
