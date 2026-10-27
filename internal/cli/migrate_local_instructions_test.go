package cli

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/spf13/cobra"
)

// SPEC-CLAUDE-INSTRUCTION-FILES-001 M2 — the migration verb's direction is
// reversed (AGENTS.* -> CLAUDE.*) and extended to BOTH pairs (shared
// AGENTS.md -> CLAUDE.md and local AGENTS.local.md -> CLAUDE.local.md), with
// the lossless principles of REQ-CIF-005..011: byte-exact backup before the
// destination write, a read->write change guard, a single atomic os.Rename as
// the final step, unconditional preservation of the partial destination on
// every failure exit, and the N12 EEXIST disambiguation (a foreign destination
// is never preserved — only a destination this process created via O_EXCL).

const (
	// File names are spelled as literals here so the tests compile — and stay
	// red — against the pre-flip implementation (E8 discipline).
	testLegacyLocalName  = "AGENTS.local.md"
	testCanonLocalName   = "CLAUDE.local.md"
	testLegacySharedName = "AGENTS.md"
	testCanonSharedName  = "CLAUDE.md"
)

// migrationPairs is the table every direction-sensitive test drives: BOTH
// paths are mandatory (AC-CIF-005; a mutant serving only one pair fails).
var migrationPairs = []struct{ legacy, canonical string }{
	{testLegacyLocalName, testCanonLocalName},
	{testLegacySharedName, testCanonSharedName},
}

// Seam injections (t.Cleanup restores the production default).
func injectMigrateGuardHook(t *testing.T, hook func()) func() {
	t.Helper()
	prev := migrateGuardHook
	migrateGuardHook = hook
	return func() { migrateGuardHook = prev }
}

func injectMigrateRename(t *testing.T, fn func(src, dst string) error) {
	t.Helper()
	prev := migrateRename
	migrateRename = fn
	t.Cleanup(func() { migrateRename = prev })
}

func injectMigrateWriteDest(t *testing.T, fn func(path string, body []byte, perm os.FileMode) error) {
	t.Helper()
	prev := migrateWriteDest
	migrateWriteDest = fn
	t.Cleanup(func() { migrateWriteDest = prev })
}

// runMigrateLocalInstructions invokes the verb against root on a fresh
// command with buffer-backed streams.
func runMigrateLocalInstructions(t *testing.T, root string) (string, string, error) {
	t.Helper()
	var out, errB bytes.Buffer
	cmd := newMigrateLocalInstructionsCmd(func() (string, error) { return root, nil })
	cmd.SetOut(&out)
	cmd.SetErr(&errB)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	return out.String(), errB.String(), err
}

func writeLocalFile(t *testing.T, root, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func localFileDigest(t *testing.T, path string) [32]byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return sha256.Sum256(b)
}

func filePerm(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	return info.Mode().Perm()
}

// backupEntries lists every file under the backup root (relative to root).
func backupEntries(t *testing.T, root string) []string {
	t.Helper()
	var found []string
	base := filepath.Join(root, localInstructionsBackupDir)
	_ = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			rel, rerr := filepath.Rel(root, path)
			if rerr == nil {
				found = append(found, filepath.ToSlash(rel))
			}
		}
		return nil
	})
	return found
}

// latestContentSurvives asserts the lossless invariant for the given bytes:
// the content exists at the canonical path, in a backup entry, or at the
// untouched legacy path (REQ-CIF-010, every exit).
func latestContentSurvives(t *testing.T, root, legacy, canonical string, want [32]byte) {
	t.Helper()
	for _, path := range []string{
		filepath.Join(root, canonical),
		filepath.Join(root, legacy),
	} {
		if _, err := os.Lstat(path); err == nil {
			if got := localFileDigest(t, path); got == want {
				return
			}
		}
	}
	for _, rel := range backupEntries(t, root) {
		if got := localFileDigest(t, filepath.Join(root, filepath.FromSlash(rel))); got == want {
			return
		}
	}
	t.Errorf("lossless invariant violated: latest content of %s survives nowhere (canonical, backup, legacy)", legacy)
}

// AC-CIF-007 (REQ-CIF-007, N8 both-pairs): with both files of a pair present
// the verb refuses — per pair — exiting non-zero, modifying neither file, and
// reporting both paths with the hand-merge instruction.
func TestMigrateLocalInstructions_RefusesCoexistence(t *testing.T) {
	for _, pair := range migrationPairs {
		t.Run(pair.legacy, func(t *testing.T) {
			root := t.TempDir()
			writeLocalFile(t, root, pair.legacy, "legacy side\n")
			writeLocalFile(t, root, pair.canonical, "canonical side\n")
			legacyBefore := localFileDigest(t, filepath.Join(root, pair.legacy))
			canonBefore := localFileDigest(t, filepath.Join(root, pair.canonical))

			_, _, err := runMigrateLocalInstructions(t, root)
			if err == nil {
				t.Fatal("migration with both files present exited 0, want a refusal")
			}
			if !strings.Contains(err.Error(), "both exist") || !strings.Contains(err.Error(), "merge") {
				t.Errorf("refusal reason = %q, want it to name the coexistence and the hand-merge instruction", err.Error())
			}
			for name, before := range map[string][32]byte{pair.legacy: legacyBefore, pair.canonical: canonBefore} {
				if got := localFileDigest(t, filepath.Join(root, name)); got != before {
					t.Errorf("%s was modified by a refused migration", name)
				}
			}
			if entries := backupEntries(t, root); len(entries) != 0 {
				t.Errorf("refused migration wrote %d backup entries, want 0", len(entries))
			}
		})
	}
}

// AC-CIF-005/006: the legacy-only case moves the content byte-for-byte with
// permission bits preserved, keeps a byte-exact backup copy written before the
// destination, and removes the legacy name from the root only through the
// final atomic rename (the original lands in the backup).
func TestMigrateLocalInstructions_MovesWithBackup(t *testing.T) {
	for _, pair := range migrationPairs {
		t.Run(pair.legacy, func(t *testing.T) {
			root := t.TempDir()
			body := "# 사용자 지침\n\n- keep bytes exactly, no trailing newline"
			path := filepath.Join(root, pair.legacy)
			if err := os.WriteFile(path, []byte(body), 0o640); err != nil {
				t.Fatal(err)
			}
			want := sha256.Sum256([]byte(body))

			out, _, err := runMigrateLocalInstructions(t, root)
			if err != nil {
				t.Fatalf("migrate: %v", err)
			}
			if got := localFileDigest(t, filepath.Join(root, pair.canonical)); got != want {
				t.Errorf("%s content differs from the original %s", pair.canonical, pair.legacy)
			}
			if got := filePerm(t, filepath.Join(root, pair.canonical)); got != 0o640 {
				t.Errorf("%s perm = %v, want the original's 0640", pair.canonical, got)
			}
			if _, err := os.Lstat(path); !os.IsNotExist(err) {
				t.Errorf("%s still present at the project root (err=%v) — the final rename must remove it", pair.legacy, err)
			}
			entries := backupEntries(t, root)
			if len(entries) < 2 {
				t.Fatalf("backup entries = %d, want at least 2 (pre-write copy + renamed original): %v", len(entries), entries)
			}
			for _, rel := range entries {
				if got := localFileDigest(t, filepath.Join(root, filepath.FromSlash(rel))); got != want {
					t.Errorf("backup entry %s is not byte-identical to the original", rel)
				}
			}
			if !strings.Contains(out, "Migrated") || !strings.Contains(out, pair.canonical) {
				t.Errorf("stdout = %q, want it to report the migration of %s", out, pair.canonical)
			}
			// Lossless invariant on the success exit.
			latestContentSurvives(t, root, pair.legacy, pair.canonical, want)
		})
	}
}

// AC-CIF-008: re-running after a completed migration is a noop — no write, no
// second backup.
func TestMigrateLocalInstructions_RerunIsNoop(t *testing.T) {
	for _, pair := range migrationPairs {
		t.Run(pair.legacy, func(t *testing.T) {
			root := t.TempDir()
			writeLocalFile(t, root, pair.legacy, "body\n")
			if _, _, err := runMigrateLocalInstructions(t, root); err != nil {
				t.Fatalf("first migrate: %v", err)
			}
			canon := localFileDigest(t, filepath.Join(root, pair.canonical))
			entries := backupEntries(t, root)

			out, _, err := runMigrateLocalInstructions(t, root)
			if err != nil {
				t.Fatalf("second migrate: %v", err)
			}
			if !strings.Contains(out, "Nothing to migrate") {
				t.Errorf("re-run stdout = %q, want a noop message", out)
			}
			if got := localFileDigest(t, filepath.Join(root, pair.canonical)); got != canon {
				t.Errorf("re-run modified %s", pair.canonical)
			}
			if got := backupEntries(t, root); len(got) != len(entries) {
				t.Errorf("backup entries after re-run = %d, want %d (no second backup)", len(got), len(entries))
			}
		})
	}
}

// Neither file present: nothing to do, exit 0, nothing created.
func TestMigrateLocalInstructions_NothingToMigrate(t *testing.T) {
	root := t.TempDir()
	if _, _, err := runMigrateLocalInstructions(t, root); err != nil {
		t.Fatalf("migrate on an empty project: %v", err)
	}
	for _, name := range []string{testLegacyLocalName, testCanonLocalName, testLegacySharedName, testCanonSharedName, ".moai"} {
		if _, err := os.Lstat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Errorf("%s created on a no-op run (err=%v)", name, err)
		}
	}
}

// AC-CIF-009: a legacy file that is not a regular file (symlink — including
// one pointing outside the root, the path-escape representative: the verb
// takes no path argument, so a fixed name that refuses to follow a link is
// the whole escape surface) is refused rather than followed, on both pairs.
func TestMigrateLocalInstructions_RefusesNonRegular(t *testing.T) {
	for _, pair := range migrationPairs {
		t.Run(pair.legacy, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(t.TempDir(), "elsewhere.md")
			if err := os.WriteFile(target, []byte("outside\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, filepath.Join(root, pair.legacy)); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
			_, _, err := runMigrateLocalInstructions(t, root)
			if err == nil {
				t.Fatalf("migration of a symlinked %s exited 0, want a refusal", pair.legacy)
			}
			if _, err := os.Lstat(filepath.Join(root, pair.canonical)); !os.IsNotExist(err) {
				t.Errorf("%s created by a refused migration (err=%v)", pair.canonical, err)
			}
			if _, err := os.Lstat(target); err != nil {
				t.Fatalf("symlink target followed or removed: %v", err)
			}
		})
	}
}

// AC-CIF-009: an unreadable legacy file (mode 0000) is refused without
// touching anything.
func TestMigrateLocalInstructions_RefusesUnreadable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: mode bits do not block reads")
	}
	for _, pair := range migrationPairs {
		t.Run(pair.legacy, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, pair.legacy)
			if err := os.WriteFile(path, []byte("secret\n"), 0o000); err != nil {
				t.Fatal(err)
			}
			_, _, err := runMigrateLocalInstructions(t, root)
			if err == nil {
				t.Fatal("migration of an unreadable file exited 0, want a refusal")
			}
			if _, err := os.Lstat(filepath.Join(root, pair.canonical)); !os.IsNotExist(err) {
				t.Errorf("%s created by a refused migration (err=%v)", pair.canonical, err)
			}
		})
	}
}

// AC-CIF-010 (b-1) (REQ-CIF-010 front guard): when the source's content or
// mtime changes between the read and the destination write, the migration
// aborts before the destination write — the changed original stays untouched
// at the legacy path, no destination exists, and the changed bytes survive.
func TestMigrateLocalInstructions_SourceChangedDuringRead(t *testing.T) {
	for _, pair := range migrationPairs {
		t.Run(pair.legacy, func(t *testing.T) {
			root := t.TempDir()
			writeLocalFile(t, root, pair.legacy, "first read\n")
			hook := func() {
				// Simulates the user editing the source inside the read->write
				// window.
				writeLocalFile(t, root, pair.legacy, "edited during the window\n")
			}
			unlock := injectMigrateGuardHook(t, hook)

			_, _, err := runMigrateLocalInstructions(t, root)
			unlock()
			if err == nil {
				t.Fatal("migration over a concurrently-changed source exited 0, want an abort")
			}
			if !strings.Contains(err.Error(), "changed") {
				t.Errorf("abort reason = %q, want it to name the concurrent change", err.Error())
			}
			if _, lerr := os.Lstat(filepath.Join(root, pair.canonical)); !os.IsNotExist(lerr) {
				t.Errorf("%s written despite the guard (err=%v)", pair.canonical, lerr)
			}
			want := localFileDigest(t, filepath.Join(root, pair.legacy))
			if want != sha256.Sum256([]byte("edited during the window\n")) {
				t.Fatal("fixture setup: the window edit did not land")
			}
			// Lossless invariant on the guard exit: latest content at the
			// untouched legacy path.
			latestContentSurvives(t, root, pair.legacy, pair.canonical, want)
		})
	}
}

// AC-CIF-010 (a)/(d) (REQ-CIF-010/011): a refused final rename leaves the
// original in place and preserves the partial destination — unconditionally,
// whatever its state — as a backup entry, reporting the exact partial outcome
// from the preserved copy; no "Migrated" is reported.
func TestMigrateLocalInstructions_FinalRenameRefusal(t *testing.T) {
	for _, tc := range []struct {
		name    string
		editDst bool
	}{
		{name: "unedited destination"},
		{name: "edited destination", editDst: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, pair := range migrationPairs {
				t.Run(pair.legacy, func(t *testing.T) {
					root := t.TempDir()
					writeLocalFile(t, root, pair.legacy, "legacy body\n")
					legacyWant := localFileDigest(t, filepath.Join(root, pair.legacy))

					var dstWant [32]byte
					injectMigrateRename(t, func(src, dst string) error {
						if !strings.HasSuffix(filepath.ToSlash(src), pair.legacy) {
							// The preserve rename (destination -> backup entry)
							// must RUN; only the FINAL step (legacy -> backup)
							// is refused.
							return os.Rename(src, dst)
						}
						if tc.editDst {
							// The user edits the destination between its
							// creation and the refused final step.
							writeLocalFile(t, root, pair.canonical, "user edited the destination\n")
						}
						dstWant = localFileDigest(t, filepath.Join(root, pair.canonical))
						return &os.LinkError{Op: "rename", Old: src, New: dst, Err: syscall.EXDEV}
					})

					out, _, err := runMigrateLocalInstructions(t, root)
					if err == nil {
						t.Fatal("final-rename refusal exited 0, want a reported failure")
					}
					if strings.Contains(out, "Migrated") {
						t.Errorf("stdout reports a migration that did not complete: %q", out)
					}
					// The original stays in place, byte-identical.
					if got := localFileDigest(t, filepath.Join(root, pair.legacy)); got != legacyWant {
						t.Error("the original was moved or modified by a failed final step")
					}
					// The destination (this process's O_EXCL creation) is
					// preserved into a backup entry — state-independently; the
					// hash of whatever state it was in survives.
					var preserved bool
					for _, rel := range backupEntries(t, root) {
						if got := localFileDigest(t, filepath.Join(root, filepath.FromSlash(rel))); got == dstWant {
							preserved = true
						}
					}
					if !preserved {
						t.Errorf("partial destination not preserved into a backup entry (want hash of its exact state)")
					}
					if _, lerr := os.Lstat(filepath.Join(root, pair.canonical)); !os.IsNotExist(lerr) {
						t.Errorf("%s left at the root after rollback (err=%v) — no half state", pair.canonical, lerr)
					}
					// Lossless invariant on this exit: original at the legacy
					// path, edited destination in the backup.
					latestContentSurvives(t, root, pair.legacy, pair.canonical, legacyWant)
					latestContentSurvives(t, root, pair.legacy, pair.canonical, dstWant)
				})
			}
		})
	}
}

// AC-CIF-010 (c) (REQ-CIF-011, N9): a failed destination write leaves the
// original untouched and preserves the partial destination into a backup
// entry (hash-compared) — a mutant that preserves on rename failure but
// deletes on write failure fails here.
func TestMigrateLocalInstructions_DestinationWriteFailure(t *testing.T) {
	for _, pair := range migrationPairs {
		t.Run(pair.legacy, func(t *testing.T) {
			root := t.TempDir()
			writeLocalFile(t, root, pair.legacy, "legacy body\n")
			legacyWant := localFileDigest(t, filepath.Join(root, pair.legacy))

			var partialWant [32]byte
			injectMigrateWriteDest(t, func(path string, body []byte, perm os.FileMode) error {
				// Simulates a mid-write failure that leaves a partial file.
				if err := os.WriteFile(path, []byte("partial bytes that made it to disk"), perm); err != nil {
					t.Fatal(err)
				}
				partialWant = localFileDigest(t, path)
				return errors.New("simulated disk-full mid-write")
			})

			out, _, err := runMigrateLocalInstructions(t, root)
			if err == nil {
				t.Fatal("destination-write failure exited 0, want a reported failure")
			}
			if strings.Contains(out, "Migrated") {
				t.Errorf("stdout reports a migration that did not complete: %q", out)
			}
			if got := localFileDigest(t, filepath.Join(root, pair.legacy)); got != legacyWant {
				t.Error("the original was modified by a failed destination write")
			}
			var preserved bool
			for _, rel := range backupEntries(t, root) {
				if got := localFileDigest(t, filepath.Join(root, filepath.FromSlash(rel))); got == partialWant {
					preserved = true
				}
			}
			if !preserved {
				t.Error("partial destination not preserved into a backup entry (hash mismatch everywhere)")
			}
			if _, lerr := os.Lstat(filepath.Join(root, pair.canonical)); !os.IsNotExist(lerr) {
				t.Errorf("%s left at the root after the write failure (err=%v) — no half state", pair.canonical, lerr)
			}
			latestContentSurvives(t, root, pair.legacy, pair.canonical, legacyWant)
		})
	}
}

// AC-CIF-010 race row (N12): an O_EXCL EEXIST after the presence check raced
// means ANOTHER process created the destination — that file is not "the
// partial destination". The verb leaves both files in place and refuses; the
// foreign destination is never renamed into the backup.
func TestMigrateLocalInstructions_EEXISTForeignDestination(t *testing.T) {
	for _, pair := range migrationPairs {
		t.Run(pair.legacy, func(t *testing.T) {
			root := t.TempDir()
			writeLocalFile(t, root, pair.legacy, "legacy body\n")
			legacyWant := localFileDigest(t, filepath.Join(root, pair.legacy))
			foreign := sha256.Sum256([]byte("written by another process\n"))

			injectMigrateWriteDest(t, func(path string, body []byte, perm os.FileMode) error {
				// The race: another process wins the destination name between
				// the presence check and this process's O_EXCL open.
				if err := os.WriteFile(path, []byte("written by another process\n"), perm); err != nil {
					t.Fatal(err)
				}
				return os.ErrExist
			})

			_, _, err := runMigrateLocalInstructions(t, root)
			if err == nil {
				t.Fatal("EEXIST race exited 0, want a refusal")
			}
			if !strings.Contains(err.Error(), "exists") {
				t.Errorf("refusal = %q, want it to name the existing destination", err.Error())
			}
			// Both files stay in place, byte-identical — the foreign
			// destination is NOT preserved into the backup.
			if got := localFileDigest(t, filepath.Join(root, pair.legacy)); got != legacyWant {
				t.Error("the original was moved or modified by the EEXIST refusal")
			}
			if got := localFileDigest(t, filepath.Join(root, pair.canonical)); got != foreign {
				t.Error("the foreign destination was removed or replaced by the refusal")
			}
			for _, rel := range backupEntries(t, root) {
				if got := localFileDigest(t, filepath.Join(root, filepath.FromSlash(rel))); got == foreign {
					t.Errorf("foreign destination renamed into the backup (%s) — only this process's own O_EXCL creation is ever preserved", rel)
				}
			}
			latestContentSurvives(t, root, pair.legacy, pair.canonical, legacyWant)
		})
	}
}

// AC-CIF-010 (e) (REQ-CIF-011 rollback-failure exit): when both the final
// rename and the rollback rename fail (backup directory not writable), the
// verb leaves the original and the partial destination untouched in place,
// exits non-zero, and reports the residual paths and the failed preservation.
func TestMigrateLocalInstructions_RollbackRenameFailure(t *testing.T) {
	for _, pair := range migrationPairs {
		t.Run(pair.legacy, func(t *testing.T) {
			root := t.TempDir()
			writeLocalFile(t, root, pair.legacy, "legacy body\n")
			legacyWant := localFileDigest(t, filepath.Join(root, pair.legacy))

			base := filepath.Join(root, localInstructionsBackupDir)
			injectMigrateRename(t, func(src, dst string) error {
				// Both renames into the backup fail: the directory is not
				// writable for this process (EACCES shape).
				if err := os.Chmod(base, 0o500); err != nil {
					t.Fatal(err)
				}
				return &os.LinkError{Op: "rename", Old: src, New: dst, Err: syscall.EACCES}
			})
			t.Cleanup(func() { _ = os.Chmod(base, 0o755) })

			_, _, err := runMigrateLocalInstructions(t, root)
			if err == nil {
				t.Fatal("rollback-failure exit was 0, want non-zero")
			}
			// Both residual files stay in place.
			if got := localFileDigest(t, filepath.Join(root, pair.legacy)); got != legacyWant {
				t.Error("the original was moved despite the failed preservation")
			}
			if _, lerr := os.Lstat(filepath.Join(root, pair.canonical)); lerr != nil {
				t.Fatalf("partial destination removed despite the failed rollback: %v", lerr)
			}
			if !strings.Contains(err.Error(), pair.legacy) || !strings.Contains(err.Error(), pair.canonical) {
				t.Errorf("error = %q, want it to report both residual paths", err.Error())
			}
			if !strings.Contains(err.Error(), "preserv") {
				t.Errorf("error = %q, want it to report the failed preservation", err.Error())
			}
			latestContentSurvives(t, root, pair.legacy, pair.canonical, legacyWant)
		})
	}
}

// The verb is reachable as `moai migrate local-instructions`.
func TestMigrateLocalInstructions_RegisteredUnderMigrate(t *testing.T) {
	var found *cobra.Command
	for _, c := range migrateCmd.Commands() {
		if c.Name() == "local-instructions" {
			found = c
		}
	}
	if found == nil {
		t.Fatal("`moai migrate local-instructions` is not registered")
	}
}
