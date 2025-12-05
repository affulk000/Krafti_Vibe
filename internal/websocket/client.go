package websocket

import (
	"encoding/json"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512 KB
)

// Client represents a WebSocket client connection
type Client struct {
	// The WebSocket connection
	conn *websocket.Conn

	// Hub that manages this client
	hub *Hub

	// Buffered channel of outbound messages
	send chan *Message

	// User ID associated with this connection
	UserID uuid.UUID

	// Tenant ID associated with this connection
	TenantID uuid.UUID

	// Client metadata
	ConnectedAt time.Time
	LastPingAt  time.Time
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, hub *Hub, userID, tenantID uuid.UUID) *Client {
	return &Client{
		conn:        conn,
		hub:         hub,
		send:        make(chan *Message, 256),
		UserID:      userID,
		TenantID:    tenantID,
		ConnectedAt: time.Now(),
		LastPingAt:  time.Now(),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.LastPingAt = time.Now()
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log unexpected close error
			}
			break
		}

		// Add timestamp if not present
		if msg.Timestamp == 0 {
			msg.Timestamp = time.Now().Unix()
		}

		// Process the message (you can add custom logic here)
		c.handleMessage(&msg)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteJSON(message)
			if err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from the client
func (c *Client) handleMessage(msg *Message) {
	// Add user context to the message
	if c.UserID != uuid.Nil {
		if msg.Data == nil {
			msg.Data = make(map[string]any)
		}
		msg.Data["sender_user_id"] = c.UserID.String()
		msg.Data["sender_tenant_id"] = c.TenantID.String()
	}

	// Route the message based on type
	switch msg.Type {
	case "ping":
		// Send pong response
		c.send <- &Message{
			Type:      "pong",
			Data:      map[string]any{"timestamp": time.Now().Unix()},
			Timestamp: time.Now().Unix(),
		}

	case "typing":
		// Broadcast typing indicator to target
		if msg.TargetID != nil {
			c.hub.BroadcastToUser(*msg.TargetID, "typing", map[string]any{
				"user_id": c.UserID.String(),
				"typing":  msg.Data["typing"],
			})
		}

	case "message":
		// Handle chat messages
		c.hub.broadcast <- msg

	default:
		// Broadcast other message types
		c.hub.broadcast <- msg
	}
}

// SendJSON sends a JSON message to the client
func (c *Client) SendJSON(v any) error {
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.conn.WriteJSON(v)
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(msgType string, data map[string]any) {
	msg := &Message{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
	select {
	case c.send <- msg:
	default:
		// Channel is full, skip message
	}
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// marshalMessage converts a Message to JSON bytes
func marshalMessage(msg *Message) ([]byte, error) {
	return json.Marshal(msg)
}
