package router

import (
	"Krafti_Vibe/internal/handler"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
)

func (r *Router) setupReviewRoutes(api fiber.Router) {
	// Initialize service
	reviewService := service.NewReviewService(r.repos, r.config.Logger)
	reviewHandler := handler.NewReviewHandler(reviewService)

	// Create review routes
	reviews := api.Group("/reviews")

	// Auth middleware configuration
	authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
		RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
	})

	// ============================================================================
	// CRUD Operations
	// ============================================================================

	// Create review
	reviews.Post("",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReviewWrite),
		reviewHandler.CreateReview,
	)

	// Get review by ID
	reviews.Get("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReviewRead),
		reviewHandler.GetReview,
	)

	// Update review
	reviews.Put("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReviewWrite),
		reviewHandler.UpdateReview,
	)

	// Delete review
	reviews.Delete("/:id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReviewWrite),
		reviewHandler.DeleteReview,
	)

	// ============================================================================
	// Query Operations
	// ============================================================================

	// List reviews (with pagination and filters)
	reviews.Post("/list",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReviewRead),
		reviewHandler.ListReviews,
	)

	// Get artisan reviews
	reviews.Get("/artisan/:artisan_id",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReviewRead),
		reviewHandler.GetArtisanReviews,
	)

	// ============================================================================
	// Review Interaction
	// ============================================================================

	// Respond to review (artisan response)
	reviews.Post("/:id/respond",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReviewWrite),
		reviewHandler.RespondToReview,
	)

	// Mark review as helpful
	reviews.Post("/:id/helpful",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReviewRead),
		reviewHandler.MarkHelpful,
	)

	// ============================================================================
	// Analytics & Statistics
	// ============================================================================

	// Get review statistics for artisan
	reviews.Get("/artisan/:artisan_id/stats",
		authMiddleware,
		middleware.RequireScopes(r.scopes.ReviewRead),
		reviewHandler.GetReviewStats,
	)
}
