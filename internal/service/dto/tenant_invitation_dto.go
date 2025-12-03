package dto

import (
	"fmt"
	"strings"
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// Invitation DTOs
// ============================================================================

// CreateInvitationRequest represents the request to create an invitation
type CreateInvitationRequest struct {
	Email      string          `json:"email" validate:"required,email"`
	Role       models.UserRole `json:"role" validate:"required"`
	ExpiryDays int             `json:"expiry_days,omitempty"` // Default: 7
}

// Validate validates the create invitation request
func (r *CreateInvitationRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if r.Role == "" {
		return fmt.Errorf("role is required")
	}
	if r.ExpiryDays < 0 || r.ExpiryDays > 30 {
		return fmt.Errorf("expiry_days must be between 0 and 30")
	}
	return nil
}

// Sanitize sanitizes the create invitation request
func (r *CreateInvitationRequest) Sanitize() {
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
}

// AcceptInvitationRequest represents the request to accept an invitation
type AcceptInvitationRequest struct {
	Token string `json:"token" validate:"required"`
}

// Validate validates the accept invitation request
func (r *AcceptInvitationRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return fmt.Errorf("token is required")
	}
	return nil
}

// InvitationFilter represents filters for listing invitations
type InvitationFilter struct {
	Status   *string `json:"status,omitempty"` // pending, accepted, expired
	Page     int     `json:"page" validate:"min=1"`
	PageSize int     `json:"page_size" validate:"min=1,max=100"`
}

// InvitationResponse represents an invitation response
type InvitationResponse struct {
	ID          uuid.UUID       `json:"id"`
	TenantID    uuid.UUID       `json:"tenant_id"`
	TenantName  string          `json:"tenant_name,omitempty"`
	Email       string          `json:"email"`
	Role        models.UserRole `json:"role"`
	Status      string          `json:"status"` // pending, accepted, expired
	ExpiresAt   time.Time       `json:"expires_at"`
	AcceptedAt  *time.Time      `json:"accepted_at,omitempty"`
	InvitedBy   uuid.UUID       `json:"invited_by"`
	InviterName string          `json:"inviter_name,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// InvitationListResponse represents a paginated list of invitations
type InvitationListResponse struct {
	Invitations []*InvitationResponse `json:"invitations"`
	Page        int                   `json:"page"`
	PageSize    int                   `json:"page_size"`
	TotalItems  int64                 `json:"total_items"`
	TotalPages  int                   `json:"total_pages"`
}

// AcceptInvitationResponse represents the response after accepting an invitation
type AcceptInvitationResponse struct {
	Success    bool            `json:"success"`
	Message    string          `json:"message"`
	TenantID   uuid.UUID       `json:"tenant_id"`
	TenantName string          `json:"tenant_name"`
	UserID     uuid.UUID       `json:"user_id"`
	Role       models.UserRole `json:"role"`
}

// ToInvitationResponse converts a models.TenantInvitation to InvitationResponse
func ToInvitationResponse(invitation *models.TenantInvitation, tenant *models.Tenant) *InvitationResponse {
	if invitation == nil {
		return nil
	}

	status := "pending"
	if invitation.AcceptedAt != nil {
		status = "accepted"
	} else if invitation.IsExpired() {
		status = "expired"
	}

	response := &InvitationResponse{
		ID:         invitation.ID,
		TenantID:   invitation.TenantID,
		Email:      invitation.Email,
		Role:       invitation.Role,
		Status:     status,
		ExpiresAt:  invitation.ExpiresAt,
		AcceptedAt: invitation.AcceptedAt,
		InvitedBy:  invitation.InvitedBy,
		CreatedAt:  invitation.CreatedAt,
	}

	if tenant != nil {
		response.TenantName = tenant.Name
	}

	if invitation.Inviter != nil {
		response.InviterName = invitation.Inviter.FirstName + " " + invitation.Inviter.LastName
	}

	return response
}

// ============================================================================
// Usage Tracking DTOs
// ============================================================================

// TrackFeatureUsageRequest represents the request to track feature usage
type TrackFeatureUsageRequest struct {
	Feature string `json:"feature" validate:"required"`
	Count   int    `json:"count" validate:"min=1"`
}

// Validate validates the track feature usage request
func (r *TrackFeatureUsageRequest) Validate() error {
	if strings.TrimSpace(r.Feature) == "" {
		return fmt.Errorf("feature is required")
	}
	if r.Count < 1 {
		r.Count = 1
	}
	return nil
}

// UsageTrackingResponse represents usage tracking data
type UsageTrackingResponse struct {
	TenantID        uuid.UUID `json:"tenant_id"`
	Date            time.Time `json:"date"`
	APICallsCount   int64     `json:"api_calls_count"`
	APICallsLimit   int64     `json:"api_calls_limit"`
	APIUsagePercent float64   `json:"api_usage_percent"`
	StorageUsedGB   int64     `json:"storage_used_gb"`
	BandwidthUsedGB int64     `json:"bandwidth_used_gb"`
	BookingsCreated int       `json:"bookings_created"`
	ProjectsCreated int       `json:"projects_created"`
	SMSSent         int       `json:"sms_sent"`
	EmailsSent      int       `json:"emails_sent"`
}

// ToUsageTrackingResponse converts a models.TenantUsageTracking to UsageTrackingResponse
func ToUsageTrackingResponse(usage *models.TenantUsageTracking) *UsageTrackingResponse {
	if usage == nil {
		return nil
	}

	usagePercent := float64(0)
	if usage.APICallsLimit > 0 {
		usagePercent = (float64(usage.APICallsCount) / float64(usage.APICallsLimit)) * 100
	}

	return &UsageTrackingResponse{
		TenantID:        usage.TenantID,
		Date:            usage.Date,
		APICallsCount:   usage.APICallsCount,
		APICallsLimit:   usage.APICallsLimit,
		APIUsagePercent: usagePercent,
		StorageUsedGB:   usage.StorageUsedGB,
		BandwidthUsedGB: usage.BandwidthUsedGB,
		BookingsCreated: usage.BookingsCreated,
		ProjectsCreated: usage.ProjectsCreated,
		SMSSent:         usage.SMSSent,
		EmailsSent:      usage.EmailsSent,
	}
}

// UsageHistoryResponse represents usage history
type UsageHistoryResponse struct {
	TenantID   uuid.UUID                `json:"tenant_id"`
	StartDate  time.Time                `json:"start_date"`
	EndDate    time.Time                `json:"end_date"`
	DailyUsage []*UsageTrackingResponse `json:"daily_usage"`
	Summary    *UsageSummary            `json:"summary"`
}

// UsageSummary represents aggregated usage summary
type UsageSummary struct {
	TotalAPICalls    int64   `json:"total_api_calls"`
	TotalBookings    int     `json:"total_bookings"`
	TotalProjects    int     `json:"total_projects"`
	TotalSMS         int     `json:"total_sms"`
	TotalEmails      int     `json:"total_emails"`
	AverageAPICalls  float64 `json:"average_api_calls_per_day"`
	PeakAPICalls     int64   `json:"peak_api_calls"`
	PeakAPICallsDate string  `json:"peak_api_calls_date"`
	TotalStorageUsed int64   `json:"total_storage_used_gb"`
	TotalBandwidth   int64   `json:"total_bandwidth_gb"`
}

// ============================================================================
// Data Export DTOs
// ============================================================================

// DataExportRequest represents the request to export data
type DataExportRequest struct {
	ExportType string   `json:"export_type" validate:"required"` // full, partial, gdpr
	Tables     []string `json:"tables,omitempty"`                // For partial exports
	Format     string   `json:"format,omitempty"`                // json, csv (default: json)
}

// Validate validates the data export request
func (r *DataExportRequest) Validate() error {
	if strings.TrimSpace(r.ExportType) == "" {
		return fmt.Errorf("export_type is required")
	}
	validTypes := map[string]bool{
		"full":    true,
		"partial": true,
		"gdpr":    true,
	}
	if !validTypes[r.ExportType] {
		return fmt.Errorf("export_type must be one of: full, partial, gdpr")
	}
	if r.ExportType == "partial" && len(r.Tables) == 0 {
		return fmt.Errorf("tables are required for partial export")
	}
	if r.Format == "" {
		r.Format = "json"
	}
	return nil
}

// DataExportFilter represents filters for listing data exports
type DataExportFilter struct {
	Status   *string `json:"status,omitempty"` // pending, processing, completed, failed, cancelled
	Page     int     `json:"page" validate:"min=1"`
	PageSize int     `json:"page_size" validate:"min=1,max=100"`
}

// DataExportResponse represents a data export response
type DataExportResponse struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	RequestedBy   uuid.UUID  `json:"requested_by"`
	RequesterName string     `json:"requester_name,omitempty"`
	ExportType    string     `json:"export_type"`
	Status        string     `json:"status"`
	FileURL       string     `json:"file_url,omitempty"`
	FileSize      int64      `json:"file_size,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	ErrorMessage  string     `json:"error_message,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// DataExportListResponse represents a paginated list of data exports
type DataExportListResponse struct {
	Exports    []*DataExportResponse `json:"exports"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalItems int64                 `json:"total_items"`
	TotalPages int                   `json:"total_pages"`
}

// ToDataExportResponse converts a models.DataExportRequest to DataExportResponse
func ToDataExportResponse(export *models.DataExportRequest) *DataExportResponse {
	if export == nil {
		return nil
	}

	response := &DataExportResponse{
		ID:          export.ID,
		TenantID:    export.TenantID,
		RequestedBy: export.RequestedBy,
		ExportType:  export.ExportType,
		Status:      export.Status,
		FileURL:     export.FileURL,
		ExpiresAt:   export.ExpiresAt,
		CompletedAt: export.CompletedAt,
		CreatedAt:   export.CreatedAt,
		UpdatedAt:   export.UpdatedAt,
	}

	if export.Requester != nil {
		response.RequesterName = export.Requester.FirstName + " " + export.Requester.LastName
	}

	return response
}

// ============================================================================
// Tenant Member DTOs
// ============================================================================

// TenantMemberResponse represents a tenant member
type TenantMemberResponse struct {
	UserID    uuid.UUID         `json:"user_id"`
	Email     string            `json:"email"`
	FirstName string            `json:"first_name"`
	LastName  string            `json:"last_name"`
	Role      models.UserRole   `json:"role"`
	Status    models.UserStatus `json:"status"`
	JoinedAt  time.Time         `json:"joined_at"`
	LastLogin *time.Time        `json:"last_login,omitempty"`
}

// TenantMembersListResponse represents a list of tenant members
type TenantMembersListResponse struct {
	Members    []*TenantMemberResponse `json:"members"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	TotalItems int64                   `json:"total_items"`
	TotalPages int                     `json:"total_pages"`
}

// UpdateMemberRoleRequest represents the request to update a member's role
type UpdateMemberRoleRequest struct {
	Role models.UserRole `json:"role" validate:"required"`
}

// Validate validates the update member role request
func (r *UpdateMemberRoleRequest) Validate() error {
	if r.Role == "" {
		return fmt.Errorf("role is required")
	}
	return nil
}

// RemoveMemberRequest represents the request to remove a member
type RemoveMemberRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Reason string    `json:"reason,omitempty"`
}

// Validate validates the remove member request
func (r *RemoveMemberRequest) Validate() error {
	if r.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	return nil
}

// ============================================================================
// Audit Log DTOs
// ============================================================================

// TenantAuditLogResponse represents an audit log entry
type TenantAuditLogResponse struct {
	ID         uuid.UUID              `json:"id"`
	TenantID   uuid.UUID              `json:"tenant_id"`
	UserID     uuid.UUID              `json:"user_id"`
	UserEmail  string                 `json:"user_email,omitempty"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id,omitempty"`
	Changes    map[string]interface{} `json:"changes,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// TenantAuditLogListResponse represents a list of audit logs
type TenantAuditLogListResponse struct {
	Logs       []*TenantAuditLogResponse `json:"logs"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalItems int64                     `json:"total_items"`
	TotalPages int                       `json:"total_pages"`
}

// AuditLogFilter represents filters for listing audit logs
type AuditLogFilter struct {
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Action    *string    `json:"action,omitempty"`
	Resource  *string    `json:"resource,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Page      int        `json:"page" validate:"min=1"`
	PageSize  int        `json:"page_size" validate:"min=1,max=100"`
}
