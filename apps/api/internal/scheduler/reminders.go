package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type reminderWindow struct {
	reminderType     string
	notificationType string
	start            time.Time
	end              time.Time
}

type reminderTask struct {
	ID      int64
	Title   string
	ActorID *int64
	DueDate *time.Time
}

type reminderEventData struct {
	TaskTitle    string  `json:"taskTitle"`
	ReminderType string  `json:"reminderType"`
	DueDate      *string `json:"dueDate"`
}

// RunDueReminders checks task due dates across active windows and creates notifications.
func (s *Scheduler) RunDueReminders(ctx context.Context, now time.Time) error {
	if now.IsZero() {
		now = time.Now()
	}
	windows := s.buildWindows(now)
	var errs []error
	for _, window := range windows {
		if err := s.processWindow(ctx, window, now); err != nil {
			s.logger.Error("failed to process reminder window",
				slog.String("reminder_type", window.reminderType),
				slog.Any("error", err),
			)
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s *Scheduler) buildWindows(now time.Time) []reminderWindow {
	return []reminderWindow{
		{
			reminderType:     "one_day_before",
			notificationType: "due_date_reminder",
			start:            now.Add(23*time.Hour + 50*time.Minute),
			end:              now.Add(24*time.Hour + 10*time.Minute),
		},
		{
			reminderType:     "one_hour_before",
			notificationType: "due_date_reminder",
			start:            now.Add(50 * time.Minute),
			end:              now.Add(70 * time.Minute),
		},
		{
			reminderType:     "overdue",
			notificationType: "task_overdue",
			start:            now.Add(-10 * time.Minute),
			end:              now,
		},
	}
}

func (s *Scheduler) processWindow(ctx context.Context, window reminderWindow, now time.Time) error {
	tasks, err := s.findTasksNeedingReminder(ctx, window)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if err := s.processReminder(ctx, task, window, now); err != nil {
			s.logger.Error("failed to process task reminder",
				slog.Int64("task_id", task.ID),
				slog.String("reminder_type", window.reminderType),
				slog.Any("error", err),
			)
		}
	}
	return nil
}

func (s *Scheduler) findTasksNeedingReminder(ctx context.Context, window reminderWindow) ([]reminderTask, error) {
	var tasks []reminderTask
	err := s.db.WithContext(ctx).
		Table("tasks").
		Select("tasks.id, tasks.title, tasks.actor_id, tasks.due_date").
		Joins("LEFT JOIN columns ON columns.id = tasks.column_id").
		Joins("LEFT JOIN task_reminder_sents ON task_reminder_sents.task_id = tasks.id AND task_reminder_sents.reminder_type = ?", window.reminderType).
		Where("tasks.actor_id IS NOT NULL").
		Where("tasks.due_date IS NOT NULL").
		Where("tasks.due_date BETWEEN ? AND ?", window.start, window.end).
		Where("task_reminder_sents.id IS NULL").
		Where("columns.is_final IS NULL OR columns.is_final = ?", false).
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *Scheduler) processReminder(ctx context.Context, task reminderTask, window reminderWindow, now time.Time) error {
	if task.ActorID == nil {
		return nil
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sent := schemas.TaskReminderSent{
			TaskID:       task.ID,
			ReminderType: window.reminderType,
			SentAt:       now,
		}
		res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&sent)
		if res.Error != nil || res.RowsAffected == 0 {
			return res.Error
		}
		return s.insertReminderNotification(tx, task, window, now)
	})
}

func (s *Scheduler) insertReminderNotification(tx *gorm.DB, task reminderTask, window reminderWindow, now time.Time) error {
	resID := strconv.FormatInt(task.ID, 10)
	resType := "task"
	eventDataJSON, err := json.Marshal(reminderEventData{
		TaskTitle:    task.Title,
		ReminderType: window.reminderType,
		DueDate:      formatDueDate(task.DueDate),
	})
	if err != nil {
		return err
	}
	notification := schemas.Notification{
		UserID:       *task.ActorID,
		Title:        task.Title,
		Type:         window.notificationType,
		EventData:    string(eventDataJSON),
		ResourceID:   &resID,
		ResourceType: &resType,
		CreatedAt:    now,
	}
	return tx.Create(&notification).Error
}

func formatDueDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}
