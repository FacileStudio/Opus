// Package task implements task API documentation.
package task

import documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"

func taskQueryRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/task/tasks/{projectId}",
			Summary:      "List project tasks",
			Description:  "Returns project details, columns, and tasks filtered and ordered.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "projectId", Type: "int64", Description: "Project ID"}},
			ResponseBody: `{"data":{"id":1,"columns":[],"archivedTasks":[],"plannedTasks":[]},"pagination":{"total":0,"page":1,"pageSize":50,"totalPages":1}}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Project not found"},
			},
		},
		{
			Method:       "GET",
			Path:         "/task/me",
			Summary:      "List my tasks",
			Description:  "Returns tasks assigned to the caller in the workspace given by the workspaceId query parameter. Archived tasks and tasks in final columns are excluded unless includeDone=true.",
			Auth:         "bearer",
			ResponseBody: `[{"id":1,"projectId":1,"projectSlug":"app","projectName":"App","projectIcon":"Layout","number":1,"title":"Task","status":"to-do","priority":"low","subtaskTotal":0,"subtaskDone":0}]`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid workspace id"},
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
			},
		},
		{
			Method:       "GET",
			Path:         "/task/{id}",
			Summary:      "Get single task",
			Description:  "Returns task details by ID including author and labels.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Task ID"}},
			ResponseBody: `{"id":1,"projectId":1,"title":"Task","status":"to-do","priority":"low"}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
	}
}

func taskBulkRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "PATCH",
			Path:         "/task/bulk",
			Summary:      "Bulk update tasks",
			Description:  "Performs bulk operations on multiple tasks in the same workspace.",
			Auth:         "bearer",
			RequestBody:  `{"taskIds":[1,2],"operation":"updateStatus","value":"done"}`,
			ResponseBody: `{"success":true,"updatedCount":2}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid operation or cross-workspace tasks"},
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Tasks not found"},
			},
		},
	}
}

func taskLifecycleRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "POST",
			Path:         "/task/{projectId}",
			Summary:      "Create a task",
			Description:  "Creates a new task in the specified project.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "projectId", Type: "int64", Description: "Project ID"}},
			RequestBody:  `{"title":"New task","status":"to-do","priority":"low"}`,
			ResponseBody: `{"id":1,"projectId":1,"title":"New task","status":"to-do","priority":"low"}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid title, status, or priority"},
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Project not found"},
			},
		},
		{
			Method:       "PUT",
			Path:         "/task/move/{id}",
			Summary:      "Move task to project",
			Description:  "Moves a task to another project within the workspace.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Task ID"}},
			RequestBody:  `{"destinationProjectId":2,"destinationStatus":"in-progress"}`,
			ResponseBody: `{"task":{"id":1},"sourceProjectId":1,"destinationProjectId":2}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid destination or cross-workspace move"},
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Task or destination project not found"},
			},
		},
		{
			Method:       "PUT",
			Path:         "/task/{id}",
			Summary:      "Update full task",
			Description:  "Replaces all updatable fields of a task.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Task ID"}},
			RequestBody:  `{"title":"Updated","status":"in-progress","priority":"high","projectId":1,"position":1}`,
			ResponseBody: `{"id":1,"title":"Updated","status":"in-progress","priority":"high"}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid field values"},
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/task/{id}",
			Summary:      "Delete task",
			Description:  "Deletes a task and its relations.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Task ID"}},
			ResponseBody: `{"id":1,"title":"Deleted"}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
	}
}

func taskPropertyRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "PUT",
			Path:         "/task/status/{id}",
			Summary:      "Update task status",
			Description:  "Updates only the status and column of a task.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Task ID"}},
			RequestBody:  `{"status":"done"}`,
			ResponseBody: `{"id":1,"status":"done"}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid status"},
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
		{
			Method:       "PUT",
			Path:         "/task/priority/{id}",
			Summary:      "Update task priority",
			Description:  "Updates only the priority of a task.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Task ID"}},
			RequestBody:  `{"priority":"urgent"}`,
			ResponseBody: `{"id":1,"priority":"urgent"}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid priority"},
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
		{
			Method:       "PUT",
			Path:         "/task/actor/{id}",
			Summary:      "Update task actor",
			Description:  "Assigns or unassigns a task.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Task ID"}},
			RequestBody:  `{"actorId":2}`,
			ResponseBody: `{"id":1,"actorId":2}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
	}
}

func taskContentRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "PUT",
			Path:         "/task/due-date/{id}",
			Summary:      "Update task due date",
			Description:  "Updates only the due date of a task.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Task ID"}},
			RequestBody:  `{"dueDate":"2026-12-31T23:59:59Z"}`,
			ResponseBody: `{"id":1,"dueDate":"2026-12-31T23:59:59Z"}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
		{
			Method:       "PUT",
			Path:         "/task/title/{id}",
			Summary:      "Update task title",
			Description:  "Updates only the title of a task.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Task ID"}},
			RequestBody:  `{"title":"New title"}`,
			ResponseBody: `{"id":1,"title":"New title"}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid title"},
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
		{
			Method:       "PUT",
			Path:         "/task/description/{id}",
			Summary:      "Update task description",
			Description:  "Updates only the description of a task.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Task ID"}},
			RequestBody:  `{"description":"Updated description"}`,
			ResponseBody: `{"id":1,"description":"Updated description"}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
	}
}

// Documentation returns the task module API reference metadata.
func Documentation() documentation.Module {
	routes := append(taskQueryRoutes(), taskBulkRoutes()...)
	routes = append(routes, taskLifecycleRoutes()...)
	routes = append(routes, taskPropertyRoutes()...)
	routes = append(routes, taskContentRoutes()...)
	return documentation.Module{
		Name:        "task",
		Description: "Task management, status workflows, and batch operations.",
		Routes:      routes,
	}
}
