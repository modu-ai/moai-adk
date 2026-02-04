// Package resilience provides resilience patterns for external service integration.
// It includes circuit breaker, retry policy, health check, and resource monitoring.
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

// HealthStatus represents the health status of a service.
type HealthStatus string

const (
	// StatusHealthy indicates the service is healthy and operational.
	StatusHealthy HealthStatus = "healthy"
	// StatusUnhealthy indicates the service is unhealthy or unavailable.
	StatusUnhealthy HealthStatus = "unhealthy"
	// StatusUnknown indicates the health status has not been determined.
	StatusUnknown HealthStatus = "unknown"
)

// String returns the string representation of the health status.
func (s HealthStatus) String() string {
	return string(s)
}

// IsHealthy returns true if the status indicates a healthy state.
func (s HealthStatus) IsHealthy() bool {
	return s == StatusHealthy
}

// ResourceStats represents system resource usage statistics.
type ResourceStats struct {
	MemoryUsedMB   uint64  `json:"memoryUsedMB"`
	MemoryTotalMB  uint64  `json:"memoryTotalMB"`
	GoroutineCount int     `json:"goroutineCount"`
	CPUPercent     float64 `json:"cpuPercent"`
}

// MemoryUsagePercent returns the memory usage as a percentage.
// Returns 0 if MemoryTotalMB is zero to avoid division by zero.
func (s ResourceStats) MemoryUsagePercent() float64 {
	if s.MemoryTotalMB == 0 {
		return 0.0
	}
	return float64(s.MemoryUsedMB) / float64(s.MemoryTotalMB) * 100.0
}

// IsMemoryHigh returns true if memory usage exceeds the given threshold percentage.
func (s ResourceStats) IsMemoryHigh(threshold float64) bool {
	return s.MemoryUsagePercent() > threshold
}

// IsGoroutineHigh returns true if goroutine count exceeds the given threshold.
func (s ResourceStats) IsGoroutineHigh(threshold int) bool {
	return s.GoroutineCount > threshold
}

// ResourceThresholds holds threshold values for resource monitoring alerts.
type ResourceThresholds struct {
	MemoryPercent  float64 `json:"memoryPercent"`
	GoroutineCount int     `json:"goroutineCount"`
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

// ClientError represents an error caused by invalid client input.
// Client errors should not be retried.
type ClientError struct {
	message string
}

// NewClientError creates a new ClientError with the given message.
func NewClientError(message string) *ClientError {
	return &ClientError{message: message}
}

// Error implements the error interface.
func (e *ClientError) Error() string {
	return "client error: " + e.message
}

// IsClientError returns true, indicating this is a client error.
func (e *ClientError) IsClientError() bool {
	return true
}

// isClientError is an interface for detecting client errors.
type isClientError interface {
	IsClientError() bool
}
