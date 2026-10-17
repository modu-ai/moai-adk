package cli

// migrate_removal_backup_test.go — REQ-RC30-004 / AC-RC30-03 (card t1668,
// SPEC-UPDATE-REMOVAL-FAMILY-001 M1): every file the project migration
// removes lands a byte-identical copy in the RUN backup before the removal,
// and the migration runs after the run's backup stage. The order witness is
// structural: the removed-asset copies must live in the SAME run directory
// the Backup stage created (the one holding backup_metadata.json) — a
// migration that ran before the backup stage would have no run directory to
// write into and would create its own.

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// runUpdateFlowWithProgress drives the full update flow (the same entry
// point `moai update` uses) in the scratch project with --yes, so the
// Backup stage, the migration, and the deploy all run in their production
// order.
func runUpdateFlowWithProgress(t *testing.T, root string) string {
	t.Helper()
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("no-hooks", true, "")
	cmd.Flags().Bool("no-plugin", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("check", "", "")
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&strings.Builder{})
	cmd.SetContext(context.Background())

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if _, err := runTemplateSyncWithProgress(cmd, true); err != nil {
		t.Fatalf("runTemplateSyncWithProgress: %v", err)
	}
	return out.String()
}

// removalBackupCopy locates the run-backup copy of a removed project asset:
// the unique .moai-backups/<run>/removed-assets/<rel> file. It returns the
// run directory and the copy's absolute path, failing the test when the copy
// is absent or ambiguous — the copy IS the criterion under judgment.
func removalBackupCopy(t *testing.T, root, rel string) (string, string) {
	t.Helper()
	pattern := filepath.Join(root, ".moai-backups", "*", "removed-assets", filepath.FromSlash(rel))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob removal backups: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected exactly one run-backup copy of %s under .moai-backups, found %d: %v", rel, len(matches), matches)
	}
	suffix := filepath.Join("removed-assets", filepath.FromSlash(rel))
	if !strings.HasSuffix(matches[0], suffix) {
		t.Fatalf("copy path %s does not carry the expected layout %s", matches[0], suffix)
	}
	return strings.TrimSuffix(matches[0], suffix), matches[0]
}

// TestMigrationBackupBeforeRemoval is AC-RC30-03 (U-02, REQ-RC30-004): with
// the removal switch on in the test fixture, the full update flow removes the
// confirmed identical skill AND (1) the run backup holds a byte-identical
// copy of every removed file, (2) the copy lives in the run directory the
// Backup stage created — the migration ran after the backup stage, and (3)
// the dropped manifest entry stays dropped after the whole flow (the
// lost-update guard's happy path).
func TestMigrationBackupBeforeRemoval(t *testing.T) {
	enableRemovalForTest(t)
	root := buildMigrationFixture(t)
	home := installMigrationUserCounterparts(t)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	runUpdateFlowWithProgress(t, root)

	// The removal really happened (the fixture meets every removal
	// precondition: template-managed record, matching bytes, confirmed
	// counterpart, switch on).
	assertFileAbsent(t, root, migIdenticalSkill)

	// (1) The run backup holds a byte-identical copy of the removed file.
	runDir, copyPath := removalBackupCopy(t, root, migIdenticalSkill)
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	wantBytes, err := fs.ReadFile(embedded, migIdenticalSkill)
	if err != nil {
		t.Fatalf("embedded read: %v", err)
	}
	gotBytes, err := os.ReadFile(copyPath)
	if err != nil {
		t.Fatalf("read removal backup copy: %v", err)
	}
	if string(gotBytes) != string(wantBytes) {
		t.Errorf("the run-backup copy of %s is not byte-identical to the removed file (%d vs %d bytes)",
			migIdenticalSkill, len(gotBytes), len(wantBytes))
	}

	// (2) The copy shares the run directory the Backup stage created: the
	// stage's own artifact (backup_metadata.json) sits beside removed-assets/.
	if _, err := os.Stat(filepath.Join(runDir, "backup_metadata.json")); err != nil {
		t.Errorf("the removed-asset copy is not in the Backup stage's run directory %s (backup_metadata.json: %v) — the migration did not run behind the backup stage", runDir, err)
	}

	// (3) The dropped manifest entry survived the FULL flow as dropped: a
	// later stale-view Save would resurrect it (the lost-update window
	// plan-audit-iter2 R1 reproduced).
	disk := manifest.NewManager()
	if _, err := disk.Load(root); err != nil {
		t.Fatalf("reload manifest after the flow: %v", err)
	}
	if _, found := disk.GetEntry(migIdenticalSkill); found {
		t.Errorf("the removed file's manifest entry resurrected after the full update flow")
	}
}
