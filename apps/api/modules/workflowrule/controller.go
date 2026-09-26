package workflowrule

import (
	"context"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// Controller validates and maps workflow rule HTTP requests.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) list(ctx context.Context, userID, projectID int64) ([]WorkflowRuleResponse, error) {
	if projectID <= 0 {
		return nil, errors.Invalid("invalid project id")
	}
	rules, err := controller.service.ListByProject(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	responses := make([]WorkflowRuleResponse, 0, len(rules))
	for _, rule := range rules {
		responses = append(responses, toResponse(&rule))
	}
	return responses, nil
}

func (controller *Controller) upsert(ctx context.Context, userID, projectID int64, req *UpsertWorkflowRuleRequest) (*WorkflowRuleResponse, error) {
	if projectID <= 0 {
		return nil, errors.Invalid("invalid project id")
	}
	req.IntegrationType = strings.TrimSpace(req.IntegrationType)
	req.EventType = strings.TrimSpace(req.EventType)
	if req.IntegrationType == "" {
		return nil, errors.Invalid("integrationType is required")
	}
	if req.EventType == "" {
		return nil, errors.Invalid("eventType is required")
	}
	if req.ColumnID <= 0 {
		return nil, errors.Invalid("columnId is required")
	}
	rule, err := controller.service.Upsert(ctx, userID, projectID, req)
	if err != nil {
		return nil, err
	}
	resp := toResponse(rule)
	return &resp, nil
}

func (controller *Controller) delete(ctx context.Context, userID, ruleID int64) (*WorkflowRuleResponse, error) {
	if ruleID <= 0 {
		return nil, errors.Invalid("invalid workflow rule id")
	}
	rule, err := controller.service.Delete(ctx, userID, ruleID)
	if err != nil {
		return nil, err
	}
	resp := toResponse(rule)
	return &resp, nil
}

func toResponse(rule *schemas.WorkflowRule) WorkflowRuleResponse {
	return WorkflowRuleResponse{
		ID:              rule.ID,
		ProjectID:       rule.ProjectID,
		IntegrationType: rule.IntegrationType,
		EventType:       rule.EventType,
		ColumnID:        rule.ColumnID,
		CreatedAt:       rule.CreatedAt,
		UpdatedAt:       rule.UpdatedAt,
	}
}
