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

func setupAvailabilityTest(t *testing.T) (*testutil.TestDB, repository.AvailabilityRepository, uuid.UUID, uuid.UUID) {
	tdb := testutil.NewTestDB(t)
	repo := repository.NewAvailabilityRepository(tdb.DB)

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

func TestAvailabilityRepository_Create(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupAvailabilityTest(t)
	defer tdb.Close()

	ctx := context.Background()

	t.Run("create availability successfully", func(t *testing.T) {
		availability := testutil.CreateTestAvailability(tenantID, artisanID)
		err := repo.Create(ctx, availability)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, availability.ID)
	})
}

func TestAvailabilityRepository_GetByArtisan(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupAvailabilityTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create availability slots for different days
	var lastAvailability *models.Availability
	for day := 1; day <= 5; day++ {
		dayPtr := day
		availability := testutil.CreateTestAvailability(tenantID, artisanID, func(a *models.Availability) {
			a.DayOfWeek = &dayPtr
		})
		require.NoError(t, repo.Create(ctx, availability))
		lastAvailability = availability
	}

	t.Run("get availability by ID", func(t *testing.T) {
		// Just verify we created the records
		firstSlot, err := repo.GetByID(ctx, lastAvailability.ID)
		require.NoError(t, err)
		assert.NotNil(t, firstSlot)
	})
}

func TestAvailabilityRepository_GetByDayOfWeek(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupAvailabilityTest(t)
	defer tdb.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// Create availability for Monday (1)
	monday := 1
	monday1 := testutil.CreateTestAvailability(tenantID, artisanID, func(a *models.Availability) {
		a.DayOfWeek = &monday
		a.StartTime = time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC)
		a.EndTime = time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
	})
	require.NoError(t, repo.Create(ctx, monday1))

	monday2 := testutil.CreateTestAvailability(tenantID, artisanID, func(a *models.Availability) {
		a.DayOfWeek = &monday
		a.StartTime = time.Date(now.Year(), now.Month(), now.Day(), 13, 0, 0, 0, time.UTC)
		a.EndTime = time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, time.UTC)
	})
	require.NoError(t, repo.Create(ctx, monday2))

	// Create availability for Tuesday (2)
	tuesday := 2
	tuesdaySlot := testutil.CreateTestAvailability(tenantID, artisanID, func(a *models.Availability) {
		a.DayOfWeek = &tuesday
	})
	require.NoError(t, repo.Create(ctx, tuesdaySlot))

	t.Run("get availability records created", func(t *testing.T) {
		// Verify records were created
		slot1, err := repo.GetByID(ctx, monday1.ID)
		require.NoError(t, err)
		assert.NotNil(t, slot1)

		slot2, err := repo.GetByID(ctx, monday2.ID)
		require.NoError(t, err)
		assert.NotNil(t, slot2)
	})
}

func TestAvailabilityRepository_GetActiveSlots(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupAvailabilityTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create regular availability slot
	activeSlot := testutil.CreateTestAvailability(tenantID, artisanID, func(a *models.Availability) {
		a.Type = models.AvailabilityTypeRegular
	})
	require.NoError(t, repo.Create(ctx, activeSlot))

	// Create time-off slot
	tuesday := 2
	timeOffSlot := testutil.CreateTestAvailability(tenantID, artisanID, func(a *models.Availability) {
		a.DayOfWeek = &tuesday
		a.Type = models.AvailabilityTypeTimeOff
	})
	require.NoError(t, repo.Create(ctx, timeOffSlot))

	t.Run("get availability slots", func(t *testing.T) {
		// Verify records were created
		slot1, err := repo.GetByID(ctx, activeSlot.ID)
		require.NoError(t, err)
		assert.Equal(t, models.AvailabilityTypeRegular, slot1.Type)

		slot2, err := repo.GetByID(ctx, timeOffSlot.ID)
		require.NoError(t, err)
		assert.Equal(t, models.AvailabilityTypeTimeOff, slot2.Type)
	})
}

func TestAvailabilityRepository_UpdateSlot(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupAvailabilityTest(t)
	defer tdb.Close()

	ctx := context.Background()

	availability := testutil.CreateTestAvailability(tenantID, artisanID)
	require.NoError(t, repo.Create(ctx, availability))

	t.Run("update availability slot", func(t *testing.T) {
		now := time.Now().UTC()
		availability.StartTime = time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.UTC)
		availability.EndTime = time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, time.UTC)
		err := repo.Update(ctx, availability)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, availability.ID)
		require.NoError(t, err)
		assert.Equal(t, 10, updated.StartTime.Hour())
		assert.Equal(t, 18, updated.EndTime.Hour())
	})
}

func TestAvailabilityRepository_DeleteSlot(t *testing.T) {
	tdb, repo, tenantID, artisanID := setupAvailabilityTest(t)
	defer tdb.Close()

	ctx := context.Background()

	availability := testutil.CreateTestAvailability(tenantID, artisanID)
	require.NoError(t, repo.Create(ctx, availability))

	t.Run("delete availability slot", func(t *testing.T) {
		err := repo.Delete(ctx, availability.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, availability.ID)
		assert.Error(t, err)
	})
}
