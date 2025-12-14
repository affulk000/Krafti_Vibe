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

	// Logto authentication configuration
	Auth LogtoConfig

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

// LogtoConfig holds Logto connection configuration
type LogtoConfig struct {
	// Core Logto configuration
	Endpoint    string `json:"endpoint"`
	Issuer      string `json:"issuer"`
	JWKSURI     string `json:"jwks_uri"`
	TokenURI    string `json:"token_uri"`
	UserInfoURI string `json:"userinfo_uri"`

	// Application credentials
	M2MAppID     string `json:"m2m_app_id"`
	M2MAppSecret string `json:"m2m_app_secret"`
	WebAppID     string `json:"web_app_id"`
	WebAppSecret string `json:"web_app_secret"`

	// API Resources
	APIResourceIndicator string `json:"api_resource_indicator"`
	AdminAPIResource     string `json:"admin_api_resource"`
	ArtisanAPIResource   string `json:"artisan_api_resource"`
	CustomerAPIResource  string `json:"customer_api_resource"`

	// Cache and performance settings
	JWKSCacheTTL       int `json:"jwks_cache_ttl"`
	JWKSRefreshWindow  int `json:"jwks_refresh_window"`
	ClockSkewTolerance int `json:"clock_skew_tolerance"`

	// Feature flags
	EnableOrganizations     bool `json:"enable_organizations"`
	EnableM2M               bool `json:"enable_m2m"`
	EnableRBAC              bool `json:"enable_rbac"`
	ValidateAudience        bool `json:"validate_audience"`
	ValidateIssuer          bool `json:"validate_issuer"`
	EnableLogging           bool `json:"enable_logging"`
	EnableBackgroundRefresh bool `json:"enable_background_refresh"`

	// Security settings
	RequiredScopes   []string `json:"required_scopes"`
	TrustedAudiences []string `json:"trusted_audiences"`
	AllowedOrigins   []string `json:"allowed_origins"`

	// Webhook settings
	WebhookSigningSecret string `json:"webhook_signing_secret"`
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

	// Legacy JWT cache TTL (replaced by LOGTO_JWKS_CACHE_TTL)
	_ = getEnv("JWT_CACHE_TTL", "3600") // Deprecated, use LOGTO_JWKS_CACHE_TTL

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
		Auth: LogtoConfig{
			// Core Logto URLs
			Endpoint:    getEnv("LOGTO_ENDPOINT", "https://gfxldq.logto.app/"),
			Issuer:      getEnv("LOGTO_ISSUER", "https://gfxldq.logto.app/oidc"),
			JWKSURI:     getEnv("LOGTO_JWKS_URI", "https://gfxldq.logto.app/oidc/jwks"),
			TokenURI:    getEnv("LOGTO_TOKEN_URI", "https://gfxldq.logto.app/oidc/token"),
			UserInfoURI: getEnv("LOGTO_USERINFO_URI", "https://gfxldq.logto.app/oidc/me"),

			// Application credentials
			M2MAppID:     getEnv("LOGTO_M2M_APP_ID", ""),
			M2MAppSecret: getEnv("LOGTO_M2M_APP_SECRET", ""),
			WebAppID:     getEnv("LOGTO_WEB_APP_ID", ""),
			WebAppSecret: getEnv("LOGTO_WEB_APP_SECRET", ""),

			// API Resources
			APIResourceIndicator: getEnv("LOGTO_API_RESOURCE", "https://api.kraftivibe.com"),
			AdminAPIResource:     getEnv("LOGTO_ADMIN_API_RESOURCE", "https://api.kraftivibe.com/admin"),
			ArtisanAPIResource:   getEnv("LOGTO_ARTISAN_API_RESOURCE", "https://api.kraftivibe.com/artisan"),
			CustomerAPIResource:  getEnv("LOGTO_CUSTOMER_API_RESOURCE", "https://api.kraftivibe.com/customer"),

			// Cache settings
			JWKSCacheTTL:       getIntEnv("LOGTO_JWKS_CACHE_TTL", 900),       // 15 minutes
			JWKSRefreshWindow:  getIntEnv("LOGTO_JWKS_REFRESH_WINDOW", 60),   // 1 minute
			ClockSkewTolerance: getIntEnv("LOGTO_CLOCK_SKEW_TOLERANCE", 300), // 5 minutes

			// Feature flags
			EnableOrganizations:     getBoolEnv("LOGTO_ENABLE_ORGANIZATIONS", true),
			EnableM2M:               getBoolEnv("LOGTO_ENABLE_M2M", true),
			EnableRBAC:              getBoolEnv("LOGTO_ENABLE_RBAC", true),
			ValidateAudience:        getBoolEnv("LOGTO_VALIDATE_AUDIENCE", true),
			ValidateIssuer:          getBoolEnv("LOGTO_VALIDATE_ISSUER", true),
			EnableLogging:           getBoolEnv("LOGTO_ENABLE_LOGGING", true),
			EnableBackgroundRefresh: getBoolEnv("LOGTO_ENABLE_BACKGROUND_REFRESH", true),

			// Security settings
			RequiredScopes:   getStringSliceEnv("LOGTO_REQUIRED_SCOPES", []string{}),
			TrustedAudiences: getStringSliceEnv("LOGTO_TRUSTED_AUDIENCES", []string{"https://api.kraftivibe.com"}),
			AllowedOrigins:   getStringSliceEnv("LOGTO_ALLOWED_ORIGINS", []string{"https://kraftivibe.com", "https://app.kraftivibe.com"}),

			// Webhook settings
			WebhookSigningSecret: getEnv("LOGTO_WEBHOOK_SECRET", ""),
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
	// Validate Logto configuration
	if err := c.Auth.Validate(); err != nil {
		return fmt.Errorf("logto configuration: %w", err)
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

// Validate validates the Logto configuration
func (lc *LogtoConfig) Validate() error {
	if lc.Endpoint == "" {
		return fmt.Errorf("LOGTO_ENDPOINT is required")
	}
	if lc.JWKSURI == "" {
		return fmt.Errorf("LOGTO_JWKS_URI is required")
	}
	if lc.Issuer == "" {
		return fmt.Errorf("LOGTO_ISSUER is required")
	}
	if lc.APIResourceIndicator == "" {
		return fmt.Errorf("LOGTO_API_RESOURCE is required")
	}

	// Validate M2M configuration if M2M is enabled
	if lc.EnableM2M {
		if lc.M2MAppID == "" {
			return fmt.Errorf("LOGTO_M2M_APP_ID is required when M2M is enabled")
		}
		if lc.M2MAppSecret == "" {
			return fmt.Errorf("LOGTO_M2M_APP_SECRET is required when M2M is enabled")
		}
	}

	// Validate cache TTL values
	if lc.JWKSCacheTTL <= 0 {
		return fmt.Errorf("LOGTO_JWKS_CACHE_TTL must be greater than 0")
	}
	if lc.JWKSRefreshWindow <= 0 {
		return fmt.Errorf("LOGTO_JWKS_REFRESH_WINDOW must be greater than 0")
	}
	if lc.ClockSkewTolerance < 0 {
		return fmt.Errorf("LOGTO_CLOCK_SKEW_TOLERANCE cannot be negative")
	}

	return nil
}

// IsM2MEnabled returns true if M2M authentication is enabled
func (lc *LogtoConfig) IsM2MEnabled() bool {
	return lc.EnableM2M && lc.M2MAppID != "" && lc.M2MAppSecret != ""
}

// IsOrganizationEnabled returns true if organization support is enabled
func (lc *LogtoConfig) IsOrganizationEnabled() bool {
	return lc.EnableOrganizations
}

// GetCacheTTLDuration returns the JWKS cache TTL as time.Duration
func (lc *LogtoConfig) GetCacheTTLDuration() time.Duration {
	return time.Duration(lc.JWKSCacheTTL) * time.Second
}

// GetRefreshWindowDuration returns the JWKS refresh window as time.Duration
func (lc *LogtoConfig) GetRefreshWindowDuration() time.Duration {
	return time.Duration(lc.JWKSRefreshWindow) * time.Second
}

// GetClockSkewToleranceDuration returns the clock skew tolerance as time.Duration
func (lc *LogtoConfig) GetClockSkewToleranceDuration() time.Duration {
	return time.Duration(lc.ClockSkewTolerance) * time.Second
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
