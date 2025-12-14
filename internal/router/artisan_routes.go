package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *Router) setupArtisanRoutes(api fiber.Router) {
	// Initialize service and handler
	artisanService := service.NewArtisanService(r.repos, r.config.Logger)
	artisanHandler := handler.NewArtisanHandler(artisanService)

	// Create artisans group
	artisans := api.Group("/artisans")

	// Apply rate limiting if cache is available
	if r.config.Cache != nil {
		zapLogger := r.config.ZapLogger
		if zapLogger == nil {
			zapLogger = zap.NewNop()
		}
		artisans.Use(middleware.RateLimitWithHeaders(middleware.DefaultRateLimitConfig(r.config.Cache, zapLogger)))
	}

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core Artisan Operations
	// ============================================================================

	// Create artisan (authenticated, requires artisan:write scope)
	artisans.Post("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanWrite),
		artisanHandler.CreateArtisan,
	)

	// Get artisan by ID (authenticated, requires artisan:read scope)
	artisans.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		artisanHandler.GetArtisan,
	)

	// Update artisan (authenticated, requires artisan:write scope)
	artisans.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanWrite),
		artisanHandler.UpdateArtisan,
	)

	// Delete artisan (authenticated, requires artisan:write scope)
	artisans.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanWrite),
		artisanHandler.DeleteArtisan,
	)

	// List artisans (authenticated, requires artisan:read scope)
	artisans.Get("/",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		artisanHandler.ListArtisans,
	)

	// ============================================================================
	// Artisan Lookup Operations
	// ============================================================================

	// Get artisan by user ID (authenticated, requires artisan:read scope)
	artisans.Get("/user/:user_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		artisanHandler.GetArtisanByUserID,
	)

	// Search artisans (authenticated, requires artisan:read scope)
	artisans.Post("/search",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		artisanHandler.SearchArtisans,
	)

	// ============================================================================
	// Artisan Discovery
	// ============================================================================

	// Get available artisans (authenticated, requires artisan:read scope)
	artisans.Get("/available",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		artisanHandler.GetAvailableArtisans,
	)

	// Get artisans by specialization (authenticated, requires artisan:read scope)
	artisans.Get("/specialization/:specialization",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		artisanHandler.GetArtisansBySpecialization,
	)

	// Get top rated artisans (authenticated, requires artisan:read scope)
	artisans.Get("/top-rated",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		artisanHandler.GetTopRatedArtisans,
	)

	// Find nearby artisans (authenticated, requires artisan:read scope)
	artisans.Post("/nearby",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		artisanHandler.FindNearbyArtisans,
	)

	// ============================================================================
	// Availability Management
	// ============================================================================

	// Update availability (authenticated, requires artisan:write scope)
	artisans.Put("/:id/availability",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanWrite),
		artisanHandler.UpdateAvailability,
	)

	// Batch update availability (authenticated, requires artisan:write scope)
	artisans.Post("/availability/batch",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanWrite),
		artisanHandler.BatchUpdateAvailability,
	)

	// ============================================================================
	// Statistics & Analytics
	// ============================================================================

	// Get artisan statistics (authenticated, requires artisan:read scope)
	artisans.Get("/:id/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		artisanHandler.GetArtisanStats,
	)

	// Get dashboard statistics (authenticated, requires artisan:read scope)
	artisans.Get("/:id/dashboard",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ArtisanRead),
		artisanHandler.GetDashboardStats,
	)
}
