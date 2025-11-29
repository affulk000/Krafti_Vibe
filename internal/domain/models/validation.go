package models

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/google/uuid"
)

// ValidationError represents a validation error for a specific field
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ValidateRequired validates that a field is not empty
func ValidateRequired(fieldName string, value any) error {
	if isEmpty(value) {
		return &ValidationError{
			Field:   fieldName,
			Message: "is required",
		}
	}
	return nil
}

// ValidateEmail validates that a string is a valid email format
func ValidateEmail(fieldName, email string) error {
	if email == "" {
		return nil // Empty is handled by required validator
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return &ValidationError{
			Field:   fieldName,
			Message: "must be a valid email address",
		}
	}
	return nil
}

// ValidateUUID validates that a string is a valid UUID
func ValidateUUID(fieldName string, id uuid.UUID) error {
	if id == uuid.Nil {
		return &ValidationError{
			Field:   fieldName,
			Message: "must be a valid UUID",
		}
	}
	return nil
}

// ValidateMinLength validates that a string meets minimum length
func ValidateMinLength(fieldName string, value string, min int) error {
	if len(value) < min {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("must be at least %d characters", min),
		}
	}
	return nil
}

// ValidateMaxLength validates that a string doesn't exceed maximum length
func ValidateMaxLength(fieldName string, value string, max int) error {
	if len(value) > max {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("must be at most %d characters", max),
		}
	}
	return nil
}

// ValidateRange validates that a number is within a range
func ValidateRange(fieldName string, value, min, max float64) error {
	if value < min || value > max {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("must be between %v and %v", min, max),
		}
	}
	return nil
}

// ValidateOneOf validates that a value is one of the allowed values
func ValidateOneOf(fieldName string, value any, allowedValues ...any) error {
	if slices.Contains(allowedValues, value) {
		return nil
	}
	return &ValidationError{
		Field:   fieldName,
		Message: fmt.Sprintf("must be one of: %v", allowedValues),
	}
}

// ValidateModel validates a model struct using reflection and struct tags
func ValidateModel(model any) ValidationErrors {
	var errors ValidationErrors
	v := reflect.ValueOf(model)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ValidationErrors{ValidationError{Field: "model", Message: "cannot be nil"}}
		}
		v = v.Elem()
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Get field name from JSON tag or use struct field name
		fieldName := getFieldName(field)
		if fieldName == "" || fieldName == "-" {
			continue
		}

		// Check for required tag
		if strings.Contains(field.Tag.Get("validate"), "required") {
			if err := ValidateRequired(fieldName, fieldValue.Interface()); err != nil {
				errors = append(errors, *err.(*ValidationError))
			}
		}

		// Check for email tag
		if strings.Contains(field.Tag.Get("validate"), "email") {
			if fieldValue.Kind() == reflect.String {
				if err := ValidateEmail(fieldName, fieldValue.String()); err != nil {
					errors = append(errors, *err.(*ValidationError))
				}
			}
		}

		// Check for min tag
		if minTag := extractTagValue(field.Tag.Get("validate"), "min"); minTag != "" {
			if fieldValue.Kind() == reflect.String {
				var min int
				fmt.Sscanf(minTag, "%d", &min)
				if err := ValidateMinLength(fieldName, fieldValue.String(), min); err != nil {
					errors = append(errors, *err.(*ValidationError))
				}
			}
		}

		// Check for max tag
		if maxTag := extractTagValue(field.Tag.Get("validate"), "max"); maxTag != "" {
			if fieldValue.Kind() == reflect.String {
				var max int
				fmt.Sscanf(maxTag, "%d", &max)
				if err := ValidateMaxLength(fieldName, fieldValue.String(), max); err != nil {
					errors = append(errors, *err.(*ValidationError))
				}
			}
		}
	}

	return errors
}

// Helper functions

func isEmpty(value any) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

func getFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag != "" && jsonTag != "-" {
		// Get the first part before comma
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			return jsonTag[:idx]
		}
		return jsonTag
	}
	return field.Name
}

func extractTagValue(tag, key string) string {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, key+"=") {
			return strings.TrimPrefix(part, key+"=")
		}
	}
	return ""
}
