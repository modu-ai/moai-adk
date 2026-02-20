package resilience

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewHealthChecker(t *testing.T) {
	t.Parallel()

	checker := NewHealthChecker(HealthCheckerConfig{
		Interval: 30 * time.Second,
		CheckFunc: func(ctx context.Context) error {
			return nil
		},
	})

	if checker == nil {
		t.Fatal("NewHealthChecker returned nil")
	}
	if checker.Status() != StatusUnknown {
		t.Errorf("initial status = %v, want %v", checker.Status(), StatusUnknown)
	}
}

func TestHealthCheckerCheck(t *testing.T) {
	t.Parallel()

	checker := NewHealthChecker(HealthCheckerConfig{
		Interval: 30 * time.Second,
		CheckFunc: func(ctx context.Context) error {
			return nil
		},
	})

	ctx := context.Background()
	status := checker.Check(ctx)

	if status != StatusHealthy {
		t.Errorf("Check() = %v, want %v", status, StatusHealthy)
	}
	if checker.Status() != StatusHealthy {
		t.Errorf("Status() after Check() = %v, want %v", checker.Status(), StatusHealthy)
	}
}

func TestHealthCheckerCheckFailure(t *testing.T) {
	t.Parallel()

	checkErr := errors.New("service unavailable")
	checker := NewHealthChecker(HealthCheckerConfig{
		Interval: 30 * time.Second,
		CheckFunc: func(ctx context.Context) error {
			return checkErr
		},
	})

	ctx := context.Background()
	status := checker.Check(ctx)

	if status != StatusUnhealthy {
		t.Errorf("Check() = %v, want %v", status, StatusUnhealthy)
	}
	if checker.Status() != StatusUnhealthy {
		t.Errorf("Status() after failed Check() = %v, want %v", checker.Status(), StatusUnhealthy)
	}
}

func TestHealthCheckerLastCheck(t *testing.T) {
	t.Parallel()

	checker := NewHealthChecker(HealthCheckerConfig{
		Interval: 30 * time.Second,
		CheckFunc: func(ctx context.Context) error {
			return nil
		},
	})

	// Initially, LastCheck should return zero time
	if !checker.LastCheck().IsZero() {
		t.Errorf("LastCheck() before any check should be zero time")
	}

	ctx := context.Background()
	before := time.Now()
	_ = checker.Check(ctx)
	after := time.Now()

	lastCheck := checker.LastCheck()
	if lastCheck.Before(before) || lastCheck.After(after) {
		t.Errorf("LastCheck() = %v, want between %v and %v", lastCheck, before, after)
	}
}

func TestHealthCheckerContextCancellation(t *testing.T) {
	t.Parallel()

	checker := NewHealthChecker(HealthCheckerConfig{
		Interval: 30 * time.Second,
		CheckFunc: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(1 * time.Second):
				return nil
			}
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	status := checker.Check(ctx)

	if status != StatusUnknown {
		t.Errorf("Check() with cancelled context = %v, want %v", status, StatusUnknown)
	}
}

func TestHealthCheckerStartStop(t *testing.T) {
	t.Parallel()

	var checkCount int
	var mu sync.Mutex

	checker := NewHealthChecker(HealthCheckerConfig{
		Interval: 20 * time.Millisecond,
		CheckFunc: func(ctx context.Context) error {
			mu.Lock()
			checkCount++
			mu.Unlock()
			return nil
		},
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Start periodic checking
	checker.Start(ctx)

	// Wait for a few checks
	time.Sleep(100 * time.Millisecond)

	// Stop checking
	cancel()

	// Give time for goroutine to stop
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	countAtStop := checkCount
	mu.Unlock()

	// Wait a bit more and verify no more checks
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	finalCount := checkCount
	mu.Unlock()

	if countAtStop < 2 {
		t.Errorf("expected at least 2 checks, got %d", countAtStop)
	}
	if finalCount != countAtStop {
		t.Errorf("checks continued after stop: had %d, now have %d", countAtStop, finalCount)
	}
}

func TestHealthCheckerConcurrentAccess(t *testing.T) {
	t.Parallel()

	checker := NewHealthChecker(HealthCheckerConfig{
		Interval: 30 * time.Second,
		CheckFunc: func(ctx context.Context) error {
			return nil
		},
	})

	ctx := context.Background()
	var wg sync.WaitGroup
	numGoroutines := 50

	for range numGoroutines {
		wg.Go(func() {
			_ = checker.Check(ctx)
			_ = checker.Status()
			_ = checker.LastCheck()
		})
	}

	wg.Wait()
	// Test passes if no race conditions occur
}

func TestHealthCheckerWithCircuitBreaker(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: 5,
		Timeout:   30 * time.Second,
	})

	checker := NewHealthChecker(HealthCheckerConfig{
		Interval:       30 * time.Second,
		CircuitBreaker: cb,
		CheckFunc: func(ctx context.Context) error {
			return nil
		},
	})

	if checker == nil {
		t.Fatal("NewHealthChecker with CircuitBreaker returned nil")
	}
}

func TestHealthCheckerOnStatusChange(t *testing.T) {
	t.Parallel()

	var statusChanges []HealthStatus
	var mu sync.Mutex

	healthy := true
	checker := NewHealthChecker(HealthCheckerConfig{
		Interval: 30 * time.Second,
		CheckFunc: func(ctx context.Context) error {
			if healthy {
				return nil
			}
			return errors.New("unhealthy")
		},
		OnStatusChange: func(from, to HealthStatus) {
			mu.Lock()
			statusChanges = append(statusChanges, to)
			mu.Unlock()
		},
	})

	ctx := context.Background()

	// First check - healthy
	_ = checker.Check(ctx)

	// Change to unhealthy
	healthy = false
	_ = checker.Check(ctx)

	// Change back to healthy
	healthy = true
	_ = checker.Check(ctx)

	mu.Lock()
	defer mu.Unlock()

	// Should have recorded status changes
	if len(statusChanges) < 2 {
		t.Errorf("expected at least 2 status changes, got %d", len(statusChanges))
	}
}

func TestHealthCheckerTimeout(t *testing.T) {
	t.Parallel()

	checker := NewHealthChecker(HealthCheckerConfig{
		Interval: 30 * time.Second,
		Timeout:  50 * time.Millisecond,
		CheckFunc: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(1 * time.Second):
				return nil
			}
		},
	})

	ctx := context.Background()
	start := time.Now()
	status := checker.Check(ctx)
	elapsed := time.Since(start)

	// Should timeout around 50ms
	if elapsed > 200*time.Millisecond {
		t.Errorf("Check() took %v, expected timeout around 50ms", elapsed)
	}
	if status != StatusUnhealthy {
		t.Errorf("Check() = %v, want %v (should timeout)", status, StatusUnhealthy)
	}
}

func TestHealthCheckerMultipleServices(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	service1Healthy := true
	service2Healthy := true

	checker1 := NewHealthChecker(HealthCheckerConfig{
		Name:     "service1",
		Interval: 30 * time.Second,
		CheckFunc: func(ctx context.Context) error {
			if service1Healthy {
				return nil
			}
			return errors.New("service1 down")
		},
	})

	checker2 := NewHealthChecker(HealthCheckerConfig{
		Name:     "service2",
		Interval: 30 * time.Second,
		CheckFunc: func(ctx context.Context) error {
			if service2Healthy {
				return nil
			}
			return errors.New("service2 down")
		},
	})

	// Both healthy
	if checker1.Check(ctx) != StatusHealthy {
		t.Error("service1 should be healthy")
	}
	if checker2.Check(ctx) != StatusHealthy {
		t.Error("service2 should be healthy")
	}

	// Service1 becomes unhealthy
	service1Healthy = false
	if checker1.Check(ctx) != StatusUnhealthy {
		t.Error("service1 should be unhealthy")
	}
	if checker2.Check(ctx) != StatusHealthy {
		t.Error("service2 should still be healthy")
	}
}

func TestHealthCheckerLastError(t *testing.T) {
	t.Parallel()

	checkErr := errors.New("service unavailable")
	checker := NewHealthChecker(HealthCheckerConfig{
		Interval: 30 * time.Second,
		CheckFunc: func(ctx context.Context) error {
			return checkErr
		},
	})

	ctx := context.Background()

	// Before any check
	if checker.LastError() != nil {
		t.Error("LastError() before any check should be nil")
	}

	_ = checker.Check(ctx)

	if !errors.Is(checker.LastError(), checkErr) {
		t.Errorf("LastError() = %v, want %v", checker.LastError(), checkErr)
	}
}

func TestHealthCheckerStop(t *testing.T) {
	t.Parallel()

	var checkCount int
	var mu sync.Mutex

	checker := NewHealthChecker(HealthCheckerConfig{
		Interval: 20 * time.Millisecond,
		CheckFunc: func(ctx context.Context) error {
			mu.Lock()
			checkCount++
			mu.Unlock()
			return nil
		},
	})

	ctx := context.Background()

	// Start periodic checking
	checker.Start(ctx)

	// Wait for a few checks
	time.Sleep(100 * time.Millisecond)

	// Stop using Stop() method
	checker.Stop()

	// Give time for goroutine to stop
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	countAtStop := checkCount
	mu.Unlock()

	// Wait and verify no more checks
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	finalCount := checkCount
	mu.Unlock()

	if countAtStop < 2 {
		t.Errorf("expected at least 2 checks, got %d", countAtStop)
	}
	// Allow for at most 1 extra check due to timing race
	if finalCount > countAtStop+1 {
		t.Errorf("checks continued after Stop(): had %d, now have %d", countAtStop, finalCount)
	}
}

func TestHealthCheckerStopWithoutStart(t *testing.T) {
	t.Parallel()

	checker := NewHealthChecker(HealthCheckerConfig{
		Interval: 30 * time.Second,
		CheckFunc: func(ctx context.Context) error {
			return nil
		},
	})

	// Stop without starting should not panic
	checker.Stop()
}

func TestHealthCheckerDoubleStart(t *testing.T) {
	t.Parallel()

	var checkCount int
	var mu sync.Mutex

	checker := NewHealthChecker(HealthCheckerConfig{
		Interval: 20 * time.Millisecond,
		CheckFunc: func(ctx context.Context) error {
			mu.Lock()
			checkCount++
			mu.Unlock()
			return nil
		},
	})

	ctx := t.Context()

	// Start twice - should only have one monitoring goroutine
	checker.Start(ctx)
	checker.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := checkCount
	mu.Unlock()

	// Should have reasonable number of checks (not doubled)
	if count > 10 {
		t.Errorf("too many checks (%d), double start may have created multiple goroutines", count)
	}
}
