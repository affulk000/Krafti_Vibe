package router

import (
	ws "Krafti_Vibe/internal/websocket"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func (r *Router) setupWebSocketRoutes(api fiber.Router, wsHandler *ws.Handler) {
	// Create WebSocket group
	wsGroup := api.Group("/ws")

	// Auth middleware configuration

	// ============================================================================
	// WebSocket Connection Endpoint
	// ============================================================================

	// WebSocket upgrade check middleware
	wsGroup.Use("/connect", func(c *fiber.Ctx) error {
		// Check if the client requested an upgrade to the WebSocket protocol
		if websocket.IsWebSocketUpgrade(c) {
			// Store fiber context for use in WebSocket handler
			c.Locals("fiber_ctx", c)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// WebSocket connection endpoint (authenticated)
	wsGroup.Get("/connect",
		r.RequireAuth(),
		websocket.New(wsHandler.HandleConnection, websocket.Config{
			EnableCompression: true,
		}),
	)

	// ============================================================================
	// WebSocket Management API (HTTP endpoints)
	// ============================================================================

	// Get WebSocket statistics (authenticated, requires admin scope)
	wsGroup.Get("/stats",
		r.RequireAuth(),
		wsHandler.GetStats,
	)

	// Check if a user is online (authenticated)
	wsGroup.Get("/users/:user_id/online",
		r.RequireAuth(),
		wsHandler.CheckUserOnline,
	)

	// ============================================================================
	// Message Broadcasting API (HTTP endpoints)
	// ============================================================================

	// Broadcast message to a specific user (authenticated, requires admin scope)
	wsGroup.Post("/broadcast/user",
		r.RequireAuth(),
		wsHandler.BroadcastToUser,
	)

	// Broadcast message to a tenant (authenticated, requires admin scope)
	wsGroup.Post("/broadcast/tenant",
		r.RequireAuth(),
		wsHandler.BroadcastToTenant,
	)

	// Broadcast message to all users (authenticated, requires admin scope)
	wsGroup.Post("/broadcast/all",
		r.RequireAuth(),
		wsHandler.BroadcastToAll,
	)
}
