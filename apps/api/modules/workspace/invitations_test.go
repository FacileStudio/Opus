package workspace

import (
	"context"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

func TestInviteMember(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	ctx := context.Background()

	invitation, err := service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "Paul@Facile.Studio", Role: "admin"})
	if err != nil {
		t.Fatalf("InviteMember: %v", err)
	}
	if invitation.Email != "paul@facile.studio" {
		t.Fatalf("email = %q, want normalized lowercase", invitation.Email)
	}
	if invitation.Status != "pending" || invitation.Role != "admin" {
		t.Fatalf("invitation = %+v, want pending admin", invitation)
	}
	if invitation.WorkspaceName != "Studio" || invitation.InviterName != "Camille" {
		t.Fatalf("invitation join = %+v, want workspace Studio and inviter Camille", invitation)
	}
	if invitation.ExpiresAt == nil || !invitation.ExpiresAt.After(time.Now().Add(6*24*time.Hour)) {
		t.Fatalf("expiresAt = %v, want ~7 days out", invitation.ExpiresAt)
	}
}

func TestInviteMemberRoleDefaultsAndGuards(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	memberID := createUser(t, db, "lea@facile.studio", "Lea")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	addMember(t, db, workspaceID, memberID, "member")
	ctx := context.Background()

	invitation, err := service.InviteMember(ctx, memberID, workspaceID, InviteMemberRequest{Email: "paul@facile.studio"})
	if errors.Status(err) != 403 {
		t.Fatalf("member InviteMember status = %d, want 403", errors.Status(err))
	}
	invitation, err = service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "lea@facile.studio"})
	if errors.Status(err) != 409 {
		t.Fatalf("invite existing member status = %d, want 409", errors.Status(err))
	}
	if _, err := service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "not-an-email"}); errors.Status(err) != 400 {
		t.Fatalf("invite bad email status = %d, want 400", errors.Status(err))
	}
	if _, err := service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "x@y.z", Role: "superuser"}); errors.Status(err) != 400 {
		t.Fatalf("invite bad role status = %d, want 400", errors.Status(err))
	}
	if invitation != nil {
		t.Fatalf("failed invite returned %+v, want nil", invitation)
	}
	first, err := service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "paul@facile.studio"})
	if err != nil {
		t.Fatalf("first invite: %v", err)
	}
	if first.Role != "member" {
		t.Fatalf("default role = %q, want member", first.Role)
	}
	if _, err := service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "paul@facile.studio"}); errors.Status(err) != 409 {
		t.Fatalf("duplicate invite status = %d, want 409", errors.Status(err))
	}
}

func TestUserInvitations(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	ctx := context.Background()

	first, err := service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "paul@facile.studio"})
	if err != nil {
		t.Fatalf("invite paul: %v", err)
	}
	second, err := service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "lea@facile.studio"})
	if err != nil {
		t.Fatalf("invite lea: %v", err)
	}
	if _, err := service.RejectInvitation(ctx, ownerID, "lea@facile.studio", second.ID); err != nil {
		t.Fatalf("reject lea's invite: %v", err)
	}

	list, err := service.UserInvitations(ctx, "paul@facile.studio")
	if err != nil {
		t.Fatalf("UserInvitations: %v", err)
	}
	if len(list) != 1 || list[0].ID != first.ID {
		t.Fatalf("paul's list = %+v, want only paul's pending invite", list)
	}
	none, err := service.UserInvitations(ctx, "nobody@facile.studio")
	if err != nil {
		t.Fatalf("UserInvitations for stranger: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("stranger list = %+v, want none", none)
	}
}

func TestWorkspaceInvitationsListsAllStatuses(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	ctx := context.Background()

	invitation, err := service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "paul@facile.studio"})
	if err != nil {
		t.Fatalf("invite: %v", err)
	}
	if _, err := service.RejectInvitation(ctx, ownerID, "paul@facile.studio", invitation.ID); err != nil {
		t.Fatalf("reject: %v", err)
	}
	if _, err := service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "lea@facile.studio"}); err != nil {
		t.Fatalf("invite lea: %v", err)
	}

	list, err := service.WorkspaceInvitations(ctx, ownerID, workspaceID)
	if err != nil {
		t.Fatalf("WorkspaceInvitations: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("workspace invitations = %d rows, want 2 (any status)", len(list))
	}
}

func TestAcceptInvitation(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	inviteeID := createUser(t, db, "paul@facile.studio", "Paul")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	ctx := context.Background()

	invitation, err := service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "paul@facile.studio", Role: "admin"})
	if err != nil {
		t.Fatalf("invite: %v", err)
	}
	if _, err := service.AcceptInvitation(ctx, ownerID, "camille@facile.studio", invitation.ID); errors.Status(err) != 403 {
		t.Fatalf("accept as non-invitee status = %d, want 403", errors.Status(err))
	}
	result, err := service.AcceptInvitation(ctx, inviteeID, "paul@facile.studio", invitation.ID)
	if err != nil {
		t.Fatalf("AcceptInvitation: %v", err)
	}
	if result.Workspace == nil || result.Workspace.ID != workspaceID {
		t.Fatalf("accept result = %+v, want the workspace", result)
	}
	member, err := service.MemberMe(ctx, inviteeID, workspaceID)
	if err != nil {
		t.Fatalf("MemberMe after accept: %v", err)
	}
	if member.Role != "admin" {
		t.Fatalf("accepted role = %q, want admin (the invitation's role)", member.Role)
	}
	if _, err := service.AcceptInvitation(ctx, inviteeID, "paul@facile.studio", invitation.ID); errors.Status(err) != 409 {
		t.Fatalf("re-accept status = %d, want 409", errors.Status(err))
	}
}

func TestRejectInvitation(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	ctx := context.Background()

	invitation, err := service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "paul@facile.studio"})
	if err != nil {
		t.Fatalf("invite: %v", err)
	}
	if _, err := service.RejectInvitation(ctx, ownerID, "paul@facile.studio", invitation.ID); err != nil {
		t.Fatalf("RejectInvitation: %v", err)
	}
	var row schemas.Invitation
	if err := db.First(&row, invitation.ID).Error; err != nil {
		t.Fatalf("load invitation: %v", err)
	}
	if row.Status != "rejected" {
		t.Fatalf("status = %q, want rejected", row.Status)
	}
}

func TestCancelInvitation(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	adminID := createUser(t, db, "lea@facile.studio", "Lea")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	addMember(t, db, workspaceID, adminID, "admin")
	ctx := context.Background()

	invitation, err := service.InviteMember(ctx, adminID, workspaceID, InviteMemberRequest{Email: "paul@facile.studio"})
	if err != nil {
		t.Fatalf("invite: %v", err)
	}
	if _, err := service.CancelInvitation(ctx, ownerID, invitation.ID); err != nil {
		t.Fatalf("admin CancelInvitation: %v", err)
	}
	var row schemas.Invitation
	if err := db.First(&row, invitation.ID).Error; err != nil {
		t.Fatalf("load invitation: %v", err)
	}
	if row.Status != "canceled" {
		t.Fatalf("status = %q, want canceled", row.Status)
	}
	second, err := service.InviteMember(ctx, adminID, workspaceID, InviteMemberRequest{Email: "lea2@facile.studio"})
	if err != nil {
		t.Fatalf("invite second: %v", err)
	}
	if _, err := service.CancelInvitation(ctx, ownerID, second.ID); err != nil {
		t.Fatalf("owner CancelInvitation of admin's invite: %v", err)
	}
}
func TestAdminCannotInviteOwner(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	adminID := createUser(t, db, "lea@facile.studio", "Lea")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	addMember(t, db, workspaceID, adminID, "admin")
	ctx := context.Background()

	if _, err := service.InviteMember(ctx, adminID, workspaceID, InviteMemberRequest{Email: "paul@facile.studio", Role: "owner"}); errors.Status(err) != 403 {
		t.Fatalf("admin inviting owner status = %d, want 403", errors.Status(err))
	}
	invitation, err := service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "paul@facile.studio", Role: "owner"})
	if err != nil {
		t.Fatalf("owner inviting owner: %v", err)
	}
	if invitation.Role != "owner" {
		t.Fatalf("role = %q, want owner", invitation.Role)
	}
}

func TestInvitationByIDGatedToInviteeOrOwnerAdmin(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	memberID := createUser(t, db, "lea@facile.studio", "Lea")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	addMember(t, db, workspaceID, memberID, "member")
	ctx := context.Background()

	invitation, err := service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "paul@facile.studio"})
	if err != nil {
		t.Fatalf("invite: %v", err)
	}
	if _, err := service.InvitationByID(ctx, memberID, "lea@facile.studio", invitation.ID); errors.Status(err) != 403 {
		t.Fatalf("unrelated member read status = %d, want 403", errors.Status(err))
	}
	if _, err := service.InvitationByID(ctx, ownerID, "camille@facile.studio", invitation.ID); err != nil {
		t.Fatalf("owner read: %v", err)
	}
	other, err := service.InvitationByID(ctx, 0, "paul@facile.studio", invitation.ID)
	if err != nil {
		t.Fatalf("invitee-by-email read: %v", err)
	}
	if other.ID != invitation.ID {
		t.Fatalf("read id = %d, want %d", other.ID, invitation.ID)
	}
}

func TestExpiredInvitationCannotBeAccepted(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	inviteeID := createUser(t, db, "paul@facile.studio", "Paul")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	ctx := context.Background()

	invitation, err := service.InviteMember(ctx, ownerID, workspaceID, InviteMemberRequest{Email: "paul@facile.studio"})
	if err != nil {
		t.Fatalf("invite: %v", err)
	}
	past := time.Now().Add(-time.Hour)
	if err := db.Model(&schemas.Invitation{}).Where("id = ?", invitation.ID).Update("expires_at", &past).Error; err != nil {
		t.Fatalf("backdate expiry: %v", err)
	}
	if _, err := service.AcceptInvitation(ctx, inviteeID, "paul@facile.studio", invitation.ID); errors.Status(err) != 409 {
		t.Fatalf("accept expired status = %d, want 409", errors.Status(err))
	}
}
