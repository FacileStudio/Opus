// Package activity provides task activity tracking and comment functionality.
package activity

import (
	"encoding/json"
	"time"
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
	TaskID    int64           `json:"taskId"`
	UserID    *int64          `json:"userId,omitempty"`
	Message   *string         `json:"message,omitempty"`
	Content   *string         `json:"content,omitempty"`
	Type      string          `json:"type"`
	EventData json.RawMessage `json:"eventData,omitempty"`
}

// CommentActivityRequest is the payload for creating a comment on a task.
type CommentActivityRequest struct {
	TaskID  int64  `json:"taskId"`
	Comment string `json:"comment"`
	Content string `json:"content,omitempty"`
}

// UpdateCommentActivityRequest is the payload for updating an existing comment.
type UpdateCommentActivityRequest struct {
	ActivityID int64  `json:"activityId"`
	Comment    string `json:"comment"`
	Content    string `json:"content,omitempty"`
}

// DeleteCommentActivityRequest is the payload for deleting a comment.
type DeleteCommentActivityRequest struct {
	ActivityID int64 `json:"activityId"`
}

// ActivityParams carries database parameters for inserting an activity record.
type ActivityParams struct {
	TaskID    int64
	UserID    *int64
	Type      string
	Content   string
	EventData string
}
