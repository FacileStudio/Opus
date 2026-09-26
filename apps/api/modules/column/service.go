package column

import (
	"context"
	"fmt"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// Service handles database operations and workspace authorization for columns.
type Service struct {
	db         *gorm.DB
	controller *Controller
}

// NewService constructs a new column service with its database handle.
func NewService(db *gorm.DB) *Service {
	service := &Service{db: db}
	service.controller = newController(service)
	return service
}

func (service *Service) verifyProjectMembership(ctx context.Context, userID, projectID int64) error {
	var project schemas.Project
	if err := service.db.WithContext(ctx).Select("id, workspace_id").First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("project not found")
		}
		return errors.Internal("failed to load project", err)
	}
	var count int64
	err := service.db.WithContext(ctx).Model(&schemas.WorkspaceMember{}).
		Where("user_id = ? AND workspace_id = ?", userID, project.WorkspaceID).
		Count(&count).Error
	if err != nil {
		return errors.Internal("failed to verify workspace membership", err)
	}
	if count == 0 {
		return errors.Forbidden("You don't have access to this workspace")
	}
	return nil
}

func (service *Service) verifyColumnMembership(ctx context.Context, userID, columnID int64) (*schemas.Column, error) {
	var col schemas.Column
	if err := service.db.WithContext(ctx).First(&col, columnID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("Column not found")
		}
		return nil, errors.Internal("failed to load column", err)
	}
	if err := service.verifyProjectMembership(ctx, userID, col.ProjectID); err != nil {
		return nil, err
	}
	return &col, nil
}

// List returns all columns for a project ordered by position.
func (service *Service) List(ctx context.Context, userID, projectID int64) ([]schemas.Column, error) {
	if err := service.verifyProjectMembership(ctx, userID, projectID); err != nil {
		return nil, err
	}
	var cols []schemas.Column
	err := service.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("position asc").
		Find(&cols).Error
	if err != nil {
		return nil, errors.Internal("failed to list columns", err)
	}
	return cols, nil
}

// Create inserts a new column after validating uniqueness and computing position.
func (service *Service) Create(ctx context.Context, userID, projectID int64, col *schemas.Column) (*schemas.Column, error) {
	if err := service.verifyProjectMembership(ctx, userID, projectID); err != nil {
		return nil, err
	}
	var existingCount int64
	err := service.db.WithContext(ctx).Model(&schemas.Column{}).
		Where("project_id = ? AND slug = ?", projectID, col.Slug).
		Count(&existingCount).Error
	if err != nil {
		return nil, errors.Internal("failed to check existing column", err)
	}
	if existingCount > 0 {
		return nil, errors.Conflict(fmt.Sprintf("Column with slug %q already exists in this project", col.Slug))
	}
	var maxPos *int
	err = service.db.WithContext(ctx).Model(&schemas.Column{}).
		Where("project_id = ?", projectID).
		Select("COALESCE(MAX(position), -1)").
		Scan(&maxPos).Error
	if err != nil {
		return nil, errors.Internal("failed to get max position", err)
	}
	pos := 0
	if maxPos != nil {
		pos = *maxPos + 1
	}
	now := time.Now()
	col.ProjectID = projectID
	col.Position = pos
	col.CreatedAt = now
	col.UpdatedAt = now
	if err := service.db.WithContext(ctx).Create(col).Error; err != nil {
		return nil, errors.Internal("failed to create column", err)
	}
	return col, nil
}

// Update modifies the requested column fields and refreshes the record.
func (service *Service) Update(ctx context.Context, userID, columnID int64, fields map[string]interface{}) (*schemas.Column, error) {
	existing, err := service.verifyColumnMembership(ctx, userID, columnID)
	if err != nil {
		return nil, err
	}
	if len(fields) > 0 {
		fields["updated_at"] = time.Now()
		if err := service.db.WithContext(ctx).Model(&schemas.Column{}).Where("id = ?", columnID).Updates(fields).Error; err != nil {
			return nil, errors.Internal("failed to update column", err)
		}
		if err := service.db.WithContext(ctx).First(existing, columnID).Error; err != nil {
			return nil, errors.Internal("failed to reload column", err)
		}
	}
	return existing, nil
}

// Delete removes a column if no tasks are assigned to it.
func (service *Service) Delete(ctx context.Context, userID, columnID int64) (*schemas.Column, error) {
	existing, err := service.verifyColumnMembership(ctx, userID, columnID)
	if err != nil {
		return nil, err
	}
	var taskCount int64
	err = service.db.WithContext(ctx).Model(&schemas.Task{}).
		Where("column_id = ?", columnID).
		Count(&taskCount).Error
	if err != nil {
		return nil, errors.Internal("failed to check tasks", err)
	}
	if taskCount > 0 {
		return nil, errors.Conflict("Cannot delete column that contains tasks. Move or delete tasks first.")
	}
	if err := service.db.WithContext(ctx).Delete(existing).Error; err != nil {
		return nil, errors.Internal("failed to delete column", err)
	}
	return existing, nil
}

// Reorder applies position changes to multiple columns within a transaction.
func (service *Service) Reorder(ctx context.Context, userID, projectID int64, orders []ColumnOrderItem) ([]schemas.Column, error) {
	if err := service.verifyProjectMembership(ctx, userID, projectID); err != nil {
		return nil, err
	}
	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range orders {
			res := tx.Model(&schemas.Column{}).
				Where("id = ? AND project_id = ?", item.ID, projectID).
				Update("position", item.Position)
			if res.Error != nil {
				return errors.Internal("failed to update column position", res.Error)
			}
			if res.RowsAffected == 0 {
				return errors.Invalid(fmt.Sprintf("Column %d does not belong to this project", item.ID))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var cols []schemas.Column
	if err := service.db.WithContext(ctx).Where("project_id = ?", projectID).Order("position asc").Find(&cols).Error; err != nil {
		return nil, errors.Internal("failed to list reordered columns", err)
	}
	return cols, nil
}
