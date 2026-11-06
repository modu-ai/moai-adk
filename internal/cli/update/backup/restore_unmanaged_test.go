package backup

import (
	"os"
	"path/filepath"
	"testing"
)

// writeUnmanagedFile creates one file with the given content, making parents.
func writeUnmanagedFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readUnmanagedFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// A file the new template does not ship is restored as it was in the backup.
func TestRestoreUnmanagedConfigDirs_RestoresMissingFile(t *testing.T) {
	root := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "20260101_000000")
	writeUnmanagedFile(t, filepath.Join(backupDir, "evaluator-profiles", "custom.md"), "my profile")

	refs, err := RestoreUnmanagedConfigDirs(root, backupDir)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if got := readUnmanagedFile(t, filepath.Join(root, ".moai", "config", "evaluator-profiles", "custom.md")); got != "my profile" {
		t.Fatalf("restored content = %q, want %q", got, "my profile")
	}
	if len(refs) != 1 || refs[0].Action != "restored" {
		t.Fatalf("refs = %+v, want one restored entry", refs)
	}
}

// When the working file differs, the new template stays in place and the backup
// copy is written beside it, so nothing the user wrote is lost and no template
// update is reverted.
func TestRestoreUnmanagedConfigDirs_KeepsNewTemplateAndWritesSideCopy(t *testing.T) {
	root := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "20260101_000000")
	writeUnmanagedFile(t, filepath.Join(backupDir, "astgrep-rules", "go", "idioms.yml"), "user rule")
	working := filepath.Join(root, ".moai", "config", "astgrep-rules", "go", "idioms.yml")
	writeUnmanagedFile(t, working, "new template rule")

	refs, err := RestoreUnmanagedConfigDirs(root, backupDir)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if got := readUnmanagedFile(t, working); got != "new template rule" {
		t.Fatalf("working file = %q; the new template must stay", got)
	}
	side := working + ".from-20260101_000000"
	if got := readUnmanagedFile(t, side); got != "user rule" {
		t.Fatalf("side copy = %q, want the backup content", got)
	}
	if len(refs) != 1 || refs[0].Action != "side-copy" {
		t.Fatalf("refs = %+v, want one side-copy entry", refs)
	}
}

// Identical content is a no-op: no side copy and no report.
func TestRestoreUnmanagedConfigDirs_IdenticalFileIsNoOp(t *testing.T) {
	root := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "20260101_000000")
	writeUnmanagedFile(t, filepath.Join(backupDir, "evaluator-profiles", "strict.md"), "same")
	working := filepath.Join(root, ".moai", "config", "evaluator-profiles", "strict.md")
	writeUnmanagedFile(t, working, "same")

	refs, err := RestoreUnmanagedConfigDirs(root, backupDir)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if len(refs) != 0 {
		t.Fatalf("refs = %+v, want none for identical content", refs)
	}
	if _, err := os.Stat(working + ".from-20260101_000000"); !os.IsNotExist(err) {
		t.Fatalf("side copy written for identical content: err=%v", err)
	}
}

// A second run over the same backup finds its side copy already in place and
// reports nothing new, so repeating the restore never duplicates or clobbers.
func TestRestoreUnmanagedConfigDirs_SecondRunIsIdempotent(t *testing.T) {
	root := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "20260101_000000")
	writeUnmanagedFile(t, filepath.Join(backupDir, "astgrep-rules", "go", "idioms.yml"), "user rule")
	working := filepath.Join(root, ".moai", "config", "astgrep-rules", "go", "idioms.yml")
	writeUnmanagedFile(t, working, "new template rule")

	if _, err := RestoreUnmanagedConfigDirs(root, backupDir); err != nil {
		t.Fatalf("first restore: %v", err)
	}
	refs, err := RestoreUnmanagedConfigDirs(root, backupDir)
	if err != nil {
		t.Fatalf("second restore: %v", err)
	}
	if len(refs) != 0 {
		t.Fatalf("second run reported %+v, want none", refs)
	}
	if got := readUnmanagedFile(t, working+".from-20260101_000000"); got != "user rule" {
		t.Fatalf("side copy changed on second run: %q", got)
	}
}

// A target whose parent is a symbolic link out of the config directory is refused.
// The refusal is reported as an entry rather than dropped, so an explicit --restore
// cannot read as a success for a file it did not write.
func TestRestoreUnmanagedConfigDirs_RefusedTargetIsReported(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "20260101_000000")
	writeUnmanagedFile(t, filepath.Join(backupDir, "evaluator-profiles", "sub", "x.md"), "backup body")
	profiles := filepath.Join(root, ".moai", "config", "evaluator-profiles")
	if err := os.MkdirAll(profiles, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(profiles, "sub")); err != nil {
		t.Skip(err)
	}

	refs, err := RestoreUnmanagedConfigDirs(root, backupDir)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "x.md")); !os.IsNotExist(err) {
		t.Fatalf("file written through a symbolic link out of the config directory: err=%v", err)
	}
	if len(refs) != 1 || refs[0].Action != "target-refused" {
		t.Fatalf("refs = %+v, want one target-refused entry", refs)
	}
	if want := ".moai/config/evaluator-profiles/sub/x.md"; refs[0].Path != want {
		t.Fatalf("refused path = %q, want %q", refs[0].Path, want)
	}
}
