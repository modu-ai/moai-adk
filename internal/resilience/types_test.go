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

