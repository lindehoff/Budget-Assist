package ai

import (
	"context"
	"time"
)

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
