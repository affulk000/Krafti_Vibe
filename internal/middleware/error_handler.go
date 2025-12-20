package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ErrorHandler creates a custom error handler for Fiber
func ErrorHandler(logger *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Default error code
		code := fiber.StatusInternalServerError

		// Retrieve the custom status code if it's a fiber.Error
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		// Log the error
		logger.Error("request error",
			zap.Error(err),
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
			zap.Int("status", code),
			zap.String("ip", c.IP()),
		)

		// Send error response
		return c.Status(code).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    code,
				"message": err.Error(),
			},
		})
	}
}
