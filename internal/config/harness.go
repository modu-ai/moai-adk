package config

// harness.go — SPEC-INIT-HARNESS-001: the closed set for the llm.harness key
// (REQ-IH-001). The set mirrors the agent_wiring wizard axis and the --llm
// flag exactly; claude is the default (DefaultHarness).

import (
	"path/filepath"
)

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

// ReadHarness returns the resolved llm.harness value for the project rooted
// at projectRoot (SPEC-INIT-HARNESS-001 REQ-IH-010/011 read side): the file
// value when it is a member of the closed set, DefaultHarness otherwise. An
// absent file, an absent key, or an out-of-set value all read as claude — the
// documented pre-SPEC fallback that keeps existing projects on today's
// update behavior. Parse failures degrade to the fallback the same way the
// Loader does (slog-warn path there; silent here, callers re-derive nothing).
func ReadHarness(projectRoot string) string {
	return ReadHarnessFrom(filepath.Join(projectRoot, ".moai", "config", "sections"))
}

// ReadHarnessFrom is ReadHarness against an explicit sections directory — the
// form update's restore step uses to read the PRE-UPDATE backup, whose
// directory layout is a sections/ tree, not the live one.
func ReadHarnessFrom(sectionsDir string) string {
	wrapper := &llmFileWrapper{}
	if _, err := loadYAMLFile(sectionsDir, "llm.yaml", wrapper); err != nil {
		return DefaultHarness
	}
	if IsValidAgentHarness(wrapper.LLM.Harness) {
		return wrapper.LLM.Harness
	}
	return DefaultHarness
}
