package workflowrule

import (
	documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"
)

func workflowRuleRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/workflow-rule/{projectId}",
			Summary:      "List workflow rules for a project",
			Description:  "Get all automation rules configured for a project.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "projectId", Type: "int64", Description: "Project ID"}},
			ResponseBody: `[{"id":int64,"projectId":int64,"integrationType":string,"eventType":string,"columnId":int64}]`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Project not found"},
			},
		},
		{
			Method:       "PUT",
			Path:         "/workflow-rule/{projectId}",
			Summary:      "Upsert a workflow rule",
			Description:  "Create or update an automation rule for a project event.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "projectId", Type: "int64", Description: "Project ID"}},
			RequestBody:  `{"integrationType":string,"eventType":string,"columnId":int64}`,
			ResponseBody: `{"id":int64,"projectId":int64,"integrationType":string,"eventType":string,"columnId":int64}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid column or parameters"},
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/workflow-rule/{id}",
			Summary:      "Delete a workflow rule",
			Description:  "Remove an existing automation rule.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Workflow rule ID"}},
			ResponseBody: `{"id":int64,"projectId":int64}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Workflow rule not found"},
			},
		},
	}
}

// Documentation returns API documentation metadata for the workflow rule module.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "workflowrule",
		Description: "Workflow rules automate column moves and assignments based on integration events.",
		Routes:      workflowRuleRoutes(),
	}
}
