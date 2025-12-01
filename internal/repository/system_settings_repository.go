package repository

import (
	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SystemSettingRepository interface remains the same...
type SystemSettingRepository interface {
	BaseRepository[models.SystemSetting]
	GetByKey(ctx context.Context, key string) (*models.SystemSetting, error)
	GetByKeys(ctx context.Context, keys []string) ([]*models.SystemSetting, error)
	SetSetting(ctx context.Context, key, value string, settingType models.SettingType) error
	DeleteByKey(ctx context.Context, key string) error
	KeyExists(ctx context.Context, key string) (bool, error)
	GetStringValue(ctx context.Context, key string, defaultValue string) (string, error)
	GetBoolValue(ctx context.Context, key string, defaultValue bool) (bool, error)
	GetIntValue(ctx context.Context, key string, defaultValue int) (int, error)
	GetFloatValue(ctx context.Context, key string, defaultValue float64) (float64, error)
	GetJSONValue(ctx context.Context, key string, dest any) error
	GetByCategory(ctx context.Context, category string) ([]*models.SystemSetting, error)
	GetByGroup(ctx context.Context, group string) ([]*models.SystemSetting, error)
	GetPublicSettings(ctx context.Context) ([]*models.SystemSetting, error)
	GetPrivateSettings(ctx context.Context) ([]*models.SystemSetting, error)
	GetEncryptedSettings(ctx context.Context) ([]*models.SystemSetting, error)
	BulkSet(ctx context.Context, settings map[string]SettingValue) error
	BulkDeleteByKeys(ctx context.Context, keys []string) error
	Search(ctx context.Context, query string, pagination PaginationParams) ([]*models.SystemSetting, PaginationResult, error)
	FindByFilters(ctx context.Context, filters SettingFilters, pagination PaginationParams) ([]*models.SystemSetting, PaginationResult, error)
	GetAllCategories(ctx context.Context) ([]string, error)
	GetAllGroups(ctx context.Context) ([]string, error)
	GetCategoriesWithCount(ctx context.Context) ([]CategoryCount, error)
	ValidateSetting(ctx context.Context, key, value string) error
	EncryptSetting(ctx context.Context, key string) error
	DecryptSetting(ctx context.Context, key string) error
	TogglePublic(ctx context.Context, key string, isPublic bool) error
	GetSettingHistory(ctx context.Context, key string, limit int) ([]SettingHistory, error)
	GetRecentChanges(ctx context.Context, hours int, pagination PaginationParams) ([]*models.SystemSetting, PaginationResult, error)
	GetSettingsByModifier(ctx context.Context, userID uuid.UUID, pagination PaginationParams) ([]*models.SystemSetting, PaginationResult, error)
	GetSettingStats(ctx context.Context) (SettingStats, error)
	GetCategoryStats(ctx context.Context, category string) (CategorySettingStats, error)
	GetMostModifiedSettings(ctx context.Context, limit int) ([]SettingModificationCount, error)
	ExportSettings(ctx context.Context, category string) (map[string]any, error)
	ImportSettings(ctx context.Context, settings map[string]any, overwrite bool) error
	ResetToDefaults(ctx context.Context, category string) error
	BackupSettings(ctx context.Context) ([]byte, error)
	RestoreSettings(ctx context.Context, backup []byte) error
	RefreshCache(ctx context.Context, key string) error
	RefreshCategoryCache(ctx context.Context, category string) error
	ClearSettingsCache(ctx context.Context) error
}

type SettingValue struct {
	Value       string             `json:"value"`
	Type        models.SettingType `json:"type"`
	Description string             `json:"description,omitempty"`
	Category    string             `json:"category,omitempty"`
	Group       string             `json:"group,omitempty"`
	IsPublic    bool               `json:"is_public"`
	IsEncrypted bool               `json:"is_encrypted"`
	Rules       map[string]any     `json:"validation_rules,omitempty"`
}

type SettingFilters struct {
	Categories     []string             `json:"categories"`
	Groups         []string             `json:"groups"`
	Types          []models.SettingType `json:"types"`
	IsPublic       *bool                `json:"is_public"`
	IsEncrypted    *bool                `json:"is_encrypted"`
	ModifiedBy     *uuid.UUID           `json:"modified_by"`
	ModifiedAfter  *time.Time           `json:"modified_after"`
	ModifiedBefore *time.Time           `json:"modified_before"`
}

type SettingHistory struct {
	SettingID  uuid.UUID `json:"setting_id"`
	Key        string    `json:"key"`
	OldValue   string    `json:"old_value"`
	NewValue   string    `json:"new_value"`
	ModifiedBy uuid.UUID `json:"modified_by"`
	ModifiedAt time.Time `json:"modified_at"`
	UserEmail  string    `json:"user_email,omitempty"`
}

type SettingStats struct {
	TotalSettings      int64                        `json:"total_settings"`
	PublicSettings     int64                        `json:"public_settings"`
	PrivateSettings    int64                        `json:"private_settings"`
	EncryptedSettings  int64                        `json:"encrypted_settings"`
	ByType             map[models.SettingType]int64 `json:"by_type"`
	ByCategory         map[string]int64             `json:"by_category"`
	RecentlyModified   int64                        `json:"recently_modified_24h"`
	MostActiveCategory string                       `json:"most_active_category"`
}

type CategorySettingStats struct {
	Category       string                       `json:"category"`
	TotalSettings  int64                        `json:"total_settings"`
	ByType         map[models.SettingType]int64 `json:"by_type"`
	PublicCount    int64                        `json:"public_count"`
	EncryptedCount int64                        `json:"encrypted_count"`
	LastModified   *time.Time                   `json:"last_modified,omitempty"`
}

type SettingModificationCount struct {
	Key               string    `json:"key"`
	Category          string    `json:"category"`
	ModificationCount int64     `json:"modification_count"`
	LastModified      time.Time `json:"last_modified"`
}

type CategoryCount struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

type systemSettingRepository struct {
	BaseRepository[models.SystemSetting]
	db        *gorm.DB
	logger    log.AllLogger
	cache     Cache
	metrics   MetricsCollector
	encryptor SettingEncryptor
}

type SettingEncryptor interface {
	Encrypt(value string) (string, error)
	Decrypt(value string) (string, error)
}

func NewSystemSettingRepository(
	db *gorm.DB,
	encryptor SettingEncryptor,
	config ...RepositoryConfig,
) SystemSettingRepository {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 15 * time.Minute
	}

	baseRepo := NewBaseRepository[models.SystemSetting](db, cfg)

	return &systemSettingRepository{
		BaseRepository: baseRepo,
		db:             db,
		logger:         cfg.Logger,
		cache:          cfg.Cache,
		metrics:        cfg.Metrics,
		encryptor:      encryptor,
	}
}

func (r *systemSettingRepository) GetByKey(ctx context.Context, key string) (*models.SystemSetting, error) {
	start := time.Now()
	defer func() {
		if r.metrics != nil {
			r.metrics.RecordOperation("get_by_key", "system_settings", time.Since(start), nil)
		}
	}()

	if key == "" {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "key cannot be empty", errors.ErrInvalidInput)
	}

	cacheKey := r.getCacheKey("key", key)
	if r.cache != nil {
		var setting models.SystemSetting
		if err := r.cache.GetJSON(ctx, cacheKey, &setting); err == nil {
			if r.metrics != nil {
				r.metrics.RecordCacheHit("system_settings")
			}
			return &setting, nil
		}
		if r.metrics != nil {
			r.metrics.RecordCacheMiss("system_settings")
		}
	}

	var setting models.SystemSetting
	if err := r.db.WithContext(ctx).
		Preload("ModifiedBy").
		Where("key = ?", key).
		First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewRepositoryError("NOT_FOUND", fmt.Sprintf("setting not found: %s", key), errors.ErrNotFound)
		}
		if r.logger != nil {
			r.logger.Error("failed to get setting by key", "key", key, "error", err)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get setting", err)
	}

	if setting.IsEncrypted && r.encryptor != nil {
		decrypted, err := r.encryptor.Decrypt(setting.Value)
		if err != nil {
			if r.logger != nil {
				r.logger.Error("failed to decrypt setting", "key", key, "error", err)
			}
			return nil, errors.NewRepositoryError("DECRYPTION_FAILED", "failed to decrypt setting", err)
		}
		setting.Value = decrypted
	}

	if r.cache != nil {
		if err := r.cache.SetJSON(ctx, cacheKey, setting, 15*time.Minute); err != nil && r.logger != nil {
			r.logger.Warn("failed to cache setting", "key", key, "error", err)
		}
	}

	return &setting, nil
}

func (r *systemSettingRepository) GetByKeys(ctx context.Context, keys []string) ([]*models.SystemSetting, error) {
	if len(keys) == 0 {
		return []*models.SystemSetting{}, nil
	}

	var settings []*models.SystemSetting
	if err := r.db.WithContext(ctx).
		Preload("ModifiedBy").
		Where("key IN ?", keys).
		Find(&settings).Error; err != nil {
		if r.logger != nil {
			r.logger.Error("failed to get settings by keys", "count", len(keys), "error", err)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get settings", err)
	}

	for _, setting := range settings {
		if setting.IsEncrypted && r.encryptor != nil {
			decrypted, err := r.encryptor.Decrypt(setting.Value)
			if err != nil {
				if r.logger != nil {
					r.logger.Error("failed to decrypt setting", "key", setting.Key, "error", err)
				}
				continue
			}
			setting.Value = decrypted
		}
	}

	return settings, nil
}

func (r *systemSettingRepository) ValidateSetting(ctx context.Context, key, value string) error {
	setting, err := r.GetByKey(ctx, key)
	if err != nil {
		return err
	}

	if len(setting.ValidationRules) == 0 {
		return nil
	}

	// FIXED: ValidationRules is already map[string]any (JSONB), use directly
	rules := setting.ValidationRules
	_ = rules // Avoid unused variable warning

	// TODO: Implement actual validation logic based on rules
	// Example:
	// if minLength, ok := rules["min_length"].(float64); ok {
	//     if len(value) < int(minLength) {
	//         return errors.NewRepositoryError("VALIDATION_FAILED", "value too short", errors.ErrInvalidInput)
	//     }
	// }

	return nil
}

func (r *systemSettingRepository) EncryptSetting(ctx context.Context, key string) error {
	if r.encryptor == nil {
		return errors.NewRepositoryError("NO_ENCRYPTOR", "encryptor not configured", errors.ErrInvalidInput)
	}

	setting, err := r.GetByKey(ctx, key)
	if err != nil {
		return err
	}

	if setting.IsEncrypted {
		return nil
	}

	encrypted, err := r.encryptor.Encrypt(setting.Value)
	if err != nil {
		return errors.NewRepositoryError("ENCRYPTION_FAILED", "failed to encrypt setting", err)
	}

	setting.Value = encrypted
	setting.IsEncrypted = true

	if err := r.Update(ctx, setting); err != nil {
		return err
	}

	if r.cache != nil {
		r.cache.Delete(ctx, r.getCacheKey("key", key))
	}

	return nil
}

func (r *systemSettingRepository) DecryptSetting(ctx context.Context, key string) error {
	if r.encryptor == nil {
		return errors.NewRepositoryError("NO_ENCRYPTOR", "encryptor not configured", errors.ErrInvalidInput)
	}

	var setting *models.SystemSetting
	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewRepositoryError("NOT_FOUND", "setting not found", errors.ErrNotFound)
		}
		return errors.NewRepositoryError("GET_FAILED", "failed to get setting", err)
	}

	if !setting.IsEncrypted {
		return nil
	}

	decrypted, err := r.encryptor.Decrypt(setting.Value)
	if err != nil {
		return errors.NewRepositoryError("DECRYPTION_FAILED", "failed to decrypt setting", err)
	}

	setting.Value = decrypted
	setting.IsEncrypted = false

	// FIXED: Pass pointer to pointer
	if err := r.Update(ctx, setting); err != nil {
		return err
	}

	if r.cache != nil {
		r.cache.Delete(ctx, r.getCacheKey("key", key))
	}

	return nil
}

func (r *systemSettingRepository) TogglePublic(ctx context.Context, key string, isPublic bool) error {
	setting, err := r.GetByKey(ctx, key)
	if err != nil {
		return err
	}

	setting.IsPublic = isPublic

	// FIXED: Pass pointer to pointer
	if err := r.Update(ctx, setting); err != nil {
		return err
	}

	if r.cache != nil {
		r.cache.Delete(ctx, r.getCacheKey("key", key))
	}

	return nil
}

func (r *systemSettingRepository) GetSettingHistory(ctx context.Context, key string, limit int) ([]SettingHistory, error) {
	// This requires an audit log table - placeholder implementation
	return []SettingHistory{}, nil
}

func (r *systemSettingRepository) GetRecentChanges(ctx context.Context, hours int, pagination PaginationParams) ([]*models.SystemSetting, PaginationResult, error) {
	pagination.Validate()

	cutoffTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Where("updated_at >= ?", cutoffTime).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count settings", err)
	}

	var settings []*models.SystemSetting
	if err := r.db.WithContext(ctx).
		Preload("ModifiedBy").
		Where("updated_at >= ?", cutoffTime).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("updated_at DESC").
		Find(&settings).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find recent changes", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return settings, paginationResult, nil
}

func (r *systemSettingRepository) GetSettingsByModifier(ctx context.Context, userID uuid.UUID, pagination PaginationParams) ([]*models.SystemSetting, PaginationResult, error) {
	pagination.Validate()

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Where("last_modified_by = ?", userID).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count settings", err)
	}

	var settings []*models.SystemSetting
	if err := r.db.WithContext(ctx).
		Preload("ModifiedBy").
		Where("last_modified_by = ?", userID).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("updated_at DESC").
		Find(&settings).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find settings by modifier", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return settings, paginationResult, nil
}

func (r *systemSettingRepository) GetSettingStats(ctx context.Context) (SettingStats, error) {
	var stats SettingStats

	if err := r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Count(&stats.TotalSettings).Error; err != nil {
		return stats, errors.NewRepositoryError("COUNT_FAILED", "failed to count total settings", err)
	}

	r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Where("is_public = ?", true).
		Count(&stats.PublicSettings)

	r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Where("is_public = ?", false).
		Count(&stats.PrivateSettings)

	r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Where("is_encrypted = ?", true).
		Count(&stats.EncryptedSettings)

	var typeResults []struct {
		Type  models.SettingType
		Count int64
	}
	r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeResults)

	stats.ByType = make(map[models.SettingType]int64)
	for _, result := range typeResults {
		stats.ByType[result.Type] = result.Count
	}

	var categoryResults []struct {
		Category string
		Count    int64
	}
	r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Select("category, COUNT(*) as count").
		Where("category IS NOT NULL AND category != ''").
		Group("category").
		Scan(&categoryResults)

	stats.ByCategory = make(map[string]int64)
	for _, result := range categoryResults {
		stats.ByCategory[result.Category] = result.Count
	}

	cutoffTime := time.Now().Add(-24 * time.Hour)
	r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Where("updated_at >= ?", cutoffTime).
		Count(&stats.RecentlyModified)

	if len(categoryResults) > 0 {
		maxCount := int64(0)
		for _, result := range categoryResults {
			if result.Count > maxCount {
				maxCount = result.Count
				stats.MostActiveCategory = result.Category
			}
		}
	}

	return stats, nil
}

func (r *systemSettingRepository) GetCategoryStats(ctx context.Context, category string) (CategorySettingStats, error) {
	var stats CategorySettingStats
	stats.Category = category

	if err := r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Where("category = ?", category).
		Count(&stats.TotalSettings).Error; err != nil {
		return stats, errors.NewRepositoryError("COUNT_FAILED", "failed to count category settings", err)
	}

	var typeResults []struct {
		Type  models.SettingType
		Count int64
	}
	r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Select("type, COUNT(*) as count").
		Where("category = ?", category).
		Group("type").
		Scan(&typeResults)

	stats.ByType = make(map[models.SettingType]int64)
	for _, result := range typeResults {
		stats.ByType[result.Type] = result.Count
	}

	r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Where("category = ? AND is_public = ?", category, true).
		Count(&stats.PublicCount)

	r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Where("category = ? AND is_encrypted = ?", category, true).
		Count(&stats.EncryptedCount)

	var lastModified time.Time
	if err := r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Where("category = ?", category).
		Order("updated_at DESC").
		Limit(1).
		Pluck("updated_at", &lastModified).Error; err == nil && !lastModified.IsZero() {
		stats.LastModified = &lastModified
	}

	return stats, nil
}

func (r *systemSettingRepository) GetMostModifiedSettings(ctx context.Context, limit int) ([]SettingModificationCount, error) {
	// This requires an audit log table - placeholder implementation
	return []SettingModificationCount{}, nil
}

func (r *systemSettingRepository) ExportSettings(ctx context.Context, category string) (map[string]any, error) {
	var settings []*models.SystemSetting
	var err error

	if category == "" {
		err = r.db.WithContext(ctx).Find(&settings).Error
	} else {
		err = r.db.WithContext(ctx).
			Where("category = ?", category).
			Find(&settings).Error
	}

	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to export settings", "category", category, "error", err)
		}
		return nil, errors.NewRepositoryError("EXPORT_FAILED", "failed to export settings", err)
	}

	result := make(map[string]any)
	for _, setting := range settings {
		result[setting.Key] = map[string]any{
			"value":            setting.Value,
			"type":             setting.Type,
			"description":      setting.Description,
			"category":         setting.Category,
			"group":            setting.Group,
			"is_public":        setting.IsPublic,
			"is_encrypted":     setting.IsEncrypted,
			"validation_rules": setting.ValidationRules,
		}
	}

	return result, nil
}

func (r *systemSettingRepository) ImportSettings(ctx context.Context, settings map[string]any, overwrite bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, value := range settings {
			settingMap, ok := value.(map[string]any)
			if !ok {
				continue
			}

			var existing models.SystemSetting
			err := tx.Where("key = ?", key).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				setting := &models.SystemSetting{
					Key:         key,
					Value:       fmt.Sprintf("%v", settingMap["value"]),
					Type:        models.SettingType(fmt.Sprintf("%v", settingMap["type"])),
					Description: fmt.Sprintf("%v", settingMap["description"]),
					Category:    fmt.Sprintf("%v", settingMap["category"]),
					Group:       fmt.Sprintf("%v", settingMap["group"]),
				}

				if isPublic, ok := settingMap["is_public"].(bool); ok {
					setting.IsPublic = isPublic
				}
				if isEncrypted, ok := settingMap["is_encrypted"].(bool); ok {
					setting.IsEncrypted = isEncrypted
				}

				if err := tx.Create(setting).Error; err != nil {
					return errors.NewRepositoryError("IMPORT_FAILED", "failed to import setting", err)
				}
			} else if err != nil {
				return err
			} else if overwrite {
				existing.Value = fmt.Sprintf("%v", settingMap["value"])
				existing.Type = models.SettingType(fmt.Sprintf("%v", settingMap["type"]))
				existing.Description = fmt.Sprintf("%v", settingMap["description"])
				existing.Category = fmt.Sprintf("%v", settingMap["category"])
				existing.Group = fmt.Sprintf("%v", settingMap["group"])

				if isPublic, ok := settingMap["is_public"].(bool); ok {
					existing.IsPublic = isPublic
				}
				if isEncrypted, ok := settingMap["is_encrypted"].(bool); ok {
					existing.IsEncrypted = isEncrypted
				}

				if err := tx.Save(&existing).Error; err != nil {
					return errors.NewRepositoryError("IMPORT_FAILED", "failed to update setting", err)
				}
			}
		}
		return nil
	})
}

func (r *systemSettingRepository) ResetToDefaults(ctx context.Context, category string) error {
	if r.logger != nil {
		r.logger.Info("reset to defaults called", "category", category)
	}
	return errors.NewRepositoryError("NOT_IMPLEMENTED", "reset to defaults not implemented", errors.ErrInvalidInput)
}

func (r *systemSettingRepository) BackupSettings(ctx context.Context) ([]byte, error) {
	settings, err := r.ExportSettings(ctx, "")
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to marshal settings backup", "error", err)
		}
		return nil, errors.NewRepositoryError("BACKUP_FAILED", "failed to marshal backup", err)
	}

	if r.logger != nil {
		r.logger.Info("settings backup created", "size", len(data))
	}
	return data, nil
}

func (r *systemSettingRepository) RestoreSettings(ctx context.Context, backup []byte) error {
	var settings map[string]any
	if err := json.Unmarshal(backup, &settings); err != nil {
		if r.logger != nil {
			r.logger.Error("failed to unmarshal settings backup", "error", err)
		}
		return errors.NewRepositoryError("RESTORE_FAILED", "failed to unmarshal backup", err)
	}

	if err := r.ImportSettings(ctx, settings, true); err != nil {
		return err
	}

	if r.cache != nil {
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("*"))
	}

	if r.logger != nil {
		r.logger.Info("settings restored from backup", "count", len(settings))
	}
	return nil
}

func (r *systemSettingRepository) RefreshCache(ctx context.Context, key string) error {
	if r.cache == nil {
		return nil
	}

	cacheKey := r.getCacheKey("key", key)
	if err := r.cache.Delete(ctx, cacheKey); err != nil && r.logger != nil {
		r.logger.Warn("failed to delete cache key", "key", cacheKey, "error", err)
	}

	_, err := r.GetByKey(ctx, key)
	return err
}

func (r *systemSettingRepository) RefreshCategoryCache(ctx context.Context, category string) error {
	if r.cache == nil {
		return nil
	}

	cacheKey := r.getCacheKey("category", category)
	if err := r.cache.Delete(ctx, cacheKey); err != nil && r.logger != nil {
		r.logger.Warn("failed to delete category cache", "category", category, "error", err)
	}

	_, err := r.GetByCategory(ctx, category)
	return err
}

func (r *systemSettingRepository) ClearSettingsCache(ctx context.Context) error {
	if r.cache == nil {
		return nil
	}

	pattern := r.getCacheKeyPattern("*")
	if err := r.cache.DeletePattern(ctx, pattern); err != nil {
		if r.logger != nil {
			r.logger.Error("failed to clear settings cache", "error", err)
		}
		return errors.NewRepositoryError("CACHE_CLEAR_FAILED", "failed to clear cache", err)
	}

	if r.logger != nil {
		r.logger.Info("settings cache cleared")
	}
	return nil
}

// Helper methods
func (r *systemSettingRepository) getCacheKey(prefix string, parts ...string) string {
	allParts := append([]string{"repo", "system_settings", prefix}, parts...)
	return strings.Join(allParts, ":")
}

func (r *systemSettingRepository) getCacheKeyPattern(pattern string) string {
	return fmt.Sprintf("repo:system_settings:%s", pattern)
}

func (r *systemSettingRepository) SetSetting(ctx context.Context, key, value string, settingType models.SettingType) error {
	if key == "" {
		return errors.NewRepositoryError("INVALID_INPUT", "key cannot be empty", errors.ErrInvalidInput)
	}

	var existing *models.SystemSetting
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		setting := &models.SystemSetting{
			Key:   key,
			Value: value,
			Type:  settingType,
		}
		if err := r.Create(ctx, setting); err != nil {
			return err
		}
	} else if err != nil {
		return errors.NewRepositoryError("GET_FAILED", "failed to check existing setting", err)
	} else {
		existing.Value = value
		existing.Type = settingType
		if err := r.Update(ctx, existing); err != nil {
			return err
		}
	}

	if r.cache != nil {
		r.cache.Delete(ctx, r.getCacheKey("key", key))
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("category:*"))
	}

	return nil
}

func (r *systemSettingRepository) DeleteByKey(ctx context.Context, key string) error {
	if key == "" {
		return errors.NewRepositoryError("INVALID_INPUT", "key cannot be empty", errors.ErrInvalidInput)
	}

	result := r.db.WithContext(ctx).Where("key = ?", key).Delete(&models.SystemSetting{})
	if result.Error != nil {
		if r.logger != nil {
			r.logger.Error("failed to delete setting", "key", key, "error", result.Error)
		}
		return errors.NewRepositoryError("DELETE_FAILED", "failed to delete setting", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewRepositoryError("NOT_FOUND", "setting not found", errors.ErrNotFound)
	}

	if r.cache != nil {
		r.cache.Delete(ctx, r.getCacheKey("key", key))
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("*"))
	}

	if r.logger != nil {
		r.logger.Info("setting deleted", "key", key)
	}
	return nil
}

func (r *systemSettingRepository) KeyExists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, nil
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Where("key = ?", key).
		Count(&count).Error; err != nil {
		return false, errors.NewRepositoryError("COUNT_FAILED", "failed to check key existence", err)
	}

	return count > 0, nil
}

func (r *systemSettingRepository) GetStringValue(ctx context.Context, key string, defaultValue string) (string, error) {
	setting, err := r.GetByKey(ctx, key)
	if err != nil {
		if errors.IsNotFound(err) {
			return defaultValue, nil
		}
		return "", err
	}
	return setting.Value, nil
}

func (r *systemSettingRepository) GetBoolValue(ctx context.Context, key string, defaultValue bool) (bool, error) {
	setting, err := r.GetByKey(ctx, key)
	if err != nil {
		if errors.IsNotFound(err) {
			return defaultValue, nil
		}
		return false, err
	}
	return setting.GetBoolValue(), nil
}

func (r *systemSettingRepository) GetIntValue(ctx context.Context, key string, defaultValue int) (int, error) {
	setting, err := r.GetByKey(ctx, key)
	if err != nil {
		if errors.IsNotFound(err) {
			return defaultValue, nil
		}
		return 0, err
	}
	return setting.GetIntValue(), nil
}

func (r *systemSettingRepository) GetFloatValue(ctx context.Context, key string, defaultValue float64) (float64, error) {
	setting, err := r.GetByKey(ctx, key)
	if err != nil {
		if errors.IsNotFound(err) {
			return defaultValue, nil
		}
		return 0, err
	}
	return setting.GetFloatValue(), nil
}

func (r *systemSettingRepository) GetJSONValue(ctx context.Context, key string, dest any) error {
	setting, err := r.GetByKey(ctx, key)
	if err != nil {
		return err
	}

	if setting.Type != models.SettingTypeJSON {
		return errors.NewRepositoryError("INVALID_TYPE", "setting is not of type JSON", errors.ErrInvalidInput)
	}

	if err := json.Unmarshal([]byte(setting.Value), dest); err != nil {
		if r.logger != nil {
			r.logger.Error("failed to unmarshal JSON setting", "key", key, "error", err)
		}
		return errors.NewRepositoryError("UNMARSHAL_FAILED", "failed to unmarshal JSON", err)
	}

	return nil
}

func (r *systemSettingRepository) GetByCategory(ctx context.Context, category string) ([]*models.SystemSetting, error) {
	if category == "" {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "category cannot be empty", errors.ErrInvalidInput)
	}

	cacheKey := r.getCacheKey("category", category)
	if r.cache != nil {
		var settings []*models.SystemSetting
		if err := r.cache.GetJSON(ctx, cacheKey, &settings); err == nil {
			if r.metrics != nil {
				r.metrics.RecordCacheHit("system_settings")
			}
			return settings, nil
		}
		if r.metrics != nil {
			r.metrics.RecordCacheMiss("system_settings")
		}
	}

	var settings []*models.SystemSetting
	if err := r.db.WithContext(ctx).
		Preload("ModifiedBy").
		Where("category = ?", category).
		Order("key ASC").
		Find(&settings).Error; err != nil {
		if r.logger != nil {
			r.logger.Error("failed to get settings by category", "category", category, "error", err)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get settings", err)
	}

	for _, setting := range settings {
		if setting.IsEncrypted && r.encryptor != nil {
			decrypted, err := r.encryptor.Decrypt(setting.Value)
			if err != nil {
				if r.logger != nil {
					r.logger.Error("failed to decrypt setting", "key", setting.Key, "error", err)
				}
				continue
			}
			setting.Value = decrypted
		}
	}

	if r.cache != nil {
		if err := r.cache.SetJSON(ctx, cacheKey, settings, 15*time.Minute); err != nil && r.logger != nil {
			r.logger.Warn("failed to cache category settings", "category", category, "error", err)
		}
	}

	return settings, nil
}

func (r *systemSettingRepository) GetByGroup(ctx context.Context, group string) ([]*models.SystemSetting, error) {
	if group == "" {
		return nil, errors.NewRepositoryError("INVALID_INPUT", "group cannot be empty", errors.ErrInvalidInput)
	}

	var settings []*models.SystemSetting
	if err := r.db.WithContext(ctx).
		Preload("ModifiedBy").
		Where("\"group\" = ?", group).
		Order("key ASC").
		Find(&settings).Error; err != nil {
		if r.logger != nil {
			r.logger.Error("failed to get settings by group", "group", group, "error", err)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get settings", err)
	}

	for _, setting := range settings {
		if setting.IsEncrypted && r.encryptor != nil {
			decrypted, err := r.encryptor.Decrypt(setting.Value)
			if err != nil {
				if r.logger != nil {
					r.logger.Error("failed to decrypt setting", "key", setting.Key, "error", err)
				}
				continue
			}
			setting.Value = decrypted
		}
	}

	return settings, nil
}

func (r *systemSettingRepository) GetPublicSettings(ctx context.Context) ([]*models.SystemSetting, error) {
	var settings []*models.SystemSetting
	if err := r.db.WithContext(ctx).
		Where("is_public = ?", true).
		Order("category ASC, key ASC").
		Find(&settings).Error; err != nil {
		if r.logger != nil {
			r.logger.Error("failed to get public settings", "error", err)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get public settings", err)
	}

	return settings, nil
}

func (r *systemSettingRepository) GetPrivateSettings(ctx context.Context) ([]*models.SystemSetting, error) {
	var settings []*models.SystemSetting
	if err := r.db.WithContext(ctx).
		Preload("ModifiedBy").
		Where("is_public = ?", false).
		Order("category ASC, key ASC").
		Find(&settings).Error; err != nil {
		if r.logger != nil {
			r.logger.Error("failed to get private settings", "error", err)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get private settings", err)
	}

	for _, setting := range settings {
		if setting.IsEncrypted && r.encryptor != nil {
			decrypted, err := r.encryptor.Decrypt(setting.Value)
			if err != nil {
				if r.logger != nil {
					r.logger.Error("failed to decrypt setting", "key", setting.Key, "error", err)
				}
				continue
			}
			setting.Value = decrypted
		}
	}

	return settings, nil
}

func (r *systemSettingRepository) GetEncryptedSettings(ctx context.Context) ([]*models.SystemSetting, error) {
	var settings []*models.SystemSetting
	if err := r.db.WithContext(ctx).
		Preload("ModifiedBy").
		Where("is_encrypted = ?", true).
		Order("category ASC, key ASC").
		Find(&settings).Error; err != nil {
		if r.logger != nil {
			r.logger.Error("failed to get encrypted settings", "error", err)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get encrypted settings", err)
	}

	return settings, nil
}

func (r *systemSettingRepository) BulkSet(ctx context.Context, settings map[string]SettingValue) error {
	if len(settings) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, val := range settings {
			var existing models.SystemSetting
			err := tx.Where("key = ?", key).First(&existing).Error

			setting := &models.SystemSetting{
				Key:         key,
				Value:       val.Value,
				Type:        val.Type,
				Description: val.Description,
				Category:    val.Category,
				Group:       val.Group,
				IsPublic:    val.IsPublic,
				IsEncrypted: val.IsEncrypted,
			}

			// FIXED: Direct assignment since both are map[string]any
			if val.Rules != nil {
				setting.ValidationRules = models.JSONB(val.Rules)
			}

			if val.IsEncrypted && r.encryptor != nil {
				encrypted, err := r.encryptor.Encrypt(val.Value)
				if err != nil {
					return errors.NewRepositoryError("ENCRYPTION_FAILED", "failed to encrypt setting", err)
				}
				setting.Value = encrypted
			}

			if err == gorm.ErrRecordNotFound {
				if err := tx.Create(setting).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				setting.ID = existing.ID
				if err := tx.Save(setting).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *systemSettingRepository) BulkDeleteByKeys(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).Where("key IN ?", keys).Delete(&models.SystemSetting{})
	if result.Error != nil {
		if r.logger != nil {
			r.logger.Error("failed to bulk delete settings", "count", len(keys), "error", result.Error)
		}
		return errors.NewRepositoryError("DELETE_FAILED", "failed to bulk delete settings", result.Error)
	}

	if r.cache != nil {
		r.cache.DeletePattern(ctx, r.getCacheKeyPattern("*"))
	}

	if r.logger != nil {
		r.logger.Info("bulk deleted settings", "count", result.RowsAffected)
	}
	return nil
}

func (r *systemSettingRepository) Search(ctx context.Context, query string, pagination PaginationParams) ([]*models.SystemSetting, PaginationResult, error) {
	pagination.Validate()
	searchPattern := "%" + query + "%"

	var totalItems int64
	if err := r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Where("key ILIKE ? OR description ILIKE ?", searchPattern, searchPattern).
		Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count settings", err)
	}

	var settings []*models.SystemSetting
	if err := r.db.WithContext(ctx).
		Preload("ModifiedBy").
		Where("key ILIKE ? OR description ILIKE ?", searchPattern, searchPattern).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("key ASC").
		Find(&settings).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to search settings", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return settings, paginationResult, nil
}

func (r *systemSettingRepository) FindByFilters(ctx context.Context, filters SettingFilters, pagination PaginationParams) ([]*models.SystemSetting, PaginationResult, error) {
	pagination.Validate()

	query := r.db.WithContext(ctx).Model(&models.SystemSetting{})

	if len(filters.Categories) > 0 {
		query = query.Where("category IN ?", filters.Categories)
	}

	if len(filters.Groups) > 0 {
		query = query.Where("\"group\" IN ?", filters.Groups)
	}

	if len(filters.Types) > 0 {
		query = query.Where("type IN ?", filters.Types)
	}

	if filters.IsPublic != nil {
		query = query.Where("is_public = ?", *filters.IsPublic)
	}

	if filters.IsEncrypted != nil {
		query = query.Where("is_encrypted = ?", *filters.IsEncrypted)
	}

	if filters.ModifiedBy != nil {
		query = query.Where("last_modified_by = ?", *filters.ModifiedBy)
	}

	if filters.ModifiedAfter != nil {
		query = query.Where("updated_at >= ?", *filters.ModifiedAfter)
	}

	if filters.ModifiedBefore != nil {
		query = query.Where("updated_at <= ?", *filters.ModifiedBefore)
	}

	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("COUNT_FAILED", "failed to count settings", err)
	}

	var settings []*models.SystemSetting
	if err := query.
		Preload("ModifiedBy").
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Order("category ASC, key ASC").
		Find(&settings).Error; err != nil {
		return nil, PaginationResult{}, errors.NewRepositoryError("FIND_FAILED", "failed to find settings", err)
	}

	paginationResult := CalculatePagination(pagination, totalItems)
	return settings, paginationResult, nil
}

func (r *systemSettingRepository) GetAllCategories(ctx context.Context) ([]string, error) {
	var categories []string
	if err := r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Distinct("category").
		Where("category IS NOT NULL AND category != ''").
		Order("category ASC").
		Pluck("category", &categories).Error; err != nil {
		if r.logger != nil {
			r.logger.Error("failed to get categories", "error", err)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get categories", err)
	}
	return categories, nil
}

func (r *systemSettingRepository) GetAllGroups(ctx context.Context) ([]string, error) {
	var groups []string
	if err := r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Distinct("\"group\"").
		Where("\"group\" IS NOT NULL AND \"group\" != ''").
		Order("\"group\" ASC").
		Pluck("\"group\"", &groups).Error; err != nil {
		if r.logger != nil {
			r.logger.Error("failed to get groups", "error", err)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get groups", err)
	}
	return groups, nil
}

func (r *systemSettingRepository) GetCategoriesWithCount(ctx context.Context) ([]CategoryCount, error) {
	var results []CategoryCount
	if err := r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Select("category, COUNT(*) as count").
		Where("category IS NOT NULL AND category != ''").
		Group("category").
		Order("count DESC").
		Scan(&results).Error; err != nil {
		if r.logger != nil {
			r.logger.Error("failed to get categories with count", "error", err)
		}
		return nil, errors.NewRepositoryError("GET_FAILED", "failed to get categories with count", err)
	}
	return results, nil
}
