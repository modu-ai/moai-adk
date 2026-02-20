package resilience

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewCircuitBreaker(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: 5,
		Timeout:   30 * time.Second,
	})

	if cb == nil {
		t.Fatal("NewCircuitBreaker returned nil")
	}
	if cb.State() != StateClosed {
		t.Errorf("initial state = %v, want %v", cb.State(), StateClosed)
	}
}

func TestNewCircuitBreakerDefaults(t *testing.T) {
	t.Parallel()

	// Test with zero values to verify defaults are applied
	cb := NewCircuitBreaker(CircuitBreakerConfig{})

	if cb == nil {
		t.Fatal("NewCircuitBreaker returned nil")
	}
	if cb.State() != StateClosed {
		t.Errorf("initial state = %v, want %v", cb.State(), StateClosed)
	}
}

func TestCircuitBreakerCallSuccess(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: 5,
		Timeout:   30 * time.Second,
	})

	ctx := context.Background()
	err := cb.Call(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Call() error = %v, want nil", err)
	}
	if cb.State() != StateClosed {
		t.Errorf("state after success = %v, want %v", cb.State(), StateClosed)
	}
}

func TestCircuitBreakerCallFailure(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: 5,
		Timeout:   30 * time.Second,
	})

	expectedErr := errors.New("operation failed")
	ctx := context.Background()
	err := cb.Call(ctx, func() error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("Call() error = %v, want %v", err, expectedErr)
	}
	// Should still be closed after one failure
	if cb.State() != StateClosed {
		t.Errorf("state after one failure = %v, want %v", cb.State(), StateClosed)
	}
}

func TestCircuitBreakerOpensAfterThreshold(t *testing.T) {
	t.Parallel()

	threshold := 5
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: threshold,
		Timeout:   30 * time.Second,
	})

	ctx := context.Background()
	operationErr := errors.New("operation failed")

	// Fail threshold times
	for range threshold {
		_ = cb.Call(ctx, func() error {
			return operationErr
		})
	}

	// Circuit should now be open
	if cb.State() != StateOpen {
		t.Errorf("state after %d failures = %v, want %v", threshold, cb.State(), StateOpen)
	}
}

func TestCircuitBreakerOpenRejectsImmediately(t *testing.T) {
	t.Parallel()

	threshold := 5
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: threshold,
		Timeout:   30 * time.Second,
	})

	ctx := context.Background()
	operationErr := errors.New("operation failed")

	// Open the circuit
	for range threshold {
		_ = cb.Call(ctx, func() error {
			return operationErr
		})
	}

	// Track if function was called
	called := false
	err := cb.Call(ctx, func() error {
		called = true
		return nil
	})

	if called {
		t.Error("function was called when circuit is open")
	}
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("Call() error = %v, want %v", err, ErrCircuitOpen)
	}
}

func TestCircuitBreakerHalfOpenAfterTimeout(t *testing.T) {
	t.Parallel()

	threshold := 2
	timeout := 50 * time.Millisecond
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: threshold,
		Timeout:   timeout,
	})

	ctx := context.Background()
	operationErr := errors.New("operation failed")

	// Open the circuit
	for range threshold {
		_ = cb.Call(ctx, func() error {
			return operationErr
		})
	}

	if cb.State() != StateOpen {
		t.Fatalf("expected open state, got %v", cb.State())
	}

	// Wait for timeout
	time.Sleep(timeout + 10*time.Millisecond)

	// Circuit should transition to half-open on next call
	if cb.State() != StateHalfOpen {
		t.Errorf("state after timeout = %v, want %v", cb.State(), StateHalfOpen)
	}
}

func TestCircuitBreakerHalfOpenSuccessCloses(t *testing.T) {
	t.Parallel()

	threshold := 2
	timeout := 50 * time.Millisecond
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: threshold,
		Timeout:   timeout,
	})

	ctx := context.Background()
	operationErr := errors.New("operation failed")

	// Open the circuit
	for range threshold {
		_ = cb.Call(ctx, func() error {
			return operationErr
		})
	}

	// Wait for timeout
	time.Sleep(timeout + 10*time.Millisecond)

	// Successful call should close circuit
	err := cb.Call(ctx, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Call() error = %v, want nil", err)
	}
	if cb.State() != StateClosed {
		t.Errorf("state after successful half-open call = %v, want %v", cb.State(), StateClosed)
	}
}

func TestCircuitBreakerHalfOpenFailureReopens(t *testing.T) {
	t.Parallel()

	threshold := 2
	timeout := 50 * time.Millisecond
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: threshold,
		Timeout:   timeout,
	})

	ctx := context.Background()
	operationErr := errors.New("operation failed")

	// Open the circuit
	for range threshold {
		_ = cb.Call(ctx, func() error {
			return operationErr
		})
	}

	// Wait for timeout
	time.Sleep(timeout + 10*time.Millisecond)

	// Failed call should reopen circuit
	_ = cb.Call(ctx, func() error {
		return operationErr
	})

	if cb.State() != StateOpen {
		t.Errorf("state after failed half-open call = %v, want %v", cb.State(), StateOpen)
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	t.Parallel()

	threshold := 2
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: threshold,
		Timeout:   30 * time.Second,
	})

	ctx := context.Background()
	operationErr := errors.New("operation failed")

	// Open the circuit
	for range threshold {
		_ = cb.Call(ctx, func() error {
			return operationErr
		})
	}

	if cb.State() != StateOpen {
		t.Fatalf("expected open state, got %v", cb.State())
	}

	// Reset the circuit
	cb.Reset()

	if cb.State() != StateClosed {
		t.Errorf("state after Reset() = %v, want %v", cb.State(), StateClosed)
	}

	// Should be able to call again
	err := cb.Call(ctx, func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Call() after Reset() error = %v, want nil", err)
	}
}

func TestCircuitBreakerContextCancellation(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: 5,
		Timeout:   30 * time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := cb.Call(ctx, func() error {
		return nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Call() error = %v, want context.Canceled", err)
	}
}

func TestCircuitBreakerConcurrentAccess(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: 100, // High threshold to avoid opening
		Timeout:   30 * time.Second,
	})

	ctx := context.Background()
	var wg sync.WaitGroup
	numGoroutines := 100

	var successCount atomic.Int32

	for range numGoroutines {
		wg.Go(func() {
			err := cb.Call(ctx, func() error {
				return nil
			})
			if err == nil {
				successCount.Add(1)
			}
		})
	}

	wg.Wait()

	if successCount.Load() != int32(numGoroutines) {
		t.Errorf("successful calls = %d, want %d", successCount.Load(), numGoroutines)
	}
}

func TestCircuitBreakerSuccessResetsFailureCount(t *testing.T) {
	t.Parallel()

	threshold := 5
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: threshold,
		Timeout:   30 * time.Second,
	})

	ctx := context.Background()
	operationErr := errors.New("operation failed")

	// Fail threshold-1 times
	for i := 0; i < threshold-1; i++ {
		_ = cb.Call(ctx, func() error {
			return operationErr
		})
	}

	// Succeed once - should reset counter
	_ = cb.Call(ctx, func() error {
		return nil
	})

	// Fail threshold-1 more times
	for i := 0; i < threshold-1; i++ {
		_ = cb.Call(ctx, func() error {
			return operationErr
		})
	}

	// Circuit should still be closed
	if cb.State() != StateClosed {
		t.Errorf("state should be closed, got %v", cb.State())
	}
}

func TestCircuitBreakerMetrics(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: 5,
		Timeout:   30 * time.Second,
	})

	ctx := context.Background()

	// Make some calls
	_ = cb.Call(ctx, func() error { return nil })
	_ = cb.Call(ctx, func() error { return nil })
	_ = cb.Call(ctx, func() error { return errors.New("fail") })

	metrics := cb.Metrics()

	if metrics.TotalCalls != 3 {
		t.Errorf("TotalCalls = %d, want 3", metrics.TotalCalls)
	}
	if metrics.SuccessCount != 2 {
		t.Errorf("SuccessCount = %d, want 2", metrics.SuccessCount)
	}
	if metrics.FailureCount != 1 {
		t.Errorf("FailureCount = %d, want 1", metrics.FailureCount)
	}
}

func TestCircuitBreakerOnStateChange(t *testing.T) {
	t.Parallel()

	var stateChanges []CircuitState
	var mu sync.Mutex

	threshold := 2
	timeout := 50 * time.Millisecond
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: threshold,
		Timeout:   timeout,
		OnStateChange: func(from, to CircuitState) {
			mu.Lock()
			stateChanges = append(stateChanges, to)
			mu.Unlock()
		},
	})

	ctx := context.Background()
	operationErr := errors.New("operation failed")

	// Open the circuit
	for range threshold {
		_ = cb.Call(ctx, func() error {
			return operationErr
		})
	}

	// Wait for async callback and timeout to transition to half-open
	time.Sleep(timeout + 20*time.Millisecond)
	_ = cb.State() // Trigger state check

	// Wait for half-open transition callback
	time.Sleep(20 * time.Millisecond)

	// Close the circuit with success
	_ = cb.Call(ctx, func() error {
		return nil
	})

	// Wait for async callback to complete
	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(stateChanges) < 2 {
		t.Errorf("expected at least 2 state changes, got %d: %v", len(stateChanges), stateChanges)
	}
}

// Test error sentinel
func TestErrCircuitOpen(t *testing.T) {
	t.Parallel()

	if ErrCircuitOpen == nil {
		t.Fatal("ErrCircuitOpen should not be nil")
	}
	if ErrCircuitOpen.Error() != "circuit breaker is open" {
		t.Errorf("ErrCircuitOpen.Error() = %q, want %q", ErrCircuitOpen.Error(), "circuit breaker is open")
	}
}
