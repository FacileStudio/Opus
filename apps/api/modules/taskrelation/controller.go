package taskrelation

import (
	"context"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// Controller handles request validation and DTO mapping for task relation endpoints.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func toSummaryResponse(summary *TaskSummary) *TaskSummaryResponse {
	if summary == nil {
		return nil
	}
	return &TaskSummaryResponse{
		ID:        summary.ID,
		Title:     summary.Title,
		Status:    summary.Status,
		Priority:  summary.Priority,
		Number:    summary.Number,
		ProjectID: summary.ProjectID,
		UserID:    summary.ActorID,
		ActorName: summary.ActorName,
	}
}

func toRecordResponse(record RelationRecord) TaskRelationResponse {
	return TaskRelationResponse{
		ID:           record.Relation.ID,
		SourceTaskID: record.Relation.SourceTaskID,
		TargetTaskID: record.Relation.TargetTaskID,
		RelationType: record.Relation.RelationType,
		CreatedAt:    record.Relation.CreatedAt,
		SourceTask:   toSummaryResponse(record.SourceTask),
		TargetTask:   toSummaryResponse(record.TargetTask),
	}
}

func toRelationResponse(relation *schemas.TaskRelation) *TaskRelationResponse {
	return &TaskRelationResponse{
		ID:           relation.ID,
		SourceTaskID: relation.SourceTaskID,
		TargetTaskID: relation.TargetTaskID,
		RelationType: relation.RelationType,
		CreatedAt:    relation.CreatedAt,
	}
}

func (controller *Controller) list(ctx context.Context, userID, taskID int64) ([]TaskRelationResponse, error) {
	if taskID <= 0 {
		return nil, errors.Invalid("invalid task id")
	}
	records, err := controller.service.ListForTask(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	responses := make([]TaskRelationResponse, 0, len(records))
	for _, record := range records {
		responses = append(responses, toRecordResponse(record))
	}
	return responses, nil
}

func (controller *Controller) create(ctx context.Context, userID int64, req *CreateTaskRelationRequest) (*TaskRelationResponse, error) {
	if req == nil {
		return nil, errors.Invalid("request body is required")
	}
	if req.SourceTaskID <= 0 || req.TargetTaskID <= 0 {
		return nil, errors.Invalid("invalid task id")
	}
	if req.SourceTaskID == req.TargetTaskID {
		return nil, errors.Invalid("Cannot create a relation between a task and itself")
	}
	relType := strings.TrimSpace(req.RelationType)
	if relType != "subtask" && relType != "blocks" && relType != "related" {
		return nil, errors.Invalid("Invalid relation type")
	}
	cleanReq := CreateTaskRelationRequest{
		SourceTaskID: req.SourceTaskID,
		TargetTaskID: req.TargetTaskID,
		RelationType: relType,
	}
	relation, err := controller.service.CreateRelation(ctx, userID, cleanReq)
	if err != nil {
		return nil, err
	}
	controller.service.notifyRelation(ctx, relation)
	return toRelationResponse(relation), nil
}

func (controller *Controller) delete(ctx context.Context, userID, id int64) (*TaskRelationResponse, error) {
	if id <= 0 {
		return nil, errors.Invalid("invalid relation id")
	}
	relation, err := controller.service.DeleteRelation(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	controller.service.notifyRelation(ctx, relation)
	return toRelationResponse(relation), nil
}
