package dto

import (
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Payment Request DTOs
// ============================================================================

// CreatePaymentRequest represents a request to create a payment
type CreatePaymentRequest struct {
	TenantID          uuid.UUID            `json:"tenant_id" validate:"required"`
	BookingID         uuid.UUID            `json:"booking_id" validate:"required"`
	CustomerID        uuid.UUID            `json:"customer_id" validate:"required"`
	ArtisanID         *uuid.UUID           `json:"artisan_id,omitempty"`
	Amount            float64              `json:"amount" validate:"required,min=0"`
	Currency          string               `json:"currency" validate:"required,len=3"`
	Method            models.PaymentMethod `json:"method" validate:"required"`
	Type              models.PaymentType   `json:"type" validate:"required"`
	ProviderName      string               `json:"provider_name,omitempty"`
	ProviderPaymentID string               `json:"provider_payment_id,omitempty"`
	CommissionRate    float64              `json:"commission_rate" validate:"min=0,max=100"`
	Metadata          models.JSONB         `json:"metadata,omitempty"`
}

// Validate validates the create payment request
func (r *CreatePaymentRequest) Validate() error {
	if r.TenantID == uuid.Nil {
		return ErrTenantIDRequired
	}
	if r.BookingID == uuid.Nil {
		return ErrBookingIDRequired
	}
	if r.CustomerID == uuid.Nil {
		return ErrCustomerIDRequired
	}
	if r.Amount <= 0 {
		return ErrInvalidAmount
	}
	if len(r.Currency) != 3 {
		return ErrInvalidCurrency
	}
	if r.CommissionRate < 0 || r.CommissionRate > 100 {
		return ErrInvalidCommissionRate
	}
	return nil
}

// PaymentFilter represents filters for payment queries
type PaymentFilter struct {
	TenantID        uuid.UUID              `json:"tenant_id"`
	Statuses        []models.PaymentStatus `json:"statuses"`
	Methods         []models.PaymentMethod `json:"methods"`
	Types           []models.PaymentType   `json:"types"`
	MinAmount       *float64               `json:"min_amount"`
	MaxAmount       *float64               `json:"max_amount"`
	CustomerID      *uuid.UUID             `json:"customer_id"`
	ArtisanID       *uuid.UUID             `json:"artisan_id"`
	BookingID       *uuid.UUID             `json:"booking_id"`
	ProcessedAfter  *time.Time             `json:"processed_after"`
	ProcessedBefore *time.Time             `json:"processed_before"`
	HasRefund       *bool                  `json:"has_refund"`
	ProviderName    *string                `json:"provider_name"`
	Page            int                    `json:"page"`
	PageSize        int                    `json:"page_size"`
}

// ============================================================================
// Payment Response DTOs
// ============================================================================

// PaymentListResponse represents a paginated list of payments
type PaymentListResponse struct {
	Payments    []*PaymentResponse `json:"payments"`
	Page        int                `json:"page"`
	PageSize    int                `json:"pageSize"`
	TotalItems  int64              `json:"totalItems"`
	TotalPages  int                `json:"totalPages"`
	HasNext     bool               `json:"hasNext"`
	HasPrevious bool               `json:"hasPrevious"`
}

// RefundRecordResponse represents a refund record
type RefundRecordResponse struct {
	PaymentID    uuid.UUID  `json:"payment_id"`
	RefundAmount float64    `json:"refund_amount"`
	RefundReason string     `json:"refund_reason"`
	RefundedAt   *time.Time `json:"refunded_at"`
	Status       string     `json:"status"`
}

// EarningsResponse represents artisan earnings
type EarningsResponse struct {
	ArtisanID uuid.UUID `json:"artisan_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
	IsPaid    bool      `json:"is_paid"`
}

// RevenueResponse represents platform revenue
type RevenueResponse struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// CommissionBreakdownResponse represents commission breakdown
type CommissionBreakdownResponse struct {
	TenantID              uuid.UUID `json:"tenant_id"`
	TotalPaid             float64   `json:"total_paid"`
	ArtisanTotal          float64   `json:"artisan_total"`
	PlatformTotal         float64   `json:"platform_total"`
	AverageCommissionRate float64   `json:"average_commission_rate"`
	StartDate             time.Time `json:"start_date"`
	EndDate               time.Time `json:"end_date"`
}

// PaymentStatsResponse represents payment statistics
type PaymentStatsResponse struct {
	TenantID                uuid.UUID                      `json:"tenant_id"`
	TotalPayments           int64                          `json:"total_payments"`
	TotalRevenue            float64                        `json:"total_revenue"`
	TotalRefunded           float64                        `json:"total_refunded"`
	AverageTransactionValue float64                        `json:"average_transaction_value"`
	SuccessRate             float64                        `json:"success_rate"`
	PendingCount            int64                          `json:"pending_count"`
	FailedCount             int64                          `json:"failed_count"`
	RefundedCount           int64                          `json:"refunded_count"`
	ByStatus                map[models.PaymentStatus]int64 `json:"by_status"`
	ByMethod                map[models.PaymentMethod]int64 `json:"by_method"`
}

// RevenueDataResponse represents revenue data for a period
type RevenueDataResponse struct {
	Period           time.Time `json:"period"`
	Revenue          float64   `json:"revenue"`
	TransactionCount int64     `json:"transaction_count"`
}

// CustomerPaymentSummaryResponse represents customer payment summary
type CustomerPaymentSummaryResponse struct {
	CustomerID         uuid.UUID            `json:"customer_id"`
	TotalSpent         float64              `json:"total_spent"`
	TotalPayments      int64                `json:"total_payments"`
	SuccessfulPayments int64                `json:"successful_payments"`
	FailedPayments     int64                `json:"failed_payments"`
	AveragePayment     float64              `json:"average_payment"`
	LastPaymentDate    *time.Time           `json:"last_payment_date"`
	PreferredMethod    models.PaymentMethod `json:"preferred_method"`
}

// CustomerPaymentDataResponse represents customer payment data
type CustomerPaymentDataResponse struct {
	CustomerID   uuid.UUID `json:"customer_id"`
	TotalSpent   float64   `json:"total_spent"`
	PaymentCount int64     `json:"payment_count"`
}

// PaymentTrendResponse represents payment trends
type PaymentTrendResponse struct {
	Date             time.Time `json:"date"`
	Revenue          float64   `json:"revenue"`
	TransactionCount int64     `json:"transaction_count"`
	SuccessfulCount  int64     `json:"successful_count"`
	FailedCount      int64     `json:"failed_count"`
}

// ProviderStatsResponse represents provider statistics
type ProviderStatsResponse struct {
	ProviderName      string  `json:"provider_name"`
	TotalTransactions int64   `json:"total_transactions"`
	SuccessfulCount   int64   `json:"successful_count"`
	FailedCount       int64   `json:"failed_count"`
	TotalRevenue      float64 `json:"total_revenue"`
	SuccessRate       float64 `json:"success_rate"`
}

// ============================================================================
// Payment Conversion Functions
// ============================================================================

// ToPaymentResponse converts a Payment model to PaymentResponse DTO
func ToPaymentResponse(payment *models.Payment) *PaymentResponse {
	if payment == nil {
		return nil
	}

	return &PaymentResponse{
		ID:             payment.ID,
		SubscriptionID: uuid.Nil, // Not applicable for booking payments
		Amount:         payment.Amount,
		Currency:       payment.Currency,
		Status:         string(payment.Status),
		Method:         string(payment.Method),
		Description:    fmt.Sprintf("Payment for booking %s", payment.BookingID),
		FailureReason:  payment.FailureReason,
		ProcessedAt:    payment.ProcessedAt,
		CreatedAt:      payment.CreatedAt,
	}
}

// ToPaymentResponses converts multiple Payment models to PaymentResponse DTOs
func ToPaymentResponses(payments []*models.Payment) []*PaymentResponse {
	if payments == nil {
		return nil
	}

	responses := make([]*PaymentResponse, len(payments))
	for i, payment := range payments {
		responses[i] = ToPaymentResponse(payment)
	}
	return responses
}

// ============================================================================
// Validation Errors
// ============================================================================

var (
	ErrTenantIDRequired      = fmt.Errorf("tenant ID is required")
	ErrBookingIDRequired     = fmt.Errorf("booking ID is required")
	ErrCustomerIDRequired    = fmt.Errorf("customer ID is required")
	ErrInvalidAmount         = fmt.Errorf("amount must be positive")
	ErrInvalidCurrency       = fmt.Errorf("currency must be a 3-letter code")
	ErrInvalidCommissionRate = fmt.Errorf("commission rate must be between 0 and 100")
)
