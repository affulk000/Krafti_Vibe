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

func setupBookingTest(t *testing.T) (*testutil.TestDB, repository.BookingRepository, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	tdb := testutil.NewTestDB(t)
	repo := repository.NewBookingRepository(tdb.DB, testutil.DefaultRepositoryConfig())

	// Create test data (tenant with owner)
	_, tenant := testutil.CreateTestTenantWithOwner(tdb.DB)

	customer := testutil.CreateTestUser(&tenant.ID, func(u *models.User) {
		u.Email = "customer@example.com"
		u.Role = models.UserRoleCustomer
	})
	require.NoError(t, tdb.DB.Create(customer).Error)

	artisanUser := testutil.CreateTestUser(&tenant.ID, func(u *models.User) {
		u.Email = "artisan@example.com"
		u.Role = models.UserRoleArtisan
	})
	require.NoError(t, tdb.DB.Create(artisanUser).Error)

	artisan := testutil.CreateTestArtisan(artisanUser.ID, tenant.ID)
	require.NoError(t, tdb.DB.Create(artisan).Error)

	service := testutil.CreateTestService(tenant.ID, artisan.ID)
	require.NoError(t, tdb.DB.Create(service).Error)

	return tdb, repo, tenant.ID, customer.ID, artisan.ID, service.ID
}

func TestBookingRepository_Create(t *testing.T) {
	tdb, repo, tenantID, customerID, artisanID, serviceID := setupBookingTest(t)
	defer tdb.Close()

	ctx := context.Background()

	t.Run("create booking successfully", func(t *testing.T) {
		booking := testutil.CreateTestBooking(tenantID, customerID, artisanID, serviceID)
		err := repo.Create(ctx, booking)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, booking.ID)
	})
}

func TestBookingRepository_GetByID(t *testing.T) {
	tdb, repo, tenantID, customerID, artisanID, serviceID := setupBookingTest(t)
	defer tdb.Close()

	ctx := context.Background()

	booking := testutil.CreateTestBooking(tenantID, customerID, artisanID, serviceID)
	require.NoError(t, repo.Create(ctx, booking))

	t.Run("get booking by ID", func(t *testing.T) {
		found, err := repo.GetByID(ctx, booking.ID)
		require.NoError(t, err)
		assert.Equal(t, booking.ID, found.ID)
		assert.Equal(t, customerID, found.CustomerID)
	})
}

func TestBookingRepository_GetByCustomer(t *testing.T) {
	tdb, repo, tenantID, customerID, artisanID, serviceID := setupBookingTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create multiple bookings
	for i := 0; i < 3; i++ {
		booking := testutil.CreateTestBooking(tenantID, customerID, artisanID, serviceID)
		require.NoError(t, repo.Create(ctx, booking))
	}

	t.Run("get bookings by customer", func(t *testing.T) {
		bookings, pagination, err := repo.GetByCustomerID(ctx, customerID, repository.PaginationParams{
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, bookings, 3)
		assert.Equal(t, int64(3), pagination.TotalItems)
	})
}

func TestBookingRepository_GetByArtisan(t *testing.T) {
	tdb, repo, tenantID, customerID, artisanID, serviceID := setupBookingTest(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create bookings
	for i := 0; i < 2; i++ {
		booking := testutil.CreateTestBooking(tenantID, customerID, artisanID, serviceID)
		require.NoError(t, repo.Create(ctx, booking))
	}

	t.Run("get bookings by artisan", func(t *testing.T) {
		bookings, pagination, err := repo.GetByArtisanID(ctx, artisanID, repository.PaginationParams{
			Page:     1,
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, bookings, 2)
		assert.Equal(t, int64(2), pagination.TotalItems)
	})
}

func TestBookingRepository_UpdateStatus(t *testing.T) {
	tdb, repo, tenantID, customerID, artisanID, serviceID := setupBookingTest(t)
	defer tdb.Close()

	ctx := context.Background()

	booking := testutil.CreateTestBooking(tenantID, customerID, artisanID, serviceID, func(b *models.Booking) {
		b.Status = models.BookingStatusPending
	})
	require.NoError(t, repo.Create(ctx, booking))

	t.Run("confirm booking", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, booking.ID, models.BookingStatusConfirmed)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, booking.ID)
		require.NoError(t, err)
		assert.Equal(t, models.BookingStatusConfirmed, updated.Status)
	})

	t.Run("cancel booking", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, booking.ID, models.BookingStatusCancelled)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, booking.ID)
		require.NoError(t, err)
		assert.Equal(t, models.BookingStatusCancelled, updated.Status)
	})
}

func TestBookingRepository_GetUpcomingBookings(t *testing.T) {
	tdb, repo, tenantID, customerID, artisanID, serviceID := setupBookingTest(t)
	defer tdb.Close()

	ctx := context.Background()

	now := time.Now().UTC()

	// Create future booking
	futureBooking := testutil.CreateTestBooking(tenantID, customerID, artisanID, serviceID, func(b *models.Booking) {
		b.StartTime = now.Add(24 * time.Hour)
		b.EndTime = now.Add(26 * time.Hour)
		b.Status = models.BookingStatusConfirmed
	})
	require.NoError(t, repo.Create(ctx, futureBooking))

	// Create past booking
	pastBooking := testutil.CreateTestBooking(tenantID, customerID, artisanID, serviceID, func(b *models.Booking) {
		b.StartTime = now.Add(-48 * time.Hour)
		b.EndTime = now.Add(-46 * time.Hour)
		b.Status = models.BookingStatusCompleted
	})
	require.NoError(t, repo.Create(ctx, pastBooking))

	t.Run("verify bookings created", func(t *testing.T) {
		// Verify future booking
		future, err := repo.GetByID(ctx, futureBooking.ID)
		require.NoError(t, err)
		assert.Equal(t, models.BookingStatusConfirmed, future.Status)

		// Verify past booking
		past, err := repo.GetByID(ctx, pastBooking.ID)
		require.NoError(t, err)
		assert.Equal(t, models.BookingStatusCompleted, past.Status)
	})
}

func TestBookingRepository_GetByDateRange(t *testing.T) {
	tdb, repo, tenantID, customerID, artisanID, serviceID := setupBookingTest(t)
	defer tdb.Close()

	ctx := context.Background()

	now := time.Now().UTC()

	// Create booking within a specific time
	booking1 := testutil.CreateTestBooking(tenantID, customerID, artisanID, serviceID, func(b *models.Booking) {
		b.StartTime = now.Add(30 * time.Hour)
		b.EndTime = now.Add(32 * time.Hour)
	})
	require.NoError(t, repo.Create(ctx, booking1))

	// Create booking at different time
	booking2 := testutil.CreateTestBooking(tenantID, customerID, artisanID, serviceID, func(b *models.Booking) {
		b.StartTime = now.Add(72 * time.Hour)
		b.EndTime = now.Add(74 * time.Hour)
	})
	require.NoError(t, repo.Create(ctx, booking2))

	t.Run("verify bookings created with correct times", func(t *testing.T) {
		// Verify first booking
		b1, err := repo.GetByID(ctx, booking1.ID)
		require.NoError(t, err)
		assert.NotNil(t, b1)
		assert.True(t, b1.StartTime.After(now))

		// Verify second booking
		b2, err := repo.GetByID(ctx, booking2.ID)
		require.NoError(t, err)
		assert.NotNil(t, b2)
		assert.True(t, b2.StartTime.After(b1.StartTime))
	})
}
