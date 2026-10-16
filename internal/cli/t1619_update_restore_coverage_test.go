package cli

// t1619_update_restore_coverage_test.go — card t1619, bundle B2: the unexecuted changed lines
// of update_restore.go (reportUncoveredRestorePaths skips targets with no backed-up copy, and
// returns early when none is uncovered). Family: TestUpdate.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// TestUpdateRestoreUncoveredReportSkipsAbsentCopies — a backup that holds no copies lists no
// uncovered paths; the settings notice still prints and the heading does not.
func TestUpdateRestoreUncoveredReportSkipsAbsentCopies(t *testing.T) {
	projectRoot := t.TempDir()
	absBackup := t.TempDir()
	var out strings.Builder
	reportUncoveredRestorePaths(&out, projectRoot, absBackup)

	heading := tui.Translate("update.restore.not_covered_heading", "en")
	if strings.Contains(out.String(), heading) {
		t.Fatalf("a backup with no copies must list no uncovered paths, got %q", out.String())
	}
	if out.Len() == 0 {
		t.Fatal("the restore report must still print its settings notice")
	}
}
