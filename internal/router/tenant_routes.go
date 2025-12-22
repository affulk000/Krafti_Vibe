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
	tenants.Use(r.zitadelMW.RequireAuth())

	// ============================================================================
	// Core Tenant Operations
	// ============================================================================

	// List tenants - platform admin only (must be before /:id)
	tenants.Get("/",
		r.zitadelMW.RequireAnyPlatformRole(),
		tenantHandler.ListTenants,
	)

	// Search tenants - platform admin only (must be before /:id)
	tenants.Post("/search",
		r.zitadelMW.RequireAnyPlatformRole(),
		tenantHandler.SearchTenants,
	)

	// Create tenant - any authenticated user can create a tenant (becomes tenant owner)
	tenants.Post("/",
		tenantHandler.CreateTenant,
	)

	// ============================================================================
	// Tenant Lookup Operations (must be before /:id)
	// ============================================================================

	// Get tenant by subdomain - any authenticated user (for tenant discovery)
	tenants.Get("/subdomain/:subdomain",
		tenantHandler.GetTenantBySubdomain,
	)

	// Get tenant by domain - any authenticated user (for tenant discovery)
	tenants.Get("/domain/:domain",
		tenantHandler.GetTenantByDomain,
	)

	// Check subdomain availability - any authenticated user (for signup)
	tenants.Get("/check-subdomain/:subdomain",
		tenantHandler.CheckSubdomainAvailability,
	)

	// Check domain availability - any authenticated user (for signup)
	tenants.Get("/check-domain/:domain",
		tenantHandler.CheckDomainAvailability,
	)

	// Get expired trials - platform admin only (must be before /:id)
	tenants.Get("/trials/expired",
		r.zitadelMW.RequireAnyPlatformRole(),
		tenantHandler.GetExpiredTrials,
	)

	// ============================================================================
	// Parameterized Routes (must be after specific routes)
	// ============================================================================

	// Get tenant by ID - tenant owner/admin can view their tenant, platform users can view all
	tenants.Get("/:id",
		middleware.RequireTenantOwnerOrAdmin(),
		tenantHandler.GetTenant,
	)

	// Update tenant - tenant owner/admin only
	tenants.Put("/:id",
		middleware.RequireTenantOwnerOrAdmin(),
		tenantHandler.UpdateTenant,
	)

	// Delete tenant - platform admin or tenant owner only
	tenants.Delete("/:id",
		middleware.RequireTenantOwner(),
		tenantHandler.DeleteTenant,
	)

	// ============================================================================
	// Tenant Status Management
	// ============================================================================

	// Activate tenant - platform admin or tenant owner
	tenants.Post("/:id/activate",
		middleware.RequireTenantOwner(),
		tenantHandler.ActivateTenant,
	)

	// Suspend tenant - platform admin only
	tenants.Post("/:id/suspend",
		r.zitadelMW.RequireAnyPlatformRole(),
		tenantHandler.SuspendTenant,
	)

	// Cancel tenant - tenant owner only
	tenants.Post("/:id/cancel",
		middleware.RequireTenantOwner(),
		tenantHandler.CancelTenant,
	)

	// ============================================================================
	// Tenant Settings & Configuration
	// ============================================================================

	// Update tenant plan - tenant owner only
	tenants.Put("/:id/plan",
		middleware.RequireTenantOwner(),
		tenantHandler.UpdateTenantPlan,
	)

	// Update tenant settings - tenant owner/admin only
	tenants.Put("/:id/settings",
		middleware.RequireTenantOwnerOrAdmin(),
		tenantHandler.UpdateTenantSettings,
	)

	// Update tenant features - tenant owner only
	tenants.Put("/:id/features",
		middleware.RequireTenantOwner(),
		tenantHandler.UpdateTenantFeatures,
	)

	// ============================================================================
	// Tenant Metrics & Information
	// ============================================================================

	// Get tenant statistics - tenant owner/admin only
	tenants.Get("/:id/stats",
		middleware.RequireTenantOwnerOrAdmin(),
		tenantHandler.GetTenantStats,
	)

	// Get tenant details - tenant owner/admin only
	tenants.Get("/:id/details",
		middleware.RequireTenantOwnerOrAdmin(),
		tenantHandler.GetTenantDetails,
	)

	// Get tenant health - tenant owner/admin only
	tenants.Get("/:id/health",
		middleware.RequireTenantOwnerOrAdmin(),
		tenantHandler.GetTenantHealth,
	)

	// Get tenant limits - tenant owner/admin only
	tenants.Get("/:id/limits",
		middleware.RequireTenantOwnerOrAdmin(),
		tenantHandler.GetTenantLimits,
	)

	// ============================================================================
	// Trial Management
	// ============================================================================

	// Extend tenant trial - platform admin only
	tenants.Post("/:id/extend-trial",
		r.zitadelMW.RequireAnyPlatformRole(),
		tenantHandler.ExtendTrial,
	)
}
