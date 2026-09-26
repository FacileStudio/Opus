// Package project manages project workspaces, columns, and workflows.
package project

import (
	"context"
	stderrors "errors"
	"fmt"
	"math"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// Service handles database operations on projects and verifies workspace access.
type Service struct {
	orm        *gorm.DB
	controller *controller
}

// NewService constructs a project Service backed by the provided database handle.
func NewService(db *gorm.DB) *Service {
	s := &Service{orm: db}
	s.controller = newController(s)
	return s
}

// RequireMember verifies that the caller belongs to the workspace.
func (s *Service) RequireMember(ctx context.Context, userID, workspaceID int64) error {
	var count int64
	err := s.orm.WithContext(ctx).
		Model(&schemas.WorkspaceMember{}).
		Where("user_id = ? AND workspace_id = ?", userID, workspaceID).
		Count(&count).Error
	if err != nil {
		return errors.Internal("failed to check workspace access", err)
	}
	if count == 0 {
		return errors.Forbidden("You don't have access to this workspace")
	}
	return nil
}

// ListByWorkspace returns projects for a workspace with their aggregate task statistics.
func (s *Service) ListByWorkspace(ctx context.Context, userID, workspaceID int64, includeArchived bool) ([]schemas.Project, map[int64]ProjectStatistics, error) {
	if err := s.RequireMember(ctx, userID, workspaceID); err != nil {
		return nil, nil, err
	}
	query := s.orm.WithContext(ctx).Where("workspace_id = ?", workspaceID)
	if !includeArchived {
		query = query.Where("archived_at IS NULL")
	}
	var projects []schemas.Project
	if err := query.Order("created_at asc").Find(&projects).Error; err != nil {
		return nil, nil, errors.Internal("failed to list projects", err)
	}
	ids := make([]int64, len(projects))
	for i, p := range projects {
		ids[i] = p.ID
	}
	statsMap, err := s.loadStatistics(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	return projects, statsMap, nil
}

// GetByID loads a project and its statistics after verifying caller workspace membership.
func (s *Service) GetByID(ctx context.Context, userID, projectID int64) (*schemas.Project, *ProjectStatistics, error) {
	project, err := s.loadProjectAndVerify(ctx, userID, projectID)
	if err != nil {
		return nil, nil, err
	}
	statsMap, err := s.loadStatistics(ctx, []int64{projectID})
	if err != nil {
		return nil, nil, err
	}
	stats := statsMap[projectID]
	return project, &stats, nil
}

// Create inserts a project and scaffolds its default columns in one transaction.
func (s *Service) Create(ctx context.Context, userID int64, project schemas.Project) (*schemas.Project, error) {
	if err := s.RequireMember(ctx, userID, project.WorkspaceID); err != nil {
		return nil, err
	}
	slug, err := s.uniqueSlug(ctx, project.WorkspaceID, project.Slug)
	if err != nil {
		return nil, err
	}
	project.Slug = slug
	project.CreatedAt = time.Now()
	err = s.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&project).Error; err != nil {
			return errors.Internal("failed to create project", err)
		}
		return createDefaultColumns(tx, project.ID, project.CreatedAt)
	})
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// Update writes partial updates to a project after verifying caller workspace access.
func (s *Service) Update(ctx context.Context, userID, projectID int64, updates map[string]any) (*schemas.Project, error) {
	project, err := s.loadProjectAndVerify(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	if len(updates) > 0 {
		err := s.orm.WithContext(ctx).Model(&schemas.Project{}).Where("id = ?", projectID).Updates(updates).Error
		if err != nil {
			return nil, errors.Internal("failed to update project", err)
		}
		if err := s.orm.WithContext(ctx).First(project, projectID).Error; err != nil {
			return nil, errors.Internal("failed to reload project", err)
		}
	}
	return project, nil
}

// Delete removes a project and its associated columns and tasks in a transaction.
func (s *Service) Delete(ctx context.Context, userID, projectID int64) (*schemas.Project, error) {
	project, err := s.loadProjectAndVerify(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	err = s.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("project_id = ?", projectID).Delete(&schemas.Column{}).Error; err != nil {
			return errors.Internal("failed to delete project columns", err)
		}
		if err := tx.Where("project_id = ?", projectID).Delete(&schemas.Task{}).Error; err != nil {
			return errors.Internal("failed to delete project tasks", err)
		}
		if err := tx.Where("id = ?", projectID).Delete(&schemas.Project{}).Error; err != nil {
			return errors.Internal("failed to delete project", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return project, nil
}

// Archive sets the archived_at timestamp on a project.
func (s *Service) Archive(ctx context.Context, userID, projectID int64) (*schemas.Project, error) {
	project, err := s.loadProjectAndVerify(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	err = s.orm.WithContext(ctx).Model(&schemas.Project{}).Where("id = ?", projectID).Update("archived_at", &now).Error
	if err != nil {
		return nil, errors.Internal("failed to archive project", err)
	}
	project.ArchivedAt = &now
	return project, nil
}

// Unarchive clears the archived_at timestamp on a project.
func (s *Service) Unarchive(ctx context.Context, userID, projectID int64) (*schemas.Project, error) {
	project, err := s.loadProjectAndVerify(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	err = s.orm.WithContext(ctx).Model(&schemas.Project{}).Where("id = ?", projectID).Update("archived_at", nil).Error
	if err != nil {
		return nil, errors.Internal("failed to unarchive project", err)
	}
	project.ArchivedAt = nil
	return project, nil
}

func (s *Service) loadProjectAndVerify(ctx context.Context, userID, projectID int64) (*schemas.Project, error) {
	var project schemas.Project
	err := s.orm.WithContext(ctx).First(&project, projectID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("Project not found")
		}
		return nil, errors.Internal("failed to load project", err)
	}
	if err := s.RequireMember(ctx, userID, project.WorkspaceID); err != nil {
		return nil, err
	}
	return &project, nil
}

func (s *Service) loadStatistics(ctx context.Context, projectIDs []int64) (map[int64]ProjectStatistics, error) {
	out := make(map[int64]ProjectStatistics, len(projectIDs))
	if len(projectIDs) == 0 {
		return out, nil
	}
	for _, id := range projectIDs {
		out[id] = ProjectStatistics{}
	}
	var tasks []schemas.Task
	err := s.orm.WithContext(ctx).Where("project_id IN ?", projectIDs).Find(&tasks).Error
	if err != nil {
		return nil, errors.Internal("failed to load project task statistics", err)
	}
	for _, t := range tasks {
		stat := out[t.ProjectID]
		stat.TotalTasks++
		if t.Status == "done" || t.Status == "archived" {
			stat.CompletionPercentage++
		}
		if t.DueDate != nil && (stat.DueDate == nil || t.DueDate.Before(*stat.DueDate)) {
			stat.DueDate = t.DueDate
		}
		out[t.ProjectID] = stat
	}
	for id, stat := range out {
		if stat.TotalTasks > 0 {
			stat.CompletionPercentage = int(math.Round(float64(stat.CompletionPercentage) / float64(stat.TotalTasks) * 100))
			out[id] = stat
		}
	}
	return out, nil
}

func (s *Service) uniqueSlug(ctx context.Context, workspaceID int64, base string) (string, error) {
	candidate := base
	for suffix := 2; ; suffix++ {
		var count int64
		err := s.orm.WithContext(ctx).Model(&schemas.Project{}).
			Where("workspace_id = ? AND slug = ?", workspaceID, candidate).
			Count(&count).Error
		if err != nil {
			return "", errors.Internal("failed to check project slug", err)
		}
		if count == 0 {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, suffix)
	}
}

func createDefaultColumns(tx *gorm.DB, projectID int64, now time.Time) error {
	columns := []schemas.Column{
		{ProjectID: projectID, Name: "To Do", Slug: "to-do", Position: 0, IsFinal: false, CreatedAt: now, UpdatedAt: now},
		{ProjectID: projectID, Name: "In Progress", Slug: "in-progress", Position: 1, IsFinal: false, CreatedAt: now, UpdatedAt: now},
		{ProjectID: projectID, Name: "In Review", Slug: "in-review", Position: 2, IsFinal: false, CreatedAt: now, UpdatedAt: now},
		{ProjectID: projectID, Name: "Done", Slug: "done", Position: 3, IsFinal: true, CreatedAt: now, UpdatedAt: now},
	}
	for i := range columns {
		if err := tx.Create(&columns[i]).Error; err != nil {
			return err
		}
	}
	return nil
}
