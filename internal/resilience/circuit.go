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

	// Metrics tracked atomically
	totalCalls    atomic.Int64
	successCount  atomic.Int64
	failureTotal  atomic.Int64
	rejectedCount atomic.Int64
}

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

	// Check and potentially update state
	cb.mu.Lock()
	state := cb.checkState()
	cb.mu.Unlock()

	if state == StateOpen {
		cb.rejectedCount.Add(1)
		return ErrCircuitOpen
	}

	// Execute the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

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
		// Call asynchronously to avoid blocking
		go cb.config.OnStateChange(oldState, newState)
	}
}
