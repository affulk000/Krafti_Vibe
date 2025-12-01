package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLogger interface for audit logging
type AuditLogger interface {
	LogCreate(ctx context.Context, entityType string, entityID uuid.UUID, entity any)
	LogUpdate(ctx context.Context, entityType string, entityID uuid.UUID, oldValues, newValues map[string]any)
	LogDelete(ctx context.Context, entityType string, entityID uuid.UUID)
	LogAction(ctx context.Context, action models.AuditAction, entityType string, entityID uuid.UUID, description string, metadata map[string]any)
}

// DatabaseAuditLogger implements AuditLogger using the database
type DatabaseAuditLogger struct {
	db     *gorm.DB
	logger interface {
		Error(msg string, keysAndValues ...any)
		Warn(msg string, keysAndValues ...any)
	}
}

// NewDatabaseAuditLogger creates a new database audit logger
func NewDatabaseAuditLogger(db *gorm.DB, logger interface {
	Error(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
}) AuditLogger {
	return &DatabaseAuditLogger{
		db:     db,
		logger: logger,
	}
}

// LogCreate logs a create action
func (l *DatabaseAuditLogger) LogCreate(ctx context.Context, entityType string, entityID uuid.UUID, entity any) {
	l.logAction(ctx, models.AuditActionCreate, entityType, entityID, fmt.Sprintf("Created %s", entityType), nil, toMap(entity), nil)
}

// LogUpdate logs an update action
func (l *DatabaseAuditLogger) LogUpdate(ctx context.Context, entityType string, entityID uuid.UUID, oldValues, newValues map[string]any) {
	description := fmt.Sprintf("Updated %s", entityType)
	l.logAction(ctx, models.AuditActionUpdate, entityType, entityID, description, nil, newValues, oldValues)
}

// LogDelete logs a delete action
func (l *DatabaseAuditLogger) LogDelete(ctx context.Context, entityType string, entityID uuid.UUID) {
	l.logAction(ctx, models.AuditActionDelete, entityType, entityID, fmt.Sprintf("Deleted %s", entityType), nil, nil, nil)
}

// LogAction logs a custom action
func (l *DatabaseAuditLogger) LogAction(ctx context.Context, action models.AuditAction, entityType string, entityID uuid.UUID, description string, metadata map[string]any) {
	l.logAction(ctx, action, entityType, entityID, description, metadata, nil, nil)
}

// logAction is the internal method that writes to the database
func (l *DatabaseAuditLogger) logAction(ctx context.Context, action models.AuditAction, entityType string, entityID uuid.UUID, description string, metadata map[string]any, newValues map[string]any, oldValues map[string]any) {
	// Extract user info from context if available
	var userID *uuid.UUID
	var userEmail string
	var userRole models.UserRole
	var tenantID *uuid.UUID
	var ipAddress string
	var userAgent string

	// Try to get user info from context (this would be set by auth middleware)
	if userIDVal := ctx.Value("user_id"); userIDVal != nil {
		if id, ok := userIDVal.(uuid.UUID); ok {
			userID = &id
		}
	}
	if emailVal := ctx.Value("user_email"); emailVal != nil {
		if email, ok := emailVal.(string); ok {
			userEmail = email
		}
	}
	if roleVal := ctx.Value("user_role"); roleVal != nil {
		if role, ok := roleVal.(models.UserRole); ok {
			userRole = role
		}
	}
	if tenantIDVal := ctx.Value("tenant_id"); tenantIDVal != nil {
		if id, ok := tenantIDVal.(uuid.UUID); ok {
			tenantID = &id
		}
	}
	if ipVal := ctx.Value("ip_address"); ipVal != nil {
		if ip, ok := ipVal.(string); ok {
			ipAddress = ip
		}
	}
	if uaVal := ctx.Value("user_agent"); uaVal != nil {
		if ua, ok := uaVal.(string); ok {
			userAgent = ua
		}
	}

	// Convert maps to JSONB
	var oldValuesJSONB models.JSONB
	var newValuesJSONB models.JSONB
	var metadataJSONB models.JSONB

	if oldValues != nil {
		oldValuesJSONB = models.JSONB(oldValues)
	}
	if newValues != nil {
		newValuesJSONB = models.JSONB(newValues)
	}
	if metadata != nil {
		metadataJSONB = models.JSONB(metadata)
	}

	auditLog := models.AuditLog{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		TenantID:    tenantID,
		UserID:      userID,
		UserEmail:   userEmail,
		UserRole:    userRole,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
		OldValues:   oldValuesJSONB,
		NewValues:   newValuesJSONB,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Metadata:    metadataJSONB,
	}

	// Write to database asynchronously to avoid blocking
	go func() {
		if err := l.db.WithContext(context.Background()).Create(&auditLog).Error; err != nil {
			if l.logger != nil {
				l.logger.Error("failed to write audit log", "error", err, "entity_type", entityType, "entity_id", entityID)
			}
		}
	}()
}

// NoOpAuditLogger is a no-op audit logger implementation for testing
type NoOpAuditLogger struct{}

func (l *NoOpAuditLogger) LogCreate(ctx context.Context, entityType string, entityID uuid.UUID, entity any) {
}
func (l *NoOpAuditLogger) LogUpdate(ctx context.Context, entityType string, entityID uuid.UUID, oldValues, newValues map[string]any) {
}
func (l *NoOpAuditLogger) LogDelete(ctx context.Context, entityType string, entityID uuid.UUID) {}
func (l *NoOpAuditLogger) LogAction(ctx context.Context, action models.AuditAction, entityType string, entityID uuid.UUID, description string, metadata map[string]any) {
}

// DefaultAuditLogger returns a no-op audit logger (for testing or when audit is disabled)
func DefaultAuditLogger() AuditLogger {
	return &NoOpAuditLogger{}
}

// helper
func toMap(v any) map[string]any {
	if v == nil {
		return nil
	}

	b, _ := json.Marshal(v)
	var m map[string]any
	json.Unmarshal(b, &m)
	return m
}
