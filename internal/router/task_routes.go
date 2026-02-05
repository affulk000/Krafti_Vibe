package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *Router) setupTaskRoutes(api fiber.Router) {
	// Initialize service and handler
	taskService := service.NewTaskService(r.repos, r.config.Logger)
	taskHandler := handler.NewTaskHandler(taskService)

	// Create tasks group
	tasks := api.Group("/tasks")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger = zap.NewNop()
		}
		tasks.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// Auth middleware configuration

	// ============================================================================
	// Core Task Operations
	// ============================================================================

	// Create task (authenticated, requires task:write scope)
	tasks.Post("/",
		r.RequireAuth(),
		taskHandler.CreateTask,
	)

	// Get task by ID (authenticated, requires task:read scope)
	tasks.Get("/:id",
		r.RequireAuth(),
		taskHandler.GetTask,
	)

	// Update task (authenticated, requires task:write scope)
	tasks.Put("/:id",
		r.RequireAuth(),
		taskHandler.UpdateTask,
	)

	// Delete task (authenticated, requires task:write scope)
	tasks.Delete("/:id",
		r.RequireAuth(),
		taskHandler.DeleteTask,
	)

	// ============================================================================
	// Task Actions
	// ============================================================================

	// Complete task (authenticated, requires task:write scope)
	tasks.Post("/:id/complete",
		r.RequireAuth(),
		taskHandler.CompleteTask,
	)

	// ============================================================================
	// Related Resource Queries
	// ============================================================================

	// Get tasks by milestone (authenticated, requires task:read scope)
	tasks.Get("/milestone/:milestone_id",
		r.RequireAuth(),
		taskHandler.GetMilestoneTasks,
	)
}
