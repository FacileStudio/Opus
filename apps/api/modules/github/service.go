package github

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
)

// Service coordinates GitHub repository integration and webhook processing.
type Service struct {
	db     *gorm.DB
	secret string
}

// NewService builds a GitHub integration service.
func NewService(db *gorm.DB, secret string) *Service {
	return &Service{db: db, secret: secret}
}

// ProcessWebhook processes incoming GitHub webhook payloads.
func (s *Service) ProcessWebhook(ctx context.Context, payload *WebhookPayload) error {
	if payload == nil {
		return errors.New("empty payload")
	}
	if payload.PullRequest != nil {
		return s.handlePullRequest(ctx, payload)
	}
	if len(payload.Commits) > 0 || payload.Ref != "" {
		return s.handlePush(ctx, payload)
	}
	return nil
}

// ListPullRequestsByTask lists pull requests linked to a given task.
func (s *Service) ListPullRequestsByTask(ctx context.Context, taskID int64) ([]schemas.TaskPullRequest, error) {
	var rows []schemas.TaskPullRequest
	err := s.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Order("updated_at DESC").
		Find(&rows).Error
	return rows, err
}

func (s *Service) handlePullRequest(ctx context.Context, payload *WebhookPayload) error {
	pr := payload.PullRequest
	repoFullName := ""
	defaultBranch := "main"
	if payload.Repository != nil {
		repoFullName = payload.Repository.FullName
		if payload.Repository.DefaultBranch != "" {
			defaultBranch = payload.Repository.DefaultBranch
		}
	}
	refs := extractTaskRefs(pr.Head.Ref, pr.Title, pr.Body)
	for _, ref := range refs {
		if err := s.processPullRequestRef(ctx, ref, repoFullName, defaultBranch, pr); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) processPullRequestRef(ctx context.Context, ref TaskRef, repoFullName, defaultBranch string, pr *PullRequestPayload) error {
	task, project, err := s.resolveTask(ctx, ref)
	if err != nil || task == nil {
		return nil
	}
	if err := s.recordPullRequest(ctx, task.ID, repoFullName, pr); err != nil {
		return err
	}
	if err := s.applyWorkflowTransition(ctx, task, project, pr, defaultBranch); err != nil {
		return err
	}
	if pr.User != nil {
		if err := s.autoAssignActor(ctx, task, pr.User.ID); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	return nil
}

func (s *Service) handlePush(ctx context.Context, payload *WebhookPayload) error {
	var texts []string
	if payload.Ref != "" {
		texts = append(texts, payload.Ref)
	}
	for _, c := range payload.Commits {
		texts = append(texts, c.Message)
	}
	refs := extractTaskRefs(texts...)
	for _, ref := range refs {
		task, project, err := s.resolveTask(ctx, ref)
		if err != nil || task == nil {
			continue
		}
		if err := s.moveTaskToInProgress(ctx, task, project); err != nil {
			return err
		}
		if payload.Sender != nil {
			if err := s.autoAssignActor(ctx, task, payload.Sender.ID); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
	}
	return nil
}

func (s *Service) resolveTask(ctx context.Context, ref TaskRef) (*schemas.Task, *schemas.Project, error) {
	var project schemas.Project
	if err := s.db.WithContext(ctx).Where("LOWER(slug) = ?", strings.ToLower(ref.ProjectSlug)).First(&project).Error; err != nil {
		return nil, nil, err
	}
	var task schemas.Task
	if err := s.db.WithContext(ctx).Where("project_id = ? AND number = ?", project.ID, ref.TaskNumber).First(&task).Error; err != nil {
		return nil, nil, err
	}
	return &task, &project, nil
}

func (s *Service) recordPullRequest(ctx context.Context, taskID int64, repoFullName string, pr *PullRequestPayload) error {
	state := derivePRState(pr)
	var authorID *string
	if pr.User != nil && pr.User.ID > 0 {
		str := strconv.FormatInt(pr.User.ID, 10)
		authorID = &str
	}
	var existing schemas.TaskPullRequest
	err := s.db.WithContext(ctx).
		Where("task_id = ? AND repo_full_name = ? AND pr_number = ?", taskID, repoFullName, pr.Number).
		First(&existing).Error
	if err == nil {
		return s.db.WithContext(ctx).Model(&existing).Updates(map[string]any{
			"title":            pr.Title,
			"url":              pr.HTMLURL,
			"state":            state,
			"head_branch":      pr.Head.Ref,
			"base_branch":      pr.Base.Ref,
			"author_github_id": authorID,
			"updated_at":       time.Now(),
		}).Error
	}
	row := schemas.TaskPullRequest{
		TaskID:         taskID,
		RepoFullName:   repoFullName,
		PRNumber:       pr.Number,
		Title:          pr.Title,
		URL:            pr.HTMLURL,
		State:          state,
		HeadBranch:     pr.Head.Ref,
		BaseBranch:     pr.Base.Ref,
		AuthorGitHubID: authorID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	return s.db.WithContext(ctx).Create(&row).Error
}

func derivePRState(pr *PullRequestPayload) string {
	if pr.Merged {
		return "merged"
	}
	if pr.State == "closed" {
		return "closed"
	}
	if pr.Draft {
		return "draft"
	}
	return "open"
}

func (s *Service) applyWorkflowTransition(ctx context.Context, task *schemas.Task, project *schemas.Project, pr *PullRequestPayload, defaultBranch string) error {
	if pr.Merged && (defaultBranch == "" || pr.Base.Ref == defaultBranch) {
		return s.moveTaskToFinal(ctx, task, project)
	}
	if pr.State == "open" && !pr.Draft {
		return s.moveTaskToReview(ctx, task, project)
	}
	if pr.Draft || pr.State == "open" {
		return s.moveTaskToInProgress(ctx, task, project)
	}
	return nil
}

func (s *Service) moveTaskToFinal(ctx context.Context, task *schemas.Task, project *schemas.Project) error {
	var finalCol schemas.Column
	err := s.db.WithContext(ctx).Where("project_id = ? AND is_final = true", project.ID).First(&finalCol).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return s.db.WithContext(ctx).Model(task).Updates(map[string]any{
		"column_id":  finalCol.ID,
		"status":     finalCol.Slug,
		"updated_at": time.Now(),
	}).Error
}

func (s *Service) moveTaskToReview(ctx context.Context, task *schemas.Task, project *schemas.Project) error {
	var reviewCol schemas.Column
	err := s.db.WithContext(ctx).
		Where("project_id = ? AND (slug = 'in-review' OR LOWER(name) = 'in review')", project.ID).
		First(&reviewCol).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return s.db.WithContext(ctx).Model(task).Updates(map[string]any{
		"column_id":  reviewCol.ID,
		"status":     reviewCol.Slug,
		"updated_at": time.Now(),
	}).Error
}

func (s *Service) moveTaskToInProgress(ctx context.Context, task *schemas.Task, project *schemas.Project) error {
	if task.Status != "to-do" && task.Status != "backlog" && task.ColumnID != nil {
		return nil
	}
	var inProgressCol schemas.Column
	err := s.db.WithContext(ctx).
		Where("project_id = ? AND (slug = 'in-progress' OR LOWER(name) = 'in progress')", project.ID).
		First(&inProgressCol).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return s.db.WithContext(ctx).Model(task).Updates(map[string]any{
		"column_id":  inProgressCol.ID,
		"status":     inProgressCol.Slug,
		"updated_at": time.Now(),
	}).Error
}

func (s *Service) autoAssignActor(ctx context.Context, task *schemas.Task, githubID int64) error {
	if task.ActorID != nil || githubID <= 0 {
		return nil
	}
	targetID := strconv.FormatInt(githubID, 10)
	var user schemas.User
	if err := s.db.WithContext(ctx).Where("github_id = ?", targetID).First(&user).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(task).Updates(map[string]any{
		"actor_id":   user.ID,
		"updated_at": time.Now(),
	}).Error
}
