# Rate Limiting Middleware

Comprehensive rate limiting middleware for the Krafti Vibe API with Redis-backed distributed rate limiting, tier-based limits, and multiple rate limiting strategies.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Rate Limit Tiers](#rate-limit-tiers)
- [Middleware Options](#middleware-options)
- [Usage Examples](#usage-examples)
- [Headers](#headers)
- [Testing](#testing)
- [Best Practices](#best-practices)

## Overview

The rate limiting middleware provides multiple strategies to protect your API from abuse and ensure fair usage across all users:

- **Redis-backed** - Distributed rate limiting across multiple servers
- **Tier-based** - Different limits for different user tiers (free, basic, premium, enterprise)
- **Multi-strategy** - Fixed window, sliding window, and token bucket algorithms
- **Flexible** - Per-user, per-IP, per-tenant, or custom key-based limiting
- **Headers** - Standard rate limit headers in responses
- **Graceful** - Fails open if Redis is unavailable

## Features

### Rate Limiting Strategies

1. **Fixed Window** - Simple counter reset at fixed intervals
2. **Sliding Window** - Weighted average of current and previous windows
3. **Token Bucket** - Allow bursts with steady refill rate

### Identification Methods

- **User-based** - Rate limit authenticated users by user ID
- **IP-based** - Rate limit by client IP address
- **Tenant-based** - Rate limit by tenant/organization
- **Custom** - Custom key generation function

### Tier System

- **Free** - 100 requests/minute
- **Basic** - 1,000 requests/minute
- **Premium** - 5,000 requests/minute
- **Enterprise** - 10,000 requests/minute
- **M2M** - 10,000 requests/minute (machine-to-machine)
- **Admin** - 50,000 requests/minute

## Quick Start

### Basic Usage

```go
package main

import (
    "time"

    "Krafti_Vibe/internal/infrastructure/cache"
    "Krafti_Vibe/internal/middleware"

    "github.com/gofiber/fiber/v2"
    "go.uber.org/zap"
)

func main() {
    app := fiber.New()

    // Initialize Redis cache
    redisClient, _ := cache.NewRedisClient(cache.RedisConfig{
        Host: "localhost",
        Port: "6379",
    }, logger)

    // Create logger
    logger, _ := zap.NewProduction()

    // Apply global rate limiting
    app.Use(middleware.GlobalRateLimit(redisClient, logger, 100, 1*time.Minute))

    app.Listen(":8080")
}
```

### With Authentication (Recommended)

```go
// Apply rate limiting that respects user tiers
api := app.Group("/api/v1")

// Rate limit based on authenticated user's tier/scopes
api.Use(middleware.RateLimitByScope(redisClient, logger))

// User routes
api.Get("/users", getUsers)
```

## Configuration

### RateLimitConfig

```go
type RateLimitConfig struct {
    // Required
    Cache   cache.Cache   // Redis client
    Logger  *zap.Logger   // Structured logger

    // Basic settings
    Enabled bool          // Enable/disable rate limiting
    Max     int           // Max requests per window
    Window  time.Duration // Time window for rate limit

    // Advanced settings
    EnableUserRateLimit        bool            // Enable per-user limiting
    EnableIPRateLimit          bool            // Enable per-IP limiting
    TierLimits                 map[string]TierLimit
    DefaultTier                string
    SkipFailedRequests         bool            // Don't count 4xx/5xx requests
    SkipSuccessfulRequests     bool            // Don't count successful requests

    // Custom functions
    KeyGenerator            func(*fiber.Ctx) string
    LimitReachedHandler     func(*fiber.Ctx) error
}
```

### Default Configuration

```go
config := middleware.DefaultRateLimitConfig(cache, logger)
// Defaults:
// - Max: 100 requests
// - Window: 1 minute
// - EnableUserRateLimit: true
// - EnableIPRateLimit: true
// - DefaultTier: "free"
```

## Rate Limit Tiers

### Default Tiers

```go
tiers := middleware.DefaultTierLimits()

// Free tier: 100 requests/minute
tiers["free"] = TierLimit{
    Name:   "free",
    Max:    100,
    Window: 1 * time.Minute,
}

// Premium tier: 5000 requests/minute
tiers["premium"] = TierLimit{
    Name:   "premium",
    Max:    5000,
    Window: 1 * time.Minute,
}
```

### Tier Detection

Tiers are automatically detected based on:

1. **M2M Authentication** - IsM2M flag → "m2m" tier
2. **Admin Scopes** - "admin:full" scope → "admin" tier
3. **Tier Scopes** - "tier:enterprise", "tier:premium", "tier:basic"
4. **Default** - "free" tier for unauthenticated or basic users

### Custom Tiers

```go
config := middleware.DefaultRateLimitConfig(cache, logger)
config.TierLimits["vip"] = middleware.TierLimit{
    Name:        "vip",
    Max:         100000,
    Window:      1 * time.Minute,
    Description: "VIP: 100K requests/minute",
}
```

## Middleware Options

### 1. Global Rate Limit

Simple rate limit for all requests:

```go
// Limit all requests to 100/minute
app.Use(middleware.GlobalRateLimit(cache, logger, 100, 1*time.Minute))
```

### 2. User Rate Limit

Per-user rate limiting (requires authentication):

```go
// Limit each authenticated user to 1000/minute
app.Use(middleware.UserRateLimit(cache, logger, 1000, 1*time.Minute))
```

### 3. IP Rate Limit

Per-IP rate limiting:

```go
// Limit each IP to 60/minute
app.Use(middleware.IPRateLimit(cache, logger, 60, 1*time.Minute))
```

### 4. Tenant Rate Limit

Per-tenant/organization rate limiting:

```go
// Limit each tenant to 10000/minute
app.Use(middleware.TenantRateLimit(cache, logger, 10000, 1*time.Minute))
```

### 5. Scope-Based Rate Limit

Automatic tier detection from user scopes:

```go
// Automatically applies appropriate tier based on user's scopes
app.Use(middleware.RateLimitByScope(cache, logger))
```

### 6. Endpoint-Specific Rate Limits

Different limits for different endpoints:

```go
limits := map[string]middleware.RateLimitConfig{
    "/api/v1/users":    {Max: 100, Window: 1 * time.Minute},
    "/api/v1/bookings": {Max: 500, Window: 1 * time.Minute},
    "/api/v1/payments": {Max: 50,  Window: 1 * time.Minute},
}

app.Use(middleware.RateLimitByEndpoint(cache, logger, limits))
```

### 7. Sliding Window Rate Limit

More accurate rate limiting using sliding windows:

```go
// Weighted average of current and previous windows
app.Use(middleware.SlidingWindowRateLimit(cache, logger, 1000, 1*time.Minute))
```

### 8. Burst Rate Limit

Token bucket algorithm allowing bursts:

```go
// Allow bursts up to 100, refill at 50/minute
app.Use(middleware.BurstRateLimit(cache, logger, 100, 50, 1*time.Minute))
```

### 9. Rate Limit with Headers

Adds standard rate limit headers to responses:

```go
config := middleware.DefaultRateLimitConfig(cache, logger)
app.Use(middleware.RateLimitWithHeaders(config))
```

## Usage Examples

### Example 1: Basic API with Tiered Rate Limiting

```go
package main

import (
    "time"

    "Krafti_Vibe/internal/infrastructure/cache"
    "Krafti_Vibe/internal/middleware"

    "github.com/gofiber/fiber/v2"
    "go.uber.org/zap"
)

func main() {
    app := fiber.New()
    logger, _ := zap.NewProduction()

    // Initialize Redis
    redisClient, _ := cache.NewRedisClient(cache.RedisConfig{
        Host: "localhost",
        Port: "6379",
    }, logger)

    // Public endpoints - strict rate limit
    public := app.Group("/public")
    public.Use(middleware.IPRateLimit(redisClient, logger, 60, 1*time.Minute))
    public.Get("/health", healthCheck)

    // API endpoints - tier-based rate limiting
    api := app.Group("/api/v1")
    api.Use(middleware.AuthMiddleware(...))  // Auth first
    api.Use(middleware.RateLimitByScope(redisClient, logger))  // Then rate limit

    api.Get("/users", getUsers)
    api.Post("/users", createUser)

    app.Listen(":8080")
}
```

### Example 2: Custom Rate Limit Handler

```go
config := middleware.DefaultRateLimitConfig(cache, logger)
config.LimitReachedHandler = func(c *fiber.Ctx) error {
    return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
        "error":       "rate_limit_exceeded",
        "message":     "You have exceeded the rate limit. Please try again later.",
        "retry_after": c.Get("Retry-After"),
        "tier":        "free",
        "upgrade_url": "https://kraftivibe.com/pricing",
    })
}

app.Use(middleware.RateLimitMiddleware(config))
```

### Example 3: Custom Key Generator

```go
config := middleware.DefaultRateLimitConfig(cache, logger)

// Rate limit by API key instead of IP/user
config.KeyGenerator = func(c *fiber.Ctx) string {
    apiKey := c.Get("X-API-Key")
    if apiKey != "" {
        return "apikey:" + apiKey
    }
    return c.IP()
}

app.Use(middleware.RateLimitMiddleware(config))
```

### Example 4: Skip Failed Requests

```go
config := middleware.DefaultRateLimitConfig(cache, logger)

// Don't count 4xx/5xx errors toward rate limit
// Useful to prevent malicious requests from consuming quota
config.SkipFailedRequests = true

app.Use(middleware.RateLimitMiddleware(config))
```

### Example 5: Multiple Rate Limits

```go
// Apply multiple rate limiters in sequence
api := app.Group("/api/v1")

// 1. Global IP-based limit (prevents IP flooding)
api.Use(middleware.IPRateLimit(cache, logger, 1000, 1*time.Minute))

// 2. User-based limit (after auth)
api.Use(middleware.AuthMiddleware(...))
api.Use(middleware.UserRateLimit(cache, logger, 500, 1*time.Minute))

// 3. Endpoint-specific limits
payments := api.Group("/payments")
payments.Use(middleware.GlobalRateLimit(cache, logger, 50, 1*time.Minute))
```

## Headers

The rate limiting middleware adds standard headers to responses:

### Response Headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640000000
X-RateLimit-Window: 1m0s
```

When rate limit is exceeded:

```http
HTTP/1.1 429 Too Many Requests
Retry-After: 60
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1640000060
```

### Header Descriptions

| Header | Description | Example |
|--------|-------------|---------|
| `X-RateLimit-Limit` | Maximum requests allowed in window | 100 |
| `X-RateLimit-Remaining` | Requests remaining in current window | 95 |
| `X-RateLimit-Reset` | Unix timestamp when limit resets | 1640000000 |
| `X-RateLimit-Window` | Duration of rate limit window | 1m0s |
| `Retry-After` | Seconds to wait before retrying (429 only) | 60 |

## Testing

### Unit Testing

```go
func TestRateLimiting(t *testing.T) {
    // Create mock cache
    mockCache := setupMockCache()
    logger, _ := zap.NewDevelopment()

    app := fiber.New()
    app.Use(middleware.GlobalRateLimit(mockCache, logger, 5, 1*time.Minute))

    app.Get("/test", func(c *fiber.Ctx) error {
        return c.SendString("OK")
    })

    // Should allow 5 requests
    for i := 0; i < 5; i++ {
        req := httptest.NewRequest("GET", "/test", nil)
        resp, _ := app.Test(req)
        assert.Equal(t, 200, resp.StatusCode)
    }

    // 6th request should be rate limited
    req := httptest.NewRequest("GET", "/test", nil)
    resp, _ := app.Test(req)
    assert.Equal(t, 429, resp.StatusCode)
}
```

### Integration Testing

```go
func TestTierRateLimiting(t *testing.T) {
    // Test with real Redis
    redisClient := setupRedisClient()
    defer redisClient.Close()

    logger, _ := zap.NewDevelopment()

    app := fiber.New()
    app.Use(middleware.RateLimitByScope(redisClient, logger))

    // Test free tier (100/min)
    testUserRequests(t, app, "free_user_token", 100, true)
    testUserRequests(t, app, "free_user_token", 1, false) // 101st should fail

    // Test premium tier (5000/min)
    testUserRequests(t, app, "premium_user_token", 5000, true)
}
```

### Load Testing

```bash
# Test rate limiting with siege
siege -c 10 -r 100 http://localhost:8080/api/v1/users

# Test rate limiting with wrk
wrk -t 10 -c 100 -d 60s http://localhost:8080/api/v1/users

# Monitor Redis
redis-cli --stat
```

## Best Practices

### 1. Use Multiple Layers

Apply multiple rate limits for defense in depth:

```go
// Layer 1: Global IP-based (prevent DDoS)
app.Use(middleware.IPRateLimit(cache, logger, 10000, 1*time.Minute))

// Layer 2: User-based (fair usage)
app.Use(middleware.UserRateLimit(cache, logger, 1000, 1*time.Minute))

// Layer 3: Endpoint-specific (protect expensive operations)
payments.Use(middleware.GlobalRateLimit(cache, logger, 50, 1*time.Minute))
```

### 2. Set Appropriate Limits

```go
// Public endpoints - strict
public.Use(middleware.IPRateLimit(cache, logger, 60, 1*time.Minute))

// Read operations - generous
reads.Use(middleware.UserRateLimit(cache, logger, 1000, 1*time.Minute))

// Write operations - moderate
writes.Use(middleware.UserRateLimit(cache, logger, 100, 1*time.Minute))

// Expensive operations - strict
reports.Use(middleware.UserRateLimit(cache, logger, 10, 1*time.Minute))
```

### 3. Monitor and Alert

```go
// Log rate limit events
config.LimitReachedHandler = func(c *fiber.Ctx) error {
    logger.Warn("rate limit exceeded",
        zap.String("ip", c.IP()),
        zap.String("path", c.Path()),
        zap.String("user_id", getUserID(c)),
    )

    // Optionally trigger alerts for repeated violations
    checkForAbuse(c)

    return c.Status(429).JSON(fiber.Map{
        "error": "rate_limit_exceeded",
    })
}
```

### 4. Use Sliding Windows for Accuracy

```go
// More accurate than fixed windows, prevents burst at window boundaries
app.Use(middleware.SlidingWindowRateLimit(cache, logger, 1000, 1*time.Minute))
```

### 5. Handle Redis Failures Gracefully

The middleware fails open by default (allows requests if Redis is unavailable). Monitor Redis health:

```go
// Health check
app.Get("/health", func(c *fiber.Ctx) error {
    if err := redisClient.Ping(c.Context()); err != nil {
        return c.Status(503).JSON(fiber.Map{
            "status": "unhealthy",
            "redis":  "unavailable",
        })
    }
    return c.JSON(fiber.Map{"status": "healthy"})
})
```

### 6. Document Rate Limits

Provide clear documentation to API consumers:

```go
app.Get("/api/rate-limits", func(c *fiber.Ctx) error {
    return c.JSON(middleware.DefaultTierLimits())
})
```

### 7. Use Burst Limiting for Spiky Traffic

```go
// Allow bursts up to 500, refill at 100/minute
app.Use(middleware.BurstRateLimit(cache, logger, 500, 100, 1*time.Minute))
```

## Troubleshooting

### Common Issues

1. **Rate limit not working**
   - Check Redis connection: `redis-cli ping`
   - Verify cache is passed to middleware
   - Check rate limit headers in response

2. **Too many rate limit rejections**
   - Review limits for your tier
   - Check if multiple limiters are applied
   - Monitor Redis keys: `redis-cli --scan --pattern ratelimit:*`

3. **Inconsistent behavior across servers**
   - Ensure all servers use same Redis instance
   - Verify system clocks are synchronized (NTP)

4. **Memory issues in Redis**
   - Rate limit keys auto-expire
   - Monitor: `redis-cli INFO memory`
   - Increase Redis memory or reduce TTL

### Debug Mode

```go
logger, _ := zap.NewDevelopment()  // Development mode shows debug logs
config := middleware.DefaultRateLimitConfig(cache, logger)
```

## Related Documentation

- [Redis Cache](../infrastructure/cache/README.md)
- [Authentication Middleware](./middleware.go)
- [Router Integration](../router/README.md)
