package events

import (
	"context"
	"sync"
	"testing"
	"time"

	antenne "github.com/FacileStudio/antenne-client/go"
	enveloppe "github.com/FacileStudio/enveloppe/go"
)

type recordedMessage struct {
	channel string
	payload any
}

type fakeTransport struct {
	mu           sync.Mutex
	connected    bool
	messages     []recordedMessage
	emitErr      error
	dialErr      error
	disconnected bool
}

func (f *fakeTransport) Connect(_ context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.dialErr != nil {
		return f.dialErr
	}
	f.connected = true
	return nil
}

func (f *fakeTransport) EmitNow(channel string, payload any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.emitErr != nil {
		return f.emitErr
	}
	f.messages = append(f.messages, recordedMessage{channel: channel, payload: payload})
	return nil
}

func (f *fakeTransport) Disconnect() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.connected = false
	f.disconnected = true
}

func (f *fakeTransport) IsConnected() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.connected
}

func (f *fakeTransport) getMessages() []recordedMessage {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]recordedMessage, len(f.messages))
	copy(out, f.messages)
	return out
}

func TestPublisher_EmitTask_Success(t *testing.T) {
	fake := &fakeTransport{connected: true}
	cfg := Config{URL: "http://localhost:8080", Secret: "test-secret"}
	pub := New(cfg, WithTransport(fake))

	ctx := context.Background()
	taskPayload := TaskPayload{
		ID:          101,
		FacileID:    "101",
		ProjectID:   202,
		WorkspaceID: 303,
		Title:       "Implement events",
		Name:        "Implement events",
		Status:      "in-progress",
		ActorEmail:  "test@facile.studio",
	}

	err := pub.EmitTaskCreated(ctx, taskPayload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msgs := fake.getMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].channel != ChannelTaskCreated {
		t.Errorf("channel = %s, want %s", msgs[0].channel, ChannelTaskCreated)
	}

	evt, ok := msgs[0].payload.(enveloppe.Event[any])
	if !ok {
		t.Fatalf("payload is not enveloppe.Event[any]: %T", msgs[0].payload)
	}
	if evt.Version != 1 {
		t.Errorf("version = %d, want 1", evt.Version)
	}
	if evt.App != enveloppe.AppOpus {
		t.Errorf("app = %s, want %s", evt.App, enveloppe.AppOpus)
	}
	if evt.Object != enveloppe.ObjectTask {
		t.Errorf("object = %s, want %s", evt.Object, enveloppe.ObjectTask)
	}
	if evt.Action != enveloppe.ActionCreated {
		t.Errorf("action = %s, want %s", evt.Action, enveloppe.ActionCreated)
	}
	if evt.FacileID != "101" {
		t.Errorf("facile_id = %s, want 101", evt.FacileID)
	}
	if _, err := time.Parse(time.RFC3339, evt.Timestamp); err != nil {
		t.Errorf("timestamp is not RFC3339: %s", evt.Timestamp)
	}
	if evt.IdempotencyKey == "" {
		t.Error("expected non-empty idempotency key")
	}

	pubCount, dropCount := pub.Stats()
	if pubCount != 1 || dropCount != 0 {
		t.Errorf("stats: published=%d, dropped=%d, want 1, 0", pubCount, dropCount)
	}
}

func TestPublisher_AllTaskChannels(t *testing.T) {
	fake := &fakeTransport{connected: true}
	pub := New(Config{URL: "http://localhost:8080", Secret: "test"}, WithTransport(fake))
	ctx := context.Background()
	payload := TaskPayload{ID: 1, Title: "Task"}

	if err := pub.EmitTaskUpdated(ctx, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := pub.EmitTaskDeleted(ctx, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := pub.EmitTaskStatusChanged(ctx, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := pub.EmitTaskActorChanged(ctx, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msgs := fake.getMessages()
	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(msgs))
	}
	expected := []struct {
		channel string
		action  enveloppe.Action
	}{
		{ChannelTaskUpdated, enveloppe.ActionUpdated},
		{ChannelTaskDeleted, enveloppe.ActionDeleted},
		{ChannelTaskStatusChanged, enveloppe.ActionUpdated},
		{ChannelTaskActorChanged, enveloppe.ActionUpdated},
	}
	for i, exp := range expected {
		if msgs[i].channel != exp.channel {
			t.Errorf("[%d] channel = %s, want %s", i, msgs[i].channel, exp.channel)
		}
		evt := msgs[i].payload.(enveloppe.Event[any])
		if evt.Action != exp.action {
			t.Errorf("[%d] action = %s, want %s", i, evt.Action, exp.action)
		}
	}
}

func TestPublisher_EmitProject_Channels(t *testing.T) {
	fake := &fakeTransport{connected: true}
	pub := New(Config{URL: "http://localhost", Secret: "sec"}, WithTransport(fake))
	ctx := context.Background()
	payload := ProjectPayload{ID: 10, Slug: "core-api", Name: "Core API"}

	if err := pub.EmitProjectCreated(ctx, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := pub.EmitProjectUpdated(ctx, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := pub.EmitProjectDeleted(ctx, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msgs := fake.getMessages()
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	if msgs[0].channel != ChannelProjectCreated {
		t.Errorf("channel = %s, want %s", msgs[0].channel, ChannelProjectCreated)
	}
	if msgs[1].channel != ChannelProjectUpdated {
		t.Errorf("channel = %s, want %s", msgs[1].channel, ChannelProjectUpdated)
	}
	if msgs[2].channel != ChannelProjectDeleted {
		t.Errorf("channel = %s, want %s", msgs[2].channel, ChannelProjectDeleted)
	}
	evt := msgs[0].payload.(enveloppe.Event[any])
	if evt.Object != enveloppe.ObjectProject {
		t.Errorf("object = %s, want %s", evt.Object, enveloppe.ObjectProject)
	}
	if evt.FacileID != "core-api" {
		t.Errorf("facile_id = %s, want core-api", evt.FacileID)
	}
}

func TestPublisher_EmitCommentAndWorkspace(t *testing.T) {
	fake := &fakeTransport{connected: true}
	pub := New(Config{URL: "http://localhost", Secret: "sec"}, WithTransport(fake))
	ctx := context.Background()

	if err := pub.EmitCommentCreated(ctx, CommentPayload{ID: 55, TaskID: 10, Content: "Hello"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := pub.EmitWorkspaceCreated(ctx, WorkspacePayload{ID: 88, Slug: "facile-team", Name: "Facile Team"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msgs := fake.getMessages()
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	evtComment := msgs[0].payload.(enveloppe.Event[any])
	if evtComment.Object != ObjectComment {
		t.Errorf("object = %s, want %s", evtComment.Object, ObjectComment)
	}
	evtWs := msgs[1].payload.(enveloppe.Event[any])
	if evtWs.Object != ObjectWorkspace {
		t.Errorf("object = %s, want %s", evtWs.Object, ObjectWorkspace)
	}
}

func TestPublisher_FailSoft_Disconnected(t *testing.T) {
	fake := &fakeTransport{connected: false, emitErr: antenne.ErrNotConnected}
	pub := New(Config{URL: "http://localhost", Secret: "sec"}, WithTransport(fake))

	err := pub.EmitTaskCreated(context.Background(), TaskPayload{ID: 1})
	if err != nil {
		t.Fatalf("expected nil error on fail-soft, got %v", err)
	}

	pubCount, dropCount := pub.Stats()
	if pubCount != 0 || dropCount != 1 {
		t.Errorf("stats: published=%d, dropped=%d, want 0, 1", pubCount, dropCount)
	}
}

func TestPublisher_FailSoft_DisabledConfig(t *testing.T) {
	pub := New(Config{})

	err := pub.EmitTaskCreated(context.Background(), TaskPayload{ID: 1})
	if err != nil {
		t.Fatalf("expected nil error when disabled, got %v", err)
	}
	pubCount, dropCount := pub.Stats()
	if pubCount != 0 || dropCount != 0 {
		t.Errorf("stats: published=%d, dropped=%d, want 0, 0", pubCount, dropCount)
	}
}

func TestPublisher_FailSoft_NilPublisher(t *testing.T) {
	var pub *Publisher
	err := pub.EmitTaskCreated(context.Background(), TaskPayload{ID: 1})
	if err != nil {
		t.Fatalf("expected nil error on nil publisher, got %v", err)
	}
	pub.Close()
	pubCount, dropCount := pub.Stats()
	if pubCount != 0 || dropCount != 0 {
		t.Errorf("stats: published=%d, dropped=%d, want 0, 0", pubCount, dropCount)
	}
}
