package handler

import (
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// PaymentHandler handles HTTP requests for payment operations
type PaymentHandler struct {
	paymentService service.PaymentService
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// ============================================================================
// Core Payment Operations
// ============================================================================

// CreatePayment godoc
// @Summary Create a new payment
// @Description Create a new payment record
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body dto.CreatePaymentRequest true "Payment creation data"
// @Success 201 {object} dto.PaymentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments [post]
func (h *PaymentHandler) CreatePayment(c *fiber.Ctx) error {
	var req dto.CreatePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	payment, err := h.paymentService.CreatePayment(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, payment, "Payment created successfully")
}

// GetPayment godoc
// @Summary Get payment by ID
// @Description Get detailed payment information by ID
// @Tags payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} dto.PaymentResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/{id} [get]
func (h *PaymentHandler) GetPayment(c *fiber.Ctx) error {
	paymentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid payment ID", err)
	}

	payment, err := h.paymentService.GetPayment(c.Context(), paymentID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, payment)
}

// GetPaymentsByBooking godoc
// @Summary Get payments by booking
// @Description Get all payments for a specific booking
// @Tags payments
// @Produce json
// @Param booking_id path string true "Booking ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/booking/{booking_id} [get]
func (h *PaymentHandler) GetPaymentsByBooking(c *fiber.Ctx) error {
	bookingID, err := uuid.Parse(c.Params("booking_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid booking ID", err)
	}

	payments, err := h.paymentService.GetPaymentsByBooking(c.Context(), bookingID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, payments)
}

// GetPaymentsByCustomer godoc
// @Summary Get payments by customer
// @Description Get all payments for a specific customer
// @Tags payments
// @Produce json
// @Param customer_id path string true "Customer ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.PaymentListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/customer/{customer_id} [get]
func (h *PaymentHandler) GetPaymentsByCustomer(c *fiber.Ctx) error {
	customerID, err := uuid.Parse(c.Params("customer_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid customer ID", err)
	}

	pagination := repository.PaginationParams{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	payments, err := h.paymentService.GetPaymentsByCustomer(c.Context(), customerID, pagination)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, payments)
}

// GetPaymentsByArtisan godoc
// @Summary Get payments by artisan
// @Description Get all payments for a specific artisan
// @Tags payments
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.PaymentListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/artisan/{artisan_id} [get]
func (h *PaymentHandler) GetPaymentsByArtisan(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("artisan_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid artisan ID", err)
	}

	pagination := repository.PaginationParams{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	payments, err := h.paymentService.GetPaymentsByArtisan(c.Context(), artisanID, pagination)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, payments)
}

// GetPaymentsByTenant godoc
// @Summary Get payments by tenant
// @Description Get all payments for a specific tenant
// @Tags payments
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.PaymentListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/tenant [get]
func (h *PaymentHandler) GetPaymentsByTenant(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	pagination := repository.PaginationParams{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	payments, err := h.paymentService.GetPaymentsByTenant(c.Context(), tenantID, pagination)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, payments)
}

// ============================================================================
// Payment Status Operations
// ============================================================================

// MarkPaymentAsPaid godoc
// @Summary Mark payment as paid
// @Description Mark a payment as successfully paid
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param status body MarkPaidRequest true "Provider payment ID"
// @Success 200 {object} dto.PaymentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/{id}/mark-paid [post]
func (h *PaymentHandler) MarkPaymentAsPaid(c *fiber.Ctx) error {
	paymentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid payment ID", err)
	}

	var req MarkPaidRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	payment, err := h.paymentService.MarkPaymentAsPaid(c.Context(), paymentID, req.ProviderPaymentID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, payment, "Payment marked as paid")
}

// MarkPaymentAsFailed godoc
// @Summary Mark payment as failed
// @Description Mark a payment as failed
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param status body MarkFailedRequest true "Failure reason"
// @Success 200 {object} dto.PaymentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/{id}/mark-failed [post]
func (h *PaymentHandler) MarkPaymentAsFailed(c *fiber.Ctx) error {
	paymentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid payment ID", err)
	}

	var req MarkFailedRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	payment, err := h.paymentService.MarkPaymentAsFailed(c.Context(), paymentID, req.Reason)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, payment, "Payment marked as failed")
}

// GetPendingPayments godoc
// @Summary Get pending payments
// @Description Get all pending payments for a tenant
// @Tags payments
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.PaymentListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/pending [get]
func (h *PaymentHandler) GetPendingPayments(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	pagination := repository.PaginationParams{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	payments, err := h.paymentService.GetPendingPayments(c.Context(), tenantID, pagination)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, payments)
}

// ============================================================================
// Refund Operations
// ============================================================================

// ProcessRefund godoc
// @Summary Process refund
// @Description Process a full or partial refund for a payment
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param refund body RefundRequest true "Refund data"
// @Success 200 {object} dto.PaymentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/{id}/refund [post]
func (h *PaymentHandler) ProcessRefund(c *fiber.Ctx) error {
	paymentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid payment ID", err)
	}

	var req RefundRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	payment, err := h.paymentService.ProcessRefund(c.Context(), paymentID, req.Amount, req.Reason)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, payment, "Refund processed successfully")
}

// GetRefundablePayments godoc
// @Summary Get refundable payments
// @Description Get all refundable payments for a booking
// @Tags payments
// @Produce json
// @Param booking_id path string true "Booking ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/booking/{booking_id}/refundable [get]
func (h *PaymentHandler) GetRefundablePayments(c *fiber.Ctx) error {
	bookingID, err := uuid.Parse(c.Params("booking_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid booking ID", err)
	}

	payments, err := h.paymentService.GetRefundablePayments(c.Context(), bookingID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, payments)
}

// ============================================================================
// Commission & Earnings
// ============================================================================

// GetArtisanEarnings godoc
// @Summary Get artisan earnings
// @Description Get earnings summary for an artisan
// @Tags payments
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Param start_date query string true "Start date (RFC3339)"
// @Param end_date query string true "End date (RFC3339)"
// @Success 200 {object} dto.EarningsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/artisan/{artisan_id}/earnings [get]
func (h *PaymentHandler) GetArtisanEarnings(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("artisan_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid artisan ID", err)
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest, "MISSING_DATES", "start_date and end_date are required", nil)
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid start_date format", err)
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid end_date format", err)
	}

	earnings, err := h.paymentService.GetArtisanEarnings(c.Context(), artisanID, startDate, endDate)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, earnings)
}

// GetPlatformRevenue godoc
// @Summary Get platform revenue
// @Description Get revenue summary for a tenant
// @Tags payments
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param start_date query string true "Start date (RFC3339)"
// @Param end_date query string true "End date (RFC3339)"
// @Success 200 {object} dto.RevenueResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/revenue [get]
func (h *PaymentHandler) GetPlatformRevenue(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest, "MISSING_DATES", "start_date and end_date are required", nil)
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid start_date format", err)
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid end_date format", err)
	}

	revenue, err := h.paymentService.GetPlatformRevenue(c.Context(), tenantID, startDate, endDate)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, revenue)
}

// ============================================================================
// Analytics & Reporting
// ============================================================================

// GetPaymentStats godoc
// @Summary Get payment statistics
// @Description Get comprehensive payment statistics
// @Tags payments
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} dto.PaymentStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/stats [get]
func (h *PaymentHandler) GetPaymentStats(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	stats, err := h.paymentService.GetPaymentStats(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, stats)
}

// GetPaymentTrends godoc
// @Summary Get payment trends
// @Description Get payment trends over time
// @Tags payments
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param days query int false "Number of days" default(30)
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/trends [get]
func (h *PaymentHandler) GetPaymentTrends(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	days := getIntQuery(c, "days", 30)

	trends, err := h.paymentService.GetPaymentTrends(c.Context(), tenantID, days)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, trends)
}

// GetPaymentsByMethod godoc
// @Summary Get payments by method
// @Description Get payments filtered by payment method
// @Tags payments
// @Produce json
// @Param method path string true "Payment method"
// @Param tenant_id query string true "Tenant ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.PaymentListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /payments/method/{method} [get]
func (h *PaymentHandler) GetPaymentsByMethod(c *fiber.Ctx) error {
	method := models.PaymentMethod(c.Params("method"))

	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	pagination := repository.PaginationParams{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	payments, err := h.paymentService.GetPaymentsByMethod(c.Context(), method, tenantID, pagination)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, payments)
}

// ============================================================================
// Request Types
// ============================================================================

type MarkPaidRequest struct {
	ProviderPaymentID string `json:"provider_payment_id"`
}

type MarkFailedRequest struct {
	Reason string `json:"reason"`
}

type RefundRequest struct {
	Amount float64 `json:"amount"`
	Reason string  `json:"reason"`
}
