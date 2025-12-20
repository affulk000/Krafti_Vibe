package enterprise

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries        int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
	JitterEnabled     bool
	RetryableErrors   []error
	OnRetry           func(attempt int, err error, delay time.Duration)
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:        3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		JitterEnabled:     true,
		RetryableErrors:   []error{},
	}
}

// RetryStrategy defines different retry strategies
type RetryStrategy int

const (
	// StrategyExponential uses exponential backoff
	StrategyExponential RetryStrategy = iota
	// StrategyLinear uses linear backoff
	StrategyLinear
	// StrategyConstant uses constant delay
	StrategyConstant
	// StrategyFibonacci uses fibonacci backoff
	StrategyFibonacci
)

// Retrier handles retry logic with various strategies
type Retrier struct {
	config   *RetryConfig
	strategy RetryStrategy
	logger   log.AllLogger
}

// NewRetrier creates a new retrier with exponential backoff
func NewRetrier(config *RetryConfig, logger log.AllLogger) *Retrier {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return &Retrier{
		config:   config,
		strategy: StrategyExponential,
		logger:   logger,
	}
}

// NewRetrierWithStrategy creates a retrier with specific strategy
func NewRetrierWithStrategy(config *RetryConfig, strategy RetryStrategy, logger log.AllLogger) *Retrier {
	retrier := NewRetrier(config, logger)
	retrier.strategy = strategy
	return retrier
}

// Execute executes a function with retry logic
func (r *Retrier) Execute(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !r.isRetryable(err) {
			return err
		}

		// Don't sleep after last attempt
		if attempt >= r.config.MaxRetries {
			break
		}

		// Calculate delay
		delay := r.calculateDelay(attempt)

		// Call retry callback if provided
		if r.config.OnRetry != nil {
			r.config.OnRetry(attempt+1, err, delay)
		}

		r.logger.Info("retrying operation",
			"attempt", attempt+1,
			"max_retries", r.config.MaxRetries,
			"delay", delay,
			"error", err.Error(),
		)

		// Wait with context awareness
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", r.config.MaxRetries, lastErr)
}

// ExecuteWithResult executes a function with retry logic and returns result
func (r *Retrier) ExecuteWithResult(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	var lastErr error
	var result interface{}

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Execute function
		res, err := fn()
		if err == nil {
			return res, nil
		}

		lastErr = err
		result = res

		// Check if error is retryable
		if !r.isRetryable(err) {
			return result, err
		}

		// Don't sleep after last attempt
		if attempt >= r.config.MaxRetries {
			break
		}

		// Calculate delay
		delay := r.calculateDelay(attempt)

		// Call retry callback if provided
		if r.config.OnRetry != nil {
			r.config.OnRetry(attempt+1, err, delay)
		}

		r.logger.Info("retrying operation",
			"attempt", attempt+1,
			"max_retries", r.config.MaxRetries,
			"delay", delay,
			"error", err.Error(),
		)

		// Wait with context awareness
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return result, fmt.Errorf("operation failed after %d retries: %w", r.config.MaxRetries, lastErr)
}

// calculateDelay calculates the delay for the current attempt
func (r *Retrier) calculateDelay(attempt int) time.Duration {
	var delay time.Duration

	switch r.strategy {
	case StrategyExponential:
		delay = r.exponentialBackoff(attempt)
	case StrategyLinear:
		delay = r.linearBackoff(attempt)
	case StrategyConstant:
		delay = r.config.InitialDelay
	case StrategyFibonacci:
		delay = r.fibonacciBackoff(attempt)
	default:
		delay = r.exponentialBackoff(attempt)
	}

	// Apply jitter if enabled
	if r.config.JitterEnabled {
		delay = r.applyJitter(delay)
	}

	// Ensure delay doesn't exceed max
	delay = min(delay, r.config.MaxDelay)

	return delay
}

// exponentialBackoff calculates exponential backoff delay
func (r *Retrier) exponentialBackoff(attempt int) time.Duration {
	multiplier := math.Pow(r.config.BackoffMultiplier, float64(attempt))
	delay := float64(r.config.InitialDelay) * multiplier
	return time.Duration(delay)
}

// linearBackoff calculates linear backoff delay
func (r *Retrier) linearBackoff(attempt int) time.Duration {
	return r.config.InitialDelay * time.Duration(attempt+1)
}

// fibonacciBackoff calculates fibonacci backoff delay
func (r *Retrier) fibonacciBackoff(attempt int) time.Duration {
	fib := fibonacci(attempt + 1)
	return r.config.InitialDelay * time.Duration(fib)
}

// applyJitter adds random jitter to delay
func (r *Retrier) applyJitter(delay time.Duration) time.Duration {
	jitter := time.Duration(rand.Int63n(int64(delay) / 2))
	return delay + jitter
}

// isRetryable checks if an error is retryable
func (r *Retrier) isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check against configured retryable errors
	for _, retryableErr := range r.config.RetryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}

	// Check using IsRetryable function
	return IsRetryable(err)
}

// fibonacci calculates fibonacci number
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// RetryWithBackoff is a helper function for simple retry with exponential backoff
func RetryWithBackoff(ctx context.Context, maxRetries int, initialDelay time.Duration, fn func() error) error {
	config := &RetryConfig{
		MaxRetries:        maxRetries,
		InitialDelay:      initialDelay,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		JitterEnabled:     true,
	}

	retrier := NewRetrier(config, log.DefaultLogger())
	return retrier.Execute(ctx, fn)
}

// RetryWithBackoffAndResult is a helper function for retry with result
func RetryWithBackoffAndResult(
	ctx context.Context,
	maxRetries int,
	initialDelay time.Duration,
	fn func() (interface{}, error),
) (interface{}, error) {
	config := &RetryConfig{
		MaxRetries:        maxRetries,
		InitialDelay:      initialDelay,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		JitterEnabled:     true,
	}

	retrier := NewRetrier(config, log.DefaultLogger())
	return retrier.ExecuteWithResult(ctx, fn)
}

// RetryMetrics holds retry operation metrics
type RetryMetrics struct {
	TotalRetries      int64
	SuccessfulRetries int64
	FailedRetries     int64
	AverageAttempts   float64
	MaxAttempts       int
	TotalDelay        time.Duration
}

// RetryTracker tracks retry metrics
type RetryTracker struct {
	metrics RetryMetrics
}

// NewRetryTracker creates a new retry tracker
func NewRetryTracker() *RetryTracker {
	return &RetryTracker{
		metrics: RetryMetrics{},
	}
}

// RecordRetry records a retry attempt
func (t *RetryTracker) RecordRetry(attempts int, success bool, delay time.Duration) {
	t.metrics.TotalRetries++
	t.metrics.TotalDelay += delay

	if success {
		t.metrics.SuccessfulRetries++
	} else {
		t.metrics.FailedRetries++
	}

	t.metrics.MaxAttempts = max(t.metrics.MaxAttempts, attempts)

	// Calculate average attempts
	if t.metrics.TotalRetries > 0 {
		t.metrics.AverageAttempts = float64(t.metrics.SuccessfulRetries+t.metrics.FailedRetries) / float64(t.metrics.TotalRetries)
	}
}

// GetMetrics returns current retry metrics
func (t *RetryTracker) GetMetrics() RetryMetrics {
	return t.metrics
}

// Reset resets retry metrics
func (t *RetryTracker) Reset() {
	t.metrics = RetryMetrics{}
}

// RetryCondition defines a condition for retry
type RetryCondition func(err error) bool

// RetryOnError retries on specific error
func RetryOnError(target error) RetryCondition {
	return func(err error) bool {
		return errors.Is(err, target)
	}
}

// RetryOnAnyError retries on any error
func RetryOnAnyError() RetryCondition {
	return func(err error) bool {
		return err != nil
	}
}

// RetryOnErrors retries on multiple errors
func RetryOnErrors(targets ...error) RetryCondition {
	return func(err error) bool {
		for _, target := range targets {
			if errors.Is(err, target) {
				return true
			}
		}
		return false
	}
}

// ConditionalRetrier retries based on conditions
type ConditionalRetrier struct {
	*Retrier
	conditions []RetryCondition
}

// NewConditionalRetrier creates a conditional retrier
func NewConditionalRetrier(config *RetryConfig, logger log.AllLogger, conditions ...RetryCondition) *ConditionalRetrier {
	return &ConditionalRetrier{
		Retrier:    NewRetrier(config, logger),
		conditions: conditions,
	}
}

// Execute executes with conditional retry
func (cr *ConditionalRetrier) Execute(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= cr.config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check all conditions
		shouldRetry := false
		for _, condition := range cr.conditions {
			if condition(err) {
				shouldRetry = true
				break
			}
		}

		if !shouldRetry {
			return err
		}

		if attempt >= cr.config.MaxRetries {
			break
		}

		delay := cr.calculateDelay(attempt)

		if cr.config.OnRetry != nil {
			cr.config.OnRetry(attempt+1, err, delay)
		}

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", cr.config.MaxRetries, lastErr)
}

// Common retryable errors
var (
	ErrTemporaryFailure = errors.New("temporary failure")
	ErrServiceBusy      = errors.New("service busy")
	ErrRateLimited      = errors.New("rate limited")
	ErrConnectionFailed = errors.New("connection failed")
	ErrTimeout          = errors.New("operation timeout")
)
