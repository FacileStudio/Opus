package github

import (
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
)

// WebhookPayload represents the GitHub webhook payload for push and pull request events.
type WebhookPayload struct {
	Action      string              `json:"action"`
	Number      int                 `json:"number"`
	PullRequest *PullRequestPayload `json:"pull_request"`
	Repository  *RepositoryPayload  `json:"repository"`
	Sender      *SenderPayload      `json:"sender"`
	Ref         string              `json:"ref"`
	Commits     []CommitPayload     `json:"commits"`
}

// PullRequestPayload holds the pull request details inside a webhook payload.
type PullRequestPayload struct {
	ID        int64          `json:"id"`
	Number    int            `json:"number"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	HTMLURL   string         `json:"html_url"`
	State     string         `json:"state"`
	Draft     bool           `json:"draft"`
	Merged    bool           `json:"merged"`
	Head      BranchRef      `json:"head"`
	Base      BranchRef      `json:"base"`
	User      *SenderPayload `json:"user"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// BranchRef contains the branch ref and commit SHA.
type BranchRef struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

// RepositoryPayload holds the repository data.
type RepositoryPayload struct {
	ID            int64  `json:"id"`
	FullName      string `json:"full_name"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
}

// SenderPayload contains the user that triggered the event.
type SenderPayload struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

// CommitPayload holds commit message details for push events.
type CommitPayload struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

// TaskRef is an extracted reference to an Opus task.
type TaskRef struct {
	ProjectSlug string
	TaskNumber  int
}

// PullRequestResponse represents a task pull request returned to clients.
type PullRequestResponse struct {
	ID             int64     `json:"id"`
	TaskID         int64     `json:"taskId"`
	RepoFullName   string    `json:"repoFullName"`
	PRNumber       int       `json:"prNumber"`
	Title          string    `json:"title"`
	URL            string    `json:"url"`
	State          string    `json:"state"`
	HeadBranch     string    `json:"headBranch"`
	BaseBranch     string    `json:"baseBranch"`
	AuthorGitHubID *string   `json:"authorGithubId,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func toPullRequestResponse(row schemas.TaskPullRequest) PullRequestResponse {
	return PullRequestResponse{
		ID:             row.ID,
		TaskID:         row.TaskID,
		RepoFullName:   row.RepoFullName,
		PRNumber:       row.PRNumber,
		Title:          row.Title,
		URL:            row.URL,
		State:          row.State,
		HeadBranch:     row.HeadBranch,
		BaseBranch:     row.BaseBranch,
		AuthorGitHubID: row.AuthorGitHubID,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}
