// Package resilience provides resilience patterns for external service integration.
// It includes a circuit breaker for preventing cascading failures.
package resilience

import "errors"

// CircuitState represents the state of a circuit breaker.
type CircuitState string

const (
	// StateClosed indicates normal operation - requests are allowed.
	StateClosed CircuitState = "closed"
	// StateOpen indicates the circuit is open - requests are rejected immediately.
	StateOpen CircuitState = "open"
	// StateHalfOpen indicates the circuit is testing recovery - one request is allowed.
	StateHalfOpen CircuitState = "half-open"
)

// String returns the string representation of the circuit state.
func (s CircuitState) String() string {
	return string(s)
}

// IsValid returns true if the circuit state is a valid value.
func (s CircuitState) IsValid() bool {
	switch s {
	case StateClosed, StateOpen, StateHalfOpen:
		return true
	default:
		return false
	}
}

// CircuitBreakerMetrics holds metrics for a circuit breaker.
type CircuitBreakerMetrics struct {
	TotalCalls    int64 `json:"totalCalls"`
	SuccessCount  int64 `json:"successCount"`
	FailureCount  int64 `json:"failureCount"`
	RejectedCount int64 `json:"rejectedCount"`
}

// Sentinel errors for the resilience package.
var (
	// ErrCircuitOpen is returned when the circuit breaker is in open state.
	ErrCircuitOpen = errors.New("circuit breaker is open")
)
