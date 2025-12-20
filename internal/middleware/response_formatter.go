package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// StandardResponse represents the standard API response format
type StandardResponse struct {
	Success   bool       `json:"success"`
	Data      any        `json:"data,omitempty"`
	Error     *ErrorInfo `json:"error,omitempty"`
	Meta      *MetaInfo  `json:"meta,omitempty"`
	RequestID string     `json:"request_id,omitempty"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// MetaInfo represents pagination and other metadata
type MetaInfo struct {
	Page       int   `json:"page,omitempty"`
	PageSize   int   `json:"page_size,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
	TotalCount int64 `json:"total_count,omitempty"`
}

// ResponseFormatterConfig holds configuration for response formatter
type ResponseFormatterConfig struct {
	Logger           *zap.Logger
	IncludeRequestID bool
}

// DefaultResponseFormatterConfig returns default response formatter configuration
func DefaultResponseFormatterConfig(logger *zap.Logger) ResponseFormatterConfig {
	return ResponseFormatterConfig{
		Logger:           logger,
		IncludeRequestID: true,
	}
}

// ResponseFormatter provides helper methods for sending standardized responses
type ResponseFormatter struct {
	c      *fiber.Ctx
	config ResponseFormatterConfig
}

// NewResponseFormatter creates a new response formatter
func NewResponseFormatter(c *fiber.Ctx, config ResponseFormatterConfig) *ResponseFormatter {
	return &ResponseFormatter{
		c:      c,
		config: config,
	}
}

// Success sends a successful response
func (rf *ResponseFormatter) Success(data any, meta *MetaInfo) error {
	response := StandardResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}

	if rf.config.IncludeRequestID {
		response.RequestID = rf.c.Get("X-Request-ID")
	}

	return rf.c.JSON(response)
}

// Created sends a 201 Created response
func (rf *ResponseFormatter) Created(data any) error {
	response := StandardResponse{
		Success: true,
		Data:    data,
	}

	if rf.config.IncludeRequestID {
		response.RequestID = rf.c.Get("X-Request-ID")
	}

	return rf.c.Status(fiber.StatusCreated).JSON(response)
}

// NoContent sends a 204 No Content response
func (rf *ResponseFormatter) NoContent() error {
	return rf.c.SendStatus(fiber.StatusNoContent)
}

// Error sends an error response
func (rf *ResponseFormatter) Error(statusCode int, code string, message string, details any) error {
	response := StandardResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	if rf.config.IncludeRequestID {
		response.RequestID = rf.c.Get("X-Request-ID")
	}

	return rf.c.Status(statusCode).JSON(response)
}

// BadRequest sends a 400 Bad Request response
func (rf *ResponseFormatter) BadRequest(message string, details any) error {
	return rf.Error(fiber.StatusBadRequest, "BAD_REQUEST", message, details)
}

// Unauthorized sends a 401 Unauthorized response
func (rf *ResponseFormatter) Unauthorized(message string) error {
	return rf.Error(fiber.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

// Forbidden sends a 403 Forbidden response
func (rf *ResponseFormatter) Forbidden(message string) error {
	return rf.Error(fiber.StatusForbidden, "FORBIDDEN", message, nil)
}

// NotFound sends a 404 Not Found response
func (rf *ResponseFormatter) NotFound(message string) error {
	return rf.Error(fiber.StatusNotFound, "NOT_FOUND", message, nil)
}

// Conflict sends a 409 Conflict response
func (rf *ResponseFormatter) Conflict(message string, details any) error {
	return rf.Error(fiber.StatusConflict, "CONFLICT", message, details)
}

// UnprocessableEntity sends a 422 Unprocessable Entity response
func (rf *ResponseFormatter) UnprocessableEntity(message string, details any) error {
	return rf.Error(fiber.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", message, details)
}

// TooManyRequests sends a 429 Too Many Requests response
func (rf *ResponseFormatter) TooManyRequests(message string) error {
	return rf.Error(fiber.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", message, nil)
}

// InternalServerError sends a 500 Internal Server Error response
func (rf *ResponseFormatter) InternalServerError(message string) error {
	return rf.Error(fiber.StatusInternalServerError, "INTERNAL_ERROR", message, nil)
}

// ServiceUnavailable sends a 503 Service Unavailable response
func (rf *ResponseFormatter) ServiceUnavailable(message string) error {
	return rf.Error(fiber.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message, nil)
}

// Paginated sends a paginated success response
func (rf *ResponseFormatter) Paginated(data any, page, pageSize int, totalCount int64) error {
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	meta := &MetaInfo{
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		TotalCount: totalCount,
	}

	return rf.Success(data, meta)
}

// RespondJSON is a helper function to easily send JSON responses
func RespondJSON(c *fiber.Ctx, logger *zap.Logger) *ResponseFormatter {
	config := DefaultResponseFormatterConfig(logger)
	return NewResponseFormatter(c, config)
}
