// Workspace lifecycle: create, update, delete, the full-organization read,
// the caller's own membership row, and member role management. The read-only
// list and the logo operations live in service.go.
package workspace

import (
	"context"
	stderrors "errors"
	"fmt"
	"strconv"
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

// Update writes the present keys of a workspace patch. Owners and admins may
// edit; the slug must stay unique against every other workspace.
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
		slug := slugify(*req.Slug)
		if slug == "" {
			return nil, errors.Invalid("Workspace slug cannot be empty")
		}
		taken, err := service.slugTaken(ctx, slug, workspaceID)
		if err != nil {
			return nil, err
		}
		if taken {
			return nil, errors.Conflict("A workspace with this slug already exists")
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

// Delete removes a workspace and everything that hangs off it: memberships
// and invitations first, then the row itself.
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

// Full answers the workspace row and every member, the shape the organization
// settings screen renders.
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
	return &FullWorkspaceResponse{Workspace: workspace, Members: members}, nil
}

// MemberMe answers the caller's own membership row in a workspace.
func (service *Service) MemberMe(ctx context.Context, userID, workspaceID int64) (*MemberResponse, error) {
	if err := service.RequireMember(ctx, userID, workspaceID); err != nil {
		return nil, err
	}
	return service.memberByUserID(ctx, workspaceID, userID)
}

// UpdateMemberRole changes one member's role. Only owners may; nobody may
// demote themselves, and the last owner cannot be demoted at all.
func (service *Service) UpdateMemberRole(ctx context.Context, callerID, workspaceID, targetUserID int64, role string) (*MemberResponse, error) {
	if err := service.requireRole(ctx, callerID, workspaceID, "owner"); err != nil {
		return nil, err
	}
	if !validRole(role) {
		return nil, errors.Invalid("Role must be owner, admin or member")
	}
	if callerID == targetUserID {
		return nil, errors.Invalid("An owner cannot change their own role")
	}
	member, err := service.memberByUserID(ctx, workspaceID, targetUserID)
	if err != nil {
		return nil, err
	}
	if member.Role == "owner" && role != "owner" {
		owners, err := service.ownerCount(ctx, workspaceID)
		if err != nil {
			return nil, err
		}
		if owners <= 1 {
			return nil, errors.Invalid("A workspace needs at least one owner")
		}
	}
	if err := service.orm.WithContext(ctx).Model(&schemas.WorkspaceMember{}).
		Where("user_id = ? AND workspace_id = ?", targetUserID, workspaceID).
		Update("role", role).Error; err != nil {
		return nil, errors.Internal("failed to update the member role", err)
	}
	member.Role = role
	return member, nil
}

// RemoveMember deletes a membership. The target names the member by id or by
// email, matching the frontend's memberIdOrEmail. The last owner cannot be
// removed, or the workspace is stranded.
func (service *Service) RemoveMember(ctx context.Context, callerID, workspaceID int64, target string) (*MessageResponse, error) {
	if err := service.requireRole(ctx, callerID, workspaceID, "owner", "admin"); err != nil {
		return nil, err
	}
	targetUserID, err := service.memberUserID(ctx, workspaceID, target)
	if err != nil {
		return nil, err
	}
	member, err := service.memberByUserID(ctx, workspaceID, targetUserID)
	if err != nil {
		return nil, err
	}
	// Only owners manage owners: an admin removing a co-owner (or the owner)
	// escalates past the role boundary invite already enforces.
	if member.Role == "owner" {
		if err := service.requireRole(ctx, callerID, workspaceID, "owner"); err != nil {
			return nil, err
		}
		owners, err := service.ownerCount(ctx, workspaceID)
		if err != nil {
			return nil, err
		}
		if owners <= 1 {
			return nil, errors.Invalid("The workspace owner cannot be removed")
		}
	}
	if err := service.orm.WithContext(ctx).
		Where("user_id = ? AND workspace_id = ?", targetUserID, workspaceID).
		Delete(&schemas.WorkspaceMember{}).Error; err != nil {
		return nil, errors.Internal("failed to remove the member", err)
	}
	return &MessageResponse{Message: "Member removed"}, nil
}

// RoleFor returns the caller's role in a workspace, forbidding the request
// when they are not a member at all.
func (service *Service) RoleFor(ctx context.Context, userID, workspaceID int64) (string, error) {
	var role string
	err := service.orm.WithContext(ctx).
		Model(&schemas.WorkspaceMember{}).
		Where("user_id = ? AND workspace_id = ?", userID, workspaceID).
		Select("role").Scan(&role).Error
	if err != nil {
		return "", errors.Internal("failed to check workspace access", err)
	}
	if role == "" {
		return "", errors.Forbidden("You don't have access to this workspace")
	}
	return role, nil
}

// requireRole gates an action behind one of the allowed roles.
func (service *Service) requireRole(ctx context.Context, userID, workspaceID int64, roles ...string) error {
	role, err := service.RoleFor(ctx, userID, workspaceID)
	if err != nil {
		return err
	}
	for _, allowed := range roles {
		if role == allowed {
			return nil
		}
	}
	return errors.Forbidden("You don't have permission to do this")
}

// workspaceByID loads one workspace row as the response shape.
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

// memberByUserID loads one member joined to their user row.
func (service *Service) memberByUserID(ctx context.Context, workspaceID, userID int64) (*MemberResponse, error) {
	var member MemberResponse
	err := service.orm.WithContext(ctx).
		Table("users").
		Select("users.id, users.name, users.email, users.image, workspace_members.role").
		Joins("JOIN workspace_members ON workspace_members.user_id = users.id").
		Where("workspace_members.workspace_id = ? AND workspace_members.user_id = ?", workspaceID, userID).
		Scan(&member).Error
	if err != nil {
		return nil, errors.Internal("failed to load the member", err)
	}
	if member.ID == 0 {
		return nil, errors.NotFound("Member not found")
	}
	return &member, nil
}

// memberUserID resolves a member target, numeric id or email, to a user id.
func (service *Service) memberUserID(ctx context.Context, workspaceID int64, target string) (int64, error) {
	if userID, err := strconv.ParseInt(target, 10, 64); err == nil {
		return userID, nil
	}
	var userID int64
	err := service.orm.WithContext(ctx).
		Table("users").
		Select("users.id").
		Joins("JOIN workspace_members ON workspace_members.user_id = users.id").
		Where("workspace_members.workspace_id = ? AND users.email = ?", workspaceID, strings.ToLower(strings.TrimSpace(target))).
		Scan(&userID).Error
	if err != nil {
		return 0, errors.Internal("failed to find the member", err)
	}
	if userID == 0 {
		return 0, errors.NotFound("Member not found")
	}
	return userID, nil
}

// ownerCount counts the owner memberships of a workspace.
func (service *Service) ownerCount(ctx context.Context, workspaceID int64) (int64, error) {
	var count int64
	err := service.orm.WithContext(ctx).Model(&schemas.WorkspaceMember{}).
		Where("workspace_id = ? AND role = ?", workspaceID, "owner").Count(&count).Error
	if err != nil {
		return 0, errors.Internal("failed to count the workspace owners", err)
	}
	return count, nil
}

// uniqueSlug returns the slug itself when free, or a numbered variant, so a
// duplicate name still produces a valid unique identifier.
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

// slugTaken reports whether another workspace already holds the slug.
func (service *Service) slugTaken(ctx context.Context, slug string, excludeID int64) (bool, error) {
	var count int64
	err := service.orm.WithContext(ctx).Model(&schemas.Workspace{}).
		Where("slug = ? AND id != ?", slug, excludeID).Count(&count).Error
	if err != nil {
		return false, errors.Internal("failed to check the slug", err)
	}
	return count > 0, nil
}

// slugify normalizes any input into a URL-safe slug: lowercase, letters and
// digits kept, everything else one dash.
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

// validRole reports whether role is one of the three membership roles.
func validRole(role string) bool {
	switch role {
	case "owner", "admin", "member":
		return true
	default:
		return false
	}
}

// workspaceResponse renders a schema row as the wire shape.
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

// Controller methods below mirror the service one for one, keeping the route
// handlers thin and the service testable without HTTP.

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

func (controller *Controller) memberMe(ctx context.Context, userID, workspaceID int64) (*MemberResponse, error) {
	return controller.service.MemberMe(ctx, userID, workspaceID)
}

func (controller *Controller) updateMemberRole(ctx context.Context, userID, workspaceID, targetUserID int64, req *UpdateMemberRoleRequest) (*MemberResponse, error) {
	return controller.service.UpdateMemberRole(ctx, userID, workspaceID, targetUserID, req.Role)
}

func (controller *Controller) removeMember(ctx context.Context, userID, workspaceID int64, target string) (*MessageResponse, error) {
	return controller.service.RemoveMember(ctx, userID, workspaceID, target)
}
