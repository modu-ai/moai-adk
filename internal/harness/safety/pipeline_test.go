// Package safety — pipeline integration test.
// REQ-HL-005~008: L1→L2→L3→L4→L5 순서 강제 및 short-circuit 검증.
package safety

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// layerCallOrder는 layer 호출 순서를 추적하는 테스트 헬퍼이다.
type layerCallOrder struct {
	order []int
}

func (l *layerCallOrder) record(layer int) {
	l.order = append(l.order, layer)
}

// TestPipeline_OrderingL1ToL5는 L1→L2→L3→L4→L5 순서를 검증한다.
// [HARD] pipeline.go의 순서는 mock counter로 검증해야 한다.
func TestPipeline_OrderingL1ToL5(t *testing.T) {
	t.Parallel()

	tracker := &layerCallOrder{}

	// 각 layer가 호출되면 순서를 기록하는 mock pipeline
	pipe := &Pipeline{
		l1FrozenCheck: func(_ harness.Proposal) bool {
			tracker.record(1)
			return false // FROZEN 아님 → 통과
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
		autoApply: false, // L5 항상 실행 (auto_apply=false이면 L5 pending)
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

	// 순서 검증: 1 → 2 → 3 → 4 → 5
	expectedOrder := []int{1, 2, 3, 4, 5}
	if len(tracker.order) != len(expectedOrder) {
		t.Errorf("호출 순서 길이 = %v, want %v", tracker.order, expectedOrder)
		return
	}
	for i, got := range tracker.order {
		if got != expectedOrder[i] {
			t.Errorf("tracker.order[%d] = %d, want %d (전체: %v)", i, got, expectedOrder[i], tracker.order)
		}
	}
}

// TestPipeline_ShortCircuitOnL1는 L1 FROZEN 감지 시 L2~L5를 호출하지 않는지 검증한다.
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

	// L1에서 거부되어야 함
	if decision.Kind != harness.DecisionRejected {
		t.Errorf("Kind = %q, want %q", decision.Kind, harness.DecisionRejected)
	}
	if decision.RejectedBy != 1 {
		t.Errorf("RejectedBy = %d, want 1", decision.RejectedBy)
	}

	// L2~L5는 호출되지 않아야 함
	for _, layer := range tracker.order {
		if layer >= 2 {
			t.Errorf("L1 거부 후 L%d가 호출됨 (short-circuit 실패)", layer)
		}
	}
}

// TestPipeline_ShortCircuitOnL2는 L2 canary 실패 시 L3~L5를 호출하지 않는지 검증한다.
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
			t.Errorf("L2 거부 후 L%d가 호출됨 (short-circuit 실패)", layer)
		}
	}
}

// TestPipeline_ShortCircuitOnL3는 L3 contradiction 탐지 시 L4~L5를 호출하지 않는지 검증한다.
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
			t.Errorf("L3 거부 후 L%d가 호출됨 (short-circuit 실패)", layer)
		}
	}
}

// TestPipeline_ShortCircuitOnL4는 L4 rate limit 초과 시 L5를 호출하지 않는지 검증한다.
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
			return false, 24 * time.Hour, nil // 차단
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
			t.Errorf("L4 거부 후 L%d가 호출됨 (short-circuit 실패)", layer)
		}
	}
}

// TestPipeline_L5PendingWhenAutoApplyFalse는 auto_apply=false이면 L5가 pending_approval을 반환하는지 검증한다.
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
		autoApply: false, // L5 항상 pending
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
		t.Error("OversightProposal이 nil이다, L5 pending 시 payload가 있어야 한다")
	}
}

// TestPipeline_ApprovedWhenAutoApplyTrue는 auto_apply=true이면 approved를 반환하는지 검증한다.
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
		autoApply: true, // L5 자동 승인
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

// TestPipeline_L2ErrorPropagated는 L2 오류가 Evaluate에서 전파되는지 검증한다.
func TestPipeline_L2ErrorPropagated(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("canary 내부 오류")

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
		t.Fatal("L2 오류가 전파되지 않았다")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("err = %v, want %v", err, expectedErr)
	}
}

// TestNewPipeline_DefaultConfig는 기본 설정으로 Pipeline을 생성하는지 검증한다.
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
		t.Fatal("NewPipeline이 nil을 반환했다")
	}

	// 기본 pipeline으로 실제 제안 평가 (FROZEN 아닌 경로)
	proposal := harness.Proposal{
		ID:         "default-test",
		TargetPath: ".moai/harness/test.yaml",
		CreatedAt:  time.Now(),
	}

	decision, err := pipe.Evaluate(proposal, nil)
	if err != nil {
		t.Fatalf("기본 Pipeline Evaluate 실패: %v", err)
	}

	// auto_apply=false이므로 pending_approval 또는 rejected
	if decision.Kind == harness.DecisionApproved {
		t.Error("auto_apply=false인데 approved가 반환됨")
	}
}
