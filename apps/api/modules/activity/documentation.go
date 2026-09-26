package activity

import (
	documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"
)

func activityCollectionRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/activity/{taskId}",
			Summary:      "Get all activities for a task",
			Description:  "Get all activities and comments for a specific task ordered by creation time.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "taskId", Type: "int64", Description: "The task ID"}},
			ResponseBody: `[{"id":int64,"taskId":int64,"userId":int64,"type":string,"content":string,"eventData":object,"createdAt":string,"updatedAt":string}]`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid task ID"},
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
		{
			Method:       "POST",
			Path:         "/activity/create",
			Summary:      "Create an activity",
			Description:  "Create a new system-generated or user activity event on a task.",
			Auth:         "bearer",
			RequestBody:  `{"taskId":int64,"userId":int64,"message":string,"content":string,"type":string,"eventData":object}`,
			ResponseBody: `{"id":int64,"taskId":int64,"userId":int64,"type":string,"content":string,"eventData":object,"createdAt":string,"updatedAt":string}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Task ID and type are required"},
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
	}
}

func activityCommentRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "POST",
			Path:         "/activity/comment",
			Summary:      "Create a comment activity",
			Description:  "Create a new comment activity entry on a task.",
			Auth:         "bearer",
			RequestBody:  `{"taskId":int64,"comment":string}`,
			ResponseBody: `{"id":int64,"taskId":int64,"userId":int64,"type":string,"content":string,"eventData":object,"createdAt":string,"updatedAt":string}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Comment text is required"},
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
		{
			Method:       "PUT",
			Path:         "/activity/comment",
			Summary:      "Update a comment activity",
			Description:  "Update an existing comment activity (author only).",
			Auth:         "bearer",
			RequestBody:  `{"activityId":int64,"comment":string}`,
			ResponseBody: `{"id":int64,"taskId":int64,"userId":int64,"type":string,"content":string,"eventData":object,"createdAt":string,"updatedAt":string}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Comment text is required"},
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Comment not found or you are not the author"},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/activity/comment",
			Summary:      "Delete a comment activity",
			Description:  "Delete an existing comment activity (author only).",
			Auth:         "bearer",
			RequestBody:  `{"activityId":int64}`,
			ResponseBody: `{"id":int64,"taskId":int64,"userId":int64,"type":string,"content":string,"eventData":object,"createdAt":string,"updatedAt":string}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Activity ID is required"},
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Comment not found or you are not the author"},
			},
		},
	}
}

// Documentation returns the API documentation metadata for the activity module.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "activity",
		Description: "Activity records state changes, comments, and audit events on tasks.",
		Routes:      append([]documentation.Route{}, append(activityCollectionRoutes(), activityCommentRoutes()...)...),
	}
}
