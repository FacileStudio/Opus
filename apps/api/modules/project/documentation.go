// Package project manages project workspaces, columns, and workflows.
package project

import documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"

func projectCollectionRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/project",
			Summary:      "List projects in a workspace",
			Description:  "Get all projects in a workspace with their aggregate task statistics.",
			Auth:         "bearer",
			ResponseBody: `[{"id":int64,"workspaceId":int64,"name":string,"slug":string,"icon":string,"description":string,"isPublic":bool,"archivedAt":string,"createdAt":string,"statistics":{...}}]`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Workspace ID is missing or invalid"},
				{Status: 403, Code: "permission_denied", Description: "Caller is not a member of the workspace"},
			},
		},
		{
			Method:       "POST",
			Path:         "/project",
			Summary:      "Create a project",
			Description:  "Creates a project in a workspace and scaffolds default workflow columns.",
			Auth:         "bearer",
			RequestBody:  `{"name":string,"workspaceId":int64,"icon":string?,"slug":string?}`,
			ResponseBody: `{"id":int64,"workspaceId":int64,"name":string,"slug":string,"icon":string,"description":string,"isPublic":bool,"archivedAt":string,"createdAt":string}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Missing required fields"},
				{Status: 403, Code: "permission_denied", Description: "Caller is not a member of the workspace"},
			},
		},
	}
}

func projectItemRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/project/{id}",
			Summary:      "Get project details",
			Description:  "Get a single project with its task statistics.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Project ID"}},
			ResponseBody: `{"id":int64,"workspaceId":int64,"name":string,"slug":string,"icon":string,"description":string,"isPublic":bool,"archivedAt":string,"createdAt":string,"statistics":{...}}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller is not a member of the workspace"},
				{Status: 404, Code: "not_found", Description: "Project not found"},
			},
		},
		{
			Method:       "PUT",
			Path:         "/project/{id}",
			Summary:      "Update a project",
			Description:  "Updates fields of an existing project.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Project ID"}},
			RequestBody:  `{"name":string?,"icon":string?,"slug":string?,"description":string?,"isPublic":bool?}`,
			ResponseBody: `{"id":int64,"workspaceId":int64,"name":string,"slug":string,"icon":string,"description":string,"isPublic":bool,"archivedAt":string,"createdAt":string}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid update fields"},
				{Status: 403, Code: "permission_denied", Description: "Caller is not a member of the workspace"},
				{Status: 404, Code: "not_found", Description: "Project not found"},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/project/{id}",
			Summary:      "Delete a project",
			Description:  "Deletes a project along with its columns and tasks.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Project ID"}},
			ResponseBody: `{"id":int64,"workspaceId":int64,"name":string,"slug":string,"icon":string,"description":string,"isPublic":bool,"archivedAt":string,"createdAt":string}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller is not a member of the workspace"},
				{Status: 404, Code: "not_found", Description: "Project not found"},
			},
		},
	}
}

func projectActionRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "PUT",
			Path:         "/project/{id}/archive",
			Summary:      "Archive a project",
			Description:  "Marks a project as archived.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Project ID"}},
			ResponseBody: `{"id":int64,"workspaceId":int64,"name":string,"slug":string,"icon":string,"description":string,"isPublic":bool,"archivedAt":string,"createdAt":string}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller is not a member of the workspace"},
				{Status: 404, Code: "not_found", Description: "Project not found"},
			},
		},
		{
			Method:       "PUT",
			Path:         "/project/{id}/unarchive",
			Summary:      "Unarchive a project",
			Description:  "Restores an archived project.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Project ID"}},
			ResponseBody: `{"id":int64,"workspaceId":int64,"name":string,"slug":string,"icon":string,"description":string,"isPublic":bool,"archivedAt":string,"createdAt":string}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller is not a member of the workspace"},
				{Status: 404, Code: "not_found", Description: "Project not found"},
			},
		},
	}
}

// Documentation returns the project module API reference metadata.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "project",
		Description: "Projects organize tasks, columns, and workflows within a workspace.",
		Routes:      append(projectCollectionRoutes(), append(projectItemRoutes(), projectActionRoutes()...)...),
	}
}
