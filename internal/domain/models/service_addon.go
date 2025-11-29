package models

import "github.com/google/uuid"

type ServiceAddon struct {
	BaseModel
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`

	Name            string  `json:"name" gorm:"not null;size:255" validate:"required"`
	Description     string  `json:"description,omitempty" gorm:"type:text"`
	Price           float64 `json:"price" gorm:"type:decimal(10,2);not null" validate:"required,min=0"`
	DurationMinutes int     `json:"duration_minutes" gorm:"default:0"`
	IsActive        bool    `json:"is_active" gorm:"default:true"`

	// Relationships
	Services []Service `json:"services,omitempty" gorm:"many2many:service_addon_relations"`
}
