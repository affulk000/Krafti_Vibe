package handler

import (
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CustomerHandler handles HTTP requests for customer operations
type CustomerHandler struct {
	customerService service.CustomerService
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(customerService service.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
	}
}

// ============================================================================
// Core Operations
// ============================================================================

// CreateCustomer godoc
// @Summary Create a new customer
// @Description Create a new customer profile
// @Tags customers
// @Accept json
// @Produce json
// @Param customer body dto.CreateCustomerRequest true "Customer creation data"
// @Success 201 {object} dto.CustomerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers [post]
func (h *CustomerHandler) CreateCustomer(c *fiber.Ctx) error {
	var req dto.CreateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	customer, err := h.customerService.CreateCustomer(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, customer, "Customer created successfully")
}

// GetCustomer godoc
// @Summary Get customer by ID
// @Description Get detailed customer information by ID
// @Tags customers
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} dto.CustomerResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{id} [get]
func (h *CustomerHandler) GetCustomer(c *fiber.Ctx) error {
	customerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid customer ID", err)
	}

	customer, err := h.customerService.GetCustomer(c.Context(), customerID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, customer)
}

// GetCustomerByUserID godoc
// @Summary Get customer by user ID
// @Description Get customer profile by user ID
// @Tags customers
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} dto.CustomerResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/user/{user_id} [get]
func (h *CustomerHandler) GetCustomerByUserID(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	customer, err := h.customerService.GetCustomerByUserID(c.Context(), userID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, customer)
}

// UpdateCustomer godoc
// @Summary Update customer
// @Description Update customer information
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param customer body dto.UpdateCustomerRequest true "Update data"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(c *fiber.Ctx) error {
	customerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid customer ID", err)
	}

	var req dto.UpdateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	customer, err := h.customerService.UpdateCustomer(c.Context(), customerID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, customer, "Customer updated successfully")
}

// DeleteCustomer godoc
// @Summary Delete customer
// @Description Soft delete a customer profile
// @Tags customers
// @Param id path string true "Customer ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(c *fiber.Ctx) error {
	customerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid customer ID", err)
	}

	if err := h.customerService.DeleteCustomer(c.Context(), customerID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewNoContentResponse(c)
}

// ListCustomers godoc
// @Summary List customers
// @Description Get a paginated list of customers with filters
// @Tags customers
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param tenant_id query string false "Filter by tenant ID"
// @Success 200 {object} dto.CustomerListResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers [get]
func (h *CustomerHandler) ListCustomers(c *fiber.Ctx) error {
	// Get auth context for tenant isolation
	authCtx := MustGetAuthContext(c)

	filter := dto.CustomerFilter{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	// Use tenant ID from auth context (for tenant isolation)
	// Platform users can optionally filter by tenant via query param
	if authCtx.TenantID != uuid.Nil {
		filter.TenantID = &authCtx.TenantID
	} else if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		// Platform users can filter by specific tenant
		if tenantID, err := uuid.Parse(tenantIDStr); err == nil {
			filter.TenantID = &tenantID
		}
	}

	customers, err := h.customerService.ListCustomers(c.Context(), filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, customers)
}

// SearchCustomers godoc
// @Summary Search customers
// @Description Search for customers
// @Tags customers
// @Accept json
// @Produce json
// @Param search body dto.CustomerSearchRequest true "Search criteria"
// @Success 200 {object} dto.CustomerListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/search [post]
func (h *CustomerHandler) SearchCustomers(c *fiber.Ctx) error {
	var req dto.CustomerSearchRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	customers, err := h.customerService.SearchCustomers(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, customers)
}

// GetActiveCustomers godoc
// @Summary Get active customers
// @Description Get all active customers
// @Tags customers
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.CustomerListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/active [get]
func (h *CustomerHandler) GetActiveCustomers(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	filter := dto.CustomerFilter{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	customers, err := h.customerService.GetActiveCustomers(c.Context(), tenantID, filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, customers)
}

// ============================================================================
// Loyalty & Rewards
// ============================================================================

// UpdateLoyaltyPoints godoc
// @Summary Update loyalty points
// @Description Update customer loyalty points
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param loyalty body dto.UpdateLoyaltyPointsRequest true "Loyalty points data"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{id}/loyalty-points [put]
func (h *CustomerHandler) UpdateLoyaltyPoints(c *fiber.Ctx) error {
	customerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid customer ID", err)
	}

	var req dto.UpdateLoyaltyPointsRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	customer, err := h.customerService.UpdateLoyaltyPoints(c.Context(), customerID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, customer, "Loyalty points updated successfully")
}

// GetLoyaltyPointsHistory godoc
// @Summary Get loyalty points history
// @Description Get customer loyalty points transaction history
// @Tags customers
// @Produce json
// @Param id path string true "Customer ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{id}/loyalty-history [get]
func (h *CustomerHandler) GetLoyaltyPointsHistory(c *fiber.Ctx) error {
	customerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid customer ID", err)
	}

	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 20)

	history, err := h.customerService.GetLoyaltyPointsHistory(c.Context(), customerID, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, history)
}

// GetTopCustomers godoc
// @Summary Get top customers
// @Description Get top customers by various criteria
// @Tags customers
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param criteria query string false "Criteria (spending, bookings, loyalty)" default("spending")
// @Param limit query int false "Limit" default(10)
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/top [get]
func (h *CustomerHandler) GetTopCustomers(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	criteria := c.Query("criteria", "spending")
	limit := getIntQuery(c, "limit", 10)

	customers, err := h.customerService.GetTopCustomers(c.Context(), tenantID, criteria, limit)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, customers)
}

// ============================================================================
// Preferences & Settings
// ============================================================================

// AddPreferredArtisan godoc
// @Summary Add preferred artisan
// @Description Add an artisan to customer's preferred list
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param preferred body dto.AddPreferredArtisanRequest true "Preferred artisan data"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{id}/preferred-artisans [post]
func (h *CustomerHandler) AddPreferredArtisan(c *fiber.Ctx) error {
	customerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid customer ID", err)
	}

	var req dto.AddPreferredArtisanRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	customer, err := h.customerService.AddPreferredArtisan(c.Context(), customerID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, customer, "Preferred artisan added successfully")
}

// RemovePreferredArtisan godoc
// @Summary Remove preferred artisan
// @Description Remove an artisan from customer's preferred list
// @Tags customers
// @Param id path string true "Customer ID"
// @Param artisan_id path string true "Artisan ID"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{id}/preferred-artisans/{artisan_id} [delete]
func (h *CustomerHandler) RemovePreferredArtisan(c *fiber.Ctx) error {
	customerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid customer ID", err)
	}

	artisanID, err := uuid.Parse(c.Params("artisan_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ARTISAN_ID", "Invalid artisan ID", err)
	}

	customer, err := h.customerService.RemovePreferredArtisan(c.Context(), customerID, artisanID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, customer, "Preferred artisan removed successfully")
}

// UpdateNotificationPreferences godoc
// @Summary Update notification preferences
// @Description Update customer notification preferences
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param preferences body dto.UpdateNotificationPreferencesRequest true "Notification preferences"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{id}/notification-preferences [put]
func (h *CustomerHandler) UpdateNotificationPreferences(c *fiber.Ctx) error {
	customerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid customer ID", err)
	}

	var req dto.UpdateNotificationPreferencesRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	customer, err := h.customerService.UpdateNotificationPreferences(c.Context(), customerID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, customer, "Notification preferences updated successfully")
}

// UpdatePrimaryLocation godoc
// @Summary Update primary location
// @Description Update customer's primary location
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param location body models.Location true "Location data"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/{id}/primary-location [put]
func (h *CustomerHandler) UpdatePrimaryLocation(c *fiber.Ctx) error {
	customerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid customer ID", err)
	}

	var location models.Location
	if err := c.BodyParser(&location); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	customer, err := h.customerService.UpdatePrimaryLocation(c.Context(), customerID, &location)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, customer, "Primary location updated successfully")
}

// ============================================================================
// Analytics & Reporting
// ============================================================================

// GetCustomerStats godoc
// @Summary Get customer statistics
// @Description Get customer statistics for a tenant
// @Tags customers
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param start_date query string true "Start date (RFC3339)"
// @Param end_date query string true "End date (RFC3339)"
// @Success 200 {object} dto.CustomerStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/stats [get]
func (h *CustomerHandler) GetCustomerStats(c *fiber.Ctx) error {
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

	stats, err := h.customerService.GetCustomerStats(c.Context(), tenantID, startDate, endDate)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, stats)
}

// GetCustomerSegments godoc
// @Summary Get customer segments
// @Description Get customer segmentation data
// @Tags customers
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /customers/segments [get]
func (h *CustomerHandler) GetCustomerSegments(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	segments, err := h.customerService.GetCustomerSegments(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, segments)
}
