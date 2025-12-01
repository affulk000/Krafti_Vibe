package repository

import (
	"context"
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageRepository defines the interface for message repository operations
type MessageRepository interface {
	BaseRepository[models.Message]

	// Conversation Management
	FindConversation(ctx context.Context, userID1, userID2 uuid.UUID, pagination PaginationParams) ([]*models.Message, PaginationResult, error)
	FindConversationByBooking(ctx context.Context, bookingID uuid.UUID, pagination PaginationParams) ([]*models.Message, PaginationResult, error)
	GetConversationList(ctx context.Context, userID uuid.UUID) ([]ConversationSummary, error)

	// Message Queries
	FindByReceiverID(ctx context.Context, receiverID uuid.UUID, pagination PaginationParams) ([]*models.Message, PaginationResult, error)
	FindBySenderID(ctx context.Context, senderID uuid.UUID, pagination PaginationParams) ([]*models.Message, PaginationResult, error)
	FindUnreadMessages(ctx context.Context, receiverID uuid.UUID) ([]*models.Message, error)
	FindMessagesByStatus(ctx context.Context, userID uuid.UUID, status models.MessageStatus, pagination PaginationParams) ([]*models.Message, PaginationResult, error)
	FindMessagesByType(ctx context.Context, tenantID uuid.UUID, messageType models.MessageType, pagination PaginationParams) ([]*models.Message, PaginationResult, error)

	// Thread Management
	FindReplies(ctx context.Context, parentMessageID uuid.UUID) ([]*models.Message, error)
	GetThreadMessages(ctx context.Context, parentMessageID uuid.UUID, pagination PaginationParams) ([]*models.Message, PaginationResult, error)

	// Status Operations
	MarkAsRead(ctx context.Context, messageID uuid.UUID) error
	MarkMultipleAsRead(ctx context.Context, messageIDs []uuid.UUID) error
	MarkConversationAsRead(ctx context.Context, receiverID, senderID uuid.UUID) error
	UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, status models.MessageStatus) error

	// Analytics & Statistics
	CountUnreadMessages(ctx context.Context, userID uuid.UUID) (int64, error)
	CountUnreadByConversation(ctx context.Context, userID, otherUserID uuid.UUID) (int64, error)
	GetMessageStats(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) (MessageStats, error)
	GetConversationStats(ctx context.Context, userID1, userID2 uuid.UUID) (ConversationStats, error)

	// Search & Filter
	SearchMessages(ctx context.Context, userID uuid.UUID, query string, pagination PaginationParams) ([]*models.Message, PaginationResult, error)
	FindMessagesByDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Message, PaginationResult, error)

	// Bulk Operations
	BulkMarkAsDelivered(ctx context.Context, receiverID uuid.UUID) error
	DeleteConversation(ctx context.Context, userID1, userID2 uuid.UUID) error
	DeleteMessagesOlderThan(ctx context.Context, tenantID uuid.UUID, duration time.Duration) error

	// Tenant Operations
	FindByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Message, PaginationResult, error)
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
}

// ConversationSummary represents a conversation overview
type ConversationSummary struct {
	OtherUserID       uuid.UUID            `json:"other_user_id"`
	OtherUserName     string               `json:"other_user_name"`
	LastMessage       string               `json:"last_message"`
	LastMessageAt     time.Time            `json:"last_message_at"`
	LastMessageType   models.MessageType   `json:"last_message_type"`
	UnreadCount       int64                `json:"unread_count"`
	TotalMessages     int64                `json:"total_messages"`
	BookingID         *uuid.UUID           `json:"booking_id,omitempty"`
	LastMessageStatus models.MessageStatus `json:"last_message_status"`
}

// MessageStats represents message statistics
type MessageStats struct {
	TotalMessages    int64                          `json:"total_messages"`
	SentMessages     int64                          `json:"sent_messages"`
	ReceivedMessages int64                          `json:"received_messages"`
	UnreadMessages   int64                          `json:"unread_messages"`
	MessagesByType   map[models.MessageType]int64   `json:"messages_by_type"`
	MessagesByStatus map[models.MessageStatus]int64 `json:"messages_by_status"`
	AveragePerDay    float64                        `json:"average_per_day"`
	LastMessageAt    *time.Time                     `json:"last_message_at,omitempty"`
}

// ConversationStats represents statistics for a specific conversation
type ConversationStats struct {
	TotalMessages       int64                        `json:"total_messages"`
	User1SentCount      int64                        `json:"user1_sent_count"`
	User2SentCount      int64                        `json:"user2_sent_count"`
	UnreadCount         int64                        `json:"unread_count"`
	FirstMessageAt      *time.Time                   `json:"first_message_at,omitempty"`
	LastMessageAt       *time.Time                   `json:"last_message_at,omitempty"`
	MessagesByType      map[models.MessageType]int64 `json:"messages_by_type"`
	AverageResponseTime time.Duration                `json:"average_response_time"`
}

// messageRepository implements MessageRepository
type messageRepository struct {
	BaseRepository[models.Message]
	db     *gorm.DB
	logger log.AllLogger
	cache  Cache
}

// NewMessageRepository creates a new MessageRepository instance
func NewMessageRepository(db *gorm.DB, config ...RepositoryConfig) MessageRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	baseRepo := NewBaseRepository[models.Message](db, cfg)

	return &messageRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
	}
}

// Helper methods for tenant isolation

// applyTenantFilter applies tenant isolation to a query based on user context
func (r *messageRepository) applyTenantFilter(query *gorm.DB, user *models.User) *gorm.DB {
	// Platform admins can access all tenants
	if user.IsPlatformAdmin() {
		return query
	}

	// Regular users can only access their tenant's data
	if user.TenantID != nil {
		return query.Where("messages.tenant_id = ?", *user.TenantID)
	}

	return query
}

// getUserWithTenantInfo retrieves user with tenant information for access control
func (r *messageRepository) getUserWithTenantInfo(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Select("id", "tenant_id", "is_platform_user", "role", "status").
		First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("USER_NOT_FOUND", "user not found", errors.ErrNotFound)
		}
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get user", err)
	}

	// Check if user is active
	if user.Status != models.UserStatusActive {
		return nil, errors.NewRepositoryError("USER_INACTIVE", "user account is not active", errors.ErrForbidden)
	}

	return &user, nil
}

// validateTenantAccess validates that users can communicate within the same tenant
func (r *messageRepository) validateTenantAccess(ctx context.Context, senderID, receiverID uuid.UUID) error {
	sender, err := r.getUserWithTenantInfo(ctx, senderID)
	if err != nil {
		return err
	}

	receiver, err := r.getUserWithTenantInfo(ctx, receiverID)
	if err != nil {
		return err
	}

	// Platform admins can message anyone
	if sender.IsPlatformAdmin() || receiver.IsPlatformAdmin() {
		return nil
	}

	// Both users must belong to the same tenant
	if sender.TenantID == nil || receiver.TenantID == nil {
		return errors.NewRepositoryError("NO_TENANT", "users must belong to a tenant", errors.ErrForbidden)
	}

	if *sender.TenantID != *receiver.TenantID {
		return errors.NewRepositoryError("TENANT_MISMATCH", "users must belong to the same tenant", errors.ErrForbidden)
	}

	return nil
}

// FindConversation retrieves all messages between two users with tenant isolation
func (r *messageRepository) FindConversation(ctx context.Context, userID1, userID2 uuid.UUID, pagination PaginationParams) ([]*models.Message, PaginationResult, error) {
	if userID1 == uuid.Nil || userID2 == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "user IDs cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Verify both users belong to the same tenant or one is a platform admin
	var user1, user2 models.User
	if err := r.db.WithContext(ctx).Select("id", "tenant_id", "is_platform_user", "role").First(&user1, userID1).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("USER_NOT_FOUND", "user1 not found", errors.ErrNotFound)
	}
	if err := r.db.WithContext(ctx).Select("id", "tenant_id", "is_platform_user", "role").First(&user2, userID2).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("USER_NOT_FOUND", "user2 not found", errors.ErrNotFound)
	}

	// Check tenant access - platform admins can access all conversations
	if !user1.IsPlatformAdmin() && !user2.IsPlatformAdmin() {
		if user1.TenantID == nil || user2.TenantID == nil || *user1.TenantID != *user2.TenantID {
			return nil, PaginationResult{}, errors.NewRepositoryError("TENANT_MISMATCH", "users must belong to the same tenant", errors.ErrForbidden)
		}
	}

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID1, userID2, userID2, userID1).
		Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count conversation messages", "user1", userID1, "user2", userID2, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count messages", err)
	}

	// Find paginated results
	var messages []*models.Message
	if err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Receiver").
		Preload("Booking").
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID1, userID2, userID2, userID1).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&messages).Error; err != nil {
		r.logger.Error("failed to find conversation", "user1", userID1, "user2", userID2, "error", err)
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find conversation", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return messages, paginationResult, nil
}

// FindConversationByBooking retrieves messages for a specific booking
func (r *messageRepository) FindConversationByBooking(ctx context.Context, bookingID uuid.UUID, pagination PaginationParams) ([]*models.Message, PaginationResult, error) {
	if bookingID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "booking_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("booking_id = ?", bookingID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count messages", err)
	}

	// Find messages
	var messages []*models.Message
	if err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Receiver").
		Preload("Booking").
		Where("booking_id = ?", bookingID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&messages).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find messages", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return messages, paginationResult, nil
}

// GetConversationList retrieves a list of all conversations for a user with tenant isolation
func (r *messageRepository) GetConversationList(ctx context.Context, userID uuid.UUID) ([]ConversationSummary, error) {
	if userID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	// Get user to check tenant and role
	var user models.User
	if err := r.db.WithContext(ctx).Select("id", "tenant_id", "is_platform_user", "role").First(&user, userID).Error; err != nil {
		return nil, errors.NewRepositoryError("USER_NOT_FOUND", "user not found", errors.ErrNotFound)
	}

	// Get distinct conversation partners with last message info
	type ConvResult struct {
		OtherUserID       uuid.UUID
		OtherUserName     string
		LastMessage       string
		LastMessageAt     time.Time
		LastMessageType   models.MessageType
		LastMessageStatus models.MessageStatus
		BookingID         *uuid.UUID
	}

	var conversations []ConvResult

	// Build query with tenant isolation (platform admins can see all)
	baseCondition := "(sender_id = ? OR receiver_id = ?)"
	params := []interface{}{userID, userID, userID, userID}

	// Add tenant isolation for non-platform users
	if !user.IsPlatformAdmin() && user.TenantID != nil {
		baseCondition = "(sender_id = ? OR receiver_id = ?) AND tenant_id = ?"
		params = append(params, *user.TenantID)
	}

	// Query for sent and received messages
	query := fmt.Sprintf(`
		WITH ranked_messages AS (
			SELECT
				CASE
					WHEN sender_id = ? THEN receiver_id
					ELSE sender_id
				END as other_user_id,
				content as last_message,
				created_at as last_message_at,
				type as last_message_type,
				status as last_message_status,
				booking_id,
				ROW_NUMBER() OVER (
					PARTITION BY CASE
						WHEN sender_id = ? THEN receiver_id
						ELSE sender_id
					END
					ORDER BY created_at DESC
				) as rn
			FROM messages
			WHERE %s
				AND deleted_at IS NULL
		)
		SELECT
			rm.other_user_id,
			COALESCE(u.first_name || ' ' || u.last_name, u.email) as other_user_name,
			rm.last_message,
			rm.last_message_at,
			rm.last_message_type,
			rm.last_message_status,
			rm.booking_id
		FROM ranked_messages rm
		LEFT JOIN users u ON u.id = rm.other_user_id
		WHERE rm.rn = 1
		ORDER BY rm.last_message_at DESC
	`, baseCondition)

	if err := r.db.WithContext(ctx).Raw(query, params...).Scan(&conversations).Error; err != nil {
		r.logger.Error("failed to get conversation list", "user_id", userID, "error", err)
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get conversation list", err)
	}

	// Get unread counts and total messages for each conversation
	summaries := make([]ConversationSummary, len(conversations))
	for i, conv := range conversations {
		summaries[i] = ConversationSummary{
			OtherUserID:       conv.OtherUserID,
			OtherUserName:     conv.OtherUserName,
			LastMessage:       conv.LastMessage,
			LastMessageAt:     conv.LastMessageAt,
			LastMessageType:   conv.LastMessageType,
			LastMessageStatus: conv.LastMessageStatus,
			BookingID:         conv.BookingID,
		}

		// Count unread messages from this conversation partner
		var unreadCount int64
		unreadQuery := r.db.WithContext(ctx).
			Model(&models.Message{}).
			Where("sender_id = ? AND receiver_id = ? AND status IN (?, ?)",
				conv.OtherUserID, userID, models.MessageStatusSent, models.MessageStatusDelivered)

		// Apply tenant filter for non-platform users
		if !user.IsPlatformAdmin() && user.TenantID != nil {
			unreadQuery = unreadQuery.Where("tenant_id = ?", *user.TenantID)
		}

		unreadQuery.Count(&unreadCount)
		summaries[i].UnreadCount = unreadCount

		// Count total messages in conversation
		var totalCount int64
		totalQuery := r.db.WithContext(ctx).
			Model(&models.Message{}).
			Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
				userID, conv.OtherUserID, conv.OtherUserID, userID)

		// Apply tenant filter for non-platform users
		if !user.IsPlatformAdmin() && user.TenantID != nil {
			totalQuery = totalQuery.Where("tenant_id = ?", *user.TenantID)
		}

		totalQuery.Count(&totalCount)
		summaries[i].TotalMessages = totalCount
	}

	return summaries, nil
}

// FindByReceiverID retrieves messages for a receiver
func (r *messageRepository) FindByReceiverID(ctx context.Context, receiverID uuid.UUID, pagination PaginationParams) ([]*models.Message, PaginationResult, error) {
	if receiverID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "receiver_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("receiver_id = ?", receiverID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count messages", err)
	}

	var messages []*models.Message
	if err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Booking").
		Where("receiver_id = ?", receiverID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&messages).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find messages", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return messages, paginationResult, nil
}

// FindBySenderID retrieves messages sent by a user
func (r *messageRepository) FindBySenderID(ctx context.Context, senderID uuid.UUID, pagination PaginationParams) ([]*models.Message, PaginationResult, error) {
	if senderID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "sender_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("sender_id = ?", senderID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count messages", err)
	}

	var messages []*models.Message
	if err := r.db.WithContext(ctx).
		Preload("Receiver").
		Preload("Booking").
		Where("sender_id = ?", senderID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&messages).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find messages", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return messages, paginationResult, nil
}

// FindUnreadMessages retrieves all unread messages for a user
func (r *messageRepository) FindUnreadMessages(ctx context.Context, receiverID uuid.UUID) ([]*models.Message, error) {
	if receiverID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "receiver_id cannot be nil", errors.ErrInvalidInput)
	}

	var messages []*models.Message
	if err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Booking").
		Where("receiver_id = ? AND status IN (?, ?)",
			receiverID, models.MessageStatusSent, models.MessageStatusDelivered).
		Order("created_at DESC").
		Find(&messages).Error; err != nil {
		r.logger.Error("failed to find unread messages", "receiver_id", receiverID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find unread messages", err)
	}

	return messages, nil
}

// FindMessagesByStatus retrieves messages by status
func (r *messageRepository) FindMessagesByStatus(ctx context.Context, userID uuid.UUID, status models.MessageStatus, pagination PaginationParams) ([]*models.Message, PaginationResult, error) {
	if userID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("(sender_id = ? OR receiver_id = ?) AND status = ?", userID, userID, status).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count messages", err)
	}

	var messages []*models.Message
	if err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Receiver").
		Where("(sender_id = ? OR receiver_id = ?) AND status = ?", userID, userID, status).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&messages).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find messages", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return messages, paginationResult, nil
}

// FindMessagesByType retrieves messages by type
func (r *messageRepository) FindMessagesByType(ctx context.Context, tenantID uuid.UUID, messageType models.MessageType, pagination PaginationParams) ([]*models.Message, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("tenant_id = ? AND type = ?", tenantID, messageType).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count messages", err)
	}

	var messages []*models.Message
	if err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Receiver").
		Where("tenant_id = ? AND type = ?", tenantID, messageType).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&messages).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find messages", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return messages, paginationResult, nil
}

// FindReplies retrieves all replies to a message
func (r *messageRepository) FindReplies(ctx context.Context, parentMessageID uuid.UUID) ([]*models.Message, error) {
	if parentMessageID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "parent_message_id cannot be nil", errors.ErrInvalidInput)
	}

	var messages []*models.Message
	if err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Receiver").
		Where("parent_message_id = ?", parentMessageID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		r.logger.Error("failed to find replies", "parent_message_id", parentMessageID, "error", err)
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find replies", err)
	}

	return messages, nil
}

// GetThreadMessages retrieves all messages in a thread with pagination
func (r *messageRepository) GetThreadMessages(ctx context.Context, parentMessageID uuid.UUID, pagination PaginationParams) ([]*models.Message, PaginationResult, error) {
	if parentMessageID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "parent_message_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("parent_message_id = ?", parentMessageID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count thread messages", err)
	}

	var messages []*models.Message
	if err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Receiver").
		Where("parent_message_id = ?", parentMessageID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find thread messages", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return messages, paginationResult, nil
}

// MarkAsRead marks a message as read
func (r *messageRepository) MarkAsRead(ctx context.Context, messageID uuid.UUID) error {
	if messageID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "message_id cannot be nil", errors.ErrInvalidInput)
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"status":  models.MessageStatusRead,
			"read_at": now,
		})

	if result.Error != nil {
		r.logger.Error("failed to mark message as read", "message_id", messageID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark message as read", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "message not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:messages:*")
	}

	return nil
}

// MarkMultipleAsRead marks multiple messages as read
func (r *messageRepository) MarkMultipleAsRead(ctx context.Context, messageIDs []uuid.UUID) error {
	if len(messageIDs) == 0 {
		return nil
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id IN ?", messageIDs).
		Updates(map[string]interface{}{
			"status":  models.MessageStatusRead,
			"read_at": now,
		})

	if result.Error != nil {
		r.logger.Error("failed to mark messages as read", "count", len(messageIDs), "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark messages as read", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:messages:*")
	}

	r.logger.Debug("marked messages as read", "count", result.RowsAffected)
	return nil
}

// MarkConversationAsRead marks all messages in a conversation as read
func (r *messageRepository) MarkConversationAsRead(ctx context.Context, receiverID, senderID uuid.UUID) error {
	if receiverID == uuid.Nil || senderID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "receiver_id and sender_id cannot be nil", errors.ErrInvalidInput)
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("receiver_id = ? AND sender_id = ? AND status IN (?, ?)",
			receiverID, senderID, models.MessageStatusSent, models.MessageStatusDelivered).
		Updates(map[string]interface{}{
			"status":  models.MessageStatusRead,
			"read_at": now,
		})

	if result.Error != nil {
		r.logger.Error("failed to mark conversation as read", "receiver_id", receiverID, "sender_id", senderID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark conversation as read", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:messages:*")
	}

	r.logger.Debug("marked conversation as read", "count", result.RowsAffected)
	return nil
}

// UpdateMessageStatus updates the status of a message
func (r *messageRepository) UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, status models.MessageStatus) error {
	if messageID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "message_id cannot be nil", errors.ErrInvalidInput)
	}

	updates := map[string]interface{}{
		"status": status,
	}

	if status == models.MessageStatusRead {
		updates["read_at"] = time.Now()
	}

	result := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id = ?", messageID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("failed to update message status", "message_id", messageID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update message status", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "message not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:messages:*")
	}

	return nil
}

// CountUnreadMessages counts unread messages for a user
func (r *messageRepository) CountUnreadMessages(ctx context.Context, userID uuid.UUID) (int64, error) {
	if userID == uuid.Nil {
		return 0, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("receiver_id = ? AND status IN (?, ?)",
			userID, models.MessageStatusSent, models.MessageStatusDelivered).
		Count(&count).Error; err != nil {
		r.logger.Error("failed to count unread messages", "user_id", userID, "error", err)
		return 0, errors.NewRepositoryError("COUNT_FAILED", "failed to count unread messages", err)
	}

	return count, nil
}

// CountUnreadByConversation counts unread messages in a specific conversation
func (r *messageRepository) CountUnreadByConversation(ctx context.Context, userID, otherUserID uuid.UUID) (int64, error) {
	if userID == uuid.Nil || otherUserID == uuid.Nil {
		return 0, errors.NewRepositoryError("INVALID_INPUT", "user IDs cannot be nil", errors.ErrInvalidInput)
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("receiver_id = ? AND sender_id = ? AND status IN (?, ?)",
			userID, otherUserID, models.MessageStatusSent, models.MessageStatusDelivered).
		Count(&count).Error; err != nil {
		return 0, errors.NewRepositoryError("COUNT_FAILED", "failed to count unread messages", err)
	}

	return count, nil
}

// GetMessageStats retrieves comprehensive message statistics
func (r *messageRepository) GetMessageStats(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) (MessageStats, error) {
	if tenantID == uuid.Nil {
		return MessageStats{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	stats := MessageStats{
		MessagesByType:   make(map[models.MessageType]int64),
		MessagesByStatus: make(map[models.MessageStatus]int64),
	}

	query := r.db.WithContext(ctx).Model(&models.Message{}).Where("tenant_id = ?", tenantID)
	if userID != nil {
		query = query.Where("sender_id = ? OR receiver_id = ?", *userID, *userID)
	}

	// Total messages
	if err := query.Count(&stats.TotalMessages).Error; err != nil {
		return stats, errors.NewRepositoryError("QUERY_FAILED", "failed to count total messages", err)
	}

	// Sent and received counts (if user specified)
	if userID != nil {
		r.db.WithContext(ctx).
			Model(&models.Message{}).
			Where("tenant_id = ? AND sender_id = ?", tenantID, *userID).
			Count(&stats.SentMessages)

		r.db.WithContext(ctx).
			Model(&models.Message{}).
			Where("tenant_id = ? AND receiver_id = ?", tenantID, *userID).
			Count(&stats.ReceivedMessages)

		// Unread count
		r.db.WithContext(ctx).
			Model(&models.Message{}).
			Where("tenant_id = ? AND receiver_id = ? AND status IN (?, ?)",
				tenantID, *userID, models.MessageStatusSent, models.MessageStatusDelivered).
			Count(&stats.UnreadMessages)
	}

	// Messages by type
	type TypeCount struct {
		Type  models.MessageType
		Count int64
	}
	var typeCounts []TypeCount
	typeQuery := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Select("type, COUNT(*) as count").
		Where("tenant_id = ?", tenantID)

	if userID != nil {
		typeQuery = typeQuery.Where("sender_id = ? OR receiver_id = ?", *userID, *userID)
	}

	if err := typeQuery.Group("type").Scan(&typeCounts).Error; err != nil {
		r.logger.Warn("failed to count messages by type", "error", err)
	} else {
		for _, tc := range typeCounts {
			stats.MessagesByType[tc.Type] = tc.Count
		}
	}

	// Messages by status
	type StatusCount struct {
		Status models.MessageStatus
		Count  int64
	}
	var statusCounts []StatusCount
	statusQuery := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID)

	if userID != nil {
		statusQuery = statusQuery.Where("sender_id = ? OR receiver_id = ?", *userID, *userID)
	}

	if err := statusQuery.Group("status").Scan(&statusCounts).Error; err != nil {
		r.logger.Warn("failed to count messages by status", "error", err)
	} else {
		for _, sc := range statusCounts {
			stats.MessagesByStatus[sc.Status] = sc.Count
		}
	}

	// Last message timestamp
	var lastMessage models.Message
	lastQuery := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID)

	if userID != nil {
		lastQuery = lastQuery.Where("sender_id = ? OR receiver_id = ?", *userID, *userID)
	}

	if err := lastQuery.Order("created_at DESC").First(&lastMessage).Error; err == nil {
		stats.LastMessageAt = &lastMessage.CreatedAt
	}

	// Calculate average messages per day
	if stats.TotalMessages > 0 {
		var firstMessage models.Message
		firstQuery := r.db.WithContext(ctx).
			Where("tenant_id = ?", tenantID)

		if userID != nil {
			firstQuery = firstQuery.Where("sender_id = ? OR receiver_id = ?", *userID, *userID)
		}

		if err := firstQuery.Order("created_at ASC").First(&firstMessage).Error; err == nil {
			daysSinceFirst := time.Since(firstMessage.CreatedAt).Hours() / 24
			if daysSinceFirst > 0 {
				stats.AveragePerDay = float64(stats.TotalMessages) / daysSinceFirst
			}
		}
	}

	return stats, nil
}

// GetConversationStats retrieves statistics for a specific conversation
func (r *messageRepository) GetConversationStats(ctx context.Context, userID1, userID2 uuid.UUID) (ConversationStats, error) {
	if userID1 == uuid.Nil || userID2 == uuid.Nil {
		return ConversationStats{}, errors.NewRepositoryError("INVALID_INPUT", "user IDs cannot be nil", errors.ErrInvalidInput)
	}

	stats := ConversationStats{
		MessagesByType: make(map[models.MessageType]int64),
	}

	baseQuery := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID1, userID2, userID2, userID1)

	// Total messages
	if err := baseQuery.Count(&stats.TotalMessages).Error; err != nil {
		return stats, errors.NewRepositoryError("QUERY_FAILED", "failed to count messages", err)
	}

	// User-specific counts
	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("sender_id = ? AND receiver_id = ?", userID1, userID2).
		Count(&stats.User1SentCount)

	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("sender_id = ? AND receiver_id = ?", userID2, userID1).
		Count(&stats.User2SentCount)

	// Unread messages
	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND status IN (?, ?)",
			userID1, userID2, userID2, userID1,
			models.MessageStatusSent, models.MessageStatusDelivered).
		Count(&stats.UnreadCount)

	// First and last message timestamps
	var firstMessage, lastMessage models.Message

	if err := r.db.WithContext(ctx).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID1, userID2, userID2, userID1).
		Order("created_at ASC").
		First(&firstMessage).Error; err == nil {
		stats.FirstMessageAt = &firstMessage.CreatedAt
	}

	if err := r.db.WithContext(ctx).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID1, userID2, userID2, userID1).
		Order("created_at DESC").
		First(&lastMessage).Error; err == nil {
		stats.LastMessageAt = &lastMessage.CreatedAt
	}

	// Messages by type
	type TypeCount struct {
		Type  models.MessageType
		Count int64
	}
	var typeCounts []TypeCount
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Select("type, COUNT(*) as count").
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID1, userID2, userID2, userID1).
		Group("type").
		Scan(&typeCounts).Error; err == nil {
		for _, tc := range typeCounts {
			stats.MessagesByType[tc.Type] = tc.Count
		}
	}

	// Calculate average response time
	type MessagePair struct {
		SendTime     time.Time
		ResponseTime time.Time
	}

	query := `
		WITH message_pairs AS (
			SELECT
				m1.created_at as send_time,
				MIN(m2.created_at) as response_time
			FROM messages m1
			LEFT JOIN messages m2 ON (
				m1.sender_id = m2.receiver_id
				AND m1.receiver_id = m2.sender_id
				AND m2.created_at > m1.created_at
			)
			WHERE (
				(m1.sender_id = ? AND m1.receiver_id = ?)
				OR (m1.sender_id = ? AND m1.receiver_id = ?)
			)
			AND m1.deleted_at IS NULL
			AND (m2.deleted_at IS NULL OR m2.id IS NULL)
			GROUP BY m1.id, m1.created_at
			HAVING MIN(m2.created_at) IS NOT NULL
		)
		SELECT AVG(EXTRACT(EPOCH FROM (response_time - send_time))) as avg_seconds
		FROM message_pairs
	`

	var avgSeconds float64
	if err := r.db.WithContext(ctx).Raw(query, userID1, userID2, userID2, userID1).Scan(&avgSeconds).Error; err == nil {
		stats.AverageResponseTime = time.Duration(avgSeconds) * time.Second
	}

	return stats, nil
}

// SearchMessages searches messages by content
func (r *messageRepository) SearchMessages(ctx context.Context, userID uuid.UUID, query string, pagination PaginationParams) ([]*models.Message, PaginationResult, error) {
	if userID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()
	searchPattern := "%" + query + "%"

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("(sender_id = ? OR receiver_id = ?) AND content ILIKE ?",
			userID, userID, searchPattern).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count messages", err)
	}

	// Find messages
	var messages []*models.Message
	if err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Receiver").
		Preload("Booking").
		Where("(sender_id = ? OR receiver_id = ?) AND content ILIKE ?",
			userID, userID, searchPattern).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&messages).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search messages", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return messages, paginationResult, nil
}

// FindMessagesByDateRange retrieves messages within a date range
func (r *messageRepository) FindMessagesByDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Message, PaginationResult, error) {
	if userID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("(sender_id = ? OR receiver_id = ?) AND created_at BETWEEN ? AND ?",
			userID, userID, startDate, endDate).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count messages", err)
	}

	// Find messages
	var messages []*models.Message
	if err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Receiver").
		Preload("Booking").
		Where("(sender_id = ? OR receiver_id = ?) AND created_at BETWEEN ? AND ?",
			userID, userID, startDate, endDate).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&messages).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find messages", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return messages, paginationResult, nil
}

// BulkMarkAsDelivered marks all sent messages to a receiver as delivered
func (r *messageRepository) BulkMarkAsDelivered(ctx context.Context, receiverID uuid.UUID) error {
	if receiverID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "receiver_id cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("receiver_id = ? AND status = ?", receiverID, models.MessageStatusSent).
		Update("status", models.MessageStatusDelivered)

	if result.Error != nil {
		r.logger.Error("failed to mark messages as delivered", "receiver_id", receiverID, "error", result.Error)
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark messages as delivered", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:messages:*")
	}

	r.logger.Debug("marked messages as delivered", "count", result.RowsAffected)
	return nil
}

// DeleteConversation deletes all messages between two users
func (r *messageRepository) DeleteConversation(ctx context.Context, userID1, userID2 uuid.UUID) error {
	if userID1 == uuid.Nil || userID2 == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "user IDs cannot be nil", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID1, userID2, userID2, userID1).
		Delete(&models.Message{})

	if result.Error != nil {
		r.logger.Error("failed to delete conversation", "user1", userID1, "user2", userID2, "error", result.Error)
		return errors.NewRepositoryError("DELETE_FAILED", "failed to delete conversation", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:messages:*")
	}

	r.logger.Info("deleted conversation", "count", result.RowsAffected)
	return nil
}

// DeleteMessagesOlderThan deletes messages older than specified duration
func (r *messageRepository) DeleteMessagesOlderThan(ctx context.Context, tenantID uuid.UUID, duration time.Duration) error {
	if tenantID == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	cutoffDate := time.Now().Add(-duration)

	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND created_at < ?", tenantID, cutoffDate).
		Delete(&models.Message{})

	if result.Error != nil {
		r.logger.Error("failed to delete old messages", "tenant_id", tenantID, "error", result.Error)
		return errors.NewRepositoryError("DELETE_FAILED", "failed to delete old messages", result.Error)
	}

	// Invalidate cache
	if r.cache != nil {
		r.cache.DeletePattern(ctx, "repo:messages:*")
	}

	r.logger.Info("deleted old messages", "tenant_id", tenantID, "count", result.RowsAffected, "older_than", duration)
	return nil
}

// FindByTenantID retrieves messages for a tenant with pagination
func (r *messageRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Message, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count messages", err)
	}

	// Find messages
	var messages []*models.Message
	if err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Receiver").
		Preload("Booking").
		Where("tenant_id = ?", tenantID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&messages).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find messages", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return messages, paginationResult, nil
}

// CountByTenant counts messages for a tenant
func (r *messageRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	if tenantID == uuid.Nil {
		return 0, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error; err != nil {
		return 0, errors.NewRepositoryError("COUNT_FAILED", "failed to count messages", err)
	}

	return count, nil
}

// Additional helper methods for message repository

// CreateWithValidation creates a message with tenant and user validation
func (r *messageRepository) CreateWithValidation(ctx context.Context, message *models.Message) error {
	if message == nil {
		return errors.NewRepositoryError("INVALID_INPUT", "message cannot be nil", errors.ErrInvalidInput)
	}

	// Validate tenant access between sender and receiver
	if err := r.validateTenantAccess(ctx, message.SenderID, message.ReceiverID); err != nil {
		return err
	}

	// Get sender to set tenant ID
	sender, err := r.getUserWithTenantInfo(ctx, message.SenderID)
	if err != nil {
		return err
	}

	// Set tenant ID from sender (platform admins' messages inherit receiver's tenant)
	if sender.IsPlatformAdmin() {
		receiver, err := r.getUserWithTenantInfo(ctx, message.ReceiverID)
		if err != nil {
			return err
		}
		if receiver.TenantID != nil {
			message.TenantID = *receiver.TenantID
		}
	} else if sender.TenantID != nil {
		message.TenantID = *sender.TenantID
	} else {
		return errors.NewRepositoryError("NO_TENANT", "sender must belong to a tenant", errors.ErrInvalidInput)
	}

	// Set default status if not set
	if message.Status == "" {
		message.Status = models.MessageStatusSent
	}

	// Set default type if not set
	if message.Type == "" {
		message.Type = models.MessageTypeText
	}

	// Use base repository create
	return r.BaseRepository.Create(ctx, message)
}

// FindByTenantAndUser retrieves messages for a specific user within a tenant
func (r *messageRepository) FindByTenantAndUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, pagination PaginationParams) ([]*models.Message, PaginationResult, error) {
	if tenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}
	if userID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "user_id cannot be nil", errors.ErrInvalidInput)
	}

	pagination.Validate()

	// Count total
	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("tenant_id = ? AND (sender_id = ? OR receiver_id = ?)", tenantID, userID, userID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count messages", err)
	}

	// Find messages
	var messages []*models.Message
	if err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Receiver").
		Preload("Booking").
		Where("tenant_id = ? AND (sender_id = ? OR receiver_id = ?)", tenantID, userID, userID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&messages).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find messages", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return messages, paginationResult, nil
}

// GetTenantMessageStats retrieves tenant-specific message statistics with plan limits check
func (r *messageRepository) GetTenantMessageStats(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "tenant_id cannot be nil", errors.ErrInvalidInput)
	}

	stats := make(map[string]interface{})

	// Get tenant info for plan limits
	var tenant models.Tenant
	if err := r.db.WithContext(ctx).Select("id", "plan", "status").First(&tenant, tenantID).Error; err != nil {
		r.logger.Warn("failed to get tenant info", "tenant_id", tenantID, "error", err)
	} else {
		stats["tenant_plan"] = tenant.Plan
		stats["tenant_status"] = tenant.Status
	}

	// Total messages
	var totalMessages int64
	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalMessages)
	stats["total_messages"] = totalMessages

	// Messages by status
	type StatusCount struct {
		Status models.MessageStatus
		Count  int64
	}
	var statusCounts []StatusCount
	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("status").
		Scan(&statusCounts)

	statusMap := make(map[string]int64)
	for _, sc := range statusCounts {
		statusMap[string(sc.Status)] = sc.Count
	}
	stats["by_status"] = statusMap

	// Messages by type
	type TypeCount struct {
		Type  models.MessageType
		Count int64
	}
	var typeCounts []TypeCount
	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Select("type, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("type").
		Scan(&typeCounts)

	typeMap := make(map[string]int64)
	for _, tc := range typeCounts {
		typeMap[string(tc.Type)] = tc.Count
	}
	stats["by_type"] = typeMap

	// Active conversations (unique user pairs)
	var activeConversations int64
	r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT CONCAT(LEAST(sender_id::text, receiver_id::text), GREATEST(sender_id::text, receiver_id::text)))
		FROM messages
		WHERE tenant_id = ? AND deleted_at IS NULL
	`, tenantID).Scan(&activeConversations)
	stats["active_conversations"] = activeConversations

	// Messages today
	var messagesToday int64
	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("tenant_id = ? AND DATE(created_at) = CURRENT_DATE", tenantID).
		Count(&messagesToday)
	stats["messages_today"] = messagesToday

	// Messages this week
	var messagesThisWeek int64
	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("tenant_id = ? AND created_at >= DATE_TRUNC('week', CURRENT_DATE)", tenantID).
		Count(&messagesThisWeek)
	stats["messages_this_week"] = messagesThisWeek

	// Messages this month
	var messagesThisMonth int64
	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("tenant_id = ? AND created_at >= DATE_TRUNC('month', CURRENT_DATE)", tenantID).
		Count(&messagesThisMonth)
	stats["messages_this_month"] = messagesThisMonth

	// Average response time (in minutes)
	var avgResponseMinutes float64
	r.db.WithContext(ctx).Raw(`
		WITH message_pairs AS (
			SELECT
				m1.created_at as send_time,
				MIN(m2.created_at) as response_time
			FROM messages m1
			LEFT JOIN messages m2 ON (
				m1.sender_id = m2.receiver_id
				AND m1.receiver_id = m2.sender_id
				AND m2.created_at > m1.created_at
				AND m1.tenant_id = m2.tenant_id
			)
			WHERE m1.tenant_id = ?
				AND m1.deleted_at IS NULL
				AND (m2.deleted_at IS NULL OR m2.id IS NULL)
			GROUP BY m1.id, m1.created_at
			HAVING MIN(m2.created_at) IS NOT NULL
		)
		SELECT AVG(EXTRACT(EPOCH FROM (response_time - send_time)) / 60)
		FROM message_pairs
	`, tenantID).Scan(&avgResponseMinutes)
	stats["avg_response_time_minutes"] = avgResponseMinutes

	// Unread messages count
	var unreadMessages int64
	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("tenant_id = ? AND status IN (?, ?)",
			tenantID, models.MessageStatusSent, models.MessageStatusDelivered).
		Count(&unreadMessages)
	stats["unread_messages"] = unreadMessages

	return stats, nil
}
