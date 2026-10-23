package schemas

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFacileIDMinted(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(&Project{}, &Task{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	project := Project{WorkspaceID: 1, Name: "P", Slug: "p", CreatedAt: time.Now()}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}
	if project.FacileID == nil || len(*project.FacileID) != 26 {
		t.Fatalf("project facile id not minted: %v", project.FacileID)
	}
}
