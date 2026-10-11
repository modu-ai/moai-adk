package cli

// factory_merge_composite_exit_test.go — card t1582 sync-audit F1 (REQ-MWQ2-006):
// a measurement failure that CO-OCCURS with an earlier failing condition must
// still reach the process exit as a deliberate non-zero code. The recorded run
// names only its FIRST failing condition (failed_condition), so an exit decision
// that reads that field alone skips the measurement failure whenever sync-audit,
// conflict-free, or tree-identity fails before re-measure-record does. The
// verdict itself (the printed line, the checks, failed_condition) is unchanged;
// only the exit decision reads the full set of recorded checks.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

func TestFactoryMergeReadyCompositeFailureResolvesNonZeroExit(t *testing.T) {
	// sync-audit fails FIRST (the sync record reads audit-ready, not complete),
	// and the re-measure record for the candidate tree is cleared, so condition
	// (d) fails as well. failed_condition names sync-audit; the process exit must
	// still carry the measurement failure.
	repo, lockRoot, specID := mergeReadyFixture(t, "audit-ready", "lane-9", "sess-lane-9")
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
	if !resolved || code != factory.MergeExitRecordInvalid {
		t.Fatalf("a measurement failure that co-occurs with sync-audit must reach the process exit as MergeExitRecordInvalid (%d): code=%d resolved=%t err=%v out=%s",
			factory.MergeExitRecordInvalid, code, resolved, err, out)
	}
	if !strings.Contains(out, `"failed_condition":"sync-audit"`) {
		t.Fatalf("the verdict must still name the first failing condition: %s", out)
	}
}
