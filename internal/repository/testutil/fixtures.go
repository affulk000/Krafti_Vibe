package testutil

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateTestTenant creates a test tenant with a placeholder owner ID
// Note: The tenant requires an owner user to exist in the database before creation
// Use CreateTestTenantWithOwner if you need to create both the owner and tenant
func CreateTestTenant(overrides ...func(*models.Tenant)) *models.Tenant {
	ownerID := uuid.New() // Create a dummy owner ID

	tenant := &models.Tenant{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		OwnerID:       ownerID,
		Name:          "Test Tenant",
		Subdomain:     "test-tenant",
		Domain:        "test.example.com",
		Status:        models.TenantStatusActive,
		BusinessEmail: "tenant@example.com",
		BusinessPhone: "+1234567890",
		Plan:          models.TenantPlanSmall,
		MaxUsers:      100,
		MaxArtisans:   50,
	}

	for _, override := range overrides {
		override(tenant)
	}

	return tenant
}

// CreateTestOwner creates a test owner user (without tenant association)
// The user is created with UserStatusPending to bypass tenant validation
func CreateTestOwner(overrides ...func(*models.User)) *models.User {
	user := &models.User{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		TenantID:      nil, // Owner created before tenant
		ZitadelUserID: uuid.New().String(),
		Email:         "owner@example.com",
		FirstName:     "Tenant",
		LastName:      "Owner",
		PhoneNumber:   "+1234567890",
		Role:          models.UserRoleTenantOwner,
		Status:        models.UserStatusPending, // Pending status allows nil TenantID
		EmailVerified: true,
		PhoneVerified: false,
		Timezone:      "UTC",
		Language:      "en",
	}

	for _, override := range overrides {
		override(user)
	}

	return user
}

// CreateTestTenantWithOwner is a helper that creates both an owner and a tenant
// This should be used in tests that need a tenant, as tenants require an owner
func CreateTestTenantWithOwner(db interface {
	Create(value interface{}) *gorm.DB
}, tenantOverrides ...func(*models.Tenant)) (*models.User, *models.Tenant) {
	// Create owner user first
	owner := CreateTestOwner()
	if err := db.Create(owner).Error; err != nil {
		panic("failed to create test owner: " + err.Error())
	}

	// Create tenant with this owner
	tenant := CreateTestTenant(func(t *models.Tenant) {
		t.OwnerID = owner.ID
		// Apply any additional overrides
		for _, override := range tenantOverrides {
			override(t)
		}
	})

	if err := db.Create(tenant).Error; err != nil {
		panic("failed to create test tenant: " + err.Error())
	}

	return owner, tenant
}

// CreateTestUser creates a test user
func CreateTestUser(tenantID *uuid.UUID, overrides ...func(*models.User)) *models.User {
	user := &models.User{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		TenantID:      tenantID,
		ZitadelUserID: uuid.New().String(),
		Email:         "user@example.com",
		FirstName:     "John",
		LastName:      "Doe",
		PhoneNumber:   "+1234567890",
		Role:          models.UserRoleCustomer,
		Status:        models.UserStatusActive,
		EmailVerified: true,
		PhoneVerified: false,
		Timezone:      "UTC",
		Language:      "en",
	}

	for _, override := range overrides {
		override(user)
	}

	return user
}

// CreateTestArtisan creates a test artisan
func CreateTestArtisan(userID, tenantID uuid.UUID, overrides ...func(*models.Artisan)) *models.Artisan {
	artisan := &models.Artisan{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		UserID:          userID,
		TenantID:        tenantID,
		Bio:             "Expert artisan with 10 years of experience",
		Specialization:  models.StringArray{"woodwork", "carpentry"},
		YearsExperience: 10,
		IsAvailable:     true,
		Rating:          4.8,
		ReviewCount:     50,
		TotalBookings:   50,
	}

	for _, override := range overrides {
		override(artisan)
	}

	return artisan
}

// CreateTestCustomer creates a test customer
func CreateTestCustomer(userID, tenantID uuid.UUID, overrides ...func(*models.Customer)) *models.Customer {
	customer := &models.Customer{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		UserID:            userID,
		TenantID:          tenantID,
		TotalBookings:     5,
		TotalSpent:        1500.00,
		LoyaltyPoints:     100,
		PreferredArtisans: []uuid.UUID{},
	}

	for _, override := range overrides {
		override(customer)
	}

	return customer
}

// CreateTestService creates a test service
func CreateTestService(tenantID, artisanID uuid.UUID, overrides ...func(*models.Service)) *models.Service {
	artisanPtr := &artisanID
	service := &models.Service{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		TenantID:        tenantID,
		ArtisanID:       artisanPtr,
		Name:            "Custom Furniture",
		Description:     "High-quality custom furniture crafting",
		Category:        models.ServiceCategoryCarpentry,
		Price:           500.00,
		Currency:        "USD",
		DurationMinutes: 120,
		IsActive:        true,
	}

	for _, override := range overrides {
		override(service)
	}

	return service
}

// CreateTestBooking creates a test booking
func CreateTestBooking(tenantID, customerID, artisanID, serviceID uuid.UUID, overrides ...func(*models.Booking)) *models.Booking {
	now := time.Now().UTC()
	startTime := now.Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	booking := &models.Booking{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		TenantID:      tenantID,
		CustomerID:    customerID,
		ArtisanID:     artisanID,
		ServiceID:     serviceID,
		Status:        models.BookingStatusPending,
		StartTime:     startTime,
		EndTime:       endTime,
		Duration:      120,
		BasePrice:     500.00,
		TotalPrice:    500.00,
		Currency:      "USD",
		PaymentStatus: models.PaymentStatusPending,
	}

	for _, override := range overrides {
		override(booking)
	}

	return booking
}

// CreateTestProject creates a test project
// TODO: Fix field names - some fields don't match actual Project model
// Commented out to allow tests to compile
/*
func CreateTestProject(tenantID, customerID, artisanID, bookingID uuid.UUID, overrides ...func(*models.Project)) *models.Project {
	customerPtr := &customerID
	artisanPtr := &artisanID
	bookingPtr := &bookingID

	project := &models.Project{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		TenantID:    tenantID,
		CustomerID:  customerPtr,
		ArtisanID:   artisanPtr,
		BookingID:   bookingPtr,
		Title:       "Custom Dining Table",
		Description: "Oak dining table for 8 people",
		Status:      models.ProjectStatusInProgress,
		Priority:    models.ProjectPriorityMedium,
	}

	for _, override := range overrides {
		override(project)
	}

	return project
}
*/

// CreateTestInvoice creates a test invoice
// TODO: Fix field names - some fields don't match actual Invoice model
// Commented out to allow tests to compile
/*
func CreateTestInvoice(tenantID, customerID, artisanID, bookingID uuid.UUID, overrides ...func(*models.Invoice)) *models.Invoice {
	customerPtr := &customerID
	bookingPtr := &bookingID

	invoice := &models.Invoice{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		TenantID:      tenantID,
		CustomerID:    customerPtr,
		BookingID:     bookingPtr,
		InvoiceNumber: "INV-001",
		Status:        models.InvoiceStatusDraft,
		Currency:      "USD",
		Subtotal:      500.00,
		TaxRate:       0.08,
		TaxAmount:     40.00,
		Total:         540.00,
		DueDate:       time.Now().UTC().Add(30 * 24 * time.Hour),
	}

	for _, override := range overrides {
		override(invoice)
	}

	return invoice
}
*/

// CreateTestPayment creates a test payment
// TODO: Fix field names - some fields don't match actual Payment model
// Commented out to allow tests to compile
/*
func CreateTestPayment(tenantID, customerID, invoiceID uuid.UUID, overrides ...func(*models.Payment)) *models.Payment {
	payment := &models.Payment{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		TenantID:         tenantID,
		CustomerID:       customerID,
		InvoiceID:        &invoiceID,
		Amount:           540.00,
		Currency:         "USD",
		Status:           models.PaymentStatusPending,
		PaymentMethod:    models.PaymentMethodCard,
		TransactionID:    uuid.New().String(),
	}

	for _, override := range overrides {
		override(payment)
	}

	return payment
}
*/

// CreateTestNotification creates a test notification
// TODO: Fix field names - TenantID type mismatch (pointer vs value)
// Commented out to allow tests to compile
/*
func CreateTestNotification(tenantID, userID uuid.UUID, overrides ...func(*models.Notification)) *models.Notification {
	notification := &models.Notification{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		TenantID: &tenantID,
		UserID:   userID,
		Type:     models.NotificationTypeBookingConfirmation,
		Title:    "Booking Confirmed",
		Message:  "Your booking has been confirmed",
		Priority: models.NotificationPriorityMedium,
		IsRead:   false,
	}

	for _, override := range overrides {
		override(notification)
	}

	return notification
}
*/

// CreateTestAvailability creates a test availability slot
func CreateTestAvailability(tenantID, artisanID uuid.UUID, overrides ...func(*models.Availability)) *models.Availability {
	dayOfWeek := 1 // Monday
	now := time.Now().UTC()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC)
	endTime := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, time.UTC)

	availability := &models.Availability{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		TenantID:    tenantID,
		ArtisanID:   artisanID,
		Type:        models.AvailabilityTypeRegular,
		DayOfWeek:   &dayOfWeek,
		StartTime:   startTime,
		EndTime:     endTime,
		IsRecurring: true,
	}

	for _, override := range overrides {
		override(availability)
	}

	return availability
}

// CreateTestSDKClient creates a test SDK client
func CreateTestSDKClient(tenantID uuid.UUID, overrides ...func(*models.SDKClient)) *models.SDKClient {
	client := &models.SDKClient{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		TenantID:    tenantID,
		Name:        "Test SDK Client",
		Description: "Test SDK client for integration testing",
		Platform:    models.SDKPlatformWeb,
		Environment: models.SDKEnvironmentDevelopment,
		IsActive:    true,
		Permissions: models.SDKPermissions{
			CanReadBookings:   true,
			CanCreateBookings: true,
			CanReadServices:   true,
		},
		Scopes: models.StringArray{"read:bookings", "write:bookings"},
	}

	for _, override := range overrides {
		override(client)
	}

	return client
}

// CreateTestWhiteLabel creates a test white label configuration
func CreateTestWhiteLabel(tenantID uuid.UUID, overrides ...func(*models.WhiteLabel)) *models.WhiteLabel {
	whiteLabel := &models.WhiteLabel{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		TenantID:       tenantID,
		CompanyName:    "Test Brand",
		LogoURL:        "https://example.com/logo.png",
		PrimaryColor:   "#007bff",
		SecondaryColor: "#6c757d",
		CustomDomain:   "brand.example.com",
		IsActive:       true,
	}

	for _, override := range overrides {
		override(whiteLabel)
	}

	return whiteLabel
}
