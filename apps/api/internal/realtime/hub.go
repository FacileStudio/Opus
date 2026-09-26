package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Broadcaster provides the public interface for broadcasting realtime events.
type Broadcaster interface {
	BroadcastToProject(projectID string, msg BroadcastMessage, excludeInitiatorID string) error
	BroadcastToWorkspace(workspaceID string, msg BroadcastMessage, excludeInitiatorID string) error
	ServeWS(w http.ResponseWriter, r *http.Request)
}

// Config specifies the runtime settings for a Hub.
type Config struct {
	Adapter       Adapter
	Authenticator Authenticator
	Logger        *slog.Logger
	CheckOrigin   func(r *http.Request) bool
}

// Hub manages active WebSocket connections and channel subscriptions.
type Hub struct {
	mu            sync.RWMutex
	projects      map[string]map[*Client]struct{}
	workspaces    map[string]map[*Client]struct{}
	adapter       Adapter
	authenticator Authenticator
	upgrader      websocket.Upgrader
	logger        *slog.Logger
	closed        bool
}

// NewHub constructs and starts a new Hub.
func NewHub(cfg Config) (*Hub, error) {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	adapter := cfg.Adapter
	if adapter == nil {
		adapter = NewInMemoryAdapter()
	}
	hub := &Hub{
		projects:      make(map[string]map[*Client]struct{}),
		workspaces:    make(map[string]map[*Client]struct{}),
		adapter:       adapter,
		authenticator: cfg.Authenticator,
		upgrader:      buildUpgrader(cfg.CheckOrigin),
		logger:        logger,
	}
	if err := adapter.Subscribe(context.Background(), hub.deliver); err != nil {
		return nil, err
	}
	return hub, nil
}

func buildUpgrader(checkOrigin func(r *http.Request) bool) websocket.Upgrader {
	if checkOrigin == nil {
		checkOrigin = func(*http.Request) bool { return true }
	}
	return websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	}
}

// Register adds a client connection to project and workspace tracking maps.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		client.Close()
		return
	}
	h.registerProject(client)
	h.registerWorkspace(client)
}

func (h *Hub) registerProject(client *Client) {
	if client.projectID == "" {
		return
	}
	clients, ok := h.projects[client.projectID]
	if !ok {
		clients = make(map[*Client]struct{})
		h.projects[client.projectID] = clients
	}
	clients[client] = struct{}{}
}

func (h *Hub) registerWorkspace(client *Client) {
	if client.workspaceID == "" {
		return
	}
	clients, ok := h.workspaces[client.workspaceID]
	if !ok {
		clients = make(map[*Client]struct{})
		h.workspaces[client.workspaceID] = clients
	}
	clients[client] = struct{}{}
}

// Unregister removes a client connection from active tracking.
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.unregisterProject(client)
	h.unregisterWorkspace(client)
	client.Close()
}

func (h *Hub) unregisterProject(client *Client) {
	if client.projectID == "" {
		return
	}
	clients, ok := h.projects[client.projectID]
	if !ok {
		return
	}
	delete(clients, client)
	if len(clients) == 0 {
		delete(h.projects, client.projectID)
	}
}

func (h *Hub) unregisterWorkspace(client *Client) {
	if client.workspaceID == "" {
		return
	}
	clients, ok := h.workspaces[client.workspaceID]
	if !ok {
		return
	}
	delete(clients, client)
	if len(clients) == 0 {
		delete(h.workspaces, client.workspaceID)
	}
}

// BroadcastToProject dispatches a message to all clients listening on a project.
func (h *Hub) BroadcastToProject(projectID string, msg BroadcastMessage, excludeInitiatorID string) error {
	msg.ProjectID = projectID
	env := Envelope{
		ProjectID:          projectID,
		Message:            msg,
		ExcludeInitiatorID: excludeInitiatorID,
	}
	return h.adapter.Publish(context.Background(), env)
}

// BroadcastToWorkspace dispatches a message to all clients listening on a workspace.
func (h *Hub) BroadcastToWorkspace(workspaceID string, msg BroadcastMessage, excludeInitiatorID string) error {
	msg.WorkspaceID = workspaceID
	env := Envelope{
		WorkspaceID:        workspaceID,
		Message:            msg,
		ExcludeInitiatorID: excludeInitiatorID,
	}
	return h.adapter.Publish(context.Background(), env)
}

func (h *Hub) deliver(env Envelope) {
	payload, err := json.Marshal(env.Message)
	if err != nil {
		h.logger.Error("failed to encode broadcast payload", slog.Any("error", err))
		return
	}
	targets := h.collectTargets(env)
	for _, client := range targets {
		client.sendPayload(payload)
	}
}

func (h *Hub) collectTargets(env Envelope) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	seen := make(map[*Client]struct{})
	var targets []*Client
	if env.ProjectID != "" {
		targets = appendTargets(targets, h.projects[env.ProjectID], env.ExcludeInitiatorID, seen)
	}
	if env.WorkspaceID != "" {
		targets = appendTargets(targets, h.workspaces[env.WorkspaceID], env.ExcludeInitiatorID, seen)
	}
	return targets
}

func appendTargets(
	targets []*Client,
	clients map[*Client]struct{},
	excludeInitiatorID string,
	seen map[*Client]struct{},
) []*Client {
	for client := range clients {
		if _, ok := seen[client]; ok {
			continue
		}
		if excludeInitiatorID != "" && client.initiatorID == excludeInitiatorID {
			continue
		}
		seen[client] = struct{}{}
		targets = append(targets, client)
	}
	return targets
}

// ProjectConnectionCount returns active connections for a project.
func (h *Hub) ProjectConnectionCount(projectID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.projects[projectID])
}

// WorkspaceConnectionCount returns active connections for a workspace.
func (h *Hub) WorkspaceConnectionCount(workspaceID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.workspaces[workspaceID])
}

// Close terminates the hub, closing all connections and the adapter.
func (h *Hub) Close() error {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return nil
	}
	h.closed = true
	allClients := h.drainClients()
	h.mu.Unlock()

	for _, client := range allClients {
		client.Close()
	}
	if h.adapter != nil {
		return h.adapter.Close()
	}
	return nil
}

func (h *Hub) drainClients() []*Client {
	var allClients []*Client
	for _, clients := range h.projects {
		for client := range clients {
			allClients = append(allClients, client)
		}
	}
	for _, clients := range h.workspaces {
		for client := range clients {
			allClients = append(allClients, client)
		}
	}
	h.projects = make(map[string]map[*Client]struct{})
	h.workspaces = make(map[string]map[*Client]struct{})
	return allClients
}
