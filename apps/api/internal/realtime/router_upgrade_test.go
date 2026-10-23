package realtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/tronc/httpx"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

func TestUpgradeThroughSharedRouterMiddleware(t *testing.T) {
	hub, err := NewHub(Config{Authenticator: fakeAuthenticator()})
	if err != nil {
		t.Fatalf("failed to create hub: %v", err)
	}
	router := httpx.NewRouter(httpx.Config{
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins:   []string{"https://opus.example.test"},
			AllowCredentials: true,
		},
	})
	router.Route("/api", func(api chi.Router) {
		api.Get("/ws/{projectId}", hub.ServeWS)
	})
	server := httptest.NewServer(router)
	t.Cleanup(func() {
		server.Close()
		hub.Close()
	})

	headers := http.Header{}
	headers.Set("Cookie", "session=cookie-token")
	headers.Set("Origin", "https://opus.example.test")

	conn := dialWS(t, wsURL(server.URL, "/api/ws/4?windowId=w1"), headers)
	defer conn.Close()

	if err := waitForCount(func() int { return hub.ProjectConnectionCount("4") }, 1); err != nil {
		t.Fatalf("expected connection through router middleware: %v", err)
	}
}

func TestUpgradeRefusedWhenAuthorizerDenies(t *testing.T) {
	hub, err := NewHub(Config{
		Authenticator: fakeAuthenticator(),
		Authorizer: func(_ context.Context, userID int64, projectID, _ string) bool {
			return userID == 42 && projectID == "4"
		},
	})
	if err != nil {
		t.Fatalf("failed to create hub: %v", err)
	}
	router := chi.NewRouter()
	router.Get("/api/ws/{projectId}", hub.ServeWS)
	server := httptest.NewServer(router)
	t.Cleanup(func() {
		server.Close()
		hub.Close()
	})

	headers := http.Header{}
	headers.Set("Cookie", "session=cookie-token")

	conn := dialWS(t, wsURL(server.URL, "/api/ws/4"), headers)
	conn.Close()

	_, resp, err := websocket.DefaultDialer.Dial(wsURL(server.URL, "/api/ws/5"), headers)
	if err == nil {
		t.Fatalf("expected the handshake for a foreign project to fail")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for a foreign project, got %v", resp)
	}
}

type recordingPublisher struct {
	projects []string
	messages []BroadcastMessage
	excluded []string
}

func (p *recordingPublisher) BroadcastToProject(projectID string, msg BroadcastMessage, exclude string) error {
	p.projects = append(p.projects, projectID)
	p.messages = append(p.messages, msg)
	p.excluded = append(p.excluded, exclude)
	return nil
}

func TestNotifySkipsTheInitiatingWindow(t *testing.T) {
	publisher := &recordingPublisher{}
	handler := InitiatorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Notify(r.Context(), publisher, TypeTaskUpdated, 7, 9)
	}))
	req := httptest.NewRequest(http.MethodPut, "/task/title/9", nil)
	req.Header.Set("X-Opus-Window-Id", "w1")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if len(publisher.messages) != 1 || publisher.projects[0] != "7" || publisher.messages[0].TaskID != "9" || publisher.excluded[0] != "w1" {
		t.Fatalf("unexpected broadcast: %+v %+v %+v", publisher.projects, publisher.messages, publisher.excluded)
	}
}
