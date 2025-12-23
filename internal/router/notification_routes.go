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

	// ============================================================================
	// Core Notification Operations
	// ============================================================================

	// Send notification (authenticated, requires notification:write scope)
	notifications.Post("/",
		r.zitadelMW.RequireAuth(),
		notificationHandler.SendNotification,
	)

	// Get user notifications (authenticated, requires notification:read scope)
	notifications.Get("/",
		r.zitadelMW.RequireAuth(),
		notificationHandler.GetUserNotifications,
	)

	// Get unread notification count (authenticated, requires notification:read scope)
	notifications.Get("/unread-count",
		r.zitadelMW.RequireAuth(),
		notificationHandler.GetUnreadCount,
	)

	// ============================================================================
	// Notification Actions
	// ============================================================================

	// Mark notification as read (authenticated, requires notification:write scope)
	notifications.Post("/:id/read",
		r.zitadelMW.RequireAuth(),
		notificationHandler.MarkAsRead,
	)

	// Mark all notifications as read (authenticated, requires notification:write scope)
	notifications.Post("/read-all",
		r.zitadelMW.RequireAuth(),
		notificationHandler.MarkAllAsRead,
	)

	// Get notification by ID (authenticated, requires notification:read scope)
	notifications.Get("/:id",
		r.zitadelMW.RequireAuth(),
		notificationHandler.GetNotification,
	)

	// Delete notification (authenticated, requires notification:write scope)
	notifications.Delete("/:id",
		r.zitadelMW.RequireAuth(),
		notificationHandler.DeleteNotification,
	)

	// Mark multiple notifications as read (authenticated, requires notification:write scope)
	notifications.Post("/mark-multiple-read",
		r.zitadelMW.RequireAuth(),
		notificationHandler.MarkMultipleAsRead,
	)

	// Send bulk notification (admin only)
	notifications.Post("/bulk",
		r.zitadelMW.RequireAuth(),
		middleware.RequireTenantOwnerOrAdmin(),
		notificationHandler.SendBulkNotification,
	)

	// Delete read notifications (authenticated)
	notifications.Delete("/cleanup/read",
		r.zitadelMW.RequireAuth(),
		notificationHandler.DeleteReadNotifications,
	)

	// Delete old notifications (platform admin only)
	notifications.Delete("/cleanup/old",
		r.zitadelMW.RequireAuth(),
		middleware.RequireTenantOwnerOrAdmin(),
		notificationHandler.DeleteOldNotifications,
	)
}
