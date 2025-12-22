package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SDKService defines the interface for SDK operations
type SDKService interface {
	// SDK Client operations
	CreateClient(ctx context.Context, tenantID uuid.UUID, req *dto.CreateSDKClientRequest) (*dto.SDKClientResponse, error)
	GetClient(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.SDKClientResponse, error)
	UpdateClient(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateSDKClientRequest) (*dto.SDKClientResponse, error)
	DeleteClient(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
	ListClients(ctx context.Context, tenantID uuid.UUID, filter *dto.SDKClientFilter) (*dto.SDKClientListResponse, error)

	// SDK Key operations
	CreateKey(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, req *dto.CreateSDKKeyRequest) (*dto.SDKKeyWithSecretResponse, error)
	GetKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.SDKKeyResponse, error)
	UpdateKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateSDKKeyRequest) (*dto.SDKKeyResponse, error)
	RevokeKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, userID uuid.UUID, req *dto.RevokeSDKKeyRequest) error
	RotateKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, userID uuid.UUID, req *dto.RotateSDKKeyRequest) (*dto.SDKKeyWithSecretResponse, error)
	ListKeys(ctx context.Context, tenantID uuid.UUID, filter *dto.SDKKeyFilter) (*dto.SDKKeyListResponse, error)

	// Key validation
	ValidateKey(ctx context.Context, apiKey string) (*models.SDKKey, error)

	// Usage tracking
	RecordUsage(ctx context.Context, usage *models.SDKUsage) error
	GetUsageStats(ctx context.Context, clientID uuid.UUID, tenantID uuid.UUID, startDate, endDate time.Time) (*dto.SDKUsageStatsResponse, error)
	ListUsage(ctx context.Context, tenantID uuid.UUID, filter *dto.SDKUsageFilter) (*dto.SDKUsageListResponse, error)
}

type sdkService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewSDKService creates a new SDK service
func NewSDKService(repos *repository.Repositories, logger log.AllLogger) SDKService {
	return &sdkService{
		repos:  repos,
		logger: logger,
	}
}

func (s *sdkService) CreateClient(ctx context.Context, tenantID uuid.UUID, req *dto.CreateSDKClientRequest) (*dto.SDKClientResponse, error) {
	// Verify tenant exists
	_, err := s.repos.Tenant.GetByID(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get tenant", "tenant_id", tenantID, "error", err)
		return nil, errors.NewNotFoundError("tenant")
	}

	// Create SDK client
	client := &models.SDKClient{
		TenantID:        tenantID,
		Name:            req.Name,
		Description:     req.Description,
		Platform:        req.Platform,
		Environment:     req.Environment,
		AppIdentifier:   req.AppIdentifier,
		AppVersion:      req.AppVersion,
		AllowedOrigins:  req.AllowedOrigins,
		AllowedIPs:      req.AllowedIPs,
		WebhookURL:      req.WebhookURL,
		Configuration:   req.Configuration,
		Scopes:          req.Scopes,
		Permissions:     req.Permissions,
		RateLimitConfig: req.RateLimitConfig,
		Tags:            req.Tags,
		IsActive:        req.IsActive,
	}

	if err := s.repos.SDKClient.Create(ctx, client); err != nil {
		s.logger.Error("failed to create SDK client", "error", err)
		return nil, errors.NewInternalError("failed to create SDK client", err)
	}

	// Reload with relationships
	created, err := s.repos.SDKClient.GetByID(ctx, client.ID)
	if err != nil {
		s.logger.Error("failed to reload SDK client", "error", err)
		return nil, errors.NewInternalError("failed to reload SDK client", err)
	}

	return dto.ToSDKClientResponse(created), nil
}

func (s *sdkService) GetClient(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.SDKClientResponse, error) {
	client, err := s.repos.SDKClient.GetByTenantID(ctx, tenantID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("SDK client")
		}
		s.logger.Error("failed to get SDK client", "id", id, "error", err)
		return nil, errors.NewInternalError("failed to get SDK client", err)
	}

	return dto.ToSDKClientResponse(client), nil
}

func (s *sdkService) UpdateClient(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateSDKClientRequest) (*dto.SDKClientResponse, error) {
	client, err := s.repos.SDKClient.GetByTenantID(ctx, tenantID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("SDK client")
		}
		s.logger.Error("failed to get SDK client", "id", id, "error", err)
		return nil, errors.NewInternalError("failed to get SDK client", err)
	}

	// Update fields
	if req.Name != nil {
		client.Name = *req.Name
	}
	if req.Description != nil {
		client.Description = *req.Description
	}
	if req.AppIdentifier != nil {
		client.AppIdentifier = *req.AppIdentifier
	}
	if req.AppVersion != nil {
		client.AppVersion = *req.AppVersion
	}
	if req.AllowedOrigins != nil {
		client.AllowedOrigins = *req.AllowedOrigins
	}
	if req.AllowedIPs != nil {
		client.AllowedIPs = *req.AllowedIPs
	}
	if req.WebhookURL != nil {
		client.WebhookURL = *req.WebhookURL
	}
	if req.Configuration != nil {
		client.Configuration = *req.Configuration
	}
	if req.Scopes != nil {
		client.Scopes = *req.Scopes
	}
	if req.Permissions != nil {
		client.Permissions = *req.Permissions
	}
	if req.RateLimitConfig != nil {
		client.RateLimitConfig = *req.RateLimitConfig
	}
	if req.Tags != nil {
		client.Tags = *req.Tags
	}
	if req.IsActive != nil {
		client.IsActive = *req.IsActive
	}

	if err := s.repos.SDKClient.Update(ctx, client); err != nil {
		s.logger.Error("failed to update SDK client", "error", err)
		return nil, errors.NewInternalError("failed to update SDK client", err)
	}

	updated, err := s.repos.SDKClient.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to reload SDK client", "error", err)
		return nil, errors.NewInternalError("failed to reload SDK client", err)
	}

	return dto.ToSDKClientResponse(updated), nil
}

func (s *sdkService) DeleteClient(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	client, err := s.repos.SDKClient.GetByTenantID(ctx, tenantID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError("SDK client")
		}
		s.logger.Error("failed to get SDK client", "id", id, "error", err)
		return errors.NewInternalError("failed to get SDK client", err)
	}

	if err := s.repos.SDKClient.Delete(ctx, client.ID); err != nil {
		s.logger.Error("failed to delete SDK client", "error", err)
		return errors.NewInternalError("failed to delete SDK client", err)
	}

	return nil
}

func (s *sdkService) ListClients(ctx context.Context, tenantID uuid.UUID, filter *dto.SDKClientFilter) (*dto.SDKClientListResponse, error) {
	var clients []*models.SDKClient
	var total int64
	var err error

	if filter.Platform != nil {
		clients, total, err = s.repos.SDKClient.ListByPlatform(ctx, tenantID, *filter.Platform, filter.Page, filter.PageSize)
	} else if filter.Environment != nil {
		clients, total, err = s.repos.SDKClient.ListByEnvironment(ctx, tenantID, *filter.Environment, filter.Page, filter.PageSize)
	} else {
		clients, total, err = s.repos.SDKClient.ListByTenant(ctx, tenantID, filter.Page, filter.PageSize)
	}

	if err != nil {
		s.logger.Error("failed to list SDK clients", "error", err)
		return nil, errors.NewInternalError("failed to list SDK clients", err)
	}

	totalPages := int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize))

	return &dto.SDKClientListResponse{
		Clients:     dto.ToSDKClientResponses(clients),
		Page:        filter.Page,
		PageSize:    filter.PageSize,
		TotalItems:  total,
		TotalPages:  totalPages,
		HasNext:     filter.Page < totalPages,
		HasPrevious: filter.Page > 1,
	}, nil
}

func (s *sdkService) CreateKey(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, req *dto.CreateSDKKeyRequest) (*dto.SDKKeyWithSecretResponse, error) {
	// Verify client exists and belongs to tenant
	_, err := s.repos.SDKClient.GetByTenantID(ctx, tenantID, req.ClientID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("SDK client")
		}
		s.logger.Error("failed to get SDK client", "client_id", req.ClientID, "error", err)
		return nil, errors.NewInternalError("failed to get SDK client", err)
	}

	// Generate API key
	apiKey, keyHash, keyPrefix, err := s.generateAPIKey()
	if err != nil {
		s.logger.Error("failed to generate API key", "error", err)
		return nil, errors.NewInternalError("failed to generate API key", err)
	}

	// Calculate expiration
	var expiresAt *time.Time
	if req.ExpiresInDays != nil {
		expiry := time.Now().AddDate(0, 0, *req.ExpiresInDays)
		expiresAt = &expiry
	}

	// Create SDK key
	key := &models.SDKKey{
		ClientID:        req.ClientID,
		TenantID:        tenantID,
		Name:            req.Name,
		Description:     req.Description,
		KeyHash:         keyHash,
		KeyPrefix:       keyPrefix,
		Environment:     req.Environment,
		Scopes:          req.Scopes,
		RateLimitConfig: req.RateLimitConfig,
		ExpiresAt:       expiresAt,
		Status:          models.SDKKeyStatusActive,
		CreatedBy:       userID,
	}

	if err := s.repos.SDKKey.Create(ctx, key); err != nil {
		s.logger.Error("failed to create SDK key", "error", err)
		return nil, errors.NewInternalError("failed to create SDK key", err)
	}

	// Reload with relationships
	created, err := s.repos.SDKKey.GetByID(ctx, key.ID)
	if err != nil {
		s.logger.Error("failed to reload SDK key", "error", err)
		return nil, errors.NewInternalError("failed to reload SDK key", err)
	}

	response := dto.ToSDKKeyResponse(created)
	return &dto.SDKKeyWithSecretResponse{
		SDKKeyResponse: response,
		APIKey:         apiKey, // Only returned once at creation
	}, nil
}

func (s *sdkService) GetKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.SDKKeyResponse, error) {
	key, err := s.repos.SDKKey.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("SDK key")
		}
		s.logger.Error("failed to get SDK key", "id", id, "error", err)
		return nil, errors.NewInternalError("failed to get SDK key", err)
	}

	// Verify tenant access
	if key.TenantID != tenantID {
		return nil, errors.NewForbiddenError("SDK key does not belong to your tenant")
	}

	return dto.ToSDKKeyResponse(key), nil
}

func (s *sdkService) UpdateKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateSDKKeyRequest) (*dto.SDKKeyResponse, error) {
	key, err := s.repos.SDKKey.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("SDK key")
		}
		s.logger.Error("failed to get SDK key", "id", id, "error", err)
		return nil, errors.NewInternalError("failed to get SDK key", err)
	}

	// Verify tenant access
	if key.TenantID != tenantID {
		return nil, errors.NewForbiddenError("SDK key does not belong to your tenant")
	}

	// Update fields
	if req.Name != nil {
		key.Name = *req.Name
	}
	if req.Description != nil {
		key.Description = *req.Description
	}
	if req.Scopes != nil {
		key.Scopes = *req.Scopes
	}
	if req.RateLimitConfig != nil {
		key.RateLimitConfig = req.RateLimitConfig
	}
	if req.Status != nil {
		key.Status = *req.Status
	}

	if err := s.repos.SDKKey.Update(ctx, key); err != nil {
		s.logger.Error("failed to update SDK key", "error", err)
		return nil, errors.NewInternalError("failed to update SDK key", err)
	}

	updated, err := s.repos.SDKKey.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to reload SDK key", "error", err)
		return nil, errors.NewInternalError("failed to reload SDK key", err)
	}

	return dto.ToSDKKeyResponse(updated), nil
}

func (s *sdkService) RevokeKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, userID uuid.UUID, req *dto.RevokeSDKKeyRequest) error {
	key, err := s.repos.SDKKey.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError("SDK key")
		}
		s.logger.Error("failed to get SDK key", "id", id, "error", err)
		return errors.NewInternalError("failed to get SDK key", err)
	}

	// Verify tenant access
	if key.TenantID != tenantID {
		return errors.NewForbiddenError("SDK key does not belong to your tenant")
	}

	if err := s.repos.SDKKey.Revoke(ctx, id, userID, req.Reason); err != nil {
		s.logger.Error("failed to revoke SDK key", "error", err)
		return errors.NewInternalError("failed to revoke SDK key", err)
	}

	return nil
}

func (s *sdkService) RotateKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, userID uuid.UUID, req *dto.RotateSDKKeyRequest) (*dto.SDKKeyWithSecretResponse, error) {
	// Get existing key
	oldKey, err := s.repos.SDKKey.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("SDK key")
		}
		s.logger.Error("failed to get SDK key", "id", id, "error", err)
		return nil, errors.NewInternalError("failed to get SDK key", err)
	}

	// Verify tenant access
	if oldKey.TenantID != tenantID {
		return nil, errors.NewForbiddenError("SDK key does not belong to your tenant")
	}

	// Generate new API key
	apiKey, keyHash, keyPrefix, err := s.generateAPIKey()
	if err != nil {
		s.logger.Error("failed to generate API key", "error", err)
		return nil, errors.NewInternalError("failed to generate API key", err)
	}

	// Calculate expiration
	var expiresAt *time.Time
	if req.ExpiresInDays != nil {
		expiry := time.Now().AddDate(0, 0, *req.ExpiresInDays)
		expiresAt = &expiry
	} else if oldKey.ExpiresAt != nil {
		// Keep same expiration duration
		duration := oldKey.ExpiresAt.Sub(oldKey.CreatedAt)
		expiry := time.Now().Add(duration)
		expiresAt = &expiry
	}

	// Create new key with same settings
	newKey := &models.SDKKey{
		ClientID:        oldKey.ClientID,
		TenantID:        oldKey.TenantID,
		Name:            fmt.Sprintf("%s (rotated)", oldKey.Name),
		Description:     oldKey.Description,
		KeyHash:         keyHash,
		KeyPrefix:       keyPrefix,
		Environment:     oldKey.Environment,
		Scopes:          oldKey.Scopes,
		RateLimitConfig: oldKey.RateLimitConfig,
		ExpiresAt:       expiresAt,
		Status:          models.SDKKeyStatusActive,
		CreatedBy:       userID,
	}

	if err := s.repos.SDKKey.Create(ctx, newKey); err != nil {
		s.logger.Error("failed to create new SDK key", "error", err)
		return nil, errors.NewInternalError("failed to create new SDK key", err)
	}

	// Revoke old key
	if err := s.repos.SDKKey.Revoke(ctx, oldKey.ID, userID, "Key rotated"); err != nil {
		s.logger.Error("failed to revoke old SDK key", "error", err)
		// Continue anyway - new key was created
	}

	// Reload with relationships
	created, err := s.repos.SDKKey.GetByID(ctx, newKey.ID)
	if err != nil {
		s.logger.Error("failed to reload SDK key", "error", err)
		return nil, errors.NewInternalError("failed to reload SDK key", err)
	}

	response := dto.ToSDKKeyResponse(created)
	return &dto.SDKKeyWithSecretResponse{
		SDKKeyResponse: response,
		APIKey:         apiKey,
	}, nil
}

func (s *sdkService) ListKeys(ctx context.Context, tenantID uuid.UUID, filter *dto.SDKKeyFilter) (*dto.SDKKeyListResponse, error) {
	var keys []*models.SDKKey
	var total int64
	var err error

	if filter.ClientID != nil {
		keys, total, err = s.repos.SDKKey.ListByClient(ctx, *filter.ClientID, filter.Page, filter.PageSize)
	} else if filter.Status != nil {
		keys, total, err = s.repos.SDKKey.ListByStatus(ctx, tenantID, *filter.Status, filter.Page, filter.PageSize)
	} else {
		keys, total, err = s.repos.SDKKey.ListByTenant(ctx, tenantID, filter.Page, filter.PageSize)
	}

	if err != nil {
		s.logger.Error("failed to list SDK keys", "error", err)
		return nil, errors.NewInternalError("failed to list SDK keys", err)
	}

	totalPages := int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize))

	return &dto.SDKKeyListResponse{
		Keys:        dto.ToSDKKeyResponses(keys),
		Page:        filter.Page,
		PageSize:    filter.PageSize,
		TotalItems:  total,
		TotalPages:  totalPages,
		HasNext:     filter.Page < totalPages,
		HasPrevious: filter.Page > 1,
	}, nil
}

func (s *sdkService) ValidateKey(ctx context.Context, apiKey string) (*models.SDKKey, error) {
	// Hash the provided key
	keyHash := s.hashKey(apiKey)

	// Get key by hash
	key, err := s.repos.SDKKey.GetByKeyHash(ctx, keyHash)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewUnauthorizedError("invalid API key")
		}
		s.logger.Error("failed to get SDK key by hash", "error", err)
		return nil, errors.NewInternalError("failed to validate API key", err)
	}

	// Check if key is valid
	if !key.IsValid() {
		return nil, errors.NewUnauthorizedError("API key is not active or has expired")
	}

	return key, nil
}

func (s *sdkService) RecordUsage(ctx context.Context, usage *models.SDKUsage) error {
	if err := s.repos.SDKUsage.Create(ctx, usage); err != nil {
		s.logger.Error("failed to record SDK usage", "error", err)
		return errors.NewInternalError("failed to record SDK usage", err)
	}

	// Update client stats
	if err := s.repos.SDKClient.UpdateUsageStats(ctx, usage.ClientID, usage.IsError); err != nil {
		s.logger.Error("failed to update client stats", "error", err)
	}

	// Update key stats if key was used
	if usage.KeyID != nil {
		if err := s.repos.SDKKey.UpdateUsageStats(ctx, *usage.KeyID, usage.IPAddress, usage.IsError); err != nil {
			s.logger.Error("failed to update key stats", "error", err)
		}
	}

	return nil
}

func (s *sdkService) GetUsageStats(ctx context.Context, clientID uuid.UUID, tenantID uuid.UUID, startDate, endDate time.Time) (*dto.SDKUsageStatsResponse, error) {
	// Verify client belongs to tenant
	_, err := s.repos.SDKClient.GetByTenantID(ctx, tenantID, clientID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("SDK client")
		}
		s.logger.Error("failed to get SDK client", "client_id", clientID, "error", err)
		return nil, errors.NewInternalError("failed to get SDK client", err)
	}

	// Get basic stats
	stats, err := s.repos.SDKUsage.GetStats(ctx, clientID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get usage stats", "error", err)
		return nil, errors.NewInternalError("failed to get usage stats", err)
	}

	// Get top endpoints
	topEndpoints, err := s.repos.SDKUsage.GetTopEndpoints(ctx, clientID, startDate, endDate, 10)
	if err != nil {
		s.logger.Error("failed to get top endpoints", "error", err)
		return nil, errors.NewInternalError("failed to get top endpoints", err)
	}

	// Get errors by type
	errorsByType, err := s.repos.SDKUsage.GetErrorsByType(ctx, clientID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get errors by type", "error", err)
		return nil, errors.NewInternalError("failed to get errors by type", err)
	}

	// Get requests by day
	requestsByDay, err := s.repos.SDKUsage.GetRequestsByDay(ctx, clientID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get requests by day", "error", err)
		return nil, errors.NewInternalError("failed to get requests by day", err)
	}

	// Get geographic distribution
	geoDist, err := s.repos.SDKUsage.GetGeographicDistribution(ctx, clientID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get geographic distribution", "error", err)
		return nil, errors.NewInternalError("failed to get geographic distribution", err)
	}

	// Convert to response DTOs
	endpointStats := make([]dto.EndpointStat, len(topEndpoints))
	for i, ep := range topEndpoints {
		endpointStats[i] = dto.EndpointStat{
			Endpoint:        ep["endpoint"].(string),
			RequestCount:    ep["request_count"].(int64),
			ErrorCount:      ep["error_count"].(int64),
			AvgResponseTime: ep["avg_response_time"].(float64),
		}
	}

	dailyStats := make([]dto.DailyStat, len(requestsByDay))
	for i, day := range requestsByDay {
		dailyStats[i] = dto.DailyStat{
			Date:     day["date"].(time.Time),
			Requests: day["requests"].(int64),
			Errors:   day["errors"].(int64),
		}
	}

	geoStats := make([]dto.GeographicStat, len(geoDist))
	for i, geo := range geoDist {
		geoStats[i] = dto.GeographicStat{
			Country:      geo["country"].(string),
			RequestCount: geo["request_count"].(int64),
		}
	}

	// Build status code distribution
	statusCodeDist := make(map[string]int64)
	// This would need a separate query - simplified for now

	return &dto.SDKUsageStatsResponse{
		ClientID:          clientID,
		TotalRequests:     stats["total_requests"].(int64),
		TotalErrors:       stats["total_errors"].(int64),
		ErrorRate:         stats["error_rate"].(float64),
		AvgResponseTime:   stats["avg_response_time"].(float64),
		TotalDataSent:     stats["total_data_sent"].(int64),
		TotalDataReceived: stats["total_data_received"].(int64),
		TopEndpoints:      endpointStats,
		StatusCodeDist:    statusCodeDist,
		RequestsByDay:     dailyStats,
		ErrorsByType:      errorsByType,
		GeographicDist:    geoStats,
	}, nil
}

func (s *sdkService) ListUsage(ctx context.Context, tenantID uuid.UUID, filter *dto.SDKUsageFilter) (*dto.SDKUsageListResponse, error) {
	var usages []*models.SDKUsage
	var total int64
	var err error

	startDate := time.Time{}
	endDate := time.Time{}
	if filter.StartDate != nil {
		startDate = *filter.StartDate
	}
	if filter.EndDate != nil {
		endDate = *filter.EndDate
	}

	if filter.ClientID != nil {
		usages, total, err = s.repos.SDKUsage.ListByClient(ctx, *filter.ClientID, startDate, endDate, filter.Page, filter.PageSize)
	} else if filter.KeyID != nil {
		usages, total, err = s.repos.SDKUsage.ListByKey(ctx, *filter.KeyID, startDate, endDate, filter.Page, filter.PageSize)
	} else {
		usages, total, err = s.repos.SDKUsage.ListByTenant(ctx, tenantID, startDate, endDate, filter.Page, filter.PageSize)
	}

	if err != nil {
		s.logger.Error("failed to list SDK usage", "error", err)
		return nil, errors.NewInternalError("failed to list SDK usage", err)
	}

	totalPages := int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize))

	return &dto.SDKUsageListResponse{
		Usage:       dto.ToSDKUsageResponses(usages),
		Page:        filter.Page,
		PageSize:    filter.PageSize,
		TotalItems:  total,
		TotalPages:  totalPages,
		HasNext:     filter.Page < totalPages,
		HasPrevious: filter.Page > 1,
	}, nil
}

// Helper functions

func (s *sdkService) generateAPIKey() (apiKey, keyHash, keyPrefix string, err error) {
	// Generate 32 random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", "", err
	}

	// Encode to base64
	apiKey = "sk_" + base64.RawURLEncoding.EncodeToString(b)

	// Hash the key for storage
	keyHash = s.hashKey(apiKey)

	// Get prefix for display (first 12 characters)
	keyPrefix = apiKey[:min(12, len(apiKey))]

	return apiKey, keyHash, keyPrefix, nil
}

func (s *sdkService) hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
