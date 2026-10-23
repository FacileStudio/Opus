package task

import (
	"context"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/Opus/apps/api/schemas"
)

func TestCreateTask(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID, col := seedProject(t, db)
	ctx := context.Background()

	now := time.Now()
	req := CreateTaskRequest{
		Title:       "Test task",
		Description: "Task description",
		Status:      col.Slug,
		Priority:    "medium",
		StartDate:   &now,
		DueDate:     &now,
	}

	task, err := service.controller.create(ctx, userID, projectID, &req)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if task.Title != "Test task" || task.Number != 1 || task.Position != 1 {
		t.Fatalf("unexpected task metadata: %+v", task)
	}

	task2, err := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{
		Title:    "Second task",
		Status:   col.Slug,
		Priority: "high",
	})
	if err != nil {
		t.Fatalf("create second task: %v", err)
	}
	if task2.Number != 2 || task2.Position != 2 {
		t.Fatalf("unexpected second task number or position: %+v", task2)
	}
}

func TestGetTask(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID, col := seedProject(t, db)
	ctx := context.Background()

	created, err := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{
		Title:  "Get task",
		Status: col.Slug,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	fetched, err := service.controller.get(ctx, userID, created.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if fetched.ID != created.ID || fetched.Title != "Get task" {
		t.Fatalf("unexpected fetched task: %+v", fetched)
	}

	if _, err := service.controller.get(ctx, userID, 9999); err == nil {
		t.Fatal("expected error for non-existent task")
	}
}

func TestListTasks(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID, col := seedProject(t, db)
	ctx := context.Background()

	for _, title := range []string{"Task 1", "Task 2", "Task 3"} {
		if _, err := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{
			Title:    title,
			Status:   col.Slug,
			Priority: "low",
		}); err != nil {
			t.Fatalf("create task %s: %v", title, err)
		}
	}

	listResp, err := service.controller.list(ctx, userID, projectID, ListTasksFilter{})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if listResp.Pagination.Total != 3 {
		t.Fatalf("expected 3 tasks, got %d", listResp.Pagination.Total)
	}
	if len(listResp.Data.Columns) != 1 || len(listResp.Data.Columns[0].Tasks) != 3 {
		t.Fatalf("expected 3 tasks in column, got %+v", listResp.Data.Columns)
	}

	page := 1
	limit := 2
	pagedResp, err := service.controller.list(ctx, userID, projectID, ListTasksFilter{
		Page:  &page,
		Limit: &limit,
	})
	if err != nil {
		t.Fatalf("list paged tasks: %v", err)
	}
	if pagedResp.Pagination.TotalPages != 2 || len(pagedResp.Data.Columns[0].Tasks) != 2 {
		t.Fatalf("unexpected paged response: %+v", pagedResp.Pagination)
	}
}

func TestUpdateTask(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID, col := seedProject(t, db)
	ctx := context.Background()

	created, err := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{
		Title:  "Original title",
		Status: col.Slug,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	updated, err := service.controller.update(ctx, userID, created.ID, &UpdateTaskRequest{
		Title:       "Updated title",
		Description: "New description",
		Status:      col.Slug,
		Priority:    "urgent",
		ProjectID:   wireid.ID(projectID),
		Position:    5,
	})
	if err != nil {
		t.Fatalf("update task: %v", err)
	}
	if updated.Title != "Updated title" || updated.Priority != "urgent" || updated.Position != 5 {
		t.Fatalf("unexpected updated task: %+v", updated)
	}
}

func TestDeleteTask(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID, col := seedProject(t, db)
	ctx := context.Background()

	created, err := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{
		Title:  "To be deleted",
		Status: col.Slug,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if err := db.Create(&schemas.Comment{
		TaskID:  created.ID,
		UserID:  userID,
		Content: "test comment",
	}).Error; err != nil {
		t.Fatalf("seed comment: %v", err)
	}

	deleted, err := service.controller.delete(ctx, userID, created.ID)
	if err != nil {
		t.Fatalf("delete task: %v", err)
	}
	if deleted.ID != created.ID {
		t.Fatalf("unexpected deleted task: %+v", deleted)
	}

	if _, err := service.controller.get(ctx, userID, created.ID); err == nil {
		t.Fatal("expected error getting deleted task")
	}
}

func TestSubtaskProgress(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID, col := seedProject(t, db)
	ctx := context.Background()

	db.Create(&schemas.Column{ProjectID: projectID, Name: "Done", Slug: "done", Position: 2, IsFinal: true, CreatedAt: time.Now(), UpdatedAt: time.Now()})

	var ids []int64
	for _, status := range []string{col.Slug, col.Slug, "done"} {
		created, err := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{Title: "t", Status: status, Priority: "low"})
		if err != nil {
			t.Fatalf("create task: %v", err)
		}
		ids = append(ids, created.ID)
	}
	for _, target := range ids[1:] {
		db.Create(&schemas.TaskRelation{SourceTaskID: ids[0], TargetTaskID: target, RelationType: "subtask", CreatedAt: time.Now()})
	}

	record, err := service.Get(ctx, userID, ids[0])
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if record.SubtaskTotal != 2 || record.SubtaskDone != 1 {
		t.Fatalf("expected 1 of 2 subtasks done, got %d of %d", record.SubtaskDone, record.SubtaskTotal)
	}
}

func TestListMineExcludesDoneAndArchived(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID, col := seedProject(t, db)
	ctx := context.Background()
	final := schemas.Column{ProjectID: projectID, Name: "Done", Slug: "done", Position: 1, IsFinal: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := db.Create(&final).Error; err != nil {
		t.Fatalf("seed final column: %v", err)
	}
	actor := wireid.ID(userID)
	for _, status := range []string{col.Slug, "done", "archived"} {
		if _, err := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{Title: status, Status: status, ActorID: &actor}); err != nil {
			t.Fatalf("create %s: %v", status, err)
		}
	}
	if _, err := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{Title: "unassigned", Status: col.Slug}); err != nil {
		t.Fatalf("create unassigned: %v", err)
	}
	var project schemas.Project
	if err := db.First(&project, projectID).Error; err != nil {
		t.Fatalf("load project: %v", err)
	}

	open, err := service.controller.listMine(ctx, userID, project.WorkspaceID, false)
	if err != nil {
		t.Fatalf("list mine: %v", err)
	}
	if len(open) != 1 || open[0].Status != col.Slug || open[0].ProjectSlug != project.Slug {
		t.Fatalf("unexpected open tasks: %+v", open)
	}
	all, err := service.controller.listMine(ctx, userID, project.WorkspaceID, true)
	if err != nil {
		t.Fatalf("list mine with done: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 tasks with includeDone, got %d", len(all))
	}
	if err := db.Model(&schemas.Project{}).Where("id = ?", projectID).Update("archived_at", time.Now()).Error; err != nil {
		t.Fatalf("archive project: %v", err)
	}
	inArchived, err := service.controller.listMine(ctx, userID, project.WorkspaceID, true)
	if err != nil {
		t.Fatalf("list mine for archived project: %v", err)
	}
	if len(inArchived) != 0 {
		t.Fatalf("expected no tasks from an archived project, got %d", len(inArchived))
	}
}
