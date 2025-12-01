package repository

import (
	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WebhookEventRepository interface {
	BaseRepository[models.WebhookEvent]

	// Core Operations
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.WebhookEvent, PaginationResult, error)
	GetByEventType(ctx context.Context, tenantID uuid.UUID, eventType models.WebhookEventType, pagination PaginationParams) ([]*models.WebhookEvent, PaginationResult, error)

	// Delivery Operations
	MarkAsDelivered(ctx context.Context, webhookID uuid.UUID, responseCode int, responseBody string) error
	MarkAsFailed(ctx context.Context, webhookID uuid.UUID, responseCode int, failureReason string) error
	IncrementAttemptCount(ctx context.Context, webhookID uuid.UUID) error
	SetNextRetryTime(ctx context.Context, webhookID uuid.UUID, nextRetryAt time.Time) error
	UpdateDeliveryStatus(ctx context.Context, webhookID uuid.UUID, delivered bool, responseCode int, responseBody, failureReason string) error

	// Retry Operations
	GetPendingRetries(ctx context.Context, limit int) ([]*models.WebhookEvent, error)
	GetFailedWebhooks(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.WebhookEvent, PaginationResult, error)
	GetWebhooksForRetry(ctx context.Context, limit int) ([]*models.WebhookEvent, error)
	ResetForRetry(ctx context.Context, webhookID uuid.UUID) error

	// Query Operations
	GetDeliveredWebhooks(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.WebhookEvent, PaginationResult, error)
	GetRecentWebhooks(ctx context.Context, tenantID uuid.UUID, hours int, pagination PaginationParams) ([]*models.WebhookEvent, PaginationResult, error)
	GetWebhooksByURL(ctx context.Context, webhookURL string, pagination PaginationParams) ([]*models.WebhookEvent, PaginationResult, error)

	// Analytics
	GetWebhookStats(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (WebhookStats, error)
	GetDeliveryRate(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error)
	GetAverageDeliveryTime(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (time.Duration, error)
	GetWebhooksByEventType(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[models.WebhookEventType]int64, error)
	GetFailureReasons(ctx context.Context, tenantID uuid.UUID, limit int) ([]FailureReasonCount, error)

	// Cleanup Operations
	DeleteOldWebhooks(ctx context.Context, olderThan time.Time) (int64, error)
	DeleteDeliveredWebhooks(ctx context.Context, olderThan time.Time) (int64, error)
	PurgeFailedWebhooks(ctx context.Context, maxAttempts int, olderThan time.Time) (int64, error)

	// Search & Filter
	FindByFilters(ctx context.Context, filters WebhookEventFilters, pagination PaginationParams) ([]*models.WebhookEvent, PaginationResult, error)
}

// WebhookStats represents webhook delivery statistics
type WebhookStats struct {
	TotalWebhooks     int64                             `json:"total_webhooks"`
	DeliveredWebhooks int64                             `json:"delivered_webhooks"`
	FailedWebhooks    int64                             `json:"failed_webhooks"`
	PendingWebhooks   int64                             `json:"pending_webhooks"`
	DeliveryRate      float64                           `json:"delivery_rate"`
	AverageAttempts   float64                           `json:"average_attempts"`
	ByEventType       map[models.WebhookEventType]int64 `json:"by_event_type"`
	ByStatus          map[string]int64                  `json:"by_status"`
}

// FailureReasonCount represents failure reason statistics
type FailureReasonCount struct {
	Reason string `json:"reason"`
	Count  int64  `json:"count"`
}

// WebhookEventFilters for advanced filtering
type WebhookEventFilters struct {
	TenantID      uuid.UUID                 `json:"tenant_id"`
	EventTypes    []models.WebhookEventType `json:"event_types"`
	Delivered     *bool                     `json:"delivered"`
	WebhookURL    string                    `json:"webhook_url"`
	MinAttempts   *int                      `json:"min_attempts"`
	MaxAttempts   *int                      `json:"max_attempts"`
	CreatedFrom   *time.Time                `json:"created_from"`
	CreatedTo     *time.Time                `json:"created_to"`
	ResponseCodes []int                     `json:"response_codes"`
}

type webhookEventRepository struct {
	BaseRepository[models.WebhookEvent]
	db      *gorm.DB
	logger  log.AllLogger
	cache   Cache
	metrics MetricsCollector
}

func NewWebhookEventRepository(db *gorm.DB, config ...RepositoryConfig) WebhookEventRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 1 * time.Minute // Short cache for webhooks
	}

	baseRepo := NewBaseRepository[models.WebhookEvent](db, cfg)

	return &webhookEventRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
		metrics:        cfg.Metrics,
	}
}

//------------------------------------------------------------
// Core Operations
//------------------------------------------------------------

func (r *webhookEventRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.WebhookEvent, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.WebhookEvent{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count webhook events", err)
	}

	var webhooks []*models.WebhookEvent
	if err := query.
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&webhooks).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find webhook events", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return webhooks, paginationResult, nil
}

func (r *webhookEventRepository) GetByEventType(ctx context.Context, tenantID uuid.UUID, eventType models.WebhookEventType, pagination PaginationParams) ([]*models.WebhookEvent, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.WebhookEvent{}).
		Where("tenant_id = ? AND event_type = ?", tenantID, eventType)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count webhook events", err)
	}

	var webhooks []*models.WebhookEvent
	if err := query.
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&webhooks).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find webhook events", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return webhooks, paginationResult, nil
}

//------------------------------------------------------------
// Delivery Operations
//------------------------------------------------------------

func (r *webhookEventRepository) MarkAsDelivered(ctx context.Context, webhookID uuid.UUID, responseCode int, responseBody string) error {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&models.WebhookEvent{}).
		Where("id = ?", webhookID).
		Updates(map[string]any{
			"delivered":         true,
			"delivered_at":      &now,
			"last_attempted_at": &now,
			"response_code":     responseCode,
			"response_body":     responseBody,
			"failure_reason":    "",
		})

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark webhook as delivered", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "webhook event not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, webhookID)
	return nil
}

func (r *webhookEventRepository) MarkAsFailed(ctx context.Context, webhookID uuid.UUID, responseCode int, failureReason string) error {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&models.WebhookEvent{}).
		Where("id = ?", webhookID).
		Updates(map[string]any{
			"last_attempted_at": &now,
			"response_code":     responseCode,
			"failure_reason":    failureReason,
		})

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark webhook as failed", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "webhook event not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, webhookID)
	return nil
}

func (r *webhookEventRepository) IncrementAttemptCount(ctx context.Context, webhookID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.WebhookEvent{}).
		Where("id = ?", webhookID).
		UpdateColumn("attempt_count", gorm.Expr("attempt_count + ?", 1))

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to increment attempt count", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "webhook event not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, webhookID)
	return nil
}

func (r *webhookEventRepository) SetNextRetryTime(ctx context.Context, webhookID uuid.UUID, nextRetryAt time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&models.WebhookEvent{}).
		Where("id = ?", webhookID).
		Update("next_retry_at", nextRetryAt)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to set next retry time", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "webhook event not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, webhookID)
	return nil
}

func (r *webhookEventRepository) UpdateDeliveryStatus(ctx context.Context, webhookID uuid.UUID, delivered bool, responseCode int, responseBody, failureReason string) error {
	now := time.Now()

	updates := map[string]any{
		"delivered":         delivered,
		"last_attempted_at": &now,
		"response_code":     responseCode,
		"response_body":     responseBody,
		"failure_reason":    failureReason,
	}

	if delivered {
		updates["delivered_at"] = &now
	}

	result := r.db.WithContext(ctx).
		Model(&models.WebhookEvent{}).
		Where("id = ?", webhookID).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update delivery status", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "webhook event not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, webhookID)
	return nil
}

//------------------------------------------------------------
// Retry Operations
//------------------------------------------------------------

func (r *webhookEventRepository) GetPendingRetries(ctx context.Context, limit int) ([]*models.WebhookEvent, error) {
	now := time.Now()

	var webhooks []*models.WebhookEvent
	if err := r.db.WithContext(ctx).
		Where("delivered = ? AND attempt_count < max_attempts AND (next_retry_at IS NULL OR next_retry_at <= ?)",
			false, now).
		Order("created_at ASC").
		Limit(limit).
		Find(&webhooks).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find pending retries", err)
	}

	return webhooks, nil
}

func (r *webhookEventRepository) GetFailedWebhooks(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.WebhookEvent, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.WebhookEvent{}).
		Where("tenant_id = ? AND delivered = ? AND attempt_count >= max_attempts", tenantID, false)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count failed webhooks", err)
	}

	var webhooks []*models.WebhookEvent
	if err := query.
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&webhooks).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find failed webhooks", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return webhooks, paginationResult, nil
}

func (r *webhookEventRepository) GetWebhooksForRetry(ctx context.Context, limit int) ([]*models.WebhookEvent, error) {
	return r.GetPendingRetries(ctx, limit)
}

func (r *webhookEventRepository) ResetForRetry(ctx context.Context, webhookID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.WebhookEvent{}).
		Where("id = ?", webhookID).
		Updates(map[string]any{
			"attempt_count":  0,
			"delivered":      false,
			"delivered_at":   nil,
			"failure_reason": "",
			"next_retry_at":  nil,
		})

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to reset webhook for retry", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "webhook event not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, webhookID)
	return nil
}

//------------------------------------------------------------
// Query Operations
//------------------------------------------------------------

func (r *webhookEventRepository) GetDeliveredWebhooks(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.WebhookEvent, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.WebhookEvent{}).
		Where("tenant_id = ? AND delivered = ?", tenantID, true)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count delivered webhooks", err)
	}

	var webhooks []*models.WebhookEvent
	if err := query.
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("delivered_at DESC").
		Find(&webhooks).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find delivered webhooks", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return webhooks, paginationResult, nil
}

func (r *webhookEventRepository) GetRecentWebhooks(ctx context.Context, tenantID uuid.UUID, hours int, pagination PaginationParams) ([]*models.WebhookEvent, PaginationResult, error) {
	pagination.Validate()

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.WebhookEvent{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, since)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count recent webhooks", err)
	}

	var webhooks []*models.WebhookEvent
	if err := query.
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&webhooks).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find recent webhooks", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return webhooks, paginationResult, nil
}

func (r *webhookEventRepository) GetWebhooksByURL(ctx context.Context, webhookURL string, pagination PaginationParams) ([]*models.WebhookEvent, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.WebhookEvent{}).Where("webhook_url = ?", webhookURL)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count webhooks", err)
	}

	var webhooks []*models.WebhookEvent
	if err := query.
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&webhooks).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find webhooks", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return webhooks, paginationResult, nil
}

//------------------------------------------------------------
// Analytics
//------------------------------------------------------------

func (r *webhookEventRepository) GetWebhookStats(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (WebhookStats, error) {
	stats := WebhookStats{
		ByEventType: make(map[models.WebhookEventType]int64),
		ByStatus:    make(map[string]int64),
	}

	query := r.db.WithContext(ctx).Model(&models.WebhookEvent{}).Where("tenant_id = ?", tenantID)
	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}

	if err := query.Count(&stats.TotalWebhooks).Error; err != nil {
		return stats, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to count total webhooks", err)
	}

	if err := query.Where("delivered = ?", true).Count(&stats.DeliveredWebhooks).Error; err != nil {
		return stats, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to count delivered webhooks", err)
	}

	if err := query.Where("delivered = ? AND attempt_count >= max_attempts", false).
		Count(&stats.FailedWebhooks).Error; err != nil {
		return stats, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to count failed webhooks", err)
	}

	if err := query.Where("delivered = ? AND attempt_count < max_attempts", false).
		Count(&stats.PendingWebhooks).Error; err != nil {
		return stats, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to count pending webhooks", err)
	}

	if stats.TotalWebhooks > 0 {
		stats.DeliveryRate = float64(stats.DeliveredWebhooks) / float64(stats.TotalWebhooks) * 100
	}

	var avgAttempts float64
	if err := query.Select("COALESCE(AVG(attempt_count), 0)").
		Scan(&avgAttempts).Error; err == nil {
		stats.AverageAttempts = avgAttempts
	}

	eventTypes := []models.WebhookEventType{
		models.WebhookEventBookingCreated,
		models.WebhookEventBookingUpdated,
		models.WebhookEventBookingCancelled,
		models.WebhookEventPaymentReceived,
		models.WebhookEventReviewCreated,
		models.WebhookEventUserCreated,
	}

	for _, eventType := range eventTypes {
		var count int64
		if err := query.Where("event_type = ?", eventType).Count(&count).Error; err == nil {
			stats.ByEventType[eventType] = count
		}
	}

	stats.ByStatus["delivered"] = stats.DeliveredWebhooks
	stats.ByStatus["failed"] = stats.FailedWebhooks
	stats.ByStatus["pending"] = stats.PendingWebhooks

	return stats, nil
}

func (r *webhookEventRepository) GetDeliveryRate(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	query := r.db.WithContext(ctx).Model(&models.WebhookEvent{}).Where("tenant_id = ?", tenantID)
	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}

	var total, delivered int64
	if err := query.Count(&total).Error; err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to count total webhooks", err)
	}

	if total == 0 {
		return 0, nil
	}

	if err := query.Where("delivered = ?", true).Count(&delivered).Error; err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to count delivered webhooks", err)
	}

	return float64(delivered) / float64(total) * 100, nil
}

func (r *webhookEventRepository) GetAverageDeliveryTime(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (time.Duration, error) {
	query := `
	SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (delivered_at - created_at))), 0)
	FROM webhook_events
	WHERE tenant_id = ? AND delivered = true
`

	args := []any{tenantID}

	if !startDate.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, startDate)
	}
	if !endDate.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, endDate)
	}

	var avgSeconds float64
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&avgSeconds).Error; err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to calculate average delivery time", err)
	}

	return time.Duration(avgSeconds) * time.Second, nil
}

func (r *webhookEventRepository) GetWebhooksByEventType(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[models.WebhookEventType]int64, error) {
	result := make(map[models.WebhookEventType]int64)

	query := r.db.WithContext(ctx).Model(&models.WebhookEvent{}).Where("tenant_id = ?", tenantID)
	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}

	rows, err := query.
		Select("event_type, COUNT(*) as count").
		Group("event_type").
		Rows()

	if err != nil {
		return nil, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to get webhooks by event type", err)
	}
	defer rows.Close()

	for rows.Next() {
		var eventType models.WebhookEventType
		var count int64
		if err := rows.Scan(&eventType, &count); err != nil {
			continue
		}
		result[eventType] = count
	}

	return result, nil
}

func (r *webhookEventRepository) GetFailureReasons(ctx context.Context, tenantID uuid.UUID, limit int) ([]FailureReasonCount, error) {
	var results []FailureReasonCount

	rows, err := r.db.WithContext(ctx).
		Model(&models.WebhookEvent{}).
		Select("failure_reason as reason, COUNT(*) as count").
		Where("tenant_id = ? AND delivered = ? AND failure_reason != ''", tenantID, false).
		Group("failure_reason").
		Order("count DESC").
		Limit(limit).
		Rows()

	if err != nil {
		return nil, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to get failure reasons", err)
	}
	defer rows.Close()

	for rows.Next() {
		var fr FailureReasonCount
		if err := rows.Scan(&fr.Reason, &fr.Count); err != nil {
			continue
		}
		results = append(results, fr)
	}

	return results, nil
}

//------------------------------------------------------------
// Cleanup Operations
//------------------------------------------------------------

func (r *webhookEventRepository) DeleteOldWebhooks(ctx context.Context, olderThan time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", olderThan).
		Delete(&models.WebhookEvent{})

	if result.Error != nil {
		return 0, errors.NewRepositoryError("DELETE_FAILED", "failed to delete old webhooks", result.Error)
	}

	return result.RowsAffected, nil
}

func (r *webhookEventRepository) DeleteDeliveredWebhooks(ctx context.Context, olderThan time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("delivered = ? AND delivered_at < ?", true, olderThan).
		Delete(&models.WebhookEvent{})

	if result.Error != nil {
		return 0, errors.NewRepositoryError("DELETE_FAILED", "failed to delete delivered webhooks", result.Error)
	}

	return result.RowsAffected, nil
}

func (r *webhookEventRepository) PurgeFailedWebhooks(ctx context.Context, maxAttempts int, olderThan time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("delivered = ? AND attempt_count >= ? AND created_at < ?", false, maxAttempts, olderThan).
		Delete(&models.WebhookEvent{})

	if result.Error != nil {
		return 0, errors.NewRepositoryError("DELETE_FAILED", "failed to purge failed webhooks", result.Error)
	}

	return result.RowsAffected, nil
}

//------------------------------------------------------------
// Search & Filter
//------------------------------------------------------------

func (r *webhookEventRepository) FindByFilters(ctx context.Context, filters WebhookEventFilters, pagination PaginationParams) ([]*models.WebhookEvent, PaginationResult, error) {
	if filters.TenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id is required", errors.ErrInvalidInput)
	}

	pagination.Validate()

	query := r.db.WithContext(ctx).Model(&models.WebhookEvent{})
	query = r.applyWebhookFilters(query, filters)

	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count webhooks", err)
	}

	var webhooks []*models.WebhookEvent
	if err := r.applyWebhookFilters(r.db.WithContext(ctx).Model(&models.WebhookEvent{}), filters).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&webhooks).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to apply filters", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return webhooks, paginationResult, nil
}

//------------------------------------------------------------
// Helper Methods
//------------------------------------------------------------

func (r *webhookEventRepository) applyWebhookFilters(query *gorm.DB, filters WebhookEventFilters) *gorm.DB {
	query = query.Where("tenant_id = ?", filters.TenantID)

	if len(filters.EventTypes) > 0 {
		query = query.Where("event_type IN ?", filters.EventTypes)
	}

	if filters.Delivered != nil {
		query = query.Where("delivered = ?", *filters.Delivered)
	}

	if filters.WebhookURL != "" {
		query = query.Where("webhook_url = ?", filters.WebhookURL)
	}

	if filters.MinAttempts != nil {
		query = query.Where("attempt_count >= ?", *filters.MinAttempts)
	}

	if filters.MaxAttempts != nil {
		query = query.Where("attempt_count <= ?", *filters.MaxAttempts)
	}

	if filters.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *filters.CreatedFrom)
	}

	if filters.CreatedTo != nil {
		query = query.Where("created_at <= ?", *filters.CreatedTo)
	}

	if len(filters.ResponseCodes) > 0 {
		query = query.Where("response_code IN ?", filters.ResponseCodes)
	}

	return query
}

func (r *webhookEventRepository) InvalidateCache(ctx context.Context, webhookID uuid.UUID) error {
	if r.cache != nil {
		cacheKey := fmt.Sprintf("webhook:%s", webhookID.String())
		if err := r.cache.Delete(ctx, cacheKey); err != nil {
			if r.logger != nil {
				r.logger.Warnf("Failed to invalidate cache for webhook %s: %v", webhookID, err)
			}
			return err
		}
	}
	return nil
}
