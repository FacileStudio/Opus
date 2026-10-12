package label

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// Service provides database operations for labels.
type Service struct {
	orm        *gorm.DB
	controller *Controller
}

// NewService wires the label service with its database handle.
func NewService(db *gorm.DB) *Service {
	service := &Service{orm: db}
	service.controller = newController(service)
	return service
}

// RequireMember verifies the caller belongs to the workspace.
func (service *Service) RequireMember(ctx context.Context, userID, workspaceID int64) error {
	var count int64
	err := service.orm.WithContext(ctx).
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

func (service *Service) taskWorkspace(ctx context.Context, taskID int64) (int64, error) {
	var task schemas.Task
	err := service.orm.WithContext(ctx).Select("workspace_id").First(&task, taskID).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errors.NotFound("Task not found")
	}
	if err != nil {
		return 0, errors.Internal("failed to lookup task workspace", err)
	}
	return task.WorkspaceID, nil
}

func (service *Service) createOnTask(ctx context.Context, userID int64, req CreateLabelRequest) (*schemas.Label, error) {
	wsID, err := service.taskWorkspace(ctx, int64(*req.TaskID))
	if err != nil {
		return nil, err
	}
	if err := service.RequireMember(ctx, userID, wsID); err != nil {
		return nil, err
	}
	var existing schemas.Label
	err = service.orm.WithContext(ctx).
		Where("task_id = ? AND name = ?", int64(*req.TaskID), req.Name).
		First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Internal("failed to query label", err)
	}
	now := time.Now()
	label := schemas.Label{
		Name:        req.Name,
		Color:       req.Color,
		TaskID:      wireid.Int64Ptr(req.TaskID),
		WorkspaceID: wsID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := service.orm.WithContext(ctx).Create(&label).Error; err != nil {
		return nil, errors.Internal("failed to create label", err)
	}
	return &label, nil
}

func (service *Service) createOnWorkspace(ctx context.Context, userID int64, req CreateLabelRequest) (*schemas.Label, error) {
	if err := service.RequireMember(ctx, userID, int64(req.WorkspaceID)); err != nil {
		return nil, err
	}
	var existing schemas.Label
	err := service.orm.WithContext(ctx).
		Where("workspace_id = ? AND task_id IS NULL AND name = ?", int64(req.WorkspaceID), req.Name).
		First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Internal("failed to query label", err)
	}
	now := time.Now()
	label := schemas.Label{
		Name:        req.Name,
		Color:       req.Color,
		TaskID:      nil,
		WorkspaceID: int64(req.WorkspaceID),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := service.orm.WithContext(ctx).Create(&label).Error; err != nil {
		return nil, errors.Internal("failed to create label", err)
	}
	return &label, nil
}

// CreateLabel inserts a new label or resolves an existing one.
func (service *Service) CreateLabel(ctx context.Context, userID int64, req CreateLabelRequest) (*schemas.Label, error) {
	if req.TaskID != nil {
		return service.createOnTask(ctx, userID, req)
	}
	return service.createOnWorkspace(ctx, userID, req)
}

// GetLabel finds a single label by its identifier after verifying membership.
func (service *Service) GetLabel(ctx context.Context, userID, labelID int64) (*schemas.Label, error) {
	var label schemas.Label
	err := service.orm.WithContext(ctx).First(&label, labelID).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("Label not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to get label", err)
	}
	if err := service.RequireMember(ctx, userID, label.WorkspaceID); err != nil {
		return nil, err
	}
	return &label, nil
}

// GetLabelsByWorkspaceID returns every label for a workspace.
func (service *Service) GetLabelsByWorkspaceID(ctx context.Context, userID, workspaceID int64) ([]schemas.Label, error) {
	if err := service.RequireMember(ctx, userID, workspaceID); err != nil {
		return nil, err
	}
	var labels []schemas.Label
	err := service.orm.WithContext(ctx).Where("workspace_id = ?", workspaceID).Order("id asc").Find(&labels).Error
	if err != nil {
		return nil, errors.Internal("failed to list labels", err)
	}
	return labels, nil
}

// GetLabelsByTaskID returns every label assigned to a task.
func (service *Service) GetLabelsByTaskID(ctx context.Context, userID, taskID int64) ([]schemas.Label, error) {
	wsID, err := service.taskWorkspace(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if err := service.RequireMember(ctx, userID, wsID); err != nil {
		return nil, err
	}
	var labels []schemas.Label
	err = service.orm.WithContext(ctx).Where("task_id = ?", taskID).Order("id asc").Find(&labels).Error
	if err != nil {
		return nil, errors.Internal("failed to list task labels", err)
	}
	return labels, nil
}

// UpdateLabel modifies the name and color of an existing label.
func (service *Service) UpdateLabel(ctx context.Context, userID, labelID int64, req UpdateLabelRequest) (*schemas.Label, error) {
	label, err := service.GetLabel(ctx, userID, labelID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	updates := map[string]any{"name": req.Name, "color": req.Color, "updated_at": now}
	if err := service.orm.WithContext(ctx).Model(label).Updates(updates).Error; err != nil {
		return nil, errors.Internal("failed to update label", err)
	}
	label.Name = req.Name
	label.Color = req.Color
	label.UpdatedAt = now
	return label, nil
}

// AssignLabelToTask links a label to a task within the same workspace.
func (service *Service) AssignLabelToTask(ctx context.Context, userID, labelID, taskID int64) (*schemas.Label, error) {
	label, err := service.GetLabel(ctx, userID, labelID)
	if err != nil {
		return nil, err
	}
	taskWsID, err := service.taskWorkspace(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if label.WorkspaceID != taskWsID {
		return nil, errors.Invalid("Label and task must belong to the same workspace")
	}
	now := time.Now()
	if err := service.orm.WithContext(ctx).Model(label).Updates(map[string]any{"task_id": taskID, "updated_at": now}).Error; err != nil {
		return nil, errors.Internal("failed to assign label to task", err)
	}
	label.TaskID = &taskID
	label.UpdatedAt = now
	return label, nil
}

// UnassignLabelFromTask removes the task association from a label.
func (service *Service) UnassignLabelFromTask(ctx context.Context, userID, labelID int64) (*schemas.Label, error) {
	label, err := service.GetLabel(ctx, userID, labelID)
	if err != nil {
		return nil, err
	}
	if label.TaskID == nil {
		return nil, errors.Invalid("Label is not assigned to a task")
	}
	now := time.Now()
	if err := service.orm.WithContext(ctx).Model(label).Updates(map[string]any{"task_id": nil, "updated_at": now}).Error; err != nil {
		return nil, errors.Internal("failed to detach label from task", err)
	}
	label.TaskID = nil
	label.UpdatedAt = now
	return label, nil
}

// DeleteLabel deletes a label associated with a task.
func (service *Service) DeleteLabel(ctx context.Context, userID, labelID int64) (*schemas.Label, error) {
	label, err := service.GetLabel(ctx, userID, labelID)
	if err != nil {
		return nil, err
	}
	if label.TaskID == nil {
		return nil, errors.Invalid("Label is not associated with a task")
	}
	if err := service.orm.WithContext(ctx).Delete(&schemas.Label{}, labelID).Error; err != nil {
		return nil, errors.Internal("failed to delete label", err)
	}
	return label, nil
}
