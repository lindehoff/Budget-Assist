package ai

import (
	"context"
	"errors"
	"time"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
}

// DefaultRetryConfig provides default retry settings
var DefaultRetryConfig = RetryConfig{
	MaxRetries:      3,
	InitialInterval: time.Second,
	MaxInterval:     30 * time.Second,
	Multiplier:      2.0,
}

// RateLimiter implements rate limiting for API requests
type RateLimiter struct {
	tokens   chan struct{}
	interval time.Duration
}

// NewRateLimiter creates a new rate limiter with the specified RPS and burst
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	interval := time.Duration(float64(time.Second) / rps)
	tokens := make(chan struct{}, burst)
	for i := 0; i < burst; i++ {
		tokens <- struct{}{}
	}

	limiter := &RateLimiter{
		tokens:   tokens,
		interval: interval,
	}

	go limiter.refill()
	return limiter
}

// Wait blocks until a token is available or the context is cancelled
func (r *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-r.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *RateLimiter) refill() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case r.tokens <- struct{}{}:
		default:
		}
	}
}

// retryWithBackoff retries an operation with exponential backoff
func retryWithBackoff(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error
	currentInterval := config.InitialInterval

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		var rateLimitErr *RateLimitError
		if !errors.As(err, &rateLimitErr) {
			return err
		}

		lastErr = err

		if attempt == config.MaxRetries {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(currentInterval):
		}

		currentInterval = time.Duration(float64(currentInterval) * config.Multiplier)
		if currentInterval > config.MaxInterval {
			currentInterval = config.MaxInterval
		}
	}

	return lastErr
}
