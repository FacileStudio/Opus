package events

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
)

func TestPayloadBuilders(t *testing.T) {
	task := schemas.Task{ID: 12, ProjectID: 34, WorkspaceID: 56, Title: "Task", Status: "open"}
	tp := NewTaskPayload(task, "alice@facile.studio")
	if tp.ID != 12 || tp.FacileID != "12" || tp.ActorEmail != "alice@facile.studio" {
		t.Errorf("unexpected task payload: %+v", tp)
	}

	proj := schemas.Project{ID: 78, Slug: "proj", Name: "Project", Description: "Desc"}
	pp := NewProjectPayload(proj, "bob@facile.studio")
	if pp.FacileID != "proj" || *pp.Description != "Desc" {
		t.Errorf("unexpected project payload: %+v", pp)
	}

	comm := schemas.Comment{ID: 99, TaskID: 12, UserID: 1, Content: "Note"}
	cp := NewCommentPayload(comm, "alice@facile.studio")
	if cp.FacileID != "99" || cp.Content != "Note" {
		t.Errorf("unexpected comment payload: %+v", cp)
	}

	ws := schemas.Workspace{ID: 100, Slug: "ws", Name: "WS"}
	wp := NewWorkspacePayload(ws, "carol@facile.studio")
	if wp.FacileID != "ws" || wp.Name != "WS" {
		t.Errorf("unexpected workspace payload: %+v", wp)
	}
}

func TestEmitterInterface(t *testing.T) {
	fake := &fakeTransport{connected: true}
	var emitter Emitter = New(Config{URL: "http://localhost", Secret: "sec"}, WithTransport(fake))
	err := emitter.EmitTask(context.Background(), ChannelTaskCreated, TaskPayload{ID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	emitter.Close()
	if !fake.disconnected {
		t.Error("expected transport to be disconnected after Close")
	}
}

func TestAllChannels(t *testing.T) {
	channels := AllChannels()
	if len(channels) != 10 {
		t.Fatalf("expected 10 channels, got %d", len(channels))
	}
}

func TestLoadConfig(t *testing.T) {
	os.Setenv("ANTENNE_URL", "http://example.com")
	os.Setenv("ANTENNE_SECRET", "supersecret")
	os.Setenv("ANTENNE_INSTANCE_ID", "inst-1")
	defer func() {
		os.Unsetenv("ANTENNE_URL")
		os.Unsetenv("ANTENNE_SECRET")
		os.Unsetenv("ANTENNE_INSTANCE_ID")
	}()

	cfg := LoadConfig()
	if cfg.URL != "http://example.com" || cfg.Secret != "supersecret" || cfg.InstanceID != "inst-1" {
		t.Errorf("unexpected config: %+v", cfg)
	}
	if !cfg.Enabled() {
		t.Error("expected config to be enabled")
	}
}

func TestPublisher_Dial_Cancel(t *testing.T) {
	fake := &fakeTransport{dialErr: errors.New("cannot connect")}
	pub := New(Config{URL: "http://localhost", Secret: "sec"}, WithTransport(fake))

	ctx, cancel := context.WithCancel(context.Background())
	pub.Start(ctx)
	cancel()

	if pub.IsConnected() {
		t.Error("expected not connected")
	}
}

func TestTaskPayloadDueDateJSON(t *testing.T) {
	due := time.Date(2026, 10, 30, 0, 0, 0, 0, time.UTC)
	task := schemas.Task{ID: 12, ProjectID: 34, WorkspaceID: 56, Title: "Task", Status: "open", DueDate: &due}

	set, err := json.Marshal(NewTaskPayload(task, ""))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(set, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["due_date"] != "2026-10-30T00:00:00Z" {
		t.Errorf("due_date = %v, want 2026-10-30T00:00:00Z", got["due_date"])
	}

	local := time.Date(2026, 10, 30, 2, 0, 0, 0, time.FixedZone("CEST", 2*3600))
	task.DueDate = &local
	if p := NewTaskPayload(task, ""); p.DueDate != "2026-10-30T00:00:00Z" {
		t.Errorf("non-UTC due date = %q, want UTC", p.DueDate)
	}

	task.DueDate = nil
	cleared, err := json.Marshal(NewTaskPayload(task, ""))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got = map[string]any{}
	if err := json.Unmarshal(cleared, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := got["due_date"]; ok {
		t.Errorf("due_date present for cleared task: %s", cleared)
	}
}
