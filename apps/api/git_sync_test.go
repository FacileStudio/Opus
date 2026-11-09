package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	antenne "github.com/FacileStudio/antenne-client/go"
	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/Opus/apps/api/internal/events"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testGitTransport struct {
	handlers map[string]antenne.EventHandler
}

func (t *testGitTransport) Connect(context.Context) error { return nil }
func (t *testGitTransport) EmitNow(string, any) error     { return nil }
func (t *testGitTransport) Disconnect()                   {}
func (t *testGitTransport) IsConnected() bool             { return true }
func (t *testGitTransport) Listen(channel string, handler antenne.EventHandler) func() {
	if t.handlers == nil {
		t.handlers = map[string]antenne.EventHandler{}
	}
	t.handlers[channel] = handler
	return func() {}
}

func TestAttachGitSync(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	transport := &testGitTransport{handlers: map[string]antenne.EventHandler{}}
	pub := events.New(events.Config{URL: "ws://localhost", Secret: "secret"}, events.WithTransport(transport))

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	appEnv := &env.Config{GitHubWebhookSecret: "test-secret"}

	attachGitSync(db, appEnv, pub, logger)

	if _, ok := transport.handlers[events.ChannelGitPullRequest]; !ok {
		t.Errorf("channel %s was not registered", events.ChannelGitPullRequest)
	}
	if _, ok := transport.handlers[events.ChannelGitPush]; !ok {
		t.Errorf("channel %s was not registered", events.ChannelGitPush)
	}

	prHandler := transport.handlers[events.ChannelGitPullRequest]
	raw := json.RawMessage(`{"action":"opened","pull_request":{"id":1,"number":10,"title":"feat: test"}}`)
	prHandler(raw, antenne.EventMeta{Offset: 1})
}
