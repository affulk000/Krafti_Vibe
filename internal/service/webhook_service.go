package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// webhookRepository defines extended webhook operations using the repository
type WebhookRepository interface {
	// Webhook Event Management
	CreateWebhookEvent(ctx context.Context, req *dto.CreateWebhookEventRequest) (*dto.WebhookEventResponse, error)
	GetWebhookEvent(ctx context.Context, eventID uuid.UUID) (*dto.WebhookEventResponse, error)
	ListWebhookEvents(ctx context.Context, filter dto.WebhookEventFilter) (*dto.WebhookEventListResponse, error)

	// Delivery Operations
	DeliverWebhook(ctx context.Context, eventID uuid.UUID) (*dto.WebhookDeliveryResponse, error)
	RetryWebhook(ctx context.Context, req *dto.RetryWebhookRequest) (*dto.WebhookDeliveryResponse, error)
	RetryFailedWebhooks(ctx context.Context, tenantID uuid.UUID, limit int) (*dto.WebhookRetryResponse, error)
	BulkRetryWebhooks(ctx context.Context, req *dto.BulkRetryRequest) (*dto.WebhookRetryResponse, error)

	// Query Operations
	GetPendingWebhooks(ctx context.Context, tenantID uuid.UUID, limit int) ([]*dto.WebhookEventResponse, error)
	GetFailedWebhooks(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.WebhookEventListResponse, error)
	GetDeliveredWebhooks(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.WebhookEventListResponse, error)
	GetRecentWebhooks(ctx context.Context, tenantID uuid.UUID, hours int, page, pageSize int) (*dto.WebhookEventListResponse, error)

	// Analytics
	GetWebhookStats(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*dto.WebhookStatsResponse, error)
	GetWebhookAnalytics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*dto.WebhookAnalyticsResponse, error)
	GetFailureReasons(ctx context.Context, tenantID uuid.UUID, limit int) ([]*dto.WebhookFailureReasonResponse, error)

	// Cleanup Operations
	CleanupOldWebhooks(ctx context.Context, olderThanDays int) (int64, error)
	CleanupDeliveredWebhooks(ctx context.Context, olderThanDays int) (int64, error)
	PurgeFailedWebhooks(ctx context.Context, maxAttempts int, olderThanDays int) (int64, error)

	// Event Triggering (from business events)
	TriggerBookingEvent(ctx context.Context, tenantID uuid.UUID, eventType models.WebhookEventType, booking any) error
	TriggerPaymentEvent(ctx context.Context, tenantID uuid.UUID, eventType models.WebhookEventType, payment any) error
	TriggerUserEvent(ctx context.Context, tenantID uuid.UUID, eventType models.WebhookEventType, user any) error

	// Background Processing
	ProcessPendingWebhooks(ctx context.Context, batchSize int) (*dto.WebhookRetryResponse, error)

	// Health & Monitoring
	HealthCheck(ctx context.Context) error
	GetServiceMetrics(ctx context.Context) map[string]any
}

// webhookRepository implements webhookRepository
type webhookRepository struct {
	repos      *repository.Repositories
	httpClient *http.Client
	logger     log.AllLogger
}

// NewwebhookRepository creates a new enhanced webhook service
func NewWebhookRepository(repos *repository.Repositories, logger log.AllLogger) WebhookRepository {
	return &webhookRepository{
		repos: repos,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// ============================================================================
// Webhook Event Management
// ============================================================================

// CreateWebhookEvent creates a new webhook event
func (s *webhookRepository) CreateWebhookEvent(ctx context.Context, req *dto.CreateWebhookEventRequest) (*dto.WebhookEventResponse, error) {
	if req.TenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant_id is required")
	}
	if req.WebhookURL == "" {
		return nil, errors.NewValidationError("webhook_url is required")
	}
	if req.Payload == nil {
		return nil, errors.NewValidationError("payload is required")
	}

	maxAttempts := req.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = 3 // Default
	}

	payloadJSON := models.JSONB{}
	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, errors.NewValidationError("invalid payload: " + err.Error())
	}
	if err := json.Unmarshal(payloadBytes, &payloadJSON); err != nil {
		return nil, errors.NewValidationError("failed to convert payload: " + err.Error())
	}

	metadataJSON := models.JSONB{}
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err == nil {
			json.Unmarshal(metadataBytes, &metadataJSON)
		}
	}

	event := &models.WebhookEvent{
		TenantID:     req.TenantID,
		EventType:    req.EventType,
		WebhookURL:   req.WebhookURL,
		Payload:      payloadJSON,
		MaxAttempts:  maxAttempts,
		AttemptCount: 0,
		Delivered:    false,
		Metadata:     metadataJSON,
	}

	if err := s.repos.WebhookEvent.Create(ctx, event); err != nil {
		return nil, errors.NewServiceError("WEBHOOK_CREATE_FAILED", "failed to create webhook event", err)
	}

	s.logger.Info("webhook event created",
		"event_id", event.ID,
		"tenant_id", req.TenantID,
		"event_type", req.EventType,
		"url", req.WebhookURL)

	return dto.ToWebhookEventResponse(event), nil
}

// GetWebhookEvent retrieves a webhook event by ID
func (s *webhookRepository) GetWebhookEvent(ctx context.Context, eventID uuid.UUID) (*dto.WebhookEventResponse, error) {
	event, err := s.repos.WebhookEvent.GetByID(ctx, eventID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("webhook event not found")
		}
		return nil, errors.NewServiceError("WEBHOOK_GET_FAILED", "failed to get webhook event", err)
	}

	return dto.ToWebhookEventResponse(event), nil
}

// ListWebhookEvents lists webhook events with filtering
func (s *webhookRepository) ListWebhookEvents(ctx context.Context, filter dto.WebhookEventFilter) (*dto.WebhookEventListResponse, error) {
	repoFilters := repository.WebhookEventFilters{
		TenantID:      filter.TenantID,
		EventTypes:    filter.EventTypes,
		Delivered:     filter.Delivered,
		WebhookURL:    filter.WebhookURL,
		MinAttempts:   filter.MinAttempts,
		MaxAttempts:   filter.MaxAttempts,
		CreatedFrom:   filter.CreatedFrom,
		CreatedTo:     filter.CreatedTo,
		ResponseCodes: filter.ResponseCodes,
	}

	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}
	pagination.Validate()

	events, paginationResult, err := s.repos.WebhookEvent.FindByFilters(ctx, repoFilters, pagination)
	if err != nil {
		return nil, errors.NewServiceError("WEBHOOK_LIST_FAILED", "failed to list webhook events", err)
	}

	return &dto.WebhookEventListResponse{
		Events:      dto.ToWebhookEventResponses(events),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ============================================================================
// Delivery Operations
// ============================================================================

// DeliverWebhook delivers a webhook event
func (s *webhookRepository) DeliverWebhook(ctx context.Context, eventID uuid.UUID) (*dto.WebhookDeliveryResponse, error) {
	event, err := s.repos.WebhookEvent.GetByID(ctx, eventID)
	if err != nil {
		return nil, errors.NewServiceError("WEBHOOK_GET_FAILED", "failed to get webhook event", err)
	}

	if event.Delivered {
		return &dto.WebhookDeliveryResponse{
			WebhookEventID: event.ID,
			Delivered:      true,
			AttemptCount:   event.AttemptCount,
			ResponseCode:   event.ResponseCode,
			DeliveredAt:    event.DeliveredAt,
		}, nil
	}

	// Increment attempt count
	if err := s.repos.WebhookEvent.IncrementAttemptCount(ctx, eventID); err != nil {
		s.logger.Error("failed to increment attempt count", "event_id", eventID, "error", err)
	}

	// Attempt delivery
	responseCode, responseBody, err := s.sendWebhook(ctx, event.WebhookURL, event.Payload)

	response := &dto.WebhookDeliveryResponse{
		WebhookEventID: event.ID,
		AttemptCount:   event.AttemptCount + 1,
		ResponseCode:   responseCode,
		ResponseBody:   responseBody,
	}

	if err != nil {
		// Delivery failed
		failureReason := err.Error()
		if err := s.repos.WebhookEvent.MarkAsFailed(ctx, eventID, responseCode, failureReason); err != nil {
			s.logger.Error("failed to mark webhook as failed", "event_id", eventID, "error", err)
		}

		response.Delivered = false
		response.FailureReason = failureReason

		// Calculate next retry time (exponential backoff)
		if event.AttemptCount+1 < event.MaxAttempts {
			nextRetry := s.calculateNextRetryTime(event.AttemptCount + 1)
			response.NextRetryAt = &nextRetry

			if err := s.repos.WebhookEvent.SetNextRetryTime(ctx, eventID, nextRetry); err != nil {
				s.logger.Error("failed to set next retry time", "event_id", eventID, "error", err)
			}
		}

		s.logger.Warn("webhook delivery failed",
			"event_id", eventID,
			"attempt", event.AttemptCount+1,
			"status_code", responseCode,
			"error", failureReason)

		return response, nil
	}

	// Delivery succeeded
	if err := s.repos.WebhookEvent.MarkAsDelivered(ctx, eventID, responseCode, responseBody); err != nil {
		s.logger.Error("failed to mark webhook as delivered", "event_id", eventID, "error", err)
	}

	now := time.Now()
	response.Delivered = true
	response.DeliveredAt = &now

	s.logger.Info("webhook delivered successfully",
		"event_id", eventID,
		"attempt", event.AttemptCount+1,
		"status_code", responseCode)

	return response, nil
}

// RetryWebhook retries a failed webhook
func (s *webhookRepository) RetryWebhook(ctx context.Context, req *dto.RetryWebhookRequest) (*dto.WebhookDeliveryResponse, error) {
	if req.ResetAttempts {
		if err := s.repos.WebhookEvent.ResetForRetry(ctx, req.WebhookEventID); err != nil {
			return nil, errors.NewServiceError("WEBHOOK_RESET_FAILED", "failed to reset webhook", err)
		}
	}

	return s.DeliverWebhook(ctx, req.WebhookEventID)
}

// RetryFailedWebhooks retries failed webhooks for a tenant
func (s *webhookRepository) RetryFailedWebhooks(ctx context.Context, tenantID uuid.UUID, limit int) (*dto.WebhookRetryResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	pagination := repository.PaginationParams{
		Page:     1,
		PageSize: limit,
	}

	failedEvents, _, err := s.repos.WebhookEvent.GetFailedWebhooks(ctx, tenantID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("WEBHOOK_RETRY_FAILED", "failed to get failed webhooks", err)
	}

	response := &dto.WebhookRetryResponse{
		RetriedCount: 0,
		SuccessCount: 0,
		FailureCount: 0,
		Errors:       []string{},
	}

	for _, event := range failedEvents {
		response.RetriedCount++

		deliveryResp, err := s.DeliverWebhook(ctx, event.ID)
		if err != nil {
			response.FailureCount++
			response.Errors = append(response.Errors, fmt.Sprintf("Event %s: %s", event.ID, err.Error()))
			continue
		}

		if deliveryResp.Delivered {
			response.SuccessCount++
		} else {
			response.FailureCount++
			if deliveryResp.FailureReason != "" {
				response.Errors = append(response.Errors, fmt.Sprintf("Event %s: %s", event.ID, deliveryResp.FailureReason))
			}
		}
	}

	s.logger.Info("failed webhooks retry completed",
		"tenant_id", tenantID,
		"retried", response.RetriedCount,
		"success", response.SuccessCount,
		"failed", response.FailureCount)

	return response, nil
}

// BulkRetryWebhooks retries webhooks in bulk based on criteria
func (s *webhookRepository) BulkRetryWebhooks(ctx context.Context, req *dto.BulkRetryRequest) (*dto.WebhookRetryResponse, error) {
	filters := repository.WebhookEventFilters{
		TenantID:  req.TenantID,
		Delivered: func() *bool { b := false; return &b }(),
	}

	if req.EventType != nil {
		filters.EventTypes = []models.WebhookEventType{*req.EventType}
	}

	if req.OlderThanHours > 0 {
		cutoff := time.Now().Add(-time.Duration(req.OlderThanHours) * time.Hour)
		filters.CreatedTo = &cutoff
	}

	pagination := repository.PaginationParams{
		Page:     1,
		PageSize: 100, // Process in batches
	}

	events, _, err := s.repos.WebhookEvent.FindByFilters(ctx, filters, pagination)
	if err != nil {
		return nil, errors.NewServiceError("WEBHOOK_BULK_RETRY_FAILED", "failed to find webhooks", err)
	}

	response := &dto.WebhookRetryResponse{
		RetriedCount: 0,
		SuccessCount: 0,
		FailureCount: 0,
		Errors:       []string{},
	}

	for _, event := range events {
		if !event.ShouldRetry() {
			continue
		}

		response.RetriedCount++

		deliveryResp, err := s.DeliverWebhook(ctx, event.ID)
		if err != nil {
			response.FailureCount++
			response.Errors = append(response.Errors, fmt.Sprintf("Event %s: %s", event.ID, err.Error()))
			continue
		}

		if deliveryResp.Delivered {
			response.SuccessCount++
		} else {
			response.FailureCount++
		}
	}

	s.logger.Info("bulk webhook retry completed",
		"tenant_id", req.TenantID,
		"retried", response.RetriedCount,
		"success", response.SuccessCount,
		"failed", response.FailureCount)

	return response, nil
}

// ============================================================================
// Query Operations
// ============================================================================

// GetPendingWebhooks retrieves pending webhooks for retry
func (s *webhookRepository) GetPendingWebhooks(ctx context.Context, tenantID uuid.UUID, limit int) ([]*dto.WebhookEventResponse, error) {
	if limit <= 0 {
		limit = 50
	}

	events, err := s.repos.WebhookEvent.GetPendingRetries(ctx, limit)
	if err != nil {
		return nil, errors.NewServiceError("WEBHOOK_PENDING_FAILED", "failed to get pending webhooks", err)
	}

	// Filter by tenant if specified
	if tenantID != uuid.Nil {
		filtered := []*models.WebhookEvent{}
		for _, event := range events {
			if event.TenantID == tenantID {
				filtered = append(filtered, event)
			}
		}
		events = filtered
	}

	return dto.ToWebhookEventResponses(events), nil
}

// GetFailedWebhooks retrieves failed webhooks
func (s *webhookRepository) GetFailedWebhooks(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.WebhookEventListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
	pagination.Validate()

	events, paginationResult, err := s.repos.WebhookEvent.GetFailedWebhooks(ctx, tenantID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("WEBHOOK_FAILED_LIST_FAILED", "failed to get failed webhooks", err)
	}

	return &dto.WebhookEventListResponse{
		Events:      dto.ToWebhookEventResponses(events),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetDeliveredWebhooks retrieves delivered webhooks
func (s *webhookRepository) GetDeliveredWebhooks(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.WebhookEventListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
	pagination.Validate()

	events, paginationResult, err := s.repos.WebhookEvent.GetDeliveredWebhooks(ctx, tenantID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("WEBHOOK_DELIVERED_LIST_FAILED", "failed to get delivered webhooks", err)
	}

	return &dto.WebhookEventListResponse{
		Events:      dto.ToWebhookEventResponses(events),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetRecentWebhooks retrieves recent webhooks
func (s *webhookRepository) GetRecentWebhooks(ctx context.Context, tenantID uuid.UUID, hours int, page, pageSize int) (*dto.WebhookEventListResponse, error) {
	if hours <= 0 {
		hours = 24
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
	pagination.Validate()

	events, paginationResult, err := s.repos.WebhookEvent.GetRecentWebhooks(ctx, tenantID, hours, pagination)
	if err != nil {
		return nil, errors.NewServiceError("WEBHOOK_RECENT_LIST_FAILED", "failed to get recent webhooks", err)
	}

	return &dto.WebhookEventListResponse{
		Events:      dto.ToWebhookEventResponses(events),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ============================================================================
// Analytics
// ============================================================================

// GetWebhookStats retrieves webhook statistics
func (s *webhookRepository) GetWebhookStats(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*dto.WebhookStatsResponse, error) {
	stats, err := s.repos.WebhookEvent.GetWebhookStats(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, errors.NewServiceError("WEBHOOK_STATS_FAILED", "failed to get webhook stats", err)
	}

	return &dto.WebhookStatsResponse{
		TenantID:          tenantID,
		TotalWebhooks:     stats.TotalWebhooks,
		DeliveredWebhooks: stats.DeliveredWebhooks,
		FailedWebhooks:    stats.FailedWebhooks,
		PendingWebhooks:   stats.PendingWebhooks,
		DeliveryRate:      stats.DeliveryRate,
		AverageAttempts:   stats.AverageAttempts,
		ByEventType:       stats.ByEventType,
		ByStatus:          stats.ByStatus,
	}, nil
}

// GetWebhookAnalytics retrieves comprehensive webhook analytics
func (s *webhookRepository) GetWebhookAnalytics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*dto.WebhookAnalyticsResponse, error) {
	stats, err := s.repos.WebhookEvent.GetWebhookStats(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, errors.NewServiceError("WEBHOOK_ANALYTICS_FAILED", "failed to get webhook analytics", err)
	}

	avgDeliveryTime, err := s.repos.WebhookEvent.GetAverageDeliveryTime(ctx, tenantID, startDate, endDate)
	if err != nil {
		s.logger.Warn("failed to get average delivery time", "error", err)
	}

	failureReasons, err := s.repos.WebhookEvent.GetFailureReasons(ctx, tenantID, 10)
	if err != nil {
		s.logger.Warn("failed to get failure reasons", "error", err)
		failureReasons = []repository.FailureReasonCount{}
	}

	topFailures := make([]*dto.WebhookFailureReasonResponse, len(failureReasons))
	for i, fr := range failureReasons {
		topFailures[i] = &dto.WebhookFailureReasonResponse{
			Reason: fr.Reason,
			Count:  fr.Count,
		}
	}

	period := "custom"
	if startDate.IsZero() && endDate.IsZero() {
		period = "all_time"
	} else if !startDate.IsZero() && !endDate.IsZero() {
		period = fmt.Sprintf("%dd", int(endDate.Sub(startDate).Hours()/24))
	}

	return &dto.WebhookAnalyticsResponse{
		TenantID:             tenantID,
		Period:               period,
		StartDate:            startDate,
		EndDate:              endDate,
		TotalEvents:          stats.TotalWebhooks,
		SuccessfulDeliveries: stats.DeliveredWebhooks,
		FailedDeliveries:     stats.FailedWebhooks,
		AverageDeliveryTime:  avgDeliveryTime.String(),
		DeliveryRate:         stats.DeliveryRate,
		EventsByType:         stats.ByEventType,
		TopFailureReasons:    topFailures,
	}, nil
}

// GetFailureReasons retrieves top failure reasons
func (s *webhookRepository) GetFailureReasons(ctx context.Context, tenantID uuid.UUID, limit int) ([]*dto.WebhookFailureReasonResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	failureReasons, err := s.repos.WebhookEvent.GetFailureReasons(ctx, tenantID, limit)
	if err != nil {
		return nil, errors.NewServiceError("WEBHOOK_FAILURE_REASONS_FAILED", "failed to get failure reasons", err)
	}

	responses := make([]*dto.WebhookFailureReasonResponse, len(failureReasons))
	for i, fr := range failureReasons {
		responses[i] = &dto.WebhookFailureReasonResponse{
			Reason: fr.Reason,
			Count:  fr.Count,
		}
	}

	return responses, nil
}

// ============================================================================
// Cleanup Operations
// ============================================================================

// CleanupOldWebhooks removes old webhook events
func (s *webhookRepository) CleanupOldWebhooks(ctx context.Context, olderThanDays int) (int64, error) {
	if olderThanDays <= 0 {
		olderThanDays = 90 // Default to 90 days
	}

	cutoff := time.Now().AddDate(0, 0, -olderThanDays)

	count, err := s.repos.WebhookEvent.DeleteOldWebhooks(ctx, cutoff)
	if err != nil {
		return 0, errors.NewServiceError("WEBHOOK_CLEANUP_FAILED", "failed to cleanup old webhooks", err)
	}

	s.logger.Info("cleaned up old webhooks",
		"count", count,
		"older_than_days", olderThanDays)

	return count, nil
}

// CleanupDeliveredWebhooks removes old delivered webhooks
func (s *webhookRepository) CleanupDeliveredWebhooks(ctx context.Context, olderThanDays int) (int64, error) {
	if olderThanDays <= 0 {
		olderThanDays = 30 // Default to 30 days
	}

	cutoff := time.Now().AddDate(0, 0, -olderThanDays)

	count, err := s.repos.WebhookEvent.DeleteDeliveredWebhooks(ctx, cutoff)
	if err != nil {
		return 0, errors.NewServiceError("WEBHOOK_CLEANUP_FAILED", "failed to cleanup delivered webhooks", err)
	}

	s.logger.Info("cleaned up delivered webhooks",
		"count", count,
		"older_than_days", olderThanDays)

	return count, nil
}

// PurgeFailedWebhooks removes permanently failed webhooks
func (s *webhookRepository) PurgeFailedWebhooks(ctx context.Context, maxAttempts int, olderThanDays int) (int64, error) {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	if olderThanDays <= 0 {
		olderThanDays = 7
	}

	cutoff := time.Now().AddDate(0, 0, -olderThanDays)

	count, err := s.repos.WebhookEvent.PurgeFailedWebhooks(ctx, maxAttempts, cutoff)
	if err != nil {
		return 0, errors.NewServiceError("WEBHOOK_PURGE_FAILED", "failed to purge failed webhooks", err)
	}

	s.logger.Info("purged failed webhooks",
		"count", count,
		"max_attempts", maxAttempts,
		"older_than_days", olderThanDays)

	return count, nil
}

// ============================================================================
// Event Triggering
// ============================================================================

// TriggerBookingEvent triggers a webhook event for booking changes
func (s *webhookRepository) TriggerBookingEvent(ctx context.Context, tenantID uuid.UUID, eventType models.WebhookEventType, booking any) error {
	// Get webhook URLs configured for this tenant (would need webhook configuration management)
	// For now, this is a placeholder that would integrate with webhook configuration
	s.logger.Info("booking event triggered",
		"tenant_id", tenantID,
		"event_type", eventType)

	return nil
}

// TriggerPaymentEvent triggers a webhook event for payment changes
func (s *webhookRepository) TriggerPaymentEvent(ctx context.Context, tenantID uuid.UUID, eventType models.WebhookEventType, payment any) error {
	s.logger.Info("payment event triggered",
		"tenant_id", tenantID,
		"event_type", eventType)

	return nil
}

// TriggerUserEvent triggers a webhook event for user changes
func (s *webhookRepository) TriggerUserEvent(ctx context.Context, tenantID uuid.UUID, eventType models.WebhookEventType, user any) error {
	s.logger.Info("user event triggered",
		"tenant_id", tenantID,
		"event_type", eventType)

	return nil
}

// ============================================================================
// Background Processing
// ============================================================================

// ProcessPendingWebhooks processes pending webhooks in batch
func (s *webhookRepository) ProcessPendingWebhooks(ctx context.Context, batchSize int) (*dto.WebhookRetryResponse, error) {
	if batchSize <= 0 {
		batchSize = 100
	}

	pendingEvents, err := s.repos.WebhookEvent.GetPendingRetries(ctx, batchSize)
	if err != nil {
		return nil, errors.NewServiceError("WEBHOOK_PROCESS_FAILED", "failed to get pending webhooks", err)
	}

	response := &dto.WebhookRetryResponse{
		RetriedCount: 0,
		SuccessCount: 0,
		FailureCount: 0,
		Errors:       []string{},
	}

	for _, event := range pendingEvents {
		if !event.CanRetryNow() {
			continue
		}

		response.RetriedCount++

		deliveryResp, err := s.DeliverWebhook(ctx, event.ID)
		if err != nil {
			response.FailureCount++
			response.Errors = append(response.Errors, fmt.Sprintf("Event %s: %s", event.ID, err.Error()))
			continue
		}

		if deliveryResp.Delivered {
			response.SuccessCount++
		} else {
			response.FailureCount++
		}
	}

	s.logger.Info("pending webhooks processed",
		"batch_size", batchSize,
		"processed", response.RetriedCount,
		"success", response.SuccessCount,
		"failed", response.FailureCount)

	return response, nil
}

// ============================================================================
// Health & Monitoring
// ============================================================================

// HealthCheck performs a health check
func (s *webhookRepository) HealthCheck(ctx context.Context) error {
	// Try to query the database
	pagination := repository.PaginationParams{Page: 1, PageSize: 1}
	_, _, err := s.repos.WebhookEvent.GetByTenantID(ctx, uuid.New(), pagination)

	if err != nil && !errors.IsNotFoundError(err) {
		return errors.NewServiceError("HEALTH_CHECK_FAILED", "webhook service health check failed", err)
	}

	return nil
}

// GetServiceMetrics retrieves service metrics
func (s *webhookRepository) GetServiceMetrics(ctx context.Context) map[string]any {
	metrics := make(map[string]any)

	// Get overall statistics
	stats, err := s.repos.WebhookEvent.GetWebhookStats(ctx, uuid.Nil, time.Time{}, time.Time{})
	if err == nil {
		metrics["total_webhooks"] = stats.TotalWebhooks
		metrics["delivered_webhooks"] = stats.DeliveredWebhooks
		metrics["failed_webhooks"] = stats.FailedWebhooks
		metrics["pending_webhooks"] = stats.PendingWebhooks
		metrics["delivery_rate"] = stats.DeliveryRate
		metrics["average_attempts"] = stats.AverageAttempts
	}

	return metrics
}

// ============================================================================
// Helper Methods
// ============================================================================

// sendWebhook sends the webhook HTTP request
func (s *webhookRepository) sendWebhook(ctx context.Context, url string, payload models.JSONB) (int, string, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return 0, "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Krafti_Vibe-Webhook/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Read response body (limit to 1MB)
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		s.logger.Warn("failed to read response body", "error", err)
		bodyBytes = []byte{}
	}

	// Check if status code indicates success (2xx)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, string(bodyBytes), fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}

	return resp.StatusCode, string(bodyBytes), nil
}

// calculateNextRetryTime calculates the next retry time with exponential backoff
func (s *webhookRepository) calculateNextRetryTime(attemptCount int) time.Time {
	// Exponential backoff: 1min, 5min, 15min, 30min, 1hr, 2hr, 4hr
	backoffMinutes := []int{1, 5, 15, 30, 60, 120, 240}

	index := attemptCount - 1
	if index >= len(backoffMinutes) {
		index = len(backoffMinutes) - 1
	}

	return time.Now().Add(time.Duration(backoffMinutes[index]) * time.Minute)
}
