package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// System Settings Request DTOs
// ============================================================================

// CreateSettingRequest represents a request to create a system setting
type CreateSettingRequest struct {
	Key             string             `json:"key" validate:"required"`
	Value           string             `json:"value" validate:"required"`
	Type            models.SettingType `json:"type" validate:"required,oneof=string number boolean json"`
	Description     string             `json:"description,omitempty"`
	Category        string             `json:"category" validate:"required"`
	Group           string             `json:"group,omitempty"`
	IsPublic        bool               `json:"is_public"`
	IsEncrypted     bool               `json:"is_encrypted"`
	ValidationRules map[string]any     `json:"validation_rules,omitempty"`
}

// UpdateSettingRequest represents a request to update a system setting
type UpdateSettingRequest struct {
	Value           *string        `json:"value,omitempty"`
	Description     *string        `json:"description,omitempty"`
	Category        *string        `json:"category,omitempty"`
	Group           *string        `json:"group,omitempty"`
	IsPublic        *bool          `json:"is_public,omitempty"`
	ValidationRules map[string]any `json:"validation_rules,omitempty"`
}

// BulkSetSettingsRequest represents a request to set multiple settings
type BulkSetSettingsRequest struct {
	Settings map[string]SettingValueRequest `json:"settings" validate:"required"`
}

// SettingValueRequest represents a setting value for bulk operations
type SettingValueRequest struct {
	Value           string             `json:"value" validate:"required"`
	Type            models.SettingType `json:"type" validate:"required"`
	Description     string             `json:"description,omitempty"`
	Category        string             `json:"category,omitempty"`
	Group           string             `json:"group,omitempty"`
	IsPublic        bool               `json:"is_public"`
	IsEncrypted     bool               `json:"is_encrypted"`
	ValidationRules map[string]any     `json:"validation_rules,omitempty"`
}

// SettingFilter represents filters for setting queries
type SettingFilter struct {
	Categories     []string             `json:"categories,omitempty"`
	Groups         []string             `json:"groups,omitempty"`
	Types          []models.SettingType `json:"types,omitempty"`
	IsPublic       *bool                `json:"is_public,omitempty"`
	IsEncrypted    *bool                `json:"is_encrypted,omitempty"`
	ModifiedBy     *uuid.UUID           `json:"modified_by,omitempty"`
	ModifiedAfter  *time.Time           `json:"modified_after,omitempty"`
	ModifiedBefore *time.Time           `json:"modified_before,omitempty"`
	SearchQuery    string               `json:"search_query,omitempty"`
	Page           int                  `json:"page"`
	PageSize       int                  `json:"page_size"`
}

// ============================================================================
// System Settings Response DTOs
// ============================================================================

// SystemSettingResponse represents a system setting
type SystemSettingResponse struct {
	ID              uuid.UUID          `json:"id"`
	Key             string             `json:"key"`
	Value           string             `json:"value"`
	Type            models.SettingType `json:"type"`
	Description     string             `json:"description,omitempty"`
	Category        string             `json:"category"`
	Group           string             `json:"group,omitempty"`
	ValidationRules models.JSONB       `json:"validation_rules,omitempty"`
	IsPublic        bool               `json:"is_public"`
	IsEncrypted     bool               `json:"is_encrypted"`
	LastModifiedBy  *uuid.UUID         `json:"last_modified_by,omitempty"`
	ModifiedBy      *UserSummary       `json:"modified_by,omitempty"`
	DisplayValue    string             `json:"display_value"`
	FullPath        string             `json:"full_path"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// SystemSettingListResponse represents a paginated list of system settings
type SystemSettingListResponse struct {
	Settings    []*SystemSettingResponse `json:"settings"`
	Page        int                      `json:"page"`
	PageSize    int                      `json:"page_size"`
	TotalItems  int64                    `json:"total_items"`
	TotalPages  int                      `json:"total_pages"`
	HasNext     bool                     `json:"has_next"`
	HasPrevious bool                     `json:"has_previous"`
}

// SettingStatsResponse represents system settings statistics
type SettingStatsResponse struct {
	TotalSettings     int64                        `json:"total_settings"`
	PublicSettings    int64                        `json:"public_settings"`
	PrivateSettings   int64                        `json:"private_settings"`
	EncryptedSettings int64                        `json:"encrypted_settings"`
	ByType            map[models.SettingType]int64 `json:"by_type"`
	ByCategory        map[string]int64             `json:"by_category"`
	ByGroup           map[string]int64             `json:"by_group"`
	RecentChanges     int64                        `json:"recent_changes_24h"`
	TotalCategories   int                          `json:"total_categories"`
	TotalGroups       int                          `json:"total_groups"`
}

// CategoryStatsResponse represents category statistics
type CategoryStatsResponse struct {
	Category      string                       `json:"category"`
	TotalSettings int64                        `json:"total_settings"`
	ByType        map[models.SettingType]int64 `json:"by_type"`
	ByGroup       map[string]int64             `json:"by_group"`
	PublicCount   int64                        `json:"public_count"`
	PrivateCount  int64                        `json:"private_count"`
}

// CategoryWithCountResponse represents a category with count
type CategoryWithCountResponse struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToSystemSettingResponse converts a SystemSetting model to SystemSettingResponse DTO
func ToSystemSettingResponse(setting *models.SystemSetting) *SystemSettingResponse {
	if setting == nil {
		return nil
	}

	resp := &SystemSettingResponse{
		ID:              setting.ID,
		Key:             setting.Key,
		Value:           setting.Value,
		Type:            setting.Type,
		Description:     setting.Description,
		Category:        setting.Category,
		Group:           setting.Group,
		ValidationRules: setting.ValidationRules,
		IsPublic:        setting.IsPublic,
		IsEncrypted:     setting.IsEncrypted,
		LastModifiedBy:  setting.LastModifiedBy,
		DisplayValue:    setting.GetDisplayValue(),
		FullPath:        setting.GetFullPath(),
		CreatedAt:       setting.CreatedAt,
		UpdatedAt:       setting.UpdatedAt,
	}

	// Add modifier if available
	if setting.ModifiedBy != nil {
		resp.ModifiedBy = &UserSummary{
			ID:        setting.ModifiedBy.ID,
			FirstName: setting.ModifiedBy.FirstName,
			LastName:  setting.ModifiedBy.LastName,
			Email:     setting.ModifiedBy.Email,
			AvatarURL: setting.ModifiedBy.AvatarURL,
		}
	}

	return resp
}

// ToSystemSettingResponses converts multiple SystemSetting models to DTOs
func ToSystemSettingResponses(settings []*models.SystemSetting) []*SystemSettingResponse {
	if settings == nil {
		return nil
	}

	responses := make([]*SystemSettingResponse, len(settings))
	for i, setting := range settings {
		responses[i] = ToSystemSettingResponse(setting)
	}
	return responses
}
