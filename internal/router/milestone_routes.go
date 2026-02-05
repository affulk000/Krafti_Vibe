package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *Router) setupMilestoneRoutes(api fiber.Router) {
	// Initialize service and handler
	milestoneService := service.NewMilestoneService(r.repos, r.config.Logger)
	milestoneHandler := handler.NewMilestoneHandler(milestoneService)

	// Create milestones group
	milestones := api.Group("/milestones")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger = zap.NewNop()
		}
		milestones.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// Auth middleware configuration

	// ============================================================================
	// Core Milestone Operations
	// ============================================================================

	// Create milestone (authenticated, requires milestone:write scope)
	milestones.Post("/",
		r.RequireAuth(),
		milestoneHandler.CreateMilestone,
	)

	// Get milestone by ID (authenticated, requires milestone:read scope)
	milestones.Get("/:id",
		r.RequireAuth(),
		milestoneHandler.GetMilestone,
	)

	// Update milestone (authenticated, requires milestone:write scope)
	milestones.Put("/:id",
		r.RequireAuth(),
		milestoneHandler.UpdateMilestone,
	)

	// Delete milestone (authenticated, requires milestone:write scope)
	milestones.Delete("/:id",
		r.RequireAuth(),
		milestoneHandler.DeleteMilestone,
	)

	// ============================================================================
	// Milestone Actions
	// ============================================================================

	// Complete milestone (authenticated, requires milestone:write scope)
	milestones.Post("/:id/complete",
		r.RequireAuth(),
		milestoneHandler.CompleteMilestone,
	)

	// ============================================================================
	// Related Resource Queries
	// ============================================================================

	// Get milestones by project (authenticated, requires milestone:read scope)
	milestones.Get("/project/:project_id",
		r.RequireAuth(),
		milestoneHandler.GetProjectMilestones,
	)
}
