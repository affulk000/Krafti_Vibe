package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QueryBuilder helps build complex queries
type QueryBuilder struct {
	db *gorm.DB
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
	return &QueryBuilder{db: db}
}

// WhereUUID adds a UUID where clause
func (qb *QueryBuilder) WhereUUID(field string, value uuid.UUID) *QueryBuilder {
	if value != uuid.Nil {
		qb.db = qb.db.Where(fmt.Sprintf("%s = ?", field), value)
	}
	return qb
}

// WhereUUIDPtr adds a UUID pointer where clause
func (qb *QueryBuilder) WhereUUIDPtr(field string, value *uuid.UUID) *QueryBuilder {
	if value != nil && *value != uuid.Nil {
		qb.db = qb.db.Where(fmt.Sprintf("%s = ?", field), *value)
	}
	return qb
}

// WhereString adds a string where clause
func (qb *QueryBuilder) WhereString(field string, value string) *QueryBuilder {
	if value != "" {
		qb.db = qb.db.Where(fmt.Sprintf("%s = ?", field), value)
	}
	return qb
}

// WhereStringLike adds a LIKE where clause
func (qb *QueryBuilder) WhereStringLike(field string, value string) *QueryBuilder {
	if value != "" {
		qb.db = qb.db.Where(fmt.Sprintf("%s ILIKE ?", field), "%"+value+"%")
	}
	return qb
}

// WhereIn adds an IN clause
func (qb *QueryBuilder) WhereIn(field string, values []uuid.UUID) *QueryBuilder {
	if len(values) > 0 {
		qb.db = qb.db.Where(fmt.Sprintf("%s IN ?", field), values)
	}
	return qb
}

// WhereDateRange adds a date range where clause
func (qb *QueryBuilder) WhereDateRange(field string, start, end *time.Time) *QueryBuilder {
	if start != nil {
		qb.db = qb.db.Where(fmt.Sprintf("%s >= ?", field), *start)
	}
	if end != nil {
		qb.db = qb.db.Where(fmt.Sprintf("%s <= ?", field), *end)
	}
	return qb
}

// WhereStatus adds a status where clause
func (qb *QueryBuilder) WhereStatus(field string, status string) *QueryBuilder {
	if status != "" {
		qb.db = qb.db.Where(fmt.Sprintf("%s = ?", field), status)
	}
	return qb
}

// WhereTenant adds tenant isolation
func (qb *QueryBuilder) WhereTenant(tenantID *uuid.UUID, allowPlatform bool) *QueryBuilder {
	if tenantID != nil && *tenantID != uuid.Nil {
		qb.db = qb.db.Where("tenant_id = ?", *tenantID)
	} else if !allowPlatform {
		// If tenant is required but not provided, return no results
		qb.db = qb.db.Where("1 = 0")
	}
	return qb
}

// OrderBy adds ordering
func (qb *QueryBuilder) OrderBy(field string, direction string) *QueryBuilder {
	if field != "" {
		if strings.ToUpper(direction) != "ASC" && strings.ToUpper(direction) != "DESC" {
			direction = "ASC"
		}
		qb.db = qb.db.Order(fmt.Sprintf("%s %s", field, direction))
	}
	return qb
}

// OrderByCreatedAt orders by created_at descending
func (qb *QueryBuilder) OrderByCreatedAt() *QueryBuilder {
	return qb.OrderBy("created_at", "DESC")
}

// Preload adds preload associations
func (qb *QueryBuilder) Preload(associations ...string) *QueryBuilder {
	for _, assoc := range associations {
		if assoc != "" {
			qb.db = qb.db.Preload(assoc)
		}
	}
	return qb
}

// Build returns the final GORM query
func (qb *QueryBuilder) Build() *gorm.DB {
	return qb.db
}

