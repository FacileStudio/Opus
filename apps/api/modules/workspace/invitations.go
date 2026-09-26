package workspace

import (
	"context"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

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

// InviteMember addresses a pending invitation.
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

// WorkspaceInvitations lists every invitation for one workspace.
func (service *Service) WorkspaceInvitations(ctx context.Context, callerID, workspaceID int64) ([]InvitationResponse, error) {
	if err := service.RequireMember(ctx, callerID, workspaceID); err != nil {
		return nil, err
	}
	return service.queryInvitations(ctx, "invitations.workspace_id = ?", workspaceID)
}

// UserInvitations lists the actionable pending invitations addressed to the caller's email.
func (service *Service) UserInvitations(ctx context.Context, email string) ([]InvitationResponse, error) {
	return service.queryInvitations(ctx, "invitations.email = ? AND invitations.status = ?", strings.ToLower(strings.TrimSpace(email)), "pending")
}

// InvitationByID loads one invitation.
func (service *Service) InvitationByID(ctx context.Context, callerID int64, callerEmail string, id int64) (*InvitationResponse, error) {
	rows, err := service.queryInvitations(ctx, "invitations.id = ?", id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.NotFound("Invitation not found")
	}
	invitation := &rows[0]
	if !strings.EqualFold(strings.TrimSpace(invitation.Email), strings.TrimSpace(callerEmail)) {
		if err := service.requireRole(ctx, callerID, invitation.WorkspaceID, "owner", "admin"); err != nil {
			return nil, err
		}
	}
	return invitation, nil
}

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