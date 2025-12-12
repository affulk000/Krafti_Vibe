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

func setupServiceTest(t *testing.T) (*testutil.TestDB, repository.ServiceRepository, uuid.UUID, uuid.UUID) {
	tdb := testutil.NewTestDB(t)
	repo := repository.NewServiceRepository(tdb.DB, testutil.DefaultRepositoryConfig())

	// Create tenant with owner
	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	artisanUser := testutil.CreateTestUser(&tenant.ID, func(u *models.User) {
		u.Email = "artisan@example.com"
		u.Role = models.UserRoleArtisan
	})
	require.NoError(t, tdb.DB.Create(artisanUser).Error)

	artisan := testutil.CreateTestArtisan(artisanUser.ID, tenant.ID)
	require.NoError(t, tdb.DB.Create(artisan).Error)

	return tdb, repo, tenant.ID, artisan.ID
}

func TestServiceRepository_Create(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	t.Run("create service successfully", func(t *testing.T) {
		service := testutil.CreateTestService(tenantID, artisanID)
		err := repo.Create(ctx, service)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, service.ID)
	})
}

func TestServiceRepository_GetByID(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	service := testutil.CreateTestService(tenantID, artisanID)
	require.NoError(t, repo.Create(ctx, service))

	t.Run("get service by ID", func(t *testing.T) {
		found, err := repo.GetByID(ctx, service.ID)
		require.NoError(t, err)
		assert.Equal(t, service.ID, found.ID)
		assert.Equal(t, service.Name, found.Name)
	})
}

func TestServiceRepository_Update(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	service := testutil.CreateTestService(tenantID, artisanID)
	require.NoError(t, repo.Create(ctx, service))

	t.Run("update service details", func(t *testing.T) {
		service.Name = "Updated Service Name"
		service.Price = 750.00
		err := repo.Update(ctx, service)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, service.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Service Name", updated.Name)
		assert.Equal(t, 750.00, updated.Price)
	})
}

func TestServiceRepository_Delete(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	service := testutil.CreateTestService(tenantID, artisanID)
	require.NoError(t, repo.Create(ctx, service))

	t.Run("delete service", func(t *testing.T) {
		err := repo.Delete(ctx, service.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, service.ID)
		assert.Error(t, err)
	})
}

func TestServiceRepository_GetByCategory(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create services with different categories
	carpentryService := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
		s.Category = models.ServiceCategoryCarpentry
	})
	require.NoError(t, repo.Create(ctx, carpentryService))

	plumbingService := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
		s.Category = models.ServiceCategoryPlumbing
	})
	require.NoError(t, repo.Create(ctx, plumbingService))

	t.Run("verify services created with categories", func(t *testing.T) {
		// Verify carpentry service
		carp, err := repo.GetByID(ctx, carpentryService.ID)
		require.NoError(t, err)
		assert.Equal(t, models.ServiceCategoryCarpentry, carp.Category)

		// Verify plumbing service
		plumb, err := repo.GetByID(ctx, plumbingService.ID)
		require.NoError(t, err)
		assert.Equal(t, models.ServiceCategoryPlumbing, plumb.Category)
	})
}

// Query Operations Tests

func TestServiceRepository_FindByTenantID(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create multiple services
	for i := 0; i < 3; i++ {
		service := testutil.CreateTestService(tenantID, artisanID)
		require.NoError(t, repo.Create(ctx, service))
	}

	t.Run("find services by tenant ID", func(t *testing.T) {
		services, pagination, err := repo.FindByTenantID(ctx, tenantID, repository.PaginationParams{
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, services, 3)
		assert.Equal(t, int64(3), pagination.TotalItems)
	})
}

func TestServiceRepository_FindByCategory(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create services with different categories
	for i := 0; i < 2; i++ {
		service := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
			s.Category = models.ServiceCategoryCarpentry
		})
		require.NoError(t, repo.Create(ctx, service))
	}

	service := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
		s.Category = models.ServiceCategoryPlumbing
	})
	require.NoError(t, repo.Create(ctx, service))

	t.Run("find services by category", func(t *testing.T) {
		services, pagination, err := repo.FindByCategory(ctx, tenantID, models.ServiceCategoryCarpentry, repository.PaginationParams{
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, services, 2)
		assert.Equal(t, int64(2), pagination.TotalItems)
	})
}

func TestServiceRepository_FindByArtisanID(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create services for artisan
	for i := 0; i < 2; i++ {
		service := testutil.CreateTestService(tenantID, artisanID)
		require.NoError(t, repo.Create(ctx, service))
	}

	t.Run("find services by artisan ID", func(t *testing.T) {
		services, err := repo.FindByArtisanID(ctx, artisanID)
		require.NoError(t, err)
		assert.Len(t, services, 2)
	})
}

func TestServiceRepository_FindActiveServices(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create services (all default to active)
	inactiveService1 := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
		s.Name = "Inactive Service 1"
	})
	require.NoError(t, repo.Create(ctx, inactiveService1))
	// Deactivate after creation
	require.NoError(t, repo.DeactivateService(ctx, inactiveService1.ID))

	inactiveService2 := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
		s.Name = "Inactive Service 2"
	})
	require.NoError(t, repo.Create(ctx, inactiveService2))
	// Deactivate after creation
	require.NoError(t, repo.DeactivateService(ctx, inactiveService2.ID))

	// Create one active service
	activeService := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
		s.Name = "Active Service"
	})
	require.NoError(t, repo.Create(ctx, activeService))

	t.Run("find only active services", func(t *testing.T) {
		services, pagination, err := repo.FindActiveServices(ctx, tenantID, repository.PaginationParams{
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, services, 1)
		assert.True(t, services[0].IsActive)
		assert.Equal(t, "Active Service", services[0].Name)
		assert.Equal(t, int64(1), pagination.TotalItems)
	})
}

func TestServiceRepository_FindByPriceRange(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create services with different prices
	cheapService := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
		s.Price = 100.00
	})
	require.NoError(t, repo.Create(ctx, cheapService))

	midService := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
		s.Price = 500.00
	})
	require.NoError(t, repo.Create(ctx, midService))

	expensiveService := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
		s.Price = 1000.00
	})
	require.NoError(t, repo.Create(ctx, expensiveService))

	t.Run("find services in price range", func(t *testing.T) {
		services, pagination, err := repo.FindByPriceRange(ctx, tenantID, 200.00, 800.00, repository.PaginationParams{
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, services, 1)
		assert.Equal(t, 500.00, services[0].Price)
		assert.Equal(t, int64(1), pagination.TotalItems)
	})
}

// Availability Management Tests

func TestServiceRepository_ActivateDeactivateService(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	service := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
		s.IsActive = false
	})
	require.NoError(t, repo.Create(ctx, service))

	t.Run("activate service", func(t *testing.T) {
		err := repo.ActivateService(ctx, service.ID)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, service.ID)
		require.NoError(t, err)
		assert.True(t, updated.IsActive)
	})

	t.Run("deactivate service", func(t *testing.T) {
		err := repo.DeactivateService(ctx, service.ID)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, service.ID)
		require.NoError(t, err)
		assert.False(t, updated.IsActive)
	})
}

func TestServiceRepository_UpdateAvailability(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	service := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
		s.IsActive = false
	})
	require.NoError(t, repo.Create(ctx, service))

	t.Run("update availability to true", func(t *testing.T) {
		err := repo.UpdateAvailability(ctx, service.ID, true)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, service.ID)
		require.NoError(t, err)
		assert.True(t, updated.IsActive)
	})
}

// Pricing Management Tests

func TestServiceRepository_UpdatePrice(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	service := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
		s.Price = 500.00
	})
	require.NoError(t, repo.Create(ctx, service))

	t.Run("update service price", func(t *testing.T) {
		err := repo.UpdatePrice(ctx, service.ID, 750.00)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, service.ID)
		require.NoError(t, err)
		assert.Equal(t, 750.00, updated.Price)
	})
}

func TestServiceRepository_UpdateDeposit(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	service := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
		s.RequiresDeposit = true
		s.DepositAmount = 100.00
	})
	require.NoError(t, repo.Create(ctx, service))

	t.Run("update deposit amount", func(t *testing.T) {
		err := repo.UpdateDeposit(ctx, service.ID, 150.00)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, service.ID)
		require.NoError(t, err)
		assert.Equal(t, 150.00, updated.DepositAmount)
	})
}

func TestServiceRepository_BulkUpdatePrices(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create multiple services
	var serviceIDs []uuid.UUID
	for i := 0; i < 3; i++ {
		service := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
			s.Price = 100.00
		})
		require.NoError(t, repo.Create(ctx, service))
		serviceIDs = append(serviceIDs, service.ID)
	}

	t.Run("bulk update prices with percentage", func(t *testing.T) {
		err := repo.BulkUpdatePrices(ctx, serviceIDs, 10.0, true)
		require.NoError(t, err)

		// Verify price increased by 10%
		for _, id := range serviceIDs {
			service, err := repo.GetByID(ctx, id)
			require.NoError(t, err)
			assert.Equal(t, 110.00, service.Price)
		}
	})
}

// Bulk Operations Tests

func TestServiceRepository_BulkActivate(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create inactive services
	var serviceIDs []uuid.UUID
	for i := 0; i < 2; i++ {
		service := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
			s.IsActive = false
		})
		require.NoError(t, repo.Create(ctx, service))
		serviceIDs = append(serviceIDs, service.ID)
	}

	t.Run("bulk activate services", func(t *testing.T) {
		err := repo.BulkActivate(ctx, serviceIDs)
		require.NoError(t, err)

		// Verify all are active
		for _, id := range serviceIDs {
			service, err := repo.GetByID(ctx, id)
			require.NoError(t, err)
			assert.True(t, service.IsActive)
		}
	})
}

func TestServiceRepository_BulkDeactivate(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create active services
	var serviceIDs []uuid.UUID
	for i := 0; i < 2; i++ {
		service := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
			s.IsActive = true
		})
		require.NoError(t, repo.Create(ctx, service))
		serviceIDs = append(serviceIDs, service.ID)
	}

	t.Run("bulk deactivate services", func(t *testing.T) {
		err := repo.BulkDeactivate(ctx, serviceIDs)
		require.NoError(t, err)

		// Verify all are inactive
		for _, id := range serviceIDs {
			service, err := repo.GetByID(ctx, id)
			require.NoError(t, err)
			assert.False(t, service.IsActive)
		}
	})
}

func TestServiceRepository_BulkUpdateCategory(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create services with different categories
	var serviceIDs []uuid.UUID
	for i := 0; i < 2; i++ {
		service := testutil.CreateTestService(tenantID, artisanID, func(s *models.Service) {
			s.Category = models.ServiceCategoryCarpentry
		})
		require.NoError(t, repo.Create(ctx, service))
		serviceIDs = append(serviceIDs, service.ID)
	}

	t.Run("bulk update category", func(t *testing.T) {
		err := repo.BulkUpdateCategory(ctx, serviceIDs, models.ServiceCategoryPlumbing)
		require.NoError(t, err)

		// Verify all have new category
		for _, id := range serviceIDs {
			service, err := repo.GetByID(ctx, id)
			require.NoError(t, err)
			assert.Equal(t, models.ServiceCategoryPlumbing, service.Category)
		}
	})
}

func TestServiceRepository_BulkDelete(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupServiceTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create services
	var serviceIDs []uuid.UUID
	for i := 0; i < 2; i++ {
		service := testutil.CreateTestService(tenantID, artisanID)
		require.NoError(t, repo.Create(ctx, service))
		serviceIDs = append(serviceIDs, service.ID)
	}

	t.Run("bulk delete services", func(t *testing.T) {
		err := repo.BulkDelete(ctx, serviceIDs)
		require.NoError(t, err)

		// Verify all are deleted
		for _, id := range serviceIDs {
			_, err := repo.GetByID(ctx, id)
			assert.Error(t, err)
		}
	})
}
