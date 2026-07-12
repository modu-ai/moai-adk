// Package safety — Frozen-rule registry (SPEC-HARNESS-EVOLVE-003 M0).
// REQ-HEV3-017: typed identifier registry of Frozen-rule identifiers.
// Seeded from the safety/frozen_guard.go frozenPrefixes list + A1 permission surfaces.
// The L3 contradiction check (M3) consults this registry to cite the rule identifier
// in the rejection Reason (audit-trail discipline).
package safety

import "path/filepath"

// FrozenRuleCategory classifies the kind of Frozen surface.
type FrozenRuleCategory string

const (
	// FrozenCategoryMoAICore covers MoAI-managed agents, skills, and rules directories.
	FrozenCategoryMoAICore FrozenRuleCategory = "moai-core"

	// FrozenCategoryPermissionSurface covers the A1 permission-surface expansion (REQ-HEV3-023).
	FrozenCategoryPermissionSurface FrozenRuleCategory = "permission-surface"

	// FrozenCategorySelfProtection covers the frozen_guard source files (REQ-HEV3-025).
	FrozenCategorySelfProtection FrozenRuleCategory = "self-protection"
)

// FrozenRule is a single registered Frozen-rule identifier.
// REQ-HEV3-017: carries a stable Name so the L3 rejection Reason can cite it
// (opaque rejections without a rule reference are not acceptable).
type FrozenRule struct {
	// PathPrefix is the path prefix that triggers the Frozen-rule match.
	// Matches the corresponding frozenPrefixes entry.
	PathPrefix string

	// Name is the stable identifier cited in the L3 rejection Reason.
	Name string

	// Category classifies the Frozen-rule surface.
	Category FrozenRuleCategory
}

// FrozenRuleRegistry is the typed registry of Frozen-rule identifiers consulted by L3.
// REQ-HEV3-015/017: seeded from the frozenPrefixes list + A1 permission surfaces.
// Each entry's PathPrefix matches a frozenPrefixes entry (verified by test
// TestFrozenRuleRegistry_EveryPrefixHasRule).
//
// @MX:ANCHOR: [AUTO] FrozenRuleRegistry is the L3 contradiction consult target.
// @MX:REASON: [AUTO] fan_in >= 3: frozen_rules_test.go, contradiction.go (M3), pipeline.go (M3)
var FrozenRuleRegistry = []FrozenRule{
	// MoAI core surfaces (seeded from frozenPrefixes — the original 3 entries).
	{PathPrefix: ".claude/agents/moai/", Name: "frozen-moai-agents", Category: FrozenCategoryMoAICore},
	{PathPrefix: ".claude/skills/moai-", Name: "frozen-moai-skills", Category: FrozenCategoryMoAICore},
	{PathPrefix: ".claude/rules/moai/", Name: "frozen-moai-rules", Category: FrozenCategoryMoAICore},

	// A1 permission surfaces (REQ-HEV3-023).
	// The hooks axis is covered by the settings.json prefix match — no separate entry.
	{PathPrefix: ".claude/settings.json", Name: "frozen-permission-settings", Category: FrozenCategoryPermissionSurface},
	{PathPrefix: ".claude/settings.local.json", Name: "frozen-permission-settings-local", Category: FrozenCategoryPermissionSurface},

	// Self-protection (REQ-HEV3-025) — the guard cannot permit its own modification.
	{PathPrefix: "internal/harness/safety/frozen_guard.go", Name: "frozen-guard-safety", Category: FrozenCategorySelfProtection},
	{PathPrefix: "internal/harness/frozen_guard.go", Name: "frozen-guard-meta", Category: FrozenCategorySelfProtection},
}

// FindFrozenRule returns the FrozenRule matching the given path, or nil if none match.
// REQ-HEV3-017: used by L3 to cite the rule identifier in the rejection Reason.
// Path normalization mirrors IsFrozen: ToSlash + Clean.
func FindFrozenRule(path string) *FrozenRule {
	if path == "" {
		return nil
	}

	norm := filepath.ToSlash(path)
	norm = filepath.ToSlash(filepath.Clean(norm))

	for i := range FrozenRuleRegistry {
		r := &FrozenRuleRegistry[i]
		if hasPrefix(norm, r.PathPrefix) {
			return r
		}
	}
	return nil
}
