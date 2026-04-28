package session

// Phase represents a moai workflow phase.
type Phase string

const (
	PhasePlan   Phase = "plan"
	PhaseRun    Phase = "run"
	PhaseSync   Phase = "sync"
	PhaseDesign Phase = "design"
	PhaseReview Phase = "review"
	PhaseFix    Phase = "fix"
	PhaseLoop   Phase = "loop"
	PhaseDB     Phase = "db"
	PhaseMX     Phase = "mx"
)

// String returns the string representation of the phase.
func (p Phase) String() string {
	return string(p)
}

// Valid checks if the phase is one of the known values.
func (p Phase) Valid() bool {
	switch p {
	case PhasePlan, PhaseRun, PhaseSync, PhaseDesign, PhaseReview,
		PhaseFix, PhaseLoop, PhaseDB, PhaseMX:
		return true
	default:
		return false
	}
}
