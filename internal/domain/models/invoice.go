package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusPartial   InvoiceStatus = "partial"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

type Invoice struct {
	BaseModel

	// Multi-tenancy
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`

	// Invoice Number
	InvoiceNumber string `json:"invoice_number" gorm:"uniqueIndex;not null;size:50" validate:"required"`

	// References
	BookingID  *uuid.UUID `json:"booking_id,omitempty" gorm:"type:uuid;index"`
	CustomerID uuid.UUID  `json:"customer_id" gorm:"type:uuid;not null;index" validate:"required"`

	// Dates
	IssueDate time.Time  `json:"issue_date" gorm:"not null" validate:"required"`
	DueDate   time.Time  `json:"due_date" gorm:"not null" validate:"required"`
	PaidAt    *time.Time `json:"paid_at,omitempty"`

	// Amounts
	SubtotalAmount float64 `json:"subtotal_amount" gorm:"type:decimal(10,2);not null" validate:"required,min=0"`
	TaxAmount      float64 `json:"tax_amount" gorm:"type:decimal(10,2);default:0"`
	DiscountAmount float64 `json:"discount_amount" gorm:"type:decimal(10,2);default:0"`
	TotalAmount    float64 `json:"total_amount" gorm:"type:decimal(10,2);not null" validate:"required,min=0"`
	PaidAmount     float64 `json:"paid_amount" gorm:"type:decimal(10,2);default:0"`
	Currency       string  `json:"currency" gorm:"size:3;default:'USD'"`

	// Status
	Status InvoiceStatus `json:"status" gorm:"type:varchar(50);not null;default:'draft'" validate:"required"`

	// Line Items
	LineItems []InvoiceLineItem `json:"line_items" gorm:"type:jsonb"`

	// Notes
	Notes           string `json:"notes,omitempty" gorm:"type:text"`
	TermsConditions string `json:"terms_conditions,omitempty" gorm:"type:text"`

	// File
	PDFFileURL string `json:"pdf_file_url,omitempty" gorm:"size:500"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Tenant   *Tenant  `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Booking  *Booking `json:"booking,omitempty" gorm:"foreignKey:BookingID"`
	Customer *User    `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
}

type InvoiceLineItem struct {
	Description string  `json:"description" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" validate:"required,min=0"`
	TotalPrice  float64 `json:"total_price" validate:"required,min=0"`
}

// Scan and Value methods for InvoiceLineItem slice
func (ili *InvoiceLineItem) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &ili)
}

func (ili InvoiceLineItem) Value() (driver.Value, error) {
	return json.Marshal(ili)
}

// Business Methods
func (i *Invoice) IsOverdue() bool {
	return i.Status != InvoiceStatusPaid && time.Now().After(i.DueDate)
}

func (i *Invoice) GetBalanceDue() float64 {
	return i.TotalAmount - i.PaidAmount
}

func (i *Invoice) IsPaid() bool {
	return i.Status == InvoiceStatusPaid
}
