package handler

import (
	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// NotificationHandler handles HTTP requests for notification operations
type NotificationHandler struct {
	notificationService service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// SendNotification sends a notification
func (h *NotificationHandler) SendNotification(c *fiber.Ctx) error {
	var req dto.CreateNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	notification, err := h.notificationService.CreateNotification(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, notification, "Notification sent successfully")
}

// GetUserNotifications retrieves notifications for a user
func (h *NotificationHandler) GetUserNotifications(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)
	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 20)

	filter := dto.NotificationFilter{
		TenantID: authCtx.TenantID,
		UserID:   authCtx.UserID,
		Page:     page,
		PageSize: pageSize,
	}

	notifications, err := h.notificationService.ListNotifications(c.Context(), filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, notifications)
}

// MarkAsRead marks a notification as read
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid notification ID", err)
	}

	if err := h.notificationService.MarkAsRead(c.Context(), notificationID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Notification marked as read")
}

// MarkAllAsRead marks all notifications as read for a user
func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)

	if err := h.notificationService.MarkAllAsRead(c.Context(), authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "All notifications marked as read")
}

// GetUnreadCount gets count of unread notifications
func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)

	count, err := h.notificationService.GetUnreadCount(c.Context(), authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, count)
}

// GetNotification godoc
// @Summary Get notification by ID
// @Description Get detailed notification information by ID
// @Tags notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} dto.NotificationResponse
// @Failure 404 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /notifications/{id} [get]
func (h *NotificationHandler) GetNotification(c *fiber.Ctx) error {
	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid notification ID", err)
	}

	notification, err := h.notificationService.GetNotification(c.Context(), notificationID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, notification)
}

// ListNotifications godoc
// @Summary List notifications
// @Description List notifications with filters
// @Tags notifications
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param is_read query bool false "Filter by read status"
// @Param notification_type query string false "Filter by notification type"
// @Success 200 {object} dto.NotificationListResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /notifications [get]
func (h *NotificationHandler) ListNotifications(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)
	page, pageSize := ParsePagination(c)

	var isRead *bool
	if c.Query("is_read") != "" {
		val := c.QueryBool("is_read")
		isRead = &val
	}

	var notifType *models.NotificationType
	if typeStr := c.Query("notification_type"); typeStr != "" {
		nt := models.NotificationType(typeStr)
		notifType = &nt
	}

	filter := dto.NotificationFilter{
		TenantID: authCtx.TenantID,
		UserID:   authCtx.UserID,
		Page:     page,
		PageSize: pageSize,
		IsRead:   isRead,
		Type:     notifType,
	}

	notifications, err := h.notificationService.ListNotifications(c.Context(), filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, notifications)
}

// DeleteNotification godoc
// @Summary Delete notification
// @Description Delete a notification by ID
// @Tags notifications
// @Param id path string true "Notification ID"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *fiber.Ctx) error {
	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid notification ID", err)
	}

	if err := h.notificationService.DeleteNotification(c.Context(), notificationID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, fiber.Map{
		"message": "Notification deleted successfully",
	})
}

// MarkMultipleAsRead godoc
// @Summary Mark multiple notifications as read
// @Description Mark multiple notifications as read
// @Tags notifications
// @Accept json
// @Produce json
// @Param ids body dto.MarkAsReadRequest true "Notification IDs"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /notifications/mark-multiple-read [post]
func (h *NotificationHandler) MarkMultipleAsRead(c *fiber.Ctx) error {
	var req struct {
		NotificationIDs []uuid.UUID `json:"notification_ids"`
	}

	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if len(req.NotificationIDs) == 0 {
		return NewErrorResponse(c, fiber.StatusBadRequest, "EMPTY_IDS", "No notification IDs provided", nil)
	}

	if err := h.notificationService.MarkMultipleAsRead(c.Context(), req.NotificationIDs); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, fiber.Map{
		"message": "Notifications marked as read",
		"count":   len(req.NotificationIDs),
	})
}

// SendBulkNotification godoc
// @Summary Send bulk notification
// @Description Send notification to multiple users
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification body dto.BulkNotificationRequest true "Bulk notification data"
// @Success 201 {object} dto.BulkNotificationResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /notifications/bulk [post]
func (h *NotificationHandler) SendBulkNotification(c *fiber.Ctx) error {
	var req dto.BulkNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	result, err := h.notificationService.SendBulkNotification(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, result, "Bulk notification sent successfully")
}

// DeleteReadNotifications godoc
// @Summary Delete read notifications
// @Description Delete all read notifications for the current user older than specified days
// @Tags notifications
// @Param older_than_days query int false "Delete read notifications older than days (default: 30)"
// @Success 200 {object} handler.SuccessResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /notifications/cleanup/read [delete]
func (h *NotificationHandler) DeleteReadNotifications(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)
	olderThanDays := c.QueryInt("older_than_days", 30)

	count, err := h.notificationService.DeleteReadNotifications(c.Context(), authCtx.UserID, olderThanDays)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, fiber.Map{
		"message":       "Read notifications deleted successfully",
		"deleted_count": count,
	})
}

// DeleteOldNotifications godoc
// @Summary Delete old notifications
// @Description Delete all notifications older than specified days (admin only)
// @Tags notifications
// @Param older_than_days query int false "Delete notifications older than days (default: 90)"
// @Success 200 {object} handler.SuccessResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /notifications/cleanup/old [delete]
func (h *NotificationHandler) DeleteOldNotifications(c *fiber.Ctx) error {
	olderThanDays := c.QueryInt("older_than_days", 90)

	count, err := h.notificationService.DeleteOldNotifications(c.Context(), olderThanDays)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, fiber.Map{
		"message":       "Old notifications deleted successfully",
		"deleted_count": count,
	})
}
