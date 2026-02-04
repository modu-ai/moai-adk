// Package resilience provides resilience patterns for external service integration.
//
// This package implements four main resilience patterns based on SPEC-HOOK-006:
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
//
// # Retry Policy (REQ-HOOK-310~313)
//
// The retry policy handles transient failures with exponential backoff and jitter.
// Client errors (invalid input) are not retried.
//
// Example usage:
//
//	policy := resilience.RetryPolicy{
//	    MaxRetries: 3,
//	    BaseDelay:  100 * time.Millisecond,
//	    MaxDelay:   5 * time.Second,
//	    UseJitter:  true,
//	}
//	err := resilience.Retry(ctx, policy, func() error {
//	    return externalService.Call()
//	})
//
// # Health Checker (REQ-HOOK-320~322)
//
// The health checker monitors service health with configurable intervals.
//
// Example usage:
//
//	checker := resilience.NewHealthChecker(resilience.HealthCheckerConfig{
//	    Interval: 30 * time.Second,
//	    CheckFunc: func(ctx context.Context) error {
//	        return externalService.Ping()
//	    },
//	})
//	status := checker.Check(ctx)
//
// # Resource Monitor (REQ-HOOK-330~332)
//
// The resource monitor tracks system resources (memory, goroutines) and triggers
// alerts when thresholds are exceeded.
//
// Example usage:
//
//	monitor := resilience.NewResourceMonitor(resilience.ResourceMonitorConfig{
//	    MemoryThreshold:    80.0,
//	    GoroutineThreshold: 1000,
//	    OnHighMemory: func(stats resilience.ResourceStats) {
//	        log.Warn("High memory usage detected")
//	    },
//	})
//	monitor.StartMonitoring(ctx, 30 * time.Second)
package resilience
