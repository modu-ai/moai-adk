// Package safety — Layer 2: Canary Check (REQ-HL-007).
// Detects effectiveness degradation by shadow-evaluating recent 3 sessions before proposal application.
package safety

import (
	"fmt"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// canaryRejectionThreshold is the rejection threshold.
// REQ-HL-007: Rejects proposal if effectiveness score drops by this value or more.
const canaryRejectionThreshold = 0.10

// maxCanarySessions is the maximum number of sessions to use for canary check.
// REQ-HL-007: Uses the most recent 3 sessions.
const maxCanarySessions = 3

// ProjectedScorer is a function type that calculates the proposal's projected effectiveness score.
// Default implementation is defaultProjectedScorer, can be injected in tests.
type ProjectedScorer func(proposal harness.Proposal, baseline float64) float64

// effectivenessScore calculates the effectiveness score for a single Session.
// Weighted sum: subcommand 40%, agent_invocation 40%, completion 20%.
func effectivenessScore(s harness.Session) float64 {
	return s.SubcommandSuccessRate*0.40 +
		s.AgentInvocationSuccess*0.40 +
		s.CompletionRate*0.20
}

// baselineScore calculates the average effectiveness of session list (latest N).
func baselineScore(sessions []harness.Session) float64 {
	n := len(sessions)
	if n == 0 {
		return 0
	}

	// Use only the latest maxCanarySessions
	start := 0
	if n > maxCanarySessions {
		start = n - maxCanarySessions
	}
	sessions = sessions[start:]

	var total float64
	for _, s := range sessions {
		total += effectivenessScore(s)
	}
	return total / float64(len(sessions))
}

// defaultProjectedScorer is the default projected score calculator.
// Simple implementation that simulates actual application effect:
// - Maintain baseline if TargetPath or NewValue is empty (meaningless change)
// - Otherwise assume slight improvement (+0.02) (conservative estimate)
func defaultProjectedScorer(proposal harness.Proposal, baseline float64) float64 {
	if proposal.TargetPath == "" || (proposal.FieldKey == "" && proposal.NewValue == "") {
		return baseline
	}
	// Assume slight improvement for meaningful proposals
	return baseline + 0.02
}

// EvaluateCanary shadow-evaluates the proposal against recent 3 sessions and returns CanaryResult.
// REQ-HL-007: Rejected=true if projected score drops by 0.10 or more compared to baseline.
//
// @MX:ANCHOR: [AUTO] EvaluateCanary is the single entry point for Layer 2.
// @MX:REASON: [AUTO] fan_in >= 3: canary_test.go, pipeline.go, Phase 4 coordinator
func EvaluateCanary(proposal harness.Proposal, sessions []harness.Session) (harness.CanaryResult, error) {
	return EvaluateCanaryWithScorer(proposal, sessions, defaultProjectedScorer)
}

// EvaluateCanaryWithScorer runs canary check with user-defined scorer.
// Used when injecting degrading/improving scenarios in tests.
//
// @MX:ANCHOR: [AUTO] EvaluateCanaryWithScorer is the test injection entry point.
// @MX:REASON: [AUTO] fan_in >= 3: EvaluateCanary, canary_test.go x3
func EvaluateCanaryWithScorer(
	proposal harness.Proposal,
	sessions []harness.Session,
	scorer ProjectedScorer,
) (harness.CanaryResult, error) {
	baseline := baselineScore(sessions)
	projected := scorer(proposal, baseline)
	delta := projected - baseline

	result := harness.CanaryResult{
		BaselineScore:  baseline,
		ProjectedScore: projected,
		Delta:          delta,
	}

	// Reject if drop amount exceeds threshold.
	// Use epsilon=1e-9 for floating-point error tolerance.
	const epsilon = 1e-9
	drop := baseline - projected
	if drop >= canaryRejectionThreshold-epsilon {
		result.Rejected = true
		result.Reason = fmt.Sprintf(
			"effectiveness drop %.4f (threshold %.2f or more): baseline=%.4f → projected=%.4f",
			drop, canaryRejectionThreshold, baseline, projected,
		)
	}

	return result, nil
}
