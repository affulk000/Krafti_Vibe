package models

import (
	"time"

	"github.com/google/uuid"
)

type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeFile   MessageType = "file"
	MessageTypeSystem MessageType = "system"
)

type Message struct {
	BaseModel

	// Multi-tenancy
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`

	// Participants
	SenderID   uuid.UUID `json:"sender_id" gorm:"type:uuid;not null;index" validate:"required"`
	ReceiverID uuid.UUID `json:"receiver_id" gorm:"type:uuid;not null;index" validate:"required"`

	// Related Booking (optional)
	BookingID *uuid.UUID `json:"booking_id,omitempty" gorm:"type:uuid;index"`

	// Content
	Type    MessageType `json:"type" gorm:"type:varchar(50);not null;default:'text'" validate:"required"`
	Content string      `json:"content" gorm:"type:text;not null" validate:"required"`
	FileURL string      `json:"file_url,omitempty" gorm:"size:500"`

	// Status
	Status MessageStatus `json:"status" gorm:"type:varchar(50);not null;default:'sent'" validate:"required"`
	ReadAt *time.Time    `json:"read_at,omitempty"`

	// Thread
	ParentMessageID *uuid.UUID `json:"parent_message_id,omitempty" gorm:"type:uuid;index"` // For replies

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Sender        *User     `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	Receiver      *User     `json:"receiver,omitempty" gorm:"foreignKey:ReceiverID"`
	Booking       *Booking  `json:"booking,omitempty" gorm:"foreignKey:BookingID"`
	ParentMessage *Message  `json:"parent_message,omitempty" gorm:"foreignKey:ParentMessageID"`
	Replies       []Message `json:"replies,omitempty" gorm:"foreignKey:ParentMessageID"`
}

// Business Methods
func (m *Message) MarkAsRead() {
	now := time.Now()
	m.Status = MessageStatusRead
	m.ReadAt = &now
}

func (m *Message) IsUnread() bool {
	return m.Status == MessageStatusSent || m.Status == MessageStatusDelivered
}
