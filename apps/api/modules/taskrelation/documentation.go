package taskrelation

import documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"

func taskRelationCollectionRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/task-relation/{taskId}",
			Summary:      "List relations for a task",
			Description:  "Get all relations for a specific task including source and target task summaries.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "taskId", Type: "int64", Description: "The task ID"}},
			ResponseBody: `[{"id":int64,"sourceTaskId":int64,"targetTaskId":int64,"relationType":string,"createdAt":string,"sourceTask":{"id":int64,"title":string,"status":string,"priority":string,"number":int,"projectId":int64,"userId":int64,"actorName":string},"targetTask":{"id":int64,"title":string,"status":string,"priority":string,"number":int,"projectId":int64,"userId":int64,"actorName":string}}]`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid task ID"},
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
		{
			Method:       "POST",
			Path:         "/task-relation",
			Summary:      "Create a task relation",
			Description:  "Create a relation between two tasks (subtask, blocks, or related).",
			Auth:         "bearer",
			RequestBody:  `{"sourceTaskId":int64,"targetTaskId":int64,"relationType":string}`,
			ResponseBody: `{"id":int64,"sourceTaskId":int64,"targetTaskId":int64,"relationType":string,"createdAt":string}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Self-relation or invalid relation type"},
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Source or target task not found"},
				{Status: 409, Code: "already_exists", Description: "Relation already exists"},
			},
		},
	}
}

func taskRelationItemRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "DELETE",
			Path:         "/task-relation/{id}",
			Summary:      "Delete a task relation",
			Description:  "Delete a task relation by its ID.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "The relation ID"}},
			ResponseBody: `{"id":int64,"sourceTaskId":int64,"targetTaskId":int64,"relationType":string,"createdAt":string}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid relation ID"},
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Task relation not found"},
			},
		},
	}
}

// Documentation returns the API documentation metadata for the task relation module.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "taskrelation",
		Description: "Task relations define dependencies between tasks (subtask, blocks, related).",
		Routes:      append(taskRelationCollectionRoutes(), taskRelationItemRoutes()...),
	}
}
