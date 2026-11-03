package workspace

import (
	"context"
	"strconv"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// MemberMe answers the caller's own membership row in a workspace.
func (service *Service) MemberMe(ctx context.Context, userID, workspaceID int64) (*MemberResponse, error) {
	if err := service.RequireMember(ctx, userID, workspaceID); err != nil {
		return nil, err
	}
	return service.memberByUserID(ctx, workspaceID, userID)
}

// UpdateMemberRole changes one member's role, unless Registre owns the workspace.
func (service *Service) UpdateMemberRole(ctx context.Context, callerID, workspaceID, targetUserID int64, role string) (*MemberResponse, error) {
	if err := service.requireRole(ctx, callerID, workspaceID, "owner"); err != nil {
		return nil, err
	}
	if err := service.requireLocal(ctx, workspaceID); err != nil {
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

// RemoveMember deletes a membership, unless Registre owns the workspace.
func (service *Service) RemoveMember(ctx context.Context, callerID, workspaceID int64, target string) (*MessageResponse, error) {
	if err := service.requireRole(ctx, callerID, workspaceID, "owner", "admin"); err != nil {
		return nil, err
	}
	if err := service.requireLocal(ctx, workspaceID); err != nil {
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

// RoleFor returns the caller's role in a workspace.
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

func (service *Service) ownerCount(ctx context.Context, workspaceID int64) (int64, error) {
	var count int64
	err := service.orm.WithContext(ctx).Model(&schemas.WorkspaceMember{}).
		Where("workspace_id = ? AND role = ?", workspaceID, "owner").Count(&count).Error
	if err != nil {
		return 0, errors.Internal("failed to count the workspace owners", err)
	}
	return count, nil
}

func validRole(role string) bool {
	switch role {
	case "owner", "admin", "member":
		return true
	default:
		return false
	}
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
