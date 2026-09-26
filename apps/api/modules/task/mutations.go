// Package task implements task mutation procedures.
package task

import (
	"context"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

func findColumnForStatus(ctx context.Context, db *gorm.DB, projectID int64, status string) (*schemas.Column, error) {
	var column schemas.Column
	err := db.WithContext(ctx).
		Where("project_id = ? AND slug = ?", projectID, status).
		First(&column).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Internal("failed to find column", err)
	}
	return &column, nil
}

func calculateNextTaskNumber(ctx context.Context, db *gorm.DB, projectID int64) (int, error) {
	var maxNumber *int
	err := db.WithContext(ctx).Model(&schemas.Task{}).
		Where("project_id = ?", projectID).
		Select("COALESCE(MAX(number), 0)").
		Scan(&maxNumber).Error
	if err != nil {
		return 0, errors.Internal("failed to calculate next task number", err)
	}
	if maxNumber != nil {
		return *maxNumber + 1, nil
	}
	return 1, nil
}

func calculateNextPosition(ctx context.Context, db *gorm.DB, projectID int64, columnID *int64, status string) (int, error) {
	query := db.WithContext(ctx).Model(&schemas.Task{}).Where("project_id = ?", projectID)
	if columnID != nil {
		query = query.Where("column_id = ?", *columnID)
	} else {
		query = query.Where("status = ?", status)
	}
	var maxPos *int
	err := query.Select("COALESCE(MAX(position), 0)").Scan(&maxPos).Error
	if err != nil {
		return 0, errors.Internal("failed to calculate next position", err)
	}
	if maxPos != nil {
		return *maxPos + 1, nil
	}
	return 1, nil
}

// Create inserts a new task in the specified project.
func (service *Service) Create(ctx context.Context, userID, projectID int64, req *CreateTaskRequest) (*TaskRecord, error) {
	project, err := service.verifyProjectMembership(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = "to-do"
	}
	if err := service.assertValidTaskStatus(ctx, projectID, status); err != nil {
		return nil, err
	}
	priority := req.Priority
	if priority == "" {
		priority = "low"
	}
	if err := assertValidPriority(priority); err != nil {
		return nil, err
	}
	column, err := findColumnForStatus(ctx, service.db, projectID, status)
	if err != nil {
		return nil, err
	}
	var columnID *int64
	if column != nil {
		columnID = &column.ID
	}
	nextNumber, err := calculateNextTaskNumber(ctx, service.db, projectID)
	if err != nil {
		return nil, err
	}
	nextPos, err := calculateNextPosition(ctx, service.db, projectID, columnID, status)
	if err != nil {
		return nil, err
	}
	actorID := req.ActorID
	if actorID == nil {
		actorID = req.UserID
	}
	now := time.Now()
	task := schemas.Task{
		ProjectID:   projectID,
		WorkspaceID: project.WorkspaceID,
		ColumnID:    columnID,
		ActorID:     actorID,
		Number:      nextNumber,
		Position:    nextPos,
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		Priority:    priority,
		StartDate:   req.StartDate,
		DueDate:     req.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := service.db.WithContext(ctx).Create(&task).Error; err != nil {
		return nil, errors.Internal("failed to create task", err)
	}
	return service.Get(ctx, userID, task.ID)
}

// Update replaces all updatable fields of a task.
func (service *Service) Update(ctx context.Context, userID, taskID int64, req *UpdateTaskRequest) (*TaskRecord, error) {
	task, err := service.verifyTaskMembership(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	if err := service.assertValidTaskStatus(ctx, task.ProjectID, req.Status); err != nil {
		return nil, err
	}
	if err := assertValidPriority(req.Priority); err != nil {
		return nil, err
	}
	column, err := findColumnForStatus(ctx, service.db, task.ProjectID, req.Status)
	if err != nil {
		return nil, err
	}
	var columnID *int64
	if column != nil {
		columnID = &column.ID
	}
	actorID := req.ActorID
	if actorID == nil {
		actorID = req.UserID
	}
	updates := map[string]interface{}{
		"title":       req.Title,
		"description": req.Description,
		"status":      req.Status,
		"priority":    req.Priority,
		"column_id":   columnID,
		"position":    req.Position,
		"start_date":  req.StartDate,
		"due_date":    req.DueDate,
		"actor_id":    actorID,
		"updated_at":  time.Now(),
	}
	if err := service.db.WithContext(ctx).Model(&schemas.Task{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
		return nil, errors.Internal("failed to update task", err)
	}
	return service.Get(ctx, userID, taskID)
}

// Delete removes a task and cascading dependencies.
func (service *Service) Delete(ctx context.Context, userID, taskID int64) (*TaskRecord, error) {
	record, err := service.Get(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	err = service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("source_task_id = ? OR target_task_id = ?", taskID, taskID).Delete(&schemas.TaskRelation{}).Error; err != nil {
			return errors.Internal("failed to delete task relations", err)
		}
		if err := tx.Where("task_id = ?", taskID).Delete(&schemas.Label{}).Error; err != nil {
			return errors.Internal("failed to delete task labels", err)
		}
		if err := tx.Where("task_id = ?", taskID).Delete(&schemas.Comment{}).Error; err != nil {
			return errors.Internal("failed to delete task comments", err)
		}
		if err := tx.Where("id = ?", taskID).Delete(&schemas.Task{}).Error; err != nil {
			return errors.Internal("failed to delete task", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return record, nil
}

func resolveDestinationColumn(cols []schemas.Column, currentStatus, requestedStatus string) (*schemas.Column, error) {
	if requestedStatus != "" {
		for i := range cols {
			if cols[i].Slug == requestedStatus {
				return &cols[i], nil
			}
		}
		return nil, errors.Invalid("Selected status is not valid for the destination project")
	}
	for i := range cols {
		if cols[i].Slug == currentStatus {
			return &cols[i], nil
		}
	}
	return &cols[0], nil
}

// Move transfers a task to another project within the same workspace.
func (service *Service) Move(ctx context.Context, userID, taskID int64, req *MoveTaskRequest) (*MoveTaskResponse, error) {
	task, err := service.verifyTaskMembership(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	if task.ProjectID == req.DestinationProjectID {
		return nil, errors.Invalid("Task is already in that project")
	}
	var destProject schemas.Project
	if err := service.db.WithContext(ctx).First(&destProject, req.DestinationProjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("Destination project not found")
		}
		return nil, errors.Internal("failed to load destination project", err)
	}
	if task.WorkspaceID != destProject.WorkspaceID {
		return nil, errors.Invalid("Tasks can only be moved within the same workspace")
	}
	var destColumns []schemas.Column
	if err := service.db.WithContext(ctx).Where("project_id = ?", req.DestinationProjectID).Order("position asc").Find(&destColumns).Error; err != nil {
		return nil, errors.Internal("failed to load destination columns", err)
	}
	if len(destColumns) == 0 {
		return nil, errors.Invalid("Destination project does not have a workflow")
	}
	resolvedColumn, err := resolveDestinationColumn(destColumns, task.Status, req.DestinationStatus)
	if err != nil {
		return nil, err
	}
	nextNumber, err := calculateNextTaskNumber(ctx, service.db, req.DestinationProjectID)
	if err != nil {
		return nil, err
	}
	nextPos, err := calculateNextPosition(ctx, service.db, req.DestinationProjectID, &resolvedColumn.ID, resolvedColumn.Slug)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"project_id": req.DestinationProjectID,
		"status":     resolvedColumn.Slug,
		"column_id":  resolvedColumn.ID,
		"number":     nextNumber,
		"position":   nextPos,
		"updated_at": time.Now(),
	}
	if err := service.db.WithContext(ctx).Model(&schemas.Task{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
		return nil, errors.Internal("failed to move task", err)
	}
	updatedRecord, err := service.Get(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	return &MoveTaskResponse{
		Task:                 toResponse(updatedRecord.Task, updatedRecord.Actor, toLabelResponses(updatedRecord.Labels)),
		SourceProjectID:      task.ProjectID,
		DestinationProjectID: req.DestinationProjectID,
	}, nil
}
