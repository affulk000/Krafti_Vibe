package router

import (
	"Krafti_Vibe/internal/infrastructure/cache"
	"Krafti_Vibe/internal/middleware"
	"Krafti_Vibe/internal/repository"
	ws "Krafti_Vibe/internal/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	swagger "github.com/swaggo/fiber-swagger"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Config holds the router configuration
type Config struct {
	DB                *gorm.DB
	Logger            log.AllLogger
	ZitadelAuthZ      *authorization.Authorizer[*oauth.IntrospectionContext]
	ZitadelMiddleware *middleware.ZitadelAuthMiddleware
	Cache             cache.Cache            // Optional: for rate limiting
	ZapLogger         *zap.Logger            // Optional: for rate limiting (zap structured logging)
	CORSConfig        *middleware.CORSConfig // Optional: for CORS
	WebhookSecret     string                 // Webhook signing secret
}

// Router handles all application routes
type Router struct {
	app       *fiber.App
	config    *Config
	repos     *repository.Repositories
	zitadelMW *middleware.ZitadelAuthMiddleware
	wsHub     *ws.Hub
	wsHandler *ws.Handler
}

// New creates a new router instance
func New(app *fiber.App, config *Config) *Router {
	// Initialize WebSocket hub and handler
	hub := ws.NewHub()
	handler := ws.NewHandler(hub)

	return &Router{
		app:       app,
		config:    config,
		zitadelMW: config.ZitadelMiddleware,
		wsHub:     hub,
		wsHandler: handler,
	}
}

// Setup initializes all routes
func (r *Router) Setup() error {
	// Initialize repositories
	r.repos = repository.NewRepositories(r.config.DB, repository.RepositoryConfig{
		Logger: r.config.Logger,
	})

	// Start WebSocket hub
	go r.wsHub.Run()
	r.config.Logger.Info("WebSocket hub started")

	// Apply global CORS middleware if configured
	if r.config.CORSConfig != nil {
		r.app.Use(middleware.CORSMiddleware(*r.config.CORSConfig))
		r.config.Logger.Info("CORS middleware enabled")
	}

	// Setup API routes
	r.setupAPIRoutes()

	return nil
}

// setupAPIRoutes sets up all API routes
func (r *Router) setupAPIRoutes() {
	// Swagger documentation (no auth required)
	r.app.Get("/swagger/*", swagger.WrapHandler)
	r.config.Logger.Info("Swagger documentation available at /swagger/index.html")

	// API v1 routes
	api := r.app.Group("/api/v1")

	// Ping godoc
	// @Summary API health check
	// @Description Simple ping endpoint to verify API is responsive
	// @Tags Health
	// @Produce json
	// @Success 200 {object} map[string]interface{} "API is healthy"
	// @Router /ping [get]
	api.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "pong",
		})
	})

	// Setup feature routes
	r.setupUserRoutes(api)
	r.setupArtisanRoutes(api)
	r.setupCustomerRoutes(api)
	r.setupBookingRoutes(api)
	r.setupInvoiceRoutes(api)
	r.setupPaymentRoutes(api)
	r.setupSubscriptionRoutes(api)
	r.setupMessageRoutes(api)
	r.setupNotificationRoutes(api)
	r.setupDataExportRoutes(api)
	r.setupTenantUsageRoutes(api)
	r.setupTenantRoutes(api)
	r.setupMilestoneRoutes(api)
	r.setupTaskRoutes(api)
	r.setupServiceRoutes(api)
	r.setupProjectRoutes(api)
	r.setupReviewRoutes(api)

	// Setup WebSocket routes
	r.setupWebSocketRoutes(api, r.wsHandler)

	// Setup Webhook routes
	r.setupWebhookRoutes(api)

	// Setup File Upload routes
	r.setupFileUploadRoutes(api)

	// Setup Report routes
	r.setupReportRoutes(api)

	// Setup Promo Code routes
	r.setupPromoRoutes(api)

	// Setup System Settings routes
	r.setupSystemSettingsRoutes(api)

	// Setup Tenant Invitation routes
	r.setupTenantInvitationRoutes(api)

	// Setup Project Update routes
	r.setupProjectUpdateRoutes(api)

	// Setup Availability routes
	r.setupAvailabilityRoutes(api)

	// Setup WhiteLabel routes
	r.setupWhiteLabelRoutes(api)

	// Setup SDK routes
	r.setupSDKRoutes(api)
}

// GetRepositories returns the repositories instance
func (r *Router) GetRepositories() *repository.Repositories {
	return r.repos
}

// RequireAuth returns authentication middleware or a no-op handler if Zitadel is not configured
// This prevents nil pointer dereference when running in development mode without Zitadel
func (r *Router) RequireAuth(opts ...authorization.CheckOption) fiber.Handler {
	if r.zitadelMW == nil {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}
	return r.zitadelMW.RequireAuth(opts...)
}
