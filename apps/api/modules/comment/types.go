// Package comment declares the comment domain types, HTTP routing and service.
package comment

import (
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
)

// CommentUserResponse holds the author profile returned with a comment.
type CommentUserResponse struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

// CommentResponse is the JSON wire shape for comment endpoints.
type CommentResponse struct {
	ID        int64                `json:"id"`
	TaskID    int64                `json:"taskId"`
	UserID    int64                `json:"userId"`
	Content   string               `json:"content"`
	CreatedAt time.Time            `json:"createdAt"`
	UpdatedAt time.Time            `json:"updatedAt"`
	User      *CommentUserResponse `json:"user,omitempty"`
}

// CreateCommentRequest is the payload for creating a comment on a task.
type CreateCommentRequest struct {
	Content string `json:"content"`
}

// UpdateCommentRequest is the payload for updating an existing comment.
type UpdateCommentRequest struct {
	Content string `json:"content"`
}

// CommentRecord pairs a comment entity with author profile metadata.
type CommentRecord struct {
	Comment     schemas.Comment
	AuthorName  string
	AuthorImage string
}
