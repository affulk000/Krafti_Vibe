-- ============================================
-- Krafti Vibe - Database Initialization Script
-- ============================================
-- This script initializes the PostgreSQL database
-- with required extensions and configurations

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enable PostGIS for geospatial queries (artisan location search)
CREATE EXTENSION IF NOT EXISTS "postgis";

-- Enable pg_trgm for fuzzy text search
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Enable unaccent for accent-insensitive search
CREATE EXTENSION IF NOT EXISTS "unaccent";

-- ============================================
-- Create Custom Types
-- ============================================

-- User Role Type
DO $$ BEGIN
    CREATE TYPE user_role AS ENUM (
        'platform_super_admin',
        'platform_admin',
        'platform_support',
        'tenant_owner',
        'tenant_admin',
        'artisan',
        'team_member',
        'customer'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- User Status Type
DO $$ BEGIN
    CREATE TYPE user_status AS ENUM (
        'active',
        'inactive',
        'suspended',
        'pending'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Tenant Plan Type
DO $$ BEGIN
    CREATE TYPE tenant_plan AS ENUM (
        'solo',
        'small',
        'corporation',
        'enterprise'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Tenant Status Type
DO $$ BEGIN
    CREATE TYPE tenant_status AS ENUM (
        'active',
        'suspended',
        'cancelled',
        'trial'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Booking Status Type
DO $$ BEGIN
    CREATE TYPE booking_status AS ENUM (
        'pending',
        'confirmed',
        'in_progress',
        'completed',
        'cancelled',
        'no_show'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Payment Status Type
DO $$ BEGIN
    CREATE TYPE payment_status AS ENUM (
        'pending',
        'deposit_paid',
        'paid',
        'refunded',
        'failed'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- ============================================
-- Create Schemas
-- ============================================

CREATE SCHEMA IF NOT EXISTS audit;
CREATE SCHEMA IF NOT EXISTS reports;

-- ============================================
-- Create Functions
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Function to set tenant context for RLS
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_uuid UUID)
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.tenant_id', tenant_uuid::text, false);
END;
$$ LANGUAGE plpgsql;

-- Function to get current tenant from context
CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS UUID AS $$
BEGIN
    RETURN current_setting('app.tenant_id', true)::UUID;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- Function to check if user is super admin
CREATE OR REPLACE FUNCTION is_super_admin()
RETURNS BOOLEAN AS $$
BEGIN
    RETURN current_setting('app.is_super_admin', true)::BOOLEAN;
EXCEPTION
    WHEN OTHERS THEN
        RETURN false;
END;
$$ LANGUAGE plpgsql STABLE;

-- ============================================
-- Audit Log Function
-- ============================================

CREATE OR REPLACE FUNCTION audit.log_change()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO audit.audit_logs (
            table_name,
            operation,
            old_data,
            user_id,
            tenant_id
        ) VALUES (
            TG_TABLE_NAME,
            TG_OP,
            row_to_json(OLD),
            current_setting('app.user_id', true)::UUID,
            current_setting('app.tenant_id', true)::UUID
        );
        RETURN OLD;
    ELSIF (TG_OP = 'UPDATE') THEN
        INSERT INTO audit.audit_logs (
            table_name,
            operation,
            old_data,
            new_data,
            user_id,
            tenant_id
        ) VALUES (
            TG_TABLE_NAME,
            TG_OP,
            row_to_json(OLD),
            row_to_json(NEW),
            current_setting('app.user_id', true)::UUID,
            current_setting('app.tenant_id', true)::UUID
        );
        RETURN NEW;
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO audit.audit_logs (
            table_name,
            operation,
            new_data,
            user_id,
            tenant_id
        ) VALUES (
            TG_TABLE_NAME,
            TG_OP,
            row_to_json(NEW),
            current_setting('app.user_id', true)::UUID,
            current_setting('app.tenant_id', true)::UUID
        );
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Create Audit Log Table
-- ============================================

CREATE TABLE IF NOT EXISTS audit.audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name TEXT NOT NULL,
    operation TEXT NOT NULL,
    old_data JSONB,
    new_data JSONB,
    user_id UUID,
    tenant_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_table_name ON audit.audit_logs(table_name);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit.audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit.audit_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit.audit_logs(user_id);

-- ============================================
-- Create Full-Text Search Configuration
-- ============================================

-- Custom text search configuration for better search
CREATE TEXT SEARCH CONFIGURATION IF NOT EXISTS kraftivibe (COPY = pg_catalog.english);

-- ============================================
-- Grant Permissions
-- ============================================

-- Grant usage on extensions
GRANT USAGE ON SCHEMA public TO PUBLIC;
GRANT USAGE ON SCHEMA audit TO PUBLIC;
GRANT USAGE ON SCHEMA reports TO PUBLIC;

-- ============================================
-- Performance Settings
-- ============================================

-- Set recommended PostgreSQL settings for SaaS workload
-- These can be overridden in postgresql.conf

-- Connection pooling settings (commented - set in postgresql.conf)
-- ALTER SYSTEM SET max_connections = 200;
-- ALTER SYSTEM SET shared_buffers = '256MB';
-- ALTER SYSTEM SET effective_cache_size = '1GB';
-- ALTER SYSTEM SET maintenance_work_mem = '64MB';
-- ALTER SYSTEM SET checkpoint_completion_target = 0.9;
-- ALTER SYSTEM SET wal_buffers = '16MB';
-- ALTER SYSTEM SET default_statistics_target = 100;
-- ALTER SYSTEM SET random_page_cost = 1.1;
-- ALTER SYSTEM SET effective_io_concurrency = 200;
-- ALTER SYSTEM SET work_mem = '4MB';
-- ALTER SYSTEM SET min_wal_size = '1GB';
-- ALTER SYSTEM SET max_wal_size = '4GB';

-- ============================================
-- Create Materialized Views for Analytics
-- ============================================

-- Tenant statistics view (will be created by migrations)
-- CREATE MATERIALIZED VIEW IF NOT EXISTS reports.tenant_stats AS ...

-- ============================================
-- Comments
-- ============================================

COMMENT ON SCHEMA audit IS 'Schema for audit logging and compliance';
COMMENT ON SCHEMA reports IS 'Schema for materialized views and reports';
COMMENT ON FUNCTION update_updated_at_column() IS 'Automatically updates updated_at timestamp';
COMMENT ON FUNCTION set_tenant_context(UUID) IS 'Sets tenant context for Row-Level Security';
COMMENT ON FUNCTION current_tenant_id() IS 'Returns current tenant ID from session';
COMMENT ON FUNCTION is_super_admin() IS 'Checks if current user is super admin';

-- ============================================
-- Success Message
-- ============================================

DO $$
BEGIN
    RAISE NOTICE '✓ Database initialization complete!';
    RAISE NOTICE '✓ Extensions enabled: uuid-ossp, pgcrypto, postgis, pg_trgm, unaccent';
    RAISE NOTICE '✓ Custom types created';
    RAISE NOTICE '✓ Helper functions created';
    RAISE NOTICE '✓ Audit logging configured';
    RAISE NOTICE '';
    RAISE NOTICE 'Next steps:';
    RAISE NOTICE '1. Run database migrations: make migrate-up';
    RAISE NOTICE '2. Seed initial data: make db-seed';
    RAISE NOTICE '3. Start the application: make run';
END $$;
