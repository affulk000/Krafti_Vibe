package handler

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// LogtoWebhookHandler handles Logto webhook events
type LogtoWebhookHandler struct {
	userService service.UserService
}

// NewLogtoWebhookHandler creates a new Logto webhook handler
func NewLogtoWebhookHandler(userService service.UserService) *LogtoWebhookHandler {
	return &LogtoWebhookHandler{
		userService: userService,
	}
}

// LogtoWebhookEvent represents a webhook event from Logto
type LogtoWebhookEvent struct {
	Event      string                 `json:"event"`      // e.g., "User.Created", "User.Data.Updated"
	CreatedAt  string                 `json:"createdAt"`  // ISO 8601 timestamp
	SessionID  string                 `json:"sessionId"`  // Optional session identifier
	UserAgent  string                 `json:"userAgent"`  // User agent of the request
	IP         string                 `json:"ip"`         // IP address
	InteractionEvent bool              `json:"interactionEvent"` // Whether this is an interaction event
	Data       map[string]interface{} `json:"data"`       // Event-specific data
}

// LogtoUserData represents user data from Logto webhooks
type LogtoUserData struct {
	ID                  string                 `json:"id"`
	Username            *string                `json:"username"`
	PrimaryEmail        *string                `json:"primaryEmail"`
	PrimaryPhone        *string                `json:"primaryPhone"`
	Name                *string                `json:"name"`
	Avatar              *string                `json:"avatar"`
	CustomData          map[string]interface{} `json:"customData"`
	Profile             map[string]interface{} `json:"profile"`
	ApplicationID       *string                `json:"applicationId"`
	IsSuspended         bool                   `json:"isSuspended"`
	LastSignInAt        *int64                 `json:"lastSignInAt"`
	CreatedAt           int64                  `json:"createdAt"`
	UpdatedAt           int64                  `json:"updatedAt"`
	IsEmailVerified     bool                   `json:"isEmailVerified,omitempty"`
	IsPhoneVerified     bool                   `json:"isPhoneVerified,omitempty"`
}

// HandleWebhook godoc
// @Summary Handle Logto webhook events
// @Description Receive and process webhook events from Logto (user creation, updates, deletion)
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param payload body LogtoWebhookEvent true "Webhook event payload"
// @Success 200 {object} SuccessResponse "Event processed successfully"
// @Failure 400 {object} ErrorResponse "Invalid payload"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /webhooks/logto [post]
func (h *LogtoWebhookHandler) HandleWebhook(c *fiber.Ctx) error {
	var event LogtoWebhookEvent
	if err := c.BodyParser(&event); err != nil {
		log.Errorf("Failed to parse webhook payload: %v", err)
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_PAYLOAD", "Invalid webhook payload", err)
	}

	// Log the event for debugging
	log.Infof("Received Logto webhook event: %s at %s", event.Event, event.CreatedAt)

	// Process the event based on type
	ctx := context.Background()

	switch event.Event {
	case "User.Created":
		return h.handleUserCreated(ctx, c, event)
	case "User.Data.Updated":
		return h.handleUserUpdated(ctx, c, event)
	case "User.Deleted":
		return h.handleUserDeleted(ctx, c, event)
	case "User.SuspensionStatus.Updated":
		return h.handleUserSuspensionUpdated(ctx, c, event)
	default:
		// Log unhandled events but return success to avoid retries
		log.Warnf("Unhandled Logto webhook event type: %s", event.Event)
		return c.JSON(SuccessResponse{
			Success: true,
			Message: "Event acknowledged but not processed",
		})
	}
}

// handleUserCreated processes User.Created events
func (h *LogtoWebhookHandler) handleUserCreated(ctx context.Context, c *fiber.Ctx, event LogtoWebhookEvent) error {
	userData, err := h.parseUserData(event.Data)
	if err != nil {
		log.Errorf("Failed to parse user data: %v", err)
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_USER_DATA", "Invalid user data in webhook", err)
	}

	log.Infof("Processing User.Created event for user: %s (email: %s)", userData.ID, getStringValue(userData.PrimaryEmail))

	// Check if user already exists
	existingUser, err := h.userService.GetUserByLogtoID(ctx, userData.ID)
	if err == nil && existingUser != nil {
		log.Infof("User already exists: %s", userData.ID)
		return c.JSON(SuccessResponse{
			Success: true,
			Message: "User already exists",
		})
	}

	// Determine user role based on custom data or default to customer
	role := h.determineUserRole(userData)

	// Create user in database
	email := getStringValue(userData.PrimaryEmail)
	if email == "" {
		// If no email, we can't create the user
		log.Warnf("Cannot create user without email: %s", userData.ID)
		return NewErrorResponse(c, fiber.StatusBadRequest, "MISSING_EMAIL", "User email is required", nil)
	}

	// Parse name into first and last name
	firstName, lastName := h.parseName(getStringValue(userData.Name))

	// Map Logto user to our user model
	isPlatformUser := h.isPlatformUser(userData)
	user := &models.User{
		LogtoUserID:    userData.ID,
		Email:          email,
		FirstName:      firstName,
		LastName:       lastName,
		AvatarURL:      getStringValueDefault(userData.Avatar),
		Role:           role,
		Status:         h.getUserStatus(userData, isPlatformUser),
		EmailVerified:  userData.IsEmailVerified,
		PhoneNumber:    getStringValueDefault(userData.PrimaryPhone),
		PhoneVerified:  userData.IsPhoneVerified,
		IsPlatformUser: isPlatformUser,
		LastLoginAt:    h.getLastLoginTime(userData),
	}
	// Set the ID manually (BaseModel field)
	user.ID = uuid.New()

	// Store custom data as JSONB metadata
	if len(userData.CustomData) > 0 {
		user.Metadata = models.JSONB(userData.CustomData)
	}

	log.Infof("Creating user in database: %s (email: %s, role: %s)", user.ID, user.Email, user.Role)

	// Create user via service
	if err := h.userService.CreateUserFromWebhook(ctx, user); err != nil {
		log.Errorf("Failed to create user from webhook: %v", err)
		return NewErrorResponse(c, fiber.StatusInternalServerError, "CREATE_FAILED", "Failed to create user", err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User created successfully",
		Data: fiber.Map{
			"logto_user_id": userData.ID,
			"email":         email,
			"role":          role,
		},
	})
}

// handleUserUpdated processes User.Data.Updated events
func (h *LogtoWebhookHandler) handleUserUpdated(ctx context.Context, c *fiber.Ctx, event LogtoWebhookEvent) error {
	userData, err := h.parseUserData(event.Data)
	if err != nil {
		log.Errorf("Failed to parse user data: %v", err)
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_USER_DATA", "Invalid user data in webhook", err)
	}

	log.Infof("Processing User.Data.Updated event for user: %s", userData.ID)

	// Get existing user
	existingUser, err := h.userService.GetUserByLogtoID(ctx, userData.ID)
	if err != nil {
		log.Errorf("User not found for update: %s", userData.ID)
		return NewErrorResponse(c, fiber.StatusNotFound, "USER_NOT_FOUND", "User not found", err)
	}

	log.Infof("Updating user: %s (current email: %s)", existingUser.ID, existingUser.Email)

	// Build updates map
	updates := make(map[string]interface{})

	// Update user fields
	if userData.PrimaryEmail != nil && *userData.PrimaryEmail != "" {
		updates["email"] = *userData.PrimaryEmail
	}
	if userData.Name != nil && *userData.Name != "" {
		firstName, lastName := h.parseName(*userData.Name)
		updates["first_name"] = firstName
		updates["last_name"] = lastName
	}
	if userData.Avatar != nil {
		updates["avatar_url"] = getStringValueDefault(userData.Avatar)
	}
	if userData.PrimaryPhone != nil {
		updates["phone_number"] = getStringValueDefault(userData.PrimaryPhone)
	}

	updates["email_verified"] = userData.IsEmailVerified
	updates["phone_verified"] = userData.IsPhoneVerified
	// Keep existing platform user status when updating
	updates["status"] = h.getUserStatus(userData, existingUser.IsPlatformUser)
	updates["updated_at"] = time.Now()

	// Update custom data
	if len(userData.CustomData) > 0 {
		customDataJSON, _ := json.Marshal(userData.CustomData)
		updates["metadata"] = customDataJSON
	}

	// Update user via service
	if err := h.userService.UpdateUserFromWebhook(ctx, existingUser.ID, updates); err != nil {
		log.Errorf("Failed to update user from webhook: %v", err)
		return NewErrorResponse(c, fiber.StatusInternalServerError, "UPDATE_FAILED", "Failed to update user", err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User updated successfully",
		Data: fiber.Map{
			"logto_user_id": userData.ID,
			"user_id":       existingUser.ID,
		},
	})
}

// handleUserDeleted processes User.Deleted events
func (h *LogtoWebhookHandler) handleUserDeleted(ctx context.Context, c *fiber.Ctx, event LogtoWebhookEvent) error {
	// Extract user ID from event data
	userID, ok := event.Data["id"].(string)
	if !ok {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DATA", "Missing user ID in webhook", nil)
	}

	log.Infof("Processing User.Deleted event for user: %s", userID)

	// Get user
	existingUser, err := h.userService.GetUserByLogtoID(ctx, userID)
	if err != nil {
		log.Warnf("User not found for deletion: %s (may already be deleted)", userID)
		return c.JSON(SuccessResponse{
			Success: true,
			Message: "User already deleted",
		})
	}

	// Mark user as deleted via service
	if err := h.userService.DeleteUserFromWebhook(ctx, existingUser.ID); err != nil {
		log.Errorf("Failed to delete user from webhook: %v", err)
		return NewErrorResponse(c, fiber.StatusInternalServerError, "DELETE_FAILED", "Failed to delete user", err)
	}

	log.Infof("Marked user as deleted: %s", existingUser.ID)

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User deleted successfully",
		Data: fiber.Map{
			"logto_user_id": userID,
			"user_id":       existingUser.ID,
		},
	})
}

// handleUserSuspensionUpdated processes User.SuspensionStatus.Updated events
func (h *LogtoWebhookHandler) handleUserSuspensionUpdated(ctx context.Context, c *fiber.Ctx, event LogtoWebhookEvent) error {
	userID, ok := event.Data["id"].(string)
	if !ok {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DATA", "Missing user ID in webhook", nil)
	}

	isSuspended, ok := event.Data["isSuspended"].(bool)
	if !ok {
		return NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_DATA", "Missing suspension status in webhook", nil)
	}

	log.Infof("Processing User.SuspensionStatus.Updated event for user: %s (suspended: %v)", userID, isSuspended)

	// Get user
	existingUser, err := h.userService.GetUserByLogtoID(ctx, userID)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusNotFound, "USER_NOT_FOUND", "User not found", err)
	}

	// Update suspension status
	var newStatus models.UserStatus
	if isSuspended {
		newStatus = models.UserStatusSuspended
	} else {
		newStatus = models.UserStatusActive
	}

	// Update via service
	if err := h.userService.UpdateUserStatusFromWebhook(ctx, existingUser.ID, newStatus); err != nil {
		log.Errorf("Failed to update user status from webhook: %v", err)
		return NewErrorResponse(c, fiber.StatusInternalServerError, "UPDATE_FAILED", "Failed to update user status", err)
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User suspension status updated",
		Data: fiber.Map{
			"logto_user_id": userID,
			"user_id":       existingUser.ID,
			"is_suspended":  isSuspended,
		},
	})
}

// Helper functions

func (h *LogtoWebhookHandler) parseUserData(data map[string]interface{}) (*LogtoUserData, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var userData LogtoUserData
	if err := json.Unmarshal(jsonData, &userData); err != nil {
		return nil, err
	}

	return &userData, nil
}

func (h *LogtoWebhookHandler) determineUserRole(userData *LogtoUserData) models.UserRole {
	// Check custom data for role
	if userData.CustomData != nil {
		if role, ok := userData.CustomData["role"].(string); ok {
			return models.UserRole(role)
		}
	}

	// Check profile for role
	if userData.Profile != nil {
		if role, ok := userData.Profile["role"].(string); ok {
			return models.UserRole(role)
		}
	}

	// Default to customer role
	return models.UserRoleCustomer
}

func (h *LogtoWebhookHandler) getUserStatus(userData *LogtoUserData, isPlatformUser bool) models.UserStatus {
	if userData.IsSuspended {
		return models.UserStatusSuspended
	}
	// Non-platform users start as pending (need to complete onboarding and get tenant)
	// Platform users can be active immediately
	if !isPlatformUser {
		return models.UserStatusPending
	}
	return models.UserStatusActive
}

func (h *LogtoWebhookHandler) isPlatformUser(userData *LogtoUserData) bool {
	// Check custom data for platform user flag
	if userData.CustomData != nil {
		if isPlatform, ok := userData.CustomData["is_platform_user"].(bool); ok {
			return isPlatform
		}
	}
	return false
}

func (h *LogtoWebhookHandler) getLastLoginTime(userData *LogtoUserData) *time.Time {
	if userData.LastSignInAt != nil && *userData.LastSignInAt > 0 {
		t := time.Unix(*userData.LastSignInAt/1000, 0)
		return &t
	}
	return nil
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getStringValueDefault(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// parseName splits a full name into first and last name
func (h *LogtoWebhookHandler) parseName(fullName string) (string, string) {
	if fullName == "" {
		return "", ""
	}

	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return "", ""
	}

	if len(parts) == 1 {
		return parts[0], ""
	}

	// First part is first name, rest is last name
	firstName := parts[0]
	lastName := strings.Join(parts[1:], " ")
	return firstName, lastName
}
