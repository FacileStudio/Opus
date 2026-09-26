// Package project manages project workspaces, columns, and workflows.
package project

import (
	"context"
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
	models := []any{
		&schemas.User{},
		&schemas.Workspace{},
		&schemas.WorkspaceMember{},
		&schemas.Project{},
		&schemas.Column{},
		&schemas.Task{},
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewService(db), db
}

func seedWorkspaceMember(t *testing.T, db *gorm.DB) (int64, int64) {
	t.Helper()
	user := schemas.User{Email: "alice@example.com", Name: "Alice"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	workspace := schemas.Workspace{Name: "Workspace 1", Slug: "ws-1", CreatedAt: time.Now()}
	if err := db.Create(&workspace).Error; err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	member := schemas.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      user.ID,
		Role:        "owner",
		JoinedAt:    time.Now(),
	}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("seed member: %v", err)
	}
	return user.ID, workspace.ID
}

func TestCreateProject(t *testing.T) {
	service, db := newTestService(t)
	userID, workspaceID := seedWorkspaceMember(t, db)
	ctx := context.Background()

	p := schemas.Project{
		WorkspaceID: workspaceID,
		Name:        "Test Project",
		Slug:        "test-project",
		Icon:        "Folder",
	}
	created, err := service.Create(ctx, userID, p)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == 0 || created.Slug != "test-project" {
		t.Fatalf("unexpected project: %+v", created)
	}

	var cols []schemas.Column
	if err := db.Where("project_id = ?", created.ID).Order("position asc").Find(&cols).Error; err != nil {
		t.Fatalf("find columns: %v", err)
	}
	if len(cols) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(cols))
	}
	if cols[0].Name != "To Do" || cols[3].Name != "Done" || !cols[3].IsFinal {
		t.Fatalf("unexpected default columns: %+v", cols)
	}
}

func TestListProjects(t *testing.T) {
	service, db := newTestService(t)
	userID, workspaceID := seedWorkspaceMember(t, db)
	ctx := context.Background()

	p1, err := service.Create(ctx, userID, schemas.Project{WorkspaceID: workspaceID, Name: "P1", Slug: "p1"})
	if err != nil {
		t.Fatalf("create p1: %v", err)
	}
	p2, err := service.Create(ctx, userID, schemas.Project{WorkspaceID: workspaceID, Name: "P2", Slug: "p2"})
	if err != nil {
		t.Fatalf("create p2: %v", err)
	}
	if _, err := service.Archive(ctx, userID, p2.ID); err != nil {
		t.Fatalf("archive p2: %v", err)
	}

	task := schemas.Task{
		ProjectID:   p1.ID,
		WorkspaceID: workspaceID,
		Title:       "Task 1",
		Status:      "done",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}

	active, stats, err := service.ListByWorkspace(ctx, userID, workspaceID, false)
	if err != nil {
		t.Fatalf("list active: %v", err)
	}
	if len(active) != 1 || active[0].ID != p1.ID {
		t.Fatalf("expected 1 active project, got %d", len(active))
	}
	p1Stats := stats[p1.ID]
	if p1Stats.TotalTasks != 1 || p1Stats.CompletionPercentage != 100 {
		t.Fatalf("unexpected stats: %+v", p1Stats)
	}

	all, _, err := service.ListByWorkspace(ctx, userID, workspaceID, true)
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 projects including archived, got %d", len(all))
	}
}

func TestGetByID(t *testing.T) {
	service, db := newTestService(t)
	userID, workspaceID := seedWorkspaceMember(t, db)
	ctx := context.Background()

	created, err := service.Create(ctx, userID, schemas.Project{WorkspaceID: workspaceID, Name: "P", Slug: "p"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	found, stats, err := service.GetByID(ctx, userID, created.ID)
	if err != nil || found.ID != created.ID || stats == nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	_, _, err = service.GetByID(ctx, userID, 99999)
	if err == nil || errors.Status(err) != 404 {
		t.Fatalf("expected 404, got %v", err)
	}

	_, _, err = service.GetByID(ctx, userID+100, created.ID)
	if err == nil || errors.Status(err) != 403 {
		t.Fatalf("expected 403, got %v", err)
	}
}

func TestUpdateProject(t *testing.T) {
	service, db := newTestService(t)
	userID, workspaceID := seedWorkspaceMember(t, db)
	ctx := context.Background()

	created, err := service.Create(ctx, userID, schemas.Project{WorkspaceID: workspaceID, Name: "P", Slug: "p"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := service.Update(ctx, userID, created.ID, map[string]any{"name": "Updated", "description": "A desc"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Updated" || updated.Description != "A desc" {
		t.Fatalf("unexpected updated project: %+v", updated)
	}
}

func TestDeleteProject(t *testing.T) {
	service, db := newTestService(t)
	userID, workspaceID := seedWorkspaceMember(t, db)
	ctx := context.Background()

	created, err := service.Create(ctx, userID, schemas.Project{WorkspaceID: workspaceID, Name: "P", Slug: "p"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := service.Delete(ctx, userID, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	var count int64
	db.Model(&schemas.Project{}).Where("id = ?", created.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected 0 projects, got %d", count)
	}

	db.Model(&schemas.Column{}).Where("project_id = ?", created.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected 0 columns, got %d", count)
	}
}

func TestArchiveAndUnarchive(t *testing.T) {
	service, db := newTestService(t)
	userID, workspaceID := seedWorkspaceMember(t, db)
	ctx := context.Background()

	created, err := service.Create(ctx, userID, schemas.Project{WorkspaceID: workspaceID, Name: "P", Slug: "p"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	archived, err := service.Archive(ctx, userID, created.ID)
	if err != nil || archived.ArchivedAt == nil {
		t.Fatalf("archive failed: %v", err)
	}

	unarchived, err := service.Unarchive(ctx, userID, created.ID)
	if err != nil || unarchived.ArchivedAt != nil {
		t.Fatalf("unarchive failed: %v", err)
	}
}
