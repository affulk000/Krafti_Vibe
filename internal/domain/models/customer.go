package models

import (
	"github.com/google/uuid"
)

type Customer struct {
	BaseModel
	UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;uniqueIndex;not null"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;index:idx_customer_tenant;not null"`

	// Preferences
	PreferredArtisans []uuid.UUID `json:"preferred_artisans,omitempty" gorm:"type:uuid[]"`
	Notes             string      `json:"notes,omitempty" gorm:"type:text"`

	// Loyalty & Rewards
	LoyaltyPoints int     `json:"loyalty_points" gorm:"default:0"`
	TotalSpent    float64 `json:"total_spent" gorm:"type:decimal(10,2);default:0"`

	// Statistics
	TotalBookings     int `json:"total_bookings" gorm:"default:0"`
	CancelledBookings int `json:"cancelled_bookings" gorm:"default:0"`
	CompletedBookings int `json:"completed_bookings" gorm:"default:0"`

	// Payment
	DefaultPaymentMethodID string `json:"default_payment_method_id,omitempty" gorm:"size:255"`

	// Location
	PrimaryLocation Location `json:"primary_location,omitempty" gorm:"type:jsonb"`

	// Communication Preferences
	EmailNotifications bool `json:"email_notifications" gorm:"default:true"`
	SMSNotifications   bool `json:"sms_notifications" gorm:"default:true"`
	PushNotifications  bool `json:"push_notifications" gorm:"default:true"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	User     *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Bookings []Booking `json:"bookings,omitempty" gorm:"foreignKey:CustomerID"`
	Reviews  []Review  `json:"reviews,omitempty" gorm:"foreignKey:CustomerID"`
	Projects []Project `json:"projects,omitempty" gorm:"foreignKey:CustomerID"`
}

// Business Methods
func (c *Customer) GetLoyaltyTier() string {
	switch {
	case c.LoyaltyPoints >= 1000:
		return "Platinum"
	case c.LoyaltyPoints >= 500:
		return "Gold"
	case c.LoyaltyPoints >= 100:
		return "Silver"
	default:
		return "Bronze"
	}
}
