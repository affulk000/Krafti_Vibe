package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *Router) setupDataExportRoutes(api fiber.Router) {
	// Initialize service and handler
	zapLogger := r.config.ZapLogger
	if zapLogger == nil {
		zapLogger = zap.NewNop()
	}
	exportService := service.NewDataExportService(r.repos, zapLogger, service.DataExportServiceConfig{})
	exportHandler := handler.NewDataExportHandler(exportService)

	// Create data-exports group
	exports := api.Group("/data-exports")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger = zap.NewNop()
		}
		exports.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// ============================================================================
	// Core Export Operations
	// ============================================================================

	// Request data export (authenticated, requires data:export scope)
	exports.Post("/",
		r.RequireAuth(),
		middleware.RequireTenantContext(),
		exportHandler.RequestDataExport,
	)

	// Get export status (authenticated, requires data:read scope)
	exports.Get("/:id",
		r.RequireAuth(),
		middleware.RequireTenantContext(),
		exportHandler.GetDataExportStatus,
	)

	// List data exports (authenticated, requires data:read scope)
	exports.Get("/",
		r.RequireAuth(),
		middleware.RequireTenantContext(),
		exportHandler.ListDataExports,
	)

	// Cancel data export (authenticated, requires data:export scope)
	exports.Post("/:id/cancel",
		r.RequireAuth(),
		middleware.RequireTenantContext(),
		exportHandler.CancelDataExport,
	)

	// ============================================================================
	// Admin Operations
	// ============================================================================

	// Get pending exports (admin only)
	exports.Get("/pending",
		r.RequireAuth(),
		middleware.RequireTenantOwnerOrAdmin(),
		exportHandler.GetPendingExports,
	)

	// Get exports by status (admin only)
	exports.Get("/by-status",
		r.RequireAuth(),
		middleware.RequireTenantOwnerOrAdmin(),
		exportHandler.GetExportsByStatus,
	)

	// Delete expired exports (admin only)
	exports.Post("/cleanup",
		r.RequireAuth(),
		middleware.RequireTenantOwnerOrAdmin(),
		exportHandler.DeleteExpiredExports,
	)
}
