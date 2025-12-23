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

// CreateBookingRequest represents the request to create a booking
type CreateBookingRequest struct {
	TenantID              uuid.UUID        `json:"tenant_id" validate:"required"`
	ArtisanID             uuid.UUID        `json:"artisan_id" validate:"required"`
	CustomerID            uuid.UUID        `json:"customer_id" validate:"required"`
	ServiceID             uuid.UUID        `json:"service_id" validate:"required"`
	StartTime             time.Time        `json:"start_time" validate:"required"`
	Duration              int              `json:"duration" validate:"required,min=15,max=480"` // 15 min to 8 hours
	Notes                 string           `json:"notes,omitempty"`
	CustomerNotes         string           `json:"customer_notes,omitempty"`
	SelectedAddons        []uuid.UUID      `json:"selected_addons,omitempty"`
	ServiceLocation       *models.Location `json:"service_location,omitempty"`
	PaymentMethodID       string           `json:"payment_method_id,omitempty"`
	RequiresDeposit       bool             `json:"requires_deposit"`
	DepositAmount         float64          `json:"deposit_amount"`
	AutoConfirm           bool             `json:"auto_confirm"`
	SendConfirmationEmail bool             `json:"send_confirmation_email"`
	SendConfirmationSMS   bool             `json:"send_confirmation_sms"`
	IsRecurring           bool             `json:"is_recurring"`
	RecurrencePattern     string           `json:"recurrence_pattern,omitempty"` // weekly, biweekly, monthly
	RecurrenceEndDate     *time.Time       `json:"recurrence_end_date,omitempty"`
	RecurrenceOccurrences *int             `json:"recurrence_occurrences,omitempty"`
	Metadata              map[string]any   `json:"metadata,omitempty"`
}

// Validate validates the create booking request
func (r *CreateBookingRequest) Validate() error {
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID is required")
	}
	if r.ArtisanID == uuid.Nil {
		return fmt.Errorf("artisan ID is required")
	}
	if r.CustomerID == uuid.Nil {
		return fmt.Errorf("customer ID is required")
	}
	if r.ServiceID == uuid.Nil {
		return fmt.Errorf("service ID is required")
	}
	if r.StartTime.IsZero() {
		return fmt.Errorf("start time is required")
	}
	if r.StartTime.Before(time.Now()) {
		return fmt.Errorf("start time cannot be in the past")
	}
	if r.Duration < 15 || r.Duration > 480 {
		return fmt.Errorf("duration must be between 15 minutes and 8 hours")
	}
	if len(r.Notes) > 2000 {
		return fmt.Errorf("notes must be 2000 characters or less")
	}
	if len(r.CustomerNotes) > 1000 {
		return fmt.Errorf("customer notes must be 1000 characters or less")
	}
	if r.RequiresDeposit && r.DepositAmount <= 0 {
		return fmt.Errorf("deposit amount must be positive when deposit is required")
	}

	// Validate recurrence settings
	if r.IsRecurring {
		if r.RecurrencePattern == "" {
			return fmt.Errorf("recurrence pattern is required for recurring bookings")
		}
		validPatterns := []string{"weekly", "biweekly", "monthly"}
		if !slices.Contains(validPatterns, r.RecurrencePattern) {
			return fmt.Errorf("invalid recurrence pattern: %s", r.RecurrencePattern)
		}
		if r.RecurrenceEndDate == nil && r.RecurrenceOccurrences == nil {
			return fmt.Errorf("either end date or number of occurrences must be specified for recurring bookings")
		}
		if r.RecurrenceEndDate != nil && r.RecurrenceEndDate.Before(r.StartTime) {
			return fmt.Errorf("recurrence end date cannot be before start time")
		}
		if r.RecurrenceOccurrences != nil && *r.RecurrenceOccurrences < 2 {
			return fmt.Errorf("recurring bookings must have at least 2 occurrences")
		}
	}

	return nil
}

// UpdateBookingRequest represents the request to update a booking
type UpdateBookingRequest struct {
	Notes              *string               `json:"notes,omitempty"`
	CustomerNotes      *string               `json:"customer_notes,omitempty"`
	InternalNotes      *string               `json:"internal_notes,omitempty"`
	ServiceLocation    *models.Location      `json:"service_location,omitempty"`
	Status             *models.BookingStatus `json:"status,omitempty"`
	PaymentStatus      *models.PaymentStatus `json:"payment_status,omitempty"`
	SelectedAddons     []uuid.UUID           `json:"selected_addons,omitempty"`
	CancellationReason *string               `json:"cancellation_reason,omitempty"`
	PaymentIntentID    *string               `json:"payment_intent_id,omitempty"`
	RefundID           *string               `json:"refund_id,omitempty"`
	BeforePhotoURLs    []string              `json:"before_photo_urls,omitempty"`
	AfterPhotoURLs     []string              `json:"after_photo_urls,omitempty"`
	ReminderSent24h    *bool                 `json:"reminder_sent_24h,omitempty"`
	ReminderSent1h     *bool                 `json:"reminder_sent_1h,omitempty"`
	Metadata           map[string]any        `json:"metadata,omitempty"`
}

// Validate validates the update booking request
func (r *UpdateBookingRequest) Validate() error {
	if r.Notes != nil && len(*r.Notes) > 2000 {
		return fmt.Errorf("notes must be 2000 characters or less")
	}
	if r.CustomerNotes != nil && len(*r.CustomerNotes) > 1000 {
		return fmt.Errorf("customer notes must be 1000 characters or less")
	}
	if r.InternalNotes != nil && len(*r.InternalNotes) > 2000 {
		return fmt.Errorf("internal notes must be 2000 characters or less")
	}
	if r.CancellationReason != nil && len(*r.CancellationReason) > 500 {
		return fmt.Errorf("cancellation reason must be 500 characters or less")
	}
	return nil
}

// RescheduleBookingRequest represents the request to reschedule a booking
type RescheduleBookingRequest struct {
	NewStartTime     time.Time `json:"new_start_time" validate:"required"`
	NewDuration      *int      `json:"new_duration,omitempty"`
	Reason           string    `json:"reason,omitempty"`
	NotifyCustomer   bool      `json:"notify_customer"`
	NotifyArtisan    bool      `json:"notify_artisan"`
	RefundDifference bool      `json:"refund_difference"` // If new booking is cheaper
}

// Validate validates the reschedule booking request
func (r *RescheduleBookingRequest) Validate() error {
	if r.NewStartTime.IsZero() {
		return fmt.Errorf("new start time is required")
	}
	if r.NewStartTime.Before(time.Now()) {
		return fmt.Errorf("new start time cannot be in the past")
	}
	if r.NewDuration != nil && (*r.NewDuration < 15 || *r.NewDuration > 480) {
		return fmt.Errorf("duration must be between 15 minutes and 8 hours")
	}
	if len(r.Reason) > 500 {
		return fmt.Errorf("reason must be 500 characters or less")
	}
	return nil
}

// CancelBookingRequest represents the request to cancel a booking
type CancelBookingRequest struct {
	Reason          string    `json:"reason" validate:"required"`
	CancelledBy     uuid.UUID `json:"cancelled_by" validate:"required"`
	RefundRequested bool      `json:"refund_requested"`
	NotifyCustomer  bool      `json:"notify_customer"`
	NotifyArtisan   bool      `json:"notify_artisan"`
}

// Validate validates the cancel booking request
func (r *CancelBookingRequest) Validate() error {
	if r.CancelledBy == uuid.Nil {
		return fmt.Errorf("cancelled by user ID is required")
	}
	if r.Reason == "" {
		return fmt.Errorf("cancellation reason is required")
	}
	if len(r.Reason) > 500 {
		return fmt.Errorf("reason must be 500 characters or less")
	}
	return nil
}

// CompleteBookingRequest represents the request to complete a booking
type CompleteBookingRequest struct {
	CompletionNotes string   `json:"completion_notes,omitempty"`
	BeforePhotoURLs []string `json:"before_photo_urls,omitempty"`
	AfterPhotoURLs  []string `json:"after_photo_urls,omitempty"`
	ActualDuration  *int     `json:"actual_duration,omitempty"` // If different from planned
	QualityRating   *int     `json:"quality_rating,omitempty"`  // 1-5 rating
	RequestReview   bool     `json:"request_review"`
}

// Validate validates the complete booking request
func (r *CompleteBookingRequest) Validate() error {
	if len(r.CompletionNotes) > 1000 {
		return fmt.Errorf("completion notes must be 1000 characters or less")
	}
	if r.ActualDuration != nil && (*r.ActualDuration < 1 || *r.ActualDuration > 600) {
		return fmt.Errorf("actual duration must be between 1 minute and 10 hours")
	}
	if r.QualityRating != nil && (*r.QualityRating < 1 || *r.QualityRating > 5) {
		return fmt.Errorf("quality rating must be between 1 and 5")
	}
	return nil
}

// ============================================================================
// Filter DTOs
// ============================================================================

// BookingFilter represents filters for listing bookings
type BookingFilter struct {
	TenantID         *uuid.UUID             `json:"tenant_id,omitempty"`
	ArtisanIDs       []uuid.UUID            `json:"artisan_ids,omitempty"`
	CustomerIDs      []uuid.UUID            `json:"customer_ids,omitempty"`
	ServiceIDs       []uuid.UUID            `json:"service_ids,omitempty"`
	Statuses         []models.BookingStatus `json:"statuses,omitempty"`
	PaymentStatuses  []models.PaymentStatus `json:"payment_statuses,omitempty"`
	StartDate        *time.Time             `json:"start_date,omitempty"`
	EndDate          *time.Time             `json:"end_date,omitempty"`
	MinDuration      *int                   `json:"min_duration,omitempty"`
	MaxDuration      *int                   `json:"max_duration,omitempty"`
	MinAmount        *float64               `json:"min_amount,omitempty"`
	MaxAmount        *float64               `json:"max_amount,omitempty"`
	HasDeposit       *bool                  `json:"has_deposit,omitempty"`
	IsRecurring      *bool                  `json:"is_recurring,omitempty"`
	HasPhotos        *bool                  `json:"has_photos,omitempty"`
	HasLocation      *bool                  `json:"has_location,omitempty"`
	CreatedAfter     *time.Time             `json:"created_after,omitempty"`
	CreatedBefore    *time.Time             `json:"created_before,omitempty"`
	UpdatedAfter     *time.Time             `json:"updated_after,omitempty"`
	UpdatedBefore    *time.Time             `json:"updated_before,omitempty"`
	Page             int                    `json:"page" validate:"min=1"`
	PageSize         int                    `json:"page_size" validate:"min=1,max=100"`
	SortBy           string                 `json:"sort_by,omitempty"`
	SortOrder        string                 `json:"sort_order,omitempty"` // asc or desc
	SearchQuery      string                 `json:"search_query,omitempty"`
	IncludeRelations []string               `json:"include_relations,omitempty"` // artisan, customer, service, payments, review
}

// Validate validates the booking filter
func (f *BookingFilter) Validate() error {
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
	if f.StartDate != nil && f.EndDate != nil && f.StartDate.After(*f.EndDate) {
		return fmt.Errorf("start date cannot be after end date")
	}
	if f.MinDuration != nil && *f.MinDuration < 0 {
		return fmt.Errorf("min duration cannot be negative")
	}
	if f.MaxDuration != nil && *f.MaxDuration < 0 {
		return fmt.Errorf("max duration cannot be negative")
	}
	if f.MinDuration != nil && f.MaxDuration != nil && *f.MinDuration > *f.MaxDuration {
		return fmt.Errorf("min duration cannot be greater than max duration")
	}
	if f.MinAmount != nil && *f.MinAmount < 0 {
		return fmt.Errorf("min amount cannot be negative")
	}
	if f.MaxAmount != nil && *f.MaxAmount < 0 {
		return fmt.Errorf("max amount cannot be negative")
	}
	if f.MinAmount != nil && f.MaxAmount != nil && *f.MinAmount > *f.MaxAmount {
		return fmt.Errorf("min amount cannot be greater than max amount")
	}
	if f.CreatedAfter != nil && f.CreatedBefore != nil && f.CreatedAfter.After(*f.CreatedBefore) {
		return fmt.Errorf("created after cannot be after created before")
	}

	// Validate sort by fields
	if f.SortBy != "" {
		validSortFields := []string{"created_at", "updated_at", "start_time", "end_time", "total_price", "status", "duration"}
		if !slices.Contains(validSortFields, f.SortBy) {
			return fmt.Errorf("invalid sort field: %s", f.SortBy)
		}
	}

	// Validate include relations
	validRelations := []string{"artisan", "customer", "service", "payments", "review", "tenant"}
	for _, relation := range f.IncludeRelations {
		if !slices.Contains(validRelations, relation) {
			return fmt.Errorf("invalid relation: %s", relation)
		}
	}

	return nil
}

// BookingSearchRequest represents a booking search request
type BookingSearchRequest struct {
	Query    string        `json:"query" validate:"required,min=2"`
	TenantID uuid.UUID     `json:"tenant_id" validate:"required"`
	Filters  BookingFilter `json:"filters,omitempty"`
}

// Validate validates the booking search request
func (r *BookingSearchRequest) Validate() error {
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID is required")
	}
	if len(strings.TrimSpace(r.Query)) < 2 {
		return fmt.Errorf("search query must be at least 2 characters")
	}
	return r.Filters.Validate()
}

// BookingAnalyticsFilter represents filters for booking analytics
type BookingAnalyticsFilter struct {
	TenantID        uuid.UUID   `json:"tenant_id" validate:"required"`
	StartDate       time.Time   `json:"start_date" validate:"required"`
	EndDate         time.Time   `json:"end_date" validate:"required"`
	GroupBy         string      `json:"group_by,omitempty"` // day, week, month, year
	ArtisanIDs      []uuid.UUID `json:"artisan_ids,omitempty"`
	ServiceIDs      []uuid.UUID `json:"service_ids,omitempty"`
	CustomerIDs     []uuid.UUID `json:"customer_ids,omitempty"`
	IncludeRefunds  bool        `json:"include_refunds"`
	IncludeDeposits bool        `json:"include_deposits"`
}

// Validate validates the booking analytics filter
func (f *BookingAnalyticsFilter) Validate() error {
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

// AvailabilityRequest represents a request to check artisan availability
type AvailabilityRequest struct {
	ArtisanID        uuid.UUID  `json:"artisan_id" validate:"required"`
	Date             time.Time  `json:"date" validate:"required"`
	Duration         int        `json:"duration" validate:"required,min=15,max=480"`
	ServiceID        *uuid.UUID `json:"service_id,omitempty"`
	ExcludeBookingID *uuid.UUID `json:"exclude_booking_id,omitempty"`
	TimeZone         string     `json:"timezone,omitempty"`
}

// Validate validates the availability request
func (r *AvailabilityRequest) Validate() error {
	if r.ArtisanID == uuid.Nil {
		return fmt.Errorf("artisan ID is required")
	}
	if r.Date.IsZero() {
		return fmt.Errorf("date is required")
	}
	if r.Duration < 15 || r.Duration > 480 {
		return fmt.Errorf("duration must be between 15 minutes and 8 hours")
	}
	return nil
}

// ============================================================================
// Response DTOs
// ============================================================================

// BookingResponse represents a booking response
type BookingResponse struct {
	ID                 uuid.UUID            `json:"id"`
	TenantID           uuid.UUID            `json:"tenant_id"`
	ArtisanID          uuid.UUID            `json:"artisan_id"`
	CustomerID         uuid.UUID            `json:"customer_id"`
	ServiceID          uuid.UUID            `json:"service_id"`
	StartTime          time.Time            `json:"start_time"`
	EndTime            time.Time            `json:"end_time"`
	Duration           int                  `json:"duration"`
	Status             models.BookingStatus `json:"status"`
	PaymentStatus      models.PaymentStatus `json:"payment_status"`
	BasePrice          float64              `json:"base_price"`
	AddonsPrice        float64              `json:"addons_price"`
	TotalPrice         float64              `json:"total_price"`
	DepositPaid        float64              `json:"deposit_paid"`
	Currency           string               `json:"currency"`
	Notes              string               `json:"notes,omitempty"`
	CustomerNotes      string               `json:"customer_notes,omitempty"`
	InternalNotes      string               `json:"internal_notes,omitempty"`
	SelectedAddons     []uuid.UUID          `json:"selected_addons,omitempty"`
	ServiceLocation    *models.Location     `json:"service_location,omitempty"`
	CancelledAt        *time.Time           `json:"cancelled_at,omitempty"`
	CancelledBy        *uuid.UUID           `json:"cancelled_by,omitempty"`
	CancellationReason string               `json:"cancellation_reason,omitempty"`
	CompletedAt        *time.Time           `json:"completed_at,omitempty"`
	BeforePhotoURLs    []string             `json:"before_photo_urls,omitempty"`
	AfterPhotoURLs     []string             `json:"after_photo_urls,omitempty"`
	PaymentIntentID    string               `json:"payment_intent_id,omitempty"`
	RefundID           string               `json:"refund_id,omitempty"`
	IsRecurring        bool                 `json:"is_recurring"`
	RecurrencePattern  string               `json:"recurrence_pattern,omitempty"`
	ParentBookingID    *uuid.UUID           `json:"parent_booking_id,omitempty"`
	RecurrenceEndDate  *time.Time           `json:"recurrence_end_date,omitempty"`
	ReminderSent24h    bool                 `json:"reminder_sent_24h"`
	ReminderSent1h     bool                 `json:"reminder_sent_1h"`
	Metadata           models.JSONB         `json:"metadata,omitempty"`

	// Related entities (populated based on include_relations)
	Artisan  *ArtisanInfoResponse   `json:"artisan,omitempty"`
	Customer *CustomerInfoResponse  `json:"customer,omitempty"`
	Service  *ServiceInfoResponse   `json:"service,omitempty"`
	Payments []*PaymentInfoResponse `json:"payments,omitempty"`
	Review   *ReviewInfoResponse    `json:"review,omitempty"`

	// Calculated fields
	CanBeCancelled   bool    `json:"can_be_cancelled"`
	CanBeRescheduled bool    `json:"can_be_rescheduled"`
	CanBeCompleted   bool    `json:"can_be_completed"`
	RefundAmount     float64 `json:"refund_amount"`
	TimeUntilStart   int64   `json:"time_until_start"` // seconds
	IsUpcoming       bool    `json:"is_upcoming"`
	IsOverdue        bool    `json:"is_overdue"`
	StatusColor      string  `json:"status_color"`
	StatusLabel      string  `json:"status_label"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ArtisanInfoResponse represents artisan information in booking response
type ArtisanInfoResponse struct {
	ID          uuid.UUID `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Rating      float64   `json:"rating"`
	ReviewCount int       `json:"review_count"`
}

// CustomerInfoResponse represents customer information in booking response
type CustomerInfoResponse struct {
	ID            uuid.UUID `json:"id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Email         string    `json:"email"`
	PhoneNumber   string    `json:"phone_number,omitempty"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	LoyaltyTier   string    `json:"loyalty_tier"`
	TotalBookings int       `json:"total_bookings"`
}

// ServiceInfoResponse represents service information in booking response
type ServiceInfoResponse struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    models.ServiceCategory `json:"category"`
	Duration    int                    `json:"duration"`
	BasePrice   float64                `json:"base_price"`
	Currency    string                 `json:"currency"`
	ImageURL    string                 `json:"image_url,omitempty"`
}

// PaymentInfoResponse represents payment information in booking response
type PaymentInfoResponse struct {
	ID          uuid.UUID            `json:"id"`
	Amount      float64              `json:"amount"`
	Currency    string               `json:"currency"`
	Method      models.PaymentMethod `json:"method"`
	Type        models.PaymentType   `json:"type"`
	Status      models.PaymentStatus `json:"status"`
	ProcessedAt *time.Time           `json:"processed_at,omitempty"`
}

// ReviewInfoResponse represents review information in booking response
type ReviewInfoResponse struct {
	ID        uuid.UUID `json:"id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// BookingPaymentResponse represents a payment response for booking operations
type BookingPaymentResponse struct {
	ID          uuid.UUID            `json:"id"`
	Amount      float64              `json:"amount"`
	Currency    string               `json:"currency"`
	Status      models.PaymentStatus `json:"status"`
	ProcessedAt *time.Time           `json:"processed_at,omitempty"`
	Method      models.PaymentMethod `json:"method,omitempty"`
	Type        models.PaymentType   `json:"type,omitempty"`
}

// BookingListResponse represents a paginated list of bookings
type BookingListResponse struct {
	Bookings    []*BookingResponse `json:"bookings"`
	Page        int                `json:"page"`
	PageSize    int                `json:"page_size"`
	TotalItems  int64              `json:"total_items"`
	TotalPages  int                `json:"total_pages"`
	HasNext     bool               `json:"has_next"`
	HasPrevious bool               `json:"has_previous"`
}

// TimeSlotResponse represents an available time slot
type TimeSlotResponse struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  int       `json:"duration"`
	Available bool      `json:"available"`
	Reason    string    `json:"reason,omitempty"` // If not available
}

// AvailabilityResponse represents artisan availability for a day
type AvailabilityResponse struct {
	ArtisanID    uuid.UUID             `json:"artisan_id"`
	Date         time.Time             `json:"date"`
	IsAvailable  bool                  `json:"is_available"`
	TimeSlots    []*TimeSlotResponse   `json:"time_slots"`
	WorkingHours *WorkingHoursResponse `json:"working_hours,omitempty"`
	Conflicts    []*ConflictResponse   `json:"conflicts,omitempty"`
}

// WorkingHoursResponse represents working hours
type WorkingHoursResponse struct {
	StartTime string `json:"start_time"` // Format: "09:00"
	EndTime   string `json:"end_time"`   // Format: "17:00"
	TimeZone  string `json:"timezone"`
}

// ConflictResponse represents a booking conflict
type ConflictResponse struct {
	ConflictType string     `json:"conflict_type"` // booking, break, unavailable
	StartTime    time.Time  `json:"start_time"`
	EndTime      time.Time  `json:"end_time"`
	BookingID    *uuid.UUID `json:"booking_id,omitempty"`
	Reason       string     `json:"reason,omitempty"`
}

// BookingStatsResponse represents booking statistics
type BookingStatsResponse struct {
	TotalBookings      int64 `json:"total_bookings"`
	PendingBookings    int64 `json:"pending_bookings"`
	ConfirmedBookings  int64 `json:"confirmed_bookings"`
	InProgressBookings int64 `json:"in_progress_bookings"`
	CompletedBookings  int64 `json:"completed_bookings"`
	CancelledBookings  int64 `json:"cancelled_bookings"`
	NoShowBookings     int64 `json:"no_show_bookings"`

	// Financial metrics
	TotalRevenue        float64 `json:"total_revenue"`
	AverageBookingValue float64 `json:"average_booking_value"`
	TotalDeposits       float64 `json:"total_deposits"`
	TotalRefunds        float64 `json:"total_refunds"`

	// Performance metrics
	CompletionRate   float64 `json:"completion_rate"`
	CancellationRate float64 `json:"cancellation_rate"`
	NoShowRate       float64 `json:"no_show_rate"`
	OnTimeRate       float64 `json:"on_time_rate"`
	AverageDuration  float64 `json:"average_duration"`

	// Trend data
	BookingTrends   []BookingTrendData   `json:"booking_trends"`
	RevenueTrends   []RevenueTrendData   `json:"revenue_trends"`
	PopularServices []PopularServiceData `json:"popular_services"`
	TopArtisans     []TopArtisanData     `json:"top_artisans"`
	TopCustomers    []TopCustomerData    `json:"top_customers"`
}

// BookingTrendData represents booking trend data
type BookingTrendData struct {
	Date     time.Time `json:"date"`
	Bookings int64     `json:"bookings"`
	Revenue  float64   `json:"revenue"`
}

// RevenueTrendData represents revenue trend data
type RevenueTrendData struct {
	Date     time.Time `json:"date"`
	Revenue  float64   `json:"revenue"`
	Bookings int64     `json:"bookings"`
}

// PopularServiceData represents popular service data
type PopularServiceData struct {
	ServiceID     uuid.UUID `json:"service_id"`
	ServiceName   string    `json:"service_name"`
	BookingCount  int64     `json:"booking_count"`
	TotalRevenue  float64   `json:"total_revenue"`
	AverageRating float64   `json:"average_rating"`
}

// TopArtisanData represents top artisan data
type TopArtisanData struct {
	ArtisanID      uuid.UUID `json:"artisan_id"`
	ArtisanName    string    `json:"artisan_name"`
	BookingCount   int64     `json:"booking_count"`
	TotalRevenue   float64   `json:"total_revenue"`
	AverageRating  float64   `json:"average_rating"`
	CompletionRate float64   `json:"completion_rate"`
}

// TopCustomerData represents top customer data
type TopCustomerData struct {
	CustomerID   uuid.UUID `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	BookingCount int64     `json:"booking_count"`
	TotalSpent   float64   `json:"total_spent"`
	LoyaltyTier  string    `json:"loyalty_tier"`
}

// ============================================================================
// Utility Functions
// ============================================================================

// ToBookingResponse converts a models.Booking to BookingResponse
func ToBookingResponse(booking *models.Booking) *BookingResponse {
	if booking == nil {
		return nil
	}

	response := &BookingResponse{
		ID:                 booking.ID,
		TenantID:           booking.TenantID,
		ArtisanID:          booking.ArtisanID,
		CustomerID:         booking.CustomerID,
		ServiceID:          booking.ServiceID,
		StartTime:          booking.StartTime,
		EndTime:            booking.EndTime,
		Duration:           booking.Duration,
		Status:             booking.Status,
		PaymentStatus:      booking.PaymentStatus,
		BasePrice:          booking.BasePrice,
		AddonsPrice:        booking.AddonsPrice,
		TotalPrice:         booking.TotalPrice,
		DepositPaid:        booking.DepositPaid,
		Currency:           booking.Currency,
		Notes:              booking.Notes,
		CustomerNotes:      booking.CustomerNotes,
		InternalNotes:      booking.InternalNotes,
		SelectedAddons:     booking.SelectedAddons,
		ServiceLocation:    booking.ServiceLocation,
		CancelledAt:        booking.CancelledAt,
		CancelledBy:        booking.CancelledBy,
		CancellationReason: booking.CancellationReason,
		CompletedAt:        booking.CompletedAt,
		BeforePhotoURLs:    booking.BeforePhotoURLs,
		AfterPhotoURLs:     booking.AfterPhotoURLs,
		PaymentIntentID:    booking.PaymentIntentID,
		RefundID:           booking.RefundID,
		IsRecurring:        booking.IsRecurring,
		RecurrencePattern:  booking.RecurrencePattern,
		ParentBookingID:    booking.ParentBookingID,
		RecurrenceEndDate:  booking.RecurrenceEndDate,
		ReminderSent24h:    booking.ReminderSent24h,
		ReminderSent1h:     booking.ReminderSent1h,
		Metadata:           booking.Metadata,
		CreatedAt:          booking.CreatedAt,
		UpdatedAt:          booking.UpdatedAt,
	}

	// Calculate derived fields
	response.CanBeCancelled = booking.CanBeCancelled()
	response.CanBeRescheduled = booking.Status == models.BookingStatusPending || booking.Status == models.BookingStatusConfirmed
	response.CanBeCompleted = booking.Status == models.BookingStatusInProgress
	response.RefundAmount = booking.CalculateRefundAmount()
	response.TimeUntilStart = int64(time.Until(booking.StartTime).Seconds())
	response.IsUpcoming = booking.IsUpcoming()
	response.IsOverdue = time.Now().After(booking.EndTime) && booking.Status != models.BookingStatusCompleted
	response.StatusColor = getStatusColor(booking.Status)
	response.StatusLabel = getStatusLabel(booking.Status)

	// Add related entities if available
	if booking.Artisan != nil {
		response.Artisan = &ArtisanInfoResponse{
			ID:          booking.Artisan.ID,
			FirstName:   booking.Artisan.FirstName,
			LastName:    booking.Artisan.LastName,
			Email:       booking.Artisan.Email,
			PhoneNumber: booking.Artisan.PhoneNumber,
			AvatarURL:   booking.Artisan.AvatarURL,
			// Rating and ReviewCount would need to be calculated separately
		}
	}

	if booking.Customer != nil {
		response.Customer = &CustomerInfoResponse{
			ID:          booking.Customer.ID,
			FirstName:   booking.Customer.FirstName,
			LastName:    booking.Customer.LastName,
			Email:       booking.Customer.Email,
			PhoneNumber: booking.Customer.PhoneNumber,
			AvatarURL:   booking.Customer.AvatarURL,
			// LoyaltyTier and TotalBookings would need to be calculated separately
		}
	}

	if booking.Service != nil {
		response.Service = &ServiceInfoResponse{
			ID:          booking.Service.ID,
			Name:        booking.Service.Name,
			Description: booking.Service.Description,
			Category:    booking.Service.Category,
			Duration:    booking.Service.DurationMinutes,
			BasePrice:   booking.Service.Price,
			Currency:    booking.Service.Currency,
			ImageURL:    booking.Service.ImageURL,
		}
	}

	// Convert payments if available
	if booking.Payments != nil {
		response.Payments = make([]*PaymentInfoResponse, len(booking.Payments))
		for i, payment := range booking.Payments {
			response.Payments[i] = &PaymentInfoResponse{
				ID:          payment.ID,
				Amount:      payment.Amount,
				Currency:    payment.Currency,
				Method:      payment.Method,
				Type:        payment.Type,
				Status:      payment.Status,
				ProcessedAt: payment.ProcessedAt,
			}
		}
	}

	if booking.Review != nil {
		response.Review = &ReviewInfoResponse{
			ID:        booking.Review.ID,
			Rating:    booking.Review.Rating,
			Comment:   booking.Review.Comment,
			CreatedAt: booking.Review.CreatedAt,
		}
	}

	return response
}

// ToBookingResponses converts multiple models.Booking to BookingResponse slice
func ToBookingResponses(bookings []*models.Booking) []*BookingResponse {
	responses := make([]*BookingResponse, 0, len(bookings))
	for _, booking := range bookings {
		responses = append(responses, ToBookingResponse(booking))
	}
	return responses
}

// Helper functions for status formatting
func getStatusColor(status models.BookingStatus) string {
	switch status {
	case models.BookingStatusPending:
		return "orange"
	case models.BookingStatusConfirmed:
		return "blue"
	case models.BookingStatusInProgress:
		return "green"
	case models.BookingStatusCompleted:
		return "gray"
	case models.BookingStatusCancelled:
		return "red"
	case models.BookingStatusNoShow:
		return "purple"
	default:
		return "gray"
	}
}

func getStatusLabel(status models.BookingStatus) string {
	switch status {
	case models.BookingStatusPending:
		return "Pending"
	case models.BookingStatusConfirmed:
		return "Confirmed"
	case models.BookingStatusInProgress:
		return "In Progress"
	case models.BookingStatusCompleted:
		return "Completed"
	case models.BookingStatusCancelled:
		return "Cancelled"
	case models.BookingStatusNoShow:
		return "No Show"
	default:
		return string(status)
	}
}
