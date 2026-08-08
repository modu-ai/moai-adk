package web

// agentfm_ordering_test.go — the agent-grid grouping and cheap-first ordering.
//
// The grid sorts agents by catalog class then by cost. Both rank functions are
// pure lookups over closed enums, and the existing tests only reach the branches
// the fixture agents happen to occupy, so several classes and effort levels were
// never evaluated. A wrong rank is not a crash — it is a silently mis-ordered
// grid — which is exactly the kind of defect a total-mapping test catches.

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/v4manifest"
)

// readDirNames lists the entry names of dir.
func readDirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}

// TestAgentGroupRankCoversEveryCatalogClass asserts the rank function is a
// total, order-correct mapping over the retained catalog: each class occupies a
// distinct rank, classes ascend in the documented display order, and anything
// unclassified lands in the final bucket.
func TestAgentGroupRankCoversEveryCatalogClass(t *testing.T) {
	byClass := map[string][]string{
		"core/manager":   {"manager-spec", "manager-develop", "manager-docs", "manager-git", "manager-design"},
		"meta/evaluator": {"plan-auditor", "sync-auditor", "super-advisor"},
		"builder":        {"builder-harness"},
		"specialist":     {"e2e-tester"},
	}
	wantRank := map[string]int{
		"core/manager": 0, "meta/evaluator": 1, "builder": 2, "specialist": 3,
	}

	for class, names := range byClass {
		for _, name := range names {
			if got := agentGroupRank(name); got != wantRank[class] {
				t.Errorf("agentGroupRank(%q) = %d, want %d (%s)", name, got, wantRank[class], class)
			}
		}
	}

	// Everything outside the named catalog falls into the trailing bucket, so an
	// agent added later sorts last rather than silently joining a class.
	for _, other := range []string{"Explore", "hns-oss-docs-specialist", "", "not-an-agent"} {
		if got := agentGroupRank(other); got != 4 {
			t.Errorf("agentGroupRank(%q) = %d, want 4 (the unclassified bucket)", other, got)
		}
	}
}

// TestAgentGroupLabelMatchesRank pins the label side: every ranked class has a
// non-empty header, the labels are distinct, and the unclassified bucket has no
// header (those rows are filtered out before the grid, so a label would be a
// heading over nothing).
func TestAgentGroupLabelMatchesRank(t *testing.T) {
	repr := map[int]string{
		0: "manager-spec",
		1: "plan-auditor",
		2: "builder-harness",
		3: "e2e-tester",
		4: "Explore",
	}
	seen := map[string]int{}
	for rank := 0; rank <= 3; rank++ {
		label := agentGroupLabel(repr[rank])
		if label == "" {
			t.Errorf("rank %d (%s) has no group label", rank, repr[rank])
			continue
		}
		if prev, dup := seen[label]; dup {
			t.Errorf("rank %d and rank %d share the label %q — the grouping is not visible", prev, rank, label)
		}
		seen[label] = rank
	}
	if got := agentGroupLabel(repr[4]); got != "" {
		t.Errorf("the unclassified bucket has label %q, want empty (its rows never reach the grid)", got)
	}
}

// TestAgentEffortCostRankIsMonotonic asserts the effort ordering is strictly
// ascending across the named levels, and that an unrecognized or empty effort
// lands at the medium tier rather than at an extreme — an unset row floating to
// the top or bottom of its group would misrepresent its cost.
func TestAgentEffortCostRankIsMonotonic(t *testing.T) {
	ordered := []string{
		v4manifest.EffortLow,
		v4manifest.EffortMedium,
		v4manifest.EffortHigh,
		v4manifest.EffortMax,
	}
	for i := 1; i < len(ordered); i++ {
		lo, hi := agentEffortCostRank(ordered[i-1]), agentEffortCostRank(ordered[i])
		if lo >= hi {
			t.Errorf("effort %q (rank %d) does not sort cheaper than %q (rank %d)",
				ordered[i-1], lo, ordered[i], hi)
		}
	}

	mediumRank := agentEffortCostRank(v4manifest.EffortMedium)
	for _, unset := range []string{"", "xhigh", "not-an-effort"} {
		if got := agentEffortCostRank(unset); got != mediumRank {
			t.Errorf("agentEffortCostRank(%q) = %d, want the medium rank %d", unset, got, mediumRank)
		}
	}
}

// TestAgentModelCostRankIsMonotonic asserts the model ordering: haiku cheaper
// than sonnet cheaper than opus, with unmapped values sorting last so they do
// not masquerade as cheap.
func TestAgentModelCostRankIsMonotonic(t *testing.T) {
	ordered := []string{v4manifest.ModelHaiku, v4manifest.ModelSonnet, v4manifest.ModelOpus}
	for i := 1; i < len(ordered); i++ {
		if agentModelCostRank(ordered[i-1]) >= agentModelCostRank(ordered[i]) {
			t.Errorf("model %q does not sort cheaper than %q", ordered[i-1], ordered[i])
		}
	}
	last := agentModelCostRank(v4manifest.ModelOpus)
	for _, unmapped := range []string{v4manifest.ModelInherit, "", "fable"} {
		if got := agentModelCostRank(unmapped); got <= last {
			t.Errorf("agentModelCostRank(%q) = %d, want a rank after opus (%d)", unmapped, got, last)
		}
	}
}

// TestAgentTierBadgeCustomOverridesModelColor pins the override sentinel: an
// effort of `max` marks the row neutral-custom regardless of model, because a
// per-agent override is no longer described by the tier's cost color.
func TestAgentTierBadgeCustomOverridesModelColor(t *testing.T) {
	custom := agentTierBadge("manager-develop", v4manifest.ModelOpus, v4manifest.EffortMax)
	if !custom.IsCustom || custom.Glyph != "custom" {
		t.Errorf("effort=max did not produce the neutral custom badge: %+v", custom)
	}

	plain := agentTierBadge("manager-develop", v4manifest.ModelOpus, v4manifest.EffortHigh)
	if plain.IsCustom {
		t.Errorf("a non-override effort was marked custom: %+v", plain)
	}
	if plain.Glyph == "custom" {
		t.Error("a non-override row reused the custom glyph")
	}
	if plain.TooltipKey == custom.TooltipKey {
		t.Error("custom and model-derived rows share a tooltip key — the distinction is not surfaced")
	}
}

// TestApplyPerfTierEditsIgnoresEmptyAndInvalid pins the two no-op guards: an
// empty tier means "no change requested", and an out-of-set value is refused
// without touching disk. Both must return nil so a save carrying neither is not
// reported as a failure.
func TestApplyPerfTierEditsIgnoresEmptyAndInvalid(t *testing.T) {
	root := t.TempDir()
	if err := applyPerfTierEdits(root, ""); err != nil {
		t.Errorf("empty tier should be a no-op, got %v", err)
	}
	if err := applyPerfTierEdits(root, "definitely-not-a-tier"); err != nil {
		t.Errorf("an out-of-set tier should be refused silently, got %v", err)
	}
	// Neither call may create config state in an empty root.
	entries, err := readDirNames(root)
	if err != nil {
		t.Fatalf("read root: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("a no-op tier edit wrote to disk: %v", entries)
	}
}
