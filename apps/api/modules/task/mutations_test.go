package task

import (
	"context"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
)

func seedSecondProject(t *testing.T, db *gorm.DB, workspaceID int64) (int64, schemas.Column) {
	t.Helper()
	destProject := schemas.Project{
		WorkspaceID: workspaceID,
		Name:        "Destination",
		Slug:        "dest",
		CreatedAt:   time.Now(),
	}
	if err := db.Create(&destProject).Error; err != nil {
		t.Fatalf("seed destination project: %v", err)
	}
	destCol := schemas.Column{
		ProjectID: destProject.ID,
		Name:      "In Progress",
		Slug:      "in-progress",
		Position:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.Create(&destCol).Error; err != nil {
		t.Fatalf("seed destination column: %v", err)
	}
	return destProject.ID, destCol
}

func TestMoveTask(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID, col := seedProject(t, db)
	ctx := context.Background()

	var p schemas.Project
	if err := db.First(&p, projectID).Error; err != nil {
		t.Fatalf("load project: %v", err)
	}
	destProjectID, destCol := seedSecondProject(t, db, p.WorkspaceID)

	created, err := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{
		Title:  "Move task",
		Status: col.Slug,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	moveResp, err := service.controller.move(ctx, userID, created.ID, &MoveTaskRequest{
		DestinationProjectID: wireid.ID(destProjectID),
		DestinationStatus:    destCol.Slug,
	})
	if err != nil {
		t.Fatalf("move task: %v", err)
	}
	if moveResp.Task.ProjectID != destProjectID || moveResp.Task.Status != destCol.Slug {
		t.Fatalf("unexpected moved task: %+v", moveResp.Task)
	}
}

func TestFieldUpdates(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID, col := seedProject(t, db)
	ctx := context.Background()

	task, err := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{
		Title:  "Field update task",
		Status: col.Slug,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	pTask, err := service.controller.updatePriority(ctx, userID, task.ID, "urgent")
	if err != nil || pTask.Priority != "urgent" {
		t.Fatalf("update priority: %v, task: %+v", err, pTask)
	}

	tTask, err := service.controller.updateTitle(ctx, userID, task.ID, "Brand new title")
	if err != nil || tTask.Title != "Brand new title" {
		t.Fatalf("update title: %v, task: %+v", err, tTask)
	}

	dTask, err := service.controller.updateDescription(ctx, userID, task.ID, "Brand new desc")
	if err != nil || dTask.Description != "Brand new desc" {
		t.Fatalf("update description: %v, task: %+v", err, dTask)
	}

	now := time.Now().Truncate(time.Second)
	dueTask, err := service.controller.updateDueDate(ctx, userID, task.ID, &UpdateDueDateRequest{DueDate: &now})
	if err != nil || dueTask.DueDate == nil {
		t.Fatalf("update due date: %v, task: %+v", err, dueTask)
	}

	actorTask, err := service.controller.updateActor(ctx, userID, task.ID, &UpdateActorRequest{UserID: wireid.FromInt64Ptr(&userID)})
	if err != nil || actorTask.ActorID == nil || *actorTask.ActorID != userID {
		t.Fatalf("update actor: %v, task: %+v", err, actorTask)
	}
}

func TestBulkOperations(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID, col := seedProject(t, db)
	ctx := context.Background()

	t1, _ := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{Title: "T1", Status: col.Slug})
	t2, _ := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{Title: "T2", Status: col.Slug})

	prioVal := "high"
	bulkPrio, err := service.controller.bulk(ctx, userID, &BulkUpdateRequest{
		TaskIDs:   wireid.FromInt64s([]int64{t1.ID, t2.ID}),
		Operation: "updatePriority",
		Value:     &prioVal,
	})
	if err != nil || bulkPrio.UpdatedCount != 2 {
		t.Fatalf("bulk update priority: %v, resp: %+v", err, bulkPrio)
	}

	dueVal := "2026-11-20T12:00:00Z"
	bulkDue, err := service.controller.bulk(ctx, userID, &BulkUpdateRequest{
		TaskIDs:   wireid.FromInt64s([]int64{t1.ID, t2.ID}),
		Operation: "updateDueDate",
		Value:     &dueVal,
	})
	if err != nil || bulkDue.UpdatedCount != 2 {
		t.Fatalf("bulk update due date: %v, resp: %+v", err, bulkDue)
	}

	bulkDel, err := service.controller.bulk(ctx, userID, &BulkUpdateRequest{
		TaskIDs:   wireid.FromInt64s([]int64{t1.ID, t2.ID}),
		Operation: "delete",
	})
	if err != nil || bulkDel.UpdatedCount != 2 {
		t.Fatalf("bulk delete: %v, resp: %+v", err, bulkDel)
	}
}
