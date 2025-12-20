package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Server ServerConfig

	// Database configuration
	Database DatabaseConfig

	// Redis configuration
	Redis RedisConfig

	// Zitadel authentication configuration
	Zitadel ZitadelConfig

	// Application settings
	App AppConfig

	// Environment
	Environment string
}

// IsProduction returns true if environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	SSLHost         string // Optional: hostname for SSL certificate verification (when using IP for Host)
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// RedisConfig holds Redis connection configuration
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

// ZitadelConfig holds Zitadel authentication configuration
type ZitadelConfig struct {
	Domain  string
	KeyPath string
}

// AppConfig holds application-specific configuration
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

var (
	globalConfig *Config
)

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file first (if exists)
	_ = godotenv.Load()

	cfg := &Config{
		Environment: getEnv("ENV", "development"),
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnv("PORT", "3000"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getDurationEnv("SERVER_IDLE_TIMEOUT", 120*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "krafti_vibe"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			SSLHost:         getEnv("DB_SSL_HOST", ""), // Optional: for SSL cert verification with IP addresses
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getDurationEnv("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
		},
		Redis: RedisConfig{
			Host:            getEnv("REDIS_HOST", "localhost"),
			Port:            getEnv("REDIS_PORT", "6379"),
			Password:        getEnv("REDIS_PASSWORD", ""),
			DB:              getIntEnv("REDIS_DB", 0),
			MaxRetries:      getIntEnv("REDIS_MAX_RETRIES", 3),
			PoolSize:        getIntEnv("REDIS_POOL_SIZE", 10),
			MinIdleConns:    getIntEnv("REDIS_MIN_IDLE_CONNS", 5),
			DialTimeout:     getDurationEnv("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:     getDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout:    getDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second),
			PoolTimeout:     getDurationEnv("REDIS_POOL_TIMEOUT", 4*time.Second),
			ConnMaxIdleTime: getDurationEnv("REDIS_CONN_MAX_IDLE_TIME", 30*time.Minute),
		},
		Zitadel: ZitadelConfig{
			Domain:  getEnv("ZITADEL_DOMAIN", ""),
			KeyPath: getEnv("ZITADEL_KEY_PATH", ""),
		},
		App: AppConfig{
			Name:           getEnv("APP_NAME", "Krafti Vibe API"),
			Version:        getEnv("APP_VERSION", "1.0.0"),
			LogLevel:       getEnv("LOG_LEVEL", "info"),
			CORSOrigins:    getStringSliceEnv("CORS_ORIGINS", []string{"*"}),
			EnableMetrics:  getBoolEnv("ENABLE_METRICS", true),
			EnableTracing:  getBoolEnv("ENABLE_TRACING", false),
			RateLimitRPS:   getIntEnv("RATE_LIMIT_RPS", 100),
			RequestTimeout: getDurationEnv("REQUEST_TIMEOUT", 30*time.Second),
		},
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	globalConfig = cfg
	return cfg, nil
}

// Get returns the global configuration instance
func Get() *Config {
	if globalConfig == nil {
		panic("configuration not loaded - call config.Load() first")
	}
	return globalConfig
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[strings.ToLower(c.App.LogLevel)] {
		return fmt.Errorf("invalid log level: %s (must be: debug, info, warn, error)", c.App.LogLevel)
	}

	return nil
}

// DatabaseURL returns the DATABASE_URL environment variable if set
func (c *Config) DatabaseURL() string {
	return os.Getenv("DATABASE_URL")
}

// DatabaseDSN returns the database connection string
func (c *Config) DatabaseDSN() string {
	// Build base DSN
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)

	// Add connection timeout to prevent hanging
	dsn += " connect_timeout=10"

	// If using IP address with SSL, add the real hostname for certificate verification
	// This allows SSL to work properly when connecting via IP (to avoid IPv6 issues)
	if c.Database.SSLHost != "" && c.Database.SSLMode != "disable" {
		dsn += fmt.Sprintf(" sslhost=%s", c.Database.SSLHost)
	}

	return dsn
}

// RedisAddr returns the Redis address
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

// ServerAddr returns the server address
func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		// Try parsing as seconds
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
}

func getStringSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		if value == "*" {
			return []string{"*"}
		}
		return strings.Split(value, ",")
	}
	return defaultValue
}
