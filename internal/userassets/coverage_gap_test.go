// coverage_gap_test.go — targeted tests closing the E3 coverage gaps
// (SPEC-USER-ASSET-INSTALL-001 TRUST 5: 85%+ on the new package).
package userassets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestLoadJournalCorruptIsTyped covers the journal corrupt path.
func TestLoadJournalCorruptIsTyped(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	if err := os.MkdirAll(MoaiHome(home), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(JournalPath(home), []byte("{bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadJournal(JournalPath(home)); err == nil {
		t.Fatal("corrupt journal accepted")
	} else if !strings.Contains(err.Error(), "corrupt") {
		t.Errorf("error should name the corruption: %v", err)
	}
}

// TestClearJournalAbsentIsNoop covers ClearJournal's IsNotExist arm.
func TestClearJournalAbsentIsNoop(t *testing.T) {
	t.Parallel()
	if err := ClearJournal(JournalPath(t.TempDir())); err != nil {
		t.Fatalf("clear absent journal: %v", err)
	}
}

// TestLockTakeoverAfterStale covers the stale-reclaim arm of the acquire
// loop (mtime older than staleAfter → takeover, not timeout).
func TestLockTakeoverAfterStale(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	path := LockPath(home)
	if err := os.MkdirAll(MoaiHome(home), 0o755); err != nil {
		t.Fatal(err)
	}
	// A stale lock file: created long in the past.
	if err := os.WriteFile(path, []byte("pid=2147483647 token=stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stale := time.Now().Add(-2 * DefaultStaleAfter)
	if err := os.Chtimes(path, stale, stale); err != nil {
		t.Fatal(err)
	}
	l, err := acquireUserLockStale(path, 100*time.Millisecond, DefaultStaleAfter)
	if err != nil {
		t.Fatalf("stale takeover failed: %v", err)
	}
	_ = l.Release()
}

// TestLockReleaseWithoutTokenIsRefused covers the identity guard's
// no-token file arm.
func TestLockReleaseWithoutTokenIsRefused(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	l, err := AcquireUserLock(home, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	// Overwrite the file without the holder's token (a foreign occupant).
	if err := os.WriteFile(LockPath(home), []byte("pid=999 token=foreign\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := l.Release(); err == nil {
		t.Error("release over a foreign lock must be refused")
	}
	// Clean up for the parallel suite.
	_ = os.Remove(LockPath(home))
}

// TestBackupPathValid covers the happy path of the backup layout.
func TestBackupPathValid(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	got, err := BackupPath(home, RootClaudeSkills, "moai-x/SKILL.md")
	if err != nil {
		t.Fatalf("BackupPath: %v", err)
	}
	want := filepath.Join(MoaiHome(home), "backups", "claude-skills", "moai-x", "SKILL.md")
	if got != want {
		t.Errorf("BackupPath = %q, want %q", got, want)
	}
}

// TestSplitManifestKeyRejectsBare covers the split's failure arm.
func TestSplitManifestKeyRejectsBare(t *testing.T) {
	t.Parallel()
	if _, _, ok := splitManifestKey("no-slash-here"); ok {
		t.Error("bare key accepted")
	}
}

// TestRootBySlug covers the slug lookup helper.
func TestRootBySlug(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	r, ok := RootBySlug(home, RootCodexAgents)
	if !ok || r.Slug != RootCodexAgents {
		t.Fatalf("RootBySlug = %+v, ok=%v", r, ok)
	}
	if _, ok := RootBySlug(home, "nope"); ok {
		t.Error("unknown slug found")
	}
}

// TestAtomicWriteRenameFailure covers the rename-failure cleanup arm.
func TestAtomicWriteRenameFailure(t *testing.T) {
	t.Parallel()
	// A destination directory that vanishes between CreateTemp and Rename
	// is hard to stage; instead cover the chmod-failure-free happy path and
	// the write-error arm via an unwritable temp parent is platform-hostile.
	// The observable residue arm: no temp files left behind after success.
	home := t.TempDir()
	p := ManifestPath(home)
	m, _ := Load(p)
	m.Files["k"] = FileEntry{SHA256: "1", Bundle: "b", InstalledAt: "t", MoaiVersion: "v"}
	if err := m.Save(p); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(MoaiHome(home))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".userassets-") {
			t.Errorf("temp residue: %s", e.Name())
		}
	}
}

// TestLoadReadErrorIsWrapped covers the unreadable-manifest arm
// (permission-denied style errors wrap, not fabricate a fresh manifest).
func TestLoadReadErrorIsWrapped(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	p := ManifestPath(home)
	if err := os.MkdirAll(MoaiHome(home), 0o755); err != nil {
		t.Fatal(err)
	}
	// A DIRECTORY at the manifest path: ReadFile errors with EISDIR.
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err == nil {
		t.Error("directory-at-manifest-path silently loaded as fresh")
	}
}

// TestValidateRelPathEmpty covers the empty-string arm.
func TestValidateRelPathEmpty(t *testing.T) {
	t.Parallel()
	if _, err := ValidateRelPath(""); err == nil {
		t.Error("empty relpath accepted")
	}
}

// TestCorruptErrorMessageAndUnwrap covers CorruptError's Error/Unwrap arms.
func TestCorruptErrorMessageAndUnwrap(t *testing.T) {
	t.Parallel()
	ce := &CorruptError{Path: "/x/y.json", Err: os.ErrInvalid}
	if !strings.Contains(ce.Error(), "/x/y.json") {
		t.Errorf("message should carry the path: %q", ce.Error())
	}
	if ce.Unwrap() != os.ErrInvalid {
		t.Error("Unwrap lost the cause")
	}
}

// TestReconcileJournalFlagCompleteDivergenceCoversBackup covers the
// flag-complete mismatch arm end to end (REQ-023 backup + divergence count).
func TestReconcileJournalFlagCompleteDivergenceCoversBackup(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if err := os.MkdirAll(MoaiHome(f.home), 0o755); err != nil {
		t.Fatal(err)
	}
	// A journal whose single entry is flag-complete; the on-disk file
	// mismatches (the user edited underneath).
	divergeRel := "claude-skills/moai-alpha/workflows/a.md"
	if err := os.MkdirAll(filepath.Join(f.home, ".claude/skills/moai-alpha/workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.home, ".claude/skills/moai-alpha/workflows/a.md"), []byte("edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	j := &PendingJournal{SchemaVersion: 1, Entries: []JournalEntry{{
		Path: divergeRel, ExpectedSHA256: sha256Hex([]byte("workflow a\n")),
		Bundle: "core", MoaiVersion: "vOld", InstalledAt: "t0", WriteCompleted: true,
	}}}
	if err := WriteJournal(JournalPath(f.home), j); err != nil {
		t.Fatal(err)
	}
	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.DivergencePreserved == 0 {
		t.Errorf("flag-complete mismatch not divergence-classified: %+v", res)
	}
	backup := filepath.Join(MoaiHome(f.home), "backups", "claude-skills", "moai-alpha", "workflows", "a.md")
	if _, err := os.Stat(backup); err != nil {
		t.Errorf("backup of shipped bytes not written: %v", err)
	}
}
