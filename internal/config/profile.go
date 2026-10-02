package config

// profile.go — the llm.profile closed-set enum, its effective-default
// resolution, and per-agent-override validation. Re-ported under
// SPEC-WEB-AGENTFM-RESTORE-001 M1 (operator card t1411) from the copy
// SPEC-AGENT-MODEL-INHERIT-001 M5 deleted (3fa8bd2ab): the console surface
// that reads and writes these keys is restored, so their reader machinery
// returns with it. The `llm.profiles` config-mirror field and the legacy
// `llm.performance_tier` read-time alias leg are deliberately NOT re-ported
// (plan §D.4-3 — the alias key stays retired and stripped; the Go default
// matrix in template is the SSOT), so the resolution chain is
// profile → default medium only.

import "strings"

// Profile closed-set values. Named constants per CLAUDE.local.md §14 (no
// magic strings for the enum). The canonical set is {high, medium, low}; the
// superseded top-column name "max" stays readable as a normalize-time alias
// (LegacyProfileMax) so pre-restore configs resolve, but nothing in this
// repository writes it back — the restored console selector persists its own
// wire vocabulary verbatim (template.ValidPerformanceTiers, REQ-AFR-003).
const (
	// ProfileHigh is the highest-quality profile column.
	ProfileHigh = "high"
	// ProfileMedium is the balanced default profile column.
	ProfileMedium = "medium"
	// ProfileLow is the economical profile column.
	ProfileLow = "low"
	// LegacyProfileMax is the superseded name of the top column. It is
	// accepted as a read-time alias for ProfileHigh and is never written back.
	LegacyProfileMax = "max"
	// DefaultProfile is the effective profile when llm.profile is absent or
	// empty: agents resolve the session model/effort (REQ-AFR-002 inheritance
	// default), and the console matrix resolves its medium column.
	DefaultProfile = ProfileMedium
)

// validProfiles is the single source of truth for the llm.profile closed set.
var validProfiles = map[string]bool{
	ProfileHigh:   true,
	ProfileMedium: true,
	ProfileLow:    true,
}

// NormalizeProfile maps a persisted profile value onto the canonical closed
// set, translating the superseded top-column name (max -> high). Any other
// value is returned verbatim so callers can still reject it as out-of-set.
//
// @MX:ANCHOR: [AUTO] NormalizeProfile — the max->high read-time alias
// @MX:REASON: [AUTO] fan_in >= 2 (EffectiveProfile + validateProfile + template tier helpers); the sole compatibility bridge for pre-restore configs
func NormalizeProfile(name string) string {
	if name == LegacyProfileMax {
		return ProfileHigh
	}
	return name
}

// IsValidProfile reports whether name is one of the closed-set profiles
// (high, medium, low) or the accepted legacy alias (max). The empty string is
// NOT a member here; callers that treat empty as "keep the effective default"
// use EffectiveProfile instead.
func IsValidProfile(name string) bool {
	return validProfiles[NormalizeProfile(name)]
}

// ValidProfiles returns the closed-set profile names for UI option lists.
// Order is stable (high, medium, low) for deterministic rendering. The legacy
// max alias is intentionally absent — it is readable but never offered.
func ValidProfiles() []string {
	return []string{ProfileHigh, ProfileMedium, ProfileLow}
}

// EffectiveProfile resolves the active profile:
//  1. a non-empty llm.profile value passes through NormalizeProfile
//     (max -> high; high/medium/low verbatim);
//  2. else the default profile ("medium").
//
// The legacy llm.performance_tier alias leg of the pre-deletion resolver is
// deliberately absent: that key stays retired (plan §D.4-3), no reader
// survives for it, and an llm.yaml still carrying the key has it stripped by
// the update pipeline. An out-of-set value is returned verbatim so callers
// can reject it (validateProfile names it in the atomic-reject set).
func (l LLMConfig) EffectiveProfile() string {
	if p := strings.TrimSpace(l.Profile); p != "" {
		return NormalizeProfile(p)
	}
	return DefaultProfile
}

// validOverrideModels is the closed set of model aliases accepted in an
// llm.agent_overrides entry (REQ-AFR-006). It matches the aliases the console
// selector offers, plus the inherit sentinel, plus "haiku" as an explicit
// user opt-in. haiku is admitted HERE ONLY: no profile column ever resolves
// an agent to haiku on its own, and no other closed set in the tree admits it.
//
// Caveat worth knowing when picking haiku: Claude's reasoning-effort levels do
// not apply to Haiku, so the effort paired with a haiku override is inert
// (the console disables the effort select and backfills the resolved value).
var validOverrideModels = map[string]bool{
	"opus":    true,
	"sonnet":  true,
	"fable":   true,
	"haiku":   true,
	"inherit": true,
}

// validOverrideEfforts is the closed set of effort levels accepted in an
// llm.agent_overrides entry (the canonical 5-tier effort vocabulary,
// REQ-AFR-006).
var validOverrideEfforts = map[string]bool{
	"low":    true,
	"medium": true,
	"high":   true,
	"xhigh":  true,
	"max":    true,
}

// retainedAgentNames is the closed set of canonical agent names an
// llm.agent_overrides entry may key on (REQ-AFR-006, REQ-MPM-007 lineage).
// It is the CURRENT 12-agent roster (.claude/agents/moai/, measured
// 2026-10-02) — manager-todo in, mission-governor out — plus the Anthropic
// built-in Explore, which keeps its entry so pre-restore overrides for it
// stay valid and resolvable. A newly retained agent must be registered here
// before an operator can target it with an override.
var retainedAgentNames = map[string]bool{
	"manager-spec":    true,
	"plan-auditor":    true,
	"sync-auditor":    true,
	"manager-develop": true,
	"super-advisor":   true,
	"manager-design":  true,
	"manager-lead":    true,
	"builder-harness": true,
	"e2e-tester":      true,
	"manager-docs":    true,
	"manager-git":     true,
	"manager-todo":    true,
	"Explore":         true,
}

// validateProfile checks the llm.profile value against the closed set. An
// empty value is the effective default (medium) and is not an error. A
// non-empty out-of-set value returns a ValidationError naming the offending
// value AND the closed set, joining the console's atomic-reject flow
// (REQ-AFR-006/007).
func validateProfile(cfg *Config) []ValidationError {
	p := strings.TrimSpace(cfg.LLM.Profile)
	if p == "" || IsValidProfile(p) {
		return nil
	}
	return []ValidationError{{
		Field:   "llm.profile",
		Message: "profile " + p + " is not in the closed set {high, medium, low}",
		Value:   p,
		Wrapped: ErrInvalidConfig,
	}}
}

// validateAgentOverrides checks each llm.agent_overrides entry (REQ-AFR-006).
// An entry naming an agent outside the retained catalog, or carrying a model
// or effort value outside the valid enums, returns a ValidationError naming
// the offending agent and field. An empty map is valid.
func validateAgentOverrides(cfg *Config) []ValidationError {
	var errs []ValidationError
	for agent, me := range cfg.LLM.AgentOverrides {
		if !retainedAgentNames[agent] {
			errs = append(errs, ValidationError{
				Field:   "llm.agent_overrides." + agent,
				Message: "agent " + agent + " is not in the retained agent catalog",
				Value:   agent,
				Wrapped: ErrInvalidConfig,
			})
			continue
		}
		if m := strings.TrimSpace(me.Model); m != "" && !validOverrideModels[m] {
			errs = append(errs, ValidationError{
				Field:   "llm.agent_overrides." + agent + ".model",
				Message: "model " + m + " is not in the valid set {opus, sonnet, fable, haiku, inherit}",
				Value:   m,
				Wrapped: ErrInvalidConfig,
			})
		}
		if e := strings.TrimSpace(me.Effort); e != "" && !validOverrideEfforts[e] {
			errs = append(errs, ValidationError{
				Field:   "llm.agent_overrides." + agent + ".effort",
				Message: "effort " + e + " is not in the valid set {low, medium, high, xhigh, max}",
				Value:   e,
				Wrapped: ErrInvalidConfig,
			})
		}
	}
	return errs
}
