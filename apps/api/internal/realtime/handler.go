package realtime

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// ServeWS handles WebSocket upgrade requests and registers the connection.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	projectID, workspaceID := extractChannels(r)
	if projectID == "" && workspaceID == "" {
		http.Error(w, "missing project or workspace channel", http.StatusBadRequest)
		return
	}
	userID, err := h.authenticateRequest(w, r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	initiatorID := extractInitiatorID(r)
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", slog.Any("error", err))
		return
	}
	client := newClient(clientConfig{
		hub:         h,
		conn:        conn,
		projectID:   projectID,
		workspaceID: workspaceID,
		initiatorID: initiatorID,
		userID:      userID,
	})
	h.Register(client)
	go client.writePump()
	go client.readPump()
}

func (h *Hub) authenticateRequest(w http.ResponseWriter, r *http.Request) (int64, error) {
	if h.authenticator == nil {
		return 0, nil
	}
	req := prepareAuthRequest(r)
	return h.authenticator.Authenticate(w, req)
}

func extractInitiatorID(r *http.Request) string {
	for _, key := range []string{"windowId", "initiatorId"} {
		if id := r.URL.Query().Get(key); id != "" {
			return id
		}
	}
	return r.Header.Get("X-Opus-Window-Id")
}

func extractChannels(r *http.Request) (string, string) {
	projectID := extractProjectID(r)
	workspaceID := extractWorkspaceID(r)
	return projectID, workspaceID
}

func extractProjectID(r *http.Request) string {
	for _, key := range []string{"projectId", "project_id"} {
		if id := chi.URLParam(r, key); id != "" {
			return id
		}
		if id := r.URL.Query().Get(key); id != "" {
			return id
		}
	}
	return parseProjectFromPath(r.URL.Path)
}

func extractWorkspaceID(r *http.Request) string {
	for _, key := range []string{"workspaceId", "workspace_id"} {
		if id := chi.URLParam(r, key); id != "" {
			return id
		}
		if id := r.URL.Query().Get(key); id != "" {
			return id
		}
	}
	return ""
}

func parseProjectFromPath(path string) string {
	cleaned := strings.TrimPrefix(path, "/")
	parts := strings.Split(cleaned, "/")
	if len(parts) >= 2 && parts[len(parts)-2] == "ws" {
		return parts[len(parts)-1]
	}
	return ""
}
