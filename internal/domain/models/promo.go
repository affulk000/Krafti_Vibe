package models

import (
	"time"

	"github.com/google/uuid"
)

type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed"
)

type PromoCode struct {
	BaseModel

	// Multi-tenancy
	TenantID *uuid.UUID `json:"tenant_id,omitempty" gorm:"type:uuid;index"` // null = platform-wide

	// Code Details
	Code        string `json:"code" gorm:"uniqueIndex;not null;size:50" validate:"required,uppercase"`
	Description string `json:"description,omitempty" gorm:"type:text"`

	// Discount
	Type           DiscountType `json:"type" gorm:"type:varchar(50);not null" validate:"required"`
	Value          float64      `json:"value" gorm:"type:decimal(10,2);not null" validate:"required,min=0"`
	MaxDiscount    float64      `json:"max_discount,omitempty" gorm:"type:decimal(10,2)"` // For percentage type
	MinOrderAmount float64      `json:"min_order_amount,omitempty" gorm:"type:decimal(10,2)"`

	// Validity
	StartsAt  time.Time  `json:"starts_at" gorm:"not null" validate:"required"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// Usage Limits
	MaxUses        int `json:"max_uses,omitempty"` // null = unlimited
	UsedCount      int `json:"used_count" gorm:"default:0"`
	MaxUsesPerUser int `json:"max_uses_per_user,omitempty"`

	// Restrictions
	ApplicableServices []uuid.UUID `json:"applicable_services,omitempty" gorm:"type:uuid[]"`
	ApplicableArtisans []uuid.UUID `json:"applicable_artisans,omitempty" gorm:"type:uuid[]"`

	// Status
	IsActive bool `json:"is_active" gorm:"default:true"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// Business Methods
func (pc *PromoCode) IsValid() bool {
	if !pc.IsActive {
		return false
	}

	now := time.Now()
	if now.Before(pc.StartsAt) {
		return false
	}

	if pc.ExpiresAt != nil && now.After(*pc.ExpiresAt) {
		return false
	}

	if pc.MaxUses > 0 && pc.UsedCount >= pc.MaxUses {
		return false
	}

	return true
}

func (pc *PromoCode) CalculateDiscount(amount float64) float64 {
	if !pc.IsValid() || amount < pc.MinOrderAmount {
		return 0
	}

	var discount float64

	switch pc.Type {
	case DiscountTypeFixed:
		discount = pc.Value
	case DiscountTypePercentage:
		discount = amount * (pc.Value / 100)
		if pc.MaxDiscount > 0 {
			discount = min(discount, pc.MaxDiscount)
		}
	}

	discount = min(discount, amount)

	return discount
}
