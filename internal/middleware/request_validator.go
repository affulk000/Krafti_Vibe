package middleware

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ValidatorConfig holds configuration for request validator
type ValidatorConfig struct {
	Logger    *zap.Logger
	Validator *validator.Validate
}

// DefaultValidatorConfig returns default validator configuration
func DefaultValidatorConfig(logger *zap.Logger) ValidatorConfig {
	return ValidatorConfig{
		Logger:    logger,
		Validator: validator.New(),
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// ValidateRequest validates the request body against a struct
func ValidateRequest(c *fiber.Ctx, v interface{}, config ValidatorConfig) error {
	// Parse body
	if err := c.BodyParser(v); err != nil {
		config.Logger.Debug("failed to parse request body",
			zap.Error(err),
			zap.String("path", c.Path()),
		)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "INVALID_REQUEST_BODY",
				"message": "Invalid request body format",
				"details": err.Error(),
			},
		})
	}

	// Validate struct
	if err := config.Validator.Struct(v); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "VALIDATION_ERROR",
					"message": "Request validation failed",
					"details": err.Error(),
				},
			})
		}

		// Build detailed validation errors
		var errors []ValidationError
		for _, fieldErr := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   fieldErr.Field(),
				Tag:     fieldErr.Tag(),
				Message: getValidationMessage(fieldErr),
			})
		}

		config.Logger.Debug("request validation failed",
			zap.String("path", c.Path()),
			zap.Int("error_count", len(errors)),
		)

		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Request validation failed",
				"details": errors,
			},
		})
	}

	return nil
}

// getValidationMessage returns a human-readable validation error message
func getValidationMessage(fieldErr validator.FieldError) string {
	field := fieldErr.Field()
	tag := fieldErr.Tag()
	param := fieldErr.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be a number", field)
	case "boolean":
		return fmt.Sprintf("%s must be a boolean", field)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, tag)
	}
}

// ValidateJSON validates JSON against a struct without consuming the request body
func ValidateJSON(c *fiber.Ctx, v interface{}, config ValidatorConfig) error {
	// Get body bytes
	body := c.Body()
	if len(body) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "EMPTY_REQUEST_BODY",
				"message": "Request body is required",
			},
		})
	}

	// Unmarshal JSON
	if err := json.Unmarshal(body, v); err != nil {
		config.Logger.Debug("failed to unmarshal JSON",
			zap.Error(err),
			zap.String("path", c.Path()),
		)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "INVALID_JSON",
				"message": "Invalid JSON format",
				"details": err.Error(),
			},
		})
	}

	// Validate struct
	if err := config.Validator.Struct(v); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "VALIDATION_ERROR",
					"message": "Request validation failed",
					"details": err.Error(),
				},
			})
		}

		// Build detailed validation errors
		var errors []ValidationError
		for _, fieldErr := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   fieldErr.Field(),
				Tag:     fieldErr.Tag(),
				Message: getValidationMessage(fieldErr),
			})
		}

		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Request validation failed",
				"details": errors,
			},
		})
	}

	return nil
}

// ValidateQuery validates query parameters
func ValidateQuery(c *fiber.Ctx, v interface{}, config ValidatorConfig) error {
	// Parse query parameters
	if err := c.QueryParser(v); err != nil {
		config.Logger.Debug("failed to parse query parameters",
			zap.Error(err),
			zap.String("path", c.Path()),
		)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "INVALID_QUERY_PARAMETERS",
				"message": "Invalid query parameters",
				"details": err.Error(),
			},
		})
	}

	// Validate struct
	if err := config.Validator.Struct(v); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "VALIDATION_ERROR",
					"message": "Query validation failed",
					"details": err.Error(),
				},
			})
		}

		// Build detailed validation errors
		var errors []ValidationError
		for _, fieldErr := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   fieldErr.Field(),
				Tag:     fieldErr.Tag(),
				Message: getValidationMessage(fieldErr),
			})
		}

		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Query validation failed",
				"details": errors,
			},
		})
	}

	return nil
}
