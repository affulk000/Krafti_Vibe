package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Invitation-related errors
var (
	// ErrInvitationNotFound is returned when invitation is not found
	ErrInvitationNotFound = errors.New("invitation not found")

	// ErrInvitationExpired is returned when invitation has expired
	ErrInvitationExpired = errors.New("invitation has expired")

	// ErrInvitationAlreadyAccepted is returned when invitation was already accepted
	ErrInvitationAlreadyAccepted = errors.New("invitation has already been accepted")

	// ErrInvitationAlreadyExists is returned when a pending invitation exists for the email
	ErrInvitationAlreadyExists = errors.New("pending invitation already exists for this email")

	// ErrUserAlreadyMember is returned when user is already a tenant member
	ErrUserAlreadyMember = errors.New("user is already a member of this tenant")

	// ErrUserBelongsToAnotherTenant is returned when user belongs to another tenant
	ErrUserBelongsToAnotherTenant = errors.New("user already belongs to another tenant")
)

// TenantInvitationService defines the interface for tenant invitation operations
type TenantInvitationService interface {
	// Core CRUD Operations
	CreateInvitation(ctx context.Context, req *dto.CreateInvitationRequest, tenantID uuid.UUID, invitedBy uuid.UUID) (*dto.InvitationResponse, error)
	GetInvitation(ctx context.Context, id uuid.UUID) (*dto.InvitationResponse, error)
	GetInvitationByToken(ctx context.Context, token string) (*dto.InvitationResponse, error)
	ListInvitations(ctx context.Context, tenantID uuid.UUID, filter *dto.InvitationFilter) (*dto.InvitationListResponse, error)

	// Invitation Actions
	AcceptInvitation(ctx context.Context, req *dto.AcceptInvitationRequest) (*dto.AcceptInvitationResponse, error)
	RevokeInvitation(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
	ResendInvitation(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.InvitationResponse, error)

	// Query Operations
	GetPendingInvitations(ctx context.Context, tenantID uuid.UUID) ([]*dto.InvitationResponse, error)
	GetInvitationsByEmail(ctx context.Context, email string) ([]*dto.InvitationResponse, error)

	// Cleanup Operations
	DeleteExpiredInvitations(ctx context.Context) (int64, error)
}

// tenantInvitationService implements TenantInvitationService
type tenantInvitationService struct {
	repos  *repository.Repositories
	logger *zap.Logger
}

// NewTenantInvitationService creates a new tenant invitation service
func NewTenantInvitationService(
	repos *repository.Repositories,
	logger *zap.Logger,
) TenantInvitationService {
	return &tenantInvitationService{
		repos:  repos,
		logger: logger,
	}
}

// ============================================================================
// Core CRUD Operations
// ============================================================================

// CreateInvitation creates a new tenant invitation
func (s *tenantInvitationService) CreateInvitation(ctx context.Context, req *dto.CreateInvitationRequest, tenantID uuid.UUID, invitedBy uuid.UUID) (*dto.InvitationResponse, error) {
	req.Sanitize()
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	s.logger.Info("creating invitation",
		zap.String("tenant_id", tenantID.String()),
		zap.String("email", req.Email),
		zap.String("role", string(req.Role)),
	)

	// Verify tenant exists and is in valid state
	tenant, err := s.repos.Tenant.GetByID(ctx, tenantID)
	if err != nil {
		return nil, ErrTenantNotFound
	}

	// Check tenant status
	if err := s.validateTenantStatus(tenant); err != nil {
		return nil, err
	}

	// Check if user already exists in tenant
	existingUser, _ := s.repos.User.GetByEmail(ctx, req.Email)
	if existingUser != nil && existingUser.TenantID != nil && *existingUser.TenantID == tenantID {
		return nil, ErrUserAlreadyMember
	}

	// Check for existing pending invitation
	if err := s.checkExistingInvitation(ctx, tenantID, req.Email); err != nil {
		return nil, err
	}

	// Check user limit
	if tenant.MaxUsers > 0 && tenant.CurrentUsers >= tenant.MaxUsers {
		return nil, ErrTenantLimitReached
	}

	// Generate secure token
	token, err := s.generateSecureToken(32)
	if err != nil {
		s.logger.Error("failed to generate invitation token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	// Set expiry (default 7 days, max 30 days)
	expiryDays := max(1, min(req.ExpiryDays, 30))
	if expiryDays <= 0 {
		expiryDays = 7
	}
	expiresAt := time.Now().AddDate(0, 0, expiryDays)

	invitation := &models.TenantInvitation{
		TenantID:  tenantID,
		Email:     req.Email,
		Role:      req.Role,
		Token:     token,
		ExpiresAt: expiresAt,
		InvitedBy: invitedBy,
	}

	if err := s.repos.TenantInvitation.Create(ctx, invitation); err != nil {
		s.logger.Error("failed to create invitation", zap.Error(err))
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	s.logger.Info("invitation created successfully",
		zap.String("invitation_id", invitation.ID.String()),
		zap.String("tenant_id", tenantID.String()),
		zap.String("email", req.Email),
	)

	// TODO: Send invitation email asynchronously via email service

	return dto.ToInvitationResponse(invitation, tenant), nil
}

// GetInvitation retrieves an invitation by ID
func (s *tenantInvitationService) GetInvitation(ctx context.Context, id uuid.UUID) (*dto.InvitationResponse, error) {
	invitation, err := s.repos.TenantInvitation.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInvitationNotFound
	}

	tenant, _ := s.repos.Tenant.GetByID(ctx, invitation.TenantID)
	return dto.ToInvitationResponse(invitation, tenant), nil
}

// GetInvitationByToken retrieves an invitation by token
func (s *tenantInvitationService) GetInvitationByToken(ctx context.Context, token string) (*dto.InvitationResponse, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	invitation, err := s.repos.TenantInvitation.FindByToken(ctx, token)
	if err != nil {
		return nil, ErrInvitationNotFound
	}

	tenant, _ := s.repos.Tenant.GetByID(ctx, invitation.TenantID)
	return dto.ToInvitationResponse(invitation, tenant), nil
}

// ListInvitations lists invitations for a tenant with pagination
func (s *tenantInvitationService) ListInvitations(ctx context.Context, tenantID uuid.UUID, filter *dto.InvitationFilter) (*dto.InvitationListResponse, error) {
	if filter == nil {
		filter = &dto.InvitationFilter{
			Page:     1,
			PageSize: 20,
		}
	}

	// Validate pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	filter.PageSize = max(1, min(filter.PageSize, 100))
	if filter.PageSize == 0 {
		filter.PageSize = 20
	}

	pagination := repository.PaginationParams{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	invitations, result, err := s.repos.TenantInvitation.FindByTenant(ctx, tenantID, pagination)
	if err != nil {
		s.logger.Error("failed to list invitations", zap.Error(err))
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}

	tenant, _ := s.repos.Tenant.GetByID(ctx, tenantID)

	// Convert to response DTOs and filter by status if requested
	responses := make([]*dto.InvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		response := dto.ToInvitationResponse(inv, tenant)

		// Apply status filter if provided
		if filter.Status != nil && *filter.Status != "" {
			if response.Status != *filter.Status {
				continue
			}
		}

		responses = append(responses, response)
	}

	return &dto.InvitationListResponse{
		Invitations: responses,
		Page:        result.Page,
		PageSize:    result.PageSize,
		TotalItems:  result.TotalItems,
		TotalPages:  result.TotalPages,
	}, nil
}

// ============================================================================
// Invitation Actions
// ============================================================================

// AcceptInvitation accepts a tenant invitation
func (s *tenantInvitationService) AcceptInvitation(ctx context.Context, req *dto.AcceptInvitationRequest) (*dto.AcceptInvitationResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	s.logger.Info("accepting invitation", zap.String("token", req.Token[:8]+"..."))

	// Find invitation by token
	invitation, err := s.repos.TenantInvitation.FindByToken(ctx, req.Token)
	if err != nil {
		return nil, ErrInvitationNotFound
	}

	// Validate invitation state
	if invitation.IsAccepted() {
		return nil, ErrInvitationAlreadyAccepted
	}

	if invitation.IsExpired() {
		return nil, ErrInvitationExpired
	}

	// Get and validate tenant
	tenant, err := s.repos.Tenant.GetByID(ctx, invitation.TenantID)
	if err != nil {
		return nil, ErrTenantNotFound
	}

	if err := s.validateTenantStatus(tenant); err != nil {
		return nil, err
	}

	// Check user limit
	if tenant.MaxUsers > 0 && tenant.CurrentUsers >= tenant.MaxUsers {
		return nil, ErrTenantLimitReached
	}

	// Handle user creation or update
	user, err := s.handleUserForInvitation(ctx, invitation)
	if err != nil {
		return nil, err
	}

	// Mark invitation as accepted
	now := time.Now()
	invitation.AcceptedAt = &now
	if err := s.repos.TenantInvitation.Update(ctx, invitation); err != nil {
		s.logger.Error("failed to update invitation", zap.Error(err))
		// Continue - user is already added
	}

	// Increment user count
	if err := s.repos.Tenant.IncrementUserCount(ctx, tenant.ID); err != nil {
		s.logger.Warn("failed to increment user count", zap.Error(err))
	}

	s.logger.Info("invitation accepted successfully",
		zap.String("invitation_id", invitation.ID.String()),
		zap.String("user_id", user.ID.String()),
		zap.String("tenant_id", tenant.ID.String()),
	)

	return &dto.AcceptInvitationResponse{
		Success:    true,
		Message:    "Invitation accepted successfully",
		TenantID:   tenant.ID,
		TenantName: tenant.Name,
		UserID:     user.ID,
		Role:       user.Role,
	}, nil
}

// RevokeInvitation revokes a pending invitation
func (s *tenantInvitationService) RevokeInvitation(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	s.logger.Info("revoking invitation",
		zap.String("invitation_id", id.String()),
		zap.String("tenant_id", tenantID.String()),
	)

	invitation, err := s.repos.TenantInvitation.GetByID(ctx, id)
	if err != nil {
		return ErrInvitationNotFound
	}

	// Verify tenant ownership
	if invitation.TenantID != tenantID {
		return ErrInvitationNotFound // Don't expose that it exists for another tenant
	}

	if invitation.IsAccepted() {
		return fmt.Errorf("cannot revoke an accepted invitation")
	}

	if err := s.repos.TenantInvitation.Delete(ctx, id); err != nil {
		s.logger.Error("failed to revoke invitation", zap.Error(err))
		return fmt.Errorf("failed to revoke invitation: %w", err)
	}

	s.logger.Info("invitation revoked successfully", zap.String("invitation_id", id.String()))
	return nil
}

// ResendInvitation resends an invitation with a new token
func (s *tenantInvitationService) ResendInvitation(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.InvitationResponse, error) {
	s.logger.Info("resending invitation",
		zap.String("invitation_id", id.String()),
		zap.String("tenant_id", tenantID.String()),
	)

	invitation, err := s.repos.TenantInvitation.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInvitationNotFound
	}

	// Verify tenant ownership
	if invitation.TenantID != tenantID {
		return nil, ErrInvitationNotFound
	}

	if invitation.IsAccepted() {
		return nil, fmt.Errorf("cannot resend an accepted invitation")
	}

	// Generate new token
	newToken, err := s.generateSecureToken(32)
	if err != nil {
		s.logger.Error("failed to generate new invitation token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate new token: %w", err)
	}

	// Extend expiry by 7 days from now
	invitation.Token = newToken
	invitation.ExpiresAt = time.Now().AddDate(0, 0, 7)

	if err := s.repos.TenantInvitation.Update(ctx, invitation); err != nil {
		s.logger.Error("failed to update invitation", zap.Error(err))
		return nil, fmt.Errorf("failed to resend invitation: %w", err)
	}

	// TODO: Send invitation email asynchronously via email service

	tenant, _ := s.repos.Tenant.GetByID(ctx, tenantID)

	s.logger.Info("invitation resent successfully", zap.String("invitation_id", id.String()))
	return dto.ToInvitationResponse(invitation, tenant), nil
}

// ============================================================================
// Query Operations
// ============================================================================

// GetPendingInvitations retrieves all pending invitations for a tenant
func (s *tenantInvitationService) GetPendingInvitations(ctx context.Context, tenantID uuid.UUID) ([]*dto.InvitationResponse, error) {
	invitations, err := s.repos.TenantInvitation.FindPendingByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get pending invitations", zap.Error(err))
		return nil, fmt.Errorf("failed to get pending invitations: %w", err)
	}

	tenant, _ := s.repos.Tenant.GetByID(ctx, tenantID)

	responses := make([]*dto.InvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		responses = append(responses, dto.ToInvitationResponse(inv, tenant))
	}

	return responses, nil
}

// GetInvitationsByEmail retrieves all invitations for an email address
func (s *tenantInvitationService) GetInvitationsByEmail(ctx context.Context, email string) ([]*dto.InvitationResponse, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	invitations, err := s.repos.TenantInvitation.FindByEmail(ctx, email)
	if err != nil {
		s.logger.Error("failed to get invitations by email", zap.Error(err))
		return nil, fmt.Errorf("failed to get invitations: %w", err)
	}

	responses := make([]*dto.InvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		tenant, _ := s.repos.Tenant.GetByID(ctx, inv.TenantID)
		responses = append(responses, dto.ToInvitationResponse(inv, tenant))
	}

	return responses, nil
}

// ============================================================================
// Cleanup Operations
// ============================================================================

// DeleteExpiredInvitations deletes all expired invitations
func (s *tenantInvitationService) DeleteExpiredInvitations(ctx context.Context) (int64, error) {
	s.logger.Info("cleaning up expired invitations")

	count, err := s.repos.TenantInvitation.DeleteExpiredInvitations(ctx)
	if err != nil {
		s.logger.Error("failed to delete expired invitations", zap.Error(err))
		return 0, fmt.Errorf("failed to delete expired invitations: %w", err)
	}

	s.logger.Info("expired invitations deleted", zap.Int64("count", count))
	return count, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// validateTenantStatus checks if tenant is in a valid state for operations
func (s *tenantInvitationService) validateTenantStatus(tenant *models.Tenant) error {
	if tenant.Status == models.TenantStatusSuspended {
		return ErrTenantSuspended
	}
	if tenant.Status == models.TenantStatusCancelled {
		return ErrTenantCancelled
	}
	return nil
}

// checkExistingInvitation checks if a pending invitation already exists
func (s *tenantInvitationService) checkExistingInvitation(ctx context.Context, tenantID uuid.UUID, email string) error {
	existingInvitations, err := s.repos.TenantInvitation.FindByTenantAndEmail(ctx, tenantID, email)
	if err != nil {
		return nil // No existing invitations
	}

	for _, inv := range existingInvitations {
		if !inv.IsExpired() && !inv.IsAccepted() {
			return ErrInvitationAlreadyExists
		}
	}

	return nil
}

// handleUserForInvitation handles user creation or update during invitation acceptance
func (s *tenantInvitationService) handleUserForInvitation(ctx context.Context, invitation *models.TenantInvitation) (*models.User, error) {
	existingUser, err := s.repos.User.GetByEmail(ctx, invitation.Email)

	if err == nil && existingUser != nil {
		// Existing user - update their tenant
		if existingUser.TenantID != nil && *existingUser.TenantID != invitation.TenantID {
			return nil, ErrUserBelongsToAnotherTenant
		}

		existingUser.TenantID = &invitation.TenantID
		existingUser.Role = invitation.Role

		if err := s.repos.User.Update(ctx, existingUser); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		return existingUser, nil
	}

	// New user - create account with pending status
	user := &models.User{
		Email:    invitation.Email,
		TenantID: &invitation.TenantID,
		Role:     invitation.Role,
		Status:   models.UserStatusPending, // They need to complete registration
	}

	if err := s.repos.User.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// generateSecureToken generates a cryptographically secure random token
func (s *tenantInvitationService) generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
