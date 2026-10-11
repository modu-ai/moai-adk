package cli

// factory_merge_measurement_exit_test.go — card t1582 item ③ (REQ-MWQ2-006):
// the measurement-failure refusal of `merge ready` must reach the process exit
// as a deliberate non-zero code. cmd/moai/main.go exits with the code
// ResolveExitCode reports for the cobra error chain, so the guard is asserted
// at that boundary: the verdict still names the failed condition, and the
// returned error resolves to a non-zero exit code. The non-measurement refusal
// (sync-audit) and the window-contest verdict (waiting) keep their zero-exit
// assertions in factory_merge_test.go.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFactoryMergeReadyMeasurementFailureResolvesNonZeroExit(t *testing.T) {
	repo, lockRoot, specID := mergeReadyFixture(t, "complete", "lane-9", "sess-lane-9")
	// Break only condition (d): clear the seeded re-measure record, exactly as
	// the RED overlay does for the same measurement-failure route.
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
	code, resolved := ResolveExitCode(err)
	if !resolved || code == 0 {
		t.Fatalf("the measurement-failure refusal must reach the process exit as a non-zero code: code=%d resolved=%t err=%v out=%s", code, resolved, err, out)
	}
	if !strings.Contains(out, `"failed_condition":"re-measure-record"`) {
		t.Fatalf("the verdict must still name the failed condition: %s", out)
	}
}
