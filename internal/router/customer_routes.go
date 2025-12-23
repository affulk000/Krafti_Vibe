package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *Router) setupCustomerRoutes(api fiber.Router) {
	// Initialize service and handler
	customerService := service.NewCustomerService(r.repos, r.config.Logger)
	customerHandler := handler.NewCustomerHandler(customerService)

	// Create customers group
	customers := api.Group("/customers")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger = zap.NewNop()
		}
		customers.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// Auth middleware configuration
	customers.Use(r.zitadelMW.RequireAuth())

	// ============================================================================
	// Core Customer Operations
	// ============================================================================

	// Create customer - any authenticated user can become a customer
	customers.Post("/",
		customerHandler.CreateCustomer,
	)

	// List customers - tenant owner/admin only
	customers.Get("/",
		middleware.RequireTenantOwnerOrAdmin(),
		customerHandler.ListCustomers,
	)

	// ============================================================================
	// Customer Lookup Operations (specific routes before parameterized routes)
	// ============================================================================

	// Get customer by user ID - self or tenant owner/admin
	customers.Get("/user/:user_id",
		middleware.RequireSelfOrAdmin(),
		customerHandler.GetCustomerByUserID,
	)

	// Search customers - tenant owner/admin only
	customers.Post("/search",
		middleware.RequireTenantOwnerOrAdmin(),
		customerHandler.SearchCustomers,
	)

	// Get active customers - tenant owner/admin only
	customers.Get("/active",
		middleware.RequireTenantOwnerOrAdmin(),
		customerHandler.GetActiveCustomers,
	)

	// Get top customers - tenant owner/admin only
	customers.Get("/top",
		middleware.RequireTenantOwnerOrAdmin(),
		customerHandler.GetTopCustomers,
	)

	// Get customer segments - tenant owner/admin only (must be before /:id)
	customers.Get("/segments",
		middleware.RequireTenantOwnerOrAdmin(),
		customerHandler.GetCustomerSegments,
	)

	// Get customer by ID - self or tenant owner/admin
	customers.Get("/:id",
		middleware.RequireSelfOrAdmin(),
		customerHandler.GetCustomer,
	)

	// Update customer - self or tenant owner/admin
	customers.Put("/:id",
		middleware.RequireSelfOrAdmin(),
		customerHandler.UpdateCustomer,
	)

	// Delete customer - tenant owner/admin only
	customers.Delete("/:id",
		middleware.RequireTenantOwnerOrAdmin(),
		customerHandler.DeleteCustomer,
	)

	// ============================================================================
	// Loyalty Program
	// ============================================================================

	// Update loyalty points - tenant owner/admin only
	customers.Put("/:id/loyalty-points",
		middleware.RequireTenantOwnerOrAdmin(),
		customerHandler.UpdateLoyaltyPoints,
	)

	// Get loyalty points history - self or tenant owner/admin
	customers.Get("/:id/loyalty-history",
		middleware.RequireSelfOrAdmin(),
		customerHandler.GetLoyaltyPointsHistory,
	)

	// ============================================================================
	// Preferences & Settings
	// ============================================================================

	// Add preferred artisan - self or tenant owner/admin
	customers.Post("/:id/preferred-artisans",
		middleware.RequireSelfOrAdmin(),
		customerHandler.AddPreferredArtisan,
	)

	// Remove preferred artisan - self or tenant owner/admin
	customers.Delete("/:id/preferred-artisans/:artisan_id",
		middleware.RequireSelfOrAdmin(),
		customerHandler.RemovePreferredArtisan,
	)

	// Update notification preferences - self or tenant owner/admin
	customers.Put("/:id/notification-preferences",
		middleware.RequireSelfOrAdmin(),
		customerHandler.UpdateNotificationPreferences,
	)

	// Update primary location - self or tenant owner/admin
	customers.Put("/:id/primary-location",
		middleware.RequireSelfOrAdmin(),
		customerHandler.UpdatePrimaryLocation,
	)

	// ============================================================================
	// Analytics & Segmentation
	// ============================================================================

	// Get customer statistics - self or tenant owner/admin
	customers.Get("/:id/stats",
		middleware.RequireSelfOrAdmin(),
		customerHandler.GetCustomerStats,
	)
}
