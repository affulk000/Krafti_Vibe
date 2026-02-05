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

	// ============================================================================
	// Core Message Operations
	// ============================================================================

	// Send message (authenticated, requires message:write scope)
	messages.Post("/",
		r.RequireAuth(),
		messageHandler.SendMessage,
	)

	// Mark message as read (authenticated, requires message:write scope)
	messages.Post("/:id/read",
		r.RequireAuth(),
		messageHandler.MarkAsRead,
	)

	// ============================================================================
	// Conversation Operations
	// ============================================================================

	// Get conversation with another user (authenticated, requires message:read scope)
	messages.Get("/conversation",
		r.RequireAuth(),
		messageHandler.GetConversation,
	)

	// Get all user conversations (authenticated, requires message:read scope)
	messages.Get("/conversations",
		r.RequireAuth(),
		messageHandler.GetUserConversations,
	)

	// Get unread message count (authenticated, requires message:read scope)
	messages.Get("/unread-count",
		r.RequireAuth(),
		messageHandler.GetUnreadCount,
	)
}
