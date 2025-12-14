package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
)

// setupPromoRoutes sets up promotional code routes
func (r *Router) setupPromoRoutes(api fiber.Router) {
	// Initialize promo code service
	promoService := service.NewPromoCodeService(r.repos, r.config.Logger)

	// Initialize promo code handler
	promoHandler := handler.NewPromoCodeHandler(promoService)

	// Promo code routes group
	promos := api.Group("/promo-codes")

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// Core CRUD Operations
	// ============================================================================

	// Create promo code
	promos.Post("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoWrite),
		promoHandler.CreatePromoCode,
	)

	// List promo codes
	promos.Get("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoRead),
		promoHandler.ListPromoCodes,
	)

	// Search promo codes
	promos.Get("/search",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoRead),
		promoHandler.SearchPromoCodes,
	)

	// Get promo code by ID
	promos.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoRead),
		promoHandler.GetPromoCode,
	)

	// Get promo code by code
	promos.Get("/code/:code",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoRead),
		promoHandler.GetPromoCodeByCode,
	)

	// Update promo code
	promos.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoWrite),
		promoHandler.UpdatePromoCode,
	)

	// Delete promo code
	promos.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoDelete),
		promoHandler.DeletePromoCode,
	)

	// ============================================================================
	// Status-Based Queries
	// ============================================================================

	// Get active promo codes
	promos.Get("/active",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoRead),
		promoHandler.GetActivePromoCodes,
	)

	// Get expired promo codes
	promos.Get("/expired",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoRead),
		promoHandler.GetExpiredPromoCodes,
	)

	// Get expiring promo codes
	promos.Get("/expiring",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoRead),
		promoHandler.GetExpiringPromoCodes,
	)

	// ============================================================================
	// Validation & Application
	// ============================================================================

	// Validate promo code
	promos.Post("/validate",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoApply),
		promoHandler.ValidatePromoCode,
	)

	// Apply promo code
	promos.Post("/apply",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoApply),
		promoHandler.ApplyPromoCode,
	)

	// ============================================================================
	// Status Management
	// ============================================================================

	// Activate promo code
	promos.Post("/:id/activate",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoWrite),
		promoHandler.ActivatePromoCode,
	)

	// Deactivate promo code
	promos.Post("/:id/deactivate",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoWrite),
		promoHandler.DeactivatePromoCode,
	)

	// ============================================================================
	// Bulk Operations
	// ============================================================================

	// Bulk activate promo codes
	promos.Post("/bulk/activate",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoWrite),
		promoHandler.BulkActivate,
	)

	// Bulk deactivate promo codes
	promos.Post("/bulk/deactivate",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoWrite),
		promoHandler.BulkDeactivate,
	)

	// Bulk delete promo codes
	promos.Delete("/bulk",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoDelete),
		promoHandler.BulkDelete,
	)

	// ============================================================================
	// Entity-Specific Queries
	// ============================================================================

	// Get valid promo codes for service
	promos.Get("/service/:service_id/valid",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoRead),
		promoHandler.GetValidPromoCodesForService,
	)

	// Get valid promo codes for artisan
	promos.Get("/artisan/:artisan_id/valid",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoRead),
		promoHandler.GetValidPromoCodesForArtisan,
	)

	// ============================================================================
	// Analytics & Statistics
	// ============================================================================

	// Get promo code statistics
	promos.Get("/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoRead),
		promoHandler.GetPromoCodeStats,
	)

	// Get top performing promo codes
	promos.Get("/analytics/top-performing",
		authMiddleware,
		middleware.RequireScopes(r.scopes.PromoRead),
		promoHandler.GetTopPerformingPromoCodes,
	)
}
