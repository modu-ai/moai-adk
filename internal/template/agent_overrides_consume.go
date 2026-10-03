package template

// agent_overrides_consume.go — the spawn-consumption resolver behind the
// llm.agent_overrides_consume opt-in gate (REQ-AFR-016,
// SPEC-WEB-AGENTFM-RESTORE-001 v0.3.0 M7, card t1421).
//
// Semantics (AC-AFR-015):
//
//   - gate OFF (the key absent or false): the resolver consumes nothing —
//     the override map stays console-stored-only and every spawn keeps the
//     session-inherit default (REQ-AFR-002, byte-for-byte today's behaviour).
//   - gate ON, entry ABSENT: plain inheritance (the session model/effort) —
//     an absent entry is the default, NOT the profile-matrix cell. This is
//     the deliberate divergence from ResolveAgentModelEffort, whose console
//     chain falls through to the agent's default cell: the profile column is
//     console-default machinery for pin/clear comparison, and feeding it
//     here would silently turn a key-less override map into spawn pins.
//     That divergence is why this resolver looks the override entry up
//     directly instead of calling ResolveAgentModelEffort and discarding its
//     cell fallback — the latter would be a second, masked derivation.
//
//   - gate ON, entry present with model `inherit` (or the model axis unset):
//     an explicit inheritance no-op.
//   - gate ON, entry present with a concrete model alias: the override wins.
//
// Single derivation: the resolver consumes the SAME validated override map
// the console chain reads (cfg.AgentOverrides — validated upstream by
// config.validateAgentOverrides against the REQ-AFR-006 closed sets) and the
// same ModelInherit sentinel. It re-derives neither the closed sets nor the
// profile resolution.

import (
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// AgentOverrideConsumption is the resolved spawn-consumption outcome for one
// agent. It is also the contract-state shape the visibility surfaces report
// (REQ-AFR-019): ConsumeEnabled is the switch state itself.
type AgentOverrideConsumption struct {
	// ConsumeEnabled mirrors the llm.agent_overrides_consume switch. False
	// means storage-only: no field below is ever set.
	ConsumeEnabled bool
	// Pinned is true only when the gate is on and the agent's override entry
	// pins a concrete model alias — the orchestrator passes Model on the
	// Agent() call (the one approved consumption channel,
	// model-policy.md's subagent model resolution order).
	Pinned bool
	// Model is the pinned model alias; empty unless Pinned.
	Model string
	// Effort is the override entry's effort value, carried for the
	// doctrine-level delivery (REQ-AFR-017 — the rules text owns effort
	// delivery). The spawn payload transports model only: the Agent tool
	// exposes no effort parameter (the agent_model_guard header contract,
	// REQ-AFR-018), so effort never takes part in a hook comparison.
	Effort string
	// InheritNoOp is true when the gate is on and the entry's model axis is
	// the explicit inherit sentinel (or unset) — explicit inheritance, not a
	// pin.
	InheritNoOp bool
}

// ResolveAgentOverrideConsumption resolves the spawn-consumption outcome for
// agent under cfg, behind the llm.agent_overrides_consume opt-in gate.
//
// @MX:ANCHOR: [AUTO] ResolveAgentOverrideConsumption — the opt-in consumption gate's single resolver
// @MX:REASON: [AUTO] fan_in >= 2 (hook advise compare + console/doctor contract-state surfaces); the gate-off zero outcome is load-bearing — any leaked pin under a closed gate breaks REQ-AFR-002's key-less-session invariant
func ResolveAgentOverrideConsumption(cfg config.LLMConfig, agent string) AgentOverrideConsumption {
	if !cfg.AgentOverridesConsume {
		// Gate closed — storage-only. Zero outcome: nothing leaks through.
		return AgentOverrideConsumption{}
	}
	out := AgentOverrideConsumption{ConsumeEnabled: true}

	ov, ok := cfg.AgentOverrides[agent]
	if !ok {
		// Absent entry — plain inheritance is the default (REQ-AFR-016):
		// ConsumeEnabled with no pin and no no-op mark.
		return out
	}

	out.Effort = ov.Effort
	switch m := strings.TrimSpace(ov.Model); m {
	case "", ModelInherit:
		// Explicit inheritance no-op (the model axis unset reads the same —
		// config.validateAgentOverrides admits an effort-only entry).
		out.InheritNoOp = true
	default:
		out.Pinned = true
		out.Model = m
	}
	return out
}
