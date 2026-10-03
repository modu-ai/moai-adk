//go:build !windows

package harness

// Tests for the ownership restriction on the prune state path (SPEC-HARNESS-
// RETENTION-HARDEN-001, AC-HRH-004, AC-HRH-005 and AC-HRH-006). The pruner
// replaces a faulty state-path entry only when the current user owns it. A
// non-root user cannot create a foreign-owned file, so the refusal branch is
// reached through the owner-check field; the real check is pinned separately.
// The file reads POSIX owner ids and therefore carries the !windows constraint.

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// captureStderr runs fn with os.Stderr redirected into a pipe and returns what fn wrote to it.
// Callers are serial tests: swapping os.Stderr cannot overlap the package's parallel tests.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	old := os.Stderr
	os.Stderr = w
	func() {
		defer func() { os.Stderr = old }()
		fn()
	}()
	_ = w.Close()
	out, _ := io.ReadAll(r)
	_ = r.Close()
	return string(out)
}

// TestPruneStateForeignOwnedLeftUntouchedAndWarns: when the owner check reports "not owned", a
// symbolic link and an unwritable regular file at the state path are left byte-identical, the prune
// is skipped with an error naming the path, and exactly one warning line is written to stderr.
func TestPruneStateForeignOwnedLeftUntouchedAndWarns(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, statePath, victim string)
	}{
		{"symlink", func(t *testing.T, statePath, victim string) {
			if err := os.Symlink(victim, statePath); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
		}},
		{"unwritable-file", func(t *testing.T, statePath, _ string) {
			if os.Geteuid() == 0 {
				t.Skip("root bypasses file permissions")
			}
			if err := os.WriteFile(statePath, []byte("2026-09-01T00:00:00Z"), 0o400); err != nil {
				t.Fatalf("write state: %v", err)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			logPath := filepath.Join(dir, "usage-log.jsonl")
			archiveDir := filepath.Join(dir, "archive")
			now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
			writeStaleLog(t, logPath, now, "stale-1")
			victim := filepath.Join(dir, "victim.txt")
			victimBytes := bytes.Repeat([]byte("V"), 64)
			if err := os.WriteFile(victim, victimBytes, 0o644); err != nil {
				t.Fatalf("write victim: %v", err)
			}
			statePath := logPath + stampSuffix
			tc.setup(t, statePath, victim)

			logBefore, _ := os.ReadFile(logPath)
			stateBefore, err := os.Lstat(statePath)
			if err != nil {
				t.Fatalf("lstat state: %v", err)
			}
			linkBefore, _ := os.Readlink(statePath)
			contentBefore, _ := os.ReadFile(statePath)

			ret := NewRetention(logPath, archiveDir, func() time.Time { return now })
			ret.ownerCheck = func(string) bool { return false }
			var perr error
			stderr := captureStderr(t, func() { perr = ret.PruneStaleEntries(30) })

			if perr == nil || !strings.Contains(perr.Error(), statePath) {
				t.Errorf("want an error naming %s, got %v", statePath, perr)
			}
			if logAfter, _ := os.ReadFile(logPath); !bytes.Equal(logBefore, logAfter) {
				t.Errorf("log changed")
			}
			if _, serr := os.Stat(archiveDir); !os.IsNotExist(serr) {
				t.Errorf("archive directory exists: %v", serr)
			}
			stateAfter, lerr := os.Lstat(statePath)
			if lerr != nil || !os.SameFile(stateBefore, stateAfter) || stateAfter.Mode() != stateBefore.Mode() {
				t.Errorf("state-path entry changed: err=%v", lerr)
			}
			if linkAfter, _ := os.Readlink(statePath); linkAfter != linkBefore {
				t.Errorf("link text changed: %q -> %q", linkBefore, linkAfter)
			}
			if contentAfter, _ := os.ReadFile(statePath); !bytes.Equal(contentBefore, contentAfter) {
				t.Errorf("state-path content changed")
			}
			if got, _ := os.ReadFile(victim); !bytes.Equal(got, victimBytes) {
				t.Errorf("victim changed: %q", got)
			}
			if strings.Count(stderr, "\n") != 1 || !strings.HasPrefix(stderr, "[WARN] harness/retention:") || !strings.Contains(stderr, statePath) {
				t.Errorf("want exactly one warning line naming %s, got %q", statePath, stderr)
			}
		})
	}
}

// TestOwnerCheckDefault: the production owner check answers for a file the user owns, for the root
// directory, for a user-owned link to the root directory (the link's own owner decides, not its
// target's), and for a link owned by someone else; the constructor installs it.
func TestOwnerCheckDefault(t *testing.T) {
	dir := t.TempDir()
	own := filepath.Join(dir, "own.txt")
	if err := os.WriteFile(own, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	ret := NewRetention(filepath.Join(dir, "usage-log.jsonl"), filepath.Join(dir, "archive"), nil)
	if !entryOwnedByCurrentUser(own) || !ret.ownerCheck(own) {
		t.Errorf("a file the test created must count as owned")
	}

	if os.Geteuid() == 0 {
		t.Skip("root owns everything the steps below probe")
	}
	if entryOwnedByCurrentUser("/") {
		t.Errorf("/ must not count as owned for a non-root user")
	}
	if ret.ownerCheck("/") {
		t.Errorf("the constructor must install the real owner check: / counted as owned")
	}

	link := filepath.Join(dir, "link-to-root")
	if err := os.Symlink("/", link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if !entryOwnedByCurrentUser(link) {
		t.Errorf("a user-owned link to / must count as owned: the link's own owner decides, not its target's")
	}

	foreign := ""
	for _, c := range []string{"/var", "/tmp", "/etc", "/bin"} {
		fi, err := os.Lstat(c)
		if err != nil || fi.Mode()&os.ModeSymlink == 0 {
			continue
		}
		if st, ok := fi.Sys().(*syscall.Stat_t); ok && int64(st.Uid) != int64(os.Geteuid()) {
			foreign = c
			break
		}
	}
	if foreign == "" {
		t.Skip("no foreign-owned symbolic link among /var, /tmp, /etc, /bin on this platform")
	}
	if entryOwnedByCurrentUser(foreign) {
		t.Errorf("%s is a symbolic link owned by another user and must not count as owned", foreign)
	}
}

// TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal drives the whole path (prune, open of the
// state file, heal) and pins that the heal removes the inspected entry only through the conditional
// removal. The owner check runs inside the heal, before the removal, so the stand-in check renames a
// fresh, healthy state file over the inspected link, as a concurrent healer would, and reports "owned".
// The fresh file carries a still-fresh stamp, so the pruner adopts it and skips the prune; had the heal
// removed the entry unconditionally, the fresh file would be gone and the pruner would create and stamp
// a new one instead.
func TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	logBefore, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	victim := filepath.Join(dir, "victim.txt")
	if err := os.WriteFile(victim, []byte("v"), 0o644); err != nil {
		t.Fatalf("write victim: %v", err)
	}
	statePath := logPath + pruneStateSuffix
	if err := os.Symlink(victim, statePath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	freshBytes := []byte(now.Add(-time.Second).UTC().Format(time.RFC3339Nano))
	var freshInfo os.FileInfo
	ret := NewRetention(logPath, archiveDir, func() time.Time { return now })
	ret.ownerCheck = func(path string) bool {
		if path != statePath {
			return true
		}
		fresh := filepath.Join(dir, "fresh.tmp")
		if err := os.WriteFile(fresh, freshBytes, 0o644); err != nil {
			t.Errorf("write fresh: %v", err)
			return true
		}
		if err := os.Rename(fresh, path); err != nil {
			t.Errorf("rename fresh: %v", err)
			return true
		}
		if freshInfo, err = os.Lstat(path); err != nil {
			t.Errorf("lstat fresh: %v", err)
		}
		return true
	}

	if perr := ret.PruneStaleEntries(30); perr != nil {
		t.Errorf("prune returned %v, want nil", perr)
	}
	if freshInfo == nil {
		t.Fatalf("the owner check never ran: the heal was not reached")
	}
	after, lerr := os.Lstat(statePath)
	if lerr != nil || !os.SameFile(freshInfo, after) {
		t.Fatalf("the fresh state file was removed or replaced by the heal: err=%v", lerr)
	}
	if got, _ := os.ReadFile(statePath); !bytes.Equal(got, freshBytes) {
		t.Errorf("fresh state file content = %q, want %q", got, freshBytes)
	}
	if logAfter, _ := os.ReadFile(logPath); !bytes.Equal(logBefore, logAfter) {
		t.Errorf("the log changed although the adopted stamp was fresh")
	}
}

// TestHealDoesNotRemoveAFreshStateFile: a state file a concurrent healer renamed over the inspected
// entry is never removed by this pruner's removal step; with the entry unchanged the step removes it.
func TestHealDoesNotRemoveAFreshStateFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	victim := filepath.Join(dir, "victim.txt")
	if err := os.WriteFile(victim, []byte("v"), 0o644); err != nil {
		t.Fatalf("write victim: %v", err)
	}
	statePath := filepath.Join(dir, "usage-log.jsonl"+pruneStateSuffix)
	if err := os.Symlink(victim, statePath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	inspected, err := os.Lstat(statePath)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}

	// Stand-in healer: creates a healthy regular state file while the link is still alive, then
	// renames it over the inspected link, so the two identities cannot coincide.
	fresh := filepath.Join(dir, "fresh.tmp")
	freshBytes := []byte("2026-10-02T00:00:00Z")
	if err := os.WriteFile(fresh, freshBytes, 0o644); err != nil {
		t.Fatalf("write fresh: %v", err)
	}
	if err := os.Rename(fresh, statePath); err != nil {
		t.Fatalf("rename fresh: %v", err)
	}
	freshInfo, err := os.Lstat(statePath)
	if err != nil {
		t.Fatalf("lstat fresh: %v", err)
	}

	removed, rerr := removeStateEntryIfUnchanged(statePath, inspected)
	if rerr != nil || removed {
		t.Errorf("removal of a changed entry: removed=%v err=%v, want false and nil", removed, rerr)
	}
	after, lerr := os.Lstat(statePath)
	if lerr != nil || !os.SameFile(freshInfo, after) {
		t.Fatalf("the fresh state file was removed or replaced: err=%v", lerr)
	}
	if got, _ := os.ReadFile(statePath); !bytes.Equal(got, freshBytes) {
		t.Errorf("fresh state file content changed: %q", got)
	}

	removed, rerr = removeStateEntryIfUnchanged(statePath, freshInfo)
	if rerr != nil || !removed {
		t.Errorf("removal of an unchanged entry: removed=%v err=%v, want true and nil", removed, rerr)
	}
	if _, serr := os.Lstat(statePath); !os.IsNotExist(serr) {
		t.Errorf("unchanged entry still present: %v", serr)
	}
}

// N1: an entry whose type and identity differ from the inspected one but whose modification time is
// equal must not be removed (the identity, type and mode comparisons, not the mtime alone, decide).
func TestHealRemovalIsNotDecidedByModTimeAlone(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	victim := filepath.Join(dir, "victim.txt")
	if err := os.WriteFile(victim, []byte("v"), 0o644); err != nil {
		t.Fatalf("write victim: %v", err)
	}
	statePath := filepath.Join(dir, "usage-log.jsonl"+stampSuffix)
	if err := os.Symlink(victim, statePath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	inspected, err := os.Lstat(statePath)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	fresh := filepath.Join(dir, "fresh.tmp")
	freshBytes := []byte("2026-10-02T00:00:00Z")
	if err := os.WriteFile(fresh, freshBytes, 0o644); err != nil {
		t.Fatalf("write fresh: %v", err)
	}
	// A coarse-timestamp file system gives the swapped-in file the inspected entry's modification time.
	if err := os.Chtimes(fresh, inspected.ModTime(), inspected.ModTime()); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
	if err := os.Rename(fresh, statePath); err != nil {
		t.Fatalf("rename fresh: %v", err)
	}
	freshInfo, err := os.Lstat(statePath)
	if err != nil {
		t.Fatalf("lstat fresh: %v", err)
	}
	if !freshInfo.ModTime().Equal(inspected.ModTime()) {
		t.Fatalf("precondition: the fresh file must carry the inspected modification time: %v vs %v", freshInfo.ModTime(), inspected.ModTime())
	}
	removed, rerr := removeStateEntryIfUnchanged(statePath, inspected)
	if rerr != nil || removed {
		t.Errorf("removal of a changed entry with an equal modification time: removed=%v err=%v, want false and nil", removed, rerr)
	}
	if got, _ := os.ReadFile(statePath); !bytes.Equal(got, freshBytes) {
		t.Errorf("fresh state file was removed or changed: %q", got)
	}
}

// N4: the file arm of the heal (an unwritable regular file) removes the entry only through the
// conditional removal; the owner-check stand-in renames a fresh, still-fresh-stamped file over the
// inspected entry, as a concurrent healer would, and reports "owned".
func TestPruneHealFileArmKeepsAFreshStateFileSwappedInDuringTheHeal(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses file permissions")
	}
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	logBefore, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	statePath := logPath + stampSuffix
	if err := os.WriteFile(statePath, []byte("2026-09-01T00:00:00Z"), 0o400); err != nil {
		t.Fatalf("write state: %v", err)
	}
	freshBytes := []byte(now.Add(-time.Second).UTC().Format(time.RFC3339Nano))
	var freshInfo os.FileInfo
	ret := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now })
	ret.ownerCheck = func(path string) bool {
		// Path guard (amendment 0.4.1, plan.md M9 and B11): once the heal lock exists the owner check is
		// asked about the heal-lock path too; the stand-in acts only on the state path and answers
		// "owned" for any other path without side effects, otherwise it would swap the heal-lock path
		// and overwrite freshInfo.
		if path != statePath {
			return true
		}
		fresh := filepath.Join(dir, "fresh.tmp")
		if err := os.WriteFile(fresh, freshBytes, 0o644); err != nil {
			t.Errorf("write fresh: %v", err)
			return true
		}
		if err := os.Rename(fresh, path); err != nil {
			t.Errorf("rename fresh: %v", err)
			return true
		}
		if freshInfo, err = os.Lstat(path); err != nil {
			t.Errorf("lstat fresh: %v", err)
		}
		return true
	}
	if perr := ret.PruneStaleEntries(30); perr != nil {
		t.Errorf("prune returned %v, want nil", perr)
	}
	if freshInfo == nil {
		t.Fatalf("the owner check never ran: the file arm of the heal was not reached")
	}
	after, lerr := os.Lstat(statePath)
	if lerr != nil || !os.SameFile(freshInfo, after) {
		t.Fatalf("the fresh state file was removed or replaced by the heal's file arm: err=%v", lerr)
	}
	if got, _ := os.ReadFile(statePath); !bytes.Equal(got, freshBytes) {
		t.Errorf("fresh state file content = %q, want %q", got, freshBytes)
	}
	if logAfter, _ := os.ReadFile(logPath); !bytes.Equal(logBefore, logAfter) {
		t.Errorf("the log changed although the adopted stamp was fresh")
	}
}
