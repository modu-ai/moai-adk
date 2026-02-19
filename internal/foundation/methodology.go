package foundation

import (
	"fmt"

	"github.com/modu-ai/moai-adk/pkg/models"
)

// Phase represents a single step within a development methodology cycle.
type Phase string

const (
	// DDD phases (ANALYZE-PRESERVE-IMPROVE).

	// PhaseAnalyze reads existing code, identifies dependencies, and maps domain boundaries.
	PhaseAnalyze Phase = "analyze"

	// PhasePreserve writes characterization tests and captures current behavior.
	PhasePreserve Phase = "preserve"

	// PhaseImprove makes incremental changes and runs tests after each change.
	PhaseImprove Phase = "improve"

	// TDD phases (RED-GREEN-REFACTOR).

	// PhaseRed writes a failing test that defines desired behavior.
	PhaseRed Phase = "red"

	// PhaseGreen writes minimal code to make the failing test pass.
	PhaseGreen Phase = "green"

	// PhaseRefactor improves the code while keeping all tests passing.
	PhaseRefactor Phase = "refactor"
)

// String returns the string representation of the Phase.
func (p Phase) String() string {
	return string(p)
}

// IsValid checks whether the Phase is one of the defined constants.
func (p Phase) IsValid() bool {
	switch p {
	case PhaseAnalyze, PhasePreserve, PhaseImprove,
		PhaseRed, PhaseGreen, PhaseRefactor:
		return true
	}
	return false
}

// CycleType identifies the type of development cycle.
type CycleType string

const (
	// CycleDDD represents the Domain-Driven Development cycle.
	CycleDDD CycleType = "ddd"

	// CycleTDD represents the Test-Driven Development cycle.
	CycleTDD CycleType = "tdd"
)

// String returns the string representation of the CycleType.
func (c CycleType) String() string {
	return string(c)
}

// DDDCycle represents the DDD development cycle: Analyze -> Preserve -> Improve.
type DDDCycle struct{}

// Phases returns the ordered phases of the DDD cycle.
func (c *DDDCycle) Phases() []Phase {
	return []Phase{PhaseAnalyze, PhasePreserve, PhaseImprove}
}

// ValidateTransition checks if the transition from one phase to another is valid.
// Valid DDD transitions: Analyze->Preserve, Preserve->Improve, Improve->Analyze (cycle).
func (c *DDDCycle) ValidateTransition(from, to Phase) error {
	valid := map[Phase]Phase{
		PhaseAnalyze:  PhasePreserve,
		PhasePreserve: PhaseImprove,
		PhaseImprove:  PhaseAnalyze,
	}
	expected, ok := valid[from]
	if !ok {
		return fmt.Errorf("%w: %s is not a valid DDD phase", ErrInvalidPhaseTransition, from)
	}
	if expected != to {
		return fmt.Errorf("%w: cannot transition from %s to %s (expected %s)", ErrInvalidPhaseTransition, from, to, expected)
	}
	return nil
}

// TDDCycle represents the TDD development cycle: Red -> Green -> Refactor.
type TDDCycle struct{}

// Phases returns the ordered phases of the TDD cycle.
func (c *TDDCycle) Phases() []Phase {
	return []Phase{PhaseRed, PhaseGreen, PhaseRefactor}
}

// ValidateTransition checks if the transition from one phase to another is valid.
// Valid TDD transitions: Red->Green, Green->Refactor, Refactor->Red (cycle).
func (c *TDDCycle) ValidateTransition(from, to Phase) error {
	valid := map[Phase]Phase{
		PhaseRed:      PhaseGreen,
		PhaseGreen:    PhaseRefactor,
		PhaseRefactor: PhaseRed,
	}
	expected, ok := valid[from]
	if !ok {
		return fmt.Errorf("%w: %s is not a valid TDD phase", ErrInvalidPhaseTransition, from)
	}
	if expected != to {
		return fmt.Errorf("%w: cannot transition from %s to %s (expected %s)", ErrInvalidPhaseTransition, from, to, expected)
	}
	return nil
}

// MethodologyConfig holds the configuration for a development methodology.
type MethodologyConfig struct {
	Mode models.DevelopmentMode `json:"mode"`
	DDD  *DDDCycle              `json:"ddd,omitempty"`
	TDD  *TDDCycle              `json:"tdd,omitempty"`
}

// NewMethodology creates a MethodologyConfig for the given development mode.
// Returns an error if the mode is not valid.
func NewMethodology(mode models.DevelopmentMode) (*MethodologyConfig, error) {
	if !mode.IsValid() {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPhaseTransition, mode)
	}

	cfg := &MethodologyConfig{Mode: mode}

	switch mode {
	case models.ModeDDD:
		cfg.DDD = &DDDCycle{}
	case models.ModeTDD:
		cfg.TDD = &TDDCycle{}
	}

	return cfg, nil
}

// ActivePhases returns the phases available for the configured methodology.
func (mc *MethodologyConfig) ActivePhases() []Phase {
	switch mc.Mode {
	case models.ModeDDD:
		if mc.DDD != nil {
			return mc.DDD.Phases()
		}
	case models.ModeTDD:
		if mc.TDD != nil {
			return mc.TDD.Phases()
		}
	}
	return nil
}
