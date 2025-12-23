package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// ProjectUpdate Request DTOs
// ============================================================================

// CreateProjectUpdateRequest represents a request to create a project update
type CreateProjectUpdateRequest struct {
	ProjectID         uuid.UUID              `json:"project_id" validate:"required"`
	UserID            uuid.UUID              `json:"user_id" validate:"required"`
	Type              models.UpdateType      `json:"type" validate:"required"`
	Title             string                 `json:"title" validate:"required,max=255"`
	Description       string                 `json:"description,omitempty"`
	VisibleToCustomer bool                   `json:"visible_to_customer"`
	AttachmentURLs    []string               `json:"attachment_urls,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateProjectUpdateRequest represents a request to update a project update
type UpdateProjectUpdateRequest struct {
	Title             *string                `json:"title,omitempty" validate:"omitempty,max=255"`
	Description       *string                `json:"description,omitempty"`
	VisibleToCustomer *bool                  `json:"visible_to_customer,omitempty"`
	AttachmentURLs    []string               `json:"attachment_urls,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ProjectUpdateFilter represents filters for project update queries
type ProjectUpdateFilter struct {
	ProjectID         uuid.UUID          `json:"project_id"`
	Type              *models.UpdateType `json:"type,omitempty"`
	VisibleToCustomer *bool              `json:"visible_to_customer,omitempty"`
	Page              int                `json:"page"`
	PageSize          int                `json:"page_size"`
}

// ============================================================================
// ProjectUpdate Response DTOs
// ============================================================================

// ProjectUpdateResponse represents a project update
type ProjectUpdateResponse struct {
	ID                uuid.UUID         `json:"id"`
	TenantID          uuid.UUID         `json:"tenant_id"`
	ProjectID         uuid.UUID         `json:"project_id"`
	UserID            uuid.UUID         `json:"user_id"`
	Type              models.UpdateType `json:"type"`
	Title             string            `json:"title"`
	Description       string            `json:"description,omitempty"`
	VisibleToCustomer bool              `json:"visible_to_customer"`
	AttachmentURLs    []string          `json:"attachment_urls,omitempty"`
	Metadata          models.JSONB      `json:"metadata,omitempty"`
	User              *UserSummary      `json:"user,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// ProjectUpdateListResponse represents a paginated list of project updates
type ProjectUpdateListResponse struct {
	Updates     []*ProjectUpdateResponse `json:"updates"`
	Page        int                      `json:"page"`
	PageSize    int                      `json:"page_size"`
	TotalItems  int64                    `json:"total_items"`
	TotalPages  int                      `json:"total_pages"`
	HasNext     bool                     `json:"has_next"`
	HasPrevious bool                     `json:"has_previous"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToProjectUpdateResponse converts a ProjectUpdate model to ProjectUpdateResponse DTO
func ToProjectUpdateResponse(update *models.ProjectUpdate) *ProjectUpdateResponse {
	if update == nil {
		return nil
	}

	resp := &ProjectUpdateResponse{
		ID:                update.ID,
		TenantID:          update.TenantID,
		ProjectID:         update.ProjectID,
		UserID:            update.UserID,
		Type:              update.Type,
		Title:             update.Title,
		Description:       update.Description,
		VisibleToCustomer: update.VisibleToCustomer,
		AttachmentURLs:    update.AttachmentURLs,
		Metadata:          update.Metadata,
		CreatedAt:         update.CreatedAt,
		UpdatedAt:         update.UpdatedAt,
	}

	// Add user if available
	if update.User != nil {
		resp.User = &UserSummary{
			ID:        update.User.ID,
			FirstName: update.User.FirstName,
			LastName:  update.User.LastName,
			Email:     update.User.Email,
			AvatarURL: update.User.AvatarURL,
		}
	}

	return resp
}

// ToProjectUpdateResponses converts multiple ProjectUpdate models to DTOs
func ToProjectUpdateResponses(updates []*models.ProjectUpdate) []*ProjectUpdateResponse {
	if updates == nil {
		return nil
	}

	responses := make([]*ProjectUpdateResponse, len(updates))
	for i, update := range updates {
		responses[i] = ToProjectUpdateResponse(update)
	}
	return responses
}
