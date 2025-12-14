package handler

import (
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// InvoiceHandler handles HTTP requests for invoice operations
type InvoiceHandler struct {
	invoiceService service.InvoiceService
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(invoiceService service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceService: invoiceService,
	}
}

// CreateInvoice creates a new invoice
func (h *InvoiceHandler) CreateInvoice(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)

	var req dto.CreateInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	invoice, err := h.invoiceService.CreateInvoice(c.Context(), authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, invoice, "Invoice created successfully")
}

// GetInvoice retrieves an invoice by ID
func (h *InvoiceHandler) GetInvoice(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)

	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid invoice ID", err)
	}

	invoice, err := h.invoiceService.GetInvoice(c.Context(), invoiceID, authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, invoice)
}

// UpdateInvoice updates an invoice
func (h *InvoiceHandler) UpdateInvoice(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)

	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid invoice ID", err)
	}

	var req dto.UpdateInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	invoice, err := h.invoiceService.UpdateInvoice(c.Context(), invoiceID, authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, invoice, "Invoice updated successfully")
}

// DeleteInvoice deletes an invoice
func (h *InvoiceHandler) DeleteInvoice(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)

	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid invoice ID", err)
	}

	if err := h.invoiceService.DeleteInvoice(c.Context(), invoiceID, authCtx.TenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewNoContentResponse(c)
}

// GetBookingInvoice retrieves invoices for a booking
func (h *InvoiceHandler) GetBookingInvoice(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)

	bookingID, err := uuid.Parse(c.Params("booking_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid booking ID", err)
	}

	invoices, err := h.invoiceService.ListInvoicesByBooking(c.Context(), bookingID, authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, invoices)
}

// GetCustomerInvoices retrieves invoices for a customer
func (h *InvoiceHandler) GetCustomerInvoices(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)

	customerID, err := uuid.Parse(c.Params("customer_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid customer ID", err)
	}

	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 20)

	invoices, err := h.invoiceService.ListInvoicesByCustomer(c.Context(), customerID, authCtx.TenantID, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, invoices)
}

// MarkInvoiceAsPaid marks an invoice as paid
func (h *InvoiceHandler) MarkInvoiceAsPaid(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)

	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid invoice ID", err)
	}

	var req dto.RecordPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	invoice, err := h.invoiceService.RecordPayment(c.Context(), invoiceID, authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, invoice, "Payment recorded successfully")
}

// SendInvoice sends an invoice to customer
func (h *InvoiceHandler) SendInvoice(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)

	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid invoice ID", err)
	}

	if err := h.invoiceService.SendInvoice(c.Context(), invoiceID, authCtx.TenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Invoice sent successfully")
}
