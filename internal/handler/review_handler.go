package handler

import (
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ReviewHandler handles HTTP requests for review operations
type ReviewHandler struct {
	reviewService service.ReviewService
}

// NewReviewHandler creates a new review handler
func NewReviewHandler(reviewService service.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

// CreateReview godoc
// @Summary Create a new review
// @Description Create a new review for a booking
// @Tags reviews
// @Accept json
// @Produce json
// @Param review body dto.CreateReviewRequest true "Review creation data"
// @Success 201 {object} dto.ReviewDetailResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reviews [post]
func (h *ReviewHandler) CreateReview(c *fiber.Ctx) error {
	var req dto.CreateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	review, err := h.reviewService.CreateReview(c.Context(), authCtx.TenantID, authCtx.UserID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, review, "Review created successfully")
}

// GetReview godoc
// @Summary Get review by ID
// @Description Get detailed review information by ID
// @Tags reviews
// @Produce json
// @Param id path string true "Review ID"
// @Success 200 {object} dto.ReviewDetailResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reviews/{id} [get]
func (h *ReviewHandler) GetReview(c *fiber.Ctx) error {
	reviewID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid review ID", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	review, err := h.reviewService.GetReview(c.Context(), reviewID, authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, review)
}

// UpdateReview godoc
// @Summary Update review
// @Description Update review information
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path string true "Review ID"
// @Param review body dto.UpdateReviewRequest true "Update data"
// @Success 200 {object} dto.ReviewDetailResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reviews/{id} [put]
func (h *ReviewHandler) UpdateReview(c *fiber.Ctx) error {
	reviewID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid review ID", err)
	}

	var req dto.UpdateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	review, err := h.reviewService.UpdateReview(c.Context(), reviewID, authCtx.UserID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, review, "Review updated successfully")
}

// DeleteReview godoc
// @Summary Delete review
// @Description Delete a review
// @Tags reviews
// @Param id path string true "Review ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reviews/{id} [delete]
func (h *ReviewHandler) DeleteReview(c *fiber.Ctx) error {
	reviewID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid review ID", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	if err := h.reviewService.DeleteReview(c.Context(), reviewID, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewNoContentResponse(c)
}

// ListReviews godoc
// @Summary List reviews
// @Description Get a paginated list of reviews with filters
// @Tags reviews
// @Accept json
// @Produce json
// @Param filter body dto.ReviewFilter true "Filter parameters"
// @Success 200 {object} dto.ReviewListResponse
// @Failure 500 {object} ErrorResponse
// @Router /reviews/list [post]
func (h *ReviewHandler) ListReviews(c *fiber.Ctx) error {
	var filter dto.ReviewFilter
	if err := c.BodyParser(&filter); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	reviews, err := h.reviewService.ListReviews(c.Context(), &filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, reviews)
}

// GetArtisanReviews godoc
// @Summary Get artisan reviews
// @Description Get all reviews for a specific artisan
// @Tags reviews
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reviews/artisan/{artisan_id} [get]
func (h *ReviewHandler) GetArtisanReviews(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("artisan_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid artisan ID", err)
	}

	reviews, err := h.reviewService.GetArtisanReviews(c.Context(), artisanID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, reviews)
}

// RespondToReview godoc
// @Summary Respond to review
// @Description Add an artisan response to a review
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path string true "Review ID"
// @Param response body dto.RespondToReviewRequest true "Response data"
// @Success 200 {object} dto.ReviewDetailResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reviews/{id}/respond [post]
func (h *ReviewHandler) RespondToReview(c *fiber.Ctx) error {
	reviewID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid review ID", err)
	}

	var req dto.RespondToReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	review, err := h.reviewService.RespondToReview(c.Context(), reviewID, authCtx.UserID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, review, "Response added successfully")
}

// GetReviewStats godoc
// @Summary Get review statistics
// @Description Get review statistics for an artisan
// @Tags reviews
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Success 200 {object} dto.ReviewStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reviews/artisan/{artisan_id}/stats [get]
func (h *ReviewHandler) GetReviewStats(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("artisan_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid artisan ID", err)
	}

	stats, err := h.reviewService.GetReviewStats(c.Context(), artisanID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, stats)
}

// MarkHelpful godoc
// @Summary Mark review as helpful
// @Description Mark a review as helpful
// @Tags reviews
// @Param id path string true "Review ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reviews/{id}/helpful [post]
func (h *ReviewHandler) MarkHelpful(c *fiber.Ctx) error {
	reviewID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid review ID", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	if err := h.reviewService.MarkHelpful(c.Context(), reviewID, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Marked as helpful")
}
