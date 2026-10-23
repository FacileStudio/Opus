package task

import (
	"context"
	"strconv"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/events"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

func (c *Controller) export(ctx context.Context, userID, projectID int64) (*ExportTasksResponse, error) {
	return c.service.ExportTasks(ctx, userID, projectID)
}

func (c *Controller) importTasks(ctx context.Context, userID, projectID int64, req *ImportTasksRequest) (*ImportTasksResponse, error) {
	if req == nil || len(req.Tasks) == 0 {
		return nil, errors.Invalid("tasks to import are required")
	}
	return c.service.ImportTasks(ctx, userID, projectID, req)
}

// ExportTasks retrieves all tasks in a project for export.
func (s *Service) ExportTasks(ctx context.Context, userID, projectID int64) (*ExportTasksResponse, error) {
	project, err := s.verifyProjectMembership(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	var tasks []schemas.Task
	err = s.db.WithContext(ctx).
		Where("project_id = ? AND archived_at IS NULL", projectID).
		Order("position asc").
		Find(&tasks).Error
	if err != nil {
		return nil, errors.Internal("failed to fetch tasks", err)
	}
	items := make([]ExportTaskItem, len(tasks))
	for i, t := range tasks {
		items[i] = ExportTaskItem{
			Title:       t.Title,
			Description: t.Description,
			Status:      t.Status,
			Priority:    t.Priority,
			DueDate:     t.DueDate,
			StartDate:   t.StartDate,
			UserID:      t.ActorID,
		}
	}
	return &ExportTasksResponse{
		Project: ExportProjectMeta{
			Name:        project.Name,
			Slug:        project.Slug,
			Description: project.Description,
			ExportedAt:  time.Now().UTC().Format(time.RFC3339),
		},
		Tasks: items,
	}, nil
}

func (s *Service) getNextTaskNumber(ctx context.Context, projectID int64) int {
	var maxNum int
	row := s.db.WithContext(ctx).Model(&schemas.Task{}).
		Where("project_id = ?", projectID).
		Select("COALESCE(MAX(number), 0)").
		Row()
	if err := row.Scan(&maxNum); err != nil {
		return 0
	}
	return maxNum
}

// ImportTasks batch inserts tasks into a project.
func (s *Service) ImportTasks(ctx context.Context, userID, projectID int64, req *ImportTasksRequest) (*ImportTasksResponse, error) {
	project, err := s.verifyProjectMembership(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	currentNum := s.getNextTaskNumber(ctx, projectID)
	successful := 0
	failed := 0
	now := time.Now()
	for _, item := range req.Tasks {
		currentNum++
		priority := item.Priority
		if priority == "" {
			priority = "low"
		}
		status := item.Status
		if status == "" {
			status = "to-do"
		}
		task := schemas.Task{
			ProjectID:   project.ID,
			WorkspaceID: project.WorkspaceID,
			ActorID:     item.UserID,
			Title:       item.Title,
			Status:      status,
			Description: item.Description,
			Priority:    priority,
			Number:      currentNum,
			Position:    currentNum,
			StartDate:   item.StartDate,
			DueDate:     item.DueDate,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := s.db.WithContext(ctx).Create(&task).Error; err != nil {
			failed++
			continue
		}
		successful++
		s.emit(ctx, userID, events.ChannelTaskCreated, task)
	}
	return &ImportTasksResponse{
		ImportedAt: now.UTC().Format(time.RFC3339),
		Results: ImportTasksSummary{
			Total:      len(req.Tasks),
			Successful: successful,
			Failed:     failed,
		},
	}, nil
}

// GetPublicProject loads public project tasks by project ID or slug.
func (s *Service) GetPublicProject(ctx context.Context, identifier string) (*ProjectTasksData, error) {
	var project schemas.Project
	query := s.db.WithContext(ctx)
	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil && id > 0 {
		query = query.Where("id = ?", id)
	} else {
		query = query.Where("slug = ?", identifier)
	}
	if err := query.First(&project).Error; err != nil {
		return nil, errors.NotFound("project not found")
	}
	if !project.IsPublic {
		return nil, errors.Forbidden("project is not public")
	}
	var columns []schemas.Column
	if err := s.db.WithContext(ctx).Where("project_id = ?", project.ID).Order("position asc").Find(&columns).Error; err != nil {
		return nil, errors.Internal("failed to load columns", err)
	}
	var tasks []schemas.Task
	if err := s.db.WithContext(ctx).Where("project_id = ?", project.ID).Order("position asc").Find(&tasks).Error; err != nil {
		return nil, errors.Internal("failed to fetch tasks", err)
	}
	records, err := s.loadTaskRecords(ctx, tasks)
	if err != nil {
		return nil, err
	}
	result := &ListTasksResult{
		Project: project,
		Columns: columns,
		Tasks:   records,
		Total:   int64(len(records)),
	}
	resp := toProjectTasksResponse(result, ListTasksFilter{})
	return &resp.Data, nil
}
