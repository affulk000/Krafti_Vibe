package handler

import (
	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ProjectUpdateHandler handles project update HTTP requests
type ProjectUpdateHandler struct {
	service service.ProjectUpdateService
}

// NewProjectUpdateHandler creates a new project update handler
func NewProjectUpdateHandler(service service.ProjectUpdateService) *ProjectUpdateHandler {
	return &ProjectUpdateHandler{
		service: service,
	}
}

// CreateProjectUpdate creates a new project update
// @Summary Create project update
// @Tags Project Updates
// @Accept json
// @Produce json
// @Param update body dto.CreateProjectUpdateRequest true "Update details"
// @Success 201 {object} dto.ProjectUpdateResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 401 {object} handler.ErrorResponse
// @Router /api/v1/project-updates [post]
func (h *ProjectUpdateHandler) CreateProjectUpdate(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.CreateProjectUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	// Set user ID from auth context
	req.UserID = authCtx.UserID

	update, err := h.service.CreateProjectUpdate(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(update)
}

// GetProjectUpdate retrieves a project update by ID
// @Summary Get project update
// @Tags Project Updates
// @Produce json
// @Param id path string true "Update ID"
// @Success 200 {object} dto.ProjectUpdateResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/project-updates/{id} [get]
func (h *ProjectUpdateHandler) GetProjectUpdate(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid update ID",
			"code":  "INVALID_UPDATE_ID",
		})
	}

	update, err := h.service.GetProjectUpdate(c.Context(), id, authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(update)
}

// UpdateProjectUpdate updates a project update
// @Summary Update project update
// @Tags Project Updates
// @Accept json
// @Produce json
// @Param id path string true "Update ID"
// @Param update body dto.UpdateProjectUpdateRequest true "Updated update details"
// @Success 200 {object} dto.ProjectUpdateResponse
// @Failure 400 {object} handler.ErrorResponse
// @Router /api/v1/project-updates/{id} [put]
func (h *ProjectUpdateHandler) UpdateProjectUpdate(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid update ID",
			"code":  "INVALID_UPDATE_ID",
		})
	}

	var req dto.UpdateProjectUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	update, err := h.service.UpdateProjectUpdate(c.Context(), id, authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(update)
}

// DeleteProjectUpdate deletes a project update
// @Summary Delete project update
// @Tags Project Updates
// @Produce json
// @Param id path string true "Update ID"
// @Success 200 {object} handler.SuccessResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/project-updates/{id} [delete]
func (h *ProjectUpdateHandler) DeleteProjectUpdate(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid update ID",
			"code":  "INVALID_UPDATE_ID",
		})
	}

	if err := h.service.DeleteProjectUpdate(c.Context(), id, authCtx.TenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "project update deleted successfully",
	})
}

// ListProjectUpdates lists project updates with filtering
// @Summary List project updates
// @Tags Project Updates
// @Produce json
// @Param project_id query string true "Project ID"
// @Param type query string false "Filter by update type"
// @Param visible_to_customer query bool false "Filter by customer visibility"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.ProjectUpdateListResponse
// @Router /api/v1/project-updates [get]
func (h *ProjectUpdateHandler) ListProjectUpdates(c *fiber.Ctx) error {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "project_id is required",
			"code":  "MISSING_PROJECT_ID",
		})
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid project_id",
			"code":  "INVALID_PROJECT_ID",
		})
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("limit", 20)

	filter := &dto.ProjectUpdateFilter{
		ProjectID: projectID,
		Page:      page,
		PageSize:  pageSize,
	}

	// Apply type filter if provided
	if updateType := c.Query("type"); updateType != "" {
		t := models.UpdateType(updateType)
		filter.Type = &t
	}

	// Apply visibility filter if provided
	if visibleStr := c.Query("visible_to_customer"); visibleStr != "" {
		visible := visibleStr == "true"
		filter.VisibleToCustomer = &visible
	}

	updates, err := h.service.ListProjectUpdates(c.Context(), filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(updates)
}

// ListByType lists project updates by type
// @Summary List project updates by type
// @Tags Project Updates
// @Produce json
// @Param project_id path string true "Project ID"
// @Param type path string true "Update type"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.ProjectUpdateListResponse
// @Router /api/v1/project-updates/project/{project_id}/type/{type} [get]
func (h *ProjectUpdateHandler) ListByType(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	projectID, err := uuid.Parse(c.Params("project_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid project_id",
			"code":  "INVALID_PROJECT_ID",
		})
	}

	updateType := models.UpdateType(c.Params("type"))
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("limit", 20)

	updates, err := h.service.ListByType(c.Context(), projectID, authCtx.TenantID, updateType, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(updates)
}

// ListCustomerVisible lists customer-visible project updates
// @Summary List customer-visible updates
// @Tags Project Updates
// @Produce json
// @Param project_id path string true "Project ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.ProjectUpdateListResponse
// @Router /api/v1/project-updates/project/{project_id}/customer-visible [get]
func (h *ProjectUpdateHandler) ListCustomerVisible(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	projectID, err := uuid.Parse(c.Params("project_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid project_id",
			"code":  "INVALID_PROJECT_ID",
		})
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("limit", 20)

	updates, err := h.service.ListCustomerVisible(c.Context(), projectID, authCtx.TenantID, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(updates)
}

// GetLatestUpdate retrieves the latest update for a project
// @Summary Get latest project update
// @Tags Project Updates
// @Produce json
// @Param project_id path string true "Project ID"
// @Success 200 {object} dto.ProjectUpdateResponse
// @Failure 404 {object} handler.ErrorResponse
// @Router /api/v1/project-updates/project/{project_id}/latest [get]
func (h *ProjectUpdateHandler) GetLatestUpdate(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	projectID, err := uuid.Parse(c.Params("project_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid project_id",
			"code":  "INVALID_PROJECT_ID",
		})
	}

	update, err := h.service.GetLatestUpdate(c.Context(), projectID, authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(update)
}
