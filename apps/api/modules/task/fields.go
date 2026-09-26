// Package task implements task field modification methods.
package task

import (
	"context"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// UpdateStatus modifies the task status.
func (service *Service) UpdateStatus(ctx context.Context, userID, taskID int64, status string) (*TaskRecord, error) {
	task, err := service.verifyTaskMembership(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	if err := service.assertValidTaskStatus(ctx, task.ProjectID, status); err != nil {
		return nil, err
	}
	column, err := findColumnForStatus(ctx, service.db, task.ProjectID, status)
	if err != nil {
		return nil, err
	}
	var columnID *int64
	if column != nil {
		columnID = &column.ID
	}
	updates := map[string]interface{}{
		"status":     status,
		"column_id":  columnID,
		"updated_at": time.Now(),
	}
	if err := service.db.WithContext(ctx).Model(&schemas.Task{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
		return nil, errors.Internal("failed to update task status", err)
	}
	return service.Get(ctx, userID, taskID)
}

// UpdatePriority modifies the task priority.
func (service *Service) UpdatePriority(ctx context.Context, userID, taskID int64, priority string) (*TaskRecord, error) {
	if _, err := service.verifyTaskMembership(ctx, userID, taskID); err != nil {
		return nil, err
	}
	if err := assertValidPriority(priority); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"priority":   priority,
		"updated_at": time.Now(),
	}
	if err := service.db.WithContext(ctx).Model(&schemas.Task{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
		return nil, errors.Internal("failed to update task priority", err)
	}
	return service.Get(ctx, userID, taskID)
}

// UpdateActor modifies the task assignee.
func (service *Service) UpdateActor(ctx context.Context, userID, taskID int64, actorID *int64) (*TaskRecord, error) {
	if _, err := service.verifyTaskMembership(ctx, userID, taskID); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"actor_id":   actorID,
		"updated_at": time.Now(),
	}
	if err := service.db.WithContext(ctx).Model(&schemas.Task{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
		return nil, errors.Internal("failed to update task actor", err)
	}
	return service.Get(ctx, userID, taskID)
}

// UpdateDueDate modifies the task due date.
func (service *Service) UpdateDueDate(ctx context.Context, userID, taskID int64, dueDate *time.Time) (*TaskRecord, error) {
	if _, err := service.verifyTaskMembership(ctx, userID, taskID); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"due_date":   dueDate,
		"updated_at": time.Now(),
	}
	if err := service.db.WithContext(ctx).Model(&schemas.Task{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
		return nil, errors.Internal("failed to update task due date", err)
	}
	return service.Get(ctx, userID, taskID)
}

// UpdateTitle modifies the task title.
func (service *Service) UpdateTitle(ctx context.Context, userID, taskID int64, title string) (*TaskRecord, error) {
	if _, err := service.verifyTaskMembership(ctx, userID, taskID); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"title":      title,
		"updated_at": time.Now(),
	}
	if err := service.db.WithContext(ctx).Model(&schemas.Task{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
		return nil, errors.Internal("failed to update task title", err)
	}
	return service.Get(ctx, userID, taskID)
}

// UpdateDescription modifies the task description.
func (service *Service) UpdateDescription(ctx context.Context, userID, taskID int64, description string) (*TaskRecord, error) {
	if _, err := service.verifyTaskMembership(ctx, userID, taskID); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"description": description,
		"updated_at":  time.Now(),
	}
	if err := service.db.WithContext(ctx).Model(&schemas.Task{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
		return nil, errors.Internal("failed to update task description", err)
	}
	return service.Get(ctx, userID, taskID)
}
