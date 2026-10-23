package task

import (
	"context"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/events"
	"github.com/FacileStudio/Opus/apps/api/schemas"
)

type recordingEmitter struct {
	channels []string
	tasks    []events.TaskPayload
}

func (r *recordingEmitter) EmitTask(_ context.Context, channel string, payload events.TaskPayload) error {
	r.channels = append(r.channels, channel)
	r.tasks = append(r.tasks, payload)
	return nil
}

func (r *recordingEmitter) EmitProject(context.Context, string, events.ProjectPayload) error {
	return nil
}

func (r *recordingEmitter) EmitComment(context.Context, string, events.CommentPayload) error {
	return nil
}

func (r *recordingEmitter) EmitWorkspace(context.Context, string, events.WorkspacePayload) error {
	return nil
}

func (r *recordingEmitter) Close() {}

func TestTaskEventsCarryFacileIDs(t *testing.T) {
	service, db := newTestService(t)
	emitter := &recordingEmitter{}
	service.SetEmitter(emitter)
	userID, projectID, _ := seedProject(t, db)
	ctx := context.Background()

	created, err := service.controller.create(ctx, userID, projectID, &CreateTaskRequest{Title: "Ship it"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := service.controller.delete(ctx, userID, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	var project schemas.Project
	if err := db.First(&project, projectID).Error; err != nil {
		t.Fatalf("load project: %v", err)
	}
	if len(emitter.tasks) != 2 {
		t.Fatalf("got %d emits, want 2", len(emitter.tasks))
	}
	first := emitter.tasks[0]
	if len(first.FacileID) != 26 || first.FacileID == "" {
		t.Errorf("task facile id not minted: %q", first.FacileID)
	}
	if first.ProjectFacileID != *project.FacileID {
		t.Errorf("project facile id = %q, want %q", first.ProjectFacileID, *project.FacileID)
	}
	if emitter.tasks[1].FacileID != first.FacileID || emitter.channels[1] != events.ChannelTaskDeleted {
		t.Errorf("delete event mismatch: %+v on %v", emitter.tasks[1], emitter.channels)
	}
}

func TestTaskEventsCarryAssigneeEmailNotEditor(t *testing.T) {
	service, db := newTestService(t)
	emitter := &recordingEmitter{}
	service.SetEmitter(emitter)
	editorID, projectID, _ := seedProject(t, db)
	assignee := schemas.User{Email: "assignee@facile.studio", Name: "Assignee"}
	if err := db.Create(&assignee).Error; err != nil {
		t.Fatalf("seed assignee: %v", err)
	}
	ctx := context.Background()

	created, err := service.controller.create(ctx, editorID, projectID, &CreateTaskRequest{Title: "Unassigned"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if got := emitter.tasks[0].ActorEmail; got != "" {
		t.Errorf("unassigned task carries actor_email %q", got)
	}
	if err := db.Model(&schemas.Task{}).Where("id = ?", created.ID).Update("actor_id", assignee.ID).Error; err != nil {
		t.Fatalf("assign: %v", err)
	}
	if _, err := service.controller.updateTitle(ctx, editorID, created.ID, "Assigned"); err != nil {
		t.Fatalf("update title: %v", err)
	}
	if got := emitter.tasks[1].ActorEmail; got != assignee.Email {
		t.Errorf("actor_email = %q, want assignee %q", got, assignee.Email)
	}
}
