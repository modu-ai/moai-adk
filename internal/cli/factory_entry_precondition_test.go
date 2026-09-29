// factory_entry_precondition_test.go — SPEC-TODO-CLASSIFY-DISPATCH-001
// AC-TCD-014: the absorption-order entry precondition, judged mechanically.
//
// REQ-TCD-013 makes the t1240 self-dispatch surface a run-phase entry
// precondition: mode-aware dispatch (M3) does not start until local develop
// carries `factory next` and `complete`. The absorption landed before this
// branch's first commit (develop `8fc5a7407` carried it; this worktree's
// HEAD `51f3e9878` is its clean absorb), and this test pins the two halves
// the AC names so a future regression is loud rather than silent:
//
//  1. the machine check — the two symbols exist in this tree's internal/cli
//     (the AC's own check command reads the source, the compile-level
//     equivalent of `git grep -l "newFactoryNextCommand" develop --
//     internal/cli` run against the tree under judgment);
//  2. the document check — spec.md (REQ-TCD-013) and plan.md (§D.1) name the
//     absorption source (branch WT-factory-self-dispatch) and the check
//     command, so the precondition stays discoverable after the run.
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFactorySelfDispatchSurfacePresent(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "internal", "cli", "factory_card.go"))
	if err != nil {
		t.Fatalf("read factory_card.go: %v", err)
	}
	body := string(source)
	for _, symbol := range []string{
		"func newFactoryNextCommand()",
		"func newFactoryCompleteCommand()",
	} {
		if !strings.Contains(body, symbol) {
			t.Errorf("absorption precondition violated: internal/cli carries no %q — the t1240 self-dispatch surface is absent and M3 (mode-aware dispatch) may not start (REQ-TCD-013)", symbol)
		}
	}
}

func TestAbsorptionPreconditionDocumented(t *testing.T) {
	specDir := filepath.Join("..", "..", ".moai", "specs", "SPEC-TODO-CLASSIFY-DISPATCH-001")
	spec, err := os.ReadFile(filepath.Join(specDir, "spec.md"))
	if err != nil {
		t.Fatalf("read spec.md: %v", err)
	}
	plan, err := os.ReadFile(filepath.Join(specDir, "plan.md"))
	if err != nil {
		t.Fatalf("read plan.md: %v", err)
	}
	if !strings.Contains(string(spec), "REQ-TCD-013") {
		t.Errorf("spec.md carries no REQ-TCD-013 — the entry precondition is undocumented")
	}
	if !strings.Contains(string(spec), "WT-factory-self-dispatch") {
		t.Errorf("spec.md does not name the absorption source WT-factory-self-dispatch")
	}
	if !strings.Contains(string(plan), "WT-factory-self-dispatch") {
		t.Errorf("plan.md does not name the absorption source WT-factory-self-dispatch")
	}
	if !strings.Contains(string(plan), "git grep") || !strings.Contains(string(plan), "newFactoryNextCommand") {
		t.Errorf("plan.md §D.1 does not record the mechanical check command (git grep newFactoryNextCommand)")
	}
}
