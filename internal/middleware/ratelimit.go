package middleware

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"Krafti_Vibe/internal/infrastructure/cache"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	// Redis cache client for distributed rate limiting
	Cache cache.Cache

	// Logger for rate limit events
	Logger *zap.Logger

	// Enabled determines if rate limiting is enabled
	Enabled bool

	// Max number of requests allowed per window
	Max int

	// Duration of the rate limit window
	Window time.Duration

	// KeyGenerator generates the rate limit key from request context
	// Default: IP address
	KeyGenerator func(*fiber.Ctx) string

	// Handler called when rate limit is exceeded
	LimitReachedHandler func(*fiber.Ctx) error

	// SkipFailedRequests determines if failed requests (4xx, 5xx) should not count toward limit
	SkipFailedRequests bool

	// SkipSuccessfulRequests determines if successful requests should not count toward limit
	SkipSuccessfulRequests bool

	// EnableUserRateLimit enables per-user rate limiting (requires auth)
	EnableUserRateLimit bool

	// EnableIPRateLimit enables per-IP rate limiting
	EnableIPRateLimit bool

	// TierLimits defines different rate limits for different user tiers/scopes
	TierLimits map[string]TierLimit

	// DefaultTier is the tier to use when no tier is specified
	DefaultTier string
}

// TierLimit defines rate limit for a specific tier
type TierLimit struct {
	Name        string
	Max         int
	Window      time.Duration
	Description string
}

// DefaultTierLimits returns default rate limit tiers
func DefaultTierLimits() map[string]TierLimit {
	return map[string]TierLimit{
		"free": {
			Name:        "free",
			Max:         100,
			Window:      1 * time.Minute,
			Description: "Free tier: 100 requests per minute",
		},
		"basic": {
			Name:        "basic",
			Max:         1000,
			Window:      1 * time.Minute,
			Description: "Basic tier: 1000 requests per minute",
		},
		"premium": {
			Name:        "premium",
			Max:         5000,
			Window:      1 * time.Minute,
			Description: "Premium tier: 5000 requests per minute",
		},
		"enterprise": {
			Name:        "enterprise",
			Max:         10000,
			Window:      1 * time.Minute,
			Description: "Enterprise tier: 10000 requests per minute",
		},
		"m2m": {
			Name:        "m2m",
			Max:         10000,
			Window:      1 * time.Minute,
			Description: "Machine-to-machine: 10000 requests per minute",
		},
		"admin": {
			Name:        "admin",
			Max:         50000,
			Window:      1 * time.Minute,
			Description: "Admin: 50000 requests per minute",
		},
	}
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig(cache cache.Cache, logger *zap.Logger) RateLimitConfig {
	return RateLimitConfig{
		Cache:                  cache,
		Logger:                 logger,
		Enabled:                true,
		Max:                    100,
		Window:                 1 * time.Minute,
		EnableUserRateLimit:    true,
		EnableIPRateLimit:      true,
		TierLimits:             DefaultTierLimits(),
		DefaultTier:            "free",
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		KeyGenerator:           defaultKeyGenerator,
		LimitReachedHandler:    defaultLimitReachedHandler,
	}
}

// defaultKeyGenerator generates rate limit key from IP address
func defaultKeyGenerator(c *fiber.Ctx) string {
	return c.IP()
}

// defaultLimitReachedHandler returns 429 when rate limit is exceeded
func defaultLimitReachedHandler(c *fiber.Ctx) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
		"error":   "rate_limit_exceeded",
		"message": "Too many requests. Please try again later.",
	})
}

// RateLimiter handles rate limiting with Redis
type RateLimiter struct {
	config RateLimitConfig
	mu     sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	if config.KeyGenerator == nil {
		config.KeyGenerator = defaultKeyGenerator
	}
	if config.LimitReachedHandler == nil {
		config.LimitReachedHandler = defaultLimitReachedHandler
	}
	if config.TierLimits == nil {
		config.TierLimits = DefaultTierLimits()
	}
	if config.DefaultTier == "" {
		config.DefaultTier = "free"
	}

	return &RateLimiter{
		config: config,
	}
}

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(config RateLimitConfig) fiber.Handler {
	limiter := NewRateLimiter(config)

	return func(c *fiber.Ctx) error {
		if !config.Enabled {
			return c.Next()
		}

		// Determine rate limit tier
		tier := limiter.determineTier(c)

		// Check user-based rate limit
		if config.EnableUserRateLimit {
			if exceeded, err := limiter.checkUserRateLimit(c, tier); err != nil {
				config.Logger.Error("failed to check user rate limit",
					zap.Error(err),
					zap.String("path", c.Path()),
				)
			} else if exceeded {
				return config.LimitReachedHandler(c)
			}
		}

		// Check IP-based rate limit
		if config.EnableIPRateLimit {
			if exceeded, err := limiter.checkIPRateLimit(c, tier); err != nil {
				config.Logger.Error("failed to check IP rate limit",
					zap.Error(err),
					zap.String("ip", c.IP()),
				)
			} else if exceeded {
				return config.LimitReachedHandler(c)
			}
		}

		// Increment counter after request completes (if needed)
		if config.SkipFailedRequests || config.SkipSuccessfulRequests {
			// Store initial values for comparison
			c.Locals("rate_limit_tier", tier)

			// Continue request
			err := c.Next()

			// Determine if we should count this request
			statusCode := c.Response().StatusCode()
			shouldSkip := false

			if config.SkipFailedRequests && statusCode >= 400 {
				shouldSkip = true
			}
			if config.SkipSuccessfulRequests && statusCode < 400 {
				shouldSkip = true
			}

			if shouldSkip {
				// Decrement the counter we incremented earlier
				limiter.decrementCounters(c, tier)
			}

			return err
		}

		return c.Next()
	}
}

// determineTier determines the rate limit tier for the request
func (rl *RateLimiter) determineTier(c *fiber.Ctx) TierLimit {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	// Check if user is authenticated
	authCtx, _ := GetAuthContext(c)
	if authCtx == nil {
		// Not authenticated, use default tier
		if tier, ok := rl.config.TierLimits[rl.config.DefaultTier]; ok {
			return tier
		}
		return TierLimit{
			Max:    rl.config.Max,
			Window: rl.config.Window,
		}
	}

	// Check for M2M
	if authCtx.IsM2M {
		if tier, ok := rl.config.TierLimits["m2m"]; ok {
			return tier
		}
	}

	// Check for admin scopes
	if authCtx.HasScope("admin:full") {
		if tier, ok := rl.config.TierLimits["admin"]; ok {
			return tier
		}
	}

	// Check for premium/enterprise scopes
	if authCtx.HasScope("tier:enterprise") {
		if tier, ok := rl.config.TierLimits["enterprise"]; ok {
			return tier
		}
	}
	if authCtx.HasScope("tier:premium") {
		if tier, ok := rl.config.TierLimits["premium"]; ok {
			return tier
		}
	}
	if authCtx.HasScope("tier:basic") {
		if tier, ok := rl.config.TierLimits["basic"]; ok {
			return tier
		}
	}

	// Default to free tier
	if tier, ok := rl.config.TierLimits[rl.config.DefaultTier]; ok {
		return tier
	}

	return TierLimit{
		Max:    rl.config.Max,
		Window: rl.config.Window,
	}
}

// HasScope checks if auth context has a specific scope
func (ac *AuthContext) HasScope(scope string) bool {
	if ac == nil {
		return false
	}
	return slices.Contains(ac.Scopes, scope)
}

// checkUserRateLimit checks if user has exceeded rate limit
func (rl *RateLimiter) checkUserRateLimit(c *fiber.Ctx, tier TierLimit) (bool, error) {
	authCtx, _ := GetAuthContext(c)
	if authCtx == nil {
		return false, nil // Not authenticated, skip user rate limit
	}

	key := fmt.Sprintf("ratelimit:user:%s:%d", authCtx.UserID.String(), time.Now().Unix()/int64(tier.Window.Seconds()))
	return rl.checkLimit(c.Context(), key, tier)
}

// checkIPRateLimit checks if IP has exceeded rate limit
func (rl *RateLimiter) checkIPRateLimit(c *fiber.Ctx, tier TierLimit) (bool, error) {
	ip := rl.config.KeyGenerator(c)
	key := fmt.Sprintf("ratelimit:ip:%s:%d", ip, time.Now().Unix()/int64(tier.Window.Seconds()))
	return rl.checkLimit(c.Context(), key, tier)
}

// checkLimit checks and increments the rate limit counter
func (rl *RateLimiter) checkLimit(ctx context.Context, key string, tier TierLimit) (bool, error) {
	// Increment counter
	count, err := rl.config.Cache.Increment(ctx, key)
	if err != nil {
		// On error, allow request (fail open)
		return false, err
	}

	// Set expiration on first increment
	if count == 1 {
		if err := rl.config.Cache.Expire(ctx, key, tier.Window); err != nil {
			rl.config.Logger.Warn("failed to set rate limit expiration",
				zap.String("key", key),
				zap.Error(err),
			)
		}
	}

	// Check if limit exceeded
	return count > int64(tier.Max), nil
}

// decrementCounters decrements rate limit counters (used when skipping requests)
func (rl *RateLimiter) decrementCounters(c *fiber.Ctx, tier TierLimit) {
	ctx := c.Context()

	// Decrement user counter
	if rl.config.EnableUserRateLimit {
		if authCtx, _ := GetAuthContext(c); authCtx != nil {
			key := fmt.Sprintf("ratelimit:user:%s:%d", authCtx.UserID.String(), time.Now().Unix()/int64(tier.Window.Seconds()))
			if _, err := rl.config.Cache.Decrement(ctx, key); err != nil {
				rl.config.Logger.Warn("failed to decrement user rate limit",
					zap.String("key", key),
					zap.Error(err),
				)
			}
		}
	}

	// Decrement IP counter
	if rl.config.EnableIPRateLimit {
		ip := rl.config.KeyGenerator(c)
		key := fmt.Sprintf("ratelimit:ip:%s:%d", ip, time.Now().Unix()/int64(tier.Window.Seconds()))
		if _, err := rl.config.Cache.Decrement(ctx, key); err != nil {
			rl.config.Logger.Warn("failed to decrement IP rate limit",
				zap.String("key", key),
				zap.Error(err),
			)
		}
	}
}

// RateLimitWithHeaders adds rate limit information to response headers
func RateLimitWithHeaders(config RateLimitConfig) fiber.Handler {
	limiter := NewRateLimiter(config)

	return func(c *fiber.Ctx) error {
		if !config.Enabled {
			return c.Next()
		}

		// Determine tier
		tier := limiter.determineTier(c)

		// Get current usage
		var remaining int64 = int64(tier.Max)
		var resetTime int64

		// Check user rate limit
		if config.EnableUserRateLimit {
			if authCtx, _ := GetAuthContext(c); authCtx != nil {
				windowStart := time.Now().Unix() / int64(tier.Window.Seconds())
				key := fmt.Sprintf("ratelimit:user:%s:%d", authCtx.UserID.String(), windowStart)

				// Get current count
				ctx := c.Context()
				if count, err := limiter.getCurrentCount(ctx, key); err == nil {
					remaining = max(0, int64(tier.Max)-count)
				}

				resetTime = (windowStart + 1) * int64(tier.Window.Seconds())
			}
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", strconv.Itoa(tier.Max))
		c.Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
		c.Set("X-RateLimit-Window", tier.Window.String())

		// Continue with rate limit check
		if config.EnableUserRateLimit {
			if exceeded, err := limiter.checkUserRateLimit(c, tier); err != nil {
				config.Logger.Error("failed to check user rate limit", zap.Error(err))
			} else if exceeded {
				c.Set("Retry-After", strconv.FormatInt(int64(tier.Window.Seconds()), 10))
				return config.LimitReachedHandler(c)
			}
		}

		if config.EnableIPRateLimit {
			if exceeded, err := limiter.checkIPRateLimit(c, tier); err != nil {
				config.Logger.Error("failed to check IP rate limit", zap.Error(err))
			} else if exceeded {
				c.Set("Retry-After", strconv.FormatInt(int64(tier.Window.Seconds()), 10))
				return config.LimitReachedHandler(c)
			}
		}

		return c.Next()
	}
}

// getCurrentCount gets the current rate limit count
func (rl *RateLimiter) getCurrentCount(ctx context.Context, key string) (int64, error) {
	val, err := rl.config.Cache.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// RateLimitByEndpoint creates rate limit middleware for specific endpoints
func RateLimitByEndpoint(cache cache.Cache, logger *zap.Logger, limits map[string]RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()

		// Find matching rate limit config
		for pattern, config := range limits {
			if strings.HasPrefix(path, pattern) {
				handler := RateLimitMiddleware(config)
				return handler(c)
			}
		}

		// No specific limit, continue
		return c.Next()
	}
}

// RateLimitByScope creates rate limit middleware based on user scopes
func RateLimitByScope(cache cache.Cache, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx, _ := GetAuthContext(c)
		if authCtx == nil {
			// Not authenticated, use default rate limit
			config := DefaultRateLimitConfig(cache, logger)
			config.Max = 60
			config.Window = 1 * time.Minute
			handler := RateLimitMiddleware(config)
			return handler(c)
		}

		// Get tier from scopes
		config := DefaultRateLimitConfig(cache, logger)

		// Admin gets highest limits
		if authCtx.HasScope("admin:full") {
			config.Max = 10000
		} else if authCtx.IsM2M {
			config.Max = 5000
		} else if authCtx.HasScope("tier:premium") {
			config.Max = 1000
		} else {
			config.Max = 100
		}

		handler := RateLimitMiddleware(config)
		return handler(c)
	}
}

// GlobalRateLimit creates a global rate limit middleware
func GlobalRateLimit(cache cache.Cache, logger *zap.Logger, max int, window time.Duration) fiber.Handler {
	config := DefaultRateLimitConfig(cache, logger)
	config.Max = max
	config.Window = window

	return RateLimitMiddleware(config)
}

// UserRateLimit creates a per-user rate limit middleware
func UserRateLimit(cache cache.Cache, logger *zap.Logger, max int, window time.Duration) fiber.Handler {
	config := DefaultRateLimitConfig(cache, logger)
	config.Max = max
	config.Window = window
	config.EnableUserRateLimit = true
	config.EnableIPRateLimit = false

	return RateLimitMiddleware(config)
}

// IPRateLimit creates a per-IP rate limit middleware
func IPRateLimit(cache cache.Cache, logger *zap.Logger, max int, window time.Duration) fiber.Handler {
	config := DefaultRateLimitConfig(cache, logger)
	config.Max = max
	config.Window = window
	config.EnableUserRateLimit = false
	config.EnableIPRateLimit = true

	return RateLimitMiddleware(config)
}

// TenantRateLimit creates a per-tenant rate limit middleware
func TenantRateLimit(cache cache.Cache, logger *zap.Logger, max int, window time.Duration) fiber.Handler {
	config := DefaultRateLimitConfig(cache, logger)
	config.Max = max
	config.Window = window
	config.KeyGenerator = func(c *fiber.Ctx) string {
		authCtx, _ := GetAuthContext(c)
		if authCtx != nil && authCtx.TenantID.String() != "" {
			return fmt.Sprintf("tenant:%s", authCtx.TenantID.String())
		}
		return c.IP()
	}

	return RateLimitMiddleware(config)
}

// SlidingWindowRateLimit implements sliding window rate limiting
// Note: This is a simplified version using fixed windows with overlap
func SlidingWindowRateLimit(cache cache.Cache, logger *zap.Logger, max int, window time.Duration) fiber.Handler {
	config := DefaultRateLimitConfig(cache, logger)
	config.Max = max
	config.Window = window

	return func(c *fiber.Ctx) error {
		if !config.Enabled {
			return c.Next()
		}

		authCtx, _ := GetAuthContext(c)
		var identifier string
		if authCtx != nil {
			identifier = authCtx.UserID.String()
		} else {
			identifier = c.IP()
		}

		ctx := c.Context()
		now := time.Now()
		currentWindow := now.Unix() / int64(window.Seconds())
		previousWindow := currentWindow - 1

		// Keys for current and previous windows
		currentKey := fmt.Sprintf("ratelimit:sliding:%s:%d", identifier, currentWindow)
		previousKey := fmt.Sprintf("ratelimit:sliding:%s:%d", identifier, previousWindow)

		// Get counts from both windows
		currentCount, _ := cache.Increment(ctx, currentKey)
		if currentCount == 1 {
			_ = cache.Expire(ctx, currentKey, window*2)
		}

		previousCountStr, err := cache.Get(ctx, previousKey)
		var previousCount int64 = 0
		if err == nil {
			previousCount, _ = strconv.ParseInt(previousCountStr, 10, 64)
		}

		// Calculate weighted count
		// Weight previous window by how much of it overlaps with current sliding window
		elapsedInCurrentWindow := float64(now.Unix() % int64(window.Seconds()))
		weight := 1.0 - (elapsedInCurrentWindow / float64(window.Seconds()))
		weightedCount := int64(float64(previousCount)*weight) + currentCount

		// Check limit
		if weightedCount > int64(max) {
			// Decrement current count since we're rejecting
			_, _ = cache.Decrement(ctx, currentKey)
			return fiber.NewError(fiber.StatusTooManyRequests, "Rate limit exceeded")
		}

		return c.Next()
	}
}

// BurstRateLimit allows bursts up to a maximum, then enforces steady rate
func BurstRateLimit(cache cache.Cache, logger *zap.Logger, burstMax int, steadyRate int, window time.Duration) fiber.Handler {
	// Token bucket algorithm
	return func(c *fiber.Ctx) error {
		authCtx, _ := GetAuthContext(c)
		var identifier string
		if authCtx != nil {
			identifier = authCtx.UserID.String()
		} else {
			identifier = c.IP()
		}

		ctx := c.Context()
		key := fmt.Sprintf("ratelimit:burst:%s", identifier)

		// Get current tokens
		tokensStr, err := cache.Get(ctx, key)
		var tokens float64 = float64(burstMax)

		if err == nil {
			if t, err := strconv.ParseFloat(tokensStr, 64); err == nil {
				tokens = t
			}
		}

		// Refill tokens based on time passed
		now := time.Now()
		lastRefillKey := fmt.Sprintf("ratelimit:burst:lastrefill:%s", identifier)
		if lastRefillStr, err := cache.Get(ctx, lastRefillKey); err == nil {
			if lastRefill, err := strconv.ParseInt(lastRefillStr, 10, 64); err == nil {
				elapsed := now.Unix() - lastRefill
				refillAmount := float64(steadyRate) * (float64(elapsed) / float64(window.Seconds()))
				tokens = min(float64(burstMax), tokens+refillAmount)
			}
		}

		// Check if we have tokens
		if tokens < 1 {
			return fiber.NewError(fiber.StatusTooManyRequests, "Rate limit exceeded - no tokens available")
		}

		// Consume token
		tokens -= 1

		// Save state
		_ = cache.Set(ctx, key, fmt.Sprintf("%.2f", tokens), window*2)
		_ = cache.Set(ctx, lastRefillKey, strconv.FormatInt(now.Unix(), 10), window*2)

		return c.Next()
	}
}
