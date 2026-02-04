package resilience

import (
	"context"
	"sync"
	"time"
)

// HealthCheckerConfig holds configuration for a health checker.
type HealthCheckerConfig struct {
	// Name is an optional identifier for the service being checked.
	Name string

	// Interval is the duration between periodic health checks.
	// Default: 30 seconds
	Interval time.Duration

	// Timeout is the maximum duration for a single health check.
	// Default: 5 seconds
	Timeout time.Duration

	// CheckFunc is the function that performs the actual health check.
	// It should return nil if the service is healthy, or an error otherwise.
	CheckFunc func(ctx context.Context) error

	// CircuitBreaker is an optional circuit breaker to integrate with.
	CircuitBreaker *CircuitBreaker

	// OnStatusChange is called when the health status changes.
	OnStatusChange func(from, to HealthStatus)
}

// HealthChecker monitors the health of a service.
type HealthChecker struct {
	config HealthCheckerConfig

	mu         sync.RWMutex
	status     HealthStatus
	lastCheck  time.Time
	lastError  error
	isRunning  bool
	cancelFunc context.CancelFunc
}

// NewHealthChecker creates a new HealthChecker with the given configuration.
func NewHealthChecker(config HealthCheckerConfig) *HealthChecker {
	// Apply defaults
	if config.Interval <= 0 {
		config.Interval = 30 * time.Second
	}
	if config.Timeout <= 0 {
		config.Timeout = 5 * time.Second
	}

	return &HealthChecker{
		config: config,
		status: StatusUnknown,
	}
}

// Check performs a health check and returns the current status.
func (hc *HealthChecker) Check(ctx context.Context) HealthStatus {
	// Check if context is already cancelled
	if ctx.Err() != nil {
		return StatusUnknown
	}

	// Apply timeout if configured
	if hc.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, hc.config.Timeout)
		defer cancel()
	}

	var newStatus HealthStatus
	var err error

	if hc.config.CheckFunc != nil {
		err = hc.config.CheckFunc(ctx)
		if err == nil {
			newStatus = StatusHealthy
		} else {
			newStatus = StatusUnhealthy
		}
	} else {
		newStatus = StatusUnknown
	}

	hc.mu.Lock()
	oldStatus := hc.status
	hc.status = newStatus
	hc.lastCheck = time.Now()
	hc.lastError = err
	hc.mu.Unlock()

	// Notify on status change
	if oldStatus != newStatus && oldStatus != StatusUnknown && hc.config.OnStatusChange != nil {
		hc.config.OnStatusChange(oldStatus, newStatus)
	}

	return newStatus
}

// Status returns the current health status without performing a check.
func (hc *HealthChecker) Status() HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.status
}

// LastCheck returns the time of the last health check.
func (hc *HealthChecker) LastCheck() time.Time {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.lastCheck
}

// LastError returns the error from the last health check, if any.
func (hc *HealthChecker) LastError() error {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.lastError
}

// Start begins periodic health checking.
// The checking continues until the context is cancelled.
func (hc *HealthChecker) Start(ctx context.Context) {
	hc.mu.Lock()
	if hc.isRunning {
		hc.mu.Unlock()
		return
	}
	hc.isRunning = true
	ctx, hc.cancelFunc = context.WithCancel(ctx)
	hc.mu.Unlock()

	go hc.runPeriodic(ctx)
}

// Stop stops periodic health checking.
func (hc *HealthChecker) Stop() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if hc.cancelFunc != nil {
		hc.cancelFunc()
		hc.cancelFunc = nil
	}
	hc.isRunning = false
}

// runPeriodic runs the health check loop.
func (hc *HealthChecker) runPeriodic(ctx context.Context) {
	// Perform initial check
	hc.Check(ctx)

	ticker := time.NewTicker(hc.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			hc.mu.Lock()
			hc.isRunning = false
			hc.mu.Unlock()
			return
		case <-ticker.C:
			hc.Check(ctx)
		}
	}
}
