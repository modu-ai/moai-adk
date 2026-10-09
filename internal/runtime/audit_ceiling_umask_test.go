//go:build darwin || linux

package runtime

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// TestAppendProgressRecordNewFileModeAppliesUmask (sync-audit-1 F2) — a
// brand-new progress.md takes the mode the pre-repair os.WriteFile gave it:
// 0644 with the process umask applied (0600 under umask 0077). The atomic
// replace path applies the destination mode with an explicit chmod, which
// bypasses the umask; a new file must reach the replace already carrying the
// umask-adjusted mode. unix-only: syscall.Umask has no Windows equivalent.
func TestAppendProgressRecordNewFileModeAppliesUmask(t *testing.T) {
	old := syscall.Umask(0o077)
	defer func() { syscall.Umask(old) }()
	specDir := t.TempDir()
	if err := appendProgressRecord(specDir, "- first record"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(specDir, "progress.md"))
	if err != nil {
		t.Fatal(err)
	}
	want := os.FileMode(0o644) &^ os.FileMode(0o077)
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("new progress.md mode %04o, want %04o (0644 with the umask applied — the pre-repair os.WriteFile semantics)", got, want)
	}
}

// TestAppendProgressRecordExistingFileModePreserved (sync-audit-1 F2, the
// existing-file arm) — an existing progress.md keeps its pre-write
// permission bits through the atomic replace: a 0600 file is still 0600
// after a record lands, regardless of the umask in effect.
func TestAppendProgressRecordExistingFileModePreserved(t *testing.T) {
	old := syscall.Umask(0o077)
	defer func() { syscall.Umask(old) }()
	specDir := t.TempDir()
	path := filepath.Join(specDir, "progress.md")
	if err := os.WriteFile(path, []byte("# progress\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := appendProgressRecord(specDir, "- second record"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("existing progress.md mode %04o, want 0600 preserved through the replace", got)
	}
}
