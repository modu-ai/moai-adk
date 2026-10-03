//go:build !windows

package harness

// DRAFT measurement instrument for the amendment 0.4.0 test debt (N1, N2, N4 of
// `.moai/reports/t1432/sync-audit-delta.md`). Not part of the change and not in the tree: injected with
// `go test -overlay`. Compiles against the tree at 7639c04c1.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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

// N2: a prune with no late event leaves the replacement log exactly the kept lines: it does not end
// with a blank line (the log reader drops blank lines, so only a raw count sees it).
func TestPruneWithoutLateEventAddsNoBlankLine(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writeStaleLog(t, logPath, now, "stale-1")
	if err := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now }).PruneStaleEntries(30); err != nil {
		t.Fatalf("prune returned %v, want nil", err)
	}
	got, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if n := strings.Count(string(got), "\n"); n != 1 || !bytes.HasSuffix(got, []byte("}\n")) {
		t.Errorf("replacement log = %q: want exactly one newline-terminated kept line and no blank line", got)
	}
}
