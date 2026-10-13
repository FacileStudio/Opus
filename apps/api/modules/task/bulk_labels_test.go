package task

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/Opus/apps/api/schemas"
)

func TestBulkAddLabel(t *testing.T) {
	service, db := newTestService(t)
	userID, projectID, col := seedProject(t, db)
	ctx := context.Background()

	first, err := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{Title: "L1", Status: col.Slug})
	if err != nil {
		t.Fatalf("create first task: %v", err)
	}
	second, err := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{Title: "L2", Status: col.Slug})
	if err != nil {
		t.Fatalf("create second task: %v", err)
	}

	var project schemas.Project
	if err := db.First(&project, projectID).Error; err != nil {
		t.Fatalf("load project: %v", err)
	}
	now := time.Now()
	label := schemas.Label{WorkspaceID: project.WorkspaceID, Name: "urgent-ish", Color: "#ff0000", CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&label).Error; err != nil {
		t.Fatalf("seed label: %v", err)
	}

	value := strconv.FormatInt(label.ID, 10)
	resp, err := service.controller.bulk(ctx, userID, &BulkUpdateRequest{
		TaskIDs:   wireid.FromInt64s([]int64{first.ID, second.ID}),
		Operation: "addLabel",
		Value:     &value,
	})
	if err != nil || resp.UpdatedCount != 2 {
		t.Fatalf("bulk add label: %v, resp: %+v", err, resp)
	}

	var count int64
	err = db.Model(&schemas.Label{}).
		Where("task_id IN ? AND name = ?", []int64{first.ID, second.ID}, "urgent-ish").
		Count(&count).Error
	if err != nil {
		t.Fatalf("count labels: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 task labels, got %d", count)
	}

	duplicate, err := service.controller.bulk(ctx, userID, &BulkUpdateRequest{
		TaskIDs:   wireid.FromInt64s([]int64{first.ID, second.ID}),
		Operation: "addLabel",
		Value:     &value,
	})
	if err != nil || duplicate.UpdatedCount != 2 {
		t.Fatalf("re-apply label: %v, resp: %+v", err, duplicate)
	}
	err = db.Model(&schemas.Label{}).
		Where("task_id IN ? AND name = ?", []int64{first.ID, second.ID}, "urgent-ish").
		Count(&count).Error
	if err != nil || count != 2 {
		t.Fatalf("expected still 2 task labels, got %d (err %v)", count, err)
	}
}
