package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPlanHashIncludesDecisionIndex: a decision-index.md edit after the
// audited SHA must change the plan-artifact hash, so the audit decider and the
// run Phase 1 skip-cache both see it (REQ-FDA-014, design §5 digest widening).
func TestPlanHashIncludesDecisionIndex(t *testing.T) {
	t.Parallel()
	base := func(index string) string {
		dir := t.TempDir()
		for _, name := range []string{"spec.md", "plan.md", "acceptance.md"} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte("initial "+name), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if index != "" {
			if err := os.WriteFile(filepath.Join(dir, "decision-index.md"), []byte(index), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}
	cache := NewInMemoryCache()
	before, err := cache.ComputeHash(base("Class: product-level\nOperator verdict:\n"))
	if err != nil {
		t.Fatal(err)
	}
	after, err := cache.ComputeHash(base("Class: implementation-level\nOperator verdict: DEFAULT-APPLIED\n"))
	if err != nil {
		t.Fatal(err)
	}
	if before == after {
		t.Fatal("reclassifying a decision-index row did NOT change the plan-artifact hash")
	}
	absent, err := cache.ComputeHash(base(""))
	if err != nil {
		t.Fatal(err)
	}
	if absent == before {
		t.Fatal("a SPEC without decision-index.md hashes like one with it")
	}
}
