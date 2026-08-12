package workspace

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

func createUser(t *testing.T, db *gorm.DB, email, name string) int64 {
	t.Helper()
	user := schemas.User{Email: email, Name: name}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user %s: %v", email, err)
	}
	return user.ID
}

func createWorkspace(t *testing.T, db *gorm.DB, name string) int64 {
	t.Helper()
	workspace := schemas.Workspace{Name: name, Slug: strings.ToLower(name), CreatedAt: time.Now()}
	if err := db.Create(&workspace).Error; err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	return workspace.ID
}

func addMember(t *testing.T, db *gorm.DB, workspaceID, userID int64, role string) {
	t.Helper()
	member := schemas.WorkspaceMember{WorkspaceID: workspaceID, UserID: userID, Role: role, JoinedAt: time.Now()}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("create member: %v", err)
	}
}

func TestCreateWorkspace(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")

	created, err := service.Create(context.Background(), ownerID, CreateWorkspaceRequest{Name: "Studio"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == 0 || created.Name != "Studio" || created.Slug != "studio" {
		t.Fatalf("Create = %+v, want name Studio slug studio", created)
	}
	member, err := service.MemberMe(context.Background(), ownerID, created.ID)
	if err != nil {
		t.Fatalf("MemberMe: %v", err)
	}
	if member.Role != "owner" {
		t.Fatalf("creator role = %q, want owner", member.Role)
	}
}

func TestCreateWorkspaceSlugUniqued(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	ctx := context.Background()

	first, err := service.Create(ctx, ownerID, CreateWorkspaceRequest{Name: "Studio", Slug: "studio"})
	if err != nil {
		t.Fatalf("Create first: %v", err)
	}
	second, err := service.Create(ctx, ownerID, CreateWorkspaceRequest{Name: "Studio", Slug: "studio"})
	if err != nil {
		t.Fatalf("Create second: %v", err)
	}
	if first.Slug == second.Slug {
		t.Fatalf("both workspaces got slug %q, want a numbered variant", first.Slug)
	}
	if !strings.HasPrefix(second.Slug, "studio-") {
		t.Fatalf("second slug = %q, want studio-<n>", second.Slug)
	}
}

func TestCreateWorkspaceRequiresName(t *testing.T) {
	service, db := newTestService(t)
	userID := createUser(t, db, "camille@facile.studio", "Camille")

	_, err := service.Create(context.Background(), userID, CreateWorkspaceRequest{Name: "  "})
	if errors.Status(err) != 400 {
		t.Fatalf("Create with empty name status = %d, want 400", errors.Status(err))
	}
}

func TestUpdateWorkspace(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	ctx := context.Background()

	name := "Studio Deux"
	description := "the team home"
	updated, err := service.Update(ctx, ownerID, workspaceID, UpdateWorkspaceRequest{Name: &name, Description: &description})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != name || updated.Description != description {
		t.Fatalf("Update = %+v, want name %q description %q", updated, name, description)
	}

	memberID := createUser(t, db, "paul@facile.studio", "Paul")
	addMember(t, db, workspaceID, memberID, "member")
	_, err = service.Update(ctx, memberID, workspaceID, UpdateWorkspaceRequest{Name: &name})
	if errors.Status(err) != 403 {
		t.Fatalf("member Update status = %d, want 403", errors.Status(err))
	}
}

func TestUpdateWorkspaceSlugConflict(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	firstID := createWorkspace(t, db, "Studio")
	secondID := createWorkspace(t, db, "Other")
	addMember(t, db, firstID, ownerID, "owner")
	addMember(t, db, secondID, ownerID, "owner")

	slug := "studio"
	_, err := service.Update(context.Background(), ownerID, secondID, UpdateWorkspaceRequest{Slug: &slug})
	if errors.Status(err) != 409 {
		t.Fatalf("Update with taken slug status = %d, want 409", errors.Status(err))
	}
}

func TestDeleteWorkspace(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	memberID := createUser(t, db, "paul@facile.studio", "Paul")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	addMember(t, db, workspaceID, memberID, "member")
	ctx := context.Background()

	if _, err := service.Delete(ctx, memberID, workspaceID); errors.Status(err) != 403 {
		t.Fatalf("member Delete status = %d, want 403", errors.Status(err))
	}
	if _, err := service.Delete(ctx, ownerID, workspaceID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	var members int64
	if err := db.Model(&schemas.WorkspaceMember{}).Where("workspace_id = ?", workspaceID).Count(&members).Error; err != nil {
		t.Fatalf("count members: %v", err)
	}
	if members != 0 {
		t.Fatalf("Delete left %d members behind", members)
	}
}

func TestFull(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	memberID := createUser(t, db, "paul@facile.studio", "Paul")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	addMember(t, db, workspaceID, memberID, "admin")

	full, err := service.Full(context.Background(), ownerID, workspaceID)
	if err != nil {
		t.Fatalf("Full: %v", err)
	}
	if full.Workspace == nil || full.Workspace.ID != workspaceID {
		t.Fatalf("Full workspace = %+v, want %d", full.Workspace, workspaceID)
	}
	if len(full.Members) != 2 {
		t.Fatalf("Full members = %d rows, want 2", len(full.Members))
	}
}

func TestMemberMe(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")

	member, err := service.MemberMe(context.Background(), ownerID, workspaceID)
	if err != nil {
		t.Fatalf("MemberMe: %v", err)
	}
	if member.ID != ownerID || member.Role != "owner" || member.Email != "camille@facile.studio" {
		t.Fatalf("MemberMe = %+v, want the caller's own row", member)
	}

	stranger := createUser(t, db, "nobody@facile.studio", "Nobody")
	if _, err := service.MemberMe(context.Background(), stranger, workspaceID); errors.Status(err) != 403 {
		t.Fatalf("MemberMe for stranger status = %d, want 403", errors.Status(err))
	}
}

func TestUpdateMemberRole(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	adminID := createUser(t, db, "paul@facile.studio", "Paul")
	memberID := createUser(t, db, "lea@facile.studio", "Lea")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	addMember(t, db, workspaceID, adminID, "admin")
	addMember(t, db, workspaceID, memberID, "member")
	ctx := context.Background()

	updated, err := service.UpdateMemberRole(ctx, ownerID, workspaceID, memberID, "admin")
	if err != nil {
		t.Fatalf("UpdateMemberRole: %v", err)
	}
	if updated.Role != "admin" {
		t.Fatalf("updated role = %q, want admin", updated.Role)
	}
	if _, err := service.UpdateMemberRole(ctx, adminID, workspaceID, memberID, "member"); errors.Status(err) != 403 {
		t.Fatalf("admin UpdateMemberRole status = %d, want 403", errors.Status(err))
	}
	if _, err := service.UpdateMemberRole(ctx, ownerID, workspaceID, ownerID, "member"); errors.Status(err) != 400 {
		t.Fatalf("self-demote status = %d, want 400", errors.Status(err))
	}
	if _, err := service.UpdateMemberRole(ctx, ownerID, workspaceID, ownerID, "admin"); errors.Status(err) != 400 {
		t.Fatalf("last-owner demote status = %d, want 400", errors.Status(err))
	}
}

func TestRemoveMember(t *testing.T) {
	service, db := newTestService(t)
	ownerID := createUser(t, db, "camille@facile.studio", "Camille")
	memberID := createUser(t, db, "lea@facile.studio", "Lea")
	workspaceID := createWorkspace(t, db, "Studio")
	addMember(t, db, workspaceID, ownerID, "owner")
	addMember(t, db, workspaceID, memberID, "member")
	ctx := context.Background()

	if _, err := service.RemoveMember(ctx, memberID, workspaceID, strconv.FormatInt(ownerID, 10)); errors.Status(err) != 403 {
		t.Fatalf("member RemoveMember status = %d, want 403", errors.Status(err))
	}
	if _, err := service.RemoveMember(ctx, ownerID, workspaceID, "lea@facile.studio"); err != nil {
		t.Fatalf("RemoveMember by email: %v", err)
	}
	if _, err := service.RemoveMember(ctx, ownerID, workspaceID, strconv.FormatInt(ownerID, 10)); errors.Status(err) != 400 {
		t.Fatalf("last-owner remove status = %d, want 400", errors.Status(err))
	}
}