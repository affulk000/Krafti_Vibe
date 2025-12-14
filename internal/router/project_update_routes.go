package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
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
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core CRUD Operations
	// ============================================================================

	// Create project update
	updates.Post("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectWrite),
		updateHandler.CreateProjectUpdate,
	)

	// List project updates (with filters)
	updates.Get("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		updateHandler.ListProjectUpdates,
	)

	// Get project update by ID
	updates.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		updateHandler.GetProjectUpdate,
	)

	// Update project update
	updates.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectWrite),
		updateHandler.UpdateProjectUpdate,
	)

	// Delete project update
	updates.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectWrite),
		updateHandler.DeleteProjectUpdate,
	)

	// ============================================================================
	// Project-Specific Query Operations
	// ============================================================================

	// Get latest update for a project
	updates.Get("/project/:project_id/latest",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		updateHandler.GetLatestUpdate,
	)

	// List customer-visible updates for a project
	updates.Get("/project/:project_id/customer-visible",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		updateHandler.ListCustomerVisible,
	)

	// List updates by type for a project
	updates.Get("/project/:project_id/type/:type",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		updateHandler.ListByType,
	)
}
