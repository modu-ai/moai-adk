package v4manifest

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// This file tests the sub-agent 4-color tier model introduced by
// SPEC-WEBCONF-SIMPLIFY-001 M1 (Option A: name-keyed lookup table,
// display-only). The tier is INDEPENDENT of each agent's effort frontmatter
// (design.md §B); these tests pin that independence by asserting the table
// is the sole source of truth for tier assignment.

// expectedAgentTiers is the hand-curated name→tier mapping per design.md §C
// (distribution 🔴×4 · 🟠×4 · 🔵×5 · 🩵×7 = 20). Every catalog agent file stem
// under .claude/agents/{moai,harness}/ MUST appear here exactly once.
var expectedAgentTiers = map[string]Tier{
	// 🔴 — deep reasoning (×4)
	"manager-spec":  TierRed,
	"plan-auditor":  TierRed,
	"super-advisor": TierRed,
	"sync-auditor":  TierRed,
	// 🟠 — heavy reasoning (×4)
	"manager-develop": TierOrange,
	"manager-design":   TierOrange,
	"builder-harness":  TierOrange,
	"e2e-tester":   TierOrange,
	// 🔵 — moderate reasoning (×5)
	"manager-docs":            TierBlue,
	"manager-git":             TierBlue,
	"quality-specialist":      TierBlue,
	"cli-template-specialist": TierBlue,
	"workflow-specialist":     TierBlue,
	// 🩵 — light / narrow-scope (×7)
	"hns-github-specialist":                   TierLightBlue,
	"hns-oss-docs-content-author-specialist":  TierLightBlue,
	"hns-oss-docs-locale-translator-specialist": TierLightBlue,
	"hns-oss-docs-structure-curator-specialist": TierLightBlue,
	"hns-release-specialist":                  TierLightBlue,
	"hns-release-update-specialist":           TierLightBlue,
	"hook-ci-specialist":                      TierLightBlue,
}

// TestAgentTier_All20ExpectedAgents verifies that AgentTier returns the
// design.md §C tier for each of the 20 expected agent names
// (AC-WC-005 + AC-WC-016 data-driven half).
func TestAgentTier_All20ExpectedAgents(t *testing.T) {
	if len(expectedAgentTiers) != 20 {
		t.Fatalf("expectedAgentTiers fixture has %d entries, want 20 — fixture is wrong", len(expectedAgentTiers))
	}
	for name, wantTier := range expectedAgentTiers {
		gotTier, ok := AgentTier(name)
		if !ok {
			t.Errorf("AgentTier(%q) returned ok=false; want an entry", name)
			continue
		}
		if gotTier != wantTier {
			t.Errorf("AgentTier(%q) = %q, want %q", name, gotTier, wantTier)
		}
	}
}

// TestAgentTier_CatalogFileCoverage enumerates the actual agent files under
// .claude/agents/{moai,harness}/ and asserts every file stem has a table
// entry, AND the table holds no orphan entry for a non-existent agent
// (AC-WC-016 dynamic half — catches drift if the catalog grows or shrinks).
func TestAgentTier_CatalogFileCoverage(t *testing.T) {
	catalog := catalogAgentStems(t)
	if len(catalog) == 0 {
		t.Skip("no agent .md files found — running outside the moai-adk-go repo layout")
	}
	table := AllAgentTiers()

	// Every catalog agent has a tier entry.
	for _, stem := range catalog {
		if _, ok := table[stem]; !ok {
			t.Errorf("catalog agent %q has no tier entry in the name table", stem)
		}
	}
	// No orphan entry: every table key corresponds to a real catalog file.
	for name := range table {
		found := false
		for _, stem := range catalog {
			if stem == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("tier table has entry %q but no matching agent file exists in .claude/agents/{moai,harness}/", name)
		}
	}
}

// TestAgentTier_Distribution pins the 🔴×4 · 🟠×4 · 🔵×5 · 🩵×7 split
// (design.md §C distribution footnote).
func TestAgentTier_Distribution(t *testing.T) {
	table := AllAgentTiers()
	counts := map[Tier]int{}
	for _, tier := range table {
		counts[tier]++
	}
	want := map[Tier]int{
		TierRed:       4,
		TierOrange:    4,
		TierBlue:      5,
		TierLightBlue: 7,
	}
	for tier, wantN := range want {
		if got := counts[tier]; got != wantN {
			t.Errorf("tier %q has %d agents, want %d", tier, got, wantN)
		}
	}
	if total := len(table); total != 20 {
		t.Errorf("table has %d entries, want 20", total)
	}
}

// TestAgentTier_AbsentEffortAgents verifies the 3 hns-oss-docs-* agents that
// lack an effort frontmatter key each get 🩵 from the name table — NOT a
// fallback and NOT "no badge" (AC-WC-021 / GWT-16). This is the Option A
// invariant: absent effort does not block tier display.
func TestAgentTier_AbsentEffortAgents(t *testing.T) {
	absent := []string{
		"hns-oss-docs-content-author-specialist",
		"hns-oss-docs-locale-translator-specialist",
		"hns-oss-docs-structure-curator-specialist",
	}
	for _, name := range absent {
		got, ok := AgentTier(name)
		if !ok {
			t.Errorf("AgentTier(%q) returned ok=false; the absent-effort agents MUST have a table entry (AC-WC-021)", name)
			continue
		}
		if got != TierLightBlue {
			t.Errorf("AgentTier(%q) = %q, want %q (🩵 from the name table, not an effort fallback)", name, got, TierLightBlue)
		}
	}
}

// TestAgentTier_UnknownAgentReturnsNotOK asserts a name with no table entry
// returns ok=false (design.md EC-6: a future agent renders no badge until
// explicitly added to the table).
func TestAgentTier_UnknownAgentReturnsNotOK(t *testing.T) {
	if _, ok := AgentTier("nonexistent-future-agent"); ok {
		t.Error(`AgentTier("nonexistent-future-agent") returned ok=true; want false for an unmapped name`)
	}
}

// TestTierSuggestedModelEffort verifies the tier → suggested-(model, effort)
// table matches the 4-row spec exactly (design.md §D / AC-WC-006). These are
// SUGGESTIONS applied on explicit user action; they do NOT change the table.
func TestTierSuggestedModelEffort(t *testing.T) {
	cases := []struct {
		tier       Tier
		wantModel  string
		wantEffort string
	}{
		{TierRed, ModelOpus, EffortXhigh},
		{TierOrange, ModelOpus, EffortHigh},
		{TierBlue, ModelSonnet, EffortMedium},
		{TierLightBlue, ModelHaiku, EffortLow},
	}
	for _, tc := range cases {
		gotModel, gotEffort := TierSuggestedModelEffort(tc.tier)
		if gotModel != tc.wantModel || gotEffort != tc.wantEffort {
			t.Errorf("TierSuggestedModelEffort(%q) = (%q, %q), want (%q, %q)",
				tc.tier, gotModel, gotEffort, tc.wantModel, tc.wantEffort)
		}
	}
}

// TestTierSuggestedModelEffort_UnknownTierReturnsEmpty asserts an unmapped
// tier returns empty strings (defensive — the closed set has 4 members, but
// a caller must not panic on a zero-value Tier).
func TestTierSuggestedModelEffort_UnknownTierReturnsEmpty(t *testing.T) {
	gotModel, gotEffort := TierSuggestedModelEffort(Tier("not-a-real-tier"))
	if gotModel != "" || gotEffort != "" {
		t.Errorf(`TierSuggestedModelEffort("not-a-real-tier") = (%q, %q), want ("", "")`, gotModel, gotEffort)
	}
}

// TestTierColor verifies each tier maps to its render glyph (design.md §C/D).
func TestTierColor(t *testing.T) {
	cases := map[Tier]string{
		TierRed:       "🔴",
		TierOrange:    "🟠",
		TierBlue:      "🔵",
		TierLightBlue: "🩵",
	}
	for tier, want := range cases {
		if got := TierColor(tier); got != want {
			t.Errorf("TierColor(%q) = %q, want %q", tier, got, want)
		}
	}
}

// TestTierColor_UnknownTierReturnsEmpty asserts an unmapped tier yields an
// empty glyph (the UI layer in M5 renders a neutral "custom"/"unmapped"
// badge on empty — design.md §D override-sentinel semantics are UI-layer,
// not M1's table).
func TestTierColor_UnknownTierReturnsEmpty(t *testing.T) {
	if got := TierColor(Tier("not-a-real-tier")); got != "" {
		t.Errorf(`TierColor("not-a-real-tier") = %q, want ""`, got)
	}
}

// catalogAgentStems reads the agent catalog from disk (.claude/agents/{moai,
// harness}/*.md) and returns the file stems (without extension), sorted. The
// stems are the lookup keys into the name→tier table and match agentfm.
// agentfm.AgentInfo.Name's contract ("file base name, extension excluded").
func catalogAgentStems(t *testing.T) []string {
	t.Helper()
	root := repoRoot(t)
	dirs := []string{
		filepath.Join(root, ".claude", "agents", "moai"),
		filepath.Join(root, ".claude", "agents", "harness"),
	}
	var stems []string
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // directory absent in this execution context — caller Skips on empty
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if filepath.Ext(name) != ".md" {
				continue
			}
			stems = append(stems, strings.TrimSuffix(name, ".md"))
		}
	}
	sort.Strings(stems)
	return stems
}

// repoRoot walks up from this test file's location until it finds go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed — cannot locate test file")
	}
	dir := filepath.Dir(file)
	for i := 0; i < 12; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("could not locate repo root (go.mod) by walking up from the test file")
	return ""
}
