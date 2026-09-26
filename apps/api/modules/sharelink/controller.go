package sharelink

import (
	"context"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// Controller adapts incoming requests to share link service calls.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) list(ctx context.Context, userID, projectID int64) ([]ShareLinkResponse, error) {
	if projectID <= 0 {
		return nil, errors.Invalid("invalid project id")
	}
	links, err := controller.service.ListByProject(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	responses := make([]ShareLinkResponse, 0, len(links))
	for i := range links {
		responses = append(responses, toShareLinkResponse(&links[i]))
	}
	return responses, nil
}

func (controller *Controller) create(ctx context.Context, userID int64, req *CreateShareLinkRequest) (*ShareLinkResponse, error) {
	if req.ProjectID <= 0 {
		return nil, errors.Invalid("projectId is required")
	}
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		return nil, errors.Invalid("expiry date must be in the future")
	}
	link, err := controller.service.Create(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	resp := toShareLinkResponse(link)
	return &resp, nil
}

func (controller *Controller) update(ctx context.Context, userID, shareLinkID int64, req *UpdateShareLinkRequest) (*ShareLinkResponse, error) {
	if shareLinkID <= 0 {
		return nil, errors.Invalid("invalid share link id")
	}
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		return nil, errors.Invalid("expiry date must be in the future")
	}
	link, err := controller.service.Update(ctx, userID, shareLinkID, req)
	if err != nil {
		return nil, err
	}
	resp := toShareLinkResponse(link)
	return &resp, nil
}

func (controller *Controller) delete(ctx context.Context, userID, shareLinkID int64) (*ShareLinkResponse, error) {
	if shareLinkID <= 0 {
		return nil, errors.Invalid("invalid share link id")
	}
	link, err := controller.service.Delete(ctx, userID, shareLinkID)
	if err != nil {
		return nil, err
	}
	resp := toShareLinkResponse(link)
	return &resp, nil
}

func (controller *Controller) getShared(ctx context.Context, token string) (*SharedProjectData, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.Invalid("token is required")
	}
	result, err := controller.service.GetSharedProject(ctx, token)
	if err != nil {
		return nil, err
	}
	return toSharedProjectData(result), nil
}

func toShareLinkResponse(link *schemas.ShareLink) ShareLinkResponse {
	return ShareLinkResponse{
		ID:             link.ID,
		ProjectID:      link.ProjectID,
		Token:          link.Token,
		Label:          link.Label,
		ExpiresAt:      link.ExpiresAt,
		IsActive:       link.IsActive,
		LastAccessedAt: link.LastAccessedAt,
		AccessCount:    link.AccessCount,
		CreatedBy:      link.CreatedBy,
		CreatedAt:      link.CreatedAt,
		UpdatedAt:      link.UpdatedAt,
	}
}

func toSharedProjectData(result *SharedProjectResult) *SharedProjectData {
	cols := make([]SharedColumnResponse, 0, len(result.Columns))
	for _, col := range result.Columns {
		cols = append(cols, SharedColumnResponse{
			ID:       col.ID,
			Name:     col.Name,
			Slug:     col.Slug,
			Position: col.Position,
			Icon:     col.Icon,
			Color:    col.Color,
			IsFinal:  col.IsFinal,
		})
	}
	tasks := make([]SharedTaskResponse, 0, len(result.Tasks))
	for _, t := range result.Tasks {
		tasks = append(tasks, SharedTaskResponse{
			ID:          t.ID,
			Number:      t.Number,
			Title:       t.Title,
			Description: t.Description,
			Status:      t.Status,
			Priority:    t.Priority,
			ColumnID:    t.ColumnID,
			DueDate:     t.DueDate,
			CreatedAt:   t.CreatedAt,
		})
	}
	return &SharedProjectData{
		ID:          result.Project.ID,
		Name:        result.Project.Name,
		Slug:        result.Project.Slug,
		Icon:        result.Project.Icon,
		Description: result.Project.Description,
		IsPublic:    result.Project.IsPublic,
		WorkspaceID: result.Project.WorkspaceID,
		Columns:     cols,
		Tasks:       tasks,
	}
}
