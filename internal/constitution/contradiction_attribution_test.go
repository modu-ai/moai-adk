package constitution

import (
	"fmt"
	"strings"
	"testing"
)

// buildRegistry returns a registry of n rules whose clauses all carry an
// obligation modal, so that a per-rule relaxation check would fire on each.
func buildRegistry(n int) *Registry {
	entries := make([]Rule, 0, n)
	for i := 1; i <= n; i++ {
		zone := ZoneEvolvable
		if i%2 == 0 {
			zone = ZoneFrozen
		}
		entries = append(entries, Rule{
			ID:     fmt.Sprintf("CONST-V3R2-%03d", i),
			Zone:   zone,
			Clause: fmt.Sprintf("Rule %d MUST hold.", i),
		})
	}
	return &Registry{Entries: entries}
}

// TestScanRelaxationIsAttributedOnlyToTheProposedRule pins the defect card t992
// measured: the relaxation verdict reads proposal.Before and proposal.After and
// nothing else, so running it per registry entry replicated one verdict across
// the whole registry. Live measurement on a 101-entry registry produced 100
// `Constraint relaxation` conflicts for a single amendment, 43 of them naming
// Evolvable rules.
//
// A single-entry registry cannot see this — the pre-existing
// TestContradictionDetector_Scan_ZoneRelaxation uses one rule and passes either
// way. The registry size is therefore the load-bearing part of this test, not
// incidental setup: it must exceed 1 for the assertion to have content.
func TestScanRelaxationIsAttributedOnlyToTheProposedRule(t *testing.T) {
	const registrySize = 40

	d := NewContradictionDetector()
	registry := buildRegistry(registrySize)

	proposal := &AmendmentProposal{
		RuleID: "CONST-V3R2-001",
		Before: "Rule 1 MUST hold.",
		After:  "Rule 1 SHOULD hold.",
	}

	result, err := d.Scan(proposal, registry)
	if err == nil {
		t.Fatal("a relaxation must still block; Scan returned no error")
	}
	if result == nil {
		t.Fatal("Scan returned a nil result alongside its error")
	}

	var relaxations []ConflictDetail
	for _, c := range result.Conflicts {
		if strings.HasPrefix(c.Description, "Constraint relaxation:") {
			relaxations = append(relaxations, c)
		}
	}

	if len(relaxations) != 1 {
		ids := make([]string, 0, len(relaxations))
		for _, c := range relaxations {
			ids = append(ids, c.ConflictingRuleID)
		}
		t.Fatalf("one relaxation must yield exactly 1 conflict, got %d over a %d-entry registry: %v",
			len(relaxations), registrySize, ids)
	}

	if got := relaxations[0].ConflictingRuleID; got != proposal.RuleID {
		t.Errorf("the relaxation must be attributed to the rule it changes: got %q, want %q",
			got, proposal.RuleID)
	}

	if !relaxations[0].IsBlocking {
		t.Error("a relaxation conflict must stay blocking")
	}
}

// TestScanRelaxationCountIsIndependentOfRegistrySize is the companion that
// makes the previous assertion non-vacuous in the other direction: if the
// verdict were ever re-coupled to registry iteration, the count would track the
// registry size. Two sizes that differ by an order of magnitude must produce
// the same count.
func TestScanRelaxationCountIsIndependentOfRegistrySize(t *testing.T) {
	d := NewContradictionDetector()
	proposal := &AmendmentProposal{
		RuleID: "CONST-V3R2-001",
		Before: "Rule 1 MUST hold.",
		After:  "Rule 1 SHOULD hold.",
	}

	counts := map[int]int{}
	for _, size := range []int{5, 50} {
		result, _ := d.Scan(proposal, buildRegistry(size))
		n := 0
		for _, c := range result.Conflicts {
			if strings.HasPrefix(c.Description, "Constraint relaxation:") {
				n++
			}
		}
		counts[size] = n
	}

	if counts[5] != counts[50] {
		t.Errorf("relaxation count tracks registry size: 5-entry=%d, 50-entry=%d",
			counts[5], counts[50])
	}
}
