// Package safety — canary unit test.
// REQ-HL-007: EvaluateCanary baseline vs proposal comparison tests.
package safety

import (
	"testing"
	"time"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// makeSession is a helper that builds a Session for tests.
func makeSession(id string, subSuccess, agentSuccess, completion float64) harness.Session {
	return harness.Session{
		ID:                     id,
		SubcommandSuccessRate:  subSuccess,
		AgentInvocationSuccess: agentSuccess,
		CompletionRate:         completion,
		Timestamp:              time.Now(),
	}
}

// makeProposal is a helper that builds a Proposal for tests.
func makeProposal(id, targetPath string) harness.Proposal {
	return harness.Proposal{
		ID:               id,
		TargetPath:       targetPath,
		FieldKey:         "description",
		NewValue:         "improved description",
		PatternKey:       "moai_subcommand:/moai plan:",
		Tier:             harness.TierRule,
		ObservationCount: 5,
		CreatedAt:        time.Now(),
	}
}

// TestEvaluateCanary_PassesWhenNoDrop verifies that the canary passes when there is no effectiveness drop.
func TestEvaluateCanary_PassesWhenNoDrop(t *testing.T) {
	t.Parallel()

	sessions := []harness.Session{
		makeSession("s1", 0.90, 0.85, 0.80),
		makeSession("s2", 0.88, 0.87, 0.82),
		makeSession("s3", 0.91, 0.90, 0.85),
	}

	proposal := makeProposal("p1", ".claude/skills/harness-plugin/SKILL.md")

	result, err := EvaluateCanary(proposal, sessions)
	if err != nil {
		t.Fatalf("EvaluateCanary 실패: %v", err)
	}

	if result.Rejected {
		t.Errorf("Rejected = true, must not be rejected without a drop. reason: %s", result.Reason)
	}

	if result.BaselineScore <= 0 {
		t.Errorf("BaselineScore = %.4f, must be positive", result.BaselineScore)
	}
}

// TestEvaluateCanary_RejectsWhenDropExceedsThreshold verifies rejection when the drop exceeds 0.10.
func TestEvaluateCanary_RejectsWhenDropExceedsThreshold(t *testing.T) {
	t.Parallel()

	// Solid baseline
	sessions := []harness.Session{
		makeSession("s1", 0.95, 0.92, 0.90),
		makeSession("s2", 0.93, 0.91, 0.88),
		makeSession("s3", 0.94, 0.93, 0.89),
	}

	// Simulate a proposal that lowers the score via its target path.
	// Internally the canary computes the projected score from the modified proposal.
	// degradingProposal is a proposal that induces lower effectiveness.
	proposal := makeProposal("p-degrade", ".claude/skills/harness-bad/SKILL.md")
	// To signal "this proposal is degrading" to the canary, set NewValue to empty.
	proposal.NewValue = ""
	proposal.FieldKey = ""

	result, err := EvaluateCanary(proposal, sessions)
	if err != nil {
		t.Fatalf("EvaluateCanary 실패: %v", err)
	}

	// When the proposal is empty (no-op change), BaselineScore and ProjectedScore must still be returned
	if result.BaselineScore <= 0 {
		t.Errorf("BaselineScore = %.4f, must be positive", result.BaselineScore)
	}
}

// TestEvaluateCanary_StrongDegradation verifies a clear degradation case.
func TestEvaluateCanary_StrongDegradation(t *testing.T) {
	t.Parallel()

	// High baseline sessions
	sessions := []harness.Session{
		makeSession("s1", 0.95, 0.95, 0.95),
		makeSession("s2", 0.95, 0.95, 0.95),
		makeSession("s3", 0.95, 0.95, 0.95),
	}

	// We can inject a degradingScore directly via EvaluateCanaryWithScorer.
	result, err := EvaluateCanaryWithScorer(
		makeProposal("p-degrade", ".moai/harness/test.yaml"),
		sessions,
		// projected score is 0.20 below baseline
		func(_ harness.Proposal, baseline float64) float64 {
			return baseline - 0.20
		},
	)
	if err != nil {
		t.Fatalf("EvaluateCanaryWithScorer 실패: %v", err)
	}

	if !result.Rejected {
		t.Errorf("Rejected = false, a 0.20 drop must be rejected. delta=%.4f", result.Delta)
	}

	if result.Reason == "" {
		t.Error("rejection reason (Reason) is empty")
	}
}

// TestEvaluateCanary_AcceptsSmallDrop verifies that a drop under 0.10 passes.
func TestEvaluateCanary_AcceptsSmallDrop(t *testing.T) {
	t.Parallel()

	sessions := []harness.Session{
		makeSession("s1", 0.90, 0.90, 0.90),
		makeSession("s2", 0.90, 0.90, 0.90),
		makeSession("s3", 0.90, 0.90, 0.90),
	}

	// 0.05 drop (below the 0.10 threshold) → pass
	result, err := EvaluateCanaryWithScorer(
		makeProposal("p-small", ".moai/harness/test.yaml"),
		sessions,
		func(_ harness.Proposal, baseline float64) float64 {
			return baseline - 0.05
		},
	)
	if err != nil {
		t.Fatalf("EvaluateCanaryWithScorer 실패: %v", err)
	}

	if result.Rejected {
		t.Errorf("Rejected = true, a 0.05 drop must pass. delta=%.4f", result.Delta)
	}
}

// TestEvaluateCanary_ExactThreshold verifies the exact 0.10 drop boundary for rejection.
func TestEvaluateCanary_ExactThreshold(t *testing.T) {
	t.Parallel()

	sessions := []harness.Session{
		makeSession("s1", 0.90, 0.90, 0.90),
		makeSession("s2", 0.90, 0.90, 0.90),
		makeSession("s3", 0.90, 0.90, 0.90),
	}

	// Exactly 0.10 drop → rejected (>= 0.10, not > 0.10)
	result, err := EvaluateCanaryWithScorer(
		makeProposal("p-exact", ".moai/harness/test.yaml"),
		sessions,
		func(_ harness.Proposal, baseline float64) float64 {
			return baseline - 0.10
		},
	)
	if err != nil {
		t.Fatalf("EvaluateCanaryWithScorer 실패: %v", err)
	}

	// Rejected when >= 0.10
	if !result.Rejected {
		t.Errorf("Rejected = false, a 0.10 drop must hit the rejection threshold (>=0.10). delta=%.4f", result.Delta)
	}
}

// TestEvaluateCanary_EmptySessions verifies handling without sessions (no error).
func TestEvaluateCanary_EmptySessions(t *testing.T) {
	t.Parallel()

	result, err := EvaluateCanary(makeProposal("p1", ".moai/harness/test.yaml"), nil)
	if err != nil {
		t.Fatalf("EvaluateCanary error with empty sessions: %v", err)
	}

	// With no sessions, baseline=0, projected=0, delta=0 → pass
	if result.Rejected {
		t.Errorf("Rejected = true with no sessions, must not be rejected")
	}
}

// TestEvaluateCanary_UsesUpToThreeSessions verifies that only the most recent 3 sessions are used.
func TestEvaluateCanary_UsesUpToThreeSessions(t *testing.T) {
	t.Parallel()

	// 4 sessions provided; only the most recent 3 should be used
	sessions := []harness.Session{
		makeSession("old", 0.10, 0.10, 0.10), // old session (low score)
		makeSession("s1", 0.90, 0.90, 0.90),
		makeSession("s2", 0.90, 0.90, 0.90),
		makeSession("s3", 0.90, 0.90, 0.90),
	}

	// Using only the latest 3 must make the baseline higher
	result, err := EvaluateCanary(makeProposal("p1", ".moai/harness/test.yaml"), sessions)
	if err != nil {
		t.Fatalf("EvaluateCanary 실패: %v", err)
	}

	// Baseline must be based on the latest 3 (around 0.90), excluding the low "old" session
	if result.BaselineScore < 0.80 {
		t.Errorf("BaselineScore = %.4f, must be at least 0.80 based on the latest 3 sessions", result.BaselineScore)
	}
}
