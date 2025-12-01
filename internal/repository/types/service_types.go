package types

import (
	"Krafti_Vibe/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

// ServiceStats represents service statistics
type ServiceStats struct {
	TotalServices        int64                            `json:"total_services"`
	ActiveServices       int64                            `json:"active_services"`
	InactiveServices     int64                            `json:"inactive_services"`
	ByCategory           map[models.ServiceCategory]int64 `json:"by_category"`
	AverageDuration      float64                          `json:"average_duration_minutes"`
	AveragePrice         float64                          `json:"average_price"`
	TotalBookings        int64                            `json:"total_bookings"`
	TotalRevenue         float64                          `json:"total_revenue"`
	ServicesWithDeposit  int64                            `json:"services_with_deposit"`
	MostPopularCategory  models.ServiceCategory           `json:"most_popular_category"`
	HighestPricedService *ServiceSummary                  `json:"highest_priced_service,omitempty"`
	MostBookedService    *ServiceSummary                  `json:"most_booked_service,omitempty"`
}

// CategoryStats represents statistics per category
type CategoryStats struct {
	Category      models.ServiceCategory `json:"category"`
	ServiceCount  int64                  `json:"service_count"`
	TotalBookings int64                  `json:"total_bookings"`
	TotalRevenue  float64                `json:"total_revenue"`
	AveragePrice  float64                `json:"average_price"`
	ActiveCount   int64                  `json:"active_count"`
}

// ServicePerformance represents performance metrics for a service
type ServicePerformance struct {
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

// ServiceRevenue represents revenue per service
type ServiceRevenue struct {
	ServiceID    uuid.UUID              `json:"service_id"`
	ServiceName  string                 `json:"service_name"`
	Category     models.ServiceCategory `json:"category"`
	Bookings     int64                  `json:"bookings"`
	Revenue      float64                `json:"revenue"`
	AveragePrice float64                `json:"average_price"`
}

// ServiceTrend represents service booking trends
type ServiceTrend struct {
	Date        time.Time `json:"date"`
	ServiceID   uuid.UUID `json:"service_id"`
	ServiceName string    `json:"service_name"`
	Bookings    int64     `json:"bookings"`
	Revenue     float64   `json:"revenue"`
}

// ServiceSummary represents a service summary
type ServiceSummary struct {
	ID       uuid.UUID              `json:"id"`
	Name     string                 `json:"name"`
	Category models.ServiceCategory `json:"category"`
	Price    float64                `json:"price"`
	Bookings int64                  `json:"bookings"`
}

// ServiceFilters for advanced filtering
type ServiceFilters struct {
	Categories      []models.ServiceCategory `json:"categories"`
	MinPrice        *float64                 `json:"min_price"`
	MaxPrice        *float64                 `json:"max_price"`
	MinDuration     *int                     `json:"min_duration"`
	MaxDuration     *int                     `json:"max_duration"`
	IsActive        *bool                    `json:"is_active"`
	ArtisanID       *uuid.UUID               `json:"artisan_id"`
	Tags            []string                 `json:"tags"`
	RequiresDeposit *bool                    `json:"requires_deposit"`
}

// CategoryCount represents count per category
type CategoryCount struct {
	Category models.ServiceCategory `json:"category"`
	Count    int64                  `json:"count"`
}
