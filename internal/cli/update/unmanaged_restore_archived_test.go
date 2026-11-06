package update

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// F33 regression (iter5 E6): an evaluator profile the reconciliation archived and
// removed must not come back from the backup in the same run. The order mirrors the
// template-sync flow: the backup is taken first, then ReconcileManagedPaths removes
// the stale file, then the unmanaged-directory restore runs against that backup.
// This is a component composition, not a full moai update run.
func TestUnmanagedRestore_DoesNotResurrectArchivedRemoved(t *testing.T) {
	const rel = ".moai/config/evaluator-profiles/retired.md"
	const staleContent = "# evaluator profile retired by the current template\n"
	root := newClassifyFixture(t, map[string]string{rel: staleContent})

	backupDir, err := backup.BackupMoaiConfig(root)
	if err != nil {
		t.Fatalf("BackupMoaiConfig: %v", err)
	}

	carried := map[string]string{".claude/rules/moai/update-note.md": "# template render\n"}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(rel, manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track retired profile: %v", err)
	}
	summary := recRunUpdate(t, root, recTmplFS(carried), renderWith(carried), mgr.Manifest(), carried)
	if !containsPath(summary.ArchivedRemoved, rel) {
		t.Fatalf("precondition: summary.ArchivedRemoved = %v, want it to list %s", summary.ArchivedRemoved, rel)
	}

	// The skip set is built the way the template-sync flow builds it from the summary.
	skipped := map[string]bool{}
	for _, a := range summary.ArchivedRemoved {
		skipped[a] = true
	}
	for _, p := range summary.Preserved {
		skipped[p] = true
	}
	if _, err := backup.RestoreUnmanagedConfigDirsSkipping(root, backupDir, func(refPath string) bool { return skipped[refPath] }); err != nil {
		t.Fatalf("RestoreUnmanagedConfigDirsSkipping: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); !os.IsNotExist(err) {
		t.Errorf("archived-removed profile came back after the unmanaged restore (stat err=%v)", err)
	}
}
