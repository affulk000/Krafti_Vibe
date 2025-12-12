package repository_test

import (
	"context"
	"testing"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/repository/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantRepository_Create(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewTenantRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	t.Run("create tenant successfully", func(t *testing.T) {
		// Create owner user first
		owner := testutil.CreateTestOwner()
		require.NoError(t, tdb.DB.Create(owner).Error)

		tenant := testutil.CreateTestTenant(func(t *models.Tenant) {
			t.OwnerID = owner.ID
		})
		err := repo.Create(ctx, tenant)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, tenant.ID)
	})

	t.Run("create tenant with duplicate slug should fail", func(t *testing.T) {
		// Create owner user first
		owner := testutil.CreateTestOwner(func(u *models.User) {
			u.Email = "owner2@example.com"
		})
		require.NoError(t, tdb.DB.Create(owner).Error)

		tenant1 := testutil.CreateTestTenant(func(t *models.Tenant) {
			t.OwnerID = owner.ID
			t.Subdomain = "unique-slug"
		})
		err := repo.Create(ctx, tenant1)
		require.NoError(t, err)

		tenant2 := testutil.CreateTestTenant(func(t *models.Tenant) {
			t.OwnerID = owner.ID
			t.Subdomain = "unique-slug"
		})
		err = repo.Create(ctx, tenant2)
		assert.Error(t, err)
	})
}

func TestTenantRepository_FindBySubdomain(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewTenantRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	// Create owner user first
	owner := testutil.CreateTestOwner()
	require.NoError(t, tdb.DB.Create(owner).Error)

	tenant := testutil.CreateTestTenant(func(t *models.Tenant) {
		t.OwnerID = owner.ID
		t.Subdomain = "test-slug"
	})
	require.NoError(t, repo.Create(ctx, tenant))

	t.Run("get tenant by slug", func(t *testing.T) {
		found, err := repo.FindBySubdomain(ctx, "test-slug")
		require.NoError(t, err)
		assert.Equal(t, tenant.ID, found.ID)
		assert.Equal(t, "test-slug", found.Subdomain)
	})

	t.Run("get tenant by non-existent slug", func(t *testing.T) {
		_, err := repo.FindBySubdomain(ctx, "non-existent")
		assert.Error(t, err)
	})
}

func TestTenantRepository_FindByDomain(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewTenantRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	// Create owner user first
	owner := testutil.CreateTestOwner()
	require.NoError(t, tdb.DB.Create(owner).Error)

	tenant := testutil.CreateTestTenant(func(t *models.Tenant) {
		t.OwnerID = owner.ID
		t.Domain = "custom.example.com"
	})
	require.NoError(t, repo.Create(ctx, tenant))

	t.Run("get tenant by domain", func(t *testing.T) {
		found, err := repo.FindByDomain(ctx, "custom.example.com")
		require.NoError(t, err)
		assert.Equal(t, tenant.ID, found.ID)
	})
}

func TestTenantRepository_UpdateStatus(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewTenantRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	// Create owner user first
	owner := testutil.CreateTestOwner()
	require.NoError(t, tdb.DB.Create(owner).Error)

	tenant := testutil.CreateTestTenant(func(t *models.Tenant) {
		t.OwnerID = owner.ID
	})
	require.NoError(t, repo.Create(ctx, tenant))

	t.Run("suspend tenant", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, tenant.ID, models.TenantStatusSuspended)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, tenant.ID)
		require.NoError(t, err)
		assert.Equal(t, models.TenantStatusSuspended, updated.Status)
	})
}

func TestTenantRepository_GetActiveTenantsCount(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewTenantRepository(tdb.DB, testutil.DefaultRepositoryConfig())
	ctx := context.Background()

	// Create owner user first
	owner := testutil.CreateTestOwner()
	require.NoError(t, tdb.DB.Create(owner).Error)

	// Create active tenants
	for i := 0; i < 3; i++ {
		tenant := testutil.CreateTestTenant(func(t *models.Tenant) {
			t.OwnerID = owner.ID
			t.Subdomain = uuid.New().String()
			t.Status = models.TenantStatusActive
		})
		require.NoError(t, repo.Create(ctx, tenant))
	}

	// Create inactive tenant
	inactiveTenant := testutil.CreateTestTenant(func(t *models.Tenant) {
		t.OwnerID = owner.ID
		t.Subdomain = "inactive-tenant"
		t.Status = models.TenantStatusSuspended
	})
	require.NoError(t, repo.Create(ctx, inactiveTenant))

	t.Run("count active tenants", func(t *testing.T) {
		count, err := repo.Count(ctx, map[string]any{"status": models.TenantStatusActive})
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})
}
