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

func TestWhiteLabelRepository_Create(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewWhiteLabelRepository(tdb.DB)
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	t.Run("create whitelabel successfully", func(t *testing.T) {
		whiteLabel := testutil.CreateTestWhiteLabel(tenant.ID)
		err := repo.Create(ctx, whiteLabel)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, whiteLabel.ID)
	})
}

func TestWhiteLabelRepository_GetByTenantID(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewWhiteLabelRepository(tdb.DB)
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	whiteLabel := testutil.CreateTestWhiteLabel(tenant.ID)
	require.NoError(t, repo.Create(ctx, whiteLabel))

	t.Run("get whitelabel by tenant", func(t *testing.T) {
		found, err := repo.GetByTenantID(ctx, tenant.ID)
		require.NoError(t, err)
		assert.Equal(t, whiteLabel.ID, found.ID)
		assert.Equal(t, tenant.ID, found.TenantID)
	})
}

func TestWhiteLabelRepository_GetByCustomDomain(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewWhiteLabelRepository(tdb.DB)
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	whiteLabel := testutil.CreateTestWhiteLabel(tenant.ID, func(w *models.WhiteLabel) {
		w.CustomDomain = "custom.example.com"
		w.CustomDomainEnabled = true
	})
	require.NoError(t, repo.Create(ctx, whiteLabel))

	t.Run("get whitelabel by domain", func(t *testing.T) {
		found, err := repo.GetByCustomDomain(ctx, "custom.example.com")
		require.NoError(t, err)
		assert.Equal(t, whiteLabel.ID, found.ID)
		assert.Equal(t, "custom.example.com", found.CustomDomain)
	})
}

func TestWhiteLabelRepository_Update(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Close()

	repo := repository.NewWhiteLabelRepository(tdb.DB)
	ctx := context.Background()

	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	whiteLabel := testutil.CreateTestWhiteLabel(tenant.ID)
	require.NoError(t, repo.Create(ctx, whiteLabel))

	t.Run("update whitelabel", func(t *testing.T) {
		whiteLabel.CompanyName = "Updated Brand"
		whiteLabel.PrimaryColor = "#ff0000"
		err := repo.Update(ctx, whiteLabel)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, whiteLabel.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Brand", updated.CompanyName)
		assert.Equal(t, "#ff0000", updated.PrimaryColor)
	})
}
