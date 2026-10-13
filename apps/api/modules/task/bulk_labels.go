package task

import (
	"context"
	stderrors "errors"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// parseBulkLabelID reads the source label id carried in a bulk addLabel value.
func parseBulkLabelID(value *string) (int64, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return 0, errors.Invalid("Label id is required")
	}
	id, err := strconv.ParseInt(strings.TrimSpace(*value), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.Invalid("Invalid label id: " + *value)
	}
	return id, nil
}

// loadBulkLabel resolves the workspace label whose name and color are copied.
func loadBulkLabel(ctx context.Context, db *gorm.DB, labelID int64) (*schemas.Label, error) {
	var label schemas.Label
	err := db.WithContext(ctx).First(&label, labelID).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("Label not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to load label", err)
	}
	return &label, nil
}

// ensureTaskLabel attaches a copy of the source label to one task when absent.
func ensureTaskLabel(ctx context.Context, db *gorm.DB, source schemas.Label, taskID int64) error {
	var existing schemas.Label
	err := db.WithContext(ctx).
		Where("task_id = ? AND workspace_id = ? AND name = ?", taskID, source.WorkspaceID, source.Name).
		First(&existing).Error
	if err == nil {
		return nil
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Internal("failed to query task label", err)
	}
	now := time.Now()
	label := schemas.Label{
		WorkspaceID: source.WorkspaceID,
		TaskID:      &taskID,
		Name:        source.Name,
		Color:       source.Color,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.WithContext(ctx).Create(&label).Error; err != nil {
		return errors.Internal("failed to add label to task", err)
	}
	return nil
}

// bulkAddLabel copies one workspace label onto every selected task.
func bulkAddLabel(ctx context.Context, service *Service, tasks []schemas.Task, value *string) (int64, error) {
	labelID, err := parseBulkLabelID(value)
	if err != nil {
		return 0, err
	}
	source, err := loadBulkLabel(ctx, service.db, labelID)
	if err != nil {
		return 0, err
	}
	for _, task := range tasks {
		if task.WorkspaceID != source.WorkspaceID {
			return 0, errors.Invalid("Label and tasks must belong to the same workspace")
		}
		if err := ensureTaskLabel(ctx, service.db, *source, task.ID); err != nil {
			return 0, err
		}
	}
	return int64(len(tasks)), nil
}
