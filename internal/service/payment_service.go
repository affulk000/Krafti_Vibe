package service

import (
	"context"
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// PaymentService defines the interface for payment service operations
type PaymentService interface {
	// Core Payment Operations
	CreatePayment(ctx context.Context, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error)
	GetPayment(ctx context.Context, paymentID uuid.UUID) (*dto.PaymentResponse, error)
	GetPaymentsByBooking(ctx context.Context, bookingID uuid.UUID) ([]*dto.PaymentResponse, error)
	GetPaymentsByCustomer(ctx context.Context, customerID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error)
	GetPaymentsByArtisan(ctx context.Context, artisanID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error)
	GetPaymentsByTenant(ctx context.Context, tenantID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error)
	GetPaymentByProviderID(ctx context.Context, providerPaymentID string) (*dto.PaymentResponse, error)

	// Payment Status Operations
	MarkPaymentAsPaid(ctx context.Context, paymentID uuid.UUID, providerPaymentID string) (*dto.PaymentResponse, error)
	MarkPaymentAsFailed(ctx context.Context, paymentID uuid.UUID, reason string) (*dto.PaymentResponse, error)
	MarkPaymentAsCancelled(ctx context.Context, paymentID uuid.UUID) (*dto.PaymentResponse, error)
	MarkPaymentAsProcessing(ctx context.Context, paymentID uuid.UUID) (*dto.PaymentResponse, error)
	GetPendingPayments(ctx context.Context, tenantID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error)
	GetFailedPayments(ctx context.Context, tenantID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error)
	GetSuccessfulPayments(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination repository.PaginationParams) (*dto.PaymentListResponse, error)

	// Refund Operations
	ProcessRefund(ctx context.Context, paymentID uuid.UUID, amount float64, reason string) (*dto.PaymentResponse, error)
	ProcessPartialRefund(ctx context.Context, paymentID uuid.UUID, amount float64, reason string) (*dto.PaymentResponse, error)
	GetRefundablePayments(ctx context.Context, bookingID uuid.UUID) ([]*dto.PaymentResponse, error)
	GetRefundedPayments(ctx context.Context, tenantID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error)
	GetPaymentRefundHistory(ctx context.Context, paymentID uuid.UUID) ([]*dto.RefundRecordResponse, error)

	// Commission & Earnings
	CalculateCommission(ctx context.Context, paymentID uuid.UUID, commissionRate float64) (*dto.PaymentResponse, error)
	GetArtisanEarnings(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time) (*dto.EarningsResponse, error)
	GetPlatformRevenue(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*dto.RevenueResponse, error)
	GetUnpaidArtisanEarnings(ctx context.Context, artisanID uuid.UUID) (*dto.EarningsResponse, error)
	GetArtisanPaymentHistory(ctx context.Context, artisanID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error)
	GetCommissionBreakdown(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*dto.CommissionBreakdownResponse, error)

	// Payment Method Operations
	GetPaymentsByMethod(ctx context.Context, method models.PaymentMethod, tenantID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error)
	GetPaymentMethodStats(ctx context.Context, tenantID uuid.UUID) (map[models.PaymentMethod]int64, error)
	GetPreferredPaymentMethods(ctx context.Context, customerID uuid.UUID) ([]models.PaymentMethod, error)

	// Analytics & Reporting
	GetPaymentStats(ctx context.Context, tenantID uuid.UUID) (*dto.PaymentStatsResponse, error)
	GetRevenueByPeriod(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, groupBy string) ([]*dto.RevenueDataResponse, error)
	GetDailyRevenue(ctx context.Context, tenantID uuid.UUID, date time.Time) (float64, error)
	GetMonthlyRevenue(ctx context.Context, tenantID uuid.UUID, year int, month time.Month) (float64, error)
	GetYearlyRevenue(ctx context.Context, tenantID uuid.UUID, year int) (float64, error)
	GetCustomerPaymentSummary(ctx context.Context, customerID uuid.UUID) (*dto.CustomerPaymentSummaryResponse, error)
	GetTopPayingCustomers(ctx context.Context, tenantID uuid.UUID, limit int, startDate, endDate time.Time) ([]*dto.CustomerPaymentDataResponse, error)
	GetPaymentTrends(ctx context.Context, tenantID uuid.UUID, days int) ([]*dto.PaymentTrendResponse, error)
	GetAverageTransactionValue(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error)

	// Search & Filter
	SearchPayments(ctx context.Context, query string, tenantID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error)
	FilterPayments(ctx context.Context, filters dto.PaymentFilter, pagination repository.PaginationParams) (*dto.PaymentListResponse, error)
	GetRecentPayments(ctx context.Context, tenantID uuid.UUID, hours int, pagination repository.PaginationParams) (*dto.PaymentListResponse, error)

	// Provider Operations
	GetPaymentsByProvider(ctx context.Context, providerName string, tenantID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error)
	GetProviderStats(ctx context.Context, tenantID uuid.UUID) (map[string]*dto.ProviderStatsResponse, error)
	GetFailedPaymentsByProvider(ctx context.Context, providerName string, tenantID uuid.UUID) ([]*dto.PaymentResponse, error)

	// Reconciliation
	GetUnreconciledPayments(ctx context.Context, tenantID uuid.UUID) ([]*dto.PaymentResponse, error)
	MarkPaymentsAsReconciled(ctx context.Context, paymentIDs []uuid.UUID) error
	GetPaymentsForReconciliation(ctx context.Context, tenantID uuid.UUID, date time.Time) ([]*dto.PaymentResponse, error)

	// Bulk Operations
	BulkMarkAsPaid(ctx context.Context, paymentIDs []uuid.UUID) error
	BulkMarkAsFailed(ctx context.Context, paymentIDs []uuid.UUID, reason string) error
}

// paymentService implements PaymentService
type paymentService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewPaymentService creates a new PaymentService instance
func NewPaymentService(repos *repository.Repositories, logger log.AllLogger) PaymentService {
	return &paymentService{
		repos:  repos,
		logger: logger,
	}
}

// ============================================================================
// Core Payment Operations
// ============================================================================

// CreatePayment creates a new payment
func (s *paymentService) CreatePayment(ctx context.Context, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.NewValidationError("invalid payment request: " + err.Error())
	}

	// Create payment model
	payment := &models.Payment{
		TenantID:          req.TenantID,
		BookingID:         req.BookingID,
		CustomerID:        req.CustomerID,
		ArtisanID:         req.ArtisanID,
		Amount:            req.Amount,
		Currency:          req.Currency,
		Method:            req.Method,
		Type:              req.Type,
		Status:            models.PaymentStatusPending,
		ProviderName:      req.ProviderName,
		ProviderPaymentID: req.ProviderPaymentID,
		CommissionRate:    req.CommissionRate,
		Metadata:          req.Metadata,
	}

	// Calculate commission split
	payment.CalculateCommission()

	// Validate payment
	if err := payment.Validate(); err != nil {
		return nil, errors.NewValidationError("payment validation failed: " + err.Error())
	}

	// Save payment
	if err := s.repos.Payment.Create(ctx, payment); err != nil {
		s.logger.Error("failed to create payment", "error", err)
		return nil, errors.NewServiceError("CREATE_FAILED", "failed to create payment", err)
	}

	s.logger.Info("payment created", "payment_id", payment.ID, "amount", payment.Amount, "currency", payment.Currency)

	// Load relationships
	payment, err := s.repos.Payment.GetByID(ctx, payment.ID)
	if err != nil {
		s.logger.Warn("failed to load payment relationships", "payment_id", payment.ID, "error", err)
	}

	return dto.ToPaymentResponse(payment), nil
}

// GetPayment retrieves a payment by ID
func (s *paymentService) GetPayment(ctx context.Context, paymentID uuid.UUID) (*dto.PaymentResponse, error) {
	if paymentID == uuid.Nil {
		return nil, errors.NewValidationError("payment ID is required")
	}

	payment, err := s.repos.Payment.GetByID(ctx, paymentID)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get payment", err)
	}

	return dto.ToPaymentResponse(payment), nil
}

// GetPaymentsByBooking retrieves all payments for a booking
func (s *paymentService) GetPaymentsByBooking(ctx context.Context, bookingID uuid.UUID) ([]*dto.PaymentResponse, error) {
	if bookingID == uuid.Nil {
		return nil, errors.NewValidationError("booking ID is required")
	}

	payments, err := s.repos.Payment.GetByBookingID(ctx, bookingID)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get payments for booking", err)
	}

	return dto.ToPaymentResponses(payments), nil
}

// GetPaymentsByCustomer retrieves all payments for a customer
func (s *paymentService) GetPaymentsByCustomer(ctx context.Context, customerID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error) {
	if customerID == uuid.Nil {
		return nil, errors.NewValidationError("customer ID is required")
	}

	payments, paginationResult, err := s.repos.Payment.GetByCustomerID(ctx, customerID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get customer payments", err)
	}

	return &dto.PaymentListResponse{
		Payments:    dto.ToPaymentResponses(payments),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetPaymentsByArtisan retrieves all payments for an artisan
func (s *paymentService) GetPaymentsByArtisan(ctx context.Context, artisanID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error) {
	if artisanID == uuid.Nil {
		return nil, errors.NewValidationError("artisan ID is required")
	}

	payments, paginationResult, err := s.repos.Payment.GetByArtisanID(ctx, artisanID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get artisan payments", err)
	}

	return &dto.PaymentListResponse{
		Payments:    dto.ToPaymentResponses(payments),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetPaymentsByTenant retrieves all payments for a tenant
func (s *paymentService) GetPaymentsByTenant(ctx context.Context, tenantID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	payments, paginationResult, err := s.repos.Payment.GetByTenantID(ctx, tenantID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get tenant payments", err)
	}

	return &dto.PaymentListResponse{
		Payments:    dto.ToPaymentResponses(payments),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetPaymentByProviderID retrieves a payment by provider payment ID
func (s *paymentService) GetPaymentByProviderID(ctx context.Context, providerPaymentID string) (*dto.PaymentResponse, error) {
	if providerPaymentID == "" {
		return nil, errors.NewValidationError("provider payment ID is required")
	}

	payment, err := s.repos.Payment.GetByProviderPaymentID(ctx, providerPaymentID)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get payment by provider ID", err)
	}

	return dto.ToPaymentResponse(payment), nil
}

// ============================================================================
// Payment Status Operations
// ============================================================================

// MarkPaymentAsPaid marks a payment as paid
func (s *paymentService) MarkPaymentAsPaid(ctx context.Context, paymentID uuid.UUID, providerPaymentID string) (*dto.PaymentResponse, error) {
	if paymentID == uuid.Nil {
		return nil, errors.NewValidationError("payment ID is required")
	}

	if err := s.repos.Payment.MarkAsPaid(ctx, paymentID, providerPaymentID); err != nil {
		return nil, errors.NewServiceError("UPDATE_FAILED", "failed to mark payment as paid", err)
	}

	s.logger.Info("payment marked as paid", "payment_id", paymentID, "provider_payment_id", providerPaymentID)

	return s.GetPayment(ctx, paymentID)
}

// MarkPaymentAsFailed marks a payment as failed
func (s *paymentService) MarkPaymentAsFailed(ctx context.Context, paymentID uuid.UUID, reason string) (*dto.PaymentResponse, error) {
	if paymentID == uuid.Nil {
		return nil, errors.NewValidationError("payment ID is required")
	}

	if err := s.repos.Payment.MarkAsFailed(ctx, paymentID, reason); err != nil {
		return nil, errors.NewServiceError("UPDATE_FAILED", "failed to mark payment as failed", err)
	}

	s.logger.Info("payment marked as failed", "payment_id", paymentID, "reason", reason)

	return s.GetPayment(ctx, paymentID)
}

// MarkPaymentAsCancelled marks a payment as cancelled
func (s *paymentService) MarkPaymentAsCancelled(ctx context.Context, paymentID uuid.UUID) (*dto.PaymentResponse, error) {
	if paymentID == uuid.Nil {
		return nil, errors.NewValidationError("payment ID is required")
	}

	if err := s.repos.Payment.MarkAsCanceled(ctx, paymentID); err != nil {
		return nil, errors.NewServiceError("UPDATE_FAILED", "failed to mark payment as cancelled", err)
	}

	s.logger.Info("payment marked as cancelled", "payment_id", paymentID)

	return s.GetPayment(ctx, paymentID)
}

// MarkPaymentAsProcessing marks a payment as processing
func (s *paymentService) MarkPaymentAsProcessing(ctx context.Context, paymentID uuid.UUID) (*dto.PaymentResponse, error) {
	if paymentID == uuid.Nil {
		return nil, errors.NewValidationError("payment ID is required")
	}

	if err := s.repos.Payment.MarkAsProcessing(ctx, paymentID); err != nil {
		return nil, errors.NewServiceError("UPDATE_FAILED", "failed to mark payment as processing", err)
	}

	return s.GetPayment(ctx, paymentID)
}

// GetPendingPayments retrieves all pending payments
func (s *paymentService) GetPendingPayments(ctx context.Context, tenantID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	payments, paginationResult, err := s.repos.Payment.GetPendingPayments(ctx, tenantID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get pending payments", err)
	}

	return &dto.PaymentListResponse{
		Payments:    dto.ToPaymentResponses(payments),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetFailedPayments retrieves all failed payments
func (s *paymentService) GetFailedPayments(ctx context.Context, tenantID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	payments, paginationResult, err := s.repos.Payment.GetFailedPayments(ctx, tenantID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get failed payments", err)
	}

	return &dto.PaymentListResponse{
		Payments:    dto.ToPaymentResponses(payments),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetSuccessfulPayments retrieves all successful payments within a date range
func (s *paymentService) GetSuccessfulPayments(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, pagination repository.PaginationParams) (*dto.PaymentListResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	payments, paginationResult, err := s.repos.Payment.GetSuccessfulPayments(ctx, tenantID, startDate, endDate, pagination)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get successful payments", err)
	}

	return &dto.PaymentListResponse{
		Payments:    dto.ToPaymentResponses(payments),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ============================================================================
// Refund Operations
// ============================================================================

// ProcessRefund processes a full or partial refund
func (s *paymentService) ProcessRefund(ctx context.Context, paymentID uuid.UUID, amount float64, reason string) (*dto.PaymentResponse, error) {
	if paymentID == uuid.Nil {
		return nil, errors.NewValidationError("payment ID is required")
	}
	if amount <= 0 {
		return nil, errors.NewValidationError("refund amount must be positive")
	}
	if reason == "" {
		return nil, errors.NewValidationError("refund reason is required")
	}

	// Get the payment
	payment, err := s.repos.Payment.GetByID(ctx, paymentID)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get payment", err)
	}

	// Validate refund
	if !payment.CanBeRefunded() {
		return nil, errors.NewValidationError(fmt.Sprintf("payment cannot be refunded (status: %s)", payment.Status))
	}

	maxRefund := payment.GetRefundableAmount()
	if amount > maxRefund {
		return nil, errors.NewValidationError(fmt.Sprintf("refund amount (%.2f) exceeds refundable amount (%.2f)", amount, maxRefund))
	}

	// Process refund
	if err := s.repos.Payment.CreateRefund(ctx, paymentID, amount, reason); err != nil {
		return nil, errors.NewServiceError("REFUND_FAILED", "failed to process refund", err)
	}

	s.logger.Info("refund processed", "payment_id", paymentID, "amount", amount, "reason", reason)

	return s.GetPayment(ctx, paymentID)
}

// ProcessPartialRefund processes a partial refund
func (s *paymentService) ProcessPartialRefund(ctx context.Context, paymentID uuid.UUID, amount float64, reason string) (*dto.PaymentResponse, error) {
	return s.ProcessRefund(ctx, paymentID, amount, reason)
}

// GetRefundablePayments retrieves all refundable payments for a booking
func (s *paymentService) GetRefundablePayments(ctx context.Context, bookingID uuid.UUID) ([]*dto.PaymentResponse, error) {
	if bookingID == uuid.Nil {
		return nil, errors.NewValidationError("booking ID is required")
	}

	payments, err := s.repos.Payment.GetRefundablePayments(ctx, bookingID)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get refundable payments", err)
	}

	return dto.ToPaymentResponses(payments), nil
}

// GetRefundedPayments retrieves all refunded payments
func (s *paymentService) GetRefundedPayments(ctx context.Context, tenantID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	payments, paginationResult, err := s.repos.Payment.GetRefundedPayments(ctx, tenantID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get refunded payments", err)
	}

	return &dto.PaymentListResponse{
		Payments:    dto.ToPaymentResponses(payments),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetPaymentRefundHistory retrieves refund history for a payment
func (s *paymentService) GetPaymentRefundHistory(ctx context.Context, paymentID uuid.UUID) ([]*dto.RefundRecordResponse, error) {
	if paymentID == uuid.Nil {
		return nil, errors.NewValidationError("payment ID is required")
	}

	refunds, err := s.repos.Payment.GetPaymentRefundHistory(ctx, paymentID)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get refund history", err)
	}

	responses := make([]*dto.RefundRecordResponse, len(refunds))
	for i, refund := range refunds {
		responses[i] = &dto.RefundRecordResponse{
			PaymentID:    refund.PaymentID,
			RefundAmount: refund.RefundAmount,
			RefundReason: refund.RefundReason,
			RefundedAt:   refund.RefundedAt,
			Status:       refund.Status,
		}
	}

	return responses, nil
}

// ============================================================================
// Commission & Earnings Operations
// ============================================================================

// CalculateCommission calculates and updates commission split
func (s *paymentService) CalculateCommission(ctx context.Context, paymentID uuid.UUID, commissionRate float64) (*dto.PaymentResponse, error) {
	if paymentID == uuid.Nil {
		return nil, errors.NewValidationError("payment ID is required")
	}
	if commissionRate < 0 || commissionRate > 100 {
		return nil, errors.NewValidationError("commission rate must be between 0 and 100")
	}

	if err := s.repos.Payment.CalculateCommissionSplit(ctx, paymentID, commissionRate); err != nil {
		return nil, errors.NewServiceError("CALCULATION_FAILED", "failed to calculate commission", err)
	}

	s.logger.Info("commission calculated", "payment_id", paymentID, "rate", commissionRate)

	return s.GetPayment(ctx, paymentID)
}

// GetArtisanEarnings retrieves total earnings for an artisan
func (s *paymentService) GetArtisanEarnings(ctx context.Context, artisanID uuid.UUID, startDate, endDate time.Time) (*dto.EarningsResponse, error) {
	if artisanID == uuid.Nil {
		return nil, errors.NewValidationError("artisan ID is required")
	}

	totalEarnings, err := s.repos.Payment.GetArtisanEarnings(ctx, artisanID, startDate, endDate)
	if err != nil {
		return nil, errors.NewServiceError("CALCULATION_FAILED", "failed to get artisan earnings", err)
	}

	return &dto.EarningsResponse{
		ArtisanID: artisanID,
		Amount:    totalEarnings,
		Currency:  "USD",
		StartDate: startDate,
		EndDate:   endDate,
		IsPaid:    false,
	}, nil
}

// GetPlatformRevenue retrieves platform revenue
func (s *paymentService) GetPlatformRevenue(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*dto.RevenueResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	revenue, err := s.repos.Payment.GetPlatformRevenue(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, errors.NewServiceError("CALCULATION_FAILED", "failed to get platform revenue", err)
	}

	return &dto.RevenueResponse{
		TenantID:  tenantID,
		Amount:    revenue,
		Currency:  "USD",
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}

// GetUnpaidArtisanEarnings retrieves unpaid earnings for an artisan
func (s *paymentService) GetUnpaidArtisanEarnings(ctx context.Context, artisanID uuid.UUID) (*dto.EarningsResponse, error) {
	if artisanID == uuid.Nil {
		return nil, errors.NewValidationError("artisan ID is required")
	}

	totalEarnings, err := s.repos.Payment.GetUnpaidArtisanEarnings(ctx, artisanID)
	if err != nil {
		return nil, errors.NewServiceError("CALCULATION_FAILED", "failed to get unpaid earnings", err)
	}

	return &dto.EarningsResponse{
		ArtisanID: artisanID,
		Amount:    totalEarnings,
		Currency:  "USD",
		IsPaid:    false,
	}, nil
}

// GetArtisanPaymentHistory retrieves payment history for an artisan
func (s *paymentService) GetArtisanPaymentHistory(ctx context.Context, artisanID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error) {
	return s.GetPaymentsByArtisan(ctx, artisanID, pagination)
}

// GetCommissionBreakdown retrieves commission breakdown
func (s *paymentService) GetCommissionBreakdown(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*dto.CommissionBreakdownResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	breakdown, err := s.repos.Payment.GetCommissionBreakdown(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, errors.NewServiceError("CALCULATION_FAILED", "failed to get commission breakdown", err)
	}

	return &dto.CommissionBreakdownResponse{
		TenantID:              tenantID,
		TotalPaid:             breakdown.TotalPaid,
		ArtisanTotal:          breakdown.ArtisanTotal,
		PlatformTotal:         breakdown.PlatformTotal,
		AverageCommissionRate: breakdown.AverageCommissionRate,
		StartDate:             startDate,
		EndDate:               endDate,
	}, nil
}

// ============================================================================
// Payment Method Operations
// ============================================================================

// GetPaymentsByMethod retrieves payments by payment method
func (s *paymentService) GetPaymentsByMethod(ctx context.Context, method models.PaymentMethod, tenantID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	payments, paginationResult, err := s.repos.Payment.GetByPaymentMethod(ctx, method, tenantID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get payments by method", err)
	}

	return &dto.PaymentListResponse{
		Payments:    dto.ToPaymentResponses(payments),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetPaymentMethodStats retrieves payment method statistics
func (s *paymentService) GetPaymentMethodStats(ctx context.Context, tenantID uuid.UUID) (map[models.PaymentMethod]int64, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	stats, err := s.repos.Payment.GetPaymentMethodStats(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("CALCULATION_FAILED", "failed to get payment method stats", err)
	}

	return stats, nil
}

// GetPreferredPaymentMethods retrieves preferred payment methods for a customer
func (s *paymentService) GetPreferredPaymentMethods(ctx context.Context, customerID uuid.UUID) ([]models.PaymentMethod, error) {
	if customerID == uuid.Nil {
		return nil, errors.NewValidationError("customer ID is required")
	}

	methods, err := s.repos.Payment.GetPreferredPaymentMethods(ctx, customerID)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get preferred payment methods", err)
	}

	return methods, nil
}

// GetPaymentStats retrieves payment statistics
func (s *paymentService) GetPaymentStats(ctx context.Context, tenantID uuid.UUID) (*dto.PaymentStatsResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	stats, err := s.repos.Payment.GetPaymentStats(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("CALCULATION_FAILED", "failed to get payment stats", err)
	}

	response := &dto.PaymentStatsResponse{
		TenantID:                tenantID,
		TotalPayments:           stats.TotalPayments,
		TotalRevenue:            stats.TotalRevenue,
		TotalRefunded:           stats.TotalRefunded,
		AverageTransactionValue: stats.AverageTransactionValue,
		SuccessRate:             stats.SuccessRate,
		PendingCount:            stats.PendingCount,
		FailedCount:             stats.FailedCount,
		RefundedCount:           stats.RefundedCount,
		ByStatus:                stats.ByStatus,
		ByMethod:                stats.ByMethod,
	}

	return response, nil
}

// GetRevenueByPeriod retrieves revenue data grouped by period
func (s *paymentService) GetRevenueByPeriod(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, groupBy string) ([]*dto.RevenueDataResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	revenueData, err := s.repos.Payment.GetRevenueByPeriod(ctx, tenantID, startDate, endDate, groupBy)
	if err != nil {
		return nil, errors.NewServiceError("CALCULATION_FAILED", "failed to get revenue by period", err)
	}

	responses := make([]*dto.RevenueDataResponse, len(revenueData))
	for i, data := range revenueData {
		responses[i] = &dto.RevenueDataResponse{
			Period:           data.Period,
			Revenue:          data.Revenue,
			TransactionCount: data.TransactionCount,
		}
	}

	return responses, nil
}

// GetDailyRevenue retrieves revenue for a specific day
func (s *paymentService) GetDailyRevenue(ctx context.Context, tenantID uuid.UUID, date time.Time) (float64, error) {
	if tenantID == uuid.Nil {
		return 0, errors.NewValidationError("tenant ID is required")
	}

	revenue, err := s.repos.Payment.GetDailyRevenue(ctx, tenantID, date)
	if err != nil {
		return 0, errors.NewServiceError("CALCULATION_FAILED", "failed to get daily revenue", err)
	}

	return revenue, nil
}

// GetMonthlyRevenue retrieves revenue for a specific month
func (s *paymentService) GetMonthlyRevenue(ctx context.Context, tenantID uuid.UUID, year int, month time.Month) (float64, error) {
	if tenantID == uuid.Nil {
		return 0, errors.NewValidationError("tenant ID is required")
	}

	revenue, err := s.repos.Payment.GetMonthlyRevenue(ctx, tenantID, year, month)
	if err != nil {
		return 0, errors.NewServiceError("CALCULATION_FAILED", "failed to get monthly revenue", err)
	}

	return revenue, nil
}

// GetYearlyRevenue retrieves revenue for a specific year
func (s *paymentService) GetYearlyRevenue(ctx context.Context, tenantID uuid.UUID, year int) (float64, error) {
	if tenantID == uuid.Nil {
		return 0, errors.NewValidationError("tenant ID is required")
	}

	revenue, err := s.repos.Payment.GetYearlyRevenue(ctx, tenantID, year)
	if err != nil {
		return 0, errors.NewServiceError("CALCULATION_FAILED", "failed to get yearly revenue", err)
	}

	return revenue, nil
}

// GetCustomerPaymentSummary retrieves payment summary for a customer
func (s *paymentService) GetCustomerPaymentSummary(ctx context.Context, customerID uuid.UUID) (*dto.CustomerPaymentSummaryResponse, error) {
	if customerID == uuid.Nil {
		return nil, errors.NewValidationError("customer ID is required")
	}

	summary, err := s.repos.Payment.GetCustomerPaymentSummary(ctx, customerID)
	if err != nil {
		return nil, errors.NewServiceError("CALCULATION_FAILED", "failed to get customer payment summary", err)
	}

	return &dto.CustomerPaymentSummaryResponse{
		CustomerID:         summary.CustomerID,
		TotalSpent:         summary.TotalSpent,
		TotalPayments:      summary.TotalPayments,
		SuccessfulPayments: summary.SuccessfulPayments,
		FailedPayments:     summary.FailedPayments,
		AveragePayment:     summary.AveragePayment,
		LastPaymentDate:    summary.LastPaymentDate,
		PreferredMethod:    summary.PreferredMethod,
	}, nil
}

// GetTopPayingCustomers retrieves top paying customers
func (s *paymentService) GetTopPayingCustomers(ctx context.Context, tenantID uuid.UUID, limit int, startDate, endDate time.Time) ([]*dto.CustomerPaymentDataResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}
	if limit <= 0 {
		limit = 10
	}

	customers, err := s.repos.Payment.GetTopPayingCustomers(ctx, tenantID, limit, startDate, endDate)
	if err != nil {
		return nil, errors.NewServiceError("CALCULATION_FAILED", "failed to get top paying customers", err)
	}

	responses := make([]*dto.CustomerPaymentDataResponse, len(customers))
	for i, customer := range customers {
		responses[i] = &dto.CustomerPaymentDataResponse{
			CustomerID:   customer.CustomerID,
			TotalSpent:   customer.TotalSpent,
			PaymentCount: customer.PaymentCount,
		}
	}

	return responses, nil
}

// GetPaymentTrends retrieves payment trends
func (s *paymentService) GetPaymentTrends(ctx context.Context, tenantID uuid.UUID, days int) ([]*dto.PaymentTrendResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}
	if days <= 0 {
		days = 30
	}

	trends, err := s.repos.Payment.GetPaymentTrends(ctx, tenantID, days)
	if err != nil {
		return nil, errors.NewServiceError("CALCULATION_FAILED", "failed to get payment trends", err)
	}

	responses := make([]*dto.PaymentTrendResponse, len(trends))
	for i, trend := range trends {
		responses[i] = &dto.PaymentTrendResponse{
			Date:             trend.Date,
			Revenue:          trend.Revenue,
			TransactionCount: trend.TransactionCount,
			SuccessfulCount:  trend.SuccessfulCount,
			FailedCount:      trend.FailedCount,
		}
	}

	return responses, nil
}

// GetAverageTransactionValue retrieves average transaction value
func (s *paymentService) GetAverageTransactionValue(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	if tenantID == uuid.Nil {
		return 0, errors.NewValidationError("tenant ID is required")
	}

	avgValue, err := s.repos.Payment.GetAverageTransactionValue(ctx, tenantID, startDate, endDate)
	if err != nil {
		return 0, errors.NewServiceError("CALCULATION_FAILED", "failed to get average transaction value", err)
	}

	return avgValue, nil
}

// ============================================================================
// Search & Filter Operations
// ============================================================================

// SearchPayments searches payments by query
func (s *paymentService) SearchPayments(ctx context.Context, query string, tenantID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	payments, paginationResult, err := s.repos.Payment.Search(ctx, query, tenantID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("SEARCH_FAILED", "failed to search payments", err)
	}

	return &dto.PaymentListResponse{
		Payments:    dto.ToPaymentResponses(payments),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// FilterPayments filters payments by criteria
func (s *paymentService) FilterPayments(ctx context.Context, filters dto.PaymentFilter, pagination repository.PaginationParams) (*dto.PaymentListResponse, error) {
	// Convert DTO filter to repository filter
	repoFilters := repository.PaymentFilters{
		TenantID:        filters.TenantID,
		Statuses:        filters.Statuses,
		Methods:         filters.Methods,
		Types:           filters.Types,
		MinAmount:       filters.MinAmount,
		MaxAmount:       filters.MaxAmount,
		CustomerID:      filters.CustomerID,
		ArtisanID:       filters.ArtisanID,
		BookingID:       filters.BookingID,
		ProcessedAfter:  filters.ProcessedAfter,
		ProcessedBefore: filters.ProcessedBefore,
		HasRefund:       filters.HasRefund,
		ProviderName:    filters.ProviderName,
	}

	payments, paginationResult, err := s.repos.Payment.FindByFilters(ctx, repoFilters, pagination)
	if err != nil {
		return nil, errors.NewServiceError("FILTER_FAILED", "failed to filter payments", err)
	}

	return &dto.PaymentListResponse{
		Payments:    dto.ToPaymentResponses(payments),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetRecentPayments retrieves recent payments
func (s *paymentService) GetRecentPayments(ctx context.Context, tenantID uuid.UUID, hours int, pagination repository.PaginationParams) (*dto.PaymentListResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}
	if hours <= 0 {
		hours = 24
	}

	payments, paginationResult, err := s.repos.Payment.GetRecentPayments(ctx, tenantID, hours, pagination)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get recent payments", err)
	}

	return &dto.PaymentListResponse{
		Payments:    dto.ToPaymentResponses(payments),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// ============================================================================
// Provider Operations
// ============================================================================

// GetPaymentsByProvider retrieves payments by provider
func (s *paymentService) GetPaymentsByProvider(ctx context.Context, providerName string, tenantID uuid.UUID, pagination repository.PaginationParams) (*dto.PaymentListResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}
	if providerName == "" {
		return nil, errors.NewValidationError("provider name is required")
	}

	payments, paginationResult, err := s.repos.Payment.GetByProvider(ctx, providerName, tenantID, pagination)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get payments by provider", err)
	}

	return &dto.PaymentListResponse{
		Payments:    dto.ToPaymentResponses(payments),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetProviderStats retrieves provider statistics
func (s *paymentService) GetProviderStats(ctx context.Context, tenantID uuid.UUID) (map[string]*dto.ProviderStatsResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	stats, err := s.repos.Payment.GetProviderStats(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("CALCULATION_FAILED", "failed to get provider stats", err)
	}

	responses := make(map[string]*dto.ProviderStatsResponse)
	for providerName, stat := range stats {
		responses[providerName] = &dto.ProviderStatsResponse{
			ProviderName:      stat.ProviderName,
			TotalTransactions: stat.TotalTransactions,
			SuccessfulCount:   stat.SuccessfulCount,
			FailedCount:       stat.FailedCount,
			TotalRevenue:      stat.TotalRevenue,
			SuccessRate:       stat.SuccessRate,
		}
	}

	return responses, nil
}

// GetFailedPaymentsByProvider retrieves failed payments for a provider
func (s *paymentService) GetFailedPaymentsByProvider(ctx context.Context, providerName string, tenantID uuid.UUID) ([]*dto.PaymentResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}
	if providerName == "" {
		return nil, errors.NewValidationError("provider name is required")
	}

	payments, err := s.repos.Payment.GetFailedPaymentsByProvider(ctx, providerName, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get failed payments by provider", err)
	}

	return dto.ToPaymentResponses(payments), nil
}

// ============================================================================
// Reconciliation Operations
// ============================================================================

// GetUnreconciledPayments retrieves unreconciled payments
func (s *paymentService) GetUnreconciledPayments(ctx context.Context, tenantID uuid.UUID) ([]*dto.PaymentResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	payments, err := s.repos.Payment.GetUnreconciledPayments(ctx, tenantID)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get unreconciled payments", err)
	}

	return dto.ToPaymentResponses(payments), nil
}

// MarkPaymentsAsReconciled marks payments as reconciled
func (s *paymentService) MarkPaymentsAsReconciled(ctx context.Context, paymentIDs []uuid.UUID) error {
	if len(paymentIDs) == 0 {
		return errors.NewValidationError("at least one payment ID is required")
	}

	if err := s.repos.Payment.MarkAsReconciled(ctx, paymentIDs); err != nil {
		return errors.NewServiceError("UPDATE_FAILED", "failed to mark payments as reconciled", err)
	}

	s.logger.Info("payments marked as reconciled", "count", len(paymentIDs))

	return nil
}

// GetPaymentsForReconciliation retrieves payments for a specific date
func (s *paymentService) GetPaymentsForReconciliation(ctx context.Context, tenantID uuid.UUID, date time.Time) ([]*dto.PaymentResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.NewValidationError("tenant ID is required")
	}

	payments, err := s.repos.Payment.GetPaymentsForReconciliation(ctx, tenantID, date)
	if err != nil {
		return nil, errors.NewServiceError("GET_FAILED", "failed to get payments for reconciliation", err)
	}

	return dto.ToPaymentResponses(payments), nil
}

// ============================================================================
// Bulk Operations
// ============================================================================

// BulkMarkAsPaid marks multiple payments as paid
func (s *paymentService) BulkMarkAsPaid(ctx context.Context, paymentIDs []uuid.UUID) error {
	if len(paymentIDs) == 0 {
		return errors.NewValidationError("at least one payment ID is required")
	}

	if err := s.repos.Payment.BulkMarkAsPaid(ctx, paymentIDs); err != nil {
		return errors.NewServiceError("BULK_UPDATE_FAILED", "failed to bulk mark as paid", err)
	}

	s.logger.Info("payments marked as paid in bulk", "count", len(paymentIDs))

	return nil
}

// BulkMarkAsFailed marks multiple payments as failed
func (s *paymentService) BulkMarkAsFailed(ctx context.Context, paymentIDs []uuid.UUID, reason string) error {
	if len(paymentIDs) == 0 {
		return errors.NewValidationError("at least one payment ID is required")
	}
	if reason == "" {
		return errors.NewValidationError("failure reason is required")
	}

	if err := s.repos.Payment.BulkMarkAsFailed(ctx, paymentIDs, reason); err != nil {
		return errors.NewServiceError("BULK_UPDATE_FAILED", "failed to bulk mark as failed", err)
	}

	s.logger.Info("payments marked as failed in bulk", "count", len(paymentIDs), "reason", reason)

	return nil
}
