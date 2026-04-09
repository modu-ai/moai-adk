// Package experiment implements the experiment loop state machine for the Self-Research System.
// It manages the complete experiment cycle: baseline measurement, mutation application,
// evaluation, and scoring.
package experiment

import (
	"time"

	"github.com/modu-ai/moai-adk/internal/research/eval"
)

// ExperimentState represents the current state of the experiment loop.
type ExperimentState string

const (
	// StateIdle is the state when the experiment loop is initialized but not yet started.
	StateIdle ExperimentState = "idle"
	// StateBaseline is the state after the baseline has been established.
	StateBaseline ExperimentState = "baseline"
	// StateMutating is the state while applying a mutation.
	StateMutating ExperimentState = "mutating"
	// StateEvaluating is the state while evaluating the mutation result.
	StateEvaluating ExperimentState = "evaluating"
	// StateScoring is the state while computing scores from evaluation results.
	StateScoring ExperimentState = "scoring"
	// StateComplete is the state when the experiment loop has finished.
	StateComplete ExperimentState = "complete"
)

// Decision represents the keep/discard decision for an experiment result.
type Decision string

const (
	// DecisionKeep accepts the experiment change.
	DecisionKeep Decision = "keep"
	// DecisionDiscard rejects the experiment change.
	DecisionDiscard Decision = "discard"
	// DecisionPending is the state before a decision has been made.
	DecisionPending Decision = "pending"
)

// ChangeRecord records the change applied in an experiment.
type ChangeRecord struct {
	Type    string `json:"type"`    // addition | modification | deletion
	Section string `json:"section"` // Name of the changed section
	Diff    string `json:"diff"`    // Diff text of the change
}

// Experiment holds the complete information for a single experiment.
type Experiment struct {
	ID         string           `json:"id"`
	Target     string           `json:"target"`
	Hypothesis string           `json:"hypothesis"`
	Change     ChangeRecord     `json:"change"`
	Result     *eval.EvalResult `json:"result"`
	Decision   Decision         `json:"decision"`
	Timestamp  time.Time        `json:"timestamp"`
}

// ChangelogEntry is a change history entry for an experiment result.
type ChangelogEntry struct {
	ExperimentID string   `json:"experiment_id"`
	Score        float64  `json:"score"`
	Change       string   `json:"change"`
	Reasoning    string   `json:"reasoning"`
	Decision     Decision `json:"decision"`
}
