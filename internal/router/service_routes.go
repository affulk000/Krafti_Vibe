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
		r.zitadelMW.RequireAuth(),
		serviceHandler.CreateService,
	)

	// Get service by ID (read-only)
	services.Get("/:id",
		r.zitadelMW.RequireAuth(),
		serviceHandler.GetServiceByID,
	)

	// Update service
	services.Put("/:id",
		r.zitadelMW.RequireAuth(),
		serviceHandler.UpdateService,
	)

	// Delete service
	services.Delete("/:id",
		r.zitadelMW.RequireAuth(),
		serviceHandler.DeleteService,
	)

	// ============================================================================
	// Query Operations
	// ============================================================================

	// List services (with pagination)
	services.Post("/list",
		r.zitadelMW.RequireAuth(),
		serviceHandler.ListServices,
	)

	// Search services
	services.Get("/search",
		r.zitadelMW.RequireAuth(),
		serviceHandler.SearchServices,
	)

	// ============================================================================
	// Status Management
	// ============================================================================

	// Activate service
	services.Post("/:id/activate",
		r.zitadelMW.RequireAuth(),
		serviceHandler.ActivateService,
	)

	// Deactivate service
	services.Post("/:id/deactivate",
		r.zitadelMW.RequireAuth(),
		serviceHandler.DeactivateService,
	)

	// ============================================================================
	// Analytics & Statistics
	// ============================================================================

	// Get service statistics
	services.Get("/stats",
		r.zitadelMW.RequireAuth(),
		serviceHandler.GetServiceStatistics,
	)

	// Get popular services
	services.Get("/popular",
		r.zitadelMW.RequireAuth(),
		serviceHandler.GetPopularServices,
	)
}
