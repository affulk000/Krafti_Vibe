package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Review Request DTOs
// ============================================================================

// CreateReviewRequest represents a request to create a review
type CreateReviewRequest struct {
	BookingID             uuid.UUID      `json:"booking_id" validate:"required"`
	Rating                int            `json:"rating" validate:"required,min=1,max=5"`
	Title                 string         `json:"title,omitempty" validate:"omitempty,max=255"`
	Comment               string         `json:"comment,omitempty"`
	QualityRating         *int           `json:"quality_rating,omitempty" validate:"omitempty,min=1,max=5"`
	ProfessionalismRating *int           `json:"professionalism_rating,omitempty" validate:"omitempty,min=1,max=5"`
	ValueRating           *int           `json:"value_rating,omitempty" validate:"omitempty,min=1,max=5"`
	TimelinessRating      *int           `json:"timeliness_rating,omitempty" validate:"omitempty,min=1,max=5"`
	PhotoURLs             []string       `json:"photo_urls,omitempty"`
	Metadata              map[string]any `json:"metadata,omitempty"`
}

// UpdateReviewRequest represents a request to update a review
type UpdateReviewRequest struct {
	Rating                *int           `json:"rating,omitempty" validate:"omitempty,min=1,max=5"`
	Title                 *string        `json:"title,omitempty" validate:"omitempty,max=255"`
	Comment               *string        `json:"comment,omitempty"`
	QualityRating         *int           `json:"quality_rating,omitempty" validate:"omitempty,min=1,max=5"`
	ProfessionalismRating *int           `json:"professionalism_rating,omitempty" validate:"omitempty,min=1,max=5"`
	ValueRating           *int           `json:"value_rating,omitempty" validate:"omitempty,min=1,max=5"`
	TimelinessRating      *int           `json:"timeliness_rating,omitempty" validate:"omitempty,min=1,max=5"`
	PhotoURLs             []string       `json:"photo_urls,omitempty"`
	Metadata              map[string]any `json:"metadata,omitempty"`
}

// RespondToReviewRequest represents a request to respond to a review
type RespondToReviewRequest struct {
	ResponseText string `json:"response_text" validate:"required"`
}

// ReviewFilter represents filters for review queries
type ReviewFilter struct {
	TenantID   uuid.UUID  `json:"tenant_id" validate:"required"`
	ArtisanID  *uuid.UUID `json:"artisan_id,omitempty"`
	CustomerID *uuid.UUID `json:"customer_id,omitempty"`
	ServiceID  *uuid.UUID `json:"service_id,omitempty"`
	MinRating  *int       `json:"min_rating,omitempty" validate:"omitempty,min=1,max=5"`
	MaxRating  *int       `json:"max_rating,omitempty" validate:"omitempty,min=1,max=5"`
	Published  *bool      `json:"published,omitempty"`
	Flagged    *bool      `json:"flagged,omitempty"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
}

// ============================================================================
// Review Response DTOs
// ============================================================================

// ReviewDetailResponse represents a detailed review with full information
type ReviewDetailResponse struct {
	ID                    uuid.UUID             `json:"id"`
	TenantID              uuid.UUID             `json:"tenant_id"`
	BookingID             uuid.UUID             `json:"booking_id"`
	ArtisanID             uuid.UUID             `json:"artisan_id"`
	CustomerID            uuid.UUID             `json:"customer_id"`
	ServiceID             uuid.UUID             `json:"service_id"`
	Rating                int                   `json:"rating"`
	Title                 string                `json:"title,omitempty"`
	Comment               string                `json:"comment,omitempty"`
	QualityRating         *int                  `json:"quality_rating,omitempty"`
	ProfessionalismRating *int                  `json:"professionalism_rating,omitempty"`
	ValueRating           *int                  `json:"value_rating,omitempty"`
	TimelinessRating      *int                  `json:"timeliness_rating,omitempty"`
	PhotoURLs             []string              `json:"photo_urls,omitempty"`
	ResponseText          string                `json:"response_text,omitempty"`
	ResponsedAt           *time.Time            `json:"responsed_at,omitempty"`
	IsPublished           bool                  `json:"is_published"`
	IsFlagged             bool                  `json:"is_flagged"`
	FlaggedReason         string                `json:"flagged_reason,omitempty"`
	HelpfulCount          int                   `json:"helpful_count"`
	NotHelpfulCount       int                   `json:"not_helpful_count"`
	Artisan               *UserSummary          `json:"artisan,omitempty"`
	Customer              *UserSummary          `json:"customer,omitempty"`
	Service               *ReviewServiceSummary `json:"service,omitempty"`
	IsPositive            bool                  `json:"is_positive"`
	HasResponse           bool                  `json:"has_response"`
	CreatedAt             time.Time             `json:"created_at"`
	UpdatedAt             time.Time             `json:"updated_at"`
}

// ReviewListResponse represents a paginated list of reviews
type ReviewListResponse struct {
	Reviews     []*ReviewDetailResponse `json:"reviews"`
	Page        int                     `json:"page"`
	PageSize    int                     `json:"page_size"`
	TotalItems  int64                   `json:"total_items"`
	TotalPages  int                     `json:"total_pages"`
	HasNext     bool                    `json:"has_next"`
	HasPrevious bool                    `json:"has_previous"`
}

// ReviewStatsResponse represents review statistics
type ReviewStatsResponse struct {
	ArtisanID          uuid.UUID     `json:"artisan_id,omitempty"`
	TotalReviews       int64         `json:"total_reviews"`
	AverageRating      float64       `json:"average_rating"`
	RatingDistribution map[int]int64 `json:"rating_distribution"`
	PositiveReviews    int64         `json:"positive_reviews"`
	NegativeReviews    int64         `json:"negative_reviews"`
	ReviewsThisMonth   int64         `json:"reviews_this_month"`
	ReviewsWithPhotos  int64         `json:"reviews_with_photos"`
	ResponseRate       float64       `json:"response_rate"`
}

// ReviewServiceSummary represents a minimal service summary for reviews
type ReviewServiceSummary struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToReviewDetailResponse converts a Review model to ReviewDetailResponse DTO
func ToReviewDetailResponse(review *models.Review) *ReviewDetailResponse {
	if review == nil {
		return nil
	}

	resp := &ReviewDetailResponse{
		ID:                    review.ID,
		TenantID:              review.TenantID,
		BookingID:             review.BookingID,
		ArtisanID:             review.ArtisanID,
		CustomerID:            review.CustomerID,
		ServiceID:             review.ServiceID,
		Rating:                review.Rating,
		Title:                 review.Title,
		Comment:               review.Comment,
		QualityRating:         review.QualityRating,
		ProfessionalismRating: review.ProfessionalismRating,
		ValueRating:           review.ValueRating,
		TimelinessRating:      review.TimelinessRating,
		PhotoURLs:             review.PhotoURLs,
		ResponseText:          review.ResponseText,
		ResponsedAt:           review.ResponsedAt,
		IsPublished:           review.IsPublished,
		IsFlagged:             review.IsFlagged,
		FlaggedReason:         review.FlaggedReason,
		HelpfulCount:          review.HelpfulCount,
		NotHelpfulCount:       review.NotHelpfulCount,
		IsPositive:            review.IsPositive(),
		HasResponse:           review.HasResponse(),
		CreatedAt:             review.CreatedAt,
		UpdatedAt:             review.UpdatedAt,
	}

	// Add artisan if available
	if review.Artisan != nil {
		resp.Artisan = &UserSummary{
			ID:        review.Artisan.ID,
			FirstName: review.Artisan.FirstName,
			LastName:  review.Artisan.LastName,
			Email:     review.Artisan.Email,
			AvatarURL: review.Artisan.AvatarURL,
		}
	}

	// Add customer if available
	if review.Customer != nil {
		resp.Customer = &UserSummary{
			ID:        review.Customer.ID,
			FirstName: review.Customer.FirstName,
			LastName:  review.Customer.LastName,
			Email:     review.Customer.Email,
			AvatarURL: review.Customer.AvatarURL,
		}
	}

	// Add service if available
	if review.Service != nil {
		resp.Service = &ReviewServiceSummary{
			ID:          review.Service.ID,
			Name:        review.Service.Name,
			Description: review.Service.Description,
		}
	}

	return resp
}

// ToReviewDetailResponses converts multiple Review models to DTOs
func ToReviewDetailResponses(reviews []*models.Review) []*ReviewDetailResponse {
	if reviews == nil {
		return nil
	}

	responses := make([]*ReviewDetailResponse, len(reviews))
	for i, review := range reviews {
		responses[i] = ToReviewDetailResponse(review)
	}
	return responses
}
