package sharelink

import (
	documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"
)

func shareLinkCollectionRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/share-link",
			Summary:      "List share links for a project",
			Description:  "Get all share links configured for a project.",
			Auth:         "bearer",
			ResponseBody: `[{"id":int64,"projectId":int64,"token":string,"label":string,"isActive":bool}]`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Project not found"},
			},
		},
		{
			Method:       "POST",
			Path:         "/share-link",
			Summary:      "Create a share link",
			Description:  "Generate a new public read-only share link for a project.",
			Auth:         "bearer",
			RequestBody:  `{"projectId":int64,"label":string,"expiresAt":string}`,
			ResponseBody: `{"id":int64,"projectId":int64,"token":string,"label":string,"isActive":bool}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid project ID or expiry date"},
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
			},
		},
	}
}

func shareLinkItemRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "PUT",
			Path:         "/share-link/{id}",
			Summary:      "Update a share link",
			Description:  "Update the label, active state, or expiration of a share link.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "The share link ID"}},
			RequestBody:  `{"label":string,"expiresAt":string,"isActive":bool}`,
			ResponseBody: `{"id":int64,"projectId":int64,"token":string,"label":string,"isActive":bool}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Share link not found"},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/share-link/{id}",
			Summary:      "Delete a share link",
			Description:  "Delete an existing share link.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "The share link ID"}},
			ResponseBody: `{"id":int64,"projectId":int64,"token":string,"label":string,"isActive":bool}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
				{Status: 404, Code: "not_found", Description: "Share link not found"},
			},
		},
		{
			Method:       "GET",
			Path:         "/shared-project/{token}",
			Summary:      "View a shared project",
			Description:  "Public endpoint to view project tasks via an active share token.",
			Auth:         "none",
			PathParams:   []documentation.Field{{Name: "token", Type: "string", Description: "The share token"}},
			ResponseBody: `{"id":int64,"name":string,"columns":[],"tasks":[]}`,
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Share link is invalid, expired, or inactive"},
			},
		},
	}
}

// Documentation returns the API documentation metadata for the share link module.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "sharelink",
		Description: "Share links provide public read-only views for projects and tasks.",
		Routes:      append(shareLinkCollectionRoutes(), shareLinkItemRoutes()...),
	}
}
