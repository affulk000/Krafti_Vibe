package database

import (
	"context"
	"fmt"
	"time"

	"Krafti_Vibe/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db *gorm.DB
)

// DB returns the global database instance
func DB() *gorm.DB {
	if db == nil {
		panic("database not initialized - call database.Initialize() first")
	}
	return db
}

// Initialize initializes the database connection
func Initialize(cfg *config.Config, zapLogger *zap.Logger) error {
	gormLogger := NewGormLogger(zapLogger, cfg.IsDevelopment())

	dsn := cfg.DatabaseURL()
	if dsn != "" {
		zapLogger.Info("using DATABASEURL", zap.String("dsn", dsn))
	} else if dsn = cfg.DatabaseDSN(); dsn == "" {
		zapLogger.Info("using DSN", zap.String("dsn", dsn))
	}

	zapLogger.Info("using DSN", zap.String("dsn", dsn))

	// Parse DSN (DO NOT CONNECT HERE)
	pgxCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse DSN: %w", err)
	}

	pgxCfg.ConnectTimeout = 10 * time.Second

	// ONE connection pool only
	sqlDB := stdlib.OpenDB(*pgxCfg)

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	db, err = gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger:                                   gormLogger,
		NowFunc:                                  func() time.Time { return time.Now().UTC() },
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true, // Disable FK constraints during migrations
	})

	if err != nil {
		return fmt.Errorf("failed to connect using GORM: %w", err)
	}

	// ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("DB ping failed: %w", err)
	}

	zapLogger.Info("Neon database connection established!")
	return nil
}

// Close closes the database connection
func Close() error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// HealthCheck performs a database health check
func HealthCheck(ctx context.Context) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check connection pool stats
	stats := sqlDB.Stats()
	if stats.OpenConnections >= stats.MaxOpenConnections {
		return fmt.Errorf("database connection pool exhausted")
	}

	return nil
}

// Stats returns database connection pool statistics
func Stats() (map[string]any, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return map[string]any{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}

// GormLogger adapts zap logger to GORM logger interface
type GormLogger struct {
	zapLogger *zap.Logger
	debug     bool
}

// NewGormLogger creates a new GORM logger from zap logger
func NewGormLogger(zapLogger *zap.Logger, debug bool) logger.Interface {
	return &GormLogger{
		zapLogger: zapLogger,
		debug:     debug,
	}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	return &newLogger
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...any) {
	l.zapLogger.Info(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...any) {
	l.zapLogger.Warn(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...any) {
	l.zapLogger.Error(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if !l.debug && err == nil {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		l.zapLogger.Error("database query failed",
			zap.Error(err),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
		)
	} else if l.debug {
		l.zapLogger.Debug("database query",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
		)
	}
}
