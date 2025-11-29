package models

import (
	"time"

	"github.com/google/uuid"
)

type WebhookEventType string

const (
	WebhookEventBookingCreated   WebhookEventType = "booking.created"
	WebhookEventBookingUpdated   WebhookEventType = "booking.updated"
	WebhookEventBookingCancelled WebhookEventType = "booking.cancelled"
	WebhookEventPaymentReceived  WebhookEventType = "payment.received"
	WebhookEventReviewCreated    WebhookEventType = "review.created"
	WebhookEventUserCreated      WebhookEventType = "user.created"
)

type WebhookEvent struct {
	BaseModel

	// Multi-tenancy
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`

	// Event Details
	EventType WebhookEventType `json:"event_type" gorm:"type:varchar(100);not null;index" validate:"required"`
	Payload   JSONB            `json:"payload" gorm:"type:jsonb;not null" validate:"required"`

	// Delivery
	WebhookURL      string     `json:"webhook_url" gorm:"size:500;not null" validate:"required,url"`
	AttemptCount    int        `json:"attempt_count" gorm:"default:0"`
	MaxAttempts     int        `json:"max_attempts" gorm:"default:3"`
	NextRetryAt     *time.Time `json:"next_retry_at,omitempty"`
	LastAttemptedAt *time.Time `json:"last_attempted_at,omitempty"`

	// Status
	Delivered     bool       `json:"delivered" gorm:"default:false;index"`
	DeliveredAt   *time.Time `json:"delivered_at,omitempty"`
	FailureReason string     `json:"failure_reason,omitempty" gorm:"type:text"`

	// Response
	ResponseCode int    `json:"response_code,omitempty"`
	ResponseBody string `json:"response_body,omitempty" gorm:"type:text"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`
}

// Business Methods
func (we *WebhookEvent) ShouldRetry() bool {
	return !we.Delivered && we.AttemptCount < we.MaxAttempts
}

func (we *WebhookEvent) CanRetryNow() bool {
	return we.ShouldRetry() && (we.NextRetryAt == nil || time.Now().After(*we.NextRetryAt))
}
