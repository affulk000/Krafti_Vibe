package handler

import (
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SDKHandler handles SDK client and key management endpoints
type SDKHandler struct {
	sdkService service.SDKService
}

// NewSDKHandler creates a new SDK handler
func NewSDKHandler(sdkService service.SDKService) *SDKHandler {
	return &SDKHandler{
		sdkService: sdkService,
	}
}

// ============================================================================
// SDK Client Endpoints
// ============================================================================

// CreateSDKClient handles creating a new SDK client
func (h *SDKHandler) CreateSDKClient(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.CreateSDKClientRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	client, err := h.sdkService.CreateClient(c.Context(), authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(client)
}

// GetSDKClient handles retrieving an SDK client by ID
func (h *SDKHandler) GetSDKClient(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	clientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid client ID",
			"code":  "INVALID_CLIENT_ID",
		})
	}

	client, err := h.sdkService.GetClient(c.Context(), clientID, authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(client)
}

// UpdateSDKClient handles updating an SDK client
func (h *SDKHandler) UpdateSDKClient(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	clientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid client ID",
			"code":  "INVALID_CLIENT_ID",
		})
	}

	var req dto.UpdateSDKClientRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	client, err := h.sdkService.UpdateClient(c.Context(), clientID, authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(client)
}

// DeleteSDKClient handles deleting an SDK client
func (h *SDKHandler) DeleteSDKClient(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	clientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid client ID",
			"code":  "INVALID_CLIENT_ID",
		})
	}

	if err := h.sdkService.DeleteClient(c.Context(), clientID, authCtx.TenantID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListSDKClients handles listing SDK clients with filtering
func (h *SDKHandler) ListSDKClients(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	filter := &dto.SDKClientFilter{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	// Parse platform filter
	if platformStr := c.Query("platform"); platformStr != "" {
		platform := models.SDKPlatform(platformStr)
		filter.Platform = &platform
	}

	// Parse environment filter
	if envStr := c.Query("environment"); envStr != "" {
		env := models.SDKEnvironment(envStr)
		filter.Environment = &env
	}

	// Parse is_active filter
	if c.Query("is_active") != "" {
		isActive := c.Query("is_active") == "true"
		filter.IsActive = &isActive
	}

	result, err := h.sdkService.ListClients(c.Context(), authCtx.TenantID, filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(result)
}

// ============================================================================
// SDK Key Endpoints
// ============================================================================

// CreateSDKKey handles creating a new SDK key
func (h *SDKHandler) CreateSDKKey(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	var req dto.CreateSDKKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	keyWithSecret, err := h.sdkService.CreateKey(c.Context(), authCtx.TenantID, authCtx.UserID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(keyWithSecret)
}

// GetSDKKey handles retrieving an SDK key by ID
func (h *SDKHandler) GetSDKKey(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	keyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid key ID",
			"code":  "INVALID_KEY_ID",
		})
	}

	key, err := h.sdkService.GetKey(c.Context(), keyID, authCtx.TenantID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(key)
}

// UpdateSDKKey handles updating an SDK key
func (h *SDKHandler) UpdateSDKKey(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	keyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid key ID",
			"code":  "INVALID_KEY_ID",
		})
	}

	var req dto.UpdateSDKKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	key, err := h.sdkService.UpdateKey(c.Context(), keyID, authCtx.TenantID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(key)
}

// RevokeSDKKey handles revoking an SDK key
func (h *SDKHandler) RevokeSDKKey(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	keyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid key ID",
			"code":  "INVALID_KEY_ID",
		})
	}

	var req dto.RevokeSDKKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	if err := h.sdkService.RevokeKey(c.Context(), keyID, authCtx.TenantID, authCtx.UserID, &req); err != nil {
		return HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "SDK key revoked successfully",
	})
}

// RotateSDKKey handles rotating an SDK key
func (h *SDKHandler) RotateSDKKey(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	keyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid key ID",
			"code":  "INVALID_KEY_ID",
		})
	}

	var req dto.RotateSDKKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	newKey, err := h.sdkService.RotateKey(c.Context(), keyID, authCtx.TenantID, authCtx.UserID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(newKey)
}

// ListSDKKeys handles listing SDK keys with filtering
func (h *SDKHandler) ListSDKKeys(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	filter := &dto.SDKKeyFilter{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	// Parse client_id filter
	if clientIDStr := c.Query("client_id"); clientIDStr != "" {
		clientID, err := uuid.Parse(clientIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid client_id",
				"code":  "INVALID_CLIENT_ID",
			})
		}
		filter.ClientID = &clientID
	}

	// Parse environment filter
	if envStr := c.Query("environment"); envStr != "" {
		env := models.SDKEnvironment(envStr)
		filter.Environment = &env
	}

	// Parse status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status := models.SDKKeyStatus(statusStr)
		filter.Status = &status
	}

	result, err := h.sdkService.ListKeys(c.Context(), authCtx.TenantID, filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(result)
}

// ValidateSDKKey handles validating an SDK key
func (h *SDKHandler) ValidateSDKKey(c *fiber.Ctx) error {
	var req struct {
		APIKey string `json:"api_key"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	if req.APIKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "api_key is required",
			"code":  "MISSING_API_KEY",
		})
	}

	key, err := h.sdkService.ValidateKey(c.Context(), req.APIKey)
	if err != nil {
		return NewUnauthorizedResponse(c, "Invalid or expired key")
	}

	return c.JSON(fiber.Map{
		"valid":     true,
		"client_id": key.ClientID,
		"key_id":    key.ID,
		"scopes":    key.Scopes,
	})
}

// ============================================================================
// SDK Usage Endpoints
// ============================================================================

// TrackSDKUsage handles recording SDK usage
func (h *SDKHandler) TrackSDKUsage(c *fiber.Ctx) error {
	var usage models.SDKUsage
	if err := c.BodyParser(&usage); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST_BODY",
		})
	}

	// Set timestamp if not provided
	if usage.Timestamp.IsZero() {
		usage.Timestamp = time.Now()
	}

	if err := h.sdkService.RecordUsage(c.Context(), &usage); err != nil {
		return HandleServiceError(c, err)
	}

	return c.SendStatus(fiber.StatusAccepted)
}

// ListSDKUsage handles listing SDK usage records
func (h *SDKHandler) ListSDKUsage(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	filter := &dto.SDKUsageFilter{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
	}

	// Parse client_id filter
	if clientIDStr := c.Query("client_id"); clientIDStr != "" {
		clientID, err := uuid.Parse(clientIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid client_id",
				"code":  "INVALID_CLIENT_ID",
			})
		}
		filter.ClientID = &clientID
	}

	// Parse key_id filter
	if keyIDStr := c.Query("key_id"); keyIDStr != "" {
		keyID, err := uuid.Parse(keyIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid key_id",
				"code":  "INVALID_KEY_ID",
			})
		}
		filter.KeyID = &keyID
	}

	// Parse is_error filter
	if c.Query("is_error") != "" {
		isError := c.Query("is_error") == "true"
		filter.IsError = &isError
	}

	// Parse date filters
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err == nil {
			filter.EndDate = &endDate
		}
	}

	result, err := h.sdkService.ListUsage(c.Context(), authCtx.TenantID, filter)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(result)
}

// GetSDKUsageStats handles retrieving SDK usage statistics
func (h *SDKHandler) GetSDKUsageStats(c *fiber.Ctx) error {
	authCtx, err := GetAuthContext(c)
	if err != nil {
		return err
	}

	clientIDStr := c.Query("client_id")
	if clientIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "client_id is required",
			"code":  "MISSING_CLIENT_ID",
		})
	}

	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid client_id",
			"code":  "INVALID_CLIENT_ID",
		})
	}

	// Parse date range
	startDate := time.Now().AddDate(0, 0, -30) // Default: last 30 days
	endDate := time.Now()

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = parsed
		}
	}

	stats, err := h.sdkService.GetUsageStats(c.Context(), clientID, authCtx.TenantID, startDate, endDate)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(stats)
}
