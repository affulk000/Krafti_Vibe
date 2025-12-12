package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents a standardized error code
type ErrorCode string

const (
	// General errors
	ErrCodeInternal      ErrorCode = "INTERNAL_ERROR"
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeInvalidInput  ErrorCode = "INVALID_INPUT"
	ErrCodeUnauthorized  ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden     ErrorCode = "FORBIDDEN"
	ErrCodeConflict      ErrorCode = "CONFLICT"
	ErrCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	// Repository errors
	ErrCodeDuplicate     ErrorCode = "DUPLICATE"
	ErrCodeDatabaseError ErrorCode = "DATABASE_ERROR"
	ErrCodeCacheError    ErrorCode = "CACHE_ERROR"

	// Business logic errors
	ErrCodeValidation ErrorCode = "VALIDATION_ERROR"
	ErrCodeBusiness   ErrorCode = "BUSINESS_ERROR"
)

// Standard errors
var (
	ErrNotFound      = errors.New("record not found")
	ErrDuplicate     = errors.New("duplicate record")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized access")
	ErrForbidden     = errors.New("forbidden access")
	ErrConflict      = errors.New("data conflict")
	ErrDatabaseError = errors.New("database error")
	ErrCacheError    = errors.New("cache error")
)

// AppError represents an application error with structured information
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
	Err        error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// NewAppErrorWithDetails creates a new application error with details
func NewAppErrorWithDetails(code ErrorCode, message, details string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: httpStatus,
	}
}

// NewAppErrorWithErr creates a new application error wrapping another error
func NewAppErrorWithErr(code ErrorCode, message string, httpStatus int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
		Details:    err.Error(),
	}
}

// RepositoryError represents a repository-specific error
type RepositoryError struct {
	Code    string
	Message string
	Err     error
}

func (e *RepositoryError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// NewRepositoryError creates a new repository error
func NewRepositoryError(code, message string, err error) *RepositoryError {
	return &RepositoryError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Helper functions to create common errors

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return NewAppError(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

// NewInvalidInputError creates an invalid input error
func NewInvalidInputError(message string) *AppError {
	return NewAppError(ErrCodeInvalidInput, message, http.StatusBadRequest)
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "unauthorized access"
	}
	return NewAppError(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "forbidden access"
	}
	return NewAppError(ErrCodeForbidden, message, http.StatusForbidden)
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return NewAppError(ErrCodeConflict, message, http.StatusConflict)
}

// NewInternalError creates an internal server error
func NewInternalError(message string, err error) *AppError {
	if message == "" {
		message = "internal server error"
	}
	return NewAppErrorWithErr(ErrCodeInternal, message, http.StatusInternalServerError, err)
}

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return NewAppError(ErrCodeValidation, message, http.StatusBadRequest)
}

// NewServiceError creates a service error
func NewServiceError(code, message string, err error) *AppError {
	return NewAppErrorWithErr(ErrorCode(code), message, http.StatusInternalServerError, err)
}

// NewTooManyRequestsError creates a rate limit error
func NewTooManyRequestsError(message string) *AppError {
	if message == "" {
		message = "too many requests"
	}
	return NewAppError(ErrCodeTooManyRequests, message, http.StatusTooManyRequests)
}

// Error checking functions

// IsNotFound checks if error is a not found error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrNotFound) {
		return true
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrCodeNotFound
	}
	return false
}

// IsNotFoundError is an alias for IsNotFound for backward compatibility
func IsNotFoundError(err error) bool {
	return IsNotFound(err)
}

// IsDuplicate checks if error is a duplicate error
func IsDuplicate(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrDuplicate) {
		return true
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrCodeDuplicate
	}
	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return repoErr.Code == "DUPLICATE"
	}
	return false
}

// IsValidationError checks if error is a validation error
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrCodeValidation || appErr.Code == ErrCodeInvalidInput
	}
	return false
}
