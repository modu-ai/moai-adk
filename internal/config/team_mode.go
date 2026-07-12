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
	// TeamModeCG is the llm.team_mode value written by `moai cg` (Claude leader +
	// GLM teammates). persistTeamMode(root, "cg").
	TeamModeCG = "cg"
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
