# Router Package

The router package provides route registration and dependency injection for the Krafti Vibe API, with integrated Logto authentication and authorization.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Logto Integration](#logto-integration)
- [Authentication Flow](#authentication-flow)
- [Route Protection](#route-protection)
- [Usage](#usage)
- [Configuration](#configuration)
- [Scopes and Permissions](#scopes-and-permissions)
- [Examples](#examples)

## Overview

The router package serves as the entry point for all HTTP routes in the application. It:

- Initializes and coordinates all feature-specific routes
- Integrates Logto JWT authentication and authorization
- Manages dependency injection for services and handlers
- Provides scope-based access control
- Supports multi-tenancy and organization context

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        main.go                              │
│  - Initialize Logto config and JWKS cache                  │
│  - Create token validator                                  │
│  - Setup database connection                               │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                    Router (router.go)                       │
│  - Accepts Fiber app and config (with Logto)               │
│  - Initializes repositories                                │
│  - Delegates to feature-specific route setup               │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              Feature Routes (user_routes.go)                │
│  - Initialize feature services                             │
│  - Initialize feature handlers                             │
│  - Register routes with auth middleware                    │
│  - Apply scope-based access control                        │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                   Handlers (handler/)                       │
│  - Extract auth context from middleware                    │
│  - Call service layer with user context                    │
│  - Return HTTP responses                                   │
└─────────────────────────────────────────────────────────────┘
```

## Logto Integration

### Components

The router integrates with Logto using the following components:

1. **Logto Config** (`logto.Config`)
   - Endpoint, App ID, App Secret
   - JWKS endpoint and cache settings
   - API resource indicator (audience)
   - Organization and RBAC settings

2. **JWKS Cache** (`logto.JWKSCache`)
   - Caches public keys for JWT verification
   - Auto-refreshes periodically
   - Reduces latency and external API calls

3. **Token Validator** (`logto.TokenValidator`)
   - Extracts JWT from Authorization header
   - Validates token signature using JWKS
   - Verifies issuer, audience, and expiration
   - Extracts claims (user ID, scopes, organization)

4. **Auth Middleware** (`middleware.AuthMiddleware`)
   - Validates JWT on every request
   - Populates auth context in Fiber locals
   - Enforces audience and scope requirements
   - Returns 401/403 for unauthorized requests

### Flow

```
┌──────────┐     ┌──────────────┐     ┌───────────────┐     ┌─────────┐
│  Client  │────▶│  Middleware  │────▶│  Token        │────▶│  JWKS   │
│          │     │              │     │  Validator    │     │  Cache  │
└──────────┘     └──────────────┘     └───────────────┘     └─────────┘
    │                   │                      │                   │
    │ 1. Send request   │                      │                   │
    │    with Bearer    │                      │                   │
    │    token          │                      │                   │
    │                   │ 2. Extract token     │                   │
    │                   ├─────────────────────▶│                   │
    │                   │                      │ 3. Fetch JWKS     │
    │                   │                      ├──────────────────▶│
    │                   │                      │ 4. Cached keys    │
    │                   │                      │◀──────────────────┤
    │                   │                      │                   │
    │                   │                      │ 5. Verify token   │
    │                   │                      │    signature      │
    │                   │ 6. Valid token       │                   │
    │                   │◀─────────────────────┤                   │
    │                   │                      │                   │
    │                   │ 7. Check audience    │                   │
    │                   │    and scopes        │                   │
    │                   │                      │                   │
    │                   │ 8. Populate auth     │                   │
    │                   │    context           │                   │
    │                   │                      │                   │
    │ 9. Request        │ 10. Call handler     │                   │
    │    proceeds       ├─────────────────────▶│                   │
    │                   │                      │                   │
```

## Authentication Flow

### 1. Public Routes (No Auth)

```go
// No authentication required
users.Post("/password-reset", userHandler.ResetPassword)
```

### 2. Protected Routes (Auth Required)

```go
// Base authentication - validates token and audience
authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
    RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
})

users.Get("/:id",
    authMiddleware,  // Validates JWT
    middleware.RequireScopes(r.scopes.UserRead),  // Validates scope
    userHandler.GetUser,
)
```

### 3. Auth Context in Handlers

```go
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
    // Get authenticated user context
    authCtx := middleware.MustGetAuthContext(c)

    // Access user information
    userID := authCtx.UserID           // UUID of authenticated user
    tenantID := authCtx.TenantID       // UUID of tenant/organization
    scopes := authCtx.Scopes           // Granted scopes
    isM2M := authCtx.IsM2M            // Machine-to-machine flag

    // Use in service calls
    user, err := h.userService.GetUser(c.Context(), requestedUserID, authCtx.UserID)
    // ...
}
```

## Route Protection

### Scope-Based Access Control

All protected routes require specific scopes:

```go
// Read operations - require "user:read" scope
users.Get("/:id", authMiddleware, middleware.RequireScopes(r.scopes.UserRead), ...)

// Write operations - require "user:write" scope
users.Post("/", authMiddleware, middleware.RequireScopes(r.scopes.UserWrite), ...)

// Delete operations - require "user:delete" scope
users.Delete("/:id", authMiddleware, middleware.RequireScopes(r.scopes.UserDelete), ...)

// Admin operations - require "user:manage" scope
users.Put("/:id/role", authMiddleware, middleware.RequireScopes(r.scopes.UserManage), ...)
```

### Additional Middleware

```go
// Require tenant context
middleware.RequireTenant()

// Require organization context
middleware.RequireOrganization()

// Require M2M authentication
middleware.RequireM2M()

// Require specific audience
middleware.RequireAudience("https://api.kraftivibe.com")
```

## Usage

### Basic Setup

```go
package main

import (
    "Krafti_Vibe/internal/infrastructure/logto"
    "Krafti_Vibe/internal/router"

    "github.com/gofiber/fiber/v2"
)

func main() {
    // 1. Initialize Logto config
    logtoConfig, _ := logto.LoadConfig(
        "https://your-tenant.logto.app",
        "your-app-id",
        "your-app-secret",
    )
    logtoConfig.APIResourceIndicator = "https://api.kraftivibe.com"

    // 2. Initialize JWKS cache
    jwksCache, _ := logto.NewJWKSCache(
        logtoConfig.JWKSEndpoint,
        logtoConfig.JWKSCacheTTL,
    )

    // 3. Create token validator
    tokenValidator := logto.NewTokenValidator(jwksCache, logtoConfig.Issuer)

    // 4. Initialize database (not shown)
    db := initDatabase()

    // 5. Create Fiber app
    app := fiber.New()

    // 6. Create router config
    routerConfig := &router.Config{
        DB:             db,
        Logger:         logger,
        LogtoConfig:    logtoConfig,
        TokenValidator: tokenValidator,
    }

    // 7. Initialize and setup router
    r := router.New(app, routerConfig)
    r.Setup()

    // 8. Start server
    app.Listen(":8080")
}
```

See [integration_example.go](./integration_example.go) for a complete example.

## Configuration

### Router Config

```go
type Config struct {
    DB             *gorm.DB                 // Database connection
    Logger         log.AllLogger            // Application logger
    LogtoConfig    *logto.Config            // Logto configuration
    TokenValidator *logto.TokenValidator    // JWT token validator
}
```

### Environment Variables

```bash
# Logto
LOGTO_ENDPOINT=https://your-tenant.logto.app
LOGTO_APP_ID=your-app-id
LOGTO_APP_SECRET=your-app-secret

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=kraftivibe
DB_PASSWORD=your-password
DB_NAME=kraftivibe

# Server
PORT=8080
CORS_ORIGINS=http://localhost:3000
```

## Scopes and Permissions

### User Management Scopes

| Scope | Description | Endpoints |
|-------|-------------|-----------|
| `user:read` | Read user information | GET /users, GET /users/:id |
| `user:write` | Create and update users | POST /users, PUT /users/:id |
| `user:delete` | Delete users | DELETE /users/:id |
| `user:manage` | Full user management | PUT /users/:id/role, POST /users/:id/unlock |

### Role-to-Scope Mappings

Configured in Logto dashboard:

- **Platform Admin**: `admin:full`, `user:manage`, `tenant:admin`
- **Tenant Admin**: `user:manage`, `booking:manage`, `service:write`
- **Artisan**: `service:read`, `service:write`, `booking:read`, `booking:write`
- **Customer**: `service:read`, `booking:read`, `booking:write`, `review:write`

### Custom Scopes

Define additional scopes in `logto.DefaultScopes()`:

```go
ServiceRead:   "service:read"
ServiceWrite:  "service:write"
BookingManage: "booking:manage"
PaymentProcess: "payment:process"
// ... etc
```

## Examples

### Making Authenticated Requests

```bash
# 1. Obtain access token from Logto (via your app or OAuth flow)
TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."

# 2. Make request with Bearer token
curl -X GET http://localhost:8080/api/v1/users/123 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"

# 3. Middleware validates:
#    - Token signature
#    - Issuer and audience
#    - Required scopes (user:read)
#    - Token expiration
```

### Testing with Invalid Token

```bash
# Missing token
curl -X GET http://localhost:8080/api/v1/users
# Response: 401 Unauthorized - "Missing or invalid authorization header"

# Invalid token
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer invalid-token"
# Response: 401 Unauthorized - "Invalid token"

# Valid token but insufficient scopes
curl -X DELETE http://localhost:8080/api/v1/users/123 \
  -H "Authorization: Bearer $TOKEN_WITHOUT_DELETE_SCOPE"
# Response: 403 Forbidden - "Insufficient permissions. Required scopes: user:delete"
```

### Adding New Feature Routes

1. Create handler in `internal/handler/`
2. Create service in `internal/service/`
3. Create route setup file:

```go
// internal/router/booking_routes.go
package router

import (
    "Krafti_Vibe/internal/handler"
    "Krafti_Vibe/internal/middleware"
    "Krafti_Vibe/internal/service"

    "github.com/gofiber/fiber/v2"
)

func (r *Router) setupBookingRoutes(api fiber.Router) {
    bookingService := service.NewBookingService(r.repos, r.config.Logger)
    bookingHandler := handler.NewBookingHandler(bookingService)

    bookings := api.Group("/bookings")

    authMiddleware := middleware.AuthMiddleware(r.tokenValidator, middleware.MiddlewareConfig{
        RequiredAudience: r.config.LogtoConfig.APIResourceIndicator,
    })

    bookings.Get("/",
        authMiddleware,
        middleware.RequireScopes(r.scopes.BookingRead),
        bookingHandler.ListBookings,
    )

    bookings.Post("/",
        authMiddleware,
        middleware.RequireScopes(r.scopes.BookingWrite),
        bookingHandler.CreateBooking,
    )

    r.config.Logger.Info("booking routes registered successfully")
}
```

4. Register in `router.go`:

```go
func (r *Router) setupAPIRoutes() {
    api := r.app.Group("/api/v1")

    r.setupUserRoutes(api)
    r.setupBookingRoutes(api)  // Add this
}
```

## Testing

### Unit Testing Routes

```go
func TestUserRoutes(t *testing.T) {
    app := fiber.New()

    // Mock dependencies
    db := setupTestDB()
    logger := setupTestLogger()
    tokenValidator := setupMockTokenValidator()

    // Setup router
    config := &router.Config{
        DB:             db,
        Logger:         logger,
        LogtoConfig:    testLogtoConfig(),
        TokenValidator: tokenValidator,
    }

    r := router.New(app, config)
    r.Setup()

    // Test request
    req := httptest.NewRequest("GET", "/api/v1/users", nil)
    req.Header.Set("Authorization", "Bearer mock-token")

    resp, _ := app.Test(req)
    assert.Equal(t, 200, resp.StatusCode)
}
```

### Integration Testing with Logto

See the test files in `internal/router/*_test.go` for examples.

## Troubleshooting

### Common Issues

1. **401 Unauthorized - "Missing or invalid authorization header"**
   - Ensure Authorization header is present
   - Format: `Authorization: Bearer <token>`

2. **401 Unauthorized - "Invalid token"**
   - Token signature verification failed
   - Check JWKS endpoint is accessible
   - Verify token was issued by correct Logto tenant

3. **403 Forbidden - "Invalid audience"**
   - Token audience doesn't match API resource indicator
   - Verify `APIResourceIndicator` in config matches token `aud` claim

4. **403 Forbidden - "Insufficient permissions"**
   - Token doesn't have required scopes
   - Check role-to-scope mappings in Logto dashboard
   - Verify user has correct role assigned

### Debug Logging

Enable auth context logging:

```go
app.Use(func(c *fiber.Ctx) error {
    defer middleware.LogAuthContext(c)
    return c.Next()
})
```

Output:
```
Auth Context - Subject: a1b2c3d4-..., ClientID: app123, OrgID: org456, TenantID: tenant789, IsM2M: false, Scopes: [user:read user:write]
```

## Security Best Practices

1. **Always validate audience** - Prevents token reuse across different APIs
2. **Use specific scopes** - Principle of least privilege
3. **Rotate JWKS regularly** - Configure appropriate cache TTL
4. **Enable HTTPS in production** - Protect tokens in transit
5. **Set proper CORS origins** - Prevent unauthorized origins
6. **Monitor failed auth attempts** - Detect potential attacks
7. **Use short token expiration** - Reduce impact of token compromise

## Related Documentation

- [Logto Documentation](https://docs.logto.io)
- [Logto RBAC Guide](https://docs.logto.io/docs/recipes/rbac/)
- [Integration Example](./integration_example.go)
- [Middleware Package](../middleware/README.md)
- [Handler Package](../handler/README.md)
