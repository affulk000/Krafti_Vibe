package service

import (
	"context"
	"fmt"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

// SystemSettingService defines the interface for system settings service operations
type SystemSettingService interface {
	// CRUD Operations
	CreateSetting(ctx context.Context, modifiedBy uuid.UUID, req *dto.CreateSettingRequest) (*dto.SystemSettingResponse, error)
	GetSetting(ctx context.Context, id uuid.UUID) (*dto.SystemSettingResponse, error)
	GetSettingByKey(ctx context.Context, key string) (*dto.SystemSettingResponse, error)
	UpdateSetting(ctx context.Context, id uuid.UUID, modifiedBy uuid.UUID, req *dto.UpdateSettingRequest) (*dto.SystemSettingResponse, error)
	DeleteSetting(ctx context.Context, id uuid.UUID, modifiedBy uuid.UUID) error
	DeleteSettingByKey(ctx context.Context, key string, modifiedBy uuid.UUID) error

	// Value Retrieval
	GetStringValue(ctx context.Context, key string, defaultValue string) (string, error)
	GetBoolValue(ctx context.Context, key string, defaultValue bool) (bool, error)
	GetIntValue(ctx context.Context, key string, defaultValue int) (int, error)
	GetFloatValue(ctx context.Context, key string, defaultValue float64) (float64, error)
	GetJSONValue(ctx context.Context, key string, dest any) error

	// Setting Management
	SetSetting(ctx context.Context, key, value string, settingType models.SettingType, modifiedBy uuid.UUID) error
	BulkSetSettings(ctx context.Context, req *dto.BulkSetSettingsRequest, modifiedBy uuid.UUID) error
	BulkDeleteSettings(ctx context.Context, keys []string, modifiedBy uuid.UUID) error
	KeyExists(ctx context.Context, key string) (bool, error)

	// Category & Group Operations
	GetByCategory(ctx context.Context, category string) ([]*dto.SystemSettingResponse, error)
	GetByGroup(ctx context.Context, group string) ([]*dto.SystemSettingResponse, error)
	GetPublicSettings(ctx context.Context) ([]*dto.SystemSettingResponse, error)
	GetPrivateSettings(ctx context.Context) ([]*dto.SystemSettingResponse, error)
	GetAllCategories(ctx context.Context) ([]string, error)
	GetAllGroups(ctx context.Context) ([]string, error)
	GetCategoriesWithCount(ctx context.Context) ([]*dto.CategoryWithCountResponse, error)

	// Advanced Queries
	ListSettings(ctx context.Context, filter *dto.SettingFilter) (*dto.SystemSettingListResponse, error)
	SearchSettings(ctx context.Context, query string, page, pageSize int) (*dto.SystemSettingListResponse, error)
	GetRecentChanges(ctx context.Context, hours int, page, pageSize int) (*dto.SystemSettingListResponse, error)
	GetSettingsByModifier(ctx context.Context, userID uuid.UUID, page, pageSize int) (*dto.SystemSettingListResponse, error)

	// Security Operations
	EncryptSetting(ctx context.Context, key string, modifiedBy uuid.UUID) error
	DecryptSetting(ctx context.Context, key string, modifiedBy uuid.UUID) error
	TogglePublic(ctx context.Context, key string, isPublic bool, modifiedBy uuid.UUID) error

	// Statistics & Analytics
	GetSettingStats(ctx context.Context) (*dto.SettingStatsResponse, error)
	GetCategoryStats(ctx context.Context, category string) (*dto.CategoryStatsResponse, error)

	// Import/Export & Backup
	ExportSettings(ctx context.Context, category string) (map[string]any, error)
	ImportSettings(ctx context.Context, settings map[string]any, overwrite bool, modifiedBy uuid.UUID) error
	BackupSettings(ctx context.Context) ([]byte, error)
	RestoreSettings(ctx context.Context, backup []byte, modifiedBy uuid.UUID) error

	// Cache Management
	RefreshCache(ctx context.Context, key string) error
	RefreshCategoryCache(ctx context.Context, category string) error
	ClearSettingsCache(ctx context.Context) error

	// Validation
	ValidateSetting(ctx context.Context, key, value string) error
}

// systemSettingService implements SystemSettingService
type systemSettingService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewSystemSettingService creates a new SystemSettingService instance
func NewSystemSettingService(repos *repository.Repositories, logger log.AllLogger) SystemSettingService {
	return &systemSettingService{
		repos:  repos,
		logger: logger,
	}
}

// CreateSetting creates a new system setting
func (s *systemSettingService) CreateSetting(ctx context.Context, modifiedBy uuid.UUID, req *dto.CreateSettingRequest) (*dto.SystemSettingResponse, error) {
	s.logger.Info("creating system setting", "key", req.Key, "modified_by", modifiedBy)

	// Check if key already exists
	exists, err := s.repos.SystemSetting.KeyExists(ctx, req.Key)
	if err != nil {
		s.logger.Error("failed to check key existence", "key", req.Key, "error", err)
		return nil, errors.NewServiceError("CHECK_FAILED", "Failed to check key existence", err)
	}

	if exists {
		return nil, errors.NewValidationError(fmt.Sprintf("Setting with key '%s' already exists", req.Key))
	}

	// Create setting model
	setting := &models.SystemSetting{
		Key:             req.Key,
		Value:           req.Value,
		Type:            req.Type,
		Description:     req.Description,
		Category:        req.Category,
		Group:           req.Group,
		IsPublic:        req.IsPublic,
		IsEncrypted:     req.IsEncrypted,
		ValidationRules: req.ValidationRules,
		LastModifiedBy:  &modifiedBy,
	}

	// Save to database
	if err := s.repos.SystemSetting.Create(ctx, setting); err != nil {
		s.logger.Error("failed to create setting", "key", req.Key, "error", err)
		return nil, errors.NewServiceError("CREATE_FAILED", "Failed to create setting", err)
	}

	s.logger.Info("setting created", "setting_id", setting.ID, "key", req.Key)

	// Load with relationships
	created, err := s.repos.SystemSetting.GetByID(ctx, setting.ID)
	if err != nil {
		s.logger.Error("failed to load setting with relationships", "setting_id", setting.ID, "error", err)
		return dto.ToSystemSettingResponse(setting), nil
	}

	return dto.ToSystemSettingResponse(created), nil
}

// GetSetting retrieves a setting by ID
func (s *systemSettingService) GetSetting(ctx context.Context, id uuid.UUID) (*dto.SystemSettingResponse, error) {
	s.logger.Info("retrieving setting", "setting_id", id)

	setting, err := s.repos.SystemSetting.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("setting not found", "setting_id", id, "error", err)
		return nil, errors.NewNotFoundError("setting")
	}

	return dto.ToSystemSettingResponse(setting), nil
}

// GetSettingByKey retrieves a setting by key
func (s *systemSettingService) GetSettingByKey(ctx context.Context, key string) (*dto.SystemSettingResponse, error) {
	s.logger.Info("retrieving setting by key", "key", key)

	if key == "" {
		return nil, errors.NewValidationError("Key cannot be empty")
	}

	setting, err := s.repos.SystemSetting.GetByKey(ctx, key)
	if err != nil {
		s.logger.Error("setting not found", "key", key, "error", err)
		return nil, errors.NewNotFoundError("setting")
	}

	return dto.ToSystemSettingResponse(setting), nil
}

// UpdateSetting updates a system setting
func (s *systemSettingService) UpdateSetting(ctx context.Context, id uuid.UUID, modifiedBy uuid.UUID, req *dto.UpdateSettingRequest) (*dto.SystemSettingResponse, error) {
	s.logger.Info("updating setting", "setting_id", id, "modified_by", modifiedBy)

	// Get existing setting
	setting, err := s.repos.SystemSetting.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("setting not found", "setting_id", id, "error", err)
		return nil, errors.NewNotFoundError("setting")
	}

	// Update fields
	if req.Value != nil {
		setting.Value = *req.Value
	}

	if req.Description != nil {
		setting.Description = *req.Description
	}

	if req.Category != nil {
		setting.Category = *req.Category
	}

	if req.Group != nil {
		setting.Group = *req.Group
	}

	if req.IsPublic != nil {
		setting.IsPublic = *req.IsPublic
	}

	if req.ValidationRules != nil {
		if setting.ValidationRules == nil {
			setting.ValidationRules = make(models.JSONB)
		}
		for k, v := range req.ValidationRules {
			setting.ValidationRules[k] = v
		}
	}

	setting.LastModifiedBy = &modifiedBy

	// Save changes
	if err := s.repos.SystemSetting.Update(ctx, setting); err != nil {
		s.logger.Error("failed to update setting", "setting_id", id, "error", err)
		return nil, errors.NewServiceError("UPDATE_FAILED", "Failed to update setting", err)
	}

	s.logger.Info("setting updated", "setting_id", id)
	return dto.ToSystemSettingResponse(setting), nil
}

// DeleteSetting deletes a setting by ID
func (s *systemSettingService) DeleteSetting(ctx context.Context, id uuid.UUID, modifiedBy uuid.UUID) error {
	s.logger.Info("deleting setting", "setting_id", id, "modified_by", modifiedBy)

	// Verify setting exists
	_, err := s.repos.SystemSetting.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("setting not found", "setting_id", id, "error", err)
		return errors.NewNotFoundError("setting")
	}

	if err := s.repos.SystemSetting.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete setting", "setting_id", id, "error", err)
		return errors.NewServiceError("DELETE_FAILED", "Failed to delete setting", err)
	}

	s.logger.Info("setting deleted", "setting_id", id)
	return nil
}

// DeleteSettingByKey deletes a setting by key
func (s *systemSettingService) DeleteSettingByKey(ctx context.Context, key string, modifiedBy uuid.UUID) error {
	s.logger.Info("deleting setting by key", "key", key, "modified_by", modifiedBy)

	if key == "" {
		return errors.NewValidationError("Key cannot be empty")
	}

	if err := s.repos.SystemSetting.DeleteByKey(ctx, key); err != nil {
		s.logger.Error("failed to delete setting", "key", key, "error", err)
		return errors.NewServiceError("DELETE_FAILED", "Failed to delete setting", err)
	}

	s.logger.Info("setting deleted", "key", key)
	return nil
}

// GetStringValue retrieves a string setting value with default
func (s *systemSettingService) GetStringValue(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := s.repos.SystemSetting.GetStringValue(ctx, key, defaultValue)
	if err != nil {
		s.logger.Error("failed to get string value", "key", key, "error", err)
		return defaultValue, errors.NewServiceError("GET_FAILED", "Failed to get string value", err)
	}
	return value, nil
}

// GetBoolValue retrieves a boolean setting value with default
func (s *systemSettingService) GetBoolValue(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.repos.SystemSetting.GetBoolValue(ctx, key, defaultValue)
	if err != nil {
		s.logger.Error("failed to get bool value", "key", key, "error", err)
		return defaultValue, errors.NewServiceError("GET_FAILED", "Failed to get bool value", err)
	}
	return value, nil
}

// GetIntValue retrieves an integer setting value with default
func (s *systemSettingService) GetIntValue(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.repos.SystemSetting.GetIntValue(ctx, key, defaultValue)
	if err != nil {
		s.logger.Error("failed to get int value", "key", key, "error", err)
		return defaultValue, errors.NewServiceError("GET_FAILED", "Failed to get int value", err)
	}
	return value, nil
}

// GetFloatValue retrieves a float setting value with default
func (s *systemSettingService) GetFloatValue(ctx context.Context, key string, defaultValue float64) (float64, error) {
	value, err := s.repos.SystemSetting.GetFloatValue(ctx, key, defaultValue)
	if err != nil {
		s.logger.Error("failed to get float value", "key", key, "error", err)
		return defaultValue, errors.NewServiceError("GET_FAILED", "Failed to get float value", err)
	}
	return value, nil
}

// GetJSONValue retrieves a JSON setting value
func (s *systemSettingService) GetJSONValue(ctx context.Context, key string, dest any) error {
	if err := s.repos.SystemSetting.GetJSONValue(ctx, key, dest); err != nil {
		s.logger.Error("failed to get JSON value", "key", key, "error", err)
		return errors.NewServiceError("GET_FAILED", "Failed to get JSON value", err)
	}
	return nil
}

// SetSetting sets a setting value (creates or updates)
func (s *systemSettingService) SetSetting(ctx context.Context, key, value string, settingType models.SettingType, modifiedBy uuid.UUID) error {
	s.logger.Info("setting value", "key", key, "type", settingType, "modified_by", modifiedBy)

	if key == "" {
		return errors.NewValidationError("Key cannot be empty")
	}

	if err := s.repos.SystemSetting.SetSetting(ctx, key, value, settingType); err != nil {
		s.logger.Error("failed to set setting", "key", key, "error", err)
		return errors.NewServiceError("SET_FAILED", "Failed to set setting", err)
	}

	// Update last modified by
	setting, err := s.repos.SystemSetting.GetByKey(ctx, key)
	if err == nil {
		setting.LastModifiedBy = &modifiedBy
		s.repos.SystemSetting.Update(ctx, setting)
	}

	return nil
}

// BulkSetSettings sets multiple settings at once
func (s *systemSettingService) BulkSetSettings(ctx context.Context, req *dto.BulkSetSettingsRequest, modifiedBy uuid.UUID) error {
	s.logger.Info("bulk setting values", "count", len(req.Settings), "modified_by", modifiedBy)

	if len(req.Settings) == 0 {
		return errors.NewValidationError("No settings provided")
	}

	// Convert DTO to repository format
	repoSettings := make(map[string]repository.SettingValue)
	for key, val := range req.Settings {
		repoSettings[key] = repository.SettingValue{
			Value:       val.Value,
			Type:        val.Type,
			Description: val.Description,
			Category:    val.Category,
			Group:       val.Group,
			IsPublic:    val.IsPublic,
			IsEncrypted: val.IsEncrypted,
			Rules:       val.ValidationRules,
		}
	}

	if err := s.repos.SystemSetting.BulkSet(ctx, repoSettings); err != nil {
		s.logger.Error("failed to bulk set settings", "error", err)
		return errors.NewServiceError("BULK_SET_FAILED", "Failed to bulk set settings", err)
	}

	s.logger.Info("bulk set completed", "count", len(req.Settings))
	return nil
}

// BulkDeleteSettings deletes multiple settings at once
func (s *systemSettingService) BulkDeleteSettings(ctx context.Context, keys []string, modifiedBy uuid.UUID) error {
	s.logger.Info("bulk deleting settings", "count", len(keys), "modified_by", modifiedBy)

	if len(keys) == 0 {
		return errors.NewValidationError("No keys provided")
	}

	if err := s.repos.SystemSetting.BulkDeleteByKeys(ctx, keys); err != nil {
		s.logger.Error("failed to bulk delete settings", "error", err)
		return errors.NewServiceError("BULK_DELETE_FAILED", "Failed to bulk delete settings", err)
	}

	s.logger.Info("bulk delete completed", "count", len(keys))
	return nil
}

// KeyExists checks if a setting key exists
func (s *systemSettingService) KeyExists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, nil
	}

	exists, err := s.repos.SystemSetting.KeyExists(ctx, key)
	if err != nil {
		s.logger.Error("failed to check key existence", "key", key, "error", err)
		return false, errors.NewServiceError("CHECK_FAILED", "Failed to check key existence", err)
	}

	return exists, nil
}

// GetByCategory retrieves all settings in a category
func (s *systemSettingService) GetByCategory(ctx context.Context, category string) ([]*dto.SystemSettingResponse, error) {
	s.logger.Info("getting settings by category", "category", category)

	if category == "" {
		return nil, errors.NewValidationError("Category cannot be empty")
	}

	settings, err := s.repos.SystemSetting.GetByCategory(ctx, category)
	if err != nil {
		s.logger.Error("failed to get settings by category", "category", category, "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get settings by category", err)
	}

	return dto.ToSystemSettingResponses(settings), nil
}

// GetByGroup retrieves all settings in a group
func (s *systemSettingService) GetByGroup(ctx context.Context, group string) ([]*dto.SystemSettingResponse, error) {
	s.logger.Info("getting settings by group", "group", group)

	if group == "" {
		return nil, errors.NewValidationError("Group cannot be empty")
	}

	settings, err := s.repos.SystemSetting.GetByGroup(ctx, group)
	if err != nil {
		s.logger.Error("failed to get settings by group", "group", group, "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get settings by group", err)
	}

	return dto.ToSystemSettingResponses(settings), nil
}

// GetPublicSettings retrieves all public settings
func (s *systemSettingService) GetPublicSettings(ctx context.Context) ([]*dto.SystemSettingResponse, error) {
	s.logger.Info("getting public settings")

	settings, err := s.repos.SystemSetting.GetPublicSettings(ctx)
	if err != nil {
		s.logger.Error("failed to get public settings", "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get public settings", err)
	}

	return dto.ToSystemSettingResponses(settings), nil
}

// GetPrivateSettings retrieves all private settings
func (s *systemSettingService) GetPrivateSettings(ctx context.Context) ([]*dto.SystemSettingResponse, error) {
	s.logger.Info("getting private settings")

	settings, err := s.repos.SystemSetting.GetPrivateSettings(ctx)
	if err != nil {
		s.logger.Error("failed to get private settings", "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get private settings", err)
	}

	return dto.ToSystemSettingResponses(settings), nil
}

// GetAllCategories retrieves all unique categories
func (s *systemSettingService) GetAllCategories(ctx context.Context) ([]string, error) {
	s.logger.Info("getting all categories")

	categories, err := s.repos.SystemSetting.GetAllCategories(ctx)
	if err != nil {
		s.logger.Error("failed to get categories", "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get categories", err)
	}

	return categories, nil
}

// GetAllGroups retrieves all unique groups
func (s *systemSettingService) GetAllGroups(ctx context.Context) ([]string, error) {
	s.logger.Info("getting all groups")

	groups, err := s.repos.SystemSetting.GetAllGroups(ctx)
	if err != nil {
		s.logger.Error("failed to get groups", "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get groups", err)
	}

	return groups, nil
}

// GetCategoriesWithCount retrieves categories with setting counts
func (s *systemSettingService) GetCategoriesWithCount(ctx context.Context) ([]*dto.CategoryWithCountResponse, error) {
	s.logger.Info("getting categories with count")

	counts, err := s.repos.SystemSetting.GetCategoriesWithCount(ctx)
	if err != nil {
		s.logger.Error("failed to get categories with count", "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get categories with count", err)
	}

	responses := make([]*dto.CategoryWithCountResponse, len(counts))
	for i, count := range counts {
		responses[i] = &dto.CategoryWithCountResponse{
			Category: count.Category,
			Count:    count.Count,
		}
	}

	return responses, nil
}

// ListSettings retrieves settings with advanced filters
func (s *systemSettingService) ListSettings(ctx context.Context, filter *dto.SettingFilter) (*dto.SystemSettingListResponse, error) {
	s.logger.Info("listing settings with filters")

	// Build repository filters
	repoFilters := repository.SettingFilters{
		Categories:     filter.Categories,
		Groups:         filter.Groups,
		Types:          filter.Types,
		IsPublic:       filter.IsPublic,
		IsEncrypted:    filter.IsEncrypted,
		ModifiedBy:     filter.ModifiedBy,
		ModifiedAfter:  filter.ModifiedAfter,
		ModifiedBefore: filter.ModifiedBefore,
	}

	// Set defaults
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	settings, paginationResult, err := s.repos.SystemSetting.FindByFilters(ctx, repoFilters, pagination)
	if err != nil {
		s.logger.Error("failed to list settings", "error", err)
		return nil, errors.NewServiceError("LIST_FAILED", "Failed to list settings", err)
	}

	return &dto.SystemSettingListResponse{
		Settings:    dto.ToSystemSettingResponses(settings),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// SearchSettings searches settings by key or description
func (s *systemSettingService) SearchSettings(ctx context.Context, query string, page, pageSize int) (*dto.SystemSettingListResponse, error) {
	s.logger.Info("searching settings", "query", query)

	// Set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	settings, paginationResult, err := s.repos.SystemSetting.Search(ctx, query, pagination)
	if err != nil {
		s.logger.Error("failed to search settings", "error", err)
		return nil, errors.NewServiceError("SEARCH_FAILED", "Failed to search settings", err)
	}

	return &dto.SystemSettingListResponse{
		Settings:    dto.ToSystemSettingResponses(settings),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetRecentChanges retrieves recently modified settings
func (s *systemSettingService) GetRecentChanges(ctx context.Context, hours int, page, pageSize int) (*dto.SystemSettingListResponse, error) {
	s.logger.Info("getting recent changes", "hours", hours)

	if hours <= 0 {
		hours = 24
	}

	// Set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	settings, paginationResult, err := s.repos.SystemSetting.GetRecentChanges(ctx, hours, pagination)
	if err != nil {
		s.logger.Error("failed to get recent changes", "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get recent changes", err)
	}

	return &dto.SystemSettingListResponse{
		Settings:    dto.ToSystemSettingResponses(settings),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// GetSettingsByModifier retrieves settings modified by a specific user
func (s *systemSettingService) GetSettingsByModifier(ctx context.Context, userID uuid.UUID, page, pageSize int) (*dto.SystemSettingListResponse, error) {
	s.logger.Info("getting settings by modifier", "user_id", userID)

	// Set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	settings, paginationResult, err := s.repos.SystemSetting.GetSettingsByModifier(ctx, userID, pagination)
	if err != nil {
		s.logger.Error("failed to get settings by modifier", "error", err)
		return nil, errors.NewServiceError("GET_FAILED", "Failed to get settings by modifier", err)
	}

	return &dto.SystemSettingListResponse{
		Settings:    dto.ToSystemSettingResponses(settings),
		Page:        paginationResult.Page,
		PageSize:    paginationResult.PageSize,
		TotalItems:  paginationResult.TotalItems,
		TotalPages:  paginationResult.TotalPages,
		HasNext:     paginationResult.HasNext,
		HasPrevious: paginationResult.HasPrev,
	}, nil
}

// EncryptSetting encrypts a setting value
func (s *systemSettingService) EncryptSetting(ctx context.Context, key string, modifiedBy uuid.UUID) error {
	s.logger.Info("encrypting setting", "key", key, "modified_by", modifiedBy)

	if key == "" {
		return errors.NewValidationError("Key cannot be empty")
	}

	if err := s.repos.SystemSetting.EncryptSetting(ctx, key); err != nil {
		s.logger.Error("failed to encrypt setting", "key", key, "error", err)
		return errors.NewServiceError("ENCRYPT_FAILED", "Failed to encrypt setting", err)
	}

	// Update last modified by
	setting, err := s.repos.SystemSetting.GetByKey(ctx, key)
	if err == nil {
		setting.LastModifiedBy = &modifiedBy
		s.repos.SystemSetting.Update(ctx, setting)
	}

	s.logger.Info("setting encrypted", "key", key)
	return nil
}

// DecryptSetting decrypts a setting value
func (s *systemSettingService) DecryptSetting(ctx context.Context, key string, modifiedBy uuid.UUID) error {
	s.logger.Info("decrypting setting", "key", key, "modified_by", modifiedBy)

	if key == "" {
		return errors.NewValidationError("Key cannot be empty")
	}

	if err := s.repos.SystemSetting.DecryptSetting(ctx, key); err != nil {
		s.logger.Error("failed to decrypt setting", "key", key, "error", err)
		return errors.NewServiceError("DECRYPT_FAILED", "Failed to decrypt setting", err)
	}

	// Update last modified by
	setting, err := s.repos.SystemSetting.GetByKey(ctx, key)
	if err == nil {
		setting.LastModifiedBy = &modifiedBy
		s.repos.SystemSetting.Update(ctx, setting)
	}

	s.logger.Info("setting decrypted", "key", key)
	return nil
}

// TogglePublic toggles the public/private status of a setting
func (s *systemSettingService) TogglePublic(ctx context.Context, key string, isPublic bool, modifiedBy uuid.UUID) error {
	s.logger.Info("toggling public status", "key", key, "is_public", isPublic, "modified_by", modifiedBy)

	if key == "" {
		return errors.NewValidationError("Key cannot be empty")
	}

	if err := s.repos.SystemSetting.TogglePublic(ctx, key, isPublic); err != nil {
		s.logger.Error("failed to toggle public status", "key", key, "error", err)
		return errors.NewServiceError("TOGGLE_FAILED", "Failed to toggle public status", err)
	}

	// Update last modified by
	setting, err := s.repos.SystemSetting.GetByKey(ctx, key)
	if err == nil {
		setting.LastModifiedBy = &modifiedBy
		s.repos.SystemSetting.Update(ctx, setting)
	}

	s.logger.Info("public status toggled", "key", key, "is_public", isPublic)
	return nil
}

// GetSettingStats retrieves overall setting statistics
func (s *systemSettingService) GetSettingStats(ctx context.Context) (*dto.SettingStatsResponse, error) {
	s.logger.Info("getting setting stats")

	stats, err := s.repos.SystemSetting.GetSettingStats(ctx)
	if err != nil {
		s.logger.Error("failed to get setting stats", "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get setting stats", err)
	}

	return &dto.SettingStatsResponse{
		TotalSettings:     stats.TotalSettings,
		PublicSettings:    stats.PublicSettings,
		PrivateSettings:   stats.PrivateSettings,
		EncryptedSettings: stats.EncryptedSettings,
		ByType:            stats.ByType,
		ByCategory:        stats.ByCategory,
		ByGroup:           make(map[string]int64), // Not in repo stats
		RecentChanges:     stats.RecentlyModified,
		TotalCategories:   len(stats.ByCategory),
		TotalGroups:       0, // Not in repo stats
	}, nil
}

// GetCategoryStats retrieves statistics for a specific category
func (s *systemSettingService) GetCategoryStats(ctx context.Context, category string) (*dto.CategoryStatsResponse, error) {
	s.logger.Info("getting category stats", "category", category)

	if category == "" {
		return nil, errors.NewValidationError("Category cannot be empty")
	}

	stats, err := s.repos.SystemSetting.GetCategoryStats(ctx, category)
	if err != nil {
		s.logger.Error("failed to get category stats", "category", category, "error", err)
		return nil, errors.NewServiceError("STATS_FAILED", "Failed to get category stats", err)
	}

	return &dto.CategoryStatsResponse{
		Category:      stats.Category,
		TotalSettings: stats.TotalSettings,
		ByType:        stats.ByType,
		ByGroup:       make(map[string]int64), // Not in repo stats
		PublicCount:   stats.PublicCount,
		PrivateCount:  stats.TotalSettings - stats.PublicCount,
	}, nil
}

// ExportSettings exports settings to a map
func (s *systemSettingService) ExportSettings(ctx context.Context, category string) (map[string]any, error) {
	s.logger.Info("exporting settings", "category", category)

	settings, err := s.repos.SystemSetting.ExportSettings(ctx, category)
	if err != nil {
		s.logger.Error("failed to export settings", "error", err)
		return nil, errors.NewServiceError("EXPORT_FAILED", "Failed to export settings", err)
	}

	s.logger.Info("settings exported", "count", len(settings))
	return settings, nil
}

// ImportSettings imports settings from a map
func (s *systemSettingService) ImportSettings(ctx context.Context, settings map[string]any, overwrite bool, modifiedBy uuid.UUID) error {
	s.logger.Info("importing settings", "count", len(settings), "overwrite", overwrite, "modified_by", modifiedBy)

	if len(settings) == 0 {
		return errors.NewValidationError("No settings provided")
	}

	if err := s.repos.SystemSetting.ImportSettings(ctx, settings, overwrite); err != nil {
		s.logger.Error("failed to import settings", "error", err)
		return errors.NewServiceError("IMPORT_FAILED", "Failed to import settings", err)
	}

	s.logger.Info("settings imported", "count", len(settings))
	return nil
}

// BackupSettings creates a backup of all settings
func (s *systemSettingService) BackupSettings(ctx context.Context) ([]byte, error) {
	s.logger.Info("backing up settings")

	backup, err := s.repos.SystemSetting.BackupSettings(ctx)
	if err != nil {
		s.logger.Error("failed to backup settings", "error", err)
		return nil, errors.NewServiceError("BACKUP_FAILED", "Failed to backup settings", err)
	}

	s.logger.Info("settings backed up", "size", len(backup))
	return backup, nil
}

// RestoreSettings restores settings from a backup
func (s *systemSettingService) RestoreSettings(ctx context.Context, backup []byte, modifiedBy uuid.UUID) error {
	s.logger.Info("restoring settings from backup", "size", len(backup), "modified_by", modifiedBy)

	if len(backup) == 0 {
		return errors.NewValidationError("Backup data is empty")
	}

	if err := s.repos.SystemSetting.RestoreSettings(ctx, backup); err != nil {
		s.logger.Error("failed to restore settings", "error", err)
		return errors.NewServiceError("RESTORE_FAILED", "Failed to restore settings", err)
	}

	s.logger.Info("settings restored from backup")
	return nil
}

// RefreshCache refreshes the cache for a specific setting
func (s *systemSettingService) RefreshCache(ctx context.Context, key string) error {
	s.logger.Info("refreshing cache for setting", "key", key)

	if key == "" {
		return errors.NewValidationError("Key cannot be empty")
	}

	if err := s.repos.SystemSetting.RefreshCache(ctx, key); err != nil {
		s.logger.Error("failed to refresh cache", "key", key, "error", err)
		return errors.NewServiceError("CACHE_REFRESH_FAILED", "Failed to refresh cache", err)
	}

	return nil
}

// RefreshCategoryCache refreshes the cache for a category
func (s *systemSettingService) RefreshCategoryCache(ctx context.Context, category string) error {
	s.logger.Info("refreshing category cache", "category", category)

	if category == "" {
		return errors.NewValidationError("Category cannot be empty")
	}

	if err := s.repos.SystemSetting.RefreshCategoryCache(ctx, category); err != nil {
		s.logger.Error("failed to refresh category cache", "category", category, "error", err)
		return errors.NewServiceError("CACHE_REFRESH_FAILED", "Failed to refresh category cache", err)
	}

	return nil
}

// ClearSettingsCache clears all settings cache
func (s *systemSettingService) ClearSettingsCache(ctx context.Context) error {
	s.logger.Info("clearing settings cache")

	if err := s.repos.SystemSetting.ClearSettingsCache(ctx); err != nil {
		s.logger.Error("failed to clear settings cache", "error", err)
		return errors.NewServiceError("CACHE_CLEAR_FAILED", "Failed to clear settings cache", err)
	}

	s.logger.Info("settings cache cleared")
	return nil
}

// ValidateSetting validates a setting value
func (s *systemSettingService) ValidateSetting(ctx context.Context, key, value string) error {
	if key == "" {
		return errors.NewValidationError("Key cannot be empty")
	}

	if err := s.repos.SystemSetting.ValidateSetting(ctx, key, value); err != nil {
		s.logger.Error("setting validation failed", "key", key, "error", err)
		return errors.NewServiceError("VALIDATION_FAILED", "Setting validation failed", err)
	}

	return nil
}
