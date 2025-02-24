package ai

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	type testCase struct {
		name     string
		rps      float64
		burst    int
		requests int
		wantErr  bool
	}

	tests := []testCase{
		{
			name:     "basic_rate_limiting",
			rps:      10.0,
			burst:    1,
			requests: 3,
			wantErr:  false,
		},
		{
			name:     "burst_handling",
			rps:      1.0,
			burst:    3,
			requests: 3,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(tt.rps, tt.burst)
			ctx := context.TODO()

			for i := 0; i < tt.requests; i++ {
				err := rl.Wait(ctx)
				if tt.wantErr {
					if err == nil {
						t.Error("expected error, got nil")
					}
					return
				}
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestRetryWithBackoff(t *testing.T) {
	type testCase struct {
		name      string
		operation func() error
		config    RetryConfig
		wantErr   bool
		wantRetry int
	}

	tests := []testCase{
		{
			name: "Successfully_complete_operation_without_retry",
			config: RetryConfig{
				MaxRetries:      3,
				InitialInterval: 10 * time.Millisecond,
				MaxInterval:     100 * time.Millisecond,
				Multiplier:      2.0,
			},
			operation: func() error {
				return nil
			},
			wantErr:   false,
			wantRetry: 0,
		},
		{
			name: "Retry_error_rate_limit_exceeded",
			config: RetryConfig{
				MaxRetries:      3,
				InitialInterval: 10 * time.Millisecond,
				MaxInterval:     100 * time.Millisecond,
				Multiplier:      2.0,
			},
			operation: func() error {
				return &RateLimitError{
					StatusCode: 429,
					Message:    "rate limit exceeded",
				}
			},
			wantErr:   true,
			wantRetry: 3,
		},
		{
			name: "Error_non_retryable_deadline_exceeded",
			config: RetryConfig{
				MaxRetries:      3,
				InitialInterval: 10 * time.Millisecond,
				MaxInterval:     100 * time.Millisecond,
				Multiplier:      2.0,
			},
			operation: func() error {
				return context.DeadlineExceeded
			},
			wantErr:   true,
			wantRetry: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retryCount := 0
			wrappedOp := func() error {
				retryCount++
				return tt.operation()
			}

			err := retryWithBackoff(context.TODO(), tt.config, wrappedOp)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
			}

			if retryCount != tt.wantRetry+1 {
				t.Errorf("retry count = %d, want %d", retryCount, tt.wantRetry+1)
			}
		})
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig

	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", config.MaxRetries)
		return
	}

	if config.InitialInterval != time.Second {
		t.Errorf("InitialInterval = %v, want %v", config.InitialInterval, time.Second)
		return
	}

	if config.MaxInterval != 30*time.Second {
		t.Errorf("MaxInterval = %v, want %v", config.MaxInterval, 30*time.Second)
		return
	}

	if config.Multiplier != 2.0 {
		t.Errorf("Multiplier = %v, want 2.0", config.Multiplier)
		return
	}
}
