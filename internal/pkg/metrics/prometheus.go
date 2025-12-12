package metrics

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// PrometheusMetrics holds all Prometheus metrics collectors
type PrometheusMetrics struct {
	// HTTP metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestSize      *prometheus.SummaryVec
	HTTPResponseSize     *prometheus.SummaryVec
	HTTPRequestsInFlight prometheus.Gauge
	HTTPErrorsTotal      *prometheus.CounterVec

	// Database metrics
	DBQueriesTotal      *prometheus.CounterVec
	DBQueryDuration     *prometheus.HistogramVec
	DBConnectionsActive prometheus.Gauge
	DBConnectionsIdle   prometheus.Gauge
	DBConnectionsMax    prometheus.Gauge
	DBTransactionsTotal *prometheus.CounterVec
	DBErrorsTotal       *prometheus.CounterVec

	// Cache metrics
	CacheHitsTotal         *prometheus.CounterVec
	CacheMissesTotal       *prometheus.CounterVec
	CacheOperationDuration *prometheus.HistogramVec
	CacheSize              prometheus.Gauge
	CacheEvictions         *prometheus.CounterVec

	// Business metrics
	BookingsTotal        *prometheus.CounterVec
	BookingStatusChanges *prometheus.CounterVec
	PaymentsTotal        *prometheus.CounterVec
	PaymentAmount        *prometheus.CounterVec
	UsersTotal           *prometheus.CounterVec
	UserRegistrations    *prometheus.CounterVec
	ArtisansActive       prometheus.Gauge
	CustomersActive      prometheus.Gauge

	// Tenant metrics
	TenantsTotal       prometheus.Gauge
	TenantsByPlan      *prometheus.GaugeVec
	TenantStorageUsage *prometheus.GaugeVec
	TenantAPIRequests  *prometheus.CounterVec

	// Authentication metrics
	AuthAttempts     *prometheus.CounterVec
	AuthFailures     *prometheus.CounterVec
	TokenValidations *prometheus.CounterVec
	SessionsActive   prometheus.Gauge

	// Notification metrics
	NotificationsSent    *prometheus.CounterVec
	NotificationsFailed  *prometheus.CounterVec
	NotificationDuration *prometheus.HistogramVec

	// Background job metrics
	JobsExecuted *prometheus.CounterVec
	JobDuration  *prometheus.HistogramVec
	JobFailures  *prometheus.CounterVec
	JobsQueued   prometheus.Gauge

	// System metrics
	GoroutinesCount prometheus.Gauge
	MemoryAlloc     prometheus.Gauge
	MemorySys       prometheus.Gauge
	MemoryHeapAlloc prometheus.Gauge
	GCDuration      prometheus.Summary

	registry *prometheus.Registry
	logger   *zap.Logger
}

// NewPrometheusMetrics creates and registers all Prometheus metrics
func NewPrometheusMetrics(namespace string, logger *zap.Logger) *PrometheusMetrics {
	registry := prometheus.NewRegistry()

	pm := &PrometheusMetrics{
		registry: registry,
		logger:   logger,

		// HTTP metrics
		HTTPRequestsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.With(registry).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestSize: promauto.With(registry).NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  namespace,
				Name:       "http_request_size_bytes",
				Help:       "HTTP request size in bytes",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: promauto.With(registry).NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  namespace,
				Name:       "http_response_size_bytes",
				Help:       "HTTP response size in bytes",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"method", "path"},
		),
		HTTPRequestsInFlight: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being processed",
			},
		),
		HTTPErrorsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_errors_total",
				Help:      "Total number of HTTP errors",
			},
			[]string{"method", "path", "error_type"},
		),

		// Database metrics
		DBQueriesTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),
		DBQueryDuration: promauto.With(registry).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{.0001, .0005, .001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"operation", "table"},
		),
		DBConnectionsActive: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_active",
				Help:      "Current number of active database connections",
			},
		),
		DBConnectionsIdle: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_idle",
				Help:      "Current number of idle database connections",
			},
		),
		DBConnectionsMax: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_max",
				Help:      "Maximum number of database connections",
			},
		),
		DBTransactionsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_transactions_total",
				Help:      "Total number of database transactions",
			},
			[]string{"status"},
		),
		DBErrorsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_errors_total",
				Help:      "Total number of database errors",
			},
			[]string{"operation", "error_type"},
		),

		// Cache metrics
		CacheHitsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
			[]string{"cache_type"},
		),
		CacheMissesTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
			[]string{"cache_type"},
		),
		CacheOperationDuration: promauto.With(registry).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "cache_operation_duration_seconds",
				Help:      "Cache operation duration in seconds",
				Buckets:   []float64{.00001, .00005, .0001, .0005, .001, .005, .01, .025, .05},
			},
			[]string{"operation", "cache_type"},
		),
		CacheSize: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "cache_size_bytes",
				Help:      "Current cache size in bytes",
			},
		),
		CacheEvictions: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_evictions_total",
				Help:      "Total number of cache evictions",
			},
			[]string{"reason"},
		),

		// Business metrics
		BookingsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "bookings_total",
				Help:      "Total number of bookings",
			},
			[]string{"status", "tenant_id"},
		),
		BookingStatusChanges: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "booking_status_changes_total",
				Help:      "Total number of booking status changes",
			},
			[]string{"from_status", "to_status", "tenant_id"},
		),
		PaymentsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "payments_total",
				Help:      "Total number of payments",
			},
			[]string{"status", "method", "tenant_id"},
		),
		PaymentAmount: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "payment_amount_total",
				Help:      "Total payment amount",
			},
			[]string{"currency", "status", "tenant_id"},
		),
		UsersTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "users_total",
				Help:      "Total number of users",
			},
			[]string{"role", "tenant_id"},
		),
		UserRegistrations: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "user_registrations_total",
				Help:      "Total number of user registrations",
			},
			[]string{"role", "tenant_id"},
		),
		ArtisansActive: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "artisans_active",
				Help:      "Current number of active artisans",
			},
		),
		CustomersActive: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "customers_active",
				Help:      "Current number of active customers",
			},
		),

		// Tenant metrics
		TenantsTotal: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "tenants_total",
				Help:      "Total number of tenants",
			},
		),
		TenantsByPlan: promauto.With(registry).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "tenants_by_plan",
				Help:      "Number of tenants by plan",
			},
			[]string{"plan"},
		),
		TenantStorageUsage: promauto.With(registry).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "tenant_storage_usage_bytes",
				Help:      "Storage usage per tenant in bytes",
			},
			[]string{"tenant_id"},
		),
		TenantAPIRequests: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "tenant_api_requests_total",
				Help:      "Total API requests per tenant",
			},
			[]string{"tenant_id"},
		),

		// Authentication metrics
		AuthAttempts: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_attempts_total",
				Help:      "Total number of authentication attempts",
			},
			[]string{"method", "status"},
		),
		AuthFailures: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_failures_total",
				Help:      "Total number of authentication failures",
			},
			[]string{"method", "reason"},
		),
		TokenValidations: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "token_validations_total",
				Help:      "Total number of token validations",
			},
			[]string{"status"},
		),
		SessionsActive: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "sessions_active",
				Help:      "Current number of active sessions",
			},
		),

		// Notification metrics
		NotificationsSent: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "notifications_sent_total",
				Help:      "Total number of notifications sent",
			},
			[]string{"channel", "type"},
		),
		NotificationsFailed: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "notifications_failed_total",
				Help:      "Total number of failed notifications",
			},
			[]string{"channel", "type", "reason"},
		),
		NotificationDuration: promauto.With(registry).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "notification_duration_seconds",
				Help:      "Notification sending duration in seconds",
				Buckets:   []float64{.1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"channel", "type"},
		),

		// Background job metrics
		JobsExecuted: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "jobs_executed_total",
				Help:      "Total number of background jobs executed",
			},
			[]string{"job_type", "status"},
		),
		JobDuration: promauto.With(registry).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "job_duration_seconds",
				Help:      "Background job duration in seconds",
				Buckets:   []float64{.1, .5, 1, 5, 10, 30, 60, 120, 300},
			},
			[]string{"job_type"},
		),
		JobFailures: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "job_failures_total",
				Help:      "Total number of job failures",
			},
			[]string{"job_type", "reason"},
		),
		JobsQueued: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "jobs_queued",
				Help:      "Current number of jobs in queue",
			},
		),

		// System metrics
		GoroutinesCount: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "goroutines",
				Help:      "Current number of goroutines",
			},
		),
		MemoryAlloc: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_alloc_bytes",
				Help:      "Bytes of allocated heap objects",
			},
		),
		MemorySys: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_sys_bytes",
				Help:      "Total bytes of memory obtained from OS",
			},
		),
		MemoryHeapAlloc: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_heap_alloc_bytes",
				Help:      "Bytes of allocated heap objects",
			},
		),
		GCDuration: promauto.With(registry).NewSummary(
			prometheus.SummaryOpts{
				Namespace:  namespace,
				Name:       "gc_duration_seconds",
				Help:       "Garbage collection duration in seconds",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
		),
	}

	// Register Go runtime metrics
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	logger.Info("Prometheus metrics initialized", zap.String("namespace", namespace))

	return pm
}

// GetRegistry returns the Prometheus registry
func (pm *PrometheusMetrics) GetRegistry() *prometheus.Registry {
	return pm.registry
}

// Handler returns an HTTP handler for Prometheus metrics
func (pm *PrometheusMetrics) Handler() fiber.Handler {
	handler := promhttp.HandlerFor(pm.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})

	return adaptor.HTTPHandler(handler)
}

// RecordHTTPRequest records HTTP request metrics
func (pm *PrometheusMetrics) RecordHTTPRequest(method, path string, status int, duration time.Duration, requestSize, responseSize int) {
	statusStr := strconv.Itoa(status)
	pm.HTTPRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
	pm.HTTPRequestDuration.WithLabelValues(method, path, statusStr).Observe(duration.Seconds())
	pm.HTTPRequestSize.WithLabelValues(method, path).Observe(float64(requestSize))
	pm.HTTPResponseSize.WithLabelValues(method, path).Observe(float64(responseSize))
}

// RecordHTTPError records HTTP error metrics
func (pm *PrometheusMetrics) RecordHTTPError(method, path, errorType string) {
	pm.HTTPErrorsTotal.WithLabelValues(method, path, errorType).Inc()
}

// IncHTTPRequestsInFlight increments in-flight requests
func (pm *PrometheusMetrics) IncHTTPRequestsInFlight() {
	pm.HTTPRequestsInFlight.Inc()
}

// DecHTTPRequestsInFlight decrements in-flight requests
func (pm *PrometheusMetrics) DecHTTPRequestsInFlight() {
	pm.HTTPRequestsInFlight.Dec()
}

// RecordDBQuery records database query metrics
func (pm *PrometheusMetrics) RecordDBQuery(operation, table, status string, duration time.Duration) {
	pm.DBQueriesTotal.WithLabelValues(operation, table, status).Inc()
	pm.DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordDBError records database error metrics
func (pm *PrometheusMetrics) RecordDBError(operation, errorType string) {
	pm.DBErrorsTotal.WithLabelValues(operation, errorType).Inc()
}

// UpdateDBConnectionStats updates database connection pool statistics
func (pm *PrometheusMetrics) UpdateDBConnectionStats(active, idle, max int) {
	pm.DBConnectionsActive.Set(float64(active))
	pm.DBConnectionsIdle.Set(float64(idle))
	pm.DBConnectionsMax.Set(float64(max))
}

// RecordCacheHit records a cache hit
func (pm *PrometheusMetrics) RecordCacheHit(cacheType string) {
	pm.CacheHitsTotal.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss records a cache miss
func (pm *PrometheusMetrics) RecordCacheMiss(cacheType string) {
	pm.CacheMissesTotal.WithLabelValues(cacheType).Inc()
}

// RecordCacheOperation records cache operation duration
func (pm *PrometheusMetrics) RecordCacheOperation(operation, cacheType string, duration time.Duration) {
	pm.CacheOperationDuration.WithLabelValues(operation, cacheType).Observe(duration.Seconds())
}

// RecordBooking records booking creation metrics
func (pm *PrometheusMetrics) RecordBooking(status, tenantID string) {
	pm.BookingsTotal.WithLabelValues(status, tenantID).Inc()
}

// RecordBookingStatusChange records booking status change
func (pm *PrometheusMetrics) RecordBookingStatusChange(fromStatus, toStatus, tenantID string) {
	pm.BookingStatusChanges.WithLabelValues(fromStatus, toStatus, tenantID).Inc()
}

// RecordPayment records payment metrics
func (pm *PrometheusMetrics) RecordPayment(status, method, tenantID string, amount float64, currency string) {
	pm.PaymentsTotal.WithLabelValues(status, method, tenantID).Inc()
	pm.PaymentAmount.WithLabelValues(currency, status, tenantID).Add(amount)
}

// RecordUserRegistration records user registration
func (pm *PrometheusMetrics) RecordUserRegistration(role, tenantID string) {
	pm.UserRegistrations.WithLabelValues(role, tenantID).Inc()
	pm.UsersTotal.WithLabelValues(role, tenantID).Inc()
}

// RecordAuthAttempt records authentication attempt
func (pm *PrometheusMetrics) RecordAuthAttempt(method, status string) {
	pm.AuthAttempts.WithLabelValues(method, status).Inc()
}

// RecordAuthFailure records authentication failure
func (pm *PrometheusMetrics) RecordAuthFailure(method, reason string) {
	pm.AuthFailures.WithLabelValues(method, reason).Inc()
}

// RecordTokenValidation records token validation
func (pm *PrometheusMetrics) RecordTokenValidation(status string) {
	pm.TokenValidations.WithLabelValues(status).Inc()
}

// RecordNotification records notification metrics
func (pm *PrometheusMetrics) RecordNotification(channel, notifType string, success bool, duration time.Duration) {
	if success {
		pm.NotificationsSent.WithLabelValues(channel, notifType).Inc()
	} else {
		pm.NotificationsFailed.WithLabelValues(channel, notifType, "unknown").Inc()
	}
	pm.NotificationDuration.WithLabelValues(channel, notifType).Observe(duration.Seconds())
}

// RecordJob records background job execution
func (pm *PrometheusMetrics) RecordJob(jobType, status string, duration time.Duration) {
	pm.JobsExecuted.WithLabelValues(jobType, status).Inc()
	pm.JobDuration.WithLabelValues(jobType).Observe(duration.Seconds())
}

// RecordJobFailure records job failure
func (pm *PrometheusMetrics) RecordJobFailure(jobType, reason string) {
	pm.JobFailures.WithLabelValues(jobType, reason).Inc()
}

// UpdateTenantsTotal updates total tenants gauge
func (pm *PrometheusMetrics) UpdateTenantsTotal(count float64) {
	pm.TenantsTotal.Set(count)
}

// UpdateTenantsByPlan updates tenants by plan gauge
func (pm *PrometheusMetrics) UpdateTenantsByPlan(plan string, count float64) {
	pm.TenantsByPlan.WithLabelValues(plan).Set(count)
}

// RecordTenantAPIRequest records API request per tenant
func (pm *PrometheusMetrics) RecordTenantAPIRequest(tenantID string) {
	pm.TenantAPIRequests.WithLabelValues(tenantID).Inc()
}
