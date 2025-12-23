package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *Router) setupTenantUsageRoutes(api fiber.Router) {
	// Initialize service and handler
	zapLogger := r.config.ZapLogger
	if zapLogger == nil {
		zapLogger = zap.NewNop()
	}
	usageService := service.NewTenantUsageService(r.repos, zapLogger)
	usageHandler := handler.NewTenantUsageHandler(usageService)

	// ============================================================================
	// Tenant-scoped Usage Routes
	// ============================================================================

	// Create tenants usage group
	tenantsUsage := api.Group("/tenants/:tenant_id/usage")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger = zap.NewNop()
		}
		tenantsUsage.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// Daily usage (authenticated, requires usage:read scope)
	tenantsUsage.Get("/daily",
		r.zitadelMW.RequireAuth(),
		middleware.RequireTenantOwnerOrAdmin(),
		usageHandler.GetDailyUsage,
	)

	// Usage history (authenticated, requires usage:read scope)
	tenantsUsage.Get("/history",
		r.zitadelMW.RequireAuth(),
		middleware.RequireTenantOwnerOrAdmin(),
		usageHandler.GetUsageHistory,
	)

	// API usage stats (authenticated, requires usage:read scope)
	tenantsUsage.Get("/api",
		r.zitadelMW.RequireAuth(),
		middleware.RequireTenantOwnerOrAdmin(),
		usageHandler.GetAPIUsageStats,
	)

	// Peak usage (authenticated, requires usage:read scope)
	tenantsUsage.Get("/peak",
		r.zitadelMW.RequireAuth(),
		middleware.RequireTenantOwnerOrAdmin(),
		usageHandler.GetPeakUsage,
	)

	// Average usage (authenticated, requires usage:read scope)
	tenantsUsage.Get("/average",
		r.zitadelMW.RequireAuth(),
		middleware.RequireTenantOwnerOrAdmin(),
		usageHandler.GetAverageUsage,
	)

	// Track feature usage (authenticated, requires usage:write scope)
	tenantsUsage.Post("/track",
		r.zitadelMW.RequireAuth(),
		middleware.RequireTenantContext(),
		usageHandler.TrackFeatureUsage,
	)

	// Increment API usage (internal use - authenticated)
	tenantsUsage.Post("/api/increment",
		r.zitadelMW.RequireAuth(),
		usageHandler.IncrementAPIUsage,
	)

	// Check API rate limit (authenticated)
	tenantsUsage.Get("/api/check",
		r.zitadelMW.RequireAuth(),
		usageHandler.CheckAPIRateLimit,
	)

	// ============================================================================
	// Admin Operations
	// ============================================================================

	// Create usage admin group
	usage := api.Group("/usage")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger = zap.NewNop()
		}
		usage.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// Delete old usage records (platform admin only)
	usage.Delete("/cleanup",
		r.zitadelMW.RequireAuth(),
		middleware.RequireTenantOwnerOrAdmin(),
		usageHandler.DeleteOldUsageRecords,
	)
}
