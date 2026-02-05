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
	invoices.Use(r.RequireAuth())

	// ============================================================================
	// Core Invoice Operations
	// ============================================================================

	// Create invoice - artisan or tenant owner/admin
	invoices.Post("/",
		middleware.RequireArtisanOrTeamMember(),
		invoiceHandler.CreateInvoice,
	)

	// Get invoice by ID - customer (owner) or artisan or tenant owner/admin
	invoices.Get("/:id",
		invoiceHandler.GetInvoice,
	)

	// Update invoice - artisan or tenant owner/admin
	invoices.Put("/:id",
		middleware.RequireArtisanOrTeamMember(),
		invoiceHandler.UpdateInvoice,
	)

	// Delete invoice - tenant owner/admin only
	invoices.Delete("/:id",
		middleware.RequireTenantOwnerOrAdmin(),
		invoiceHandler.DeleteInvoice,
	)

	// ============================================================================
	// Invoice Actions
	// ============================================================================

	// Mark invoice as paid - tenant owner/admin or customer (self)
	invoices.Post("/:id/pay",
		invoiceHandler.MarkInvoiceAsPaid,
	)

	// Send invoice to customer - artisan or tenant owner/admin
	invoices.Post("/:id/send",
		middleware.RequireArtisanOrTeamMember(),
		invoiceHandler.SendInvoice,
	)

	// ============================================================================
	// Related Resource Queries
	// ============================================================================

	// Get invoices by booking - booking owner or tenant owner/admin
	invoices.Get("/booking/:booking_id",
		invoiceHandler.GetBookingInvoice,
	)

	// Get invoices by customer - customer (self) or tenant owner/admin
	invoices.Get("/customer/:customer_id",
		middleware.RequireSelfOrAdmin(),
		invoiceHandler.GetCustomerInvoices,
	)
}
