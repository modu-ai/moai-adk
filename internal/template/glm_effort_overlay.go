package template

// glm_effort_overlay.go — SPEC-MODEL-TIER-PLANTYPE-001 M5: the GLM backend
// effort overlay (§B.8, REQ-MTP-026..030). Under a GLM-backed session, sub-agent
// model calls route to z.ai through the Anthropic-compat shim; z.ai does NOT
// implement Claude's 5-level effort vocabulary. This overlay collapses the Claude
// effort ({low, medium, high, xhigh, max}) that the plan_type tier profile
// produced (§B.6/§B.7) onto z.ai's 3-state reasoning control (thinking-off +
// reasoning_effort ∈ {high, max}), with a coding-max override for the
// code-producing agents. It is an OVERLAY, not a third plan_type: plan_type stays
// {api, subscription} and the overlay remaps only the EFFORT dimension. The MODEL
// dimension is already carried under GLM by the llm.glm.models tier→GLM env
// mapping (ANTHROPIC_DEFAULT_*), so the overlay never touches model.
//
// Sources for the collapse mapping (research.md §B, z.ai official docs):
//   - thinking is a binary enabled/disabled toggle
//   - reasoning_effort accepts ONLY {high, max}; max is the omit-default; high is
//     explicit; reasoning_effort is moot when thinking is disabled
//   - z.ai recommends reasoning_effort: max for coding tasks

import "github.com/modu-ai/moai-adk/internal/config"

// GLM canonical reasoning-state names (REQ-MTP-027). z.ai's reasoning control has
// exactly three reachable states; these named constants are the collapse output
// domain (no magic literals per CLAUDE.local.md §14).
const (
	// GLMStateThinkingOff is the "thinking disabled" state (reasoning phase off).
	GLMStateThinkingOff = "thinking-off"
	// GLMStateReasoningHigh is thinking enabled with reasoning_effort=high.
	GLMStateReasoningHigh = "reasoning-high"
	// GLMStateReasoningMax is thinking enabled with reasoning_effort=max.
	GLMStateReasoningMax = "reasoning-max"
)

// GLM wire field names + value tokens (REQ-MTP-027, no magic literals per §14).
// The wire field names (reasoning_effort / thinking) are the z.ai request-body
// controls the overlay resolves toward; the value tokens are z.ai's closed sets.
const (
	// GLMReasoningEffortKey is the z.ai request field controlling reasoning depth.
	GLMReasoningEffortKey = "reasoning_effort"
	// GLMThinkingKey is the z.ai request field toggling the reasoning phase.
	GLMThinkingKey = "thinking"
	// GLMReasoningEffortHigh is the z.ai economical-loop reasoning level.
	GLMReasoningEffortHigh = "high"
	// GLMReasoningEffortMax is the z.ai coding-task / omit-default reasoning level.
	GLMReasoningEffortMax = "max"
	// GLMThinkingEnabledVal / GLMThinkingDisabledVal are the z.ai thinking toggle
	// values.
	GLMThinkingEnabledVal  = "enabled"
	GLMThinkingDisabledVal = "disabled"
)

// GLMReasoningState is one collapsed z.ai reasoning-control state. It carries the
// canonical state Name (for tests / display) plus the wire representation
// (ThinkingEnabled + ReasoningEffort). ReasoningEffort is "" when thinking is
// disabled (z.ai fact: reasoning_effort is moot without thinking).
type GLMReasoningState struct {
	// Name is the canonical state name (GLMState* constant).
	Name string
	// ThinkingEnabled is the z.ai thinking toggle (false = thinking disabled).
	ThinkingEnabled bool
	// ReasoningEffort is the z.ai reasoning_effort value (high | max); empty when
	// thinking is disabled.
	ReasoningEffort string
}

// The three reachable GLM reasoning states, built from the named value constants.
var (
	glmReasoningThinkingOff = GLMReasoningState{Name: GLMStateThinkingOff, ThinkingEnabled: false, ReasoningEffort: ""}
	glmReasoningHigh        = GLMReasoningState{Name: GLMStateReasoningHigh, ThinkingEnabled: true, ReasoningEffort: GLMReasoningEffortHigh}
	glmReasoningMax         = GLMReasoningState{Name: GLMStateReasoningMax, ThinkingEnabled: true, ReasoningEffort: GLMReasoningEffortMax}
)

// CollapseClaudeEffortToGLM maps a Claude effort level (EffortLevel* constants)
// onto the GLM canonical reasoning state (REQ-MTP-027):
//
//	low   → thinking disabled     (fastest, least thorough → skip reasoning)
//	medium → reasoning_effort high (balanced → economical-loop level)
//	high  → reasoning_effort high (deep but not maximal; z.ai has no level between
//	                               high and max, so medium/high converge)
//	xhigh → reasoning_effort max  (extended reasoning → ceiling)
//	max   → reasoning_effort max  (ceiling → ceiling)
//
// The mapping is lossy (5→3) by design — z.ai offers fewer levels than Claude —
// and TOTAL over the input: an unrecognized effort maps to the GLM omit-default
// state (reasoning-max), so a drifted effort string never panics and never
// under-reasons.
func CollapseClaudeEffortToGLM(effort string) GLMReasoningState {
	switch effort {
	case EffortLevelLow:
		return glmReasoningThinkingOff
	case EffortLevelMedium, EffortLevelHigh:
		return glmReasoningHigh
	case EffortLevelXHigh, EffortLevelMax:
		return glmReasoningMax
	default:
		// Totality clause: unrecognized effort → z.ai omit-default (reasoning-max).
		return glmReasoningMax
	}
}

// glmCodingMaxOverrideAgents is the coding-max override set (REQ-MTP-028): the two
// code-producing retained agents that z.ai recommends reasoning_effort: max for
// coding tasks. manager-develop runs run-phase implementation; builder-harness
// generates dynamic specialists. No other retained agent is a code-producer, so
// no other agent is overridden. Named constant collection per §14.
var glmCodingMaxOverrideAgents = map[string]bool{
	"manager-develop": true,
	"builder-harness": true,
}

// IsGLMCodingMaxOverrideAgent reports whether the agent is in the coding-max
// override set (REQ-MTP-028).
func IsGLMCodingMaxOverrideAgent(agentName string) bool {
	return glmCodingMaxOverrideAgents[agentName]
}

// GLMCodingMaxOverrideAgents returns a copy of the coding-max override set as a
// slice (unordered), for tests and display surfaces to assert the set membership
// is exactly {manager-develop, builder-harness}.
func GLMCodingMaxOverrideAgents() []string {
	out := make([]string, 0, len(glmCodingMaxOverrideAgents))
	for name := range glmCodingMaxOverrideAgents {
		out = append(out, name)
	}
	return out
}

// ResolveGLMReasoning returns the per-agent GLM reasoning state under a GLM
// backend (REQ-MTP-028): agents in the coding-max override set force reasoning-max
// REGARDLESS of their collapse result; every other agent uses the collapse of its
// Claude effort. This is the per-agent overlay logic; it stays defined and
// unit-tested even when the delivery wire carries only a session-level value (the
// delivery-granularity limitation, research.md §D).
func ResolveGLMReasoning(agentName, claudeEffort string) GLMReasoningState {
	if IsGLMCodingMaxOverrideAgent(agentName) {
		// Coding-max override (z.ai coding-task recommendation) — lifts the collapse
		// result to reasoning-max for the code-producing agents.
		return glmReasoningMax
	}
	return CollapseClaudeEffortToGLM(claudeEffort)
}

// ApplyGLMEffortOverlay applies the GLM effort overlay to a plan_type profile's
// {model, effort} pair (REQ-MTP-029). Under a GLM backend it changes ONLY the
// effort representation — the returned Model is byte-identical to the input
// (the overlay never rewrites model:, whose GLM mapping is owned by
// llm.glm.models), and the returned Effort is the collapsed GLM canonical state
// name. Under a non-GLM backend (glmBackend == false) the overlay is an identity
// no-op: the pair is returned unchanged (the plan_type profile output).
func ApplyGLMEffortOverlay(entry TierProfileEntry, agentName string, glmBackend bool) TierProfileEntry {
	if !glmBackend {
		return entry // identity no-op under a Claude backend (REQ-MTP-029)
	}
	state := ResolveGLMReasoning(agentName, entry.Effort)
	return TierProfileEntry{
		Model:  entry.Model, // byte-identical — the overlay never touches model
		Effort: state.Name,  // effort representation remapped to the GLM canonical state
	}
}

// SessionGLMReasoningState derives the SESSION-GLOBAL GLM reasoning state for the
// Branch-B explicit-write delivery (REQ-MTP-030). Env vars and the
// settings.local.json env block are session-global (no per-agent reasoning-control
// channel through the z.ai shim), so the per-agent overlay collapses to one
// session value: the coding-max override level (reasoning-max), because a
// code-producing agent (manager-develop / builder-harness, in the override set)
// is the representative active spawn under a GLM-backed MoAI session. The
// per-agent collapse+override logic (ResolveGLMReasoning) remains defined and
// unit-tested; the wire carries this session-level derived value (the documented
// delivery-granularity limitation, research.md §D).
func SessionGLMReasoningState() GLMReasoningState {
	return glmReasoningMax
}

// SessionGLMReasoningStateForEffort derives the MAIN-SESSION GLM reasoning state
// from the web-set effort preference. It is the prefs-driven counterpart to
// SessionGLMReasoningState() (the coding-max session default used for sub-agents
// and the empty-effort fallback): when effort is non-empty it collapses the
// Claude effort onto z.ai's reasoning control, so a web-set effort actually
// reaches z.ai instead of being silently dropped (z.ai does NOT implement
// Claude's 5-level effort). When effort is empty it falls back to
// SessionGLMReasoningState() (the coding-max default), preserving the sub-agent
// / empty-effort behavior and existing tests.
func SessionGLMReasoningStateForEffort(effort string) GLMReasoningState {
	if effort != "" {
		return CollapseClaudeEffortToGLM(effort)
	}
	return SessionGLMReasoningState()
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
