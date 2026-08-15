package atomicfile

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// countOrphanTemps counts leftover temp files matching the helper's pattern in
// dir. AC-CAW-006 requires this to be zero on both success and error paths.
func countOrphanTemps(t *testing.T, dir string) int {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, tempPattern))
	if err != nil {
		t.Fatalf("glob temp pattern: %v", err)
	}
	return len(matches)
}

// AC-CAW-001 — Atomicity: the rename completes and the new content fully lands,
// replacing the prior content with no partial write.
func TestWrite_ReplacesContentAtomically(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target.yaml")

	if err := os.WriteFile(target, []byte("prior: content\n"), defs.FilePerm); err != nil {
		t.Fatalf("seed prior file: %v", err)
	}

	if err := Write(target, []byte("new: content\n"), defs.FilePerm); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "new: content\n" {
		t.Fatalf("content = %q, want %q", got, "new: content\n")
	}
	if countOrphanTemps(t, dir) != 0 {
		t.Fatalf("orphan temp survived success path: %d", countOrphanTemps(t, dir))
	}
}

// AC-CAW-002 — Mode-preservation round-trip. The destination's pre-existing
// mode MUST survive the write unchanged (not narrowed to 0600, not widened to
// 0644). Table-driven over the three modes acceptance.md requires.
func TestWrite_PreservesExistingMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX mode bits are not stored on Windows; os.Chmod only toggles the read-only flag, so the seeded mode drifts before Write is called")
	}
	t.Parallel()

	cases := []struct {
		name string
		mode os.FileMode
	}{
		{"0600", 0o600},
		{"0644", 0o644},
		{"0750", 0o750},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			target := filepath.Join(dir, "section.yaml")

			if err := os.WriteFile(target, []byte("old"), tc.mode); err != nil {
				t.Fatalf("seed at mode %v: %v", tc.mode, err)
			}
			// Sanity-check the seed landed at the mode we asked for (umask can
			// interfere on some systems; a failure here means the fixture is
			// wrong, not the helper).
			if info, err := os.Stat(target); err != nil {
				t.Fatalf("stat seed: %v", err)
			} else if got := info.Mode().Perm(); got != tc.mode {
				t.Fatalf("seed mode drifted to %v before Write was called", got)
			}

			if err := Write(target, []byte("new"), defs.FilePerm); err != nil {
				t.Fatalf("Write: %v", err)
			}

			info, err := os.Stat(target)
			if err != nil {
				t.Fatalf("stat after Write: %v", err)
			}
			if got := info.Mode().Perm(); got != tc.mode {
				t.Fatalf("mode = %v, want preserved %v", got, tc.mode)
			}
		})
	}
}

// AC-CAW-003 — Destination-not-present default mode. When the file does not yet
// exist, the helper creates it at the caller-supplied defaultMode (or the
// canonical default defs.FilePerm) — never at os.CreateTemp's 0600.
func TestWrite_NewFileUsesDefaultMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX mode bits are not stored on Windows; a fresh file reads back as 0666 regardless of the requested mode")
	}
	t.Parallel()

	t.Run("defs.FilePerm default", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		target := filepath.Join(dir, "fresh.yaml")

		if err := Write(target, []byte("body"), defs.FilePerm); err != nil {
			t.Fatalf("Write: %v", err)
		}

		info, err := os.Stat(target)
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if got := info.Mode().Perm(); got != defs.FilePerm {
			t.Fatalf("new-file mode = %v, want defs.FilePerm %v", got, defs.FilePerm)
		}
	})

	t.Run("caller-declared secret 0o600 wins", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		target := filepath.Join(dir, "secret.yaml")
		const secretMode os.FileMode = 0o600

		if err := Write(target, []byte("body"), secretMode); err != nil {
			t.Fatalf("Write: %v", err)
		}

		info, err := os.Stat(target)
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if got := info.Mode().Perm(); got != secretMode {
			t.Fatalf("new-file mode = %v, want caller default %v", got, secretMode)
		}
	})
}

// AC-CAW-006 — Temp cleanup on the error path. When the write fails after the
// temp file is created but before rename completes, the defer os.Remove must
// clean up the orphan. We force a rename failure by making the target an
// existing non-empty directory: renaming a file onto a non-empty directory
// fails on both darwin and linux, after the temp has already been created and
// chmod'd in the same directory.
func TestWrite_CleansUpTempOnError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Target is a non-empty directory: CreateTemp in `dir` succeeds, the
	// rename of the temp onto this directory path fails.
	target := filepath.Join(dir, "target-blocker")
	if err := os.MkdirAll(target, defs.DirPerm); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "occupant"), []byte("x"), defs.FilePerm); err != nil {
		t.Fatalf("seed occupant: %v", err)
	}

	err := Write(target, []byte("new"), defs.FilePerm)
	if err == nil {
		t.Fatal("Write unexpectedly succeeded against a directory target; test setup is invalid")
	}

	if n := countOrphanTemps(t, dir); n != 0 {
		t.Fatalf("orphan temp survived error path: %d (err was: %v)", n, err)
	}
}

// AC-CAW-001 (error leg) — when the temp file cannot be created (parent dir
// missing), Write surfaces a wrapped error and creates nothing.
func TestWrite_FailsWhenParentDirMissing(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "nonexistent-subdir", "target.yaml")

	err := Write(target, []byte("body"), defs.FilePerm)
	if err == nil {
		t.Fatal("Write unexpectedly succeeded with a missing parent dir")
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Fatalf("target should not exist after failed Write; statErr=%v", statErr)
	}
}

// Stat-error branch: when os.Stat fails with a non-IsNotExist error (here a
// permission error from a non-searchable parent dir), the error is wrapped and
// returned before any temp is created. Unix-only: Windows ignores Unix perm
// bits, so the test is skipped there.
func TestWrite_FailsOnStatPermissionError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits are not honored on Windows")
	}
	dir := t.TempDir()
	// Create a subdir, make it non-searchable so stat of a path inside it
	// fails with EACCES rather than ENOENT.
	blocked := filepath.Join(dir, "blocked")
	if err := os.MkdirAll(blocked, defs.DirPerm); err != nil {
		t.Fatalf("mkdir blocked: %v", err)
	}
	target := filepath.Join(blocked, "target.yaml")
	if err := os.Chmod(blocked, 0o000); err != nil {
		t.Fatalf("chmod blocked: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(blocked, defs.DirPerm) })

	err := Write(target, []byte("body"), defs.FilePerm)
	if err == nil {
		t.Fatal("Write unexpectedly succeeded against a non-searchable parent dir")
	}
}
