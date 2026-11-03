package timeentry

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// Service coordinates time entry business logic and storage.
type Service struct {
	db         *gorm.DB
	controller *Controller
}

// NewService instantiates a new time entry service.
func NewService(db *gorm.DB) *Service {
	service := &Service{db: db}
	service.controller = newController(service)
	return service
}

// RequireMember verifies caller belongs to the given workspace.
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

func (service *Service) verifyTaskAccess(ctx context.Context, userID, taskID int64) error {
	var task schemas.Task
	err := service.db.WithContext(ctx).
		Select("id, workspace_id").
		First(&task, "id = ?", taskID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("Task not found")
		}
		return errors.Internal("failed to query task", err)
	}
	return service.RequireMember(ctx, userID, task.WorkspaceID)
}

// requireNative refuses to change a time entry mirrored from Sablier, where
// it belongs.
func requireNative(entry *schemas.TimeEntry) error {
	if entry.Source == schemas.TimeEntrySourceSablier {
		return errors.Conflict("This time entry comes from Sablier and can only be changed there")
	}
	return nil
}

func (service *Service) verifyTimeEntryAccess(ctx context.Context, userID, entryID int64) (*schemas.TimeEntry, error) {
	var entry schemas.TimeEntry
	err := service.db.WithContext(ctx).First(&entry, "id = ?", entryID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("Time entry not found")
		}
		return nil, errors.Internal("failed to query time entry", err)
	}
	if err := service.verifyTaskAccess(ctx, userID, entry.TaskID); err != nil {
		return nil, err
	}
	return &entry, nil
}

type timeEntryRow struct {
	ID          int64
	TaskID      int64
	UserID      *int64
	Description string
	StartTime   time.Time
	EndTime     *time.Time
	Duration    int
	CreatedAt   time.Time
	Source      string
	UserName    string
	UserImage   string
}

// ListForTask queries all time entries for a given task.
func (service *Service) ListForTask(ctx context.Context, userID, taskID int64) ([]TimeEntryRecord, error) {
	if err := service.verifyTaskAccess(ctx, userID, taskID); err != nil {
		return nil, err
	}
	var rows []timeEntryRow
	err := service.db.WithContext(ctx).
		Table("time_entries").
		Select("time_entries.*, users.name as user_name, users.image as user_image").
		Joins("LEFT JOIN users ON users.id = time_entries.user_id").
		Where("time_entries.task_id = ?", taskID).
		Order("time_entries.start_time asc").
		Scan(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to list time entries", err)
	}
	records := make([]TimeEntryRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, toRecord(row))
	}
	return records, nil
}

// Get returns a single time entry record.
func (service *Service) Get(ctx context.Context, userID, entryID int64) (*TimeEntryRecord, error) {
	entry, err := service.verifyTimeEntryAccess(ctx, userID, entryID)
	if err != nil {
		return nil, err
	}
	var row timeEntryRow
	service.db.WithContext(ctx).
		Table("time_entries").
		Select("time_entries.*, users.name as user_name, users.image as user_image").
		Joins("LEFT JOIN users ON users.id = time_entries.user_id").
		Where("time_entries.id = ?", entry.ID).
		Scan(&row)
	rec := toRecord(row)
	return &rec, nil
}

func calculateDuration(start time.Time, end *time.Time) int {
	if end == nil {
		return 0
	}
	diff := int(end.Sub(start).Seconds())
	if diff < 0 {
		return 0
	}
	return diff
}

// Create records a new time entry against a task.
func (service *Service) Create(ctx context.Context, userID int64, req *CreateTimeEntryRequest) (*schemas.TimeEntry, error) {
	if err := service.verifyTaskAccess(ctx, userID, int64(req.TaskID)); err != nil {
		return nil, err
	}
	duration := calculateDuration(req.StartTime, req.EndTime)
	entry := schemas.TimeEntry{
		TaskID:      int64(req.TaskID),
		UserID:      &userID,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Duration:    duration,
		CreatedAt:   time.Now(),
	}
	if err := service.db.WithContext(ctx).Create(&entry).Error; err != nil {
		return nil, errors.Internal("failed to create time entry", err)
	}
	return &entry, nil
}

// Update edits timestamps and description of an existing time entry.
func (service *Service) Update(ctx context.Context, userID, entryID int64, req *UpdateTimeEntryRequest) (*schemas.TimeEntry, error) {
	entry, err := service.verifyTimeEntryAccess(ctx, userID, entryID)
	if err != nil {
		return nil, err
	}
	if err := requireNative(entry); err != nil {
		return nil, err
	}
	if req.StartTime != nil {
		entry.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		entry.EndTime = req.EndTime
	}
	if req.Description != nil {
		entry.Description = *req.Description
	}
	entry.Duration = calculateDuration(entry.StartTime, entry.EndTime)
	if err := service.db.WithContext(ctx).Save(entry).Error; err != nil {
		return nil, errors.Internal("failed to update time entry", err)
	}
	return entry, nil
}

// Delete removes a time entry.
func (service *Service) Delete(ctx context.Context, userID, entryID int64) (*schemas.TimeEntry, error) {
	entry, err := service.verifyTimeEntryAccess(ctx, userID, entryID)
	if err != nil {
		return nil, err
	}
	if err := requireNative(entry); err != nil {
		return nil, err
	}
	if err := service.db.WithContext(ctx).Delete(entry).Error; err != nil {
		return nil, errors.Internal("failed to delete time entry", err)
	}
	return entry, nil
}

func toRecord(row timeEntryRow) TimeEntryRecord {
	return TimeEntryRecord{
		TimeEntry: schemas.TimeEntry{
			ID:          row.ID,
			TaskID:      row.TaskID,
			UserID:      row.UserID,
			Description: row.Description,
			StartTime:   row.StartTime,
			EndTime:     row.EndTime,
			Duration:    row.Duration,
			CreatedAt:   row.CreatedAt,
			Source:      row.Source,
		},
		UserName:  row.UserName,
		UserImage: row.UserImage,
	}
}
