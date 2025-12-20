package auth

import (
	"Krafti_Vibe/internal/config"
	"context"
	"crypto/tls"
	"net/http"

	"github.com/gofiber/fiber/v2/log"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

type ZitadelAuth struct {
	AuthZ *authorization.Authorizer[*oauth.IntrospectionContext]
}

func NewZitadelAuth(cfg *config.Config) (*ZitadelAuth, error) {
	ctx := context.Background()

	// For local development with self-signed certificates
	// WARNING: Only use in development! Never in production!
	if cfg.IsDevelopment() {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		log.Warn("⚠️  TLS certificate verification disabled for local development")
	}

	authZ, err := authorization.New(
		ctx,
		zitadel.New(cfg.Zitadel.Domain),
		oauth.DefaultAuthorization(cfg.Zitadel.KeyPath),
	)
	if err != nil {
		log.Errorf("failed to create zitadel auth: %v", err)
		return nil, err
	}

	log.Infof("✅ Zitadel authentication initialized for domain: %s", cfg.Zitadel.Domain)
	return &ZitadelAuth{
		AuthZ: authZ,
	}, nil
}
