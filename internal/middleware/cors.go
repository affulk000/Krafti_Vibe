package middleware

import (
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	// AllowedOrigins is a list of allowed origins
	// Use "*" to allow all origins (not recommended for production with credentials)
	AllowedOrigins []string

	// AllowedMethods is a list of allowed HTTP methods
	// Default: GET, POST, PUT, PATCH, DELETE, OPTIONS
	AllowedMethods []string

	// AllowedHeaders is a list of allowed request headers
	// Default: Accept, Authorization, Content-Type, X-Request-ID, X-Tenant-ID, X-Organization-ID
	AllowedHeaders []string

	// ExposedHeaders is a list of headers exposed to the client
	// Default: X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
	ExposedHeaders []string

	// AllowCredentials indicates whether credentials are allowed
	// When true, AllowedOrigins cannot contain "*"
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) preflight results can be cached
	// Default: 86400 (24 hours)
	MaxAge int

	// AllowPrivateNetwork allows requests from private networks (RFC 1918)
	// Useful for development and internal tools
	AllowPrivateNetwork bool

	// Logger for logging CORS events
	Logger *zap.Logger

	// OriginValidator is a custom origin validation function
	// If provided, it takes precedence over AllowedOrigins
	OriginValidator func(origin string) bool

	// SkipPreflight allows skipping preflight for specific routes
	SkipPreflight func(*fiber.Ctx) bool

	// Enabled determines if CORS is enabled
	Enabled bool

	// Development mode allows all origins (overrides AllowedOrigins)
	Development bool
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig(logger *zap.Logger) CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Request-ID",
			"X-Tenant-ID",
			"X-Organization-ID",
			"X-API-Key",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
			"X-RateLimit-Window",
		},
		AllowCredentials:    false,
		MaxAge:              86400, // 24 hours
		AllowPrivateNetwork: false,
		Enabled:             true,
		Development:         false,
		Logger:              logger,
	}
}

// ProductionCORSConfig returns production-safe CORS configuration
func ProductionCORSConfig(allowedOrigins []string, logger *zap.Logger) CORSConfig {
	config := DefaultCORSConfig(logger)
	config.AllowedOrigins = allowedOrigins
	config.AllowCredentials = true
	config.Development = false

	// Production should have explicit origins, not wildcards
	for _, origin := range allowedOrigins {
		if origin == "*" {
			logger.Warn("CORS: Using wildcard (*) origin in production is not recommended with credentials")
		}
	}

	return config
}

// DevelopmentCORSConfig returns development-friendly CORS configuration
func DevelopmentCORSConfig(logger *zap.Logger) CORSConfig {
	config := DefaultCORSConfig(logger)
	config.Development = true
	config.AllowCredentials = true
	config.AllowPrivateNetwork = true
	config.AllowedOrigins = []string{"*"}

	return config
}

// CORSMiddleware creates CORS middleware
func CORSMiddleware(config CORSConfig) fiber.Handler {
	// Validate configuration
	if err := validateCORSConfig(&config); err != nil {
		config.Logger.Error("Invalid CORS configuration", zap.Error(err))
		// Use safe defaults
		config = DefaultCORSConfig(config.Logger)
	}

	// Log configuration
	if config.Logger != nil {
		config.Logger.Info("CORS middleware initialized",
			zap.Strings("allowed_origins", config.AllowedOrigins),
			zap.Bool("allow_credentials", config.AllowCredentials),
			zap.Bool("development", config.Development),
		)
	}

	return func(c *fiber.Ctx) error {
		// Skip if CORS is disabled
		if !config.Enabled {
			return c.Next()
		}

		origin := c.Get("Origin")

		// No origin header, not a CORS request
		if origin == "" {
			return c.Next()
		}

		// Validate origin
		if !isOriginAllowed(origin, config) {
			if config.Logger != nil {
				config.Logger.Warn("CORS: Origin not allowed",
					zap.String("origin", origin),
					zap.String("path", c.Path()),
				)
			}
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "cors_origin_not_allowed",
				"message": "Origin not allowed by CORS policy",
			})
		}

		// Set Access-Control-Allow-Origin
		if config.Development || slices.Contains(config.AllowedOrigins, "*") {
			if config.AllowCredentials {
				// Cannot use wildcard with credentials, must echo origin
				c.Set("Access-Control-Allow-Origin", origin)
				c.Set("Vary", "Origin")
			} else {
				c.Set("Access-Control-Allow-Origin", "*")
			}
		} else {
			// Echo the allowed origin
			c.Set("Access-Control-Allow-Origin", origin)
			c.Set("Vary", "Origin")
		}

		// Set Access-Control-Allow-Credentials
		if config.AllowCredentials {
			c.Set("Access-Control-Allow-Credentials", "true")
		}

		// Set Access-Control-Expose-Headers
		if len(config.ExposedHeaders) > 0 {
			c.Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
		}

		// Handle preflight request
		if c.Method() == fiber.MethodOptions {
			// Check if we should skip preflight
			if config.SkipPreflight != nil && config.SkipPreflight(c) {
				return c.Next()
			}

			// Set Access-Control-Allow-Methods
			if len(config.AllowedMethods) > 0 {
				c.Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			}

			// Set Access-Control-Allow-Headers
			requestedHeaders := c.Get("Access-Control-Request-Headers")
			if requestedHeaders != "" {
				// Validate requested headers
				if areHeadersAllowed(requestedHeaders, config.AllowedHeaders) {
					c.Set("Access-Control-Allow-Headers", requestedHeaders)
				} else {
					// Return allowed headers
					c.Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
				}
			} else if len(config.AllowedHeaders) > 0 {
				c.Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			}

			// Set Access-Control-Max-Age
			if config.MaxAge > 0 {
				c.Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}

			// Handle private network access
			if config.AllowPrivateNetwork {
				requestPrivateNetwork := c.Get("Access-Control-Request-Private-Network")
				if requestPrivateNetwork == "true" {
					c.Set("Access-Control-Allow-Private-Network", "true")
				}
			}

			// Return 204 No Content for preflight
			return c.SendStatus(fiber.StatusNoContent)
		}

		// Continue to actual request
		return c.Next()
	}
}

// isOriginAllowed checks if an origin is allowed
func isOriginAllowed(origin string, config CORSConfig) bool {
	// Development mode allows all
	if config.Development {
		return true
	}

	// Custom validator takes precedence
	if config.OriginValidator != nil {
		return config.OriginValidator(origin)
	}

	// Check if wildcard
	if slices.Contains(config.AllowedOrigins, "*") {
		return true
	}

	// Check exact match
	if slices.Contains(config.AllowedOrigins, origin) {
		return true
	}

	// Check pattern match (e.g., *.example.com)
	for _, allowed := range config.AllowedOrigins {
		if matchOriginPattern(origin, allowed) {
			return true
		}
	}

	return false
}

// matchOriginPattern matches origin against a pattern with wildcards
func matchOriginPattern(origin, pattern string) bool {
	// Parse URLs
	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	patternURL, err := url.Parse(pattern)
	if err != nil {
		// If pattern is not a valid URL, try as host pattern
		return matchHostPattern(originURL.Host, pattern)
	}

	// Scheme must match (unless pattern has no scheme)
	if patternURL.Scheme != "" && originURL.Scheme != patternURL.Scheme {
		return false
	}

	// Port must match (unless pattern has no port)
	if patternURL.Port() != "" && originURL.Port() != patternURL.Port() {
		return false
	}

	// Match hostname (supports wildcards)
	return matchHostPattern(originURL.Hostname(), patternURL.Hostname())
}

// matchHostPattern matches hostname against a pattern with wildcards
func matchHostPattern(host, pattern string) bool {
	// Exact match
	if host == pattern {
		return true
	}

	// Wildcard subdomain match (e.g., *.example.com)
	if strings.HasPrefix(pattern, "*.") {
		domain := pattern[2:] // Remove *.
		return strings.HasSuffix(host, "."+domain) || host == domain
	}

	return false
}

func areHeadersAllowed(requestedHeaders string, allowedHeaders []string) bool {
	// Normalize allowed headers only once
	normalizedAllowed := make([]string, len(allowedHeaders))
	for i, h := range allowedHeaders {
		normalizedAllowed[i] = strings.ToLower(strings.TrimSpace(h))
	}

	// Range over SplitSeq iterator (one variable only)
	for h := range strings.SplitSeq(requestedHeaders, ",") {
		h = strings.ToLower(strings.TrimSpace(h))

		if !slices.Contains(normalizedAllowed, h) {
			return false
		}
	}

	return true
}

// validateCORSConfig validates CORS configuration
func validateCORSConfig(config *CORSConfig) error {
	// If credentials are enabled, cannot use wildcard origin
	if config.AllowCredentials && slices.Contains(config.AllowedOrigins, "*") && !config.Development {
		// Log warning and fix
		if config.Logger != nil {
			config.Logger.Warn("CORS: Cannot use wildcard origin with credentials in production. Using specific origins.")
		}
		// Remove wildcard if there are other origins
		if len(config.AllowedOrigins) > 1 {
			filtered := make([]string, 0, len(config.AllowedOrigins))
			for _, origin := range config.AllowedOrigins {
				if origin != "*" {
					filtered = append(filtered, origin)
				}
			}
			config.AllowedOrigins = filtered
		}
	}

	// Set defaults if empty
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
		}
	}

	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Request-ID",
		}
	}

	return nil
}

// CORSWithConfig creates CORS middleware with custom configuration
func CORSWithConfig(allowedOrigins []string, allowCredentials bool, logger *zap.Logger) fiber.Handler {
	config := DefaultCORSConfig(logger)
	config.AllowedOrigins = allowedOrigins
	config.AllowCredentials = allowCredentials

	return CORSMiddleware(config)
}

// StrictCORS creates strict CORS middleware for production
func StrictCORS(allowedOrigins []string, logger *zap.Logger) fiber.Handler {
	return CORSMiddleware(ProductionCORSConfig(allowedOrigins, logger))
}

// PermissiveCORS creates permissive CORS middleware for development
func PermissiveCORS(logger *zap.Logger) fiber.Handler {
	return CORSMiddleware(DevelopmentCORSConfig(logger))
}

// DynamicCORS creates CORS middleware with dynamic origin validation
func DynamicCORS(validator func(string) bool, logger *zap.Logger) fiber.Handler {
	config := DefaultCORSConfig(logger)
	config.OriginValidator = validator
	config.AllowCredentials = true

	return CORSMiddleware(config)
}

// MultiOriginCORS creates CORS middleware that allows multiple specific origins
func MultiOriginCORS(origins []string, allowCredentials bool, logger *zap.Logger) fiber.Handler {
	config := DefaultCORSConfig(logger)
	config.AllowedOrigins = origins
	config.AllowCredentials = allowCredentials

	// Add common headers for API usage
	config.AllowedHeaders = append(config.AllowedHeaders,
		"X-API-Version",
		"X-Client-ID",
		"X-Session-ID",
	)

	config.ExposedHeaders = append(config.ExposedHeaders,
		"X-Total-Count",
		"X-Page",
		"X-Page-Size",
		"Link",
	)

	return CORSMiddleware(config)
}

// TenantAwareCORS creates CORS middleware that validates origins against tenant configuration
func TenantAwareCORS(getTenantOrigins func(tenantID string) []string, logger *zap.Logger) fiber.Handler {
	config := DefaultCORSConfig(logger)
	config.AllowCredentials = true

	config.OriginValidator = func(origin string) bool {
		// This is a placeholder - you would get tenant ID from request context
		// and fetch allowed origins for that tenant
		return true // Implement actual tenant-based validation
	}

	return CORSMiddleware(config)
}

// CORSStatus returns CORS configuration status for debugging
func CORSStatus(config CORSConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"cors_enabled":         config.Enabled,
			"allowed_origins":      config.AllowedOrigins,
			"allowed_methods":      config.AllowedMethods,
			"allowed_headers":      config.AllowedHeaders,
			"exposed_headers":      config.ExposedHeaders,
			"allow_credentials":    config.AllowCredentials,
			"max_age":              config.MaxAge,
			"development_mode":     config.Development,
			"private_network":      config.AllowPrivateNetwork,
			"has_custom_validator": config.OriginValidator != nil,
		})
	}
}
