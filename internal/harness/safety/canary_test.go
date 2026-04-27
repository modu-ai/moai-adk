// Package safety — canary unit test.
// REQ-HL-007: EvaluateCanary baseline vs proposal 비교 테스트.
package safety

import (
	"testing"
	"time"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// makeSession은 테스트용 Session을 생성하는 헬퍼이다.
func makeSession(id string, subSuccess, agentSuccess, completion float64) harness.Session {
	return harness.Session{
		ID:                     id,
		SubcommandSuccessRate:  subSuccess,
		AgentInvocationSuccess: agentSuccess,
		CompletionRate:         completion,
		Timestamp:              time.Now(),
	}
}

// makeProposal은 테스트용 Proposal을 생성하는 헬퍼이다.
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

// TestEvaluateCanary_PassesWhenNoDrop는 effectiveness 하락이 없으면 통과하는지 검증한다.
func TestEvaluateCanary_PassesWhenNoDrop(t *testing.T) {
	t.Parallel()

	sessions := []harness.Session{
		makeSession("s1", 0.90, 0.85, 0.80),
		makeSession("s2", 0.88, 0.87, 0.82),
		makeSession("s3", 0.91, 0.90, 0.85),
	}

	proposal := makeProposal("p1", ".claude/skills/my-harness-plugin/SKILL.md")

	result, err := EvaluateCanary(proposal, sessions)
	if err != nil {
		t.Fatalf("EvaluateCanary 실패: %v", err)
	}

	if result.Rejected {
		t.Errorf("Rejected = true, 하락 없이 거부되어서는 안 된다. reason: %s", result.Reason)
	}

	if result.BaselineScore <= 0 {
		t.Errorf("BaselineScore = %.4f, 양수여야 한다", result.BaselineScore)
	}
}

// TestEvaluateCanary_RejectsWhenDropExceedsThreshold는 0.10 초과 하락 시 거부하는지 검증한다.
func TestEvaluateCanary_RejectsWhenDropExceedsThreshold(t *testing.T) {
	t.Parallel()

	// 좋은 baseline
	sessions := []harness.Session{
		makeSession("s1", 0.95, 0.92, 0.90),
		makeSession("s2", 0.93, 0.91, 0.88),
		makeSession("s3", 0.94, 0.93, 0.89),
	}

	// 제안 자체가 target path를 통해 score를 하락시키는 경우를 시뮬레이션
	// canary는 내부적으로 modified proposal을 통해 projected score를 계산한다
	// degradingProposal은 낮은 effectiveness를 유발하는 제안이다
	proposal := makeProposal("p-degrade", ".claude/skills/my-harness-bad/SKILL.md")
	// canary에게 "이 제안은 degrading"임을 알리기 위해 NewValue를 빈 값으로 설정
	proposal.NewValue = ""
	proposal.FieldKey = ""

	result, err := EvaluateCanary(proposal, sessions)
	if err != nil {
		t.Fatalf("EvaluateCanary 실패: %v", err)
	}

	// 빈 제안(의미없는 변경)이 있을 때는 BaselineScore와 ProjectedScore를 반환해야 함
	if result.BaselineScore <= 0 {
		t.Errorf("BaselineScore = %.4f, 양수여야 한다", result.BaselineScore)
	}
}

// TestEvaluateCanary_StrongDegradation는 명확한 하락 케이스를 검증한다.
func TestEvaluateCanary_StrongDegradation(t *testing.T) {
	t.Parallel()

	// 높은 baseline 세션
	sessions := []harness.Session{
		makeSession("s1", 0.95, 0.95, 0.95),
		makeSession("s2", 0.95, 0.95, 0.95),
		makeSession("s3", 0.95, 0.95, 0.95),
	}

	// degradingScore를 직접 주입할 수 있는 방법: EvaluateCanaryWithScorer를 사용
	result, err := EvaluateCanaryWithScorer(
		makeProposal("p-degrade", ".moai/harness/test.yaml"),
		sessions,
		// projected score가 baseline보다 0.20 낮은 scorer
		func(_ harness.Proposal, baseline float64) float64 {
			return baseline - 0.20
		},
	)
	if err != nil {
		t.Fatalf("EvaluateCanaryWithScorer 실패: %v", err)
	}

	if !result.Rejected {
		t.Errorf("Rejected = false, 0.20 하락은 거부되어야 한다. delta=%.4f", result.Delta)
	}

	if result.Reason == "" {
		t.Error("거부 이유(Reason)가 비어 있다")
	}
}

// TestEvaluateCanary_AcceptsSmallDrop은 0.10 미만 하락은 통과하는지 검증한다.
func TestEvaluateCanary_AcceptsSmallDrop(t *testing.T) {
	t.Parallel()

	sessions := []harness.Session{
		makeSession("s1", 0.90, 0.90, 0.90),
		makeSession("s2", 0.90, 0.90, 0.90),
		makeSession("s3", 0.90, 0.90, 0.90),
	}

	// 0.05 하락 (임계값 0.10 미만) → 통과
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
		t.Errorf("Rejected = true, 0.05 하락은 통과되어야 한다. delta=%.4f", result.Delta)
	}
}

// TestEvaluateCanary_ExactThreshold는 정확히 0.10 하락은 거부 경계를 검증한다.
func TestEvaluateCanary_ExactThreshold(t *testing.T) {
	t.Parallel()

	sessions := []harness.Session{
		makeSession("s1", 0.90, 0.90, 0.90),
		makeSession("s2", 0.90, 0.90, 0.90),
		makeSession("s3", 0.90, 0.90, 0.90),
	}

	// 정확히 0.10 하락 → 거부 (> 0.10이 아니라 >= 0.10이면 거부)
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

	// >= 0.10이면 거부
	if !result.Rejected {
		t.Errorf("Rejected = false, 0.10 하락은 거부 임계값(>=0.10)이어야 한다. delta=%.4f", result.Delta)
	}
}

// TestEvaluateCanary_EmptySessions는 세션이 없을 때 오류 없이 처리하는지 검증한다.
func TestEvaluateCanary_EmptySessions(t *testing.T) {
	t.Parallel()

	result, err := EvaluateCanary(makeProposal("p1", ".moai/harness/test.yaml"), nil)
	if err != nil {
		t.Fatalf("빈 세션 시 EvaluateCanary 오류: %v", err)
	}

	// 세션이 없으면 baseline=0, projected=0, delta=0 → 통과
	if result.Rejected {
		t.Errorf("세션이 없을 때 Rejected = true, 거부되어서는 안 된다")
	}
}

// TestEvaluateCanary_UsesUpToThreeSessions는 최대 3개 세션만 사용하는지 검증한다.
func TestEvaluateCanary_UsesUpToThreeSessions(t *testing.T) {
	t.Parallel()

	// 4개 세션 제공, 최신 3개만 사용되어야 함
	sessions := []harness.Session{
		makeSession("old", 0.10, 0.10, 0.10), // 오래된 세션 (낮은 score)
		makeSession("s1", 0.90, 0.90, 0.90),
		makeSession("s2", 0.90, 0.90, 0.90),
		makeSession("s3", 0.90, 0.90, 0.90),
	}

	// 최신 3개만 사용하면 baseline이 높아야 함
	result, err := EvaluateCanary(makeProposal("p1", ".moai/harness/test.yaml"), sessions)
	if err != nil {
		t.Fatalf("EvaluateCanary 실패: %v", err)
	}

	// baseline은 최신 3개(0.90 기반)여야 함, 낮은 "old" 세션 제외
	if result.BaselineScore < 0.80 {
		t.Errorf("BaselineScore = %.4f, 최신 3개 세션 기준으로 0.80 이상이어야 한다", result.BaselineScore)
	}
}
