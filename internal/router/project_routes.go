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
	projects.Use(r.zitadelMW.RequireAuth())

	// ============================================================================
	// CRUD Operations
	// ============================================================================

	// Create project - artisan, customer, or tenant owner/admin
	projects.Post("",
		projectHandler.CreateProject,
	)

	// Get project (basic info) - owner (artisan/customer) or tenant owner/admin
	projects.Get("/:id",
		projectHandler.GetProject,
	)

	// Get project with full details - owner (artisan/customer) or tenant owner/admin
	projects.Get("/:id/details",
		projectHandler.GetProjectWithDetails,
	)

	// Update project - artisan (assigned) or tenant owner/admin
	projects.Put("/:id",
		middleware.RequireArtisanOrTeamMember(),
		projectHandler.UpdateProject,
	)

	// Delete project - tenant owner/admin only
	projects.Delete("/:id",
		middleware.RequireTenantOwnerOrAdmin(),
		projectHandler.DeleteProject,
	)

	// ============================================================================
	// Query Operations
	// ============================================================================

	// List projects - tenant owner/admin only
	projects.Get("",
		middleware.RequireTenantOwnerOrAdmin(),
		projectHandler.ListProjects,
	)

	// Search projects - tenant owner/admin only
	projects.Get("/search",
		middleware.RequireTenantOwnerOrAdmin(),
		projectHandler.SearchProjects,
	)

	// Get projects by artisan - artisan (self) or tenant owner/admin
	projects.Get("/artisan/:artisan_id",
		middleware.RequireArtisanOrTeamMember(),
		projectHandler.GetProjectsByArtisan,
	)

	// Get projects by customer - customer (self) or tenant owner/admin
	projects.Get("/customer/:customer_id",
		projectHandler.GetProjectsByCustomer,
	)

	// Get overdue projects - tenant owner/admin only
	projects.Get("/overdue",
		middleware.RequireTenantOwnerOrAdmin(),
		projectHandler.GetOverdueProjects,
	)

	// Get active projects - tenant owner/admin only
	projects.Get("/active",
		middleware.RequireTenantOwnerOrAdmin(),
		projectHandler.GetActiveProjects,
	)

	// ============================================================================
	// Status Management
	// ============================================================================

	// Start project - artisan (assigned) or tenant owner/admin
	projects.Post("/:id/start",
		middleware.RequireArtisanOrTeamMember(),
		projectHandler.StartProject,
	)

	// Pause project - artisan (assigned) or tenant owner/admin
	projects.Post("/:id/pause",
		middleware.RequireArtisanOrTeamMember(),
		projectHandler.PauseProject,
	)

	// Resume project - artisan (assigned) or tenant owner/admin
	projects.Post("/:id/resume",
		middleware.RequireArtisanOrTeamMember(),
		projectHandler.ResumeProject,
	)

	// Complete project - artisan (assigned) or tenant owner/admin
	projects.Post("/:id/complete",
		middleware.RequireArtisanOrTeamMember(),
		projectHandler.CompleteProject,
	)

	// Cancel project - artisan (assigned), customer (owner), or tenant owner/admin
	projects.Post("/:id/cancel",
		projectHandler.CancelProject,
	)

	// ============================================================================
	// Analytics & Statistics
	// ============================================================================

	// Get project statistics - tenant owner/admin only
	projects.Get("/stats",
		middleware.RequireTenantOwnerOrAdmin(),
		projectHandler.GetProjectStats,
	)

	// Get artisan project statistics - artisan (self) or tenant owner/admin
	projects.Get("/artisan/:artisan_id/stats",
		middleware.RequireArtisanOrTeamMember(),
		projectHandler.GetArtisanProjectStats,
	)

	// Get project health - owner (artisan/customer) or tenant owner/admin
	projects.Get("/:id/health",
		projectHandler.GetProjectHealth,
	)

	// Get project timeline - owner (artisan/customer) or tenant owner/admin
	projects.Get("/:id/timeline",
		projectHandler.GetProjectTimeline,
	)

	// ============================================================================
	// Dashboard
	// ============================================================================

	// Get artisan dashboard - artisan (self) or tenant owner/admin
	projects.Get("/artisan/:artisan_id/dashboard",
		middleware.RequireArtisanOrTeamMember(),
		projectHandler.GetArtisanDashboard,
	)

	// Get tenant dashboard - tenant owner/admin only
	projects.Get("/dashboard",
		middleware.RequireTenantOwnerOrAdmin(),
		projectHandler.GetTenantDashboard,
	)

	// ============================================================================
	// Bulk Operations
	// ============================================================================

	// Bulk update project status - tenant owner/admin only
	projects.Put("/bulk/status",
		middleware.RequireTenantOwnerOrAdmin(),
		projectHandler.BulkUpdateStatus,
	)

	// Archive completed projects - tenant owner/admin only
	projects.Post("/archive",
		middleware.RequireTenantOwnerOrAdmin(),
		projectHandler.ArchiveCompletedProjects,
	)
}
