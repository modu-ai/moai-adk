package config

// harness.go — SPEC-INIT-HARNESS-001: the closed set for the llm.harness key
// (REQ-IH-001). The set mirrors the agent_wiring wizard axis and the --llm
// flag exactly; claude is the default (DefaultHarness).

// validAgentHarnesses is the closed set of llm.harness values. claude = the
// full .claude/ deployment (today's behavior), both = claude plus the Codex
// wiring, codex = Codex-only deployment (AGENTS.md + Codex surfaces only).
var validAgentHarnesses = map[string]struct{}{
	"claude": {},
	"codex":  {},
	"both":   {},
}

// IsValidAgentHarness reports whether value is one of the closed-set harness
// values (claude, codex, both). The empty string is NOT a member: callers that
// mean "no selection recorded" handle that case themselves (it resolves to
// claude as the documented pre-SPEC fallback).
func IsValidAgentHarness(value string) bool {
	_, ok := validAgentHarnesses[value]
	return ok
}

// ValidAgentHarnesses returns the closed-set harness names for UI option
// lists. Order is stable (claude, codex, both) and matches the wizard's
// agent_wiring option order.
func ValidAgentHarnesses() []string {
	return []string{"claude", "codex", "both"}
}
