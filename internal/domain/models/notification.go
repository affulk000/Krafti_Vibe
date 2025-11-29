package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeBookingCreated   NotificationType = "booking_created"
	NotificationTypeBookingConfirmed NotificationType = "booking_confirmed"
	NotificationTypeBookingCancelled NotificationType = "booking_cancelled"
	NotificationTypeBookingReminder  NotificationType = "booking_reminder"
	NotificationTypeBookingCompleted NotificationType = "booking_completed"
	NotificationTypePaymentReceived  NotificationType = "payment_received"
	NotificationTypeReviewReceived   NotificationType = "review_received"
	NotificationTypeMessageReceived  NotificationType = "message_received"
	NotificationTypeSystem           NotificationType = "system"
)

type NotificationChannel string

const (
	NotificationChannelInApp NotificationChannel = "in_app"
	NotificationChannelEmail NotificationChannel = "email"
	NotificationChannelSMS   NotificationChannel = "sms"
	NotificationChannelPush  NotificationChannel = "push"
)

type Notification struct {
	BaseModel

	// Multi-tenancy
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`

	// Recipient
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index" validate:"required"`

	// Type & Content
	Type    NotificationType `json:"type" gorm:"type:varchar(50);not null;index" validate:"required"`
	Title   string           `json:"title" gorm:"not null;size:255" validate:"required"`
	Message string           `json:"message" gorm:"type:text;not null" validate:"required"`

	// Channels
	Channels []NotificationChannel `json:"channels" gorm:"type:text[];not null" validate:"required,min=1"`

	// Status per Channel
	SentViaInApp bool `json:"sent_via_in_app" gorm:"default:false"`
	SentViaEmail bool `json:"sent_via_email" gorm:"default:false"`
	SentViaSMS   bool `json:"sent_via_sms" gorm:"default:false"`
	SentViaPush  bool `json:"sent_via_push" gorm:"default:false"`

	// Read Status
	IsRead bool       `json:"is_read" gorm:"default:false;index"`
	ReadAt *time.Time `json:"read_at,omitempty"`

	// Action
	ActionURL  string `json:"action_url,omitempty" gorm:"size:500"`
	ActionText string `json:"action_text,omitempty" gorm:"size:100"`

	// Related Entity
	RelatedEntityType string     `json:"related_entity_type,omitempty" gorm:"size:50"` // booking, payment, review
	RelatedEntityID   *uuid.UUID `json:"related_entity_id,omitempty" gorm:"type:uuid"`

	// Priority
	Priority int `json:"priority" gorm:"default:5;check:priority >= 1 AND priority <= 10"` // 1=highest, 10=lowest

	// Expiry
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Business Methods
func (n *Notification) MarkAsRead() {
	now := time.Now()
	n.IsRead = true
	n.ReadAt = &now
}

func (n *Notification) IsExpired() bool {
	return n.ExpiresAt != nil && time.Now().After(*n.ExpiresAt)
}

func (n *Notification) ShouldSendVia(channel NotificationChannel) bool {
	for _, ch := range n.Channels {
		if ch == channel {
			return true
		}
	}
	return false
}
