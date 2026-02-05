package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
)

func (r *Router) setupWebhookRoutes(api fiber.Router) {
	// Initialize webhook service and handler
	webhookService := service.NewWebhookRepository(r.repos, r.config.Logger)
	webhookHandler := handler.NewWebhookHandler(webhookService)

	// Create webhook routes group
	webhooks := api.Group("/webhooks")

	// Auth middleware configuration

	// ============================================================================
	// Webhook Event Management
	// ============================================================================

	// Create webhook event
	webhooks.Post("",
		r.RequireAuth(),
		webhookHandler.CreateWebhookEvent,
	)

	// Get webhook event by ID
	webhooks.Get("/:id",
		r.RequireAuth(),
		webhookHandler.GetWebhookEvent,
	)

	// List webhook events
	webhooks.Post("/list",
		r.RequireAuth(),
		webhookHandler.ListWebhookEvents,
	)

	// ============================================================================
	// Delivery Operations
	// ============================================================================

	// Deliver webhook (manual trigger)
	webhooks.Post("/:id/deliver",
		r.RequireAuth(),
		webhookHandler.DeliverWebhook,
	)

	// Retry webhook
	webhooks.Post("/:id/retry",
		r.RequireAuth(),
		webhookHandler.RetryWebhook,
	)

	// Retry failed webhooks for a tenant
	webhooks.Post("/tenant/:tenantId/retry-failed",
		r.RequireAuth(),
		webhookHandler.RetryFailedWebhooks,
	)

	// Bulk retry webhooks
	webhooks.Post("/bulk-retry",
		r.RequireAuth(),
		webhookHandler.BulkRetryWebhooks,
	)

	// ============================================================================
	// Query Operations
	// ============================================================================

	// Get pending webhooks
	webhooks.Get("/pending",
		r.RequireAuth(),
		webhookHandler.GetPendingWebhooks,
	)

	// Get failed webhooks
	webhooks.Get("/failed",
		r.RequireAuth(),
		webhookHandler.GetFailedWebhooks,
	)

	// Get delivered webhooks
	webhooks.Get("/delivered",
		r.RequireAuth(),
		webhookHandler.GetDeliveredWebhooks,
	)

	// Get recent webhooks
	webhooks.Get("/recent",
		r.RequireAuth(),
		webhookHandler.GetRecentWebhooks,
	)

	// ============================================================================
	// Analytics
	// ============================================================================

	// Get webhook statistics
	webhooks.Get("/stats",
		r.RequireAuth(),
		webhookHandler.GetWebhookStats,
	)

	// Get webhook analytics
	webhooks.Get("/analytics",
		r.RequireAuth(),
		webhookHandler.GetWebhookAnalytics,
	)

	// Get failure reasons
	webhooks.Get("/failure-reasons",
		r.RequireAuth(),
		webhookHandler.GetFailureReasons,
	)

	// ============================================================================
	// Cleanup Operations (Admin/Manage access required)
	// ============================================================================

	// Cleanup old webhooks
	webhooks.Post("/cleanup/old",
		r.RequireAuth(),
		webhookHandler.CleanupOldWebhooks,
	)

	// Cleanup delivered webhooks
	webhooks.Post("/cleanup/delivered",
		r.RequireAuth(),
		webhookHandler.CleanupDeliveredWebhooks,
	)

	// Purge failed webhooks
	webhooks.Post("/cleanup/purge",
		r.RequireAuth(),
		webhookHandler.PurgeFailedWebhooks,
	)

	// ============================================================================
	// Background Processing (Admin/Manage access required)
	// ============================================================================

	// Process pending webhooks
	webhooks.Post("/process-pending",
		r.RequireAuth(),
		webhookHandler.ProcessPendingWebhooks,
	)

	// ============================================================================
	// Health & Monitoring
	// ============================================================================

	// Health check (read access)
	webhooks.Get("/health",
		r.RequireAuth(),
		webhookHandler.HealthCheck,
	)

	// Service metrics (read access)
	webhooks.Get("/metrics",
		r.RequireAuth(),
		webhookHandler.GetServiceMetrics,
	)

	r.config.Logger.Info("webhook routes registered successfully")
}
