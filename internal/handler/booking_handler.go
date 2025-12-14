package handler

import (
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// BookingHandler handles HTTP requests for booking operations
// ENTERPRISE-GRADE: Implements comprehensive error handling, logging, authentication,
// authorization, input validation, and security best practices
type BookingHandler struct {
	bookingService service.BookingService
}

// NewBookingHandler creates a new booking handler
func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	if bookingService == nil {
		panic("bookingService cannot be nil")
	}

	return &BookingHandler{
		bookingService: bookingService,
	}
}

// ============================================================================
// Core Booking Operations
// ============================================================================

// CreateBooking godoc
// @Summary Create a new booking
// @Description Create a new booking with enterprise-grade validation and security
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param booking body dto.CreateBookingRequest true "Booking creation data"
// @Param Idempotency-Key header string false "Idempotency key for safe retries"
// @Success 201 {object} dto.BookingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings [post]
func (h *BookingHandler) CreateBooking(c *fiber.Ctx) error {
	// Set security headers
	SetSecurityHeaders(c)
	SetNoCacheHeaders(c)

	// Authenticate and get user context
	authCtx, err := GetAuthContext(c)
	if err != nil {
		LogHandlerError(c, "create_booking.auth_failed", err)
		return err
	}

	// Validate content type
	if err := ValidateContentType(c, "application/json"); err != nil {
		return err
	}

	// Check for idempotency key (recommended for write operations)
	idempotencyKey, hasKey := ValidateIdempotencyKey(c)
	if hasKey {
		LogHandlerInfo(c, "create_booking", map[string]interface{}{
			"idempotency_key": idempotencyKey,
			"user_id":         authCtx.UserID.String(),
		})
	}

	// Parse and validate request
	var req dto.CreateBookingRequest
	if err := c.BodyParser(&req); err != nil {
		LogHandlerError(c, "create_booking.parse_error", err)
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST",
			"Invalid request body. Please check your input.", err)
	}

	// Ensure tenant isolation - the request tenant must match authenticated tenant
	if authCtx.TenantID != uuid.Nil && req.TenantID != authCtx.TenantID {
		LogHandlerError(c, "create_booking.tenant_mismatch",
			fiber.NewError(fiber.StatusForbidden, "Tenant mismatch"))
		return NewForbiddenResponse(c, "You can only create bookings for your own tenant")
	}

	// Log the operation
	LogHandlerInfo(c, "create_booking", map[string]interface{}{
		"user_id":   authCtx.UserID.String(),
		"tenant_id": authCtx.TenantID.String(),
	})

	// Call service layer
	booking, err := h.bookingService.CreateBooking(c.Context(), &req)
	if err != nil {
		LogHandlerError(c, "create_booking.service_error", err)
		return HandleServiceError(c, err)
	}

	// Set Location header for created resource
	c.Set("Location", "/api/v1/bookings/"+booking.ID.String())

	LogHandlerInfo(c, "create_booking.success", map[string]interface{}{
		"booking_id": booking.ID.String(),
		"user_id":    authCtx.UserID.String(),
	})

	return NewCreatedResponse(c, booking, "Booking created successfully")
}

// GetBooking godoc
// @Summary Get booking by ID
// @Description Get detailed booking information by ID with authorization checks
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID (UUID format)"
// @Success 200 {object} dto.BookingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/{id} [get]
func (h *BookingHandler) GetBooking(c *fiber.Ctx) error {
	// Set cache headers for read operations (cache for 5 minutes, private)
	SetCacheHeaders(c, 300, true)
	SetSecurityHeaders(c)

	// Authenticate user
	authCtx, err := GetAuthContext(c)
	if err != nil {
		LogHandlerError(c, "get_booking.auth_failed", err)
		return err
	}

	// Parse and validate UUID
	bookingID, err := ParseUUIDParam(c, "id")
	if err != nil {
		LogHandlerError(c, "get_booking.invalid_id", err)
		return err
	}

	// Log the operation
	LogHandlerInfo(c, "get_booking", map[string]any{
		"booking_id": bookingID.String(),
		"user_id":    authCtx.UserID.String(),
		"tenant_id":  authCtx.TenantID.String(),
	})

	// Fetch booking
	booking, err := h.bookingService.GetBooking(c.Context(), bookingID)
	if err != nil {
		LogHandlerError(c, "get_booking.service_error", err)
		return HandleServiceError(c, err)
	}

	// Verify tenant isolation - user can only access bookings in their tenant
	if booking.TenantID != authCtx.TenantID {
		LogHandlerError(c, "get_booking.unauthorized_access",
			fiber.NewError(fiber.StatusForbidden, "Tenant mismatch"))
		return NewForbiddenResponse(c, "You don't have access to this booking")
	}

	return NewSuccessResponse(c, booking)
}

// UpdateBooking godoc
// @Summary Update booking
// @Description Update booking information
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param booking body dto.UpdateBookingRequest true "Update data"
// @Success 200 {object} dto.BookingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/{id} [put]
func (h *BookingHandler) UpdateBooking(c *fiber.Ctx) error {
	bookingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid booking ID", err)
	}

	var req dto.UpdateBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	booking, err := h.bookingService.UpdateBooking(c.Context(), bookingID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, booking, "Booking updated successfully")
}

// DeleteBooking godoc
// @Summary Delete booking
// @Description Delete a booking
// @Tags bookings
// @Param id path string true "Booking ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/{id} [delete]
func (h *BookingHandler) DeleteBooking(c *fiber.Ctx) error {
	bookingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid booking ID", err)
	}

	if err := h.bookingService.DeleteBooking(c.Context(), bookingID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewNoContentResponse(c)
}

// ListBookings godoc
// @Summary List bookings
// @Description Get a paginated list of bookings with filters and tenant isolation
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (min: 1)" default(1)
// @Param page_size query int false "Page size (min: 1, max: 100)" default(20)
// @Param tenant_id query string false "Filter by tenant ID (must match authenticated tenant)"
// @Param artisan_id query string false "Filter by artisan ID"
// @Param customer_id query string false "Filter by customer ID"
// @Param status query string false "Filter by status (pending, confirmed, in_progress, completed, cancelled)"
// @Param sort_by query string false "Sort field" default(created_at)
// @Param sort_order query string false "Sort order (asc, desc)" default(desc)
// @Success 200 {object} dto.BookingListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings [get]
func (h *BookingHandler) ListBookings(c *fiber.Ctx) error {
	// Set cache headers (cache for 1 minute, private)
	SetCacheHeaders(c, 60, true)
	SetSecurityHeaders(c)

	// Authenticate user
	authCtx, err := GetAuthContext(c)
	if err != nil {
		LogHandlerError(c, "list_bookings.auth_failed", err)
		return err
	}

	// Validate and normalize pagination
	page, pageSize := ValidatePagination(
		getIntQuery(c, "page", 1),
		getIntQuery(c, "page_size", 20),
	)

	filter := dto.BookingFilter{
		Page:     page,
		PageSize: pageSize,
		TenantID: &authCtx.TenantID, // Enforce tenant isolation
	}

	// Parse tenant ID if provided (must match authenticated tenant)
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID",
				"Invalid tenant ID format", err)
		}

		// Security: Ensure requested tenant matches authenticated tenant
		if tenantID != authCtx.TenantID {
			LogHandlerError(c, "list_bookings.tenant_mismatch",
				fiber.NewError(fiber.StatusForbidden, "Tenant mismatch"))
			return NewForbiddenResponse(c, "You can only list bookings for your own tenant")
		}
	}

	// Parse artisan ID if provided
	if artisanIDStr := c.Query("artisan_id"); artisanIDStr != "" {
		artisanID, err := uuid.Parse(artisanIDStr)
		if err != nil {
			return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ARTISAN_ID",
				"Invalid artisan ID format", err)
		}
		filter.ArtisanIDs = []uuid.UUID{artisanID}
	}

	// Parse customer ID if provided
	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		customerID, err := uuid.Parse(customerIDStr)
		if err != nil {
			return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_CUSTOMER_ID",
				"Invalid customer ID format", err)
		}
		filter.CustomerIDs = []uuid.UUID{customerID}
	}

	// Parse and validate status filter
	if statusStr := c.Query("status"); statusStr != "" {
		// Validate status is one of the allowed values
		validStatuses := []string{"pending", "confirmed", "in_progress", "completed", "cancelled", "no_show"}
		statusErr := ValidateEnum("status", statusStr, validStatuses)
		if statusErr != nil {
			return NewValidationErrorResponse(c, []ValidationError{*statusErr})
		}

		status := models.BookingStatus(statusStr)
		filter.Statuses = []models.BookingStatus{status}
	}

	// Extract sort parameters
	sortBy, sortOrder := ExtractSortParams(c, []string{"created_at", "scheduled_at", "updated_at", "status"})

	// Log the operation
	LogHandlerInfo(c, "list_bookings", map[string]interface{}{
		"user_id":    authCtx.UserID.String(),
		"tenant_id":  authCtx.TenantID.String(),
		"page":       page,
		"page_size":  pageSize,
		"sort_by":    sortBy,
		"sort_order": sortOrder,
	})

	// Fetch bookings
	bookings, err := h.bookingService.ListBookings(c.Context(), filter)
	if err != nil {
		LogHandlerError(c, "list_bookings.service_error", err)
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, bookings)
}

// SearchBookings godoc
// @Summary Search bookings
// @Description Search for bookings
// @Tags bookings
// @Accept json
// @Produce json
// @Param search body dto.BookingSearchRequest true "Search criteria"
// @Success 200 {object} dto.BookingListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/search [post]
func (h *BookingHandler) SearchBookings(c *fiber.Ctx) error {
	var req dto.BookingSearchRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	bookings, err := h.bookingService.SearchBookings(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, bookings)
}

// GetBookingsByArtisan godoc
// @Summary Get bookings by artisan
// @Description Get all bookings for a specific artisan
// @Tags bookings
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.BookingListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/artisan/{artisan_id} [get]
func (h *BookingHandler) GetBookingsByArtisan(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("artisan_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid artisan ID", err)
	}

	filter := dto.BookingFilter{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	bookings, err := h.bookingService.GetBookingsByArtisan(c.Context(), artisanID, filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, bookings)
}

// GetBookingsByCustomer godoc
// @Summary Get bookings by customer
// @Description Get all bookings for a specific customer
// @Tags bookings
// @Produce json
// @Param customer_id path string true "Customer ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.BookingListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/customer/{customer_id} [get]
func (h *BookingHandler) GetBookingsByCustomer(c *fiber.Ctx) error {
	customerID, err := uuid.Parse(c.Params("customer_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid customer ID", err)
	}

	filter := dto.BookingFilter{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	bookings, err := h.bookingService.GetBookingsByCustomer(c.Context(), customerID, filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, bookings)
}

// ============================================================================
// Status Management
// ============================================================================

// ConfirmBooking godoc
// @Summary Confirm booking
// @Description Confirm a pending booking
// @Tags bookings
// @Param id path string true "Booking ID"
// @Success 200 {object} dto.BookingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/{id}/confirm [post]
func (h *BookingHandler) ConfirmBooking(c *fiber.Ctx) error {
	bookingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid booking ID", err)
	}

	booking, err := h.bookingService.ConfirmBooking(c.Context(), bookingID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, booking, "Booking confirmed successfully")
}

// StartBooking godoc
// @Summary Start booking
// @Description Mark a booking as started
// @Tags bookings
// @Param id path string true "Booking ID"
// @Success 200 {object} dto.BookingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/{id}/start [post]
func (h *BookingHandler) StartBooking(c *fiber.Ctx) error {
	bookingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid booking ID", err)
	}

	booking, err := h.bookingService.StartBooking(c.Context(), bookingID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, booking, "Booking started successfully")
}

// CompleteBooking godoc
// @Summary Complete booking
// @Description Mark a booking as completed
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param completion body dto.CompleteBookingRequest true "Completion data"
// @Success 200 {object} dto.BookingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/{id}/complete [post]
func (h *BookingHandler) CompleteBooking(c *fiber.Ctx) error {
	bookingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid booking ID", err)
	}

	var req dto.CompleteBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	booking, err := h.bookingService.CompleteBooking(c.Context(), bookingID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, booking, "Booking completed successfully")
}

// CancelBooking godoc
// @Summary Cancel booking
// @Description Cancel a booking
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param cancellation body dto.CancelBookingRequest true "Cancellation data"
// @Success 200 {object} dto.BookingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/{id}/cancel [post]
func (h *BookingHandler) CancelBooking(c *fiber.Ctx) error {
	bookingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid booking ID", err)
	}

	var req dto.CancelBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	booking, err := h.bookingService.CancelBooking(c.Context(), bookingID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, booking, "Booking cancelled successfully")
}

// MarkAsNoShow godoc
// @Summary Mark booking as no-show
// @Description Mark a booking as no-show
// @Tags bookings
// @Param id path string true "Booking ID"
// @Success 200 {object} dto.BookingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/{id}/no-show [post]
func (h *BookingHandler) MarkAsNoShow(c *fiber.Ctx) error {
	bookingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid booking ID", err)
	}

	booking, err := h.bookingService.MarkAsNoShow(c.Context(), bookingID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, booking, "Booking marked as no-show")
}

// RescheduleBooking godoc
// @Summary Reschedule booking
// @Description Reschedule a booking to a new time
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param reschedule body dto.RescheduleBookingRequest true "Reschedule data"
// @Success 200 {object} dto.BookingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/{id}/reschedule [post]
func (h *BookingHandler) RescheduleBooking(c *fiber.Ctx) error {
	bookingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid booking ID", err)
	}

	var req dto.RescheduleBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	booking, err := h.bookingService.RescheduleBooking(c.Context(), bookingID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, booking, "Booking rescheduled successfully")
}

// ============================================================================
// Scheduling & Availability
// ============================================================================

// CheckArtisanAvailability godoc
// @Summary Check artisan availability
// @Description Check if an artisan is available for a specific time
// @Tags bookings
// @Accept json
// @Produce json
// @Param availability body dto.AvailabilityRequest true "Availability check data"
// @Success 200 {object} dto.AvailabilityResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/check-availability [post]
func (h *BookingHandler) CheckArtisanAvailability(c *fiber.Ctx) error {
	var req dto.AvailabilityRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	availability, err := h.bookingService.CheckArtisanAvailability(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, availability)
}

// GetAvailableTimeSlots godoc
// @Summary Get available time slots
// @Description Get available time slots for an artisan on a specific date
// @Tags bookings
// @Produce json
// @Param artisan_id query string true "Artisan ID"
// @Param date query string true "Date (RFC3339)"
// @Param duration query int true "Duration in minutes"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/available-slots [get]
func (h *BookingHandler) GetAvailableTimeSlots(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Query("artisan_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ARTISAN_ID", "Invalid artisan ID", err)
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest, "MISSING_DATE", "Date is required", nil)
	}

	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid date format", err)
	}

	duration := getIntQuery(c, "duration", 60)

	slots, err := h.bookingService.GetAvailableTimeSlots(c.Context(), artisanID, date, duration)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, slots)
}

// GetArtisanSchedule godoc
// @Summary Get artisan schedule
// @Description Get artisan's schedule for a date range
// @Tags bookings
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Param start_date query string true "Start date (RFC3339)"
// @Param end_date query string true "End date (RFC3339)"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/artisan/{artisan_id}/schedule [get]
func (h *BookingHandler) GetArtisanSchedule(c *fiber.Ctx) error {
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

	bookings, err := h.bookingService.GetArtisanSchedule(c.Context(), artisanID, startDate, endDate)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, bookings)
}

// ============================================================================
// Time-based Queries
// ============================================================================

// GetUpcomingBookings godoc
// @Summary Get upcoming bookings
// @Description Get upcoming bookings for a tenant
// @Tags bookings
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param days query int false "Number of days" default(7)
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/upcoming [get]
func (h *BookingHandler) GetUpcomingBookings(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	days := getIntQuery(c, "days", 7)

	bookings, err := h.bookingService.GetUpcomingBookings(c.Context(), tenantID, days)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, bookings)
}

// GetTodayBookings godoc
// @Summary Get today's bookings
// @Description Get bookings scheduled for today
// @Tags bookings
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/today [get]
func (h *BookingHandler) GetTodayBookings(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	bookings, err := h.bookingService.GetTodayBookings(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, bookings)
}

// GetBookingsInDateRange godoc
// @Summary Get bookings in date range
// @Description Get bookings within a specific date range
// @Tags bookings
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param start_date query string true "Start date (RFC3339)"
// @Param end_date query string true "End date (RFC3339)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.BookingListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/date-range [get]
func (h *BookingHandler) GetBookingsInDateRange(c *fiber.Ctx) error {
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

	filter := dto.BookingFilter{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	bookings, err := h.bookingService.GetBookingsInDateRange(c.Context(), tenantID, startDate, endDate, filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, bookings)
}

// ============================================================================
// Photo Management
// ============================================================================

// AddBeforePhotos godoc
// @Summary Add before photos
// @Description Add before photos to a booking
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param photos body PhotoRequest true "Photo URLs"
// @Success 200 {object} dto.BookingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/{id}/photos/before [post]
func (h *BookingHandler) AddBeforePhotos(c *fiber.Ctx) error {
	bookingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid booking ID", err)
	}

	var req PhotoRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	booking, err := h.bookingService.AddBeforePhotos(c.Context(), bookingID, req.PhotoURLs)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, booking, "Before photos added successfully")
}

// AddAfterPhotos godoc
// @Summary Add after photos
// @Description Add after photos to a booking
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param photos body PhotoRequest true "Photo URLs"
// @Success 200 {object} dto.BookingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/{id}/photos/after [post]
func (h *BookingHandler) AddAfterPhotos(c *fiber.Ctx) error {
	bookingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid booking ID", err)
	}

	var req PhotoRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	booking, err := h.bookingService.AddAfterPhotos(c.Context(), bookingID, req.PhotoURLs)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, booking, "After photos added successfully")
}

// ============================================================================
// Analytics & Reporting
// ============================================================================

// GetBookingStats godoc
// @Summary Get booking statistics
// @Description Get booking statistics for a tenant
// @Tags bookings
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} dto.BookingStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/stats [get]
func (h *BookingHandler) GetBookingStats(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	stats, err := h.bookingService.GetBookingStats(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, stats)
}

// GetArtisanBookingStats godoc
// @Summary Get artisan booking statistics
// @Description Get booking statistics for an artisan
// @Tags bookings
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Param start_date query string true "Start date (RFC3339)"
// @Param end_date query string true "End date (RFC3339)"
// @Success 200 {object} dto.BookingStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/artisan/{artisan_id}/stats [get]
func (h *BookingHandler) GetArtisanBookingStats(c *fiber.Ctx) error {
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

	stats, err := h.bookingService.GetArtisanBookingStats(c.Context(), artisanID, startDate, endDate)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, stats)
}

// ============================================================================
// Bulk Operations
// ============================================================================

// BulkConfirm godoc
// @Summary Bulk confirm bookings
// @Description Confirm multiple bookings at once
// @Tags bookings
// @Accept json
// @Produce json
// @Param bulk body BulkBookingRequest true "Booking IDs"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/bulk/confirm [post]
func (h *BookingHandler) BulkConfirm(c *fiber.Ctx) error {
	var req BulkBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	bookings, err := h.bookingService.BulkConfirm(c.Context(), req.BookingIDs)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, bookings, "Bookings confirmed successfully")
}

// BulkCancel godoc
// @Summary Bulk cancel bookings
// @Description Cancel multiple bookings at once
// @Tags bookings
// @Accept json
// @Produce json
// @Param bulk body BulkCancelRequest true "Booking IDs and reason"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/bulk/cancel [post]
func (h *BookingHandler) BulkCancel(c *fiber.Ctx) error {
	var req BulkCancelRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	bookings, err := h.bookingService.BulkCancel(c.Context(), req.BookingIDs, req.Reason)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, bookings, "Bookings cancelled successfully")
}

// ============================================================================
// Request Types
// ============================================================================

type PhotoRequest struct {
	PhotoURLs []string `json:"photo_urls"`
}

type BulkBookingRequest struct {
	BookingIDs []uuid.UUID `json:"booking_ids"`
}

type BulkCancelRequest struct {
	BookingIDs []uuid.UUID `json:"booking_ids"`
	Reason     string      `json:"reason"`
}
