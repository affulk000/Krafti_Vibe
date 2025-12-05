package main

import (
	"flag"
	"fmt"
	"os"

	"Krafti_Vibe/internal/config"
	"Krafti_Vibe/internal/infrastructure/database"
	"Krafti_Vibe/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// Define command-line flags
	var (
		action       = flag.String("action", "up", "Migration action: up, down, status, seed")
		dryRun       = flag.Bool("dry-run", false, "Perform a dry run without applying changes")
		force        = flag.Bool("force", false, "Force migration even if risky")
		skipSeed     = flag.Bool("skip-seed", false, "Skip data seeding")
		skipExtensions = flag.Bool("skip-extensions", false, "Skip PostgreSQL extension creation")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	zapLogger, err := logger.Initialize("info", cfg.Environment)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	zapLogger.Info("migration tool started",
		zap.String("action", *action),
		zap.Bool("dry_run", *dryRun),
		zap.Bool("force", *force),
	)

	// Initialize database
	if err := database.Initialize(cfg, zapLogger); err != nil {
		zapLogger.Fatal("failed to initialize database", zap.Error(err))
	}
	defer database.Close()

	db := database.DB()

	// Execute migration action
	switch *action {
	case "up":
		if err := migrateUp(db, zapLogger, cfg, *dryRun, *force, *skipSeed, *skipExtensions); err != nil {
			zapLogger.Fatal("migration up failed", zap.Error(err))
		}

	case "down":
		if err := migrateDown(db, zapLogger); err != nil {
			zapLogger.Fatal("migration down failed", zap.Error(err))
		}

	case "status":
		if err := migrationStatus(db, zapLogger); err != nil {
			zapLogger.Fatal("failed to get migration status", zap.Error(err))
		}

	case "seed":
		if err := seedData(db, zapLogger); err != nil {
			zapLogger.Fatal("data seeding failed", zap.Error(err))
		}

	default:
		zapLogger.Fatal("unknown action",
			zap.String("action", *action),
			zap.String("valid_actions", "up, down, status, seed"),
		)
	}

	zapLogger.Info("migration tool completed successfully")
}

// migrateUp runs database migrations
func migrateUp(db *gorm.DB, logger *zap.Logger, cfg *config.Config, dryRun, force, skipSeed, skipExtensions bool) error {
	logger.Info("running migrations up")

	migrationConfig := database.MigrationConfig{
		AutoMigrate:    true,
		SeedData:       !skipSeed && cfg.IsDevelopment(),
		Force:          force,
		DryRun:         dryRun,
		Logger:         logger,
		SkipExtensions: skipExtensions,
	}

	if err := database.RunMigrations(db, migrationConfig); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Show migration status
	migrations, err := database.GetMigrationStatus(db)
	if err != nil {
		logger.Warn("failed to get migration status", zap.Error(err))
	} else {
		logger.Info("migration completed",
			zap.Int("total_migrations", len(migrations)),
		)
		if len(migrations) > 0 {
			logger.Info("last migration",
				zap.String("version", migrations[0].Version),
				zap.Time("applied_at", migrations[0].AppliedAt),
			)
		}
	}

	return nil
}

// migrateDown rolls back the last migration
func migrateDown(db *gorm.DB, logger *zap.Logger) error {
	logger.Warn("rolling back last migration")

	// Show confirmation warning
	fmt.Println("\n⚠️  WARNING: You are about to rollback the last migration!")
	fmt.Println("This will delete the migration record but NOT automatically reverse schema changes.")
	fmt.Println("Manual schema rollback may be required.")
	fmt.Print("\nType 'yes' to confirm: ")

	var confirmation string
	fmt.Scanln(&confirmation)

	if confirmation != "yes" {
		logger.Info("rollback cancelled")
		return nil
	}

	if err := database.RollbackLastMigration(db, logger); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	logger.Info("rollback completed")
	return nil
}

// migrationStatus shows the current migration status
func migrationStatus(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("fetching migration status")

	migrations, err := database.GetMigrationStatus(db)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	if len(migrations) == 0 {
		fmt.Println("\n📊 Migration Status:")
		fmt.Println("No migrations have been applied yet.")
		return nil
	}

	fmt.Println("\n📊 Migration Status:")
	fmt.Printf("Total migrations applied: %d\n\n", len(migrations))

	// Print table header
	fmt.Printf("%-5s | %-35s | %-50s | %s\n", "ID", "Version", "Description", "Applied At")
	fmt.Println("------|-------------------------------------|-----------------------------------------------------|-------------------------")

	// Print migrations
	for _, migration := range migrations {
		fmt.Printf("%-5d | %-35s | %-50s | %s\n",
			migration.ID,
			migration.Version,
			truncate(migration.Description, 50),
			migration.AppliedAt.Format("2006-01-02 15:04:05"),
		)
	}

	fmt.Println()
	return nil
}

// seedData seeds initial data
func seedData(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("seeding initial data")

	migrationConfig := database.MigrationConfig{
		AutoMigrate:    false,
		SeedData:       true,
		Force:          false,
		DryRun:         false,
		Logger:         logger,
		SkipExtensions: true,
	}

	if err := database.RunMigrations(db, migrationConfig); err != nil {
		return fmt.Errorf("seeding failed: %w", err)
	}

	logger.Info("data seeding completed")
	return nil
}

// truncate truncates a string to the specified length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
