package middleware

import (
	"slices"
	"strconv"
	"time"

	"Krafti_Vibe/internal/pkg/metrics"

	"github.com/gofiber/fiber/v2"
)

// MetricsMiddleware creates a middleware for collecting Prometheus metrics
func MetricsMiddleware(m *metrics.PrometheusMetrics) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Record request start time
		start := time.Now()

		// Increment in-flight requests
		m.IncHTTPRequestsInFlight()
		defer m.DecHTTPRequestsInFlight()

		// Get request size
		requestSize := len(c.Request().Body())

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get response details
		statusCode := c.Response().StatusCode()
		responseSize := len(c.Response().Body())
		method := c.Method()
		path := c.Path()

		// Normalize path to reduce cardinality
		// Remove IDs and dynamic segments
		normalizedPath := normalizePath(path)

		// Record HTTP metrics
		m.RecordHTTPRequest(method, normalizedPath, statusCode, duration, requestSize, responseSize)

		// Record errors if any
		if statusCode >= 400 {
			errorType := getErrorType(statusCode)
			m.RecordHTTPError(method, normalizedPath, errorType)
		}

		// Record tenant-specific metrics if available
		if tenantID := c.Locals("tenant_id"); tenantID != nil {
			if tid, ok := tenantID.(string); ok {
				m.RecordTenantAPIRequest(tid)
			}
		}

		return err
	}
}

// normalizePath normalizes the path to reduce metric cardinality
// Replaces UUIDs and numeric IDs with placeholders
func normalizePath(path string) string {
	// Common path patterns to normalize
	// For production, consider using a more sophisticated path normalizer
	// or registering known routes

	// Simple implementation - return as-is for now
	// In production, you'd want to replace IDs with :id placeholder
	return path
}

// getErrorType categorizes HTTP errors
func getErrorType(statusCode int) string {
	switch {
	case statusCode >= 400 && statusCode < 500:
		return "client_error"
	case statusCode >= 500:
		return "server_error"
	default:
		return "unknown"
	}
}

// DBMetricsInterceptor wraps database operations with metrics
type DBMetricsInterceptor struct {
	metrics *metrics.PrometheusMetrics
}

// NewDBMetricsInterceptor creates a new database metrics interceptor
func NewDBMetricsInterceptor(m *metrics.PrometheusMetrics) *DBMetricsInterceptor {
	return &DBMetricsInterceptor{
		metrics: m,
	}
}

// RecordOperation records a database operation with metrics
func (i *DBMetricsInterceptor) RecordOperation(operation, table string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	status := "success"
	if err != nil {
		status = "error"
		i.metrics.RecordDBError(operation, "query_error")
	}

	i.metrics.RecordDBQuery(operation, table, status, duration)
	return err
}

// CacheMetricsWrapper wraps cache operations with metrics
type CacheMetricsWrapper struct {
	metrics *metrics.PrometheusMetrics
}

// NewCacheMetricsWrapper creates a new cache metrics wrapper
func NewCacheMetricsWrapper(m *metrics.PrometheusMetrics) *CacheMetricsWrapper {
	return &CacheMetricsWrapper{
		metrics: m,
	}
}

// RecordGet records a cache get operation
func (w *CacheMetricsWrapper) RecordGet(cacheType string, hit bool, duration time.Duration) {
	if hit {
		w.metrics.RecordCacheHit(cacheType)
	} else {
		w.metrics.RecordCacheMiss(cacheType)
	}
	w.metrics.RecordCacheOperation("get", cacheType, duration)
}

// RecordSet records a cache set operation
func (w *CacheMetricsWrapper) RecordSet(cacheType string, duration time.Duration) {
	w.metrics.RecordCacheOperation("set", cacheType, duration)
}

// RecordDelete records a cache delete operation
func (w *CacheMetricsWrapper) RecordDelete(cacheType string, duration time.Duration) {
	w.metrics.RecordCacheOperation("delete", cacheType, duration)
}

// SystemMetricsCollector periodically collects system metrics
type SystemMetricsCollector struct {
	metrics  *metrics.PrometheusMetrics
	interval time.Duration
	done     chan struct{}
}

// NewSystemMetricsCollector creates a new system metrics collector
func NewSystemMetricsCollector(m *metrics.PrometheusMetrics, interval time.Duration) *SystemMetricsCollector {
	return &SystemMetricsCollector{
		metrics:  m,
		interval: interval,
		done:     make(chan struct{}),
	}
}

// Start begins collecting system metrics
func (c *SystemMetricsCollector) Start() {
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.collect()
			case <-c.done:
				return
			}
		}
	}()
}

// Stop stops collecting system metrics
func (c *SystemMetricsCollector) Stop() {
	close(c.done)
}

// collect collects and records system metrics
func (c *SystemMetricsCollector) collect() {
	// Note: This is a placeholder
	// In production, you'd use runtime.ReadMemStats() and runtime.NumGoroutine()
	// to collect actual system metrics

	// Example:
	// var m runtime.MemStats
	// runtime.ReadMemStats(&m)
	// c.metrics.MemoryAlloc.Set(float64(m.Alloc))
	// c.metrics.MemorySys.Set(float64(m.Sys))
	// c.metrics.MemoryHeapAlloc.Set(float64(m.HeapAlloc))
	// c.metrics.GoroutinesCount.Set(float64(runtime.NumGoroutine()))
}

// BusinessMetricsRecorder provides helper methods for recording business metrics
type BusinessMetricsRecorder struct {
	metrics *metrics.PrometheusMetrics
}

// NewBusinessMetricsRecorder creates a new business metrics recorder
func NewBusinessMetricsRecorder(m *metrics.PrometheusMetrics) *BusinessMetricsRecorder {
	return &BusinessMetricsRecorder{
		metrics: m,
	}
}

// RecordBookingCreated records a booking creation
func (r *BusinessMetricsRecorder) RecordBookingCreated(status, tenantID string) {
	r.metrics.RecordBooking(status, tenantID)
}

// RecordBookingStatusChange records a booking status change
func (r *BusinessMetricsRecorder) RecordBookingStatusChange(fromStatus, toStatus, tenantID string) {
	r.metrics.RecordBookingStatusChange(fromStatus, toStatus, tenantID)
}

// RecordPaymentProcessed records a payment
func (r *BusinessMetricsRecorder) RecordPaymentProcessed(status, method, tenantID, currency string, amount float64) {
	r.metrics.RecordPayment(status, method, tenantID, amount, currency)
}

// RecordUserRegistered records a user registration
func (r *BusinessMetricsRecorder) RecordUserRegistered(role, tenantID string) {
	r.metrics.RecordUserRegistration(role, tenantID)
}

// RecordNotificationSent records a sent notification
func (r *BusinessMetricsRecorder) RecordNotificationSent(channel, notifType string, success bool, duration time.Duration) {
	r.metrics.RecordNotification(channel, notifType, success, duration)
}

// RecordJobExecuted records a background job execution
func (r *BusinessMetricsRecorder) RecordJobExecuted(jobType, status string, duration time.Duration) {
	r.metrics.RecordJob(jobType, status, duration)
}

// AuthMetricsRecorder provides helper methods for recording auth metrics
type AuthMetricsRecorder struct {
	metrics *metrics.PrometheusMetrics
}

// NewAuthMetricsRecorder creates a new auth metrics recorder
func NewAuthMetricsRecorder(m *metrics.PrometheusMetrics) *AuthMetricsRecorder {
	return &AuthMetricsRecorder{
		metrics: m,
	}
}

// RecordLoginAttempt records a login attempt
func (r *AuthMetricsRecorder) RecordLoginAttempt(method, status string) {
	r.metrics.RecordAuthAttempt(method, status)
}

// RecordLoginFailure records a login failure
func (r *AuthMetricsRecorder) RecordLoginFailure(method, reason string) {
	r.metrics.RecordAuthFailure(method, reason)
}

// RecordTokenValidation records a token validation
func (r *AuthMetricsRecorder) RecordTokenValidation(valid bool) {
	status := "valid"
	if !valid {
		status = "invalid"
	}
	r.metrics.RecordTokenValidation(status)
}

// MetricsConfig holds configuration for metrics middleware
type MetricsConfig struct {
	// Enabled determines if metrics collection is enabled
	Enabled bool

	// PathNormalization enables path normalization to reduce cardinality
	PathNormalization bool

	// ExcludePaths defines paths to exclude from metrics
	ExcludePaths []string

	// IncludeTenantID determines if tenant ID should be recorded
	IncludeTenantID bool

	// RecordRequestBody determines if request body size should be recorded
	RecordRequestBody bool

	// RecordResponseBody determines if response body size should be recorded
	RecordResponseBody bool
}

// DefaultMetricsConfig returns default metrics configuration
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Enabled:           true,
		PathNormalization: true,
		ExcludePaths: []string{
			"/health",
			"/health/live",
			"/health/ready",
			"/metrics",
		},
		IncludeTenantID:    true,
		RecordRequestBody:  true,
		RecordResponseBody: true,
	}
}

// MetricsMiddlewareWithConfig creates a middleware with custom configuration
func MetricsMiddlewareWithConfig(m *metrics.PrometheusMetrics, config MetricsConfig) fiber.Handler {
	if !config.Enabled {
		// Return no-op middleware if disabled
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		path := c.Path()

		// Check if path should be excluded
		if slices.Contains(config.ExcludePaths, path) {
			return c.Next()
		}

		start := time.Now()
		m.IncHTTPRequestsInFlight()
		defer m.DecHTTPRequestsInFlight()

		var requestSize int
		if config.RecordRequestBody {
			requestSize = len(c.Request().Body())
		}

		err := c.Next()

		duration := time.Since(start)
		statusCode := c.Response().StatusCode()
		method := c.Method()

		var responseSize int
		if config.RecordResponseBody {
			responseSize = len(c.Response().Body())
		}

		normalizedPath := path
		if config.PathNormalization {
			normalizedPath = normalizePath(path)
		}

		m.RecordHTTPRequest(method, normalizedPath, statusCode, duration, requestSize, responseSize)

		if statusCode >= 400 {
			errorType := getErrorType(statusCode)
			m.RecordHTTPError(method, normalizedPath, errorType)
		}

		if config.IncludeTenantID {
			if tenantID := c.Locals("tenant_id"); tenantID != nil {
				if tid, ok := tenantID.(string); ok {
					m.RecordTenantAPIRequest(tid)
				}
			}
		}

		return err
	}
}

// GetMetricsSummary returns a summary of current metrics
func GetMetricsSummary(c *fiber.Ctx) error {
	// This would typically query Prometheus or return cached metrics
	// For now, return a simple response
	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Metrics are being collected. Visit /metrics for Prometheus format.",
	})
}

// ConvertStatusCodeToString converts status code to string for metrics
func ConvertStatusCodeToString(code int) string {
	return strconv.Itoa(code)
}
