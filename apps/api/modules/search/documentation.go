package search

import (
	documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"
)

func searchRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/search",
			Summary:      "Search across resources",
			Description:  "Multi-domain search querying tasks, projects, workspaces, and comments.",
			Auth:         "bearer",
			ResponseBody: `{"results":[],"totalCount":0,"searchQuery":string}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Query parameter is required"},
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
			},
		},
	}
}

// Documentation returns API documentation metadata for the search module.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "search",
		Description: "Global full-text search across accessible workspaces and projects.",
		Routes:      searchRoutes(),
	}
}
