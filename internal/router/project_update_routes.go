package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
)

// setupProjectUpdateRoutes sets up project update routes
func (r *Router) setupProjectUpdateRoutes(api fiber.Router) {
	// Initialize project update service
	updateService := service.NewProjectUpdateService(r.repos, r.config.Logger)

	// Initialize project update handler
	updateHandler := handler.NewProjectUpdateHandler(updateService)

	// Project update routes group
	updates := api.Group("/project-updates")

	// Auth middleware configuration

	// ============================================================================
	// Core CRUD Operations
	// ============================================================================

	// Create project update
	updates.Post("",
		r.RequireAuth(),
		updateHandler.CreateProjectUpdate,
	)

	// List project updates (with filters)
	updates.Get("",
		r.RequireAuth(),
		updateHandler.ListProjectUpdates,
	)

	// Get project update by ID
	updates.Get("/:id",
		r.RequireAuth(),
		updateHandler.GetProjectUpdate,
	)

	// Update project update
	updates.Put("/:id",
		r.RequireAuth(),
		updateHandler.UpdateProjectUpdate,
	)

	// Delete project update
	updates.Delete("/:id",
		r.RequireAuth(),
		updateHandler.DeleteProjectUpdate,
	)

	// ============================================================================
	// Project-Specific Query Operations
	// ============================================================================

	// Get latest update for a project
	updates.Get("/project/:project_id/latest",
		r.RequireAuth(),
		updateHandler.GetLatestUpdate,
	)

	// List customer-visible updates for a project
	updates.Get("/project/:project_id/customer-visible",
		r.RequireAuth(),
		updateHandler.ListCustomerVisible,
	)

	// List updates by type for a project
	updates.Get("/project/:project_id/type/:type",
		r.RequireAuth(),
		updateHandler.ListByType,
	)
}
