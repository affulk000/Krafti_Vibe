package repository

import (
	"context"

	"gorm.io/gorm"
)

// Repositories aggregates all repository instances for dependency injection.
// It provides a centralized access point to all domain repositories.
type Repositories struct {
	// Core Entities
	User   UserRepository
	Tenant TenantRepository

	// Business Operations
	Booking      BookingRepository
	Service      ServiceRepository
	ServiceAddon ServiceAddonRepository
	Payment      PaymentRepository
	Invoice      InvoiceRepository
	PromoCode    PromoCodeRepository

	// Project Management
	Project          ProjectRepository
	ProjectMilestone ProjectMilestoneRepository
	ProjectTask      ProjectTaskRepository
	ProjectUpdate    ProjectUpdateRepository

	// User Management
	Artisan  ArtisanRepository
	Customer CustomerRepository
	Review   *ReviewRepository

	// Communication & Files
	Message      MessageRepository
	FileUpload   FileUploadRepository
	Notification NotificationRepository

	// Analytics & Administration
	Report              ReportRepository
	Subscription        SubscriptionRepository
	SystemSetting       SystemSettingRepository
	TenantInvitation    TenantInvitationRepository
	TenantUsageTracking TenantUsageTrackingRepository
	DataExport          DataExportRequestRepository
	WebhookEvent        WebhookEventRepository
}

// NewRepositories creates a new instance of all repositories with the given database connection.
// It accepts optional RepositoryConfig for configuring logger, cache, metrics, and audit logging.
//
// Example:
//
//	repos := NewRepositories(db, RepositoryConfig{
//		Logger: logger,
//		Cache: cache,
//	})
func NewRepositories(db *gorm.DB, config ...RepositoryConfig) *Repositories {
	var cfg RepositoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return &Repositories{
		// Core Entities
		User:   NewUserRepository(db, cfg),
		Tenant: NewTenantRepository(db, cfg),

		// Business Operations
		Booking:      NewBookingRepository(db, cfg),
		Service:      NewServiceRepository(db, cfg),
		ServiceAddon: NewServiceAddonRepository(db, cfg),
		Payment:      NewPaymentRepository(db, cfg),
		Invoice:      NewInvoiceRepository(db, cfg),
		PromoCode:    NewPromoCodeRepository(db, cfg),

		// Project Management
		Project:          NewProjectRepository(db, cfg),
		ProjectMilestone: NewProjectMilestoneRepository(db, cfg),
		ProjectTask:      NewProjectTaskRepository(db, cfg),
		ProjectUpdate:    NewProjectUpdateRepository(db, cfg),

		// User Management
		Artisan:  NewArtisanRepository(db, cfg),
		Customer: NewCustomerRepository(db, cfg),
		Review:   NewReviewRepository(db, cfg.Logger),

		// Communication & Files
		Message:      NewMessageRepository(db, cfg),
		FileUpload:   NewFileUploadRepository(db, cfg),
		Notification: NewNotificationRepository(db, cfg),

		// Analytics & Administration
		Report:              NewReportRepository(db, cfg),
		Subscription:        NewSubscriptionRepository(db, cfg),
		SystemSetting:       NewSystemSettingRepository(db, nil, cfg),
		TenantInvitation:    NewTenantInvitationRepository(db, cfg),
		TenantUsageTracking: NewTenantUsageTrackingRepository(db, cfg),
		DataExport:          NewDataExportRequestRepository(db, cfg),
		WebhookEvent:        NewWebhookEventRepository(db, cfg),
	}
}

// UnitOfWork provides transactional operations across multiple repositories.
// It ensures that a series of repository operations either all succeed or all fail together.
type UnitOfWork struct {
	db *gorm.DB
}

// NewUnitOfWork creates a new unit of work instance.
func NewUnitOfWork(db *gorm.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

// Execute runs a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
//
// Example:
//
//	uow := NewUnitOfWork(db)
//	err := uow.Execute(ctx, func(repos *Repositories) error {
//		// Create user
//		if err := repos.User.Create(ctx, &user); err != nil {
//			return err
//		}
//		// Create associated profile
//		return repos.Profile.Create(ctx, &profile)
//	})
func (uow *UnitOfWork) Execute(ctx context.Context, fn func(*Repositories) error) error {
	return uow.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repos := NewRepositories(tx)
		return fn(repos)
	})
}
