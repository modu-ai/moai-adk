package cli

// update_backup_retention_report_test.go — AC-UBR-12 at the update output surface
// (SPEC-UPDATE-BACKUP-RETENTION-001): the rendered lines name every removed run,
// every retained excess run with its reason, and the unmanaged entries as a count
// and a size. The pruner-level report is asserted in the backup package.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
)

func TestBackupRetentionReportLines_NamesEveryDecision(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "proj")
	ns := filepath.Join(root, ".moai", "backups")
	cfg := filepath.Join(root, ".moai-backups")
	rep := backup.PruneReport{
		Runs: 5, Window: 1, Protected: 2, Removed: 3,
		Items: []backup.PruneItem{
			{Run: filepath.Join(ns, "update-2026-10-04T00-00-00Z"), Dirs: []string{filepath.Join(ns, "update-2026-10-04T00-00-00Z")}, Action: backup.RetentionActionRemoved},
			{Run: filepath.Join(cfg, "20261003_000000"), Dirs: []string{filepath.Join(cfg, "20261003_000000")}, Action: backup.RetentionActionRetained, Reason: backup.RetentionReasonProtectsVersions, Protects: 2},
			{Run: filepath.Join(ns, "update-2026-10-02T00-00-00Z"), Dirs: []string{filepath.Join(ns, "update-2026-10-02T00-00-00Z")}, Action: backup.RetentionActionRetained, Reason: backup.RetentionReasonUnclassifiable, Detail: "pre-clean/pipe"},
			{Run: filepath.Join(ns, "update-2026-10-01T00-00-00Z"), Dirs: []string{filepath.Join(ns, "update-2026-10-01T00-00-00Z")}, Action: backup.RetentionActionRetained, Reason: backup.RetentionReasonRemovalFailed, Detail: "permission denied"},
		},
		Unmanaged: backup.UnmanagedEntries{Count: 2, Bytes: 30},
	}

	got := strings.Join(backupRetentionReportLines(rep, root), "\n")

	for _, want := range []string{
		"removed", "update-2026-10-04T00-00-00Z",
		"retained", "20261003_000000", backup.RetentionReasonProtectsVersions, "2",
		"update-2026-10-02T00-00-00Z", backup.RetentionReasonUnclassifiable, "pre-clean/pipe",
		"update-2026-10-01T00-00-00Z", backup.RetentionReasonRemovalFailed, "permission denied",
		"unmanaged", "30",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("AC-UBR-12: the report lines do not contain %q:\n%s", want, got)
		}
	}
}

func TestBackupRetentionReportLines_QuietWhenNothingHappened(t *testing.T) {
	if lines := backupRetentionReportLines(backup.PruneReport{Runs: 3, Window: 3}, "/proj"); len(lines) != 0 {
		t.Errorf("a prune with no excess runs and no unmanaged entries rendered %d line(s): %v", len(lines), lines)
	}
}
