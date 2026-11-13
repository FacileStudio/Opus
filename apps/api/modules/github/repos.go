package github

import (
	"context"
	stderrors "errors"
	"slices"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/githubapp"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

const (
	appMissing = "GitHub is not set up on this Opus: the Facile Opus App needs OPUS_GITHUB_APP_CLIENT_ID and OPUS_GITHUB_APP_PRIVATE_KEY"
	idMissing  = "Link your GitHub account in your Facile account, then sign in to Opus again"
)

var adminRoles = []string{"owner", "admin"}

// Repos links projects to GitHub repositories and reads their issues
// through the Facile Opus App. A nil app turns every call but unlinking
// into a 412.
type Repos struct {
	db  *gorm.DB
	app *githubapp.Client
}

// NewRepos builds the linked-repository service.
func NewRepos(db *gorm.DB, app *githubapp.Client) *Repos {
	return &Repos{db: db, app: app}
}

// List returns the repositories linked to a project. Any member may read it.
func (r *Repos) List(ctx context.Context, userID, projectID int64) ([]RepoResponse, error) {
	if err := r.requireApp(); err != nil {
		return nil, err
	}
	if err := r.requireRole(ctx, userID, projectID); err != nil {
		return nil, err
	}
	rows, err := r.linked(ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]RepoResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toRepoResponse(row))
	}
	return out, nil
}

// Link connects a repository. The caller must own or administer the
// project's workspace, have a GitHub account on file, and hold write or
// admin on the repository.
func (r *Repos) Link(ctx context.Context, userID, projectID int64, name string) (*RepoResponse, error) {
	if err := r.requireApp(); err != nil {
		return nil, err
	}
	name = strings.Trim(strings.TrimSpace(name), "/")
	if !githubapp.ValidRepo(name) {
		return nil, errors.Invalid("Write the repository as owner/repo")
	}
	if err := r.requireRole(ctx, userID, projectID, adminRoles...); err != nil {
		return nil, err
	}
	if err := r.requireWrite(ctx, userID, name); err != nil {
		return nil, err
	}
	repo, err := r.app.Repository(ctx, name)
	if err != nil {
		return nil, appError(err, name)
	}
	row := schemas.ProjectGitHubRepo{ProjectID: projectID, RepoFullName: repo.FullName, DefaultBranch: repo.DefaultBranch, CreatedAt: time.Now()}
	if row.DefaultBranch == "" {
		row.DefaultBranch = "main"
	}
	result := r.db.WithContext(ctx).
		Where("project_id = ? AND lower(repo_full_name) = lower(?)", projectID, repo.FullName).
		FirstOrCreate(&row)
	if result.Error != nil {
		return nil, errors.Internal("failed to link the repository", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, errors.Conflict("This repository is already linked")
	}
	response := toRepoResponse(row)
	return &response, nil
}

// Unlink removes a linked repository. It needs workspace owner or admin
// rights, not the App, so a link can always be removed.
func (r *Repos) Unlink(ctx context.Context, userID, projectID, repoID int64) error {
	if err := r.requireRole(ctx, userID, projectID, adminRoles...); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", repoID, projectID).
		Delete(&schemas.ProjectGitHubRepo{})
	if result.Error != nil {
		return errors.Internal("failed to unlink the repository", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("Linked repository not found")
	}
	return nil
}

func (r *Repos) linked(ctx context.Context, projectID int64) ([]schemas.ProjectGitHubRepo, error) {
	var rows []schemas.ProjectGitHubRepo
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("repo_full_name").Find(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to list linked repositories", err)
	}
	return rows, nil
}

func (r *Repos) requireApp() error {
	if r.app == nil {
		return errors.Failed(appMissing)
	}
	return nil
}

func (r *Repos) requireRole(ctx context.Context, userID, projectID int64, roles ...string) error {
	var project schemas.Project
	err := r.db.WithContext(ctx).Select("id, workspace_id").First(&project, projectID).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NotFound("project not found")
	}
	if err != nil {
		return errors.Internal("failed to load project", err)
	}
	var role string
	err = r.db.WithContext(ctx).Model(&schemas.WorkspaceMember{}).
		Where("user_id = ? AND workspace_id = ?", userID, project.WorkspaceID).
		Select("role").Scan(&role).Error
	if err != nil {
		return errors.Internal("failed to check workspace access", err)
	}
	if role == "" {
		return errors.Forbidden("You don't have access to this workspace")
	}
	if len(roles) > 0 && !slices.Contains(roles, role) {
		return errors.Forbidden("Only workspace owners and admins can change linked repositories")
	}
	return nil
}

func (r *Repos) requireWrite(ctx context.Context, userID int64, name string) error {
	var githubID *string
	err := r.db.WithContext(ctx).Model(&schemas.User{}).Where("id = ?", userID).Select("github_id").Scan(&githubID).Error
	if err != nil {
		return errors.Internal("failed to load the GitHub account", err)
	}
	if githubID == nil || strings.TrimSpace(*githubID) == "" {
		return errors.Failed(idMissing)
	}
	level, err := r.app.Permission(ctx, name, *githubID)
	if err != nil {
		return appError(err, name)
	}
	if !githubapp.CanWrite(level) {
		return errors.Forbidden("You need write access to " + name + " on GitHub to link it")
	}
	return nil
}

func appError(err error, name string) error {
	switch {
	case stderrors.Is(err, githubapp.ErrNotInstalled):
		return errors.Failed("The Facile Opus App is not installed on " + name + ": install it on the GitHub account or organization that owns the repository, then link it again")
	case stderrors.Is(err, githubapp.ErrInvalidRepo):
		return errors.Invalid("Write the repository as owner/repo")
	case stderrors.Is(err, githubapp.ErrNotFound):
		return errors.NotFound(name + " was not found on GitHub")
	default:
		return errors.New("unavailable", "GitHub did not answer for "+name, err)
	}
}
