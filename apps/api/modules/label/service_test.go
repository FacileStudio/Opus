package label

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
	err = db.AutoMigrate(
		&schemas.User{},
		&schemas.Workspace{},
		&schemas.WorkspaceMember{},
		&schemas.Project{},
		&schemas.Task{},
		&schemas.Label{},
	)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewService(db), db
}

func seedFixture(t *testing.T, db *gorm.DB) (int64, int64, int64) {
	t.Helper()
	now := time.Now()
	user := schemas.User{Email: "alex@facile.studio", Name: "Alex"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	ws := schemas.Workspace{Name: "Core Workspace", Slug: "core-ws", CreatedAt: now}
	if err := db.Create(&ws).Error; err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	member := schemas.WorkspaceMember{WorkspaceID: ws.ID, UserID: user.ID, Role: "admin", JoinedAt: now}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("seed member: %v", err)
	}
	proj := schemas.Project{WorkspaceID: ws.ID, Name: "Opus", Slug: "opus", CreatedAt: now}
	if err := db.Create(&proj).Error; err != nil {
		t.Fatalf("seed project: %v", err)
	}
	task := schemas.Task{
		ProjectID:   proj.ID,
		WorkspaceID: ws.ID,
		Number:      1,
		Title:       "First Task",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("seed task: %v", err)
	}
	return user.ID, ws.ID, task.ID
}

func TestCreateLabel(t *testing.T) {
	service, db := newTestService(t)
	userID, wsID, taskID := seedFixture(t, db)
	ctx := context.Background()

	wsLabel, err := service.CreateLabel(ctx, userID, CreateLabelRequest{
		Name:        "Bug",
		Color:       "#ff0000",
		WorkspaceID: wsID,
	})
	if err != nil || wsLabel.ID == 0 || wsLabel.TaskID != nil || wsLabel.Name != "Bug" {
		t.Fatalf("CreateLabel workspace = %v, err %v", wsLabel, err)
	}

	dupWsLabel, err := service.CreateLabel(ctx, userID, CreateLabelRequest{
		Name:        "Bug",
		Color:       "#00ff00",
		WorkspaceID: wsID,
	})
	if err != nil || dupWsLabel.ID != wsLabel.ID {
		t.Fatalf("CreateLabel duplicate workspace = %v, err %v", dupWsLabel, err)
	}

	taskLabel, err := service.CreateLabel(ctx, userID, CreateLabelRequest{
		Name:        "Feature",
		Color:       "#0000ff",
		WorkspaceID: wsID,
		TaskID:      &taskID,
	})
	if err != nil || taskLabel.ID == 0 || taskLabel.TaskID == nil || *taskLabel.TaskID != taskID {
		t.Fatalf("CreateLabel task = %v, err %v", taskLabel, err)
	}

	dupTaskLabel, err := service.CreateLabel(ctx, userID, CreateLabelRequest{
		Name:        "Feature",
		Color:       "#ffff00",
		WorkspaceID: wsID,
		TaskID:      &taskID,
	})
	if err != nil || dupTaskLabel.ID != taskLabel.ID {
		t.Fatalf("CreateLabel duplicate task = %v, err %v", dupTaskLabel, err)
	}
}

func TestGetLabel(t *testing.T) {
	service, db := newTestService(t)
	userID, wsID, _ := seedFixture(t, db)
	ctx := context.Background()

	created, err := service.CreateLabel(ctx, userID, CreateLabelRequest{
		Name:        "Urgent",
		Color:       "#f00",
		WorkspaceID: wsID,
	})
	if err != nil {
		t.Fatalf("CreateLabel: %v", err)
	}

	got, err := service.GetLabel(ctx, userID, created.ID)
	if err != nil || got.ID != created.ID {
		t.Fatalf("GetLabel = %v, err %v", got, err)
	}

	_, err = service.GetLabel(ctx, userID, 99999)
	if errors.Status(err) != 404 {
		t.Fatalf("GetLabel missing status = %d, want 404", errors.Status(err))
	}

	_, err = service.GetLabel(ctx, userID+1000, created.ID)
	if errors.Status(err) != 403 {
		t.Fatalf("GetLabel stranger status = %d, want 403", errors.Status(err))
	}
}

func TestGetLabelsByWorkspaceAndTask(t *testing.T) {
	service, db := newTestService(t)
	userID, wsID, taskID := seedFixture(t, db)
	ctx := context.Background()

	_, err := service.CreateLabel(ctx, userID, CreateLabelRequest{Name: "L1", Color: "#111", WorkspaceID: wsID})
	if err != nil {
		t.Fatalf("create L1: %v", err)
	}
	_, err = service.CreateLabel(ctx, userID, CreateLabelRequest{
		Name:        "L2",
		Color:       "#222",
		WorkspaceID: wsID,
		TaskID:      &taskID,
	})
	if err != nil {
		t.Fatalf("create L2: %v", err)
	}

	wsLabels, err := service.GetLabelsByWorkspaceID(ctx, userID, wsID)
	if err != nil || len(wsLabels) != 2 {
		t.Fatalf("GetLabelsByWorkspaceID len = %d, want 2", len(wsLabels))
	}

	taskLabels, err := service.GetLabelsByTaskID(ctx, userID, taskID)
	if err != nil || len(taskLabels) != 1 || taskLabels[0].Name != "L2" {
		t.Fatalf("GetLabelsByTaskID = %v, want 1 label L2", taskLabels)
	}

	_, err = service.GetLabelsByWorkspaceID(ctx, userID+1000, wsID)
	if errors.Status(err) != 403 {
		t.Fatalf("GetLabelsByWorkspaceID stranger status = %d, want 403", errors.Status(err))
	}
}

func TestUpdateAndAssign(t *testing.T) {
	service, db := newTestService(t)
	userID, wsID, taskID := seedFixture(t, db)
	ctx := context.Background()

	label, err := service.CreateLabel(ctx, userID, CreateLabelRequest{Name: "Old", Color: "#aaa", WorkspaceID: wsID})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := service.UpdateLabel(ctx, userID, label.ID, UpdateLabelRequest{Name: "New", Color: "#bbb"})
	if err != nil || updated.Name != "New" || updated.Color != "#bbb" {
		t.Fatalf("UpdateLabel = %v, err %v", updated, err)
	}

	assigned, err := service.AssignLabelToTask(ctx, userID, label.ID, taskID)
	if err != nil || assigned.TaskID == nil || *assigned.TaskID != taskID {
		t.Fatalf("AssignLabelToTask = %v, err %v", assigned, err)
	}

	unassigned, err := service.UnassignLabelFromTask(ctx, userID, label.ID)
	if err != nil || unassigned.TaskID != nil {
		t.Fatalf("UnassignLabelFromTask = %v, err %v", unassigned, err)
	}

	_, err = service.UnassignLabelFromTask(ctx, userID, label.ID)
	if errors.Status(err) != 400 {
		t.Fatalf("Unassign already unassigned status = %d, want 400", errors.Status(err))
	}
}

func TestDeleteLabel(t *testing.T) {
	service, db := newTestService(t)
	userID, wsID, taskID := seedFixture(t, db)
	ctx := context.Background()

	wsLabel, err := service.CreateLabel(ctx, userID, CreateLabelRequest{Name: "WsOnly", Color: "#111", WorkspaceID: wsID})
	if err != nil {
		t.Fatalf("create ws label: %v", err)
	}

	_, err = service.DeleteLabel(ctx, userID, wsLabel.ID)
	if errors.Status(err) != 400 {
		t.Fatalf("DeleteLabel on workspace label status = %d, want 400", errors.Status(err))
	}

	taskLabel, err := service.CreateLabel(ctx, userID, CreateLabelRequest{
		Name:        "TaskOnly",
		Color:       "#222",
		WorkspaceID: wsID,
		TaskID:      &taskID,
	})
	if err != nil {
		t.Fatalf("create task label: %v", err)
	}

	deleted, err := service.DeleteLabel(ctx, userID, taskLabel.ID)
	if err != nil || deleted.ID != taskLabel.ID {
		t.Fatalf("DeleteLabel = %v, err %v", deleted, err)
	}

	_, err = service.GetLabel(ctx, userID, taskLabel.ID)
	if errors.Status(err) != 404 {
		t.Fatalf("GetLabel after delete status = %d, want 404", errors.Status(err))
	}
}
