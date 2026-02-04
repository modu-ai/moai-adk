package resilience

import (
	"testing"
)

func TestCircuitStateConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state CircuitState
		want  string
	}{
		{"StateClosed", StateClosed, "closed"},
		{"StateOpen", StateOpen, "open"},
		{"StateHalfOpen", StateHalfOpen, "half-open"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.state) != tt.want {
				t.Errorf("got %q, want %q", tt.state, tt.want)
			}
		})
	}
}

func TestCircuitStateString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state CircuitState
		want  string
	}{
		{"closed state", StateClosed, "closed"},
		{"open state", StateOpen, "open"},
		{"half-open state", StateHalfOpen, "half-open"},
		{"unknown state", CircuitState("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.state.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCircuitStateIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state CircuitState
		want  bool
	}{
		{"closed is valid", StateClosed, true},
		{"open is valid", StateOpen, true},
		{"half-open is valid", StateHalfOpen, true},
		{"empty is invalid", CircuitState(""), false},
		{"random is invalid", CircuitState("random"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.state.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHealthStatusConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status HealthStatus
		want   string
	}{
		{"StatusHealthy", StatusHealthy, "healthy"},
		{"StatusUnhealthy", StatusUnhealthy, "unhealthy"},
		{"StatusUnknown", StatusUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.status) != tt.want {
				t.Errorf("got %q, want %q", tt.status, tt.want)
			}
		})
	}
}

func TestHealthStatusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status HealthStatus
		want   string
	}{
		{"healthy status", StatusHealthy, "healthy"},
		{"unhealthy status", StatusUnhealthy, "unhealthy"},
		{"unknown status", StatusUnknown, "unknown"},
		{"custom status", HealthStatus("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHealthStatusIsHealthy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status HealthStatus
		want   bool
	}{
		{"healthy returns true", StatusHealthy, true},
		{"unhealthy returns false", StatusUnhealthy, false},
		{"unknown returns false", StatusUnknown, false},
		{"custom returns false", HealthStatus("custom"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.status.IsHealthy(); got != tt.want {
				t.Errorf("IsHealthy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceStatsZeroValue(t *testing.T) {
	t.Parallel()

	var stats ResourceStats
	if stats.MemoryUsedMB != 0 {
		t.Errorf("MemoryUsedMB: got %d, want 0", stats.MemoryUsedMB)
	}
	if stats.MemoryTotalMB != 0 {
		t.Errorf("MemoryTotalMB: got %d, want 0", stats.MemoryTotalMB)
	}
	if stats.GoroutineCount != 0 {
		t.Errorf("GoroutineCount: got %d, want 0", stats.GoroutineCount)
	}
	if stats.CPUPercent != 0.0 {
		t.Errorf("CPUPercent: got %f, want 0.0", stats.CPUPercent)
	}
}

func TestResourceStatsCreation(t *testing.T) {
	t.Parallel()

	stats := ResourceStats{
		MemoryUsedMB:   1024,
		MemoryTotalMB:  8192,
		GoroutineCount: 50,
		CPUPercent:     25.5,
	}

	if stats.MemoryUsedMB != 1024 {
		t.Errorf("MemoryUsedMB: got %d, want 1024", stats.MemoryUsedMB)
	}
	if stats.MemoryTotalMB != 8192 {
		t.Errorf("MemoryTotalMB: got %d, want 8192", stats.MemoryTotalMB)
	}
	if stats.GoroutineCount != 50 {
		t.Errorf("GoroutineCount: got %d, want 50", stats.GoroutineCount)
	}
	if stats.CPUPercent != 25.5 {
		t.Errorf("CPUPercent: got %f, want 25.5", stats.CPUPercent)
	}
}

func TestResourceStatsMemoryUsagePercent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		stats ResourceStats
		want  float64
	}{
		{
			name:  "50% usage",
			stats: ResourceStats{MemoryUsedMB: 4096, MemoryTotalMB: 8192},
			want:  50.0,
		},
		{
			name:  "25% usage",
			stats: ResourceStats{MemoryUsedMB: 2048, MemoryTotalMB: 8192},
			want:  25.0,
		},
		{
			name:  "zero total returns zero",
			stats: ResourceStats{MemoryUsedMB: 1024, MemoryTotalMB: 0},
			want:  0.0,
		},
		{
			name:  "zero values",
			stats: ResourceStats{},
			want:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.stats.MemoryUsagePercent(); got != tt.want {
				t.Errorf("MemoryUsagePercent() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestResourceStatsIsMemoryHigh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		stats     ResourceStats
		threshold float64
		want      bool
	}{
		{
			name:      "above 80% threshold",
			stats:     ResourceStats{MemoryUsedMB: 8500, MemoryTotalMB: 10000},
			threshold: 80.0,
			want:      true,
		},
		{
			name:      "below 80% threshold",
			stats:     ResourceStats{MemoryUsedMB: 7000, MemoryTotalMB: 10000},
			threshold: 80.0,
			want:      false,
		},
		{
			name:      "exactly at threshold",
			stats:     ResourceStats{MemoryUsedMB: 8000, MemoryTotalMB: 10000},
			threshold: 80.0,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.stats.IsMemoryHigh(tt.threshold); got != tt.want {
				t.Errorf("IsMemoryHigh(%f) = %v, want %v", tt.threshold, got, tt.want)
			}
		})
	}
}

func TestResourceStatsIsGoroutineHigh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		stats     ResourceStats
		threshold int
		want      bool
	}{
		{
			name:      "above 1000 threshold",
			stats:     ResourceStats{GoroutineCount: 1001},
			threshold: 1000,
			want:      true,
		},
		{
			name:      "below 1000 threshold",
			stats:     ResourceStats{GoroutineCount: 500},
			threshold: 1000,
			want:      false,
		},
		{
			name:      "exactly at threshold",
			stats:     ResourceStats{GoroutineCount: 1000},
			threshold: 1000,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.stats.IsGoroutineHigh(tt.threshold); got != tt.want {
				t.Errorf("IsGoroutineHigh(%d) = %v, want %v", tt.threshold, got, tt.want)
			}
		})
	}
}
