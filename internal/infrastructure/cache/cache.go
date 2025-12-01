package cache

import (
	"context"
	"errors"
	"time"
)

// Common cache errors
var (
	ErrCacheMiss      = errors.New("cache: key not found")
	ErrCacheSetFailed = errors.New("cache: failed to set value")
	ErrCacheDelFailed = errors.New("cache: failed to delete key")
	ErrInvalidTTL     = errors.New("cache: invalid TTL value")
)

// Cache defines the interface for cache operations
type Cache interface {
	// Get retrieves a value by key
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value with TTL
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// GetJSON retrieves and unmarshals a JSON value
	GetJSON(ctx context.Context, key string, dest any) error

	// SetJSON marshals and stores a JSON value
	SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error

	// Delete removes one or more keys
	Delete(ctx context.Context, keys ...string) error

	// DeletePattern removes all keys matching a pattern
	DeletePattern(ctx context.Context, pattern string) error

	// Exists checks if keys exist
	Exists(ctx context.Context, keys ...string) (int64, error)

	// Expire sets TTL on an existing key
	Expire(ctx context.Context, key string, ttl time.Duration) error

	// Increment increments a counter
	Increment(ctx context.Context, key string) (int64, error)

	// IncrementBy increments a counter by a specific amount
	IncrementBy(ctx context.Context, key string, value int64) (int64, error)

	// Decrement decrements a counter
	Decrement(ctx context.Context, key string) (int64, error)

	// SetNX sets a key only if it doesn't exist
	SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error)

	// GetTTL returns the remaining TTL of a key
	GetTTL(ctx context.Context, key string) (time.Duration, error)

	// Ping checks if the cache is responsive
	Ping(ctx context.Context) error

	// Close closes the cache connection
	Close() error
}

// CacheKeys provides standardized cache key patterns
type CacheKeys struct {
	// Session keys
	SessionPrefix string

	// Tenant keys
	TenantPrefix string

	// User keys
	UserPrefix string

	// Artisan keys
	ArtisanPrefix string

	// Booking keys
	BookingPrefix string

	// Rate limit keys
	RateLimitPrefix string
}

// NewCacheKeys returns default cache key patterns
func NewCacheKeys() *CacheKeys {
	return &CacheKeys{
		SessionPrefix:   "session",
		TenantPrefix:    "tenant",
		UserPrefix:      "user",
		ArtisanPrefix:   "artisan",
		BookingPrefix:   "booking",
		RateLimitPrefix: "ratelimit",
	}
}

// SessionKey generates a session cache key
func (ck *CacheKeys) SessionKey(sessionID string) string {
	return ck.SessionPrefix + ":" + sessionID
}

// TenantKey generates a tenant cache key
func (ck *CacheKeys) TenantKey(tenantID string) string {
	return ck.TenantPrefix + ":" + tenantID
}

// TenantConfigKey generates a tenant config cache key
func (ck *CacheKeys) TenantConfigKey(tenantID string) string {
	return ck.TenantPrefix + ":" + tenantID + ":config"
}

// UserKey generates a user cache key
func (ck *CacheKeys) UserKey(userID string) string {
	return ck.UserPrefix + ":" + userID
}

// UserByEmailKey generates a user by email cache key
func (ck *CacheKeys) UserByEmailKey(email string) string {
	return ck.UserPrefix + ":email:" + email
}

// ArtisanKey generates an artisan cache key
func (ck *CacheKeys) ArtisanKey(artisanID string) string {
	return ck.ArtisanPrefix + ":" + artisanID
}

// ArtisanAvailabilityKey generates an artisan availability cache key
func (ck *CacheKeys) ArtisanAvailabilityKey(artisanID string, date string) string {
	return ck.ArtisanPrefix + ":" + artisanID + ":availability:" + date
}

// BookingKey generates a booking cache key
func (ck *CacheKeys) BookingKey(bookingID string) string {
	return ck.BookingPrefix + ":" + bookingID
}

// RateLimitKey generates a rate limit cache key
func (ck *CacheKeys) RateLimitKey(identifier string, window string) string {
	return ck.RateLimitPrefix + ":" + identifier + ":" + window
}

// CacheStrategy defines caching strategies
type CacheStrategy struct {
	// TTL is the time-to-live for cached items
	TTL time.Duration

	// Enabled indicates if caching is enabled
	Enabled bool

	// WarmupEnabled indicates if cache warmup is enabled
	WarmupEnabled bool
}

// DefaultCacheStrategies returns default caching strategies
func DefaultCacheStrategies() map[string]CacheStrategy {
	return map[string]CacheStrategy{
		"session": {
			TTL:           1 * time.Hour,
			Enabled:       true,
			WarmupEnabled: false,
		},
		"tenant_config": {
			TTL:           24 * time.Hour,
			Enabled:       true,
			WarmupEnabled: true,
		},
		"user": {
			TTL:           30 * time.Minute,
			Enabled:       true,
			WarmupEnabled: false,
		},
		"artisan_profile": {
			TTL:           15 * time.Minute,
			Enabled:       true,
			WarmupEnabled: false,
		},
		"artisan_availability": {
			TTL:           5 * time.Minute,
			Enabled:       true,
			WarmupEnabled: false,
		},
		"booking": {
			TTL:           10 * time.Minute,
			Enabled:       true,
			WarmupEnabled: false,
		},
		"rate_limit": {
			TTL:           1 * time.Minute,
			Enabled:       true,
			WarmupEnabled: false,
		},
		"query_result": {
			TTL:           5 * time.Minute,
			Enabled:       true,
			WarmupEnabled: false,
		},
	}
}

// CacheManager provides high-level caching operations
type CacheManager struct {
	cache      Cache
	keys       *CacheKeys
	strategies map[string]CacheStrategy
}

// NewCacheManager creates a new cache manager
func NewCacheManager(cache Cache) *CacheManager {
	return &CacheManager{
		cache:      cache,
		keys:       NewCacheKeys(),
		strategies: DefaultCacheStrategies(),
	}
}

// GetCache returns the underlying cache
func (cm *CacheManager) GetCache() Cache {
	return cm.cache
}

// GetKeys returns the cache keys helper
func (cm *CacheManager) GetKeys() *CacheKeys {
	return cm.keys
}

// GetStrategy returns the cache strategy for a given type
func (cm *CacheManager) GetStrategy(strategyType string) CacheStrategy {
	if strategy, ok := cm.strategies[strategyType]; ok {
		return strategy
	}
	// Return default strategy
	return CacheStrategy{
		TTL:           5 * time.Minute,
		Enabled:       true,
		WarmupEnabled: false,
	}
}

// SetStrategy sets a custom cache strategy
func (cm *CacheManager) SetStrategy(strategyType string, strategy CacheStrategy) {
	cm.strategies[strategyType] = strategy
}

// CacheWithStrategy caches a value using the specified strategy
func (cm *CacheManager) CacheWithStrategy(ctx context.Context, key string, value any, strategyType string) error {
	strategy := cm.GetStrategy(strategyType)
	if !strategy.Enabled {
		return nil // Caching disabled for this type
	}
	return cm.cache.SetJSON(ctx, key, value, strategy.TTL)
}

// GetFromCache retrieves a value from cache using the specified strategy
func (cm *CacheManager) GetFromCache(ctx context.Context, key string, dest any, strategyType string) error {
	strategy := cm.GetStrategy(strategyType)
	if !strategy.Enabled {
		return ErrCacheMiss // Caching disabled, treat as miss
	}
	return cm.cache.GetJSON(ctx, key, dest)
}

// InvalidatePattern invalidates all keys matching a pattern
func (cm *CacheManager) InvalidatePattern(ctx context.Context, pattern string) error {
	return cm.cache.DeletePattern(ctx, pattern)
}

// InvalidateTenant invalidates all cache for a tenant
func (cm *CacheManager) InvalidateTenant(ctx context.Context, tenantID string) error {
	pattern := cm.keys.TenantPrefix + ":" + tenantID + ":*"
	return cm.cache.DeletePattern(ctx, pattern)
}

// InvalidateUser invalidates all cache for a user
func (cm *CacheManager) InvalidateUser(ctx context.Context, userID string) error {
	pattern := cm.keys.UserPrefix + ":" + userID + "*"
	return cm.cache.DeletePattern(ctx, pattern)
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits          int64
	Misses        int64
	Sets          int64
	Deletes       int64
	HitRate       float64
	TotalRequests int64
}

// CalculateHitRate calculates the cache hit rate
func (cs *CacheStats) CalculateHitRate() {
	cs.TotalRequests = cs.Hits + cs.Misses
	if cs.TotalRequests > 0 {
		cs.HitRate = float64(cs.Hits) / float64(cs.TotalRequests) * 100
	}
}
