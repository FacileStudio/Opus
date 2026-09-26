package timeentry

import (
	"context"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// Controller validates and maps HTTP payloads for time entries.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) listForTask(ctx context.Context, userID, taskID int64) ([]TimeEntryResponse, error) {
	if taskID <= 0 {
		return nil, errors.Invalid("invalid task id")
	}
	records, err := controller.service.ListForTask(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	responses := make([]TimeEntryResponse, 0, len(records))
	for _, rec := range records {
		responses = append(responses, toResponse(rec))
	}
	return responses, nil
}

func (controller *Controller) get(ctx context.Context, userID, entryID int64) (*TimeEntryResponse, error) {
	if entryID <= 0 {
		return nil, errors.Invalid("invalid time entry id")
	}
	rec, err := controller.service.Get(ctx, userID, entryID)
	if err != nil {
		return nil, err
	}
	resp := toResponse(*rec)
	return &resp, nil
}

func (controller *Controller) create(ctx context.Context, userID int64, req *CreateTimeEntryRequest) (*TimeEntryResponse, error) {
	if req.TaskID <= 0 {
		return nil, errors.Invalid("taskId is required")
	}
	if req.StartTime.IsZero() {
		return nil, errors.Invalid("startTime is required")
	}
	if req.EndTime != nil && req.EndTime.Before(req.StartTime) {
		return nil, errors.Invalid("endTime cannot be earlier than startTime")
	}
	entry, err := controller.service.Create(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	resp := toModelResponse(entry)
	return &resp, nil
}

func (controller *Controller) update(ctx context.Context, userID, entryID int64, req *UpdateTimeEntryRequest) (*TimeEntryResponse, error) {
	if entryID <= 0 {
		return nil, errors.Invalid("invalid time entry id")
	}
	if req.StartTime != nil && req.EndTime != nil && req.EndTime.Before(*req.StartTime) {
		return nil, errors.Invalid("endTime cannot be earlier than startTime")
	}
	entry, err := controller.service.Update(ctx, userID, entryID, req)
	if err != nil {
		return nil, err
	}
	resp := toModelResponse(entry)
	return &resp, nil
}

func (controller *Controller) delete(ctx context.Context, userID, entryID int64) (*TimeEntryResponse, error) {
	if entryID <= 0 {
		return nil, errors.Invalid("invalid time entry id")
	}
	entry, err := controller.service.Delete(ctx, userID, entryID)
	if err != nil {
		return nil, err
	}
	resp := toModelResponse(entry)
	return &resp, nil
}

func toResponse(rec TimeEntryRecord) TimeEntryResponse {
	resp := toModelResponse(&rec.TimeEntry)
	if rec.TimeEntry.UserID != nil {
		resp.User = &TimeEntryUserResponse{
			ID:    *rec.TimeEntry.UserID,
			Name:  rec.UserName,
			Image: rec.UserImage,
		}
	}
	return resp
}

func toModelResponse(entry *schemas.TimeEntry) TimeEntryResponse {
	return TimeEntryResponse{
		ID:          entry.ID,
		TaskID:      entry.TaskID,
		UserID:      entry.UserID,
		Description: entry.Description,
		StartTime:   entry.StartTime,
		EndTime:     entry.EndTime,
		Duration:    entry.Duration,
		CreatedAt:   entry.CreatedAt,
	}
}
