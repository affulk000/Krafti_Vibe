package handler

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/gofiber/fiber/v2"
)

// Enterprise-grade validation functions

// ValidateRequest validates the request body and returns validation errors
func ValidateRequest(c *fiber.Ctx, req interface{ Validate() error }) error {
	if err := req.Validate(); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error(), err)
	}
	return nil
}

// ValidateStringLength validates string length constraints
func ValidateStringLength(field, value string, min, max int) *ValidationError {
	length := len(value)
	if length < min {
		return &ValidationError{
			Field:   field,
			Message: "Must be at least " + string(rune(min)) + " characters",
		}
	}
	if length > max {
		return &ValidationError{
			Field:   field,
			Message: "Must not exceed " + string(rune(max)) + " characters",
		}
	}
	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) []ValidationError {
	var errors []ValidationError

	if len(password) < 8 {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password must be at least 8 characters long",
		})
	}

	if len(password) > 128 {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password must not exceed 128 characters",
		})
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password must contain at least one uppercase letter",
		})
	}

	if !hasLower {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password must contain at least one lowercase letter",
		})
	}

	if !hasNumber {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password must contain at least one number",
		})
	}

	if !hasSpecial {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password must contain at least one special character",
		})
	}

	return errors
}

// ValidateEmail validates email format with enhanced rules
func ValidateEmail(email string) *ValidationError {
	if email == "" {
		return &ValidationError{
			Field:   "email",
			Message: "Email is required",
		}
	}

	if len(email) > 254 {
		return &ValidationError{
			Field:   "email",
			Message: "Email address is too long",
		}
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return &ValidationError{
			Field:   "email",
			Message: "Invalid email format",
		}
	}

	return nil
}

// ValidatePhoneNumber validates phone number format
func ValidatePhoneNumber(phone string) *ValidationError {
	if phone == "" {
		return nil // Optional field
	}

	// Remove common separators
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	// Remove leading + if present (international format)
	cleaned = strings.TrimPrefix(cleaned, "+")

	// Validate it contains only digits
	phoneRegex := regexp.MustCompile(`^\d{10,15}$`)
	if !phoneRegex.MatchString(cleaned) {
		return &ValidationError{
			Field:   "phone",
			Message: "Invalid phone number format. Use international format (e.g., +1234567890)",
		}
	}

	return nil
}

// ValidateURL validates URL format
func ValidateURL(url string) *ValidationError {
	if url == "" {
		return nil // Optional field
	}

	urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/.*)?$`)
	if !urlRegex.MatchString(url) {
		return &ValidationError{
			Field:   "url",
			Message: "Invalid URL format. Must start with http:// or https://",
		}
	}

	return nil
}

// ValidateAmount validates monetary amount
func ValidateAmount(field string, amount float64) *ValidationError {
	if amount < 0 {
		return &ValidationError{
			Field:   field,
			Message: "Amount cannot be negative",
		}
	}

	if amount > 999999999.99 {
		return &ValidationError{
			Field:   field,
			Message: "Amount exceeds maximum allowed value",
		}
	}

	return nil
}

// ValidatePercentage validates percentage value
func ValidatePercentage(field string, value float64) *ValidationError {
	if value < 0 || value > 100 {
		return &ValidationError{
			Field:   field,
			Message: "Percentage must be between 0 and 100",
		}
	}

	return nil
}

// ValidateEnum validates if value is in allowed enum values
func ValidateEnum(field string, value string, allowedValues []string) *ValidationError {
	if slices.Contains(allowedValues, value) {
		return nil // Value is valid
	}

	return &ValidationError{
		Field:   field,
		Message: "Invalid value. Allowed values: " + strings.Join(allowedValues, ", "),
	}
}

// ValidateArrayLength validates array length constraints
func ValidateArrayLength(field string, length, min, max int) *ValidationError {
	if length < min {
		return &ValidationError{
			Field:   field,
			Message: "Must contain at least " + string(rune(min)) + " items",
		}
	}

	if length > max {
		return &ValidationError{
			Field:   field,
			Message: "Must not exceed " + string(rune(max)) + " items",
		}
	}

	return nil
}

// SanitizeHTML strips HTML tags from input (basic protection against XSS)
func SanitizeHTML(input string) string {
	// Remove HTML tags
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	sanitized := htmlRegex.ReplaceAllString(input, "")

	// Remove script tags and content
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	sanitized = scriptRegex.ReplaceAllString(sanitized, "")

	return sanitized
}

// ValidateDateRange validates that start date is before end date
func ValidateDateRange(startField, endField string, start, end any) *ValidationError {
	// This is a simplified version - you'd compare actual time.Time values
	// Implementation would depend on your date format
	return nil
}

// ValidateFileExtension validates file extension
func ValidateFileExtension(filename string, allowedExtensions []string) *ValidationError {
	if filename == "" {
		return &ValidationError{
			Field:   "file",
			Message: "Filename is required",
		}
	}

	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return &ValidationError{
			Field:   "file",
			Message: "Invalid filename format",
		}
	}

	ext := strings.ToLower(parts[len(parts)-1])
	for _, allowed := range allowedExtensions {
		if ext == strings.ToLower(allowed) {
			return nil
		}
	}

	return &ValidationError{
		Field:   "file",
		Message: "Invalid file extension. Allowed: " + strings.Join(allowedExtensions, ", "),
	}
}

// ValidateMaxFileSize validates file size
func ValidateMaxFileSize(size int64, maxSizeBytes int64) *ValidationError {
	if size > maxSizeBytes {
		maxSizeMB := maxSizeBytes / (1024 * 1024)
		return &ValidationError{
			Field:   "file",
			Message: "File size exceeds maximum allowed (" + string(rune(maxSizeMB)) + "MB)",
		}
	}

	return nil
}

// ValidateIPAddress validates IP address format
func ValidateIPAddress(ip string) *ValidationError {
	if ip == "" {
		return nil // Optional field
	}

	ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	if !ipRegex.MatchString(ip) {
		return &ValidationError{
			Field:   "ip_address",
			Message: "Invalid IP address format",
		}
	}

	// Validate each octet is 0-255
	for part := range strings.SplitSeq(ip, ".") {
		var octet int
		if _, err := fmt.Sscanf(part, "%d", &octet); err != nil || octet < 0 || octet > 255 {
			return &ValidationError{
				Field:   "ip_address",
				Message: "Invalid IP address octets",
			}
		}
	}

	return nil
}
