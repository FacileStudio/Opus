package realtime

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
	sendBufferSize = 256
)

type clientConfig struct {
	hub         *Hub
	conn        *websocket.Conn
	projectID   string
	workspaceID string
	initiatorID string
	userID      int64
}

// Client represents an active WebSocket connection.
type Client struct {
	hub         *Hub
	conn        *websocket.Conn
	send        chan []byte
	projectID   string
	workspaceID string
	initiatorID string
	userID      int64
	done        chan struct{}
	closeOnce   sync.Once
}

func newClient(cfg clientConfig) *Client {
	return &Client{
		hub:         cfg.hub,
		conn:        cfg.conn,
		send:        make(chan []byte, sendBufferSize),
		projectID:   cfg.projectID,
		workspaceID: cfg.workspaceID,
		initiatorID: cfg.initiatorID,
		userID:      cfg.userID,
		done:        make(chan struct{}),
	}
}

// UserID returns the authenticated user id for this connection.
func (c *Client) UserID() int64 {
	return c.userID
}

// InitiatorID returns the client's window or initiator identifier.
func (c *Client) InitiatorID() string {
	return c.initiatorID
}

// ProjectID returns the bound project identifier.
func (c *Client) ProjectID() string {
	return c.projectID
}

// WorkspaceID returns the bound workspace identifier.
func (c *Client) WorkspaceID() string {
	return c.workspaceID
}

// Close terminates the client connection.
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.done)
		c.conn.Close()
	})
}

func (c *Client) sendPayload(payload []byte) bool {
	select {
	case <-c.done:
		return false
	default:
	}
	select {
	case c.send <- payload:
		return true
	default:
		return false
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c)
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}
