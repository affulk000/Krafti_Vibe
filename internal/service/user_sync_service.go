package service

import (
	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"go.uber.org/zap"
)

// UserSyncService handles synchronization between Zitadel and application database
type UserSyncService struct {
	userRepo repository.UserRepository
	logger   *zap.Logger
}

// NewUserSyncService creates a new user sync service
func NewUserSyncService(userRepo repository.UserRepository, logger *zap.Logger) *UserSyncService {
	return &UserSyncService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// SyncUserFromZitadel synchronizes user data from Zitadel to local database
// This should be called on every authentication to keep user data in sync
func (s *UserSyncService) SyncUserFromZitadel(ctx context.Context, authCtx *oauth.IntrospectionContext) (*models.User, error) {
	if authCtx == nil {
		return nil, fmt.Errorf("introspection context is nil")
	}

	zitadelUserID := authCtx.UserID()
	if zitadelUserID == "" {
		return nil, fmt.Errorf("zitadel user ID is empty")
	}

	s.logger.Info("syncing user from Zitadel",
		zap.String("zitadel_user_id", zitadelUserID),
		zap.String("email", authCtx.Email),
		zap.String("username", authCtx.Username),
	)

	// Try to find existing user by Zitadel ID
	existingUser, err := s.userRepo.GetByZitadelID(ctx, zitadelUserID)
	if err == nil && existingUser != nil {
		// User exists - update and return
		return s.updateExistingUser(ctx, existingUser, authCtx)
	}

	// User doesn't exist - try to find by email (for migration scenarios)
	if authCtx.Email != "" {
		existingUser, err = s.userRepo.GetByEmail(ctx, authCtx.Email)
		if err == nil && existingUser != nil {
			// Link existing user to Zitadel
			return s.linkExistingUser(ctx, existingUser, authCtx)
		}
	}

	// User doesn't exist - create new user
	return s.createNewUser(ctx, authCtx)
}

// updateExistingUser updates an existing user with latest Zitadel data
func (s *UserSyncService) updateExistingUser(ctx context.Context, user *models.User, authCtx *oauth.IntrospectionContext) (*models.User, error) {
	s.logger.Debug("updating existing user",
		zap.String("user_id", user.ID.String()),
		zap.String("zitadel_user_id", user.ZitadelUserID),
	)

	now := time.Now()
	user.LastLoginAt = &now

	// Update basic info from Zitadel if changed
	if authCtx.Email != "" && authCtx.Email != user.Email {
		user.Email = authCtx.Email
	}

	// Update name if available
	if authCtx.Username != "" {
		// Parse username if it contains first/last name
		// For now, just update if empty
		if user.FirstName == "" {
			user.FirstName = authCtx.Username
		}
	}

	// Check for platform-level roles from Zitadel
	zitadelRoles := extractZitadelRoles(authCtx)
	platformRole := determinePlatformRole(zitadelRoles)
	if platformRole != "" {
		user.Role = platformRole
		user.IsPlatformUser = true
		user.TenantID = nil // Platform users have no tenant
	}

	// Save updates
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update user", zap.Error(err))
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info("user updated successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("role", string(user.Role)),
	)

	return user, nil
}

// linkExistingUser links an existing user (found by email) to their Zitadel account
func (s *UserSyncService) linkExistingUser(ctx context.Context, user *models.User, authCtx *oauth.IntrospectionContext) (*models.User, error) {
	s.logger.Info("linking existing user to Zitadel",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
		zap.String("zitadel_user_id", authCtx.UserID()),
	)

	now := time.Now()
	user.ZitadelUserID = authCtx.UserID()
	user.AuthProvider = "zitadel"
	user.MigrationStatus = "completed"
	user.MigratedAt = &now
	user.LastLoginAt = &now

	// Check for platform roles
	zitadelRoles := extractZitadelRoles(authCtx)
	platformRole := determinePlatformRole(zitadelRoles)
	if platformRole != "" {
		user.Role = platformRole
		user.IsPlatformUser = true
		user.TenantID = nil
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to link user", zap.Error(err))
		return nil, fmt.Errorf("failed to link user: %w", err)
	}

	s.logger.Info("user linked successfully", zap.String("user_id", user.ID.String()))
	return user, nil
}

// createNewUser creates a new user from Zitadel authentication context
func (s *UserSyncService) createNewUser(ctx context.Context, authCtx *oauth.IntrospectionContext) (*models.User, error) {
	s.logger.Info("creating new user from Zitadel",
		zap.String("zitadel_user_id", authCtx.UserID()),
		zap.String("email", authCtx.Email),
	)

	// Determine role based on Zitadel roles
	zitadelRoles := extractZitadelRoles(authCtx)
	platformRole := determinePlatformRole(zitadelRoles)

	var role models.UserRole
	var isPlatformUser bool
	var tenantID *uuid.UUID

	if platformRole != "" {
		// Platform user
		role = platformRole
		isPlatformUser = true
		tenantID = nil
	} else {
		// Regular tenant user - default to customer
		role = models.UserRoleCustomer
		isPlatformUser = false

		// Try to extract tenant from organization ID
		if authCtx.OrganizationID() != "" {
			if tid, err := uuid.Parse(authCtx.OrganizationID()); err == nil {
				tenantID = &tid
			}
		}
	}

	// Parse names from username or email
	firstName, lastName := parseUsername(authCtx.Username, authCtx.Email)

	now := time.Now()
	user := &models.User{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		ZitadelUserID:   authCtx.UserID(),
		AuthProvider:    "zitadel",
		MigrationStatus: "completed",
		MigratedAt:      &now,
		Email:           authCtx.Email,
		FirstName:       firstName,
		LastName:        lastName,
		Role:            role,
		Status:          models.UserStatusActive,
		EmailVerified:   authCtx.Email != "", // Assume verified if from Zitadel
		IsPlatformUser:  isPlatformUser,
		TenantID:        tenantID,
		Timezone:        "UTC",
		Language:        "en",
		LastLoginAt:     &now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("user created successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("role", string(user.Role)),
		zap.Bool("is_platform_user", user.IsPlatformUser),
	)

	return user, nil
}

// Helper functions

// extractZitadelRoles extracts roles from Zitadel introspection context
func extractZitadelRoles(authCtx *oauth.IntrospectionContext) []string {
	if authCtx == nil || authCtx.Claims == nil {
		return []string{}
	}

	roles := []string{}

	// Check for roles in the standard Zitadel claim
	if rolesRaw, ok := authCtx.Claims["urn:zitadel:iam:org:project:roles"]; ok {
		if rolesMap, ok := rolesRaw.(map[string]interface{}); ok {
			for role := range rolesMap {
				roles = append(roles, role)
			}
		}
	}

	return roles
}

// determinePlatformRole determines if user has a platform role from Zitadel
func determinePlatformRole(zitadelRoles []string) models.UserRole {
	// Check roles in priority order
	for _, role := range zitadelRoles {
		switch role {
		case "platform_super_admin":
			return models.UserRolePlatformSuperAdmin
		case "platform_admin":
			return models.UserRolePlatformAdmin
		case "platform_support":
			return models.UserRolePlatformSupport
		}
	}

	return "" // No platform role
}

// parseUsername attempts to extract first and last name from username or email
func parseUsername(username, email string) (string, string) {
	if username != "" {
		// If username is "john.doe" or "john doe"
		// This is a simple implementation - enhance as needed
		return username, ""
	}

	if email != "" {
		// Extract name from email like "john.doe@example.com"
		return email, ""
	}

	return "User", ""
}

// GetOrSyncUser gets a user from DB or syncs from Zitadel if needed
func (s *UserSyncService) GetOrSyncUser(ctx context.Context, zitadelUserID string, authCtx *oauth.IntrospectionContext) (interface{}, error) {
	// Try to get from database first
	user, err := s.userRepo.GetByZitadelID(ctx, zitadelUserID)
	if err == nil && user != nil {
		// User exists, update last login
		now := time.Now()
		user.LastLoginAt = &now
		_ = s.userRepo.Update(ctx, user)
		return user, nil
	}

	// User not found, sync from Zitadel
	syncedUser, syncErr := s.SyncUserFromZitadel(ctx, authCtx)
	if syncErr != nil {
		return nil, syncErr
	}

	// Ensure we return nil interface if user is nil
	if syncedUser == nil {
		return nil, nil
	}

	return syncedUser, nil
}
