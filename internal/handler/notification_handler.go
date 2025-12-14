package handler

import (
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

	return NewSuccessResponse(c, map[string]interface{}{"count": count})
}
