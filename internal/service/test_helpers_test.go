package service_test

import (
	"context"
	"io"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/repository"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ============================================================================
// Mock User Repository
// ============================================================================

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByLogtoID(ctx context.Context, logtoID string) (*models.User, error) {
	args := m.Called(ctx, logtoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByZitadelID(ctx context.Context, zitadelID string) (*models.User, error) {
	args := m.Called(ctx, zitadelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) Search(ctx context.Context, query string, tenantID *uuid.UUID, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	args := m.Called(ctx, query, tenantID, pagination)
	if args.Get(0) == nil {
		return nil, repository.PaginationResult{}, args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(repository.PaginationResult), args.Error(2)
}

func (m *MockUserRepository) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) VerifyPhone(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) EnableMFA(ctx context.Context, userID uuid.UUID, secret string) error {
	args := m.Called(ctx, userID, secret)
	return args.Error(0)
}

func (m *MockUserRepository) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateRole(ctx context.Context, userID uuid.UUID, role models.UserRole) error {
	args := m.Called(ctx, userID, role)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status models.UserStatus) error {
	args := m.Called(ctx, userID, status)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmailWithTenant(ctx context.Context, email string, tenantID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, email, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByFilters(ctx context.Context, filters repository.UserFilters, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	args := m.Called(ctx, filters, pagination)
	if args.Get(0) == nil {
		return nil, repository.PaginationResult{}, args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(repository.PaginationResult), args.Error(2)
}

func (m *MockUserRepository) GetActiveUsers(ctx context.Context, tenantID *uuid.UUID, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	args := m.Called(ctx, tenantID, pagination)
	if args.Get(0) == nil {
		return nil, repository.PaginationResult{}, args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(repository.PaginationResult), args.Error(2)
}

func (m *MockUserRepository) GetRecentlyActive(ctx context.Context, tenantID *uuid.UUID, hours int, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	args := m.Called(ctx, tenantID, hours, pagination)
	if args.Get(0) == nil {
		return nil, repository.PaginationResult{}, args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(repository.PaginationResult), args.Error(2)
}

func (m *MockUserRepository) GetTenantUsersByRole(ctx context.Context, tenantID uuid.UUID, role models.UserRole, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	args := m.Called(ctx, tenantID, role, pagination)
	if args.Get(0) == nil {
		return nil, repository.PaginationResult{}, args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(repository.PaginationResult), args.Error(2)
}

func (m *MockUserRepository) GetByRole(ctx context.Context, role models.UserRole, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	args := m.Called(ctx, role, pagination)
	if args.Get(0) == nil {
		return nil, repository.PaginationResult{}, args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(repository.PaginationResult), args.Error(2)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	args := m.Called(ctx, userID, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, updates map[string]any) error {
	args := m.Called(ctx, userID, updates)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	args := m.Called(ctx, userID, avatarURL)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePreferences(ctx context.Context, userID uuid.UUID, timezone, language string) error {
	args := m.Called(ctx, userID, timezone, language)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateConsent(ctx context.Context, userID uuid.UUID, dataProcessing, marketing bool) error {
	args := m.Called(ctx, userID, dataProcessing, marketing)
	return args.Error(0)
}

func (m *MockUserRepository) AcceptTerms(ctx context.Context, userID uuid.UUID, version string) error {
	args := m.Called(ctx, userID, version)
	return args.Error(0)
}

func (m *MockUserRepository) AcceptPrivacyPolicy(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) MarkForDeletion(ctx context.Context, userID uuid.UUID, scheduledDate time.Time) error {
	args := m.Called(ctx, userID, scheduledDate)
	return args.Error(0)
}

func (m *MockUserRepository) GetUsersMarkedForDeletion(ctx context.Context) ([]*models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) PermanentlyDeleteUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) RecordLogin(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) RecordFailedLogin(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) UnlockUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) GetLockedUsers(ctx context.Context) ([]*models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserStats(ctx context.Context, tenantID *uuid.UUID) (repository.UserStats, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return repository.UserStats{}, args.Error(1)
	}
	return args.Get(0).(repository.UserStats), args.Error(1)
}

func (m *MockUserRepository) GetRegistrationStats(ctx context.Context, startDate, endDate time.Time) (repository.RegistrationStats, error) {
	args := m.Called(ctx, startDate, endDate)
	if args.Get(0) == nil {
		return repository.RegistrationStats{}, args.Error(1)
	}
	return args.Get(0).(repository.RegistrationStats), args.Error(1)
}

func (m *MockUserRepository) GetUserGrowth(ctx context.Context, tenantID *uuid.UUID, months int) ([]repository.UserGrowthData, error) {
	args := m.Called(ctx, tenantID, months)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.UserGrowthData), args.Error(1)
}

func (m *MockUserRepository) ActiveUsers(ctx context.Context, tenantID *uuid.UUID) (int64, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) GetTenantAdmins(ctx context.Context, tenantID uuid.UUID) ([]*models.User, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateSessionToken(ctx context.Context, userID uuid.UUID, token string, time time.Time) error {
	args := m.Called(ctx, userID, token, time)
	return args.Error(0)
}

// Stub methods for interface satisfaction
func (m *MockUserRepository) CreateBatch(ctx context.Context, users []*models.User) error { return nil }
func (m *MockUserRepository) GetByIDWithTenant(ctx context.Context, id uuid.UUID, tenantID *uuid.UUID) (*models.User, error) {
	return nil, nil
}
func (m *MockUserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockUserRepository) Restore(ctx context.Context, id uuid.UUID) error    { return nil }
func (m *MockUserRepository) Find(ctx context.Context, filters map[string]any) ([]*models.User, error) {
	return nil, nil
}
func (m *MockUserRepository) FindWithPagination(ctx context.Context, filters map[string]any, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	return nil, repository.PaginationResult{}, nil
}
func (m *MockUserRepository) Count(ctx context.Context, filters map[string]any) (int64, error) {
	return 0, nil
}
func (m *MockUserRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return false, nil
}
func (m *MockUserRepository) InvalidateCache(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockUserRepository) InvalidateCachePattern(ctx context.Context, pattern string) error {
	return nil
}
func (m *MockUserRepository) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return nil
}
func (m *MockUserRepository) GetDB() *gorm.DB { return nil }
func (m *MockUserRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error) {
	return nil, nil
}
func (m *MockUserRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	return nil, repository.PaginationResult{}, nil
}
func (m *MockUserRepository) FindByRole(ctx context.Context, tenantID uuid.UUID, role models.UserRole, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	return nil, repository.PaginationResult{}, nil
}
func (m *MockUserRepository) FindByStatus(ctx context.Context, tenantID uuid.UUID, status models.UserStatus, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	return nil, repository.PaginationResult{}, nil
}
func (m *MockUserRepository) ActivateUser(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *MockUserRepository) SuspendUser(ctx context.Context, userID uuid.UUID, reason string) error {
	return nil
}
func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *MockUserRepository) GetRoleDistribution(ctx context.Context, tenantID uuid.UUID) (map[models.UserRole]int64, error) {
	return nil, nil
}
func (m *MockUserRepository) BulkUpdateStatus(ctx context.Context, userIDs []uuid.UUID, status models.UserStatus) error {
	return nil
}
func (m *MockUserRepository) BulkUpdateRole(ctx context.Context, userIDs []uuid.UUID, role models.UserRole) error {
	return nil
}
func (m *MockUserRepository) FindInactiveUsers(ctx context.Context, tenantID uuid.UUID, days int) ([]*models.User, error) {
	return nil, nil
}
func (m *MockUserRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *MockUserRepository) CountByRole(ctx context.Context, tenantID uuid.UUID, role models.UserRole) (int64, error) {
	return 0, nil
}
func (m *MockUserRepository) DeactivateUser(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *MockUserRepository) GetUserStatsByTenant(ctx context.Context, tenantID uuid.UUID) (any, error) {
	return nil, nil
}
func (m *MockUserRepository) GetArtisans(ctx context.Context, tenantID uuid.UUID, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	return nil, repository.PaginationResult{}, nil
}
func (m *MockUserRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	return nil, repository.PaginationResult{}, nil
}
func (m *MockUserRepository) GetCustomers(ctx context.Context, tenantID uuid.UUID, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	return nil, repository.PaginationResult{}, nil
}
func (m *MockUserRepository) GetInactiveUsers(ctx context.Context, tenantID *uuid.UUID, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	return nil, repository.PaginationResult{}, nil
}
func (m *MockUserRepository) GetPlatformUsers(ctx context.Context, pagination repository.PaginationParams) ([]*models.User, repository.PaginationResult, error) {
	return nil, repository.PaginationResult{}, nil
}

// ============================================================================
// Mock Tenant Repository
// ============================================================================

type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) Update(ctx context.Context, tenant *models.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) FindBySubdomain(ctx context.Context, subdomain string) (*models.Tenant, error) {
	args := m.Called(ctx, subdomain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) FindByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	args := m.Called(ctx, domain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) ActivateTenant(ctx context.Context, tenantID uuid.UUID) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *MockTenantRepository) SuspendTenant(ctx context.Context, tenantID uuid.UUID, reason string) error {
	args := m.Called(ctx, tenantID, reason)
	return args.Error(0)
}

func (m *MockTenantRepository) CancelTenant(ctx context.Context, tenantID uuid.UUID, reason string) error {
	args := m.Called(ctx, tenantID, reason)
	return args.Error(0)
}

func (m *MockTenantRepository) IncrementUserCount(ctx context.Context, tenantID uuid.UUID) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *MockTenantRepository) DecrementUserCount(ctx context.Context, tenantID uuid.UUID) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *MockTenantRepository) UpdateStorageUsed(ctx context.Context, tenantID uuid.UUID, bytesUsed int64) error {
	args := m.Called(ctx, tenantID, bytesUsed)
	return args.Error(0)
}

func (m *MockTenantRepository) FindExpiredTrials(ctx context.Context) ([]*models.Tenant, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) UpdateSettings(ctx context.Context, tenantID uuid.UUID, settings models.JSONB) error {
	args := m.Called(ctx, tenantID, settings)
	return args.Error(0)
}

func (m *MockTenantRepository) UpdateFeatures(ctx context.Context, tenantID uuid.UUID, features models.TenantFeatures) error {
	args := m.Called(ctx, tenantID, features)
	return args.Error(0)
}

func (m *MockTenantRepository) Search(ctx context.Context, query string, pagination repository.PaginationParams) ([]*models.Tenant, repository.PaginationResult, error) {
	args := m.Called(ctx, query, pagination)
	if args.Get(0) == nil {
		return nil, repository.PaginationResult{}, args.Error(2)
	}
	return args.Get(0).([]*models.Tenant), args.Get(1).(repository.PaginationResult), args.Error(2)
}

func (m *MockTenantRepository) FindByFilters(ctx context.Context, filters repository.TenantFilters, pagination repository.PaginationParams) ([]*models.Tenant, repository.PaginationResult, error) {
	args := m.Called(ctx, filters, pagination)
	if args.Get(0) == nil {
		return nil, repository.PaginationResult{}, args.Error(2)
	}
	return args.Get(0).([]*models.Tenant), args.Get(1).(repository.PaginationResult), args.Error(2)
}

// Stub methods for interface satisfaction
func (m *MockTenantRepository) CreateBatch(ctx context.Context, tenants []*models.Tenant) error {
	return nil
}
func (m *MockTenantRepository) GetByIDWithTenant(ctx context.Context, id uuid.UUID, tenantID *uuid.UUID) (*models.Tenant, error) {
	return nil, nil
}
func (m *MockTenantRepository) Delete(ctx context.Context, id uuid.UUID) error     { return nil }
func (m *MockTenantRepository) SoftDelete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockTenantRepository) Restore(ctx context.Context, id uuid.UUID) error    { return nil }
func (m *MockTenantRepository) Find(ctx context.Context, filters map[string]any) ([]*models.Tenant, error) {
	return nil, nil
}
func (m *MockTenantRepository) FindWithPagination(ctx context.Context, filters map[string]any, pagination repository.PaginationParams) ([]*models.Tenant, repository.PaginationResult, error) {
	return nil, repository.PaginationResult{}, nil
}
func (m *MockTenantRepository) Count(ctx context.Context, filters map[string]any) (int64, error) {
	return 0, nil
}
func (m *MockTenantRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return false, nil
}
func (m *MockTenantRepository) InvalidateCache(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockTenantRepository) InvalidateCachePattern(ctx context.Context, pattern string) error {
	return nil
}
func (m *MockTenantRepository) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return nil
}
func (m *MockTenantRepository) GetDB() *gorm.DB { return nil }
func (m *MockTenantRepository) FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*models.Tenant, error) {
	return nil, nil
}
func (m *MockTenantRepository) FindActiveBySubdomain(ctx context.Context, subdomain string) (*models.Tenant, error) {
	return nil, nil
}
func (m *MockTenantRepository) UpdatePlan(ctx context.Context, tenantID uuid.UUID, plan models.TenantPlan) error {
	return nil
}
func (m *MockTenantRepository) IncrementArtisanCount(ctx context.Context, tenantID uuid.UUID) error {
	return nil
}
func (m *MockTenantRepository) DecrementArtisanCount(ctx context.Context, tenantID uuid.UUID) error {
	return nil
}
func (m *MockTenantRepository) GetTenantStats(ctx context.Context, tenantID uuid.UUID) (any, error) {
	return nil, nil
}
func (m *MockTenantRepository) FindExpiringSoon(ctx context.Context, days int) ([]*models.Tenant, error) {
	return nil, nil
}
func (m *MockTenantRepository) BulkUpdateStatus(ctx context.Context, tenantIDs []uuid.UUID, status models.TenantStatus) error {
	return nil
}
func (m *MockTenantRepository) BulkSuspend(ctx context.Context, tenantIDs []uuid.UUID, reason string) error {
	return nil
}
func (m *MockTenantRepository) DeleteTenant(ctx context.Context, tenantID uuid.UUID) error {
	return nil
}
func (m *MockTenantRepository) CheckStorageLimit(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	return true, nil
}
func (m *MockTenantRepository) CheckUserLimit(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	return true, nil
}
func (m *MockTenantRepository) AddAdmin(ctx context.Context, tenantID, userID uuid.UUID) error {
	return nil
}
func (m *MockTenantRepository) RemoveAdmin(ctx context.Context, tenantID, userID uuid.UUID) error {
	return nil
}
func (m *MockTenantRepository) GetAdmins(ctx context.Context, tenantID uuid.UUID) ([]*models.User, error) {
	return nil, nil
}
func (m *MockTenantRepository) ConvertTrialToActive(ctx context.Context, tenantID uuid.UUID, plan models.TenantPlan) error {
	return nil
}
func (m *MockTenantRepository) DowngradePlan(ctx context.Context, tenantID uuid.UUID, plan models.TenantPlan) error {
	return nil
}
func (m *MockTenantRepository) FindActiveTenants(ctx context.Context, pagination repository.PaginationParams) ([]*models.Tenant, repository.PaginationResult, error) {
	return nil, repository.PaginationResult{}, nil
}
func (m *MockTenantRepository) GetArtisans(ctx context.Context, tenantID uuid.UUID) ([]*models.User, error) {
	return nil, nil
}
func (m *MockTenantRepository) UpgradePlan(ctx context.Context, tenantID uuid.UUID, plan models.TenantPlan) error {
	return nil
}
func (m *MockTenantRepository) DeleteInactiveTenants(ctx context.Context, days int) error { return nil }
func (m *MockTenantRepository) DisableFeature(ctx context.Context, tenantID uuid.UUID, feature string) error {
	return nil
}
func (m *MockTenantRepository) EnableFeature(ctx context.Context, tenantID uuid.UUID, feature string) error {
	return nil
}
func (m *MockTenantRepository) FindByPlan(ctx context.Context, plan models.TenantPlan, pagination repository.PaginationParams) ([]*models.Tenant, repository.PaginationResult, error) {
	return nil, repository.PaginationResult{}, nil
}
func (m *MockTenantRepository) FindByStatus(ctx context.Context, status models.TenantStatus, pagination repository.PaginationParams) ([]*models.Tenant, repository.PaginationResult, error) {
	return nil, repository.PaginationResult{}, nil
}

// ============================================================================
// Mock Logger
// ============================================================================

type MockLogger struct{}

func (m *MockLogger) Trace(v ...any)                                   {}
func (m *MockLogger) Debug(v ...any)                                   {}
func (m *MockLogger) Info(v ...any)                                    {}
func (m *MockLogger) Warn(v ...any)                                    {}
func (m *MockLogger) Error(v ...any)                                   {}
func (m *MockLogger) Fatal(v ...any)                                   {}
func (m *MockLogger) Panic(v ...any)                                   {}
func (m *MockLogger) Tracef(format string, v ...any)                   {}
func (m *MockLogger) Debugf(format string, v ...any)                   {}
func (m *MockLogger) Infof(format string, v ...any)                    {}
func (m *MockLogger) Warnf(format string, v ...any)                    {}
func (m *MockLogger) Errorf(format string, v ...any)                   {}
func (m *MockLogger) Fatalf(format string, v ...any)                   {}
func (m *MockLogger) Panicf(format string, v ...any)                   {}
func (m *MockLogger) Tracew(msg string, keysAndValues ...any)          {}
func (m *MockLogger) Debugw(msg string, keysAndValues ...any)          {}
func (m *MockLogger) Infow(msg string, keysAndValues ...any)           {}
func (m *MockLogger) Warnw(msg string, keysAndValues ...any)           {}
func (m *MockLogger) Errorw(msg string, keysAndValues ...any)          {}
func (m *MockLogger) Fatalw(msg string, keysAndValues ...any)          {}
func (m *MockLogger) Panicw(msg string, keysAndValues ...any)          {}
func (m *MockLogger) SetLevel(level log.Level)                         {}
func (m *MockLogger) SetOutput(writer io.Writer)                       {}
func (m *MockLogger) WithContext(ctx context.Context) log.CommonLogger { return m }
