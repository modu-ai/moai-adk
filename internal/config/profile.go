package config

// profile.go — SPEC-MODEL-PROFILE-MATRIX-001 M1: the llm.profile closed-set
// enum, its effective-default resolution (with the legacy performance_tier
// read-time alias), the {model, effort} pair type, and per-agent-override
// validation. The profile selects the active per-agent-group model+effort
// column {max, medium, low}, replacing the retired plan_type × tier axis.

import "strings"

// Profile closed-set values (REQ-MPM-001). Named constants per CLAUDE.local.md
// §14 (no magic strings for the enum), mirroring the validPlanTypes pattern.
//
// The top column is named "high" (formerly "max"). The rename unifies three
// vocabularies that previously disagreed: llm.profile, the legacy
// llm.performance_tier, and template.ModelPolicy are now all {high, medium,
// low}, so the former performance_tier high->max projection is an identity and
// no migration pass is required. Configs written before the rename carry
// profile: max; LegacyProfileMax keeps them resolving via NormalizeProfile.
const (
	// ProfileHigh is the highest-quality profile column.
	ProfileHigh = "high"
	// ProfileMedium is the balanced default profile column (DECISION-002).
	ProfileMedium = "medium"
	// ProfileLow is the economical profile column.
	ProfileLow = "low"
	// LegacyProfileMax is the superseded name of the top column. It is accepted
	// as a read-time alias for ProfileHigh and is never written back.
	LegacyProfileMax = "max"
	// DefaultProfile is the effective profile when llm.profile is absent/empty
	// and no legacy performance_tier alias applies (REQ-MPM-002, DECISION-002).
	DefaultProfile = ProfileMedium
)

// validProfiles is the single source of truth for the llm.profile closed set.
var validProfiles = map[string]bool{
	ProfileHigh:   true,
	ProfileMedium: true,
	ProfileLow:    true,
}

// NormalizeProfile maps a persisted profile value onto the canonical closed set,
// translating the superseded top-column name (max -> high). Any other value is
// returned verbatim so callers can still reject it as out-of-set.
//
// @MX:ANCHOR: [AUTO] NormalizeProfile — the max->high read-time alias
// @MX:REASON: [AUTO] fan_in >= 3 (EffectiveProfile + validateProfile + template NormalizeToTier); the sole compatibility bridge for pre-rename configs
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

// ModelEffort carries a {model, effort} assignment for one agent or one
// profile group cell. Model is a Claude Code short alias (opus/sonnet/fable/
// inherit); Effort is a reasoning effort level (low/medium/high/xhigh/max).
type ModelEffort struct {
	Model  string `yaml:"model"`
	Effort string `yaml:"effort"`
}

// EffectiveProfile resolves the active profile (REQ-MPM-002):
//  1. a non-empty llm.profile value passes through NormalizeProfile
//     (max -> high; high/medium/low verbatim);
//  2. else the legacy llm.performance_tier is used the same way — since the
//     top column is now named "high", the former high->max projection is an
//     identity and a pre-rename "max" resolves to "high";
//  3. else the default profile ("medium", DECISION-002).
//
// The separate init-selection constant DefaultModelPolicy = "medium"
// (template package) is NOT consulted here; since
// SPEC-CLI-WIZARD-RESTRUCTURE-001 it happens to agree with this function's own
// default. With the top column renamed from "max" to "high", the profile and
// performance_tier axes share one vocabulary, so no separate projection
// survives — only the max -> high read alias remains.
func (l LLMConfig) EffectiveProfile() string {
	if p := strings.TrimSpace(l.Profile); p != "" {
		return NormalizeProfile(p)
	}
	if pt := strings.TrimSpace(l.PerformanceTier); pt != "" {
		return NormalizeProfile(pt)
	}
	return DefaultProfile
}

// validOverrideModels is the closed set of model aliases accepted in an
// llm.agent_overrides entry. It matches the aliases used by Matrix A, plus the
// inherit sentinel, plus "haiku" as an explicit user opt-in.
//
// haiku is admitted HERE ONLY. The relaxation is deliberately scoped to the
// per-agent override surface — a hand-picked, per-agent economy choice — and
// changes nothing else:
//   - defaultProfileMatrix (template package) stays haiku-free: no profile
//     column ever resolves an agent to haiku on its own.
//   - validRoutingModels (model_routing.go) stays haiku-free: the
//     model_routing_profiles closed set is a separate surface.
//
// The tier layer already contemplates haiku for mechanical agents (see
// template.ModelPolicyMedium/Low), so admitting it as an explicit override
// removes an asymmetry rather than introducing one.
//
// Caveat worth knowing when picking haiku: Claude's reasoning-effort levels do
// not apply to Haiku, so the effort paired with a haiku override is inert.
var validOverrideModels = map[string]bool{
	"opus":    true,
	"sonnet":  true,
	"fable":   true,
	"haiku":   true,
	"inherit": true,
}

// validOverrideEfforts is the closed set of effort levels accepted in an
// llm.agent_overrides entry (the canonical 5-tier effort vocabulary).
var validOverrideEfforts = map[string]bool{
	"low":    true,
	"medium": true,
	"high":   true,
	"xhigh":  true,
	"max":    true,
}

// retainedAgentNames is the closed set of canonical retained-agent names an
// llm.agent_overrides entry may key on (REQ-MPM-007). The 10 MoAI-custom
// agents plus the Anthropic built-in Explore.
var retainedAgentNames = map[string]bool{
	"manager-spec":    true,
	"plan-auditor":    true,
	"sync-auditor":    true,
	"manager-develop": true,
	"super-advisor":   true,
	"manager-design":  true,
	"builder-harness": true,
	"e2e-tester":      true,
	"manager-docs":    true,
	"manager-git":     true,
	"Explore":         true,
}

// validateProfile checks the llm.profile value against the closed set
// (REQ-MPM-008). An empty value is the effective default (medium) and is not an
// error. A non-empty out-of-set value returns a ValidationError naming the
// offending value AND the closed set {max, medium, low}.
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

// validateAgentOverrides checks each llm.agent_overrides entry (REQ-MPM-007).
// An entry naming an agent outside the retained catalog, or carrying a model or
// effort value outside the valid enums, returns a ValidationError naming the
// offending agent and field. An empty map is valid.
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
