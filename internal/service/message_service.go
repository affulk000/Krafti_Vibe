package service

import (
	"context"
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// MessageService defines the interface for message operations
type MessageService interface {
	// CRUD Operations
	SendMessage(ctx context.Context, tenantID, senderID uuid.UUID, req *dto.SendMessageRequest) (*dto.MessageResponse, error)
	GetMessage(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*dto.MessageResponse, error)
	UpdateMessage(ctx context.Context, id uuid.UUID, senderID uuid.UUID, req *dto.UpdateMessageRequest) (*dto.MessageResponse, error)
	DeleteMessage(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Conversation Management
	GetConversation(ctx context.Context, req *dto.ConversationRequest) (*dto.MessageListResponse, error)
	GetConversationByBooking(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID, page, pageSize int) (*dto.MessageListResponse, error)
	GetConversationList(ctx context.Context, userID uuid.UUID) (*dto.ConversationListResponse, error)

	// Message Queries
	ListMessages(ctx context.Context, filter *dto.MessageFilter) (*dto.MessageListResponse, error)
	ListUnreadMessages(ctx context.Context, userID uuid.UUID) ([]*dto.MessageResponse, error)
	ListSentMessages(ctx context.Context, senderID uuid.UUID, page, pageSize int) (*dto.MessageListResponse, error)
	ListReceivedMessages(ctx context.Context, receiverID uuid.UUID, page, pageSize int) (*dto.MessageListResponse, error)
	SearchMessages(ctx context.Context, userID uuid.UUID, query string, page, pageSize int) (*dto.MessageListResponse, error)

	// Thread Management
	GetReplies(ctx context.Context, parentMessageID uuid.UUID, userID uuid.UUID) ([]*dto.MessageResponse, error)
	GetThread(ctx context.Context, parentMessageID uuid.UUID, userID uuid.UUID, page, pageSize int) (*dto.MessageListResponse, error)

	// Status Operations
	MarkAsRead(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error
	MarkMultipleAsRead(ctx context.Context, messageIDs []uuid.UUID, userID uuid.UUID) error
	MarkConversationAsRead(ctx context.Context, userID, otherUserID uuid.UUID) error
	UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, userID uuid.UUID, status models.MessageStatus) error

	// Statistics & Analytics
	GetMessageStats(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) (*dto.MessageStatsResponse, error)
	GetConversationStats(ctx context.Context, userID1, userID2 uuid.UUID) (*dto.ConversationStatsResponse, error)
	CountUnreadMessages(ctx context.Context, userID uuid.UUID) (int64, error)

	// Utilities
	DeleteConversation(ctx context.Context, userID1, userID2 uuid.UUID) error
	BulkMarkAsDelivered(ctx context.Context, receiverID uuid.UUID) error
}

// messageService implements MessageService
type messageService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewMessageService creates a new message service
func NewMessageService(repos *repository.Repositories, logger log.AllLogger) MessageService {
	return &messageService{
		repos:  repos,
		logger: logger,
	}
}

// SendMessage sends a new message
func (s *messageService) SendMessage(ctx context.Context, tenantID, senderID uuid.UUID, req *dto.SendMessageRequest) (*dto.MessageResponse, error) {
	// Verify sender exists
	sender, err := s.repos.User.GetByID(ctx, senderID)
	if err != nil {
		s.logger.Error("failed to get sender", "sender_id", senderID, "error", err)
		return nil, errors.NewNotFoundError("sender")
	}

	// Verify receiver exists
	receiver, err := s.repos.User.GetByID(ctx, req.ReceiverID)
	if err != nil {
		s.logger.Error("failed to get receiver", "receiver_id", req.ReceiverID, "error", err)
		return nil, errors.NewNotFoundError("receiver")
	}

	// Verify users belong to the same tenant (unless one is platform admin)
	if !sender.IsPlatformAdmin() && !receiver.IsPlatformAdmin() {
		if sender.TenantID == nil || receiver.TenantID == nil || *sender.TenantID != *receiver.TenantID {
			return nil, errors.NewValidationError("Users must belong to the same tenant")
		}
	}

	// Verify booking if provided
	if req.BookingID != nil {
		booking, err := s.repos.Booking.GetByID(ctx, *req.BookingID)
		if err != nil {
			return nil, errors.NewNotFoundError("booking")
		}
		if booking.TenantID != tenantID {
			return nil, errors.NewValidationError("Booking does not belong to tenant")
		}
	}

	// Verify parent message if provided
	if req.ParentMessageID != nil {
		parentMsg, err := s.repos.Message.GetByID(ctx, *req.ParentMessageID)
		if err != nil {
			return nil, errors.NewNotFoundError("parent message")
		}
		// Ensure parent message is part of the same conversation
		if (parentMsg.SenderID != senderID && parentMsg.ReceiverID != senderID) ||
			(parentMsg.SenderID != req.ReceiverID && parentMsg.ReceiverID != req.ReceiverID) {
			return nil, errors.NewValidationError("Parent message not part of this conversation")
		}
	}

	// Create message
	message := &models.Message{
		TenantID:        tenantID,
		SenderID:        senderID,
		ReceiverID:      req.ReceiverID,
		BookingID:       req.BookingID,
		Type:            req.Type,
		Content:         req.Content,
		FileURL:         req.FileURL,
		Status:          models.MessageStatusSent,
		ParentMessageID: req.ParentMessageID,
		Metadata:        req.Metadata,
	}

	if err := s.repos.Message.Create(ctx, message); err != nil {
		s.logger.Error("failed to create message", "error", err)
		return nil, errors.NewRepositoryError("CREATE_FAILED", "Failed to send message", err)
	}

	// Load relationships
	message.Sender = sender
	message.Receiver = receiver

	if req.BookingID != nil {
		booking, _ := s.repos.Booking.GetByID(ctx, *req.BookingID)
		message.Booking = booking
	}

	s.logger.Info("message sent", "message_id", message.ID, "sender_id", senderID, "receiver_id", req.ReceiverID)
	return dto.ToMessageResponse(message), nil
}

// GetMessage retrieves a message by ID
func (s *messageService) GetMessage(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*dto.MessageResponse, error) {
	message, err := s.repos.Message.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get message", "id", id, "error", err)
		return nil, errors.NewNotFoundError("message")
	}

	// Verify user is sender or receiver
	if message.SenderID != userID && message.ReceiverID != userID {
		return nil, errors.NewNotFoundError("message")
	}

	// Load relationships
	if message.SenderID != uuid.Nil {
		sender, err := s.repos.User.GetByID(ctx, message.SenderID)
		if err == nil {
			message.Sender = sender
		}
	}

	if message.ReceiverID != uuid.Nil {
		receiver, err := s.repos.User.GetByID(ctx, message.ReceiverID)
		if err == nil {
			message.Receiver = receiver
		}
	}

	if message.BookingID != nil {
		booking, err := s.repos.Booking.GetByID(ctx, *message.BookingID)
		if err == nil {
			message.Booking = booking
		}
	}

	return dto.ToMessageResponse(message), nil
}

// UpdateMessage updates a message
func (s *messageService) UpdateMessage(ctx context.Context, id uuid.UUID, senderID uuid.UUID, req *dto.UpdateMessageRequest) (*dto.MessageResponse, error) {
	// Get existing message
	message, err := s.repos.Message.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get message", "id", id, "error", err)
		return nil, errors.NewNotFoundError("message")
	}

	// Verify user is the sender
	if message.SenderID != senderID {
		return nil, errors.NewValidationError("Only the sender can update a message")
	}

	// Don't allow updating read messages
	if message.Status == models.MessageStatusRead {
		return nil, errors.NewValidationError("Cannot update read messages")
	}

	// Update fields
	if req.Content != nil {
		message.Content = *req.Content
	}
	if req.Metadata != nil {
		message.Metadata = req.Metadata
	}

	// Save message
	if err := s.repos.Message.Update(ctx, message); err != nil {
		s.logger.Error("failed to update message", "id", id, "error", err)
		return nil, errors.NewRepositoryError("UPDATE_FAILED", "Failed to update message", err)
	}

	s.logger.Info("message updated", "id", id)

	// Get updated message
	updated, err := s.repos.Message.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewServiceError("FIND_FAILED", "Failed to retrieve updated message", err)
	}

	return dto.ToMessageResponse(updated), nil
}

// DeleteMessage deletes a message
func (s *messageService) DeleteMessage(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get message
	message, err := s.repos.Message.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get message", "id", id, "error", err)
		return errors.NewNotFoundError("message")
	}

	// Verify user is sender or receiver
	if message.SenderID != userID && message.ReceiverID != userID {
		return errors.NewNotFoundError("message")
	}

	if err := s.repos.Message.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete message", "id", id, "error", err)
		return errors.NewRepositoryError("DELETE_FAILED", "Failed to delete message", err)
	}

	s.logger.Info("message deleted", "id", id)
	return nil
}

// GetConversation retrieves conversation between two users
func (s *messageService) GetConversation(ctx context.Context, req *dto.ConversationRequest) (*dto.MessageListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     max(req.Page, 1),
		PageSize: min(max(req.PageSize, 1), 100),
	}

	messages, paginationResult, err := s.repos.Message.FindConversation(ctx, req.UserID1, req.UserID2, pagination)
	if err != nil {
		s.logger.Error("failed to get conversation", "user1", req.UserID1, "user2", req.UserID2, "error", err)
		return nil, errors.NewServiceError("FIND_FAILED", "Failed to get conversation", err)
	}

	return &dto.MessageListResponse{
		Messages:    dto.ToMessageResponses(messages),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetConversationByBooking retrieves conversation for a booking
func (s *messageService) GetConversationByBooking(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID, page, pageSize int) (*dto.MessageListResponse, error) {
	// Verify booking exists and user has access
	booking, err := s.repos.Booking.GetByID(ctx, bookingID)
	if err != nil {
		return nil, errors.NewNotFoundError("booking")
	}

	// Verify user is involved in the booking
	if booking.CustomerID != userID && booking.ArtisanID != userID {
		return nil, errors.NewValidationError("User is not involved in this booking")
	}

	pagination := repository.PaginationParams{
		Page:     max(page, 1),
		PageSize: min(max(pageSize, 1), 100),
	}

	messages, paginationResult, err := s.repos.Message.FindConversationByBooking(ctx, bookingID, pagination)
	if err != nil {
		s.logger.Error("failed to get booking conversation", "booking_id", bookingID, "error", err)
		return nil, errors.NewServiceError("FIND_FAILED", "Failed to get booking conversation", err)
	}

	return &dto.MessageListResponse{
		Messages:    dto.ToMessageResponses(messages),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetConversationList retrieves all conversations for a user
func (s *messageService) GetConversationList(ctx context.Context, userID uuid.UUID) (*dto.ConversationListResponse, error) {
	conversations, err := s.repos.Message.GetConversationList(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get conversation list", "user_id", userID, "error", err)
		return nil, errors.NewServiceError("FIND_FAILED", "Failed to get conversation list", err)
	}

	// Convert to response DTOs
	responses := make([]*dto.ConversationSummaryResponse, len(conversations))
	var totalUnread int64
	for i, conv := range conversations {
		responses[i] = &dto.ConversationSummaryResponse{
			OtherUserID:       conv.OtherUserID,
			OtherUserName:     conv.OtherUserName,
			LastMessage:       conv.LastMessage,
			LastMessageAt:     conv.LastMessageAt,
			LastMessageType:   conv.LastMessageType,
			LastMessageStatus: conv.LastMessageStatus,
			UnreadCount:       conv.UnreadCount,
			TotalMessages:     conv.TotalMessages,
			BookingID:         conv.BookingID,
		}
		totalUnread += conv.UnreadCount
	}

	return &dto.ConversationListResponse{
		Conversations: responses,
		TotalUnread:   totalUnread,
	}, nil
}

// ListMessages lists messages with filtering
func (s *messageService) ListMessages(ctx context.Context, filter *dto.MessageFilter) (*dto.MessageListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     max(filter.Page, 1),
		PageSize: max(filter.PageSize, 1),
	}
	pagination.PageSize = min(pagination.PageSize, 100)

	var messages []*models.Message
	var paginationResult repository.PaginationResult
	var err error

	// Use appropriate repository method based on filters
	if filter.SearchQuery != "" && filter.UserID != nil {
		messages, paginationResult, err = s.repos.Message.SearchMessages(ctx, *filter.UserID, filter.SearchQuery, pagination)
	} else if filter.BookingID != nil {
		messages, paginationResult, err = s.repos.Message.FindConversationByBooking(ctx, *filter.BookingID, pagination)
	} else if filter.Status != nil && filter.UserID != nil {
		messages, paginationResult, err = s.repos.Message.FindMessagesByStatus(ctx, *filter.UserID, *filter.Status, pagination)
	} else if filter.Type != nil {
		messages, paginationResult, err = s.repos.Message.FindMessagesByType(ctx, filter.TenantID, *filter.Type, pagination)
	} else if filter.SenderID != nil {
		messages, paginationResult, err = s.repos.Message.FindBySenderID(ctx, *filter.SenderID, pagination)
	} else if filter.ReceiverID != nil {
		messages, paginationResult, err = s.repos.Message.FindByReceiverID(ctx, *filter.ReceiverID, pagination)
	} else if filter.StartDate != nil && filter.EndDate != nil && filter.UserID != nil {
		messages, paginationResult, err = s.repos.Message.FindMessagesByDateRange(ctx, *filter.UserID, *filter.StartDate, *filter.EndDate, pagination)
	} else {
		messages, paginationResult, err = s.repos.Message.FindByTenantID(ctx, filter.TenantID, pagination)
	}

	if err != nil {
		s.logger.Error("failed to list messages", "error", err)
		return nil, errors.NewRepositoryError("LIST_FAILED", "Failed to list messages", err)
	}

	return &dto.MessageListResponse{
		Messages:    dto.ToMessageResponses(messages),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ListUnreadMessages lists unread messages for a user
func (s *messageService) ListUnreadMessages(ctx context.Context, userID uuid.UUID) ([]*dto.MessageResponse, error) {
	messages, err := s.repos.Message.FindUnreadMessages(ctx, userID)
	if err != nil {
		s.logger.Error("failed to list unread messages", "user_id", userID, "error", err)
		return nil, errors.NewServiceError("FIND_FAILED", "Failed to list unread messages", err)
	}

	return dto.ToMessageResponses(messages), nil
}

// ListSentMessages lists messages sent by a user
func (s *messageService) ListSentMessages(ctx context.Context, senderID uuid.UUID, page, pageSize int) (*dto.MessageListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     max(page, 1),
		PageSize: min(max(pageSize, 1), 100),
	}

	messages, paginationResult, err := s.repos.Message.FindBySenderID(ctx, senderID, pagination)
	if err != nil {
		s.logger.Error("failed to list sent messages", "sender_id", senderID, "error", err)
		return nil, errors.NewServiceError("FIND_FAILED", "Failed to list sent messages", err)
	}

	return &dto.MessageListResponse{
		Messages:    dto.ToMessageResponses(messages),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ListReceivedMessages lists messages received by a user
func (s *messageService) ListReceivedMessages(ctx context.Context, receiverID uuid.UUID, page, pageSize int) (*dto.MessageListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     max(page, 1),
		PageSize: min(max(pageSize, 1), 100),
	}

	messages, paginationResult, err := s.repos.Message.FindByReceiverID(ctx, receiverID, pagination)
	if err != nil {
		s.logger.Error("failed to list received messages", "receiver_id", receiverID, "error", err)
		return nil, errors.NewServiceError("FIND_FAILED", "Failed to list received messages", err)
	}

	return &dto.MessageListResponse{
		Messages:    dto.ToMessageResponses(messages),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// SearchMessages searches messages by content
func (s *messageService) SearchMessages(ctx context.Context, userID uuid.UUID, query string, page, pageSize int) (*dto.MessageListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     max(page, 1),
		PageSize: min(max(pageSize, 1), 100),
	}

	messages, paginationResult, err := s.repos.Message.SearchMessages(ctx, userID, query, pagination)
	if err != nil {
		s.logger.Error("failed to search messages", "query", query, "error", err)
		return nil, errors.NewServiceError("SEARCH_FAILED", "Failed to search messages", err)
	}

	return &dto.MessageListResponse{
		Messages:    dto.ToMessageResponses(messages),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetReplies retrieves all replies to a message
func (s *messageService) GetReplies(ctx context.Context, parentMessageID uuid.UUID, userID uuid.UUID) ([]*dto.MessageResponse, error) {
	// Verify parent message exists and user has access
	parentMsg, err := s.repos.Message.GetByID(ctx, parentMessageID)
	if err != nil {
		return nil, errors.NewNotFoundError("parent message")
	}

	// Verify user is involved in the conversation
	if parentMsg.SenderID != userID && parentMsg.ReceiverID != userID {
		return nil, errors.NewValidationError("User is not involved in this conversation")
	}

	messages, err := s.repos.Message.FindReplies(ctx, parentMessageID)
	if err != nil {
		s.logger.Error("failed to get replies", "parent_message_id", parentMessageID, "error", err)
		return nil, errors.NewServiceError("FIND_FAILED", "Failed to get replies", err)
	}

	return dto.ToMessageResponses(messages), nil
}

// GetThread retrieves message thread with pagination
func (s *messageService) GetThread(ctx context.Context, parentMessageID uuid.UUID, userID uuid.UUID, page, pageSize int) (*dto.MessageListResponse, error) {
	// Verify parent message exists and user has access
	parentMsg, err := s.repos.Message.GetByID(ctx, parentMessageID)
	if err != nil {
		return nil, errors.NewNotFoundError("parent message")
	}

	// Verify user is involved in the conversation
	if parentMsg.SenderID != userID && parentMsg.ReceiverID != userID {
		return nil, errors.NewValidationError("User is not involved in this conversation")
	}

	pagination := repository.PaginationParams{
		Page:     max(page, 1),
		PageSize: min(max(pageSize, 1), 100),
	}

	messages, paginationResult, err := s.repos.Message.GetThreadMessages(ctx, parentMessageID, pagination)
	if err != nil {
		s.logger.Error("failed to get thread", "parent_message_id", parentMessageID, "error", err)
		return nil, errors.NewServiceError("FIND_FAILED", "Failed to get thread", err)
	}

	return &dto.MessageListResponse{
		Messages:    dto.ToMessageResponses(messages),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// MarkAsRead marks a message as read
func (s *messageService) MarkAsRead(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error {
	// Verify message exists and user is receiver
	message, err := s.repos.Message.GetByID(ctx, messageID)
	if err != nil {
		return errors.NewNotFoundError("message")
	}

	if message.ReceiverID != userID {
		return errors.NewValidationError("Only the receiver can mark a message as read")
	}

	if err := s.repos.Message.MarkAsRead(ctx, messageID); err != nil {
		s.logger.Error("failed to mark message as read", "message_id", messageID, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "Failed to mark message as read", err)
	}

	s.logger.Info("message marked as read", "message_id", messageID)
	return nil
}

// MarkMultipleAsRead marks multiple messages as read
func (s *messageService) MarkMultipleAsRead(ctx context.Context, messageIDs []uuid.UUID, userID uuid.UUID) error {
	// TODO: Verify all messages belong to the user as receiver
	if err := s.repos.Message.MarkMultipleAsRead(ctx, messageIDs); err != nil {
		s.logger.Error("failed to mark messages as read", "count", len(messageIDs), "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "Failed to mark messages as read", err)
	}

	s.logger.Info("messages marked as read", "count", len(messageIDs))
	return nil
}

// MarkConversationAsRead marks all messages in a conversation as read
func (s *messageService) MarkConversationAsRead(ctx context.Context, userID, otherUserID uuid.UUID) error {
	if err := s.repos.Message.MarkConversationAsRead(ctx, userID, otherUserID); err != nil {
		s.logger.Error("failed to mark conversation as read", "user_id", userID, "other_user_id", otherUserID, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "Failed to mark conversation as read", err)
	}

	s.logger.Info("conversation marked as read", "user_id", userID, "other_user_id", otherUserID)
	return nil
}

// UpdateMessageStatus updates message status
func (s *messageService) UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, userID uuid.UUID, status models.MessageStatus) error {
	// Verify message exists and user has access
	message, err := s.repos.Message.GetByID(ctx, messageID)
	if err != nil {
		return errors.NewNotFoundError("message")
	}

	if message.ReceiverID != userID {
		return errors.NewValidationError("Only the receiver can update message status")
	}

	if err := s.repos.Message.UpdateMessageStatus(ctx, messageID, status); err != nil {
		s.logger.Error("failed to update message status", "message_id", messageID, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "Failed to update message status", err)
	}

	s.logger.Info("message status updated", "message_id", messageID, "status", status)
	return nil
}

// GetMessageStats retrieves message statistics
func (s *messageService) GetMessageStats(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) (*dto.MessageStatsResponse, error) {
	stats, err := s.repos.Message.GetMessageStats(ctx, tenantID, userID)
	if err != nil {
		s.logger.Error("failed to get message stats", "tenant_id", tenantID, "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get message stats", err)
	}

	return &dto.MessageStatsResponse{
		UserID:           userID,
		TenantID:         tenantID,
		TotalMessages:    stats.TotalMessages,
		SentMessages:     stats.SentMessages,
		ReceivedMessages: stats.ReceivedMessages,
		UnreadMessages:   stats.UnreadMessages,
		MessagesByType:   stats.MessagesByType,
		MessagesByStatus: stats.MessagesByStatus,
		AveragePerDay:    stats.AveragePerDay,
		LastMessageAt:    stats.LastMessageAt,
	}, nil
}

// GetConversationStats retrieves conversation statistics
func (s *messageService) GetConversationStats(ctx context.Context, userID1, userID2 uuid.UUID) (*dto.ConversationStatsResponse, error) {
	stats, err := s.repos.Message.GetConversationStats(ctx, userID1, userID2)
	if err != nil {
		s.logger.Error("failed to get conversation stats", "user1", userID1, "user2", userID2, "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get conversation stats", err)
	}

	// Format average response time
	avgResponseTime := formatDuration(stats.AverageResponseTime)
	avgResponseSec := stats.AverageResponseTime.Seconds()

	return &dto.ConversationStatsResponse{
		UserID1:             userID1,
		UserID2:             userID2,
		TotalMessages:       stats.TotalMessages,
		User1SentCount:      stats.User1SentCount,
		User2SentCount:      stats.User2SentCount,
		UnreadCount:         stats.UnreadCount,
		FirstMessageAt:      stats.FirstMessageAt,
		LastMessageAt:       stats.LastMessageAt,
		MessagesByType:      stats.MessagesByType,
		AverageResponseTime: avgResponseTime,
		AverageResponseSec:  avgResponseSec,
	}, nil
}

// CountUnreadMessages counts unread messages for a user
func (s *messageService) CountUnreadMessages(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.repos.Message.CountUnreadMessages(ctx, userID)
	if err != nil {
		s.logger.Error("failed to count unread messages", "user_id", userID, "error", err)
		return 0, errors.NewServiceError("COUNT_FAILED", "Failed to count unread messages", err)
	}

	return count, nil
}

// DeleteConversation deletes all messages in a conversation
func (s *messageService) DeleteConversation(ctx context.Context, userID1, userID2 uuid.UUID) error {
	if err := s.repos.Message.DeleteConversation(ctx, userID1, userID2); err != nil {
		s.logger.Error("failed to delete conversation", "user1", userID1, "user2", userID2, "error", err)
		return errors.NewRepositoryError("DELETE_FAILED", "Failed to delete conversation", err)
	}

	s.logger.Info("conversation deleted", "user1", userID1, "user2", userID2)
	return nil
}

// BulkMarkAsDelivered marks all sent messages to a receiver as delivered
func (s *messageService) BulkMarkAsDelivered(ctx context.Context, receiverID uuid.UUID) error {
	if err := s.repos.Message.BulkMarkAsDelivered(ctx, receiverID); err != nil {
		s.logger.Error("failed to bulk mark as delivered", "receiver_id", receiverID, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "Failed to mark messages as delivered", err)
	}

	s.logger.Info("messages marked as delivered", "receiver_id", receiverID)
	return nil
}

// Helper function to format duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f seconds", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1f minutes", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1f hours", d.Hours())
	}
	return fmt.Sprintf("%.1f days", d.Hours()/24)
}
