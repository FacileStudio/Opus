package poolsync

import (
	"log/slog"

	"github.com/FacileStudio/Opus/apps/api/internal/realtime"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func findTask(tx *gorm.DB, facileID string) (*schemas.Task, error) {
	var found []schemas.Task
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("facile_id = ?", facileID).Limit(1).Find(&found).Error; err != nil || len(found) == 0 {
		return nil, err
	}
	return &found[0], nil
}

func projectColumns(tx *gorm.DB, projectID int64) ([]schemas.Column, error) {
	var columns []schemas.Column
	err := tx.Where("project_id = ?", projectID).Order("position asc, id asc").Find(&columns).Error
	return columns, err
}

func (s *Syncer) applyTask(tx *gorm.DB, action string, evt inbound) (outcome, error) {
	existing, err := findTask(tx, evt.facileID)
	if err != nil {
		return outcome{}, err
	}
	switch {
	case action == "deleted":
		return deleteTask(tx, existing)
	case existing == nil && action == "created":
		return s.createTask(tx, evt)
	case existing == nil:
		return outcome{}, nil
	default:
		return updateTask(tx, existing, evt)
	}
}

func deleteTask(tx *gorm.DB, existing *schemas.Task) (outcome, error) {
	if existing == nil {
		return outcome{}, nil
	}
	if err := tx.Where("id = ?", existing.ID).Delete(&schemas.Task{}).Error; err != nil {
		return outcome{}, err
	}
	return outcome{msgType: realtime.TypeTaskDeleted, projectID: existing.ProjectID, taskID: existing.ID}, nil
}

func (s *Syncer) createTask(tx *gorm.DB, evt inbound) (outcome, error) {
	parent, err := findProject(tx, evt.ProjectFacileID)
	if err != nil {
		return outcome{}, err
	}
	if parent == nil || evt.Name == "" {
		s.logger.Warn("pool sync: task dropped, unknown project or no name",
			slog.String("task", evt.facileID), slog.String("project", evt.ProjectFacileID))
		return outcome{skipped: true}, nil
	}
	columns, err := projectColumns(tx, parent.ID)
	if err != nil {
		return outcome{}, err
	}
	place, found := resolveStatus(columns, evt.Status, "")
	if !found {
		place = defaultStatus(columns)
	}
	record, err := buildTask(tx, parent, evt, place)
	if err != nil {
		return outcome{}, err
	}
	if err := tx.Create(&record).Error; err != nil {
		return outcome{}, err
	}
	return outcome{msgType: realtime.TypeTaskCreated, projectID: parent.ID, taskID: record.ID}, nil
}

func buildTask(tx *gorm.DB, parent *schemas.Project, evt inbound, place target) (schemas.Task, error) {
	var number, position int
	if err := tx.Model(&schemas.Task{}).Where("project_id = ?", parent.ID).
		Select("COALESCE(MAX(number), 0)").Scan(&number).Error; err != nil {
		return schemas.Task{}, err
	}
	slot := tx.Model(&schemas.Task{}).Where("project_id = ?", parent.ID)
	if place.columnID != nil {
		slot = slot.Where("column_id = ?", *place.columnID)
	} else {
		slot = slot.Where("status = ?", place.slug)
	}
	if err := slot.Select("COALESCE(MAX(position), 0)").Scan(&position).Error; err != nil {
		return schemas.Task{}, err
	}
	actorID, err := resolveActor(tx, parent.WorkspaceID, evt.ActorEmail)
	if err != nil {
		return schemas.Task{}, err
	}
	facileID := evt.facileID
	now := evt.at
	return schemas.Task{
		FacileRef:   schemas.FacileRef{FacileID: &facileID},
		ProjectID:   parent.ID,
		WorkspaceID: parent.WorkspaceID,
		ColumnID:    place.columnID,
		ActorID:     actorID,
		Number:      number + 1,
		Position:    position + 1,
		Title:       evt.Name,
		Status:      place.slug,
		Priority:    "low",
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func updateTask(tx *gorm.DB, existing *schemas.Task, evt inbound) (outcome, error) {
	if evt.at.Before(existing.UpdatedAt) {
		return outcome{}, nil
	}
	updates := map[string]any{}
	if evt.Name != "" && evt.Name != existing.Title {
		updates["title"] = evt.Name
	}
	columns, err := projectColumns(tx, existing.ProjectID)
	if err != nil {
		return outcome{}, err
	}
	if place, found := resolveStatus(columns, evt.Status, existing.Status); found && place.slug != existing.Status {
		updates["status"] = place.slug
		updates["column_id"] = place.columnID
	}
	actorID, err := resolveActor(tx, existing.WorkspaceID, evt.ActorEmail)
	if err != nil {
		return outcome{}, err
	}
	if actorID != nil && (existing.ActorID == nil || *existing.ActorID != *actorID) {
		updates["actor_id"] = *actorID
	}
	if len(updates) == 0 {
		return outcome{}, nil
	}
	updates["updated_at"] = evt.at
	if err := tx.Model(&schemas.Task{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
		return outcome{}, err
	}
	return outcome{msgType: realtime.TypeTaskUpdated, projectID: existing.ProjectID, taskID: existing.ID}, nil
}

func resolveActor(tx *gorm.DB, workspaceID int64, email string) (*int64, error) {
	if email == "" {
		return nil, nil
	}
	var ids []int64
	err := tx.Model(&schemas.User{}).
		Where("lower(email) = lower(?)", email).
		Where("id IN (?)", tx.Model(&schemas.WorkspaceMember{}).Select("user_id").Where("workspace_id = ?", workspaceID)).
		Limit(1).Pluck("id", &ids).Error
	if err != nil || len(ids) == 0 {
		return nil, err
	}
	return &ids[0], nil
}
