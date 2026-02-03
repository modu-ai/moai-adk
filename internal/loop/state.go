// Package loop implements the Ralph Feedback Loop Engine state machine,
// persistence, and lifecycle management for MoAI-ADK.
package loop

import (
	"context"
	"errors"
	"time"
)

// Sentinel errors for loop operations.
var (
	ErrLoopAlreadyRunning   = errors.New("loop: already running")
	ErrLoopNotRunning       = errors.New("loop: not running")
	ErrLoopNotPaused        = errors.New("loop: not paused")
	ErrInvalidTransition    = errors.New("loop: invalid phase transition")
	ErrCorruptedState       = errors.New("loop: corrupted state file")
	ErrMaxIterationsReached = errors.New("loop: max iterations reached")
)

// LoopPhase represents a phase in the feedback loop cycle.
type LoopPhase string

const (
	PhaseAnalyze   LoopPhase = "analyze"
	PhaseImplement LoopPhase = "implement"
	PhaseTest      LoopPhase = "test"
	PhaseReview    LoopPhase = "review"
)

// Action constants for Decision.Action field.
const (
	ActionContinue      = "continue"
	ActionConverge      = "converge"
	ActionRequestReview = "request_review"
	ActionAbort         = "abort"
)

// DefaultCoverageTarget is the minimum coverage percentage for quality gate.
const DefaultCoverageTarget = 85.0

// validTransitions maps each phase to its only valid next phase.
var validTransitions = map[LoopPhase]LoopPhase{
	PhaseAnalyze:   PhaseImplement,
	PhaseImplement: PhaseTest,
	PhaseTest:      PhaseReview,
	PhaseReview:    PhaseAnalyze,
}

// ValidTransition checks if transitioning from current to next is a valid move.
func ValidTransition(current, next LoopPhase) bool {
	expected, ok := validTransitions[current]
	if !ok {
		return false
	}
	return expected == next
}

// NextPhase returns the next phase in the loop cycle after the given phase.
// Returns PhaseAnalyze for unknown phases.
func NextPhase(current LoopPhase) LoopPhase {
	next, ok := validTransitions[current]
	if !ok {
		return PhaseAnalyze
	}
	return next
}

// IsValidPhase checks if the given phase is a recognized loop phase.
func IsValidPhase(phase LoopPhase) bool {
	_, ok := validTransitions[phase]
	return ok
}

// LoopState represents the persistent state of a feedback loop.
type LoopState struct {
	SpecID    string     `json:"spec_id"`
	Phase     LoopPhase  `json:"phase"`
	Iteration int        `json:"iteration"`
	MaxIter   int        `json:"max_iterations"`
	Feedback  []Feedback `json:"feedback"`
	StartedAt time.Time  `json:"started_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// ToStatus creates a read-only status snapshot from the current state.
func (s *LoopState) ToStatus(running, converged bool) *LoopStatus {
	if s == nil {
		return &LoopStatus{}
	}
	return &LoopStatus{
		SpecID:    s.SpecID,
		Phase:     s.Phase,
		Iteration: s.Iteration,
		MaxIter:   s.MaxIter,
		Converged: converged,
		Running:   running,
	}
}

// LoopStatus is a read-only snapshot of the current loop state.
type LoopStatus struct {
	SpecID    string    `json:"spec_id"`
	Phase     LoopPhase `json:"phase"`
	Iteration int       `json:"iteration"`
	MaxIter   int       `json:"max_iterations"`
	Converged bool      `json:"converged"`
	Running   bool      `json:"running"`
}

// Feedback captures the results of a loop phase execution.
type Feedback struct {
	Phase        LoopPhase     `json:"phase"`
	Iteration    int           `json:"iteration"`
	TestsPassed  int           `json:"tests_passed"`
	TestsFailed  int           `json:"tests_failed"`
	LintErrors   int           `json:"lint_errors"`
	BuildSuccess bool          `json:"build_success"`
	Coverage     float64       `json:"coverage"`
	Duration     time.Duration `json:"duration"`
	Notes        string        `json:"notes"`
}

// Decision represents the decision engine's output after evaluating feedback.
type Decision struct {
	Action    string    `json:"action"`
	NextPhase LoopPhase `json:"next_phase"`
	Converged bool      `json:"converged"`
	Reason    string    `json:"reason"`
}

// Controller orchestrates the Ralph feedback loop lifecycle.
type Controller interface {
	Start(ctx context.Context, specID string) error
	Pause() error
	Resume(ctx context.Context) error
	Cancel() error
	Status() *LoopStatus
	RecordFeedback(feedback Feedback) error
}

// Storage persists loop state for session resumption.
type Storage interface {
	SaveState(state *LoopState) error
	LoadState(specID string) (*LoopState, error)
	DeleteState(specID string) error
}

// FeedbackGenerator collects feedback from build, test, and lint results.
type FeedbackGenerator interface {
	Collect(ctx context.Context) (*Feedback, error)
}

// DecisionEngine determines the next loop action based on state and feedback.
type DecisionEngine interface {
	Decide(ctx context.Context, state *LoopState, feedback *Feedback) (*Decision, error)
}
