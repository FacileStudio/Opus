package poolsync

import (
	"log/slog"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/realtime"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
)

func (s *Syncer) updateTask(tx *gorm.DB, existing *schemas.Task, evt inbound) (outcome, error) {
	if evt.updated.Before(existing.UpdatedAt) {
		return outcome{}, nil
	}
	updates := taskFields(evt)
	if evt.Name != "" && evt.Name != existing.Title {
		updates["title"] = evt.Name
	}
	projectID, err := s.placeTask(tx, existing, evt, updates)
	if err != nil {
		return outcome{}, err
	}
	if projectID == 0 {
		return s.unplaced(evt, "task moves to an unknown project")
	}
	actorID, err := s.resolveActor(tx, existing.WorkspaceID, evt)
	if err != nil {
		return outcome{}, err
	}
	if actorID != nil && (existing.ActorID == nil || *existing.ActorID != *actorID) {
		updates["actor_id"] = *actorID
	}
	changed := len(updates) > 0
	updates["updated_at"] = evt.updated
	updates["synced_at"] = time.Now()
	if err := tx.Model(&schemas.Task{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
		return outcome{}, err
	}
	if !changed {
		return outcome{}, nil
	}
	return outcome{msgType: realtime.TypeTaskUpdated, projectID: projectID, taskID: existing.ID}, nil
}

// placeTask sets the status, and the project when the event moves the task,
// in updates. It answers the task's project id, or 0 when the event names a
// project Opus does not hold.
func (s *Syncer) placeTask(tx *gorm.DB, existing *schemas.Task, evt inbound, updates map[string]any) (int64, error) {
	columns, err := projectColumns(tx, existing.ProjectID)
	if err != nil {
		return 0, err
	}
	destination, err := s.moveTarget(tx, existing, evt)
	if err != nil || destination == 0 {
		return 0, err
	}
	if destination != existing.ProjectID {
		wire := evt.Status
		if wire == "" {
			wire = classify(columns, existing.Status)
		}
		return destination, moveTask(tx, destination, wire, updates)
	}
	if place, found := resolveStatus(columns, evt.Status, existing.Status); found && place.slug != existing.Status {
		updates["status"] = place.slug
		updates["column_id"] = place.columnID
	}
	return existing.ProjectID, nil
}

// moveTarget is the project the event puts the task in. Opus only moves tasks
// within a workspace, so a move across workspaces keeps the current project.
func (s *Syncer) moveTarget(tx *gorm.DB, existing *schemas.Task, evt inbound) (int64, error) {
	if evt.ProjectFacileID == "" {
		return existing.ProjectID, nil
	}
	parent, err := findProject(tx, evt.ProjectFacileID)
	if err != nil || parent == nil {
		return 0, err
	}
	if parent.WorkspaceID != existing.WorkspaceID {
		s.logger.Warn("pool sync: task move across workspaces ignored",
			slog.String("task", evt.facileID), slog.String("project", evt.ProjectFacileID))
		return existing.ProjectID, nil
	}
	return parent.ID, nil
}

func moveTask(tx *gorm.DB, projectID int64, wire string, updates map[string]any) error {
	columns, err := projectColumns(tx, projectID)
	if err != nil {
		return err
	}
	place, found := resolveStatus(columns, wire, "")
	if !found {
		place = defaultStatus(columns)
	}
	number, position, err := nextSlot(tx, projectID, place)
	if err != nil {
		return err
	}
	updates["project_id"] = projectID
	updates["number"] = number
	updates["position"] = position
	updates["status"] = place.slug
	updates["column_id"] = place.columnID
	return nil
}
