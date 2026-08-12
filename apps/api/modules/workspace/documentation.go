package workspace

import documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "workspace",
	Description: "Workspaces are the tenant boundary: a team's projects, columns and tasks live under one, and a user's membership in one is what the workspace-scoped access checks gate on.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/workspace",
			Summary:      "List the caller's workspaces",
			Description:  "Every workspace the authenticated user is a member of, ordered by name.",
			Auth:         "bearer",
			ResponseBody: `[{"id":int64,"name":string,"slug":string,"logo":string,"description":string,"createdAt":string}]`,
		},
		{
			Method:       "GET",
			Path:         "/workspace/{workspaceId}/members",
			Summary:      "List a workspace's members",
			Description:  "Member rows joined to their user profiles. The caller must be a member of the workspace.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}},
			ResponseBody: `[{"id":int64,"name":string,"email":string,"image":string,"role":string}]`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The caller is not a member of this workspace"},
			},
		},
		{
			Method:      "POST",
			Path:        "/workspace/{workspaceId}/logo",
			Summary:     "Upload a workspace logo",
			Description: "Multipart form with a 'logo' file field. Allowed types: png, jpeg, webp, svg, gif. Maximum 2MB. Stored under UPLOAD_DIR/workspaces/{id} and served at GET /workspace/{workspaceId}/logo.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}},
			RequestBody: "multipart/form-data: logo=<image file>",
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The caller is not a member of this workspace"},
				{Status: 400, Code: "invalid_argument", Description: "Missing logo field or disallowed image type"},
				{Status: 413, Code: "resource_exhausted", Description: "File larger than 2MB"},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/workspace/{workspaceId}/logo",
			Summary:     "Delete a workspace logo",
			Description: "Removes the stored files and clears the logo column.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}},
		},
		{
			Method:      "GET",
			Path:        "/workspace/{workspaceId}/logo",
			Summary:     "Serve a workspace logo",
			Description: "The stored image bytes, cache-busted by the caller's ?v= stamp. Public, matching the legacy API.",
			Auth:        "public",
			PathParams:  []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "The workspace's id"}},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "The workspace has no logo stored"},
			},
		},
	},
}
