package cli

// factory_merge_ready_red_t1582_test.go — plan-phase RED reproduction for
// card t1582 item ③ (measurement-failure REFUSED exits 0). The fixture
// reuses mergeReadyFixture with sync_status "complete" — conditions (a)
// sync-audit, (b) conflict-free and (c) tree-identity all pass — then clears
// the seeded re-measure record so ONLY condition (d) fails (the D4 fixture
// design: mergeReadyFixture's own comment names "or clears the store" as
// the (d)-failure route). RED on the pre-repair tree: the REFUSED verdict
// renders with the measurement failure named and the command still exits 0
// (err nil from the cobra Execute). GREEN after the M3 repair: the same
// refusal exits non-zero.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRedT1582MergeReadyMeasurementFailureStillExitsZero(t *testing.T) {
	repo, lockRoot, specID := mergeReadyFixture(t, "complete", "lane-9", "sess-lane-9")
	// Break ONLY condition (d): clear the seeded re-measure record —
	// (a) sync-audit passes (sync_status complete), (b)+(c) pass on the
	// absorbed same-repo fixture.
	tree := strings.TrimSpace(func() string {
		cmd := exec.Command("git", "rev-parse", "WT-card^{tree}")
		cmd.Dir = repo
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("read candidate tree: %v", err)
		}
		return string(out)
	}())
	if err := os.Remove(filepath.Join(lockRoot, ".moai", "state", "remeasure", tree+".json")); err != nil {
		t.Fatalf("clear the seeded record: %v", err)
	}
	out, err := runFactoryMerge(t, "merge", "ready", "--card", "t9001", "--spec", specID, "--branch", "WT-card", "--develop", "develop", "--session", "sess-lane-9", "--json")
	if err == nil {
		t.Fatalf("RED t1582-AC005: the measurement-failure REFUSED verdict still exits 0 (err nil) — out: %s", out)
	}
	// The refusal keeps naming the measurement failure whatever the exit
	// code becomes (the verdict output is preserved across the repair).
	if !strings.Contains(out, "re-measure-record") {
		t.Fatalf("the refusal must keep naming the re-measure-record condition — out: %s", out)
	}
}
