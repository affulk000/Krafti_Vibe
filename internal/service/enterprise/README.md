# Enterprise-Grade Service Implementation Guide

This document provides a comprehensive guide for implementing enterprise-grade features in the service layer of Krafti Vibe.

## Table of Contents

1. [Overview](#overview)
2. [Enterprise Features](#enterprise-features)
3. [Architecture Patterns](#architecture-patterns)
4. [Implementation Guide](#implementation-guide)
5. [Configuration](#configuration)
6. [Monitoring & Observability](#monitoring--observability)
7. [Best Practices](#best-practices)

## Overview

The enterprise-grade service implementation enhances the existing `service_service.go` with production-ready features including:

- **Circuit Breaker Pattern** - Fault tolerance and graceful degradation
- **Caching Strategy** - Multi-level caching with Redis
- **Metrics & Observability** - Prometheus metrics and distributed tracing
- **Event-Driven Architecture** - Asynchronous event publishing
- **Retry Logic** - Exponential backoff with jitter
- **Rate Limiting** - Request throttling and quota management
- **Audit Logging** - Compliance and security tracking
- **Batch Processing** - Optimized bulk operations
- **Health Checks** - Service health monitoring
- **Connection Pooling** - Resource optimization

## Enterprise Features

### 1. Circuit Breaker Pattern

Protects against cascading failures by monitoring errors and temporarily blocking requests when failure thresholds are exceeded.

**Configuration:**
```go
CircuitBreakerName:         "service-service"
CircuitBreakerMaxRequests:  3
CircuitBreakerInterval:     60 * time.Second
CircuitBreakerTimeout:      30 * time.Second
CircuitBreakerFailureRatio: 0.6
```

**States:**
- **Closed**: Normal operation, requests flow through
- **Open**: Failure threshold exceeded, requests fail fast
- **Half-Open**: Testing if service recovered

**Implementation:**
```go
result, err := s.circuitBreaker.Execute(func() (interface{}, error) {
    return s.serviceRepo.GetByID(ctx, serviceID)
})
```

### 2. Caching Strategy

Multi-level caching with Redis for improved performance and reduced database load.

**Cache Layers:**
- **L1 Cache**: In-memory (optional)
- **L2 Cache**: Redis with configurable TTL
- **Cache-Aside Pattern**: Load on miss, invalidate on write

**Key Patterns:**
```
service:{serviceID}
services:tenant:{tenantID}:page:{page}:size:{size}
services:category:{category}:page:{page}
services:artisan:{artisanID}
```

**Cache Invalidation Strategies:**
- Time-based expiration (TTL)
- Event-driven invalidation (on updates/deletes)
- Pattern-based invalidation (tenant-wide, category-wide)

### 3. Metrics & Observability

Comprehensive metrics collection using Prometheus and structured logging.

**Metric Types:**

**HTTP Metrics:**
- Request duration histogram
- Request count by endpoint
- Response size histogram
- Error rate by status code
- In-flight requests gauge

**Database Metrics:**
- Query duration by operation
- Query count by table
- Connection pool statistics
- Error count by error type

**Cache Metrics:**
- Hit/miss ratio
- Operation duration
- Eviction count
- Memory usage

**Business Metrics:**
- Services created/updated/deleted count
- Active vs inactive services
- Revenue by service
- Booking conversion rate

**Implementation:**
```go
// Record HTTP metrics
s.metrics.RecordHTTPRequest(method, path, statusCode, duration, requestSize, responseSize)

// Record database metrics
s.metrics.RecordDBQuery(operation, table, status, duration)

// Record cache metrics
s.metrics.RecordCacheHit(cacheType)
s.metrics.RecordCacheMiss(cacheType)

// Record business metrics
s.metrics.RecordServiceCreated(tenantID)
s.metrics.RecordServicePriceChange(serviceID, oldPrice, newPrice)
```

### 4. Event-Driven Architecture

Asynchronous event publishing for decoupled microservices communication.

**Event Types:**
- `service.created` - New service created
- `service.updated` - Service details updated
- `service.deleted` - Service soft deleted
- `service.activated` - Service made active
- `service.deactivated` - Service made inactive
- `service.price_changed` - Price updated
- `service.bulk_operation` - Bulk operation completed

**Event Structure:**
```go
type ServiceEvent struct {
    EventID     uuid.UUID              `json:"event_id"`
    EventType   string                 `json:"event_type"`
    ServiceID   uuid.UUID              `json:"service_id"`
    TenantID    uuid.UUID              `json:"tenant_id"`
    Timestamp   time.Time              `json:"timestamp"`
    ActorID     *uuid.UUID             `json:"actor_id,omitempty"`
    Payload     map[string]interface{} `json:"payload"`
    Metadata    map[string]string      `json:"metadata,omitempty"`
}
```

**Publisher Interface:**
```go
type EventPublisher interface {
    Publish(ctx context.Context, event *ServiceEvent) error
    PublishBatch(ctx context.Context, events []*ServiceEvent) error
}
```

**Implementation Options:**
- **Kafka** - High throughput, distributed
- **RabbitMQ** - Reliable message delivery
- **NATS** - Lightweight, cloud-native
- **Redis Streams** - Simple, fast
- **AWS SNS/SQS** - Managed service

### 5. Retry Logic with Exponential Backoff

Automatically retries failed operations with increasing delays.

**Configuration:**
```go
MaxRetries:             3
RetryDelay:             100 * time.Millisecond
RetryBackoffMultiplier: 2.0
```

**Retry Delays:**
- Attempt 1: 100ms
- Attempt 2: 200ms
- Attempt 3: 400ms

**Retryable Errors:**
- Network timeouts
- Database connection errors
- Temporary service unavailability
- Rate limit exceeded (with longer backoff)

**Non-Retryable Errors:**
- Validation errors
- Authorization errors
- Resource not found
- Business logic violations

### 6. Rate Limiting

Protects services from abuse and ensures fair resource allocation.

**Strategies:**

**Token Bucket:**
```go
type RateLimiter interface {
    Allow(ctx context.Context, key string) (bool, error)
    AllowN(ctx context.Context, key string, n int) (bool, error)
}
```

**Sliding Window:**
- Track requests in time windows
- More accurate than fixed windows
- Prevent burst traffic

**Rate Limit Tiers:**
- Per User: 100 requests/minute
- Per Tenant: 1000 requests/minute
- Per IP: 500 requests/minute
- Anonymous: 10 requests/minute

**Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1640000000
```

### 7. Audit Logging

Comprehensive audit trail for compliance and security.

**Audit Log Structure:**
```go
type AuditLog struct {
    ID          uuid.UUID              `json:"id"`
    TenantID    uuid.UUID              `json:"tenant_id"`
    ServiceID   *uuid.UUID             `json:"service_id,omitempty"`
    Action      string                 `json:"action"`
    ActorID     *uuid.UUID             `json:"actor_id,omitempty"`
    Changes     map[string]interface{} `json:"changes,omitempty"`
    Timestamp   time.Time              `json:"timestamp"`
    IPAddress   string                 `json:"ip_address,omitempty"`
    UserAgent   string                 `json:"user_agent,omitempty"`
    Success     bool                   `json:"success"`
    ErrorMsg    string                 `json:"error_msg,omitempty"`
}
```

**Actions to Audit:**
- CREATE: New service creation
- UPDATE: Service modifications
- DELETE: Service deletion
- ACTIVATE: Service activation
- DEACTIVATE: Service deactivation
- PRICE_CHANGE: Price updates
- BULK_OPERATION: Bulk operations

**Storage Options:**
- Database table (structured)
- Elasticsearch (searchable)
- S3/Object storage (archive)
- Audit service (centralized)

### 8. Batch Processing

Optimized processing for bulk operations.

**Features:**
- Configurable batch size
- Parallel processing with goroutines
- Worker pool pattern
- Progress tracking
- Partial failure handling
- Transaction management

**Implementation:**
```go
func (s *service) BulkCreateServices(
    ctx context.Context,
    requests []*dto.CreateServiceRequest,
) ([]*dto.ServiceResponse, []error) {
    batchSize := s.config.BatchSize
    responses := make([]*dto.ServiceResponse, len(requests))
    errors := make([]error, len(requests))
    
    // Process in batches
    for i := 0; i < len(requests); i += batchSize {
        end := min(i+batchSize, len(requests))
        batch := requests[i:end]
        
        // Parallel processing
        var wg sync.WaitGroup
        for j, req := range batch {
            wg.Add(1)
            go func(index int, request *dto.CreateServiceRequest) {
                defer wg.Done()
                resp, err := s.CreateService(ctx, request)
                responses[i+index] = resp
                errors[i+index] = err
            }(j, req)
        }
        wg.Wait()
    }
    
    return responses, errors
}
```

### 9. Health Checks

Service health monitoring for orchestration and load balancing.

**Health Check Levels:**

**Liveness Probe:**
- Service is running
- Can accept requests
- Basic connectivity

**Readiness Probe:**
- Service is ready to handle traffic
- Dependencies available
- Warm caches loaded

**Health Check Response:**
```json
{
    "status": "healthy",
    "timestamp": "2024-01-15T10:30:00Z",
    "checks": {
        "database": {
            "status": "up",
            "latency_ms": 5
        },
        "cache": {
            "status": "up",
            "latency_ms": 2
        },
        "circuit_breaker": {
            "status": "closed",
            "failure_count": 0
        }
    },
    "metrics": {
        "uptime_seconds": 86400,
        "total_requests": 1000000,
        "error_rate": 0.001
    }
}
```

### 10. Connection Pooling

Efficient database and cache connection management.

**Database Pool Configuration:**
```go
MaxOpenConns:    100
MaxIdleConns:    25
ConnMaxLifetime: 5 * time.Minute
ConnMaxIdleTime: 10 * time.Minute
```

**Redis Pool Configuration:**
```go
PoolSize:        50
MinIdleConns:    10
MaxConnAge:      30 * time.Minute
PoolTimeout:     4 * time.Second
IdleTimeout:     5 * time.Minute
```

## Architecture Patterns

### Layered Architecture

```
┌─────────────────────────────────────────────┐
│            API/Handler Layer                │
│  (HTTP/gRPC handlers, input validation)    │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│         Enterprise Service Layer            │
│  (Business logic, circuit breaker, cache)   │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│          Repository Layer                   │
│  (Data access, queries, transactions)       │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│            Database Layer                   │
│  (PostgreSQL, MySQL, etc.)                  │
└─────────────────────────────────────────────┘
```

### Middleware Stack

```
Request Flow:
1. Rate Limiting Middleware
2. Authentication Middleware
3. Authorization Middleware
4. Metrics Middleware
5. Logging Middleware
6. Recovery Middleware
7. Request ID Middleware
8. Handler
```

### Caching Strategy

```
Cache-Aside Pattern:

┌─────────────┐
│   Request   │
└──────┬──────┘
       │
       ▼
┌─────────────┐    Cache Hit    ┌──────────────┐
│Check Cache  ├────────────────►│Return Cached │
└──────┬──────┘                 └──────────────┘
       │ Cache Miss
       ▼
┌─────────────┐
│Query DB     │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│Update Cache │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│Return Data  │
└─────────────┘
```

## Implementation Guide

### Step 1: Install Dependencies

```bash
go get github.com/sony/gobreaker
go get github.com/redis/go-redis/v9
go get github.com/prometheus/client_golang/prometheus
go get go.uber.org/zap
```

### Step 2: Configure Environment Variables

```env
# Circuit Breaker
CIRCUIT_BREAKER_MAX_REQUESTS=3
CIRCUIT_BREAKER_INTERVAL=60s
CIRCUIT_BREAKER_TIMEOUT=30s
CIRCUIT_BREAKER_FAILURE_RATIO=0.6

# Cache
CACHE_ENABLED=true
CACHE_TTL=15m
CACHE_WARMUP_ENABLED=true

# Retry
MAX_RETRIES=3
RETRY_DELAY=100ms
RETRY_BACKOFF_MULTIPLIER=2.0

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_PER_SECOND=100

# Batch Processing
BATCH_SIZE=50
BATCH_TIMEOUT=5s

# Audit
AUDIT_ENABLED=true
```

### Step 3: Initialize Enterprise Service

See `enterprise_service.go` for complete implementation.

### Step 4: Wire Dependencies

```go
// Initialize dependencies
cache := cache.NewRedisClient(redisConfig, logger)
metrics := metrics.NewPrometheusMetrics()
eventPublisher := events.NewKafkaPublisher(kafkaConfig)
auditLogger := audit.NewDatabaseAuditLogger(db)

// Create enterprise config
config := enterprise.DefaultEnterpriseConfig()

// Initialize service
serviceService := enterprise.NewEnterpriseServiceService(
    serviceRepo,
    tenantRepo,
    userRepo,
    logger,
    cache,
    metrics,
    eventPublisher,
    auditLogger,
    config,
)
```

### Step 5: Add Middleware

```go
// Add metrics middleware
app.Use(middleware.MetricsMiddleware(metrics))

// Add rate limiting middleware
app.Use(middleware.RateLimitMiddleware(ratelimiter))

// Add request ID middleware
app.Use(middleware.RequestIDMiddleware())

// Add logging middleware
app.Use(middleware.LoggingMiddleware(logger))
```

## Configuration

### Enterprise Service Configuration

```go
type EnterpriseServiceConfig struct {
    // Circuit Breaker
    CircuitBreakerName          string
    CircuitBreakerMaxRequests   uint32
    CircuitBreakerInterval      time.Duration
    CircuitBreakerTimeout       time.Duration
    CircuitBreakerFailureRatio  float64

    // Cache
    CacheTTL                    time.Duration
    CacheEnabled                bool
    CacheWarmupEnabled          bool

    // Retry
    MaxRetries                  int
    RetryDelay                  time.Duration
    RetryBackoffMultiplier      float64

    // Rate Limiting
    RateLimitEnabled            bool
    RateLimitPerSecond          int

    // Batch Processing
    BatchSize                   int
    BatchTimeout                time.Duration

    // Audit
    AuditEnabled                bool
}
```

### Default Configuration

```go
func DefaultEnterpriseConfig() *EnterpriseServiceConfig {
    return &EnterpriseServiceConfig{
        CircuitBreakerName:         "service-service",
        CircuitBreakerMaxRequests:  3,
        CircuitBreakerInterval:     60 * time.Second,
        CircuitBreakerTimeout:      30 * time.Second,
        CircuitBreakerFailureRatio: 0.6,
        
        CacheTTL:                   15 * time.Minute,
        CacheEnabled:               true,
        CacheWarmupEnabled:         false,
        
        MaxRetries:                 3,
        RetryDelay:                 100 * time.Millisecond,
        RetryBackoffMultiplier:     2.0,
        
        RateLimitEnabled:           true,
        RateLimitPerSecond:         100,
        
        BatchSize:                  50,
        BatchTimeout:               5 * time.Second,
        
        AuditEnabled:               true,
    }
}
```

## Monitoring & Observability

### Prometheus Metrics Endpoints

```
GET /metrics
```

### Key Metrics to Monitor

**Service Health:**
- `service_health_status` - Service health status
- `circuit_breaker_state` - Circuit breaker state
- `cache_hit_ratio` - Cache effectiveness

**Performance:**
- `http_request_duration_seconds` - Request latency
- `db_query_duration_seconds` - Database query time
- `cache_operation_duration_seconds` - Cache operation time

**Throughput:**
- `http_requests_total` - Total HTTP requests
- `services_created_total` - Total services created
- `bulk_operations_total` - Total bulk operations

**Errors:**
- `http_requests_errors_total` - HTTP errors
- `db_errors_total` - Database errors
- `circuit_breaker_failures_total` - Circuit breaker trips

### Grafana Dashboards

Create dashboards for:
1. Service Overview
2. Performance Metrics
3. Error Rates
4. Cache Statistics
5. Business Metrics

### Alerting Rules

```yaml
groups:
  - name: service_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_errors_total[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          
      - alert: CircuitBreakerOpen
        expr: circuit_breaker_state == 2
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Circuit breaker is open"
          
      - alert: LowCacheHitRate
        expr: cache_hit_ratio < 0.5
        for: 10m
        labels:
          severity: info
        annotations:
          summary: "Cache hit rate is low"
```

## Best Practices

### 1. Error Handling

- Use typed errors for better handling
- Log errors with context
- Return user-friendly error messages
- Track error rates

### 2. Logging

- Use structured logging (JSON)
- Include request IDs
- Log at appropriate levels
- Avoid logging sensitive data

### 3. Security

- Validate all inputs
- Sanitize user data
- Use parameterized queries
- Implement proper authorization
- Audit sensitive operations

### 4. Performance

- Use connection pooling
- Implement caching strategically
- Optimize database queries
- Use batch operations
- Profile and benchmark

### 5. Testing

- Write unit tests
- Write integration tests
- Test failure scenarios
- Load testing
- Chaos engineering

### 6. Documentation

- Document API contracts
- Update architecture diagrams
- Maintain runbooks
- Document deployment procedures

### 7. Monitoring

- Set up comprehensive metrics
- Create meaningful dashboards
- Configure alerting
- Monitor SLOs/SLIs

### 8. Scalability

- Design for horizontal scaling
- Use stateless services
- Implement proper caching
- Optimize database access
- Use async processing where appropriate

## References

- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Cache-Aside Pattern](https://docs.microsoft.com/en-us/azure/architecture/patterns/cache-aside)
- [Retry Pattern](https://docs.microsoft.com/en-us/azure/architecture/patterns/retry)
- [Event-Driven Architecture](https://martinfowler.com/articles/201701-event-driven.html)
- [Observability Engineering](https://www.oreilly.com/library/view/observability-engineering/9781492076438/)

## Support

For questions or issues, please contact the platform engineering team.