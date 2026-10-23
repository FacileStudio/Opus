// Package task implements task request validation and response mapping.
package task

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// Controller handles request validation and response mapping for tasks.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func toLabelResponses(labels []schemas.Label) []TaskLabelResponse {
	if len(labels) == 0 {
		return []TaskLabelResponse{}
	}
	out := make([]TaskLabelResponse, len(labels))
	for i, l := range labels {
		out[i] = TaskLabelResponse{
			ID:    l.ID,
			Name:  l.Name,
			Color: l.Color,
		}
	}
	return out
}

func recordResponse(rec TaskRecord) TaskResponse {
	resp := toResponse(rec.Task, rec.Actor, toLabelResponses(rec.Labels))
	resp.SubtaskTotal = rec.SubtaskTotal
	resp.SubtaskDone = rec.SubtaskDone
	return resp
}

func toResponse(task schemas.Task, actor *schemas.User, labels []TaskLabelResponse) TaskResponse {
	resp := TaskResponse{
		ID:          task.ID,
		ProjectID:   task.ProjectID,
		WorkspaceID: task.WorkspaceID,
		ColumnID:    task.ColumnID,
		Number:      task.Number,
		Position:    task.Position,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		TaskActor: TaskActor{
			ActorID: task.ActorID,
			UserID:  task.ActorID,
		},
		TaskTimestamps: TaskTimestamps{
			StartDate: task.StartDate,
			DueDate:   task.DueDate,
			CreatedAt: task.CreatedAt,
			UpdatedAt: task.UpdatedAt,
		},
		Labels: labels,
	}
	if labels == nil {
		resp.Labels = []TaskLabelResponse{}
	}
	if actor != nil {
		resp.ActorName = actor.Name
		resp.ActorImage = actor.Image
	}
	return resp
}

func calculatePagination(total int64, pagePtr, limitPtr *int) (int, int, int) {
	page := 1
	if pagePtr != nil && *pagePtr > 0 {
		page = *pagePtr
	}
	if pagePtr == nil && limitPtr == nil {
		return page, int(total), 1
	}
	pageSize := 50
	if limitPtr != nil && *limitPtr > 0 {
		pageSize = *limitPtr
		if pageSize > 100 {
			pageSize = 100
		}
	}
	totalPages := 1
	if pageSize > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
		if totalPages < 1 {
			totalPages = 1
		}
	}
	return page, pageSize, totalPages
}

func toProjectTasksResponse(result *ListTasksResult, filter ListTasksFilter) ProjectTasksResponse {
	cols := make([]ProjectColumnTasksResponse, len(result.Columns))
	for i, col := range result.Columns {
		tasks := make([]TaskResponse, 0)
		for _, rec := range result.Tasks {
			if rec.Task.Status == col.Slug {
				tasks = append(tasks, recordResponse(rec))
			}
		}
		cols[i] = ProjectColumnTasksResponse{
			ID:      col.Slug,
			Slug:    col.Slug,
			Name:    col.Name,
			IsFinal: col.IsFinal,
			Tasks:   tasks,
		}
	}
	archived := make([]TaskResponse, 0)
	planned := make([]TaskResponse, 0)
	for _, rec := range result.Tasks {
		if rec.Task.Status == "archived" {
			archived = append(archived, recordResponse(rec))
		} else if rec.Task.Status == "planned" {
			planned = append(planned, recordResponse(rec))
		}
	}
	page, pageSize, totalPages := calculatePagination(result.Total, filter.Page, filter.Limit)
	return ProjectTasksResponse{
		Data: ProjectTasksData{
			ID:            result.Project.ID,
			Name:          result.Project.Name,
			Slug:          result.Project.Slug,
			Icon:          result.Project.Icon,
			Description:   result.Project.Description,
			IsPublic:      result.Project.IsPublic,
			WorkspaceID:   result.Project.WorkspaceID,
			Columns:       cols,
			ArchivedTasks: archived,
			PlannedTasks:  planned,
		},
		Pagination: PaginationResponse{
			Total:      result.Total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	}
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
	return &resp, nil
}

func (controller *Controller) delete(ctx context.Context, userID, taskID int64) (*TaskResponse, error) {
	record, err := controller.service.Delete(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	resp := recordResponse(*record)
	return &resp, nil
}

func (controller *Controller) move(ctx context.Context, userID, taskID int64, req *MoveTaskRequest) (*MoveTaskResponse, error) {
	if req == nil || req.DestinationProjectID <= 0 {
		return nil, errors.Invalid("destinationProjectId is required")
	}
	return controller.service.Move(ctx, userID, taskID, req)
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
	return &resp, nil
}

func (controller *Controller) updateDescription(ctx context.Context, userID, taskID int64, description string) (*TaskResponse, error) {
	record, err := controller.service.UpdateDescription(ctx, userID, taskID, description)
	if err != nil {
		return nil, err
	}
	resp := recordResponse(*record)
	return &resp, nil
}

func (controller *Controller) bulk(ctx context.Context, userID int64, req *BulkUpdateRequest) (*BulkUpdateResponse, error) {
	if req == nil || len(req.TaskIDs) == 0 {
		return nil, errors.Invalid("taskIds list cannot be empty")
	}
	return controller.service.Bulk(ctx, userID, req)
}
