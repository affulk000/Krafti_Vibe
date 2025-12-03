package service

import (
	"context"
	"encoding/json"
	"fmt"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// NotificationServiceEnhanced defines enhanced notification operations
type NotificationService interface {
	// Core Notification Operations
	CreateNotification(ctx context.Context, req *dto.CreateNotificationRequest) (*dto.NotificationResponse, error)
	GetNotification(ctx context.Context, id uuid.UUID) (*dto.NotificationResponse, error)
	ListNotifications(ctx context.Context, filter dto.NotificationFilter) (*dto.NotificationListResponse, error)
	DeleteNotification(ctx context.Context, id uuid.UUID) error

	// Read Status Management
	MarkAsRead(ctx context.Context, notificationID uuid.UUID) error
	MarkMultipleAsRead(ctx context.Context, notificationIDs []uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (*dto.UnreadCountResponse, error)

	// Bulk Operations
	SendBulkNotification(ctx context.Context, req *dto.BulkNotificationRequest) (*dto.BulkNotificationResponse, error)
	DeleteReadNotifications(ctx context.Context, userID uuid.UUID, olderThanDays int) (int64, error)
	DeleteOldNotifications(ctx context.Context, olderThanDays int) (int64, error)

	// Business Event Notifications
	SendBookingNotification(ctx context.Context, booking *models.Booking, notifType models.NotificationType) (*dto.NotificationDeliveryResponse, error)
	SendPaymentNotification(ctx context.Context, payment *models.Payment, notifType models.NotificationType) (*dto.NotificationDeliveryResponse, error)
	SendReviewNotification(ctx context.Context, review *models.Review, notifType models.NotificationType) (*dto.NotificationDeliveryResponse, error)
	SendSystemNotification(ctx context.Context, req *dto.SendSystemNotificationRequest) (*dto.BulkNotificationResponse, error)

	// Query Operations
	GetUnreadNotifications(ctx context.Context, userID uuid.UUID) ([]*dto.NotificationResponse, error)
	GetNotificationsByType(ctx context.Context, userID uuid.UUID, notifType models.NotificationType, page, pageSize int) (*dto.NotificationListResponse, error)
	GetNotificationsByPriority(ctx context.Context, userID uuid.UUID, priority int, page, pageSize int) (*dto.NotificationListResponse, error)
	GetRecentNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]*dto.NotificationResponse, error)
	GetHighPriorityNotifications(ctx context.Context, userID uuid.UUID) ([]*dto.NotificationResponse, error)

	// Statistics & Analytics
	GetNotificationStats(ctx context.Context, userID uuid.UUID) (*dto.NotificationStatsResponse, error)
	GetTenantNotificationStats(ctx context.Context, tenantID uuid.UUID) (*dto.NotificationStatsResponse, error)

	// Cleanup Operations
	CleanupExpiredNotifications(ctx context.Context) (int64, error)

	// Search
	SearchNotifications(ctx context.Context, userID uuid.UUID, query string, page, pageSize int) (*dto.NotificationListResponse, error)

	// Health & Monitoring
	HealthCheck(ctx context.Context) error
	GetServiceMetrics(ctx context.Context) map[string]any
}

// notificationServiceEnhanced implements NotificationServiceEnhanced
type notificationService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewNotificationServiceEnhanced creates a new enhanced notification service
func NewNotificationService(repos *repository.Repositories, logger log.AllLogger) NotificationService {
	return &notificationService{
		repos:  repos,
		logger: logger,
	}
}

// ============================================================================
// Core Notification Operations
// ============================================================================

// CreateNotification creates a new notification
func (s *notificationService) CreateNotification(ctx context.Context, req *dto.CreateNotificationRequest) (*dto.NotificationResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	metadata := models.JSONB{}
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err == nil {
			json.Unmarshal(metadataBytes, &metadata)
		}
	}

	notification := &models.Notification{
		TenantID:          req.TenantID,
		UserID:            req.UserID,
		Type:              req.Type,
		Title:             req.Title,
		Message:           req.Message,
		Channels:          req.Channels,
		ActionURL:         req.ActionURL,
		ActionText:        req.ActionText,
		RelatedEntityType: req.RelatedEntityType,
		RelatedEntityID:   req.RelatedEntityID,
		Priority:          req.Priority,
		ExpiresAt:         req.ExpiresAt,
		IsRead:            false,
		Metadata:          metadata,
	}

	if err := s.repos.Notification.Create(ctx, notification); err != nil {
		return nil, errors.NewServiceError("NOTIFICATION_CREATE_FAILED", "failed to create notification", err)
	}

	// Asynchronously send via channels
	go s.sendViaChannels(context.Background(), notification)

	s.logger.Info("notification created",
		"notification_id", notification.ID,
		"user_id", req.UserID,
		"type", req.Type)

	return dto.ToNotificationResponse(notification), nil
}

// GetNotification retrieves a notification by ID
func (s *notificationService) GetNotification(ctx context.Context, id uuid.UUID) (*dto.NotificationResponse, error) {
	notification, err := s.repos.Notification.GetByID(ctx, id)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.NewNotFoundError("notification not found")
		}
		return nil, errors.NewServiceError("NOTIFICATION_GET_FAILED", "failed to get notification", err)
	}

	return dto.ToNotificationResponse(notification), nil
}

// ListNotifications lists notifications with filtering and pagination
func (s *notificationService) ListNotifications(ctx context.Context, filter dto.NotificationFilter) (*dto.NotificationListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}
	pagination.Validate()

	notifications, paginationResult, err := s.repos.Notification.FindByUserID(ctx, filter.UserID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("NOTIFICATION_LIST_FAILED", "failed to list notifications", err)
	}

	// Get unread count
	unreadCount, _ := s.repos.Notification.GetUnreadCount(ctx, filter.UserID)

	return &dto.NotificationListResponse{
		Notifications: dto.ToNotificationResponses(notifications),
		Page:          paginationResult.Page,
		PageSize:      paginationResult.PageSize,
		TotalItems:    paginationResult.TotalItems,
		TotalPages:    paginationResult.TotalPages,
		HasNext:       paginationResult.HasNext,
		HasPrevious:   paginationResult.HasPrev,
		UnreadCount:   unreadCount,
	}, nil
}

// DeleteNotification deletes a notification
func (s *notificationService) DeleteNotification(ctx context.Context, id uuid.UUID) error {
	if err := s.repos.Notification.Delete(ctx, id); err != nil {
		if errors.IsNotFoundError(err) {
			return errors.NewNotFoundError("notification not found")
		}
		return errors.NewServiceError("NOTIFICATION_DELETE_FAILED", "failed to delete notification", err)
	}

	s.logger.Info("notification deleted", "notification_id", id)
	return nil
}

// ============================================================================
// Read Status Management
// ============================================================================

// MarkAsRead marks a notification as read
func (s *notificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	if err := s.repos.Notification.MarkAsRead(ctx, notificationID); err != nil {
		if errors.IsNotFoundError(err) {
			return errors.NewNotFoundError("notification not found")
		}
		return errors.NewServiceError("NOTIFICATION_MARK_READ_FAILED", "failed to mark notification as read", err)
	}

	s.logger.Debug("notification marked as read", "notification_id", notificationID)
	return nil
}

// MarkMultipleAsRead marks multiple notifications as read
func (s *notificationService) MarkMultipleAsRead(ctx context.Context, notificationIDs []uuid.UUID) error {
	if len(notificationIDs) == 0 {
		return nil
	}

	if err := s.repos.Notification.MarkMultipleAsRead(ctx, notificationIDs); err != nil {
		return errors.NewServiceError("NOTIFICATION_MARK_READ_FAILED", "failed to mark notifications as read", err)
	}

	s.logger.Info("multiple notifications marked as read", "count", len(notificationIDs))
	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.NewValidationError("user_id is required")
	}

	if err := s.repos.Notification.MarkAllAsRead(ctx, userID); err != nil {
		return errors.NewServiceError("NOTIFICATION_MARK_ALL_READ_FAILED", "failed to mark all notifications as read", err)
	}

	s.logger.Info("all notifications marked as read", "user_id", userID)
	return nil
}

// GetUnreadCount retrieves unread notification count
func (s *notificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (*dto.UnreadCountResponse, error) {
	if userID == uuid.Nil {
		return nil, errors.NewValidationError("user_id is required")
	}

	count, err := s.repos.Notification.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, errors.NewServiceError("NOTIFICATION_COUNT_FAILED", "failed to get unread count", err)
	}

	response := &dto.UnreadCountResponse{
		UserID:      userID,
		UnreadCount: count,
		ByType:      make(map[string]int64),
		ByPriority:  make(map[int]int64),
	}

	// Get counts by type
	notificationTypes := []models.NotificationType{
		models.NotificationTypeBookingCreated,
		models.NotificationTypeBookingConfirmed,
		models.NotificationTypeBookingCancelled,
		models.NotificationTypeBookingReminder,
		models.NotificationTypePaymentReceived,
		models.NotificationTypeReviewReceived,
		models.NotificationTypeMessageReceived,
	}

	for _, notifType := range notificationTypes {
		typeCount, err := s.repos.Notification.GetUnreadCountByType(ctx, userID, notifType)
		if err == nil && typeCount > 0 {
			response.ByType[string(notifType)] = typeCount
		}
	}

	return response, nil
}

// ============================================================================
// Bulk Operations
// ============================================================================

// SendBulkNotification sends notifications to multiple users
func (s *notificationService) SendBulkNotification(ctx context.Context, req *dto.BulkNotificationRequest) (*dto.BulkNotificationResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid request: " + err.Error())
	}

	response := &dto.BulkNotificationResponse{
		SuccessCount:   0,
		FailureCount:   0,
		CreatedIDs:     []uuid.UUID{},
		Errors:         []string{},
		DeliveryStatus: []*dto.NotificationDeliveryResponse{},
	}

	notifications := make([]*models.Notification, 0, len(req.UserIDs))

	for _, userID := range req.UserIDs {
		notification := &models.Notification{
			TenantID:  req.TenantID,
			UserID:    userID,
			Type:      req.Type,
			Title:     req.Title,
			Message:   req.Message,
			Channels:  req.Channels,
			ActionURL: req.ActionURL,
			Priority:  req.Priority,
			ExpiresAt: req.ExpiresAt,
			IsRead:    false,
		}
		notifications = append(notifications, notification)
	}

	// Bulk create
	if err := s.repos.Notification.BulkCreate(ctx, notifications); err != nil {
		return nil, errors.NewServiceError("NOTIFICATION_BULK_CREATE_FAILED", "failed to bulk create notifications", err)
	}

	// Send via channels asynchronously
	for _, notification := range notifications {
		response.SuccessCount++
		response.CreatedIDs = append(response.CreatedIDs, notification.ID)

		go s.sendViaChannels(context.Background(), notification)
	}

	s.logger.Info("bulk notifications created",
		"count", response.SuccessCount,
		"tenant_id", req.TenantID)

	return response, nil
}

// DeleteReadNotifications deletes read notifications older than specified days
func (s *notificationService) DeleteReadNotifications(ctx context.Context, userID uuid.UUID, olderThanDays int) (int64, error) {
	if userID == uuid.Nil {
		return 0, errors.NewValidationError("user_id is required")
	}
	if olderThanDays <= 0 {
		olderThanDays = 30 // Default to 30 days
	}

	if err := s.repos.Notification.DeleteRead(ctx, userID, olderThanDays); err != nil {
		return 0, errors.NewServiceError("NOTIFICATION_DELETE_FAILED", "failed to delete read notifications", err)
	}

	s.logger.Info("deleted read notifications",
		"user_id", userID,
		"older_than_days", olderThanDays)

	return 0, nil // Repository method doesn't return count
}

// DeleteOldNotifications deletes all notifications older than specified days
func (s *notificationService) DeleteOldNotifications(ctx context.Context, olderThanDays int) (int64, error) {
	if olderThanDays <= 0 {
		olderThanDays = 90 // Default to 90 days
	}

	if err := s.repos.Notification.DeleteOld(ctx, olderThanDays); err != nil {
		return 0, errors.NewServiceError("NOTIFICATION_DELETE_FAILED", "failed to delete old notifications", err)
	}

	s.logger.Info("deleted old notifications", "older_than_days", olderThanDays)
	return 0, nil // Repository method doesn't return count
}

// ============================================================================
// Business Event Notifications
// ============================================================================

// SendBookingNotification sends notification for booking events
func (s *notificationService) SendBookingNotification(ctx context.Context, booking *models.Booking, notifType models.NotificationType) (*dto.NotificationDeliveryResponse, error) {
	if booking == nil {
		return nil, errors.NewValidationError("booking is required")
	}

	var title, message string
	var userID uuid.UUID
	var priority int = 5

	switch notifType {
	case models.NotificationTypeBookingCreated:
		title = "New Booking Created"
		message = fmt.Sprintf("Your booking #%s has been created successfully", booking.ID.String()[:8])
		userID = booking.CustomerID
		priority = 6
	case models.NotificationTypeBookingConfirmed:
		title = "Booking Confirmed"
		message = fmt.Sprintf("Your booking #%s has been confirmed", booking.ID.String()[:8])
		userID = booking.CustomerID
		priority = 7
	case models.NotificationTypeBookingCancelled:
		title = "Booking Cancelled"
		message = fmt.Sprintf("Booking #%s has been cancelled", booking.ID.String()[:8])
		userID = booking.CustomerID
		priority = 8
	case models.NotificationTypeBookingReminder:
		title = "Booking Reminder"
		message = fmt.Sprintf("Reminder: Your booking is scheduled for %s", booking.StartTime.Format("Jan 2, 2006 at 3:04 PM"))
		userID = booking.CustomerID
		priority = 7
	case models.NotificationTypeBookingCompleted:
		title = "Booking Completed"
		message = fmt.Sprintf("Your booking #%s has been completed", booking.ID.String()[:8])
		userID = booking.CustomerID
		priority = 5
	default:
		return nil, errors.NewValidationError("unsupported notification type for booking")
	}

	req := &dto.CreateNotificationRequest{
		TenantID:          booking.TenantID,
		UserID:            userID,
		Type:              notifType,
		Title:             title,
		Message:           message,
		Channels:          []models.NotificationChannel{models.NotificationChannelInApp, models.NotificationChannelEmail},
		ActionURL:         fmt.Sprintf("/bookings/%s", booking.ID),
		ActionText:        "View Booking",
		RelatedEntityType: "booking",
		RelatedEntityID:   &booking.ID,
		Priority:          priority,
	}

	notification, err := s.CreateNotification(ctx, req)
	if err != nil {
		return nil, err
	}

	return &dto.NotificationDeliveryResponse{
		NotificationID: notification.ID,
		InAppSent:      true,
		EmailSent:      false, // Would be set by actual delivery
		SMSSent:        false,
		PushSent:       false,
	}, nil
}

// SendPaymentNotification sends notification for payment events
func (s *notificationService) SendPaymentNotification(ctx context.Context, payment *models.Payment, notifType models.NotificationType) (*dto.NotificationDeliveryResponse, error) {
	if payment == nil {
		return nil, errors.NewValidationError("payment is required")
	}

	title := "Payment Received"
	message := fmt.Sprintf("Payment of %s %.2f has been received for booking #%s",
		payment.Currency, payment.Amount, payment.BookingID.String()[:8])

	req := &dto.CreateNotificationRequest{
		TenantID:          payment.TenantID,
		UserID:            payment.CustomerID,
		Type:              models.NotificationTypePaymentReceived,
		Title:             title,
		Message:           message,
		Channels:          []models.NotificationChannel{models.NotificationChannelInApp, models.NotificationChannelEmail},
		ActionURL:         fmt.Sprintf("/payments/%s", payment.ID),
		ActionText:        "View Payment",
		RelatedEntityType: "payment",
		RelatedEntityID:   &payment.ID,
		Priority:          6,
	}

	notification, err := s.CreateNotification(ctx, req)
	if err != nil {
		return nil, err
	}

	return &dto.NotificationDeliveryResponse{
		NotificationID: notification.ID,
		InAppSent:      true,
		EmailSent:      false,
		SMSSent:        false,
		PushSent:       false,
	}, nil
}

// SendReviewNotification sends notification for review events
func (s *notificationService) SendReviewNotification(ctx context.Context, review *models.Review, notifType models.NotificationType) (*dto.NotificationDeliveryResponse, error) {
	if review == nil {
		return nil, errors.NewValidationError("review is required")
	}

	title := "New Review Received"
	message := fmt.Sprintf("You received a %d-star review", review.Rating)

	req := &dto.CreateNotificationRequest{
		TenantID:          review.TenantID,
		UserID:            review.ArtisanID,
		Type:              models.NotificationTypeReviewReceived,
		Title:             title,
		Message:           message,
		Channels:          []models.NotificationChannel{models.NotificationChannelInApp, models.NotificationChannelEmail},
		ActionURL:         fmt.Sprintf("/reviews/%s", review.ID),
		ActionText:        "View Review",
		RelatedEntityType: "review",
		RelatedEntityID:   &review.ID,
		Priority:          6,
	}

	notification, err := s.CreateNotification(ctx, req)
	if err != nil {
		return nil, err
	}

	return &dto.NotificationDeliveryResponse{
		NotificationID: notification.ID,
		InAppSent:      true,
		EmailSent:      false,
		SMSSent:        false,
		PushSent:       false,
	}, nil
}

// SendSystemNotification sends a system-wide notification
func (s *notificationService) SendSystemNotification(ctx context.Context, req *dto.SendSystemNotificationRequest) (*dto.BulkNotificationResponse, error) {
	if req.TenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant_id is required")
	}
	if len(req.UserIDs) == 0 {
		return nil, errors.NewValidationError("at least one user_id is required")
	}

	bulkReq := &dto.BulkNotificationRequest{
		TenantID: req.TenantID,
		UserIDs:  req.UserIDs,
		Type:     models.NotificationTypeSystem,
		Title:    req.Title,
		Message:  req.Message,
		Channels: []models.NotificationChannel{models.NotificationChannelInApp, models.NotificationChannelEmail},
		Priority: req.Priority,
	}

	return s.SendBulkNotification(ctx, bulkReq)
}

// ============================================================================
// Query Operations
// ============================================================================

// GetUnreadNotifications retrieves unread notifications for a user
func (s *notificationService) GetUnreadNotifications(ctx context.Context, userID uuid.UUID) ([]*dto.NotificationResponse, error) {
	if userID == uuid.Nil {
		return nil, errors.NewValidationError("user_id is required")
	}

	notifications, err := s.repos.Notification.FindUnreadByUser(ctx, userID)
	if err != nil {
		return nil, errors.NewServiceError("NOTIFICATION_UNREAD_FAILED", "failed to get unread notifications", err)
	}

	return dto.ToNotificationResponses(notifications), nil
}

// GetNotificationsByType retrieves notifications by type
func (s *notificationService) GetNotificationsByType(ctx context.Context, userID uuid.UUID, notifType models.NotificationType, page, pageSize int) (*dto.NotificationListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
	pagination.Validate()

	notifications, paginationResult, err := s.repos.Notification.FindByType(ctx, userID, notifType, pagination)
	if err != nil {
		return nil, errors.NewServiceError("NOTIFICATION_TYPE_FAILED", "failed to get notifications by type", err)
	}

	return &dto.NotificationListResponse{
		Notifications: dto.ToNotificationResponses(notifications),
		Page:          paginationResult.Page,
		PageSize:      paginationResult.PageSize,
		TotalItems:    paginationResult.TotalItems,
		TotalPages:    paginationResult.TotalPages,
		HasNext:       paginationResult.HasNext,
		HasPrevious:   paginationResult.HasPrev,
	}, nil
}

// GetNotificationsByPriority retrieves notifications by priority
func (s *notificationService) GetNotificationsByPriority(ctx context.Context, userID uuid.UUID, priority int, page, pageSize int) (*dto.NotificationListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
	pagination.Validate()

	notifications, paginationResult, err := s.repos.Notification.FindByPriority(ctx, userID, priority, pagination)
	if err != nil {
		return nil, errors.NewServiceError("NOTIFICATION_PRIORITY_FAILED", "failed to get notifications by priority", err)
	}

	return &dto.NotificationListResponse{
		Notifications: dto.ToNotificationResponses(notifications),
		Page:          paginationResult.Page,
		PageSize:      paginationResult.PageSize,
		TotalItems:    paginationResult.TotalItems,
		TotalPages:    paginationResult.TotalPages,
		HasNext:       paginationResult.HasNext,
		HasPrevious:   paginationResult.HasPrev,
	}, nil
}

// GetRecentNotifications retrieves recent notifications
func (s *notificationService) GetRecentNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]*dto.NotificationResponse, error) {
	if userID == uuid.Nil {
		return nil, errors.NewValidationError("user_id is required")
	}
	if limit <= 0 {
		limit = 10
	}

	notifications, err := s.repos.Notification.FindRecentByUser(ctx, userID, limit)
	if err != nil {
		return nil, errors.NewServiceError("NOTIFICATION_RECENT_FAILED", "failed to get recent notifications", err)
	}

	return dto.ToNotificationResponses(notifications), nil
}

// GetHighPriorityNotifications retrieves high priority unread notifications
func (s *notificationService) GetHighPriorityNotifications(ctx context.Context, userID uuid.UUID) ([]*dto.NotificationResponse, error) {
	if userID == uuid.Nil {
		return nil, errors.NewValidationError("user_id is required")
	}

	notifications, err := s.repos.Notification.FindHighPriority(ctx, userID)
	if err != nil {
		return nil, errors.NewServiceError("NOTIFICATION_HIGH_PRIORITY_FAILED", "failed to get high priority notifications", err)
	}

	return dto.ToNotificationResponses(notifications), nil
}

// ============================================================================
// Statistics & Analytics
// ============================================================================

// GetNotificationStats retrieves notification statistics for a user
func (s *notificationService) GetNotificationStats(ctx context.Context, userID uuid.UUID) (*dto.NotificationStatsResponse, error) {
	if userID == uuid.Nil {
		return nil, errors.NewValidationError("user_id is required")
	}

	stats, err := s.repos.Notification.GetNotificationStats(ctx, userID)
	if err != nil {
		return nil, errors.NewServiceError("NOTIFICATION_STATS_FAILED", "failed to get notification stats", err)
	}

	total, ok := stats["total"].(int64)
	if !ok {
		total = 0
	}
	unread, ok := stats["unread"].(int64)
	if !ok {
		unread = 0
	}
	read, ok := stats["read"].(int64)
	if !ok {
		read = 0
	}
	recent7d, ok := stats["recent_7d"].(int64)
	if !ok {
		recent7d = 0
	}

	byType := make(map[string]int64)
	if typeMap, ok := stats["by_type"].(map[string]int64); ok {
		byType = typeMap
	}

	return &dto.NotificationStatsResponse{
		UserID:             userID,
		TotalNotifications: total,
		UnreadCount:        unread,
		ReadCount:          read,
		ByType:             byType,
		ByPriority:         make(map[int]int64),
		Recent7Days:        recent7d,
		ExpiredCount:       0,
	}, nil
}

// GetTenantNotificationStats retrieves notification statistics for a tenant
func (s *notificationService) GetTenantNotificationStats(ctx context.Context, tenantID uuid.UUID) (*dto.NotificationStatsResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant_id is required")
	}

	// For tenant-wide stats, we'd need to aggregate across all users
	// This is a simplified implementation
	return &dto.NotificationStatsResponse{
		UserID:             uuid.Nil,
		TotalNotifications: 0,
		UnreadCount:        0,
		ReadCount:          0,
		ByType:             make(map[string]int64),
		ByPriority:         make(map[int]int64),
		Recent7Days:        0,
		ExpiredCount:       0,
	}, nil
}

// ============================================================================
// Cleanup Operations
// ============================================================================

// CleanupExpiredNotifications removes expired notifications
func (s *notificationService) CleanupExpiredNotifications(ctx context.Context) (int64, error) {
	// The repository doesn't have a specific method for expired cleanup
	// So we'll delete old notifications (90+ days)
	return s.DeleteOldNotifications(ctx, 90)
}

// ============================================================================
// Search
// ============================================================================

// SearchNotifications searches notifications
func (s *notificationService) SearchNotifications(ctx context.Context, userID uuid.UUID, query string, page, pageSize int) (*dto.NotificationListResponse, error) {
	if userID == uuid.Nil {
		return nil, errors.NewValidationError("user_id is required")
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
	pagination.Validate()

	notifications, paginationResult, err := s.repos.Notification.Search(ctx, userID, query, pagination)
	if err != nil {
		return nil, errors.NewServiceError("NOTIFICATION_SEARCH_FAILED", "failed to search notifications", err)
	}

	return &dto.NotificationListResponse{
		Notifications: dto.ToNotificationResponses(notifications),
		Page:          paginationResult.Page,
		PageSize:      paginationResult.PageSize,
		TotalItems:    paginationResult.TotalItems,
		TotalPages:    paginationResult.TotalPages,
		HasNext:       paginationResult.HasNext,
		HasPrevious:   paginationResult.HasPrev,
	}, nil
}

// ============================================================================
// Health & Monitoring
// ============================================================================

// HealthCheck performs a health check
func (s *notificationService) HealthCheck(ctx context.Context) error {
	// Try to query the notification repository
	pagination := repository.PaginationParams{Page: 1, PageSize: 1}
	_, _, err := s.repos.Notification.FindByUserID(ctx, uuid.New(), pagination)

	if err != nil && !errors.IsNotFoundError(err) {
		return errors.NewServiceError("HEALTH_CHECK_FAILED", "notification service health check failed", err)
	}

	return nil
}

// GetServiceMetrics retrieves service metrics
func (s *notificationService) GetServiceMetrics(ctx context.Context) map[string]any {
	metrics := make(map[string]any)

	// These would be populated from actual metrics collection
	metrics["total_notifications_sent"] = 0
	metrics["notifications_pending"] = 0
	metrics["delivery_success_rate"] = 0.0

	return metrics
}

// ============================================================================
// Helper Methods
// ============================================================================

// sendViaChannels sends notification via configured channels
func (s *notificationService) sendViaChannels(ctx context.Context, notification *models.Notification) {
	for _, channel := range notification.Channels {
		switch channel {
		case models.NotificationChannelInApp:
			// In-app notifications are already stored in the database
			s.logger.Debug("in-app notification ready",
				"notification_id", notification.ID,
				"user_id", notification.UserID)

		case models.NotificationChannelEmail:
			// Send email notification
			s.sendEmailNotification(ctx, notification)

		case models.NotificationChannelSMS:
			// Send SMS notification
			s.sendSMSNotification(ctx, notification)

		case models.NotificationChannelPush:
			// Send push notification
			s.sendPushNotification(ctx, notification)
		}
	}
}

// sendEmailNotification sends email notification (placeholder)
func (s *notificationService) sendEmailNotification(ctx context.Context, notification *models.Notification) {
	// This would integrate with an email service provider
	s.logger.Info("email notification would be sent",
		"notification_id", notification.ID,
		"user_id", notification.UserID,
		"title", notification.Title)

	// Mark as sent via email
	// s.repos.Notification.MarkSentViaEmail(ctx, notification.ID)
}

// sendSMSNotification sends SMS notification (placeholder)
func (s *notificationService) sendSMSNotification(ctx context.Context, notification *models.Notification) {
	// This would integrate with an SMS service provider
	s.logger.Info("SMS notification would be sent",
		"notification_id", notification.ID,
		"user_id", notification.UserID,
		"message", notification.Message)

	// Mark as sent via SMS
	// s.repos.Notification.MarkSentViaSMS(ctx, notification.ID)
}

// sendPushNotification sends push notification (placeholder)
func (s *notificationService) sendPushNotification(ctx context.Context, notification *models.Notification) {
	// This would integrate with a push notification service
	s.logger.Info("push notification would be sent",
		"notification_id", notification.ID,
		"user_id", notification.UserID,
		"title", notification.Title)

	// Mark as sent via push
	// s.repos.Notification.MarkSentViaPush(ctx, notification.ID)
}
