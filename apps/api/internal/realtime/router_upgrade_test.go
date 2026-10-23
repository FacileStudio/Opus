package realtime

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/tronc/httpx"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/go-chi/chi/v5"
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
