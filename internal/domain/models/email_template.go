package models

import "strings"

type EmailTemplate struct {
	BaseModel

	// Template Details
	Name        string `json:"name" gorm:"uniqueIndex;not null;size:255" validate:"required"`
	Subject     string `json:"subject" gorm:"not null;size:255" validate:"required"`
	Description string `json:"description,omitempty" gorm:"type:text"`

	// Content
	HTMLBody string `json:"html_body" gorm:"type:text;not null" validate:"required"`
	TextBody string `json:"text_body,omitempty" gorm:"type:text"`

	// Variables
	Variables []string `json:"variables,omitempty" gorm:"type:text[]"` // {{customer_name}}, {{booking_date}}

	// Category
	Category string `json:"category" gorm:"size:100;index"` // booking, payment, reminder

	// Status
	IsActive bool `json:"is_active" gorm:"default:true"`

	// Multi-language Support
	Language string `json:"language" gorm:"size:10;default:'en'"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`
}

// Business Methods
func (et *EmailTemplate) RenderSubject(variables map[string]string) string {
	result := et.Subject
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		result = strings.Replace(result, placeholder, value, -1)
	}
	return result
}

func (et *EmailTemplate) RenderBody(variables map[string]string) string {
	result := et.HTMLBody
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		result = strings.Replace(result, placeholder, value, -1)
	}
	return result
}
