package workspace

import (
	"context"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// Create makes a workspace and grants its creator the owner membership in one
// transaction, so a workspace can never exist without an owner.
func (service *Service) Create(ctx context.Context, userID int64, req CreateWorkspaceRequest) (*WorkspaceResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.Invalid("Workspace name is required")
	}
	slug := slugify(name)
	if strings.TrimSpace(req.Slug) != "" {
		slug = slugify(req.Slug)
	}
	unique, err := service.uniqueSlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	workspace := schemas.Workspace{Name: name, Slug: unique}
	if req.Description != nil {
		workspace.Description = *req.Description
	}
	err = service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&workspace).Error; err != nil {
			return errors.Internal("failed to create the workspace", err)
		}
		member := schemas.WorkspaceMember{
			WorkspaceID: workspace.ID,
			UserID:      userID,
			Role:        "owner",
			JoinedAt:    time.Now(),
		}
		if err := tx.Create(&member).Error; err != nil {
			return errors.Internal("failed to create the owner membership", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return workspaceResponse(workspace), nil
}

func (service *Service) resolveUpdateSlug(ctx context.Context, rawSlug string, workspaceID int64) (string, error) {
	slug := slugify(rawSlug)
	if slug == "" {
		return "", errors.Invalid("Workspace slug cannot be empty")
	}
	taken, err := service.slugTaken(ctx, slug, workspaceID)
	if err != nil {
		return "", err
	}
	if taken {
		return "", errors.Conflict("A workspace with this slug already exists")
	}
	return slug, nil
}

// Update writes the present keys of a workspace patch.
func (service *Service) Update(ctx context.Context, userID, workspaceID int64, req UpdateWorkspaceRequest) (*WorkspaceResponse, error) {
	if err := service.requireRole(ctx, userID, workspaceID, "owner", "admin"); err != nil {
		return nil, err
	}
	if _, err := service.workspaceByID(ctx, workspaceID); err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, errors.Invalid("Workspace name cannot be empty")
		}
		updates["name"] = name
	}
	if req.Slug != nil {
		slug, err := service.resolveUpdateSlug(ctx, *req.Slug, workspaceID)
		if err != nil {
			return nil, err
		}
		updates["slug"] = slug
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if len(updates) > 0 {
		if err := service.orm.WithContext(ctx).Model(&schemas.Workspace{}).
			Where("id = ?", workspaceID).Updates(updates).Error; err != nil {
			return nil, errors.Internal("failed to update the workspace", err)
		}
	}
	return service.workspaceByID(ctx, workspaceID)
}

// Delete removes a workspace and its memberships and invitations.
func (service *Service) Delete(ctx context.Context, userID, workspaceID int64) (*MessageResponse, error) {
	if err := service.requireRole(ctx, userID, workspaceID, "owner"); err != nil {
		return nil, err
	}
	err := service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workspace_id = ?", workspaceID).Delete(&schemas.WorkspaceMember{}).Error; err != nil {
			return errors.Internal("failed to clear the workspace members", err)
		}
		if err := tx.Where("workspace_id = ?", workspaceID).Delete(&schemas.Invitation{}).Error; err != nil {
			return errors.Internal("failed to clear the workspace invitations", err)
		}
		if err := tx.Where("id = ?", workspaceID).Delete(&schemas.Workspace{}).Error; err != nil {
			return errors.Internal("failed to delete the workspace", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &MessageResponse{Message: "Workspace deleted"}, nil
}

// Full answers the workspace row and every member.
func (service *Service) Full(ctx context.Context, userID, workspaceID int64) (*FullWorkspaceResponse, error) {
	if err := service.RequireMember(ctx, userID, workspaceID); err != nil {
		return nil, err
	}
	workspace, err := service.workspaceByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	members, err := service.Members(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	invitations, err := service.queryInvitations(ctx, "invitations.workspace_id = ?", workspaceID)
	if err != nil {
		return nil, err
	}
	return &FullWorkspaceResponse{Workspace: workspace, Members: members, Invitations: invitations}, nil
}

func (service *Service) workspaceByID(ctx context.Context, workspaceID int64) (*WorkspaceResponse, error) {
	var workspace schemas.Workspace
	err := service.orm.WithContext(ctx).First(&workspace, workspaceID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("Workspace not found")
		}
		return nil, errors.Internal("failed to load the workspace", err)
	}
	return workspaceResponse(workspace), nil
}

func (service *Service) uniqueSlug(ctx context.Context, base string) (string, error) {
	candidate := base
	for suffix := 2; ; suffix++ {
		var count int64
		if err := service.orm.WithContext(ctx).Model(&schemas.Workspace{}).
			Where("slug = ?", candidate).Count(&count).Error; err != nil {
			return "", errors.Internal("failed to check the slug", err)
		}
		if count == 0 {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, suffix)
	}
}

func (service *Service) slugTaken(ctx context.Context, slug string, excludeID int64) (bool, error) {
	var count int64
	err := service.orm.WithContext(ctx).Model(&schemas.Workspace{}).
		Where("slug = ? AND id != ?", slug, excludeID).Count(&count).Error
	if err != nil {
		return false, errors.Internal("failed to check the slug", err)
	}
	return count > 0, nil
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
		return "workspace"
	}
	return slug
}

func workspaceResponse(workspace schemas.Workspace) *WorkspaceResponse {
	return &WorkspaceResponse{
		ID:          workspace.ID,
		Name:        workspace.Name,
		Slug:        workspace.Slug,
		Logo:        workspace.Logo,
		Description: workspace.Description,
		CreatedAt:   workspace.CreatedAt,
	}
}

func (controller *Controller) create(ctx context.Context, userID int64, req *CreateWorkspaceRequest) (*WorkspaceResponse, error) {
	return controller.service.Create(ctx, userID, *req)
}

func (controller *Controller) update(ctx context.Context, userID, workspaceID int64, req *UpdateWorkspaceRequest) (*WorkspaceResponse, error) {
	return controller.service.Update(ctx, userID, workspaceID, *req)
}

func (controller *Controller) delete(ctx context.Context, userID, workspaceID int64) (*MessageResponse, error) {
	return controller.service.Delete(ctx, userID, workspaceID)
}

func (controller *Controller) full(ctx context.Context, userID, workspaceID int64) (*FullWorkspaceResponse, error) {
	return controller.service.Full(ctx, userID, workspaceID)
}
