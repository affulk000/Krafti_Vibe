package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
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
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Webhook Event Management
	// ============================================================================

	// Create webhook event
	webhooks.Post("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookWrite),
		webhookHandler.CreateWebhookEvent,
	)

	// Get webhook event by ID
	webhooks.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookRead),
		webhookHandler.GetWebhookEvent,
	)

	// List webhook events
	webhooks.Post("/list",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookRead),
		webhookHandler.ListWebhookEvents,
	)

	// ============================================================================
	// Delivery Operations
	// ============================================================================

	// Deliver webhook (manual trigger)
	webhooks.Post("/:id/deliver",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookWrite),
		webhookHandler.DeliverWebhook,
	)

	// Retry webhook
	webhooks.Post("/:id/retry",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookWrite),
		webhookHandler.RetryWebhook,
	)

	// Retry failed webhooks for a tenant
	webhooks.Post("/tenant/:tenantId/retry-failed",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookManage),
		webhookHandler.RetryFailedWebhooks,
	)

	// Bulk retry webhooks
	webhooks.Post("/bulk-retry",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookManage),
		webhookHandler.BulkRetryWebhooks,
	)

	// ============================================================================
	// Query Operations
	// ============================================================================

	// Get pending webhooks
	webhooks.Get("/pending",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookRead),
		webhookHandler.GetPendingWebhooks,
	)

	// Get failed webhooks
	webhooks.Get("/failed",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookRead),
		webhookHandler.GetFailedWebhooks,
	)

	// Get delivered webhooks
	webhooks.Get("/delivered",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookRead),
		webhookHandler.GetDeliveredWebhooks,
	)

	// Get recent webhooks
	webhooks.Get("/recent",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookRead),
		webhookHandler.GetRecentWebhooks,
	)

	// ============================================================================
	// Analytics
	// ============================================================================

	// Get webhook statistics
	webhooks.Get("/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookRead),
		webhookHandler.GetWebhookStats,
	)

	// Get webhook analytics
	webhooks.Get("/analytics",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookRead),
		webhookHandler.GetWebhookAnalytics,
	)

	// Get failure reasons
	webhooks.Get("/failure-reasons",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookRead),
		webhookHandler.GetFailureReasons,
	)

	// ============================================================================
	// Cleanup Operations (Admin/Manage access required)
	// ============================================================================

	// Cleanup old webhooks
	webhooks.Post("/cleanup/old",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookManage),
		webhookHandler.CleanupOldWebhooks,
	)

	// Cleanup delivered webhooks
	webhooks.Post("/cleanup/delivered",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookManage),
		webhookHandler.CleanupDeliveredWebhooks,
	)

	// Purge failed webhooks
	webhooks.Post("/cleanup/purge",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookManage),
		webhookHandler.PurgeFailedWebhooks,
	)

	// ============================================================================
	// Background Processing (Admin/Manage access required)
	// ============================================================================

	// Process pending webhooks
	webhooks.Post("/process-pending",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookManage),
		webhookHandler.ProcessPendingWebhooks,
	)

	// ============================================================================
	// Health & Monitoring
	// ============================================================================

	// Health check (read access)
	webhooks.Get("/health",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookRead),
		webhookHandler.HealthCheck,
	)

	// Service metrics (read access)
	webhooks.Get("/metrics",
		authMiddleware,
		middleware.RequireScopes(r.scopes.WebhookRead),
		webhookHandler.GetServiceMetrics,
	)

	// ============================================================================
	// External Webhook Endpoints (No Auth Required - Verified by Signature)
	// ============================================================================

	// Logto webhook endpoint for user sync
	r.setupLogtoWebhookRoutes(webhooks)

	r.config.Logger.Info("webhook routes registered successfully")
}

// setupLogtoWebhookRoutes configures Logto webhook endpoints
func (r *Router) setupLogtoWebhookRoutes(webhooks fiber.Router) {
	// Initialize Logto webhook handler
	userService := service.NewUserService(r.repos, r.config.Logger)
	logtoWebhookHandler := handler.NewLogtoWebhookHandler(userService)

	// Get webhook signing secret from config
	webhookSecret := r.config.WebhookSecret
	skipVerification := webhookSecret == ""

	if skipVerification {
		r.config.Logger.Warn("Logto webhook signature verification is DISABLED - set LOGTO_WEBHOOK_SECRET to enable")
	}

	// Logto webhook endpoint - public endpoint with signature verification
	webhooks.Post("/logto",
		middleware.LogWebhookRequest(),
		middleware.VerifyLogtoWebhook(middleware.WebhookConfig{
			SigningSecret:    webhookSecret,
			SignatureHeader:  "Logto-Signature-Sha-256",
			SkipVerification: skipVerification,
		}),
		logtoWebhookHandler.HandleWebhook,
	)

	r.config.Logger.Info("Logto webhook endpoint registered at /api/v1/webhooks/logto")
}
