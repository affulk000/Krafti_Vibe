package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Promo Code Request DTOs
// ============================================================================

// CreatePromoCodeRequest represents a request to create a promo code
type CreatePromoCodeRequest struct {
	Code               string              `json:"code" validate:"required,uppercase,min=3,max=50"`
	Description        string              `json:"description,omitempty"`
	Type               models.DiscountType `json:"type" validate:"required,oneof=percentage fixed"`
	Value              float64             `json:"value" validate:"required,min=0"`
	MaxDiscount        float64             `json:"max_discount,omitempty" validate:"omitempty,min=0"`
	MinOrderAmount     float64             `json:"min_order_amount,omitempty" validate:"omitempty,min=0"`
	StartsAt           time.Time           `json:"starts_at" validate:"required"`
	ExpiresAt          *time.Time          `json:"expires_at,omitempty"`
	MaxUses            int                 `json:"max_uses,omitempty" validate:"omitempty,min=1"`
	MaxUsesPerUser     int                 `json:"max_uses_per_user,omitempty" validate:"omitempty,min=1"`
	ApplicableServices []uuid.UUID         `json:"applicable_services,omitempty"`
	ApplicableArtisans []uuid.UUID         `json:"applicable_artisans,omitempty"`
	IsActive           bool                `json:"is_active"`
	Metadata           map[string]any      `json:"metadata,omitempty"`
}

// UpdatePromoCodeRequest represents a request to update a promo code
type UpdatePromoCodeRequest struct {
	Description        *string        `json:"description,omitempty"`
	MaxDiscount        *float64       `json:"max_discount,omitempty" validate:"omitempty,min=0"`
	MinOrderAmount     *float64       `json:"min_order_amount,omitempty" validate:"omitempty,min=0"`
	ExpiresAt          *time.Time     `json:"expires_at,omitempty"`
	MaxUses            *int           `json:"max_uses,omitempty" validate:"omitempty,min=1"`
	MaxUsesPerUser     *int           `json:"max_uses_per_user,omitempty" validate:"omitempty,min=1"`
	ApplicableServices []uuid.UUID    `json:"applicable_services,omitempty"`
	ApplicableArtisans []uuid.UUID    `json:"applicable_artisans,omitempty"`
	IsActive           *bool          `json:"is_active,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

// ValidatePromoCodeRequest represents a request to validate a promo code
type ValidatePromoCodeRequest struct {
	Code      string     `json:"code" validate:"required"`
	Amount    float64    `json:"amount" validate:"required,min=0"`
	ServiceID *uuid.UUID `json:"service_id,omitempty"`
	ArtisanID *uuid.UUID `json:"artisan_id,omitempty"`
}

// PromoCodeFilter represents filters for promo code queries
type PromoCodeFilter struct {
	TenantID  *uuid.UUID           `json:"tenant_id,omitempty"`
	Type      *models.DiscountType `json:"type,omitempty"`
	IsActive  *bool                `json:"is_active,omitempty"`
	IsExpired *bool                `json:"is_expired,omitempty"`
	ServiceID *uuid.UUID           `json:"service_id,omitempty"`
	ArtisanID *uuid.UUID           `json:"artisan_id,omitempty"`
	Page      int                  `json:"page"`
	PageSize  int                  `json:"page_size"`
}

// ============================================================================
// Promo Code Response DTOs
// ============================================================================

// PromoCodeResponse represents a promo code
type PromoCodeResponse struct {
	ID                 uuid.UUID           `json:"id"`
	TenantID           *uuid.UUID          `json:"tenant_id,omitempty"`
	Code               string              `json:"code"`
	Description        string              `json:"description,omitempty"`
	Type               models.DiscountType `json:"type"`
	Value              float64             `json:"value"`
	MaxDiscount        float64             `json:"max_discount,omitempty"`
	MinOrderAmount     float64             `json:"min_order_amount,omitempty"`
	StartsAt           time.Time           `json:"starts_at"`
	ExpiresAt          *time.Time          `json:"expires_at,omitempty"`
	MaxUses            int                 `json:"max_uses,omitempty"`
	UsedCount          int                 `json:"used_count"`
	MaxUsesPerUser     int                 `json:"max_uses_per_user,omitempty"`
	ApplicableServices []uuid.UUID         `json:"applicable_services,omitempty"`
	ApplicableArtisans []uuid.UUID         `json:"applicable_artisans,omitempty"`
	IsActive           bool                `json:"is_active"`
	IsValid            bool                `json:"is_valid"`
	IsExpired          bool                `json:"is_expired"`
	RemainingUses      int                 `json:"remaining_uses"`
	UsagePercentage    float64             `json:"usage_percentage"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

// PromoCodeListResponse represents a paginated list of promo codes
type PromoCodeListResponse struct {
	PromoCodes  []*PromoCodeResponse `json:"promo_codes"`
	Page        int                  `json:"page"`
	PageSize    int                  `json:"page_size"`
	TotalItems  int64                `json:"total_items"`
	TotalPages  int                  `json:"total_pages"`
	HasNext     bool                 `json:"has_next"`
	HasPrevious bool                 `json:"has_previous"`
}

// PromoCodeValidationResponse represents the result of promo code validation
type PromoCodeValidationResponse struct {
	IsValid        bool               `json:"is_valid"`
	PromoCode      *PromoCodeResponse `json:"promo_code,omitempty"`
	DiscountAmount float64            `json:"discount_amount"`
	FinalAmount    float64            `json:"final_amount"`
	Message        string             `json:"message,omitempty"`
}

// PromoCodeStatsResponse represents promo code statistics
type PromoCodeStatsResponse struct {
	PromoCodeID     uuid.UUID `json:"promo_code_id"`
	Code            string    `json:"code"`
	TotalUses       int       `json:"total_uses"`
	UniqueUsers     int       `json:"unique_users"`
	TotalDiscount   float64   `json:"total_discount"`
	AverageDiscount float64   `json:"average_discount"`
	TotalRevenue    float64   `json:"total_revenue"`
	ConversionRate  float64   `json:"conversion_rate"`
	RemainingUses   int       `json:"remaining_uses"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToPromoCodeResponse converts a PromoCode model to PromoCodeResponse DTO
func ToPromoCodeResponse(promo *models.PromoCode) *PromoCodeResponse {
	if promo == nil {
		return nil
	}

	now := time.Now()
	isExpired := promo.ExpiresAt != nil && now.After(*promo.ExpiresAt)

	remainingUses := 0
	if promo.MaxUses > 0 {
		remainingUses = promo.MaxUses - promo.UsedCount
		if remainingUses < 0 {
			remainingUses = 0
		}
	}

	usagePercentage := 0.0
	if promo.MaxUses > 0 {
		usagePercentage = (float64(promo.UsedCount) / float64(promo.MaxUses)) * 100
	}

	resp := &PromoCodeResponse{
		ID:                 promo.ID,
		TenantID:           promo.TenantID,
		Code:               promo.Code,
		Description:        promo.Description,
		Type:               promo.Type,
		Value:              promo.Value,
		MaxDiscount:        promo.MaxDiscount,
		MinOrderAmount:     promo.MinOrderAmount,
		StartsAt:           promo.StartsAt,
		ExpiresAt:          promo.ExpiresAt,
		MaxUses:            promo.MaxUses,
		UsedCount:          promo.UsedCount,
		MaxUsesPerUser:     promo.MaxUsesPerUser,
		ApplicableServices: promo.ApplicableServices,
		ApplicableArtisans: promo.ApplicableArtisans,
		IsActive:           promo.IsActive,
		IsValid:            promo.IsValid(),
		IsExpired:          isExpired,
		RemainingUses:      remainingUses,
		UsagePercentage:    usagePercentage,
		CreatedAt:          promo.CreatedAt,
		UpdatedAt:          promo.UpdatedAt,
	}

	return resp
}

// ToPromoCodeResponses converts multiple PromoCode models to DTOs
func ToPromoCodeResponses(promos []*models.PromoCode) []*PromoCodeResponse {
	if promos == nil {
		return nil
	}

	responses := make([]*PromoCodeResponse, len(promos))
	for i, promo := range promos {
		responses[i] = ToPromoCodeResponse(promo)
	}
	return responses
}
