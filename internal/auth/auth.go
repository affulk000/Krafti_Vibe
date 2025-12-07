package auth

import (
	"Krafti_Vibe/internal/config"
	"Krafti_Vibe/internal/infrastructure/logto"
	"Krafti_Vibe/internal/service"
	"fmt"

	"github.com/gofiber/fiber/v2/log"
)

// AuthService represents the authentication service
type AuthService struct {
	logtoService *service.Service
	config       *config.LogtoConfig
}

// NewAuthService creates a new authentication service
func NewAuthService(cfg *config.Config) (*AuthService, error) {
	// Convert config to logto config
	logtoConfig, err := convertConfig(cfg.Auth)
	if err != nil {
		return nil, fmt.Errorf("failed to convert auth config: %w", err)
	}

	// Create Logto service
	logtoService, err := service.NewService(logtoConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Logto service: %w", err)
	}

	log.Info("Authentication service initialized successfully")

	return &AuthService{
		logtoService: logtoService,
		config:       &cfg.Auth,
	}, nil
}

// GetLogtoService returns the Logto service
func (a *AuthService) GetLogtoService() *service.Service {
	return a.logtoService
}

// GetConfig returns the auth configuration
func (a *AuthService) GetConfig() *config.LogtoConfig {
	return a.config
}

// Close closes the auth service
func (a *AuthService) Close() error {
	if a.logtoService != nil {
		return a.logtoService.Close()
	}
	return nil
}

// convertConfig converts the application config to Logto config
func convertConfig(cfg config.LogtoConfig) (*logto.Config, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("Logto endpoint is required")
	}

	// Create default config
	logtoConfig := logto.DefaultConfig()

	// Core configuration
	logtoConfig.Endpoint = cfg.Endpoint
	logtoConfig.Issuer = cfg.Issuer
	logtoConfig.JWKSEndpoint = cfg.JWKSURI
	logtoConfig.TokenEndpoint = cfg.TokenURI
	logtoConfig.AppID = cfg.M2MAppID
	logtoConfig.AppSecret = cfg.M2MAppSecret

	// API Resource configuration
	logtoConfig.APIResourceIndicator = cfg.APIResourceIndicator

	// Cache settings
	logtoConfig.JWKSCacheTTL = cfg.GetCacheTTLDuration()
	logtoConfig.JWKSRefreshWindow = cfg.GetRefreshWindowDuration()
	logtoConfig.ClockSkewTolerance = cfg.GetClockSkewToleranceDuration()

	// Feature flags
	logtoConfig.EnableOrganizations = cfg.EnableOrganizations
	logtoConfig.EnableM2M = cfg.EnableM2M
	logtoConfig.EnableRBAC = cfg.EnableRBAC
	logtoConfig.ValidateAudience = cfg.ValidateAudience
	logtoConfig.ValidateIssuer = cfg.ValidateIssuer
	logtoConfig.EnableLogging = cfg.EnableLogging

	// Validate the converted config
	if err := logtoConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Logto configuration: %w", err)
	}

	return logtoConfig, nil
}

// ValidateConfig validates the authentication configuration
func ValidateConfig(cfg config.LogtoConfig) error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("LOGTO_ENDPOINT is required")
	}
	if cfg.JWKSURI == "" {
		return fmt.Errorf("LOGTO_JWKS_URI is required")
	}
	if cfg.Issuer == "" {
		return fmt.Errorf("LOGTO_ISSUER is required")
	}
	return nil
}