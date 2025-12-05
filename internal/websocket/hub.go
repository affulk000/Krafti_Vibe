package websocket

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Clients organized by user ID for direct messaging
	userClients map[uuid.UUID]map[*Client]bool

	// Clients organized by tenant ID for tenant-wide broadcasts
	tenantClients map[uuid.UUID]map[*Client]bool

	// Inbound messages from the clients
	broadcast chan *Message

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`
	Target    string                 `json:"target,omitempty"` // user, tenant, broadcast
	TargetID  *uuid.UUID             `json:"target_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		userClients:   make(map[uuid.UUID]map[*Client]bool),
		tenantClients: make(map[uuid.UUID]map[*Client]bool),
		broadcast:     make(chan *Message, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true

	// Register by user ID
	if client.UserID != uuid.Nil {
		if h.userClients[client.UserID] == nil {
			h.userClients[client.UserID] = make(map[*Client]bool)
		}
		h.userClients[client.UserID][client] = true
	}

	// Register by tenant ID
	if client.TenantID != uuid.Nil {
		if h.tenantClients[client.TenantID] == nil {
			h.tenantClients[client.TenantID] = make(map[*Client]bool)
		}
		h.tenantClients[client.TenantID][client] = true
	}
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)

		// Unregister from user clients
		if client.UserID != uuid.Nil {
			if clients, ok := h.userClients[client.UserID]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.userClients, client.UserID)
				}
			}
		}

		// Unregister from tenant clients
		if client.TenantID != uuid.Nil {
			if clients, ok := h.tenantClients[client.TenantID]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.tenantClients, client.TenantID)
				}
			}
		}
	}
}

// broadcastMessage broadcasts a message to appropriate clients
func (h *Hub) broadcastMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	switch message.Target {
	case "user":
		if message.TargetID != nil {
			h.sendToUser(*message.TargetID, message)
		}
	case "tenant":
		if message.TargetID != nil {
			h.sendToTenant(*message.TargetID, message)
		}
	case "broadcast":
		h.sendToAll(message)
	default:
		h.sendToAll(message)
	}
}

// sendToUser sends a message to all connections of a specific user
func (h *Hub) sendToUser(userID uuid.UUID, message *Message) {
	if clients, ok := h.userClients[userID]; ok {
		for client := range clients {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
				delete(clients, client)
			}
		}
	}
}

// sendToTenant sends a message to all connections in a tenant
func (h *Hub) sendToTenant(tenantID uuid.UUID, message *Message) {
	if clients, ok := h.tenantClients[tenantID]; ok {
		for client := range clients {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
				delete(clients, client)
			}
		}
	}
}

// sendToAll sends a message to all connected clients
func (h *Hub) sendToAll(message *Message) {
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(userID uuid.UUID, msgType string, data map[string]interface{}) {
	message := &Message{
		Type:      msgType,
		Target:    "user",
		TargetID:  &userID,
		Data:      data,
		Timestamp: currentTimestamp(),
	}
	h.broadcast <- message
}

// BroadcastToTenant sends a message to all users in a tenant
func (h *Hub) BroadcastToTenant(tenantID uuid.UUID, msgType string, data map[string]interface{}) {
	message := &Message{
		Type:      msgType,
		Target:    "tenant",
		TargetID:  &tenantID,
		Data:      data,
		Timestamp: currentTimestamp(),
	}
	h.broadcast <- message
}

// BroadcastToAll sends a message to all connected clients
func (h *Hub) BroadcastToAll(msgType string, data map[string]interface{}) {
	message := &Message{
		Type:      msgType,
		Target:    "broadcast",
		Data:      data,
		Timestamp: currentTimestamp(),
	}
	h.broadcast <- message
}

// GetConnectedUsers returns the count of connected users
func (h *Hub) GetConnectedUsers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.userClients)
}

// GetTotalConnections returns the total number of connections
func (h *Hub) GetTotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// IsUserOnline checks if a user is currently connected
func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.userClients[userID]
	return ok && len(clients) > 0
}

func currentTimestamp() int64 {
	return time.Now().Unix()
}
