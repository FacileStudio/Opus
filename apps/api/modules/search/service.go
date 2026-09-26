package search

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// Service coordinates global search queries across domains.
type Service struct {
	db         *gorm.DB
	controller *Controller
}

// NewService instantiates a new search service.
func NewService(db *gorm.DB) *Service {
	service := &Service{db: db}
	service.controller = newController(service)
	return service
}

func (service *Service) resolveWorkspaces(ctx context.Context, userID int64, requestedID *int64) ([]int64, error) {
	if requestedID != nil {
		var count int64
		err := service.db.WithContext(ctx).
			Model(&schemas.WorkspaceMember{}).
			Where("user_id = ? AND workspace_id = ?", userID, *requestedID).
			Count(&count).Error
		if err != nil {
			return nil, errors.Internal("failed to verify workspace access", err)
		}
		if count == 0 {
			return nil, errors.Forbidden("You don't have access to this workspace")
		}
		return []int64{*requestedID}, nil
	}
	var wsIDs []int64
	err := service.db.WithContext(ctx).
		Model(&schemas.WorkspaceMember{}).
		Where("user_id = ?", userID).
		Pluck("workspace_id", &wsIDs).Error
	if err != nil {
		return nil, errors.Internal("failed to list user workspaces", err)
	}
	return wsIDs, nil
}

// Search executes a multi-domain query scoped to the user's workspaces.
func (service *Service) Search(ctx context.Context, userID int64, filter SearchFilter) (*SearchResponse, error) {
	wsIDs, err := service.resolveWorkspaces(ctx, userID, filter.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if len(wsIDs) == 0 {
		return &SearchResponse{Results: []SearchItem{}, TotalCount: 0, SearchQuery: filter.Query}, nil
	}

	term := "%" + strings.ToLower(filter.Query) + "%"
	var all []SearchItem

	if filter.Type == "all" || filter.Type == "tasks" {
		all = append(all, service.queryTasks(ctx, term, wsIDs, filter.ProjectID)...)
	}
	if filter.Type == "all" || filter.Type == "projects" {
		all = append(all, service.queryProjects(ctx, term, wsIDs)...)
	}
	if filter.Type == "all" || filter.Type == "workspaces" {
		all = append(all, service.queryWorkspaces(ctx, term, wsIDs)...)
	}
	if filter.Type == "all" || filter.Type == "comments" {
		all = append(all, service.queryComments(ctx, term, wsIDs)...)
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].RelevanceScore != all[j].RelevanceScore {
			return all[i].RelevanceScore > all[j].RelevanceScore
		}
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	total := len(all)
	if filter.Limit > 0 && len(all) > filter.Limit {
		all = all[:filter.Limit]
	}

	return &SearchResponse{Results: all, TotalCount: total, SearchQuery: filter.Query}, nil
}

type taskRow struct {
	ID          int64
	Title       string
	Description string
	Status      string
	Priority    string
	Number      int
	ProjectID   int64
	ProjectName string
	ProjectSlug string
	WorkspaceID int64
	CreatedAt   time.Time
}

func (service *Service) queryTasks(ctx context.Context, term string, wsIDs []int64, projectID *int64) []SearchItem {
	query := service.db.WithContext(ctx).
		Table("tasks").
		Select("tasks.id, tasks.title, tasks.description, tasks.status, tasks.priority, tasks.number, tasks.project_id, tasks.workspace_id, tasks.created_at, projects.name as project_name, projects.slug as project_slug").
		Joins("JOIN projects ON projects.id = tasks.project_id").
		Where("tasks.workspace_id IN (?) AND (LOWER(tasks.title) LIKE ? OR LOWER(tasks.description) LIKE ?)", wsIDs, term, term)
	if projectID != nil {
		query = query.Where("tasks.project_id = ?", *projectID)
	}
	var rows []taskRow
	query.Order("tasks.created_at desc").Limit(50).Scan(&rows)
	items := make([]SearchItem, 0, len(rows))
	for _, r := range rows {
		pID := strconv.FormatInt(r.ProjectID, 10)
		wID := strconv.FormatInt(r.WorkspaceID, 10)
		num := r.Number
		prio := r.Priority
		stat := r.Status
		desc := r.Description
		items = append(items, SearchItem{
			ID:             strconv.FormatInt(r.ID, 10),
			Type:           "task",
			Title:          r.Title,
			Description:    &desc,
			CreatedAt:      r.CreatedAt,
			RelevanceScore: 10,
			TaskNumber:     &num,
			Priority:       &prio,
			Status:         &stat,
			SearchMetadata: SearchMetadata{
				ProjectID:   &pID,
				ProjectName: &r.ProjectName,
				ProjectSlug: &r.ProjectSlug,
				WorkspaceID: &wID,
			},
		})
	}
	return items
}

func (service *Service) queryProjects(ctx context.Context, term string, wsIDs []int64) []SearchItem {
	var projects []schemas.Project
	service.db.WithContext(ctx).
		Where("workspace_id IN (?) AND (LOWER(name) LIKE ? OR LOWER(description) LIKE ?)", wsIDs, term, term).
		Order("created_at desc").Limit(20).
		Find(&projects)
	items := make([]SearchItem, 0, len(projects))
	for _, p := range projects {
		pID := strconv.FormatInt(p.ID, 10)
		wID := strconv.FormatInt(p.WorkspaceID, 10)
		desc := p.Description
		items = append(items, SearchItem{
			ID:             pID,
			Type:           "project",
			Title:          p.Name,
			Description:    &desc,
			CreatedAt:      p.CreatedAt,
			RelevanceScore: 8,
			SearchMetadata: SearchMetadata{
				ProjectID:   &pID,
				ProjectName: &p.Name,
				ProjectSlug: &p.Slug,
				WorkspaceID: &wID,
			},
		})
	}
	return items
}

func (service *Service) queryWorkspaces(ctx context.Context, term string, wsIDs []int64) []SearchItem {
	var workspaces []schemas.Workspace
	service.db.WithContext(ctx).
		Where("id IN (?) AND (LOWER(name) LIKE ? OR LOWER(description) LIKE ?)", wsIDs, term, term).
		Order("created_at desc").Limit(10).
		Find(&workspaces)
	items := make([]SearchItem, 0, len(workspaces))
	for _, w := range workspaces {
		wID := strconv.FormatInt(w.ID, 10)
		desc := w.Description
		items = append(items, SearchItem{
			ID:             wID,
			Type:           "workspace",
			Title:          w.Name,
			Description:    &desc,
			CreatedAt:      w.CreatedAt,
			RelevanceScore: 5,
			SearchMetadata: SearchMetadata{
				WorkspaceID:   &wID,
				WorkspaceName: &w.Name,
			},
		})
	}
	return items
}

type commentRow struct {
	ID          int64
	Content     string
	TaskID      int64
	TaskTitle   string
	ProjectID   int64
	WorkspaceID int64
	CreatedAt   time.Time
}

func (service *Service) queryComments(ctx context.Context, term string, wsIDs []int64) []SearchItem {
	var rows []commentRow
	service.db.WithContext(ctx).
		Table("comments").
		Select("comments.id, comments.content, comments.task_id, comments.created_at, tasks.title as task_title, tasks.project_id, tasks.workspace_id").
		Joins("JOIN tasks ON tasks.id = comments.task_id").
		Where("tasks.workspace_id IN (?) AND LOWER(comments.content) LIKE ?", wsIDs, term).
		Order("comments.created_at desc").Limit(20).
		Scan(&rows)
	items := make([]SearchItem, 0, len(rows))
	for _, c := range rows {
		pID := strconv.FormatInt(c.ProjectID, 10)
		wID := strconv.FormatInt(c.WorkspaceID, 10)
		content := c.Content
		items = append(items, SearchItem{
			ID:             strconv.FormatInt(c.ID, 10),
			Type:           "comment",
			Title:          "Comment on " + c.TaskTitle,
			Content:        &content,
			CreatedAt:      c.CreatedAt,
			RelevanceScore: 4,
			SearchMetadata: SearchMetadata{
				ProjectID:   &pID,
				WorkspaceID: &wID,
			},
		})
	}
	return items
}
