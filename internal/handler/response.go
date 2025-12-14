package handler

import (
	"errors"
	"time"

	pkgErrors "Krafti_Vibe/internal/pkg/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents validation errors
type ValidationErrorResponse struct {
	Success   bool              `json:"success"`
	Error     string            `json:"error"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Errors    []ValidationError `json:"errors,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(c *fiber.Ctx, status int, code, message string, err error) error {
	response := ErrorResponse{
		Success: false,
		Error:   message,
		Code:    code,
		Message: message,
	}

	// Add request ID if available
	if reqID := c.Locals("request_id"); reqID != nil {
		if id, ok := reqID.(string); ok {
			response.RequestID = id
		}
	}

	// Log error if provided
	if err != nil {
		// You could add logging here
		response.Message = err.Error()
	}

	return c.Status(status).JSON(response)
}

// HandleServiceError handles errors from the service layer
func HandleServiceError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	// Get request ID
	var requestID string
	if reqID := c.Locals("request_id"); reqID != nil {
		if id, ok := reqID.(string); ok {
			requestID = id
		}
	}

	// Handle AppError type
	var appErr *pkgErrors.AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.HTTPStatus).JSON(ErrorResponse{
			Success:   false,
			Error:     appErr.Message,
			Code:      string(appErr.Code),
			Message:   appErr.Error(),
			RequestID: requestID,
		})
	}

	// Generic error
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Success:   false,
		Error:     "Internal Server Error",
		Code:      "INTERNAL_ERROR",
		Message:   err.Error(),
		RequestID: requestID,
	})
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(c *fiber.Ctx, data any, message ...string) error {
	response := SuccessResponse{
		Success: true,
		Data:    data,
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	return c.JSON(response)
}

// NewCreatedResponse creates a new created response (201)
func NewCreatedResponse(c *fiber.Ctx, data any, message ...string) error {
	response := SuccessResponse{
		Success: true,
		Data:    data,
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// NewNoContentResponse creates a new no content response (204)
func NewNoContentResponse(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message,omitempty"`
	Data       any         `json:"data"`
	Pagination *Pagination `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page         int   `json:"page"`
	PageSize     int   `json:"page_size"`
	TotalItems   int64 `json:"total_items"`
	TotalPages   int   `json:"total_pages"`
	HasNext      bool  `json:"has_next"`
	HasPrevious  bool  `json:"has_previous"`
	NextPage     *int  `json:"next_page,omitempty"`
	PreviousPage *int  `json:"previous_page,omitempty"`
}

// NewPaginatedResponse creates a paginated response
func NewPaginatedResponse(c *fiber.Ctx, data any, pagination *Pagination, message ...string) error {
	response := PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	return c.JSON(response)
}

// NewValidationErrorResponse creates a validation error response
func NewValidationErrorResponse(c *fiber.Ctx, validationErrors []ValidationError) error {
	requestID := getRequestID(c)

	response := ValidationErrorResponse{
		Success:   false,
		Error:     "Validation failed",
		Code:      "VALIDATION_ERROR",
		Message:   "One or more fields failed validation",
		Errors:    validationErrors,
		RequestID: requestID,
	}

	return c.Status(fiber.StatusBadRequest).JSON(response)
}

// NewUnauthorizedResponse creates an unauthorized error response
func NewUnauthorizedResponse(c *fiber.Ctx, message string) error {
	requestID := getRequestID(c)

	response := ErrorResponse{
		Success:   false,
		Error:     "Unauthorized",
		Code:      "UNAUTHORIZED",
		Message:   message,
		RequestID: requestID,
	}

	return c.Status(fiber.StatusUnauthorized).JSON(response)
}

// NewForbiddenResponse creates a forbidden error response
func NewForbiddenResponse(c *fiber.Ctx, message string) error {
	requestID := getRequestID(c)

	response := ErrorResponse{
		Success:   false,
		Error:     "Forbidden",
		Code:      "FORBIDDEN",
		Message:   message,
		RequestID: requestID,
	}

	return c.Status(fiber.StatusForbidden).JSON(response)
}

// NewNotFoundResponse creates a not found error response
func NewNotFoundResponse(c *fiber.Ctx, resource string) error {
	requestID := getRequestID(c)

	response := ErrorResponse{
		Success:   false,
		Error:     "Not Found",
		Code:      "NOT_FOUND",
		Message:   resource + " not found",
		RequestID: requestID,
	}

	return c.Status(fiber.StatusNotFound).JSON(response)
}

// NewConflictResponse creates a conflict error response
func NewConflictResponse(c *fiber.Ctx, message string) error {
	requestID := getRequestID(c)

	response := ErrorResponse{
		Success:   false,
		Error:     "Conflict",
		Code:      "CONFLICT",
		Message:   message,
		RequestID: requestID,
	}

	return c.Status(fiber.StatusConflict).JSON(response)
}

// NewRateLimitResponse creates a rate limit error response
func NewRateLimitResponse(c *fiber.Ctx) error {
	requestID := getRequestID(c)

	response := ErrorResponse{
		Success:   false,
		Error:     "Too Many Requests",
		Code:      "RATE_LIMIT_EXCEEDED",
		Message:   "Rate limit exceeded. Please try again later.",
		RequestID: requestID,
	}

	// Add Retry-After header
	c.Set("Retry-After", "60")

	return c.Status(fiber.StatusTooManyRequests).JSON(response)
}

// NewAcceptedResponse creates an accepted response (202)
func NewAcceptedResponse(c *fiber.Ctx, message string, location ...string) error {
	response := SuccessResponse{
		Success: true,
		Message: message,
	}

	// Add Location header if provided
	if len(location) > 0 {
		c.Set("Location", location[0])
	}

	return c.Status(fiber.StatusAccepted).JSON(response)
}

// getRequestID retrieves the request ID from context
func getRequestID(c *fiber.Ctx) string {
	// Try to get from header first
	if reqID := c.Get("X-Request-ID"); reqID != "" {
		return reqID
	}

	// Try to get from locals
	if reqID := c.Locals("request_id"); reqID != nil {
		if id, ok := reqID.(string); ok {
			return id
		}
	}

	return ""
}

// LogHandlerError logs handler errors with context
func LogHandlerError(c *fiber.Ctx, operation string, err error) {
	log.Errorw("Handler error",
		"operation", operation,
		"error", err.Error(),
		"path", c.Path(),
		"method", c.Method(),
		"request_id", getRequestID(c),
		"timestamp", time.Now().UTC(),
	)
}

// LogHandlerInfo logs handler info with context
func LogHandlerInfo(c *fiber.Ctx, operation string, details map[string]interface{}) {
	fields := []interface{}{
		"operation", operation,
		"path", c.Path(),
		"method", c.Method(),
		"request_id", getRequestID(c),
		"timestamp", time.Now().UTC(),
	}

	for k, v := range details {
		fields = append(fields, k, v)
	}

	log.Infow("Handler operation", fields...)
}
