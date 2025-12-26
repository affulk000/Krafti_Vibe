package repository_test

import (
	"context"
	"testing"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/repository/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	// Create test tenant with owner
	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	t.Run("create user successfully", func(t *testing.T) {
		user := testutil.CreateTestUser(&tenant.ID, func(u *models.User) {
			u.Email = "new@example.com"
		})

		err := repo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.NotZero(t, user.CreatedAt)
	})

	t.Run("create user with duplicate email should fail", func(t *testing.T) {
		user1 := testutil.CreateTestUser(&tenant.ID, func(u *models.User) {
			u.Email = "duplicate@example.com"
		})
		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		user2 := testutil.CreateTestUser(&tenant.ID, func(u *models.User) {
			u.Email = "duplicate@example.com"
		})
		err = repo.Create(ctx, user2)
		assert.Error(t, err)
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	// Setup
	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	user := testutil.CreateTestUser(&tenant.ID)
	require.NoError(t, repo.Create(ctx, user))

	t.Run("get existing user", func(t *testing.T) {
		found, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("get non-existent user", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	user := testutil.CreateTestUser(&tenant.ID, func(u *models.User) {
		u.Email = "find@example.com"
	})
	require.NoError(t, repo.Create(ctx, user))

	t.Run("get user by email", func(t *testing.T) {
		found, err := repo.GetByEmail(ctx, "find@example.com")
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, "find@example.com", found.Email)
	})

	t.Run("get user by non-existent email", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
	})
}

func TestUserRepository_GetByZitadelID(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	zitadelID := uuid.New().String()
	user := testutil.CreateTestUser(&tenant.ID, func(u *models.User) {
		u.ZitadelUserID = zitadelID
	})
	require.NoError(t, repo.Create(ctx, user))

	t.Run("get user by zitadel ID", func(t *testing.T) {
		found, err := repo.GetByZitadelID(ctx, zitadelID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, zitadelID, found.ZitadelUserID)
	})
}

func TestUserRepository_Update(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	user := testutil.CreateTestUser(&tenant.ID)
	require.NoError(t, repo.Create(ctx, user))

	t.Run("update user successfully", func(t *testing.T) {
		user.FirstName = "Jane"
		user.LastName = "Smith"
		err := repo.Update(ctx, user)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "Jane", updated.FirstName)
		assert.Equal(t, "Smith", updated.LastName)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	user := testutil.CreateTestUser(&tenant.ID)
	require.NoError(t, repo.Create(ctx, user))

	t.Run("soft delete user", func(t *testing.T) {
		err := repo.SoftDelete(ctx, user.ID)
		require.NoError(t, err)

		// Should not be found in regular queries
		_, err = repo.GetByID(ctx, user.ID)
		assert.Error(t, err)
	})
}

func TestUserRepository_GetByTenantID(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	// Create tenant1 with unique owner email
	owner1 := testutil.CreateTestOwner(func(u *models.User) {
		u.Email = "owner1@example.com"
	})
	require.NoError(t, tdb.DB.Create(owner1).Error)
	tenant1 := testutil.CreateTestTenant(func(t *models.Tenant) {
		t.OwnerID = owner1.ID
		t.Subdomain = "tenant1"
	})
	require.NoError(t, tdb.DB.Create(tenant1).Error)

	// Create tenant2 with unique owner email
	owner2 := testutil.CreateTestOwner(func(u *models.User) {
		u.Email = "owner2@example.com"
	})
	require.NoError(t, tdb.DB.Create(owner2).Error)
	tenant2 := testutil.CreateTestTenant(func(t *models.Tenant) {
		t.OwnerID = owner2.ID
		t.Subdomain = "tenant2"
	})
	require.NoError(t, tdb.DB.Create(tenant2).Error)

	// Create users for tenant1
	for i := 0; i < 3; i++ {
		user := testutil.CreateTestUser(&tenant1.ID, func(u *models.User) {
			u.Email = uuid.New().String() + "@tenant1.com"
		})
		require.NoError(t, repo.Create(ctx, user))
	}

	// Create users for tenant2
	for i := 0; i < 2; i++ {
		user := testutil.CreateTestUser(&tenant2.ID, func(u *models.User) {
			u.Email = uuid.New().String() + "@tenant2.com"
		})
		require.NoError(t, repo.Create(ctx, user))
	}

	t.Run("get users by tenant", func(t *testing.T) {
		users, pagination, err := repo.GetByTenantID(ctx, tenant1.ID, repository.PaginationParams{
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, users, 3)
		assert.Equal(t, int64(3), pagination.TotalItems)
	})
}

func TestUserRepository_UpdateRole(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	user := testutil.CreateTestUser(&tenant.ID, func(u *models.User) {
		u.Role = models.UserRoleCustomer
	})
	require.NoError(t, repo.Create(ctx, user))

	t.Run("update user role", func(t *testing.T) {
		err := repo.UpdateRole(ctx, user.ID, models.UserRoleArtisan)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, models.UserRoleArtisan, updated.Role)
	})
}

func TestUserRepository_UpdateStatus(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	user := testutil.CreateTestUser(&tenant.ID)
	require.NoError(t, repo.Create(ctx, user))

	t.Run("deactivate user", func(t *testing.T) {
		err := repo.DeactivateUser(ctx, user.ID)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, models.UserStatusInactive, updated.Status)
	})

	t.Run("activate user", func(t *testing.T) {
		err := repo.ActivateUser(ctx, user.ID)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, models.UserStatusActive, updated.Status)
	})

	t.Run("suspend user", func(t *testing.T) {
		err := repo.SuspendUser(ctx, user.ID, "Policy violation")
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, models.UserStatusSuspended, updated.Status)
	})
}

func TestUserRepository_RecordLogin(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	user := testutil.CreateTestUser(&tenant.ID)
	require.NoError(t, repo.Create(ctx, user))

	t.Run("record successful login", func(t *testing.T) {
		err := repo.RecordLogin(ctx, user.ID)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.NotNil(t, updated.LastLoginAt)
		assert.Equal(t, 0, updated.FailedLoginAttempts)
	})

	t.Run("record failed login", func(t *testing.T) {
		err := repo.RecordFailedLogin(ctx, user.ID)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, updated.FailedLoginAttempts)
	})
}

func TestUserRepository_VerifyEmail(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	user := testutil.CreateTestUser(&tenant.ID, func(u *models.User) {
		u.EmailVerified = false
	})
	require.NoError(t, repo.Create(ctx, user))

	t.Run("verify email", func(t *testing.T) {
		err := repo.VerifyEmail(ctx, user.ID)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.True(t, updated.EmailVerified)
		assert.NotNil(t, updated.EmailVerified)
	})
}

func TestUserRepository_VerifyPhone(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	user := testutil.CreateTestUser(&tenant.ID, func(u *models.User) {
		u.PhoneVerified = false
	})
	require.NoError(t, repo.Create(ctx, user))

	t.Run("verify phone", func(t *testing.T) {
		err := repo.VerifyPhone(ctx, user.ID)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.True(t, updated.PhoneVerified)
		assert.NotNil(t, updated.PhoneVerified)
	})
}

func TestUserRepository_Search(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	// Create test users
	users := []*models.User{
		testutil.CreateTestUser(&tenant.ID, func(u *models.User) {
			u.FirstName = "Alice"
			u.LastName = "Anderson"
			u.Email = "alice@example.com"
		}),
		testutil.CreateTestUser(&tenant.ID, func(u *models.User) {
			u.FirstName = "Bob"
			u.LastName = "Brown"
			u.Email = "bob@example.com"
		}),
		testutil.CreateTestUser(&tenant.ID, func(u *models.User) {
			u.FirstName = "Alice"
			u.LastName = "Brown"
			u.Email = "alice.brown@example.com"
		}),
	}

	for _, user := range users {
		require.NoError(t, repo.Create(ctx, user))
	}

	t.Run("search by first name", func(t *testing.T) {
		results, _, err := repo.Search(ctx, "Alice", &tenant.ID, repository.PaginationParams{
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("search by email", func(t *testing.T) {
		results, _, err := repo.Search(ctx, "bob@example.com", &tenant.ID, repository.PaginationParams{
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Bob", results[0].FirstName)
	})
}

func TestUserRepository_AcceptTerms(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	user := testutil.CreateTestUser(&tenant.ID)
	require.NoError(t, repo.Create(ctx, user))

	t.Run("accept terms of service", func(t *testing.T) {
		err := repo.AcceptTerms(ctx, user.ID, "v1.0")
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.NotNil(t, updated.TermsAcceptedAt)
		assert.Equal(t, "v1.0", updated.TermsVersion)
	})
}

func TestUserRepository_MarkForDeletion(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewUserRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	user := testutil.CreateTestUser(&tenant.ID)
	require.NoError(t, repo.Create(ctx, user))

	t.Run("mark user for deletion", func(t *testing.T) {
		scheduledDate := time.Now().UTC().Add(30 * 24 * time.Hour)
		err := repo.MarkForDeletion(ctx, user.ID, scheduledDate)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.True(t, updated.MarkedForDeletion)
		assert.NotNil(t, updated.DeletionScheduledAt)
	})

	t.Run("get users marked for deletion", func(t *testing.T) {
		users, err := repo.GetUsersMarkedForDeletion(ctx)
		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, user.ID, users[0].ID)
	})
}
