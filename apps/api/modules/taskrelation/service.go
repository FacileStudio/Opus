package taskrelation

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// Service provides database operations for task relations.
type Service struct {
	db         *gorm.DB
	controller *Controller
}

// NewService wires the task relation service with its database handle.
func NewService(db *gorm.DB) *Service {
	service := &Service{db: db}
	service.controller = newController(service)
	return service
}

// RequireMember verifies that the caller belongs to the given workspace.
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

func (service *Service) findTask(ctx context.Context, taskID int64, notFoundMsg string) (*schemas.Task, error) {
	var task schemas.Task
	err := service.db.WithContext(ctx).First(&task, "id = ?", taskID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound(notFoundMsg)
		}
		return nil, errors.Internal("failed to query task", err)
	}
	return &task, nil
}

func (service *Service) fetchTaskSummaries(ctx context.Context, relations []schemas.TaskRelation) (map[int64]*TaskSummary, error) {
	idSet := make(map[int64]struct{})
	for _, rel := range relations {
		idSet[rel.SourceTaskID] = struct{}{}
		idSet[rel.TargetTaskID] = struct{}{}
	}
	ids := make([]int64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	type taskRow struct {
		ID        int64
		Title     string
		Status    string
		Priority  string
		Number    int
		ProjectID int64
		ActorID   *int64
		ActorName *string
	}
	var rows []taskRow
	err := service.db.WithContext(ctx).
		Table("tasks").
		Select("tasks.id, tasks.title, tasks.status, tasks.priority, tasks.number, tasks.project_id, tasks.actor_id, users.name as actor_name").
		Joins("LEFT JOIN users ON users.id = tasks.actor_id").
		Where("tasks.id IN ?", ids).
		Scan(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to query related tasks", err)
	}
	taskMap := make(map[int64]*TaskSummary, len(rows))
	for _, row := range rows {
		taskMap[row.ID] = &TaskSummary{
			ID:        row.ID,
			Title:     row.Title,
			Status:    row.Status,
			Priority:  row.Priority,
			Number:    row.Number,
			ProjectID: row.ProjectID,
			ActorID:   row.ActorID,
			ActorName: row.ActorName,
		}
	}
	return taskMap, nil
}

// ListForTask returns all relations for a task with associated task summaries.
func (service *Service) ListForTask(ctx context.Context, userID, taskID int64) ([]RelationRecord, error) {
	task, err := service.findTask(ctx, taskID, "Task not found")
	if err != nil {
		return nil, err
	}
	if err := service.RequireMember(ctx, userID, task.WorkspaceID); err != nil {
		return nil, err
	}
	var relations []schemas.TaskRelation
	err = service.db.WithContext(ctx).
		Where("source_task_id = ? OR target_task_id = ?", taskID, taskID).
		Find(&relations).Error
	if err != nil {
		return nil, errors.Internal("failed to query task relations", err)
	}
	if len(relations) == 0 {
		return []RelationRecord{}, nil
	}
	taskMap, err := service.fetchTaskSummaries(ctx, relations)
	if err != nil {
		return nil, err
	}
	records := make([]RelationRecord, 0, len(relations))
	for _, rel := range relations {
		records = append(records, RelationRecord{
			Relation:   rel,
			SourceTask: taskMap[rel.SourceTaskID],
			TargetTask: taskMap[rel.TargetTaskID],
		})
	}
	return records, nil
}

// CreateRelation stores a new task relation after validating task existence and access.
func (service *Service) CreateRelation(ctx context.Context, userID int64, req CreateTaskRelationRequest) (*schemas.TaskRelation, error) {
	sourceTask, err := service.findTask(ctx, req.SourceTaskID, "Source task not found")
	if err != nil {
		return nil, err
	}
	if err := service.RequireMember(ctx, userID, sourceTask.WorkspaceID); err != nil {
		return nil, err
	}
	targetTask, err := service.findTask(ctx, req.TargetTaskID, "Target task not found")
	if err != nil {
		return nil, err
	}
	if err := service.RequireMember(ctx, userID, targetTask.WorkspaceID); err != nil {
		return nil, err
	}
	var count int64
	err = service.db.WithContext(ctx).
		Model(&schemas.TaskRelation{}).
		Where(
			"relation_type = ? AND ((source_task_id = ? AND target_task_id = ?) OR (source_task_id = ? AND target_task_id = ?))",
			req.RelationType, req.SourceTaskID, req.TargetTaskID, req.TargetTaskID, req.SourceTaskID,
		).
		Count(&count).Error
	if err != nil {
		return nil, errors.Internal("failed to check existing relation", err)
	}
	if count > 0 {
		return nil, errors.Conflict("This relation already exists")
	}
	relation := schemas.TaskRelation{
		SourceTaskID: req.SourceTaskID,
		TargetTaskID: req.TargetTaskID,
		RelationType: req.RelationType,
		CreatedAt:    time.Now(),
	}
	if err := service.db.WithContext(ctx).Create(&relation).Error; err != nil {
		return nil, errors.Internal("failed to create task relation", err)
	}
	return &relation, nil
}

// DeleteRelation removes a task relation by its ID.
func (service *Service) DeleteRelation(ctx context.Context, userID, id int64) (*schemas.TaskRelation, error) {
	var relation schemas.TaskRelation
	err := service.db.WithContext(ctx).First(&relation, "id = ?", id).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("Task relation not found")
		}
		return nil, errors.Internal("failed to query task relation", err)
	}
	task, err := service.findTask(ctx, relation.SourceTaskID, "Task not found")
	if err != nil {
		return nil, err
	}
	if err := service.RequireMember(ctx, userID, task.WorkspaceID); err != nil {
		return nil, err
	}
	if err := service.db.WithContext(ctx).Delete(&relation).Error; err != nil {
		return nil, errors.Internal("failed to delete task relation", err)
	}
	return &relation, nil
}
