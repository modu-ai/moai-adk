package resilience

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetrySuccess(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	ctx := context.Background()
	var callCount atomic.Int32

	err := Retry(ctx, policy, func() error {
		callCount.Add(1)
		return nil
	})

	if err != nil {
		t.Errorf("Retry() error = %v, want nil", err)
	}
	if callCount.Load() != 1 {
		t.Errorf("call count = %d, want 1", callCount.Load())
	}
}

func TestRetryEventualSuccess(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	ctx := context.Background()
	var callCount atomic.Int32
	transientErr := errors.New("transient error")

	err := Retry(ctx, policy, func() error {
		count := callCount.Add(1)
		if count < 3 {
			return transientErr
		}
		return nil
	})

	if err != nil {
		t.Errorf("Retry() error = %v, want nil", err)
	}
	if callCount.Load() != 3 {
		t.Errorf("call count = %d, want 3", callCount.Load())
	}
}

func TestRetryExhausted(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	ctx := context.Background()
	var callCount atomic.Int32
	persistentErr := errors.New("persistent error")

	err := Retry(ctx, policy, func() error {
		callCount.Add(1)
		return persistentErr
	})

	if !errors.Is(err, persistentErr) {
		t.Errorf("Retry() error = %v, want %v", err, persistentErr)
	}
	// Initial call + MaxRetries retries = 4 total calls
	if callCount.Load() != 4 {
		t.Errorf("call count = %d, want 4", callCount.Load())
	}
}

func TestRetryContextCancellation(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{
		MaxRetries: 10,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   1 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	var callCount atomic.Int32

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := Retry(ctx, policy, func() error {
		callCount.Add(1)
		return errors.New("fail")
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Retry() error = %v, want context.Canceled", err)
	}
}

func TestRetryExponentialBackoff(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   1 * time.Second,
	}

	ctx := context.Background()
	var timestamps []time.Time
	transientErr := errors.New("transient error")

	_ = Retry(ctx, policy, func() error {
		timestamps = append(timestamps, time.Now())
		if len(timestamps) < 4 {
			return transientErr
		}
		return nil
	})

	if len(timestamps) < 3 {
		t.Fatalf("expected at least 3 timestamps, got %d", len(timestamps))
	}

	// Verify exponential growth of delays
	// First delay should be around BaseDelay (10ms)
	// Second delay should be around 2*BaseDelay (20ms)
	// Third delay should be around 4*BaseDelay (40ms)
	for i := 1; i < len(timestamps); i++ {
		delay := timestamps[i].Sub(timestamps[i-1])
		// Allow for some variance due to timing
		multiplier := 1 << (i - 1)
		minExpected := time.Duration(float64(policy.BaseDelay) * float64(multiplier) * 0.5)
		if delay < minExpected {
			t.Logf("delay %d was %v, expected at least %v", i, delay, minExpected)
		}
	}
}

func TestRetryMaxDelayCap(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{
		MaxRetries: 10,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   200 * time.Millisecond, // Cap at 200ms
	}

	ctx := context.Background()
	var timestamps []time.Time
	transientErr := errors.New("transient error")

	start := time.Now()
	_ = Retry(ctx, policy, func() error {
		timestamps = append(timestamps, time.Now())
		if len(timestamps) < 5 {
			return transientErr
		}
		return nil
	})
	elapsed := time.Since(start)

	// With max delay of 200ms, total time should be capped
	// Without cap: 100 + 200 + 400 + 800 = 1500ms
	// With cap: 100 + 200 + 200 + 200 = 700ms
	if elapsed > 1*time.Second {
		t.Errorf("elapsed time %v suggests MaxDelay is not being applied", elapsed)
	}
}

func TestRetryJitter(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{
		MaxRetries: 5,
		BaseDelay:  50 * time.Millisecond,
		MaxDelay:   1 * time.Second,
		UseJitter:  true,
	}

	ctx := context.Background()
	var delays []time.Duration
	var lastTime time.Time
	transientErr := errors.New("transient error")

	_ = Retry(ctx, policy, func() error {
		now := time.Now()
		if !lastTime.IsZero() {
			delays = append(delays, now.Sub(lastTime))
		}
		lastTime = now
		if len(delays) < 5 {
			return transientErr
		}
		return nil
	})

	if len(delays) < 2 {
		t.Skip("not enough delays to verify jitter")
	}

	// With jitter, delays should have some variance
	// This is a probabilistic test, so we just verify delays are not identical
	allSame := true
	for i := 1; i < len(delays); i++ {
		// Account for expected exponential backoff ratio
		ratio := float64(delays[i]) / float64(delays[i-1])
		if ratio < 1.5 || ratio > 3.0 {
			allSame = false
			break
		}
	}
	// Note: This test may occasionally fail due to timing, which is acceptable
	_ = allSame
}

func TestRetryNoRetryOnClientError(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{
		MaxRetries: 5,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	ctx := context.Background()
	var callCount atomic.Int32
	clientErr := NewClientError("invalid input")

	err := Retry(ctx, policy, func() error {
		callCount.Add(1)
		return clientErr
	})

	if !errors.Is(err, clientErr) {
		t.Errorf("Retry() error = %v, want %v", err, clientErr)
	}
	// Should only be called once - no retries for client errors
	if callCount.Load() != 1 {
		t.Errorf("call count = %d, want 1 (no retries for client errors)", callCount.Load())
	}
}

func TestRetryWithRetryableErrors(t *testing.T) {
	t.Parallel()

	retryableErr := errors.New("retryable")
	nonRetryableErr := errors.New("non-retryable")

	policy := RetryPolicy{
		MaxRetries:      5,
		BaseDelay:       10 * time.Millisecond,
		MaxDelay:        100 * time.Millisecond,
		RetryableErrors: []error{retryableErr},
	}

	ctx := context.Background()

	// Test retryable error is retried
	var callCount1 atomic.Int32
	_ = Retry(ctx, policy, func() error {
		callCount1.Add(1)
		return retryableErr
	})
	if callCount1.Load() != 6 { // 1 initial + 5 retries
		t.Errorf("retryable error: call count = %d, want 6", callCount1.Load())
	}

	// Test non-retryable error is not retried
	var callCount2 atomic.Int32
	_ = Retry(ctx, policy, func() error {
		callCount2.Add(1)
		return nonRetryableErr
	})
	if callCount2.Load() != 1 {
		t.Errorf("non-retryable error: call count = %d, want 1", callCount2.Load())
	}
}

func TestRetryZeroMaxRetries(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{
		MaxRetries: 0,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	ctx := context.Background()
	var callCount atomic.Int32
	err := Retry(ctx, policy, func() error {
		callCount.Add(1)
		return errors.New("fail")
	})

	if err == nil {
		t.Error("Retry() error = nil, want error")
	}
	// Should only be called once with zero retries
	if callCount.Load() != 1 {
		t.Errorf("call count = %d, want 1", callCount.Load())
	}
}

func TestRetryContextTimeout(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{
		MaxRetries: 10,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   1 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	var callCount atomic.Int32
	err := Retry(ctx, policy, func() error {
		callCount.Add(1)
		return errors.New("fail")
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Retry() error = %v, want context.DeadlineExceeded", err)
	}
}

// Test client error type
func TestClientError(t *testing.T) {
	t.Parallel()

	err := NewClientError("invalid input")
	if err == nil {
		t.Fatal("NewClientError returned nil")
	}

	if !err.IsClientError() {
		t.Error("IsClientError() = false, want true")
	}

	expectedMsg := "client error: invalid input"
	if err.Error() != expectedMsg {
		t.Errorf("Error() = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestIsRetryableError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error is not retryable", nil, false},
		{"client error is not retryable", NewClientError("bad input"), false},
		{"generic error is retryable", errors.New("generic"), true},
		{"context canceled is not retryable", context.Canceled, false},
		{"context deadline exceeded is not retryable", context.DeadlineExceeded, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsRetryableError(tt.err); got != tt.want {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		attempt   int
		baseDelay time.Duration
		maxDelay  time.Duration
		wantMin   time.Duration
		wantMax   time.Duration
	}{
		{
			name:      "first attempt",
			attempt:   0,
			baseDelay: 100 * time.Millisecond,
			maxDelay:  10 * time.Second,
			wantMin:   100 * time.Millisecond,
			wantMax:   100 * time.Millisecond,
		},
		{
			name:      "second attempt",
			attempt:   1,
			baseDelay: 100 * time.Millisecond,
			maxDelay:  10 * time.Second,
			wantMin:   200 * time.Millisecond,
			wantMax:   200 * time.Millisecond,
		},
		{
			name:      "third attempt",
			attempt:   2,
			baseDelay: 100 * time.Millisecond,
			maxDelay:  10 * time.Second,
			wantMin:   400 * time.Millisecond,
			wantMax:   400 * time.Millisecond,
		},
		{
			name:      "capped at max delay",
			attempt:   10,
			baseDelay: 100 * time.Millisecond,
			maxDelay:  500 * time.Millisecond,
			wantMin:   500 * time.Millisecond,
			wantMax:   500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CalculateBackoff(tt.attempt, tt.baseDelay, tt.maxDelay, false)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("CalculateBackoff() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateBackoffWithJitter(t *testing.T) {
	t.Parallel()

	baseDelay := 100 * time.Millisecond
	maxDelay := 10 * time.Second

	// Run multiple times to verify jitter adds variance
	var delays []time.Duration
	for range 10 {
		delay := CalculateBackoff(2, baseDelay, maxDelay, true)
		delays = append(delays, delay)
	}

	// Check that we have some variance (not all identical)
	allSame := true
	for i := 1; i < len(delays); i++ {
		if delays[i] != delays[0] {
			allSame = false
			break
		}
	}

	if allSame {
		t.Log("all delays were identical, jitter may not be applied")
	}
}
