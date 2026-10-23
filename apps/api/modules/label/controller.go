package label

import (
	"context"
	"strconv"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/internal/realtime"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// Controller handles request validation and DTO mapping for label endpoints.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func toResponse(label *schemas.Label) *LabelResponse {
	return &LabelResponse{
		ID:          label.ID,
		Name:        label.Name,
		Color:       label.Color,
		CreatedAt:   label.CreatedAt,
		TaskID:      label.TaskID,
		WorkspaceID: label.WorkspaceID,
	}
}

func toListResponse(labels []schemas.Label) []LabelResponse {
	out := make([]LabelResponse, 0, len(labels))
	for i := range labels {
		out = append(out, *toResponse(&labels[i]))
	}
	return out
}

func (controller *Controller) create(ctx context.Context, userID int64, req *CreateLabelRequest) (*LabelResponse, error) {
	if req == nil {
		return nil, errors.Invalid("request body is required")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.Invalid("Label name is required")
	}
	color := strings.TrimSpace(req.Color)
	if color == "" {
		return nil, errors.Invalid("Label color is required")
	}
	if req.TaskID == nil && req.WorkspaceID <= 0 {
		return nil, errors.Invalid("Workspace ID is required")
	}
	cleanReq := CreateLabelRequest{
		Name:        name,
		Color:       color,
		WorkspaceID: req.WorkspaceID,
		TaskID:      req.TaskID,
	}
	label, err := controller.service.CreateLabel(ctx, userID, cleanReq)
	if err != nil {
		return nil, err
	}
	return toResponse(label), nil
}

func (controller *Controller) get(ctx context.Context, userID, labelID int64) (*LabelResponse, error) {
	if labelID <= 0 {
		return nil, errors.Invalid("invalid label id")
	}
	label, err := controller.service.GetLabel(ctx, userID, labelID)
	if err != nil {
		return nil, err
	}
	return toResponse(label), nil
}

func (controller *Controller) listByWorkspace(ctx context.Context, userID, workspaceID int64) ([]LabelResponse, error) {
	if workspaceID <= 0 {
		return nil, errors.Invalid("invalid workspace id")
	}
	labels, err := controller.service.GetLabelsByWorkspaceID(ctx, userID, workspaceID)
	if err != nil {
		return nil, err
	}
	return toListResponse(labels), nil
}

func (controller *Controller) listByTask(ctx context.Context, userID, taskID int64) ([]LabelResponse, error) {
	if taskID <= 0 {
		return nil, errors.Invalid("invalid task id")
	}
	labels, err := controller.service.GetLabelsByTaskID(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	return toListResponse(labels), nil
}

func (controller *Controller) update(ctx context.Context, userID, labelID int64, req *UpdateLabelRequest) (*LabelResponse, error) {
	if labelID <= 0 {
		return nil, errors.Invalid("invalid label id")
	}
	if req == nil {
		return nil, errors.Invalid("request body is required")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.Invalid("Label name is required")
	}
	color := strings.TrimSpace(req.Color)
	if color == "" {
		return nil, errors.Invalid("Label color is required")
	}
	cleanReq := UpdateLabelRequest{Name: name, Color: color}
	label, err := controller.service.UpdateLabel(ctx, userID, labelID, cleanReq)
	if err != nil {
		return nil, err
	}
	return toResponse(label), nil
}

func (controller *Controller) assignToTask(ctx context.Context, userID, labelID int64, req *AssignTaskRequest) (*LabelResponse, error) {
	if labelID <= 0 {
		return nil, errors.Invalid("invalid label id")
	}
	if req == nil || req.TaskID <= 0 {
		return nil, errors.Invalid("valid task ID is required")
	}
	label, err := controller.service.AssignLabelToTask(ctx, userID, labelID, int64(req.TaskID))
	if err != nil {
		return nil, err
	}
	controller.service.notify(ctx, realtime.BroadcastMessage{Type: realtime.TypeTaskLabelUpdated, TaskID: strconv.FormatInt(int64(req.TaskID), 10)}, int64(req.TaskID))
	return toResponse(label), nil
}

func (controller *Controller) unassignFromTask(ctx context.Context, userID, labelID int64) (*LabelResponse, error) {
	if labelID <= 0 {
		return nil, errors.Invalid("invalid label id")
	}
	before, err := controller.service.GetLabel(ctx, userID, labelID)
	if err != nil {
		return nil, err
	}
	label, err := controller.service.UnassignLabelFromTask(ctx, userID, labelID)
	if err != nil {
		return nil, err
	}
	if before.TaskID != nil {
		controller.service.notify(ctx, realtime.BroadcastMessage{Type: realtime.TypeTaskLabelUpdated, TaskID: strconv.FormatInt(*before.TaskID, 10)}, *before.TaskID)
	}
	return toResponse(label), nil
}

func (controller *Controller) delete(ctx context.Context, userID, labelID int64) (*LabelResponse, error) {
	if labelID <= 0 {
		return nil, errors.Invalid("invalid label id")
	}
	label, err := controller.service.DeleteLabel(ctx, userID, labelID)
	if err != nil {
		return nil, err
	}
	if label.TaskID != nil {
		controller.service.notify(ctx, realtime.BroadcastMessage{Type: realtime.TypeTaskLabelUpdated, TaskID: strconv.FormatInt(*label.TaskID, 10)}, *label.TaskID)
	}
	return toResponse(label), nil
}
