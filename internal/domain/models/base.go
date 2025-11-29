package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel provides common fields and behavior for all entities
type BaseModel struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime;index"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	Version   int        `json:"version" gorm:"default:0;not null"` // Optimistic locking
}

// BeforeCreate hook to ensure UUID is generated and version is initialized
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	if b.Version == 0 {
		b.Version = 1
	}
	return nil
}

// BeforeUpdate hook to increment version for optimistic locking
func (b *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	b.Version++
	return nil
}

// IsDeleted checks if the entity is soft deleted
func (b *BaseModel) IsDeleted() bool {
	return b.DeletedAt != nil && !b.DeletedAt.IsZero()
}

// GetAge returns the age of the entity in seconds
func (b *BaseModel) GetAge() time.Duration {
	return time.Since(b.CreatedAt)
}

// JSONB is a type alias for JSONB database columns
type JSONB map[string]any

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}
