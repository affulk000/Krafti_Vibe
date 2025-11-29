package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type SettingType string

const (
	SettingTypeString  SettingType = "string"
	SettingTypeNumber  SettingType = "number"
	SettingTypeBoolean SettingType = "boolean"
	SettingTypeJSON    SettingType = "json"
)

type SystemSetting struct {
	BaseModel

	// Setting Details
	Key         string      `json:"key" gorm:"uniqueIndex;not null;size:255" validate:"required"`
	Value       string      `json:"value" gorm:"type:text;not null" validate:"required"`
	Type        SettingType `json:"type" gorm:"type:varchar(50);not null" validate:"required"`
	Description string      `json:"description,omitempty" gorm:"type:text"`

	// Categorization
	Category string `json:"category" gorm:"size:100;index"`
	Group    string `json:"group,omitempty" gorm:"size:100"`

	// Validation
	ValidationRules JSONB `json:"validation_rules,omitempty" gorm:"type:jsonb"`

	// Access Control
	IsPublic    bool `json:"is_public" gorm:"default:false"`
	IsEncrypted bool `json:"is_encrypted" gorm:"default:false"`

	// Modified By
	LastModifiedBy *uuid.UUID `json:"last_modified_by,omitempty" gorm:"type:uuid"`

	// Relationships
	ModifiedBy *User `json:"modified_by,omitempty" gorm:"foreignKey:LastModifiedBy"`
}

// TableName specifies the table name for SystemSetting
func (SystemSetting) TableName() string {
	return "system_settings"
}

// Business Methods

// GetBoolValue returns the setting value as a boolean
func (ss *SystemSetting) GetBoolValue() bool {
	val := strings.ToLower(strings.TrimSpace(ss.Value))
	return val == "true" || val == "1" || val == "yes" || val == "on"
}

// GetBoolValueWithError returns the setting value as a boolean with error handling
func (ss *SystemSetting) GetBoolValueWithError() (bool, error) {
	if ss.Type != SettingTypeBoolean {
		return false, fmt.Errorf("setting type is %s, not boolean", ss.Type)
	}
	return strconv.ParseBool(ss.Value)
}

// GetIntValue returns the setting value as an integer
func (ss *SystemSetting) GetIntValue() int {
	val, _ := strconv.Atoi(ss.Value)
	return val
}

// GetIntValueWithError returns the setting value as an integer with error handling
func (ss *SystemSetting) GetIntValueWithError() (int, error) {
	if ss.Type != SettingTypeNumber {
		return 0, fmt.Errorf("setting type is %s, not number", ss.Type)
	}
	return strconv.Atoi(ss.Value)
}

// GetInt64Value returns the setting value as int64
func (ss *SystemSetting) GetInt64Value() int64 {
	val, _ := strconv.ParseInt(ss.Value, 10, 64)
	return val
}

// GetInt64ValueWithError returns the setting value as int64 with error handling
func (ss *SystemSetting) GetInt64ValueWithError() (int64, error) {
	if ss.Type != SettingTypeNumber {
		return 0, fmt.Errorf("setting type is %s, not number", ss.Type)
	}
	return strconv.ParseInt(ss.Value, 10, 64)
}

// GetFloatValue returns the setting value as a float64
func (ss *SystemSetting) GetFloatValue() float64 {
	val, _ := strconv.ParseFloat(ss.Value, 64)
	return val
}

// GetFloatValueWithError returns the setting value as a float64 with error handling
func (ss *SystemSetting) GetFloatValueWithError() (float64, error) {
	if ss.Type != SettingTypeNumber {
		return 0, fmt.Errorf("setting type is %s, not number", ss.Type)
	}
	return strconv.ParseFloat(ss.Value, 64)
}

// GetStringValue returns the setting value as a string
func (ss *SystemSetting) GetStringValue() string {
	return ss.Value
}

// GetJSONValue unmarshals the JSON value into the provided destination
func (ss *SystemSetting) GetJSONValue(dest interface{}) error {
	if ss.Type != SettingTypeJSON {
		return fmt.Errorf("setting type is %s, not json", ss.Type)
	}
	return json.Unmarshal([]byte(ss.Value), dest)
}

// SetValue sets the value and automatically determines the type
func (ss *SystemSetting) SetValue(value interface{}) error {
	switch v := value.(type) {
	case string:
		ss.Value = v
		ss.Type = SettingTypeString
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		ss.Value = fmt.Sprintf("%v", v)
		ss.Type = SettingTypeNumber
	case float32, float64:
		ss.Value = fmt.Sprintf("%v", v)
		ss.Type = SettingTypeNumber
	case bool:
		ss.Value = strconv.FormatBool(v)
		ss.Type = SettingTypeBoolean
	default:
		// Assume JSON for complex types
		jsonData, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value to JSON: %w", err)
		}
		ss.Value = string(jsonData)
		ss.Type = SettingTypeJSON
	}
	return nil
}

// Validate performs basic validation on the setting
func (ss *SystemSetting) Validate() error {
	if ss.Key == "" {
		return fmt.Errorf("setting key cannot be empty")
	}

	if ss.Value == "" {
		return fmt.Errorf("setting value cannot be empty")
	}

	// Validate type-specific values
	switch ss.Type {
	case SettingTypeBoolean:
		if _, err := strconv.ParseBool(ss.Value); err != nil {
			return fmt.Errorf("invalid boolean value: %s", ss.Value)
		}
	case SettingTypeNumber:
		if _, err := strconv.ParseFloat(ss.Value, 64); err != nil {
			return fmt.Errorf("invalid number value: %s", ss.Value)
		}
	case SettingTypeJSON:
		var js json.RawMessage
		if err := json.Unmarshal([]byte(ss.Value), &js); err != nil {
			return fmt.Errorf("invalid JSON value: %w", err)
		}
	case SettingTypeString:
		// String is always valid
	default:
		return fmt.Errorf("unknown setting type: %s", ss.Type)
	}

	return nil
}

// IsValidType checks if the setting type is valid
func (ss *SystemSetting) IsValidType() bool {
	switch ss.Type {
	case SettingTypeString, SettingTypeNumber, SettingTypeBoolean, SettingTypeJSON:
		return true
	default:
		return false
	}
}

// GetDisplayValue returns a human-readable display value
func (ss *SystemSetting) GetDisplayValue() string {
	if ss.IsEncrypted && ss.Value != "" {
		return "***encrypted***"
	}

	switch ss.Type {
	case SettingTypeJSON:
		// Pretty print JSON for display
		var js json.RawMessage
		if err := json.Unmarshal([]byte(ss.Value), &js); err == nil {
			if pretty, err := json.MarshalIndent(js, "", "  "); err == nil {
				return string(pretty)
			}
		}
	}

	return ss.Value
}

// Clone creates a deep copy of the setting
func (ss *SystemSetting) Clone() *SystemSetting {
	clone := *ss

	// Clone validation rules
	if ss.ValidationRules != nil {
		clone.ValidationRules = make(JSONB)
		for k, v := range ss.ValidationRules {
			clone.ValidationRules[k] = v
		}
	}

	// Clone UUID pointers
	if ss.LastModifiedBy != nil {
		id := *ss.LastModifiedBy
		clone.LastModifiedBy = &id
	}

	return &clone
}

// GetFullPath returns the full hierarchical path of the setting
func (ss *SystemSetting) GetFullPath() string {
	parts := []string{}

	if ss.Category != "" {
		parts = append(parts, ss.Category)
	}

	if ss.Group != "" {
		parts = append(parts, ss.Group)
	}

	parts = append(parts, ss.Key)

	return strings.Join(parts, ".")
}

// HasValidationRule checks if a specific validation rule exists
func (ss *SystemSetting) HasValidationRule(ruleName string) bool {
	if ss.ValidationRules == nil {
		return false
	}
	_, exists := ss.ValidationRules[ruleName]
	return exists
}

// GetValidationRule retrieves a specific validation rule
func (ss *SystemSetting) GetValidationRule(ruleName string) (interface{}, bool) {
	if ss.ValidationRules == nil {
		return nil, false
	}
	val, exists := ss.ValidationRules[ruleName]
	return val, exists
}

// SetValidationRule sets a specific validation rule
func (ss *SystemSetting) SetValidationRule(ruleName string, value interface{}) {
	if ss.ValidationRules == nil {
		ss.ValidationRules = make(JSONB)
	}
	ss.ValidationRules[ruleName] = value
}

// RemoveValidationRule removes a specific validation rule
func (ss *SystemSetting) RemoveValidationRule(ruleName string) {
	if ss.ValidationRules != nil {
		delete(ss.ValidationRules, ruleName)
	}
}

// IsModifiedBy checks if the setting was last modified by a specific user
func (ss *SystemSetting) IsModifiedBy(userID uuid.UUID) bool {
	return ss.LastModifiedBy != nil && *ss.LastModifiedBy == userID
}

// String returns a string representation of the setting
func (ss *SystemSetting) String() string {
	return fmt.Sprintf("Setting{Key: %s, Type: %s, Category: %s, Public: %t}",
		ss.Key, ss.Type, ss.Category, ss.IsPublic)
}
