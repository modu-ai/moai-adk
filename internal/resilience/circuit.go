package resilience

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// CircuitBreakerConfig holds configuration for a circuit breaker.
type CircuitBreakerConfig struct {
	// Threshold is the number of consecutive failures before opening the circuit.
	// Default: 5
	Threshold int

	// Timeout is the duration the circuit stays open before transitioning to half-open.
	// Default: 30 seconds
	Timeout time.Duration

	// OnStateChange is called when the circuit state changes.
	OnStateChange func(from, to CircuitState)
}

// CircuitBreaker implements the circuit breaker pattern.
// It prevents cascading failures by failing fast when a service is unavailable.
type CircuitBreaker struct {
	config CircuitBreakerConfig

	mu              sync.RWMutex
	state           CircuitState
	failureCount    int
	lastFailureTime time.Time
	lastStateChange time.Time

	// halfOpenInFlight guards the half-open single-trial invariant: while the
	// circuit is half-open, at most one trial request may execute fn(); all other
	// concurrent callers are rejected with ErrCircuitOpen until the trial resolves.
	// Guarded by cb.mu. Cleared on every Call resolution and on Reset.
	// SPEC-SEC-HARDEN-001 §M5 REQ-SEC-M5-001/002.
	halfOpenInFlight bool

	// Metrics tracked atomically
	totalCalls    atomic.Int64
	successCount  atomic.Int64
	failureTotal  atomic.Int64
	rejectedCount atomic.Int64
}

// @MX:ANCHOR: [AUTO] Circuit breaker pattern factory - core entry point for preventing failure propagation
// @MX:REASON: [AUTO] fan_in=3+ (HealthChecker, external callers, etc.); all circuit breaker instances are created through this function
// NewCircuitBreaker creates a new CircuitBreaker with the given configuration.
// Default values are applied if not specified.
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	// Apply defaults
	if config.Threshold <= 0 {
		config.Threshold = 5
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// Call executes the given function with circuit breaker protection.
// If the circuit is open, it returns ErrCircuitOpen immediately.
// If the circuit is half-open, only one request is allowed through.
func (cb *CircuitBreaker) Call(ctx context.Context, fn func() error) error {
	// Check context first
	if err := ctx.Err(); err != nil {
		return err
	}

	cb.totalCalls.Add(1)

	// Check and potentially update state, and atomically claim the half-open trial
	// permit under the same lock so exactly one caller proceeds while half-open.
	cb.mu.Lock()
	state := cb.checkState()
	if state == StateOpen {
		cb.mu.Unlock()
		cb.rejectedCount.Add(1)
		return ErrCircuitOpen
	}
	if state == StateHalfOpen {
		// SPEC-SEC-HARDEN-001 §M5: admit at most one in-flight half-open trial.
		if cb.halfOpenInFlight {
			cb.mu.Unlock()
			cb.rejectedCount.Add(1)
			return ErrCircuitOpen
		}
		cb.halfOpenInFlight = true
	}
	cb.mu.Unlock()

	// Execute the function (lock-free, preserving the existing non-blocking model).
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Release the half-open trial permit on every resolution path (success or
	// failure). The subsequent recordSuccess/recordFailure transitions out of
	// half-open; clearing here also covers the case where the breaker is still
	// half-open after this call (no transition fired).
	cb.halfOpenInFlight = false

	if err != nil {
		cb.recordFailure()
		cb.failureTotal.Add(1)
	} else {
		cb.recordSuccess()
		cb.successCount.Add(1)
	}

	return err
}

// State returns the current state of the circuit breaker.
// It also handles state transitions based on timeout.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.checkState()
}

// Reset resets the circuit breaker to the closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	oldState := cb.state
	cb.state = StateClosed
	cb.failureCount = 0
	cb.lastFailureTime = time.Time{}
	cb.lastStateChange = time.Now()
	// Clear any leftover half-open trial permit on reset (SPEC-SEC-HARDEN-001 §M5).
	cb.halfOpenInFlight = false

	// NOTE: Reset's OnStateChange is invoked SYNCHRONOUSLY by design (distinct from
	// the asynchronous transitionTo goroutine). AC-SEC-M5-006 requires this path
	// unchanged, so M5 leaves it as-is (no recover wrapper, no async dispatch).
	if oldState != StateClosed && cb.config.OnStateChange != nil {
		cb.config.OnStateChange(oldState, StateClosed)
	}
}

// Metrics returns the current metrics for the circuit breaker.
func (cb *CircuitBreaker) Metrics() CircuitBreakerMetrics {
	return CircuitBreakerMetrics{
		TotalCalls:    cb.totalCalls.Load(),
		SuccessCount:  cb.successCount.Load(),
		FailureCount:  cb.failureTotal.Load(),
		RejectedCount: cb.rejectedCount.Load(),
	}
}

// checkState checks the current state and handles state transitions.
// Must be called with the mutex held.
func (cb *CircuitBreaker) checkState() CircuitState {
	switch cb.state {
	case StateOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailureTime) >= cb.config.Timeout {
			cb.transitionTo(StateHalfOpen)
		}
	case StateHalfOpen:
		// State transitions happen in recordSuccess/recordFailure
	case StateClosed:
		// Normal operation
	}
	return cb.state
}

// recordFailure records a failure and potentially opens the circuit.
// Must be called with the mutex held.
func (cb *CircuitBreaker) recordFailure() {
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failureCount++
		if cb.failureCount >= cb.config.Threshold {
			cb.transitionTo(StateOpen)
		}
	case StateHalfOpen:
		// Single failure in half-open reopens the circuit
		cb.transitionTo(StateOpen)
	}
}

// recordSuccess records a success and potentially closes the circuit.
// Must be called with the mutex held.
func (cb *CircuitBreaker) recordSuccess() {
	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failureCount = 0
	case StateHalfOpen:
		// Success in half-open closes the circuit
		cb.transitionTo(StateClosed)
		cb.failureCount = 0
	}
}

// transitionTo transitions to the given state.
// Must be called with the mutex held.
func (cb *CircuitBreaker) transitionTo(newState CircuitState) {
	if cb.state == newState {
		return
	}

	oldState := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()

	if cb.config.OnStateChange != nil {
		// @MX:NOTE: [AUTO] OnStateChange callback invoked as a goroutine while the mutex is held by the caller; the goroutine body recovers from panics so a panicking callback no longer crashes the process (SPEC-SEC-HARDEN-001 §M5)
		// Call asynchronously to avoid blocking. The recover wrapper ensures a
		// panic inside the user-supplied callback is contained within this
		// goroutine and does not propagate to the Go runtime (which would abort
		// the host process). Context cancellation is intentionally not propagated
		// here (out of M5 scope; preserved existing dispatch shape).
		onStateChange := cb.config.OnStateChange
		go func() {
			defer func() { _ = recover() }()
			onStateChange(oldState, newState)
		}()
	}
}
