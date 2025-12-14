package handler

import (
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SystemSettingsHandler handles system settings HTTP requests
type SystemSettingsHandler struct {
	service service.SystemSettingService
}

// NewSystemSettingsHandler creates a new system settings handler
func NewSystemSettingsHandler(service service.SystemSettingService) *SystemSettingsHandler {
	return &SystemSettingsHandler{
		service: service,
	}
}

// CreateSetting creates a new system setting
// @Summary Create system setting
// @Tags System Settings
// @Accept json
// @Produce json
// @Param setting body dto.CreateSettingRequest true "Setting details"
// @Success 201 {object} dto.SystemSettingResponse
// @Failure 400 {object} fiber.Map
// @Failure 401 {object} fiber.Map
// @Router /api/v1/settings [post]
func (h *SystemSettingsHandler) CreateSetting(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.CreateSettingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	setting, err := h.service.CreateSetting(c.Context(), authCtx.UserID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(setting)
}

// GetSetting retrieves a system setting by ID
// @Summary Get system setting
// @Tags System Settings
// @Produce json
// @Param id path string true "Setting ID"
// @Success 200 {object} dto.SystemSettingResponse
// @Failure 404 {object} fiber.Map
// @Router /api/v1/settings/{id} [get]
func (h *SystemSettingsHandler) GetSetting(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid setting ID",
			"code":  "INVALID_SETTING_ID",
		})
	}

	setting, err := h.service.GetSetting(c.Context(), id)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(setting)
}

// GetSettingByKey retrieves a system setting by key
// @Summary Get system setting by key
// @Tags System Settings
// @Produce json
// @Param key path string true "Setting key"
// @Success 200 {object} dto.SystemSettingResponse
// @Failure 404 {object} fiber.Map
// @Router /api/v1/settings/key/{key} [get]
func (h *SystemSettingsHandler) GetSettingByKey(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "setting key is required",
			"code":  "MISSING_KEY",
		})
	}

	setting, err := h.service.GetSettingByKey(c.Context(), key)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(setting)
}

// UpdateSetting updates a system setting
// @Summary Update system setting
// @Tags System Settings
// @Accept json
// @Produce json
// @Param id path string true "Setting ID"
// @Param setting body dto.UpdateSettingRequest true "Updated setting details"
// @Success 200 {object} dto.SystemSettingResponse
// @Failure 400 {object} fiber.Map
// @Router /api/v1/settings/{id} [put]
func (h *SystemSettingsHandler) UpdateSetting(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid setting ID",
			"code":  "INVALID_SETTING_ID",
		})
	}

	var req dto.UpdateSettingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	setting, err := h.service.UpdateSetting(c.Context(), id, authCtx.UserID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(setting)
}

// DeleteSetting deletes a system setting
// @Summary Delete system setting
// @Tags System Settings
// @Produce json
// @Param id path string true "Setting ID"
// @Success 200 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Router /api/v1/settings/{id} [delete]
func (h *SystemSettingsHandler) DeleteSetting(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid setting ID",
			"code":  "INVALID_SETTING_ID",
		})
	}

	if err := h.service.DeleteSetting(c.Context(), id, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "setting deleted successfully",
	})
}

// ListSettings lists all system settings with filtering and pagination
// @Summary List system settings
// @Tags System Settings
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Filter by category"
// @Param group query string false "Filter by group"
// @Param is_public query bool false "Filter by public/private"
// @Success 200 {object} dto.SystemSettingListResponse
// @Router /api/v1/settings [get]
func (h *SystemSettingsHandler) ListSettings(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("limit", 20)

	filter := &dto.SettingFilter{
		Page:     page,
		PageSize: pageSize,
	}

	// Apply filters
	if category := c.Query("category"); category != "" {
		filter.Categories = []string{category}
	}

	if group := c.Query("group"); group != "" {
		filter.Groups = []string{group}
	}

	if isPublic := c.Query("is_public"); isPublic != "" {
		val := isPublic == "true"
		filter.IsPublic = &val
	}

	settings, err := h.service.ListSettings(c.Context(), filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(settings)
}

// GetSettingsByCategory retrieves settings by category
// @Summary Get settings by category
// @Tags System Settings
// @Produce json
// @Param category path string true "Category name"
// @Success 200 {object} []dto.SystemSettingResponse
// @Router /api/v1/settings/category/{category} [get]
func (h *SystemSettingsHandler) GetSettingsByCategory(c *fiber.Ctx) error {
	category := c.Params("category")
	if category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "category is required",
			"code":  "MISSING_CATEGORY",
		})
	}

	settings, err := h.service.GetByCategory(c.Context(), category)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(settings)
}

// GetSettingsByGroup retrieves settings by group
// @Summary Get settings by group
// @Tags System Settings
// @Produce json
// @Param group path string true "Group name"
// @Success 200 {object} []dto.SystemSettingResponse
// @Router /api/v1/settings/group/{group} [get]
func (h *SystemSettingsHandler) GetSettingsByGroup(c *fiber.Ctx) error {
	group := c.Params("group")
	if group == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "group is required",
			"code":  "MISSING_GROUP",
		})
	}

	settings, err := h.service.GetByGroup(c.Context(), group)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(settings)
}

// BulkSetSettings updates multiple settings at once
// @Summary Bulk set settings
// @Tags System Settings
// @Accept json
// @Produce json
// @Param settings body dto.BulkSetSettingsRequest true "Settings to set"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Router /api/v1/settings/bulk [post]
func (h *SystemSettingsHandler) BulkSetSettings(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.BulkSetSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	err = h.service.BulkSetSettings(c.Context(), &req, authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "settings updated successfully",
	})
}

// BulkDeleteSettings deletes multiple settings at once
// @Summary Bulk delete settings
// @Tags System Settings
// @Accept json
// @Produce json
// @Param keys body []string true "Setting keys to delete"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Router /api/v1/settings/bulk/delete [post]
func (h *SystemSettingsHandler) BulkDeleteSettings(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var keys []string
	if err := c.BodyParser(&keys); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	err = h.service.BulkDeleteSettings(c.Context(), keys, authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "settings deleted successfully",
	})
}

// GetPublicSettings retrieves all public settings
// @Summary Get public settings
// @Tags System Settings
// @Produce json
// @Success 200 {object} []dto.SystemSettingResponse
// @Router /api/v1/settings/public [get]
func (h *SystemSettingsHandler) GetPublicSettings(c *fiber.Ctx) error {
	settings, err := h.service.GetPublicSettings(c.Context())
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(settings)
}

// GetPrivateSettings retrieves all private settings
// @Summary Get private settings
// @Tags System Settings
// @Produce json
// @Success 200 {object} []dto.SystemSettingResponse
// @Router /api/v1/settings/private [get]
func (h *SystemSettingsHandler) GetPrivateSettings(c *fiber.Ctx) error {
	settings, err := h.service.GetPrivateSettings(c.Context())
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(settings)
}

// SearchSettings searches settings by key or description
// @Summary Search settings
// @Tags System Settings
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.SystemSettingListResponse
// @Router /api/v1/settings/search [get]
func (h *SystemSettingsHandler) SearchSettings(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "search query is required",
			"code":  "MISSING_QUERY",
		})
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("limit", 20)

	settings, err := h.service.SearchSettings(c.Context(), query, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(settings)
}

// GetSettingCategories retrieves all available setting categories
// @Summary Get setting categories
// @Tags System Settings
// @Produce json
// @Success 200 {object} []string
// @Router /api/v1/settings/categories [get]
func (h *SystemSettingsHandler) GetSettingCategories(c *fiber.Ctx) error {
	categories, err := h.service.GetAllCategories(c.Context())
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(categories)
}

// GetSettingGroups retrieves all available setting groups
// @Summary Get setting groups
// @Tags System Settings
// @Produce json
// @Success 200 {object} []string
// @Router /api/v1/settings/groups [get]
func (h *SystemSettingsHandler) GetSettingGroups(c *fiber.Ctx) error {
	groups, err := h.service.GetAllGroups(c.Context())
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(groups)
}

// GetCategoriesWithCount retrieves categories with setting counts
// @Summary Get categories with counts
// @Tags System Settings
// @Produce json
// @Success 200 {object} []dto.CategoryWithCountResponse
// @Router /api/v1/settings/categories/stats [get]
func (h *SystemSettingsHandler) GetCategoriesWithCount(c *fiber.Ctx) error {
	categories, err := h.service.GetCategoriesWithCount(c.Context())
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(categories)
}

// DeleteSettingByKey deletes a setting by its key
// @Summary Delete setting by key
// @Tags System Settings
// @Produce json
// @Param key path string true "Setting key"
// @Success 200 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Router /api/v1/settings/key/{key} [delete]
func (h *SystemSettingsHandler) DeleteSettingByKey(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	key := c.Params("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "setting key is required",
			"code":  "MISSING_KEY",
		})
	}

	if err := h.service.DeleteSettingByKey(c.Context(), key, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "setting deleted successfully",
	})
}
