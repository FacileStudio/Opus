package column

import (
	"context"
	"fmt"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// Controller handles request validation and response mapping for column endpoints.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func toResponse(col schemas.Column) ColumnResponse {
	return ColumnResponse{
		ID:        col.ID,
		ProjectID: col.ProjectID,
		Name:      col.Name,
		Slug:      col.Slug,
		Position:  col.Position,
		Icon:      col.Icon,
		Color:     col.Color,
		IsFinal:   col.IsFinal,
		CreatedAt: col.CreatedAt,
		UpdatedAt: col.UpdatedAt,
	}
}

func toResponses(cols []schemas.Column) []ColumnResponse {
	out := make([]ColumnResponse, len(cols))
	for i, col := range cols {
		out[i] = toResponse(col)
	}
	return out
}

func isVirtualStatus(slug string) bool {
	return slug == "planned" || slug == "archived"
}

func toSlug(name string) string {
	var b strings.Builder
	lastHyphen := true
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastHyphen = false
		} else if !lastHyphen {
			b.WriteByte('-')
			lastHyphen = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

func (controller *Controller) list(ctx context.Context, userID, projectID int64) ([]ColumnResponse, error) {
	cols, err := controller.service.List(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	return toResponses(cols), nil
}

func (controller *Controller) create(ctx context.Context, userID, projectID int64, req *CreateColumnRequest) (*ColumnResponse, error) {
	if req == nil {
		return nil, errors.Invalid("request body is required")
	}
	slug := toSlug(req.Name)
	if slug == "" {
		return nil, errors.Invalid("Column name must contain at least one alphanumeric character")
	}
	if isVirtualStatus(slug) {
		return nil, errors.Conflict(fmt.Sprintf("Column slug %q is reserved for virtual task statuses", slug))
	}
	col := schemas.Column{
		Name: strings.TrimSpace(req.Name),
		Slug: slug,
	}
	if req.Icon != nil {
		col.Icon = *req.Icon
	}
	if req.Color != nil {
		col.Color = *req.Color
	}
	if req.IsFinal != nil {
		col.IsFinal = *req.IsFinal
	}
	created, err := controller.service.Create(ctx, userID, projectID, &col)
	if err != nil {
		return nil, err
	}
	resp := toResponse(*created)
	return &resp, nil
}

func (controller *Controller) update(ctx context.Context, userID, columnID int64, req *UpdateColumnRequest) (*ColumnResponse, error) {
	if req == nil {
		return nil, errors.Invalid("request body is required")
	}
	fields := make(map[string]interface{})
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, errors.Invalid("Column name cannot be empty")
		}
		fields["name"] = name
	}
	if req.Icon != nil {
		fields["icon"] = *req.Icon
	}
	if req.Color != nil {
		fields["color"] = *req.Color
	}
	if req.IsFinal != nil {
		fields["is_final"] = *req.IsFinal
	}
	updated, err := controller.service.Update(ctx, userID, columnID, fields)
	if err != nil {
		return nil, err
	}
	resp := toResponse(*updated)
	return &resp, nil
}

func (controller *Controller) delete(ctx context.Context, userID, columnID int64) (*ColumnResponse, error) {
	deleted, err := controller.service.Delete(ctx, userID, columnID)
	if err != nil {
		return nil, err
	}
	resp := toResponse(*deleted)
	return &resp, nil
}

func (controller *Controller) reorder(ctx context.Context, userID, projectID int64, req *ReorderColumnsRequest) ([]ColumnResponse, error) {
	if req == nil || len(req.Columns) == 0 {
		return nil, errors.Invalid("columns list cannot be empty")
	}
	cols, err := controller.service.Reorder(ctx, userID, projectID, req.Columns)
	if err != nil {
		return nil, err
	}
	return toResponses(cols), nil
}
