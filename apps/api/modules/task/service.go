// Package task implements task management, querying and bulk operations.
package task

import (
	"context"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// Service handles database operations and authorization for tasks.
type Service struct {
	db         *gorm.DB
	controller *Controller
}

// NewService constructs a new task service with its database handle.
func NewService(db *gorm.DB) *Service {
	service := &Service{db: db}
	service.controller = newController(service)
	return service
}

func (service *Service) verifyWorkspaceMembership(ctx context.Context, userID, workspaceID int64) error {
	var count int64
	err := service.db.WithContext(ctx).Model(&schemas.WorkspaceMember{}).
		Where("user_id = ? AND workspace_id = ?", userID, workspaceID).
		Count(&count).Error
	if err != nil {
		return errors.Internal("failed to verify workspace membership", err)
	}
	if count == 0 {
		return errors.Forbidden("You don't have access to this workspace")
	}
	return nil
}

func (service *Service) verifyProjectMembership(ctx context.Context, userID, projectID int64) (*schemas.Project, error) {
	var project schemas.Project
	if err := service.db.WithContext(ctx).First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("Project not found")
		}
		return nil, errors.Internal("failed to load project", err)
	}
	if err := service.verifyWorkspaceMembership(ctx, userID, project.WorkspaceID); err != nil {
		return nil, err
	}
	return &project, nil
}

func (service *Service) verifyTaskMembership(ctx context.Context, userID, taskID int64) (*schemas.Task, error) {
	var task schemas.Task
	if err := service.db.WithContext(ctx).First(&task, taskID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("Task not found")
		}
		return nil, errors.Internal("failed to load task", err)
	}
	if err := service.verifyWorkspaceMembership(ctx, userID, task.WorkspaceID); err != nil {
		return nil, err
	}
	return &task, nil
}

func assertValidPriority(priority string) error {
	switch priority {
	case "no-priority", "low", "medium", "high", "urgent":
		return nil
	default:
		return errors.Invalid("Invalid priority: " + priority)
	}
}

func (service *Service) getValidStatuses(ctx context.Context, projectID int64) ([]string, error) {
	var columns []schemas.Column
	err := service.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("position asc").
		Find(&columns).Error
	if err != nil {
		return nil, errors.Internal("failed to fetch columns", err)
	}
	statuses := make([]string, 0, len(columns)+2)
	for _, col := range columns {
		statuses = append(statuses, col.Slug)
	}
	statuses = append(statuses, "planned", "archived")
	return statuses, nil
}

func (service *Service) assertValidTaskStatus(ctx context.Context, projectID int64, status string) error {
	statuses, err := service.getValidStatuses(ctx, projectID)
	if err != nil {
		return err
	}
	for _, s := range statuses {
		if s == status {
			return nil
		}
	}
	return errors.Invalid("Invalid status: " + status)
}

func buildOrderClause(sortBy, sortOrder string) string {
	direction := "asc"
	if strings.ToLower(sortOrder) == "desc" {
		direction = "desc"
	}
	switch sortBy {
	case "createdAt":
		return "created_at " + direction
	case "priority":
		return "CASE WHEN priority = 'urgent' THEN 4 WHEN priority = 'high' THEN 3 WHEN priority = 'medium' THEN 2 WHEN priority = 'low' THEN 1 ELSE 0 END " + direction
	case "dueDate":
		return "due_date " + direction
	case "title":
		return "title " + direction
	case "number":
		return "number " + direction
	default:
		return "position " + direction + ", id " + direction
	}
}

func applyListFilters(query *gorm.DB, filter ListTasksFilter) *gorm.DB {
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Priority != "" {
		query = query.Where("priority = ?", filter.Priority)
	}
	if filter.ActorID != nil {
		query = query.Where("actor_id = ?", *filter.ActorID)
	}
	if filter.DueBefore != nil {
		query = query.Where("due_date <= ?", *filter.DueBefore)
	}
	if filter.DueAfter != nil {
		query = query.Where("due_date >= ?", *filter.DueAfter)
	}
	return query
}

func fetchTaskLabelsMap(ctx context.Context, db *gorm.DB, taskIDs []int64) (map[int64][]schemas.Label, error) {
	var labels []schemas.Label
	if err := db.WithContext(ctx).Where("task_id IN ?", taskIDs).Find(&labels).Error; err != nil {
		return nil, errors.Internal("failed to fetch task labels", err)
	}
	taskLabels := make(map[int64][]schemas.Label)
	for _, l := range labels {
		if l.TaskID != nil {
			taskLabels[*l.TaskID] = append(taskLabels[*l.TaskID], l)
		}
	}
	return taskLabels, nil
}

func fetchTaskActorsMap(ctx context.Context, db *gorm.DB, actorIDsMap map[int64]bool) (map[int64]*schemas.User, error) {
	actorUsers := make(map[int64]*schemas.User)
	if len(actorIDsMap) == 0 {
		return actorUsers, nil
	}
	actorIDs := make([]int64, 0, len(actorIDsMap))
	for id := range actorIDsMap {
		actorIDs = append(actorIDs, id)
	}
	var users []schemas.User
	if err := db.WithContext(ctx).Where("id IN ?", actorIDs).Find(&users).Error; err != nil {
		return nil, errors.Internal("failed to fetch task actors", err)
	}
	for i := range users {
		actorUsers[users[i].ID] = &users[i]
	}
	return actorUsers, nil
}

func (service *Service) loadTaskRecords(ctx context.Context, tasks []schemas.Task) ([]TaskRecord, error) {
	if len(tasks) == 0 {
		return []TaskRecord{}, nil
	}
	taskIDs := make([]int64, len(tasks))
	actorIDsMap := make(map[int64]bool)
	for i, t := range tasks {
		taskIDs[i] = t.ID
		if t.ActorID != nil {
			actorIDsMap[*t.ActorID] = true
		}
	}
	taskLabels, err := fetchTaskLabelsMap(ctx, service.db, taskIDs)
	if err != nil {
		return nil, err
	}
	actorUsers, err := fetchTaskActorsMap(ctx, service.db, actorIDsMap)
	if err != nil {
		return nil, err
	}
	records := make([]TaskRecord, len(tasks))
	for i, t := range tasks {
		var actor *schemas.User
		if t.ActorID != nil {
			actor = actorUsers[*t.ActorID]
		}
		records[i] = TaskRecord{
			Task:   t,
			Actor:  actor,
			Labels: taskLabels[t.ID],
		}
	}
	return records, nil
}

// Get loads a single task record including author and labels.
func (service *Service) Get(ctx context.Context, userID, taskID int64) (*TaskRecord, error) {
	task, err := service.verifyTaskMembership(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	records, err := service.loadTaskRecords(ctx, []schemas.Task{*task})
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.NotFound("Task not found")
	}
	return &records[0], nil
}

func applyPagination(query *gorm.DB, pagePtr, limitPtr *int) *gorm.DB {
	if pagePtr == nil && limitPtr == nil {
		return query
	}
	page := 1
	if pagePtr != nil && *pagePtr > 0 {
		page = *pagePtr
	}
	pageSize := 50
	if limitPtr != nil && *limitPtr > 0 {
		pageSize = *limitPtr
		if pageSize > 100 {
			pageSize = 100
		}
	}
	return query.Limit(pageSize).Offset((page - 1) * pageSize)
}

// List returns project metadata, workflow columns, and matching task records.
func (service *Service) List(ctx context.Context, userID, projectID int64, filter ListTasksFilter) (*ListTasksResult, error) {
	project, err := service.verifyProjectMembership(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	var columns []schemas.Column
	if err := service.db.WithContext(ctx).Where("project_id = ?", projectID).Order("position asc").Find(&columns).Error; err != nil {
		return nil, errors.Internal("failed to load columns", err)
	}
	query := service.db.WithContext(ctx).Model(&schemas.Task{}).Where("project_id = ?", projectID)
	query = applyListFilters(query, filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, errors.Internal("failed to count tasks", err)
	}
	query = query.Order(buildOrderClause(filter.SortBy, filter.SortOrder))
	query = applyPagination(query, filter.Page, filter.Limit)
	var tasks []schemas.Task
	if err := query.Find(&tasks).Error; err != nil {
		return nil, errors.Internal("failed to fetch tasks", err)
	}
	records, err := service.loadTaskRecords(ctx, tasks)
	if err != nil {
		return nil, err
	}
	return &ListTasksResult{
		Project: *project,
		Columns: columns,
		Tasks:   records,
		Total:   total,
	}, nil
}
