package timeentry

import (
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/Opus/apps/api/schemas"
)

type TimeEntryUserResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

// TimeEntryResponse is the JSON wire representation of a time entry.
type TimeEntryResponse struct {
	ID          int64                  `json:"id"`
	TaskID      int64                  `json:"taskId"`
	UserID      *int64                 `json:"userId,omitempty"`
	Description string                 `json:"description"`
	StartTime   time.Time              `json:"startTime"`
	EndTime     *time.Time             `json:"endTime,omitempty"`
	Duration    int                    `json:"duration"`
	CreatedAt   time.Time              `json:"createdAt"`
	Source      string                 `json:"source"`
	User        *TimeEntryUserResponse `json:"user,omitempty"`
}

// CreateTimeEntryRequest carries the payload for logging time on a task.
type CreateTimeEntryRequest struct {
	TaskID      wireid.ID  `json:"taskId"`
	StartTime   time.Time  `json:"startTime"`
	EndTime     *time.Time `json:"endTime,omitempty"`
	Description string     `json:"description,omitempty"`
}

// UpdateTimeEntryRequest carries modifications for an existing time entry.
type UpdateTimeEntryRequest struct {
	StartTime   *time.Time `json:"startTime,omitempty"`
	EndTime     *time.Time `json:"endTime,omitempty"`
	Description *string    `json:"description,omitempty"`
}

// TimeEntryRecord bundles the database model with user metadata.
type TimeEntryRecord struct {
	TimeEntry schemas.TimeEntry
	UserName  string
	UserImage string
}
