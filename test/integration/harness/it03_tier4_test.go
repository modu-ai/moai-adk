//go:build integration
// +build integration

// Package harness_integration — T-P5-03: Tier 4 mock 승인 플로우 검증 (REQ-HL-005).
package harness_integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// mockApprovedEvaluator는 항상 DecisionApproved를 반환하는 mock SafetyEvaluator이다.
// T-P5-03: orchestrator가 OversightProposal을 승인한 이후의 applier 동작을 검증한다.
type mockApprovedEvaluator struct{}

// Evaluate는 항상 DecisionApproved를 반환한다 (orchestrator 승인 완료 시뮬레이션).
func (m *mockApprovedEvaluator) Evaluate(
	_ harness.Proposal,
	_ []harness.Session,
) (harness.Decision, error) {
	return harness.Decision{Kind: harness.DecisionApproved}, nil
}

// mockPendingEvaluator는 항상 DecisionPendingApproval을 반환하는 mock SafetyEvaluator이다.
// T-P5-03: CLI가 OversightProposal payload를 orchestrator에게 반환하는지 검증한다.
type mockPendingEvaluator struct {
	// returnedProposal은 Evaluate 호출 시 반환할 OversightProposal이다.
	returnedProposal *harness.OversightProposal
}

// Evaluate는 항상 DecisionPendingApproval을 반환한다 (L5 human oversight 시뮬레이션).
func (m *mockPendingEvaluator) Evaluate(
	proposal harness.Proposal,
	_ []harness.Session,
) (harness.Decision, error) {
	op := harness.OversightProposal{
		Question:   "자동 학습 업데이트 승인 요청",
		ProposalID: proposal.ID,
		Context:    "mock oversight",
		Options: []harness.OversightOption{
			{Label: "승인 (권장)", Value: "approve", Recommended: true},
			{Label: "거부", Value: "reject"},
		},
	}
	m.returnedProposal = &op
	return harness.Decision{
		Kind:              harness.DecisionPendingApproval,
		OversightProposal: &op,
	}, nil
}

// TestIT03_Tier4MockApprovalFlow는 Tier 4 패턴에 대해 두 가지 플로우를 검증한다.
// 1. PendingApproval: pipeline이 pending 반환 → ApplyPendingError 발생, payload 확인
// 2. Approved: mock orchestrator가 승인 → applier가 파일을 수정
// REQ-HL-005: Tier 4 도달 시 system이 OversightProposal을 orchestrator에게 반환한다.
func TestIT03_Tier4MockApprovalFlow(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := setupSkillFile(t, dir, "my-harness-x", `---
name: my-harness-x
description: X 스킬
---

# X Skill Body
`)

	snapshotBase := filepath.Join(dir, "snapshots")

	proposal := harness.Proposal{
		ID:               "prop-tier4-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "Tier 4 자동 업데이트 값",
		PatternKey:       "agent_invocation:expert-backend",
		Tier:             harness.TierAutoUpdate,
		ObservationCount: 10,
		CreatedAt:        time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
	}

	// ── 시나리오 A: pending_approval 반환 검증 ────────────────────────────
	t.Run("pending_approval_returns_payload", func(t *testing.T) {
		pendingEval := &mockPendingEvaluator{}
		applier := harness.NewApplier()

		err := applier.Apply(proposal, pendingEval, snapshotBase, nil)
		if err == nil {
			t.Fatal("Apply가 error를 반환해야 하는데 nil 반환")
		}

		// ApplyPendingError 타입 검증
		pendingErr, ok := err.(*harness.ApplyPendingError)
		if !ok {
			t.Fatalf("error 타입 = %T, want *harness.ApplyPendingError", err)
		}

		// OversightPayload 내용 검증
		if pendingErr.OversightPayload == nil {
			t.Fatal("OversightPayload가 nil이어서는 안 된다")
		}
		if pendingErr.OversightPayload.ProposalID != "prop-tier4-001" {
			t.Errorf("ProposalID = %q, want prop-tier4-001", pendingErr.OversightPayload.ProposalID)
		}
		if len(pendingErr.OversightPayload.Options) == 0 {
			t.Error("OversightPayload.Options가 비어 있음")
		}
		// 첫 번째 옵션이 권장 선택지인지 검증
		firstOpt := pendingErr.OversightPayload.Options[0]
		if !firstOpt.Recommended {
			t.Error("첫 번째 옵션이 Recommended=true여야 한다")
		}
	})

	// ── 시나리오 B: orchestrator 승인 후 파일 수정 검증 ──────────────────
	t.Run("approved_applies_modification", func(t *testing.T) {
		approvedDir := t.TempDir()
		approvedSkillPath := setupSkillFile(t, approvedDir, "my-harness-y", `---
name: my-harness-y
description: Y 스킬
---

# Y Skill Body
`)
		approvedProposal := harness.Proposal{
			ID:               "prop-tier4-002",
			TargetPath:       approvedSkillPath,
			FieldKey:         "description",
			NewValue:         "orchestrator 승인 후 적용값",
			PatternKey:       "agent_invocation:expert-backend",
			Tier:             harness.TierAutoUpdate,
			ObservationCount: 12,
			CreatedAt:        time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
		}

		approvedEval := &mockApprovedEvaluator{}
		applier := harness.NewApplier()
		snapshotDir := filepath.Join(approvedDir, "snapshots")

		if err := applier.Apply(approvedProposal, approvedEval, snapshotDir, nil); err != nil {
			t.Fatalf("Apply 실패 (승인 후): %v", err)
		}

		// 파일이 수정되었는지 검증
		content, err := os.ReadFile(approvedSkillPath)
		if err != nil {
			t.Fatalf("수정된 SKILL.md 읽기 실패: %v", err)
		}
		if !strings.Contains(string(content), "orchestrator 승인 후 적용값") {
			t.Errorf("파일에 NewValue가 반영되지 않음:\n%s", content)
		}

		// 스냅샷 생성 검증
		snapEntries, err := os.ReadDir(snapshotDir)
		if err != nil {
			t.Fatalf("스냅샷 디렉토리 읽기 실패: %v", err)
		}
		if len(snapEntries) == 0 {
			t.Fatal("스냅샷이 생성되지 않음")
		}
	})
}

// setupSkillFile은 지정된 content로 SKILL.md 파일을 생성하고 경로를 반환한다.
func setupSkillFile(t *testing.T, dir, skillName, content string) string {
	t.Helper()
	skillDir := filepath.Join(dir, "skills", skillName)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("skill 디렉토리 생성 실패: %v", err)
	}
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(content), 0o644); err != nil {
		t.Fatalf("SKILL.md 쓰기 실패: %v", err)
	}
	return skillPath
}
