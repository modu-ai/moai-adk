// Package safety — Layer 통합 Pipeline (REQ-HL-005~008).
// L1(Frozen Guard) → L2(Canary) → L3(Contradiction) → L4(Rate Limiter) → L5(Oversight) 순서로 실행.
// 각 layer에서 거부되면 short-circuit으로 이후 layer를 건너뛴다.
package safety

import (
	"fmt"
	"time"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// Pipeline은 5-Layer Safety Architecture를 순서대로 실행한다.
// [HARD] L1→L2→L3→L4→L5 순서는 불변이다.
//
// @MX:ANCHOR: [AUTO] Pipeline.Evaluate는 5-Layer Safety Architecture 진입점이다.
// @MX:REASON: [AUTO] fan_in >= 3: pipeline_test.go, Phase 4 coordinator, harness CLI
// @MX:WARN: [AUTO] l1~l5 함수 필드는 테스트 주입용이며 프로덕션에서는 NewPipeline으로만 생성.
// @MX:REASON: [AUTO] 직접 초기화 시 nil 함수 호출로 패닉 발생 가능.
type Pipeline struct {
	// l1FrozenCheck는 Layer 1: FROZEN 경로 확인 함수이다.
	l1FrozenCheck func(proposal harness.Proposal) bool

	// l2CanaryCheck는 Layer 2: Canary effectiveness 확인 함수이다.
	l2CanaryCheck func(proposal harness.Proposal, sessions []harness.Session) (harness.CanaryResult, error)

	// l3ContradictionCheck는 Layer 3: Contradiction 탐지 함수이다.
	l3ContradictionCheck func(proposal harness.Proposal) harness.ContradictionReport

	// l4RateLimitCheck는 Layer 4: Rate limit 확인 함수이다.
	l4RateLimitCheck func() (allowed bool, retryAfter time.Duration, err error)

	// l5OversightProposal는 Layer 5: Human Oversight payload 생성 함수이다.
	l5OversightProposal func(proposal harness.Proposal, reason string) harness.OversightProposal

	// autoApply는 true이면 모든 layer 통과 시 approved를 반환한다.
	// false이면 L5에서 pending_approval을 반환하며 orchestrator가 사용자에게 확인한다.
	autoApply bool
}

// PipelineConfig는 NewPipeline 생성에 필요한 설정이다.
type PipelineConfig struct {
	// ViolationLogPath는 frozen-guard-violations.jsonl 파일 경로이다.
	ViolationLogPath string

	// RateLimitPath는 rate-limit-state.json 파일 경로이다.
	RateLimitPath string

	// AutoApply는 true이면 모든 layer 통과 시 자동 적용한다.
	// REQ-HL-005: 기본값 false (사용자 승인 필요).
	AutoApply bool
}

// NewPipeline은 기본 구현으로 Pipeline을 생성한다.
func NewPipeline(cfg PipelineConfig) *Pipeline {
	rl := NewRateLimiter(cfg.RateLimitPath)

	return &Pipeline{
		// Layer 1: hardcoded FROZEN 접두사 검사
		l1FrozenCheck: func(proposal harness.Proposal) bool {
			return IsFrozen(proposal.TargetPath)
		},

		// Layer 2: canary effectiveness check
		l2CanaryCheck: func(proposal harness.Proposal, sessions []harness.Session) (harness.CanaryResult, error) {
			return EvaluateCanary(proposal, sessions)
		},

		// Layer 3: contradiction detection (빈 trigger 목록으로 호출 — 실제 triggers는 Phase 4에서 주입)
		l3ContradictionCheck: func(_ harness.Proposal) harness.ContradictionReport {
			// Phase 3에서는 항상 빈 보고서 반환 (Phase 4에서 실제 skill trigger 로딩)
			return harness.ContradictionReport{}
		},

		// Layer 4: sliding-window rate limit
		l4RateLimitCheck: rl.CheckLimit,

		// Layer 5: oversight payload 생성
		l5OversightProposal: BuildOversightProposal,

		autoApply: cfg.AutoApply,
	}
}

// Evaluate는 제안을 L1→L2→L3→L4→L5 순서로 평가하여 Decision을 반환한다.
// [HARD] 순서는 항상 1→2→3→4→5이며, 각 layer에서 거부되면 즉시 반환(short-circuit).
//
// sessions는 L2 canary check에 사용되는 최근 세션 목록이다 (nil이면 skip).
func (p *Pipeline) Evaluate(proposal harness.Proposal, sessions []harness.Session) (harness.Decision, error) {
	// ── Layer 1: Frozen Guard ──────────────────────────────────────────────
	if p.l1FrozenCheck(proposal) {
		return harness.Decision{
			Kind:       harness.DecisionRejected,
			RejectedBy: 1,
			Reason: fmt.Sprintf(
				"L1 FROZEN Guard: 경로 '%s'는 moai-managed FROZEN 영역입니다",
				proposal.TargetPath,
			),
		}, nil
	}

	// ── Layer 2: Canary Check ─────────────────────────────────────────────
	canaryResult, err := p.l2CanaryCheck(proposal, sessions)
	if err != nil {
		return harness.Decision{}, fmt.Errorf("safety/pipeline: L2 canary check 오류: %w", err)
	}
	if canaryResult.Rejected {
		return harness.Decision{
			Kind:       harness.DecisionRejected,
			RejectedBy: 2,
			Reason:     fmt.Sprintf("L2 Canary Check: %s", canaryResult.Reason),
		}, nil
	}

	// ── Layer 3: Contradiction Detector ────────────────────────────────────
	contradictionReport := p.l3ContradictionCheck(proposal)
	if contradictionReport.HasContradiction() {
		return harness.Decision{
			Kind:       harness.DecisionRejected,
			RejectedBy: 3,
			Reason: fmt.Sprintf(
				"L3 Contradiction: %d개 모순 탐지됨",
				len(contradictionReport.Items),
			),
		}, nil
	}

	// ── Layer 4: Rate Limiter ──────────────────────────────────────────────
	allowed, retryAfter, err := p.l4RateLimitCheck()
	if err != nil {
		return harness.Decision{}, fmt.Errorf("safety/pipeline: L4 rate limit check 오류: %w", err)
	}
	if !allowed {
		return harness.Decision{
			Kind:       harness.DecisionRejected,
			RejectedBy: 4,
			Reason: fmt.Sprintf(
				"L4 Rate Limiter: %v 후 재시도 가능",
				retryAfter.Round(time.Minute),
			),
		}, nil
	}

	// ── Layer 5: Human Oversight ───────────────────────────────────────────
	// autoApply=false이면 항상 사용자 승인 요청
	// autoApply=true이면 자동 승인
	if !p.autoApply {
		op := p.l5OversightProposal(proposal, "모든 safety layer 통과 — 사용자 최종 승인 필요")
		return harness.Decision{
			Kind:              harness.DecisionPendingApproval,
			OversightProposal: &op,
		}, nil
	}

	return harness.Decision{
		Kind: harness.DecisionApproved,
	}, nil
}
