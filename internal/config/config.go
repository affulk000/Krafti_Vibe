package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	Zitadel     ZitadelConfig
	App         AppConfig
	Environment string
}

func (c *Config) IsProduction() bool { return c.Environment == "production" }
func (c *Config) IsDevelopment() bool { return c.Environment == "development" }

type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	SSLHost         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type RedisConfig struct {
	Host            string
	Port            string
	Password        string
	DB              int
	MaxRetries      int
	PoolSize        int
	MinIdleConns    int
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolTimeout     time.Duration
	ConnMaxIdleTime time.Duration
}

type ZitadelConfig struct {
	Domain  string
	KeyPath string
}

type AppConfig struct {
	Name           string
	Version        string
	LogLevel       string
	CORSOrigins    []string
	EnableMetrics  bool
	EnableTracing  bool
	RateLimitRPS   int
	RequestTimeout time.Duration
}

var globalConfig *Config

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	// A .env file is optional; production deployments should provide environment variables directly.
	_ = godotenv.Load()

	cfg := &Config{
		Environment: getEnv("ENV", "development"),
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"), Port: getEnv("PORT", "3000"),
			ReadTimeout: getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second), WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout: getDurationEnv("SERVER_IDLE_TIMEOUT", 120*time.Second), ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"), User: getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""), DBName: getEnv("DB_NAME", "krafti_vibe"), SSLMode: getEnv("DB_SSLMODE", "disable"), SSLHost: getEnv("DB_SSL_HOST", ""),
			MaxOpenConns: getIntEnv("DB_MAX_OPEN_CONNS", 25), MaxIdleConns: getIntEnv("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute), ConnMaxIdleTime: getDurationEnv("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"), Port: getEnv("REDIS_PORT", "6379"), Password: getEnv("REDIS_PASSWORD", ""), DB: getIntEnv("REDIS_DB", 0),
			MaxRetries: getIntEnv("REDIS_MAX_RETRIES", 3), PoolSize: getIntEnv("REDIS_POOL_SIZE", 10), MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 5),
			DialTimeout: getDurationEnv("REDIS_DIAL_TIMEOUT", 5*time.Second), ReadTimeout: getDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second), WriteTimeout: getDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second),
			PoolTimeout: getDurationEnv("REDIS_POOL_TIMEOUT", 4*time.Second), ConnMaxIdleTime: getDurationEnv("REDIS_CONN_MAX_IDLE_TIME", 30*time.Minute),
		},
		Zitadel: ZitadelConfig{Domain: getEnv("ZITADEL_DOMAIN", ""), KeyPath: getEnv("ZITADEL_KEY_PATH", "")},
		App: AppConfig{
			Name: getEnv("APP_NAME", "Krafti Vibe API"), Version: getEnv("APP_VERSION", "1.0.0"), LogLevel: getEnv("LOG_LEVEL", "info"),
			CORSOrigins: getStringSliceEnv("CORS_ORIGINS", []string{"*"}), EnableMetrics: getBoolEnv("ENABLE_METRICS", true), EnableTracing: getBoolEnv("ENABLE_TRACING", false),
			RateLimitRPS: getIntEnv("RATE_LIMIT_RPS", 100), RequestTimeout: getDurationEnv("REQUEST_TIMEOUT", 30*time.Second),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	globalConfig = cfg
	return cfg, nil
}

func Get() *Config {
	if globalConfig == nil { panic("configuration not loaded - call config.Load() first") }
	return globalConfig
}

// Validate rejects unsafe or internally inconsistent production configuration instead of allowing insecure defaults.
func (c *Config) Validate() error {
	if c == nil { return fmt.Errorf("configuration is nil") }
	if c.Environment != "development" && c.Environment != "test" && c.Environment != "staging" && c.Environment != "production" {
		return fmt.Errorf("invalid environment: %s", c.Environment)
	}
	if c.Database.Host == "" || c.Database.DBName == "" || c.Database.User == "" { return fmt.Errorf("database host, user, and name are required") }
	if c.Database.Password == "" && !c.IsDevelopment() && c.Environment != "test" { return fmt.Errorf("database password is required outside development") }
	if c.IsProduction() && c.Database.SSLMode == "disable" { return fmt.Errorf("DB_SSLMODE=disable is not allowed in production") }
	if c.Redis.Host == "" || c.Redis.Port == "" { return fmt.Errorf("redis host and port are required") }
	if c.IsProduction() && c.Redis.Password == "" { return fmt.Errorf("redis password is required in production") }
	if c.IsProduction() && len(c.App.CORSOrigins) == 0 { return fmt.Errorf("at least one CORS origin is required in production") }
	if c.IsProduction() {
		for _, origin := range c.App.CORSOrigins { if strings.TrimSpace(origin) == "*" { return fmt.Errorf("wildcard CORS origin is not allowed in production") } }
		if c.Zitadel.Domain == "" { return fmt.Errorf("ZITADEL_DOMAIN is required in production") }
		if c.Zitadel.KeyPath == "" { return fmt.Errorf("ZITADEL_KEY_PATH is required in production") }
	}

	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[strings.ToLower(c.App.LogLevel)] { return fmt.Errorf("invalid log level: %s (must be: debug, info, warn, error)", c.App.LogLevel) }
	if port, err := strconv.Atoi(c.Server.Port); err != nil || port < 1 || port > 65535 { return fmt.Errorf("invalid server port: %s", c.Server.Port) }
	if c.Database.MaxOpenConns < 1 || c.Database.MaxIdleConns < 0 || c.Database.MaxIdleConns > c.Database.MaxOpenConns { return fmt.Errorf("invalid database connection pool settings") }
	if c.Redis.PoolSize < 1 || c.Redis.MinIdleConns < 0 || c.Redis.MinIdleConns > c.Redis.PoolSize { return fmt.Errorf("invalid redis connection pool settings") }
	if c.App.RateLimitRPS < 1 { return fmt.Errorf("RATE_LIMIT_RPS must be greater than zero") }
	if c.App.RequestTimeout <= 0 || c.Server.ReadTimeout <= 0 || c.Server.WriteTimeout <= 0 || c.Server.IdleTimeout <= 0 || c.Server.ShutdownTimeout <= 0 { return fmt.Errorf("server and request timeouts must be positive") }
	if c.Database.SSLHost != "" && net.ParseIP(c.Database.Host) != nil && c.Database.SSLMode == "disable" { return fmt.Errorf("DB_SSL_HOST cannot be used while DB_SSLMODE=disable") }
	return nil
}

func (c *Config) DatabaseURL() string { return os.Getenv("DATABASE_URL") }

func (c *Config) DatabaseDSN() string {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.DBName, c.Database.SSLMode)
	dsn += " connect_timeout=10"
	if c.Database.SSLHost != "" && c.Database.SSLMode != "disable" { dsn += fmt.Sprintf(" sslhost=%s", c.Database.SSLHost) }
	return dsn
}

func (c *Config) RedisAddr() string { return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port) }
func (c *Config) ServerAddr() string { return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port) }

func getEnv(key, defaultValue string) string { if value := os.Getenv(key); value != "" { return value }; return defaultValue }
func getIntEnv(key string, defaultValue int) int { if value := os.Getenv(key); value != "" { if v, err := strconv.Atoi(value); err == nil { return v } }; return defaultValue }
func getBoolEnv(key string, defaultValue bool) bool { if value := os.Getenv(key); value != "" { if v, err := strconv.ParseBool(value); err == nil { return v } }; return defaultValue }
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil { return duration }
		if seconds, err := strconv.Atoi(value); err == nil { return time.Duration(seconds) * time.Second }
	}
	return defaultValue
}
func getStringSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		if value == "*" { return []string{"*"} }
		parts := strings.Split(value, ",")
		origins := make([]string, 0, len(parts))
		for _, part := range parts { if origin := strings.TrimSpace(part); origin != "" { origins = append(origins, origin) } }
		return origins
	}
	return defaultValue
}
