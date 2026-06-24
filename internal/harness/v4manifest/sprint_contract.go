// Package v4manifest — Sprint Contract + conditional evaluator (REQ-HV4-008).
//
// The Sprint Contract is the Anthropic-GAN-inspired Generator-Evaluator
// separation: the agent that generates output is distinct from the agent that
// evaluates it, and the evaluator earns its cost only when the task exceeds
// the model's solo reliable range (C-HV4-001 simplest-solution-first).
//
// This file provides the Go-side EvaluatorDecision: a pure function that
// decides whether the evaluator should be invoked or skipped for a given task
// profile, and records the rationale either way (AC-HV4-008b). The Runner
// template (runner_template.go) mirrors this logic in JS as
// applySprintContract().
package v4manifest

// TaskProfile characterizes a task against the model's solo reliable range.
// The Runner/ACTIVATE phase populates this before deciding whether to invoke
// the evaluator (REQ-HV4-008 / AC-HV4-008b).
type TaskProfile struct {
	// WithinSoloRange is true when the task is simple and well-bounded enough
	// that the model handles it reliably solo (no adversarial verification
	// needed). When true, the evaluator is SKIPPED per C-HV4-001.
	WithinSoloRange bool

	// ComplexitySignals is an optional human-readable summary of the signals
	// that informed the WithinSoloRange decision (e.g. "single-skill
	// generation, no adversarial-verification need"). Recorded for audit.
	ComplexitySignals string
}

// EvaluatorDecision is the result of deciding whether to invoke the evaluator.
// It is returned by DecideEvaluator and recorded in the run log so a third
// party can audit whether the evaluator ran and why (NFR-HV4-002).
type EvaluatorDecision struct {
	// Invoked is true when the evaluator runs (task exceeds solo range),
	// false when the evaluator is skipped (task within solo range).
	Invoked bool

	// Rationale records the reason for the decision. Always non-empty.
	Rationale string

	// Dimensions is the Sprint Contract dimensions the evaluator would grade
	// (echoed from the manifest for auditability when Invoked).
	Dimensions []string
}

// DecideEvaluator applies the conditional-evaluator rule (REQ-HV4-008 /
// AC-HV4-008b). When the task is within the model's solo reliable range, the
// evaluator is SKIPPED and the skip is recorded with rationale. Otherwise the
// evaluator is invoked and the Sprint Contract dimensions are carried.
//
// The manifest's SprintContract is passed so the decision can echo the graded
// dimensions when the evaluator runs (NFR-HV4-002 observability).
func DecideEvaluator(profile TaskProfile, contract SprintContract) EvaluatorDecision {
	if profile.WithinSoloRange {
		rationale := "task within solo range, evaluator skipped per simplest-solution-first"
		if profile.ComplexitySignals != "" {
			rationale += " (" + profile.ComplexitySignals + ")"
		}
		return EvaluatorDecision{
			Invoked:  false,
			Rationale: rationale,
		}
	}
	return EvaluatorDecision{
		Invoked:    true,
		Rationale:  "task exceeds solo range, evaluator invoked as skeptical adversarial reviewer",
		Dimensions: contract.Dimensions,
	}
}
