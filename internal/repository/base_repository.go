package repository

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"Krafti_Vibe/internal/infrastructure/cache"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/pkg/metrics"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseRepository defines the interface for base repository operations
type BaseRepository[T any] interface {
	// Basic CRUD
	Create(ctx context.Context, entity *T) error
	CreateBatch(ctx context.Context, entities []*T) error
	GetByID(ctx context.Context, id uuid.UUID) (*T, error)
	GetByIDWithTenant(ctx context.Context, id uuid.UUID, tenantID *uuid.UUID) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error

	// Query operations
	Find(ctx context.Context, filters map[string]any) ([]*T, error)
	FindWithPagination(ctx context.Context, filters map[string]any, pagination PaginationParams) ([]*T, PaginationResult, error)
	Count(ctx context.Context, filters map[string]any) (int64, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// Cache operations
	InvalidateCache(ctx context.Context, id uuid.UUID) error
	InvalidateCachePattern(ctx context.Context, pattern string) error

	// Transaction support
	WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error
	GetDB() *gorm.DB
}

// baseRepository implements BaseRepository interface
type baseRepository[T any] struct {
	db          *gorm.DB
	logger      log.AllLogger
	tableName   string
	auditLogger AuditLogger
	cache       Cache
	metrics     MetricsCollector
	cacheTTL    time.Duration
}

// RepositoryConfig holds configuration for repository
type RepositoryConfig struct {
	Logger      log.AllLogger
	AuditLogger AuditLogger
	Cache       Cache
	Metrics     MetricsCollector
	CacheTTL    time.Duration // Default cache TTL
}

// Cache interface for caching operations
type Cache interface {
	GetJSON(ctx context.Context, key string, dest any) error
	SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
}

// MetricsCollector interface for collecting metrics
type MetricsCollector interface {
	RecordOperation(operation string, table string, duration time.Duration, err error)
	RecordCacheHit(table string)
	RecordCacheMiss(table string)
	RecordQueryCount(table string, count int64)
}

// DefaultCacheAdapter adapts our Redis client to the Cache interface
type DefaultCacheAdapter struct {
	client *cache.RedisClient
}

// NewDefaultCacheAdapter creates a new cache adapter
func NewDefaultCacheAdapter(client *cache.RedisClient) Cache {
	return &DefaultCacheAdapter{client: client}
}

func (a *DefaultCacheAdapter) GetJSON(ctx context.Context, key string, dest any) error {
	return a.client.GetJSON(ctx, key, dest)
}

func (a *DefaultCacheAdapter) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	return a.client.SetJSON(ctx, key, value, ttl)
}

func (a *DefaultCacheAdapter) Delete(ctx context.Context, keys ...string) error {
	return a.client.Delete(ctx, keys...)
}

func (a *DefaultCacheAdapter) DeletePattern(ctx context.Context, pattern string) error {
	return a.client.DeletePattern(ctx, pattern)
}

func (a *DefaultCacheAdapter) Exists(ctx context.Context, keys ...string) (int64, error) {
	return a.client.Exists(ctx, keys...)
}

// DefaultMetricsCollector implements MetricsCollector using Prometheus
type DefaultMetricsCollector struct {
	prometheus *metrics.PrometheusMetrics
}

// NewDefaultMetricsCollector creates a new metrics collector
func NewDefaultMetricsCollector(prometheus *metrics.PrometheusMetrics) MetricsCollector {
	return &DefaultMetricsCollector{prometheus: prometheus}
}

func (m *DefaultMetricsCollector) RecordOperation(operation string, table string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	m.prometheus.RecordDBQuery(operation, table, status, duration)
}

func (m *DefaultMetricsCollector) RecordCacheHit(table string) {
	m.prometheus.RecordCacheHit(table)
}

func (m *DefaultMetricsCollector) RecordCacheMiss(table string) {
	m.prometheus.RecordCacheMiss(table)
}

func (m *DefaultMetricsCollector) RecordQueryCount(table string, count int64) {
	// Prometheus doesn't have a direct count metric, but we can use it differently
	// For now, we'll just log it
}

// NewBaseRepository creates a new base repository
func NewBaseRepository[T any](db *gorm.DB, config ...RepositoryConfig) BaseRepository[T] {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	// Set default cache TTL if not provided
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 5 * time.Minute
	}

	// Get table name from type
	var entity T
	tableName := getTableName(entity)

	return &baseRepository[T]{
		db:          db,
		logger:      cfg.Logger,
		tableName:   tableName,
		auditLogger: cfg.AuditLogger,
		cache:       cfg.Cache,
		metrics:     cfg.Metrics,
		cacheTTL:    cfg.CacheTTL,
	}
}

// Create creates a new entity
func (r *baseRepository[T]) Create(ctx context.Context, entity *T) error {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("create", r.tableName, time.Since(start), nil)
		}
	}()

	if entity == nil {
		return errors.NewRepositoryError("INVALID_INPUT", "entity cannot be nil", errors.ErrInvalidInput)
	}

	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		r.logger.Error("failed to create entity", "table", r.tableName, "error", err)

		if r.metrics != nil {
			r.metrics.RecordOperation("create", r.tableName, time.Since(start), err)
		}

		// Check for duplicate key error
		if isDuplicateError(err) {
			return errors.NewRepositoryError("DUPLICATE", "entity already exists", errors.ErrDuplicate)
		}

		return errors.NewRepositoryError("CREATE_FAILED", "failed to create entity", err)
	}

	// Invalidate list cache after creation
	if r.cache != nil {
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("list:*"))
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("count:*"))
	}

	// Log audit trail
	if r.auditLogger != nil {
		if id := getEntityID(entity); id != uuid.Nil {
			r.auditLogger.LogCreate(ctx, r.tableName, id, entity)
		}
	}

	r.logger.Debug("entity created", "table", r.tableName, "id", getEntityID(entity))
	return nil
}

// CreateBatch creates multiple entities in a batch
func (r *baseRepository[T]) CreateBatch(ctx context.Context, entities []*T) error {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("create_batch", r.tableName, time.Since(start), nil)
		}
	}()

	if len(entities) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(entities, 100).Error; err != nil {
		r.logger.Error("failed to create batch", "table", r.tableName, "count", len(entities), "error", err)

		if r.metrics != nil {
			r.metrics.RecordOperation("create_batch", r.tableName, time.Since(start), err)
		}

		return errors.NewRepositoryError("CREATE_BATCH_FAILED", "failed to create batch", err)
	}

	// Invalidate list cache after batch creation
	if r.cache != nil {
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("list:*"))
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("count:*"))
	}

	r.logger.Debug("batch created", "table", r.tableName, "count", len(entities))
	return nil
}

// GetByID retrieves an entity by ID with caching
func (r *baseRepository[T]) GetByID(ctx context.Context, id uuid.UUID) (*T, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("get_by_id", r.tableName, time.Since(start), nil)
		}
	}()

	if id == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "id cannot be nil", errors.ErrInvalidInput)
	}

	// Try to get from cache first
	cacheKey := r.getCacheKey("id", id.String())
	if r.cache != nil {
		var entity T
		if err := r.cache.GetJSON(ctx, cacheKey, &entity); err == nil {
			if r.metrics != nil {
				r.metrics.RecordCacheHit(r.tableName)
			}
			r.logger.Debug("cache hit", "table", r.tableName, "id", id)
			return &entity, nil
		}
		if r.metrics != nil {
			r.metrics.RecordCacheMiss(r.tableName)
		}
	}

	// Cache miss - fetch from database
	var entity T
	if err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "entity not found", errors.ErrNotFound)
		}
		r.logger.Error("failed to get entity by id", "table", r.tableName, "id", id, "error", err)

		if r.metrics != nil {
			r.metrics.RecordOperation("get_by_id", r.tableName, time.Since(start), err)
		}

		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get entity", err)
	}

	// Store in cache
	if r.cache != nil {
		if err := r.cache.SetJSON(ctx, cacheKey, entity, r.cacheTTL); err != nil {
			r.logger.Warn("failed to cache entity", "table", r.tableName, "id", id, "error", err)
		}
	}

	return &entity, nil
}

// GetByIDWithTenant retrieves an entity by ID with tenant isolation
func (r *baseRepository[T]) GetByIDWithTenant(ctx context.Context, id uuid.UUID, tenantID *uuid.UUID) (*T, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("get_by_id_tenant", r.tableName, time.Since(start), nil)
		}
	}()

	if id == uuid.Nil {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "id cannot be nil", errors.ErrInvalidInput)
	}

	// Build cache key with tenant
	var cacheKey string
	if tenantID != nil && *tenantID != uuid.Nil {
		cacheKey = r.getCacheKey("id:tenant", id.String(), tenantID.String())
	} else {
		cacheKey = r.getCacheKey("id", id.String())
	}

	// Try to get from cache first
	if r.cache != nil {
		var entity T
		if err := r.cache.GetJSON(ctx, cacheKey, &entity); err == nil {
			if r.metrics != nil {
				r.metrics.RecordCacheHit(r.tableName)
			}
			return &entity, nil
		}
		if r.metrics != nil {
			r.metrics.RecordCacheMiss(r.tableName)
		}
	}

	var entity T
	query := r.db.WithContext(ctx).Where("id = ?", id)

	// Apply tenant isolation if tenantID is provided
	if tenantID != nil && *tenantID != uuid.Nil {
		query = query.Where("tenant_id = ?", *tenantID)
	}

	if err := query.First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", "entity not found", errors.ErrNotFound)
		}
		r.logger.Error("failed to get entity by id with tenant", "table", r.tableName, "id", id, "tenant_id", tenantID, "error", err)

		if r.metrics != nil {
			r.metrics.RecordOperation("get_by_id_tenant", r.tableName, time.Since(start), err)
		}

		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get entity", err)
	}

	// Store in cache
	if r.cache != nil {
		if err := r.cache.SetJSON(ctx, cacheKey, entity, r.cacheTTL); err != nil {
			r.logger.Warn("failed to cache entity", "table", r.tableName, "id", id, "error", err)
		}
	}

	return &entity, nil
}

// Update updates an entity with optimistic locking support
func (r *baseRepository[T]) Update(ctx context.Context, entity *T) error {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("update", r.tableName, time.Since(start), nil)
		}
	}()

	if entity == nil {
		return errors.NewRepositoryError("INVALID_INPUT", "entity cannot be nil", errors.ErrInvalidInput)
	}

	id := getEntityID(entity)
	if id == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "entity id cannot be nil", errors.ErrInvalidInput)
	}

	// Get old values for audit trail and version check
	var oldEntity T
	var oldValues map[string]any
	var currentVersion int

	if err := r.db.WithContext(ctx).First(&oldEntity, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "entity not found", errors.ErrNotFound)
		}
		r.logger.Error("failed to get entity for update", "table", r.tableName, "id", id, "error", err)
		return errors.NewRepositoryError("GET_FAILED", "failed to get entity", err)
	}

	// Extract version for optimistic locking
	currentVersion = getEntityVersion(&oldEntity)
	entityVersion := getEntityVersion(entity)

	// Check optimistic locking - version must match
	if currentVersion != entityVersion {
		return errors.NewRepositoryError("CONFLICT", "entity was modified by another process", errors.ErrConflict)
	}

	if r.auditLogger != nil {
		oldValues = entityToMap(&oldEntity)
	}

	// Update with version check in WHERE clause for additional safety
	result := r.db.WithContext(ctx).
		Model(entity).
		Where("id = ? AND version = ?", id, currentVersion).
		Updates(entity)

	if result.Error != nil {
		r.logger.Error("failed to update entity", "table", r.tableName, "id", id, "error", result.Error)

		if r.metrics != nil {
			r.metrics.RecordOperation("update", r.tableName, time.Since(start), result.Error)
		}

		return errors.NewRepositoryError("UPDATE_FAILED", "failed to update entity", result.Error)
	}

	// Check if update actually affected a row (optimistic locking check)
	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("CONFLICT", "entity was modified by another process", errors.ErrConflict)
	}

	// Invalidate cache for this entity
	if r.cache != nil {
		r.InvalidateCache(ctx, id)
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("list:*"))
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("count:*"))
	}

	// Log audit trail
	if r.auditLogger != nil {
		newValues := entityToMap(entity)
		r.auditLogger.LogUpdate(ctx, r.tableName, id, oldValues, newValues)
	}

	r.logger.Debug("entity updated", "table", r.tableName, "id", id)
	return nil
}

// Delete permanently deletes an entity
func (r *baseRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("delete", r.tableName, time.Since(start), nil)
		}
	}()

	if id == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "id cannot be nil", errors.ErrInvalidInput)
	}

	var entity T
	result := r.db.WithContext(ctx).Unscoped().Delete(&entity, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("failed to delete entity", "table", r.tableName, "id", id, "error", result.Error)

		if r.metrics != nil {
			r.metrics.RecordOperation("delete", r.tableName, time.Since(start), result.Error)
		}

		return errors.NewRepositoryError("DELETE_FAILED", "failed to delete entity", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "entity not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.InvalidateCache(ctx, id)
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("list:*"))
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("count:*"))
	}

	// Log audit trail
	if r.auditLogger != nil {
		r.auditLogger.LogDelete(ctx, r.tableName, id)
	}

	r.logger.Debug("entity deleted", "table", r.tableName, "id", id)
	return nil
}

// SoftDelete soft deletes an entity
func (r *baseRepository[T]) SoftDelete(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("soft_delete", r.tableName, time.Since(start), nil)
		}
	}()

	if id == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "id cannot be nil", errors.ErrInvalidInput)
	}

	var entity T
	result := r.db.WithContext(ctx).Delete(&entity, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("failed to soft delete entity", "table", r.tableName, "id", id, "error", result.Error)

		if r.metrics != nil {
			r.metrics.RecordOperation("soft_delete", r.tableName, time.Since(start), result.Error)
		}

		return errors.NewRepositoryError("SOFT_DELETE_FAILED", "failed to soft delete entity", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "entity not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.InvalidateCache(ctx, id)
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("list:*"))
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("count:*"))
	}

	// Log audit trail
	if r.auditLogger != nil {
		r.auditLogger.LogDelete(ctx, r.tableName, id)
	}

	r.logger.Debug("entity soft deleted", "table", r.tableName, "id", id)
	return nil
}

// Restore restores a soft-deleted entity
func (r *baseRepository[T]) Restore(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("restore", r.tableName, time.Since(start), nil)
		}
	}()

	if id == uuid.Nil {
		return errors.NewRepositoryError("INVALID_INPUT", "id cannot be nil", errors.ErrInvalidInput)
	}

	var entity T
	result := r.db.WithContext(ctx).Unscoped().Model(&entity).Where("id = ?", id).Update("deleted_at", nil)
	if result.Error != nil {
		r.logger.Error("failed to restore entity", "table", r.tableName, "id", id, "error", result.Error)

		if r.metrics != nil {
			r.metrics.RecordOperation("restore", r.tableName, time.Since(start), result.Error)
		}

		return errors.NewRepositoryError("RESTORE_FAILED", "failed to restore entity", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "entity not found", errors.ErrNotFound)
	}

	// Invalidate cache
	if r.cache != nil {
		r.InvalidateCache(ctx, id)
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("list:*"))
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("count:*"))
	}

	r.logger.Debug("entity restored", "table", r.tableName, "id", id)
	return nil
}

// Find finds entities matching filters with caching
func (r *baseRepository[T]) Find(ctx context.Context, filters map[string]any) ([]*T, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("find", r.tableName, time.Since(start), nil)
		}
	}()

	// Build cache key from filters
	cacheKey := r.getCacheKeyFromFilters("list", filters)

	// Try cache first
	if r.cache != nil {
		var entities []*T
		if err := r.cache.GetJSON(ctx, cacheKey, &entities); err == nil {
			if r.metrics != nil {
				r.metrics.RecordCacheHit(r.tableName)
				r.metrics.RecordQueryCount(r.tableName, int64(len(entities)))
			}
			return entities, nil
		}
		if r.metrics != nil {
			r.metrics.RecordCacheMiss(r.tableName)
		}
	}

	query := r.db.WithContext(ctx)
	query = applyFilters(query, filters)

	var entities []*T
	if err := query.Find(&entities).Error; err != nil {
		r.logger.Error("failed to find entities", "table", r.tableName, "error", err)

		if r.metrics != nil {
			r.metrics.RecordOperation("find", r.tableName, time.Since(start), err)
		}

		return nil, errors.NewRepositoryError("FIND_FAILED", "failed to find entities", err)
	}

	// Cache results
	if r.cache != nil {
		if err := r.cache.SetJSON(ctx, cacheKey, entities, r.cacheTTL); err != nil {
			r.logger.Warn("failed to cache results", "table", r.tableName, "error", err)
		}
	}

	if r.metrics != nil {
		r.metrics.RecordQueryCount(r.tableName, int64(len(entities)))
	}

	return entities, nil
}

// FindWithPagination finds entities with pagination and caching
func (r *baseRepository[T]) FindWithPagination(ctx context.Context, filters map[string]any, pagination PaginationParams) ([]*T, PaginationResult, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("find_paginated", r.tableName, time.Since(start), nil)
		}
	}()

	pagination.Validate()

	// Build cache key
	cacheKey := r.getCacheKeyFromFiltersPagination("list:page", filters, pagination)

	// Try cache first
	type cachedResult struct {
		Entities   []*T
		Pagination PaginationResult
	}

	if r.cache != nil {
		var cached cachedResult
		if err := r.cache.GetJSON(ctx, cacheKey, &cached); err == nil {
			if r.metrics != nil {
				r.metrics.RecordCacheHit(r.tableName)
				r.metrics.RecordQueryCount(r.tableName, int64(len(cached.Entities)))
			}
			return cached.Entities, cached.Pagination, nil
		}
		if r.metrics != nil {
			r.metrics.RecordCacheMiss(r.tableName)
		}
	}

	// Build query with filters
	query := r.db.WithContext(ctx)
	query = applyFilters(query, filters)

	// Count total items
	var totalItems int64
	if err := query.Model(new(T)).Count(&totalItems).Error; err != nil {
		r.logger.Error("failed to count entities", "table", r.tableName, "error", err)

		if r.metrics != nil {
			r.metrics.RecordOperation("find_paginated", r.tableName, time.Since(start), err)
		}

		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count entities", err)
	}

	// Apply pagination
	query = query.Offset(pagination.Offset()).Limit(pagination.Limit())

	var entities []*T
	if err := query.Find(&entities).Error; err != nil {
		r.logger.Error("failed to find entities with pagination", "table", r.tableName, "error", err)

		if r.metrics != nil {
			r.metrics.RecordOperation("find_paginated", r.tableName, time.Since(start), err)
		}

		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find entities", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)

	// Cache results
	if r.cache != nil {
		cached := cachedResult{
			Entities:   entities,
			Pagination: paginationResult,
		}
		if err := r.cache.SetJSON(ctx, cacheKey, cached, r.cacheTTL); err != nil {
			r.logger.Warn("failed to cache paginated results", "table", r.tableName, "error", err)
		}
	}

	if r.metrics != nil {
		r.metrics.RecordQueryCount(r.tableName, totalItems)
	}

	return entities, paginationResult, nil
}

// Count counts entities matching filters with caching
func (r *baseRepository[T]) Count(ctx context.Context, filters map[string]any) (int64, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("count", r.tableName, time.Since(start), nil)
		}
	}()

	// Build cache key
	cacheKey := r.getCacheKeyFromFilters("count", filters)

	// Try cache first
	if r.cache != nil {
		var count int64
		if err := r.cache.GetJSON(ctx, cacheKey, &count); err == nil {
			if r.metrics != nil {
				r.metrics.RecordCacheHit(r.tableName)
			}
			return count, nil
		}
		if r.metrics != nil {
			r.metrics.RecordCacheMiss(r.tableName)
		}
	}

	query := r.db.WithContext(ctx).Model(new(T))
	query = applyFilters(query, filters)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		r.logger.Error("failed to count entities", "table", r.tableName, "error", err)

		if r.metrics != nil {
			r.metrics.RecordOperation("count", r.tableName, time.Since(start), err)
		}

		return 0, errors.NewRepositoryError("COUNT_FAILED", "failed to count entities", err)
	}

	// Cache count
	if r.cache != nil {
		if err := r.cache.SetJSON(ctx, cacheKey, count, r.cacheTTL); err != nil {
			r.logger.Warn("failed to cache count", "table", r.tableName, "error", err)
		}
	}

	return count, nil
}

// Exists checks if an entity exists with caching
func (r *baseRepository[T]) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("exists", r.tableName, time.Since(start), nil)
		}
	}()

	if id == uuid.Nil {
		return false, nil
	}

	// Try to get from cache (which will tell us if it exists)
	cacheKey := r.getCacheKey("exists", id.String())
	if r.cache != nil {
		var exists bool
		if err := r.cache.GetJSON(ctx, cacheKey, &exists); err == nil {
			if r.metrics != nil {
				r.metrics.RecordCacheHit(r.tableName)
			}
			return exists, nil
		}
		if r.metrics != nil {
			r.metrics.RecordCacheMiss(r.tableName)
		}
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Count(&count).Error; err != nil {
		r.logger.Error("failed to check existence", "table", r.tableName, "id", id, "error", err)

		if r.metrics != nil {
			r.metrics.RecordOperation("exists", r.tableName, time.Since(start), err)
		}

		return false, errors.NewRepositoryError("EXISTS_CHECK_FAILED", "failed to check existence", err)
	}

	exists := count > 0

	// Cache existence check
	if r.cache != nil {
		if err := r.cache.SetJSON(ctx, cacheKey, exists, r.cacheTTL); err != nil {
			r.logger.Warn("failed to cache exists check", "table", r.tableName, "error", err)
		}
	}

	return exists, nil
}

// InvalidateCache invalidates cache for a specific entity
func (r *baseRepository[T]) InvalidateCache(ctx context.Context, id uuid.UUID) error {
	if r.cache == nil {
		return nil
	}

	// Delete all cache keys related to this entity
	patterns := []string{
		r.getCacheKey("id", id.String()),
		r.getCacheKey("id:tenant", id.String(), "*"),
		r.getCacheKey("exists", id.String()),
	}

	for _, pattern := range patterns {
		if err := r.cache.DeletePattern(ctx, pattern); err != nil {
			r.logger.Warn("failed to invalidate cache", "pattern", pattern, "error", err)
		}
	}

	return nil
}

// InvalidateCachePattern invalidates cache based on a pattern
func (r *baseRepository[T]) InvalidateCachePattern(ctx context.Context, pattern string) error {
	if r.cache == nil {
		return nil
	}

	fullPattern := r.getCacheKeyPattern(pattern)
	if err := r.cache.DeletePattern(ctx, fullPattern); err != nil {
		r.logger.Warn("failed to invalidate cache pattern", "pattern", fullPattern, "error", err)
		return err
	}

	return nil
}

// WithTransaction executes a function within a transaction
func (r *baseRepository[T]) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("transaction", r.tableName, time.Since(start), nil)
		}
	}()

	return r.db.WithContext(ctx).Transaction(fn)
}

// GetDB returns the underlying GORM database instance
func (r *baseRepository[T]) GetDB() *gorm.DB {
	return r.db
}

// Cache key helpers

func (r *baseRepository[T]) getCacheKey(prefix string, parts ...string) string {
	allParts := append([]string{"repo", r.tableName, prefix}, parts...)
	return strings.Join(allParts, ":")
}

func (r *baseRepository[T]) getCacheKeyPattern(pattern string) string {
	return fmt.Sprintf("repo:%s:%s", r.tableName, pattern)
}

func (r *baseRepository[T]) getCacheKeyFromFilters(prefix string, filters map[string]any) string {
	if len(filters) == 0 {
		return r.getCacheKey(prefix, "all")
	}

	// Create a deterministic key from filters
	var parts []string
	for k, v := range filters {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	// Join all filter parts into a single string
	filterStr := strings.Join(parts, ":")
	return r.getCacheKey(prefix, filterStr)
}

func (r *baseRepository[T]) getCacheKeyFromFiltersPagination(prefix string, filters map[string]any, pagination PaginationParams) string {
	filterKey := r.getCacheKeyFromFilters(prefix, filters)
	return fmt.Sprintf("%s:page=%d:size=%d", filterKey, pagination.Page, pagination.PageSize)
}

// Helper functions

func getTableName(entity any) string {
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return fmt.Sprintf("%ss", toSnakeCase(t.Name()))
}

func getEntityID(entity any) uuid.UUID {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Try to get ID field from BaseModel
	idField := v.FieldByName("ID")
	if idField.IsValid() && idField.CanInterface() {
		if id, ok := idField.Interface().(uuid.UUID); ok {
			return id
		}
	}

	return uuid.Nil
}

func getEntityVersion(entity any) int {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Try to get Version field from BaseModel
	versionField := v.FieldByName("Version")
	if versionField.IsValid() && versionField.CanInterface() {
		if version, ok := versionField.Interface().(int); ok {
			return version
		}
	}

	return 0
}

func applyFilters(query *gorm.DB, filters map[string]any) *gorm.DB {
	for key, value := range filters {
		if value != nil {
			query = query.Where(fmt.Sprintf("%s = ?", key), value)
		}
	}
	return query
}

func entityToMap(entity any) map[string]any {
	result := make(map[string]any)
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields
		if !value.CanInterface() {
			continue
		}

		// Get JSON tag or use field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Remove omitempty and other options
		if commaIdx := strings.IndexByte(jsonTag, ','); commaIdx != -1 {
			jsonTag = jsonTag[:commaIdx]
		}

		if jsonTag != "" {
			result[jsonTag] = value.Interface()
		}
	}

	return result
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

func isDuplicateError(err error) bool {
	// Check for PostgreSQL duplicate key error
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique constraint")
}
