package middleware

import (
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuditLogConfig holds configuration for audit logging
type AuditLogConfig struct {
	Logger *zap.Logger
	// LogRequestBody determines if request body should be logged
	LogRequestBody bool
	// LogResponseBody determines if response body should be logged
	LogResponseBody bool
	// SensitiveHeaders are headers to redact from logs
	SensitiveHeaders []string
	// MethodsToLog specifies which HTTP methods to log (empty = all)
	MethodsToLog []string
	// PathsToSkip contains paths that should not be logged
	PathsToSkip []string
}

// DefaultAuditLogConfig returns default audit log configuration
func DefaultAuditLogConfig(logger *zap.Logger) AuditLogConfig {
	return AuditLogConfig{
		Logger:           logger,
		LogRequestBody:   false, // Don't log by default for performance
		LogResponseBody:  false,
		SensitiveHeaders: []string{"Authorization", "Cookie", "X-Api-Key"},
		MethodsToLog:     []string{"POST", "PUT", "PATCH", "DELETE"}, // Only mutating operations
		PathsToSkip: []string{
			"/health/live",
			"/health/ready",
			"/api/v1/ping",
		},
	}
}

// AuditLog returns middleware that logs API requests for audit purposes
func AuditLog(config AuditLogConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip paths
		path := c.Path()
		if slices.Contains(config.PathsToSkip, path) {
			return c.Next()
		}

		// Check if method should be logged
		method := c.Method()
		if !slices.Contains(config.MethodsToLog, method) {
			return c.Next()
		}

		// Capture start time
		start := time.Now()

		// Extract user information
		var userID string
		var tenantID string
		var userEmail string
		var userRole string

		if authCtx, ok := GetAuthContext(c); ok {
			userID = authCtx.UserID.String()
			tenantID = authCtx.TenantID.String()
			userEmail = authCtx.Email
		}

		if dbUser, ok := GetDatabaseUser(c); ok {
			userRole = string(dbUser.Role)
		}

		// Get request ID
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Execute request
		err := c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build audit log entry
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.Get("User-Agent")),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", latency),
			zap.Int("response_size", len(c.Response().Body())),
		}

		// Add user context if available
		if userID != "" {
			fields = append(fields,
				zap.String("user_id", userID),
				zap.String("user_email", userEmail),
				zap.String("user_role", userRole),
			)
		}

		if tenantID != "" && tenantID != uuid.Nil.String() {
			fields = append(fields, zap.String("tenant_id", tenantID))
		}

		// Add error if present
		if err != nil {
			fields = append(fields, zap.Error(err))
		}

		// Log based on status code
		statusCode := c.Response().StatusCode()
		if statusCode >= 500 {
			config.Logger.Error("api_audit", fields...)
		} else if statusCode >= 400 {
			config.Logger.Warn("api_audit", fields...)
		} else {
			config.Logger.Info("api_audit", fields...)
		}

		return err
	}
}

// AuditAction logs a specific action for audit trail
func AuditAction(logger *zap.Logger, c *fiber.Ctx, action string, resource string, details map[string]any) {
	fields := []zap.Field{
		zap.String("action", action),
		zap.String("resource", resource),
		zap.String("request_id", c.Get("X-Request-ID")),
		zap.String("ip", c.IP()),
		zap.Time("timestamp", time.Now()),
	}

	// Add user context
	if authCtx, ok := GetAuthContext(c); ok {
		fields = append(fields,
			zap.String("user_id", authCtx.UserID.String()),
			zap.String("user_email", authCtx.Email),
		)
	}

	// Add tenant context
	if tenantID, ok := GetTenantID(c); ok {
		fields = append(fields, zap.String("tenant_id", tenantID.String()))
	}

	// Add custom details
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	logger.Info("audit_action", fields...)
}
