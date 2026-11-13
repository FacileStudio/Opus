package github

import (
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
)

// LinkRepoRequest is the body of POST /project/{projectId}/github-repos.
type LinkRepoRequest struct {
	Repo string `json:"repo"`
}

// RepoResponse is a repository linked to a project.
type RepoResponse struct {
	ID            int64     `json:"id"`
	ProjectID     int64     `json:"projectId"`
	FullName      string    `json:"fullName"`
	DefaultBranch string    `json:"defaultBranch"`
	URL           string    `json:"url"`
	CreatedAt     time.Time `json:"createdAt"`
}

// IssuesResponse lists issues across a project's linked repositories.
type IssuesResponse struct {
	Issues      []IssueResponse   `json:"issues"`
	Unavailable []UnavailableRepo `json:"unavailable"`
}

// UnavailableRepo is a linked repository GitHub refused, with the reason.
type UnavailableRepo struct {
	Repo   string `json:"repo"`
	Reason string `json:"reason"`
}

// IssueResponse is one GitHub issue.
type IssueResponse struct {
	Repo      string       `json:"repo"`
	Number    int          `json:"number"`
	Title     string       `json:"title"`
	URL       string       `json:"url"`
	State     string       `json:"state"`
	Comments  int          `json:"comments"`
	Labels    []IssueLabel `json:"labels"`
	Author    *IssueAuthor `json:"author"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

// IssueLabel is a label on an issue; Color is hex without the #.
type IssueLabel struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// IssueAuthor is the GitHub account that opened an issue.
type IssueAuthor struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatarUrl"`
	URL       string `json:"url"`
}

func toRepoResponse(row schemas.ProjectGitHubRepo) RepoResponse {
	return RepoResponse{
		ID:            row.ID,
		ProjectID:     row.ProjectID,
		FullName:      row.RepoFullName,
		DefaultBranch: row.DefaultBranch,
		URL:           "https://github.com/" + row.RepoFullName,
		CreatedAt:     row.CreatedAt,
	}
}
