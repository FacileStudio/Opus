package github

import documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"

// Documentation returns OpenAPI documentation for the GitHub integration module.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "github",
		Description: "GitHub repository integration and webhook handling.",
		Routes: []documentation.Route{
			{
				Method:      "POST",
				Path:        "/integrations/github/webhook",
				Summary:     "Receive GitHub webhook",
				Description: "Handles incoming GitHub pull request and push webhook events.",
				Auth:        "none",
			},
			{
				Method:      "GET",
				Path:        "/project/{projectId}/task/{taskId}/pull-requests",
				Summary:     "List task pull requests",
				Description: "Returns all pull requests linked to a given task.",
				Auth:        "bearer",
			},
		},
	}
}
