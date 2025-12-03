package service

import (
	"context"
	"fmt"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// InvoiceService defines the interface for invoice operations
type InvoiceService interface {
	// CRUD Operations
	CreateInvoice(ctx context.Context, tenantID uuid.UUID, req *dto.CreateInvoiceRequest) (*dto.InvoiceResponse, error)
	GetInvoice(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.InvoiceResponse, error)
	GetInvoiceByNumber(ctx context.Context, invoiceNumber string, tenantID uuid.UUID) (*dto.InvoiceResponse, error)
	UpdateInvoice(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateInvoiceRequest) (*dto.InvoiceResponse, error)
	DeleteInvoice(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error

	// List & Filter Operations
	ListInvoices(ctx context.Context, filter *dto.InvoiceFilter) (*dto.InvoiceListResponse, error)
	ListInvoicesByCustomer(ctx context.Context, customerID uuid.UUID, tenantID uuid.UUID, page, pageSize int) (*dto.InvoiceListResponse, error)
	ListInvoicesByBooking(ctx context.Context, bookingID uuid.UUID, tenantID uuid.UUID) ([]*dto.InvoiceResponse, error)
	ListOverdueInvoices(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.InvoiceListResponse, error)
	ListUnpaidInvoices(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.InvoiceListResponse, error)

	// Status Operations
	SendInvoice(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
	RecordPayment(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.RecordPaymentRequest) (*dto.InvoiceResponse, error)
	CancelInvoice(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error

	// Statistics & Analytics
	GetInvoiceStats(ctx context.Context, tenantID uuid.UUID) (*dto.InvoiceStatsResponse, error)
}

// invoiceService implements InvoiceService
type invoiceService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(repos *repository.Repositories, logger log.AllLogger) InvoiceService {
	return &invoiceService{
		repos:  repos,
		logger: logger,
	}
}

// CreateInvoice creates a new invoice
func (s *invoiceService) CreateInvoice(ctx context.Context, tenantID uuid.UUID, req *dto.CreateInvoiceRequest) (*dto.InvoiceResponse, error) {
	// Verify customer exists
	customer, err := s.repos.User.GetByID(ctx, req.CustomerID)
	if err != nil {
		s.logger.Error("failed to get customer", "customer_id", req.CustomerID, "error", err)
		return nil, errors.NewNotFoundError("customer")
	}

	// Verify customer belongs to tenant
	if customer.TenantID != nil && *customer.TenantID != tenantID {
		return nil, errors.NewValidationError("Customer does not belong to tenant")
	}

	// Generate invoice number
	invoiceNumber, err := s.repos.Invoice.GenerateInvoiceNumber(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to generate invoice number", "error", err)
		return nil, errors.NewServiceError("GENERATION_FAILED", "Failed to generate invoice number", err)
	}

	// Calculate amounts
	var subtotal float64
	for _, item := range req.LineItems {
		subtotal += item.TotalPrice
	}

	totalAmount := subtotal + req.TaxAmount - req.DiscountAmount

	// Create invoice
	invoice := &models.Invoice{
		TenantID:        tenantID,
		InvoiceNumber:   invoiceNumber,
		BookingID:       req.BookingID,
		CustomerID:      req.CustomerID,
		IssueDate:       req.IssueDate,
		DueDate:         req.DueDate,
		SubtotalAmount:  subtotal,
		TaxAmount:       req.TaxAmount,
		DiscountAmount:  req.DiscountAmount,
		TotalAmount:     totalAmount,
		PaidAmount:      0,
		Currency:        req.Currency,
		Status:          models.InvoiceStatusDraft,
		LineItems:       req.LineItems,
		Notes:           req.Notes,
		TermsConditions: req.TermsConditions,
		Metadata:        req.Metadata,
	}

	if err := s.repos.Invoice.Create(ctx, invoice); err != nil {
		s.logger.Error("failed to create invoice", "error", err)
		return nil, errors.NewRepositoryError("CREATE_FAILED", "Failed to create invoice", err)
	}

	// Load customer relationship
	invoice.Customer = customer

	s.logger.Info("invoice created", "invoice_id", invoice.ID, "invoice_number", invoiceNumber)
	return dto.ToInvoiceResponse(invoice), nil
}

// GetInvoice retrieves an invoice by ID
func (s *invoiceService) GetInvoice(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.InvoiceResponse, error) {
	invoice, err := s.repos.Invoice.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get invoice", "id", id, "error", err)
		return nil, errors.NewNotFoundError("invoice")
	}

	// Verify tenant access
	if invoice.TenantID != tenantID {
		return nil, errors.NewNotFoundError("invoice")
	}

	// Load relationships
	if invoice.CustomerID != uuid.Nil {
		customer, err := s.repos.User.GetByID(ctx, invoice.CustomerID)
		if err == nil {
			invoice.Customer = customer
		}
	}

	return dto.ToInvoiceResponse(invoice), nil
}

// GetInvoiceByNumber retrieves an invoice by invoice number
func (s *invoiceService) GetInvoiceByNumber(ctx context.Context, invoiceNumber string, tenantID uuid.UUID) (*dto.InvoiceResponse, error) {
	invoice, err := s.repos.Invoice.GetByInvoiceNumber(ctx, invoiceNumber)
	if err != nil {
		s.logger.Error("failed to get invoice by number", "invoice_number", invoiceNumber, "error", err)
		return nil, errors.NewNotFoundError("invoice")
	}

	// Verify tenant access
	if invoice.TenantID != tenantID {
		return nil, errors.NewNotFoundError("invoice")
	}

	// Load relationships
	if invoice.CustomerID != uuid.Nil {
		customer, err := s.repos.User.GetByID(ctx, invoice.CustomerID)
		if err == nil {
			invoice.Customer = customer
		}
	}

	return dto.ToInvoiceResponse(invoice), nil
}

// UpdateInvoice updates an invoice
func (s *invoiceService) UpdateInvoice(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateInvoiceRequest) (*dto.InvoiceResponse, error) {
	// Get existing invoice
	invoice, err := s.repos.Invoice.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get invoice", "id", id, "error", err)
		return nil, errors.NewNotFoundError("invoice")
	}

	// Verify tenant access
	if invoice.TenantID != tenantID {
		return nil, errors.NewNotFoundError("invoice")
	}

	// Check if invoice can be updated
	if invoice.Status == models.InvoiceStatusPaid || invoice.Status == models.InvoiceStatusCancelled {
		return nil, errors.NewValidationError(fmt.Sprintf("Cannot update invoice with status %s", invoice.Status))
	}

	// Update fields
	if req.DueDate != nil {
		invoice.DueDate = *req.DueDate
	}
	if req.TaxAmount != nil {
		invoice.TaxAmount = *req.TaxAmount
	}
	if req.DiscountAmount != nil {
		invoice.DiscountAmount = *req.DiscountAmount
	}
	if req.Notes != nil {
		invoice.Notes = *req.Notes
	}
	if req.TermsConditions != nil {
		invoice.TermsConditions = *req.TermsConditions
	}
	if req.Metadata != nil {
		invoice.Metadata = req.Metadata
	}

	// Update line items if provided
	if len(req.LineItems) > 0 {
		if err := s.repos.Invoice.UpdateLineItems(ctx, id, req.LineItems); err != nil {
			s.logger.Error("failed to update line items", "id", id, "error", err)
			return nil, errors.NewRepositoryError("UPDATE_FAILED", "Failed to update line items", err)
		}
		invoice.LineItems = req.LineItems
	}

	// Recalculate amounts
	if err := s.repos.Invoice.RecalculateAmounts(ctx, id); err != nil {
		s.logger.Error("failed to recalculate amounts", "id", id, "error", err)
		return nil, errors.NewRepositoryError("UPDATE_FAILED", "Failed to recalculate amounts", err)
	}

	// Save invoice
	if err := s.repos.Invoice.Update(ctx, invoice); err != nil {
		s.logger.Error("failed to update invoice", "id", id, "error", err)
		return nil, errors.NewRepositoryError("UPDATE_FAILED", "Failed to update invoice", err)
	}

	s.logger.Info("invoice updated", "id", id)

	// Get updated invoice
	updated, err := s.repos.Invoice.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewServiceError("FIND_FAILED", "Failed to retrieve updated invoice", err)
	}

	return dto.ToInvoiceResponse(updated), nil
}

// DeleteInvoice deletes an invoice
func (s *invoiceService) DeleteInvoice(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	// Get invoice
	invoice, err := s.repos.Invoice.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get invoice", "id", id, "error", err)
		return errors.NewNotFoundError("invoice")
	}

	// Verify tenant access
	if invoice.TenantID != tenantID {
		return errors.NewNotFoundError("invoice")
	}

	// Check if invoice can be deleted
	if invoice.Status == models.InvoiceStatusPaid {
		return errors.NewValidationError("Cannot delete paid invoices")
	}

	if err := s.repos.Invoice.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete invoice", "id", id, "error", err)
		return errors.NewRepositoryError("DELETE_FAILED", "Failed to delete invoice", err)
	}

	s.logger.Info("invoice deleted", "id", id)
	return nil
}

// ListInvoices lists invoices with filtering
func (s *invoiceService) ListInvoices(ctx context.Context, filter *dto.InvoiceFilter) (*dto.InvoiceListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     max(filter.Page, 1),
		PageSize: max(filter.PageSize, 1),
	}
	pagination.PageSize = min(pagination.PageSize, 100)

	var invoices []*models.Invoice
	var paginationResult repository.PaginationResult
	var err error

	// Use appropriate repository method based on filters
	if filter.Status != nil {
		invoices, paginationResult, err = s.repos.Invoice.GetByStatus(ctx, filter.TenantID, *filter.Status, pagination)
	} else if filter.IsOverdue != nil && *filter.IsOverdue {
		invoices, paginationResult, err = s.repos.Invoice.GetOverdueInvoices(ctx, filter.TenantID, pagination)
	} else if filter.CustomerID != nil {
		invoices, paginationResult, err = s.repos.Invoice.GetByCustomerID(ctx, *filter.CustomerID, pagination)
	} else {
		invoices, paginationResult, err = s.repos.Invoice.GetByTenantID(ctx, filter.TenantID, pagination)
	}

	if err != nil {
		s.logger.Error("failed to list invoices", "error", err)
		return nil, errors.NewRepositoryError("LIST_FAILED", "Failed to list invoices", err)
	}

	return &dto.InvoiceListResponse{
		Invoices:    dto.ToInvoiceResponses(invoices),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ListInvoicesByCustomer lists invoices for a customer
func (s *invoiceService) ListInvoicesByCustomer(ctx context.Context, customerID uuid.UUID, tenantID uuid.UUID, page, pageSize int) (*dto.InvoiceListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     max(page, 1),
		PageSize: min(max(pageSize, 1), 100),
	}

	invoices, paginationResult, err := s.repos.Invoice.GetByCustomerID(ctx, customerID, pagination)
	if err != nil {
		s.logger.Error("failed to list invoices by customer", "customer_id", customerID, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "Failed to list invoices", err)
	}

	return &dto.InvoiceListResponse{
		Invoices:    dto.ToInvoiceResponses(invoices),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ListInvoicesByBooking lists invoices for a booking
func (s *invoiceService) ListInvoicesByBooking(ctx context.Context, bookingID uuid.UUID, tenantID uuid.UUID) ([]*dto.InvoiceResponse, error) {
	invoices, err := s.repos.Invoice.GetByBookingID(ctx, bookingID)
	if err != nil {
		s.logger.Error("failed to list invoices by booking", "booking_id", bookingID, "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "Failed to list invoices", err)
	}

	return dto.ToInvoiceResponses(invoices), nil
}

// ListOverdueInvoices lists overdue invoices
func (s *invoiceService) ListOverdueInvoices(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.InvoiceListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     max(page, 1),
		PageSize: min(max(pageSize, 1), 100),
	}

	invoices, paginationResult, err := s.repos.Invoice.GetOverdueInvoices(ctx, tenantID, pagination)
	if err != nil {
		s.logger.Error("failed to list overdue invoices", "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "Failed to list overdue invoices", err)
	}

	return &dto.InvoiceListResponse{
		Invoices:    dto.ToInvoiceResponses(invoices),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ListUnpaidInvoices lists unpaid invoices
func (s *invoiceService) ListUnpaidInvoices(ctx context.Context, tenantID uuid.UUID, page, pageSize int) (*dto.InvoiceListResponse, error) {
	pagination := repository.PaginationParams{
		Page:     max(page, 1),
		PageSize: min(max(pageSize, 1), 100),
	}

	invoices, paginationResult, err := s.repos.Invoice.GetUnpaidInvoices(ctx, tenantID, pagination)
	if err != nil {
		s.logger.Error("failed to list unpaid invoices", "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "Failed to list unpaid invoices", err)
	}

	return &dto.InvoiceListResponse{
		Invoices:    dto.ToInvoiceResponses(invoices),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// SendInvoice marks an invoice as sent
func (s *invoiceService) SendInvoice(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	// Get invoice
	invoice, err := s.repos.Invoice.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get invoice", "id", id, "error", err)
		return errors.NewNotFoundError("invoice")
	}

	// Verify tenant access
	if invoice.TenantID != tenantID {
		return errors.NewNotFoundError("invoice")
	}

	// Check if invoice is in draft status
	if invoice.Status != models.InvoiceStatusDraft {
		return errors.NewValidationError(fmt.Sprintf("Cannot send invoice with status %s", invoice.Status))
	}

	if err := s.repos.Invoice.MarkAsSent(ctx, id); err != nil {
		s.logger.Error("failed to mark invoice as sent", "id", id, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "Failed to send invoice", err)
	}

	s.logger.Info("invoice sent", "id", id)
	return nil
}

// RecordPayment records a payment for an invoice
func (s *invoiceService) RecordPayment(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.RecordPaymentRequest) (*dto.InvoiceResponse, error) {
	// Get invoice
	invoice, err := s.repos.Invoice.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get invoice", "id", id, "error", err)
		return nil, errors.NewNotFoundError("invoice")
	}

	// Verify tenant access
	if invoice.TenantID != tenantID {
		return nil, errors.NewNotFoundError("invoice")
	}

	// Validate payment amount
	balanceDue := invoice.GetBalanceDue()
	if req.Amount > balanceDue {
		return nil, errors.NewValidationError("Payment amount exceeds balance due")
	}

	// Record payment
	if err := s.repos.Invoice.RecordPayment(ctx, id, req.Amount, req.PaymentDate); err != nil {
		s.logger.Error("failed to record payment", "id", id, "error", err)
		return nil, errors.NewRepositoryError("UPDATE_FAILED", "Failed to record payment", err)
	}

	s.logger.Info("payment recorded", "id", id, "amount", req.Amount)

	// Get updated invoice
	updated, err := s.repos.Invoice.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewServiceError("FIND_FAILED", "Failed to retrieve updated invoice", err)
	}

	return dto.ToInvoiceResponse(updated), nil
}

// CancelInvoice cancels an invoice
func (s *invoiceService) CancelInvoice(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	// Get invoice
	invoice, err := s.repos.Invoice.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get invoice", "id", id, "error", err)
		return errors.NewNotFoundError("invoice")
	}

	// Verify tenant access
	if invoice.TenantID != tenantID {
		return errors.NewNotFoundError("invoice")
	}

	// Check if invoice can be cancelled
	if invoice.Status == models.InvoiceStatusPaid {
		return errors.NewValidationError("Cannot cancel paid invoices")
	}

	if err := s.repos.Invoice.MarkAsCancelled(ctx, id); err != nil {
		s.logger.Error("failed to cancel invoice", "id", id, "error", err)
		return errors.NewRepositoryError("UPDATE_FAILED", "Failed to cancel invoice", err)
	}

	s.logger.Info("invoice cancelled", "id", id)
	return nil
}

// GetInvoiceStats retrieves invoice statistics
func (s *invoiceService) GetInvoiceStats(ctx context.Context, tenantID uuid.UUID) (*dto.InvoiceStatsResponse, error) {
	stats, err := s.repos.Invoice.GetInvoiceStats(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get invoice stats", "tenant_id", tenantID, "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get invoice stats", err)
	}

	return &dto.InvoiceStatsResponse{
		TenantID:          tenantID,
		TotalInvoices:     stats.TotalInvoices,
		DraftInvoices:     stats.DraftInvoices,
		SentInvoices:      stats.SentInvoices,
		PaidInvoices:      stats.PaidInvoices,
		OverdueInvoices:   stats.OverdueInvoices,
		CancelledInvoices: stats.CancelledInvoices,
		TotalAmount:       stats.TotalRevenue,
		PaidAmount:        stats.TotalRevenue,
		OutstandingAmount: stats.TotalReceivables,
		OverdueAmount:     stats.TotalOverdue,
	}, nil
}
