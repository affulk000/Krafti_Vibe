package middleware

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	zitadelhttp "github.com/zitadel/zitadel-go/v3/pkg/http/middleware"
)

// AuthContextKey is the key used to store authentication context in Fiber locals
const AuthContextKey = "auth_context"

// AuthContext holds the authenticated user information
type AuthContext struct {
	UserID        uuid.UUID
	Email         string
	Username      string
	Roles         []string
	Permissions   []string
	OrgID         string
	ProjectID     string
	TenantID      uuid.UUID
	Scopes        []string
	Claims        map[string]interface{}
	IsM2M         bool
	IntrospectCtx *oauth.IntrospectionContext
}

// UserSyncer defines the interface for syncing users
type UserSyncer interface {
	GetOrSyncUser(ctx context.Context, zitadelUserID string, authCtx *oauth.IntrospectionContext) (interface{}, error)
}

// ZitadelAuthMiddleware wraps the official Zitadel HTTP middleware for Fiber
type ZitadelAuthMiddleware struct {
	mw        *zitadelhttp.Interceptor[*oauth.IntrospectionContext]
	userSyncer UserSyncer
}

// NewZitadelAuthMiddleware creates a new Zitadel authentication middleware using the official package
func NewZitadelAuthMiddleware(
	authZ *authorization.Authorizer[*oauth.IntrospectionContext],
	userSyncer UserSyncer,
) *ZitadelAuthMiddleware {
	return &ZitadelAuthMiddleware{
		mw:        zitadelhttp.New(authZ),
		userSyncer: userSyncer,
	}
}

// RequireAuth creates a Fiber handler that requires authentication using official Zitadel middleware
func (m *ZitadelAuthMiddleware) RequireAuth(opts ...authorization.CheckOption) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create a custom http.ResponseWriter to capture auth failures
		authFailed := false
		rw := &responseWriter{
			onError: func() {
				authFailed = true
			},
		}

		// Create http.Request from Fiber context
		req, err := http.NewRequest(
			c.Method(),
			c.OriginalURL(),
			nil,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to process request",
				},
			})
		}

		// Copy headers from Fiber to http.Request
		c.Request().Header.VisitAll(func(key, value []byte) {
			req.Header.Add(string(key), string(value))
		})

		// Create a done channel to signal when the handler completes
		done := make(chan struct{})
		var finalReq *http.Request

		// Wrap with the official Zitadel middleware
		handler := m.mw.RequireAuthorization(opts...)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			finalReq = r
			close(done)
		}))

		// Execute the middleware chain
		handler.ServeHTTP(rw, req)

		// Wait for completion or check if auth failed
		select {
		case <-done:
			// Auth successful, extract context
		default:
			// Auth failed
			if authFailed || rw.statusCode == http.StatusUnauthorized || rw.statusCode == http.StatusForbidden {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"error": fiber.Map{
						"code":    "INVALID_TOKEN",
						"message": "Invalid or expired token",
					},
				})
			}
		}

		// Extract the introspection context using the official middleware's Context method
		var authCtx *oauth.IntrospectionContext
		if finalReq != nil {
			authCtx = m.mw.Context(finalReq.Context())
		}

		if authCtx == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INVALID_TOKEN",
					"message": "Invalid or expired token",
				},
			})
		}

		// Extract user information
		userIDStr := authCtx.UserID()
		if userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INVALID_TOKEN",
					"message": "Invalid token: no user ID",
				},
			})
		}

		// Parse user ID as UUID
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			userID = uuid.Nil
		}

		// Parse organization ID as tenant ID
		orgID := authCtx.OrganizationID()
		tenantID, err := uuid.Parse(orgID)
		if err != nil && orgID != "" {
			tenantID = uuid.Nil
		}

		// Sync user from Zitadel to database (if syncer is configured)
		var dbUser interface{}
		if m.userSyncer != nil {
			var err error
			dbUser, err = m.userSyncer.GetOrSyncUser(c.Context(), userIDStr, authCtx)
			if err != nil {
				// Log error but don't fail authentication
				// User can still use the API with Zitadel identity
				c.Locals("user_sync_error", err.Error())
				// Continue - authentication succeeded, sync failure is not critical
			} else if dbUser != nil {
				c.Locals("user_synced", true)
			}
		}

		// Build auth context
		authContext := &AuthContext{
			UserID:        userID,
			Email:         authCtx.Email,
			Username:      authCtx.Username,
			Roles:         extractRoles(authCtx),
			Permissions:   extractPermissions(authCtx),
			OrgID:         orgID,
			TenantID:      tenantID,
			Scopes:        []string(authCtx.Scope),
			Claims:        extractClaims(authCtx.Claims),
			IsM2M:         authCtx.Subject == authCtx.ClientID,
			IntrospectCtx: authCtx,
		}

		// Store auth context in Fiber locals
		c.Locals(AuthContextKey, authContext)

		// Store database user if available
		if dbUser != nil {
			c.Locals("db_user", dbUser)
		}

		return c.Next()
	}
}

// RequireRole creates a middleware that requires specific roles
func (m *ZitadelAuthMiddleware) RequireRole(roles ...string) fiber.Handler {
	var opts []authorization.CheckOption
	for _, role := range roles {
		opts = append(opts, authorization.WithRole(role))
	}
	return m.RequireAuth(opts...)
}

// GetAuthContext retrieves the authentication context from Fiber locals
func GetAuthContext(c *fiber.Ctx) (*AuthContext, bool) {
	val := c.Locals(AuthContextKey)
	if val == nil {
		return nil, false
	}
	authCtx, ok := val.(*AuthContext)
	return authCtx, ok
}

// MustGetAuthContext retrieves the authentication context or panics if not found
func MustGetAuthContext(c *fiber.Ctx) *AuthContext {
	authCtx, ok := GetAuthContext(c)
	if !ok {
		panic("authentication context not found - did you forget to use the auth middleware?")
	}
	return authCtx
}

// responseWriter implements http.ResponseWriter to capture responses
type responseWriter struct {
	header     http.Header
	statusCode int
	body       []byte
	onError    func()
}

func (w *responseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return len(b), nil
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	if (statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden) && w.onError != nil {
		w.onError()
	}
}

// Helper functions

func extractRoles(ctx *oauth.IntrospectionContext) []string {
	if ctx == nil || ctx.Claims == nil {
		return []string{}
	}

	roles := []string{}
	if rolesRaw, ok := ctx.Claims["urn:zitadel:iam:org:project:roles"]; ok {
		if rolesMap, ok := rolesRaw.(map[string]interface{}); ok {
			for role := range rolesMap {
				roles = append(roles, role)
			}
		}
	}
	return roles
}

func extractPermissions(ctx *oauth.IntrospectionContext) []string {
	if ctx == nil {
		return []string{}
	}

	permissions := []string{}
	if len(ctx.Scope) > 0 {
		permissions = append(permissions, []string(ctx.Scope)...)
	}
	return permissions
}

func extractClaims(claims map[string]interface{}) map[string]interface{} {
	if claims == nil {
		return make(map[string]interface{})
	}
	return claims
}

// Context returns the introspection context from the request context
func (m *ZitadelAuthMiddleware) Context(ctx context.Context) *oauth.IntrospectionContext {
	return m.mw.Context(ctx)
}
