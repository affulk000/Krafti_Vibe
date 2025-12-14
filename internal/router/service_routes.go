package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
)

func (r *Router) setupServiceRoutes(api fiber.Router) {
	// Initialize service
	serviceService := service.NewServiceService(r.repos.Service, r.repos.Tenant, r.repos.User, r.config.Logger)
	serviceHandler := handler.NewServiceHandler(serviceService)

	// Create service catalog routes
	services := api.Group("/services")

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// CRUD Operations
	// ============================================================================

	// Create service
	services.Post("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ServiceWrite),
		serviceHandler.CreateService,
	)

	// Get service by ID (read-only)
	services.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ServiceRead),
		serviceHandler.GetServiceByID,
	)

	// Update service
	services.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ServiceWrite),
		serviceHandler.UpdateService,
	)

	// Delete service
	services.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ServiceWrite),
		serviceHandler.DeleteService,
	)

	// ============================================================================
	// Query Operations
	// ============================================================================

	// List services (with pagination)
	services.Post("/list",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ServiceRead),
		serviceHandler.ListServices,
	)

	// Search services
	services.Get("/search",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ServiceRead),
		serviceHandler.SearchServices,
	)

	// ============================================================================
	// Status Management
	// ============================================================================

	// Activate service
	services.Post("/:id/activate",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ServiceWrite),
		serviceHandler.ActivateService,
	)

	// Deactivate service
	services.Post("/:id/deactivate",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ServiceWrite),
		serviceHandler.DeactivateService,
	)

	// ============================================================================
	// Analytics & Statistics
	// ============================================================================

	// Get service statistics
	services.Get("/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ServiceRead),
		serviceHandler.GetServiceStatistics,
	)

	// Get popular services
	services.Get("/popular",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ServiceRead),
		serviceHandler.GetPopularServices,
	)
}
