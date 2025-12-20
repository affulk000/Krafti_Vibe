package repository

import (
	"context"
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SDKClientRepository defines the interface for SDK client data operations
type SDKClientRepository interface {
	// CRUD operations
	Create(ctx context.Context, client *models.SDKClient) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.SDKClient, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.SDKClient, error)
	Update(ctx context.Context, client *models.SDKClient) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Query operations
	ListByTenant(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*models.SDKClient, int64, error)
	ListByPlatform(ctx context.Context, tenantID uuid.UUID, platform models.SDKPlatform, page, pageSize int) ([]*models.SDKClient, int64, error)
	ListByEnvironment(ctx context.Context, tenantID uuid.UUID, environment models.SDKEnvironment, page, pageSize int) ([]*models.SDKClient, int64, error)

	// Stats
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
	UpdateUsageStats(ctx context.Context, clientID uuid.UUID, isError bool) error
}

// SDKKeyRepository defines the interface for SDK key data operations
type SDKKeyRepository interface {
	// CRUD operations
	Create(ctx context.Context, key *models.SDKKey) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.SDKKey, error)
	GetByKeyHash(ctx context.Context, keyHash string) (*models.SDKKey, error)
	Update(ctx context.Context, key *models.SDKKey) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Query operations
	ListByClient(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]*models.SDKKey, int64, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*models.SDKKey, int64, error)
	ListByStatus(ctx context.Context, tenantID uuid.UUID, status models.SDKKeyStatus, page, pageSize int) ([]*models.SDKKey, int64, error)

	// Key operations
	Revoke(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID, reason string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.SDKKeyStatus) error
	UpdateUsageStats(ctx context.Context, keyID uuid.UUID, ipAddress string, isError bool) error

	// Expiration
	GetExpiredKeys(ctx context.Context) ([]*models.SDKKey, error)
	MarkAsExpired(ctx context.Context, ids []uuid.UUID) error
}

// SDKUsageRepository defines the interface for SDK usage data operations
type SDKUsageRepository interface {
	// Create usage record
	Create(ctx context.Context, usage *models.SDKUsage) error
	BulkCreate(ctx context.Context, usages []*models.SDKUsage) error

	// Query operations
	ListByClient(ctx context.Context, clientID uuid.UUID, startDate, endDate time.Time, page, pageSize int) ([]*models.SDKUsage, int64, error)
	ListByKey(ctx context.Context, keyID uuid.UUID, startDate, endDate time.Time, page, pageSize int) ([]*models.SDKUsage, int64, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, page, pageSize int) ([]*models.SDKUsage, int64, error)

	// Analytics
	GetStats(ctx context.Context, clientID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error)
	GetTopEndpoints(ctx context.Context, clientID uuid.UUID, startDate, endDate time.Time, limit int) ([]map[string]interface{}, error)
	GetErrorsByType(ctx context.Context, clientID uuid.UUID, startDate, endDate time.Time) (map[string]int64, error)
	GetRequestsByDay(ctx context.Context, clientID uuid.UUID, startDate, endDate time.Time) ([]map[string]interface{}, error)
	GetGeographicDistribution(ctx context.Context, clientID uuid.UUID, startDate, endDate time.Time) ([]map[string]interface{}, error)
}

// Implementation

type sdkClientRepository struct {
	db *gorm.DB
}

func NewSDKClientRepository(db *gorm.DB) SDKClientRepository {
	return &sdkClientRepository{db: db}
}

func (r *sdkClientRepository) Create(ctx context.Context, client *models.SDKClient) error {
	return r.db.WithContext(ctx).Create(client).Error
}

func (r *sdkClientRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SDKClient, error) {
	var client models.SDKClient
	err := r.db.WithContext(ctx).
		Preload("APIKeys").
		Where("id = ?", id).
		First(&client).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *sdkClientRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.SDKClient, error) {
	var client models.SDKClient
	err := r.db.WithContext(ctx).
		Preload("APIKeys").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&client).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *sdkClientRepository) Update(ctx context.Context, client *models.SDKClient) error {
	return r.db.WithContext(ctx).Save(client).Error
}

func (r *sdkClientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.SDKClient{}, id).Error
}

func (r *sdkClientRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*models.SDKClient, int64, error) {
	var clients []*models.SDKClient
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).
		Model(&models.SDKClient{}).
		Where("tenant_id = ?", tenantID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).
		Preload("APIKeys").
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&clients).Error; err != nil {
		return nil, 0, err
	}

	return clients, total, nil
}

func (r *sdkClientRepository) ListByPlatform(ctx context.Context, tenantID uuid.UUID, platform models.SDKPlatform, page, pageSize int) ([]*models.SDKClient, int64, error) {
	var clients []*models.SDKClient
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).
		Model(&models.SDKClient{}).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).
		Preload("APIKeys").
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&clients).Error; err != nil {
		return nil, 0, err
	}

	return clients, total, nil
}

func (r *sdkClientRepository) ListByEnvironment(ctx context.Context, tenantID uuid.UUID, environment models.SDKEnvironment, page, pageSize int) ([]*models.SDKClient, int64, error) {
	var clients []*models.SDKClient
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).
		Model(&models.SDKClient{}).
		Where("tenant_id = ? AND environment = ?", tenantID, environment).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).
		Preload("APIKeys").
		Where("tenant_id = ? AND environment = ?", tenantID, environment).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&clients).Error; err != nil {
		return nil, 0, err
	}

	return clients, total, nil
}

func (r *sdkClientRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.SDKClient{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	return count, err
}

func (r *sdkClientRepository) UpdateUsageStats(ctx context.Context, clientID uuid.UUID, isError bool) error {
	updates := map[string]interface{}{
		"total_requests": gorm.Expr("total_requests + ?", 1),
		"last_used_at":   time.Now(),
	}

	if isError {
		updates["total_errors"] = gorm.Expr("total_errors + ?", 1)
	}

	return r.db.WithContext(ctx).
		Model(&models.SDKClient{}).
		Where("id = ?", clientID).
		Updates(updates).Error
}

// SDKKey Repository

type sdkKeyRepository struct {
	db *gorm.DB
}

func NewSDKKeyRepository(db *gorm.DB) SDKKeyRepository {
	return &sdkKeyRepository{db: db}
}

func (r *sdkKeyRepository) Create(ctx context.Context, key *models.SDKKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

func (r *sdkKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SDKKey, error) {
	var key models.SDKKey
	err := r.db.WithContext(ctx).
		Preload("Client").
		Where("id = ?", id).
		First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *sdkKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (*models.SDKKey, error) {
	var key models.SDKKey
	err := r.db.WithContext(ctx).
		Preload("Client").
		Where("key_hash = ?", keyHash).
		First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *sdkKeyRepository) Update(ctx context.Context, key *models.SDKKey) error {
	return r.db.WithContext(ctx).Save(key).Error
}

func (r *sdkKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.SDKKey{}, id).Error
}

func (r *sdkKeyRepository) ListByClient(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]*models.SDKKey, int64, error) {
	var keys []*models.SDKKey
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).
		Model(&models.SDKKey{}).
		Where("client_id = ?", clientID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).
		Where("client_id = ?", clientID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&keys).Error; err != nil {
		return nil, 0, err
	}

	return keys, total, nil
}

func (r *sdkKeyRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*models.SDKKey, int64, error) {
	var keys []*models.SDKKey
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).
		Model(&models.SDKKey{}).
		Where("tenant_id = ?", tenantID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).
		Preload("Client").
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&keys).Error; err != nil {
		return nil, 0, err
	}

	return keys, total, nil
}

func (r *sdkKeyRepository) ListByStatus(ctx context.Context, tenantID uuid.UUID, status models.SDKKeyStatus, page, pageSize int) ([]*models.SDKKey, int64, error) {
	var keys []*models.SDKKey
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).
		Model(&models.SDKKey{}).
		Where("tenant_id = ? AND status = ?", tenantID, status).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).
		Preload("Client").
		Where("tenant_id = ? AND status = ?", tenantID, status).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&keys).Error; err != nil {
		return nil, 0, err
	}

	return keys, total, nil
}

func (r *sdkKeyRepository) Revoke(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID, reason string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.SDKKey{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        models.SDKKeyStatusRevoked,
			"revoked_at":    &now,
			"revoked_by":    &revokedBy,
			"revoke_reason": reason,
		}).Error
}

func (r *sdkKeyRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.SDKKeyStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.SDKKey{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *sdkKeyRepository) UpdateUsageStats(ctx context.Context, keyID uuid.UUID, ipAddress string, isError bool) error {
	updates := map[string]interface{}{
		"total_requests": gorm.Expr("total_requests + ?", 1),
		"last_used_at":   time.Now(),
		"last_used_ip":   ipAddress,
	}

	if isError {
		updates["total_errors"] = gorm.Expr("total_errors + ?", 1)
	}

	return r.db.WithContext(ctx).
		Model(&models.SDKKey{}).
		Where("id = ?", keyID).
		Updates(updates).Error
}

func (r *sdkKeyRepository) GetExpiredKeys(ctx context.Context) ([]*models.SDKKey, error) {
	var keys []*models.SDKKey
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ? AND status = ?", now, models.SDKKeyStatusActive).
		Find(&keys).Error
	return keys, err
}

func (r *sdkKeyRepository) MarkAsExpired(ctx context.Context, ids []uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.SDKKey{}).
		Where("id IN ?", ids).
		Update("status", models.SDKKeyStatusExpired).Error
}

// SDKUsage Repository

type sdkUsageRepository struct {
	db *gorm.DB
}

func NewSDKUsageRepository(db *gorm.DB) SDKUsageRepository {
	return &sdkUsageRepository{db: db}
}

func (r *sdkUsageRepository) Create(ctx context.Context, usage *models.SDKUsage) error {
	return r.db.WithContext(ctx).Create(usage).Error
}

func (r *sdkUsageRepository) BulkCreate(ctx context.Context, usages []*models.SDKUsage) error {
	return r.db.WithContext(ctx).CreateInBatches(usages, 100).Error
}

func (r *sdkUsageRepository) ListByClient(ctx context.Context, clientID uuid.UUID, startDate, endDate time.Time, page, pageSize int) ([]*models.SDKUsage, int64, error) {
	var usages []*models.SDKUsage
	var total int64

	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Model(&models.SDKUsage{}).
		Where("client_id = ?", clientID)

	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order("timestamp DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&usages).Error; err != nil {
		return nil, 0, err
	}

	return usages, total, nil
}

func (r *sdkUsageRepository) ListByKey(ctx context.Context, keyID uuid.UUID, startDate, endDate time.Time, page, pageSize int) ([]*models.SDKUsage, int64, error) {
	var usages []*models.SDKUsage
	var total int64

	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Model(&models.SDKUsage{}).
		Where("key_id = ?", keyID)

	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order("timestamp DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&usages).Error; err != nil {
		return nil, 0, err
	}

	return usages, total, nil
}

func (r *sdkUsageRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, page, pageSize int) ([]*models.SDKUsage, int64, error) {
	var usages []*models.SDKUsage
	var total int64

	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Model(&models.SDKUsage{}).
		Where("tenant_id = ?", tenantID)

	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order("timestamp DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&usages).Error; err != nil {
		return nil, 0, err
	}

	return usages, total, nil
}

func (r *sdkUsageRepository) GetStats(ctx context.Context, clientID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	var result struct {
		TotalRequests    int64
		TotalErrors      int64
		AvgResponseTime  float64
		TotalDataSent    int64
		TotalDataReceived int64
	}

	query := r.db.WithContext(ctx).Model(&models.SDKUsage{}).
		Where("client_id = ?", clientID)

	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	err := query.
		Select("COUNT(*) as total_requests, "+
			"SUM(CASE WHEN is_error THEN 1 ELSE 0 END) as total_errors, "+
			"AVG(response_time) as avg_response_time, "+
			"SUM(request_size) as total_data_sent, "+
			"SUM(response_size) as total_data_received").
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	errorRate := float64(0)
	if result.TotalRequests > 0 {
		errorRate = float64(result.TotalErrors) / float64(result.TotalRequests) * 100
	}

	return map[string]interface{}{
		"total_requests":     result.TotalRequests,
		"total_errors":       result.TotalErrors,
		"error_rate":         errorRate,
		"avg_response_time":  result.AvgResponseTime,
		"total_data_sent":    result.TotalDataSent,
		"total_data_received": result.TotalDataReceived,
	}, nil
}

func (r *sdkUsageRepository) GetTopEndpoints(ctx context.Context, clientID uuid.UUID, startDate, endDate time.Time, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := r.db.WithContext(ctx).Model(&models.SDKUsage{}).
		Where("client_id = ?", clientID)

	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	err := query.
		Select("endpoint, "+
			"COUNT(*) as request_count, "+
			"SUM(CASE WHEN is_error THEN 1 ELSE 0 END) as error_count, "+
			"AVG(response_time) as avg_response_time").
		Group("endpoint").
		Order("request_count DESC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

func (r *sdkUsageRepository) GetErrorsByType(ctx context.Context, clientID uuid.UUID, startDate, endDate time.Time) (map[string]int64, error) {
	var results []struct {
		ErrorCode string
		Count     int64
	}

	query := r.db.WithContext(ctx).Model(&models.SDKUsage{}).
		Where("client_id = ? AND is_error = ?", clientID, true)

	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	err := query.
		Select("error_code, COUNT(*) as count").
		Group("error_code").
		Order("count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	errorMap := make(map[string]int64)
	for _, r := range results {
		errorMap[r.ErrorCode] = r.Count
	}

	return errorMap, nil
}

func (r *sdkUsageRepository) GetRequestsByDay(ctx context.Context, clientID uuid.UUID, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := r.db.WithContext(ctx).Model(&models.SDKUsage{}).
		Where("client_id = ?", clientID)

	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	err := query.
		Select("DATE(timestamp) as date, "+
			"COUNT(*) as requests, "+
			"SUM(CASE WHEN is_error THEN 1 ELSE 0 END) as errors").
		Group("DATE(timestamp)").
		Order("date ASC").
		Scan(&results).Error

	return results, err
}

func (r *sdkUsageRepository) GetGeographicDistribution(ctx context.Context, clientID uuid.UUID, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := r.db.WithContext(ctx).Model(&models.SDKUsage{}).
		Where("client_id = ? AND country IS NOT NULL AND country != ''", clientID)

	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	err := query.
		Select("country, COUNT(*) as request_count").
		Group("country").
		Order("request_count DESC").
		Limit(20).
		Scan(&results).Error

	return results, err
}
