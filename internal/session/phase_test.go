package session

import "testing"

func TestPhaseString(t *testing.T) {
	tests := []struct {
		name     string
		phase    Phase
		expected string
	}{
		{"plan phase", PhasePlan, "plan"},
		{"run phase", PhaseRun, "run"},
		{"sync phase", PhaseSync, "sync"},
		{"design phase", PhaseDesign, "design"},
		{"review phase", PhaseReview, "review"},
		{"fix phase", PhaseFix, "fix"},
		{"loop phase", PhaseLoop, "loop"},
		{"db phase", PhaseDB, "db"},
		{"mx phase", PhaseMX, "mx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.phase.String(); got != tt.expected {
				t.Errorf("Phase.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPhaseValid(t *testing.T) {
	validPhases := []Phase{
		PhasePlan, PhaseRun, PhaseSync, PhaseDesign, PhaseReview,
		PhaseFix, PhaseLoop, PhaseDB, PhaseMX,
	}

	for _, phase := range validPhases {
		t.Run(phase.String()+" valid", func(t *testing.T) {
			if !phase.Valid() {
				t.Errorf("Phase.Valid() = false, want true for %v", phase)
			}
		})
	}

	invalidPhases := []Phase{
		Phase("invalid"),
		Phase(""),
		Phase("Plan"), // case sensitive
	}

	for _, phase := range invalidPhases {
		t.Run(string(phase)+" invalid", func(t *testing.T) {
			if phase.Valid() {
				t.Errorf("Phase.Valid() = true, want false for %v", phase)
			}
		})
	}
}
