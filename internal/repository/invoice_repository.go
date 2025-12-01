package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InvoiceRepository defines the interface for invoice repository operations
type InvoiceRepository interface {
	BaseRepository[models.Invoice]

	// Core Operations
	GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error)
	GetByCustomerID(ctx context.Context, customerID uuid.UUID, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*models.Invoice, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error)
	GenerateInvoiceNumber(ctx context.Context, tenantID uuid.UUID) (string, error)

	// Status Operations
	MarkAsPaid(ctx context.Context, invoiceID uuid.UUID, paidAt time.Time) error
	MarkAsPartiallyPaid(ctx context.Context, invoiceID uuid.UUID, paidAmount float64) error
	MarkAsSent(ctx context.Context, invoiceID uuid.UUID) error
	MarkAsCancelled(ctx context.Context, invoiceID uuid.UUID) error
	UpdateStatus(ctx context.Context, invoiceID uuid.UUID, status models.InvoiceStatus) error

	// Status Queries
	GetByStatus(ctx context.Context, tenantID uuid.UUID, status models.InvoiceStatus, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error)
	GetDraftInvoices(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error)
	GetUnpaidInvoices(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error)
	GetOverdueInvoices(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error)
	GetPaidInvoices(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error)

	// Date-based Queries
	GetInvoicesDueInRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*models.Invoice, error)
	GetInvoicesIssuedInRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error)
	GetInvoicesDueSoon(ctx context.Context, tenantID uuid.UUID, days int) ([]*models.Invoice, error)

	// Payment Operations
	RecordPayment(ctx context.Context, invoiceID uuid.UUID, amount float64, paidAt time.Time) error
	GetTotalReceivables(ctx context.Context, tenantID uuid.UUID) (float64, error)
	GetTotalOverdue(ctx context.Context, tenantID uuid.UUID) (float64, error)
	GetCustomerBalance(ctx context.Context, customerID uuid.UUID) (float64, error)

	// Line Items
	UpdateLineItems(ctx context.Context, invoiceID uuid.UUID, lineItems []models.InvoiceLineItem) error
	RecalculateAmounts(ctx context.Context, invoiceID uuid.UUID) error

	// PDF & Files
	UpdatePDFURL(ctx context.Context, invoiceID uuid.UUID, pdfURL string) error
	GetInvoicesWithoutPDF(ctx context.Context, tenantID uuid.UUID) ([]*models.Invoice, error)

	// Analytics & Reporting
	GetInvoiceStats(ctx context.Context, tenantID uuid.UUID) (InvoiceStats, error)
	GetRevenueByPeriod(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, groupBy string) ([]InvoiceRevenueData, error)
	GetCustomerInvoiceSummary(ctx context.Context, customerID uuid.UUID) (CustomerInvoiceSummary, error)
	GetAverageDaysToPay(ctx context.Context, tenantID uuid.UUID) (float64, error)
	GetTopCustomersByRevenue(ctx context.Context, tenantID uuid.UUID, limit int, startDate, endDate time.Time) ([]CustomerRevenueData, error)
	GetInvoiceAgingReport(ctx context.Context, tenantID uuid.UUID) (InvoiceAgingReport, error)

	// Search & Filter
	Search(ctx context.Context, query string, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error)
	FindByFilters(ctx context.Context, filters InvoiceFilters, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error)

	// Bulk Operations
	BulkMarkAsSent(ctx context.Context, invoiceIDs []uuid.UUID) error
	BulkMarkAsCancelled(ctx context.Context, invoiceIDs []uuid.UUID) error
	BulkUpdateDueDate(ctx context.Context, invoiceIDs []uuid.UUID, newDueDate time.Time) error
}

// InvoiceStats represents comprehensive invoice statistics
type InvoiceStats struct {
	TotalInvoices       int64                          `json:"total_invoices"`
	DraftInvoices       int64                          `json:"draft_invoices"`
	SentInvoices        int64                          `json:"sent_invoices"`
	PaidInvoices        int64                          `json:"paid_invoices"`
	OverdueInvoices     int64                          `json:"overdue_invoices"`
	PartialInvoices     int64                          `json:"partial_invoices"`
	CancelledInvoices   int64                          `json:"cancelled_invoices"`
	TotalRevenue        float64                        `json:"total_revenue"`
	TotalReceivables    float64                        `json:"total_receivables"`
	TotalOverdue        float64                        `json:"total_overdue"`
	AverageInvoiceValue float64                        `json:"average_invoice_value"`
	AverageDaysToPay    float64                        `json:"average_days_to_pay"`
	ByStatus            map[models.InvoiceStatus]int64 `json:"by_status"`
	ThisMonthRevenue    float64                        `json:"this_month_revenue"`
	LastMonthRevenue    float64                        `json:"last_month_revenue"`
}

// InvoiceRevenueData represents revenue data for a period
type InvoiceRevenueData struct {
	Period       string  `json:"period"`
	Revenue      float64 `json:"revenue"`
	InvoiceCount int64   `json:"invoice_count"`
	PaidCount    int64   `json:"paid_count"`
	AverageValue float64 `json:"average_value"`
}

// CustomerInvoiceSummary represents invoice summary for a customer
type CustomerInvoiceSummary struct {
	CustomerID       uuid.UUID  `json:"customer_id"`
	CustomerName     string     `json:"customer_name"`
	TotalInvoices    int64      `json:"total_invoices"`
	TotalAmount      float64    `json:"total_amount"`
	TotalPaid        float64    `json:"total_paid"`
	TotalOutstanding float64    `json:"total_outstanding"`
	OverdueAmount    float64    `json:"overdue_amount"`
	OverdueCount     int64      `json:"overdue_count"`
	AverageDaysToPay float64    `json:"average_days_to_pay"`
	LastInvoiceDate  *time.Time `json:"last_invoice_date,omitempty"`
	LastPaymentDate  *time.Time `json:"last_payment_date,omitempty"`
}

// CustomerRevenueData represents revenue data per customer
type CustomerRevenueData struct {
	CustomerID   uuid.UUID `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	Revenue      float64   `json:"revenue"`
	InvoiceCount int64     `json:"invoice_count"`
	AverageValue float64   `json:"average_value"`
}

// InvoiceAgingReport represents aging analysis of invoices
type InvoiceAgingReport struct {
	Current         AgingBucket `json:"current"`       // 0-30 days
	Days30to60      AgingBucket `json:"days_30_to_60"` // 31-60 days
	Days60to90      AgingBucket `json:"days_60_to_90"` // 61-90 days
	Over90Days      AgingBucket `json:"over_90_days"`  // 90+ days
	TotalReceivable float64     `json:"total_receivable"`
}

// AgingBucket represents an aging bucket
type AgingBucket struct {
	Amount float64 `json:"amount"`
	Count  int64   `json:"count"`
}

// InvoiceFilters for advanced filtering
type InvoiceFilters struct {
	TenantID      uuid.UUID              `json:"tenant_id"`
	CustomerIDs   []uuid.UUID            `json:"customer_ids"`
	Statuses      []models.InvoiceStatus `json:"statuses"`
	MinAmount     *float64               `json:"min_amount"`
	MaxAmount     *float64               `json:"max_amount"`
	IssueDateFrom *time.Time             `json:"issue_date_from"`
	IssueDateTo   *time.Time             `json:"issue_date_to"`
	DueDateFrom   *time.Time             `json:"due_date_from"`
	DueDateTo     *time.Time             `json:"due_date_to"`
	IsOverdue     *bool                  `json:"is_overdue"`
	HasBooking    *bool                  `json:"has_booking"`
}

// invoiceRepository implements InvoiceRepository
type invoiceRepository struct {
	BaseRepository[models.Invoice]
	db      *gorm.DB
	logger  log.AllLogger
	cache   Cache
	metrics MetricsCollector
}

// NewInvoiceRepository creates a new InvoiceRepository instance
func NewInvoiceRepository(db *gorm.DB, config ...RepositoryConfig) InvoiceRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 5 * time.Minute
	}

	baseRepo := NewBaseRepository[models.Invoice](db, cfg)

	return &invoiceRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
		metrics:        cfg.Metrics,
	}
}

// ------------------------------------------------------------------------------------------------
// CORE & STATUS IMPLEMENTATIONS (from previous turn, included for completeness)
// ------------------------------------------------------------------------------------------------

// GetByInvoiceNumber retrieves an invoice by invoice number
func (r *invoiceRepository) GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("get_by_invoice_number", "invoices", time.Since(start), nil)
		}
	}()

	if invoiceNumber == "" {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "invoice number cannot be empty", errors.ErrInvalidInput)
	}

	// Try cache first
	cacheKey := r.getCacheKey("number", invoiceNumber)
	if r.cache != nil {
		var invoice models.Invoice
		if err := r.cache.GetJSON(ctx, cacheKey, &invoice); err == nil {
			if r.metrics != nil {
				r.metrics.RecordCacheHit("invoices")
			}
			return &invoice, nil
		}
		if r.metrics != nil {
			r.metrics.RecordCacheMiss("invoices")
		}
	}

	var invoice models.Invoice
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Booking").
		Where("invoice_number = ?", invoiceNumber).
		First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "invoice not found", errors.ErrNotFound)
		}
		r.logger.Error("failed to get invoice by number", "invoice_number", invoiceNumber, "error", err)
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get invoice", err)
	}

	// Cache the result
	if r.cache != nil {
		if err := r.cache.SetJSON(ctx, cacheKey, invoice, 5*time.Minute); err != nil {
			r.logger.Warn("failed to cache invoice", "invoice_number", invoiceNumber, "error", err)
		}
	}

	return &invoice, nil
}

// GetByCustomerID retrieves all invoices for a customer
func (r *invoiceRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Invoice{}).Where("customer_id = ?", customerID)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count invoices", err)
	}

	var invoices []*models.Invoice
	if err := query.
		Preload("Customer").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("issue_date DESC").
		Find(&invoices).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find invoices", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return invoices, paginationResult, nil
}

// GetByBookingID retrieves all invoices for a booking
func (r *invoiceRepository) GetByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*models.Invoice, error) {
	var invoices []*models.Invoice
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Booking").
		Where("booking_id = ?", bookingID).
		Order("issue_date DESC").
		Find(&invoices).Error; err != nil {
		r.logger.Error("failed to get invoices by booking ID", "booking_id", bookingID, "error", err)
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get invoices", err)
	}
	return invoices, nil
}

// GetByTenantID retrieves all invoices for a tenant
func (r *invoiceRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Invoice{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count invoices", err)
	}

	var invoices []*models.Invoice
	if err := query.
		Preload("Customer").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("issue_date DESC").
		Find(&invoices).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find invoices", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return invoices, paginationResult, nil
}

// GenerateInvoiceNumber generates a unique invoice number
func (r *invoiceRepository) GenerateInvoiceNumber(ctx context.Context, tenantID uuid.UUID) (string, error) {
	now := time.Now()
	prefix := fmt.Sprintf("INV-%d%02d-", now.Year(), now.Month())

	// Get the count of invoices for this month
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Where("tenant_id = ? AND invoice_number LIKE ?", tenantID, prefix+"%").
		Count(&count).Error; err != nil {
		return "", errors.NewRepositoryError("GENERATION_FAILED", "failed to generate invoice number", err)
	}

	invoiceNumber := fmt.Sprintf("%s%04d", prefix, count+1)
	return invoiceNumber, nil
}

// MarkAsPaid marks an invoice as paid
func (r *invoiceRepository) MarkAsPaid(ctx context.Context, invoiceID uuid.UUID, paidAt time.Time) error {
	invoice, err := r.GetByID(ctx, invoiceID)
	if err != nil {
		return err
	}

	invoice.Status = models.InvoiceStatusPaid
	invoice.PaidAmount = invoice.TotalAmount
	invoice.PaidAt = &paidAt

	if err := r.Update(ctx, invoice); err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateInvoiceCache(ctx, invoiceID, invoice.InvoiceNumber)

	r.logger.Info("invoice marked as paid", "invoice_id", invoiceID, "invoice_number", invoice.InvoiceNumber)
	return nil
}

// MarkAsPartiallyPaid marks an invoice as partially paid
func (r *invoiceRepository) MarkAsPartiallyPaid(ctx context.Context, invoiceID uuid.UUID, paidAmount float64) error {
	invoice, err := r.GetByID(ctx, invoiceID)
	if err != nil {
		return err
	}

	if paidAmount <= 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "paid amount must be positive", errors.ErrInvalidInput)
	}

	if paidAmount > invoice.TotalAmount {
		return errors.NewRepositoryError("INVALID_INPUT", "paid amount exceeds total amount", errors.ErrInvalidInput)
	}

	invoice.PaidAmount += paidAmount

	if invoice.PaidAmount >= invoice.TotalAmount {
		invoice.Status = models.InvoiceStatusPaid
		now := time.Now()
		invoice.PaidAt = &now
	} else {
		invoice.Status = models.InvoiceStatusPartial
	}

	if err := r.Update(ctx, invoice); err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateInvoiceCache(ctx, invoiceID, invoice.InvoiceNumber)

	r.logger.Info("invoice marked as partially paid", "invoice_id", invoiceID, "amount", paidAmount)
	return nil
}

// MarkAsSent marks an invoice as sent
func (r *invoiceRepository) MarkAsSent(ctx context.Context, invoiceID uuid.UUID) error {
	return r.UpdateStatus(ctx, invoiceID, models.InvoiceStatusSent)
}

// MarkAsCancelled marks an invoice as cancelled
func (r *invoiceRepository) MarkAsCancelled(ctx context.Context, invoiceID uuid.UUID) error {
	return r.UpdateStatus(ctx, invoiceID, models.InvoiceStatusCancelled)
}

// UpdateStatus updates an invoice status
func (r *invoiceRepository) UpdateStatus(ctx context.Context, invoiceID uuid.UUID, status models.InvoiceStatus) error {
	result := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Where("id = ?", invoiceID).
		Update("status", status)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update status", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "invoice not found", errors.ErrNotFound)
	}

	r.logger.Info("invoice status updated", "invoice_id", invoiceID, "status", status)
	return nil
}

// GetByStatus retrieves invoices by status
func (r *invoiceRepository) GetByStatus(ctx context.Context, tenantID uuid.UUID, status models.InvoiceStatus, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Where("tenant_id = ? AND status = ?", tenantID, status)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count invoices", err)
	}

	var invoices []*models.Invoice
	if err := query.
		Preload("Customer").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("issue_date DESC").
		Find(&invoices).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find invoices", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return invoices, paginationResult, nil
}

// GetDraftInvoices retrieves all draft invoices
func (r *invoiceRepository) GetDraftInvoices(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error) {
	return r.GetByStatus(ctx, tenantID, models.InvoiceStatusDraft, pagination)
}

// GetUnpaidInvoices retrieves all unpaid invoices
func (r *invoiceRepository) GetUnpaidInvoices(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Where("tenant_id = ? AND status IN ?", tenantID, []models.InvoiceStatus{
			models.InvoiceStatusSent,
			models.InvoiceStatusPartial,
			models.InvoiceStatusOverdue,
		})

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count invoices", err)
	}

	var invoices []*models.Invoice
	if err := query.
		Preload("Customer").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("due_date ASC").
		Find(&invoices).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find invoices", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return invoices, paginationResult, nil
}

// GetOverdueInvoices retrieves all overdue invoices
func (r *invoiceRepository) GetOverdueInvoices(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error) {
	pagination.Validate()

	now := time.Now()
	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Where("tenant_id = ? AND status NOT IN ? AND due_date < ?",
			tenantID,
			[]models.InvoiceStatus{models.InvoiceStatusPaid, models.InvoiceStatusCancelled},
			now)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count invoices", err)
	}

	var invoices []*models.Invoice
	if err := query.
		Preload("Customer").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("due_date ASC").
		Find(&invoices).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find invoices", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return invoices, paginationResult, nil
}

// GetPaidInvoices retrieves paid invoices for a tenant within a date range
func (r *invoiceRepository) GetPaidInvoices(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error) {
	pagination.Validate()

	countQuery := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.InvoiceStatusPaid)
	countQuery = r.applyDateRange(countQuery, "paid_at", startDate, endDate)

	var totalItems int64
	if err := countQuery.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count paid invoices", err)
	}

	dataQuery := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.InvoiceStatusPaid)
	dataQuery = r.applyDateRange(dataQuery, "paid_at", startDate, endDate)

	var invoices []*models.Invoice
	if err := dataQuery.
		Preload("Customer").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("paid_at DESC").
		Find(&invoices).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find paid invoices", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return invoices, paginationResult, nil
}

// GetInvoicesDueInRange retrieves invoices due within a range
func (r *invoiceRepository) GetInvoicesDueInRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*models.Invoice, error) {
	query := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Where("tenant_id = ?", tenantID)
	query = r.applyDateRange(query, "due_date", startDate, endDate)

	var invoices []*models.Invoice
	if err := query.
		Preload("Customer").
		Preload("Booking").
		Order("due_date ASC").
		Find(&invoices).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find invoices by due date", err)
	}

	return invoices, nil
}

// GetInvoicesIssuedInRange retrieves invoices issued in a date range
func (r *invoiceRepository) GetInvoicesIssuedInRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error) {
	pagination.Validate()

	countQuery := r.db.WithContext(ctx).Model(&models.Invoice{}).Where("tenant_id = ?", tenantID)
	countQuery = r.applyDateRange(countQuery, "issue_date", startDate, endDate)

	var totalItems int64
	if err := countQuery.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count invoices", err)
	}

	dataQuery := r.db.WithContext(ctx).Model(&models.Invoice{}).Where("tenant_id = ?", tenantID)
	dataQuery = r.applyDateRange(dataQuery, "issue_date", startDate, endDate)

	var invoices []*models.Invoice
	if err := dataQuery.
		Preload("Customer").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("issue_date DESC").
		Find(&invoices).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find invoices", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return invoices, paginationResult, nil
}

// GetInvoicesDueSoon retrieves invoices due within the next N days
func (r *invoiceRepository) GetInvoicesDueSoon(ctx context.Context, tenantID uuid.UUID, days int) ([]*models.Invoice, error) {
	if days <= 0 {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "days must be greater than zero", errors.ErrInvalidInput)
	}

	now := time.Now()
	deadline := now.AddDate(0, 0, days)

	query := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Where("tenant_id = ? AND status NOT IN ? AND due_date BETWEEN ? AND ?",
			tenantID,
			[]models.InvoiceStatus{models.InvoiceStatusPaid, models.InvoiceStatusCancelled},
			now,
			deadline)

	var invoices []*models.Invoice
	if err := query.
		Preload("Customer").
		Preload("Booking").
		Order("due_date ASC").
		Find(&invoices).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find due soon invoices", err)
	}

	return invoices, nil
}

// RecordPayment records a payment against an invoice
func (r *invoiceRepository) RecordPayment(ctx context.Context, invoiceID uuid.UUID, amount float64, paidAt time.Time) error {
	if amount <= 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "amount must be greater than zero", errors.ErrInvalidInput)
	}

	var invoice models.Invoice
	if err := r.db.WithContext(ctx).First(&invoice, "id = ?", invoiceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "invoice not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("GET_FAILED", "failed to load invoice", err)
	}

	invoice.PaidAmount += amount
	if invoice.PaidAmount >= invoice.TotalAmount {
		invoice.PaidAmount = invoice.TotalAmount
		invoice.Status = models.InvoiceStatusPaid
		if paidAt.IsZero() {
			now := time.Now()
			invoice.PaidAt = &now
		} else {
			invoice.PaidAt = &paidAt
		}
	} else {
		invoice.Status = models.InvoiceStatusPartial
	}

	if err := r.db.WithContext(ctx).Save(&invoice).Error; err != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to record payment", err)
	}

	r.invalidateInvoiceCache(ctx, invoice.ID, invoice.InvoiceNumber)
	return nil
}

// GetTotalReceivables calculates outstanding receivables
func (r *invoiceRepository) GetTotalReceivables(ctx context.Context, tenantID uuid.UUID) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Select("COALESCE(SUM(total_amount - paid_amount), 0)").
		Where("tenant_id = ? AND status NOT IN ?", tenantID, []models.InvoiceStatus{
			models.InvoiceStatusPaid,
			models.InvoiceStatusCancelled,
		}).
		Scan(&total).Error
	if err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to calculate receivables", err)
	}
	return total, nil
}

// GetTotalOverdue calculates overdue amount
func (r *invoiceRepository) GetTotalOverdue(ctx context.Context, tenantID uuid.UUID) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Select("COALESCE(SUM(total_amount - paid_amount), 0)").
		Where("tenant_id = ? AND status NOT IN ? AND due_date < ?", tenantID, []models.InvoiceStatus{
			models.InvoiceStatusPaid,
			models.InvoiceStatusCancelled,
		}, time.Now()).
		Scan(&total).Error
	if err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to calculate overdue amount", err)
	}
	return total, nil
}

// GetCustomerBalance returns outstanding balance for a customer
func (r *invoiceRepository) GetCustomerBalance(ctx context.Context, customerID uuid.UUID) (float64, error) {
	var balance float64
	err := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Select("COALESCE(SUM(total_amount - paid_amount), 0)").
		Where("customer_id = ? AND status NOT IN ?", customerID, []models.InvoiceStatus{
			models.InvoiceStatusPaid,
			models.InvoiceStatusCancelled,
		}).
		Scan(&balance).Error
	if err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to calculate customer balance", err)
	}
	return balance, nil
}

// UpdateLineItems replaces invoice line items and recalculates totals
func (r *invoiceRepository) UpdateLineItems(ctx context.Context, invoiceID uuid.UUID, lineItems []models.InvoiceLineItem) error {
	if len(lineItems) == 0 {
		return errors.NewRepositoryError("INVALID_INPUT", "line items cannot be empty", errors.ErrInvalidInput)
	}

	var invoice models.Invoice
	if err := r.db.WithContext(ctx).First(&invoice, "id = ?", invoiceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "invoice not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("GET_FAILED", "failed to load invoice", err)
	}

	subtotal := calculateLineItemSubtotal(lineItems)
	total := subtotal + invoice.TaxAmount - invoice.DiscountAmount
	if total < 0 {
		total = 0
	}

	update := map[string]interface{}{
		"line_items":      lineItems,
		"subtotal_amount": subtotal,
		"total_amount":    total,
	}

	result := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Where("id = ?", invoiceID).
		Updates(update)
	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update line items", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "invoice not found", errors.ErrNotFound)
	}

	r.invalidateInvoiceCache(ctx, invoice.ID, invoice.InvoiceNumber)
	return nil
}

// RecalculateAmounts recalculates invoice totals
func (r *invoiceRepository) RecalculateAmounts(ctx context.Context, invoiceID uuid.UUID) error {
	var invoice models.Invoice
	if err := r.db.WithContext(ctx).First(&invoice, "id = ?", invoiceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "invoice not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("GET_FAILED", "failed to load invoice", err)
	}

	subtotal := calculateLineItemSubtotal(invoice.LineItems)
	total := subtotal + invoice.TaxAmount - invoice.DiscountAmount
	if total < 0 {
		total = 0
	}

	result := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Where("id = ?", invoiceID).
		Updates(map[string]interface{}{
			"subtotal_amount": subtotal,
			"total_amount":    total,
		})
	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to recalculate invoice", result.Error)
	}

	r.invalidateInvoiceCache(ctx, invoice.ID, invoice.InvoiceNumber)
	return nil
}

// UpdatePDFURL updates the stored PDF URL
func (r *invoiceRepository) UpdatePDFURL(ctx context.Context, invoiceID uuid.UUID, pdfURL string) error {
	result := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Where("id = ?", invoiceID).
		Update("pdf_file_url", pdfURL)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update pdf url", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "invoice not found", errors.ErrNotFound)
	}

	r.InvalidateCache(ctx, invoiceID)
	return nil
}

// GetInvoicesWithoutPDF finds invoices missing PDF artifacts
func (r *invoiceRepository) GetInvoicesWithoutPDF(ctx context.Context, tenantID uuid.UUID) ([]*models.Invoice, error) {
	var invoices []*models.Invoice
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND (pdf_file_url = '' OR pdf_file_url IS NULL)", tenantID).
		Order("issue_date DESC").
		Find(&invoices).Error
	if err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find invoices without pdf", err)
	}
	return invoices, nil
}

// GetInvoiceStats aggregates stats for a tenant
func (r *invoiceRepository) GetInvoiceStats(ctx context.Context, tenantID uuid.UUID) (InvoiceStats, error) {
	stats := InvoiceStats{
		ByStatus: make(map[models.InvoiceStatus]int64),
	}

	if err := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalInvoices).Error; err != nil {
		return stats, errors.NewRepositoryError("COUNT_FAILED", "failed to count invoices", err)
	}

	type statusCount struct {
		Status models.InvoiceStatus
		Count  int64
	}
	var counts []statusCount
	if err := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("status").
		Find(&counts).Error; err != nil {
		return stats, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to aggregate statuses", err)
	}
	for _, c := range counts {
		stats.ByStatus[c.Status] = c.Count
		switch c.Status {
		case models.InvoiceStatusDraft:
			stats.DraftInvoices = c.Count
		case models.InvoiceStatusSent:
			stats.SentInvoices = c.Count
		case models.InvoiceStatusPaid:
			stats.PaidInvoices = c.Count
		case models.InvoiceStatusOverdue:
			stats.OverdueInvoices = c.Count
		case models.InvoiceStatusPartial:
			stats.PartialInvoices = c.Count
		case models.InvoiceStatusCancelled:
			stats.CancelledInvoices = c.Count
		}
	}

	var avgValue float64
	if err := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Select("COALESCE(AVG(total_amount), 0)").
		Where("tenant_id = ?", tenantID).
		Scan(&avgValue).Error; err == nil {
		stats.AverageInvoiceValue = avgValue
	}

	totalRevenue, _ := r.sumRevenueInRange(ctx, tenantID, time.Time{}, time.Time{})
	stats.TotalRevenue = totalRevenue

	if receivables, err := r.GetTotalReceivables(ctx, tenantID); err == nil {
		stats.TotalReceivables = receivables
	}

	if overdue, err := r.GetTotalOverdue(ctx, tenantID); err == nil {
		stats.TotalOverdue = overdue
	}

	if avgDays, err := r.GetAverageDaysToPay(ctx, tenantID); err == nil {
		stats.AverageDaysToPay = avgDays
	}

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := startOfMonth.AddDate(0, 1, 0)
	lastMonth := startOfMonth.AddDate(0, -1, 0)

	stats.ThisMonthRevenue, _ = r.sumRevenueInRange(ctx, tenantID, startOfMonth, nextMonth)
	stats.LastMonthRevenue, _ = r.sumRevenueInRange(ctx, tenantID, lastMonth, startOfMonth)

	return stats, nil
}

// GetRevenueByPeriod groups revenue by a period granularity
func (r *invoiceRepository) GetRevenueByPeriod(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, groupBy string) ([]InvoiceRevenueData, error) {
	if endDate.IsZero() {
		endDate = time.Now()
	}
	dateTrunc := "month"
	format := "YYYY-MM"

	switch strings.ToLower(groupBy) {
	case "day":
		dateTrunc = "day"
		format = "YYYY-MM-DD"
	case "week":
		dateTrunc = "week"
		format = "IYYY-IW"
	}

	var results []InvoiceRevenueData
	err := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Select(fmt.Sprintf("to_char(date_trunc('%s', issue_date), '%s') as period, COALESCE(SUM(total_amount), 0) as revenue, COUNT(*) as invoice_count, SUM(CASE WHEN status = '%s' THEN 1 ELSE 0 END) as paid_count, COALESCE(AVG(total_amount), 0) as average_value",
			dateTrunc, format, models.InvoiceStatusPaid)).
		Where("tenant_id = ? AND issue_date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Group("period").
		Order("period").
		Scan(&results).Error
	if err != nil {
		return nil, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to calculate revenue by period", err)
	}

	return results, nil
}

// GetCustomerInvoiceSummary returns aggregates for a customer
func (r *invoiceRepository) GetCustomerInvoiceSummary(ctx context.Context, customerID uuid.UUID) (CustomerInvoiceSummary, error) {
	summary := CustomerInvoiceSummary{
		CustomerID: customerID,
	}

	err := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Select("COALESCE(SUM(total_amount), 0) as total_amount, COALESCE(SUM(paid_amount), 0) as total_paid, COUNT(*) as total_invoices, COALESCE(SUM(total_amount - paid_amount), 0) as total_outstanding, COALESCE(SUM(CASE WHEN due_date < NOW() AND status NOT IN ('paid','cancelled') THEN total_amount - paid_amount ELSE 0 END), 0) as overdue_amount, COALESCE(SUM(CASE WHEN due_date < NOW() AND status NOT IN ('paid','cancelled') THEN 1 ELSE 0 END),0) as overdue_count").
		Where("customer_id = ?", customerID).
		Scan(&summary).Error
	if err != nil {
		return summary, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to summarize customer invoices", err)
	}

	var lastInvoice time.Time
	if err := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Select("MAX(issue_date)").
		Where("customer_id = ?", customerID).
		Scan(&lastInvoice).Error; err == nil && !lastInvoice.IsZero() {
		summary.LastInvoiceDate = &lastInvoice
	}

	var lastPayment time.Time
	if err := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Select("MAX(paid_at)").
		Where("customer_id = ? AND paid_at IS NOT NULL", customerID).
		Scan(&lastPayment).Error; err == nil && !lastPayment.IsZero() {
		summary.LastPaymentDate = &lastPayment
	}

	summary.CustomerName = r.getUserFullName(ctx, customerID)

	return summary, nil
}

// GetAverageDaysToPay returns avg days to pay invoices
func (r *invoiceRepository) GetAverageDaysToPay(ctx context.Context, tenantID uuid.UUID) (float64, error) {
	var avg sql.NullFloat64
	err := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Select("AVG(EXTRACT(EPOCH FROM (paid_at - issue_date)) / 86400)").
		Where("tenant_id = ? AND status = ? AND paid_at IS NOT NULL", tenantID, models.InvoiceStatusPaid).
		Scan(&avg).Error
	if err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to calculate average days to pay", err)
	}
	if !avg.Valid {
		return 0, nil
	}
	return avg.Float64, nil
}

// GetTopCustomersByRevenue returns top revenue customers
func (r *invoiceRepository) GetTopCustomersByRevenue(ctx context.Context, tenantID uuid.UUID, limit int, startDate, endDate time.Time) ([]CustomerRevenueData, error) {
	if limit <= 0 {
		limit = 5
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	var results []CustomerRevenueData
	err := r.db.WithContext(ctx).
		Table("invoices").
		Select("customer_id, COALESCE(SUM(total_amount), 0) AS revenue, COUNT(*) AS invoice_count").
		Where("tenant_id = ? AND issue_date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Group("customer_id").
		Order("revenue DESC").
		Limit(limit).
		Scan(&results).Error
	if err != nil {
		return nil, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to fetch top customers", err)
	}

	for i := range results {
		results[i].CustomerName = r.getUserFullName(ctx, results[i].CustomerID)
		if results[i].InvoiceCount > 0 {
			results[i].AverageValue = results[i].Revenue / float64(results[i].InvoiceCount)
		}
	}

	return results, nil
}

// GetInvoiceAgingReport builds an aging summary
func (r *invoiceRepository) GetInvoiceAgingReport(ctx context.Context, tenantID uuid.UUID) (InvoiceAgingReport, error) {
	report := InvoiceAgingReport{}

	type agingRow struct {
		Bucket string
		Amount float64
		Count  int64
	}

	var rows []agingRow
	query := `
		SELECT bucket, COALESCE(SUM(outstanding), 0) AS amount, COUNT(*) AS count
		FROM (
			SELECT
				CASE
					WHEN due_date >= NOW() THEN 'current'
					WHEN DATE_PART('day', NOW() - due_date) BETWEEN 0 AND 30 THEN 'current'
					WHEN DATE_PART('day', NOW() - due_date) BETWEEN 31 AND 60 THEN 'days_30_to_60'
					WHEN DATE_PART('day', NOW() - due_date) BETWEEN 61 AND 90 THEN 'days_60_to_90'
					ELSE 'over_90_days'
				END AS bucket,
				total_amount - paid_amount AS outstanding
			FROM invoices
			WHERE tenant_id = ? AND status NOT IN ('paid', 'cancelled')
		) sub
		GROUP BY bucket`

	if err := r.db.WithContext(ctx).Raw(query, tenantID).Scan(&rows).Error; err != nil {
		return report, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to build aging report", err)
	}

	for _, row := range rows {
		switch row.Bucket {
		case "current":
			report.Current.Amount = row.Amount
			report.Current.Count = row.Count
		case "days_30_to_60":
			report.Days30to60.Amount = row.Amount
			report.Days30to60.Count = row.Count
		case "days_60_to_90":
			report.Days60to90.Amount = row.Amount
			report.Days60to90.Count = row.Count
		case "over_90_days":
			report.Over90Days.Amount = row.Amount
			report.Over90Days.Count = row.Count
		}
		report.TotalReceivable += row.Amount
	}

	return report, nil
}

// Search performs a simple search across invoices
func (r *invoiceRepository) Search(ctx context.Context, query string, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error) {
	pagination.Validate()
	like := fmt.Sprintf("%%%s%%", strings.TrimSpace(query))

	countQuery := r.db.WithContext(ctx).Model(&models.Invoice{}).Where("tenant_id = ?", tenantID)
	if query != "" {
		countQuery = countQuery.Where("invoice_number ILIKE ? OR notes ILIKE ?", like, like)
	}

	var totalItems int64
	if err := countQuery.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count invoices", err)
	}

	dataQuery := r.db.WithContext(ctx).Model(&models.Invoice{}).Where("tenant_id = ?", tenantID)
	if query != "" {
		dataQuery = dataQuery.Where("invoice_number ILIKE ? OR notes ILIKE ?", like, like)
	}

	var invoices []*models.Invoice
	if err := dataQuery.
		Preload("Customer").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("issue_date DESC").
		Find(&invoices).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search invoices", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return invoices, paginationResult, nil
}

// FindByFilters applies advanced filtering with pagination
func (r *invoiceRepository) FindByFilters(ctx context.Context, filters InvoiceFilters, pagination PaginationParams) ([]*models.Invoice, PaginationResult, error) {
	if filters.TenantID == uuid.Nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("INVALID_INPUT", "tenant_id is required", errors.ErrInvalidInput)
	}

	pagination.Validate()

	query := r.db.WithContext(ctx).Model(&models.Invoice{})
	query = r.applyInvoiceFilters(query, filters)

	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count invoices", err)
	}

	var invoices []*models.Invoice
	if err := r.applyInvoiceFilters(r.db.WithContext(ctx).Model(&models.Invoice{}), filters).
		Preload("Customer").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("issue_date DESC").
		Find(&invoices).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to apply filters", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return invoices, paginationResult, nil
}

// BulkMarkAsSent updates multiple invoices to sent status
func (r *invoiceRepository) BulkMarkAsSent(ctx context.Context, invoiceIDs []uuid.UUID) error {
	if len(invoiceIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Where("id IN ?", invoiceIDs).
		Update("status", models.InvoiceStatusSent)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk mark as sent", result.Error)
	}

	for _, id := range invoiceIDs {
		r.InvalidateCache(ctx, id)
	}
	return nil
}

// BulkMarkAsCancelled cancels multiple invoices
func (r *invoiceRepository) BulkMarkAsCancelled(ctx context.Context, invoiceIDs []uuid.UUID) error {
	if len(invoiceIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Where("id IN ?", invoiceIDs).
		Update("status", models.InvoiceStatusCancelled)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk cancel invoices", result.Error)
	}

	for _, id := range invoiceIDs {
		r.InvalidateCache(ctx, id)
	}
	return nil
}

// BulkUpdateDueDate updates due dates for multiple invoices
func (r *invoiceRepository) BulkUpdateDueDate(ctx context.Context, invoiceIDs []uuid.UUID, newDueDate time.Time) error {
	if len(invoiceIDs) == 0 {
		return nil
	}
	if newDueDate.IsZero() {
		return errors.NewRepositoryError("INVALID_INPUT", "new due date is required", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Where("id IN ?", invoiceIDs).
		Update("due_date", newDueDate)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk update due dates", result.Error)
	}

	for _, id := range invoiceIDs {
		r.InvalidateCache(ctx, id)
	}
	return nil
}

// Helper Methods
func (r *invoiceRepository) getCacheKey(prefix string, parts ...string) string {
	allParts := append([]string{"repo", "invoices", prefix}, parts...)
	return strings.Join(allParts, ":")
}

func (r *invoiceRepository) getCacheKeyPattern(pattern string) string {
	return fmt.Sprintf("repo:invoices:%s", pattern)
}

func (r *invoiceRepository) invalidateInvoiceCache(ctx context.Context, invoiceID uuid.UUID, invoiceNumber string) {
	if r.cache == nil {
		return
	}

	keys := []string{r.getCacheKey("id", invoiceID.String())}
	if invoiceNumber != "" {
		keys = append(keys, r.getCacheKey("number", invoiceNumber))
	}

	if err := r.cache.Delete(ctx, keys...); err != nil && r.logger != nil {
		r.logger.Warn("failed to delete invoice cache keys", "error", err)
	}

	_ = r.cache.DeletePattern(ctx, r.getCacheKeyPattern("list:*"))
	_ = r.cache.DeletePattern(ctx, r.getCacheKeyPattern("count:*"))
}

func (r *invoiceRepository) applyDateRange(query *gorm.DB, column string, startDate, endDate time.Time) *gorm.DB {
	if !startDate.IsZero() {
		query = query.Where(fmt.Sprintf("%s >= ?", column), startDate)
	}
	if !endDate.IsZero() {
		query = query.Where(fmt.Sprintf("%s <= ?", column), endDate)
	}
	return query
}

func (r *invoiceRepository) applyInvoiceFilters(query *gorm.DB, filters InvoiceFilters) *gorm.DB {
	query = query.Where("tenant_id = ?", filters.TenantID)

	if len(filters.CustomerIDs) > 0 {
		query = query.Where("customer_id IN ?", filters.CustomerIDs)
	}

	if len(filters.Statuses) > 0 {
		query = query.Where("status IN ?", filters.Statuses)
	}

	if filters.MinAmount != nil {
		query = query.Where("total_amount >= ?", *filters.MinAmount)
	}

	if filters.MaxAmount != nil {
		query = query.Where("total_amount <= ?", *filters.MaxAmount)
	}

	if filters.IssueDateFrom != nil {
		query = query.Where("issue_date >= ?", *filters.IssueDateFrom)
	}
	if filters.IssueDateTo != nil {
		query = query.Where("issue_date <= ?", *filters.IssueDateTo)
	}
	if filters.DueDateFrom != nil {
		query = query.Where("due_date >= ?", *filters.DueDateFrom)
	}
	if filters.DueDateTo != nil {
		query = query.Where("due_date <= ?", *filters.DueDateTo)
	}

	if filters.IsOverdue != nil {
		if *filters.IsOverdue {
			query = query.Where("due_date < ? AND status NOT IN ?", time.Now(), []models.InvoiceStatus{
				models.InvoiceStatusPaid,
				models.InvoiceStatusCancelled,
			})
		} else {
			query = query.Where("due_date >= ?", time.Now())
		}
	}

	if filters.HasBooking != nil {
		if *filters.HasBooking {
			query = query.Where("booking_id IS NOT NULL")
		} else {
			query = query.Where("booking_id IS NULL")
		}
	}

	return query
}

func (r *invoiceRepository) sumRevenueInRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	query := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Select("COALESCE(SUM(total_amount), 0)").
		Where("tenant_id = ? AND status = ?", tenantID, models.InvoiceStatusPaid)

	if !startDate.IsZero() {
		query = query.Where("paid_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("paid_at < ?", endDate)
	}

	var total float64
	if err := query.Scan(&total).Error; err != nil {
		return 0, errors.NewRepositoryError("AGGREGATION_FAILED", "failed to sum revenue", err)
	}
	return total, nil
}

func calculateLineItemSubtotal(items []models.InvoiceLineItem) float64 {
	var subtotal float64
	for i := range items {
		itemTotal := items[i].TotalPrice
		if itemTotal == 0 {
			itemTotal = float64(items[i].Quantity) * items[i].UnitPrice
		}
		subtotal += itemTotal
	}
	return subtotal
}

func (r *invoiceRepository) getUserFullName(ctx context.Context, userID uuid.UUID) string {
	type userName struct {
		FirstName string
		LastName  string
	}

	var name userName
	if err := r.db.WithContext(ctx).
		Table("users").
		Select("first_name, last_name").
		Where("id = ?", userID).
		Scan(&name).Error; err != nil {
		return ""
	}

	fullName := strings.TrimSpace(fmt.Sprintf("%s %s", name.FirstName, name.LastName))
	return fullName
}
