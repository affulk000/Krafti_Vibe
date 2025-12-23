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

	// Apply authentication to all review routes
	reviews.Use(r.zitadelMW.RequireAuth())

	// ============================================================================
	// CRUD Operations
	// ============================================================================

	// Create review - customer (after booking completion)
	reviews.Post("", reviewHandler.CreateReview)

	// Get review by ID - any authenticated user
	reviews.Get("/:id", reviewHandler.GetReview)

	// Update review - review author or tenant owner/admin
	reviews.Put("/:id", reviewHandler.UpdateReview)

	// Delete review - review author or tenant owner/admin
	reviews.Delete("/:id", reviewHandler.DeleteReview)

	// ============================================================================
	// Query Operations
	// ============================================================================

	// List reviews - any authenticated user
	reviews.Post("/list", reviewHandler.ListReviews)

	// Get artisan reviews - any authenticated user
	reviews.Get("/artisan/:artisan_id", reviewHandler.GetArtisanReviews)

	// ============================================================================
	// Review Interaction
	// ============================================================================

	// Respond to review - artisan (reviewed) or tenant owner/admin
	reviews.Post("/:id/respond", middleware.RequireArtisanOrTeamMember(), reviewHandler.RespondToReview)

	// Mark review as helpful - any authenticated user
	reviews.Post("/:id/helpful", reviewHandler.MarkHelpful)

	// ============================================================================
	// Analytics & Statistics
	// ============================================================================

	// Get review statistics for artisan - artisan (self) or tenant owner/admin
	reviews.Get("/artisan/:artisan_id/stats", middleware.RequireArtisanOrTeamMember(), reviewHandler.GetReviewStats)
}
