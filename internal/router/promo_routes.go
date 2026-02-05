package router

import (
	"Krafti_Vibe/internal/handler"
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

	// ============================================================================
	// Core CRUD Operations
	// ============================================================================

	// Create promo code
	promos.Post("",
		r.RequireAuth(),
		promoHandler.CreatePromoCode,
	)

	// List promo codes
	promos.Get("",
		r.RequireAuth(),
		promoHandler.ListPromoCodes,
	)

	// Search promo codes
	promos.Get("/search",
		r.RequireAuth(),
		promoHandler.SearchPromoCodes,
	)

	// Get promo code by ID
	promos.Get("/:id",
		r.RequireAuth(),
		promoHandler.GetPromoCode,
	)

	// Get promo code by code
	promos.Get("/code/:code",
		r.RequireAuth(),
		promoHandler.GetPromoCodeByCode,
	)

	// Update promo code
	promos.Put("/:id",
		r.RequireAuth(),
		promoHandler.UpdatePromoCode,
	)

	// Delete promo code
	promos.Delete("/:id",
		r.RequireAuth(),
		promoHandler.DeletePromoCode,
	)

	// ============================================================================
	// Status-Based Queries
	// ============================================================================

	// Get active promo codes
	promos.Get("/active",
		r.RequireAuth(),
		promoHandler.GetActivePromoCodes,
	)

	// Get expired promo codes
	promos.Get("/expired",
		r.RequireAuth(),
		promoHandler.GetExpiredPromoCodes,
	)

	// Get expiring promo codes
	promos.Get("/expiring",
		r.RequireAuth(),
		promoHandler.GetExpiringPromoCodes,
	)

	// ============================================================================
	// Validation & Application
	// ============================================================================

	// Validate promo code
	promos.Post("/validate",
		r.RequireAuth(),
		promoHandler.ValidatePromoCode,
	)

	// Apply promo code
	promos.Post("/apply",
		r.RequireAuth(),
		promoHandler.ApplyPromoCode,
	)

	// ============================================================================
	// Status Management
	// ============================================================================

	// Activate promo code
	promos.Post("/:id/activate",
		r.RequireAuth(),
		promoHandler.ActivatePromoCode,
	)

	// Deactivate promo code
	promos.Post("/:id/deactivate",
		r.RequireAuth(),
		promoHandler.DeactivatePromoCode,
	)

	// ============================================================================
	// Bulk Operations
	// ============================================================================

	// Bulk activate promo codes
	promos.Post("/bulk/activate",
		r.RequireAuth(),
		promoHandler.BulkActivate,
	)

	// Bulk deactivate promo codes
	promos.Post("/bulk/deactivate",
		r.RequireAuth(),
		promoHandler.BulkDeactivate,
	)

	// Bulk delete promo codes
	promos.Delete("/bulk",
		r.RequireAuth(),
		promoHandler.BulkDelete,
	)

	// ============================================================================
	// Entity-Specific Queries
	// ============================================================================

	// Get valid promo codes for service
	promos.Get("/service/:service_id/valid",
		r.RequireAuth(),
		promoHandler.GetValidPromoCodesForService,
	)

	// Get valid promo codes for artisan
	promos.Get("/artisan/:artisan_id/valid",
		r.RequireAuth(),
		promoHandler.GetValidPromoCodesForArtisan,
	)

	// ============================================================================
	// Analytics & Statistics
	// ============================================================================

	// Get promo code statistics
	promos.Get("/stats",
		r.RequireAuth(),
		promoHandler.GetPromoCodeStats,
	)

	// Get top performing promo codes
	promos.Get("/analytics/top-performing",
		r.RequireAuth(),
		promoHandler.GetTopPerformingPromoCodes,
	)
}
