package database

import (
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MigrationVersion tracks database schema versions
type MigrationVersion struct {
	ID          uint      `gorm:"primaryKey"`
	Version     string    `gorm:"uniqueIndex;not null"`
	Description string    `gorm:"not null"`
	AppliedAt   time.Time `gorm:"autoCreateTime"`
}

// TableName specifies the table name for MigrationVersion
func (MigrationVersion) TableName() string {
	return "schema_migrations"
}

// MigrationConfig holds migration configuration
type MigrationConfig struct {
	AutoMigrate    bool // Run auto-migration
	SeedData       bool // Seed initial data
	Force          bool // Force migration even if risky
	DryRun         bool // Don't actually apply migrations
	Logger         *zap.Logger
	SkipExtensions bool // Skip PostgreSQL extension creation
}

// DefaultMigrationConfig returns default migration configuration
func DefaultMigrationConfig(logger *zap.Logger) MigrationConfig {
	return MigrationConfig{
		AutoMigrate:    true,
		SeedData:       false,
		Force:          false,
		DryRun:         false,
		Logger:         logger,
		SkipExtensions: false,
	}
}

// RunMigrations executes database migrations
func RunMigrations(db *gorm.DB, config MigrationConfig) error {
	if config.Logger == nil {
		return fmt.Errorf("logger is required for migrations")
	}

	logger := config.Logger
	logger.Info("starting database migrations")

	// Ensure migration version table exists
	if err := ensureMigrationTable(db, logger); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// Enable PostgreSQL extensions
	if !config.SkipExtensions {
		if err := enableExtensions(db, logger); err != nil {
			logger.Warn("failed to enable PostgreSQL extensions", zap.Error(err))
		}
	}

	// Run auto-migration
	if config.AutoMigrate {
		if config.DryRun {
			logger.Info("dry-run mode: skipping actual migration")
			return nil
		}

		if err := autoMigrate(db, logger); err != nil {
			return fmt.Errorf("auto-migration failed: %w", err)
		}

		// Record migration version
		if err := recordMigration(db, logger); err != nil {
			logger.Warn("failed to record migration version", zap.Error(err))
		}
	}

	// Seed initial data
	if config.SeedData {
		if err := seedData(db, logger); err != nil {
			logger.Error("data seeding failed", zap.Error(err))
			if !config.Force {
				return fmt.Errorf("data seeding failed: %w", err)
			}
		}
	}

	logger.Info("database migrations completed successfully")
	return nil
}

// ensureMigrationTable creates the schema_migrations table if it doesn't exist
func ensureMigrationTable(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("ensuring migration version table exists")

	if err := db.AutoMigrate(&MigrationVersion{}); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	return nil
}

// enableExtensions enables required PostgreSQL extensions
func enableExtensions(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("enabling PostgreSQL extensions")

	extensions := []string{
		"uuid-ossp",  // UUID generation
		"pg_trgm",    // Trigram matching for text search
		"btree_gin",  // GIN indexes for better performance
		"btree_gist", // GIST indexes
		"pgcrypto",   // Cryptographic functions
	}

	for _, ext := range extensions {
		logger.Info("enabling extension", zap.String("extension", ext))
		if err := db.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\"", ext)).Error; err != nil {
			logger.Warn("failed to enable extension",
				zap.String("extension", ext),
				zap.Error(err),
			)
			// Continue with other extensions
		}
	}

	return nil
}

// autoMigrate runs GORM auto-migration for all models
func autoMigrate(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("running auto-migration for all models")

	// Note: Foreign key constraint creation is disabled in database.go config
	// This prevents circular dependency issues during migration
	logger.Info("migrating all models (without foreign key constraints)")

	allModels := []any{
		// Core entities (no dependencies)
		&models.Tenant{},
		&models.User{},

		// User-related entities (depend on User/Tenant)
		&models.Artisan{},
		&models.Customer{},

		// Service-related entities
		&models.Service{},
		&models.ServiceAddon{},

		// Booking and scheduling
		&models.Availability{},
		&models.Booking{},

		// Project management
		&models.Project{},
		&models.ProjectMilestone{},
		&models.ProjectTask{},
		&models.ProjectUpdate{},

		// Financial entities
		&models.Payment{},
		&models.Invoice{},
		&models.PromoCode{},
		&models.Subscription{},

		// Communication
		&models.Message{},
		&models.Notification{},
		&models.EmailTemplate{},

		// File management
		&models.FileUpload{},

		// Reviews and ratings
		&models.Review{},

		// Analytics and reporting
		&models.AnalyticsEvent{},
		&models.Report{},

		// System and administration
		&models.SystemSetting{},
		&models.TenantInvitation{},
		&models.TenantUsageTracking{},
		&models.DataExportRequest{},
		&models.WebhookEvent{},
		&models.AuditLog{},
		&models.APIKey{},

		// Branding and customization
		&models.WhiteLabel{},
	}

	// Run migration for all models at once
	logger.Info("creating/updating all tables", zap.Int("model_count", len(allModels)))
	if err := db.AutoMigrate(allModels...); err != nil {
		logger.Error("auto-migration failed", zap.Error(err))
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	logger.Info("all models migrated successfully")

	// Create indexes for better performance
	if err := createIndexes(db, logger); err != nil {
		logger.Warn("failed to create some indexes", zap.Error(err))
	}

	logger.Info("auto-migration completed successfully")
	return nil
}

// createIndexes creates additional database indexes for performance
func createIndexes(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("creating additional indexes")

	indexes := []struct {
		table   string
		name    string
		columns string
	}{
		// User indexes
		{"users", "idx_users_email", "email"},
		{"users", "idx_users_tenant_email", "tenant_id, email"},
		{"users", "idx_users_status", "status"},

		// Booking indexes
		{"bookings", "idx_bookings_tenant", "tenant_id"},
		{"bookings", "idx_bookings_customer", "customer_id"},
		{"bookings", "idx_bookings_artisan", "artisan_id"},
		{"bookings", "idx_bookings_status", "status"},
		{"bookings", "idx_bookings_start_time", "start_time"},

		// Service indexes
		{"services", "idx_services_tenant", "tenant_id"},
		{"services", "idx_services_is_active", "is_active"},
		{"services", "idx_services_category", "category"},

		// Payment indexes
		{"payments", "idx_payments_tenant", "tenant_id"},
		{"payments", "idx_payments_booking", "booking_id"},
		{"payments", "idx_payments_status", "status"},
		{"payments", "idx_payments_method", "method"},

		// Project indexes
		{"projects", "idx_projects_tenant", "tenant_id"},
		{"projects", "idx_projects_customer", "customer_id"},
		{"projects", "idx_projects_artisan", "artisan_id"},
		{"projects", "idx_projects_status", "status"},

		// Notification indexes
		{"notifications", "idx_notifications_user", "user_id"},
		{"notifications", "idx_notifications_read", "read_at"},

		// Webhook indexes
		{"webhook_events", "idx_webhook_tenant", "tenant_id"},
		{"webhook_events", "idx_webhook_delivered", "delivered"},
		{"webhook_events", "idx_webhook_next_retry", "next_retry_at"},

		// Audit log indexes
		{"audit_logs", "idx_audit_tenant", "tenant_id"},
		{"audit_logs", "idx_audit_user", "user_id"},
		{"audit_logs", "idx_audit_action", "action"},
		{"audit_logs", "idx_audit_created", "created_at"},
	}

	for _, idx := range indexes {
		sql := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", idx.name, idx.table, idx.columns)
		if err := db.Exec(sql).Error; err != nil {
			logger.Warn("failed to create index",
				zap.String("index", idx.name),
				zap.String("table", idx.table),
				zap.Error(err),
			)
			// Continue with other indexes
		}
	}

	// Create full-text search indexes
	ftsIndexes := []struct {
		table  string
		name   string
		column string
	}{
		{"services", "idx_services_name_fts", "name"},
		{"services", "idx_services_description_fts", "description"},
		{"artisans", "idx_artisans_bio_fts", "bio"},
	}

	for _, idx := range ftsIndexes {
		sql := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s USING gin(to_tsvector('english', %s))",
			idx.name, idx.table, idx.column)
		if err := db.Exec(sql).Error; err != nil {
			logger.Warn("failed to create full-text search index",
				zap.String("index", idx.name),
				zap.String("table", idx.table),
				zap.Error(err),
			)
		}
	}

	logger.Info("index creation completed")
	return nil
}

// recordMigration records the migration in the schema_migrations table
func recordMigration(db *gorm.DB, logger *zap.Logger) error {
	version := fmt.Sprintf("v1.0.0_%s", time.Now().Format("20060102_150405"))
	migration := &MigrationVersion{
		Version:     version,
		Description: "Auto-migration of all models",
	}

	// Check if this version already exists
	var existing MigrationVersion
	if err := db.Where("version = ?", version).First(&existing).Error; err == nil {
		logger.Info("migration version already recorded", zap.String("version", version))
		return nil
	}

	if err := db.Create(migration).Error; err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	logger.Info("migration version recorded", zap.String("version", version))
	return nil
}

// seedData seeds initial data into the database
func seedData(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("seeding initial data")

	// Check if data already exists
	var count int64
	if err := db.Model(&models.SystemSetting{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing data: %w", err)
	}

	if count > 0 {
		logger.Info("data already seeded, skipping")
		return nil
	}

	// Seed system settings
	systemSettings := []models.SystemSetting{
		{
			Key:         "app.name",
			Value:       "Krafti Vibe",
			Category:    "general",
			Description: "Application name",
			IsPublic:    true,
		},
		{
			Key:         "app.version",
			Value:       "1.0.0",
			Category:    "general",
			Description: "Application version",
			IsPublic:    true,
		},
		{
			Key:         "features.booking.enabled",
			Value:       "true",
			Category:    "features",
			Description: "Enable booking feature",
			IsPublic:    false,
		},
		{
			Key:         "features.payments.enabled",
			Value:       "true",
			Category:    "features",
			Description: "Enable payments feature",
			IsPublic:    false,
		},
		{
			Key:         "features.projects.enabled",
			Value:       "true",
			Category:    "features",
			Description: "Enable projects feature",
			IsPublic:    false,
		},
		{
			Key:         "limits.max_file_size",
			Value:       "10485760",
			Category:    "limits",
			Description: "Maximum file upload size in bytes (10MB)",
			IsPublic:    false,
		},
		{
			Key:         "limits.max_bookings_per_day",
			Value:       "50",
			Category:    "limits",
			Description: "Maximum bookings per artisan per day",
			IsPublic:    false,
		},
	}

	for _, setting := range systemSettings {
		if err := db.Create(&setting).Error; err != nil {
			logger.Error("failed to seed system setting",
				zap.String("key", setting.Key),
				zap.Error(err),
			)
			return fmt.Errorf("failed to seed system setting %s: %w", setting.Key, err)
		}
	}

	logger.Info("initial data seeded successfully",
		zap.Int("system_settings", len(systemSettings)),
	)

	return nil
}

// GetMigrationStatus returns the current migration status
func GetMigrationStatus(db *gorm.DB) ([]MigrationVersion, error) {
	var migrations []MigrationVersion
	if err := db.Order("applied_at DESC").Find(&migrations).Error; err != nil {
		return nil, fmt.Errorf("failed to get migration status: %w", err)
	}
	return migrations, nil
}

// RollbackLastMigration rolls back the last migration (WARNING: Not fully implemented)
func RollbackLastMigration(db *gorm.DB, logger *zap.Logger) error {
	logger.Warn("rollback is not fully implemented - use with caution")

	// Get the last migration
	var lastMigration MigrationVersion
	if err := db.Order("applied_at DESC").First(&lastMigration).Error; err != nil {
		return fmt.Errorf("no migrations to rollback: %w", err)
	}

	logger.Info("rolling back migration",
		zap.String("version", lastMigration.Version),
		zap.Time("applied_at", lastMigration.AppliedAt),
	)

	// Delete the migration record
	if err := db.Delete(&lastMigration).Error; err != nil {
		return fmt.Errorf("failed to delete migration record: %w", err)
	}

	logger.Warn("migration record deleted - manual schema rollback may be required")
	return nil
}
