// Package resilience provides a circuit breaker pattern for external service integration.
//
// # Circuit Breaker (REQ-HOOK-300~303)
//
// The circuit breaker prevents cascading failures by failing fast when a service
// is unavailable. It has three states: Closed (normal operation), Open (failing fast),
// and Half-Open (testing recovery).
//
// Example usage:
//
//	cb := resilience.NewCircuitBreaker(resilience.CircuitBreakerConfig{
//	    Threshold: 5,
//	    Timeout:   30 * time.Second,
//	})
//	err := cb.Call(ctx, func() error {
//	    return externalService.Call()
//	})
package resilience
