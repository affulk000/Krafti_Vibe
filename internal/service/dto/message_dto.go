package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Message Request DTOs
// ============================================================================

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	ReceiverID      uuid.UUID         `json:"receiver_id" validate:"required"`
	Type            models.MessageType `json:"type" validate:"required,oneof=text image file system"`
	Content         string            `json:"content" validate:"required"`
	FileURL         string            `json:"file_url,omitempty" validate:"omitempty,url"`
	BookingID       *uuid.UUID        `json:"booking_id,omitempty"`
	ParentMessageID *uuid.UUID        `json:"parent_message_id,omitempty"`
	Metadata        map[string]any    `json:"metadata,omitempty"`
}

// UpdateMessageRequest represents a request to update a message
type UpdateMessageRequest struct {
	Content  *string        `json:"content,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// MessageFilter represents filters for message queries
type MessageFilter struct {
	TenantID   uuid.UUID              `json:"tenant_id" validate:"required"`
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	SenderID   *uuid.UUID             `json:"sender_id,omitempty"`
	ReceiverID *uuid.UUID             `json:"receiver_id,omitempty"`
	BookingID  *uuid.UUID             `json:"booking_id,omitempty"`
	Type       *models.MessageType    `json:"type,omitempty"`
	Status     *models.MessageStatus  `json:"status,omitempty"`
	IsUnread   *bool                  `json:"is_unread,omitempty"`
	StartDate  *time.Time             `json:"start_date,omitempty"`
	EndDate    *time.Time             `json:"end_date,omitempty"`
	SearchQuery string                `json:"search_query,omitempty"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
}

// ConversationRequest represents a request to get a conversation
type ConversationRequest struct {
	UserID1  uuid.UUID `json:"user_id_1" validate:"required"`
	UserID2  uuid.UUID `json:"user_id_2" validate:"required"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}

// ============================================================================
// Message Response DTOs
// ============================================================================

// MessageResponse represents a message
type MessageResponse struct {
	ID              uuid.UUID             `json:"id"`
	TenantID        uuid.UUID             `json:"tenant_id"`
	SenderID        uuid.UUID             `json:"sender_id"`
	ReceiverID      uuid.UUID             `json:"receiver_id"`
	BookingID       *uuid.UUID            `json:"booking_id,omitempty"`
	Type            models.MessageType    `json:"type"`
	Content         string                `json:"content"`
	FileURL         string                `json:"file_url,omitempty"`
	Status          models.MessageStatus  `json:"status"`
	ReadAt          *time.Time            `json:"read_at,omitempty"`
	ParentMessageID *uuid.UUID            `json:"parent_message_id,omitempty"`
	Metadata        models.JSONB          `json:"metadata,omitempty"`
	Sender          *UserSummary          `json:"sender,omitempty"`
	Receiver        *UserSummary          `json:"receiver,omitempty"`
	Booking         *BookingSummary       `json:"booking,omitempty"`
	IsUnread        bool                  `json:"is_unread"`
	ReplyCount      int                   `json:"reply_count,omitempty"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

// MessageListResponse represents a paginated list of messages
type MessageListResponse struct {
	Messages    []*MessageResponse `json:"messages"`
	Page        int                `json:"page"`
	PageSize    int                `json:"page_size"`
	TotalItems  int64              `json:"total_items"`
	TotalPages  int                `json:"total_pages"`
	HasNext     bool               `json:"has_next"`
	HasPrevious bool               `json:"has_previous"`
}

// ConversationSummaryResponse represents a conversation overview
type ConversationSummaryResponse struct {
	OtherUserID       uuid.UUID            `json:"other_user_id"`
	OtherUserName     string               `json:"other_user_name"`
	OtherUserAvatar   string               `json:"other_user_avatar,omitempty"`
	LastMessage       string               `json:"last_message"`
	LastMessageAt     time.Time            `json:"last_message_at"`
	LastMessageType   models.MessageType   `json:"last_message_type"`
	LastMessageStatus models.MessageStatus `json:"last_message_status"`
	UnreadCount       int64                `json:"unread_count"`
	TotalMessages     int64                `json:"total_messages"`
	BookingID         *uuid.UUID           `json:"booking_id,omitempty"`
}

// ConversationListResponse represents a list of conversations
type ConversationListResponse struct {
	Conversations []*ConversationSummaryResponse `json:"conversations"`
	TotalUnread   int64                          `json:"total_unread"`
}

// MessageStatsResponse represents message statistics
type MessageStatsResponse struct {
	UserID           *uuid.UUID                         `json:"user_id,omitempty"`
	TenantID         uuid.UUID                          `json:"tenant_id"`
	TotalMessages    int64                              `json:"total_messages"`
	SentMessages     int64                              `json:"sent_messages"`
	ReceivedMessages int64                              `json:"received_messages"`
	UnreadMessages   int64                              `json:"unread_messages"`
	MessagesByType   map[models.MessageType]int64       `json:"messages_by_type"`
	MessagesByStatus map[models.MessageStatus]int64     `json:"messages_by_status"`
	AveragePerDay    float64                            `json:"average_per_day"`
	LastMessageAt    *time.Time                         `json:"last_message_at,omitempty"`
}

// ConversationStatsResponse represents statistics for a specific conversation
type ConversationStatsResponse struct {
	UserID1             uuid.UUID                      `json:"user_id_1"`
	UserID2             uuid.UUID                      `json:"user_id_2"`
	TotalMessages       int64                          `json:"total_messages"`
	User1SentCount      int64                          `json:"user1_sent_count"`
	User2SentCount      int64                          `json:"user2_sent_count"`
	UnreadCount         int64                          `json:"unread_count"`
	FirstMessageAt      *time.Time                     `json:"first_message_at,omitempty"`
	LastMessageAt       *time.Time                     `json:"last_message_at,omitempty"`
	MessagesByType      map[models.MessageType]int64   `json:"messages_by_type"`
	AverageResponseTime string                         `json:"average_response_time"` // Human readable format
	AverageResponseSec  float64                        `json:"average_response_seconds"`
}

// BookingSummary represents a minimal booking summary for messages
type BookingSummary struct {
	ID     uuid.UUID            `json:"id"`
	Status models.BookingStatus `json:"status"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToMessageResponse converts a Message model to MessageResponse DTO
func ToMessageResponse(message *models.Message) *MessageResponse {
	if message == nil {
		return nil
	}

	resp := &MessageResponse{
		ID:              message.ID,
		TenantID:        message.TenantID,
		SenderID:        message.SenderID,
		ReceiverID:      message.ReceiverID,
		BookingID:       message.BookingID,
		Type:            message.Type,
		Content:         message.Content,
		FileURL:         message.FileURL,
		Status:          message.Status,
		ReadAt:          message.ReadAt,
		ParentMessageID: message.ParentMessageID,
		Metadata:        message.Metadata,
		IsUnread:        message.IsUnread(),
		ReplyCount:      len(message.Replies),
		CreatedAt:       message.CreatedAt,
		UpdatedAt:       message.UpdatedAt,
	}

	// Add sender if available
	if message.Sender != nil {
		resp.Sender = &UserSummary{
			ID:        message.Sender.ID,
			FirstName: message.Sender.FirstName,
			LastName:  message.Sender.LastName,
			Email:     message.Sender.Email,
			AvatarURL: message.Sender.AvatarURL,
		}
	}

	// Add receiver if available
	if message.Receiver != nil {
		resp.Receiver = &UserSummary{
			ID:        message.Receiver.ID,
			FirstName: message.Receiver.FirstName,
			LastName:  message.Receiver.LastName,
			Email:     message.Receiver.Email,
			AvatarURL: message.Receiver.AvatarURL,
		}
	}

	// Add booking if available
	if message.Booking != nil {
		resp.Booking = &BookingSummary{
			ID:     message.Booking.ID,
			Status: message.Booking.Status,
		}
	}

	return resp
}

// ToMessageResponses converts multiple Message models to DTOs
func ToMessageResponses(messages []*models.Message) []*MessageResponse {
	if messages == nil {
		return nil
	}

	responses := make([]*MessageResponse, len(messages))
	for i, message := range messages {
		responses[i] = ToMessageResponse(message)
	}
	return responses
}
