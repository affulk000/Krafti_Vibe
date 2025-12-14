package router

import (
	"Krafti_Vibe/internal/middleware"
	ws "Krafti_Vibe/internal/websocket"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func (r *Router) setupWebSocketRoutes(api fiber.Router, wsHandler *ws.Handler) {
	// Create WebSocket group
	wsGroup := api.Group("/ws")

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

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
		authMiddleware,
		websocket.New(wsHandler.HandleConnection, websocket.Config{
			EnableCompression: true,
		}),
	)

	// ============================================================================
	// WebSocket Management API (HTTP endpoints)
	// ============================================================================

	// Get WebSocket statistics (authenticated, requires admin scope)
	wsGroup.Get("/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.AdminRead),
		wsHandler.GetStats,
	)

	// Check if a user is online (authenticated)
	wsGroup.Get("/users/:user_id/online",
		authMiddleware,
		wsHandler.CheckUserOnline,
	)

	// ============================================================================
	// Message Broadcasting API (HTTP endpoints)
	// ============================================================================

	// Broadcast message to a specific user (authenticated, requires admin scope)
	wsGroup.Post("/broadcast/user",
		authMiddleware,
		middleware.RequireScopes(r.scopes.AdminWrite),
		wsHandler.BroadcastToUser,
	)

	// Broadcast message to a tenant (authenticated, requires admin scope)
	wsGroup.Post("/broadcast/tenant",
		authMiddleware,
		middleware.RequireScopes(r.scopes.AdminWrite),
		wsHandler.BroadcastToTenant,
	)

	// Broadcast message to all users (authenticated, requires admin scope)
	wsGroup.Post("/broadcast/all",
		authMiddleware,
		middleware.RequireScopes(r.scopes.AdminFull),
		wsHandler.BroadcastToAll,
	)
}
