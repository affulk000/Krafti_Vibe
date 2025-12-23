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
		r.zitadelMW.RequireAuth(),
		webhookHandler.CreateWebhookEvent,
	)

	// Get webhook event by ID
	webhooks.Get("/:id",
		r.zitadelMW.RequireAuth(),
		webhookHandler.GetWebhookEvent,
	)

	// List webhook events
	webhooks.Post("/list",
		r.zitadelMW.RequireAuth(),
		webhookHandler.ListWebhookEvents,
	)

	// ============================================================================
	// Delivery Operations
	// ============================================================================

	// Deliver webhook (manual trigger)
	webhooks.Post("/:id/deliver",
		r.zitadelMW.RequireAuth(),
		webhookHandler.DeliverWebhook,
	)

	// Retry webhook
	webhooks.Post("/:id/retry",
		r.zitadelMW.RequireAuth(),
		webhookHandler.RetryWebhook,
	)

	// Retry failed webhooks for a tenant
	webhooks.Post("/tenant/:tenantId/retry-failed",
		r.zitadelMW.RequireAuth(),
		webhookHandler.RetryFailedWebhooks,
	)

	// Bulk retry webhooks
	webhooks.Post("/bulk-retry",
		r.zitadelMW.RequireAuth(),
		webhookHandler.BulkRetryWebhooks,
	)

	// ============================================================================
	// Query Operations
	// ============================================================================

	// Get pending webhooks
	webhooks.Get("/pending",
		r.zitadelMW.RequireAuth(),
		webhookHandler.GetPendingWebhooks,
	)

	// Get failed webhooks
	webhooks.Get("/failed",
		r.zitadelMW.RequireAuth(),
		webhookHandler.GetFailedWebhooks,
	)

	// Get delivered webhooks
	webhooks.Get("/delivered",
		r.zitadelMW.RequireAuth(),
		webhookHandler.GetDeliveredWebhooks,
	)

	// Get recent webhooks
	webhooks.Get("/recent",
		r.zitadelMW.RequireAuth(),
		webhookHandler.GetRecentWebhooks,
	)

	// ============================================================================
	// Analytics
	// ============================================================================

	// Get webhook statistics
	webhooks.Get("/stats",
		r.zitadelMW.RequireAuth(),
		webhookHandler.GetWebhookStats,
	)

	// Get webhook analytics
	webhooks.Get("/analytics",
		r.zitadelMW.RequireAuth(),
		webhookHandler.GetWebhookAnalytics,
	)

	// Get failure reasons
	webhooks.Get("/failure-reasons",
		r.zitadelMW.RequireAuth(),
		webhookHandler.GetFailureReasons,
	)

	// ============================================================================
	// Cleanup Operations (Admin/Manage access required)
	// ============================================================================

	// Cleanup old webhooks
	webhooks.Post("/cleanup/old",
		r.zitadelMW.RequireAuth(),
		webhookHandler.CleanupOldWebhooks,
	)

	// Cleanup delivered webhooks
	webhooks.Post("/cleanup/delivered",
		r.zitadelMW.RequireAuth(),
		webhookHandler.CleanupDeliveredWebhooks,
	)

	// Purge failed webhooks
	webhooks.Post("/cleanup/purge",
		r.zitadelMW.RequireAuth(),
		webhookHandler.PurgeFailedWebhooks,
	)

	// ============================================================================
	// Background Processing (Admin/Manage access required)
	// ============================================================================

	// Process pending webhooks
	webhooks.Post("/process-pending",
		r.zitadelMW.RequireAuth(),
		webhookHandler.ProcessPendingWebhooks,
	)

	// ============================================================================
	// Health & Monitoring
	// ============================================================================

	// Health check (read access)
	webhooks.Get("/health",
		r.zitadelMW.RequireAuth(),
		webhookHandler.HealthCheck,
	)

	// Service metrics (read access)
	webhooks.Get("/metrics",
		r.zitadelMW.RequireAuth(),
		webhookHandler.GetServiceMetrics,
	)

	r.config.Logger.Info("webhook routes registered successfully")
}
