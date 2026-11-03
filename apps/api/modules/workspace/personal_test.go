package workspace

import (
	"context"
	"log/slog"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
)

func TestPersonalWorkspaceIsCreatedOnceAndStaysPrivate(t *testing.T) {
	service, db := newPostgresService(t)
	ctx := context.Background()
	alice := createUser(t, db, "alice@facile.studio", "Alice")
	addIdentity(t, db, alice, "sub-alice")
	team := createWorkspace(t, db, "Team")
	addMember(t, db, team, alice, "owner")

	for range 2 {
		if _, err := service.ListForUser(ctx, alice); err != nil {
			t.Fatal(err)
		}
	}
	listed, err := service.ListForUser(ctx, alice)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 || !listed[0].Personal || listed[0].Name != personalName || listed[1].Personal {
		t.Fatalf("want Personal first then Team, got %+v", listed)
	}
	personal := listed[0].ID
	if _, err := service.Delete(ctx, alice, personal); err == nil {
		t.Fatal("deleting the personal workspace should be refused")
	}
	if _, err := service.InviteMember(ctx, alice, personal, InviteMemberRequest{Email: "bob@facile.studio"}); err == nil {
		t.Fatal("inviting into the personal workspace should be refused")
	}

	imported, err := service.importSpaces(ctx, testIssuer)
	if err != nil {
		t.Fatal(err)
	}
	if len(imported) != 1 || imported[0].Name != "Team" {
		t.Fatalf("only Team should be imported, got %+v", imported)
	}
	d := NewDirectory(service, directory.Client{Issuer: testIssuer}, slog.Default())
	snapshot := directory.Snapshot{Version: 1, Spaces: []directory.Space{{ID: "r-personal", Name: "Personal",
		Members: []directory.Member{{Subject: "sub-alice", Email: "alice@facile.studio", Role: directory.RoleOwner}}}}}
	if err := d.Apply(ctx, snapshot); err != nil {
		t.Fatal(err)
	}
	var mine schemas.Workspace
	if err := db.First(&mine, "id = ?", personal).Error; err != nil || mine.RegistreSpaceID != nil {
		t.Fatalf("a shared space named Personal must not link the personal workspace: %+v (%v)", mine, err)
	}
}
