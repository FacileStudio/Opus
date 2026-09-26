package realtime

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func newTestHub(t *testing.T, auth Authenticator) (*Hub, *httptest.Server) {
	t.Helper()
	hub, err := NewHub(Config{
		Authenticator: auth,
	})
	if err != nil {
		t.Fatalf("failed to create hub: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	t.Cleanup(func() {
		server.Close()
		hub.Close()
	})
	return hub, server
}

func wsURL(serverURL, pathAndQuery string) string {
	wsBase := "ws" + strings.TrimPrefix(serverURL, "http")
	return wsBase + pathAndQuery
}

func dialWS(t *testing.T, u string, header http.Header) *websocket.Conn {
	t.Helper()
	dialer := websocket.Dialer{
		HandshakeTimeout: 3 * time.Second,
	}
	conn, _, err := dialer.Dial(u, header)
	if err != nil {
		t.Fatalf("websocket dial %q failed: %v", u, err)
	}
	return conn
}

func TestHubRegistrationAndUnregistration(t *testing.T) {
	hub, server := newTestHub(t, nil)

	conn := dialWS(t, wsURL(server.URL, "/?projectId=p1&windowId=w1"), nil)
	defer conn.Close()

	if err := waitForCount(func() int { return hub.ProjectConnectionCount("p1") }, 1); err != nil {
		t.Fatalf("expected 1 connection: %v", err)
	}

	conn.Close()
	if err := waitForCount(func() int { return hub.ProjectConnectionCount("p1") }, 0); err != nil {
		t.Fatalf("expected 0 connections: %v", err)
	}
}

func TestHubMessageDelivery(t *testing.T) {
	hub, server := newTestHub(t, nil)

	conn := dialWS(t, wsURL(server.URL, "/?projectId=p1&windowId=w1"), nil)
	defer conn.Close()

	if err := waitForCount(func() int { return hub.ProjectConnectionCount("p1") }, 1); err != nil {
		t.Fatalf("expected connection: %v", err)
	}

	msg := BroadcastMessage{
		Type:         TypeTaskCreated,
		ProjectID:    "p1",
		TaskID:       "task-123",
		SourceTaskID: "src-1",
		TargetTaskID: "tgt-2",
	}
	if err := hub.BroadcastToProject("p1", msg, ""); err != nil {
		t.Fatalf("broadcast failed: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message failed: %v", err)
	}

	var received BroadcastMessage
	if err := json.Unmarshal(payload, &received); err != nil {
		t.Fatalf("unmarshal payload failed: %v", err)
	}
	if received.Type != TypeTaskCreated {
		t.Fatalf("got type %q, want %q", received.Type, TypeTaskCreated)
	}
	if received.TaskID != "task-123" {
		t.Fatalf("got taskID %q, want task-123", received.TaskID)
	}
}

func TestHubInitiatorExclusion(t *testing.T) {
	hub, server := newTestHub(t, nil)

	connA := dialWS(t, wsURL(server.URL, "/?projectId=p1&windowId=win-A"), nil)
	defer connA.Close()

	connB := dialWS(t, wsURL(server.URL, "/?projectId=p1&windowId=win-B"), nil)
	defer connB.Close()

	if err := waitForCount(func() int { return hub.ProjectConnectionCount("p1") }, 2); err != nil {
		t.Fatalf("expected 2 connections: %v", err)
	}

	msg := BroadcastMessage{
		Type:      TypeTaskUpdated,
		ProjectID: "p1",
		TaskID:    "task-999",
	}
	if err := hub.BroadcastToProject("p1", msg, "win-A"); err != nil {
		t.Fatalf("broadcast failed: %v", err)
	}

	connB.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, payloadB, err := connB.ReadMessage()
	if err != nil {
		t.Fatalf("connB should have received message: %v", err)
	}

	var receivedB BroadcastMessage
	if err := json.Unmarshal(payloadB, &receivedB); err != nil {
		t.Fatalf("unmarshal receivedB failed: %v", err)
	}
	if receivedB.TaskID != "task-999" {
		t.Fatalf("got taskID %q, want task-999", receivedB.TaskID)
	}

	connA.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, _, errA := connA.ReadMessage()
	if errA == nil {
		t.Fatalf("connA should have been excluded but received message")
	}
}

func TestHubWorkspaceBroadcast(t *testing.T) {
	hub, server := newTestHub(t, nil)

	conn := dialWS(t, wsURL(server.URL, "/?workspace_id=ws-test&windowId=w-ws"), nil)
	defer conn.Close()

	if err := waitForCount(func() int { return hub.WorkspaceConnectionCount("ws-test") }, 1); err != nil {
		t.Fatalf("expected workspace connection: %v", err)
	}

	msg := BroadcastMessage{
		Type:        TypeCommentUpdated,
		WorkspaceID: "ws-test",
		TaskID:      "t-comment",
	}
	if err := hub.BroadcastToWorkspace("ws-test", msg, ""); err != nil {
		t.Fatalf("broadcast failed: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message failed: %v", err)
	}

	var received BroadcastMessage
	if err := json.Unmarshal(payload, &received); err != nil {
		t.Fatalf("unmarshal payload failed: %v", err)
	}
	if received.Type != TypeCommentUpdated {
		t.Fatalf("got type %q, want %q", received.Type, TypeCommentUpdated)
	}
}

func waitForCount(fn func() int, target int) error {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if fn() == target {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return websocket.ErrCloseSent
}
