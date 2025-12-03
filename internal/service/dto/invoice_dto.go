package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Invoice Request DTOs
// ============================================================================

// CreateInvoiceRequest represents a request to create an invoice
type CreateInvoiceRequest struct {
	BookingID       *uuid.UUID               `json:"booking_id,omitempty"`
	CustomerID      uuid.UUID                `json:"customer_id" validate:"required"`
	IssueDate       time.Time                `json:"issue_date" validate:"required"`
	DueDate         time.Time                `json:"due_date" validate:"required,gtfield=IssueDate"`
	LineItems       []models.InvoiceLineItem `json:"line_items" validate:"required,min=1"`
	TaxAmount       float64                  `json:"tax_amount" validate:"min=0"`
	DiscountAmount  float64                  `json:"discount_amount" validate:"min=0"`
	Currency        string                   `json:"currency" validate:"required,len=3"`
	Notes           string                   `json:"notes,omitempty"`
	TermsConditions string                   `json:"terms_conditions,omitempty"`
	Metadata        map[string]any           `json:"metadata,omitempty"`
}

// UpdateInvoiceRequest represents a request to update an invoice
type UpdateInvoiceRequest struct {
	DueDate         *time.Time               `json:"due_date,omitempty"`
	LineItems       []models.InvoiceLineItem `json:"line_items,omitempty"`
	TaxAmount       *float64                 `json:"tax_amount,omitempty" validate:"omitempty,min=0"`
	DiscountAmount  *float64                 `json:"discount_amount,omitempty" validate:"omitempty,min=0"`
	Notes           *string                  `json:"notes,omitempty"`
	TermsConditions *string                  `json:"terms_conditions,omitempty"`
	Metadata        map[string]any           `json:"metadata,omitempty"`
}

// RecordPaymentRequest represents a request to record a payment
type RecordPaymentRequest struct {
	Amount      float64   `json:"amount" validate:"required,min=0"`
	PaymentDate time.Time `json:"payment_date" validate:"required"`
	Notes       string    `json:"notes,omitempty"`
}

// InvoiceFilter represents filters for invoice queries
type InvoiceFilter struct {
	TenantID   uuid.UUID             `json:"tenant_id" validate:"required"`
	CustomerID *uuid.UUID            `json:"customer_id,omitempty"`
	BookingID  *uuid.UUID            `json:"booking_id,omitempty"`
	Status     *models.InvoiceStatus `json:"status,omitempty"`
	StartDate  *time.Time            `json:"start_date,omitempty"`
	EndDate    *time.Time            `json:"end_date,omitempty"`
	IsOverdue  *bool                 `json:"is_overdue,omitempty"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
}

// ============================================================================
// Invoice Response DTOs
// ============================================================================

// InvoiceResponse represents an invoice
type InvoiceResponse struct {
	ID              uuid.UUID                `json:"id"`
	TenantID        uuid.UUID                `json:"tenant_id"`
	InvoiceNumber   string                   `json:"invoice_number"`
	BookingID       *uuid.UUID               `json:"booking_id,omitempty"`
	CustomerID      uuid.UUID                `json:"customer_id"`
	IssueDate       time.Time                `json:"issue_date"`
	DueDate         time.Time                `json:"due_date"`
	PaidAt          *time.Time               `json:"paid_at,omitempty"`
	SubtotalAmount  float64                  `json:"subtotal_amount"`
	TaxAmount       float64                  `json:"tax_amount"`
	DiscountAmount  float64                  `json:"discount_amount"`
	TotalAmount     float64                  `json:"total_amount"`
	PaidAmount      float64                  `json:"paid_amount"`
	BalanceDue      float64                  `json:"balance_due"`
	Currency        string                   `json:"currency"`
	Status          models.InvoiceStatus     `json:"status"`
	LineItems       []models.InvoiceLineItem `json:"line_items"`
	Notes           string                   `json:"notes,omitempty"`
	TermsConditions string                   `json:"terms_conditions,omitempty"`
	PDFFileURL      string                   `json:"pdf_file_url,omitempty"`
	IsOverdue       bool                     `json:"is_overdue"`
	IsPaid          bool                     `json:"is_paid"`
	Customer        *UserSummary             `json:"customer,omitempty"`
	Metadata        models.JSONB             `json:"metadata,omitempty"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
}

// InvoiceListResponse represents a paginated list of invoices
type InvoiceListResponse struct {
	Invoices    []*InvoiceResponse `json:"invoices"`
	Page        int                `json:"page"`
	PageSize    int                `json:"page_size"`
	TotalItems  int64              `json:"total_items"`
	TotalPages  int                `json:"total_pages"`
	HasNext     bool               `json:"has_next"`
	HasPrevious bool               `json:"has_previous"`
}

// InvoiceStatsResponse represents invoice statistics
type InvoiceStatsResponse struct {
	TenantID          uuid.UUID `json:"tenant_id"`
	TotalInvoices     int64     `json:"total_invoices"`
	DraftInvoices     int64     `json:"draft_invoices"`
	SentInvoices      int64     `json:"sent_invoices"`
	PaidInvoices      int64     `json:"paid_invoices"`
	OverdueInvoices   int64     `json:"overdue_invoices"`
	CancelledInvoices int64     `json:"cancelled_invoices"`
	TotalAmount       float64   `json:"total_amount"`
	PaidAmount        float64   `json:"paid_amount"`
	OutstandingAmount float64   `json:"outstanding_amount"`
	OverdueAmount     float64   `json:"overdue_amount"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToInvoiceResponse converts an Invoice model to InvoiceResponse DTO
func ToInvoiceResponse(invoice *models.Invoice) *InvoiceResponse {
	if invoice == nil {
		return nil
	}

	resp := &InvoiceResponse{
		ID:              invoice.ID,
		TenantID:        invoice.TenantID,
		InvoiceNumber:   invoice.InvoiceNumber,
		BookingID:       invoice.BookingID,
		CustomerID:      invoice.CustomerID,
		IssueDate:       invoice.IssueDate,
		DueDate:         invoice.DueDate,
		PaidAt:          invoice.PaidAt,
		SubtotalAmount:  invoice.SubtotalAmount,
		TaxAmount:       invoice.TaxAmount,
		DiscountAmount:  invoice.DiscountAmount,
		TotalAmount:     invoice.TotalAmount,
		PaidAmount:      invoice.PaidAmount,
		BalanceDue:      invoice.GetBalanceDue(),
		Currency:        invoice.Currency,
		Status:          invoice.Status,
		LineItems:       invoice.LineItems,
		Notes:           invoice.Notes,
		TermsConditions: invoice.TermsConditions,
		PDFFileURL:      invoice.PDFFileURL,
		IsOverdue:       invoice.IsOverdue(),
		IsPaid:          invoice.IsPaid(),
		Metadata:        invoice.Metadata,
		CreatedAt:       invoice.CreatedAt,
		UpdatedAt:       invoice.UpdatedAt,
	}

	// Add customer if available
	if invoice.Customer != nil {
		resp.Customer = &UserSummary{
			ID:        invoice.Customer.ID,
			FirstName: invoice.Customer.FirstName,
			LastName:  invoice.Customer.LastName,
			Email:     invoice.Customer.Email,
			AvatarURL: invoice.Customer.AvatarURL,
		}
	}

	return resp
}

// ToInvoiceResponses converts multiple Invoice models to DTOs
func ToInvoiceResponses(invoices []*models.Invoice) []*InvoiceResponse {
	if invoices == nil {
		return nil
	}

	responses := make([]*InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		responses[i] = ToInvoiceResponse(invoice)
	}
	return responses
}
