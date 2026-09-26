package comment

import (
	documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"
)

func commentCollectionRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/comment/{taskId}",
			Summary:      "List comments for a task",
			Description:  "Get all comments for a specific task ordered by creation date.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "taskId", Type: "int64", Description: "The task ID"}},
			ResponseBody: `[{"id":int64,"taskId":int64,"userId":int64,"content":string,"createdAt":string,"updatedAt":string,"user":{"name":string,"image":string}}]`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
		{
			Method:       "POST",
			Path:         "/comment/{taskId}",
			Summary:      "Create a comment on a task",
			Description:  "Create a new discussion comment on a task.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "taskId", Type: "int64", Description: "The task ID"}},
			RequestBody:  `{"content":string}`,
			ResponseBody: `{"id":int64,"taskId":int64,"userId":int64,"content":string,"createdAt":string,"updatedAt":string}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Content is required"},
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
	}
}

func commentItemRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "PUT",
			Path:         "/comment/{id}",
			Summary:      "Update a comment",
			Description:  "Update an existing comment (author only).",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "The comment ID"}},
			RequestBody:  `{"content":string}`,
			ResponseBody: `{"id":int64,"taskId":int64,"userId":int64,"content":string,"createdAt":string,"updatedAt":string}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Content is required"},
				{Status: 403, Code: "permission_denied", Description: "Only the author can edit this comment"},
				{Status: 404, Code: "not_found", Description: "Comment not found"},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/comment/{id}",
			Summary:      "Delete a comment",
			Description:  "Delete an existing comment (author only).",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "The comment ID"}},
			ResponseBody: `{"id":int64,"taskId":int64,"userId":int64,"content":string,"createdAt":string,"updatedAt":string}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Only the author can delete this comment"},
				{Status: 404, Code: "not_found", Description: "Comment not found"},
			},
		},
	}
}

// Documentation returns the API documentation metadata for the comment module.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "comment",
		Description: "Comments are discussion entries on tasks, scoped to the task workspace.",
		Routes:      append(commentCollectionRoutes(), commentItemRoutes()...),
	}
}
