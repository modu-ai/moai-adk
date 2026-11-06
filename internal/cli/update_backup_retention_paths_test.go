package cli

// update_backup_retention_paths_test.go — AC-UBR-21 at the update output surface
// (SPEC-UPDATE-BACKUP-RETENTION-001, REQ-UBR-020): the full-reinstall path and the
// template-sync path each print the retention report rows for the runs they prune.
// Each test fails when the call that prints those rows is removed from its path,
// so the rows are pinned where they are produced, not only in the renderer
// (update_backup_retention_report_test.go).

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedStaleConfigBackupRuns creates n timestamped, empty config-backup runs under
// .moai-backups/ — the shape the writer produces and the pruner may remove.
func seedStaleConfigBackupRuns(t *testing.T, root string, n int) []string {
	t.Helper()
	var names []string
	for i := 1; i <= n; i++ {
		name := fmt.Sprintf("2025010%d_000000", i)
		makeTestDir(t, root, filepath.Join(".moai-backups", name))
		names = append(names, name)
	}
	return names
}

// TestCleanReinstall_PrintsBackupRetentionRows pins the full-reinstall path:
// pruning stale runs prints one row per removed run and the retention summary.
func TestCleanReinstall_PrintsBackupRetentionRows(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v2.16.1\n")
	writeTestFile(t, root, ".claude/agents/moai/manager-strategy.md", "retired\n")
	writeTestFile(t, root, ".moai/specs/SPEC-USER-RET/spec.md", "user spec\n")
	stale := seedStaleConfigBackupRuns(t, root, 7)

	var out bytes.Buffer
	if _, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
		Out:              &out,
		Deployer:         &stubDeployer{},
		RunMigrateAgency: (&stubMigrateRunner{}).Run,
	}); err != nil {
		t.Fatalf("runCleanReinstall: %v", err)
	}
	got := out.String()

	removedRow := ""
	for _, name := range stale {
		if _, err := os.Stat(filepath.Join(root, ".moai-backups", name)); os.IsNotExist(err) {
			removedRow = "[clean-reinstall] old backup run removed: " + filepath.ToSlash(filepath.Join(".moai-backups", name))
			break
		}
	}
	if removedRow == "" {
		t.Fatalf("AC-UBR-21: fixture did not prune any seeded run; output:\n%s", got)
	}
	if !strings.Contains(got, removedRow) {
		t.Errorf("AC-UBR-21: the full-reinstall output has no per-run removal row %q:\n%s", removedRow, got)
	}
	if !strings.Contains(got, "[clean-reinstall] backup retention: ") {
		t.Errorf("AC-UBR-21: the full-reinstall output has no retention summary row:\n%s", got)
	}
}

// TestTemplateSync_RecordsBackupRetentionRows pins the template-sync path: the
// same prune records the report rows in the update ledger.
func TestTemplateSync_RecordsBackupRetentionRows(t *testing.T) {
	root := makeLLMPreserveFixture(t, false, false)
	stale := seedStaleConfigBackupRuns(t, root, 7)

	updateLedger.reset()
	t.Cleanup(updateLedger.reset)
	runTemplateSyncAt(t, root)

	var rows []string
	for _, r := range updateLedger.reference {
		rows = append(rows, r.msg)
	}
	joined := strings.Join(rows, "\n")

	removedRow := ""
	for _, name := range stale {
		if _, err := os.Stat(filepath.Join(root, ".moai-backups", name)); os.IsNotExist(err) {
			removedRow = "old backup run removed: " + filepath.ToSlash(filepath.Join(".moai-backups", name))
			break
		}
	}
	if removedRow == "" {
		t.Fatalf("AC-UBR-21: fixture did not prune any seeded run; ledger rows:\n%s", joined)
	}
	if !strings.Contains(joined, removedRow) {
		t.Errorf("AC-UBR-21: the template-sync ledger has no per-run removal row %q:\n%s", removedRow, joined)
	}
	if !strings.Contains(joined, "backup retention: ") {
		t.Errorf("AC-UBR-21: the template-sync ledger has no retention summary row:\n%s", joined)
	}
}
