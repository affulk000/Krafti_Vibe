package router

import (
	"Krafti_Vibe/internal/handler"
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

	// ============================================================================
	// CRUD Operations
	// ============================================================================

	// Create service
	services.Post("",
		r.RequireAuth(),
		serviceHandler.CreateService,
	)

	// Get service by ID (read-only)
	services.Get("/:id",
		r.RequireAuth(),
		serviceHandler.GetServiceByID,
	)

	// Update service
	services.Put("/:id",
		r.RequireAuth(),
		serviceHandler.UpdateService,
	)

	// Delete service
	services.Delete("/:id",
		r.RequireAuth(),
		serviceHandler.DeleteService,
	)

	// ============================================================================
	// Query Operations
	// ============================================================================

	// List services (with pagination)
	services.Post("/list",
		r.RequireAuth(),
		serviceHandler.ListServices,
	)

	// Search services
	services.Get("/search",
		r.RequireAuth(),
		serviceHandler.SearchServices,
	)

	// ============================================================================
	// Status Management
	// ============================================================================

	// Activate service
	services.Post("/:id/activate",
		r.RequireAuth(),
		serviceHandler.ActivateService,
	)

	// Deactivate service
	services.Post("/:id/deactivate",
		r.RequireAuth(),
		serviceHandler.DeactivateService,
	)

	// ============================================================================
	// Analytics & Statistics
	// ============================================================================

	// Get service statistics
	services.Get("/stats",
		r.RequireAuth(),
		serviceHandler.GetServiceStatistics,
	)

	// Get popular services
	services.Get("/popular",
		r.RequireAuth(),
		serviceHandler.GetPopularServices,
	)
}
