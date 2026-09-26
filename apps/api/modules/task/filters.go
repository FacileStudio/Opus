// Package task implements query filtering and parsing for task listings.
package task

import (
	"net/http"
	"strconv"
	"time"

	"github.com/FacileStudio/tronc/errors"
)

func parseDateString(s string) (*time.Time, error) {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return &t, nil
		}
	}
	return nil, errors.Invalid("invalid date format: " + s)
}

func parseListFilter(r *http.Request) (ListTasksFilter, error) {
	q := r.URL.Query()
	filter := ListTasksFilter{
		Status:    q.Get("status"),
		Priority:  q.Get("priority"),
		SortBy:    q.Get("sortBy"),
		SortOrder: q.Get("sortOrder"),
	}
	if actorStr := q.Get("actorId"); actorStr != "" {
		if id, err := strconv.ParseInt(actorStr, 10, 64); err == nil {
			filter.ActorID = &id
		}
	}
	if pageStr := q.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			filter.Page = &page
		}
	}
	if limitStr := q.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = &limit
		}
	}
	if s := q.Get("dueBefore"); s != "" {
		if t, err := parseDateString(s); err == nil {
			filter.DueBefore = t
		}
	}
	if s := q.Get("dueAfter"); s != "" {
		if t, err := parseDateString(s); err == nil {
			filter.DueAfter = t
		}
	}
	return filter, nil
}
