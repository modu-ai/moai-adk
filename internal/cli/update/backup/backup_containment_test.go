package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// A backup root that is a symbolic link leads out of the project. Namespace
// pruning must not remove an entry that the link resolves to outside it.
func TestCleanupNamespaceBackups_RefusesSymlinkedBackupRoot(t *testing.T) {
	project := t.TempDir()
	outside := filepath.Join(t.TempDir(), "backups")
	for i := 1; i <= 6; i++ {
		writeUnmanagedFile(t, filepath.Join(outside, fmt.Sprintf("update-2026-01-%02dT00-00-00Z", i), "a.yml"), "shared")
	}
	if err := os.MkdirAll(filepath.Join(project, defs.MoAIDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(project, defs.MoAIDir, defs.NamespaceBackupsSubdir)); err != nil {
		t.Skip(err)
	}

	if n := CleanupNamespaceBackups(project, 5); n != 0 {
		t.Fatalf("CleanupNamespaceBackups removed %d entries through a symlinked root", n)
	}
	if _, err := os.Stat(filepath.Join(outside, "update-2026-01-01T00-00-00Z")); err != nil {
		t.Fatalf("entry outside the project removed: %v", err)
	}
}

// The same holds for the timestamped root .moai-backups.
func TestCleanupOldBackups_RefusesSymlinkedBackupRoot(t *testing.T) {
	project := t.TempDir()
	outside := filepath.Join(t.TempDir(), "backups")
	for i := 1; i <= 6; i++ {
		writeUnmanagedFile(t, filepath.Join(outside, fmt.Sprintf("202601%02d_000000", i), "a.yml"), "shared")
	}
	if err := os.Symlink(outside, filepath.Join(project, defs.BackupsDir)); err != nil {
		t.Skip(err)
	}

	if n := CleanupOldBackups(project, 5); n != 0 {
		t.Fatalf("CleanupOldBackups removed %d entries through a symlinked root", n)
	}
	if _, err := os.Stat(filepath.Join(outside, "20260101_000000")); err != nil {
		t.Fatalf("entry outside the project removed: %v", err)
	}
}

// A side-copy path that is a dangling symbolic link points outside the config
// directory. The side copy must not be written through the link.
func TestRestoreUnmanagedConfigDirs_SideCopyNeverWritesThroughSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "20260101_000000")
	writeUnmanagedFile(t, filepath.Join(backupDir, "evaluator-profiles", "x.md"), "backup body")
	configDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)
	target := filepath.Join(configDir, "evaluator-profiles", "x.md")
	writeUnmanagedFile(t, target, "template body")
	escape := filepath.Join(outside, "escaped.md")
	if err := os.Symlink(escape, target+".from-20260101_000000"); err != nil {
		t.Skip(err)
	}

	refs, _ := RestoreUnmanagedConfigDirs(root, backupDir)
	if _, err := os.Stat(escape); err == nil {
		t.Fatalf("side copy written through a symlink outside the config directory; refs=%+v", refs)
	}
	if len(refs) != 1 || refs[0].Action != "side-copy-refused" {
		t.Fatalf("refs = %+v, want one side-copy-refused entry", refs)
	}
}

// An existing side copy that differs from the backup is kept as it is, and nothing
// is written beside it.
func TestRestoreUnmanagedConfigDirs_SideCopyBlockedKeepsExistingCopy(t *testing.T) {
	root := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "20260101_000000")
	writeUnmanagedFile(t, filepath.Join(backupDir, "evaluator-profiles", "x.md"), "backup body")
	configDir := filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir)
	target := filepath.Join(configDir, "evaluator-profiles", "x.md")
	writeUnmanagedFile(t, target, "template body")
	side := target + ".from-20260101_000000"
	writeUnmanagedFile(t, side, "earlier copy")

	refs, err := RestoreUnmanagedConfigDirs(root, backupDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || refs[0].Action != "side-copy-blocked" {
		t.Fatalf("refs = %+v, want one side-copy-blocked entry", refs)
	}
	got, err := os.ReadFile(side)
	if err != nil || string(got) != "earlier copy" {
		t.Fatalf("existing side copy changed: got %q (err %v)", got, err)
	}
}

// The refusal predicates name a root the cleanup will not touch, and only that: a
// missing root and a real directory are not refused.
func TestBackupRootRefused_OnlySymlinkedRootIsRefused(t *testing.T) {
	project := t.TempDir()
	if OldBackupRootRefused(project) || NamespaceBackupRootRefused(project) {
		t.Fatal("missing backup roots reported as refused")
	}
	if err := os.MkdirAll(filepath.Join(project, defs.MoAIDir, defs.NamespaceBackupsSubdir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(project, defs.BackupsDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if OldBackupRootRefused(project) || NamespaceBackupRootRefused(project) {
		t.Fatal("real backup roots reported as refused")
	}

	linked := t.TempDir()
	if err := os.MkdirAll(filepath.Join(linked, defs.MoAIDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(linked, defs.BackupsDir)); err != nil {
		t.Skip(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(linked, defs.MoAIDir, defs.NamespaceBackupsSubdir)); err != nil {
		t.Skip(err)
	}
	if !OldBackupRootRefused(linked) || !NamespaceBackupRootRefused(linked) {
		t.Fatal("symlinked backup roots not reported as refused")
	}
}
