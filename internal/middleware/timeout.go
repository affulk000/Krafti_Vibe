package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// TimeoutConfig holds configuration for timeout middleware
type TimeoutConfig struct {
	// Timeout is the maximum duration for request processing
	Timeout time.Duration
	// Logger for logging timeout events
	Logger *zap.Logger
	// SkipPaths contains paths that should skip timeout (e.g., websockets, long-polling)
	SkipPaths []string
}

// DefaultTimeoutConfig returns default timeout configuration
func DefaultTimeoutConfig(logger *zap.Logger) TimeoutConfig {
	return TimeoutConfig{
		Timeout: 30 * time.Second,
		Logger:  logger,
		SkipPaths: []string{
			"/api/v1/ws",
			"/api/v1/websocket",
			"/health/live",
			"/health/ready",
			"/debug/pprof",
		},
	}
}

// Timeout returns a middleware that enforces request timeout
func Timeout(config TimeoutConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if path should skip timeout
		path := c.Path()
		for _, skipPath := range config.SkipPaths {
			if path == skipPath || len(path) >= len(skipPath) && path[:len(skipPath)] == skipPath {
				return c.Next()
			}
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.UserContext(), config.Timeout)
		defer cancel()

		// Replace context
		c.SetUserContext(ctx)

		// Channel to signal completion
		done := make(chan error, 1)

		// Execute handler in goroutine
		go func() {
			done <- c.Next()
		}()

		// Wait for completion or timeout
		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			// Log timeout event
			config.Logger.Warn("request timeout",
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.Duration("timeout", config.Timeout),
				zap.String("ip", c.IP()),
				zap.String("request_id", c.Get("X-Request-ID")),
			)

			return c.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "REQUEST_TIMEOUT",
					"message": "Request processing timeout exceeded",
					"timeout": config.Timeout.String(),
				},
			})
		}
	}
}
