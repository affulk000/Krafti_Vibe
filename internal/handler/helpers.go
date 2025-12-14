package handler

import (
	"Krafti_Vibe/internal/middleware"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Enterprise-grade helper functions for handlers
// Note: getIntQuery is defined in user_handler.go to avoid duplication

// getBoolQuery retrieves a boolean query parameter with a default value
func getBoolQuery(c *fiber.Ctx, key string, defaultValue bool) bool {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

// getFloatQuery retrieves a float64 query parameter with a default value
func getFloatQuery(c *fiber.Ctx, key string, defaultValue float64) float64 {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}

	return floatValue
}

// ValidatePagination validates and normalizes pagination parameters
func ValidatePagination(page, pageSize int) (int, int) {
	// Ensure page is at least 1
	if page < 1 {
		page = 1
	}

	// Ensure pageSize is within acceptable range
	if pageSize < 1 {
		pageSize = 20 // default
	} else if pageSize > 100 {
		pageSize = 100 // max
	}

	return page, pageSize
}

// ParseUUIDParam parses UUID from path parameter
func ParseUUIDParam(c *fiber.Ctx, paramName string) (uuid.UUID, error) {
	idStr := c.Params(paramName)
	if idStr == "" {
		return uuid.Nil, NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_PARAMETER", "Missing "+paramName, nil)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid "+paramName+" format", err)
	}

	return id, nil
}

// ParseUUIDQuery parses UUID from query parameter
func ParseUUIDQuery(c *fiber.Ctx, paramName string) (*uuid.UUID, error) {
	idStr := c.Query(paramName)
	if idStr == "" {
		return nil, nil // Optional parameter
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, NewErrorResponse(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid "+paramName+" format", err)
	}

	return &id, nil
}

// GetAuthContext retrieves and validates authentication context
func GetAuthContext(c *fiber.Ctx) (*middleware.AuthContext, error) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		return nil, NewUnauthorizedResponse(c, "Authentication required")
	}

	// Validate tenant ID
	if authCtx.TenantID == uuid.Nil {
		return nil, NewForbiddenResponse(c, "Tenant context required")
	}

	return authCtx, nil
}

// MustGetAuthContext retrieves authentication context or panics
// This should only be used after authentication middleware has run
func MustGetAuthContext(c *fiber.Ctx) *middleware.AuthContext {
	authCtx := middleware.MustGetAuthContext(c)
	return authCtx
}

// ValidateRequiredFields checks if required fields are present in a map
func ValidateRequiredFields(data map[string]interface{}, requiredFields []string) []ValidationError {
	var errors []ValidationError

	for _, field := range requiredFields {
		value, exists := data[field]
		if !exists || value == nil || value == "" {
			errors = append(errors, ValidationError{
				Field:   field,
				Message: "This field is required",
			})
		}
	}

	return errors
}

// SanitizeString sanitizes user input by trimming whitespace and removing potentially dangerous characters
func SanitizeString(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	return input
}

// BuildPagination creates pagination metadata
func BuildPagination(page, pageSize int, totalItems int64) *Pagination {
	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	pagination := &Pagination{
		Page:        page,
		PageSize:    pageSize,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}

	if pagination.HasNext {
		nextPage := page + 1
		pagination.NextPage = &nextPage
	}

	if pagination.HasPrevious {
		prevPage := page - 1
		pagination.PreviousPage = &prevPage
	}

	return pagination
}

// ValidateEmailFormat validates email format (basic check)
func ValidateEmailFormat(email string) bool {
	if email == "" {
		return false
	}

	// Basic email validation
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}

	if !strings.Contains(parts[1], ".") {
		return false
	}

	return true
}

// SetCacheHeaders sets cache control headers for responses
func SetCacheHeaders(c *fiber.Ctx, maxAge int, isPrivate bool) {
	if isPrivate {
		c.Set("Cache-Control", "private, max-age="+strconv.Itoa(maxAge))
	} else {
		c.Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
	}
}

// SetNoCacheHeaders sets no-cache headers for responses
func SetNoCacheHeaders(c *fiber.Ctx) {
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")
}

// ValidateContentType validates the Content-Type header
func ValidateContentType(c *fiber.Ctx, expectedType string) error {
	contentType := c.Get("Content-Type")
	if !strings.Contains(contentType, expectedType) {
		return NewErrorResponse(c, fiber.StatusUnsupportedMediaType, "INVALID_CONTENT_TYPE",
			"Expected Content-Type: "+expectedType, nil)
	}
	return nil
}

// ExtractSortParams extracts and validates sort parameters
func ExtractSortParams(c *fiber.Ctx, allowedFields []string) (string, string) {
	sortBy := c.Query("sort_by", "created_at")
	sortOrder := strings.ToLower(c.Query("sort_order", "desc"))

	// Validate sortBy is in allowed fields
	isValid := false
	for _, field := range allowedFields {
		if sortBy == field {
			isValid = true
			break
		}
	}
	if !isValid {
		sortBy = "created_at" // default
	}

	// Validate sortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc" // default
	}

	return sortBy, sortOrder
}

// CheckTenantAccess verifies the authenticated user has access to the specified tenant
func CheckTenantAccess(c *fiber.Ctx, targetTenantID uuid.UUID) error {
	authCtx := MustGetAuthContext(c)

	// Check if the user's tenant matches the target tenant
	if authCtx.TenantID != targetTenantID {
		LogHandlerError(c, "tenant_access_check",
			fiber.NewError(fiber.StatusForbidden, "Access denied: tenant mismatch"))
		return NewForbiddenResponse(c, "You don't have access to this resource")
	}

	return nil
}

// ValidateIDempotencyKey validates and retrieves idempotency key for write operations
func ValidateIdempotencyKey(c *fiber.Ctx) (string, bool) {
	key := c.Get("Idempotency-Key")
	if key == "" {
		return "", false
	}

	// Validate key format (should be a UUID or similar)
	if len(key) < 16 || len(key) > 128 {
		return "", false
	}

	return key, true
}

// SetSecurityHeaders sets standard security headers
func SetSecurityHeaders(c *fiber.Ctx) {
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("X-Frame-Options", "DENY")
	c.Set("X-XSS-Protection", "1; mode=block")
	c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
}
