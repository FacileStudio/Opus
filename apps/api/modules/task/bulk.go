// Package task implements bulk operations on tasks.
package task

import (
	"context"
	"strconv"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

func parseBulkDate(val *string) (*time.Time, error) {
	if val == nil || *val == "" {
		return nil, nil
	}
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, *val); err == nil {
			return &t, nil
		}
	}
	return nil, errors.Invalid("Invalid date value: " + *val)
}

func updateProjectTasksStatus(ctx context.Context, service *Service, projectID int64, tasks []schemas.Task, value string) (int64, error) {
	if err := service.assertValidTaskStatus(ctx, projectID, value); err != nil {
		return 0, err
	}
	col, err := findColumnForStatus(ctx, service.db, projectID, value)
	if err != nil {
		return 0, err
	}
	var columnID *int64
	if col != nil {
		columnID = &col.ID
	}
	var projectTaskIDs []int64
	for _, t := range tasks {
		if t.ProjectID == projectID {
			projectTaskIDs = append(projectTaskIDs, t.ID)
		}
	}
	res := service.db.WithContext(ctx).Model(&schemas.Task{}).
		Where("id IN ?", projectTaskIDs).
		Updates(map[string]interface{}{
			"status":     value,
			"column_id":  columnID,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return 0, errors.Internal("failed to update status", res.Error)
	}
	return res.RowsAffected, nil
}

func bulkUpdateStatus(ctx context.Context, service *Service, tasks []schemas.Task, value *string) (int64, error) {
	if value == nil || *value == "" {
		return 0, errors.Invalid("Status value is required")
	}
	projectIDsMap := make(map[int64]bool)
	for _, t := range tasks {
		projectIDsMap[t.ProjectID] = true
	}
	var totalUpdated int64
	for projectID := range projectIDsMap {
		updated, err := updateProjectTasksStatus(ctx, service, projectID, tasks, *value)
		if err != nil {
			return 0, err
		}
		totalUpdated += updated
	}
	return totalUpdated, nil
}

func bulkUpdatePriority(ctx context.Context, db *gorm.DB, taskIDs []int64, value *string) (int64, error) {
	if value == nil || *value == "" {
		return 0, errors.Invalid("Priority value is required")
	}
	if err := assertValidPriority(*value); err != nil {
		return 0, err
	}
	res := db.WithContext(ctx).Model(&schemas.Task{}).
		Where("id IN ?", taskIDs).
		Updates(map[string]interface{}{
			"priority":   *value,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return 0, errors.Internal("failed to update priority", res.Error)
	}
	return res.RowsAffected, nil
}

func bulkUpdateActor(ctx context.Context, db *gorm.DB, taskIDs []int64, value *string) (int64, error) {
	var actorID *int64
	if value != nil && *value != "" && *value != "0" {
		parsed, err := strconv.ParseInt(*value, 10, 64)
		if err != nil {
			return 0, errors.Invalid("Invalid actor id: " + *value)
		}
		actorID = &parsed
	}
	res := db.WithContext(ctx).Model(&schemas.Task{}).
		Where("id IN ?", taskIDs).
		Updates(map[string]interface{}{
			"actor_id":   actorID,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return 0, errors.Internal("failed to update actor", res.Error)
	}
	return res.RowsAffected, nil
}

func bulkDelete(ctx context.Context, db *gorm.DB, taskIDs []int64) (int64, error) {
	var count int64
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("source_task_id IN ? OR target_task_id IN ?", taskIDs, taskIDs).Delete(&schemas.TaskRelation{}).Error; err != nil {
			return errors.Internal("failed to delete task relations", err)
		}
		if err := tx.Where("task_id IN ?", taskIDs).Delete(&schemas.Label{}).Error; err != nil {
			return errors.Internal("failed to delete task labels", err)
		}
		if err := tx.Where("task_id IN ?", taskIDs).Delete(&schemas.Comment{}).Error; err != nil {
			return errors.Internal("failed to delete task comments", err)
		}
		res := tx.Where("id IN ?", taskIDs).Delete(&schemas.Task{})
		if res.Error != nil {
			return errors.Internal("failed to delete tasks", res.Error)
		}
		count = res.RowsAffected
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func bulkUpdateDueDate(ctx context.Context, db *gorm.DB, taskIDs []int64, value *string) (int64, error) {
	parsedDate, err := parseBulkDate(value)
	if err != nil {
		return 0, err
	}
	res := db.WithContext(ctx).Model(&schemas.Task{}).
		Where("id IN ?", taskIDs).
		Updates(map[string]interface{}{
			"due_date":   parsedDate,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return 0, errors.Internal("failed to update due date", res.Error)
	}
	return res.RowsAffected, nil
}

// Bulk applies batch operations across tasks in the same workspace.
func (service *Service) Bulk(ctx context.Context, userID int64, req *BulkUpdateRequest) (*BulkUpdateResponse, error) {
	if req == nil || len(req.TaskIDs) == 0 {
		return nil, errors.Invalid("taskIds list cannot be empty")
	}
	var tasks []schemas.Task
	if err := service.db.WithContext(ctx).Where("id IN ?", wireid.Int64s(req.TaskIDs)).Find(&tasks).Error; err != nil {
		return nil, errors.Internal("failed to fetch tasks", err)
	}
	if len(tasks) == 0 {
		return nil, errors.NotFound("No tasks found")
	}
	workspaceIDsMap := make(map[int64]bool)
	taskIDs := make([]int64, len(tasks))
	for i, t := range tasks {
		workspaceIDsMap[t.WorkspaceID] = true
		taskIDs[i] = t.ID
	}
	if len(workspaceIDsMap) > 1 {
		return nil, errors.Invalid("All tasks must belong to the same workspace")
	}
	if err := service.verifyWorkspaceMembership(ctx, userID, tasks[0].WorkspaceID); err != nil {
		return nil, err
	}
	var updatedCount int64
	var err error
	switch req.Operation {
	case "updateStatus":
		updatedCount, err = bulkUpdateStatus(ctx, service, tasks, req.Value)
	case "updatePriority":
		updatedCount, err = bulkUpdatePriority(ctx, service.db, taskIDs, req.Value)
	case "updateActor":
		updatedCount, err = bulkUpdateActor(ctx, service.db, taskIDs, req.Value)
	case "delete":
		updatedCount, err = bulkDelete(ctx, service.db, taskIDs)
	case "updateDueDate":
		updatedCount, err = bulkUpdateDueDate(ctx, service.db, taskIDs, req.Value)
	case "addLabel":
		updatedCount, err = bulkAddLabel(ctx, service, tasks, req.Value)
	default:
		return nil, errors.Invalid("Unknown operation: " + req.Operation)
	}
	if err != nil {
		return nil, err
	}
	return &BulkUpdateResponse{Success: true, UpdatedCount: updatedCount}, nil
}
