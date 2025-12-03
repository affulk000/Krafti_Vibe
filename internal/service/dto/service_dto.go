package dto

import (
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/repository/types"

	"github.com/google/uuid"
)

// ============================================================================
// Request DTOs
// ============================================================================

// CreateServiceRequest represents a request to create a service
type CreateServiceRequest struct {
	TenantID        uuid.UUID              `json:"tenant_id" validate:"required"`
	ArtisanID       *uuid.UUID             `json:"artisan_id,omitempty"`
	Name            string                 `json:"name" validate:"required,min=2,max=255"`
	Description     string                 `json:"description,omitempty"`
	Category        models.ServiceCategory `json:"category" validate:"required"`
	Price           float64                `json:"price" validate:"required,min=0"`
	Currency        string                 `json:"currency" validate:"required,len=3"`
	DepositAmount   float64                `json:"deposit_amount,omitempty"`
	DurationMinutes int                    `json:"duration_minutes" validate:"required,min=5"`
	BufferMinutes   int                    `json:"buffer_minutes" validate:"min=0"`
	IsActive        bool                   `json:"is_active"`
	MaxBookingsDay  int                    `json:"max_bookings_day" validate:"min=0"`
	ImageURL        string                 `json:"image_url,omitempty"`
	RequiresDeposit bool                   `json:"requires_deposit"`
	Tags            []string               `json:"tags,omitempty"`
	Metadata        models.JSONB           `json:"metadata,omitempty"`
}

// Validate validates the create service request
func (r *CreateServiceRequest) Validate() error {
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID is required")
	}
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.Price < 0 {
		return fmt.Errorf("price cannot be negative")
	}
	if r.DurationMinutes < 5 {
		return fmt.Errorf("duration must be at least 5 minutes")
	}
	if r.Currency == "" {
		r.Currency = "USD"
	}
	return nil
}

// UpdateServiceRequest represents a request to update a service
type UpdateServiceRequest struct {
	Name            *string                 `json:"name,omitempty"`
	Description     *string                 `json:"description,omitempty"`
	Category        *models.ServiceCategory `json:"category,omitempty"`
	Price           *float64                `json:"price,omitempty"`
	Currency        *string                 `json:"currency,omitempty"`
	DepositAmount   *float64                `json:"deposit_amount,omitempty"`
	DurationMinutes *int                    `json:"duration_minutes,omitempty"`
	BufferMinutes   *int                    `json:"buffer_minutes,omitempty"`
	IsActive        *bool                   `json:"is_active,omitempty"`
	MaxBookingsDay  *int                    `json:"max_bookings_day,omitempty"`
	ImageURL        *string                 `json:"image_url,omitempty"`
	RequiresDeposit *bool                   `json:"requires_deposit,omitempty"`
	Tags            []string                `json:"tags,omitempty"`
	Metadata        models.JSONB            `json:"metadata,omitempty"`
}

// ListServicesRequest represents a request to list services
type ListServicesRequest struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	Page     int       `json:"page" validate:"min=1"`
	PageSize int       `json:"page_size" validate:"min=1,max=100"`
}

// Validate validates the list services request
func (r *ListServicesRequest) Validate() error {
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID is required")
	}
	if r.Page < 1 {
		r.Page = 1
	}
	r.PageSize = max(1, min(r.PageSize, 100))
	if r.PageSize == 0 {
		r.PageSize = 20
	}
	return nil
}

// ServiceFilterRequest represents advanced filtering criteria
type ServiceFilterRequest struct {
	Categories      []models.ServiceCategory `json:"categories,omitempty"`
	MinPrice        *float64                 `json:"min_price,omitempty"`
	MaxPrice        *float64                 `json:"max_price,omitempty"`
	MinDuration     *int                     `json:"min_duration,omitempty"`
	MaxDuration     *int                     `json:"max_duration,omitempty"`
	IsActive        *bool                    `json:"is_active,omitempty"`
	ArtisanID       *uuid.UUID               `json:"artisan_id,omitempty"`
	Tags            []string                 `json:"tags,omitempty"`
	RequiresDeposit *bool                    `json:"requires_deposit,omitempty"`
	Page            int                      `json:"page" validate:"min=1"`
	PageSize        int                      `json:"page_size" validate:"min=1,max=100"`
}

// Validate validates the service filter request
func (r *ServiceFilterRequest) Validate() error {
	if r.Page < 1 {
		r.Page = 1
	}
	r.PageSize = max(1, min(r.PageSize, 100))
	if r.PageSize == 0 {
		r.PageSize = 20
	}
	return nil
}

// BulkPriceUpdateRequest represents a bulk price update request
type BulkPriceUpdateRequest struct {
	ServiceIDs   []uuid.UUID `json:"service_ids" validate:"required,min=1"`
	Adjustment   float64     `json:"adjustment" validate:"required"`
	IsPercentage bool        `json:"is_percentage"`
}

// Validate validates the bulk price update request
func (r *BulkPriceUpdateRequest) Validate() error {
	if len(r.ServiceIDs) == 0 {
		return fmt.Errorf("at least one service ID is required")
	}
	return nil
}

// ============================================================================
// Response DTOs
// ============================================================================

// ServiceResponse represents a service response
type ServiceResponse struct {
	ID              uuid.UUID              `json:"id"`
	TenantID        uuid.UUID              `json:"tenant_id"`
	ArtisanID       *uuid.UUID             `json:"artisan_id,omitempty"`
	ArtisanName     string                 `json:"artisan_name,omitempty"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	Category        models.ServiceCategory `json:"category"`
	Price           float64                `json:"price"`
	Currency        string                 `json:"currency"`
	DepositAmount   float64                `json:"deposit_amount"`
	DurationMinutes int                    `json:"duration_minutes"`
	BufferMinutes   int                    `json:"buffer_minutes"`
	TotalDuration   int                    `json:"total_duration"`
	IsActive        bool                   `json:"is_active"`
	MaxBookingsDay  int                    `json:"max_bookings_day"`
	ImageURL        string                 `json:"image_url,omitempty"`
	RequiresDeposit bool                   `json:"requires_deposit"`
	Tags            []string               `json:"tags,omitempty"`
	Metadata        models.JSONB           `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ListServicesResponse represents a paginated list of services
type ListServicesResponse struct {
	Services   []*ServiceResponse `json:"services"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalItems int64              `json:"total_items"`
	TotalPages int                `json:"total_pages"`
}

// ServiceAddonResponse represents a service addon response
type ServiceAddonResponse struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Price       float64   `json:"price"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ServiceStatistics represents service statistics
type ServiceStatistics struct {
	TotalServices        int64                            `json:"total_services"`
	ActiveServices       int64                            `json:"active_services"`
	InactiveServices     int64                            `json:"inactive_services"`
	ByCategory           map[models.ServiceCategory]int64 `json:"by_category"`
	AverageDuration      float64                          `json:"average_duration"`
	AveragePrice         float64                          `json:"average_price"`
	TotalBookings        int64                            `json:"total_bookings"`
	TotalRevenue         float64                          `json:"total_revenue"`
	ServicesWithDeposit  int64                            `json:"services_with_deposit"`
	MostPopularCategory  models.ServiceCategory           `json:"most_popular_category"`
	HighestPricedService *types.ServiceSummary            `json:"highest_priced_service,omitempty"`
	MostBookedService    *types.ServiceSummary            `json:"most_booked_service,omitempty"`
}

// CategoryStatisticsResponse represents category statistics
type CategoryStatisticsResponse struct {
	Category      models.ServiceCategory `json:"category"`
	ServiceCount  int64                  `json:"service_count"`
	TotalBookings int64                  `json:"total_bookings"`
	TotalRevenue  float64                `json:"total_revenue"`
	AveragePrice  float64                `json:"average_price"`
	ActiveCount   int64                  `json:"active_count"`
}

// ServicePerformanceResponse represents service performance metrics
type ServicePerformanceResponse struct {
	ServiceID         uuid.UUID `json:"service_id"`
	ServiceName       string    `json:"service_name"`
	TotalBookings     int64     `json:"total_bookings"`
	CompletedBookings int64     `json:"completed_bookings"`
	CancelledBookings int64     `json:"cancelled_bookings"`
	TotalRevenue      float64   `json:"total_revenue"`
	AverageRating     float64   `json:"average_rating"`
	ReviewCount       int64     `json:"review_count"`
	CompletionRate    float64   `json:"completion_rate"`
	CancellationRate  float64   `json:"cancellation_rate"`
	BookingsThisMonth int64     `json:"bookings_this_month"`
	RevenueThisMonth  float64   `json:"revenue_this_month"`
	PopularityScore   float64   `json:"popularity_score"`
}

// ServiceRevenueResponse represents service revenue
type ServiceRevenueResponse struct {
	ServiceID    uuid.UUID              `json:"service_id"`
	ServiceName  string                 `json:"service_name"`
	Category     models.ServiceCategory `json:"category"`
	Bookings     int64                  `json:"bookings"`
	Revenue      float64                `json:"revenue"`
	AveragePrice float64                `json:"average_price"`
}

// ServiceTrendResponse represents service booking trends
type ServiceTrendResponse struct {
	Date        time.Time `json:"date"`
	ServiceID   uuid.UUID `json:"service_id"`
	ServiceName string    `json:"service_name"`
	Bookings    int64     `json:"bookings"`
	Revenue     float64   `json:"revenue"`
}

// CategoryCountResponse represents category with count
type CategoryCountResponse struct {
	Category models.ServiceCategory `json:"category"`
	Count    int64                  `json:"count"`
}
