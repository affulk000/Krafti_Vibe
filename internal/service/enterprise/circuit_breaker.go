package enterprise

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/sony/gobreaker"
)

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Name          string
	MaxRequests   uint32
	Interval      time.Duration
	Timeout       time.Duration
	FailureRatio  float64
	OnStateChange func(name string, from gobreaker.State, to gobreaker.State)
}

// DefaultCircuitBreakerConfig returns default configuration
func DefaultCircuitBreakerConfig(name string) *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		Name:         name,
		MaxRequests:  3,
		Interval:     60 * time.Second,
		Timeout:      30 * time.Second,
		FailureRatio: 0.6,
	}
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*gobreaker.CircuitBreaker
	mu       sync.RWMutex
	logger   log.AllLogger
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(logger log.AllLogger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
		logger:   logger,
	}
}

// GetOrCreate returns existing circuit breaker or creates a new one
func (m *CircuitBreakerManager) GetOrCreate(config *CircuitBreakerConfig) *gobreaker.CircuitBreaker {
	m.mu.RLock()
	cb, exists := m.breakers[config.Name]
	m.mu.RUnlock()

	if exists {
		return cb
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, exists := m.breakers[config.Name]; exists {
		return cb
	}

	// Create new circuit breaker
	settings := gobreaker.Settings{
		Name:        config.Name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= config.FailureRatio
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			m.logger.Warn("circuit breaker state changed",
				"name", name,
				"from", from.String(),
				"to", to.String(),
			)
			if config.OnStateChange != nil {
				config.OnStateChange(name, from, to)
			}
		},
	}

	cb = gobreaker.NewCircuitBreaker(settings)
	m.breakers[config.Name] = cb

	m.logger.Info("circuit breaker created", "name", config.Name)
	return cb
}

// Get returns an existing circuit breaker
func (m *CircuitBreakerManager) Get(name string) (*gobreaker.CircuitBreaker, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cb, exists := m.breakers[name]
	if !exists {
		return nil, fmt.Errorf("circuit breaker %s not found", name)
	}
	return cb, nil
}

// GetState returns the current state of a circuit breaker
func (m *CircuitBreakerManager) GetState(name string) (gobreaker.State, error) {
	cb, err := m.Get(name)
	if err != nil {
		return gobreaker.StateClosed, err
	}
	return cb.State(), nil
}

// GetAllStates returns states of all circuit breakers
func (m *CircuitBreakerManager) GetAllStates() map[string]gobreaker.State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make(map[string]gobreaker.State)
	for name, cb := range m.breakers {
		states[name] = cb.State()
	}
	return states
}

// Reset resets a circuit breaker to closed state
func (m *CircuitBreakerManager) Reset(name string) error {
	_, err := m.Get(name)
	if err != nil {
		return err
	}

	// Circuit breakers don't have a public reset, but we can recreate
	m.mu.Lock()
	defer m.mu.Unlock()

	if oldCb, exists := m.breakers[name]; exists {
		// Create new breaker with same settings
		settings := gobreaker.Settings{
			Name:        name,
			MaxRequests: 3,
			Interval:    60 * time.Second,
			Timeout:     30 * time.Second,
		}
		m.breakers[name] = gobreaker.NewCircuitBreaker(settings)
		m.logger.Info("circuit breaker reset", "name", name, "old_state", oldCb.State().String())
	}

	return nil
}

// CircuitBreakerWrapper wraps operations with circuit breaker protection
type CircuitBreakerWrapper struct {
	cb     *gobreaker.CircuitBreaker
	logger log.AllLogger
}

// NewCircuitBreakerWrapper creates a new wrapper
func NewCircuitBreakerWrapper(cb *gobreaker.CircuitBreaker, logger log.AllLogger) *CircuitBreakerWrapper {
	return &CircuitBreakerWrapper{
		cb:     cb,
		logger: logger,
	}
}

// Execute executes a function with circuit breaker protection
func (w *CircuitBreakerWrapper) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	return w.cb.Execute(func() (any, error) {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return fn()
	})
}

// ExecuteWithFallback executes with circuit breaker and fallback
func (w *CircuitBreakerWrapper) ExecuteWithFallback(
	ctx context.Context,
	fn func() (any, error),
	fallback func() (any, error),
) (any, error) {
	result, err := w.Execute(ctx, fn)
	if err != nil {
		w.logger.Warn("circuit breaker execution failed, using fallback", "error", err)
		return fallback()
	}
	return result, nil
}

// IsOpen returns true if circuit breaker is open
func (w *CircuitBreakerWrapper) IsOpen() bool {
	return w.cb.State() == gobreaker.StateOpen
}

// State returns current circuit breaker state
func (w *CircuitBreakerWrapper) State() gobreaker.State {
	return w.cb.State()
}

// CircuitBreakerStats holds circuit breaker statistics
type CircuitBreakerStats struct {
	Name             string           `json:"name"`
	State            string           `json:"state"`
	Counts           gobreaker.Counts `json:"counts"`
	LastStateChange  time.Time        `json:"last_state_change,omitempty"`
	ConsecutiveFails uint32           `json:"consecutive_fails"`
}

// GetStats returns circuit breaker statistics
func (m *CircuitBreakerManager) GetStats(name string) (*CircuitBreakerStats, error) {
	cb, err := m.Get(name)
	if err != nil {
		return nil, err
	}

	return &CircuitBreakerStats{
		Name:  name,
		State: cb.State().String(),
	}, nil
}

// HealthCheck checks if all circuit breakers are healthy
func (m *CircuitBreakerManager) HealthCheck() (bool, []string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	healthy := true
	issues := []string{}

	for name, cb := range m.breakers {
		if cb.State() == gobreaker.StateOpen {
			healthy = false
			issues = append(issues, fmt.Sprintf("circuit breaker %s is open", name))
		}
	}

	return healthy, issues
}

// CircuitBreakerMetrics for prometheus
type CircuitBreakerMetrics struct {
	Requests             int64
	TotalSuccesses       int64
	TotalFailures        int64
	ConsecutiveSuccesses int64
	ConsecutiveFailures  int64
}

// GetMetrics returns metrics for a circuit breaker
func (m *CircuitBreakerManager) GetMetrics(name string) (*CircuitBreakerMetrics, error) {
	cb, err := m.Get(name)
	if err != nil {
		return nil, err
	}

	counts := cb.Counts()
	return &CircuitBreakerMetrics{
		Requests:             int64(counts.Requests),
		TotalSuccesses:       int64(counts.TotalSuccesses),
		TotalFailures:        int64(counts.TotalFailures),
		ConsecutiveSuccesses: int64(counts.ConsecutiveSuccesses),
		ConsecutiveFailures:  int64(counts.ConsecutiveFailures),
	}, nil
}

// RetryableError wraps an error that should be retried
type RetryableError struct {
	Err     error
	Retries int
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable error (attempt %d): %v", e.Retries, e.Err)
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryable checks if an error should be retried
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types
	switch err.(type) {
	case *RetryableError:
		return true
	default:
		// Check error message for common patterns
		errMsg := err.Error()
		retryablePatterns := []string{
			"connection refused",
			"connection reset",
			"timeout",
			"temporary failure",
			"service unavailable",
			"too many requests",
		}

		for _, pattern := range retryablePatterns {
			if contains(errMsg, pattern) {
				return true
			}
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
