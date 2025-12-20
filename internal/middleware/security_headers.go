package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// SecurityHeadersConfig defines configuration for security headers
type SecurityHeadersConfig struct {
	// XSSProtection enables XSS protection header
	XSSProtection bool
	// ContentTypeNosniff enables X-Content-Type-Options header
	ContentTypeNosniff bool
	// XFrameOptions sets X-Frame-Options header (DENY, SAMEORIGIN, ALLOW-FROM uri)
	XFrameOptions string
	// HSTSMaxAge sets Strict-Transport-Security max-age
	HSTSMaxAge int
	// HSTSExcludeSubdomains excludes subdomains from HSTS
	HSTSExcludeSubdomains bool
	// ContentSecurityPolicy sets CSP header
	ContentSecurityPolicy string
	// CSPReportOnly sets CSP in report-only mode
	CSPReportOnly bool
	// ReferrerPolicy sets Referrer-Policy header
	ReferrerPolicy string
	// PermissionsPolicy sets Permissions-Policy header
	PermissionsPolicy string
}

// DefaultSecurityHeadersConfig returns default security headers configuration
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		XSSProtection:      true,
		ContentTypeNosniff: true,
		XFrameOptions:      "SAMEORIGIN",
		HSTSMaxAge:         31536000, // 1 year
		HSTSExcludeSubdomains: false,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'",
		CSPReportOnly: false,
		ReferrerPolicy: "strict-origin-when-cross-origin",
		PermissionsPolicy: "geolocation=(), microphone=(), camera=()",
	}
}

// SecurityHeaders returns a middleware that sets security headers
func SecurityHeaders(config SecurityHeadersConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// X-XSS-Protection
		if config.XSSProtection {
			c.Set("X-XSS-Protection", "1; mode=block")
		}

		// X-Content-Type-Options
		if config.ContentTypeNosniff {
			c.Set("X-Content-Type-Options", "nosniff")
		}

		// X-Frame-Options
		if config.XFrameOptions != "" {
			c.Set("X-Frame-Options", config.XFrameOptions)
		}

		// Strict-Transport-Security (HSTS)
		if config.HSTSMaxAge > 0 {
			hsts := "max-age=" + string(rune(config.HSTSMaxAge))
			if !config.HSTSExcludeSubdomains {
				hsts += "; includeSubDomains"
			}
			// Only set HSTS on HTTPS
			if c.Protocol() == "https" {
				c.Set("Strict-Transport-Security", hsts)
			}
		}

		// Content-Security-Policy
		if config.ContentSecurityPolicy != "" {
			headerName := "Content-Security-Policy"
			if config.CSPReportOnly {
				headerName = "Content-Security-Policy-Report-Only"
			}
			c.Set(headerName, config.ContentSecurityPolicy)
		}

		// Referrer-Policy
		if config.ReferrerPolicy != "" {
			c.Set("Referrer-Policy", config.ReferrerPolicy)
		}

		// Permissions-Policy (formerly Feature-Policy)
		if config.PermissionsPolicy != "" {
			c.Set("Permissions-Policy", config.PermissionsPolicy)
		}

		// Remove server identifying headers
		c.Set("X-Powered-By", "")

		return c.Next()
	}
}
