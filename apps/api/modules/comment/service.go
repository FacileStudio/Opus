package comment

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// Service handles persistence and workspace membership checks for comments.
type Service struct {
	db         *gorm.DB
	controller *Controller
}

// NewService wires a comment service with its database handle.
func NewService(db *gorm.DB) *Service {
	service := &Service{db: db}
	service.controller = newController(service)
	return service
}

// RequireMember verifies that the caller belongs to the given workspace.
func (service *Service) RequireMember(ctx context.Context, userID, workspaceID int64) error {
	var count int64
	err := service.db.WithContext(ctx).
		Model(&schemas.WorkspaceMember{}).
		Where("user_id = ? AND workspace_id = ?", userID, workspaceID).
		Count(&count).Error
	if err != nil {
		return errors.Internal("failed to check workspace access", err)
	}
	if count == 0 {
		return errors.Forbidden("You don't have access to this workspace")
	}
	return nil
}

func (service *Service) verifyTaskAccess(ctx context.Context, userID, taskID int64) error {
	var task schemas.Task
	err := service.db.WithContext(ctx).
		Select("id, workspace_id").
		First(&task, "id = ?", taskID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("Task not found")
		}
		return errors.Internal("failed to query task", err)
	}
	return service.RequireMember(ctx, userID, task.WorkspaceID)
}

// ListForTask returns all comments for a task ordered by creation time.
func (service *Service) ListForTask(ctx context.Context, userID, taskID int64) ([]CommentRecord, error) {
	if err := service.verifyTaskAccess(ctx, userID, taskID); err != nil {
		return nil, err
	}
	type commentRow struct {
		ID          int64
		TaskID      int64
		UserID      int64
		Content     string
		CreatedAt   time.Time
		UpdatedAt   time.Time
		AuthorName  string
		AuthorImage string
	}
	var rows []commentRow
	err := service.db.WithContext(ctx).
		Table("comments").
		Select("comments.id, comments.task_id, comments.user_id, comments.content, comments.created_at, comments.updated_at, users.name as author_name, users.image as author_image").
		Joins("LEFT JOIN users ON users.id = comments.user_id").
		Where("comments.task_id = ?", taskID).
		Order("comments.created_at asc").
		Scan(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to list comments", err)
	}
	records := make([]CommentRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, CommentRecord{
			Comment: schemas.Comment{
				ID:        row.ID,
				TaskID:    row.TaskID,
				UserID:    row.UserID,
				Content:   row.Content,
				CreatedAt: row.CreatedAt,
				UpdatedAt: row.UpdatedAt,
			},
			AuthorName:  row.AuthorName,
			AuthorImage: row.AuthorImage,
		})
	}
	return records, nil
}

// Create adds a new comment to a task after verifying workspace access.
func (service *Service) Create(ctx context.Context, userID, taskID int64, content string) (*schemas.Comment, error) {
	if err := service.verifyTaskAccess(ctx, userID, taskID); err != nil {
		return nil, err
	}
	now := time.Now()
	comment := schemas.Comment{
		TaskID:    taskID,
		UserID:    userID,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := service.db.WithContext(ctx).Create(&comment).Error; err != nil {
		return nil, errors.Internal("failed to create comment", err)
	}
	return &comment, nil
}

// Update modifies comment content when the caller is the author and in workspace.
func (service *Service) Update(ctx context.Context, userID, commentID int64, content string) (*schemas.Comment, error) {
	var comment schemas.Comment
	err := service.db.WithContext(ctx).First(&comment, "id = ?", commentID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("Comment not found")
		}
		return nil, errors.Internal("failed to load comment", err)
	}
	if comment.UserID != userID {
		return nil, errors.Forbidden("Only the author can edit this comment")
	}
	if err := service.verifyTaskAccess(ctx, userID, comment.TaskID); err != nil {
		return nil, err
	}
	comment.Content = content
	comment.UpdatedAt = time.Now()
	if err := service.db.WithContext(ctx).Save(&comment).Error; err != nil {
		return nil, errors.Internal("failed to update comment", err)
	}
	return &comment, nil
}

// Delete removes a comment when the caller is the author and in workspace.
func (service *Service) Delete(ctx context.Context, userID, commentID int64) (*schemas.Comment, error) {
	var comment schemas.Comment
	err := service.db.WithContext(ctx).First(&comment, "id = ?", commentID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("Comment not found")
		}
		return nil, errors.Internal("failed to load comment", err)
	}
	if comment.UserID != userID {
		return nil, errors.Forbidden("Only the author can delete this comment")
	}
	if err := service.verifyTaskAccess(ctx, userID, comment.TaskID); err != nil {
		return nil, err
	}
	if err := service.db.WithContext(ctx).Delete(&comment).Error; err != nil {
		return nil, errors.Internal("failed to delete comment", err)
	}
	return &comment, nil
}
