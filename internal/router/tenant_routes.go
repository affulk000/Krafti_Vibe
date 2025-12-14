package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *Router) setupTenantRoutes(api fiber.Router) {
	// Initialize service and handler
	zapLogger := r.config.ZapLogger
	if zapLogger == nil {
		zapLogger = zap.NewNop()
	}
	tenantService := service.NewTenantService(r.repos, zapLogger)
	tenantHandler := handler.NewTenantHandler(tenantService)

	// Create tenants group
	tenants := api.Group("/tenants")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger = zap.NewNop()
		}
		tenants.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core Tenant Operations
	// ============================================================================

	// Create tenant (authenticated, requires tenant:write scope)
	tenants.Post("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantWrite),
		tenantHandler.CreateTenant,
	)

	// Get tenant by ID (authenticated, requires tenant:read scope)
	tenants.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		tenantHandler.GetTenant,
	)

	// Update tenant (authenticated, requires tenant:write scope)
	tenants.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantWrite),
		tenantHandler.UpdateTenant,
	)

	// Delete tenant (authenticated, requires tenant:write scope)
	tenants.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantWrite),
		tenantHandler.DeleteTenant,
	)

	// List tenants (authenticated, requires tenant:read scope)
	tenants.Get("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		tenantHandler.ListTenants,
	)

	// Search tenants (authenticated, requires tenant:read scope)
	tenants.Post("/search",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		tenantHandler.SearchTenants,
	)

	// ============================================================================
	// Tenant Lookup Operations
	// ============================================================================

	// Get tenant by subdomain (authenticated, requires tenant:read scope)
	tenants.Get("/subdomain/:subdomain",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		tenantHandler.GetTenantBySubdomain,
	)

	// Get tenant by domain (authenticated, requires tenant:read scope)
	tenants.Get("/domain/:domain",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		tenantHandler.GetTenantByDomain,
	)

	// Check subdomain availability (authenticated, requires tenant:read scope)
	tenants.Get("/check-subdomain/:subdomain",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		tenantHandler.CheckSubdomainAvailability,
	)

	// Check domain availability (authenticated, requires tenant:read scope)
	tenants.Get("/check-domain/:domain",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		tenantHandler.CheckDomainAvailability,
	)

	// ============================================================================
	// Tenant Status Management
	// ============================================================================

	// Activate tenant (authenticated, requires tenant:write scope)
	tenants.Post("/:id/activate",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantWrite),
		tenantHandler.ActivateTenant,
	)

	// Suspend tenant (authenticated, requires tenant:write scope)
	tenants.Post("/:id/suspend",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantWrite),
		tenantHandler.SuspendTenant,
	)

	// Cancel tenant (authenticated, requires tenant:write scope)
	tenants.Post("/:id/cancel",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantWrite),
		tenantHandler.CancelTenant,
	)

	// ============================================================================
	// Tenant Settings & Configuration
	// ============================================================================

	// Update tenant plan (authenticated, requires tenant:write scope)
	tenants.Put("/:id/plan",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantWrite),
		tenantHandler.UpdateTenantPlan,
	)

	// Update tenant settings (authenticated, requires tenant:write scope)
	tenants.Put("/:id/settings",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantWrite),
		tenantHandler.UpdateTenantSettings,
	)

	// Update tenant features (authenticated, requires tenant:write scope)
	tenants.Put("/:id/features",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantWrite),
		tenantHandler.UpdateTenantFeatures,
	)

	// ============================================================================
	// Tenant Metrics & Information
	// ============================================================================

	// Get tenant statistics (authenticated, requires tenant:read scope)
	tenants.Get("/:id/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		tenantHandler.GetTenantStats,
	)

	// Get tenant details (authenticated, requires tenant:read scope)
	tenants.Get("/:id/details",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		tenantHandler.GetTenantDetails,
	)

	// Get tenant health (authenticated, requires tenant:read scope)
	tenants.Get("/:id/health",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		tenantHandler.GetTenantHealth,
	)

	// Get tenant limits (authenticated, requires tenant:read scope)
	tenants.Get("/:id/limits",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		tenantHandler.GetTenantLimits,
	)

	// ============================================================================
	// Trial Management
	// ============================================================================

	// Extend tenant trial (authenticated, requires tenant:write scope)
	tenants.Post("/:id/extend-trial",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantWrite),
		tenantHandler.ExtendTrial,
	)

	// Get expired trials (authenticated, requires tenant:read scope)
	tenants.Get("/trials/expired",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TenantRead),
		tenantHandler.GetExpiredTrials,
	)
}
