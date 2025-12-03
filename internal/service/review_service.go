package service

import (
	"context"
	"maps"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// ReviewService defines the interface for review service operations
type ReviewService interface {
	// CRUD Operations
	CreateReview(ctx context.Context, tenantID, customerID uuid.UUID, req *dto.CreateReviewRequest) (*dto.ReviewDetailResponse, error)
	GetReview(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*dto.ReviewDetailResponse, error)
	UpdateReview(ctx context.Context, id uuid.UUID, customerID uuid.UUID, req *dto.UpdateReviewRequest) (*dto.ReviewDetailResponse, error)
	DeleteReview(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Review Queries
	ListReviews(ctx context.Context, filter *dto.ReviewFilter) (*dto.ReviewListResponse, error)
	GetReviewByBooking(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID) (*dto.ReviewDetailResponse, error)
	GetArtisanReviews(ctx context.Context, artisanID uuid.UUID) ([]*dto.ReviewDetailResponse, error)
	GetCustomerReviews(ctx context.Context, customerID uuid.UUID) ([]*dto.ReviewDetailResponse, error)

	// Review Management
	RespondToReview(ctx context.Context, reviewID uuid.UUID, artisanID uuid.UUID, req *dto.RespondToReviewRequest) (*dto.ReviewDetailResponse, error)
	PublishReview(ctx context.Context, reviewID uuid.UUID, moderatorID uuid.UUID) error
	UnpublishReview(ctx context.Context, reviewID uuid.UUID, moderatorID uuid.UUID) error
	FlagReview(ctx context.Context, reviewID uuid.UUID, reason string, moderatorID uuid.UUID) error
	UnflagReview(ctx context.Context, reviewID uuid.UUID, moderatorID uuid.UUID) error

	// Moderation
	GetPendingModeration(ctx context.Context, tenantID uuid.UUID) ([]*dto.ReviewDetailResponse, error)
	GetFlaggedReviews(ctx context.Context, tenantID uuid.UUID) ([]*dto.ReviewDetailResponse, error)

	// Voting
	MarkHelpful(ctx context.Context, reviewID uuid.UUID, userID uuid.UUID) error
	MarkNotHelpful(ctx context.Context, reviewID uuid.UUID, userID uuid.UUID) error

	// Statistics & Analytics
	GetReviewStats(ctx context.Context, artisanID uuid.UUID) (*dto.ReviewStatsResponse, error)
	GetAverageRating(ctx context.Context, artisanID uuid.UUID) (float64, error)
	GetRatingDistribution(ctx context.Context, artisanID uuid.UUID) (map[int]int64, error)
}

// reviewService implements ReviewService
type reviewService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewReviewService creates a new ReviewService instance
func NewReviewService(repos *repository.Repositories, logger log.AllLogger) ReviewService {
	return &reviewService{
		repos:  repos,
		logger: logger,
	}
}

// CreateReview creates a new review
func (s *reviewService) CreateReview(ctx context.Context, tenantID, customerID uuid.UUID, req *dto.CreateReviewRequest) (*dto.ReviewDetailResponse, error) {
	s.logger.Info("creating review", "tenant_id", tenantID, "customer_id", customerID, "booking_id", req.BookingID)

	// Verify booking exists
	booking, err := s.repos.Booking.GetByID(ctx, req.BookingID)
	if err != nil {
		s.logger.Error("booking not found", "booking_id", req.BookingID, "error", err)
		return nil, errors.NewNotFoundError("booking")
	}

	// Verify booking belongs to tenant
	if booking.TenantID != tenantID {
		s.logger.Warn("booking does not belong to tenant", "booking_id", req.BookingID, "tenant_id", tenantID)
		return nil, errors.NewValidationError("Booking does not belong to this tenant")
	}

	// Verify customer is the booking customer
	if booking.CustomerID != customerID {
		s.logger.Warn("customer is not the booking customer", "customer_id", customerID, "booking_customer_id", booking.CustomerID)
		return nil, errors.NewValidationError("Only the booking customer can create a review")
	}

	// Check if review already exists for this booking
	existingReview, err := s.repos.Review.FindByBookingID(ctx, req.BookingID)
	if err == nil && existingReview != nil {
		return nil, errors.NewValidationError("Review already exists for this booking")
	}

	// Verify booking is completed
	if booking.Status != models.BookingStatusCompleted {
		return nil, errors.NewValidationError("Can only review completed bookings")
	}

	// Create review model
	review := &models.Review{
		TenantID:              tenantID,
		BookingID:             req.BookingID,
		ArtisanID:             booking.ArtisanID,
		CustomerID:            customerID,
		ServiceID:             booking.ServiceID,
		Rating:                req.Rating,
		Title:                 req.Title,
		Comment:               req.Comment,
		QualityRating:         req.QualityRating,
		ProfessionalismRating: req.ProfessionalismRating,
		ValueRating:           req.ValueRating,
		TimelinessRating:      req.TimelinessRating,
		PhotoURLs:             req.PhotoURLs,
		Metadata:              req.Metadata,
		IsPublished:           true, // Auto-publish by default
		IsFlagged:             false,
		HelpfulCount:          0,
		NotHelpfulCount:       0,
	}

	// Save to database
	if err := s.repos.Review.Create(ctx, review); err != nil {
		s.logger.Error("failed to create review", "error", err)
		return nil, errors.NewServiceError("CREATE_FAILED", "Failed to create review", err)
	}

	s.logger.Info("review created", "review_id", review.ID)

	// Load with relationships
	created, err := s.repos.Review.GetByID(ctx, review.ID)
	if err != nil {
		s.logger.Error("failed to load review with relationships", "review_id", review.ID, "error", err)
		return dto.ToReviewDetailResponse(review), nil
	}

	return dto.ToReviewDetailResponse(created), nil
}

// GetReview retrieves a review by ID
func (s *reviewService) GetReview(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*dto.ReviewDetailResponse, error) {
	s.logger.Info("retrieving review", "review_id", id, "user_id", userID)

	review, err := s.repos.Review.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("review not found", "review_id", id, "error", err)
		return nil, errors.NewNotFoundError("review")
	}

	// Verify user has access (customer, artisan, or admin)
	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user")
	}

	// Check access: must be customer, artisan, or from same tenant
	hasAccess := review.CustomerID == userID ||
		review.ArtisanID == userID ||
		(user.TenantID != nil && *user.TenantID == review.TenantID)

	if !hasAccess {
		s.logger.Warn("unauthorized access attempt", "review_id", id, "user_id", userID)
		return nil, errors.NewValidationError("You do not have access to this review")
	}

	return dto.ToReviewDetailResponse(review), nil
}

// UpdateReview updates a review
func (s *reviewService) UpdateReview(ctx context.Context, id uuid.UUID, customerID uuid.UUID, req *dto.UpdateReviewRequest) (*dto.ReviewDetailResponse, error) {
	s.logger.Info("updating review", "review_id", id, "customer_id", customerID)

	// Get existing review
	review, err := s.repos.Review.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("review not found", "review_id", id, "error", err)
		return nil, errors.NewNotFoundError("review")
	}

	// Verify customer is the reviewer
	if review.CustomerID != customerID {
		s.logger.Warn("unauthorized update attempt", "review_id", id, "customer_id", customerID)
		return nil, errors.NewValidationError("Only the reviewer can update this review")
	}

	// Cannot update if artisan has responded
	if review.ResponseText != "" {
		return nil, errors.NewValidationError("Cannot update review after artisan has responded")
	}

	// Update fields
	if req.Rating != nil {
		review.Rating = *req.Rating
	}

	if req.Title != nil {
		review.Title = *req.Title
	}

	if req.Comment != nil {
		review.Comment = *req.Comment
	}

	if req.QualityRating != nil {
		review.QualityRating = req.QualityRating
	}

	if req.ProfessionalismRating != nil {
		review.ProfessionalismRating = req.ProfessionalismRating
	}

	if req.ValueRating != nil {
		review.ValueRating = req.ValueRating
	}

	if req.TimelinessRating != nil {
		review.TimelinessRating = req.TimelinessRating
	}

	if req.PhotoURLs != nil {
		review.PhotoURLs = req.PhotoURLs
	}

	if req.Metadata != nil {
		if review.Metadata == nil {
			review.Metadata = make(models.JSONB)
		}
		maps.Copy(review.Metadata, req.Metadata)
	}

	// Save changes
	if err := s.repos.Review.Update(ctx, review); err != nil {
		s.logger.Error("failed to update review", "review_id", id, "error", err)
		return nil, errors.NewServiceError("UPDATE_FAILED", "Failed to update review", err)
	}

	s.logger.Info("review updated", "review_id", id)
	return dto.ToReviewDetailResponse(review), nil
}

// DeleteReview deletes a review
func (s *reviewService) DeleteReview(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	s.logger.Info("deleting review", "review_id", id, "user_id", userID)

	// Get existing review
	review, err := s.repos.Review.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("review not found", "review_id", id, "error", err)
		return errors.NewNotFoundError("review")
	}

	// Verify user is the customer or admin
	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return errors.NewNotFoundError("user")
	}

	canDelete := review.CustomerID == userID ||
		(user.TenantID != nil && *user.TenantID == review.TenantID &&
			(user.Role == models.UserRoleTenantAdmin || user.Role == models.UserRoleTenantOwner))

	if !canDelete {
		s.logger.Warn("unauthorized delete attempt", "review_id", id, "user_id", userID)
		return errors.NewValidationError("Only the reviewer or admin can delete this review")
	}

	if err := s.repos.Review.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete review", "review_id", id, "error", err)
		return errors.NewServiceError("DELETE_FAILED", "Failed to delete review", err)
	}

	s.logger.Info("review deleted", "review_id", id)
	return nil
}

// ListReviews retrieves reviews with filters
func (s *reviewService) ListReviews(ctx context.Context, filter *dto.ReviewFilter) (*dto.ReviewListResponse, error) {
	s.logger.Info("listing reviews", "tenant_id", filter.TenantID)

	// Set defaults
	page := max(1, filter.Page)
	pageSize := min(100, max(1, filter.PageSize))

	// Build query based on filters
	var reviews []models.Review
	var total int64
	var err error

	// For now, use simple filtering - you can extend this with more sophisticated repository methods
	if filter.ArtisanID != nil {
		reviews, err = s.repos.Review.FindByArtisanID(ctx, *filter.ArtisanID)
		total = int64(len(reviews))
	} else if filter.CustomerID != nil {
		reviews, err = s.repos.Review.FindByCustomerID(ctx, *filter.CustomerID)
		total = int64(len(reviews))
	} else {
		reviews, total, err = s.repos.Review.FindByTenantID(ctx, filter.TenantID, page, pageSize)
	}

	if err != nil {
		s.logger.Error("failed to list reviews", "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "Failed to list reviews", err)
	}

	// Apply additional filters in memory (can be optimized with repository-level filtering)
	var filteredReviews []models.Review
	for _, review := range reviews {
		// Apply min/max rating filters
		if filter.MinRating != nil && review.Rating < *filter.MinRating {
			continue
		}
		if filter.MaxRating != nil && review.Rating > *filter.MaxRating {
			continue
		}

		// Apply published filter
		if filter.Published != nil && review.IsPublished != *filter.Published {
			continue
		}

		// Apply flagged filter
		if filter.Flagged != nil && review.IsFlagged != *filter.Flagged {
			continue
		}

		filteredReviews = append(filteredReviews, review)
	}

	// Convert to pointers for DTO conversion
	reviewPtrs := make([]*models.Review, len(filteredReviews))
	for i := range filteredReviews {
		reviewPtrs[i] = &filteredReviews[i]
	}

	// Calculate pagination
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.ReviewListResponse{
		Reviews:     dto.ToReviewDetailResponses(reviewPtrs),
		Page:        page,
		PageSize:    pageSize,
		TotalItems:  total,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}, nil
}

// GetReviewByBooking retrieves review for a specific booking
func (s *reviewService) GetReviewByBooking(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID) (*dto.ReviewDetailResponse, error) {
	s.logger.Info("getting review by booking", "booking_id", bookingID, "user_id", userID)

	review, err := s.repos.Review.FindByBookingID(ctx, bookingID)
	if err != nil {
		s.logger.Error("review not found for booking", "booking_id", bookingID, "error", err)
		return nil, errors.NewNotFoundError("review")
	}

	// Verify user has access
	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user")
	}

	hasAccess := review.CustomerID == userID ||
		review.ArtisanID == userID ||
		(user.TenantID != nil && *user.TenantID == review.TenantID)

	if !hasAccess {
		s.logger.Warn("unauthorized access attempt", "booking_id", bookingID, "user_id", userID)
		return nil, errors.NewValidationError("You do not have access to this review")
	}

	return dto.ToReviewDetailResponse(review), nil
}

// GetArtisanReviews retrieves all published reviews for an artisan
func (s *reviewService) GetArtisanReviews(ctx context.Context, artisanID uuid.UUID) ([]*dto.ReviewDetailResponse, error) {
	s.logger.Info("getting artisan reviews", "artisan_id", artisanID)

	reviews, err := s.repos.Review.FindByArtisanID(ctx, artisanID)
	if err != nil {
		s.logger.Error("failed to get artisan reviews", "artisan_id", artisanID, "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get artisan reviews", err)
	}

	// Convert to pointers
	reviewPtrs := make([]*models.Review, len(reviews))
	for i := range reviews {
		reviewPtrs[i] = &reviews[i]
	}

	return dto.ToReviewDetailResponses(reviewPtrs), nil
}

// GetCustomerReviews retrieves all reviews by a customer
func (s *reviewService) GetCustomerReviews(ctx context.Context, customerID uuid.UUID) ([]*dto.ReviewDetailResponse, error) {
	s.logger.Info("getting customer reviews", "customer_id", customerID)

	reviews, err := s.repos.Review.FindByCustomerID(ctx, customerID)
	if err != nil {
		s.logger.Error("failed to get customer reviews", "customer_id", customerID, "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get customer reviews", err)
	}

	// Convert to pointers
	reviewPtrs := make([]*models.Review, len(reviews))
	for i := range reviews {
		reviewPtrs[i] = &reviews[i]
	}

	return dto.ToReviewDetailResponses(reviewPtrs), nil
}

// RespondToReview adds an artisan's response to a review
func (s *reviewService) RespondToReview(ctx context.Context, reviewID uuid.UUID, artisanID uuid.UUID, req *dto.RespondToReviewRequest) (*dto.ReviewDetailResponse, error) {
	s.logger.Info("responding to review", "review_id", reviewID, "artisan_id", artisanID)

	// Get existing review
	review, err := s.repos.Review.GetByID(ctx, reviewID)
	if err != nil {
		s.logger.Error("review not found", "review_id", reviewID, "error", err)
		return nil, errors.NewNotFoundError("review")
	}

	// Verify artisan is the review subject
	if review.ArtisanID != artisanID {
		s.logger.Warn("unauthorized response attempt", "review_id", reviewID, "artisan_id", artisanID)
		return nil, errors.NewValidationError("Only the reviewed artisan can respond")
	}

	// Check if already responded
	if review.ResponseText != "" {
		return nil, errors.NewValidationError("Review already has a response")
	}

	// Add response
	if err := s.repos.Review.AddResponse(ctx, reviewID, req.ResponseText, artisanID); err != nil {
		s.logger.Error("failed to add response", "review_id", reviewID, "error", err)
		return nil, errors.NewServiceError("RESPONSE_FAILED", "Failed to add response", err)
	}

	s.logger.Info("response added to review", "review_id", reviewID)

	// Reload review
	updated, err := s.repos.Review.GetByID(ctx, reviewID)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "Failed to reload review", err)
	}

	return dto.ToReviewDetailResponse(updated), nil
}

// PublishReview publishes a review (makes it public)
func (s *reviewService) PublishReview(ctx context.Context, reviewID uuid.UUID, moderatorID uuid.UUID) error {
	s.logger.Info("publishing review", "review_id", reviewID, "moderator_id", moderatorID)

	// Verify review exists
	_, err := s.repos.Review.GetByID(ctx, reviewID)
	if err != nil {
		s.logger.Error("review not found", "review_id", reviewID, "error", err)
		return errors.NewNotFoundError("review")
	}

	if err := s.repos.Review.UpdatePublishStatus(ctx, reviewID, true); err != nil {
		s.logger.Error("failed to publish review", "review_id", reviewID, "error", err)
		return errors.NewServiceError("PUBLISH_FAILED", "Failed to publish review", err)
	}

	s.logger.Info("review published", "review_id", reviewID)
	return nil
}

// UnpublishReview unpublishes a review (makes it private)
func (s *reviewService) UnpublishReview(ctx context.Context, reviewID uuid.UUID, moderatorID uuid.UUID) error {
	s.logger.Info("unpublishing review", "review_id", reviewID, "moderator_id", moderatorID)

	// Verify review exists
	_, err := s.repos.Review.GetByID(ctx, reviewID)
	if err != nil {
		s.logger.Error("review not found", "review_id", reviewID, "error", err)
		return errors.NewNotFoundError("review")
	}

	if err := s.repos.Review.UpdatePublishStatus(ctx, reviewID, false); err != nil {
		s.logger.Error("failed to unpublish review", "review_id", reviewID, "error", err)
		return errors.NewServiceError("UNPUBLISH_FAILED", "Failed to unpublish review", err)
	}

	s.logger.Info("review unpublished", "review_id", reviewID)
	return nil
}

// FlagReview flags a review for moderation
func (s *reviewService) FlagReview(ctx context.Context, reviewID uuid.UUID, reason string, moderatorID uuid.UUID) error {
	s.logger.Info("flagging review", "review_id", reviewID, "moderator_id", moderatorID, "reason", reason)

	// Get existing review
	review, err := s.repos.Review.GetByID(ctx, reviewID)
	if err != nil {
		s.logger.Error("review not found", "review_id", reviewID, "error", err)
		return errors.NewNotFoundError("review")
	}

	if review.IsFlagged {
		return errors.NewValidationError("Review is already flagged")
	}

	review.IsFlagged = true
	review.FlaggedReason = reason

	if err := s.repos.Review.Update(ctx, review); err != nil {
		s.logger.Error("failed to flag review", "review_id", reviewID, "error", err)
		return errors.NewServiceError("FLAG_FAILED", "Failed to flag review", err)
	}

	s.logger.Info("review flagged", "review_id", reviewID)
	return nil
}

// UnflagReview removes flag from a review
func (s *reviewService) UnflagReview(ctx context.Context, reviewID uuid.UUID, moderatorID uuid.UUID) error {
	s.logger.Info("unflagging review", "review_id", reviewID, "moderator_id", moderatorID)

	// Get existing review
	review, err := s.repos.Review.GetByID(ctx, reviewID)
	if err != nil {
		s.logger.Error("review not found", "review_id", reviewID, "error", err)
		return errors.NewNotFoundError("review")
	}

	if !review.IsFlagged {
		return errors.NewValidationError("Review is not flagged")
	}

	review.IsFlagged = false
	review.FlaggedReason = ""

	if err := s.repos.Review.Update(ctx, review); err != nil {
		s.logger.Error("failed to unflag review", "review_id", reviewID, "error", err)
		return errors.NewServiceError("UNFLAG_FAILED", "Failed to unflag review", err)
	}

	s.logger.Info("review unflagged", "review_id", reviewID)
	return nil
}

// GetPendingModeration retrieves reviews pending moderation
func (s *reviewService) GetPendingModeration(ctx context.Context, tenantID uuid.UUID) ([]*dto.ReviewDetailResponse, error) {
	s.logger.Info("getting pending moderation reviews", "tenant_id", tenantID)

	reviews, err := s.repos.Review.FindPendingModeration(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get pending moderation reviews", "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get pending moderation reviews", err)
	}

	// Convert to pointers
	reviewPtrs := make([]*models.Review, len(reviews))
	for i := range reviews {
		reviewPtrs[i] = &reviews[i]
	}

	return dto.ToReviewDetailResponses(reviewPtrs), nil
}

// GetFlaggedReviews retrieves flagged reviews
func (s *reviewService) GetFlaggedReviews(ctx context.Context, tenantID uuid.UUID) ([]*dto.ReviewDetailResponse, error) {
	s.logger.Info("getting flagged reviews", "tenant_id", tenantID)

	reviews, err := s.repos.Review.FindFlagged(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get flagged reviews", "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get flagged reviews", err)
	}

	// Convert to pointers
	reviewPtrs := make([]*models.Review, len(reviews))
	for i := range reviews {
		reviewPtrs[i] = &reviews[i]
	}

	return dto.ToReviewDetailResponses(reviewPtrs), nil
}

// MarkHelpful marks a review as helpful
func (s *reviewService) MarkHelpful(ctx context.Context, reviewID uuid.UUID, userID uuid.UUID) error {
	s.logger.Info("marking review as helpful", "review_id", reviewID, "user_id", userID)

	// Verify review exists
	_, err := s.repos.Review.GetByID(ctx, reviewID)
	if err != nil {
		s.logger.Error("review not found", "review_id", reviewID, "error", err)
		return errors.NewNotFoundError("review")
	}

	if err := s.repos.Review.IncrementHelpfulCount(ctx, reviewID); err != nil {
		s.logger.Error("failed to increment helpful count", "review_id", reviewID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to mark review as helpful", err)
	}

	s.logger.Info("review marked as helpful", "review_id", reviewID)
	return nil
}

// MarkNotHelpful marks a review as not helpful
func (s *reviewService) MarkNotHelpful(ctx context.Context, reviewID uuid.UUID, userID uuid.UUID) error {
	s.logger.Info("marking review as not helpful", "review_id", reviewID, "user_id", userID)

	// Verify review exists
	_, err := s.repos.Review.GetByID(ctx, reviewID)
	if err != nil {
		s.logger.Error("review not found", "review_id", reviewID, "error", err)
		return errors.NewNotFoundError("review")
	}

	if err := s.repos.Review.IncrementNotHelpfulCount(ctx, reviewID); err != nil {
		s.logger.Error("failed to increment not helpful count", "review_id", reviewID, "error", err)
		return errors.NewServiceError("UPDATE_FAILED", "Failed to mark review as not helpful", err)
	}

	s.logger.Info("review marked as not helpful", "review_id", reviewID)
	return nil
}

// GetReviewStats retrieves review statistics for an artisan
func (s *reviewService) GetReviewStats(ctx context.Context, artisanID uuid.UUID) (*dto.ReviewStatsResponse, error) {
	s.logger.Info("getting review stats", "artisan_id", artisanID)

	// Get all reviews for artisan
	reviews, err := s.repos.Review.FindByArtisanID(ctx, artisanID)
	if err != nil {
		s.logger.Error("failed to get artisan reviews", "artisan_id", artisanID, "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get artisan reviews", err)
	}

	// Get average rating
	avgRating, err := s.repos.Review.GetAverageRating(ctx, artisanID)
	if err != nil {
		s.logger.Error("failed to get average rating", "artisan_id", artisanID, "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get average rating", err)
	}

	// Get rating distribution
	distribution, err := s.repos.Review.GetRatingDistribution(ctx, artisanID)
	if err != nil {
		s.logger.Error("failed to get rating distribution", "artisan_id", artisanID, "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get rating distribution", err)
	}

	// Calculate stats
	totalReviews := int64(len(reviews))
	var positiveReviews, negativeReviews, reviewsWithPhotos int64
	var responsedCount int64

	for _, review := range reviews {
		if review.IsPositive() {
			positiveReviews++
		} else {
			negativeReviews++
		}

		if len(review.PhotoURLs) > 0 {
			reviewsWithPhotos++
		}

		if review.HasResponse() {
			responsedCount++
		}
	}

	// Calculate response rate
	responseRate := 0.0
	if totalReviews > 0 {
		responseRate = (float64(responsedCount) / float64(totalReviews)) * 100
	}

	// Reviews this month (simplified - would need time filtering in repository)
	reviewsThisMonth := int64(0)

	return &dto.ReviewStatsResponse{
		ArtisanID:          artisanID,
		TotalReviews:       totalReviews,
		AverageRating:      avgRating,
		RatingDistribution: distribution,
		PositiveReviews:    positiveReviews,
		NegativeReviews:    negativeReviews,
		ReviewsThisMonth:   reviewsThisMonth,
		ReviewsWithPhotos:  reviewsWithPhotos,
		ResponseRate:       responseRate,
	}, nil
}

// GetAverageRating retrieves the average rating for an artisan
func (s *reviewService) GetAverageRating(ctx context.Context, artisanID uuid.UUID) (float64, error) {
	s.logger.Info("getting average rating", "artisan_id", artisanID)

	avgRating, err := s.repos.Review.GetAverageRating(ctx, artisanID)
	if err != nil {
		s.logger.Error("failed to get average rating", "artisan_id", artisanID, "error", err)
		return 0, errors.NewServiceError("GET_FAILED", "Failed to get average rating", err)
	}

	return avgRating, nil
}

// GetRatingDistribution retrieves the rating distribution for an artisan
func (s *reviewService) GetRatingDistribution(ctx context.Context, artisanID uuid.UUID) (map[int]int64, error) {
	s.logger.Info("getting rating distribution", "artisan_id", artisanID)

	distribution, err := s.repos.Review.GetRatingDistribution(ctx, artisanID)
	if err != nil {
		s.logger.Error("failed to get rating distribution", "artisan_id", artisanID, "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get rating distribution", err)
	}

	return distribution, nil
}
