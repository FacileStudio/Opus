// Invitations, the collaborative half of the workspace module: an invite is a
// row addressed to an email, and accepting it is what creates the membership.
// No invite ever auto-adds a member; every one starts pending and moves to
// accepted, rejected or canceled by an explicit action.
package workspace

import (
	"context"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// invitationRow is the internal joined shape an invitation query scans into.
// InviterID is carried for the cancel check and kept off the wire.
type invitationRow struct {
	ID            int64
	Email         string
	Role          string
	Status        string
	CreatedAt     time.Time
	ExpiresAt     *time.Time
	WorkspaceID   int64
	WorkspaceName string
	InviterID     int64
	InviterName   string
}

// InviteMember addresses a pending invitation. Owners and admins invite; the
// role defaults to member. An existing member or a duplicate pending invite
// for the same address is a conflict, never a second row.
func (service *Service) InviteMember(ctx context.Context, callerID, workspaceID int64, req InviteMemberRequest) (*InvitationResponse, error) {
	if err := service.requireRole(ctx, callerID, workspaceID, "owner", "admin"); err != nil {
		return nil, err
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.Invalid("A valid email is required")
	}
	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "member"
	}
	if !validRole(role) {
		return nil, errors.Invalid("Role must be owner, admin or member")
	}
	// Owners may grant any role; admins may not confer owner, or an admin
	// escalates to a workspace they cannot otherwise reach.
	if role == "owner" {
		if err := service.requireRole(ctx, callerID, workspaceID, "owner"); err != nil {
			return nil, err
		}
	}
	var memberCount int64
	err := service.orm.WithContext(ctx).
		Table("workspace_members").
		Joins("JOIN users ON users.id = workspace_members.user_id").
		Where("workspace_members.workspace_id = ? AND users.email = ?", workspaceID, email).
		Count(&memberCount).Error
	if err != nil {
		return nil, errors.Internal("failed to check the workspace members", err)
	}
	if memberCount > 0 {
		return nil, errors.Conflict("This person is already a member")
	}
	var existing int64
	err = service.orm.WithContext(ctx).Model(&schemas.Invitation{}).
		Where("workspace_id = ? AND email = ? AND status = ?", workspaceID, email, "pending").
		Count(&existing).Error
	if err != nil {
		return nil, errors.Internal("failed to check existing invitations", err)
	}
	if existing > 0 {
		return nil, errors.Conflict("An invitation is already pending for this email")
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	invitation := schemas.Invitation{
		WorkspaceID: workspaceID,
		Email:       email,
		Role:        role,
		Status:      "pending",
		InviterID:   callerID,
		CreatedAt:   time.Now(),
		ExpiresAt:   &expiresAt,
	}
	if err := service.orm.WithContext(ctx).Create(&invitation).Error; err != nil {
		return nil, errors.Internal("failed to create the invitation", err)
	}
	return service.InvitationByID(ctx, callerID, email, invitation.ID)
}

// WorkspaceInvitations lists every invitation for one workspace, newest first,
// for the workspace's own invite management screen.
func (service *Service) WorkspaceInvitations(ctx context.Context, callerID, workspaceID int64) ([]InvitationResponse, error) {
	if err := service.RequireMember(ctx, callerID, workspaceID); err != nil {
		return nil, err
	}
	return service.queryInvitations(ctx, "invitations.workspace_id = ?", workspaceID)
}

// UserInvitations lists the actionable pending invitations addressed to the
// caller's email.
func (service *Service) UserInvitations(ctx context.Context, email string) ([]InvitationResponse, error) {
	return service.queryInvitations(ctx, "invitations.email = ? AND invitations.status = ?", strings.ToLower(strings.TrimSpace(email)), "pending")
}

// InvitationByID loads one invitation, for the accept page and the frontend's
// getInvitation call.
func (service *Service) InvitationByID(ctx context.Context, callerID int64, callerEmail string, id int64) (*InvitationResponse, error) {
	rows, err := service.queryInvitations(ctx, "invitations.id = ?", id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.NotFound("Invitation not found")
	}
	invitation := &rows[0]
	// An invitation is readable by the person it is addressed to, or by the
	// workspace's owners/admins. Anything else is an enumeration oracle on
	// sequential int64 ids.
	if !strings.EqualFold(strings.TrimSpace(invitation.Email), strings.TrimSpace(callerEmail)) {
		if err := service.requireRole(ctx, callerID, invitation.WorkspaceID, "owner", "admin"); err != nil {
			return nil, err
		}
	}
	return invitation, nil
}

// AcceptInvitation moves a pending invitation to accepted and, in the same
// transaction, creates the membership with the invitation's role. Only the
// invited address may accept.
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

// RejectInvitation marks a pending invitation rejected. The invited address is
// the only one that can decline.
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

// CancelInvitation marks a pending invitation canceled. The inviter cancels
// their own invite; an owner or admin of the workspace can cancel any.
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

// expired reports whether a pending invitation has outlived its expires_at.
// The 7-day default set at invite time is stamped onto the row; a nil expiry
// means the row predates the window and is treated as live (the legacy
// better-auth invites it replaces enforced their own window on the accept
// path, but Go owns accept now).
func (service *Service) expired(invitation *InvitationResponse) bool {
	return invitation.ExpiresAt != nil && !time.Now().Before(*invitation.ExpiresAt)
}

// queryInvitations runs the joined invitation select under one predicate,
// shared by the list and detail reads.
func (service *Service) queryInvitations(ctx context.Context, where string, args ...any) ([]InvitationResponse, error) {
	rows := []invitationRow{}
	err := service.orm.WithContext(ctx).
		Table("invitations").
		Select("invitations.id, invitations.email, invitations.role, invitations.status, invitations.created_at, invitations.expires_at, invitations.workspace_id, invitations.inviter_id, workspaces.name AS workspace_name, users.name AS inviter_name").
		Joins("JOIN workspaces ON workspaces.id = invitations.workspace_id").
		Joins("JOIN users ON users.id = invitations.inviter_id").
		Where(where, args...).
		Order("invitations.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to query invitations", err)
	}
	out := make([]InvitationResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, invitationResponse(row))
	}
	return out, nil
}

// invitationResponse renders the internal joined row as the wire shape.
func invitationResponse(row invitationRow) InvitationResponse {
	return InvitationResponse{
		ID:            row.ID,
		Email:         row.Email,
		Role:          row.Role,
		Status:        row.Status,
		CreatedAt:     row.CreatedAt,
		ExpiresAt:     row.ExpiresAt,
		WorkspaceID:   row.WorkspaceID,
		WorkspaceName: row.WorkspaceName,
		InviterName:   row.InviterName,
		InviterID:     row.InviterID,
	}
}

// Controller methods below mirror the service one for one.

func (controller *Controller) inviteMember(ctx context.Context, callerID, workspaceID int64, req *InviteMemberRequest) (*InvitationResponse, error) {
	return controller.service.InviteMember(ctx, callerID, workspaceID, *req)
}

func (controller *Controller) workspaceInvitations(ctx context.Context, callerID, workspaceID int64) ([]InvitationResponse, error) {
	return controller.service.WorkspaceInvitations(ctx, callerID, workspaceID)
}

func (controller *Controller) userInvitations(ctx context.Context, callerEmail string) ([]InvitationResponse, error) {
	return controller.service.UserInvitations(ctx, callerEmail)
}

func (controller *Controller) invitationByID(ctx context.Context, callerID int64, callerEmail string, id int64) (*InvitationResponse, error) {
	return controller.service.InvitationByID(ctx, callerID, callerEmail, id)
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