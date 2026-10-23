package project

import (
	"context"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/events"
	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
)

type recordedEmit struct {
	channel string
	payload events.ProjectPayload
}

type recordingEmitter struct {
	projects []recordedEmit
}

func (r *recordingEmitter) EmitTask(context.Context, string, events.TaskPayload) error { return nil }

func (r *recordingEmitter) EmitProject(_ context.Context, channel string, payload events.ProjectPayload) error {
	r.projects = append(r.projects, recordedEmit{channel: channel, payload: payload})
	return nil
}

func (r *recordingEmitter) EmitComment(context.Context, string, events.CommentPayload) error {
	return nil
}

func (r *recordingEmitter) EmitWorkspace(context.Context, string, events.WorkspacePayload) error {
	return nil
}

func (r *recordingEmitter) Close() {}

func TestProjectControllerEmitsLifecycleEvents(t *testing.T) {
	service, db := newTestService(t)
	emitter := &recordingEmitter{}
	service.SetEmitter(emitter)
	userID, workspaceID := seedWorkspaceMember(t, db)
	ctx := context.Background()

	created, err := service.controller.create(ctx, userID, &CreateProjectRequest{Name: "Sync me", WorkspaceID: wireid.ID(workspaceID)})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	newName := "Renamed"
	if _, err := service.controller.update(ctx, userID, created.ID, &UpdateProjectRequest{Name: &newName}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := service.controller.delete(ctx, userID, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	want := []string{events.ChannelProjectCreated, events.ChannelProjectUpdated, events.ChannelProjectDeleted}
	if len(emitter.projects) != len(want) {
		t.Fatalf("got %d emits, want %d", len(emitter.projects), len(want))
	}
	facileID := emitter.projects[0].payload.FacileID
	if len(facileID) != 26 {
		t.Fatalf("facile id not minted: %q", facileID)
	}
	for i, channel := range want {
		got := emitter.projects[i]
		if got.channel != channel || got.payload.FacileID != facileID {
			t.Errorf("emit %d = %s/%s, want %s/%s", i, got.channel, got.payload.FacileID, channel, facileID)
		}
	}
	if emitter.projects[1].payload.Name != "Renamed" || emitter.projects[0].payload.ActorEmail != "alice@example.com" {
		t.Errorf("unexpected payloads: %+v", emitter.projects)
	}
}
