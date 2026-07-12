// Package safety — Frozen-rule registry tests (SPEC-HARNESS-EVOLVE-003 M0).
// REQ-HEV3-017: typed identifier registry of Frozen-rule identifiers consulted by L3.
package safety

import "testing"

// TestFrozenRuleRegistry_Seeded verifies the Frozen-rule registry is non-empty and seeded.
// H-3: seeded from the frozenPrefixes list + A1 permission surfaces.
func TestFrozenRuleRegistry_Seeded(t *testing.T) {
	t.Parallel()

	if len(FrozenRuleRegistry) == 0 {
		t.Fatal("FrozenRuleRegistry is empty — must be seeded from frozenPrefixes + A1 surfaces")
	}
}

// TestFrozenRuleRegistry_EveryPrefixHasRule verifies every frozenPrefixes entry has a
// corresponding FrozenRuleRegistry entry (the registry stays in sync with the L1 block-list).
func TestFrozenRuleRegistry_EveryPrefixHasRule(t *testing.T) {
	t.Parallel()

	rulePrefixes := make(map[string]bool, len(FrozenRuleRegistry))
	for _, r := range FrozenRuleRegistry {
		rulePrefixes[r.PathPrefix] = true
	}
	for _, prefix := range frozenPrefixes {
		if !rulePrefixes[prefix] {
			t.Errorf("frozenPrefixes entry %q has no corresponding FrozenRuleRegistry entry", prefix)
		}
	}
}

// TestFrozenRuleRegistry_StableNames verifies every rule carries a non-empty stable Name
// and a non-empty Category (REQ-HEV3-017: the L3 rejection Reason cites the rule identifier).
func TestFrozenRuleRegistry_StableNames(t *testing.T) {
	t.Parallel()

	for _, r := range FrozenRuleRegistry {
		if r.Name == "" {
			t.Errorf("FrozenRule PathPrefix=%q has empty Name — REQ-HEV3-017 requires a stable identifier", r.PathPrefix)
		}
		if r.Category == "" {
			t.Errorf("FrozenRule PathPrefix=%q has empty Category", r.PathPrefix)
		}
	}
}

// TestFindFrozenRule_PermissionSurface verifies FindFrozenRule returns the rule for a
// permission-surface path (A1 expansion entry).
func TestFindFrozenRule_PermissionSurface(t *testing.T) {
	t.Parallel()

	r := FindFrozenRule(".claude/settings.json")
	if r == nil {
		t.Fatal("FindFrozenRule(.claude/settings.json) = nil, must find the permission-surface rule")
	}
	if r.Name == "" {
		t.Error("FindFrozenRule returned a rule with empty Name")
	}
}

// TestFindFrozenRule_SelfProtection verifies FindFrozenRule returns the rule for a
// frozen_guard source file (REQ-HEV3-025 self-protection).
func TestFindFrozenRule_SelfProtection(t *testing.T) {
	t.Parallel()

	r := FindFrozenRule("internal/harness/safety/frozen_guard.go")
	if r == nil {
		t.Fatal("FindFrozenRule(internal/harness/safety/frozen_guard.go) = nil, must find the self-protection rule")
	}
}

// TestFindFrozenRule_NoMatch verifies FindFrozenRule returns nil for non-Frozen paths.
func TestFindFrozenRule_NoMatch(t *testing.T) {
	t.Parallel()

	r := FindFrozenRule(".claude/agents/harness/foo.md")
	if r != nil {
		t.Errorf("FindFrozenRule(harness path) = %+v, must be nil", r)
	}
}
