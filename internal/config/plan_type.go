package config

// plan_type.go — SPEC-MODEL-TIER-PLANTYPE-001 M1: the llm.plan_type closed-set
// enum, its effective-default resolution, and its validation. plan_type selects
// which model tier profile (API-metered vs subscription billing) applies to the
// shipped agents; it describes the Claude billing context, orthogonal to the
// model_routing cost axis.

import "strings"

// Plan type closed-set values (REQ-MTP-001). These are named constants rather
// than inline literals per CLAUDE.local.md §14 (no magic strings for the enum),
// mirroring the validRoutingPerfTiers pattern in model_routing.go.
const (
	// PlanTypeAPI selects the API-metered billing tier profile.
	PlanTypeAPI = "api"
	// PlanTypeSubscription selects the subscription billing tier profile.
	PlanTypeSubscription = "subscription"
	// DefaultPlanType is the effective plan type when llm.plan_type is absent or
	// empty (REQ-MTP-002 backward compatibility: existing projects resolve to the
	// subscription branch without any config edit).
	DefaultPlanType = PlanTypeSubscription
)

// validPlanTypes is the single source of truth for the llm.plan_type closed set,
// mirroring how validMergeMethods is the SSOT for the merge_method enum.
var validPlanTypes = map[string]bool{
	PlanTypeAPI:          true,
	PlanTypeSubscription: true,
}

// IsValidPlanType reports whether name is one of the closed-set plan types
// (api, subscription). The empty string is NOT a member here; callers that treat
// empty as "keep the effective default" must guard for empty (or use
// EffectivePlanType, which resolves empty to DefaultPlanType).
func IsValidPlanType(name string) bool {
	return validPlanTypes[name]
}

// ValidPlanTypes returns the closed-set plan type names as a slice, for
// populating UI option lists (web <select>, CLI flag help). Order is not
// guaranteed (sourced from a map) — callers that need stable ordering must sort.
func ValidPlanTypes() []string {
	names := make([]string, 0, len(validPlanTypes))
	for name := range validPlanTypes {
		names = append(names, name)
	}
	return names
}

// EffectivePlanType resolves the effective plan type for the LLM section
// (REQ-MTP-002). An absent, empty, or whitespace-only plan_type resolves to
// DefaultPlanType (subscription); any explicit value passes through verbatim.
func (l LLMConfig) EffectivePlanType() string {
	if strings.TrimSpace(l.PlanType) == "" {
		return DefaultPlanType
	}
	return l.PlanType
}

// validatePlanType checks the llm.plan_type value against the closed set
// (REQ-MTP-003). An empty value is treated as the effective default
// (subscription) and is not an error — identical to how
// validateGitConventionConfig guards gc.Convention != "". A non-empty out-of-set
// value returns a ValidationError whose message names the offending value AND the
// closed set {api, subscription}.
func validatePlanType(cfg *Config) []ValidationError {
	pt := cfg.LLM.PlanType
	if pt == "" || validPlanTypes[pt] {
		return nil
	}
	return []ValidationError{{
		Field:   "llm.plan_type",
		Message: "plan_type " + pt + " is not in the closed set {api, subscription}",
		Value:   pt,
		Wrapped: ErrInvalidConfig,
	}}
}
