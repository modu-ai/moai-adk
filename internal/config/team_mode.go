package config

// team_mode.go — SPEC-MODEL-TIER-PLANTYPE-001 M5: named constants for the
// llm.team_mode and llm.mode field VALUES that signal a GLM backend. These are
// the values persistTeamMode (internal/cli/glm.go) actually writes; the GLM
// effort-overlay backend-detection predicate (template.IsGLMBackend, REQ-MTP-026)
// reads them. Named constants per CLAUDE.local.md §14 (no magic strings for the
// closed set), mirroring the plan_type.go / model_routing.go constant pattern.

const (
	// TeamModeGLM is the llm.team_mode value written by `moai glm` (all-GLM
	// session) — the PRIMARY GLM backend signal. persistTeamMode(root, "glm").
	TeamModeGLM = "glm"
	// TeamModeGPT is the llm.team_mode value that signals a gateway (gpt)
	// backend (template.IsGatewayBackend). No launcher persists it: a gateway
	// launch owns its child environment and never mutates llm.yaml (the
	// TestUnifiedGatewayLaunchSkipsLegacyModeMutation invariant), so at runtime
	// the value arrives through LLMConfig.WithLaunchProvider folding the
	// launcher-owned MOAI_LAUNCH_PROVIDER env value. It is also parse-accepted
	// from llm.yaml so an operator (or a test) can pin the backend explicitly.
	TeamModeGPT = "gpt"
	// LegacyTeamModeCG identifies historical data requiring explicit migration.
	// It is not an executable backend or a provider selection.
	LegacyTeamModeCG = "cg"
	// TeamModeClaude is a legacy non-GLM team_mode value (Claude-only). Retained
	// for backward-compat parsing; not a GLM backend signal.
	TeamModeClaude = "claude"
	// TeamModeHybrid is a legacy non-GLM team_mode value. Retained for
	// backward-compat parsing; not a GLM backend signal.
	TeamModeHybrid = "hybrid"
)

// LLMModeGLM is the llm.mode value that would signal a GLM backend. The llm.mode
// field is currently DORMANT — no non-test writer sets it (`moai glm` writes
// team_mode="glm", NOT mode="glm"). Retained only as a defensive OR in the
// backend-detection predicate (template.IsGLMBackend) for a future launch path
// that populates mode.
const LLMModeGLM = "glm"

// LLMModeGPT is the dormant llm.mode counterpart of TeamModeGPT, kept as the
// same defensive OR in template.IsGatewayBackend. No writer sets it.
const LLMModeGPT = "gpt"

// WithLaunchProvider folds the launcher-owned initial-provider value
// (MOAI_LAUNCH_PROVIDER, set by prepareGatewayLaunch on the child env) into
// the config-level backend signal. A gateway launch persists nothing into
// llm.yaml, so a config-reading surface (the web console, `moai model
// profile`) running inside a gpt session would otherwise read a Claude
// backend and offer per-agent model selection the launcher makes meaningless.
//
// Only provider "gpt" folds, and only when llm.yaml carries no team_mode of its
// own — an explicit team_mode is the operator's intent and wins. "glm" is not
// folded here because `moai glm` already persists team_mode="glm"; "claude"
// and any other value leave the config untouched. Value semantics: the
// receiver is copied, never mutated.
func (c LLMConfig) WithLaunchProvider(provider string) LLMConfig {
	if provider == TeamModeGPT && c.TeamMode == "" {
		c.TeamMode = TeamModeGPT
	}
	return c
}
