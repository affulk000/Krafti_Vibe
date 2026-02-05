package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *Router) setupSubscriptionRoutes(api fiber.Router) {
	// Initialize dependent services
	paymentService := service.NewPaymentService(r.repos, r.config.Logger)

	// Initialize subscription service with dependencies
	subscriptionService := service.NewSubscriptionService(r.repos, paymentService, r.config.Logger)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)

	// Create subscriptions group
	subscriptions := api.Group("/subscriptions")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger = zap.NewNop()
		}
		subscriptions.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// Auth middleware configuration

	// ============================================================================
	// Core Subscription Operations
	// ============================================================================

	// Create subscription (authenticated, requires subscription:write scope)
	subscriptions.Post("/",
		r.RequireAuth(),
		subscriptionHandler.CreateSubscription,
	)

	// Get subscription by ID (authenticated, requires subscription:read scope)
	subscriptions.Get("/:id",
		r.RequireAuth(),
		subscriptionHandler.GetSubscription,
	)

	// Update subscription (authenticated, requires subscription:write scope)
	subscriptions.Put("/:id",
		r.RequireAuth(),
		subscriptionHandler.UpdateSubscription,
	)

	// List subscriptions (authenticated, requires subscription:read scope)
	subscriptions.Get("/",
		r.RequireAuth(),
		subscriptionHandler.ListSubscriptions,
	)

	// ============================================================================
	// Subscription Actions
	// ============================================================================

	// Cancel subscription (authenticated, requires subscription:write scope)
	subscriptions.Post("/:id/cancel",
		r.RequireAuth(),
		subscriptionHandler.CancelSubscription,
	)
}
