package activity

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// Service handles persistence and workspace access verification for activities.
type Service struct {
	db         *gorm.DB
	controller *Controller
}

// NewService wires an activity service with its database handle.
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

func (service *Service) findComment(ctx context.Context, userID, activityID int64) (*schemas.Activity, error) {
	var activity schemas.Activity
	err := service.db.WithContext(ctx).First(&activity, "id = ?", activityID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("Comment not found or you are not the author")
		}
		return nil, errors.Internal("failed to load comment activity", err)
	}
	if activity.UserID == nil || *activity.UserID != userID {
		return nil, errors.NotFound("Comment not found or you are not the author")
	}
	if err := service.verifyTaskAccess(ctx, userID, activity.TaskID); err != nil {
		return nil, err
	}
	return &activity, nil
}

// ListForTask returns all activities for a task ordered by creation time descending.
func (service *Service) ListForTask(ctx context.Context, userID, taskID int64) ([]schemas.Activity, error) {
	if err := service.verifyTaskAccess(ctx, userID, taskID); err != nil {
		return nil, err
	}
	var activities []schemas.Activity
	err := service.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Order("created_at desc").
		Find(&activities).Error
	if err != nil {
		return nil, errors.Internal("failed to list activities", err)
	}
	return activities, nil
}

// Create adds a new activity entry after verifying workspace access.
func (service *Service) Create(ctx context.Context, callerUserID int64, params ActivityParams) (*schemas.Activity, error) {
	if err := service.verifyTaskAccess(ctx, callerUserID, params.TaskID); err != nil {
		return nil, err
	}
	now := time.Now()
	activity := schemas.Activity{
		TaskID:    params.TaskID,
		UserID:    params.UserID,
		Type:      params.Type,
		Content:   params.Content,
		EventData: params.EventData,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := service.db.WithContext(ctx).Create(&activity).Error; err != nil {
		return nil, errors.Internal("failed to create activity", err)
	}
	return &activity, nil
}

// CreateComment creates a comment activity after verifying workspace access.
func (service *Service) CreateComment(ctx context.Context, userID, taskID int64, content string) (*schemas.Activity, error) {
	if err := service.verifyTaskAccess(ctx, userID, taskID); err != nil {
		return nil, err
	}
	now := time.Now()
	activity := schemas.Activity{
		TaskID:    taskID,
		UserID:    &userID,
		Type:      "comment",
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := service.db.WithContext(ctx).Create(&activity).Error; err != nil {
		return nil, errors.Internal("failed to create comment activity", err)
	}
	return &activity, nil
}

// UpdateComment modifies comment activity content if the caller is the author.
func (service *Service) UpdateComment(ctx context.Context, userID, activityID int64, content string) (*schemas.Activity, error) {
	activity, err := service.findComment(ctx, userID, activityID)
	if err != nil {
		return nil, err
	}
	activity.Content = content
	activity.UpdatedAt = time.Now()
	if err := service.db.WithContext(ctx).Save(activity).Error; err != nil {
		return nil, errors.Internal("failed to update comment activity", err)
	}
	return activity, nil
}

// DeleteComment removes a comment activity if the caller is the author.
func (service *Service) DeleteComment(ctx context.Context, userID, activityID int64) (*schemas.Activity, error) {
	activity, err := service.findComment(ctx, userID, activityID)
	if err != nil {
		return nil, err
	}
	if err := service.db.WithContext(ctx).Delete(activity).Error; err != nil {
		return nil, errors.Internal("failed to delete comment activity", err)
	}
	return activity, nil
}
