package poolsync

import (
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/events"
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
	case existing != nil:
		return s.updateTask(tx, existing, evt)
	case action == "created":
		return s.createTask(tx, evt)
	default:
		return outcome{}, nil
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
		return s.unplaced(evt, "task has an unknown project or no name")
	}
	var twins []schemas.Task
	if err := tx.Where("project_id = ? AND lower(title) = lower(?) AND facile_id <> ? AND synced_at IS NULL",
		parent.ID, evt.Name, evt.facileID).Order("id").Limit(1).Find(&twins).Error; err != nil {
		return outcome{}, err
	}
	if len(twins) > 0 {
		return s.mergeTask(tx, &twins[0], evt)
	}
	columns, err := projectColumns(tx, parent.ID)
	if err != nil {
		return outcome{}, err
	}
	place, found := resolveStatus(columns, evt.Status, "")
	if !found {
		place = defaultStatus(columns)
	}
	record, err := s.buildTask(tx, parent, evt, place)
	if err != nil {
		return outcome{}, err
	}
	if err := tx.Create(&record).Error; err != nil {
		return outcome{}, err
	}
	return outcome{msgType: realtime.TypeTaskCreated, projectID: parent.ID, taskID: record.ID}, nil
}

// mergeTask links an unlinked local twin to the incoming task, keeping the
// smaller facile_id on both sides like mergeProject.
func (s *Syncer) mergeTask(tx *gorm.DB, twin *schemas.Task, evt inbound) (outcome, error) {
	if twin.FacileID != nil && evt.facileID >= *twin.FacileID {
		return outcome{}, nil
	}
	if err := tx.Model(&schemas.Task{}).Where("id = ?", twin.ID).
		Updates(map[string]any{"facile_id": evt.facileID, "updated_at": twin.UpdatedAt, "synced_at": time.Now()}).Error; err != nil {
		return outcome{}, err
	}
	twin.FacileID = &evt.facileID
	return s.updateTask(tx, twin, evt)
}

func nextSlot(tx *gorm.DB, projectID int64, place target) (int, int, error) {
	var number, position int
	if err := tx.Model(&schemas.Task{}).Where("project_id = ?", projectID).
		Select("COALESCE(MAX(number), 0)").Scan(&number).Error; err != nil {
		return 0, 0, err
	}
	slot := tx.Model(&schemas.Task{}).Where("project_id = ?", projectID)
	if place.columnID != nil {
		slot = slot.Where("column_id = ?", *place.columnID)
	} else {
		slot = slot.Where("status = ?", place.slug)
	}
	if err := slot.Select("COALESCE(MAX(position), 0)").Scan(&position).Error; err != nil {
		return 0, 0, err
	}
	return number + 1, position + 1, nil
}

func (s *Syncer) buildTask(tx *gorm.DB, parent *schemas.Project, evt inbound, place target) (schemas.Task, error) {
	number, position, err := nextSlot(tx, parent.ID, place)
	if err != nil {
		return schemas.Task{}, err
	}
	actorID, err := s.resolveActor(tx, parent.WorkspaceID, evt)
	if err != nil {
		return schemas.Task{}, err
	}
	facileID, now := evt.facileID, time.Now()
	record := schemas.Task{
		FacileRef:   schemas.FacileRef{FacileID: &facileID},
		ProjectID:   parent.ID,
		WorkspaceID: parent.WorkspaceID,
		ColumnID:    place.columnID,
		ActorID:     actorID,
		Number:      number,
		Position:    position,
		Title:       evt.Name,
		Status:      place.slug,
		Priority:    "low",
		CreatedAt:   evt.at,
		UpdatedAt:   evt.updated,
		SyncedAt:    &now,
	}
	fields := taskFields(evt)
	if description, ok := fields["description"].(string); ok {
		record.Description = description
	}
	if priority, ok := fields["priority"].(string); ok {
		record.Priority = priority
	}
	if due, ok := fields["due_date"].(*time.Time); ok {
		record.DueDate = due
	}
	return record, nil
}

// taskFields collects the description, priority and due date an event sets.
// A field the event does not carry is left out, so it never clears anything.
func taskFields(evt inbound) map[string]any {
	fields := map[string]any{}
	if evt.Description != nil {
		fields["description"] = *evt.Description
	}
	if evt.Priority != nil {
		if priority, ok := events.LocalPriority(*evt.Priority); ok {
			fields["priority"] = priority
		}
	}
	if due, ok := evt.dueDate(); ok && due != nil {
		fields["due_date"] = due
	} else if ok {
		fields["due_date"] = nil
	}
	return fields
}

func (s *Syncer) resolveActor(tx *gorm.DB, workspaceID int64, evt inbound) (*int64, error) {
	userID, err := s.resolveUser(tx, evt.ActorSubject, evt.ActorEmail)
	if err != nil || userID == 0 {
		return nil, err
	}
	var count int64
	err = tx.Model(&schemas.WorkspaceMember{}).Where("workspace_id = ? AND user_id = ?", workspaceID, userID).Count(&count).Error
	if err != nil || count == 0 {
		return nil, err
	}
	return &userID, nil
}
