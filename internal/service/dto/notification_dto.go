package dto

import (
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Notification Request DTOs
// ============================================================================

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	TenantID          uuid.UUID                    `json:"tenant_id" validate:"required"`
	UserID            uuid.UUID                    `json:"user_id" validate:"required"`
	Type              models.NotificationType      `json:"type" validate:"required"`
	Title             string                       `json:"title" validate:"required,max=255"`
	Message           string                       `json:"message" validate:"required"`
	Channels          []models.NotificationChannel `json:"channels" validate:"required,min=1"`
	ActionURL         string                       `json:"action_url,omitempty"`
	ActionText        string                       `json:"action_text,omitempty"`
	RelatedEntityType string                       `json:"related_entity_type,omitempty"`
	RelatedEntityID   *uuid.UUID                   `json:"related_entity_id,omitempty"`
	Priority          int                          `json:"priority" validate:"min=1,max=10"`
	ExpiresAt         *time.Time                   `json:"expires_at,omitempty"`
	Metadata          map[string]any               `json:"metadata,omitempty"`
}

// Validate validates the create notification request
func (r *CreateNotificationRequest) Validate() error {
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("tenant_id is required")
	}
	if r.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if r.Title == "" {
		return fmt.Errorf("title is required")
	}
	if r.Message == "" {
		return fmt.Errorf("message is required")
	}
	if len(r.Channels) == 0 {
		return fmt.Errorf("at least one channel is required")
	}
	if r.Priority < 1 || r.Priority > 10 {
		return fmt.Errorf("priority must be between 1 and 10")
	}
	return nil
}

// BulkNotificationRequest represents bulk notification request
type BulkNotificationRequest struct {
	TenantID  uuid.UUID                    `json:"tenant_id" validate:"required"`
	UserIDs   []uuid.UUID                  `json:"user_ids" validate:"required,min=1"`
	Type      models.NotificationType      `json:"type" validate:"required"`
	Title     string                       `json:"title" validate:"required,max=255"`
	Message   string                       `json:"message" validate:"required"`
	Channels  []models.NotificationChannel `json:"channels" validate:"required,min=1"`
	ActionURL string                       `json:"action_url,omitempty"`
	Priority  int                          `json:"priority" validate:"min=1,max=10"`
	ExpiresAt *time.Time                   `json:"expires_at,omitempty"`
}

// Validate validates the bulk notification request
func (r *BulkNotificationRequest) Validate() error {
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("tenant_id is required")
	}
	if len(r.UserIDs) == 0 {
		return fmt.Errorf("at least one user_id is required")
	}
	if r.Title == "" {
		return fmt.Errorf("title is required")
	}
	if r.Message == "" {
		return fmt.Errorf("message is required")
	}
	if len(r.Channels) == 0 {
		return fmt.Errorf("at least one channel is required")
	}
	r.Priority = max(1, min(r.Priority, 10))
	if r.Priority == 0 {
		r.Priority = 5 // Default priority
	}
	return nil
}

// NotificationFilter represents filters for notification queries
type NotificationFilter struct {
	TenantID          uuid.UUID                `json:"tenant_id"`
	UserID            uuid.UUID                `json:"user_id"`
	Type              *models.NotificationType `json:"type,omitempty"`
	IsRead            *bool                    `json:"is_read,omitempty"`
	Priority          *int                     `json:"priority,omitempty"`
	RelatedEntityType *string                  `json:"related_entity_type,omitempty"`
	FromDate          *time.Time               `json:"from_date,omitempty"`
	ToDate            *time.Time               `json:"to_date,omitempty"`
	Page              int                      `json:"page"`
	PageSize          int                      `json:"page_size"`
}

// MarkAsReadRequest represents a request to mark notifications as read
type MarkAsReadRequest struct {
	NotificationIDs []uuid.UUID `json:"notification_ids" validate:"required,min=1"`
}

// SendBookingNotificationRequest represents booking notification request
type SendBookingNotificationRequest struct {
	BookingID uuid.UUID               `json:"booking_id" validate:"required"`
	Type      models.NotificationType `json:"type" validate:"required"`
}

// SendPaymentNotificationRequest represents payment notification request
type SendPaymentNotificationRequest struct {
	PaymentID uuid.UUID               `json:"payment_id" validate:"required"`
	Type      models.NotificationType `json:"type" validate:"required"`
}

// SendSystemNotificationRequest represents system notification request
type SendSystemNotificationRequest struct {
	TenantID uuid.UUID   `json:"tenant_id" validate:"required"`
	UserIDs  []uuid.UUID `json:"user_ids" validate:"required,min=1"`
	Title    string      `json:"title" validate:"required,max=255"`
	Message  string      `json:"message" validate:"required"`
	Priority int         `json:"priority" validate:"min=1,max=10"`
}

// ============================================================================
// Notification Response DTOs
// ============================================================================

// NotificationResponse represents a notification
type NotificationResponse struct {
	ID                uuid.UUID                    `json:"id"`
	TenantID          uuid.UUID                    `json:"tenant_id"`
	UserID            uuid.UUID                    `json:"user_id"`
	Type              models.NotificationType      `json:"type"`
	Title             string                       `json:"title"`
	Message           string                       `json:"message"`
	Channels          []models.NotificationChannel `json:"channels"`
	SentViaInApp      bool                         `json:"sent_via_in_app"`
	SentViaEmail      bool                         `json:"sent_via_email"`
	SentViaSMS        bool                         `json:"sent_via_sms"`
	SentViaPush       bool                         `json:"sent_via_push"`
	IsRead            bool                         `json:"is_read"`
	ReadAt            *time.Time                   `json:"read_at,omitempty"`
	ActionURL         string                       `json:"action_url,omitempty"`
	ActionText        string                       `json:"action_text,omitempty"`
	RelatedEntityType string                       `json:"related_entity_type,omitempty"`
	RelatedEntityID   *uuid.UUID                   `json:"related_entity_id,omitempty"`
	Priority          int                          `json:"priority"`
	ExpiresAt         *time.Time                   `json:"expires_at,omitempty"`
	Metadata          models.JSONB                 `json:"metadata,omitempty"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
}

// NotificationListResponse represents a paginated list of notifications
type NotificationListResponse struct {
	Notifications []*NotificationResponse `json:"notifications"`
	Page          int                     `json:"page"`
	PageSize      int                     `json:"page_size"`
	TotalItems    int64                   `json:"total_items"`
	TotalPages    int                     `json:"total_pages"`
	HasNext       bool                    `json:"has_next"`
	HasPrevious   bool                    `json:"has_previous"`
	UnreadCount   int64                   `json:"unread_count,omitempty"`
}

// NotificationStatsResponse represents notification statistics
type NotificationStatsResponse struct {
	UserID             uuid.UUID        `json:"user_id"`
	TotalNotifications int64            `json:"total_notifications"`
	UnreadCount        int64            `json:"unread_count"`
	ReadCount          int64            `json:"read_count"`
	ByType             map[string]int64 `json:"by_type"`
	ByPriority         map[int]int64    `json:"by_priority"`
	Recent7Days        int64            `json:"recent_7_days"`
	ExpiredCount       int64            `json:"expired_count"`
}

// NotificationPreferencesResponse represents user notification preferences
type NotificationPreferencesResponse struct {
	UserID               uuid.UUID                    `json:"user_id"`
	EnabledChannels      []models.NotificationChannel `json:"enabled_channels"`
	EnabledTypes         []models.NotificationType    `json:"enabled_types"`
	QuietHoursStart      *time.Time                   `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd        *time.Time                   `json:"quiet_hours_end,omitempty"`
	EmailDigestEnabled   bool                         `json:"email_digest_enabled"`
	EmailDigestFrequency string                       `json:"email_digest_frequency,omitempty"`
}

// NotificationDeliveryResponse represents delivery status
type NotificationDeliveryResponse struct {
	NotificationID uuid.UUID `json:"notification_id"`
	InAppSent      bool      `json:"in_app_sent"`
	EmailSent      bool      `json:"email_sent"`
	SMSSent        bool      `json:"sms_sent"`
	PushSent       bool      `json:"push_sent"`
	FailedChannels []string  `json:"failed_channels,omitempty"`
	Errors         []string  `json:"errors,omitempty"`
}

// BulkNotificationResponse represents bulk operation result
type BulkNotificationResponse struct {
	SuccessCount   int                             `json:"success_count"`
	FailureCount   int                             `json:"failure_count"`
	CreatedIDs     []uuid.UUID                     `json:"created_ids,omitempty"`
	Errors         []string                        `json:"errors,omitempty"`
	DeliveryStatus []*NotificationDeliveryResponse `json:"delivery_status,omitempty"`
}

// UnreadCountResponse represents unread notification count
type UnreadCountResponse struct {
	UserID      uuid.UUID        `json:"user_id"`
	UnreadCount int64            `json:"unread_count"`
	ByType      map[string]int64 `json:"by_type,omitempty"`
	ByPriority  map[int]int64    `json:"by_priority,omitempty"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToNotificationResponse converts a Notification model to NotificationResponse DTO
func ToNotificationResponse(notification *models.Notification) *NotificationResponse {
	if notification == nil {
		return nil
	}

	return &NotificationResponse{
		ID:                notification.ID,
		TenantID:          notification.TenantID,
		UserID:            notification.UserID,
		Type:              notification.Type,
		Title:             notification.Title,
		Message:           notification.Message,
		Channels:          notification.Channels,
		SentViaInApp:      notification.SentViaInApp,
		SentViaEmail:      notification.SentViaEmail,
		SentViaSMS:        notification.SentViaSMS,
		SentViaPush:       notification.SentViaPush,
		IsRead:            notification.IsRead,
		ReadAt:            notification.ReadAt,
		ActionURL:         notification.ActionURL,
		ActionText:        notification.ActionText,
		RelatedEntityType: notification.RelatedEntityType,
		RelatedEntityID:   notification.RelatedEntityID,
		Priority:          notification.Priority,
		ExpiresAt:         notification.ExpiresAt,
		Metadata:          notification.Metadata,
		CreatedAt:         notification.CreatedAt,
		UpdatedAt:         notification.UpdatedAt,
	}
}

// ToNotificationResponses converts multiple Notification models to DTOs
func ToNotificationResponses(notifications []*models.Notification) []*NotificationResponse {
	if notifications == nil {
		return nil
	}

	responses := make([]*NotificationResponse, len(notifications))
	for i, notification := range notifications {
		responses[i] = ToNotificationResponse(notification)
	}
	return responses
}
