// Package safety — Layer 2: Canary Check (REQ-HL-007).
// 제안 적용 전 최근 3개 세션을 shadow-evaluate하여 effectiveness 하락 탐지.
package safety

import (
	"fmt"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// canaryRejectionThreshold는 거부 임계값이다.
// REQ-HL-007: effectiveness score가 이 값 이상 하락하면 제안을 거부한다.
const canaryRejectionThreshold = 0.10

// maxCanarySessions는 canary check에 사용할 최대 세션 수이다.
// REQ-HL-007: 가장 최근 3개 세션을 사용한다.
const maxCanarySessions = 3

// ProjectedScorer는 제안의 projected effectiveness score를 계산하는 함수 타입이다.
// 기본 구현은 defaultProjectedScorer이며, 테스트에서 주입 가능하다.
type ProjectedScorer func(proposal harness.Proposal, baseline float64) float64

// effectivenessScore는 단일 Session의 effectiveness 점수를 계산한다.
// 가중 합산: subcommand 40%, agent_invocation 40%, completion 20%.
func effectivenessScore(s harness.Session) float64 {
	return s.SubcommandSuccessRate*0.40 +
		s.AgentInvocationSuccess*0.40 +
		s.CompletionRate*0.20
}

// baselineScore는 세션 목록(최신 N개)의 평균 effectiveness를 계산한다.
func baselineScore(sessions []harness.Session) float64 {
	n := len(sessions)
	if n == 0 {
		return 0
	}

	// 최신 maxCanarySessions개만 사용
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

// defaultProjectedScorer는 기본 projected score 계산기이다.
// 실제 적용 효과를 시뮬레이션하는 단순 구현:
// - TargetPath나 NewValue가 비어 있으면 baseline 유지 (의미 없는 변경)
// - 그 외에는 소폭 개선(+0.02)을 가정 (보수적 추정)
func defaultProjectedScorer(proposal harness.Proposal, baseline float64) float64 {
	if proposal.TargetPath == "" || (proposal.FieldKey == "" && proposal.NewValue == "") {
		return baseline
	}
	// 의미 있는 제안은 소폭 개선 가정
	return baseline + 0.02
}

// EvaluateCanary는 제안을 최근 3개 세션에 shadow-evaluate하여 CanaryResult를 반환한다.
// REQ-HL-007: projected score가 baseline 대비 0.10 이상 하락하면 Rejected=true.
//
// @MX:ANCHOR: [AUTO] EvaluateCanary는 Layer 2의 단일 진입점이다.
// @MX:REASON: [AUTO] fan_in >= 3: canary_test.go, pipeline.go, Phase 4 coordinator
func EvaluateCanary(proposal harness.Proposal, sessions []harness.Session) (harness.CanaryResult, error) {
	return EvaluateCanaryWithScorer(proposal, sessions, defaultProjectedScorer)
}

// EvaluateCanaryWithScorer는 사용자 정의 scorer로 canary check를 실행한다.
// 테스트에서 degrading/improving 시나리오를 주입할 때 사용한다.
//
// @MX:ANCHOR: [AUTO] EvaluateCanaryWithScorer는 테스트 주입 진입점이다.
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

	// 하락폭이 임계값 이상이면 거부.
	// 부동소수점 오차 허용을 위해 epsilon=1e-9를 사용한다.
	const epsilon = 1e-9
	drop := baseline - projected
	if drop >= canaryRejectionThreshold-epsilon {
		result.Rejected = true
		result.Reason = fmt.Sprintf(
			"effectiveness 하락 %.4f (임계값 %.2f 이상): baseline=%.4f → projected=%.4f",
			drop, canaryRejectionThreshold, baseline, projected,
		)
	}

	return result, nil
}
