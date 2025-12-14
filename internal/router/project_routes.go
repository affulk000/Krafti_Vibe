package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
)

func (r *Router) setupProjectRoutes(api fiber.Router) {
	// Initialize service
	projectService := service.NewProjectService(r.repos, r.config.Logger)
	projectHandler := handler.NewProjectHandler(projectService)

	// Create project routes
	projects := api.Group("/projects")

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// CRUD Operations
	// ============================================================================

	// Create project
	projects.Post("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectWrite),
		projectHandler.CreateProject,
	)

	// Get project (basic info)
	projects.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		projectHandler.GetProject,
	)

	// Get project with full details
	projects.Get("/:id/details",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		projectHandler.GetProjectWithDetails,
	)

	// Update project
	projects.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectWrite),
		projectHandler.UpdateProject,
	)

	// Delete project
	projects.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectWrite),
		projectHandler.DeleteProject,
	)

	// ============================================================================
	// Query Operations
	// ============================================================================

	// List projects (with pagination and filters)
	projects.Get("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		projectHandler.ListProjects,
	)

	// Search projects
	projects.Get("/search",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		projectHandler.SearchProjects,
	)

	// Get projects by artisan
	projects.Get("/artisan/:artisan_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		projectHandler.GetProjectsByArtisan,
	)

	// Get projects by customer
	projects.Get("/customer/:customer_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		projectHandler.GetProjectsByCustomer,
	)

	// Get overdue projects
	projects.Get("/overdue",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		projectHandler.GetOverdueProjects,
	)

	// Get active projects
	projects.Get("/active",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		projectHandler.GetActiveProjects,
	)

	// ============================================================================
	// Status Management
	// ============================================================================

	// Start project
	projects.Post("/:id/start",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectWrite),
		projectHandler.StartProject,
	)

	// Pause project
	projects.Post("/:id/pause",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectWrite),
		projectHandler.PauseProject,
	)

	// Resume project
	projects.Post("/:id/resume",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectWrite),
		projectHandler.ResumeProject,
	)

	// Complete project
	projects.Post("/:id/complete",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectWrite),
		projectHandler.CompleteProject,
	)

	// Cancel project
	projects.Post("/:id/cancel",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectWrite),
		projectHandler.CancelProject,
	)

	// ============================================================================
	// Analytics & Statistics
	// ============================================================================

	// Get project statistics
	projects.Get("/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		projectHandler.GetProjectStats,
	)

	// Get artisan project statistics
	projects.Get("/artisan/:artisan_id/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		projectHandler.GetArtisanProjectStats,
	)

	// Get project health
	projects.Get("/:id/health",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		projectHandler.GetProjectHealth,
	)

	// Get project timeline
	projects.Get("/:id/timeline",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		projectHandler.GetProjectTimeline,
	)

	// ============================================================================
	// Dashboard
	// ============================================================================

	// Get artisan dashboard
	projects.Get("/artisan/:artisan_id/dashboard",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		projectHandler.GetArtisanDashboard,
	)

	// Get tenant dashboard
	projects.Get("/dashboard",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectRead),
		projectHandler.GetTenantDashboard,
	)

	// ============================================================================
	// Bulk Operations
	// ============================================================================

	// Bulk update project status
	projects.Put("/bulk/status",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectWrite),
		projectHandler.BulkUpdateStatus,
	)

	// Archive completed projects
	projects.Post("/archive",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ProjectWrite),
		projectHandler.ArchiveCompletedProjects,
	)
}
