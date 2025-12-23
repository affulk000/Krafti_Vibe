package dto

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Request DTOs
// ============================================================================

// CreateCustomerRequest represents the request to create a customer
type CreateCustomerRequest struct {
	UserID             uuid.UUID        `json:"user_id" validate:"required"`
	TenantID           uuid.UUID        `json:"tenant_id" validate:"required"`
	Notes              string           `json:"notes,omitempty"`
	PreferredArtisans  []uuid.UUID      `json:"preferred_artisans,omitempty"`
	PrimaryLocation    *models.Location `json:"primary_location,omitempty"`
	EmailNotifications bool             `json:"email_notifications"`
	SMSNotifications   bool             `json:"sms_notifications"`
	PushNotifications  bool             `json:"push_notifications"`
	Metadata           map[string]any   `json:"metadata,omitempty"`
}

// Validate validates the create customer request
func (r *CreateCustomerRequest) Validate() error {
	if r.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID is required")
	}
	if len(r.Notes) > 2000 {
		return fmt.Errorf("notes must be 2000 characters or less")
	}
	return nil
}

// UpdateCustomerRequest represents the request to update a customer
type UpdateCustomerRequest struct {
	Notes                  *string          `json:"notes,omitempty"`
	PreferredArtisans      []uuid.UUID      `json:"preferred_artisans,omitempty"`
	PrimaryLocation        *models.Location `json:"primary_location,omitempty"`
	DefaultPaymentMethodID *string          `json:"default_payment_method_id,omitempty"`
	EmailNotifications     *bool            `json:"email_notifications,omitempty"`
	SMSNotifications       *bool            `json:"sms_notifications,omitempty"`
	PushNotifications      *bool            `json:"push_notifications,omitempty"`
	Metadata               map[string]any   `json:"metadata,omitempty"`
}

// Validate validates the update customer request
func (r *UpdateCustomerRequest) Validate() error {
	if r.Notes != nil && len(*r.Notes) > 2000 {
		return fmt.Errorf("notes must be 2000 characters or less")
	}
	return nil
}

// UpdateLoyaltyPointsRequest represents the request to update loyalty points
type UpdateLoyaltyPointsRequest struct {
	Points    int    `json:"points" validate:"required"`
	Operation string `json:"operation" validate:"required"` // add, subtract, set
	Reason    string `json:"reason,omitempty"`
}

// Validate validates the update loyalty points request
func (r *UpdateLoyaltyPointsRequest) Validate() error {
	validOps := []string{"add", "subtract", "set"}
	if !slices.Contains(validOps, r.Operation) {
		return fmt.Errorf("invalid operation: %s", r.Operation)
	}
	if r.Points < 0 {
		return fmt.Errorf("points cannot be negative")
	}
	return nil
}

// AddPreferredArtisanRequest represents the request to add preferred artisan
type AddPreferredArtisanRequest struct {
	ArtisanID uuid.UUID `json:"artisan_id" validate:"required"`
}

// Validate validates the add preferred artisan request
func (r *AddPreferredArtisanRequest) Validate() error {
	if r.ArtisanID == uuid.Nil {
		return fmt.Errorf("artisan ID is required")
	}
	return nil
}

// UpdateNotificationPreferencesRequest represents notification preferences update
type UpdateNotificationPreferencesRequest struct {
	EmailNotifications *bool `json:"email_notifications,omitempty"`
	SMSNotifications   *bool `json:"sms_notifications,omitempty"`
	PushNotifications  *bool `json:"push_notifications,omitempty"`
}

// ============================================================================
// Filter DTOs
// ============================================================================

// CustomerFilter represents filters for listing customers
type CustomerFilter struct {
	TenantID           *uuid.UUID  `json:"tenant_id,omitempty"`
	LoyaltyTiers       []string    `json:"loyalty_tiers,omitempty"`
	MinTotalSpent      *float64    `json:"min_total_spent,omitempty"`
	MaxTotalSpent      *float64    `json:"max_total_spent,omitempty"`
	MinLoyaltyPoints   *int        `json:"min_loyalty_points,omitempty"`
	MaxLoyaltyPoints   *int        `json:"max_loyalty_points,omitempty"`
	MinBookings        *int        `json:"min_bookings,omitempty"`
	MaxBookings        *int        `json:"max_bookings,omitempty"`
	HasPrimaryLocation *bool       `json:"has_primary_location,omitempty"`
	CreatedAfter       *time.Time  `json:"created_after,omitempty"`
	CreatedBefore      *time.Time  `json:"created_before,omitempty"`
	LastBookingAfter   *time.Time  `json:"last_booking_after,omitempty"`
	LastBookingBefore  *time.Time  `json:"last_booking_before,omitempty"`
	PreferredArtisans  []uuid.UUID `json:"preferred_artisans,omitempty"`
	Page               int         `json:"page" validate:"min=1"`
	PageSize           int         `json:"page_size" validate:"min=1,max=100"`
	SortBy             string      `json:"sort_by,omitempty"`
	SortOrder          string      `json:"sort_order,omitempty"` // asc or desc
	SearchQuery        string      `json:"search_query,omitempty"`
}

// Validate validates the customer filter
func (f *CustomerFilter) Validate() error {
	if f.Page < 1 {
		f.Page = 1
	}
	f.PageSize = max(1, min(f.PageSize, 100))
	if f.PageSize == 0 {
		f.PageSize = 20
	}
	if f.SortOrder != "" && f.SortOrder != "asc" && f.SortOrder != "desc" {
		return fmt.Errorf("sort order must be 'asc' or 'desc'")
	}
	if f.MinTotalSpent != nil && *f.MinTotalSpent < 0 {
		return fmt.Errorf("min total spent cannot be negative")
	}
	if f.MaxTotalSpent != nil && *f.MaxTotalSpent < 0 {
		return fmt.Errorf("max total spent cannot be negative")
	}
	if f.MinTotalSpent != nil && f.MaxTotalSpent != nil && *f.MinTotalSpent > *f.MaxTotalSpent {
		return fmt.Errorf("min total spent cannot be greater than max total spent")
	}
	if f.MinLoyaltyPoints != nil && *f.MinLoyaltyPoints < 0 {
		return fmt.Errorf("min loyalty points cannot be negative")
	}
	if f.MaxLoyaltyPoints != nil && *f.MaxLoyaltyPoints < 0 {
		return fmt.Errorf("max loyalty points cannot be negative")
	}

	// Validate loyalty tiers
	if len(f.LoyaltyTiers) > 0 {
		validTiers := []string{"Bronze", "Silver", "Gold", "Platinum"}
		for _, tier := range f.LoyaltyTiers {
			if !slices.Contains(validTiers, tier) {
				return fmt.Errorf("invalid loyalty tier: %s", tier)
			}
		}
	}

	return nil
}

// CustomerAnalyticsFilter represents filters for customer analytics
type CustomerAnalyticsFilter struct {
	TenantID    uuid.UUID   `json:"tenant_id" validate:"required"`
	StartDate   time.Time   `json:"start_date" validate:"required"`
	EndDate     time.Time   `json:"end_date" validate:"required"`
	GroupBy     string      `json:"group_by,omitempty"` // day, week, month, year
	CustomerIDs []uuid.UUID `json:"customer_ids,omitempty"`
}

// Validate validates the customer analytics filter
func (f *CustomerAnalyticsFilter) Validate() error {
	if f.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID is required")
	}
	if f.StartDate.IsZero() {
		return fmt.Errorf("start date is required")
	}
	if f.EndDate.IsZero() {
		return fmt.Errorf("end date is required")
	}
	if f.StartDate.After(f.EndDate) {
		return fmt.Errorf("start date cannot be after end date")
	}

	if f.GroupBy != "" {
		validGroupBy := []string{"day", "week", "month", "year"}
		if !slices.Contains(validGroupBy, f.GroupBy) {
			return fmt.Errorf("invalid group by: %s", f.GroupBy)
		}
	}

	return nil
}

// ============================================================================
// Response DTOs
// ============================================================================

// CustomerResponse represents a customer response
type CustomerResponse struct {
	ID                     uuid.UUID        `json:"id"`
	UserID                 uuid.UUID        `json:"user_id"`
	TenantID               uuid.UUID        `json:"tenant_id"`
	PreferredArtisans      []uuid.UUID      `json:"preferred_artisans"`
	Notes                  string           `json:"notes,omitempty"`
	LoyaltyPoints          int              `json:"loyalty_points"`
	LoyaltyTier            string           `json:"loyalty_tier"`
	TotalSpent             float64          `json:"total_spent"`
	TotalBookings          int              `json:"total_bookings"`
	CancelledBookings      int              `json:"cancelled_bookings"`
	CompletedBookings      int              `json:"completed_bookings"`
	DefaultPaymentMethodID string           `json:"default_payment_method_id,omitempty"`
	PrimaryLocation        *models.Location `json:"primary_location,omitempty"`
	EmailNotifications     bool             `json:"email_notifications"`
	SMSNotifications       bool             `json:"sms_notifications"`
	PushNotifications      bool             `json:"push_notifications"`
	Metadata               models.JSONB     `json:"metadata,omitempty"`

	// User information
	User *UserInfoResponse `json:"user,omitempty"`

	// Calculated fields
	BookingCompletionRate float64    `json:"booking_completion_rate"`
	BookingCancelRate     float64    `json:"booking_cancel_rate"`
	AverageBookingValue   float64    `json:"average_booking_value"`
	LastBookingDate       *time.Time `json:"last_booking_date,omitempty"`
	DaysSinceLastBooking  *int       `json:"days_since_last_booking,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserInfoResponse represents user information in customer response
type UserInfoResponse struct {
	ID          uuid.UUID `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Status      string    `json:"status"`
}

// CustomerListResponse represents a paginated list of customers
type CustomerListResponse struct {
	Customers   []*CustomerResponse `json:"customers"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
	TotalItems  int64               `json:"total_items"`
	TotalPages  int                 `json:"total_pages"`
	HasNext     bool                `json:"has_next"`
	HasPrevious bool                `json:"has_previous"`
}

// CustomerStatsResponse represents customer statistics
type CustomerStatsResponse struct {
	TotalCustomers    int64 `json:"total_customers"`
	NewCustomers      int64 `json:"new_customers"`      // This period
	ActiveCustomers   int64 `json:"active_customers"`   // With bookings this period
	InactiveCustomers int64 `json:"inactive_customers"` // No bookings this period

	// Loyalty breakdown
	ByLoyaltyTier map[string]int64 `json:"by_loyalty_tier"`

	// Financial metrics
	TotalRevenue            float64 `json:"total_revenue"`
	AverageSpentPerCustomer float64 `json:"average_spent_per_customer"`
	HighestSpendingCustomer float64 `json:"highest_spending_customer"`

	// Booking metrics
	TotalBookings              int64   `json:"total_bookings"`
	AverageBookingsPerCustomer float64 `json:"average_bookings_per_customer"`
	BookingCompletionRate      float64 `json:"booking_completion_rate"`
	BookingCancelRate          float64 `json:"booking_cancel_rate"`

	// Growth metrics
	CustomerGrowthRate    float64 `json:"customer_growth_rate"`
	RevenueGrowthRate     float64 `json:"revenue_growth_rate"`
	CustomerRetentionRate float64 `json:"customer_retention_rate"`

	// Top customers
	TopCustomersBySpending []*CustomerResponse `json:"top_customers_by_spending"`
	TopCustomersByBookings []*CustomerResponse `json:"top_customers_by_bookings"`
}

// CustomerAnalyticsResponse represents customer analytics data
type CustomerAnalyticsResponse struct {
	Period     string                       `json:"period"`
	TenantID   uuid.UUID                    `json:"tenant_id"`
	GroupBy    string                       `json:"group_by,omitempty"`
	DataPoints []CustomerAnalyticsDataPoint `json:"data_points"`
	Summary    *CustomerAnalyticsSummary    `json:"summary"`
}

// CustomerAnalyticsDataPoint represents a single analytics data point
type CustomerAnalyticsDataPoint struct {
	Date            time.Time `json:"date"`
	NewCustomers    int64     `json:"new_customers"`
	ActiveCustomers int64     `json:"active_customers"`
	Revenue         float64   `json:"revenue"`
	Bookings        int64     `json:"bookings"`
	LoyaltyPoints   int64     `json:"loyalty_points_awarded"`
}

// CustomerAnalyticsSummary represents analytics summary
type CustomerAnalyticsSummary struct {
	TotalNewCustomers     int64   `json:"total_new_customers"`
	TotalRevenue          float64 `json:"total_revenue"`
	TotalBookings         int64   `json:"total_bookings"`
	AverageBookingValue   float64 `json:"average_booking_value"`
	CustomerGrowthRate    float64 `json:"customer_growth_rate"`
	RevenueGrowthRate     float64 `json:"revenue_growth_rate"`
	CustomerRetentionRate float64 `json:"customer_retention_rate"`
}

// LoyaltyPointsHistoryResponse represents loyalty points history
type LoyaltyPointsHistoryResponse struct {
	ID         uuid.UUID  `json:"id"`
	CustomerID uuid.UUID  `json:"customer_id"`
	Points     int        `json:"points"`
	Operation  string     `json:"operation"`
	Reason     string     `json:"reason"`
	BookingID  *uuid.UUID `json:"booking_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// CustomerSegmentResponse represents customer segment data
type CustomerSegmentResponse struct {
	SegmentName     string  `json:"segment_name"`
	CustomerCount   int64   `json:"customer_count"`
	Percentage      float64 `json:"percentage"`
	TotalRevenue    float64 `json:"total_revenue"`
	AverageSpending float64 `json:"average_spending"`
	Description     string  `json:"description"`
}

// ============================================================================
// Utility Functions
// ============================================================================

// ToCustomerResponse converts a models.Customer to CustomerResponse
func ToCustomerResponse(customer *models.Customer) *CustomerResponse {
	if customer == nil {
		return nil
	}

	response := &CustomerResponse{
		ID:                     customer.ID,
		UserID:                 customer.UserID,
		TenantID:               customer.TenantID,
		PreferredArtisans:      customer.PreferredArtisans,
		Notes:                  customer.Notes,
		LoyaltyPoints:          customer.LoyaltyPoints,
		LoyaltyTier:            customer.GetLoyaltyTier(),
		TotalSpent:             customer.TotalSpent,
		TotalBookings:          customer.TotalBookings,
		CancelledBookings:      customer.CancelledBookings,
		CompletedBookings:      customer.CompletedBookings,
		DefaultPaymentMethodID: customer.DefaultPaymentMethodID,
		PrimaryLocation:        &customer.PrimaryLocation,
		EmailNotifications:     customer.EmailNotifications,
		SMSNotifications:       customer.SMSNotifications,
		PushNotifications:      customer.PushNotifications,
		Metadata:               customer.Metadata,
		CreatedAt:              customer.CreatedAt,
		UpdatedAt:              customer.UpdatedAt,
	}

	// Calculate booking rates
	if customer.TotalBookings > 0 {
		response.BookingCompletionRate = (float64(customer.CompletedBookings) / float64(customer.TotalBookings)) * 100
		response.BookingCancelRate = (float64(customer.CancelledBookings) / float64(customer.TotalBookings)) * 100
	}

	// Calculate average booking value
	if customer.CompletedBookings > 0 {
		response.AverageBookingValue = customer.TotalSpent / float64(customer.CompletedBookings)
	}

	// Add user information if available
	if customer.User != nil {
		response.User = &UserInfoResponse{
			ID:          customer.User.ID,
			FirstName:   customer.User.FirstName,
			LastName:    customer.User.LastName,
			Email:       customer.User.Email,
			PhoneNumber: customer.User.PhoneNumber,
			AvatarURL:   customer.User.AvatarURL,
			Status:      string(customer.User.Status),
		}
	}

	return response
}

// ToCustomerResponses converts multiple models.Customer to CustomerResponse slice
func ToCustomerResponses(customers []*models.Customer) []*CustomerResponse {
	responses := make([]*CustomerResponse, 0, len(customers))
	for _, customer := range customers {
		responses = append(responses, ToCustomerResponse(customer))
	}
	return responses
}

// CustomerSearchRequest represents a customer search request
type CustomerSearchRequest struct {
	Query    string         `json:"query" validate:"required,min=2"`
	TenantID uuid.UUID      `json:"tenant_id,omitempty"` // Optional - will be extracted from auth context if not provided
	Filters  CustomerFilter `json:"filters,omitempty"`
}

// Validate validates the customer search request
func (r *CustomerSearchRequest) Validate() error {
	if len(strings.TrimSpace(r.Query)) < 2 {
		return fmt.Errorf("search query must be at least 2 characters")
	}
	// Note: TenantID is now optional and will be set from auth context by the handler
	return r.Filters.Validate()
}
