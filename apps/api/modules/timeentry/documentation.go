package timeentry

import (
	documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"
)

func timeEntryCollectionRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/time-entry/task/{taskId}",
			Summary:      "List time entries for a task",
			Description:  "Get all logged time entries for a task.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "taskId", Type: "int64", Description: "Task ID"}},
			ResponseBody: `[{"id":int64,"taskId":int64,"duration":int}]`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
		{
			Method:       "POST",
			Path:         "/time-entry",
			Summary:      "Create a time entry",
			Description:  "Log work duration on a task.",
			Auth:         "bearer",
			RequestBody:  `{"taskId":int64,"startTime":string,"endTime":string,"description":string}`,
			ResponseBody: `{"id":int64,"taskId":int64,"duration":int}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid timestamps or task ID"},
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
			},
		},
	}
}

func timeEntryItemRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/time-entry/{id}",
			Summary:      "Get a time entry",
			Description:  "Fetch details of a single time entry.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Time entry ID"}},
			ResponseBody: `{"id":int64,"taskId":int64,"duration":int}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Time entry not found"},
			},
		},
		{
			Method:       "PUT",
			Path:         "/time-entry/{id}",
			Summary:      "Update a time entry",
			Description:  "Update the timestamps or description of a time entry.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Time entry ID"}},
			RequestBody:  `{"startTime":string,"endTime":string,"description":string}`,
			ResponseBody: `{"id":int64,"taskId":int64,"duration":int}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid timestamps"},
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Time entry not found"},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/time-entry/{id}",
			Summary:      "Delete a time entry",
			Description:  "Remove an existing time entry.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Time entry ID"}},
			ResponseBody: `{"id":int64,"taskId":int64}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Time entry not found"},
			},
		},
	}
}

// Documentation returns API documentation metadata for the time tracking module.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "timeentry",
		Description: "Time entries track work duration against tasks.",
		Routes:      append(timeEntryCollectionRoutes(), timeEntryItemRoutes()...),
	}
}
