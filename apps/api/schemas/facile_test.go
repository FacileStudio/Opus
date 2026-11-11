package schemas

import (
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
)

func TestFacileIDMinted(t *testing.T) {
	db := dbtest.Open(t)
	workspace := Workspace{Name: "W", Slug: "w", CreatedAt: time.Now()}
	if err := db.Create(&workspace).Error; err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	project := Project{WorkspaceID: workspace.ID, Name: "P", Slug: "p", CreatedAt: time.Now()}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}
	if project.FacileID == nil || len(*project.FacileID) != 26 {
		t.Fatalf("project facile id not minted: %v", project.FacileID)
	}
}
