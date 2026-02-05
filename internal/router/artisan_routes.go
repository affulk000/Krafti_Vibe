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
	artisans.Use(r.RequireAuth())

	// ============================================================================
	// Core Artisan Operations
	// ============================================================================

	// Create artisan - tenant owner/admin can create artisans
	artisans.Post("/",
		middleware.RequireTenantOwnerOrAdmin(),
		artisanHandler.CreateArtisan,
	)

	// Get artisan by ID - any authenticated user can view artisan profiles
	artisans.Get("/:id",
		artisanHandler.GetArtisan,
	)

	// Update artisan - self (artisan) or tenant owner/admin
	artisans.Put("/:id",
		middleware.RequireSelfOrAdmin(),
		artisanHandler.UpdateArtisan,
	)

	// Delete artisan - tenant owner/admin only
	artisans.Delete("/:id",
		middleware.RequireTenantOwnerOrAdmin(),
		artisanHandler.DeleteArtisan,
	)

	// List artisans - any authenticated user (for customer discovery)
	artisans.Get("/",
		artisanHandler.ListArtisans,
	)

	// ============================================================================
	// Artisan Lookup Operations
	// ============================================================================

	// Get artisan by user ID - any authenticated user
	artisans.Get("/user/:user_id",
		artisanHandler.GetArtisanByUserID,
	)

	// Search artisans - any authenticated user
	artisans.Post("/search",
		artisanHandler.SearchArtisans,
	)

	// ============================================================================
	// Artisan Discovery
	// ============================================================================

	// Get available artisans - any authenticated user (for booking)
	artisans.Get("/available",
		artisanHandler.GetAvailableArtisans,
	)

	// Get artisans by specialization - any authenticated user
	artisans.Get("/specialization/:specialization",
		artisanHandler.GetArtisansBySpecialization,
	)

	// Get top rated artisans - any authenticated user
	artisans.Get("/top-rated",
		artisanHandler.GetTopRatedArtisans,
	)

	// Find nearby artisans - any authenticated user
	artisans.Post("/nearby",
		artisanHandler.FindNearbyArtisans,
	)

	// ============================================================================
	// Availability Management
	// ============================================================================

	// Update availability - artisan (self) or tenant owner/admin
	artisans.Put("/:id/availability",
		middleware.RequireArtisanOrTeamMember(),
		artisanHandler.UpdateAvailability,
	)

	// Batch update availability - artisan or tenant owner/admin
	artisans.Post("/availability/batch",
		middleware.RequireArtisanOrTeamMember(),
		artisanHandler.BatchUpdateAvailability,
	)

	// ============================================================================
	// Statistics & Analytics
	// ============================================================================

	// Get artisan statistics - artisan (self) or tenant owner/admin
	artisans.Get("/:id/stats",
		middleware.RequireArtisanOrTeamMember(),
		artisanHandler.GetArtisanStats,
	)

	// Get dashboard statistics - artisan (self) or tenant owner/admin
	artisans.Get("/:id/dashboard",
		middleware.RequireArtisanOrTeamMember(),
		artisanHandler.GetDashboardStats,
	)
}
