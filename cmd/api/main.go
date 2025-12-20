package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"Krafti_Vibe/internal/auth"
	"Krafti_Vibe/internal/config"
	"Krafti_Vibe/internal/infrastructure/cache"
	"Krafti_Vibe/internal/infrastructure/database"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/pkg/health"
	"Krafti_Vibe/internal/pkg/logger"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/router"
	"Krafti_Vibe/internal/service"

	_ "Krafti_Vibe/docs" // Import generated swagger docs

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// @title Krafti Vibe API
// @version 1.0
// @description Multi-tenant marketplace platform API for connecting artisans with customers. Supports booking management, payments, projects, and more.
// @description
// @description ## Authentication
// @description This API uses Bearer token authentication via Zitadel.
// @description Get your access token from the Zitadel authentication endpoint and include it in the Authorization header.
// @description
// @description ## Rate Limiting
// @description API requests are rate limited per tenant/IP. Default: 100 requests per second.
// @description
// @description ## Multi-tenancy
// @description Most endpoints require a valid tenant context. Include tenant ID in request headers or URL parameters.

// @contact.name Krafti Vibe API Support
// @contact.url https://kraftivibe.com/support
// @contact.email support@kraftivibe.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3000
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your bearer token in the format: Bearer {token}

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key for M2M authentication

// @tag.name Health
// @tag.description Health check and monitoring endpoints

// @tag.name Users
// @tag.description User management endpoints (CRUD, authentication, MFA)

// @tag.name Tenants
// @tag.description Tenant/organization management (platform admin only)

// @tag.name Services
// @tag.description Service catalog management

// @tag.name Bookings
// @tag.description Booking and appointment management

// @tag.name Projects
// @tag.description Project management and tracking

// @tag.name Payments
// @tag.description Payment processing and invoice management

// @tag.name Messages
// @tag.description Messaging and notifications

// @tag.name Reviews
// @tag.description Review and rating management

// @tag.name Artisans
// @tag.description Artisan profile management

// @tag.name Customers
// @tag.description Customer profile management

// @tag.name Admin
// @tag.description Platform administration endpoints

// Build information (set via ldflags during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
	GoVersion = runtime.Version()
)

func main() {
	// Run the application and handle exit
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}

// run contains the main application logic
func run() error {
	// ============================================================================
	// Configuration
	// ============================================================================

	fmt.Println("Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// ============================================================================
	// Logging
	// ============================================================================

	fmt.Println("Initializing logger...")
	zapLogger, err := logger.Initialize(cfg.App.LogLevel, cfg.Environment)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	// Log application startup information
	zapLogger.Info("starting application",
		zap.String("name", cfg.App.Name),
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
		zap.String("git_commit", GitCommit),
		zap.String("go_version", GoVersion),
		zap.String("environment", cfg.Environment),
		zap.String("log_level", cfg.App.LogLevel),
	)

	// Log system information
	zapLogger.Info("system information",
		zap.Int("num_cpu", runtime.NumCPU()),
		zap.String("os", runtime.GOOS),
		zap.String("arch", runtime.GOARCH),
	)

	// ============================================================================
	// Database
	// ============================================================================

	zapLogger.Info("initializing database connection",
		zap.String("host", cfg.Database.Host),
		zap.String("port", cfg.Database.Port),
		zap.String("database", cfg.Database.DBName),
	)

	if err := database.Initialize(cfg, zapLogger); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() {
		zapLogger.Info("closing database connection")
		database.Close()
	}()

	db := database.DB()
	zapLogger.Info("database connection established")

	// Run database migrations
	if err := runMigrations(db, zapLogger, cfg); err != nil {
		zapLogger.Error("database migrations failed", zap.Error(err))
		// Don't fail startup on migration errors in production
		if !cfg.IsProduction() {
			return fmt.Errorf("migration failed: %w", err)
		}
		zapLogger.Warn("continuing startup despite migration errors (production mode)")
	}

	// ============================================================================
	// Redis Cache
	// ============================================================================

	zapLogger.Info("initializing redis cache",
		zap.String("host", cfg.Redis.Host),
		zap.String("port", cfg.Redis.Port),
	)

	redisCache, err := cache.NewRedisClient(cache.RedisConfig{
		Host:            cfg.Redis.Host,
		Port:            cfg.Redis.Port,
		Password:        cfg.Redis.Password,
		DB:              cfg.Redis.DB,
		MaxRetries:      cfg.Redis.MaxRetries,
		PoolSize:        cfg.Redis.PoolSize,
		MinIdleConns:    cfg.Redis.MinIdleConns,
		DialTimeout:     cfg.Redis.DialTimeout,
		ReadTimeout:     cfg.Redis.ReadTimeout,
		WriteTimeout:    cfg.Redis.WriteTimeout,
		PoolTimeout:     cfg.Redis.PoolTimeout,
		ConnMaxIdleTime: cfg.Redis.ConnMaxIdleTime,
		KeyPrefix:       "kraftivibe",
	}, zapLogger)
	if err != nil {
		zapLogger.Error("failed to initialize redis cache", zap.Error(err))
		// Continue without cache in development
		if !cfg.IsProduction() {
			zapLogger.Warn("continuing without cache (development mode)")
		} else {
			return fmt.Errorf("redis cache required in production: %w", err)
		}
	} else {
		defer func() {
			zapLogger.Info("closing redis connection")
			if err := redisCache.Close(); err != nil {
				zapLogger.Error("error closing redis", zap.Error(err))
			}
		}()
		zapLogger.Info("redis cache initialized")
	}

	// ============================================================================
	// Fiber Application
	// ============================================================================

	zapLogger.Info("initializing fiber application")

	app := fiber.New(fiber.Config{
		AppName:               fmt.Sprintf("%s v%s", cfg.App.Name, Version),
		ServerHeader:          cfg.App.Name,
		ErrorHandler:          middleware.ErrorHandler(zapLogger),
		ReadTimeout:           cfg.Server.ReadTimeout,
		WriteTimeout:          cfg.Server.WriteTimeout,
		IdleTimeout:           cfg.Server.IdleTimeout,
		DisableStartupMessage: cfg.IsProduction(),
		EnablePrintRoutes:     cfg.IsDevelopment(),
		Prefork:               false, // Set to true for production multi-process
		CaseSensitive:         true,
		StrictRouting:         false,
		BodyLimit:             10 * 1024 * 1024, // 10MB
		Concurrency:           256 * 1024,       // Maximum number of concurrent connections
		ReadBufferSize:        4096,
		WriteBufferSize:       4096,
		CompressedFileSuffix:  ".fiber.gz",
		ProxyHeader:           fiber.HeaderXForwardedFor,
		GETOnly:               false,
		DisableKeepalive:      false,
		ReduceMemoryUsage:     cfg.IsProduction(),
	})

	// ============================================================================
	// Global Middleware
	// ============================================================================

	zapLogger.Info("configuring middleware")

	// Recovery middleware - must be first to catch panics
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.IsDevelopment(),
		StackTraceHandler: func(c *fiber.Ctx, e any) {
			zapLogger.Error("panic recovered",
				zap.Any("panic", e),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.String("ip", c.IP()),
			)
		},
	}))

	// Request ID middleware - for request tracing
	app.Use(requestid.New(requestid.Config{
		Header:     "X-Request-ID",
		ContextKey: "request_id",
		Generator: func() string {
			return fmt.Sprintf("%d", time.Now().UnixNano())
		},
	}))

	// Logger middleware - for request logging
	if cfg.IsDevelopment() {
		app.Use(fiberLogger.New(fiberLogger.Config{
			Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
			TimeFormat: "15:04:05",
			TimeZone:   "Local",
		}))
	}

	// Compression middleware - for response compression
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // LevelBestSpeed for performance, LevelBestCompression for size
	}))

	// Profiling endpoints (development and staging only)
	if !cfg.IsProduction() {
		zapLogger.Info("enabling pprof profiling endpoints")
		app.Use(pprof.New())
	}

	// ============================================================================
	// Health Check Endpoints
	// ============================================================================

	zapLogger.Info("setting up health check endpoints")

	healthChecker := health.NewHealthChecker(
		&health.DatabaseChecker{},
	)

	// Liveness probe - simple check if server is running
	app.Get("/health/live", health.LivenessHandler())

	// Readiness probe - checks if service is ready to accept traffic
	app.Get("/health/ready", health.ReadinessHandler(healthChecker))

	// Combined health check with detailed information
	app.Get("/health", health.Handler(healthChecker))

	// Version endpoint
	app.Get("/version", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"version":    Version,
			"build_time": BuildTime,
			"git_commit": GitCommit,
			"go_version": GoVersion,
		})
	})

	// ============================================================================
	// Zitadel Authentication
	// ============================================================================

	zapLogger.Info("initializing Zitadel authentication",
		zap.String("domain", cfg.Zitadel.Domain),
	)

	// Initialize Zitadel authentication
	zitadelAuth, err := auth.NewZitadelAuth(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize Zitadel authentication: %w", err)
	}

	zapLogger.Info("Zitadel authentication initialized successfully")

	// ============================================================================
	// User Sync Service (for Zitadel user synchronization)
	// ============================================================================

	zapLogger.Info("initializing user sync service")

	// Create fiber logger adapter (defined later)
	fiberLogger := logger.NewFiberLogger(zapLogger)

	// Create user repository for sync service
	userRepo := repository.NewUserRepository(db, repository.RepositoryConfig{
		Logger: fiberLogger,
	})

	// Create user sync service
	userSyncService := service.NewUserSyncService(userRepo, zapLogger)

	// Create Zitadel authentication middleware with user sync
	zitadelMiddleware := middleware.NewZitadelAuthMiddleware(
		zitadelAuth.AuthZ,
		userSyncService,
	)

	zapLogger.Info("user sync service initialized - users will be auto-synced on authentication")

	// ============================================================================
	// API Router Setup
	// ============================================================================

	zapLogger.Info("setting up API routes")

	// Create CORS configuration
	corsConfig := &middleware.CORSConfig{
		AllowedOrigins:   cfg.App.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		AllowCredentials: true,
		ExposedHeaders:   []string{"X-Request-ID"},
		MaxAge:           3600,
		Logger:           zapLogger,
	}

	// Initialize router with all dependencies
	routerConfig := &router.Config{
		DB:                db,
		Logger:            fiberLogger,
		ZitadelAuthZ:      zitadelAuth.AuthZ,
		ZitadelMiddleware: zitadelMiddleware,
		Cache:             redisCache,
		ZapLogger:         zapLogger,
		CORSConfig:        corsConfig,
		WebhookSecret:     "",
	}

	apiRouter := router.New(app, routerConfig)
	if err := apiRouter.Setup(); err != nil {
		return fmt.Errorf("failed to setup routes: %w", err)
	}

	zapLogger.Info("API routes configured")

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":      "route not found",
			"code":       "NOT_FOUND",
			"path":       c.Path(),
			"method":     c.Method(),
			"request_id": c.Locals("request_id"),
		})
	})

	// ============================================================================
	// Start Server
	// ============================================================================

	serverAddr := cfg.ServerAddr()
	zapLogger.Info("starting server",
		zap.String("address", serverAddr),
		zap.String("environment", cfg.Environment),
	)

	// Print banner in development
	if cfg.IsDevelopment() {
		printBanner(cfg, serverAddr)
	}

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := app.Listen(serverAddr); err != nil {
			serverErr <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	zapLogger.Info("server started successfully")

	// ============================================================================
	// Graceful Shutdown
	// ============================================================================

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-serverErr:
		zapLogger.Error("server error occurred", zap.Error(err))
		return err
	case sig := <-quit:
		zapLogger.Info("shutdown signal received",
			zap.String("signal", sig.String()),
		)
	}

	// Graceful shutdown with timeout
	zapLogger.Info("initiating graceful shutdown",
		zap.Duration("timeout", cfg.Server.ShutdownTimeout),
	)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		zapLogger.Error("server forced to shutdown", zap.Error(err))
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	zapLogger.Info("server gracefully stopped")
	zapLogger.Info("application shutdown complete")

	return nil
}

// runMigrations runs database migrations
func runMigrations(db *gorm.DB, logger *zap.Logger, cfg *config.Config) error {
	logger.Info("checking database migrations")

	// Configure migration settings based on environment
	migrationConfig := database.MigrationConfig{
		AutoMigrate:    true,
		SeedData:       cfg.IsDevelopment(), // Only seed in development
		Force:          false,
		DryRun:         false,
		Logger:         logger,
		SkipExtensions: false,
	}

	// Run migrations
	if err := database.RunMigrations(db, migrationConfig); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Log migration status
	migrations, err := database.GetMigrationStatus(db)
	if err != nil {
		logger.Warn("failed to get migration status", zap.Error(err))
	} else if len(migrations) > 0 {
		logger.Info("migration status",
			zap.Int("total_migrations", len(migrations)),
			zap.String("last_migration", migrations[0].Version),
			zap.Time("last_applied", migrations[0].AppliedAt),
		)
	}

	logger.Info("database migrations completed")
	return nil
}

// printBanner prints startup banner in development mode
func printBanner(cfg *config.Config, addr string) {
	banner := fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║                   🎨 KRAFTI VIBE API 🎨                     ║
║                                                              ║
║  Version:     %-47s ║
║  Environment: %-47s ║
║  Address:     %-47s ║
║                                                              ║
║  📚 API Documentation:                                       ║
║     http://%s/api/v1                                ║
║                                                              ║
║  💓 Health Checks:                                           ║
║     http://%s/health                                ║
║     http://%s/health/live                           ║
║     http://%s/health/ready                          ║
║                                                              ║
║  🔍 Profiling (Dev Only):                                    ║
║     http://%s/debug/pprof/                          ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
`, Version, cfg.Environment, addr, addr, addr, addr, addr, addr)

	fmt.Println(banner)
}
