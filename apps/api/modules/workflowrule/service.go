package workflowrule

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// Service provides persistence and validation for workflow automation rules.
type Service struct {
	db         *gorm.DB
	controller *Controller
}

// NewService instantiates a new workflow rule service.
func NewService(db *gorm.DB) *Service {
	service := &Service{db: db}
	service.controller = newController(service)
	return service
}

// RequireMember verifies caller belongs to the given workspace.
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

func (service *Service) verifyProjectAccess(ctx context.Context, userID, projectID int64) error {
	var project schemas.Project
	err := service.db.WithContext(ctx).
		Select("id, workspace_id").
		First(&project, "id = ?", projectID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("Project not found")
		}
		return errors.Internal("failed to query project", err)
	}
	return service.RequireMember(ctx, userID, project.WorkspaceID)
}

func (service *Service) verifyRuleAccess(ctx context.Context, userID, ruleID int64) (*schemas.WorkflowRule, error) {
	var rule schemas.WorkflowRule
	err := service.db.WithContext(ctx).First(&rule, "id = ?", ruleID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("Workflow rule not found")
		}
		return nil, errors.Internal("failed to query workflow rule", err)
	}
	if err := service.verifyProjectAccess(ctx, userID, rule.ProjectID); err != nil {
		return nil, err
	}
	return &rule, nil
}

// ListByProject returns all workflow rules for a project.
func (service *Service) ListByProject(ctx context.Context, userID, projectID int64) ([]schemas.WorkflowRule, error) {
	if err := service.verifyProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}
	var rules []schemas.WorkflowRule
	err := service.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at asc").
		Find(&rules).Error
	if err != nil {
		return nil, errors.Internal("failed to list workflow rules", err)
	}
	return rules, nil
}

// Upsert creates or updates a rule for a given project, integration, and event type.
func (service *Service) Upsert(ctx context.Context, userID, projectID int64, req *UpsertWorkflowRuleRequest) (*schemas.WorkflowRule, error) {
	if err := service.verifyProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}
	var column schemas.Column
	err := service.db.WithContext(ctx).First(&column, "id = ? AND project_id = ?", int64(req.ColumnID), projectID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Invalid("Column not found in project")
		}
		return nil, errors.Internal("failed to query column", err)
	}
	var existing schemas.WorkflowRule
	err = service.db.WithContext(ctx).
		Where("project_id = ? AND integration_type = ? AND event_type = ?", projectID, req.IntegrationType, req.EventType).
		First(&existing).Error
	now := time.Now()
	if err == nil {
		existing.ColumnID = int64(req.ColumnID)
		existing.UpdatedAt = now
		if err := service.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return nil, errors.Internal("failed to update workflow rule", err)
		}
		return &existing, nil
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Internal("failed to lookup workflow rule", err)
	}
	newRule := schemas.WorkflowRule{
		ProjectID:       projectID,
		IntegrationType: req.IntegrationType,
		EventType:       req.EventType,
		ColumnID:        int64(req.ColumnID),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := service.db.WithContext(ctx).Create(&newRule).Error; err != nil {
		return nil, errors.Internal("failed to create workflow rule", err)
	}
	return &newRule, nil
}

// Delete removes a workflow rule.
func (service *Service) Delete(ctx context.Context, userID, ruleID int64) (*schemas.WorkflowRule, error) {
	rule, err := service.verifyRuleAccess(ctx, userID, ruleID)
	if err != nil {
		return nil, err
	}
	if err := service.db.WithContext(ctx).Delete(rule).Error; err != nil {
		return nil, errors.Internal("failed to delete workflow rule", err)
	}
	return rule, nil
}
