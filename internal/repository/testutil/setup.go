package testutil

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/repository"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDB sets up a PostgreSQL database for testing using testcontainers
type TestDB struct {
	DB        *gorm.DB
	Container *postgrescontainer.PostgresContainer
	ctx       context.Context
}

// NewTestDB creates a new test database using PostgreSQL in a container
func NewTestDB(t *testing.T) *TestDB {
	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := postgrescontainer.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgrescontainer.WithDatabase("testdb"),
		postgrescontainer.WithUsername("testuser"),
		postgrescontainer.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Connect with GORM
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		DisableForeignKeyConstraintWhenMigrating: true, // Disable FK constraints to handle circular dependencies
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate all models
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return &TestDB{
		DB:        db,
		Container: pgContainer,
		ctx:       ctx,
	}
}

// Close closes the test database connection and terminates the container
func (tdb *TestDB) Close() error {
	sqlDB, err := tdb.DB.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Close(); err != nil {
		return err
	}
	if tdb.Container != nil {
		return tdb.Container.Terminate(tdb.ctx)
	}
	return nil
}

// Cleanup removes all data from the database
func (tdb *TestDB) Cleanup() error {
	// Get all table names from PostgreSQL
	var tables []string
	if err := tdb.DB.Raw(`
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
	`).Scan(&tables).Error; err != nil {
		return err
	}

	// Disable triggers and delete all data from each table
	for _, table := range tables {
		if err := tdb.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			return err
		}
	}

	return nil
}

// AutoMigrate runs migrations for all models
// PostgreSQL supports CHECK constraints properly, so we can use standard AutoMigrate
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Tenant{},
		&models.TenantInvitation{},
		&models.Artisan{},
		&models.Customer{},
		&models.Service{},
		&models.ServiceAddon{},
		&models.Availability{},
		&models.Booking{},
		&models.Project{},
		&models.ProjectMilestone{},
		&models.ProjectTask{},
		&models.ProjectUpdate{},
		&models.Review{},
		&models.Invoice{},
		&models.Payment{},
		&models.Message{},
		&models.Notification{},
		&models.EmailTemplate{},
		&models.FileUpload{},
		&models.Subscription{},
		&models.PromoCode{},
		&models.Report{},
		&models.AnalyticsEvent{},
		&models.WebhookEvent{},
		&models.AuditLog{},
		&models.SystemSetting{},
		&models.TenantUsageTracking{},
		&models.DataExportRequest{},
		&models.APIKey{},
		&models.SDKClient{},
		&models.WhiteLabel{},
	)
}

// MockCache is a mock implementation of repository.Cache
type MockCache struct {
	data map[string]any
}

// NewMockCache creates a new mock cache
func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]any),
	}
}

func (m *MockCache) GetJSON(ctx context.Context, key string, dest any) error {
	val, exists := m.data[key]
	if !exists {
		return fmt.Errorf("key not found")
	}
	// Simple type assertion for testing
	switch v := dest.(type) {
	case *string:
		*v = val.(string)
	case *int:
		*v = val.(int)
	default:
		// For complex types, you'd need proper JSON marshaling
	}
	return nil
}

func (m *MockCache) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *MockCache) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

func (m *MockCache) DeletePattern(ctx context.Context, pattern string) error {
	// Simple pattern matching for testing
	for key := range m.data {
		delete(m.data, key)
	}
	return nil
}

func (m *MockCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	count := int64(0)
	for _, key := range keys {
		if _, exists := m.data[key]; exists {
			count++
		}
	}
	return count, nil
}

// MockMetrics is a mock implementation of repository.MetricsCollector
type MockMetrics struct{}

func NewMockMetrics() *MockMetrics {
	return &MockMetrics{}
}

func (m *MockMetrics) RecordOperation(operation string, table string, duration time.Duration, err error) {
}

func (m *MockMetrics) RecordCacheHit(table string) {
}

func (m *MockMetrics) RecordCacheMiss(table string) {
}

func (m *MockMetrics) RecordQueryCount(table string, count int64) {
}

// MockAuditLogger is a mock implementation of repository.AuditLogger
type MockAuditLogger struct {
	logs []AuditEntry
}

type AuditEntry struct {
	UserID    uuid.UUID
	TenantID  *uuid.UUID
	Action    string
	Entity    string
	EntityID  uuid.UUID
	Changes   map[string]any
	Timestamp time.Time
}

func NewMockAuditLogger() *MockAuditLogger {
	return &MockAuditLogger{
		logs: make([]AuditEntry, 0),
	}
}

func (m *MockAuditLogger) LogCreate(ctx context.Context, entityType string, entityID uuid.UUID, entity any) {
	m.logs = append(m.logs, AuditEntry{
		Action:    "create",
		Entity:    entityType,
		EntityID:  entityID,
		Timestamp: time.Now().UTC(),
	})
}

func (m *MockAuditLogger) LogUpdate(ctx context.Context, entityType string, entityID uuid.UUID, oldValues, newValues map[string]any) {
	m.logs = append(m.logs, AuditEntry{
		Action:    "update",
		Entity:    entityType,
		EntityID:  entityID,
		Changes:   newValues,
		Timestamp: time.Now().UTC(),
	})
}

func (m *MockAuditLogger) LogDelete(ctx context.Context, entityType string, entityID uuid.UUID) {
	m.logs = append(m.logs, AuditEntry{
		Action:    "delete",
		Entity:    entityType,
		EntityID:  entityID,
		Timestamp: time.Now().UTC(),
	})
}

func (m *MockAuditLogger) LogAction(ctx context.Context, action models.AuditAction, entityType string, entityID uuid.UUID, description string, metadata map[string]any) {
	m.logs = append(m.logs, AuditEntry{
		Action:    string(action),
		Entity:    entityType,
		EntityID:  entityID,
		Changes:   metadata,
		Timestamp: time.Now().UTC(),
	})
}

func (m *MockAuditLogger) GetLogs() []AuditEntry {
	return m.logs
}

// MockLogger is a mock implementation of log.AllLogger
type MockLogger struct{}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Trace(v ...any)                          {}
func (m *MockLogger) Debug(v ...any)                          {}
func (m *MockLogger) Info(v ...any)                           {}
func (m *MockLogger) Warn(v ...any)                           {}
func (m *MockLogger) Error(v ...any)                          {}
func (m *MockLogger) Fatal(v ...any)                          {}
func (m *MockLogger) Panic(v ...any)                          {}
func (m *MockLogger) Tracef(format string, v ...any)          {}
func (m *MockLogger) Debugf(format string, v ...any)          {}
func (m *MockLogger) Infof(format string, v ...any)           {}
func (m *MockLogger) Warnf(format string, v ...any)           {}
func (m *MockLogger) Errorf(format string, v ...any)          {}
func (m *MockLogger) Fatalf(format string, v ...any)          {}
func (m *MockLogger) Panicf(format string, v ...any)          {}
func (m *MockLogger) Tracew(msg string, keysAndValues ...any) {}
func (m *MockLogger) Debugw(msg string, keysAndValues ...any) {}
func (m *MockLogger) Infow(msg string, keysAndValues ...any)  {}
func (m *MockLogger) Warnw(msg string, keysAndValues ...any)  {}
func (m *MockLogger) Errorw(msg string, keysAndValues ...any) {}
func (m *MockLogger) Fatalw(msg string, keysAndValues ...any) {}
func (m *MockLogger) Panicw(msg string, keysAndValues ...any) {}
func (m *MockLogger) SetLevel(level log.Level)                {}
func (m *MockLogger) SetOutput(writer io.Writer)              {}
func (m *MockLogger) WithContext(ctx context.Context) log.CommonLogger {
	return m
}

// DefaultRepositoryConfig creates a default repository configuration for testing
func DefaultRepositoryConfig() repository.RepositoryConfig {
	return repository.RepositoryConfig{
		Logger:      NewMockLogger(),
		AuditLogger: NewMockAuditLogger(),
		Cache:       NewMockCache(),
		Metrics:     NewMockMetrics(),
		CacheTTL:    5 * time.Minute,
	}
}
