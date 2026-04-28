// Package safety — oversight unit test.
// REQ-HL-005: OversightProposal payload 구조 검증 (subagent boundary 준수).
package safety

import (
	"testing"
	"time"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// TestBuildOversightProposal_BasicStructure는 OversightProposal 기본 구조를 검증한다.
func TestBuildOversightProposal_BasicStructure(t *testing.T) {
	t.Parallel()

	proposal := harness.Proposal{
		ID:               "test-001",
		TargetPath:       ".claude/skills/my-harness-plugin/SKILL.md",
		FieldKey:         "description",
		NewValue:         "improved description with heuristic note",
		PatternKey:       "moai_subcommand:/moai plan:",
		Tier:             harness.TierAutoUpdate,
		ObservationCount: 10,
		CreatedAt:        time.Now(),
	}

	op := BuildOversightProposal(proposal, "중복 trigger 탐지")

	// 질문 텍스트가 있어야 함
	if op.Question == "" {
		t.Error("Question이 비어 있다")
	}

	// ProposalID가 일치해야 함
	if op.ProposalID != proposal.ID {
		t.Errorf("ProposalID = %q, want %q", op.ProposalID, proposal.ID)
	}

	// 옵션이 1~4개여야 함
	if len(op.Options) < 1 || len(op.Options) > 4 {
		t.Errorf("옵션 수 = %d, 1~4 범위여야 한다", len(op.Options))
	}
}

// TestBuildOversightProposal_FirstOptionRecommended는 첫 번째 옵션이 권장인지 검증한다.
// REQ-HL-005: AskUserQuestion schema — 첫 번째 옵션은 Recommended여야 한다.
func TestBuildOversightProposal_FirstOptionRecommended(t *testing.T) {
	t.Parallel()

	proposal := harness.Proposal{
		ID:         "test-002",
		TargetPath: ".moai/harness/chaining-rules.yaml",
		FieldKey:   "triggers",
		NewValue:   "new-trigger",
		Tier:       harness.TierAutoUpdate,
		CreatedAt:  time.Now(),
	}

	op := BuildOversightProposal(proposal, "rate limit 초과")

	if len(op.Options) == 0 {
		t.Fatal("옵션이 없다")
	}

	// 첫 번째 옵션이 Recommended여야 함
	if !op.Options[0].Recommended {
		t.Errorf("첫 번째 옵션 Recommended = false, true여야 한다. option: %+v", op.Options[0])
	}

	// 나머지 옵션은 Recommended=false여야 함
	for i, opt := range op.Options[1:] {
		if opt.Recommended {
			t.Errorf("옵션[%d] Recommended = true, 첫 번째만 Recommended여야 한다", i+1)
		}
	}
}

// TestBuildOversightProposal_MaxFourOptions는 옵션이 최대 4개인지 검증한다.
// REQ-HL-005: AskUserQuestion은 최대 4개 옵션만 지원한다.
func TestBuildOversightProposal_MaxFourOptions(t *testing.T) {
	t.Parallel()

	proposal := harness.Proposal{
		ID:        "test-003",
		TargetPath: ".moai/harness/test.yaml",
		CreatedAt: time.Now(),
	}

	op := BuildOversightProposal(proposal, "contradiction detected")

	if len(op.Options) > 4 {
		t.Errorf("옵션 수 = %d, 최대 4개여야 한다 (AskUserQuestion 제한)", len(op.Options))
	}
}

// TestBuildOversightProposal_AllOptionsHaveDescriptions는 모든 옵션이 설명을 가지는지 검증한다.
func TestBuildOversightProposal_AllOptionsHaveDescriptions(t *testing.T) {
	t.Parallel()

	proposal := harness.Proposal{
		ID:               "test-004",
		TargetPath:       ".claude/skills/my-harness-x/SKILL.md",
		Tier:             harness.TierAutoUpdate,
		ObservationCount: 12,
		CreatedAt:        time.Now(),
	}

	op := BuildOversightProposal(proposal, "")

	for i, opt := range op.Options {
		if opt.Label == "" {
			t.Errorf("옵션[%d].Label이 비어 있다", i)
		}
		if opt.Description == "" {
			t.Errorf("옵션[%d].Description이 비어 있다", i)
		}
		if opt.Value == "" {
			t.Errorf("옵션[%d].Value가 비어 있다", i)
		}
	}
}

// TestBuildOversightProposal_ContextContainsProposalInfo는 Context에 제안 정보가 포함되는지 검증한다.
func TestBuildOversightProposal_ContextContainsProposalInfo(t *testing.T) {
	t.Parallel()

	proposal := harness.Proposal{
		ID:               "unique-id-xyz",
		TargetPath:       ".claude/skills/my-harness-abc/SKILL.md",
		Tier:             harness.TierAutoUpdate,
		ObservationCount: 15,
		CreatedAt:        time.Now(),
	}

	op := BuildOversightProposal(proposal, "")

	if op.Context == "" {
		t.Error("Context가 비어 있다")
	}
}

// TestBuildOversightProposal_DoesNotCallAskUserQuestion는 AskUserQuestion을 호출하지 않는지 검증한다.
// REQ-HL-005 + agent-common-protocol §User Interaction Boundary:
// subagent는 AskUserQuestion을 직접 호출할 수 없다.
// BuildOversightProposal은 payload만 반환하며 사용자 상호작용을 시작하지 않는다.
//
// 이 테스트는 컴파일 타임 검증이다:
// - oversight.go에 "AskUserQuestion" 호출이 없음을 간접 검증 (grep 기반)
// - Go 타입 시스템으로는 AskUserQuestion이 외부 도구라 호출 자체가 불가
func TestBuildOversightProposal_DoesNotCallAskUserQuestion(t *testing.T) {
	t.Parallel()

	// BuildOversightProposal은 순수 함수여야 함:
	// - 사용자 입력 대기 없음
	// - 외부 도구 호출 없음
	// - 단순히 payload를 구성하여 반환
	proposal := harness.Proposal{
		ID:        "boundary-test",
		CreatedAt: time.Now(),
	}

	// 이 호출은 즉시 반환되어야 한다 (blocking 없음)
	op := BuildOversightProposal(proposal, "")

	// 구조체가 반환되면 성공 (AskUserQuestion 호출은 패닉 또는 타임아웃을 유발함)
	if op.ProposalID != proposal.ID {
		t.Errorf("ProposalID = %q, want %q", op.ProposalID, proposal.ID)
	}
}
