package models

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingStatusPending    BookingStatus = "pending"
	BookingStatusConfirmed  BookingStatus = "confirmed"
	BookingStatusInProgress BookingStatus = "in_progress"
	BookingStatusCompleted  BookingStatus = "completed"
	BookingStatusCancelled  BookingStatus = "cancelled"
	BookingStatusNoShow     BookingStatus = "no_show"
)

type Booking struct {
	BaseModel

	// Multi-tenancy
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index:idx_booking_tenant_date;index:idx_booking_tenant_artisan"`

	// Parties
	ArtisanID  uuid.UUID `json:"artisan_id" gorm:"type:uuid;not null;index:idx_booking_tenant_artisan;index:idx_booking_artisan_status" validate:"required"`
	CustomerID uuid.UUID `json:"customer_id" gorm:"type:uuid;not null;index:idx_booking_customer_status" validate:"required"`
	ServiceID  uuid.UUID `json:"service_id" gorm:"type:uuid;not null;index" validate:"required"`

	// Timing
	StartTime time.Time `json:"start_time" gorm:"not null;index:idx_booking_tenant_date;index:idx_booking_artisan_status" validate:"required"`
	EndTime   time.Time `json:"end_time" gorm:"not null" validate:"required,gtfield=StartTime"`
	Duration  int       `json:"duration" gorm:"not null"` // minutes

	// Status
	Status        BookingStatus `json:"status" gorm:"type:varchar(50);not null;default:'pending';index:idx_booking_artisan_status;index:idx_booking_customer_status" validate:"required"`
	PaymentStatus PaymentStatus `json:"payment_status" gorm:"type:varchar(50);not null;default:'pending';index" validate:"required"`

	// Pricing
	BasePrice   float64 `json:"base_price" gorm:"type:decimal(10,2);not null" validate:"required,min=0"`
	AddonsPrice float64 `json:"addons_price" gorm:"type:decimal(10,2);default:0"`
	TotalPrice  float64 `json:"total_price" gorm:"type:decimal(10,2);not null" validate:"required,min=0"`
	DepositPaid float64 `json:"deposit_paid" gorm:"type:decimal(10,2);default:0"`
	Currency    string  `json:"currency" gorm:"size:3;default:'USD'"`

	// Details
	Notes          string      `json:"notes,omitempty" gorm:"type:text"`
	CustomerNotes  string      `json:"customer_notes,omitempty" gorm:"type:text"`
	InternalNotes  string      `json:"internal_notes,omitempty" gorm:"type:text"`
	SelectedAddons []uuid.UUID `json:"selected_addons,omitempty" gorm:"type:uuid[]"`

	// Location (for mobile services)
	ServiceLocation *Location `json:"service_location,omitempty" gorm:"type:jsonb"`

	// Cancellation
	CancelledAt        *time.Time `json:"cancelled_at,omitempty"`
	CancelledBy        *uuid.UUID `json:"cancelled_by,omitempty"`
	CancellationReason string     `json:"cancellation_reason,omitempty" gorm:"type:text"`

	// Completion
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	BeforePhotoURLs []string   `json:"before_photo_urls,omitempty" gorm:"type:text[]"`
	AfterPhotoURLs  []string   `json:"after_photo_urls,omitempty" gorm:"type:text[]"`

	// Payment
	PaymentIntentID string `json:"payment_intent_id,omitempty" gorm:"size:255"`
	RefundID        string `json:"refund_id,omitempty" gorm:"size:255"`

	// Recurrence
	IsRecurring       bool       `json:"is_recurring" gorm:"default:false"`
	RecurrencePattern string     `json:"recurrence_pattern,omitempty" gorm:"size:50"` // weekly, biweekly, monthly
	ParentBookingID   *uuid.UUID `json:"parent_booking_id,omitempty" gorm:"type:uuid;index"`
	RecurrenceEndDate *time.Time `json:"recurrence_end_date,omitempty"`

	// Reminders
	ReminderSent24h bool `json:"reminder_sent_24h" gorm:"default:false"`
	ReminderSent1h  bool `json:"reminder_sent_1h" gorm:"default:false"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Tenant        *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Artisan       *User     `json:"artisan,omitempty" gorm:"foreignKey:ArtisanID"`
	Customer      *User     `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Service       *Service  `json:"service,omitempty" gorm:"foreignKey:ServiceID"`
	Payments      []Payment `json:"payments,omitempty" gorm:"foreignKey:BookingID"`
	Review        *Review   `json:"review,omitempty" gorm:"foreignKey:BookingID"`
	ParentBooking *Booking  `json:"parent_booking,omitempty" gorm:"foreignKey:ParentBookingID"`
	ChildBookings []Booking `json:"child_bookings,omitempty" gorm:"foreignKey:ParentBookingID"`
}

// Business Methods
func (b *Booking) CanBeCancelled() bool {
	return b.Status == BookingStatusPending || b.Status == BookingStatusConfirmed
}

func (b *Booking) IsUpcoming() bool {
	return time.Now().Before(b.StartTime) && (b.Status == BookingStatusPending || b.Status == BookingStatusConfirmed)
}

func (b *Booking) IsInProgress() bool {
	now := time.Now()
	return now.After(b.StartTime) && now.Before(b.EndTime) && b.Status == BookingStatusInProgress
}

func (b *Booking) CalculateRefundAmount() float64 {
	hoursUntil := time.Until(b.StartTime).Hours()

	switch {
	case hoursUntil >= 24:
		return b.TotalPrice // Full refund
	case hoursUntil >= 12:
		return b.TotalPrice * 0.75 // 75% refund
	case hoursUntil >= 6:
		return b.TotalPrice * 0.50 // 50% refund
	default:
		return 0 // No refund
	}
}

func (b *Booking) RequiresDeposit() bool {
	return b.DepositPaid > 0
}
