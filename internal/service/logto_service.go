package service

import (
	"Krafti_Vibe/internal/infrastructure/logto"
	"Krafti_Vibe/internal/middleware"
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// Service manages Logto authentication and authorization
type Service struct {
	config         *logto.Config
	jwksCache      *logto.JWKSCache
	tokenValidator *logto.TokenValidator
	scopes         *logto.Scopes
}

// NewService creates a new Logto service
func NewService(config *logto.Config) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize JWKS cache
	jwksCache, err := logto.NewJWKSCache(config.JWKSEndpoint, config.JWKSCacheTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWKS cache: %w", err)
	}

	// Initialize token validator
	tokenValidator := logto.NewTokenValidator(jwksCache, config.Issuer)

	service := &Service{
		config:         config,
		jwksCache:      jwksCache,
		tokenValidator: tokenValidator,
		scopes:         logto.DefaultScopes(),
	}

	// Log configuration
	config.LogConfig()

	log.Info("Logto service initialized successfully")

	return service, nil
}

// GetTokenValidator returns the token validator
func (s *Service) GetTokenValidator() *logto.TokenValidator {
	return s.tokenValidator
}

// GetConfig returns the configuration
func (s *Service) GetConfig() *logto.Config {
	return s.config
}

// GetScopes returns the scopes configuration
func (s *Service) GetScopes() *logto.Scopes {
	return s.scopes
}

// RefreshJWKS manually refreshes the JWKS cache
func (s *Service) RefreshJWKS(ctx context.Context) error {
	return s.jwksCache.Refresh(ctx)
}

// CreateAuthMiddleware creates authentication middleware with default config
func (s *Service) CreateAuthMiddleware() fiber.Handler {
	return middleware.AuthMiddleware(s.tokenValidator)
}

// CreateAuthMiddlewareWithConfig creates authentication middleware with custom config
func (s *Service) CreateAuthMiddlewareWithConfig(config middleware.MiddlewareConfig) fiber.Handler {
	return middleware.AuthMiddleware(s.tokenValidator, config)
}

// CreateScopedMiddleware creates middleware that requires specific scopes
func (s *Service) CreateScopedMiddleware(scopes ...string) fiber.Handler {
	return middleware.RequireScopes(scopes...)
}

// CreateOrganizationMiddleware creates middleware for organization validation
func (s *Service) CreateOrganizationMiddleware() fiber.Handler {
	if !s.config.EnableOrganizations {
		log.Warn("Organizations are not enabled in configuration")
	}
	return middleware.RequireOrganization()
}

// CreateTenantMiddleware creates middleware for tenant validation
func (s *Service) CreateTenantMiddleware() fiber.Handler {
	if !s.config.EnableOrganizations {
		log.Warn("Organizations/Tenants are not enabled in configuration")
	}
	return middleware.RequireTenant()
}

// CreateM2MMiddleware creates middleware for M2M validation
func (s *Service) CreateM2MMiddleware() fiber.Handler {
	if !s.config.EnableM2M {
		log.Warn("M2M is not enabled in configuration")
	}
	return middleware.RequireM2M()
}

// ValidateScope checks if a scope is valid
func (s *Service) ValidateScope(scope string) bool {
	// Check against defined scopes
	definedScopes := []string{
		s.scopes.ServiceRead, s.scopes.ServiceWrite, s.scopes.ServiceDelete,
		s.scopes.BookingRead, s.scopes.BookingWrite, s.scopes.BookingDelete, s.scopes.BookingManage,
		s.scopes.ProjectRead, s.scopes.ProjectWrite, s.scopes.ProjectDelete, s.scopes.ProjectManage,
		s.scopes.UserRead, s.scopes.UserWrite, s.scopes.UserDelete, s.scopes.UserManage,
		s.scopes.ArtisanRead, s.scopes.ArtisanWrite, s.scopes.ArtisanManage,
		s.scopes.CustomerRead, s.scopes.CustomerWrite, s.scopes.CustomerManage,
		s.scopes.PaymentRead, s.scopes.PaymentWrite, s.scopes.PaymentProcess,
		s.scopes.TenantRead, s.scopes.TenantWrite, s.scopes.TenantManage, s.scopes.TenantAdmin,
		s.scopes.ReviewRead, s.scopes.ReviewWrite, s.scopes.ReviewModerate,
		s.scopes.ReportRead, s.scopes.ReportGenerate, s.scopes.ReportExport,
		s.scopes.AdminRead, s.scopes.AdminWrite, s.scopes.AdminFull,
	}

	return slices.Contains(definedScopes, scope)
}

// GetRoleScopes returns the scopes for a given role
func (s *Service) GetRoleScopes(roleName string) ([]string, error) {
	roles := logto.DefaultRoles()
	for _, role := range roles {
		if role.Name == roleName {
			return role.Scopes, nil
		}
	}
	return nil, fmt.Errorf("role not found: %s", roleName)
}

// HealthCheck performs a health check on the Logto service
func (s *Service) HealthCheck(ctx context.Context) error {
	// Check if JWKS can be fetched
	keySet := s.jwksCache.GetKeySet()
	if keySet == nil {
		return fmt.Errorf("JWKS not available")
	}

	// Try to refresh JWKS
	if err := s.jwksCache.Refresh(ctx); err != nil {
		return fmt.Errorf("failed to refresh JWKS: %w", err)
	}

	return nil
}

// Close closes the service and cleans up resources
func (s *Service) Close() error {
	log.Info("Closing Logto service")
	// Add cleanup logic if needed
	// s.jwksCache can be cleaned up here if needed
	return nil
}

// M2MClient represents a machine-to-machine client configuration
type M2MClient struct {
	ClientID     string
	ClientSecret string
	Scopes       []string
}

// ObtainM2MToken obtains an access token for M2M authentication
func (s *Service) ObtainM2MToken(ctx context.Context, client M2MClient, resourceIndicator string) (*M2MTokenResponse, error) {
	if !s.config.EnableM2M {
		return nil, fmt.Errorf("M2M is not enabled")
	}

	if client.ClientID == "" || client.ClientSecret == "" {
		return nil, fmt.Errorf("client ID and secret are required")
	}

	// This would normally make an HTTP request to the token endpoint
	// For now, return a placeholder - implement actual token fetching
	log.Warnf("M2M token fetching not fully implemented - use Logto SDK or HTTP client")

	return &M2MTokenResponse{
		AccessToken: "",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "",
	}, fmt.Errorf("not implemented - use external HTTP client to fetch token from %s", s.config.TokenEndpoint)
}

// M2MTokenResponse represents the response from M2M token endpoint
type M2MTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// StartBackgroundRefresh starts a background goroutine to refresh JWKS periodically
func (s *Service) StartBackgroundRefresh(ctx context.Context) {
	ticker := time.NewTicker(s.config.JWKSRefreshWindow)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Info("Stopping JWKS background refresh")
				return
			case <-ticker.C:
				if err := s.jwksCache.Refresh(ctx); err != nil {
					log.Errorf("Failed to refresh JWKS in background: %v", err)
				} else {
					log.Debug("JWKS refreshed successfully in background")
				}
			}
		}
	}()
	log.Info("JWKS background refresh started")
}

// GetAPIResources returns the default API resources configuration
func (s *Service) GetAPIResources() []logto.APIResource {
	return logto.DefaultAPIResources()
}

// GetRoles returns the default role definitions
func (s *Service) GetRoles() []logto.RoleDefinition {
	return logto.DefaultRoles()
}

// ValidateAudience checks if the audience is valid for the API
func (s *Service) ValidateAudience(audience string) bool {
	if s.config.APIResourceIndicator != "" {
		return audience == s.config.APIResourceIndicator
	}

	// Check against default API resources
	resources := logto.DefaultAPIResources()
	for _, resource := range resources {
		if audience == resource.Indicator {
			return true
		}
	}

	return false
}

