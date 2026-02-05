package router

import (
	"Krafti_Vibe/internal/handler"
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

	// ============================================================================
	// Core CRUD Operations
	// ============================================================================

	// Create availability
	availability.Post("",
		r.RequireAuth(),
		availabilityHandler.CreateAvailability,
	)

	// List availabilities (with filters)
	availability.Get("",
		r.RequireAuth(),
		availabilityHandler.ListAvailabilities,
	)

	// Get availability by ID
	availability.Get("/:id",
		r.RequireAuth(),
		availabilityHandler.GetAvailability,
	)

	// Update availability
	availability.Put("/:id",
		r.RequireAuth(),
		availabilityHandler.UpdateAvailability,
	)

	// Delete availability
	availability.Delete("/:id",
		r.RequireAuth(),
		availabilityHandler.DeleteAvailability,
	)

	// ============================================================================
	// Availability Checks
	// ============================================================================

	// Check availability for a time slot
	availability.Post("/check",
		r.RequireAuth(),
		availabilityHandler.CheckAvailability,
	)

	// ============================================================================
	// Bulk Operations
	// ============================================================================

	// Bulk create availability
	availability.Post("/bulk",
		r.RequireAuth(),
		availabilityHandler.BulkCreateAvailability,
	)

	// ============================================================================
	// Artisan-Specific Query Operations
	// ============================================================================

	// Get weekly schedule for artisan
	availability.Get("/artisan/:artisan_id/weekly",
		r.RequireAuth(),
		availabilityHandler.GetWeeklySchedule,
	)

	// Get availability by day of week
	availability.Get("/artisan/:artisan_id/day/:day",
		r.RequireAuth(),
		availabilityHandler.GetByDayOfWeek,
	)

	// Get availability by type
	availability.Get("/artisan/:artisan_id/type/:type",
		r.RequireAuth(),
		availabilityHandler.ListByType,
	)

	// Delete availability by type
	availability.Delete("/artisan/:artisan_id/type/:type",
		r.RequireAuth(),
		availabilityHandler.DeleteByType,
	)
}
