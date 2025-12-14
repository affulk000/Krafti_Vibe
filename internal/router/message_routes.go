package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *Router) setupMessageRoutes(api fiber.Router) {
	// Initialize service and handler
	messageService := service.NewMessageService(r.repos, r.config.Logger)
	messageHandler := handler.NewMessageHandler(messageService)

	// Create messages group
	messages := api.Group("/messages")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger = zap.NewNop()
		}
		messages.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core Message Operations
	// ============================================================================

	// Send message (authenticated, requires message:write scope)
	messages.Post("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.MessageWrite),
		messageHandler.SendMessage,
	)

	// Mark message as read (authenticated, requires message:write scope)
	messages.Post("/:id/read",
		authMiddleware,
		middleware.RequireScopes(r.scopes.MessageWrite),
		messageHandler.MarkAsRead,
	)

	// ============================================================================
	// Conversation Operations
	// ============================================================================

	// Get conversation with another user (authenticated, requires message:read scope)
	messages.Get("/conversation",
		authMiddleware,
		middleware.RequireScopes(r.scopes.MessageRead),
		messageHandler.GetConversation,
	)

	// Get all user conversations (authenticated, requires message:read scope)
	messages.Get("/conversations",
		authMiddleware,
		middleware.RequireScopes(r.scopes.MessageRead),
		messageHandler.GetUserConversations,
	)

	// Get unread message count (authenticated, requires message:read scope)
	messages.Get("/unread-count",
		authMiddleware,
		middleware.RequireScopes(r.scopes.MessageRead),
		messageHandler.GetUnreadCount,
	)
}
