package handler

import (
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MessageHandler handles HTTP requests for messaging operations
type MessageHandler struct {
	messageService service.MessageService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageService service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// SendMessage sends a message
func (h *MessageHandler) SendMessage(c *fiber.Ctx) error {
	var req dto.SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	message, err := h.messageService.SendMessage(c.Context(), authCtx.TenantID, authCtx.UserID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, message, "Message sent successfully")
}

// GetConversation retrieves messages in a conversation
func (h *MessageHandler) GetConversation(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)

	otherUserID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 50)

	req := &dto.ConversationRequest{
		UserID1:  authCtx.UserID,
		UserID2:  otherUserID,
		Page:     page,
		PageSize: pageSize,
	}

	messages, err := h.messageService.GetConversation(c.Context(), req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, messages)
}

// GetUserConversations retrieves all conversations for a user
func (h *MessageHandler) GetUserConversations(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)

	conversations, err := h.messageService.GetConversationList(c.Context(), authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, conversations)
}

// MarkAsRead marks messages as read
func (h *MessageHandler) MarkAsRead(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)

	messageID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid message ID", err)
	}

	if err := h.messageService.MarkAsRead(c.Context(), messageID, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Message marked as read")
}

// GetUnreadCount gets count of unread messages
func (h *MessageHandler) GetUnreadCount(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)

	count, err := h.messageService.CountUnreadMessages(c.Context(), authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, map[string]any{"count": count})
}
