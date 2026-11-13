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
			{
				Method:      "GET",
				Path:        "/project/{projectId}/github-repos",
				Summary:     "List linked GitHub repositories",
				Description: "Returns the repositories linked to a project. Any workspace member. 412 when the Opus GitHub App is not configured.",
				Auth:        "bearer",
			},
			{
				Method:      "POST",
				Path:        "/project/{projectId}/github-repos",
				Summary:     "Link a GitHub repository",
				Description: "Body {\"repo\": \"owner/repo\"}. Needs workspace owner or admin, a GitHub account linked in Registre, and write or admin on the repository. 412 when the App or the GitHub account is missing, or the App is not installed on the repository.",
				Auth:        "bearer",
			},
			{
				Method:      "DELETE",
				Path:        "/project/{projectId}/github-repos/{repoId}",
				Summary:     "Unlink a GitHub repository",
				Description: "Removes a linked repository. Needs workspace owner or admin.",
				Auth:        "bearer",
			},
			{
				Method:      "GET",
				Path:        "/project/{projectId}/github-issues",
				Summary:     "List GitHub issues",
				Description: "Issues of every linked repository, pull requests excluded, at most 100 per repository, cached 60 seconds. Query state=open (default) or closed. Any workspace member. Repositories GitHub refuses are listed under unavailable with a reason.",
				Auth:        "bearer",
			},
		},
	}
}
