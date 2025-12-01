package logto

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// JWKSCache manages the JWKS cache with auto-refresh
type JWKSCache struct {
	endpoint  string
	keySet    jwk.Set
	ttl       time.Duration
	lastFetch time.Time
}

// NewJWKSCache creates a new JWKS cache
func NewJWKSCache(endpoint string, ttl time.Duration) (*JWKSCache, error) {
	cache := &JWKSCache{
		endpoint: endpoint,
		ttl:      ttl,
	}

	// Initial fetch
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := cache.Refresh(ctx); err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	log.Infof("JWKS cache initialized from %s", endpoint)

	return cache, nil
}

// Refresh fetches and updates the JWKS
func (c *JWKSCache) Refresh(ctx context.Context) error {
	keySet, err := jwk.Fetch(ctx, c.endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	c.keySet = keySet
	c.lastFetch = time.Now()

	log.Debugf("JWKS refreshed, key count: %d", keySet.Len())

	return nil
}

// GetKeySet returns the cached key set
func (c *JWKSCache) GetKeySet() jwk.Set {
	return c.keySet
}

// ShouldRefresh checks if the cache should be refreshed
func (c *JWKSCache) ShouldRefresh() bool {
	return time.Since(c.lastFetch) > c.ttl
}

// GetLastFetchTime returns the last fetch time
func (c *JWKSCache) GetLastFetchTime() time.Time {
	return c.lastFetch
}

type TokenValidator struct {
	jwksCache *JWKSCache
	issuer    string
}

func NewTokenValidator(jwksCache *JWKSCache, issuer string) *TokenValidator {
	return &TokenValidator{
		jwksCache: jwksCache,
		issuer:    issuer,
	}
}

func (tv *TokenValidator) ExtractToken(c *fiber.Ctx) (string, error) {
	const bearerPrefix = "Bearer "
	authorization := c.Get("Authorization")

	if authorization == "" {
		return "", fmt.Errorf("authorization header is missing")
	}

	if !strings.HasPrefix(authorization, bearerPrefix) {
		return "", fmt.Errorf("authorization header must start with 'Bearer '")
	}

	return strings.TrimPrefix(authorization, bearerPrefix), nil
}

func (tv *TokenValidator) ValidateToken(tokenString string) (jwt.Token, error) {
	keySet := tv.jwksCache.GetKeySet()

	token, err := jwt.Parse([]byte(tokenString), jwt.WithKeySet(keySet))
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Verify issuer
	issuer, ok := token.Issuer()
	if !ok || issuer != tv.issuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	return token, nil
}

func (tv *TokenValidator) GetClaims(token jwt.Token) *TokenClaims {
	subject, _ := token.Subject()
	audience, _ := token.Audience()

	return &TokenClaims{
		Subject:        subject,
		ClientID:       getStringClaim(token, "client_id"),
		OrganizationID: getStringClaim(token, "organization_id"),
		Scopes:         getScopesFromToken(token),
		Audience:       audience,
	}
}

type TokenClaims struct {
	Subject        string
	ClientID       string
	OrganizationID string
	Scopes         []string
	Audience       []string
}

func (tc *TokenClaims) HasAudience(aud string) bool {
	return slices.Contains(tc.Audience, aud)
}

func (tc *TokenClaims) HasScope(scope string) bool {
	return slices.Contains(tc.Scopes, scope)
}

func (tc *TokenClaims) HasScopes(scopes []string) bool {
	for _, required := range scopes {
		if !tc.HasScope(required) {
			return false
		}
	}
	return true
}

func getStringClaim(token jwt.Token, key string) string {
	var val any
	if err := token.Get(key, &val); err == nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getScopesFromToken(token jwt.Token) []string {
	var val any
	if err := token.Get("scope", &val); err == nil {
		if scope, ok := val.(string); ok && scope != "" {
			return strings.Split(scope, " ")
		}
	}
	return []string{}
}
