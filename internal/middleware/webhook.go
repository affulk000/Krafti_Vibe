package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// WebhookConfig holds webhook middleware configuration
type WebhookConfig struct {
	// Secret used to verify webhook signatures
	SigningSecret string
	// Header name containing the signature
	SignatureHeader string
	// Skip signature verification (NOT recommended for production)
	SkipVerification bool
}

// VerifyLogtoWebhook creates middleware to verify Logto webhook signatures
func VerifyLogtoWebhook(config WebhookConfig) fiber.Handler {
	// Set defaults
	if config.SignatureHeader == "" {
		config.SignatureHeader = "Logto-Signature-Sha-256"
	}

	return func(c *fiber.Ctx) error {
		// Skip verification if configured (development only)
		if config.SkipVerification {
			log.Warn("Webhook signature verification is DISABLED - not recommended for production")
			return c.Next()
		}

		// Get signature from header
		signature := c.Get(config.SignatureHeader)
		if signature == "" {
			log.Warnf("Missing webhook signature header: %s", config.SignatureHeader)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing_signature",
				"message": "Webhook signature is required",
			})
		}

		// Get raw body
		body := c.Body()
		if len(body) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "empty_body",
				"message": "Request body is empty",
			})
		}

		// Verify signature
		if !verifySignature(body, signature, config.SigningSecret) {
			log.Errorf("Invalid webhook signature. Expected signature header: %s", config.SignatureHeader)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid_signature",
				"message": "Webhook signature verification failed",
			})
		}

		log.Debug("Webhook signature verified successfully")
		return c.Next()
	}
}

// verifySignature verifies the HMAC SHA-256 signature
func verifySignature(payload []byte, signature, secret string) bool {
	// Create HMAC hash
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)

	// Clean up signature (remove any prefix if present)
	signature = strings.TrimPrefix(signature, "sha256=")
	signature = strings.ToLower(signature)

	// Compare signatures (constant time comparison)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// LogWebhookRequest logs webhook requests for debugging
func LogWebhookRequest() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Infof("Webhook request received: %s %s from %s",
			c.Method(),
			c.Path(),
			c.IP(),
		)

		// Log headers (excluding sensitive ones)
		headers := c.GetReqHeaders()
		for key, values := range headers {
			if !isSensitiveHeader(key) {
				log.Debugf("  %s: %v", key, values)
			}
		}

		return c.Next()
	}
}

func isSensitiveHeader(header string) bool {
	sensitive := []string{
		"authorization",
		"cookie",
		"set-cookie",
		"x-api-key",
		"logto-signature",
	}

	headerLower := strings.ToLower(header)
	for _, s := range sensitive {
		if strings.Contains(headerLower, s) {
			return true
		}
	}
	return false
}
