package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Webhook Event Request DTOs
// ============================================================================

// CreateWebhookEventRequest represents a request to create a webhook event
type CreateWebhookEventRequest struct {
	TenantID    uuid.UUID               `json:"tenant_id" validate:"required"`
	EventType   models.WebhookEventType `json:"event_type" validate:"required"`
	WebhookURL  string                  `json:"webhook_url" validate:"required,url"`
	Payload     map[string]any          `json:"payload" validate:"required"`
	MaxAttempts int                     `json:"max_attempts" validate:"min=1,max=10"`
	Metadata    map[string]any          `json:"metadata,omitempty"`
}

// WebhookEventFilter represents filters for webhook events
type WebhookEventFilter struct {
	TenantID      uuid.UUID                 `json:"tenant_id"`
	EventTypes    []models.WebhookEventType `json:"event_types"`
	Delivered     *bool                     `json:"delivered"`
	WebhookURL    string                    `json:"webhook_url"`
	MinAttempts   *int                      `json:"min_attempts"`
	MaxAttempts   *int                      `json:"max_attempts"`
	CreatedFrom   *time.Time                `json:"created_from"`
	CreatedTo     *time.Time                `json:"created_to"`
	ResponseCodes []int                     `json:"response_codes"`
	Page          int                       `json:"page"`
	PageSize      int                       `json:"page_size"`
}

// RetryWebhookRequest represents a request to retry a failed webhook
type RetryWebhookRequest struct {
	WebhookEventID uuid.UUID `json:"webhook_event_id" validate:"required"`
	ResetAttempts  bool      `json:"reset_attempts"`
}

// BulkRetryRequest represents a bulk retry request
type BulkRetryRequest struct {
	TenantID       uuid.UUID                `json:"tenant_id" validate:"required"`
	EventType      *models.WebhookEventType `json:"event_type,omitempty"`
	OlderThanHours int                      `json:"older_than_hours" validate:"min=1,max=168"`
}

// WebhookDeliveryRequest represents a webhook delivery request
type WebhookDeliveryRequest struct {
	WebhookEventID uuid.UUID `json:"webhook_event_id" validate:"required"`
	ForceDelivery  bool      `json:"force_delivery"`
}

// ============================================================================
// Webhook Event Response DTOs
// ============================================================================

// WebhookEventResponse represents a webhook event
type WebhookEventResponse struct {
	ID              uuid.UUID               `json:"id"`
	TenantID        uuid.UUID               `json:"tenant_id"`
	EventType       models.WebhookEventType `json:"event_type"`
	Payload         models.JSONB            `json:"payload"`
	WebhookURL      string                  `json:"webhook_url"`
	AttemptCount    int                     `json:"attempt_count"`
	MaxAttempts     int                     `json:"max_attempts"`
	Delivered       bool                    `json:"delivered"`
	DeliveredAt     *time.Time              `json:"delivered_at,omitempty"`
	NextRetryAt     *time.Time              `json:"next_retry_at,omitempty"`
	LastAttemptedAt *time.Time              `json:"last_attempted_at,omitempty"`
	ResponseCode    int                     `json:"response_code,omitempty"`
	ResponseBody    string                  `json:"response_body,omitempty"`
	FailureReason   string                  `json:"failure_reason,omitempty"`
	Metadata        models.JSONB            `json:"metadata,omitempty"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

// WebhookEventListResponse represents a paginated list of webhook events
type WebhookEventListResponse struct {
	Events      []*WebhookEventResponse `json:"events"`
	Page        int                     `json:"page"`
	PageSize    int                     `json:"page_size"`
	TotalItems  int64                   `json:"total_items"`
	TotalPages  int                     `json:"total_pages"`
	HasNext     bool                    `json:"has_next"`
	HasPrevious bool                    `json:"has_previous"`
}

// WebhookStatsResponse represents webhook delivery statistics
type WebhookStatsResponse struct {
	TenantID          uuid.UUID                         `json:"tenant_id"`
	TotalWebhooks     int64                             `json:"total_webhooks"`
	DeliveredWebhooks int64                             `json:"delivered_webhooks"`
	FailedWebhooks    int64                             `json:"failed_webhooks"`
	PendingWebhooks   int64                             `json:"pending_webhooks"`
	DeliveryRate      float64                           `json:"delivery_rate"`
	AverageAttempts   float64                           `json:"average_attempts"`
	ByEventType       map[models.WebhookEventType]int64 `json:"by_event_type"`
	ByStatus          map[string]int64                  `json:"by_status"`
}

// WebhookDeliveryResponse represents delivery attempt result
type WebhookDeliveryResponse struct {
	WebhookEventID uuid.UUID  `json:"webhook_event_id"`
	Delivered      bool       `json:"delivered"`
	AttemptCount   int        `json:"attempt_count"`
	ResponseCode   int        `json:"response_code"`
	ResponseBody   string     `json:"response_body,omitempty"`
	FailureReason  string     `json:"failure_reason,omitempty"`
	NextRetryAt    *time.Time `json:"next_retry_at,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
}

// WebhookFailureReasonResponse represents failure reason statistics
type WebhookFailureReasonResponse struct {
	Reason string `json:"reason"`
	Count  int64  `json:"count"`
}

// WebhookAnalyticsResponse represents webhook analytics
type WebhookAnalyticsResponse struct {
	TenantID             uuid.UUID                         `json:"tenant_id"`
	Period               string                            `json:"period"`
	StartDate            time.Time                         `json:"start_date"`
	EndDate              time.Time                         `json:"end_date"`
	TotalEvents          int64                             `json:"total_events"`
	SuccessfulDeliveries int64                             `json:"successful_deliveries"`
	FailedDeliveries     int64                             `json:"failed_deliveries"`
	AverageDeliveryTime  string                            `json:"average_delivery_time"`
	DeliveryRate         float64                           `json:"delivery_rate"`
	EventsByType         map[models.WebhookEventType]int64 `json:"events_by_type"`
	TopFailureReasons    []*WebhookFailureReasonResponse   `json:"top_failure_reasons"`
}

// WebhookRetryResponse represents retry operation result
type WebhookRetryResponse struct {
	RetriedCount int      `json:"retried_count"`
	SuccessCount int      `json:"success_count"`
	FailureCount int      `json:"failure_count"`
	Errors       []string `json:"errors,omitempty"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToWebhookEventResponse converts a WebhookEvent model to WebhookEventResponse DTO
func ToWebhookEventResponse(event *models.WebhookEvent) *WebhookEventResponse {
	if event == nil {
		return nil
	}

	return &WebhookEventResponse{
		ID:              event.ID,
		TenantID:        event.TenantID,
		EventType:       event.EventType,
		Payload:         event.Payload,
		WebhookURL:      event.WebhookURL,
		AttemptCount:    event.AttemptCount,
		MaxAttempts:     event.MaxAttempts,
		Delivered:       event.Delivered,
		DeliveredAt:     event.DeliveredAt,
		NextRetryAt:     event.NextRetryAt,
		LastAttemptedAt: event.LastAttemptedAt,
		ResponseCode:    event.ResponseCode,
		ResponseBody:    event.ResponseBody,
		FailureReason:   event.FailureReason,
		Metadata:        event.Metadata,
		CreatedAt:       event.CreatedAt,
		UpdatedAt:       event.UpdatedAt,
	}
}

// ToWebhookEventResponses converts multiple WebhookEvent models to DTOs
func ToWebhookEventResponses(events []*models.WebhookEvent) []*WebhookEventResponse {
	if events == nil {
		return nil
	}

	responses := make([]*WebhookEventResponse, len(events))
	for i, event := range events {
		responses[i] = ToWebhookEventResponse(event)
	}
	return responses
}
