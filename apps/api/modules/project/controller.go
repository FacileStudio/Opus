// Package project manages project workspaces, columns, and workflows.
package project

import (
	"context"
	"strconv"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

type controller struct {
	service *Service
}

func newController(service *Service) *controller {
	return &controller{service: service}
}

func (c *controller) list(ctx context.Context, userID int64, workspaceIDStr string, includeArchived bool) ([]ProjectResponse, error) {
	workspaceID, err := strconv.ParseInt(strings.TrimSpace(workspaceIDStr), 10, 64)
	if err != nil || workspaceID <= 0 {
		return nil, errors.Invalid("workspaceId is required")
	}
	projects, statsMap, err := c.service.ListByWorkspace(ctx, userID, workspaceID, includeArchived)
	if err != nil {
		return nil, err
	}
	out := make([]ProjectResponse, 0, len(projects))
	for i := range projects {
		p := &projects[i]
		stat := statsMap[p.ID]
		out = append(out, *toProjectResponse(p, &stat))
	}
	return out, nil
}

func (c *controller) create(ctx context.Context, userID int64, req *CreateProjectRequest) (*ProjectResponse, error) {
	if req == nil {
		return nil, errors.Invalid("request body is required")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.Invalid("project name is required")
	}
	if req.WorkspaceID <= 0 {
		return nil, errors.Invalid("workspaceId is required")
	}
	icon := strings.TrimSpace(req.Icon)
	if icon == "" {
		icon = "Layout"
	}
	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = slugify(name)
	} else {
		slug = slugify(slug)
	}
	project := schemas.Project{
		WorkspaceID: req.WorkspaceID,
		Name:        name,
		Slug:        slug,
		Icon:        icon,
	}
	created, err := c.service.Create(ctx, userID, project)
	if err != nil {
		return nil, err
	}
	stats := &ProjectStatistics{}
	return toProjectResponse(created, stats), nil
}

func (c *controller) get(ctx context.Context, userID, projectID int64) (*ProjectResponse, error) {
	if projectID <= 0 {
		return nil, errors.Invalid("invalid project id")
	}
	project, stats, err := c.service.GetByID(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	return toProjectResponse(project, stats), nil
}

func (c *controller) update(ctx context.Context, userID, projectID int64, req *UpdateProjectRequest) (*ProjectResponse, error) {
	if projectID <= 0 {
		return nil, errors.Invalid("invalid project id")
	}
	if req == nil {
		return nil, errors.Invalid("request body is required")
	}
	updates := map[string]any{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, errors.Invalid("project name cannot be empty")
		}
		updates["name"] = name
	}
	if req.Icon != nil {
		updates["icon"] = strings.TrimSpace(*req.Icon)
	}
	if req.Slug != nil {
		slug := slugify(*req.Slug)
		if slug == "" {
			return nil, errors.Invalid("project slug cannot be empty")
		}
		updates["slug"] = slug
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	updated, err := c.service.Update(ctx, userID, projectID, updates)
	if err != nil {
		return nil, err
	}
	statsMap, _ := c.service.loadStatistics(ctx, []int64{projectID})
	stat := statsMap[projectID]
	return toProjectResponse(updated, &stat), nil
}

func (c *controller) delete(ctx context.Context, userID, projectID int64) (*ProjectResponse, error) {
	if projectID <= 0 {
		return nil, errors.Invalid("invalid project id")
	}
	deleted, err := c.service.Delete(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	return toProjectResponse(deleted, nil), nil
}

func (c *controller) archive(ctx context.Context, userID, projectID int64) (*ProjectResponse, error) {
	if projectID <= 0 {
		return nil, errors.Invalid("invalid project id")
	}
	archived, err := c.service.Archive(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	return toProjectResponse(archived, nil), nil
}

func (c *controller) unarchive(ctx context.Context, userID, projectID int64) (*ProjectResponse, error) {
	if projectID <= 0 {
		return nil, errors.Invalid("invalid project id")
	}
	unarchived, err := c.service.Unarchive(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	return toProjectResponse(unarchived, nil), nil
}

func toProjectResponse(project *schemas.Project, stats *ProjectStatistics) *ProjectResponse {
	return &ProjectResponse{
		ID:          project.ID,
		WorkspaceID: project.WorkspaceID,
		Slug:        project.Slug,
		Icon:        project.Icon,
		Name:        project.Name,
		Description: project.Description,
		IsPublic:    project.IsPublic,
		ArchivedAt:  project.ArchivedAt,
		CreatedAt:   project.CreatedAt,
		Statistics:  stats,
	}
}

func slugify(input string) string {
	lower := strings.ToLower(strings.TrimSpace(input))
	var builder strings.Builder
	lastDash := false
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
	}
	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		return "project"
	}
	return slug
}
