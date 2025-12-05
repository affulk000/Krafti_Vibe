package middleware

import (
	"Krafti_Vibe/internal/infrastructure/logto"
	"fmt"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// AuthContext contains authentication and authorization information
type AuthContext struct {
	Token          jwt.Token
	Subject        string
	ClientID       string
	OrganizationID string
	TenantID       uuid.UUID
	Scopes         []string
	Audience       []string
	UserID         uuid.UUID
	IsM2M          bool
	Claims         *logto.TokenClaims
}

// MiddlewareConfig holds configuration for auth middleware
type MiddlewareConfig struct {
	// Required audience for token validation
	RequiredAudience string
	// Required scopes for access
	RequiredScopes []string
	// Whether to validate organization context
	ValidateOrganization bool
	// Skip authentication (for public routes)
	Skip func(*fiber.Ctx) bool
	// Custom error handler
	ErrorHandler func(*fiber.Ctx, error) error
}

// Default error handler
func defaultErrorHandler(c *fiber.Ctx, err error) error {
	if authErr, ok := err.(*AuthorizationError); ok {
		return c.Status(authErr.Status).JSON(fiber.Map{
			"error":   "authorization_error",
			"message": authErr.Message,
		})
	}
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error":   "unauthorized",
		"message": err.Error(),
	})
}

// AuthorizationError represents an authorization error with HTTP status
type AuthorizationError struct {
	Message string
	Status  int
}

func (e *AuthorizationError) Error() string {
	return e.Message
}

// NewAuthorizationError creates a new authorization error
func NewAuthorizationError(message string, status ...int) *AuthorizationError {
	statusCode := fiber.StatusForbidden
	if len(status) > 0 {
		statusCode = status[0]
	}
	return &AuthorizationError{
		Message: message,
		Status:  statusCode,
	}
}

// AuthMiddleware creates a new authentication middleware
func AuthMiddleware(validator *logto.TokenValidator, config ...MiddlewareConfig) fiber.Handler {
	cfg := MiddlewareConfig{
		ErrorHandler: defaultErrorHandler,
	}

	if len(config) > 0 {
		cfg = config[0]
		if cfg.ErrorHandler == nil {
			cfg.ErrorHandler = defaultErrorHandler
		}
	}

	return func(c *fiber.Ctx) error {
		// Skip if configured
		if cfg.Skip != nil && cfg.Skip(c) {
			return c.Next()
		}

		// Extract and validate token
		tokenString, err := validator.ExtractToken(c)
		if err != nil {
			return cfg.ErrorHandler(c, NewAuthorizationError("Missing or invalid authorization header", fiber.StatusUnauthorized))
		}

		token, err := validator.ValidateToken(tokenString)
		if err != nil {
			return cfg.ErrorHandler(c, err)
		}

		// Get claims
		claims := validator.GetClaims(token)

		// Validate audience if required
		if cfg.RequiredAudience != "" && !claims.HasAudience(cfg.RequiredAudience) {
			return cfg.ErrorHandler(c, NewAuthorizationError("Invalid audience", fiber.StatusForbidden))
		}

		// Validate scopes if required
		if len(cfg.RequiredScopes) > 0 && !claims.HasScopes(cfg.RequiredScopes) {
			return cfg.ErrorHandler(c, NewAuthorizationError(
				fmt.Sprintf("Insufficient permissions. Required scopes: %s", strings.Join(cfg.RequiredScopes, ", ")),
				fiber.StatusForbidden,
			))
		}

		// Build auth context
		authCtx, err := buildAuthContext(token, claims)
		if err != nil {
			return cfg.ErrorHandler(c, err)
		}

		// Validate organization context if required
		if cfg.ValidateOrganization {
			if err := validateOrganizationContext(c, authCtx); err != nil {
				return cfg.ErrorHandler(c, err)
			}
		}

		// Store auth context in locals
		c.Locals("auth", authCtx)
		c.Locals("token", token)
		c.Locals("claims", claims)

		return c.Next()
	}
}

// RequireScopes creates middleware that requires specific scopes
func RequireScopes(scopes ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx := GetAuthContext(c)
		if authCtx == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
		}

		for _, required := range scopes {
			found := false
			found = slices.Contains(authCtx.Scopes, required)
			if !found {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":    "insufficient_scope",
					"message":  fmt.Sprintf("Required scope missing: %s", required),
					"required": scopes,
					"current":  authCtx.Scopes,
				})
			}
		}

		return c.Next()
	}
}

// RequireAudience creates middleware that requires specific audience
func RequireAudience(audience string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx := GetAuthContext(c)
		if authCtx == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
		}

		if !authCtx.Claims.HasAudience(audience) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "invalid_audience",
				"message": fmt.Sprintf("Required audience: %s", audience),
			})
		}

		return c.Next()
	}
}

// RequireOrganization creates middleware that validates organization context
func RequireOrganization() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx := GetAuthContext(c)
		if authCtx == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
		}

		if authCtx.OrganizationID == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "no_organization",
				"message": "Organization context required",
			})
		}

		// Extract organization ID from request (path param, query, or header)
		requestOrgID := extractOrganizationFromRequest(c)
		if requestOrgID != "" && requestOrgID != authCtx.OrganizationID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "organization_mismatch",
				"message": "Token organization does not match requested organization",
			})
		}

		return c.Next()
	}
}

// RequireTenant creates middleware that validates tenant context
func RequireTenant() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx := GetAuthContext(c)
		if authCtx == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
		}

		if authCtx.TenantID == uuid.Nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "no_tenant",
				"message": "Tenant context required",
			})
		}

		// Extract tenant ID from request
		requestTenantID := extractTenantFromRequest(c)
		if requestTenantID != uuid.Nil && requestTenantID != authCtx.TenantID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "tenant_mismatch",
				"message": "Token tenant does not match requested tenant",
			})
		}

		return c.Next()
	}
}

// RequireM2M ensures the request is from a machine-to-machine client
func RequireM2M() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx := GetAuthContext(c)
		if authCtx == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
		}

		if !authCtx.IsM2M {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "invalid_grant_type",
				"message": "Machine-to-machine authentication required",
			})
		}

		return c.Next()
	}
}

// GetAuthContext retrieves the auth context from fiber locals
func GetAuthContext(c *fiber.Ctx) *AuthContext {
	if authCtx := c.Locals("auth"); authCtx != nil {
		if ctx, ok := authCtx.(*AuthContext); ok {
			return ctx
		}
	}
	return nil
}

// GetToken retrieves the JWT token from fiber locals
func GetToken(c *fiber.Ctx) jwt.Token {
	if token := c.Locals("token"); token != nil {
		if t, ok := token.(jwt.Token); ok {
			return t
		}
	}
	return nil
}

// GetClaims retrieves the token claims from fiber locals
func GetClaims(c *fiber.Ctx) *logto.TokenClaims {
	if claims := c.Locals("claims"); claims != nil {
		if c, ok := claims.(*logto.TokenClaims); ok {
			return c
		}
	}
	return nil
}

// MustGetAuthContext retrieves auth context or panics
func MustGetAuthContext(c *fiber.Ctx) *AuthContext {
	authCtx := GetAuthContext(c)
	if authCtx == nil {
		panic("auth context not found - ensure auth middleware is applied")
	}
	return authCtx
}

// Helper functions

func buildAuthContext(token jwt.Token, claims *logto.TokenClaims) (*AuthContext, error) {
	authCtx := &AuthContext{
		Token:          token,
		Subject:        claims.Subject,
		ClientID:       claims.ClientID,
		OrganizationID: claims.OrganizationID,
		Scopes:         claims.Scopes,
		Audience:       claims.Audience,
		Claims:         claims,
	}

	// Determine if this is M2M (no user subject, only client_id)
	authCtx.IsM2M = claims.Subject == "" || strings.HasPrefix(claims.Subject, "client_")

	// Parse user ID from subject
	if claims.Subject != "" && !authCtx.IsM2M {
		if userID, err := uuid.Parse(claims.Subject); err == nil {
			authCtx.UserID = userID
		}
	}

	// Parse tenant ID from organization ID or custom claim
	if claims.OrganizationID != "" {
		// Remove urn:logto:organization: prefix if present
		orgID := strings.TrimPrefix(claims.OrganizationID, "urn:logto:organization:")
		if tenantID, err := uuid.Parse(orgID); err == nil {
			authCtx.TenantID = tenantID
		}
	}

	return authCtx, nil
}

func validateOrganizationContext(c *fiber.Ctx, authCtx *AuthContext) error {
	if authCtx.OrganizationID == "" {
		return NewAuthorizationError("Organization context required for this operation", fiber.StatusForbidden)
	}

	// Check if the organization audience is present
	hasOrgAudience := false
	for _, aud := range authCtx.Audience {
		if strings.HasPrefix(aud, "urn:logto:organization:") {
			hasOrgAudience = true
			break
		}
	}

	if !hasOrgAudience {
		return NewAuthorizationError("Invalid organization audience", fiber.StatusForbidden)
	}

	return nil
}

func extractOrganizationFromRequest(c *fiber.Ctx) string {
	// Try path parameter
	if orgID := c.Params("organizationId"); orgID != "" {
		return orgID
	}
	if orgID := c.Params("orgId"); orgID != "" {
		return orgID
	}

	// Try query parameter
	if orgID := c.Query("organizationId"); orgID != "" {
		return orgID
	}
	if orgID := c.Query("orgId"); orgID != "" {
		return orgID
	}

	// Try header
	if orgID := c.Get("X-Organization-ID"); orgID != "" {
		return orgID
	}

	return ""
}

func extractTenantFromRequest(c *fiber.Ctx) uuid.UUID {
	// Try path parameter
	if tenantID := c.Params("tenantId"); tenantID != "" {
		if id, err := uuid.Parse(tenantID); err == nil {
			return id
		}
	}

	// Try query parameter
	if tenantID := c.Query("tenantId"); tenantID != "" {
		if id, err := uuid.Parse(tenantID); err == nil {
			return id
		}
	}

	// Try header
	if tenantID := c.Get("X-Tenant-ID"); tenantID != "" {
		if id, err := uuid.Parse(tenantID); err == nil {
			return id
		}
	}

	return uuid.Nil
}

// ErrorHandler returns a Fiber error handler with structured logging
func ErrorHandler(logger interface{}) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := "Internal Server Error"

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = e.Message
		}

		return c.Status(code).JSON(fiber.Map{
			"error":      message,
			"code":       code,
			"request_id": c.Get("X-Request-ID", "unknown"),
		})
	}
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(c *fiber.Ctx) string {
	// Try header first
	if reqID := c.Get("X-Request-ID"); reqID != "" {
		return reqID
	}
	// Try locals
	if reqID := c.Locals("requestid"); reqID != nil {
		if id, ok := reqID.(string); ok {
			return id
		}
	}
	return "unknown"
}

// RequestID is a middleware that adds request ID to context
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Request ID is already added by Fiber's requestid middleware
		// This is just a pass-through for compatibility
		return c.Next()
	}
}

// RequireArtisan returns a middleware that requires artisan role
func RequireArtisan() fiber.Handler {
	return RequireScopes("artisan")
}

// RequireOrganizationContext is a middleware that requires organization context
func RequireOrganizationContext(c *fiber.Ctx) error {
	return RequireOrganization()(c)
}

// GetOrganizationIDFromContext retrieves organization ID from context
func GetOrganizationIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	authCtx := GetAuthContext(c)
	if authCtx == nil {
		return uuid.Nil, fmt.Errorf("no auth context")
	}

	if authCtx.OrganizationID == "" {
		return uuid.Nil, fmt.Errorf("no organization ID in context")
	}

	return uuid.Parse(authCtx.OrganizationID)
}

// VerifyAccessToken is an alias for AuthMiddleware for backward compatibility
func VerifyAccessToken(validator *logto.TokenValidator, config ...MiddlewareConfig) fiber.Handler {
	return AuthMiddleware(validator, config...)
}

// LogAuthContext logs authentication context (for debugging)
func LogAuthContext(c *fiber.Ctx) {
	authCtx := GetAuthContext(c)
	if authCtx != nil {
		log.Infof("Auth Context - Subject: %s, ClientID: %s, OrgID: %s, TenantID: %s, IsM2M: %v, Scopes: %v",
			authCtx.Subject,
			authCtx.ClientID,
			authCtx.OrganizationID,
			authCtx.TenantID,
			authCtx.IsM2M,
			authCtx.Scopes,
		)
	} else {
		log.Info("No auth context found")
	}
}
