// Package safety — pipeline integration tests.
// REQ-HL-005~008: enforce L1→L2→L3→L4→L5 ordering and verify short-circuit behavior.
package safety

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// layerCallOrder is a test helper that tracks the call order of layers.
type layerCallOrder struct {
	order []int
}

func (l *layerCallOrder) record(layer int) {
	l.order = append(l.order, layer)
}

// TestPipeline_OrderingL1ToL5 verifies the L1→L2→L3→L4→L5 ordering.
// [HARD] The order in pipeline.go must be verified with a mock counter.
func TestPipeline_OrderingL1ToL5(t *testing.T) {
	t.Parallel()

	tracker := &layerCallOrder{}

	// Mock pipeline that records the order whenever each layer is called
	pipe := &Pipeline{
		l1FrozenCheck: func(_ harness.Proposal) bool {
			tracker.record(1)
			return false // not FROZEN → pass
		},
		l2CanaryCheck: func(_ harness.Proposal, _ []harness.Session) (harness.CanaryResult, error) {
			tracker.record(2)
			return harness.CanaryResult{Rejected: false}, nil
		},
		l3ContradictionCheck: func(_ harness.Proposal) harness.ContradictionReport {
			tracker.record(3)
			return harness.ContradictionReport{}
		},
		l4RateLimitCheck: func() (bool, time.Duration, error) {
			tracker.record(4)
			return true, 0, nil
		},
		l5OversightProposal: func(p harness.Proposal, reason string) harness.OversightProposal {
			tracker.record(5)
			return BuildOversightProposal(p, reason)
		},
		autoApply: false, // L5 always runs (with auto_apply=false, L5 is pending)
	}

	proposal := harness.Proposal{
		ID:         "order-test",
		TargetPath: ".moai/harness/test.yaml",
		CreatedAt:  time.Now(),
	}

	_, err := pipe.Evaluate(proposal, nil)
	if err != nil {
		t.Fatalf("Evaluate 실패: %v", err)
	}

	// Verify ordering: 1 → 2 → 3 → 4 → 5
	expectedOrder := []int{1, 2, 3, 4, 5}
	if len(tracker.order) != len(expectedOrder) {
		t.Errorf("call order length = %v, want %v", tracker.order, expectedOrder)
		return
	}
	for i, got := range tracker.order {
		if got != expectedOrder[i] {
			t.Errorf("tracker.order[%d] = %d, want %d (full: %v)", i, got, expectedOrder[i], tracker.order)
		}
	}
}

// TestPipeline_ShortCircuitOnL1 verifies that L2~L5 are not called when L1 FROZEN is detected.
func TestPipeline_ShortCircuitOnL1(t *testing.T) {
	t.Parallel()

	tracker := &layerCallOrder{}

	pipe := &Pipeline{
		l1FrozenCheck: func(_ harness.Proposal) bool {
			tracker.record(1)
			return true // FROZEN → short-circuit
		},
		l2CanaryCheck: func(_ harness.Proposal, _ []harness.Session) (harness.CanaryResult, error) {
			tracker.record(2)
			return harness.CanaryResult{}, nil
		},
		l3ContradictionCheck: func(_ harness.Proposal) harness.ContradictionReport {
			tracker.record(3)
			return harness.ContradictionReport{}
		},
		l4RateLimitCheck: func() (bool, time.Duration, error) {
			tracker.record(4)
			return true, 0, nil
		},
		l5OversightProposal: func(p harness.Proposal, reason string) harness.OversightProposal {
			tracker.record(5)
			return BuildOversightProposal(p, reason)
		},
	}

	proposal := harness.Proposal{
		ID:         "frozen-test",
		TargetPath: ".claude/agents/moai/evil.md",
		CreatedAt:  time.Now(),
	}

	decision, err := pipe.Evaluate(proposal, nil)
	if err != nil {
		t.Fatalf("Evaluate 실패: %v", err)
	}

	// Must be rejected at L1
	if decision.Kind != harness.DecisionRejected {
		t.Errorf("Kind = %q, want %q", decision.Kind, harness.DecisionRejected)
	}
	if decision.RejectedBy != 1 {
		t.Errorf("RejectedBy = %d, want 1", decision.RejectedBy)
	}

	// L2~L5 must not be called
	for _, layer := range tracker.order {
		if layer >= 2 {
			t.Errorf("L%d was called after L1 rejection (short-circuit failed)", layer)
		}
	}
}

// TestPipeline_ShortCircuitOnL2 verifies that L3~L5 are not called when L2 canary fails.
func TestPipeline_ShortCircuitOnL2(t *testing.T) {
	t.Parallel()

	tracker := &layerCallOrder{}

	pipe := &Pipeline{
		l1FrozenCheck: func(_ harness.Proposal) bool {
			tracker.record(1)
			return false
		},
		l2CanaryCheck: func(_ harness.Proposal, _ []harness.Session) (harness.CanaryResult, error) {
			tracker.record(2)
			return harness.CanaryResult{Rejected: true, Reason: "mock canary rejection"}, nil
		},
		l3ContradictionCheck: func(_ harness.Proposal) harness.ContradictionReport {
			tracker.record(3)
			return harness.ContradictionReport{}
		},
		l4RateLimitCheck: func() (bool, time.Duration, error) {
			tracker.record(4)
			return true, 0, nil
		},
		l5OversightProposal: func(p harness.Proposal, reason string) harness.OversightProposal {
			tracker.record(5)
			return BuildOversightProposal(p, reason)
		},
	}

	proposal := harness.Proposal{ID: "canary-test", CreatedAt: time.Now()}

	decision, err := pipe.Evaluate(proposal, nil)
	if err != nil {
		t.Fatalf("Evaluate 실패: %v", err)
	}

	if decision.Kind != harness.DecisionRejected {
		t.Errorf("Kind = %q, want %q", decision.Kind, harness.DecisionRejected)
	}
	if decision.RejectedBy != 2 {
		t.Errorf("RejectedBy = %d, want 2", decision.RejectedBy)
	}

	for _, layer := range tracker.order {
		if layer >= 3 {
			t.Errorf("L%d was called after L2 rejection (short-circuit failed)", layer)
		}
	}
}

// TestPipeline_ShortCircuitOnL3 verifies that L4~L5 are not called when L3 detects a contradiction.
func TestPipeline_ShortCircuitOnL3(t *testing.T) {
	t.Parallel()

	tracker := &layerCallOrder{}

	pipe := &Pipeline{
		l1FrozenCheck: func(_ harness.Proposal) bool {
			tracker.record(1)
			return false
		},
		l2CanaryCheck: func(_ harness.Proposal, _ []harness.Session) (harness.CanaryResult, error) {
			tracker.record(2)
			return harness.CanaryResult{Rejected: false}, nil
		},
		l3ContradictionCheck: func(_ harness.Proposal) harness.ContradictionReport {
			tracker.record(3)
			return harness.ContradictionReport{
				Items: []harness.ContradictionItem{
					{Type: harness.ContradictionOverlappingTriggers, Description: "conflict"},
				},
			}
		},
		l4RateLimitCheck: func() (bool, time.Duration, error) {
			tracker.record(4)
			return true, 0, nil
		},
		l5OversightProposal: func(p harness.Proposal, reason string) harness.OversightProposal {
			tracker.record(5)
			return BuildOversightProposal(p, reason)
		},
	}

	proposal := harness.Proposal{ID: "contradiction-test", CreatedAt: time.Now()}

	decision, err := pipe.Evaluate(proposal, nil)
	if err != nil {
		t.Fatalf("Evaluate 실패: %v", err)
	}

	if decision.Kind != harness.DecisionRejected {
		t.Errorf("Kind = %q, want %q", decision.Kind, harness.DecisionRejected)
	}
	if decision.RejectedBy != 3 {
		t.Errorf("RejectedBy = %d, want 3", decision.RejectedBy)
	}

	for _, layer := range tracker.order {
		if layer >= 4 {
			t.Errorf("L%d was called after L3 rejection (short-circuit failed)", layer)
		}
	}
}

// TestPipeline_ShortCircuitOnL4 verifies that L5 is not called when L4 rate-limit is exceeded.
func TestPipeline_ShortCircuitOnL4(t *testing.T) {
	t.Parallel()

	tracker := &layerCallOrder{}

	pipe := &Pipeline{
		l1FrozenCheck: func(_ harness.Proposal) bool {
			tracker.record(1)
			return false
		},
		l2CanaryCheck: func(_ harness.Proposal, _ []harness.Session) (harness.CanaryResult, error) {
			tracker.record(2)
			return harness.CanaryResult{Rejected: false}, nil
		},
		l3ContradictionCheck: func(_ harness.Proposal) harness.ContradictionReport {
			tracker.record(3)
			return harness.ContradictionReport{}
		},
		l4RateLimitCheck: func() (bool, time.Duration, error) {
			tracker.record(4)
			return false, 24 * time.Hour, nil // blocked
		},
		l5OversightProposal: func(p harness.Proposal, reason string) harness.OversightProposal {
			tracker.record(5)
			return BuildOversightProposal(p, reason)
		},
	}

	proposal := harness.Proposal{ID: "rate-test", CreatedAt: time.Now()}

	decision, err := pipe.Evaluate(proposal, nil)
	if err != nil {
		t.Fatalf("Evaluate 실패: %v", err)
	}

	if decision.Kind != harness.DecisionRejected {
		t.Errorf("Kind = %q, want %q", decision.Kind, harness.DecisionRejected)
	}
	if decision.RejectedBy != 4 {
		t.Errorf("RejectedBy = %d, want 4", decision.RejectedBy)
	}

	for _, layer := range tracker.order {
		if layer >= 5 {
			t.Errorf("L%d was called after L4 rejection (short-circuit failed)", layer)
		}
	}
}

// TestPipeline_L5PendingWhenAutoApplyFalse verifies that with auto_apply=false, L5 returns pending_approval.
func TestPipeline_L5PendingWhenAutoApplyFalse(t *testing.T) {
	t.Parallel()

	pipe := &Pipeline{
		l1FrozenCheck: func(_ harness.Proposal) bool { return false },
		l2CanaryCheck: func(_ harness.Proposal, _ []harness.Session) (harness.CanaryResult, error) {
			return harness.CanaryResult{Rejected: false}, nil
		},
		l3ContradictionCheck: func(_ harness.Proposal) harness.ContradictionReport {
			return harness.ContradictionReport{}
		},
		l4RateLimitCheck: func() (bool, time.Duration, error) { return true, 0, nil },
		l5OversightProposal: func(p harness.Proposal, reason string) harness.OversightProposal {
			return BuildOversightProposal(p, reason)
		},
		autoApply: false, // L5 always pending
	}

	proposal := harness.Proposal{ID: "pending-test", CreatedAt: time.Now()}

	decision, err := pipe.Evaluate(proposal, nil)
	if err != nil {
		t.Fatalf("Evaluate 실패: %v", err)
	}

	if decision.Kind != harness.DecisionPendingApproval {
		t.Errorf("Kind = %q, want %q", decision.Kind, harness.DecisionPendingApproval)
	}

	if decision.OversightProposal == nil {
		t.Error("OversightProposal is nil; L5 pending must include a payload")
	}
}

// TestPipeline_ApprovedWhenAutoApplyTrue verifies that with auto_apply=true approved is returned.
func TestPipeline_ApprovedWhenAutoApplyTrue(t *testing.T) {
	t.Parallel()

	pipe := &Pipeline{
		l1FrozenCheck: func(_ harness.Proposal) bool { return false },
		l2CanaryCheck: func(_ harness.Proposal, _ []harness.Session) (harness.CanaryResult, error) {
			return harness.CanaryResult{Rejected: false}, nil
		},
		l3ContradictionCheck: func(_ harness.Proposal) harness.ContradictionReport {
			return harness.ContradictionReport{}
		},
		l4RateLimitCheck: func() (bool, time.Duration, error) { return true, 0, nil },
		l5OversightProposal: func(p harness.Proposal, reason string) harness.OversightProposal {
			return BuildOversightProposal(p, reason)
		},
		autoApply: true, // L5 auto-approval
	}

	proposal := harness.Proposal{ID: "auto-approve-test", CreatedAt: time.Now()}

	decision, err := pipe.Evaluate(proposal, nil)
	if err != nil {
		t.Fatalf("Evaluate 실패: %v", err)
	}

	if decision.Kind != harness.DecisionApproved {
		t.Errorf("Kind = %q, want %q", decision.Kind, harness.DecisionApproved)
	}
}

// TestPipeline_L2ErrorPropagated verifies that an L2 error is propagated by Evaluate.
func TestPipeline_L2ErrorPropagated(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("canary internal error")

	pipe := &Pipeline{
		l1FrozenCheck: func(_ harness.Proposal) bool { return false },
		l2CanaryCheck: func(_ harness.Proposal, _ []harness.Session) (harness.CanaryResult, error) {
			return harness.CanaryResult{}, expectedErr
		},
		l3ContradictionCheck: func(_ harness.Proposal) harness.ContradictionReport {
			return harness.ContradictionReport{}
		},
		l4RateLimitCheck: func() (bool, time.Duration, error) { return true, 0, nil },
		l5OversightProposal: func(p harness.Proposal, reason string) harness.OversightProposal {
			return BuildOversightProposal(p, reason)
		},
	}

	proposal := harness.Proposal{ID: "error-test", CreatedAt: time.Now()}

	_, err := pipe.Evaluate(proposal, nil)
	if err == nil {
		t.Fatal("L2 error was not propagated")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("err = %v, want %v", err, expectedErr)
	}
}

// TestNewPipeline_DefaultConfig verifies Pipeline creation with default config.
func TestNewPipeline_DefaultConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	violationLogPath := filepath.Join(dir, "violations.jsonl")
	rateLimitPath := filepath.Join(dir, "rate-limit-state.json")

	pipe := NewPipeline(PipelineConfig{
		ViolationLogPath: violationLogPath,
		RateLimitPath:    rateLimitPath,
		AutoApply:        false,
	})

	if pipe == nil {
		t.Fatal("NewPipeline returned nil")
	}

	// Evaluate an actual proposal with the default pipeline (non-FROZEN path)
	proposal := harness.Proposal{
		ID:         "default-test",
		TargetPath: ".moai/harness/test.yaml",
		CreatedAt:  time.Now(),
	}

	decision, err := pipe.Evaluate(proposal, nil)
	if err != nil {
		t.Fatalf("default Pipeline Evaluate failed: %v", err)
	}

	// With auto_apply=false, result must be pending_approval or rejected
	if decision.Kind == harness.DecisionApproved {
		t.Error("auto_apply=false but approved was returned")
	}
}
