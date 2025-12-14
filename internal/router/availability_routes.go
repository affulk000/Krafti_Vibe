package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
)

// setupAvailabilityRoutes sets up availability routes
func (r *Router) setupAvailabilityRoutes(api fiber.Router) {
	// Initialize availability service
	availabilityService := service.NewAvailabilityService(r.repos, r.config.Logger)

	// Initialize availability handler
	availabilityHandler := handler.NewAvailabilityHandler(availabilityService)

	// Availability routes group
	availability := api.Group("/availability")

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core CRUD Operations
	// ============================================================================

	// Create availability
	availability.Post("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanWrite),
		availabilityHandler.CreateAvailability,
	)

	// List availabilities (with filters)
	availability.Get("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		availabilityHandler.ListAvailabilities,
	)

	// Get availability by ID
	availability.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		availabilityHandler.GetAvailability,
	)

	// Update availability
	availability.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanWrite),
		availabilityHandler.UpdateAvailability,
	)

	// Delete availability
	availability.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanWrite),
		availabilityHandler.DeleteAvailability,
	)

	// ============================================================================
	// Availability Checks
	// ============================================================================

	// Check availability for a time slot
	availability.Post("/check",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		availabilityHandler.CheckAvailability,
	)

	// ============================================================================
	// Bulk Operations
	// ============================================================================

	// Bulk create availability
	availability.Post("/bulk",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanWrite),
		availabilityHandler.BulkCreateAvailability,
	)

	// ============================================================================
	// Artisan-Specific Query Operations
	// ============================================================================

	// Get weekly schedule for artisan
	availability.Get("/artisan/:artisan_id/weekly",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		availabilityHandler.GetWeeklySchedule,
	)

	// Get availability by day of week
	availability.Get("/artisan/:artisan_id/day/:day",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		availabilityHandler.GetByDayOfWeek,
	)

	// Get availability by type
	availability.Get("/artisan/:artisan_id/type/:type",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		availabilityHandler.ListByType,
	)

	// Delete availability by type
	availability.Delete("/artisan/:artisan_id/type/:type",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanManage),
		availabilityHandler.DeleteByType,
	)
}
