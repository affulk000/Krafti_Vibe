package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *Router) setupNotificationRoutes(api fiber.Router) {
	// Initialize service and handler
	notificationService := service.NewNotificationService(r.repos, r.config.Logger)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// Create notifications group
	notifications := api.Group("/notifications")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger = zap.NewNop()
		}
		notifications.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core Notification Operations
	// ============================================================================

	// Send notification (authenticated, requires notification:write scope)
	notifications.Post("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.NotificationWrite),
		notificationHandler.SendNotification,
	)

	// Get user notifications (authenticated, requires notification:read scope)
	notifications.Get("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.NotificationRead),
		notificationHandler.GetUserNotifications,
	)

	// Get unread notification count (authenticated, requires notification:read scope)
	notifications.Get("/unread-count",
		authMiddleware,
		middleware.RequireScopes(r.scopes.NotificationRead),
		notificationHandler.GetUnreadCount,
	)

	// ============================================================================
	// Notification Actions
	// ============================================================================

	// Mark notification as read (authenticated, requires notification:write scope)
	notifications.Post("/:id/read",
		authMiddleware,
		middleware.RequireScopes(r.scopes.NotificationWrite),
		notificationHandler.MarkAsRead,
	)

	// Mark all notifications as read (authenticated, requires notification:write scope)
	notifications.Post("/read-all",
		authMiddleware,
		middleware.RequireScopes(r.scopes.NotificationWrite),
		notificationHandler.MarkAllAsRead,
	)
}
