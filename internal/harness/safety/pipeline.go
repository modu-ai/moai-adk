// Package safety — Layer Integration Pipeline (REQ-HL-005~008).
// Executes in order: L1(Frozen Guard) → L2(Canary) → L3(Contradiction) → L4(Rate Limiter) → L5(Oversight).
// If rejected at any layer, short-circuits and skips subsequent layers.
package safety

import (
	"fmt"
	"strings"
	"time"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// Pipeline executes the 5-Layer Safety Architecture in order.
// [HARD] L1→L2→L3→L4→L5 order is immutable.
//
// @MX:ANCHOR: [AUTO] Pipeline.Evaluate is the 5-Layer Safety Architecture entry point.
// @MX:REASON: [AUTO] fan_in >= 3: pipeline_test.go, Phase 4 coordinator, harness CLI
// @MX:WARN: [AUTO] l1~l5 function fields are for test injection, use only NewPipeline in production.
// @MX:REASON: [AUTO] Direct initialization can cause panic due to nil function calls.
type Pipeline struct {
	// l1FrozenCheck is Layer 1: FROZEN path verification function.
	l1FrozenCheck func(proposal harness.Proposal) bool

	// l2CanaryCheck is Layer 2: Canary effectiveness verification function.
	l2CanaryCheck func(proposal harness.Proposal, sessions []harness.Session) (harness.CanaryResult, error)

	// l3ContradictionCheck is Layer 3: Contradiction detection function.
	l3ContradictionCheck func(proposal harness.Proposal) harness.ContradictionReport

	// l4RateLimitCheck is Layer 4: Rate limit verification function.
	l4RateLimitCheck func() (allowed bool, retryAfter time.Duration, err error)

	// l5OversightProposal is Layer 5: Human Oversight payload generation function.
	l5OversightProposal func(proposal harness.Proposal, reason string) harness.OversightProposal

	// autoApply: if true, returns approved when all layers pass.
	// If false, returns pending_approval at L5 and orchestrator prompts user for confirmation.
	autoApply bool
}

// PipelineConfig is the configuration required to create NewPipeline.
type PipelineConfig struct {
	// ViolationLogPath is the path to frozen-guard-violations.jsonl file.
	ViolationLogPath string

	// RateLimitPath is the path to rate-limit-state.json file.
	RateLimitPath string

	// AutoApply: if true, automatically applies when all layers pass.
	// REQ-HL-005: default is false (user approval required).
	AutoApply bool
}

// NewPipeline creates a Pipeline with default implementation.
func NewPipeline(cfg PipelineConfig) *Pipeline {
	rl := NewRateLimiter(cfg.RateLimitPath)

	return &Pipeline{
		// Layer 1: hardcoded FROZEN prefix check
		l1FrozenCheck: func(proposal harness.Proposal) bool {
			return IsFrozen(proposal.TargetPath)
		},

		// Layer 2: canary effectiveness check
		l2CanaryCheck: func(proposal harness.Proposal, sessions []harness.Session) (harness.CanaryResult, error) {
			return EvaluateCanary(proposal, sessions)
		},

		// Layer 3: contradiction detection — Frozen-rules consult for the Curator
		// write path (REQ-HEV3-015, SPEC-HARNESS-EVOLVE-003 M3). Replaces the prior
		// no-op (the old "Phase 4" deferral that was never honored). The existing
		// DetectOverlappingTriggers / DetectChainRuleContradictions (harness-applier
		// path) are NOT removed; DetectFrozenRuleContradictions is the Curator-path
		// sibling (design.md §C.4). L1→L2→L3→L4→L5 order is immutable (pipeline.go:14).
		l3ContradictionCheck: DetectFrozenRuleContradictions,

		// Layer 4: sliding-window rate limit
		l4RateLimitCheck: rl.CheckLimit,

		// Layer 5: oversight payload generation
		l5OversightProposal: BuildOversightProposal,

		autoApply: cfg.AutoApply,
	}
}

// Evaluate evaluates the proposal in L1→L2→L3→L4→L5 order and returns Decision.
// [HARD] Order is always 1→2→3→4→5, with immediate return on rejection at any layer (short-circuit).
//
// sessions is the list of recent sessions used for L2 canary check (skip if nil).
func (p *Pipeline) Evaluate(proposal harness.Proposal, sessions []harness.Session) (harness.Decision, error) {
	// ── Layer 1: Frozen Guard ──────────────────────────────────────────────
	if p.l1FrozenCheck(proposal) {
		return harness.Decision{
			Kind:       harness.DecisionRejected,
			RejectedBy: 1,
			Reason: fmt.Sprintf(
				"L1 FROZEN Guard: path '%s' is in moai-managed FROZEN zone",
				proposal.TargetPath,
			),
		}, nil
	}

	// ── Layer 2: Canary Check ─────────────────────────────────────────────
	canaryResult, err := p.l2CanaryCheck(proposal, sessions)
	if err != nil {
		return harness.Decision{}, fmt.Errorf("safety/pipeline: L2 canary check error: %w", err)
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
		// REQ-HEV3-017: cite the contradiction item Descriptions (which carry the
		// Frozen-rule stable identifier for the Curator path, or the trigger/chain
		// detail for the harness-applier path) in the rejection Reason.
		var descs []string
		for _, item := range contradictionReport.Items {
			descs = append(descs, item.Description)
		}
		return harness.Decision{
			Kind:       harness.DecisionRejected,
			RejectedBy: 3,
			Reason: fmt.Sprintf(
				"L3 Contradiction: %d detected: %s",
				len(contradictionReport.Items),
				strings.Join(descs, "; "),
			),
		}, nil
	}

	// ── Layer 4: Rate Limiter ──────────────────────────────────────────────
	allowed, retryAfter, err := p.l4RateLimitCheck()
	if err != nil {
		return harness.Decision{}, fmt.Errorf("safety/pipeline: L4 rate limit check error: %w", err)
	}
	if !allowed {
		return harness.Decision{
			Kind:       harness.DecisionRejected,
			RejectedBy: 4,
			Reason: fmt.Sprintf(
				"L4 Rate Limiter: retry available after %v",
				retryAfter.Round(time.Minute),
			),
		}, nil
	}

	// ── Layer 5: Human Oversight ───────────────────────────────────────────
	// If autoApply=false, always request user approval
	// If autoApply=true, auto-approve
	if !p.autoApply {
		op := p.l5OversightProposal(proposal, "All safety layers passed — user final approval required")
		return harness.Decision{
			Kind:              harness.DecisionPendingApproval,
			OversightProposal: &op,
		}, nil
	}

	return harness.Decision{
		Kind: harness.DecisionApproved,
	}, nil
}
