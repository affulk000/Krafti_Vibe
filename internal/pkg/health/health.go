package health

import (
	"Krafti_Vibe/internal/infrastructure/database"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// HealthStatus represents the health status
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    HealthStatus           `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks,omitempty"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Status       HealthStatus `json:"status"`
	Message      string       `json:"message,omitempty"`
	ResponseTime string       `json:"response_time,omitempty"`
}

// Checker defines the interface for health checkers
type Checker interface {
	Check(ctx context.Context) error
	Name() string
}

// HealthChecker manages health checks
type HealthChecker struct {
	checkers []Checker
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(checkers ...Checker) *HealthChecker {
	return &HealthChecker{
		checkers: checkers,
	}
}

// Register adds a new health checker
func (h *HealthChecker) Register(checker Checker) {
	h.checkers = append(h.checkers, checker)
}

// Check performs all health checks
func (h *HealthChecker) Check(ctx context.Context) HealthResponse {
	checks := make(map[string]CheckResult)
	overallStatus := StatusHealthy

	for _, checker := range h.checkers {
		start := time.Now()
		err := checker.Check(ctx)
		duration := time.Since(start)

		result := CheckResult{
			Status:       StatusHealthy,
			ResponseTime: duration.String(),
		}

		if err != nil {
			result.Status = StatusUnhealthy
			result.Message = err.Error()
			overallStatus = StatusUnhealthy
		}

		checks[checker.Name()] = result
	}

	return HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	}
}

// DatabaseChecker checks database health
type DatabaseChecker struct{}

func (c *DatabaseChecker) Name() string {
	return "database"
}

func (c *DatabaseChecker) Check(ctx context.Context) error {
	return database.HealthCheck(ctx)
}

// Handler returns a Fiber handler for health checks
func Handler(healthChecker *HealthChecker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		response := healthChecker.Check(ctx)

		statusCode := fiber.StatusOK
		if response.Status == StatusUnhealthy {
			statusCode = fiber.StatusServiceUnavailable
		}

		return c.Status(statusCode).JSON(response)
	}
}

// ReadinessHandler returns a handler for readiness checks
func ReadinessHandler(healthChecker *HealthChecker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		response := healthChecker.Check(ctx)

		// Readiness requires all checks to pass
		if response.Status != StatusHealthy {
			return c.Status(fiber.StatusServiceUnavailable).JSON(response)
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// LivenessHandler returns a handler for liveness checks
func LivenessHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":    "alive",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}
