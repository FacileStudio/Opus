package task

import (
	"math"

	"github.com/FacileStudio/Opus/apps/api/schemas"
)

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
	if task.FacileID != nil {
		resp.FacileID = *task.FacileID
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
