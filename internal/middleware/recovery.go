package middleware

import (
	"Krafti_Vibe/internal/pkg/errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Recovery creates a recovery middleware with structured logging
func Recovery(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic
				logger.Error("panic recovered",
					zap.Any("panic", r),
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
					zap.String("request_id", c.Get("X-Request-ID", "unknown")),
					zap.Stack("stack"),
				)

				// Return error response
				appErr := errors.NewInternalError("internal server error", nil)
				c.Status(appErr.HTTPStatus).JSON(fiber.Map{
					"error":   appErr.Message,
					"code":    appErr.Code,
					"details": "An unexpected error occurred",
				})
			}
		}()

		return c.Next()
	}
}
