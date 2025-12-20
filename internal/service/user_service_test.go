package service_test

import (
	"context"
	"testing"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service"
	"Krafti_Vibe/internal/service/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================================
// Test Setup
// ============================================================================

func setupUserServiceTest() (*repository.Repositories, service.UserService) {
	userRepo := new(MockUserRepository)
	logger := &MockLogger{}

	repos := &repository.Repositories{
		User: userRepo,
	}

	svc := service.NewUserService(repos, logger)

	return repos, svc
}

// ============================================================================
// Tests
// ============================================================================

func TestUserService_CreateUser(t *testing.T) {
	repos, svc := setupUserServiceTest()
	userRepo := repos.User.(*MockUserRepository)
	ctx := context.Background()

	t.Run("create user successfully", func(t *testing.T) {
		tenantID := uuid.New()

		// User doesn't exist yet
		userRepo.On("GetByEmail", ctx, "user@example.com").Return(nil, errors.ErrNotFound).Once()
		userRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil).Once()

		req := &dto.CreateUserRequest{
			TenantID:  &tenantID,
			Email:     "user@example.com",
			Password:  "password123",
			FirstName: "John",
			LastName:  "Doe",
			Role:      models.UserRoleCustomer,
		}

		resp, err := svc.CreateUser(ctx, req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "user@example.com", resp.Email)
		assert.Equal(t, "John", resp.FirstName)
		assert.Equal(t, "Doe", resp.LastName)
		userRepo.AssertExpectations(t)
	})

	t.Run("fail when email already exists", func(t *testing.T) {
		existingUser := &models.User{
			BaseModel: models.BaseModel{ID: uuid.New()},
			Email:     "existing@example.com",
		}

		userRepo.On("GetByEmail", ctx, "existing@example.com").Return(existingUser, nil).Once()

		req := &dto.CreateUserRequest{
			IsPlatformUser: true,
			Email:          "existing@example.com",
			Password:       "password123",
			FirstName:      "John",
			LastName:       "Doe",
			Role:           models.UserRoleCustomer,
		}

		resp, err := svc.CreateUser(ctx, req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "email already exists")
		userRepo.AssertExpectations(t)
	})

	t.Run("fail when tenant required but not provided", func(t *testing.T) {
		req := &dto.CreateUserRequest{
			IsPlatformUser: false,
			Email:          "user@example.com",
			Password:       "password123",
			FirstName:      "John",
			LastName:       "Doe",
			Role:           models.UserRoleCustomer,
		}

		resp, err := svc.CreateUser(ctx, req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "tenant_id is required")
	})

	t.Run("password is hashed correctly", func(t *testing.T) {
		tenantID := uuid.New()
		password := "mySecurePassword123"

		userRepo.On("GetByEmail", ctx, "test@example.com").Return(nil, errors.ErrNotFound).Once()

		var capturedUser *models.User
		userRepo.On("Create", ctx, mock.MatchedBy(func(u *models.User) bool {
			capturedUser = u
			return true
		})).Return(nil).Once()

		req := &dto.CreateUserRequest{
			TenantID:  &tenantID,
			Email:     "test@example.com",
			Password:  password,
			FirstName: "Test",
			LastName:  "User",
			Role:      models.UserRoleCustomer,
		}

		_, err := svc.CreateUser(ctx, req)

		require.NoError(t, err)
		assert.NotEmpty(t, capturedUser.PasswordHash)

		// Verify password was hashed
		err = bcrypt.CompareHashAndPassword([]byte(capturedUser.PasswordHash), []byte(password))
		assert.NoError(t, err, "Password should be hashed correctly")

		userRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUser(t *testing.T) {
	repos, svc := setupUserServiceTest()
	userRepo := repos.User.(*MockUserRepository)
	ctx := context.Background()

	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("user can get their own data", func(t *testing.T) {
		user := &models.User{
			BaseModel: models.BaseModel{ID: userID},
			Email:     "user@example.com",
			FirstName: "John",
			LastName:  "Doe",
		}

		// GetUser calls GetByID twice: once for the target user, once for the requesting user
		// Since userID == requestingUserID, both calls are with the same ID
		userRepo.On("GetByID", ctx, userID).Return(user, nil).Twice()

		resp, err := svc.GetUser(ctx, userID, userID)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, userID, resp.ID)
		userRepo.AssertExpectations(t)
	})

	t.Run("platform admin can access any user", func(t *testing.T) {
		targetUser := &models.User{
			BaseModel: models.BaseModel{ID: uuid.New()},
			Email:     "target@example.com",
		}

		adminID := uuid.New()
		adminUser := &models.User{
			BaseModel:      models.BaseModel{ID: adminID},
			Role:           models.UserRolePlatformAdmin,
			IsPlatformUser: true,
		}

		userRepo.On("GetByID", ctx, targetUser.ID).Return(targetUser, nil).Once()
		userRepo.On("GetByID", ctx, adminID).Return(adminUser, nil).Once()

		resp, err := svc.GetUser(ctx, targetUser.ID, adminID)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		userRepo.AssertExpectations(t)
	})

	t.Run("tenant admin can access users in their tenant", func(t *testing.T) {
		targetUserID := uuid.New()
		targetUser := &models.User{
			BaseModel: models.BaseModel{ID: targetUserID},
			TenantID:  &tenantID,
			Email:     "target@example.com",
		}

		adminID := uuid.New()
		adminUser := &models.User{
			BaseModel: models.BaseModel{ID: adminID},
			TenantID:  &tenantID,
			Role:      models.UserRoleTenantAdmin,
		}

		userRepo.On("GetByID", ctx, targetUserID).Return(targetUser, nil).Once()
		userRepo.On("GetByID", ctx, adminID).Return(adminUser, nil).Once()

		resp, err := svc.GetUser(ctx, targetUserID, adminID)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		userRepo.AssertExpectations(t)
	})

	t.Run("fail when user not found", func(t *testing.T) {
		userRepo.On("GetByID", ctx, userID).Return(nil, errors.ErrNotFound).Once()

		resp, err := svc.GetUser(ctx, userID, uuid.New())

		require.Error(t, err)
		assert.Nil(t, resp)
		userRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByEmail(t *testing.T) {
	repos, svc := setupUserServiceTest()
	userRepo := repos.User.(*MockUserRepository)
	ctx := context.Background()

	t.Run("get user by email successfully", func(t *testing.T) {
		user := &models.User{
			BaseModel: models.BaseModel{ID: uuid.New()},
			Email:     "user@example.com",
		}

		userRepo.On("GetByEmail", ctx, "user@example.com").Return(user, nil).Once()

		resp, err := svc.GetUserByEmail(ctx, "user@example.com", nil)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "user@example.com", resp.Email)
		userRepo.AssertExpectations(t)
	})

	t.Run("get user by email with tenant filter", func(t *testing.T) {
		tenantID := uuid.New()
		user := &models.User{
			BaseModel: models.BaseModel{ID: uuid.New()},
			TenantID:  &tenantID,
			Email:     "user@example.com",
		}

		userRepo.On("GetByEmailWithTenant", ctx, "user@example.com", tenantID).Return(user, nil).Once()

		resp, err := svc.GetUserByEmail(ctx, "user@example.com", &tenantID)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "user@example.com", resp.Email)
		userRepo.AssertExpectations(t)
	})
}

func TestUserService_VerifyEmail(t *testing.T) {
	repos, svc := setupUserServiceTest()
	userRepo := repos.User.(*MockUserRepository)
	ctx := context.Background()

	userID := uuid.New()

	t.Run("verify email successfully", func(t *testing.T) {
		userRepo.On("VerifyEmail", ctx, userID).Return(nil).Once()

		err := svc.VerifyEmail(ctx, userID)

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
	})
}

func TestUserService_EnableMFA(t *testing.T) {
	repos, svc := setupUserServiceTest()
	userRepo := repos.User.(*MockUserRepository)
	ctx := context.Background()

	userID := uuid.New()

	t.Run("enable MFA successfully", func(t *testing.T) {
		secret := "test-secret"
		userRepo.On("EnableMFA", ctx, userID, secret).Return(nil).Once()

		req := &dto.EnableMFARequest{
			Secret: secret,
			Code:   "123456",
		}

		err := svc.EnableMFA(ctx, userID, req)

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateRole(t *testing.T) {
	repos, svc := setupUserServiceTest()
	userRepo := repos.User.(*MockUserRepository)
	ctx := context.Background()

	t.Run("platform admin can update roles", func(t *testing.T) {
		userID := uuid.New()
		adminID := uuid.New()

		adminUser := &models.User{
			BaseModel:      models.BaseModel{ID: adminID},
			Role:           models.UserRolePlatformAdmin,
			IsPlatformUser: true,
		}

		userRepo.On("GetByID", ctx, adminID).Return(adminUser, nil).Once()
		userRepo.On("UpdateRole", ctx, userID, models.UserRoleArtisan).Return(nil).Once()

		err := svc.UpdateRole(ctx, userID, models.UserRoleArtisan, adminID)

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
	})

	t.Run("non-admin cannot update roles", func(t *testing.T) {
		userID := uuid.New()
		regularUserID := uuid.New()

		regularUser := &models.User{
			BaseModel: models.BaseModel{ID: regularUserID},
			Role:      models.UserRoleCustomer,
		}

		userRepo.On("GetByID", ctx, regularUserID).Return(regularUser, nil).Once()

		err := svc.UpdateRole(ctx, userID, models.UserRoleArtisan, regularUserID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "permission")
		userRepo.AssertExpectations(t)
	})
}
