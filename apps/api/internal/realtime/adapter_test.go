package realtime

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestInMemoryAdapterPublishSubscribe(t *testing.T) {
	adapter := NewInMemoryAdapter()
	defer func() {
		_ = adapter.Close()
	}()

	var received Envelope
	var wg sync.WaitGroup
	wg.Add(1)

	err := adapter.Subscribe(context.Background(), func(env Envelope) {
		received = env
		wg.Done()
	})
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	sent := Envelope{
		ProjectID:          "proj-123",
		ExcludeInitiatorID: "win-1",
		Message: BroadcastMessage{
			Type:      TypeTaskCreated,
			ProjectID: "proj-123",
			TaskID:    "task-456",
		},
	}

	if err := adapter.Publish(context.Background(), sent); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	wg.Wait()
	if received.ProjectID != sent.ProjectID {
		t.Fatalf("got projectID %q, want %q", received.ProjectID, sent.ProjectID)
	}
	if received.Message.TaskID != "task-456" {
		t.Fatalf("got taskID %q, want task-456", received.Message.TaskID)
	}
}

func TestInMemoryAdapterMultipleSubscribers(t *testing.T) {
	adapter := NewInMemoryAdapter()
	defer func() {
		_ = adapter.Close()
	}()

	var count int
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	handler := func(env Envelope) {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
	}

	if err := adapter.Subscribe(context.Background(), handler); err != nil {
		t.Fatalf("sub 1 failed: %v", err)
	}
	if err := adapter.Subscribe(context.Background(), handler); err != nil {
		t.Fatalf("sub 2 failed: %v", err)
	}

	if err := adapter.Publish(context.Background(), Envelope{ProjectID: "p1"}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	wg.Wait()
	mu.Lock()
	defer mu.Unlock()
	if count != 2 {
		t.Fatalf("got count %d, want 2", count)
	}
}

func TestInMemoryAdapterClose(t *testing.T) {
	adapter := NewInMemoryAdapter()
	if err := adapter.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if err := adapter.Publish(context.Background(), Envelope{}); err == nil {
		t.Fatalf("publish on closed adapter did not error")
	}
	if err := adapter.Subscribe(context.Background(), func(Envelope) {}); err == nil {
		t.Fatalf("subscribe on closed adapter did not error")
	}
}

func TestRedisAdapterFallbackEmptyURL(t *testing.T) {
	adapter := NewRedisAdapter("", nil)
	defer func() {
		_ = adapter.Close()
	}()

	if !adapter.UsingFallback() {
		t.Fatalf("expected fallback for empty redis url")
	}

	var received Envelope
	receivedCh := make(chan Envelope, 1)
	err := adapter.Subscribe(context.Background(), func(env Envelope) {
		receivedCh <- env
	})
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	testEnv := Envelope{ProjectID: "fallback-proj"}
	if err := adapter.Publish(context.Background(), testEnv); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	select {
	case received = <-receivedCh:
		if received.ProjectID != "fallback-proj" {
			t.Fatalf("got projectID %q, want fallback-proj", received.ProjectID)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for fallback message delivery")
	}
}

func TestRedisAdapterFallbackInvalidHost(t *testing.T) {
	adapter := NewRedisAdapter("redis://127.0.0.1:54321", nil)
	defer func() {
		_ = adapter.Close()
	}()

	if !adapter.UsingFallback() {
		t.Fatalf("expected fallback for unreachable redis host")
	}

	receivedCh := make(chan Envelope, 1)
	err := adapter.Subscribe(context.Background(), func(env Envelope) {
		receivedCh <- env
	})
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	if err := adapter.Publish(context.Background(), Envelope{ProjectID: "unreachable"}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	select {
	case env := <-receivedCh:
		if env.ProjectID != "unreachable" {
			t.Fatalf("got projectID %q, want unreachable", env.ProjectID)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for fallback delivery")
	}
}

func TestNewAdapterHelper(t *testing.T) {
	mem := NewAdapter("", nil)
	defer func() {
		_ = mem.Close()
	}()
	if _, ok := mem.(*InMemoryAdapter); !ok {
		t.Fatalf("expected *InMemoryAdapter for empty url")
	}

	red := NewAdapter("redis://127.0.0.1:54321", nil)
	defer func() {
		_ = red.Close()
	}()
	if _, ok := red.(*RedisAdapter); !ok {
		t.Fatalf("expected *RedisAdapter for redis url")
	}
}
