//go:build !windows

package harness

// DRAFT measurement instrument for the SPEC-HARNESS-RETENTION-HARDEN-001 amendment 0.4.0 (card t1432).
// Not part of the change and not in the tree: it is injected with `go test -overlay`. It compiles against
// the tree at 7639c04c1, where no heal lock exists: it names the heal-lock file by the literal
// ".prune-heal" and takes the lock itself with syscall.Flock, so every assertion is about what the
// pruner does while ANOTHER descriptor holds that lock.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

const healSuffixDraft = ".prune-heal"

// holdHealLockDraft opens <log>.prune-heal (creating it) and takes an exclusive flock on it, as another
// healer would. The returned release is idempotent and also runs at test cleanup.
func holdHealLockDraft(t *testing.T, logPath string) (release func()) {
	t.Helper()
	f, err := os.OpenFile(logPath+healSuffixDraft, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		t.Fatalf("open heal lock: %v", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock heal lock: %v", err)
	}
	var once sync.Once
	release = func() {
		once.Do(func() {
			_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
			_ = f.Close()
		})
	}
	t.Cleanup(release)
	return release
}

// healFixture builds a log with one stale event and a state path that is a symbolic link owned by the
// current user, pointing at a victim file.
type healFixture struct {
	dir, logPath, archiveDir, statePath, healPath, victim string
	now                                                   time.Time
	logBefore                                             []byte
	inspected                                             os.FileInfo
}

func newHealFixture(t *testing.T) healFixture {
	t.Helper()
	dir := t.TempDir()
	fx := healFixture{
		dir:        dir,
		logPath:    filepath.Join(dir, "usage-log.jsonl"),
		archiveDir: filepath.Join(dir, "archive"),
		now:        time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC),
	}
	fx.statePath = fx.logPath + stampSuffix
	fx.healPath = fx.logPath + healSuffixDraft
	writeStaleLog(t, fx.logPath, fx.now, "stale-1")
	var err error
	if fx.logBefore, err = os.ReadFile(fx.logPath); err != nil {
		t.Fatalf("read log: %v", err)
	}
	fx.victim = filepath.Join(dir, "victim.txt")
	if err := os.WriteFile(fx.victim, bytes.Repeat([]byte("V"), 64), 0o644); err != nil {
		t.Fatalf("write victim: %v", err)
	}
	if err := os.Symlink(fx.victim, fx.statePath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if fx.inspected, err = os.Lstat(fx.statePath); err != nil {
		t.Fatalf("lstat state: %v", err)
	}
	return fx
}

// TestPruneHealSerializesOnTheHealLock (AC-HRH-006 case b): while another descriptor holds the heal
// lock, the pruner neither removes nor replaces the faulty entry; after the holder renames a fresh
// regular state file over the entry (as a winning healer does) and releases the lock, the pruner
// re-inspects, leaves that file untouched and adopts its fresh stamp.
func TestPruneHealSerializesOnTheHealLock(t *testing.T) {
	t.Parallel()
	fx := newHealFixture(t)
	release := holdHealLockDraft(t, fx.logPath)

	ret := NewRetention(fx.logPath, fx.archiveDir, func() time.Time { return fx.now })
	done := make(chan error, 1)
	go func() { done <- ret.PruneStaleEntries(30) }()

	returned := false
	var perr error
	select {
	case perr = <-done:
		returned = true
		t.Errorf("the pruner returned (%v) while another descriptor held the heal lock: it did not wait", perr)
	case <-time.After(300 * time.Millisecond):
	}
	if cur, lerr := os.Lstat(fx.statePath); lerr != nil || !os.SameFile(fx.inspected, cur) {
		t.Errorf("the state-path entry was removed or replaced while the heal lock was held by another descriptor: err=%v", lerr)
	}
	if t.Failed() {
		release()
		if !returned {
			select {
			case <-done:
			case <-time.After(5 * time.Second):
			}
		}
		return
	}

	// The winner: a healthy regular state file with a still-fresh stamp replaces the faulty link.
	freshBytes := []byte(fx.now.Add(-time.Second).UTC().Format(time.RFC3339Nano))
	fresh := filepath.Join(fx.dir, "fresh.tmp")
	if err := os.WriteFile(fresh, freshBytes, 0o644); err != nil {
		t.Fatalf("write fresh: %v", err)
	}
	if err := os.Rename(fresh, fx.statePath); err != nil {
		t.Fatalf("rename fresh: %v", err)
	}
	freshInfo, err := os.Lstat(fx.statePath)
	if err != nil {
		t.Fatalf("lstat fresh: %v", err)
	}
	release()

	select {
	case perr = <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("the pruner did not finish after the heal lock was released")
	}
	if perr != nil {
		t.Errorf("the pruner returned %v, want nil", perr)
	}
	after, lerr := os.Lstat(fx.statePath)
	if lerr != nil || !os.SameFile(freshInfo, after) {
		t.Fatalf("the winner's fresh state file was removed or replaced: err=%v", lerr)
	}
	if got, _ := os.ReadFile(fx.statePath); !bytes.Equal(got, freshBytes) {
		t.Errorf("fresh state file content = %q, want %q", got, freshBytes)
	}
	if logAfter, _ := os.ReadFile(fx.logPath); !bytes.Equal(fx.logBefore, logAfter) {
		t.Errorf("the log changed although the adopted stamp was fresh")
	}
}

// TestPruneHealLockHeldPastTheBoundFailsClosed (AC-HRH-015): the heal lock stays held for the whole
// call; the pruner gives up within the bound without touching anything, returns an error naming the
// heal lock and writes one warning line. Serial: it swaps os.Stderr.
func TestPruneHealLockHeldPastTheBoundFailsClosed(t *testing.T) {
	fx := newHealFixture(t)
	release := holdHealLockDraft(t, fx.logPath)
	healBefore, err := os.Lstat(fx.healPath)
	if err != nil {
		t.Fatalf("lstat heal lock: %v", err)
	}

	ret := NewRetention(fx.logPath, fx.archiveDir, func() time.Time { return fx.now })
	var perr error
	var elapsed time.Duration
	stderr := captureStderr(t, func() {
		done := make(chan error, 1)
		start := time.Now()
		go func() { done <- ret.PruneStaleEntries(30) }()
		select {
		case perr = <-done:
		case <-time.After(8 * time.Second):
			t.Errorf("hang: PruneStaleEntries did not return within 8 s while the heal lock was held")
			release() // let the goroutine finish so the test ends
			perr = <-done
		}
		elapsed = time.Since(start)
	})

	const ceiling = 4500 * time.Millisecond
	if elapsed >= ceiling {
		t.Errorf("the call took %v, want under %v (the hook timeout is 5 s)", elapsed, ceiling)
	}
	if perr == nil || !strings.Contains(perr.Error(), fx.healPath) {
		t.Errorf("want an error naming %s, got %v", fx.healPath, perr)
	}
	if cur, lerr := os.Lstat(fx.statePath); lerr != nil || !os.SameFile(fx.inspected, cur) || cur.Mode() != fx.inspected.Mode() {
		t.Errorf("the state-path entry was removed or replaced although the heal lock was not acquired: err=%v", lerr)
	}
	if got, _ := os.ReadFile(fx.victim); !bytes.Equal(got, bytes.Repeat([]byte("V"), 64)) {
		t.Errorf("victim changed: %q", got)
	}
	if logAfter, _ := os.ReadFile(fx.logPath); !bytes.Equal(fx.logBefore, logAfter) {
		t.Errorf("log changed")
	}
	if _, serr := os.Stat(fx.archiveDir); !os.IsNotExist(serr) {
		t.Errorf("archive directory exists: %v", serr)
	}
	if strings.Count(stderr, "\n") != 1 || !strings.HasPrefix(stderr, "[WARN] harness/retention:") || !strings.Contains(stderr, fx.healPath) {
		t.Errorf("want exactly one warning line naming %s, got %q", fx.healPath, stderr)
	}
	if healAfter, lerr := os.Lstat(fx.healPath); lerr != nil || !os.SameFile(healBefore, healAfter) || healAfter.Size() != 0 {
		t.Errorf("the heal-lock entry was removed, replaced or truncated: err=%v", lerr)
	}
}

// TestPruneHealLockHostileEntryFailsClosed (AC-HRH-016): a symbolic link, a directory, a FIFO or a
// not-owned file at the heal-lock path makes the heal fail closed, never followed, never truncated,
// never removed, and the call returns within the bound (no new hang path). Serial: swaps os.Stderr.
func TestPruneHealLockHostileEntryFailsClosed(t *testing.T) {
	type hostile struct {
		name  string
		setup func(t *testing.T, fx healFixture) (check func(t *testing.T))
		owner func(path string, fx healFixture) bool
	}
	victim2Bytes := bytes.Repeat([]byte("W"), 64)
	cases := []hostile{
		{"symlink", func(t *testing.T, fx healFixture) func(*testing.T) {
			v2 := filepath.Join(fx.dir, "victim2.txt")
			if err := os.WriteFile(v2, victim2Bytes, 0o644); err != nil {
				t.Fatalf("write victim2: %v", err)
			}
			if err := os.Symlink(v2, fx.healPath); err != nil {
				t.Fatalf("symlink heal lock: %v", err)
			}
			return func(t *testing.T) {
				if got, _ := os.ReadFile(v2); !bytes.Equal(got, victim2Bytes) {
					t.Errorf("the link target behind the heal-lock path changed: %q", got)
				}
				if fi, err := os.Lstat(fx.healPath); err != nil || fi.Mode()&os.ModeSymlink == 0 {
					t.Errorf("the heal-lock symbolic link was removed or replaced: err=%v", err)
				}
			}
		}, nil},
		{"directory", func(t *testing.T, fx healFixture) func(*testing.T) {
			if err := os.Mkdir(fx.healPath, 0o755); err != nil {
				t.Fatalf("mkdir heal lock: %v", err)
			}
			return func(t *testing.T) {
				if fi, err := os.Lstat(fx.healPath); err != nil || !fi.IsDir() {
					t.Errorf("the heal-lock directory was removed or replaced: err=%v", err)
				}
			}
		}, nil},
		{"fifo", func(t *testing.T, fx healFixture) func(*testing.T) {
			if err := syscall.Mkfifo(fx.healPath, 0o644); err != nil {
				t.Skipf("mkfifo unavailable: %v", err)
			}
			return func(t *testing.T) {
				if fi, err := os.Lstat(fx.healPath); err != nil || fi.Mode()&os.ModeNamedPipe == 0 {
					t.Errorf("the heal-lock FIFO was removed or replaced: err=%v", err)
				}
			}
		}, nil},
		{"not-owned", func(t *testing.T, fx healFixture) func(*testing.T) {
			if err := os.WriteFile(fx.healPath, nil, 0o600); err != nil {
				t.Fatalf("write heal lock: %v", err)
			}
			before, _ := os.Lstat(fx.healPath)
			return func(t *testing.T) {
				if fi, err := os.Lstat(fx.healPath); err != nil || !os.SameFile(before, fi) || fi.Size() != 0 {
					t.Errorf("the heal-lock file was removed, replaced or truncated: err=%v", err)
				}
			}
		}, func(path string, fx healFixture) bool { return path != fx.healPath }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fx := newHealFixture(t)
			check := tc.setup(t, fx)
			ret := NewRetention(fx.logPath, fx.archiveDir, func() time.Time { return fx.now })
			if tc.owner != nil {
				ret.ownerCheck = func(p string) bool { return tc.owner(p, fx) }
			}
			var perr error
			stderr := captureStderr(t, func() {
				done := make(chan error, 1)
				go func() { done <- ret.PruneStaleEntries(30) }()
				select {
				case perr = <-done:
				case <-time.After(4500 * time.Millisecond):
					t.Errorf("hang: PruneStaleEntries did not return within 4.5 s")
				}
			})
			if perr == nil || !strings.Contains(perr.Error(), fx.healPath) {
				t.Errorf("want an error naming %s, got %v", fx.healPath, perr)
			}
			if cur, lerr := os.Lstat(fx.statePath); lerr != nil || !os.SameFile(fx.inspected, cur) {
				t.Errorf("the state-path entry was removed or replaced although the heal lock was unusable: err=%v", lerr)
			}
			if got, _ := os.ReadFile(fx.victim); !bytes.Equal(got, bytes.Repeat([]byte("V"), 64)) {
				t.Errorf("victim changed: %q", got)
			}
			if logAfter, _ := os.ReadFile(fx.logPath); !bytes.Equal(fx.logBefore, logAfter) {
				t.Errorf("log changed")
			}
			if strings.Count(stderr, "\n") != 1 || !strings.HasPrefix(stderr, "[WARN] harness/retention:") || !strings.Contains(stderr, fx.healPath) {
				t.Errorf("want exactly one warning line naming %s, got %q", fx.healPath, stderr)
			}
			check(t)
		})
	}
}

// TestPruneHealLockIsFreeWhilePrunerIsInItsArchiveStep (AC-HRH-006 case d): the heal lock is held only
// across the re-inspection and the removal. A pruner that healed a faulty entry and is now blocked in
// its archive step (the month archive path is a FIFO) must not be holding the heal lock: a
// non-blocking exclusive flock by the test succeeds.
func TestPruneHealLockIsFreeWhilePrunerIsInItsArchiveStep(t *testing.T) {
	t.Parallel()
	fx := newHealFixture(t)
	release := blockedPruner(t, fx.dir, fx.now, fx.logPath)
	defer func() { _ = release() }()

	if fi, err := os.Lstat(fx.statePath); err != nil || !fi.Mode().IsRegular() {
		t.Fatalf("precondition: the pruner must have replaced the faulty link by now: err=%v", err)
	}
	f, err := os.OpenFile(fx.healPath, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		t.Fatalf("open heal lock: %v", err)
	}
	defer func() { _ = f.Close() }()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Errorf("the heal lock is held while the pruner is in its archive step: %v", err)
	}
}

// TestPruneCommonPathCreatesNoHealLock (AC-HRH-006 case c): a healthy state file and an absent state
// path both prune without ever creating the heal-lock file.
func TestPruneCommonPathCreatesNoHealLock(t *testing.T) {
	for _, name := range []string{"absent", "healthy-expired-stamp"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			logPath := filepath.Join(dir, "usage-log.jsonl")
			now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
			writeStaleLog(t, logPath, now, "stale-1")
			if name == "healthy-expired-stamp" {
				if err := os.WriteFile(logPath+stampSuffix, []byte("2026-09-01T00:00:00Z"), 0o644); err != nil {
					t.Fatalf("write state: %v", err)
				}
			}
			if err := NewRetention(logPath, filepath.Join(dir, "archive"), func() time.Time { return now }).PruneStaleEntries(30); err != nil {
				t.Fatalf("prune returned %v, want nil", err)
			}
			if logHasSubject(t, logPath, "stale-1") {
				t.Errorf("the prune did not run (stale event still in the log)")
			}
			if _, err := os.Lstat(logPath + healSuffixDraft); !os.IsNotExist(err) {
				t.Errorf("a heal-lock entry exists after a prune that needed no heal: err=%v", err)
			}
		})
	}
}
