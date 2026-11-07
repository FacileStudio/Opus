package workspace

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&schemas.User{}, &schemas.Workspace{}, &schemas.WorkspaceMember{}, &schemas.Invitation{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	service := NewService(db, t.TempDir())
	return service, db
}

func seedMember(t *testing.T, db *gorm.DB) (int64, int64) {
	t.Helper()
	user := schemas.User{Email: "camille@facile.studio", Name: "Camille"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	workspace := schemas.Workspace{Name: "Studio", Slug: "studio", CreatedAt: time.Now()}
	if err := db.Create(&workspace).Error; err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	member := schemas.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      user.ID,
		Role:        "admin",
		JoinedAt:    time.Now(),
	}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("seed member: %v", err)
	}
	return user.ID, workspace.ID
}

func TestListForUser(t *testing.T) {
	service, db := newTestService(t)
	userID, workspaceID := seedMember(t, db)

	workspaces, err := service.ListForUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListForUser: %v", err)
	}
	if len(workspaces) != 2 || !workspaces[0].Personal || workspaces[1].ID != workspaceID {
		t.Fatalf("ListForUser = %+v, want the personal workspace then %d", workspaces, workspaceID)
	}

	other, err := service.ListForUser(context.Background(), userID+1000)
	if err != nil {
		t.Fatalf("ListForUser for stranger: %v", err)
	}
	if len(other) != 0 {
		t.Fatalf("ListForUser for stranger = %+v, want none", other)
	}
}

func TestRequireMember(t *testing.T) {
	service, db := newTestService(t)
	userID, workspaceID := seedMember(t, db)
	ctx := context.Background()

	if err := service.RequireMember(ctx, userID, workspaceID); err != nil {
		t.Fatalf("RequireMember for member: %v", err)
	}
	err := service.RequireMember(ctx, userID+1000, workspaceID)
	if err == nil {
		t.Fatal("RequireMember for stranger succeeded, want Forbidden")
	}
	if errors.Status(err) != 403 {
		t.Fatalf("RequireMember for stranger status = %d, want 403", errors.Status(err))
	}
}

func TestMembers(t *testing.T) {
	service, db := newTestService(t)
	userID, workspaceID := seedMember(t, db)

	members, err := service.Members(context.Background(), workspaceID)
	if err != nil {
		t.Fatalf("Members: %v", err)
	}
	if len(members) != 1 || members[0].ID != userID || members[0].Role != "admin" {
		t.Fatalf("Members = %+v, want one admin member %d", members, userID)
	}
}

func TestLogoRoundTrip(t *testing.T) {
	service, db := newTestService(t)
	_, workspaceID := seedMember(t, db)
	ctx := context.Background()

	url, err := service.SetLogo(ctx, workspaceID, []byte("fake-png-bytes"), "png")
	if err != nil {
		t.Fatalf("SetLogo: %v", err)
	}
	if !strings.HasPrefix(url, "/api/workspace/") || !strings.Contains(url, "/logo?v=") {
		t.Fatalf("SetLogo url = %q, want cache-busted workspace url", url)
	}

	var workspace schemas.Workspace
	if err := db.First(&workspace, workspaceID).Error; err != nil {
		t.Fatalf("load workspace: %v", err)
	}
	if workspace.Logo == "" {
		t.Fatal("SetLogo left the logo column empty")
	}

	logo, err := service.Logo(ctx, workspaceID)
	if err != nil || logo == nil || string(logo.Data) != "fake-png-bytes" || logo.Extension != "png" {
		t.Fatalf("Logo = %v, %v; want stored bytes", logo, err)
	}

	if err := service.ClearLogo(ctx, workspaceID); err != nil {
		t.Fatalf("ClearLogo: %v", err)
	}
	logo, err = service.Logo(ctx, workspaceID)
	if err != nil || logo != nil {
		t.Fatalf("Logo after ClearLogo: logo=%v err=%v, want gone", logo, err)
	}
	if err := db.First(&workspace, workspaceID).Error; err != nil {
		t.Fatalf("reload workspace: %v", err)
	}
	if workspace.Logo != "" {
		t.Fatalf("ClearLogo left the logo column as %q", workspace.Logo)
	}
}
