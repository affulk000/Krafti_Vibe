package websocket

import (
	"Krafti_Vibe/internal/middleware"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler handles WebSocket connections
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

// HandleConnection handles WebSocket connections
func (h *Handler) HandleConnection(c *websocket.Conn) {
	// Get authentication context from fiber locals
	// The auth middleware should have run before this
	fiberCtx := c.Locals("fiber_ctx")
	if fiberCtx == nil {
		c.WriteJSON(fiber.Map{
			"error": "Authentication required",
		})
		c.Close()
		return
	}

	ctx, ok := fiberCtx.(*fiber.Ctx)
	if !ok {
		c.WriteJSON(fiber.Map{
			"error": "Invalid context",
		})
		c.Close()
		return
	}

	// Get auth context
	authCtx := middleware.GetAuthContext(ctx)
	if authCtx == nil {
		c.WriteJSON(fiber.Map{
			"error": "Unauthorized",
		})
		c.Close()
		return
	}

	// Create new client
	client := NewClient(c, h.hub, authCtx.UserID, authCtx.TenantID)

	// Register client with hub
	h.hub.register <- client

	// Send welcome message
	client.SendMessage("connected", map[string]interface{}{
		"message":   "Successfully connected to WebSocket",
		"user_id":   authCtx.UserID.String(),
		"tenant_id": authCtx.TenantID.String(),
	})

	// Start reading and writing
	go client.WritePump()
	client.ReadPump()
}

// GetStats returns WebSocket statistics
func (h *Handler) GetStats(c *fiber.Ctx) error {
	stats := fiber.Map{
		"connected_users":   h.hub.GetConnectedUsers(),
		"total_connections": h.hub.GetTotalConnections(),
	}
	return c.JSON(stats)
}

// CheckUserOnline checks if a user is online
func (h *Handler) CheckUserOnline(c *fiber.Ctx) error {
	userIDStr := c.Params("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	isOnline := h.hub.IsUserOnline(userID)
	return c.JSON(fiber.Map{
		"user_id":   userID.String(),
		"is_online": isOnline,
	})
}

// BroadcastToUser sends a message to a specific user
func (h *Handler) BroadcastToUser(c *fiber.Ctx) error {
	var req struct {
		UserID uuid.UUID              `json:"user_id"`
		Type   string                 `json:"type"`
		Data   map[string]interface{} `json:"data"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	h.hub.BroadcastToUser(req.UserID, req.Type, req.Data)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Message sent",
	})
}

// BroadcastToTenant sends a message to all users in a tenant
func (h *Handler) BroadcastToTenant(c *fiber.Ctx) error {
	var req struct {
		TenantID uuid.UUID              `json:"tenant_id"`
		Type     string                 `json:"type"`
		Data     map[string]interface{} `json:"data"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	h.hub.BroadcastToTenant(req.TenantID, req.Type, req.Data)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Message broadcast to tenant",
	})
}

// BroadcastToAll sends a message to all connected clients
func (h *Handler) BroadcastToAll(c *fiber.Ctx) error {
	var req struct {
		Type string                 `json:"type"`
		Data map[string]interface{} `json:"data"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	h.hub.BroadcastToAll(req.Type, req.Data)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Message broadcast to all",
	})
}

// GetHub returns the WebSocket hub
func (h *Handler) GetHub() *Hub {
	return h.hub
}
