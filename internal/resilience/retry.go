package resilience

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// RetryPolicy defines the retry behavior for operations.
type RetryPolicy struct {
	// MaxRetries is the maximum number of retry attempts (not including initial call).
	MaxRetries int `json:"maxRetries"`

	// BaseDelay is the initial delay before the first retry.
	BaseDelay time.Duration `json:"baseDelay"`

	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration `json:"maxDelay"`

	// UseJitter adds randomness to delays to prevent thundering herd.
	UseJitter bool `json:"useJitter"`

	// RetryableErrors is a list of errors that should be retried.
	// If empty, all errors except client errors are retried.
	RetryableErrors []error `json:"retryableErrors"`
}

// Retry executes the given function with the specified retry policy.
// It returns the error from the last attempt if all retries are exhausted.
func Retry(ctx context.Context, policy RetryPolicy, fn func() error) error {
	var lastErr error

	// Initial attempt plus retries
	maxAttempts := policy.MaxRetries + 1

	for attempt := range maxAttempts {
		// Check context before each attempt
		if err := ctx.Err(); err != nil {
			return err
		}

		// Execute the function
		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if !isErrorRetryable(err, policy.RetryableErrors) {
			return err
		}

		// Don't delay after the last attempt
		if attempt < maxAttempts-1 {
			delay := CalculateBackoff(attempt, policy.BaseDelay, policy.MaxDelay, policy.UseJitter)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	return lastErr
}

// CalculateBackoff calculates the backoff delay for a given attempt.
// The delay grows exponentially: baseDelay * 2^attempt, capped at maxDelay.
func CalculateBackoff(attempt int, baseDelay, maxDelay time.Duration, useJitter bool) time.Duration {
	if baseDelay <= 0 {
		baseDelay = 100 * time.Millisecond
	}
	if maxDelay <= 0 {
		maxDelay = 30 * time.Second
	}

	// Calculate exponential backoff: baseDelay * 2^attempt
	delay := baseDelay
	for range attempt {
		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
			break
		}
	}

	// Apply jitter if enabled (random value between 0.5 and 1.5 of calculated delay)
	if useJitter {
		jitterFactor := 0.5 + rand.Float64() // 0.5 to 1.5
		delay = time.Duration(float64(delay) * jitterFactor)
	}

	// Ensure we don't exceed max delay
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// IsRetryableError determines if an error should be retried.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Context errors are not retryable
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Client errors are not retryable
	var clientErr isClientError
	if errors.As(err, &clientErr) && clientErr.IsClientError() {
		return false
	}

	return true
}

// isErrorRetryable checks if the error should be retried based on the policy.
func isErrorRetryable(err error, retryableErrors []error) bool {
	if err == nil {
		return false
	}

	// Context errors are never retryable
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Client errors are never retryable
	var clientErr isClientError
	if errors.As(err, &clientErr) && clientErr.IsClientError() {
		return false
	}

	// If specific retryable errors are defined, check against them
	if len(retryableErrors) > 0 {
		for _, retryable := range retryableErrors {
			if errors.Is(err, retryable) {
				return true
			}
		}
		return false
	}

	// Default: all errors except client errors are retryable
	return true
}
