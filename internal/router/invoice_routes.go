package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *Router) setupInvoiceRoutes(api fiber.Router) {
	// Initialize service and handler
	invoiceService := service.NewInvoiceService(r.repos, r.config.Logger)
	invoiceHandler := handler.NewInvoiceHandler(invoiceService)

	// Create invoices group
	invoices := api.Group("/invoices")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger = zap.NewNop()
		}
		invoices.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core Invoice Operations
	// ============================================================================

	// Create invoice (authenticated, requires invoice:write scope)
	invoices.Post("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.InvoiceWrite),
		invoiceHandler.CreateInvoice,
	)

	// Get invoice by ID (authenticated, requires invoice:read scope)
	invoices.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.InvoiceRead),
		invoiceHandler.GetInvoice,
	)

	// Update invoice (authenticated, requires invoice:write scope)
	invoices.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.InvoiceWrite),
		invoiceHandler.UpdateInvoice,
	)

	// Delete invoice (authenticated, requires invoice:write scope)
	invoices.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.InvoiceWrite),
		invoiceHandler.DeleteInvoice,
	)

	// ============================================================================
	// Invoice Actions
	// ============================================================================

	// Mark invoice as paid (authenticated, requires invoice:write scope)
	invoices.Post("/:id/pay",
		authMiddleware,
		middleware.RequireScopes(r.scopes.InvoiceWrite),
		invoiceHandler.MarkInvoiceAsPaid,
	)

	// Send invoice to customer (authenticated, requires invoice:write scope)
	invoices.Post("/:id/send",
		authMiddleware,
		middleware.RequireScopes(r.scopes.InvoiceWrite),
		invoiceHandler.SendInvoice,
	)

	// ============================================================================
	// Related Resource Queries
	// ============================================================================

	// Get invoices by booking (authenticated, requires invoice:read scope)
	invoices.Get("/booking/:booking_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.InvoiceRead),
		invoiceHandler.GetBookingInvoice,
	)

	// Get invoices by customer (authenticated, requires invoice:read scope)
	invoices.Get("/customer/:customer_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.InvoiceRead),
		invoiceHandler.GetCustomerInvoices,
	)
}
