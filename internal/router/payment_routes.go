package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *Router) setupPaymentRoutes(api fiber.Router) {
	// Initialize service and handler
	paymentService := service.NewPaymentService(r.repos, r.config.Logger)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// Create payments group
	payments := api.Group("/payments")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger = zap.NewNop()
		}
		payments.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core Payment Operations
	// ============================================================================

	// Create payment (authenticated, requires payment:write scope)
	payments.Post("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentWrite),
		paymentHandler.CreatePayment,
	)

	// Get payment by ID (authenticated, requires payment:read scope)
	payments.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentRead),
		paymentHandler.GetPayment,
	)

	// ============================================================================
	// Payment Queries by Resource
	// ============================================================================

	// Get payments by booking (authenticated, requires payment:read scope)
	payments.Get("/booking/:booking_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentRead),
		paymentHandler.GetPaymentsByBooking,
	)

	// Get payments by customer (authenticated, requires payment:read scope)
	payments.Get("/customer/:customer_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentRead),
		paymentHandler.GetPaymentsByCustomer,
	)

	// Get payments by artisan (authenticated, requires payment:read scope)
	payments.Get("/artisan/:artisan_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentRead),
		paymentHandler.GetPaymentsByArtisan,
	)

	// Get payments by tenant (authenticated, requires payment:read scope)
	payments.Get("/tenant/:tenant_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentRead),
		paymentHandler.GetPaymentsByTenant,
	)

	// Get payments by method (authenticated, requires payment:read scope)
	payments.Get("/method/:payment_method",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentRead),
		paymentHandler.GetPaymentsByMethod,
	)

	// ============================================================================
	// Payment Status Management
	// ============================================================================

	// Mark payment as paid (authenticated, requires payment:process scope)
	payments.Post("/:id/mark-paid",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentProcess),
		paymentHandler.MarkPaymentAsPaid,
	)

	// Mark payment as failed (authenticated, requires payment:process scope)
	payments.Post("/:id/mark-failed",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentProcess),
		paymentHandler.MarkPaymentAsFailed,
	)

	// Get pending payments (authenticated, requires payment:read scope)
	payments.Get("/pending",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentRead),
		paymentHandler.GetPendingPayments,
	)

	// ============================================================================
	// Refunds
	// ============================================================================

	// Process refund (authenticated, requires payment:process scope)
	payments.Post("/:id/refund",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentProcess),
		paymentHandler.ProcessRefund,
	)

	// Get refundable payments (authenticated, requires payment:read scope)
	payments.Get("/refundable",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentRead),
		paymentHandler.GetRefundablePayments,
	)

	// ============================================================================
	// Financial Analytics
	// ============================================================================

	// Get artisan earnings (authenticated, requires payment:read scope)
	payments.Get("/earnings/artisan/:artisan_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentRead),
		paymentHandler.GetArtisanEarnings,
	)

	// Get platform revenue (authenticated, requires payment:read scope)
	payments.Get("/revenue/platform",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentRead),
		paymentHandler.GetPlatformRevenue,
	)

	// Get payment statistics (authenticated, requires payment:read scope)
	payments.Get("/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentRead),
		paymentHandler.GetPaymentStats,
	)

	// Get payment trends (authenticated, requires payment:read scope)
	payments.Get("/trends",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PaymentRead),
		paymentHandler.GetPaymentTrends,
	)
}
