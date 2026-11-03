package workspace

import (
	"context"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// AcceptInvitation moves a pending invitation to accepted and creates the membership.
func (service *Service) AcceptInvitation(ctx context.Context, callerID int64, callerEmail string, id int64) (*AcceptInvitationResponse, error) {
	invitation, err := service.InvitationByID(ctx, callerID, callerEmail, id)
	if err != nil {
		return nil, err
	}
	if invitation.Status != "pending" {
		return nil, errors.Conflict("This invitation is no longer pending")
	}
	if service.expired(invitation) {
		return nil, errors.Conflict("This invitation has expired")
	}
	if !strings.EqualFold(strings.TrimSpace(invitation.Email), strings.TrimSpace(callerEmail)) {
		return nil, errors.Forbidden("This invitation was not sent to you")
	}
	if err := service.requireLocal(ctx, invitation.WorkspaceID); err != nil {
		return nil, err
	}
	var memberCount int64
	err = service.orm.WithContext(ctx).Model(&schemas.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ?", invitation.WorkspaceID, callerID).
		Count(&memberCount).Error
	if err != nil {
		return nil, errors.Internal("failed to check the membership", err)
	}
	if memberCount > 0 {
		return nil, errors.Conflict("You are already a member of this workspace")
	}
	err = service.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&schemas.Invitation{}).Where("id = ?", id).Update("status", "accepted").Error; err != nil {
			return errors.Internal("failed to accept the invitation", err)
		}
		member := schemas.WorkspaceMember{
			WorkspaceID: invitation.WorkspaceID,
			UserID:      callerID,
			Role:        invitation.Role,
			JoinedAt:    time.Now(),
		}
		if err := tx.Create(&member).Error; err != nil {
			return errors.Internal("failed to create the membership", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	workspace, err := service.workspaceByID(ctx, invitation.WorkspaceID)
	if err != nil {
		return nil, err
	}
	return &AcceptInvitationResponse{Workspace: workspace, Message: "Invitation accepted"}, nil
}

// RejectInvitation marks a pending invitation rejected.
func (service *Service) RejectInvitation(ctx context.Context, callerID int64, callerEmail string, id int64) (*MessageResponse, error) {
	invitation, err := service.InvitationByID(ctx, callerID, callerEmail, id)
	if err != nil {
		return nil, err
	}
	if invitation.Status != "pending" {
		return nil, errors.Conflict("This invitation is no longer pending")
	}
	if service.expired(invitation) {
		return nil, errors.Conflict("This invitation has expired")
	}
	if !strings.EqualFold(strings.TrimSpace(invitation.Email), strings.TrimSpace(callerEmail)) {
		return nil, errors.Forbidden("This invitation was not sent to you")
	}
	if err := service.orm.WithContext(ctx).Model(&schemas.Invitation{}).
		Where("id = ?", id).Update("status", "rejected").Error; err != nil {
		return nil, errors.Internal("failed to reject the invitation", err)
	}
	return &MessageResponse{Message: "Invitation rejected"}, nil
}

// CancelInvitation marks a pending invitation canceled.
func (service *Service) CancelInvitation(ctx context.Context, callerID int64, id int64) (*MessageResponse, error) {
	invitation, err := service.InvitationByID(ctx, callerID, "", id)
	if err != nil {
		return nil, err
	}
	if invitation.Status != "pending" {
		return nil, errors.Conflict("This invitation is no longer pending")
	}
	if invitation.InviterID != callerID {
		if err := service.requireRole(ctx, callerID, invitation.WorkspaceID, "owner", "admin"); err != nil {
			return nil, err
		}
	}
	if err := service.orm.WithContext(ctx).Model(&schemas.Invitation{}).
		Where("id = ?", id).Update("status", "canceled").Error; err != nil {
		return nil, errors.Internal("failed to cancel the invitation", err)
	}
	return &MessageResponse{Message: "Invitation canceled"}, nil
}

func (service *Service) expired(invitation *InvitationResponse) bool {
	return invitation.ExpiresAt != nil && !time.Now().Before(*invitation.ExpiresAt)
}

func (controller *Controller) acceptInvitation(ctx context.Context, callerID int64, callerEmail string, id int64) (*AcceptInvitationResponse, error) {
	return controller.service.AcceptInvitation(ctx, callerID, callerEmail, id)
}

func (controller *Controller) rejectInvitation(ctx context.Context, callerID int64, callerEmail string, id int64) (*MessageResponse, error) {
	return controller.service.RejectInvitation(ctx, callerID, callerEmail, id)
}

func (controller *Controller) cancelInvitation(ctx context.Context, callerID int64, id int64) (*MessageResponse, error) {
	return controller.service.CancelInvitation(ctx, callerID, id)
}
