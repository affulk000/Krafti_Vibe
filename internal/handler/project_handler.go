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

// ProjectHandler handles HTTP requests for project operations
type ProjectHandler struct {
	projectService service.ProjectService
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(projectService service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// ============================================================================
// CRUD Operations
// ============================================================================

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new project
// @Tags projects
// @Accept json
// @Produce json
// @Param project body dto.CreateProjectRequest true "Project creation data"
// @Success 201 {object} dto.ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects [post]
func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	var req dto.CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	project, err := h.projectService.CreateProject(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewCreatedResponse(c, project, "Project created successfully")
}

// GetProject godoc
// @Summary Get project by ID
// @Description Get basic project information by ID
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/{id} [get]
func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid project ID", err)
	}

	project, err := h.projectService.GetProject(c.Context(), projectID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, project)
}

// GetProjectWithDetails godoc
// @Summary Get project with full details
// @Description Get detailed project information including milestones and tasks
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/{id}/details [get]
func (h *ProjectHandler) GetProjectWithDetails(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid project ID", err)
	}

	project, err := h.projectService.GetProjectWithDetails(c.Context(), projectID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, project)
}

// UpdateProject godoc
// @Summary Update project
// @Description Update project information
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param project body dto.UpdateProjectRequest true "Update data"
// @Success 200 {object} dto.ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/{id} [put]
func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid project ID", err)
	}

	var req dto.UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	project, err := h.projectService.UpdateProject(c.Context(), projectID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, project, "Project updated successfully")
}

// DeleteProject godoc
// @Summary Delete project
// @Description Delete a project
// @Tags projects
// @Param id path string true "Project ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid project ID", err)
	}

	if err := h.projectService.DeleteProject(c.Context(), projectID); err != nil {
		return HandleServiceError(c, err)
	}

	return NewNoContentResponse(c)
}

// ============================================================================
// Query Operations
// ============================================================================

// ListProjects godoc
// @Summary List projects
// @Description Get a paginated list of projects with filters
// @Tags projects
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param tenant_id query string false "Filter by tenant ID"
// @Param status query string false "Filter by status"
// @Success 200 {object} dto.ProjectListResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects [get]
func (h *ProjectHandler) ListProjects(c *fiber.Ctx) error {
	filter := dto.ProjectFilter{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	// Parse tenant ID if provided
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		if tenantID, err := uuid.Parse(tenantIDStr); err == nil {
			filter.TenantID = tenantID
		}
	}

	// Parse status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status := models.ProjectStatus(statusStr)
		filter.Status = &status
	}

	projects, err := h.projectService.ListProjects(c.Context(), filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, projects)
}

// SearchProjects godoc
// @Summary Search projects
// @Description Search for projects
// @Tags projects
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.ProjectListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/search [get]
func (h *ProjectHandler) SearchProjects(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	query := c.Query("q")
	if query == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest, "MISSING_QUERY", "Search query is required", nil)
	}

	pagination := repository.PaginationParams{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	projects, err := h.projectService.SearchProjects(c.Context(), tenantID, query, pagination)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, projects)
}

// GetProjectsByArtisan godoc
// @Summary Get projects by artisan
// @Description Get all projects for a specific artisan
// @Tags projects
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.ProjectListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/artisan/{artisan_id} [get]
func (h *ProjectHandler) GetProjectsByArtisan(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("artisan_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid artisan ID", err)
	}

	pagination := repository.PaginationParams{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	projects, err := h.projectService.GetProjectsByArtisan(c.Context(), artisanID, pagination)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, projects)
}

// GetProjectsByCustomer godoc
// @Summary Get projects by customer
// @Description Get all projects for a specific customer
// @Tags projects
// @Produce json
// @Param customer_id path string true "Customer ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.ProjectListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/customer/{customer_id} [get]
func (h *ProjectHandler) GetProjectsByCustomer(c *fiber.Ctx) error {
	customerID, err := uuid.Parse(c.Params("customer_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid customer ID", err)
	}

	pagination := repository.PaginationParams{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	projects, err := h.projectService.GetProjectsByCustomer(c.Context(), customerID, pagination)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, projects)
}

// GetOverdueProjects godoc
// @Summary Get overdue projects
// @Description Get all overdue projects
// @Tags projects
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/overdue [get]
func (h *ProjectHandler) GetOverdueProjects(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	projects, err := h.projectService.GetOverdueProjects(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, projects)
}

// GetActiveProjects godoc
// @Summary Get active projects
// @Description Get all active projects
// @Tags projects
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/active [get]
func (h *ProjectHandler) GetActiveProjects(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	projects, err := h.projectService.GetActiveProjects(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, projects)
}

// ============================================================================
// Status Management
// ============================================================================

// StartProject godoc
// @Summary Start project
// @Description Start a project
// @Tags projects
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/{id}/start [post]
func (h *ProjectHandler) StartProject(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid project ID", err)
	}

	project, err := h.projectService.StartProject(c.Context(), projectID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, project, "Project started successfully")
}

// PauseProject godoc
// @Summary Pause project
// @Description Pause a project
// @Tags projects
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/{id}/pause [post]
func (h *ProjectHandler) PauseProject(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid project ID", err)
	}

	project, err := h.projectService.PauseProject(c.Context(), projectID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, project, "Project paused successfully")
}

// ResumeProject godoc
// @Summary Resume project
// @Description Resume a paused project
// @Tags projects
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/{id}/resume [post]
func (h *ProjectHandler) ResumeProject(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid project ID", err)
	}

	project, err := h.projectService.ResumeProject(c.Context(), projectID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, project, "Project resumed successfully")
}

// CompleteProject godoc
// @Summary Complete project
// @Description Mark a project as completed
// @Tags projects
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/{id}/complete [post]
func (h *ProjectHandler) CompleteProject(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid project ID", err)
	}

	project, err := h.projectService.CompleteProject(c.Context(), projectID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, project, "Project completed successfully")
}

// CancelProject godoc
// @Summary Cancel project
// @Description Cancel a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param cancel body CancelProjectRequest true "Cancellation reason"
// @Success 200 {object} dto.ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/{id}/cancel [post]
func (h *ProjectHandler) CancelProject(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid project ID", err)
	}

	var req CancelProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	project, err := h.projectService.CancelProject(c.Context(), projectID, req.Reason)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, project, "Project cancelled successfully")
}

// ============================================================================
// Analytics & Statistics
// ============================================================================

// GetProjectStats godoc
// @Summary Get project statistics
// @Description Get comprehensive project statistics
// @Tags projects
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} dto.ProjectStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/stats [get]
func (h *ProjectHandler) GetProjectStats(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	stats, err := h.projectService.GetProjectStats(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, stats)
}

// GetArtisanProjectStats godoc
// @Summary Get artisan project statistics
// @Description Get project statistics for an artisan
// @Tags projects
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Success 200 {object} dto.ArtisanProjectStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/artisan/{artisan_id}/stats [get]
func (h *ProjectHandler) GetArtisanProjectStats(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("artisan_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid artisan ID", err)
	}

	stats, err := h.projectService.GetArtisanProjectStats(c.Context(), artisanID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, stats)
}

// GetProjectHealth godoc
// @Summary Get project health
// @Description Get project health metrics
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectHealthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/{id}/health [get]
func (h *ProjectHandler) GetProjectHealth(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid project ID", err)
	}

	health, err := h.projectService.GetProjectHealth(c.Context(), projectID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, health)
}

// GetProjectTimeline godoc
// @Summary Get project timeline
// @Description Get project timeline
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectTimelineResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/{id}/timeline [get]
func (h *ProjectHandler) GetProjectTimeline(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid project ID", err)
	}

	timeline, err := h.projectService.GetProjectTimeline(c.Context(), projectID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, timeline)
}

// ============================================================================
// Dashboard
// ============================================================================

// GetArtisanDashboard godoc
// @Summary Get artisan dashboard
// @Description Get dashboard data for an artisan
// @Tags projects
// @Produce json
// @Param artisan_id path string true "Artisan ID"
// @Success 200 {object} dto.ArtisanDashboardResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/artisan/{artisan_id}/dashboard [get]
func (h *ProjectHandler) GetArtisanDashboard(c *fiber.Ctx) error {
	artisanID, err := uuid.Parse(c.Params("artisan_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid artisan ID", err)
	}

	dashboard, err := h.projectService.GetArtisanDashboard(c.Context(), artisanID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, dashboard)
}

// GetTenantDashboard godoc
// @Summary Get tenant dashboard
// @Description Get dashboard data for a tenant
// @Tags projects
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} dto.TenantProjectDashboardResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/dashboard [get]
func (h *ProjectHandler) GetTenantDashboard(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	dashboard, err := h.projectService.GetTenantDashboard(c.Context(), tenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, dashboard)
}

// ============================================================================
// Bulk Operations
// ============================================================================

// BulkUpdateStatus godoc
// @Summary Bulk update project status
// @Description Update status for multiple projects
// @Tags projects
// @Accept json
// @Produce json
// @Param bulk body dto.BulkProjectUpdateRequest true "Bulk update data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/bulk/status [put]
func (h *ProjectHandler) BulkUpdateStatus(c *fiber.Ctx) error {
	var req dto.BulkProjectUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.projectService.BulkUpdateStatus(c.Context(), &req); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Project statuses updated successfully")
}

// ArchiveCompletedProjects godoc
// @Summary Archive completed projects
// @Description Archive projects completed before a certain date
// @Tags projects
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param days query int false "Days ago" default(90)
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /projects/archive [post]
func (h *ProjectHandler) ArchiveCompletedProjects(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID", err)
	}

	days := getIntQuery(c, "days", 90)
	olderThan := time.Duration(days) * 24 * time.Hour

	if err := h.projectService.ArchiveCompletedProjects(c.Context(), tenantID, olderThan); err != nil {
		return HandleServiceError(c, err)
	}

	return NewSuccessResponse(c, nil, "Completed projects archived successfully")
}

// ============================================================================
// Request Types
// ============================================================================

type CancelProjectRequest struct {
	Reason string `json:"reason"`
}
