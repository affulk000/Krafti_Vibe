package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentRepository defines the interface for payment repository operations
type PaymentRepository interface {
	BaseRepository[models.Payment]

	// Core Operations
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*models.Payment, error)
	GetByCustomerID(ctx context.Context, customerID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error)
	GetByArtisanID(ctx context.Context, artisanID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error)
	GetByProviderPaymentID(ctx context.Context, providerPaymentID string) (*models.Payment, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error)

	// Payment Status Operations
	MarkAsPaid(ctx context.Context, paymentID uuid.UUID, providerPaymentID string) error
	MarkAsFailed(ctx context.Context, paymentID uuid.UUID, reason string) error
	MarkAsCanceled(ctx context.Context, paymentID uuid.UUID) error
	MarkAsProcessing(ctx context.Context, paymentID uuid.UUID) error
	GetPendingPayments(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error)
	GetFailedPayments(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error)
	GetSuccessfulPayments(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Payment, PaginationResult, error)

	// Refund Operations
	CreateRefund(ctx context.Context, paymentID uuid.UUID, amount float64, reason string) error
	PartialRefund(ctx context.Context, paymentID uuid.UUID, amount float64, reason string) error
	GetRefundablePayments(ctx context.Context, bookingID uuid.UUID) ([]*models.Payment, error)
	GetRefundedPayments(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error)
	GetPaymentRefundHistory(ctx context.Context, paymentID uuid.UUID) ([]RefundRecord, error)

	// Commission & Financial Operations
	CalculateCommissionSplit(ctx context.Context, paymentID uuid.UUID, commissionRate float64) error
	GetArtisanEarnings(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time) (float64, error)
	GetPlatformRevenue(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error)
	GetUnpaidArtisanEarnings(ctx context.Context, artisanID uuid.UUID) (float64, error)
	GetArtisanPaymentHistory(ctx context.Context, artisanID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error)
	GetCommissionBreakdown(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (CommissionBreakdown, error)

	// Payment Method Operations
	GetByPaymentMethod(ctx context.Context, method models.PaymentMethod, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error)
	GetPaymentMethodStats(ctx context.Context, tenantID uuid.UUID) (map[models.PaymentMethod]int64, error)
	GetPreferredPaymentMethods(ctx context.Context, customerID uuid.UUID) ([]models.PaymentMethod, error)

	// Analytics & Reporting
	GetPaymentStats(ctx context.Context, tenantID uuid.UUID) (PaymentStats, error)
	GetRevenueByPeriod(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, groupBy string) ([]RevenueData, error)
	GetDailyRevenue(ctx context.Context, tenantID uuid.UUID, date time.Time) (float64, error)
	GetMonthlyRevenue(ctx context.Context, tenantID uuid.UUID, year int, month time.Month) (float64, error)
	GetYearlyRevenue(ctx context.Context, tenantID uuid.UUID, year int) (float64, error)
	GetCustomerPaymentSummary(ctx context.Context, customerID uuid.UUID) (CustomerPaymentSummary, error)
	GetTopPayingCustomers(ctx context.Context, tenantID uuid.UUID, limit int, startDate, endDate time.Time) ([]CustomerPaymentData, error)
	GetPaymentTrends(ctx context.Context, tenantID uuid.UUID, days int) ([]PaymentTrend, error)
	GetAverageTransactionValue(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error)

	// Search & Filter
	Search(ctx context.Context, query string, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error)
	FindByFilters(ctx context.Context, filters PaymentFilters, pagination PaginationParams) ([]*models.Payment, PaginationResult, error)
	GetRecentPayments(ctx context.Context, tenantID uuid.UUID, hours int, pagination PaginationParams) ([]*models.Payment, PaginationResult, error)

	// Provider-specific Operations
	GetByProvider(ctx context.Context, providerName string, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error)
	GetProviderStats(ctx context.Context, tenantID uuid.UUID) (map[string]ProviderStats, error)
	GetFailedPaymentsByProvider(ctx context.Context, providerName string, tenantID uuid.UUID) ([]*models.Payment, error)

	// Reconciliation
	GetUnreconciledPayments(ctx context.Context, tenantID uuid.UUID) ([]*models.Payment, error)
	MarkAsReconciled(ctx context.Context, paymentIDs []uuid.UUID) error
	GetPaymentsForReconciliation(ctx context.Context, tenantID uuid.UUID, date time.Time) ([]*models.Payment, error)

	// Bulk Operations
	BulkMarkAsPaid(ctx context.Context, paymentIDs []uuid.UUID) error
	BulkMarkAsFailed(ctx context.Context, paymentIDs []uuid.UUID, reason string) error
}

// PaymentFilters for advanced filtering
type PaymentFilters struct {
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
}

// PaymentStats represents overall payment statistics
type PaymentStats struct {
	TotalPayments           int64                          `json:"total_payments"`
	TotalRevenue            float64                        `json:"total_revenue"`
	TotalRefunded           float64                        `json:"total_refunded"`
	AverageTransactionValue float64                        `json:"average_transaction_value"`
	SuccessRate             float64                        `json:"success_rate"`
	ByStatus                map[models.PaymentStatus]int64 `json:"by_status"`
	ByMethod                map[models.PaymentMethod]int64 `json:"by_method"`
	PendingCount            int64                          `json:"pending_count"`
	FailedCount             int64                          `json:"failed_count"`
	RefundedCount           int64                          `json:"refunded_count"`
}

// RevenueData represents revenue data for a specific period
type RevenueData struct {
	Period           time.Time `json:"period"`
	Revenue          float64   `json:"revenue"`
	TransactionCount int64     `json:"transaction_count"`
}

// CustomerPaymentSummary represents payment summary for a customer
type CustomerPaymentSummary struct {
	CustomerID         uuid.UUID            `json:"customer_id"`
	TotalSpent         float64              `json:"total_spent"`
	TotalPayments      int64                `json:"total_payments"`
	SuccessfulPayments int64                `json:"successful_payments"`
	FailedPayments     int64                `json:"failed_payments"`
	AveragePayment     float64              `json:"average_payment"`
	LastPaymentDate    *time.Time           `json:"last_payment_date"`
	PreferredMethod    models.PaymentMethod `json:"preferred_method"`
}

// CustomerPaymentData represents customer payment data for analytics
type CustomerPaymentData struct {
	CustomerID   uuid.UUID `json:"customer_id"`
	TotalSpent   float64   `json:"total_spent"`
	PaymentCount int64     `json:"payment_count"`
}

// PaymentTrend represents payment trends over time
type PaymentTrend struct {
	Date             time.Time `json:"date"`
	Revenue          float64   `json:"revenue"`
	TransactionCount int64     `json:"transaction_count"`
	SuccessfulCount  int64     `json:"successful_count"`
	FailedCount      int64     `json:"failed_count"`
}

// CommissionBreakdown represents commission split breakdown
type CommissionBreakdown struct {
	TotalPaid             float64 `json:"total_paid"`
	ArtisanTotal          float64 `json:"artisan_total"`
	PlatformTotal         float64 `json:"platform_total"`
	AverageCommissionRate float64 `json:"average_commission_rate"`
}

// RefundRecord represents a refund record
type RefundRecord struct {
	PaymentID    uuid.UUID  `json:"payment_id"`
	RefundAmount float64    `json:"refund_amount"`
	RefundReason string     `json:"refund_reason"`
	RefundedAt   *time.Time `json:"refunded_at"`
	Status       string     `json:"status"`
}

// ProviderStats represents statistics for a payment provider
type ProviderStats struct {
	ProviderName      string  `json:"provider_name"`
	TotalTransactions int64   `json:"total_transactions"`
	SuccessfulCount   int64   `json:"successful_count"`
	FailedCount       int64   `json:"failed_count"`
	TotalRevenue      float64 `json:"total_revenue"`
	SuccessRate       float64 `json:"success_rate"`
}

// PaymentReconciliation represents a reconciliation record
type PaymentReconciliation struct {
	Date              time.Time `json:"date"`
	TotalPayments     int64     `json:"total_payments"`
	TotalAmount       float64   `json:"total_amount"`
	ReconciledCount   int64     `json:"reconciled_count"`
	UnreconciledCount int64     `json:"unreconciled_count"`
	Discrepancies     int64     `json:"discrepancies"`
}

// ArtisanEarningsReport represents artisan earnings report
type ArtisanEarningsReport struct {
	ArtisanID        uuid.UUID `json:"artisan_id"`
	TotalEarnings    float64   `json:"total_earnings"`
	PaidAmount       float64   `json:"paid_amount"`
	UnpaidAmount     float64   `json:"unpaid_amount"`
	TransactionCount int64     `json:"transaction_count"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
}

// PaymentMethodPreference represents payment method preferences
type PaymentMethodPreference struct {
	Method     models.PaymentMethod `json:"method"`
	Count      int64                `json:"count"`
	Percentage float64              `json:"percentage"`
}

// paymentRepository implements PaymentRepository
// paymentRepository implements PaymentRepository
type paymentRepository struct {
	BaseRepository[models.Payment]
	db      *gorm.DB
	logger  log.AllLogger
	cache   Cache
	metrics MetricsCollector
}

// NewPaymentRepository creates a new PaymentRepository instance
func NewPaymentRepository(db *gorm.DB, config ...RepositoryConfig) PaymentRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 3 * time.Minute // Shorter cache for financial data
	}

	baseRepo := NewBaseRepository[models.Payment](db, cfg)

	return &paymentRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
		metrics:        cfg.Metrics,
	}
}

// GetByBookingID retrieves all payments for a booking
func (r *paymentRepository) GetByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*models.Payment, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("get_by_booking_id", "payments", time.Since(start), nil)
		}
	}()

	if bookingID == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "booking ID cannot be nil", errors.ErrInvalidInput)
	}

	// try cache
	cacheKey := r.getCacheKey("booking", bookingID.String())
	if r.cache != nil {
		var payments []*models.Payment
		if err := r.cache.GetJSON(ctx, cacheKey, &payments); err == nil {
			if r.metrics != nil {
				r.metrics.RecordCacheHit("payments")
			}
			return payments, nil
		}
		if r.metrics != nil {
			r.metrics.RecordCacheMiss("payments")
		}
	}

	var payments []*models.Payment
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Where("booking_id = ?", bookingID).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		r.logger.Error("failed to get payments by booking ID", "booking_id", bookingID, "error", err)
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get payments", err)
	}

	// Cache the result
	if r.cache != nil {
		if err := r.cache.SetJSON(ctx, cacheKey, payments, 3*time.Minute); err != nil {
			r.logger.Warn("failed to cache payments", "booking_id", bookingID, "error", err)
		}
	}

	return payments, nil

}

// GetByCustomerID retrieves all payments for a customer
func (r *paymentRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Payment{}).Where("customer_id = ?", customerID)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count payments", err)
	}

	var payments []*models.Payment
	if err := query.
		Preload("Booking").
		Preload("Artisan").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find payments", err)
	}

	PaginationResult := CalculatePagination(pagination, totalItems)
	return payments, PaginationResult, nil
}

// GetByArtisanID retrieves all payments for an artisan
func (r *paymentRepository) GetByArtisanID(ctx context.Context, artisanID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("artisan_id = ?", artisanID)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count payments", err)
	}

	var payments []*models.Payment
	if err := query.
		Preload("Customer").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find payments", err)
	}

	PaginationResult := CalculatePagination(pagination, totalItems)
	return payments, PaginationResult, nil
}

// GetByProviderPaymentID retrieves a payment by provider payment ID
func (r *paymentRepository) GetByProviderPaymentID(ctx context.Context, providerPaymentID string) (*models.Payment, error) {
	if providerPaymentID == "" {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "provider payment ID cannot be empty", errors.ErrInvalidInput)
	}

	// Try cache first
	cacheKey := r.getCacheKey("provider", providerPaymentID)
	if r.cache != nil {
		var payment models.Payment
		if err := r.cache.GetJSON(ctx, cacheKey, &payment); err == nil {
			if r.metrics != nil {
				r.metrics.RecordCacheHit("payments")
			}
			return &payment, nil
		}
		if r.metrics != nil {
			r.metrics.RecordCacheMiss("payments")
		}
	}

	var payment models.Payment
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Where("provider_payment_id = ?", providerPaymentID).
		First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "payment not found", errors.ErrNotFound)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get payment", err)
	}

	// Cache the result
	if r.cache != nil {
		if err := r.cache.SetJSON(ctx, cacheKey, payment, 3*time.Minute); err != nil {
			r.logger.Warn("failed to cache payment", "provider_payment_id", providerPaymentID, "error", err)
		}
	}

	return &payment, nil
}

// GetByTenantID retrieves all payments for a tenant
func (r *paymentRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Payment{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count payments", err)
	}

	var payments []*models.Payment
	if err := query.
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find payments", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return payments, paginationResult, nil
}

// MarkAsPaid retrievies marked as paid payments
func (r *paymentRepository) MarkAsPaid(ctx context.Context, paymentID uuid.UUID, providerPaymentID string) error {
	payment, err := r.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}
	var payments models.Payment
	payments.MarkAsPaid()
	if providerPaymentID != "" {
		payments.ProviderPaymentID = providerPaymentID
	}

	if err := r.Update(ctx, payment); err != nil {
		return err
	}

	// Invalidate cache
	r.invalidatePaymentCache(ctx, paymentID, payments.BookingID)

	r.logger.Info("payment marked as paid", "payment_id", paymentID, "provider_payment_id", providerPaymentID)
	return nil
}
func (r *paymentRepository) MarkAsFailed(ctx context.Context, paymentID uuid.UUID, reason string) error {
	payment, err := r.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	var payments models.Payment
	payments.MarkAsFailed(reason)

	if err := r.Update(ctx, payment); err != nil {
		return err
	}

	// Invalidate cache
	r.invalidatePaymentCache(ctx, paymentID, payments.BookingID)

	r.logger.Info("payment marked as failed", "payment_id", paymentID, "reason", reason)
	return nil
}
func (r *paymentRepository) MarkAsCanceled(ctx context.Context, paymentID uuid.UUID) error {
	payment, err := r.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	var payments models.Payment
	payments.MarkAsCancelled()

	if err := r.Update(ctx, payment); err != nil {
		return err
	}

	// Invalidate cache
	r.invalidatePaymentCache(ctx, paymentID, payments.BookingID)

	r.logger.Info("payment marked as canceled", "payment_id", paymentID)
	return nil
}
func (r *paymentRepository) MarkAsProcessing(ctx context.Context, paymentID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("id = ?", paymentID).
		Update("status", models.PaymentStatusProcessing)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark as processing", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "payment not found", errors.ErrNotFound)
	}

	return nil
}
func (r *paymentRepository) GetPendingPayments(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.PaymentStatusPending)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count payments", err)
	}

	var payments []*models.Payment
	if err := query.
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at ASC").
		Find(&payments).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find payments", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return payments, paginationResult, nil
}
func (r *paymentRepository) GetFailedPayments(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("tenant_id = ? AND status = ?", tenantID, models.PaymentStatusFailed)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count payments", err)
	}

	var payments []*models.Payment
	if err := query.
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find payments", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return payments, paginationResult, nil
}
func (r *paymentRepository) GetSuccessfulPayments(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination PaginationParams) ([]*models.Payment, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("tenant_id = ? AND status = ? AND processed_at BETWEEN ? AND ?",
			tenantID, models.PaymentStatusPaid, startDate, endDate)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count payments", err)
	}

	var payments []*models.Payment
	if err := query.
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("processed_at DESC").
		Find(&payments).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find payments", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return payments, paginationResult, nil
}

func (r *paymentRepository) CreateRefund(ctx context.Context, paymentID uuid.UUID, amount float64, reason string) error {
	payment, err := r.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	var payments models.Payment
	if err := payments.ProcessRefund(amount, reason); err != nil {
		return errors.NewRepositoryError("REFUND_FAILED", err.Error(), errors.ErrInvalidInput)
	}

	if err := r.Update(ctx, payment); err != nil {
		return err
	}

	// Invalidate cache
	r.invalidatePaymentCache(ctx, paymentID, payments.BookingID)

	r.logger.Info("refund created", "payment_id", paymentID, "amount", amount, "reason", reason)
	return nil
}
func (r *paymentRepository) PartialRefund(ctx context.Context, paymentID uuid.UUID, amount float64, reason string) error {
	return r.CreateRefund(ctx, paymentID, amount, reason)
}
func (r *paymentRepository) GetRefundablePayments(ctx context.Context, bookingID uuid.UUID) ([]*models.Payment, error) {
	var payments []*models.Payment
	if err := r.db.WithContext(ctx).
		Where("booking_id = ? AND status = ? AND refunded_amount < amount",
			bookingID, models.PaymentStatusPaid).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find refundable payments", err)
	}
	return payments, nil
}
func (r *paymentRepository) GetRefundedPayments(ctx context.Context, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("tenant_id = ? AND status IN ?", tenantID, []models.PaymentStatus{
			models.PaymentStatusRefunded,
			models.PaymentStatusPartialRefund,
		})

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count payments", err)
	}

	var payments []*models.Payment
	if err := query.
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("refunded_at DESC").
		Find(&payments).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find payments", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return payments, paginationResult, nil
}

// GetPaymentRefundHistory retrieves refund history for a payment
func (r *paymentRepository) GetPaymentRefundHistory(ctx context.Context, paymentID uuid.UUID) ([]RefundRecord, error) {
	_, err := r.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	var payments models.Payment
	// Check if payment has any refunds
	if payments.RefundedAmount == 0 {
		return []RefundRecord{}, nil
	}

	refunedRecord := []RefundRecord{
		{
			PaymentID:    payments.ID,
			RefundAmount: payments.RefundedAmount,
			RefundReason: payments.RefundReason,
			RefundedAt:   payments.RefundedAt,
			Status:       string(payments.Status),
		},
	}

	return refunedRecord, nil
}

func (r *paymentRepository) CalculateCommissionSplit(ctx context.Context, paymentID uuid.UUID, commissionRate float64) error {
	payment, err := r.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	var payments models.Payment
	if err := payments.SetCommissionRate(commissionRate); err != nil {
		return errors.NewRepositoryError("COMMISSION_FAILED", err.Error(), errors.ErrInvalidInput)
	}

	if err := r.Update(ctx, payment); err != nil {
		return err
	}

	// Invalidate cache
	r.invalidatePaymentCache(ctx, paymentID, payments.BookingID)

	r.logger.Info("commission split calculated", "payment_id", paymentID, "rate", commissionRate)
	return nil
}
func (r *paymentRepository) GetArtisanEarnings(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	var totalEarnings float64
	if err := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(SUM(artisan_amount), 0)").
		Where("artisan_id = ? AND status = ? AND processed_at BETWEEN ? AND ?",
			artisanID, models.PaymentStatusPaid, startDate, endDate).
		Scan(&totalEarnings).Error; err != nil {
		return 0, errors.NewRepositoryError("CALCULATION_FAILED", "failed to calculate earnings", err)
	}
	return totalEarnings, nil
}
func (r *paymentRepository) GetPlatformRevenue(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	var totalRevenue float64
	if err := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(SUM(platform_amount), 0)").
		Where("tenant_id = ? AND status = ? AND processed_at BETWEEN ? AND ?",
			tenantID, models.PaymentStatusPaid, startDate, endDate).
		Scan(&totalRevenue).Error; err != nil {
		return 0, errors.NewRepositoryError("CALCULATION_FAILED", "failed to calculate revenue", err)
	}
	return totalRevenue, nil
}
func (r *paymentRepository) GetUnpaidArtisanEarnings(ctx context.Context, artisanID uuid.UUID) (float64, error) {
	var totalEarnings float64
	if err := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(SUM(artisan_amount), 0)").
		Where("artisan_id = ? AND status = ? AND (metadata->>'paid_to_artisan')::boolean IS NOT TRUE",
			artisanID, models.PaymentStatusPaid).
		Scan(&totalEarnings).Error; err != nil {
		return 0, errors.NewRepositoryError("CALCULATION_FAILED", "failed to calculate unpaid earnings", err)
	}
	return totalEarnings, nil
}
func (r *paymentRepository) GetArtisanPaymentHistory(ctx context.Context, artisanID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error) {
	return r.GetByArtisanID(ctx, artisanID, pagination)
}
func (r *paymentRepository) GetCommissionBreakdown(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (CommissionBreakdown, error) {
	var breakdown CommissionBreakdown

	// Total paid
	r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("tenant_id = ? AND status = ? AND processed_at BETWEEN ? AND ?",
			tenantID, models.PaymentStatusPaid, startDate, endDate).
		Scan(&breakdown.TotalPaid)

	// Artisan total
	r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(SUM(artisan_amount), 0)").
		Where("tenant_id = ? AND status = ? AND processed_at BETWEEN ? AND ?",
			tenantID, models.PaymentStatusPaid, startDate, endDate).
		Scan(&breakdown.ArtisanTotal)

	// Platform total
	r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(SUM(platform_amount), 0)").
		Where("tenant_id = ? AND status = ? AND processed_at BETWEEN ? AND ?",
			tenantID, models.PaymentStatusPaid, startDate, endDate).
		Scan(&breakdown.PlatformTotal)

	// Average commission rate
	r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(AVG(commission_rate), 0)").
		Where("tenant_id = ? AND status = ? AND processed_at BETWEEN ? AND ?",
			tenantID, models.PaymentStatusPaid, startDate, endDate).
		Scan(&breakdown.AverageCommissionRate)

	return breakdown, nil
}

func (r *paymentRepository) GetByPaymentMethod(ctx context.Context, method models.PaymentMethod, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("tenant_id = ? AND method = ?", tenantID, method)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count payments", err)
	}

	var payments []*models.Payment
	if err := query.
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find payments", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return payments, paginationResult, nil
}
func (r *paymentRepository) GetPaymentMethodStats(ctx context.Context, tenantID uuid.UUID) (map[models.PaymentMethod]int64, error) {
	var results []struct {
		Method models.PaymentMethod
		Count  int64
	}

	if err := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("method, COUNT(*) as count").
		Where("tenant_id = ? AND status = ?", tenantID, models.PaymentStatusPaid).
		Group("method").
		Scan(&results).Error; err != nil {
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get payment method stats", err)
	}

	stats := make(map[models.PaymentMethod]int64)
	for _, result := range results {
		stats[result.Method] = result.Count
	}

	return stats, nil
}
func (r *paymentRepository) GetPreferredPaymentMethods(ctx context.Context, customerID uuid.UUID) ([]models.PaymentMethod, error) {
	var results []struct {
		Method models.PaymentMethod
		Count  int64
	}

	if err := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("method, COUNT(*) as count").
		Where("customer_id = ? AND status = ?", customerID, models.PaymentStatusPaid).
		Group("method").
		Order("count DESC").
		Limit(3).
		Scan(&results).Error; err != nil {
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get preferred methods", err)
	}

	methods := make([]models.PaymentMethod, len(results))
	for i, result := range results {
		methods[i] = result.Method
	}

	return methods, nil
}

func (r *paymentRepository) GetPaymentStats(ctx context.Context, tenantID uuid.UUID) (PaymentStats, error) {
	var stats PaymentStats

	// Total payments
	r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalPayments)

	// By status
	var statusResults []struct {
		Status models.PaymentStatus
		Count  int64
	}
	r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("status").
		Scan(&statusResults)

	stats.ByStatus = make(map[models.PaymentStatus]int64)
	for _, result := range statusResults {
		stats.ByStatus[result.Status] = result.Count
	}

	// Total revenue (paid payments)
	r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("tenant_id = ? AND status = ?", tenantID, models.PaymentStatusPaid).
		Scan(&stats.TotalRevenue)

	// Total refunded
	r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(SUM(refunded_amount), 0)").
		Where("tenant_id = ?", tenantID).
		Scan(&stats.TotalRefunded)

	// Average transaction value
	r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(AVG(amount), 0)").
		Where("tenant_id = ? AND status = ?", tenantID, models.PaymentStatusPaid).
		Scan(&stats.AverageTransactionValue)

	// By payment method
	var methodResults []struct {
		Method models.PaymentMethod
		Count  int64
	}
	r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("method, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("method").
		Scan(&methodResults)

	stats.ByMethod = make(map[models.PaymentMethod]int64)
	for _, result := range methodResults {
		stats.ByMethod[result.Method] = result.Count
	}

	// Success rate
	if stats.TotalPayments > 0 {
		successCount := stats.ByStatus[models.PaymentStatusPaid]
		stats.SuccessRate = (float64(successCount) / float64(stats.TotalPayments)) * 100
	}

	return stats, nil
}
func (r *paymentRepository) GetRevenueByPeriod(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, groupBy string) ([]RevenueData, error) {
	var results []RevenueData

	var dateFormat string
	switch groupBy {
	case "day":
		dateFormat = "DATE(processed_at)"
	case "week":
		dateFormat = "DATE_TRUNC('week', processed_at)"
	case "month":
		dateFormat = "DATE_TRUNC('month', processed_at)"
	default:
		dateFormat = "DATE(processed_at)"
	}

	query := fmt.Sprintf(`
			SELECT
				%s as period,
				COALESCE(SUM(amount), 0) as revenue,
				COUNT(*) as transaction_count
			FROM payments
			WHERE tenant_id = ? AND status = ? AND processed_at BETWEEN ? AND ?
			GROUP BY period
			ORDER BY period ASC
		`, dateFormat)

	if err := r.db.WithContext(ctx).Raw(query, tenantID, models.PaymentStatusPaid, startDate, endDate).Scan(&results).Error; err != nil {
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get revenue by period", err)
	}

	return results, nil
}
func (r *paymentRepository) GetDailyRevenue(ctx context.Context, tenantID uuid.UUID, date time.Time) (float64, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var revenue float64
	if err := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("tenant_id = ? AND status = ? AND processed_at BETWEEN ? AND ?",
			tenantID, models.PaymentStatusPaid, startOfDay, endOfDay).
		Scan(&revenue).Error; err != nil {
		return 0, errors.NewRepositoryError("CALCULATION_FAILED", "failed to calculate daily revenue", err)
	}

	return revenue, nil
}

// GetMonthlyRevenue retrieves revenue for a specific month
func (r *paymentRepository) GetMonthlyRevenue(ctx context.Context, tenantID uuid.UUID, year int, month time.Month) (float64, error) {
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	var revenue float64
	if err := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("tenant_id = ? AND status = ? AND processed_at BETWEEN ? AND ?",
			tenantID, models.PaymentStatusPaid, startOfMonth, endOfMonth).
		Scan(&revenue).Error; err != nil {
		return 0, errors.NewRepositoryError("CALCULATION_FAILED", "failed to calculate monthly revenue", err)
	}

	return revenue, nil
}

// GetYearlyRevenue retrieves revenue for a specific year
func (r *paymentRepository) GetYearlyRevenue(ctx context.Context, tenantID uuid.UUID, year int) (float64, error) {
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := startOfYear.AddDate(1, 0, 0)

	var revenue float64
	if err := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("tenant_id = ? AND status = ? AND processed_at BETWEEN ? AND ?",
			tenantID, models.PaymentStatusPaid, startOfYear, endOfYear).
		Scan(&revenue).Error; err != nil {
		return 0, errors.NewRepositoryError("CALCULATION_FAILED", "failed to calculate yearly revenue", err)
	}

	return revenue, nil
}

// GetCustomerPaymentSummary retrieves payment summary for a customer
func (r *paymentRepository) GetCustomerPaymentSummary(ctx context.Context, customerID uuid.UUID) (CustomerPaymentSummary, error) {
	var summary CustomerPaymentSummary
	summary.CustomerID = customerID

	// Total spent
	r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("customer_id = ? AND status = ?", customerID, models.PaymentStatusPaid).
		Scan(&summary.TotalSpent)

	// Total payments
	r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("customer_id = ?", customerID).
		Count(&summary.TotalPayments)

	// Successful payments
	r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("customer_id = ? AND status = ?", customerID, models.PaymentStatusPaid).
		Count(&summary.SuccessfulPayments)

	// Failed payments
	r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("customer_id = ? AND status = ?", customerID, models.PaymentStatusFailed).
		Count(&summary.FailedPayments)

	// Average payment
	if summary.SuccessfulPayments > 0 {
		summary.AveragePayment = summary.TotalSpent / float64(summary.SuccessfulPayments)
	}

	// Last payment date
	var lastPayment models.Payment
	if err := r.db.WithContext(ctx).
		Where("customer_id = ? AND status = ?", customerID, models.PaymentStatusPaid).
		Order("processed_at DESC").
		First(&lastPayment).Error; err == nil {
		summary.LastPaymentDate = lastPayment.ProcessedAt
	}

	// Preferred payment method
	var methodResult struct {
		Method models.PaymentMethod
	}
	if err := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("method").
		Where("customer_id = ? AND status = ?", customerID, models.PaymentStatusPaid).
		Group("method").
		Order("COUNT(*) DESC").
		Limit(1).
		Scan(&methodResult).Error; err == nil {
		summary.PreferredMethod = methodResult.Method
	}

	return summary, nil
}

// GetTopPayingCustomers retrieves top paying customers
func (r *paymentRepository) GetTopPayingCustomers(ctx context.Context, tenantID uuid.UUID, limit int, startDate, endDate time.Time) ([]CustomerPaymentData, error) {
	var results []CustomerPaymentData

	query := `
		SELECT
			customer_id,
			COALESCE(SUM(amount), 0) as total_spent,
			COUNT(*) as payment_count
		FROM payments
		WHERE tenant_id = ? AND status = ? AND processed_at BETWEEN ? AND ?
		GROUP BY customer_id
		ORDER BY total_spent DESC
		LIMIT ?
	`

	if err := r.db.WithContext(ctx).Raw(query, tenantID, models.PaymentStatusPaid, startDate, endDate, limit).Scan(&results).Error; err != nil {
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get top paying customers", err)
	}

	return results, nil
}

// GetPaymentTrends retrieves payment trends for the specified number of days
func (r *paymentRepository) GetPaymentTrends(ctx context.Context, tenantID uuid.UUID, days int) ([]PaymentTrend, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	var results []PaymentTrend

	query := `
		SELECT
			DATE(processed_at) as date,
			COALESCE(SUM(amount), 0) as revenue,
			COUNT(*) as transaction_count,
			COUNT(CASE WHEN status = ? THEN 1 END) as successful_count,
			COUNT(CASE WHEN status = ? THEN 1 END) as failed_count
		FROM payments
		WHERE tenant_id = ? AND processed_at >= ?
		GROUP BY DATE(processed_at)
		ORDER BY date ASC
	`

	if err := r.db.WithContext(ctx).Raw(query, models.PaymentStatusPaid, models.PaymentStatusFailed, tenantID, startDate).Scan(&results).Error; err != nil {
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get payment trends", err)
	}

	return results, nil
}

// GetAverageTransactionValue retrieves average transaction value
func (r *paymentRepository) GetAverageTransactionValue(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	var avgValue float64
	if err := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("COALESCE(AVG(amount), 0)").
		Where("tenant_id = ? AND status = ? AND processed_at BETWEEN ? AND ?",
			tenantID, models.PaymentStatusPaid, startDate, endDate).
		Scan(&avgValue).Error; err != nil {
		return 0, errors.NewRepositoryError("CALCULATION_FAILED", "failed to calculate average transaction value", err)
	}

	return avgValue, nil
}

// Search searches payments by query string
func (r *paymentRepository) Search(ctx context.Context, query string, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error) {
	pagination.Validate()
	searchPattern := "%" + strings.ToLower(query) + "%"

	dbQuery := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Joins("LEFT JOIN users AS customers ON customers.id = payments.customer_id").
		Where("payments.tenant_id = ?", tenantID).
		Where("LOWER(customers.first_name) LIKE ? OR LOWER(customers.last_name) LIKE ? OR LOWER(payments.provider_payment_id) LIKE ?",
			searchPattern, searchPattern, searchPattern)

	var totalItems int64
	if err := dbQuery.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count payments", err)
	}

	var payments []*models.Payment
	if err := dbQuery.
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("payments.created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search payments", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return payments, paginationResult, nil
}

// FindByFilters retrieves payments using advanced filters
func (r *paymentRepository) FindByFilters(ctx context.Context, filters PaymentFilters, pagination PaginationParams) ([]*models.Payment, PaginationResult, error) {
	pagination.Validate()

	query := r.db.WithContext(ctx).Model(&models.Payment{})

	// Apply filters
	if filters.TenantID != uuid.Nil {
		query = query.Where("tenant_id = ?", filters.TenantID)
	}

	if len(filters.Statuses) > 0 {
		query = query.Where("status IN ?", filters.Statuses)
	}

	if len(filters.Methods) > 0 {
		query = query.Where("method IN ?", filters.Methods)
	}

	if len(filters.Types) > 0 {
		query = query.Where("type IN ?", filters.Types)
	}

	if filters.MinAmount != nil {
		query = query.Where("amount >= ?", *filters.MinAmount)
	}

	if filters.MaxAmount != nil {
		query = query.Where("amount <= ?", *filters.MaxAmount)
	}

	if filters.CustomerID != nil {
		query = query.Where("customer_id = ?", *filters.CustomerID)
	}

	if filters.ArtisanID != nil {
		query = query.Where("artisan_id = ?", *filters.ArtisanID)
	}

	if filters.BookingID != nil {
		query = query.Where("booking_id = ?", *filters.BookingID)
	}

	if filters.ProcessedAfter != nil {
		query = query.Where("processed_at >= ?", *filters.ProcessedAfter)
	}

	if filters.ProcessedBefore != nil {
		query = query.Where("processed_at <= ?", *filters.ProcessedBefore)
	}

	if filters.HasRefund != nil && *filters.HasRefund {
		query = query.Where("refunded_amount > 0")
	}

	if filters.ProviderName != nil {
		query = query.Where("provider_name = ?", *filters.ProviderName)
	}

	// Count total
	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count payments", err)
	}

	// Find payments
	var payments []*models.Payment
	if err := query.
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find payments", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return payments, paginationResult, nil
}

// GetRecentPayments retrieves recent payments within specified hours
func (r *paymentRepository) GetRecentPayments(ctx context.Context, tenantID uuid.UUID, hours int, pagination PaginationParams) ([]*models.Payment, PaginationResult, error) {
	pagination.Validate()

	cutoffTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, cutoffTime)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count payments", err)
	}

	var payments []*models.Payment
	if err := query.
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find payments", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return payments, paginationResult, nil
}

// GetByProvider retrieves payments by provider
func (r *paymentRepository) GetByProvider(ctx context.Context, providerName string, tenantID uuid.UUID, pagination PaginationParams) ([]*models.Payment, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	query := r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("tenant_id = ? AND provider_name = ?", tenantID, providerName)

	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count payments", err)
	}

	var payments []*models.Payment
	if err := query.
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find payments", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return payments, paginationResult, nil
}

// GetProviderStats retrieves statistics by provider
func (r *paymentRepository) GetProviderStats(ctx context.Context, tenantID uuid.UUID) (map[string]ProviderStats, error) {
	var results []struct {
		ProviderName      string
		TotalTransactions int64
		SuccessfulCount   int64
		FailedCount       int64
		TotalRevenue      float64
	}

	query := `
		SELECT
			provider_name,
			COUNT(*) as total_transactions,
			COUNT(CASE WHEN status = ? THEN 1 END) as successful_count,
			COUNT(CASE WHEN status = ? THEN 1 END) as failed_count,
			COALESCE(SUM(CASE WHEN status = ? THEN amount ELSE 0 END), 0) as total_revenue
		FROM payments
		WHERE tenant_id = ? AND provider_name IS NOT NULL AND provider_name != ''
		GROUP BY provider_name
	`

	if err := r.db.WithContext(ctx).Raw(query,
		models.PaymentStatusPaid, models.PaymentStatusFailed, models.PaymentStatusPaid, tenantID).
		Scan(&results).Error; err != nil {
		return nil, errors.NewRepositoryError("QUERY_FAILED", "failed to get provider stats", err)
	}

	stats := make(map[string]ProviderStats)
	for _, result := range results {
		providerStat := ProviderStats{
			ProviderName:      result.ProviderName,
			TotalTransactions: result.TotalTransactions,
			SuccessfulCount:   result.SuccessfulCount,
			FailedCount:       result.FailedCount,
			TotalRevenue:      result.TotalRevenue,
		}
		if result.TotalTransactions > 0 {
			providerStat.SuccessRate = (float64(result.SuccessfulCount) / float64(result.TotalTransactions)) * 100
		}
		stats[result.ProviderName] = providerStat
	}

	return stats, nil
}

// GetFailedPaymentsByProvider retrieves failed payments for a provider
func (r *paymentRepository) GetFailedPaymentsByProvider(ctx context.Context, providerName string, tenantID uuid.UUID) ([]*models.Payment, error) {
	var payments []*models.Payment
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Where("tenant_id = ? AND provider_name = ? AND status = ?",
			tenantID, providerName, models.PaymentStatusFailed).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find failed payments", err)
	}
	return payments, nil
}

// GetUnreconciledPayments retrieves payments that need reconciliation
func (r *paymentRepository) GetUnreconciledPayments(ctx context.Context, tenantID uuid.UUID) ([]*models.Payment, error) {
	var payments []*models.Payment
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Where("tenant_id = ? AND status = ? AND (metadata->>'reconciled')::boolean IS NOT TRUE",
			tenantID, models.PaymentStatusPaid).
		Order("processed_at ASC").
		Find(&payments).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find unreconciled payments", err)
	}
	return payments, nil
}

// MarkAsReconciled marks payments as reconciled
func (r *paymentRepository) MarkAsReconciled(ctx context.Context, paymentIDs []uuid.UUID) error {
	if len(paymentIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("id IN ?", paymentIDs).
		Update("metadata", gorm.Expr("jsonb_set(COALESCE(metadata, '{}'::jsonb), '{reconciled}', 'true'::jsonb)"))

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to mark as reconciled", result.Error)
	}

	r.logger.Info("payments marked as reconciled", "count", result.RowsAffected)
	return nil
}

// GetPaymentsForReconciliation retrieves payments for a specific date
func (r *paymentRepository) GetPaymentsForReconciliation(ctx context.Context, tenantID uuid.UUID, date time.Time) ([]*models.Payment, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var payments []*models.Payment
	if err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Artisan").
		Preload("Booking").
		Where("tenant_id = ? AND status = ? AND processed_at BETWEEN ? AND ?",
			tenantID, models.PaymentStatusPaid, startOfDay, endOfDay).
		Order("processed_at ASC").
		Find(&payments).Error; err != nil {
		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find payments for reconciliation", err)
	}
	return payments, nil
}

// BulkMarkAsPaid marks multiple payments as paid
func (r *paymentRepository) BulkMarkAsPaid(ctx context.Context, paymentIDs []uuid.UUID) error {
	if len(paymentIDs) == 0 {
		return nil
	}

	now := time.Now()
	updates := map[string]any{
		"status":       models.PaymentStatusPaid,
		"processed_at": now,
	}

	result := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("id IN ?", paymentIDs).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk mark as paid", result.Error)
	}

	r.logger.Info("payments marked as paid in bulk", "count", result.RowsAffected)
	return nil
}

// BulkMarkAsFailed marks multiple payments as failed
func (r *paymentRepository) BulkMarkAsFailed(ctx context.Context, paymentIDs []uuid.UUID, reason string) error {
	if len(paymentIDs) == 0 {
		return nil
	}

	updates := map[string]any{
		"status":         models.PaymentStatusFailed,
		"failure_reason": reason,
	}

	result := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("id IN ?", paymentIDs).
		Updates(updates)

	if result.Error != nil {
		return errors.NewRepositoryError("UPDATE_FAILED", "failed to bulk mark as failed", result.Error)
	}

	r.logger.Info("payments marked as failed in bulk", "count", result.RowsAffected, "reason", reason)
	return nil
}

// Helper methods

func (r *paymentRepository) getCacheKey(prefix string, parts ...string) string {
	allParts := append([]string{"repo", "payments", prefix}, parts...)
	return strings.Join(allParts, ":")
}

func (r *paymentRepository) invalidatePaymentCache(ctx context.Context, paymentID, bookingID uuid.UUID) {
	if r.cache == nil {
		return
	}

	patterns := []string{
		r.getCacheKey("id", paymentID.String()),
		r.getCacheKey("booking", bookingID.String()),
		r.getCacheKey("provider", "*"),
	}

	for _, pattern := range patterns {
		r.cache.DeletePattern(ctx, pattern)
	}
}
