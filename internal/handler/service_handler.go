package handler

import (
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ServiceHandler handles HTTP requests for service catalog operations
type ServiceHandler struct {
	serviceService service.ServiceService
}

// NewServiceHandler creates a new service handler
func NewServiceHandler(serviceService service.ServiceService) *ServiceHandler {
	return &ServiceHandler{
		serviceService: serviceService,
	}
}

// CreateService godoc
// @Summary Create a new service
// @Description Create a new service in the catalog
// @Tags services
// @Accept json
// @Produce json
// @Param service body dto.CreateServiceRequest true "Service creation data"
// @Success 201 {object} dto.ServiceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /services [post]
func (h *ServiceHandler) CreateService(c *fiber.Ctx) error {
	var req dto.CreateServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	svc, err := h.serviceService.CreateService(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, svc, "Service created successfully")
}

// GetServiceByID godoc
// @Summary Get service by ID
// @Description Get detailed service information by ID
// @Tags services
// @Produce json
// @Param id path string true "Service ID"
// @Success 200 {object} dto.ServiceResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /services/{id} [get]
func (h *ServiceHandler) GetServiceByID(c *fiber.Ctx) error {
	serviceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid service ID", err)
	}

	svc, err := h.serviceService.GetServiceByID(c.Context(), serviceID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, svc)
}

// UpdateService godoc
// @Summary Update service
// @Description Update service information
// @Tags services
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param service body dto.UpdateServiceRequest true "Update data"
// @Success 200 {object} dto.ServiceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /services/{id} [put]
func (h *ServiceHandler) UpdateService(c *fiber.Ctx) error {
	serviceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid service ID", err)
	}

	var req dto.UpdateServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	svc, err := h.serviceService.UpdateService(c.Context(), serviceID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, svc, "Service updated successfully")
}

// DeleteService godoc
// @Summary Delete service
// @Description Delete a service from the catalog
// @Tags services
// @Param id path string true "Service ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /services/{id} [delete]
func (h *ServiceHandler) DeleteService(c *fiber.Ctx) error {
	serviceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid service ID", err)
	}

	if err := h.serviceService.DeleteService(c.Context(), serviceID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewNoContentResponse(c)
}

// ListServices godoc
// @Summary List services
// @Description Get a paginated list of services
// @Tags services
// @Accept json
// @Produce json
// @Param req body dto.ListServicesRequest true "List parameters"
// @Success 200 {object} dto.ListServicesResponse
// @Failure 500 {object} ErrorResponse
// @Router /services/list [post]
func (h *ServiceHandler) ListServices(c *fiber.Ctx) error {
	var req dto.ListServicesRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	services, err := h.serviceService.ListServices(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, services)
}

// SearchServices godoc
// @Summary Search services
// @Description Search for services
// @Tags services
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.ListServicesResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /services/search [get]
func (h *ServiceHandler) SearchServices(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	query := c.Query("q")
	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 20)

	services, err := h.serviceService.SearchServices(c.Context(), tenantID, query, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, services)
}

// ActivateService godoc
// @Summary Activate service
// @Description Activate a service
// @Tags services
// @Param id path string true "Service ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /services/{id}/activate [post]
func (h *ServiceHandler) ActivateService(c *fiber.Ctx) error {
	serviceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid service ID", err)
	}

	if err := h.serviceService.ActivateService(c.Context(), serviceID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Service activated successfully")
}

// DeactivateService godoc
// @Summary Deactivate service
// @Description Deactivate a service
// @Tags services
// @Param id path string true "Service ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /services/{id}/deactivate [post]
func (h *ServiceHandler) DeactivateService(c *fiber.Ctx) error {
	serviceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid service ID", err)
	}

	if err := h.serviceService.DeactivateService(c.Context(), serviceID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Service deactivated successfully")
}

// GetServiceStatistics godoc
// @Summary Get service statistics
// @Description Get comprehensive service statistics
// @Tags services
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} dto.ServiceStatistics
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /services/stats [get]
func (h *ServiceHandler) GetServiceStatistics(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	stats, err := h.serviceService.GetServiceStatistics(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, stats)
}

// GetPopularServices godoc
// @Summary Get popular services
// @Description Get most popular services
// @Tags services
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param limit query int false "Limit" default(10)
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /services/popular [get]
func (h *ServiceHandler) GetPopularServices(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	limit := getIntQuery(c, "limit", 10)

	services, err := h.serviceService.GetPopularServices(c.Context(), tenantID, limit)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, services)
}
