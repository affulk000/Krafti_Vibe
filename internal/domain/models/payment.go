package models

import (
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/google/uuid"
)

type PaymentMethod string

const (
	PaymentMethodCard     PaymentMethod = "card"
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodBank     PaymentMethod = "bank_transfer"
	PaymentMethodWallet   PaymentMethod = "wallet"
	PaymentMethodPayStack PaymentMethod = "paystack"
	PaymentMethodStripe   PaymentMethod = "stripe"
)

type PaymentType string

const (
	PaymentTypeDeposit PaymentType = "deposit"
	PaymentTypeFull    PaymentType = "full"
	PaymentTypeRefund  PaymentType = "refund"
	PaymentTypeTip     PaymentType = "tip"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending       PaymentStatus = "pending"
	PaymentStatusProcessing    PaymentStatus = "processing"
	PaymentStatusPaid          PaymentStatus = "paid"
	PaymentStatusFailed        PaymentStatus = "failed"
	PaymentStatusCancelled     PaymentStatus = "cancelled"
	PaymentStatusRefunded      PaymentStatus = "refunded"
	PaymentStatusPartialRefund PaymentStatus = "partial_refund"
)

type Payment struct {
	BaseModel

	// Multi-tenancy
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`

	// References
	BookingID  uuid.UUID  `json:"booking_id" gorm:"type:uuid;not null;index" validate:"required"`
	CustomerID uuid.UUID  `json:"customer_id" gorm:"type:uuid;not null;index" validate:"required"`
	ArtisanID  *uuid.UUID `json:"artisan_id,omitempty" gorm:"type:uuid;index"`

	// Payment Details
	Amount   float64       `json:"amount" gorm:"type:decimal(10,2);not null" validate:"required,min=0"`
	Currency string        `json:"currency" gorm:"size:3;default:'USD'"`
	Method   PaymentMethod `json:"method" gorm:"type:varchar(50);not null" validate:"required"`
	Type     PaymentType   `json:"type" gorm:"type:varchar(50);not null" validate:"required"`
	Status   PaymentStatus `json:"status" gorm:"type:varchar(50);not null" validate:"required"`

	// External References
	ProviderPaymentID string `json:"provider_payment_id,omitempty" gorm:"size:255;index"` // Stripe, PayPal ID
	ProviderName      string `json:"provider_name,omitempty" gorm:"size:50"`

	// Commission Split
	ArtisanAmount  float64 `json:"artisan_amount" gorm:"type:decimal(10,2);default:0"`
	PlatformAmount float64 `json:"platform_amount" gorm:"type:decimal(10,2);default:0"`
	CommissionRate float64 `json:"commission_rate" gorm:"type:decimal(5,2);default:0"`

	// Processing
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
	FailureReason string     `json:"failure_reason,omitempty" gorm:"type:text"`

	// Refund
	RefundedAmount float64    `json:"refunded_amount" gorm:"type:decimal(10,2);default:0"`
	RefundedAt     *time.Time `json:"refunded_at,omitempty"`
	RefundReason   string     `json:"refund_reason,omitempty" gorm:"type:text"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Booking  *Booking `json:"booking,omitempty" gorm:"foreignKey:BookingID"`
	Customer *User    `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Artisan  *User    `json:"artisan,omitempty" gorm:"foreignKey:ArtisanID"`
}

// TableName specifies the table name for Payment
func (Payment) TableName() string {
	return "payments"
}

// Business Methods

// IsSuccessful checks if the payment was successful
func (p *Payment) IsSuccessful() bool {
	return p.Status == PaymentStatusPaid
}

// IsPending checks if the payment is pending
func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

// IsProcessing checks if the payment is being processed
func (p *Payment) IsProcessing() bool {
	return p.Status == PaymentStatusProcessing
}

// IsFailed checks if the payment failed
func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed
}

// IsCancelled checks if the payment was cancelled
func (p *Payment) IsCancelled() bool {
	return p.Status == PaymentStatusCancelled
}

// IsRefunded checks if the payment was fully refunded
func (p *Payment) IsRefunded() bool {
	return p.Status == PaymentStatusRefunded || p.RefundedAmount >= p.Amount
}

// IsPartiallyRefunded checks if the payment was partially refunded
func (p *Payment) IsPartiallyRefunded() bool {
	return p.Status == PaymentStatusPartialRefund || (p.RefundedAmount > 0 && p.RefundedAmount < p.Amount)
}

// CanBeRefunded checks if the payment can be refunded
func (p *Payment) CanBeRefunded() bool {
	return p.Status == PaymentStatusPaid && p.RefundedAmount < p.Amount
}

// GetRefundableAmount returns the amount that can still be refunded
func (p *Payment) GetRefundableAmount() float64 {
	if !p.CanBeRefunded() {
		return 0
	}
	return p.Amount - p.RefundedAmount
}

// IsFullRefund checks if a refund amount is a full refund
func (p *Payment) IsFullRefund(refundAmount float64) bool {
	return (p.RefundedAmount + refundAmount) >= p.Amount
}

// MarkAsPaid marks the payment as paid and sets processed timestamp
func (p *Payment) MarkAsPaid() {
	p.Status = PaymentStatusPaid
	now := time.Now()
	p.ProcessedAt = &now
}

// MarkAsFailed marks the payment as failed with a reason
func (p *Payment) MarkAsFailed(reason string) {
	p.Status = PaymentStatusFailed
	p.FailureReason = reason
}

// MarkAsCancelled marks the payment as cancelled
func (p *Payment) MarkAsCancelled() {
	p.Status = PaymentStatusCancelled
}

// ProcessRefund processes a refund for this payment
func (p *Payment) ProcessRefund(amount float64, reason string) error {
	if !p.CanBeRefunded() {
		return fmt.Errorf("payment cannot be refunded (status: %s)", p.Status)
	}

	if amount <= 0 {
		return fmt.Errorf("refund amount must be positive")
	}

	if amount > p.GetRefundableAmount() {
		return fmt.Errorf("refund amount (%.2f) exceeds refundable amount (%.2f)",
			amount, p.GetRefundableAmount())
	}

	p.RefundedAmount += amount
	now := time.Now()
	p.RefundedAt = &now
	p.RefundReason = reason

	// Update status based on refund amount
	if p.RefundedAmount >= p.Amount {
		p.Status = PaymentStatusRefunded
	} else {
		p.Status = PaymentStatusPartialRefund
	}

	return nil
}

// CalculateCommission calculates platform and artisan amounts based on commission rate
func (p *Payment) CalculateCommission() {
	if p.CommissionRate > 0 {
		p.PlatformAmount = p.Amount * (p.CommissionRate / 100.0)
		p.ArtisanAmount = p.Amount - p.PlatformAmount
	} else {
		p.ArtisanAmount = p.Amount
		p.PlatformAmount = 0
	}
}

// SetCommissionRate sets the commission rate and recalculates the split
func (p *Payment) SetCommissionRate(rate float64) error {
	if rate < 0 || rate > 100 {
		return fmt.Errorf("commission rate must be between 0 and 100")
	}
	p.CommissionRate = rate
	p.CalculateCommission()
	return nil
}

// IsThirdPartyPayment checks if payment was made through a third-party provider
func (p *Payment) IsThirdPartyPayment() bool {
	return p.Method == PaymentMethodPayStack ||
		p.Method == PaymentMethodStripe ||
		p.ProviderPaymentID != ""
}

// IsCashPayment checks if payment is cash
func (p *Payment) IsCashPayment() bool {
	return p.Method == PaymentMethodCash
}

// GetPaymentAge returns how long ago the payment was made
func (p *Payment) GetPaymentAge() time.Duration {
	if p.ProcessedAt != nil {
		return time.Since(*p.ProcessedAt)
	}
	return time.Since(p.CreatedAt)
}

// CanBeModified checks if the payment can still be modified
func (p *Payment) CanBeModified() bool {
	return p.Status == PaymentStatusPending || p.Status == PaymentStatusProcessing
}

// GetNetAmount returns the net amount after refunds
func (p *Payment) GetNetAmount() float64 {
	return p.Amount - p.RefundedAmount
}

// IsDeposit checks if this is a deposit payment
func (p *Payment) IsDeposit() bool {
	return p.Type == PaymentTypeDeposit
}

// IsFullPayment checks if this is a full payment
func (p *Payment) IsFullPayment() bool {
	return p.Type == PaymentTypeFull
}

// IsTip checks if this is a tip payment
func (p *Payment) IsTip() bool {
	return p.Type == PaymentTypeTip
}

// IsRefundPayment checks if this is a refund payment
func (p *Payment) IsRefundPayment() bool {
	return p.Type == PaymentTypeRefund
}

// GetStatusLabel returns a human-readable status label
func (p *Payment) GetStatusLabel() string {
	switch p.Status {
	case PaymentStatusPending:
		return "Pending"
	case PaymentStatusProcessing:
		return "Processing"
	case PaymentStatusPaid:
		return "Paid"
	case PaymentStatusFailed:
		return "Failed"
	case PaymentStatusCancelled:
		return "Cancelled"
	case PaymentStatusRefunded:
		return "Refunded"
	case PaymentStatusPartialRefund:
		return "Partially Refunded"
	default:
		return string(p.Status)
	}
}

// GetMethodLabel returns a human-readable payment method label
func (p *Payment) GetMethodLabel() string {
	switch p.Method {
	case PaymentMethodCard:
		return "Credit/Debit Card"
	case PaymentMethodCash:
		return "Cash"
	case PaymentMethodBank:
		return "Bank Transfer"
	case PaymentMethodWallet:
		return "Digital Wallet"
	case PaymentMethodPayStack:
		return "Paystack"
	case PaymentMethodStripe:
		return "Stripe"
	default:
		return string(p.Method)
	}
}

// Validate performs business logic validation
func (p *Payment) Validate() error {
	if p.Amount <= 0 {
		return fmt.Errorf("payment amount must be positive")
	}

	if p.BookingID == uuid.Nil {
		return fmt.Errorf("booking ID is required")
	}

	if p.CustomerID == uuid.Nil {
		return fmt.Errorf("customer ID is required")
	}

	if p.CommissionRate < 0 || p.CommissionRate > 100 {
		return fmt.Errorf("commission rate must be between 0 and 100")
	}

	if p.RefundedAmount > p.Amount {
		return fmt.Errorf("refunded amount cannot exceed payment amount")
	}

	// Validate method
	validMethods := []PaymentMethod{
		PaymentMethodCard, PaymentMethodCash, PaymentMethodBank,
		PaymentMethodWallet, PaymentMethodPayStack, PaymentMethodStripe,
	}
	if !slices.Contains(validMethods, p.Method) {
		return fmt.Errorf("invalid payment method: %s", p.Method)
	}

	// Validate type
	validTypes := []PaymentType{
		PaymentTypeDeposit, PaymentTypeFull, PaymentTypeRefund, PaymentTypeTip,
	}

	if !slices.Contains(validTypes, p.Type) {
		return fmt.Errorf("invalid payment type: %s", p.Type)
	}

	return nil
}

// String returns a string representation of the payment
func (p *Payment) String() string {
	return fmt.Sprintf("Payment{ID: %s, Amount: %.2f %s, Status: %s, Method: %s}",
		p.ID, p.Amount, p.Currency, p.Status, p.Method)
}

// Clone creates a copy of the payment
func (p *Payment) Clone() *Payment {
	clone := *p

	// Clone pointer fields
	if p.ArtisanID != nil {
		id := *p.ArtisanID
		clone.ArtisanID = &id
	}

	if p.ProcessedAt != nil {
		t := *p.ProcessedAt
		clone.ProcessedAt = &t
	}

	if p.RefundedAt != nil {
		t := *p.RefundedAt
		clone.RefundedAt = &t
	}

	// Clone metadata
	if p.Metadata != nil {
		clone.Metadata = make(JSONB)
		maps.Copy(clone.Metadata, p.Metadata)
	}

	return &clone
}

// GetTransactionReference generates a unique transaction reference
func (p *Payment) GetTransactionReference() string {
	if p.ProviderPaymentID != "" {
		return p.ProviderPaymentID
	}
	return fmt.Sprintf("PAY-%s", p.ID.String()[:8])
}

// HasArtisan checks if payment has an associated artisan
func (p *Payment) HasArtisan() bool {
	return p.ArtisanID != nil && *p.ArtisanID != uuid.Nil
}

// GetRefundPercentage returns the percentage of amount that has been refunded
func (p *Payment) GetRefundPercentage() float64 {
	if p.Amount == 0 {
		return 0
	}
	return (p.RefundedAmount / p.Amount) * 100
}

// IsProcessed checks if the payment has been processed
func (p *Payment) IsProcessed() bool {
	return p.ProcessedAt != nil
}

// GetProcessingTime returns the time taken to process the payment
func (p *Payment) GetProcessingTime() time.Duration {
	if p.ProcessedAt == nil {
		return 0
	}
	return p.ProcessedAt.Sub(p.CreatedAt)
}
