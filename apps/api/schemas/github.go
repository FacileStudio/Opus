package schemas

import "time"

// ProjectGitHubRepo links an Opus project to a repository.
type ProjectGitHubRepo struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ProjectID     int64     `gorm:"column:project_id;not null;index"`
	RepoFullName  string    `gorm:"column:repo_full_name;not null;index"`
	DefaultBranch string    `gorm:"column:default_branch;not null;default:'main'"`
	CreatedAt     time.Time `gorm:"column:created_at;not null"`
}

// TableName returns the table name for ProjectGitHubRepo.
func (ProjectGitHubRepo) TableName() string {
	return "project_github_repos"
}

// TaskPullRequest records a pull request linked to a task.
type TaskPullRequest struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TaskID         int64     `gorm:"column:task_id;not null;index"`
	RepoFullName   string    `gorm:"column:repo_full_name;not null"`
	PRNumber       int       `gorm:"column:pr_number;not null"`
	Title          string    `gorm:"column:title;not null"`
	URL            string    `gorm:"column:url;not null"`
	State          string    `gorm:"column:state;not null"`
	HeadBranch     string    `gorm:"column:head_branch;not null;default:''"`
	BaseBranch     string    `gorm:"column:base_branch;not null;default:''"`
	AuthorGitHubID *string   `gorm:"column:author_github_id"`
	CreatedAt      time.Time `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null"`
}

// TableName returns the table name for TaskPullRequest.
func (TaskPullRequest) TableName() string {
	return "task_pull_requests"
}
