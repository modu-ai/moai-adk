package template

// glm_effort_overlay.go — SPEC-MODEL-TIER-PLANTYPE-001 M5: the GLM backend
// effort overlay (§B.8, REQ-MTP-026..030). Under a GLM-backed session, sub-agent
// model calls route to z.ai through the Anthropic-compat shim; z.ai does NOT
// implement Claude's 5-level effort vocabulary. This overlay collapses the Claude
// effort ({low, medium, high, xhigh, max}) that the plan_type tier profile
// produced (§B.6/§B.7) onto GLM-5.3's 3-level reasoning_effort control
// ({low, high, max}), with a coding-max override for the code-producing
// run-phase agent (manager-develop). It is an OVERLAY, not a third plan_type:
// plan_type stays {api, subscription} and the overlay remaps only the EFFORT
// dimension. The MODEL dimension is already carried under GLM by the
// llm.glm.models tier→GLM env mapping (ANTHROPIC_DEFAULT_*), so the overlay
// never touches model.
//
// Sources for the collapse mapping (z.ai official docs, https://docs.z.ai/guides/llm/glm-5.3):
//   - GLM-5.3 reasons ALWAYS. "Disabling reasoning is no longer supported" — a
//     request carrying thinking.type "disabled" FAILS against glm-5.3.
//   - reasoning_effort accepts {low, high, max}; max is the default.
//   - Official migration guidance for callers that previously disabled thinking:
//     "change it to enabled and set reasoning_effort to low before updating the
//     model ID to glm-5.3. Otherwise, the request will fail."
//   - z.ai recommends reasoning_effort: max for coding tasks.
//
// This is why the collapse floor is the WIRE-REAL `low` state (thinking enabled,
// reasoning_effort=low) rather than the pre-5.3 thinking-off state: under
// glm-5.3 the thinking-off state is not merely weaker, it is unreachable.

import (
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// GLM canonical reasoning-state names (REQ-MTP-027). GLM-5.3's reasoning control
// has exactly three reachable levels; these named constants are the collapse
// output domain (no magic literals per CLAUDE.local.md §14).
//
// The state names are the z.ai reasoning_effort tokens themselves — the pre-5.3
// {thinking-off, reasoning-high, reasoning-max} vocabulary described a wire shape
// (a thinking toggle crossed with a 2-value effort) that glm-5.3 no longer has.
// Naming the state after the wire value it carries removes the translation step
// and keeps the settings widget domain and the wire domain identical.
const (
	// GLMStateLow is thinking enabled with reasoning_effort=low (the floor under
	// glm-5.3 — the pre-5.3 thinking-off state is unreachable, see the file doc).
	GLMStateLow = "low"
	// GLMStateHigh is thinking enabled with reasoning_effort=high. Since
	// SPEC-GLM-EFFORT-MAX-001 no Claude effort collapses to this state; the
	// name stays live for the settings-widget domain (GLMReasoningStateNames)
	// and the stored-only defaults in internal/settings/schema_sections.go.
	GLMStateHigh = "high"
	// GLMStateMax is thinking enabled with reasoning_effort=max (z.ai default).
	GLMStateMax = "max"
)

// GLM wire field names + value tokens (REQ-MTP-027, no magic literals per §14).
// The wire field names (reasoning_effort / thinking) are the z.ai request-body
// controls the overlay resolves toward; the value tokens are z.ai's closed sets.
const (
	// GLMReasoningEffortKey is the z.ai request field controlling reasoning depth.
	GLMReasoningEffortKey = "reasoning_effort"
	// GLMThinkingKey is the z.ai request field toggling the reasoning phase.
	GLMThinkingKey = "thinking"
	// GLMReasoningEffortLow is the z.ai lightweight reasoning level. Under
	// glm-5.3 it REPLACES the retired thinking-disabled state — z.ai's own
	// migration guidance is to set reasoning_effort=low where thinking was
	// previously disabled.
	GLMReasoningEffortLow = "low"
	// GLMReasoningEffortHigh is the z.ai economical-loop reasoning level. It
	// stays declared as wire vocabulary — `high` is a legal z.ai wire value and
	// a selectable stored state — but since SPEC-GLM-EFFORT-MAX-001 no Claude
	// effort collapses to it (medium/high raise to max), so by design no
	// in-repo constructor references it.
	GLMReasoningEffortHigh = "high"
	// GLMReasoningEffortMax is the z.ai coding-task / omit-default reasoning level.
	GLMReasoningEffortMax = "max"
	// GLMThinkingEnabledVal is the z.ai thinking toggle value. glm-5.3 accepts
	// only this value — the "disabled" counterpart was retired upstream and is
	// deliberately not declared here, so no caller can reach an unreachable state.
	GLMThinkingEnabledVal = "enabled"
)

// GLMReasoningState is one collapsed z.ai reasoning-control state. It carries the
// canonical state Name (for tests / display) plus the wire representation
// (ThinkingEnabled + ReasoningEffort).
//
// ThinkingEnabled is true in every reachable glm-5.3 state — the field is kept
// (rather than dropped) because it is the shape the wire writers already branch
// on, and because a future backend that reintroduces a reasoning-off mode would
// need it back. Under glm-5.3 it is invariantly true, so ReasoningEffort is
// always populated.
type GLMReasoningState struct {
	// Name is the canonical state name (GLMState* constant).
	Name string
	// ThinkingEnabled is the z.ai thinking toggle. Invariantly true under
	// glm-5.3: "Disabling reasoning is no longer supported."
	ThinkingEnabled bool
	// ReasoningEffort is the z.ai reasoning_effort value (low | high | max).
	ReasoningEffort string
}

// The GLM reasoning states built from the named value constants. Post
// SPEC-GLM-EFFORT-MAX-001 only low and max are constructed: every Claude
// effort above low collapses to max, so no glmReasoningHigh value exists —
// the `high` state survives only as the name/wire constants (widget domain +
// stored defaults; see GLMStateHigh / GLMReasoningEffortHigh).
var (
	glmReasoningLow = GLMReasoningState{Name: GLMStateLow, ThinkingEnabled: true, ReasoningEffort: GLMReasoningEffortLow}
	glmReasoningMax = GLMReasoningState{Name: GLMStateMax, ThinkingEnabled: true, ReasoningEffort: GLMReasoningEffortMax}
)

// CollapseClaudeEffortToGLM maps a Claude effort level (EffortLevel* constants)
// onto the GLM canonical reasoning state (REQ-MTP-027, ceiling raised by
// REQ-GEM-001):
//
//	low    → reasoning_effort low  (the wire-real floor under glm-5.3: thinking
//	                                 cannot be disabled, so low is the floor)
//	medium → reasoning_effort max  (ceiling raise, SPEC-GLM-EFFORT-MAX-001)
//	high   → reasoning_effort max
//	xhigh  → reasoning_effort max
//	max    → reasoning_effort max  (ceiling → ceiling)
//
// Every effort above `low` converges on `max` (the operator directive behind
// card t175): the former medium/high→high convergence was raised so the
// collapse never withholds the ceiling from a mid-tier effort. The `high`
// state remains a legal wire value and a selectable stored state, but no
// Claude effort maps to it anymore.
//
// The mapping is lossy (5→2) by design — z.ai offers fewer levels than Claude —
// and TOTAL over the input: an unrecognized effort maps to the GLM default
// level (max), so a drifted effort string never panics and never under-reasons.
func CollapseClaudeEffortToGLM(effort string) GLMReasoningState {
	switch effort {
	case EffortLevelLow:
		return glmReasoningLow
	case EffortLevelMedium, EffortLevelHigh, EffortLevelXHigh, EffortLevelMax:
		return glmReasoningMax
	default:
		// Totality clause: unrecognized effort → z.ai default level (max).
		return glmReasoningMax
	}
}

// IsGLMFlashModel reports whether the resolved session model is glm-5.3-flash
// (the flash variant). The check is substring-based and case-insensitive so a
// decorated id (e.g. a suffix the caller forwards verbatim) still matches.
func IsGLMFlashModel(model string) bool {
	return strings.Contains(strings.ToLower(model), config.DefaultGLM53Flash)
}

// CollapseClaudeEffortToGLMForModel is the model-aware collapse. Under
// glm-5.3-flash it returns the `max` state for EVERY Claude effort level —
// including `low`: the 3-level GLM reasoning control (low/high/max) does not
// exist on flash, which accepts reasoning_effort: max only, so emitting the
// `low` state would send a wire value the model rejects. For every non-flash
// model it delegates to the glm-5.3-family collapse unchanged
// (CollapseClaudeEffortToGLM), preserving low→low, above-low→max, and the
// unrecognized→max totality clause.
func CollapseClaudeEffortToGLMForModel(model, effort string) GLMReasoningState {
	if IsGLMFlashModel(model) {
		return glmReasoningMax
	}
	return CollapseClaudeEffortToGLM(effort)
}

// GLMReasoningStateNames returns the three canonical z.ai reasoning-state names
// in descending reasoning depth. It exists so a settings surface can offer the
// state set as a closed widget domain without re-declaring the literals — the
// same domain CollapseClaudeEffortToGLM produces.
func GLMReasoningStateNames() []string {
	return []string{GLMStateMax, GLMStateHigh, GLMStateLow}
}

// glmCodingMaxOverrideAgents is the coding-max override set (REQ-MTP-028): the
// single code-producing run-phase agent that z.ai recommends reasoning_effort:
// max for coding tasks. manager-develop runs run-phase implementation at Claude
// effort xhigh. builder-harness was removed by SPEC-GLM-EFFORT-TUNE-001 P1 — its
// artifact-scaffolding role falls under the standard collapse (which, post
// SPEC-GLM-EFFORT-MAX-001, already yields max for every effort above low) per
// the CG-mode cost-reduction goal. No other retained agent is overridden. Named
// constant collection per §14.
var glmCodingMaxOverrideAgents = map[string]bool{
	"manager-develop": true,
}

// IsGLMCodingMaxOverrideAgent reports whether the agent is in the coding-max
// override set (REQ-MTP-028).
func IsGLMCodingMaxOverrideAgent(agentName string) bool {
	return glmCodingMaxOverrideAgents[agentName]
}

// GLMCodingMaxOverrideAgents returns a copy of the coding-max override set as a
// slice (unordered), for tests and display surfaces to assert the set membership
// is exactly {manager-develop} (singleton post SPEC-GLM-EFFORT-TUNE-001 P1).
func GLMCodingMaxOverrideAgents() []string {
	out := make([]string, 0, len(glmCodingMaxOverrideAgents))
	for name := range glmCodingMaxOverrideAgents {
		out = append(out, name)
	}
	return out
}

// ResolveGLMReasoning returns the per-agent GLM reasoning state under a GLM
// backend (REQ-MTP-028): agents in the coding-max override set force the `max` level
// REGARDLESS of their collapse result; every other agent uses the collapse of its
// Claude effort. This is the per-agent overlay logic; it stays defined and
// unit-tested even when the delivery wire carries only a session-level value (the
// delivery-granularity limitation, research.md §D).
func ResolveGLMReasoning(agentName, claudeEffort string) GLMReasoningState {
	if IsGLMCodingMaxOverrideAgent(agentName) {
		// Coding-max override (z.ai coding-task recommendation) — lifts the collapse
		// result to the `max` level for the code-producing run-phase agent (manager-develop).
		return glmReasoningMax
	}
	return CollapseClaudeEffortToGLM(claudeEffort)
}

// ResolveGLMReasoningForModel is the model-aware per-agent resolution: under
// glm-5.3-flash the result is the `max` state regardless of agent or effort
// (flash accepts reasoning_effort: max only — see
// CollapseClaudeEffortToGLMForModel); for every non-flash model it delegates
// to the model-unaware ResolveGLMReasoning (coding-max override + collapse)
// unchanged.
func ResolveGLMReasoningForModel(model, agentName, claudeEffort string) GLMReasoningState {
	if IsGLMFlashModel(model) {
		return glmReasoningMax
	}
	return ResolveGLMReasoning(agentName, claudeEffort)
}

// SessionGLMReasoningState derives the SESSION-GLOBAL GLM reasoning state for the
// Branch-B explicit-write delivery (REQ-MTP-030, raised to the `max` state by
// REQ-GEM-002 — lead-ratified 2026-08-22, superseding REQ-GER-004 of the stalled
// SPEC-GLM-EFFORT-REBALANCE-001 draft; spec.md §1.3). Env vars and the
// settings.local.json env block are session-global (no per-agent reasoning-control
// channel through the z.ai shim), so the per-agent overlay collapses to ONE session
// value. That value is the `max` state, on three grounds:
//
//  1. Under Branch-B delivery this session-global env var is the ONLY reasoning
//     channel every spawn pays — a `high` default while the collapse maps
//     high→max would reach only the prefs-driven path and silently withhold the
//     operator's ceiling from the default session (sub-agents and the
//     empty-effort fallback).
//  2. t127 measured trivial no-tool spawns at ≈ 0 subagent tokens — the "paid by
//     every spawn" premise of the former `high` floor is not where cost lives;
//     reasoning-token cost scales on large calls, which is exactly where the
//     deeper reasoning is wanted (.moai/reports/t175/measurements.md §5).
//  3. `max` is z.ai's own omit-default and its coding-task recommendation — the
//     session default stops fighting the backend's native default.
//
// The per-agent collapse+override logic (ResolveGLMReasoning) remains defined and
// unit-tested — the wire carries this session-level derived value (the documented
// delivery-granularity limitation, research.md §D).
//
// Shim consumption is now MEASURED (t175, .moai/reports/t175/measurements.md §3):
// the z.ai Anthropic-compat shim honors the Anthropic `thinking` parameter
// (thinking blocks returned; depth scales with the budget) and silently IGNORES
// a top-level z.ai-style `reasoning_effort` field. The live session's own
// thinking-block responses are indirect end-to-end evidence that the
// env→reasoning chain is live; no claim is made here about delivered spend.
func SessionGLMReasoningState() GLMReasoningState {
	return glmReasoningMax
}

// SessionGLMReasoningStateForEffort derives the MAIN-SESSION GLM reasoning state
// from the web-set effort preference. It is the prefs-driven counterpart to
// SessionGLMReasoningState() (the session default used for sub-agents and the
// empty-effort fallback): when effort is non-empty it collapses the
// Claude effort onto z.ai's reasoning control, so a web-set effort actually
// reaches z.ai instead of being silently dropped (z.ai does NOT implement
// Claude's 5-level effort). When effort is empty it falls back to
// SessionGLMReasoningState() (the session default), preserving the sub-agent
// / empty-effort behavior and existing tests.
func SessionGLMReasoningStateForEffort(effort string) GLMReasoningState {
	if effort != "" {
		return CollapseClaudeEffortToGLM(effort)
	}
	return SessionGLMReasoningState()
}

// SessionGLMReasoningStateForModel is the model-aware main-session
// derivation: under glm-5.3-flash the session value is the `max` state both
// for a web-set effort (including low — flash accepts reasoning_effort: max
// only) and for the empty-effort fallback; for every non-flash model it
// delegates to SessionGLMReasoningStateForEffort unchanged.
func SessionGLMReasoningStateForModel(model, effort string) GLMReasoningState {
	if IsGLMFlashModel(model) {
		return glmReasoningMax
	}
	return SessionGLMReasoningStateForEffort(effort)
}

// IsGLMBackend reports whether the effective session backend is GLM (REQ-MTP-026).
// It reads the two llm.yaml-persisted intent signals ONLY: team_mode ∈ {cg, glm}
// (the ACTUAL persisted GLM signals — `moai glm` writes team_mode="glm",
// `moai cg` writes team_mode="cg", both via persistTeamMode) OR mode == "glm" (a
// defensive OR for the currently-dormant llm.mode field, which has no non-test
// writer today). team_mode is the real signal; mode is the defensive fallback —
// the predicate does NOT rely on mode alone (that would leave the primary all-GLM
// `moai glm` session, team_mode="glm", undetected — the inert-headline hazard).
//
// This is the CONFIG-level intent signal. It deliberately does NOT re-implement
// the stricter RUNTIME tmux-session + GLM-marker detector in
// internal/tmux/cg_detect.go IsCGMode (which reads team_mode == "cg" AND the tmux
// session env) — it cross-references it. When neither field indicates GLM the
// predicate resolves FALSE and the overlay is an identity no-op.
func IsGLMBackend(cfg config.LLMConfig) bool {
	switch cfg.TeamMode {
	case config.TeamModeGLM, config.TeamModeCG:
		return true
	}
	return cfg.Mode == config.LLMModeGLM
}
