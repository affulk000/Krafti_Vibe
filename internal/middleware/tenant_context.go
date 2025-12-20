package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TenantContextConfig holds configuration for tenant context middleware
type TenantContextConfig struct {
	Logger *zap.Logger
	// HeaderName is the name of the header containing tenant ID
	HeaderName string
	// QueryParam is the name of query parameter containing tenant ID
	QueryParam string
	// Required determines if tenant context is mandatory
	Required bool
}

// DefaultTenantContextConfig returns default tenant context configuration
func DefaultTenantContextConfig(logger *zap.Logger) TenantContextConfig {
	return TenantContextConfig{
		Logger:     logger,
		HeaderName: "X-Tenant-ID",
		QueryParam: "tenant_id",
		Required:   false,
	}
}

// TenantContext extracts and validates tenant context
func TenantContext(config TenantContextConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tenantID uuid.UUID
		var tenantIDStr string

		// Try to get tenant ID from header first
		tenantIDStr = c.Get(config.HeaderName)

		// If not in header, try query parameter
		if tenantIDStr == "" {
			tenantIDStr = c.Query(config.QueryParam)
		}

		// If not in query, try to extract from authenticated user
		if tenantIDStr == "" {
			if authCtx, ok := GetAuthContext(c); ok && authCtx.TenantID != uuid.Nil {
				tenantID = authCtx.TenantID
				tenantIDStr = tenantID.String()
			}
		}

		// If we have a tenant ID string, validate and parse it
		if tenantIDStr != "" {
			parsed, err := uuid.Parse(tenantIDStr)
			if err != nil {
				config.Logger.Warn("invalid tenant ID format",
					zap.String("tenant_id", tenantIDStr),
					zap.Error(err),
				)
				if config.Required {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"success": false,
						"error": fiber.Map{
							"code":    "INVALID_TENANT_ID",
							"message": "Invalid tenant ID format",
						},
					})
				}
			} else {
				tenantID = parsed
			}
		}

		// If tenant is required but not found
		if config.Required && tenantID == uuid.Nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "MISSING_TENANT_ID",
					"message": "Tenant ID is required",
				},
			})
		}

		// Store tenant ID in context
		if tenantID != uuid.Nil {
			c.Locals("tenant_id", tenantID)

			// Also check if db_user exists and verify tenant match (for security)
			if dbUser, ok := GetDatabaseUser(c); ok {
				// Platform users can access any tenant
				if !dbUser.IsPlatformUser {
					// Verify user belongs to this tenant
					if dbUser.TenantID != nil && *dbUser.TenantID != tenantID {
						config.Logger.Warn("tenant ID mismatch",
							zap.String("requested_tenant", tenantID.String()),
							zap.String("user_tenant", dbUser.TenantID.String()),
							zap.String("user_id", dbUser.ID.String()),
						)
						return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
							"success": false,
							"error": fiber.Map{
								"code":    "TENANT_ACCESS_DENIED",
								"message": "Access denied to this tenant",
							},
						})
					}
				}
			}
		}

		return c.Next()
	}
}

// GetTenantID retrieves tenant ID from context
func GetTenantID(c *fiber.Ctx) (uuid.UUID, bool) {
	tenantID := c.Locals("tenant_id")
	if tenantID == nil {
		return uuid.Nil, false
	}
	id, ok := tenantID.(uuid.UUID)
	return id, ok
}

// MustGetTenantID retrieves tenant ID or panics
func MustGetTenantID(c *fiber.Ctx) uuid.UUID {
	id, ok := GetTenantID(c)
	if !ok {
		panic("tenant ID not found - did you forget to use tenant context middleware?")
	}
	return id
}

// RequireTenantContext is middleware that requires tenant context to be set
func RequireTenantContext() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if _, ok := GetTenantID(c); !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "MISSING_TENANT_CONTEXT",
					"message": "Tenant context is required for this operation",
				},
			})
		}
		return c.Next()
	}
}

// IsolateTenantData ensures queries are filtered by tenant
// This is a helper for handlers to enforce tenant isolation
func IsolateTenantData(c *fiber.Ctx) func(db any) any {
	return func(db any) any {
		tenantID, hasTenant := GetTenantID(c)

		// If user is platform user, don't filter by tenant
		user, hasUser := GetDatabaseUser(c)
		if hasUser && user.IsPlatformUser {
			return db
		}

		// Apply tenant filter if tenant context exists
		if hasTenant {
			// Type assertion for GORM (you'll need to adjust based on your DB interface)
			// This is a placeholder - implement based on your actual DB abstraction
			_ = tenantID // Use tenantID to filter
		}

		return db
	}
}

// ExtractSubdomain extracts subdomain from request host
func ExtractSubdomain(c *fiber.Ctx) string {
	host := c.Hostname()
	parts := strings.Split(host, ".")

	// If we have at least 3 parts (subdomain.domain.tld)
	if len(parts) >= 3 {
		// Ignore common prefixes
		if parts[0] != "www" && parts[0] != "api" {
			return parts[0]
		}
	}

	return ""
}

// TenantFromSubdomain middleware extracts tenant from subdomain
func TenantFromSubdomain(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		subdomain := ExtractSubdomain(c)
		if subdomain != "" {
			// Store subdomain in context
			c.Locals("tenant_subdomain", subdomain)

			logger.Debug("extracted tenant subdomain",
				zap.String("subdomain", subdomain),
				zap.String("host", c.Hostname()),
			)
		}

		return c.Next()
	}
}

// GetTenantSubdomain retrieves tenant subdomain from context
func GetTenantSubdomain(c *fiber.Ctx) (string, bool) {
	subdomain := c.Locals("tenant_subdomain")
	if subdomain == nil {
		return "", false
	}
	s, ok := subdomain.(string)
	return s, ok
}
