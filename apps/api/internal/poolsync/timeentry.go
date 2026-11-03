package poolsync

import (
	"log/slog"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func findTimeEntry(tx *gorm.DB, facileID string) (*schemas.TimeEntry, error) {
	var found []schemas.TimeEntry
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("facile_id = ?", facileID).Limit(1).Find(&found).Error; err != nil || len(found) == 0 {
		return nil, err
	}
	return &found[0], nil
}

// applyTimeEntry mirrors one Sablier time entry. Created and updated both
// upsert; an entry no longer on a task, or deleted, loses its mirror.
func (s *Syncer) applyTimeEntry(tx *gorm.DB, action string, evt inbound) (outcome, error) {
	existing, err := findTimeEntry(tx, evt.facileID)
	if err != nil {
		return outcome{}, err
	}
	if existing != nil && existing.Source != schemas.TimeEntrySourceSablier {
		return outcome{}, nil
	}
	if existing != nil && action != "deleted" && evt.updated.Before(existing.UpdatedAt) {
		return outcome{}, nil
	}
	if action == "deleted" || evt.TaskFacileID == "" {
		return outcome{}, tx.Where("facile_id = ? AND source = ?", evt.facileID, schemas.TimeEntrySourceSablier).Delete(&schemas.TimeEntry{}).Error
	}
	task, err := findTask(tx, evt.TaskFacileID)
	if err != nil {
		return outcome{}, err
	}
	if task == nil {
		return s.unplaced(evt, "time entry on an unknown task")
	}
	started, valid := startOf(evt)
	if !valid {
		s.logger.Warn("pool sync: time entry without a valid started_at dropped", slog.String("facile_id", evt.facileID))
		return outcome{}, nil
	}
	userID, err := s.resolveUser(tx, evt.UserSubject, evt.UserEmail)
	if err != nil {
		return outcome{}, err
	}
	record := mirrorEntry(evt, task.ID, userID, started)
	if existing == nil {
		return outcome{}, tx.Create(&record).Error
	}
	return outcome{}, tx.Model(&schemas.TimeEntry{}).Where("id = ?", existing.ID).Updates(map[string]any{
		"task_id": record.TaskID, "user_id": record.UserID, "description": record.Description,
		"start_time": record.StartTime, "end_time": record.EndTime, "duration": record.Duration,
		"updated_at": record.UpdatedAt,
	}).Error
}

func startOf(evt inbound) (time.Time, bool) {
	started, err := time.Parse(time.RFC3339, evt.StartedAt)
	return started, err == nil
}

// mirrorEntry builds the local row. A running entry has no end and a zero
// duration, like a running Opus timer; a stopped one counts its paused time out.
func mirrorEntry(evt inbound, taskID, userID int64, started time.Time) schemas.TimeEntry {
	facileID := evt.facileID
	record := schemas.TimeEntry{
		TaskID:    taskID,
		StartTime: started,
		CreatedAt: time.Now(),
		UpdatedAt: evt.updated,
		FacileID:  &facileID,
		Source:    schemas.TimeEntrySourceSablier,
	}
	if userID != 0 {
		record.UserID = &userID
	}
	if evt.Description != nil {
		record.Description = *evt.Description
	}
	if evt.StoppedAt == nil {
		return record
	}
	if stopped, err := time.Parse(time.RFC3339, *evt.StoppedAt); err == nil {
		record.EndTime = &stopped
		worked := stopped.Sub(started) - time.Duration(evt.PausedDurationMS)*time.Millisecond
		record.Duration = int(max(worked, 0).Seconds())
	}
	return record
}
