package dto

import (
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Artisan Request DTOs
// ============================================================================

// CreateArtisanRequest represents a request to create an artisan profile
type CreateArtisanRequest struct {
	UserID               uuid.UUID              `json:"user_id" validate:"required"`
	TenantID             uuid.UUID              `json:"tenant_id" validate:"required"`
	Bio                  string                 `json:"bio,omitempty"`
	Specialization       []string               `json:"specialization" validate:"required,min=1"`
	YearsExperience      int                    `json:"years_experience" validate:"min=0"`
	Certifications       []models.Certification `json:"certifications,omitempty"`
	Portfolio            []models.PortfolioItem `json:"portfolio,omitempty"`
	CommissionRate       float64                `json:"commission_rate" validate:"min=0,max=100"`
	PaymentAccountID     string                 `json:"payment_account_id,omitempty"`
	AutoAcceptBookings   bool                   `json:"auto_accept_bookings"`
	BookingLeadTime      int                    `json:"booking_lead_time" validate:"min=0"`
	MaxAdvanceBooking    int                    `json:"max_advance_booking" validate:"min=0"`
	SimultaneousBookings int                    `json:"simultaneous_bookings" validate:"min=1"`
	Location             *models.Location       `json:"location,omitempty"`
	ServiceRadius        int                    `json:"service_radius" validate:"min=0"`
	Metadata             map[string]any         `json:"metadata,omitempty"`
}

// Validate validates the create artisan request
func (r *CreateArtisanRequest) Validate() error {
	if r.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("tenant_id is required")
	}
	if len(r.Specialization) == 0 {
		return fmt.Errorf("at least one specialization is required")
	}
	if r.YearsExperience < 0 {
		return fmt.Errorf("years_experience cannot be negative")
	}
	if r.CommissionRate < 0 || r.CommissionRate > 100 {
		return fmt.Errorf("commission_rate must be between 0 and 100")
	}
	if r.BookingLeadTime < 0 {
		return fmt.Errorf("booking_lead_time cannot be negative")
	}
	if r.MaxAdvanceBooking < 0 {
		return fmt.Errorf("max_advance_booking cannot be negative")
	}
	if r.SimultaneousBookings < 1 {
		return fmt.Errorf("simultaneous_bookings must be at least 1")
	}
	if r.ServiceRadius < 0 {
		return fmt.Errorf("service_radius cannot be negative")
	}
	return nil
}

// UpdateArtisanRequest represents a request to update an artisan profile
type UpdateArtisanRequest struct {
	Bio                  *string                `json:"bio,omitempty"`
	Specialization       []string               `json:"specialization,omitempty"`
	YearsExperience      *int                   `json:"years_experience,omitempty" validate:"omitempty,min=0"`
	Certifications       []models.Certification `json:"certifications,omitempty"`
	Portfolio            []models.PortfolioItem `json:"portfolio,omitempty"`
	CommissionRate       *float64               `json:"commission_rate,omitempty" validate:"omitempty,min=0,max=100"`
	PaymentAccountID     *string                `json:"payment_account_id,omitempty"`
	AutoAcceptBookings   *bool                  `json:"auto_accept_bookings,omitempty"`
	BookingLeadTime      *int                   `json:"booking_lead_time,omitempty" validate:"omitempty,min=0"`
	MaxAdvanceBooking    *int                   `json:"max_advance_booking,omitempty" validate:"omitempty,min=0"`
	SimultaneousBookings *int                   `json:"simultaneous_bookings,omitempty" validate:"omitempty,min=1"`
	Location             *models.Location       `json:"location,omitempty"`
	ServiceRadius        *int                   `json:"service_radius,omitempty" validate:"omitempty,min=0"`
	Metadata             map[string]any         `json:"metadata,omitempty"`
}

// UpdateAvailabilityRequest represents a request to update artisan availability
type UpdateAvailabilityRequest struct {
	IsAvailable      bool   `json:"is_available"`
	AvailabilityNote string `json:"availability_note,omitempty" validate:"max=500"`
}

// ArtisanFilter represents filters for artisan queries
type ArtisanFilter struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	IsAvailable    *bool     `json:"is_available,omitempty"`
	Specialization *string   `json:"specialization,omitempty"`
	MinRating      *float64  `json:"min_rating,omitempty"`
	MinExperience  *int      `json:"min_experience,omitempty"`
	Latitude       *float64  `json:"latitude,omitempty"`
	Longitude      *float64  `json:"longitude,omitempty"`
	RadiusKm       *int      `json:"radius_km,omitempty"`
	Page           int       `json:"page"`
	PageSize       int       `json:"page_size"`
}

// SearchArtisanRequest represents search parameters
type SearchArtisanRequest struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	Query    string    `json:"query" validate:"required,min=2"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}

// AddCertificationRequest represents request to add a certification
type AddCertificationRequest struct {
	Name       string     `json:"name" validate:"required"`
	IssuedBy   string     `json:"issued_by" validate:"required"`
	IssuedDate time.Time  `json:"issued_date" validate:"required"`
	ExpiryDate *time.Time `json:"expiry_date,omitempty"`
	FileURL    string     `json:"file_url,omitempty"`
}

// AddPortfolioItemRequest represents request to add a portfolio item
type AddPortfolioItemRequest struct {
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description,omitempty"`
	ImageURLs   []string  `json:"image_urls,omitempty"`
	Date        time.Time `json:"date" validate:"required"`
	Tags        []string  `json:"tags,omitempty"`
}

// ============================================================================
// Artisan Response DTOs
// ============================================================================

// ArtisanResponse represents an artisan profile
type ArtisanResponse struct {
	ID                   uuid.UUID              `json:"id"`
	UserID               uuid.UUID              `json:"user_id"`
	TenantID             uuid.UUID              `json:"tenant_id"`
	Bio                  string                 `json:"bio,omitempty"`
	Specialization       []string               `json:"specialization"`
	YearsExperience      int                    `json:"years_experience"`
	Certifications       []models.Certification `json:"certifications,omitempty"`
	Portfolio            []models.PortfolioItem `json:"portfolio,omitempty"`
	Rating               float64                `json:"rating"`
	ReviewCount          int                    `json:"review_count"`
	TotalBookings        int                    `json:"total_bookings"`
	IsAvailable          bool                   `json:"is_available"`
	AvailabilityNote     string                 `json:"availability_note,omitempty"`
	CommissionRate       float64                `json:"commission_rate"`
	PaymentAccountID     string                 `json:"payment_account_id,omitempty"`
	AutoAcceptBookings   bool                   `json:"auto_accept_bookings"`
	BookingLeadTime      int                    `json:"booking_lead_time"`
	MaxAdvanceBooking    int                    `json:"max_advance_booking"`
	SimultaneousBookings int                    `json:"simultaneous_bookings"`
	Location             models.Location        `json:"location,omitempty"`
	ServiceRadius        int                    `json:"service_radius"`
	Metadata             models.JSONB           `json:"metadata,omitempty"`
	UserName             string                 `json:"user_name,omitempty"`
	UserEmail            string                 `json:"user_email,omitempty"`
	UserPhone            string                 `json:"user_phone,omitempty"`
	UserAvatar           string                 `json:"user_avatar,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// ArtisanListResponse represents a paginated list of artisans
type ArtisanListResponse struct {
	Artisans    []*ArtisanResponse `json:"artisans"`
	Page        int                `json:"page"`
	PageSize    int                `json:"page_size"`
	TotalItems  int64              `json:"total_items"`
	TotalPages  int                `json:"total_pages"`
	HasNext     bool               `json:"has_next"`
	HasPrevious bool               `json:"has_previous"`
}

// ArtisanStatsResponse represents artisan statistics
type ArtisanStatsResponse struct {
	ArtisanID         uuid.UUID       `json:"artisan_id"`
	TotalBookings     int64           `json:"total_bookings"`
	CompletedBookings int64           `json:"completed_bookings"`
	CancelledBookings int64           `json:"cancelled_bookings"`
	ActiveProjects    int64           `json:"active_projects"`
	TotalProjects     int64           `json:"total_projects"`
	AverageRating     float64         `json:"average_rating"`
	ReviewCount       int64           `json:"review_count"`
	TotalEarnings     float64         `json:"total_earnings"`
	CompletionRate    float64         `json:"completion_rate"`
	TotalServices     int64           `json:"total_services"`
	YearsExperience   int             `json:"years_experience"`
	IsAvailable       bool            `json:"is_available"`
	Recent30Days      BookingStats30d `json:"recent_30_days,omitempty"`
}

// BookingStats30d represents 30-day booking statistics
type BookingStats30d struct {
	TotalBookings     int     `json:"total_bookings"`
	CompletedBookings int     `json:"completed_bookings"`
	Revenue           float64 `json:"revenue"`
}

// ArtisanDashboardStatsResponse represents dashboard statistics
type ArtisanDashboardStatsResponse struct {
	ProjectsActive   int        `json:"projects_active"`
	TasksOpen        int        `json:"tasks_open"`
	TasksOverdue     int        `json:"tasks_overdue"`
	NextDueAt        *time.Time `json:"next_due_at,omitempty"`
	UpcomingBookings int        `json:"upcoming_bookings"`
	PendingReviews   int        `json:"pending_reviews"`
	UnreadMessages   int        `json:"unread_messages"`
	WeeklyEarnings   float64    `json:"weekly_earnings"`
}

// ArtisanDetailedResponse represents detailed artisan profile with relationships
type ArtisanDetailedResponse struct {
	*ArtisanResponse
	Services      []ServiceSummary `json:"services,omitempty"`
	RecentReviews []ReviewSummary  `json:"recent_reviews,omitempty"`
}

// ServiceSummary represents a service summary
type ServiceSummary struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Category string    `json:"category"`
	Price    float64   `json:"price"`
	Duration int       `json:"duration"`
	IsActive bool      `json:"is_active"`
}

// ReviewSummary represents a review summary
type ReviewSummary struct {
	ID           uuid.UUID `json:"id"`
	Rating       float64   `json:"rating"`
	Comment      string    `json:"comment"`
	CustomerName string    `json:"customer_name"`
	CreatedAt    time.Time `json:"created_at"`
}

// ReviewResponse represents a detailed review response
type ReviewResponse struct {
	ID           uuid.UUID `json:"id"`
	Rating       float64   `json:"rating"`
	Comment      string    `json:"comment"`
	CustomerName string    `json:"customer_name"`
	CreatedAt    time.Time `json:"created_at"`
}

// ArtisanServiceDashboardResponse represents dashboard data for an artisan service
type ArtisanServiceDashboardResponse struct {
	Stats            *ArtisanStatsResponse `json:"stats"`
	RecentBookings   []*BookingResponse    `json:"recent_bookings"`
	UpcomingBookings []*BookingResponse    `json:"upcoming_bookings"`
	RecentReviews    []*ReviewResponse     `json:"recent_reviews"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToArtisanResponse converts an Artisan model to ArtisanResponse DTO
func ToArtisanResponse(artisan *models.Artisan) *ArtisanResponse {
	if artisan == nil {
		return nil
	}

	resp := &ArtisanResponse{
		ID:                   artisan.ID,
		UserID:               artisan.UserID,
		TenantID:             artisan.TenantID,
		Bio:                  artisan.Bio,
		Specialization:       artisan.Specialization,
		YearsExperience:      artisan.YearsExperience,
		Certifications:       artisan.Certifications,
		Portfolio:            artisan.Portfolio,
		Rating:               artisan.Rating,
		ReviewCount:          artisan.ReviewCount,
		TotalBookings:        artisan.TotalBookings,
		IsAvailable:          artisan.IsAvailable,
		AvailabilityNote:     artisan.AvailabilityNote,
		CommissionRate:       artisan.CommissionRate,
		PaymentAccountID:     artisan.PaymentAccountID,
		AutoAcceptBookings:   artisan.AutoAcceptBookings,
		BookingLeadTime:      artisan.BookingLeadTime,
		MaxAdvanceBooking:    artisan.MaxAdvanceBooking,
		SimultaneousBookings: artisan.SimultaneousBookings,
		Location:             artisan.Location,
		ServiceRadius:        artisan.ServiceRadius,
		Metadata:             artisan.Metadata,
		CreatedAt:            artisan.CreatedAt,
		UpdatedAt:            artisan.UpdatedAt,
	}

	// Add user information if available
	if artisan.User != nil {
		resp.UserName = artisan.User.FirstName + " " + artisan.User.LastName
		resp.UserEmail = artisan.User.Email
		resp.UserPhone = artisan.User.PhoneNumber
		resp.UserAvatar = artisan.User.AvatarURL
	}

	return resp
}

// ToArtisanResponses converts multiple Artisan models to DTOs
func ToArtisanResponses(artisans []*models.Artisan) []*ArtisanResponse {
	if artisans == nil {
		return nil
	}

	responses := make([]*ArtisanResponse, len(artisans))
	for i, artisan := range artisans {
		responses[i] = ToArtisanResponse(artisan)
	}
	return responses
}

// ToArtisanDetailedResponse converts an Artisan model with relationships
func ToArtisanDetailedResponse(artisan *models.Artisan) *ArtisanDetailedResponse {
	if artisan == nil {
		return nil
	}

	baseResp := ToArtisanResponse(artisan)
	detailed := &ArtisanDetailedResponse{
		ArtisanResponse: baseResp,
	}

	// Add services if available
	if len(artisan.Services) > 0 {
		services := make([]ServiceSummary, len(artisan.Services))
		for i, svc := range artisan.Services {
			services[i] = ServiceSummary{
				ID:       svc.ID,
				Name:     svc.Name,
				Category: string(svc.Category),
				Price:    svc.Price,
				Duration: svc.DurationMinutes,
				IsActive: svc.IsActive,
			}
		}
		detailed.Services = services
	}

	// Add recent reviews if available
	if len(artisan.Reviews) > 0 {
		// Take up to 5 most recent reviews
		reviewCount := min(len(artisan.Reviews), 5)
		reviews := make([]ReviewSummary, reviewCount)
		for i := 0; i < reviewCount; i++ {
			review := artisan.Reviews[i]
			customerName := ""
			// Customer is directly a User, not a nested structure
			if review.Customer != nil {
				customerName = review.Customer.FirstName + " " + review.Customer.LastName
			}
			reviews[i] = ReviewSummary{
				ID:           review.ID,
				Rating:       float64(review.Rating),
				Comment:      review.Comment,
				CustomerName: customerName,
				CreatedAt:    review.CreatedAt,
			}
		}
		detailed.RecentReviews = reviews
	}

	return detailed
}
