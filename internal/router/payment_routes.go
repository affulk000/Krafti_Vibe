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
	payments.Use(r.zitadelMW.RequireAuth())

	// ============================================================================
	// Core Payment Operations
	// ============================================================================

	// Create payment - customer when paying for booking
	payments.Post("/",
		paymentHandler.CreatePayment,
	)

	// Get payment by ID - owner (customer/artisan) or tenant owner/admin
	payments.Get("/:id",
		paymentHandler.GetPayment,
	)

	// ============================================================================
	// Payment Queries by Resource
	// ============================================================================

	// Get payments by booking - booking owner or tenant owner/admin
	payments.Get("/booking/:booking_id",
		paymentHandler.GetPaymentsByBooking,
	)

	// Get payments by customer - customer (self) or tenant owner/admin
	payments.Get("/customer/:customer_id",
		middleware.RequireSelfOrAdmin(),
		paymentHandler.GetPaymentsByCustomer,
	)

	// Get payments by artisan - artisan (self) or tenant owner/admin
	payments.Get("/artisan/:artisan_id",
		middleware.RequireArtisanOrTeamMember(),
		paymentHandler.GetPaymentsByArtisan,
	)

	// Get payments by tenant - tenant owner/admin only
	payments.Get("/tenant/:tenant_id",
		middleware.RequireTenantOwnerOrAdmin(),
		paymentHandler.GetPaymentsByTenant,
	)

	// Get payments by method - tenant owner/admin only
	payments.Get("/method/:payment_method",
		middleware.RequireTenantOwnerOrAdmin(),
		paymentHandler.GetPaymentsByMethod,
	)

	// ============================================================================
	// Payment Status Management
	// ============================================================================

	// Mark payment as paid - tenant owner/admin only
	payments.Post("/:id/mark-paid",
		middleware.RequireTenantOwnerOrAdmin(),
		paymentHandler.MarkPaymentAsPaid,
	)

	// Mark payment as failed - tenant owner/admin only
	payments.Post("/:id/mark-failed",
		middleware.RequireTenantOwnerOrAdmin(),
		paymentHandler.MarkPaymentAsFailed,
	)

	// Get pending payments - tenant owner/admin only
	payments.Get("/pending",
		middleware.RequireTenantOwnerOrAdmin(),
		paymentHandler.GetPendingPayments,
	)

	// ============================================================================
	// Refunds
	// ============================================================================

	// Process refund - tenant owner/admin only
	payments.Post("/:id/refund",
		middleware.RequireTenantOwnerOrAdmin(),
		paymentHandler.ProcessRefund,
	)

	// Get refundable payments - tenant owner/admin only
	payments.Get("/refundable",
		middleware.RequireTenantOwnerOrAdmin(),
		paymentHandler.GetRefundablePayments,
	)

	// ============================================================================
	// Financial Analytics
	// ============================================================================

	// Get artisan earnings - artisan (self) or tenant owner/admin
	payments.Get("/earnings/artisan/:artisan_id",
		middleware.RequireArtisanOrTeamMember(),
		paymentHandler.GetArtisanEarnings,
	)

	// Get platform revenue - platform admin only
	payments.Get("/revenue/platform",
		r.zitadelMW.RequireAnyPlatformRole(),
		paymentHandler.GetPlatformRevenue,
	)

	// Get payment statistics - tenant owner/admin only
	payments.Get("/stats",
		middleware.RequireTenantOwnerOrAdmin(),
		paymentHandler.GetPaymentStats,
	)

	// Get payment trends - tenant owner/admin only
	payments.Get("/trends",
		middleware.RequireTenantOwnerOrAdmin(),
		paymentHandler.GetPaymentTrends,
	)
}
