package foundation

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/pkg/models"
)

func TestPhaseString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Phase
		want string
	}{
		{name: "Analyze", p: PhaseAnalyze, want: "analyze"},
		{name: "Preserve", p: PhasePreserve, want: "preserve"},
		{name: "Improve", p: PhaseImprove, want: "improve"},
		{name: "Red", p: PhaseRed, want: "red"},
		{name: "Green", p: PhaseGreen, want: "green"},
		{name: "Refactor", p: PhaseRefactor, want: "refactor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.p.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPhaseIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Phase
		want bool
	}{
		{name: "Analyze", p: PhaseAnalyze, want: true},
		{name: "Preserve", p: PhasePreserve, want: true},
		{name: "Improve", p: PhaseImprove, want: true},
		{name: "Red", p: PhaseRed, want: true},
		{name: "Green", p: PhaseGreen, want: true},
		{name: "Refactor", p: PhaseRefactor, want: true},
		{name: "empty", p: Phase(""), want: false},
		{name: "invalid", p: Phase("invalid"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.p.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCycleTypeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		c    CycleType
		want string
	}{
		{name: "DDD", c: CycleDDD, want: "ddd"},
		{name: "TDD", c: CycleTDD, want: "tdd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.c.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDDDCyclePhases(t *testing.T) {
	t.Parallel()

	c := &DDDCycle{}
	phases := c.Phases()

	if len(phases) != 3 {
		t.Fatalf("DDDCycle.Phases() returned %d phases, want 3", len(phases))
	}

	expected := []Phase{PhaseAnalyze, PhasePreserve, PhaseImprove}
	for i, p := range phases {
		if p != expected[i] {
			t.Errorf("phase[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestDDDCycleValidateTransition(t *testing.T) {
	t.Parallel()

	c := &DDDCycle{}

	tests := []struct {
		name    string
		from    Phase
		to      Phase
		wantErr bool
	}{
		{name: "Analyze_to_Preserve", from: PhaseAnalyze, to: PhasePreserve, wantErr: false},
		{name: "Preserve_to_Improve", from: PhasePreserve, to: PhaseImprove, wantErr: false},
		{name: "Improve_to_Analyze_cycle", from: PhaseImprove, to: PhaseAnalyze, wantErr: false},
		{name: "Analyze_to_Improve_skip", from: PhaseAnalyze, to: PhaseImprove, wantErr: true},
		{name: "Improve_to_Preserve_backward", from: PhaseImprove, to: PhasePreserve, wantErr: true},
		{name: "invalid_from_phase", from: PhaseRed, to: PhasePreserve, wantErr: true},
		{name: "invalid_to_phase", from: PhaseAnalyze, to: PhaseGreen, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := c.ValidateTransition(tt.from, tt.to)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantErr && err != nil {
				if !errors.Is(err, ErrInvalidPhaseTransition) {
					t.Errorf("error should wrap ErrInvalidPhaseTransition, got: %v", err)
				}
			}
		})
	}
}

func TestTDDCyclePhases(t *testing.T) {
	t.Parallel()

	c := &TDDCycle{}
	phases := c.Phases()

	if len(phases) != 3 {
		t.Fatalf("TDDCycle.Phases() returned %d phases, want 3", len(phases))
	}

	expected := []Phase{PhaseRed, PhaseGreen, PhaseRefactor}
	for i, p := range phases {
		if p != expected[i] {
			t.Errorf("phase[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestTDDCycleValidateTransition(t *testing.T) {
	t.Parallel()

	c := &TDDCycle{}

	tests := []struct {
		name    string
		from    Phase
		to      Phase
		wantErr bool
	}{
		{name: "Red_to_Green", from: PhaseRed, to: PhaseGreen, wantErr: false},
		{name: "Green_to_Refactor", from: PhaseGreen, to: PhaseRefactor, wantErr: false},
		{name: "Refactor_to_Red_cycle", from: PhaseRefactor, to: PhaseRed, wantErr: false},
		{name: "Red_to_Refactor_skip", from: PhaseRed, to: PhaseRefactor, wantErr: true},
		{name: "Refactor_to_Green_backward", from: PhaseRefactor, to: PhaseGreen, wantErr: true},
		{name: "invalid_from_phase", from: PhaseAnalyze, to: PhaseGreen, wantErr: true},
		{name: "invalid_to_phase", from: PhaseRed, to: PhaseImprove, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := c.ValidateTransition(tt.from, tt.to)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantErr && err != nil {
				if !errors.Is(err, ErrInvalidPhaseTransition) {
					t.Errorf("error should wrap ErrInvalidPhaseTransition, got: %v", err)
				}
			}
		})
	}
}

func TestNewMethodology(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mode    models.DevelopmentMode
		wantErr bool
		hasDDD  bool
		hasTDD  bool
	}{
		{
			name:    "DDD_mode",
			mode:    models.ModeDDD,
			wantErr: false,
			hasDDD:  true,
			hasTDD:  false,
		},
		{
			name:    "TDD_mode",
			mode:    models.ModeTDD,
			wantErr: false,
			hasDDD:  false,
			hasTDD:  true,
		},
		{
			name:    "invalid_mode",
			mode:    models.DevelopmentMode("invalid"),
			wantErr: true,
		},
		{
			name:    "empty_mode",
			mode:    models.DevelopmentMode(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg, err := NewMethodology(tt.mode)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Mode != tt.mode {
				t.Errorf("Mode = %q, want %q", cfg.Mode, tt.mode)
			}
			if (cfg.DDD != nil) != tt.hasDDD {
				t.Errorf("DDD present = %v, want %v", cfg.DDD != nil, tt.hasDDD)
			}
			if (cfg.TDD != nil) != tt.hasTDD {
				t.Errorf("TDD present = %v, want %v", cfg.TDD != nil, tt.hasTDD)
			}
		})
	}
}

func TestMethodologyConfigActivePhases(t *testing.T) {
	t.Parallel()

	t.Run("DDD_phases", func(t *testing.T) {
		t.Parallel()
		cfg, err := NewMethodology(models.ModeDDD)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		phases := cfg.ActivePhases()
		if len(phases) != 3 {
			t.Fatalf("DDD ActivePhases() returned %d, want 3", len(phases))
		}
		expected := []Phase{PhaseAnalyze, PhasePreserve, PhaseImprove}
		for i, p := range phases {
			if p != expected[i] {
				t.Errorf("phase[%d] = %q, want %q", i, p, expected[i])
			}
		}
	})

	t.Run("TDD_phases", func(t *testing.T) {
		t.Parallel()
		cfg, err := NewMethodology(models.ModeTDD)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		phases := cfg.ActivePhases()
		if len(phases) != 3 {
			t.Fatalf("TDD ActivePhases() returned %d, want 3", len(phases))
		}
		expected := []Phase{PhaseRed, PhaseGreen, PhaseRefactor}
		for i, p := range phases {
			if p != expected[i] {
				t.Errorf("phase[%d] = %q, want %q", i, p, expected[i])
			}
		}
	})

}

func TestDDDCycleFullLoop(t *testing.T) {
	t.Parallel()

	c := &DDDCycle{}
	phases := c.Phases()

	// Validate full cycle: Analyze -> Preserve -> Improve -> Analyze.
	for i := range phases {
		from := phases[i]
		to := phases[(i+1)%len(phases)]
		if err := c.ValidateTransition(from, to); err != nil {
			t.Errorf("transition %s -> %s should be valid: %v", from, to, err)
		}
	}
}

func TestTDDCycleFullLoop(t *testing.T) {
	t.Parallel()

	c := &TDDCycle{}
	phases := c.Phases()

	// Validate full cycle: Red -> Green -> Refactor -> Red.
	for i := range phases {
		from := phases[i]
		to := phases[(i+1)%len(phases)]
		if err := c.ValidateTransition(from, to); err != nil {
			t.Errorf("transition %s -> %s should be valid: %v", from, to, err)
		}
	}
}
