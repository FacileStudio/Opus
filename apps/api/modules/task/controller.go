// Package task implements task request validation and response mapping.
package task

import (
	"context"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/realtime"
	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/tronc/errors"
)

// Controller handles request validation and response mapping for tasks.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) list(ctx context.Context, userID, projectID int64, filter ListTasksFilter) (*ProjectTasksResponse, error) {
	result, err := controller.service.List(ctx, userID, projectID, filter)
	if err != nil {
		return nil, err
	}
	resp := toProjectTasksResponse(result, filter)
	return &resp, nil
}

func (controller *Controller) get(ctx context.Context, userID, taskID int64) (*TaskResponse, error) {
	record, err := controller.service.Get(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	resp := recordResponse(*record)
	return &resp, nil
}

func (controller *Controller) create(ctx context.Context, userID, projectID int64, req *CreateTaskRequest) (*TaskResponse, error) {
	if req == nil {
		return nil, errors.Invalid("request body is required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.Invalid("Title is required")
	}
	record, err := controller.service.Create(ctx, userID, projectID, req)
	if err != nil {
		return nil, err
	}
	resp := recordResponse(*record)
	controller.service.publish(ctx, realtime.TypeTaskCreated, resp.ProjectID, resp.ID)
	return &resp, nil
}

func (controller *Controller) update(ctx context.Context, userID, taskID int64, req *UpdateTaskRequest) (*TaskResponse, error) {
	if req == nil {
		return nil, errors.Invalid("request body is required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.Invalid("Title is required")
	}
	record, err := controller.service.Update(ctx, userID, taskID, req)
	if err != nil {
		return nil, err
	}
	resp := recordResponse(*record)
	controller.service.publish(ctx, realtime.TypeTaskUpdated, resp.ProjectID, resp.ID)
	return &resp, nil
}

func (controller *Controller) delete(ctx context.Context, userID, taskID int64) (*TaskResponse, error) {
	record, err := controller.service.Delete(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	resp := recordResponse(*record)
	controller.service.publish(ctx, realtime.TypeTaskDeleted, resp.ProjectID, resp.ID)
	return &resp, nil
}

func (controller *Controller) move(ctx context.Context, userID, taskID int64, req *MoveTaskRequest) (*MoveTaskResponse, error) {
	if req == nil || req.DestinationProjectID <= 0 {
		return nil, errors.Invalid("destinationProjectId is required")
	}
	resp, err := controller.service.Move(ctx, userID, taskID, req)
	if err != nil {
		return nil, err
	}
	controller.service.publish(ctx, realtime.TypeTaskMoved, resp.SourceProjectID, resp.Task.ID)
	controller.service.publish(ctx, realtime.TypeTaskMoved, resp.DestinationProjectID, resp.Task.ID)
	return resp, nil
}

func (controller *Controller) updateStatus(ctx context.Context, userID, taskID int64, status string) (*TaskResponse, error) {
	if strings.TrimSpace(status) == "" {
		return nil, errors.Invalid("Status is required")
	}
	record, err := controller.service.UpdateStatus(ctx, userID, taskID, status)
	if err != nil {
		return nil, err
	}
	resp := recordResponse(*record)
	controller.service.publish(ctx, realtime.TypeTaskUpdated, resp.ProjectID, resp.ID)
	return &resp, nil
}

func (controller *Controller) updatePriority(ctx context.Context, userID, taskID int64, priority string) (*TaskResponse, error) {
	if strings.TrimSpace(priority) == "" {
		return nil, errors.Invalid("Priority is required")
	}
	record, err := controller.service.UpdatePriority(ctx, userID, taskID, priority)
	if err != nil {
		return nil, err
	}
	resp := recordResponse(*record)
	controller.service.publish(ctx, realtime.TypeTaskUpdated, resp.ProjectID, resp.ID)
	return &resp, nil
}

func (controller *Controller) updateActor(ctx context.Context, userID, taskID int64, req *UpdateActorRequest) (*TaskResponse, error) {
	var actorID *int64
	if req != nil {
		actorID = wireid.Int64Ptr(req.ActorID)
		if actorID == nil {
			actorID = wireid.Int64Ptr(req.UserID)
		}
	}
	record, err := controller.service.UpdateActor(ctx, userID, taskID, actorID)
	if err != nil {
		return nil, err
	}
	resp := recordResponse(*record)
	controller.service.publish(ctx, realtime.TypeTaskUpdated, resp.ProjectID, resp.ID)
	return &resp, nil
}

func (controller *Controller) updateDueDate(ctx context.Context, userID, taskID int64, req *UpdateDueDateRequest) (*TaskResponse, error) {
	var dueDate *time.Time
	if req != nil {
		dueDate = req.DueDate
	}
	record, err := controller.service.UpdateDueDate(ctx, userID, taskID, dueDate)
	if err != nil {
		return nil, err
	}
	resp := recordResponse(*record)
	controller.service.publish(ctx, realtime.TypeTaskUpdated, resp.ProjectID, resp.ID)
	return &resp, nil
}

func (controller *Controller) updateTitle(ctx context.Context, userID, taskID int64, title string) (*TaskResponse, error) {
	if strings.TrimSpace(title) == "" {
		return nil, errors.Invalid("Title is required")
	}
	record, err := controller.service.UpdateTitle(ctx, userID, taskID, title)
	if err != nil {
		return nil, err
	}
	resp := recordResponse(*record)
	controller.service.publish(ctx, realtime.TypeTaskUpdated, resp.ProjectID, resp.ID)
	return &resp, nil
}

func (controller *Controller) updateDescription(ctx context.Context, userID, taskID int64, description string) (*TaskResponse, error) {
	record, err := controller.service.UpdateDescription(ctx, userID, taskID, description)
	if err != nil {
		return nil, err
	}
	resp := recordResponse(*record)
	controller.service.publish(ctx, realtime.TypeTaskUpdated, resp.ProjectID, resp.ID)
	return &resp, nil
}

func (controller *Controller) bulk(ctx context.Context, userID int64, req *BulkUpdateRequest) (*BulkUpdateResponse, error) {
	if req == nil || len(req.TaskIDs) == 0 {
		return nil, errors.Invalid("taskIds list cannot be empty")
	}
	projectIDs := controller.service.projectIDsOf(ctx, req.TaskIDs)
	resp, err := controller.service.Bulk(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	for _, projectID := range projectIDs {
		controller.service.publish(ctx, realtime.TypeTaskUpdated, projectID, 0)
	}
	return resp, nil
}
