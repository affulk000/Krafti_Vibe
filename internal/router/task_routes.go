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
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core Task Operations
	// ============================================================================

	// Create task (authenticated, requires task:write scope)
	tasks.Post("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TaskWrite),
		taskHandler.CreateTask,
	)

	// Get task by ID (authenticated, requires task:read scope)
	tasks.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TaskRead),
		taskHandler.GetTask,
	)

	// Update task (authenticated, requires task:write scope)
	tasks.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TaskWrite),
		taskHandler.UpdateTask,
	)

	// Delete task (authenticated, requires task:write scope)
	tasks.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TaskWrite),
		taskHandler.DeleteTask,
	)

	// ============================================================================
	// Task Actions
	// ============================================================================

	// Complete task (authenticated, requires task:write scope)
	tasks.Post("/:id/complete",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TaskWrite),
		taskHandler.CompleteTask,
	)

	// ============================================================================
	// Related Resource Queries
	// ============================================================================

	// Get tasks by milestone (authenticated, requires task:read scope)
	tasks.Get("/milestone/:milestone_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.TaskRead),
		taskHandler.GetMilestoneTasks,
	)
}
