package schemas

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFacileIDMintedAndBackfilled(t *testing.T) {
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
	if err := db.Exec("INSERT INTO tasks (project_id, workspace_id, title, status, priority, created_at, updated_at) VALUES (1, 1, 'old', 'to-do', 'low', ?, ?)", time.Now(), time.Now()).Error; err != nil {
		t.Fatalf("insert legacy task: %v", err)
	}
	if err := BackfillFacileIDs(db); err != nil {
		t.Fatalf("backfill: %v", err)
	}
	var task Task
	if err := db.First(&task).Error; err != nil || task.FacileID == nil || *task.FacileID == "" {
		t.Fatalf("task not backfilled: %+v %v", task, err)
	}
	var again Project
	db.First(&again)
	if *again.FacileID != *project.FacileID {
		t.Error("backfill rewrote an existing facile id")
	}
}
