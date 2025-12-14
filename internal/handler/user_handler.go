package handler

import (
	"strconv"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ============================================================================
// CRUD Operations
// ============================================================================

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user account. Requires user:write scope.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body dto.CreateUserRequest true "User creation data"
// @Success 201 {object} dto.UserDetailResponse "User created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Unauthorized - missing or invalid token"
// @Failure 403 {object} ErrorResponse "Forbidden - insufficient permissions"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req dto.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	user, err := h.userService.CreateUser(c.Context(), &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Data:    user,
	})
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get detailed user information by ID. Requires user:read scope.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID format)"
// @Success 200 {object} dto.UserDetailResponse "User details retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid user ID format"
// @Failure 401 {object} ErrorResponse "Unauthorized - missing or invalid token"
// @Failure 403 {object} ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	user, err := h.userService.GetUser(c.Context(), userID, authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Data:    user,
	})
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body dto.UpdateUserRequest true "Update data"
// @Success 200 {object} dto.UserDetailResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	user, err := h.userService.UpdateUser(c.Context(), userID, authCtx.UserID, &req)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Data:    user,
	})
}

// DeleteUser godoc
// @Summary Delete user
// @Description Soft delete a user account
// @Tags users
// @Param id path string true "User ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	if err := h.userService.DeleteUser(c.Context(), userID, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ============================================================================
// User Queries
// ============================================================================

// ListUsers godoc
// @Summary List users
// @Description Get a paginated list of users with filters
// @Tags users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param role query string false "Filter by role"
// @Param status query string false "Filter by status"
// @Param search query string false "Search query"
// @Success 200 {object} dto.UserListResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [get]
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	filter := &dto.UserFilter{
		Page:        getIntQuery(c, "page", 1),
		PageSize:    getIntQuery(c, "page_size", 20),
		SearchQuery: c.Query("search"),
	}

	// Parse tenant ID if provided
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		if tenantID, err := uuid.Parse(tenantIDStr); err == nil {
			filter.TenantID = &tenantID
		}
	}

	// Parse role filter
	if roleStr := c.Query("role"); roleStr != "" {
		filter.Roles = []models.UserRole{models.UserRole(roleStr)}
	}

	// Parse status filter
	if statusStr := c.Query("status"); statusStr != "" {
		filter.Statuses = []models.UserStatus{models.UserStatus(statusStr)}
	}

	// Parse boolean filters
	if emailVerified := c.Query("email_verified"); emailVerified != "" {
		val := emailVerified == "true"
		filter.EmailVerified = &val
	}

	if phoneVerified := c.Query("phone_verified"); phoneVerified != "" {
		val := phoneVerified == "true"
		filter.PhoneVerified = &val
	}

	if mfaEnabled := c.Query("mfa_enabled"); mfaEnabled != "" {
		val := mfaEnabled == "true"
		filter.MFAEnabled = &val
	}

	authCtx := middleware.MustGetAuthContext(c)
	users, err := h.userService.ListUsers(c.Context(), filter, authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Data:    users,
	})
}

// SearchUsers godoc
// @Summary Search users
// @Description Search for users by name or email
// @Tags users
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.UserListResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/search [get]
func (h *UserHandler) SearchUsers(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest, "MISSING_QUERY", "Search query is required", nil)
	}

	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 20)

	var tenantID *uuid.UUID
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		if tid, err := uuid.Parse(tenantIDStr); err == nil {
			tenantID = &tid
		}
	}

	users, err := h.userService.SearchUsers(c.Context(), query, tenantID, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Data:    users,
	})
}

// GetUsersByRole godoc
// @Summary Get users by role
// @Description Get all users with a specific role
// @Tags users
// @Produce json
// @Param role path string true "User role"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.UserListResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/by-role/{role} [get]
func (h *UserHandler) GetUsersByRole(c *fiber.Ctx) error {
	role := models.UserRole(c.Params("role"))
	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 20)

	var tenantID *uuid.UUID
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		if tid, err := uuid.Parse(tenantIDStr); err == nil {
			tenantID = &tid
		}
	}

	users, err := h.userService.GetUsersByRole(c.Context(), tenantID, role, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Data:    users,
	})
}

// GetActiveUsers godoc
// @Summary Get active users
// @Description Get all active users
// @Tags users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.UserListResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/active [get]
func (h *UserHandler) GetActiveUsers(c *fiber.Ctx) error {
	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 20)

	var tenantID *uuid.UUID
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		if tid, err := uuid.Parse(tenantIDStr); err == nil {
			tenantID = &tid
		}
	}

	users, err := h.userService.GetActiveUsers(c.Context(), tenantID, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Data:    users,
	})
}

// GetRecentlyActive godoc
// @Summary Get recently active users
// @Description Get users who were active within the specified hours
// @Tags users
// @Produce json
// @Param hours query int false "Hours" default(24)
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.UserListResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/recently-active [get]
func (h *UserHandler) GetRecentlyActive(c *fiber.Ctx) error {
	hours := getIntQuery(c, "hours", 24)
	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 20)

	var tenantID *uuid.UUID
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		if tid, err := uuid.Parse(tenantIDStr); err == nil {
			tenantID = &tid
		}
	}

	users, err := h.userService.GetRecentlyActive(c.Context(), tenantID, hours, page, pageSize)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Data:    users,
	})
}

// ============================================================================
// Authentication & Security
// ============================================================================

// UpdatePassword godoc
// @Summary Update password
// @Description Update user password
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param password body dto.UpdatePasswordRequest true "Password data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/password [put]
func (h *UserHandler) UpdatePassword(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	var req dto.UpdatePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.userService.UpdatePassword(c.Context(), userID, &req); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Password updated successfully",
	})
}

// ResetPassword godoc
// @Summary Reset password
// @Description Request password reset
// @Tags users
// @Accept json
// @Produce json
// @Param reset body dto.ResetPasswordRequest true "Reset data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/password-reset [post]
func (h *UserHandler) ResetPassword(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.userService.ResetPassword(c.Context(), &req); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Password reset email sent",
	})
}

// ConfirmPasswordReset godoc
// @Summary Confirm password reset
// @Description Confirm password reset with token
// @Tags users
// @Accept json
// @Produce json
// @Param confirm body dto.ConfirmPasswordResetRequest true "Confirmation data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/password-reset/confirm [post]
func (h *UserHandler) ConfirmPasswordReset(c *fiber.Ctx) error {
	var req dto.ConfirmPasswordResetRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.userService.ConfirmPasswordReset(c.Context(), &req); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Password reset successful",
	})
}

// VerifyEmail godoc
// @Summary Verify email
// @Description Mark user email as verified
// @Tags users
// @Param id path string true "User ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/verify-email [post]
func (h *UserHandler) VerifyEmail(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	if err := h.userService.VerifyEmail(c.Context(), userID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Email verified successfully",
	})
}

// VerifyPhone godoc
// @Summary Verify phone
// @Description Mark user phone as verified
// @Tags users
// @Param id path string true "User ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/verify-phone [post]
func (h *UserHandler) VerifyPhone(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	if err := h.userService.VerifyPhone(c.Context(), userID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Phone verified successfully",
	})
}

// UnlockUser godoc
// @Summary Unlock user
// @Description Unlock a locked user account
// @Tags users
// @Param id path string true "User ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/unlock [post]
func (h *UserHandler) UnlockUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	if err := h.userService.UnlockUser(c.Context(), userID, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User unlocked successfully",
	})
}

// GetLockedUsers godoc
// @Summary Get locked users
// @Description Get all locked user accounts
// @Tags users
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/locked [get]
func (h *UserHandler) GetLockedUsers(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)
	users, err := h.userService.GetLockedUsers(c.Context(), authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Data:    users,
	})
}

// ============================================================================
// MFA Management
// ============================================================================

// SetupMFA godoc
// @Summary Setup MFA
// @Description Generate MFA secret and QR code
// @Tags users
// @Param id path string true "User ID"
// @Success 200 {object} dto.MFASetupResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/mfa/setup [post]
func (h *UserHandler) SetupMFA(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	setup, err := h.userService.SetupMFA(c.Context(), userID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Data:    setup,
	})
}

// EnableMFA godoc
// @Summary Enable MFA
// @Description Enable multi-factor authentication
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param mfa body dto.EnableMFARequest true "MFA data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/mfa/enable [post]
func (h *UserHandler) EnableMFA(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	var req dto.EnableMFARequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.userService.EnableMFA(c.Context(), userID, &req); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "MFA enabled successfully",
	})
}

// DisableMFA godoc
// @Summary Disable MFA
// @Description Disable multi-factor authentication
// @Tags users
// @Param id path string true "User ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/mfa/disable [post]
func (h *UserHandler) DisableMFA(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	if err := h.userService.DisableMFA(c.Context(), userID, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "MFA disabled successfully",
	})
}

// VerifyMFA godoc
// @Summary Verify MFA
// @Description Verify MFA code
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param code body dto.VerifyMFARequest true "MFA code"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/mfa/verify [post]
func (h *UserHandler) VerifyMFA(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	var req dto.VerifyMFARequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.userService.VerifyMFA(c.Context(), userID, &req); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "MFA verified successfully",
	})
}

// ============================================================================
// Role & Status Management
// ============================================================================

// UpdateRole godoc
// @Summary Update user role
// @Description Update a user's role
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param role body RoleUpdateRequest true "Role data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/role [put]
func (h *UserHandler) UpdateRole(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	var req struct {
		Role models.UserRole `json:"role"`
	}
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	if err := h.userService.UpdateRole(c.Context(), userID, req.Role, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User role updated successfully",
	})
}

// UpdateStatus godoc
// @Summary Update user status
// @Description Update a user's status
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param status body StatusUpdateRequest true "Status data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/status [put]
func (h *UserHandler) UpdateStatus(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	var req struct {
		Status models.UserStatus `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	if err := h.userService.UpdateStatus(c.Context(), userID, req.Status, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User status updated successfully",
	})
}

// ActivateUser godoc
// @Summary Activate user
// @Description Activate a user account
// @Tags users
// @Param id path string true "User ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/activate [post]
func (h *UserHandler) ActivateUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	if err := h.userService.ActivateUser(c.Context(), userID, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User activated successfully",
	})
}

// DeactivateUser godoc
// @Summary Deactivate user
// @Description Deactivate a user account
// @Tags users
// @Param id path string true "User ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/deactivate [post]
func (h *UserHandler) DeactivateUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	if err := h.userService.DeactivateUser(c.Context(), userID, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User deactivated successfully",
	})
}

// SuspendUser godoc
// @Summary Suspend user
// @Description Suspend a user account
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param suspend body dto.SuspendUserRequest true "Suspension data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/suspend [post]
func (h *UserHandler) SuspendUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	var req dto.SuspendUserRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	if err := h.userService.SuspendUser(c.Context(), userID, &req, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User suspended successfully",
	})
}

// ============================================================================
// Profile Management
// ============================================================================

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update user profile information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param profile body map[string]interface{} true "Profile data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/profile [put]
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.userService.UpdateProfile(c.Context(), userID, updates); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Profile updated successfully",
	})
}

// UpdateAvatar godoc
// @Summary Update user avatar
// @Description Update user avatar URL
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param avatar body AvatarUpdateRequest true "Avatar data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/avatar [put]
func (h *UserHandler) UpdateAvatar(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	var req struct {
		AvatarURL string `json:"avatar_url"`
	}
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.userService.UpdateAvatar(c.Context(), userID, req.AvatarURL); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Avatar updated successfully",
	})
}

// UpdatePreferences godoc
// @Summary Update user preferences
// @Description Update user preferences (timezone, language, consent)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param preferences body dto.UpdatePreferencesRequest true "Preferences data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/preferences [put]
func (h *UserHandler) UpdatePreferences(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	var req dto.UpdatePreferencesRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.userService.UpdatePreferences(c.Context(), userID, &req); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Preferences updated successfully",
	})
}

// ============================================================================
// Compliance & GDPR
// ============================================================================

// AcceptTerms godoc
// @Summary Accept terms
// @Description Record terms of service acceptance
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param terms body TermsAcceptRequest true "Terms version"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/accept-terms [post]
func (h *UserHandler) AcceptTerms(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	var req struct {
		Version string `json:"version"`
	}
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.userService.AcceptTerms(c.Context(), userID, req.Version); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Terms accepted successfully",
	})
}

// AcceptPrivacyPolicy godoc
// @Summary Accept privacy policy
// @Description Record privacy policy acceptance
// @Tags users
// @Param id path string true "User ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/accept-privacy [post]
func (h *UserHandler) AcceptPrivacyPolicy(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	if err := h.userService.AcceptPrivacyPolicy(c.Context(), userID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Privacy policy accepted successfully",
	})
}

// UpdateConsent godoc
// @Summary Update consent
// @Description Update user consent preferences
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param consent body ConsentUpdateRequest true "Consent data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/consent [put]
func (h *UserHandler) UpdateConsent(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	var req struct {
		DataProcessing bool `json:"data_processing"`
		Marketing      bool `json:"marketing"`
	}
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.userService.UpdateConsent(c.Context(), userID, req.DataProcessing, req.Marketing); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Consent updated successfully",
	})
}

// MarkForDeletion godoc
// @Summary Mark user for deletion
// @Description Schedule user account for deletion
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param deletion body DeletionRequest true "Deletion schedule"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/mark-for-deletion [post]
func (h *UserHandler) MarkForDeletion(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	var req struct {
		ScheduledDate time.Time `json:"scheduled_date"`
	}
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
	}

	if err := h.userService.MarkForDeletion(c.Context(), userID, req.ScheduledDate); err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User marked for deletion successfully",
	})
}

// GetUsersMarkedForDeletion godoc
// @Summary Get users marked for deletion
// @Description Get all users scheduled for deletion
// @Tags users
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/marked-for-deletion [get]
func (h *UserHandler) GetUsersMarkedForDeletion(c *fiber.Ctx) error {
	authCtx := middleware.MustGetAuthContext(c)
	users, err := h.userService.GetUsersMarkedForDeletion(c.Context(), authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Data:    users,
	})
}

// PermanentlyDeleteUser godoc
// @Summary Permanently delete user
// @Description Permanently delete a user account (irreversible)
// @Tags users
// @Param id path string true "User ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/permanent [delete]
func (h *UserHandler) PermanentlyDeleteUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID", err)
	}

	authCtx := middleware.MustGetAuthContext(c)
	if err := h.userService.PermanentlyDeleteUser(c.Context(), userID, authCtx.UserID); err != nil {
		return HandleServiceError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ============================================================================
// Analytics
// ============================================================================

// GetUserStats godoc
// @Summary Get user statistics
// @Description Get comprehensive user statistics
// @Tags users
// @Produce json
// @Param tenant_id query string false "Tenant ID"
// @Success 200 {object} dto.UserStatsResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/stats [get]
func (h *UserHandler) GetUserStats(c *fiber.Ctx) error {
	var tenantID *uuid.UUID
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		if tid, err := uuid.Parse(tenantIDStr); err == nil {
			tenantID = &tid
		}
	}

	authCtx := middleware.MustGetAuthContext(c)
	stats, err := h.userService.GetUserStats(c.Context(), tenantID, authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Data:    stats,
	})
}

// GetRegistrationStats godoc
// @Summary Get registration statistics
// @Description Get user registration statistics over time
// @Tags users
// @Produce json
// @Param start_date query string true "Start date (RFC3339)"
// @Param end_date query string true "End date (RFC3339)"
// @Success 200 {object} dto.RegistrationStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/registration-stats [get]
func (h *UserHandler) GetRegistrationStats(c *fiber.Ctx) error {
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

	authCtx := middleware.MustGetAuthContext(c)
	stats, err := h.userService.GetRegistrationStats(c.Context(), startDate, endDate, authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Data:    stats,
	})
}

// GetUserGrowth godoc
// @Summary Get user growth
// @Description Get user growth data over time
// @Tags users
// @Produce json
// @Param tenant_id query string false "Tenant ID"
// @Param months query int false "Number of months" default(12)
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/growth [get]
func (h *UserHandler) GetUserGrowth(c *fiber.Ctx) error {
	var tenantID *uuid.UUID
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		if tid, err := uuid.Parse(tenantIDStr); err == nil {
			tenantID = &tid
		}
	}

	months := getIntQuery(c, "months", 12)

	authCtx := middleware.MustGetAuthContext(c)
	growth, err := h.userService.GetUserGrowth(c.Context(), tenantID, months, authCtx.UserID)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Data:    growth,
	})
}

// ============================================================================
// Helper Functions
// ============================================================================

// getIntQuery gets an integer query parameter with a default value
func getIntQuery(c *fiber.Ctx, key string, defaultValue int) int {
	if val := c.Query(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// Request types for swagger documentation
type RoleUpdateRequest struct {
	Role models.UserRole `json:"role"`
}

type StatusUpdateRequest struct {
	Status models.UserStatus `json:"status"`
}

type AvatarUpdateRequest struct {
	AvatarURL string `json:"avatar_url"`
}

type TermsAcceptRequest struct {
	Version string `json:"version"`
}

type ConsentUpdateRequest struct {
	DataProcessing bool `json:"data_processing"`
	Marketing      bool `json:"marketing"`
}

type DeletionRequest struct {
	ScheduledDate time.Time `json:"scheduled_date"`
}
