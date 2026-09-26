package realtime

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func fakeAuthenticator() Authenticator {
	return AuthFunc(func(_ http.ResponseWriter, r *http.Request) (int64, error) {
		auth := r.Header.Get("Authorization")
		if auth == "Bearer secret-token" {
			return 42, nil
		}
		cookie, err := r.Cookie("session")
		if err == nil && cookie.Value == "cookie-token" {
			return 42, nil
		}
		return 0, errors.New("unauthorized")
	})
}

func TestHubAuthSuccessWithBearerHeader(t *testing.T) {
	hub, server := newTestHub(t, fakeAuthenticator())
	defer server.Close()
	defer func() {
		_ = hub.Close()
	}()

	headers := http.Header{}
	headers.Set("Authorization", "Bearer secret-token")

	conn := dialWS(t, wsURL(server.URL, "/?projectId=p1&windowId=w1"), headers)
	defer conn.Close()

	if err := waitForCount(func() int { return hub.ProjectConnectionCount("p1") }, 1); err != nil {
		t.Fatalf("expected authenticated connection: %v", err)
	}
}

func TestHubAuthSuccessWithQueryToken(t *testing.T) {
	hub, server := newTestHub(t, fakeAuthenticator())
	defer server.Close()
	defer func() {
		_ = hub.Close()
	}()

	conn := dialWS(t, wsURL(server.URL, "/?projectId=p1&windowId=w1&token=secret-token"), nil)
	defer conn.Close()

	if err := waitForCount(func() int { return hub.ProjectConnectionCount("p1") }, 1); err != nil {
		t.Fatalf("expected authenticated connection via query param: %v", err)
	}
}

func TestHubAuthSuccessWithCookie(t *testing.T) {
	hub, server := newTestHub(t, fakeAuthenticator())
	defer server.Close()
	defer func() {
		_ = hub.Close()
	}()

	headers := http.Header{}
	headers.Set("Cookie", "session=cookie-token")

	conn := dialWS(t, wsURL(server.URL, "/?projectId=p1&windowId=w1"), headers)
	defer conn.Close()

	if err := waitForCount(func() int { return hub.ProjectConnectionCount("p1") }, 1); err != nil {
		t.Fatalf("expected authenticated connection via cookie: %v", err)
	}
}

func TestHubAuthFailure(t *testing.T) {
	hub, server := newTestHub(t, fakeAuthenticator())
	defer server.Close()
	defer func() {
		_ = hub.Close()
	}()

	dialer := websocket.Dialer{
		HandshakeTimeout: 2 * time.Second,
	}
	_, resp, err := dialer.Dial(wsURL(server.URL, "/?projectId=p1&windowId=w1"), nil)
	if err == nil {
		t.Fatalf("expected handshake failure without auth")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got status %v, want 401 Unauthorized", resp)
	}
}

func TestHubMissingChannelRejection(t *testing.T) {
	hub, server := newTestHub(t, nil)
	defer server.Close()
	defer func() {
		_ = hub.Close()
	}()

	dialer := websocket.Dialer{
		HandshakeTimeout: 2 * time.Second,
	}
	_, resp, err := dialer.Dial(wsURL(server.URL, "/?windowId=w1"), nil)
	if err == nil {
		t.Fatalf("expected rejection for missing channel")
	}
	if resp == nil || resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("got status %v, want 400 Bad Request", resp)
	}
}
