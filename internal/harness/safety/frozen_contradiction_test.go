// Package safety — L3 Frozen-rule contradiction detector tests (SPEC-HARNESS-EVOLVE-003 M3).
// REQ-HEV3-015/016/017/033: replace the L3 no-op with a real Frozen-rules consult.
package safety

import (
	"strings"
	"testing"
	"time"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// TestDetectFrozenRuleContradictions_FrozenPath verifies REQ-HEV3-015/016/017:
// a proposal targeting a registered Frozen-rule path returns a ContradictionReport
// with HasContradiction()==true and cites the rule identifier in the Description.
func TestDetectFrozenRuleContradictions_FrozenPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		targetPath  string
		wantRuleName string // non-empty — the rule Name that should be cited
	}{
		{"moai rules", ".claude/rules/moai/core/moai-constitution.md", "frozen-moai-rules"},
		{"moai agents", ".claude/agents/moai/manager-develop.md", "frozen-moai-agents"},
		{"settings.json (A1)", ".claude/settings.json", "frozen-permission-settings"},
		{"guard source (self-protection)", "internal/harness/safety/frozen_guard.go", "frozen-guard-safety"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			proposal := harness.Proposal{TargetPath: tc.targetPath}
			report := DetectFrozenRuleContradictions(proposal)

			if !report.HasContradiction() {
				t.Fatalf("HasContradiction() = false for %q, want true", tc.targetPath)
			}
			if len(report.Items) == 0 {
				t.Fatal("report.Items empty, want at least 1 contradiction item")
			}
			// REQ-HEV3-017: the rejection Reason cites the rule identifier
			desc := report.Items[0].Description
			if !strings.Contains(desc, tc.wantRuleName) {
				t.Errorf("Description %q does not cite rule Name %q (REQ-HEV3-017)", desc, tc.wantRuleName)
			}
		})
	}
}

// TestDetectFrozenRuleContradictions_NonFrozenPath verifies a proposal targeting
// a non-Frozen path returns an empty report.
func TestDetectFrozenRuleContradictions_NonFrozenPath(t *testing.T) {
	t.Parallel()

	proposal := harness.Proposal{TargetPath: ".claude/agents/harness/my-agent.md"}
	report := DetectFrozenRuleContradictions(proposal)

	if report.HasContradiction() {
		t.Errorf("HasContradiction() = true for non-Frozen path, want false")
	}
}

// TestPipeline_L3FrozenRule_RejectedBy3 verifies AC-HEV3-016 / REQ-HEV3-033:
// a Curator proposal that contradicts a registered Frozen rule is rejected with
// RejectedBy == 3 (L3 reachability).
//
// The test uses a Pipeline with L1 configured to pass-through (return false) so
// the proposal reaches L3 — simulating the case where L1 does not cover the path
// but L3's typed registry does. This proves the REAL l3ContradictionCheck fires.
func TestPipeline_L3FrozenRule_RejectedBy3(t *testing.T) {
	t.Parallel()

	pipe := &Pipeline{
		// L1 pass-through — let the proposal reach L3 (test injection)
		l1FrozenCheck: func(_ harness.Proposal) bool { return false },
		// L2 pass-through
		l2CanaryCheck: func(_ harness.Proposal, _ []harness.Session) (harness.CanaryResult, error) {
			return harness.CanaryResult{}, nil
		},
		// L3: the REAL Frozen-rules consult (the M3 activation)
		l3ContradictionCheck: func(proposal harness.Proposal) harness.ContradictionReport {
			return DetectFrozenRuleContradictions(proposal)
		},
		// L4 pass-through
		l4RateLimitCheck: func() (bool, time.Duration, error) { return true, 0, nil },
		// L5 not reached (short-circuit at L3)
		l5OversightProposal: BuildOversightProposal,
		autoApply:           true,
	}

	proposal := harness.Proposal{
		ID:         "test-l3-001",
		TargetPath: ".claude/rules/moai/core/moai-constitution.md",
		PatternKey: "feature+plan+autopilot+success",
	}
	decision, err := pipe.Evaluate(proposal, nil)
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}

	if decision.Kind != harness.DecisionRejected {
		t.Errorf("Kind = %q, want %q", decision.Kind, harness.DecisionRejected)
	}
	if decision.RejectedBy != 3 {
		t.Errorf("RejectedBy = %d, want 3 (L3 Frozen-rule contradiction — AC-HEV3-016)", decision.RejectedBy)
	}
	// REQ-HEV3-017: the Reason cites the rule identifier
	if !strings.Contains(decision.Reason, "frozen-moai-rules") {
		t.Errorf("Reason %q does not cite the rule identifier (REQ-HEV3-017)", decision.Reason)
	}
}
