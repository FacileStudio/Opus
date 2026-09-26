package label

import documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"

func collectionRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/label/workspace/{workspaceId}",
			Summary:      "Get all labels for a workspace",
			Description:  "List all labels created within a specific workspace.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace id"}},
			ResponseBody: `[{"id":int64,"name":string,"color":string,"createdAt":string,"taskId":int64?,"workspaceId":int64}]`,
		},
		{
			Method:       "GET",
			Path:         "/label/task/{taskId}",
			Summary:      "Get all labels for a task",
			Description:  "List all labels assigned to a specific task.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "taskId", Type: "int64", Description: "The task id"}},
			ResponseBody: `[{"id":int64,"name":string,"color":string,"createdAt":string,"taskId":int64,"workspaceId":int64}]`,
		},
		{
			Method:       "POST",
			Path:         "/label",
			Summary:      "Create a new label",
			Description:  "Create a label attached to a workspace or assigned directly to a task.",
			Auth:         "bearer",
			RequestBody:  `{"name":string,"color":string,"workspaceId":int64,"taskId":int64?}`,
			ResponseBody: `{"id":int64,"name":string,"color":string,"createdAt":string,"taskId":int64?,"workspaceId":int64}`,
		},
	}
}

func itemRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/label/{id}",
			Summary:      "Get a label by ID",
			Description:  "Retrieve a specific label by its identifier.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "The label id"}},
			ResponseBody: `{"id":int64,"name":string,"color":string,"createdAt":string,"taskId":int64?,"workspaceId":int64}`,
		},
		{
			Method:       "PUT",
			Path:         "/label/{id}",
			Summary:      "Update a label",
			Description:  "Modify name and color of an existing label.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "The label id"}},
			RequestBody:  `{"name":string,"color":string}`,
			ResponseBody: `{"id":int64,"name":string,"color":string,"createdAt":string,"taskId":int64?,"workspaceId":int64}`,
		},
		{
			Method:       "DELETE",
			Path:         "/label/{id}",
			Summary:      "Delete a label",
			Description:  "Delete an existing label associated with a task.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "The label id"}},
			ResponseBody: `{"id":int64,"name":string,"color":string,"createdAt":string,"taskId":int64?,"workspaceId":int64}`,
		},
	}
}

func taskBindingRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "PUT",
			Path:         "/label/{id}/task",
			Summary:      "Assign label to task",
			Description:  "Attach an existing label to a specific task in the same workspace.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "The label id"}},
			RequestBody:  `{"taskId":int64}`,
			ResponseBody: `{"id":int64,"name":string,"color":string,"createdAt":string,"taskId":int64,"workspaceId":int64}`,
		},
		{
			Method:       "DELETE",
			Path:         "/label/{id}/task",
			Summary:      "Detach label from task",
			Description:  "Detach a label from its current task.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "The label id"}},
			ResponseBody: `{"id":int64,"name":string,"color":string,"createdAt":string,"taskId":int64?,"workspaceId":int64}`,
		},
	}
}

// Documentation returns the label module API reference metadata.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "label",
		Description: "Labels organize and categorize tasks across workspaces.",
		Routes:      append(collectionRoutes(), append(itemRoutes(), taskBindingRoutes()...)...),
	}
}
