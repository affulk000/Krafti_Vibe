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
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core Customer Operations
	// ============================================================================

	// Create customer (authenticated, requires customer:write scope)
	customers.Post("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerWrite),
		customerHandler.CreateCustomer,
	)

	// Get customer by ID (authenticated, requires customer:read scope)
	customers.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerRead),
		customerHandler.GetCustomer,
	)

	// Update customer (authenticated, requires customer:write scope)
	customers.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerWrite),
		customerHandler.UpdateCustomer,
	)

	// Delete customer (authenticated, requires customer:write scope)
	customers.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerWrite),
		customerHandler.DeleteCustomer,
	)

	// List customers (authenticated, requires customer:read scope)
	customers.Get("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerRead),
		customerHandler.ListCustomers,
	)

	// ============================================================================
	// Customer Lookup Operations
	// ============================================================================

	// Get customer by user ID (authenticated, requires customer:read scope)
	customers.Get("/user/:user_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerRead),
		customerHandler.GetCustomerByUserID,
	)

	// Search customers (authenticated, requires customer:read scope)
	customers.Post("/search",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerRead),
		customerHandler.SearchCustomers,
	)

	// Get active customers (authenticated, requires customer:read scope)
	customers.Get("/active",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerRead),
		customerHandler.GetActiveCustomers,
	)

	// Get top customers (authenticated, requires customer:read scope)
	customers.Get("/top",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerRead),
		customerHandler.GetTopCustomers,
	)

	// ============================================================================
	// Loyalty Program
	// ============================================================================

	// Update loyalty points (authenticated, requires customer:write scope)
	customers.Put("/:id/loyalty-points",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerWrite),
		customerHandler.UpdateLoyaltyPoints,
	)

	// Get loyalty points history (authenticated, requires customer:read scope)
	customers.Get("/:id/loyalty-history",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerRead),
		customerHandler.GetLoyaltyPointsHistory,
	)

	// ============================================================================
	// Preferences & Settings
	// ============================================================================

	// Add preferred artisan (authenticated, requires customer:write scope)
	customers.Post("/:id/preferred-artisans",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerWrite),
		customerHandler.AddPreferredArtisan,
	)

	// Remove preferred artisan (authenticated, requires customer:write scope)
	customers.Delete("/:id/preferred-artisans/:artisan_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerWrite),
		customerHandler.RemovePreferredArtisan,
	)

	// Update notification preferences (authenticated, requires customer:write scope)
	customers.Put("/:id/notification-preferences",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerWrite),
		customerHandler.UpdateNotificationPreferences,
	)

	// Update primary location (authenticated, requires customer:write scope)
	customers.Put("/:id/primary-location",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerWrite),
		customerHandler.UpdatePrimaryLocation,
	)

	// ============================================================================
	// Analytics & Segmentation
	// ============================================================================

	// Get customer statistics (authenticated, requires customer:read scope)
	customers.Get("/:id/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerRead),
		customerHandler.GetCustomerStats,
	)

	// Get customer segments (authenticated, requires customer:read scope)
	customers.Get("/segments",
		authMiddleware,
		middleware.RequireScopes(r.scopes.CustomerRead),
		customerHandler.GetCustomerSegments,
	)
}
