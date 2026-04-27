// Package safety — Layer 5: Human Oversight (REQ-HL-005).
// 자동 업데이트 승인 여부를 orchestrator에게 위임하기 위한 proposal payload를 생성한다.
//
// [HARD] 이 파일은 AskUserQuestion을 직접 호출하지 않는다.
// agent-common-protocol §User Interaction Boundary:
// subagent는 사용자와 직접 상호작용 불가. orchestrator(MoAI)가 이 payload를 받아
// AskUserQuestion으로 사용자에게 표시한다.
package safety

import (
	"fmt"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// BuildOversightProposal은 orchestrator가 AskUserQuestion에 사용할 OversightProposal을 생성한다.
// REQ-HL-005: Tier 4 패턴 자동 업데이트는 사용자 승인이 필요하다.
//
// reason은 이 제안이 인간 감독을 요청하는 이유이다 (예: "contradiction detected", "rate limit exceeded").
//
// [HARD] 이 함수는 AskUserQuestion을 호출하지 않는다. payload만 반환한다.
// orchestrator가 반환값을 받아 AskUserQuestion(options=payload.Options)으로 사용자에게 표시한다.
//
// @MX:ANCHOR: [AUTO] BuildOversightProposal은 L5 subagent→orchestrator 인터페이스이다.
// @MX:REASON: [AUTO] fan_in >= 3: oversight_test.go, pipeline.go, Phase 4 coordinator
func BuildOversightProposal(proposal harness.Proposal, reason string) harness.OversightProposal {
	// 컨텍스트 정보 구성
	contextStr := buildContext(proposal, reason)

	// 질문 텍스트
	question := fmt.Sprintf(
		"자동 학습 업데이트 승인 요청: '%s' 파일의 '%s' 필드를 수정하려 합니다. (%d회 관찰)",
		proposal.TargetPath, proposal.FieldKey, proposal.ObservationCount,
	)
	if proposal.TargetPath == "" {
		question = "자동 학습 업데이트 승인 요청: 패턴 기반 설정 변경을 승인하시겠습니까?"
	}

	// AskUserQuestion 옵션 구성 (최대 4개, 첫 번째 권장)
	// [HARD] 첫 번째 옵션은 Recommended=true
	options := []harness.OversightOption{
		{
			Label:       "승인 (권장)",
			Description: "제안된 자동 업데이트를 승인합니다. 시스템이 즉시 변경을 적용합니다.",
			Value:       "approve",
			Recommended: true,
		},
		{
			Label:       "거부",
			Description: "이 업데이트를 거부합니다. 시스템은 현재 설정을 유지합니다.",
			Value:       "reject",
			Recommended: false,
		},
		{
			Label:       "7일 연기",
			Description: "이 업데이트를 7일 후 재검토합니다. 현재 설정은 유지됩니다.",
			Value:       "defer",
			Recommended: false,
		},
	}

	return harness.OversightProposal{
		Question:   question,
		Options:    options,
		ProposalID: proposal.ID,
		Context:    contextStr,
	}
}

// buildContext는 OversightProposal의 Context 필드를 구성한다.
func buildContext(proposal harness.Proposal, reason string) string {
	ctx := fmt.Sprintf(
		"패턴 키: %s | Tier: %s | 관찰 횟수: %d",
		proposal.PatternKey, proposal.Tier.String(), proposal.ObservationCount,
	)
	if proposal.TargetPath != "" {
		ctx += fmt.Sprintf(" | 대상 파일: %s", proposal.TargetPath)
	}
	if proposal.FieldKey != "" {
		ctx += fmt.Sprintf(" | 수정 필드: %s", proposal.FieldKey)
	}
	if proposal.NewValue != "" {
		ctx += fmt.Sprintf(" | 새 값: %s", proposal.NewValue)
	}
	if reason != "" {
		ctx += fmt.Sprintf(" | 요청 이유: %s", reason)
	}
	return ctx
}
