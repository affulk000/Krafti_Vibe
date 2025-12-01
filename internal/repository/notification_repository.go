package repository

import (
	"context"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationRepository defines the interface for notification repository operations
type NotificationRepository interface {
	BaseRepository[models.Notification]

	// FindByUserID retrieves notifications for a user with pagination
	FindByUserID(ctx context.Context, userID uuid.UUID, pagination PaginationParams) ([]*models.Notification, PaginationResult, error)

	// FindUnreadByUser retrieves unread notifications for a user
	FindUnreadByUser(ctx context.Context, userID uuid.UUID) ([]*models.Notification, error)

	// FindByType retrieves notifications by type
	FindByType(ctx context.Context, userID uuid.UUID, notificationType models.NotificationType, pagination PaginationParams) ([]*models.Notification, PaginationResult, error)

	// FindByPriority retrieves notifications by priority
	FindByPriority(ctx context.Context, userID uuid.UUID, priority int, pagination PaginationParams) ([]*models.Notification, PaginationResult, error)

	// MarkAsRead marks a notification as read
	MarkAsRead(ctx context.Context, notificationID uuid.UUID) error

	// MarkAllAsRead marks all notifications as read for a user
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error

	// MarkMultipleAsRead marks multiple notifications as read
	MarkMultipleAsRead(ctx context.Context, notificationIDs []uuid.UUID) error

	// DeleteRead deletes read notifications older than specified days
	DeleteRead(ctx context.Context, userID uuid.UUID, olderThanDays int) error

	// DeleteOld deletes all notifications older than specified days
	DeleteOld(ctx context.Context, olderThanDays int) error

	// GetUnreadCount retrieves count of unread notifications for a user
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)

	// GetUnreadCountByType retrieves count of unread notifications by type
	GetUnreadCountByType(ctx context.Context, userID uuid.UUID, notificationType models.NotificationType) (int64, error)

	// FindPending retrieves pending notifications (not sent yet)
	FindPending(ctx context.Context, limit int) ([]*models.Notification, error)

	// MarkAsSent marks a notification as sent
	MarkAsSent(ctx context.Context, notificationID uuid.UUID) error

	// MarkAsFailed marks a notification as failed
	MarkAsFailed(ctx context.Context, notificationID uuid.UUID, errorMessage string) error

	// FindByRelatedEntity retrieves notifications related to an entity
	FindByRelatedEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.Notification, error)

	// GetNotificationStats retrieves notification statistics for a user
	GetNotificationStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)

	// FindRecentByUser retrieves recent notifications
	FindRecentByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*models.Notification, error)

	// BulkCreate creates multiple notifications
	BulkCreate(ctx context.Context, notifications []*models.Notification) error

	// FindScheduled retrieves notifications scheduled for sending
	FindScheduled(ctx context.Context, before time.Time) ([]*models.Notification, error)

	// Search searches notifications
	Search(ctx context.Context, userID uuid.UUID, query string, pagination PaginationParams) ([]*models.Notification, PaginationResult, error)

	// FindHighPriority retrieves high priority unread notifications
	FindHighPriority(ctx context.Context, userID uuid.UUID) ([]*models.Notification, error)
}

// notificationRepository implements NotificationRepository
type notificationRepository struct {
	BaseRepository[models.Notification]
	db     *gorm.DB
	logger log.AllLogger
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *gorm.DB, config ...RepositoryConfig) NotificationRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.Notification](db, cfg)

	return &notificationRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
	}
}

// FindByUserID retrieves notifications for a user with pagination
func (r *notificationRepository) FindByUserID(ctx context.Context, userID uuid.UUID, pagination PaginationParams) ([]*models.Notification, PaginationResult, error) {
	if userID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count notifications", "user_id", userID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count notifications", err)
	}

	// Find paginated results
	var notifications []*models.Notification
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		r.logger.Error("failed to find notifications", "user_id", userID, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find notifications", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return notifications, paginationResult, nil
}

// FindUnreadByUser retrieves unread notifications for a user
func (r *notificationRepository) FindUnreadByUser(ctx context.Context, userID uuid.UUID) ([]*models.Notification, error) {
	if userID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	var notifications []*models.Notification
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_read = ?", userID, false).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		r.logger.Error("failed to find unread notifications", "user_id", userID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find unread notifications", err)
	}

	return notifications, nil
}

// FindByType retrieves notifications by type
func (r *notificationRepository) FindByType(ctx context.Context, userID uuid.UUID, notificationType models.NotificationType, pagination PaginationParams) ([]*models.Notification, PaginationResult, error) {
	if userID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND type = ?", userID, notificationType).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count notifications by type", "user_id", userID, "type", notificationType, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count notifications", err)
	}

	// Find paginated results
	var notifications []*models.Notification
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, notificationType).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		r.logger.Error("failed to find notifications by type", "user_id", userID, "type", notificationType, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find notifications", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return notifications, paginationResult, nil
}

// FindByPriority retrieves notifications by priority
func (r *notificationRepository) FindByPriority(ctx context.Context, userID uuid.UUID, priority int, pagination PaginationParams) ([]*models.Notification, PaginationResult, error) {
	if userID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND priority = ?", userID, priority).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count notifications by priority", "user_id", userID, "priority", priority, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count notifications", err)
	}

	// Find paginated results
	var notifications []*models.Notification
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND priority = ?", userID, priority).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		r.logger.Error("failed to find notifications by priority", "user_id", userID, "priority", priority, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find notifications", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return notifications, paginationResult, nil
}

// MarkAsRead marks a notification as read
func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	if notificationID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "notification_id cannot be nil", errors.ErrInvalidInput)
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ?", notificationID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		})

	if result.Error != nil {
		r.logger.Error("failed to mark notification as read", "notification_id", notificationID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark notification as read", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "notification not found", errors.ErrNotFound)
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		})

	if result.Error != nil {
		r.logger.Error("failed to mark all notifications as read", "user_id", userID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark all notifications as read", result.Error)
	}

	r.logger.Info("marked all notifications as read", "user_id", userID, "count", result.RowsAffected)
	return nil
}

// MarkMultipleAsRead marks multiple notifications as read
func (r *notificationRepository) MarkMultipleAsRead(ctx context.Context, notificationIDs []uuid.UUID) error {
	if len(notificationIDs) == 0 {
		return nil
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id IN ?", notificationIDs).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		})

	if result.Error != nil {
		r.logger.Error("failed to mark notifications as read", "count", len(notificationIDs), "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark notifications as read", result.Error)
	}

	return nil
}

// DeleteRead deletes read notifications older than specified days
func (r *notificationRepository) DeleteRead(ctx context.Context, userID uuid.UUID, olderThanDays int) error {
	if userID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	if olderThanDays <= 0 {
		olderThanDays = 30
	}

	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND is_read = ? AND created_at < ?", userID, true, cutoffDate).
		Delete(&models.Notification{})

	if result.Error != nil {
		r.logger.Error("failed to delete read notifications", "user_id", userID, "error", result.Error)
		return errors.NewRepositoryError("DELETE_FAILED", "failed to delete read notifications", result.Error)
	}

	r.logger.Info("deleted read notifications", "user_id", userID, "count", result.RowsAffected, "older_than_days", olderThanDays)
	return nil
}

// DeleteOld deletes all notifications older than specified days
func (r *notificationRepository) DeleteOld(ctx context.Context, olderThanDays int) error {
	if olderThanDays <= 0 {
		olderThanDays = 90
	}

	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)

	result := r.db.WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Delete(&models.Notification{})

	if result.Error != nil {
		r.logger.Error("failed to delete old notifications", "error", result.Error)
		return errors.NewRepositoryError("DELETE_FAILED", "failed to delete old notifications", result.Error)
	}

	r.logger.Info("deleted old notifications", "count", result.RowsAffected, "older_than_days", olderThanDays)
	return nil
}

// GetUnreadCount retrieves count of unread notifications for a user
func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	if userID == uuid.Nil {
		return 0, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		r.logger.Error("failed to count unread notifications", "user_id", userID, "error", err)
		return 0, errors.NewRepositoryError("COUNT_FAILED", "failed to count unread notifications", err)
	}

	return count, nil
}

// GetUnreadCountByType retrieves count of unread notifications by type
func (r *notificationRepository) GetUnreadCountByType(ctx context.Context, userID uuid.UUID, notificationType models.NotificationType) (int64, error) {
	if userID == uuid.Nil {
		return 0, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND type = ? AND is_read = ?", userID, notificationType, false).
		Count(&count).Error; err != nil {
		r.logger.Error("failed to count unread notifications by type", "user_id", userID, "type", notificationType, "error", err)
		return 0, errors.NewRepositoryError("COUNT_FAILED", "failed to count unread notifications", err)
	}

	return count, nil
}

// FindPending retrieves pending notifications (not sent yet)
func (r *notificationRepository) FindPending(ctx context.Context, limit int) ([]*models.Notification, error) {
	if limit <= 0 {
		limit = 100
	}

	var notifications []*models.Notification
	if err := r.db.WithContext(ctx).
		Where("sent_via_in_app = ? AND sent_via_email = ? AND sent_via_sms = ? AND sent_via_push = ?", false, false, false, false).
		Order("created_at ASC").
		Limit(limit).
		Find(&notifications).Error; err != nil {
		r.logger.Error("failed to find pending notifications", "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find pending notifications", err)
	}

	return notifications, nil
}

// MarkAsSent marks a notification as sent
func (r *notificationRepository) MarkAsSent(ctx context.Context, notificationID uuid.UUID) error {
	if notificationID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "notification_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ?", notificationID).
		Updates(map[string]interface{}{
			"sent_via_in_app": true,
		})

	if result.Error != nil {
		r.logger.Error("failed to mark notification as sent", "notification_id", notificationID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark notification as sent", result.Error)
	}

	return nil
}

// MarkAsFailed marks a notification as failed (stores error in metadata)
func (r *notificationRepository) MarkAsFailed(ctx context.Context, notificationID uuid.UUID, errorMessage string) error {
	if notificationID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "notification_id cannot be nil", errors.ErrInvalidInput)
	}

	metadata := models.JSONB{"error": errorMessage, "failed_at": time.Now()}
	result := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ?", notificationID).
		Update("metadata", metadata)

	if result.Error != nil {
		r.logger.Error("failed to mark notification as failed", "notification_id", notificationID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark notification as failed", result.Error)
	}

	return nil
}

// FindByRelatedEntity retrieves notifications related to an entity
func (r *notificationRepository) FindByRelatedEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.Notification, error) {
	if entityType == "" || entityID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "entity_type and entity_id are required", errors.ErrInvalidInput)
	}

	var notifications []*models.Notification
	if err := r.db.WithContext(ctx).
		Where("related_entity_type = ? AND related_entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		r.logger.Error("failed to find notifications by entity", "entity_type", entityType, "entity_id", entityID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find notifications", err)
	}

	return notifications, nil
}

// GetNotificationStats retrieves notification statistics for a user
func (r *notificationRepository) GetNotificationStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	if userID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	stats := make(map[string]interface{})

	// Total notifications
	var total int64
	r.db.WithContext(ctx).Model(&models.Notification{}).Where("user_id = ?", userID).Count(&total)
	stats["total"] = total

	// Unread count
	unreadCount, _ := r.GetUnreadCount(ctx, userID)
	stats["unread"] = unreadCount
	stats["read"] = total - unreadCount

	// By type
	type TypeCount struct {
		Type  models.NotificationType
		Count int64
	}
	var typeCounts []TypeCount
	r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Select("type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("type").
		Scan(&typeCounts)

	typeMap := make(map[string]int64)
	for _, tc := range typeCounts {
		typeMap[string(tc.Type)] = tc.Count
	}
	stats["by_type"] = typeMap

	// Recent (last 7 days)
	var recentCount int64
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND created_at >= ?", userID, sevenDaysAgo).
		Count(&recentCount)
	stats["recent_7d"] = recentCount

	return stats, nil
}

// FindRecentByUser retrieves recent notifications
func (r *notificationRepository) FindRecentByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*models.Notification, error) {
	if userID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	if limit <= 0 {
		limit = 10
	}

	var notifications []*models.Notification
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error; err != nil {
		r.logger.Error("failed to find recent notifications", "user_id", userID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find recent notifications", err)
	}

	return notifications, nil
}

// BulkCreate creates multiple notifications
func (r *notificationRepository) BulkCreate(ctx context.Context, notifications []*models.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(notifications, 100).Error; err != nil {
		r.logger.Error("failed to bulk create notifications", "count", len(notifications), "error", err)
		return errors.NewRepositoryError("CREATE_FAILED", "failed to bulk create notifications", err)
	}

	r.logger.Info("bulk created notifications", "count", len(notifications))
	return nil
}

// FindScheduled retrieves notifications scheduled for sending (based on expires_at)
func (r *notificationRepository) FindScheduled(ctx context.Context, before time.Time) ([]*models.Notification, error) {
	var notifications []*models.Notification
	if err := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at <= ?", before).
		Where("sent_via_in_app = ?", false).
		Order("expires_at ASC").
		Find(&notifications).Error; err != nil {
		r.logger.Error("failed to find scheduled notifications", "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find scheduled notifications", err)
	}

	return notifications, nil
}

// Search searches notifications
func (r *notificationRepository) Search(ctx context.Context, userID uuid.UUID, query string, pagination PaginationParams) ([]*models.Notification, PaginationResult, error) {
	if userID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()
	searchPattern := "%" + query + "%"

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND (title ILIKE ? OR message ILIKE ?)", userID, searchPattern, searchPattern).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count search results", "user_id", userID, "query", query, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count notifications", err)
	}

	// Find paginated results
	var notifications []*models.Notification
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND (title ILIKE ? OR message ILIKE ?)", userID, searchPattern, searchPattern).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		r.logger.Error("failed to search notifications", "user_id", userID, "query", query, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search notifications", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return notifications, paginationResult, nil
}

// FindHighPriority retrieves high priority unread notifications
func (r *notificationRepository) FindHighPriority(ctx context.Context, userID uuid.UUID) ([]*models.Notification, error) {
	if userID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	var notifications []*models.Notification
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_read = ? AND priority <= ?", userID, false, 3).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		r.logger.Error("failed to find high priority notifications", "user_id", userID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find high priority notifications", err)
	}

	return notifications, nil
}
