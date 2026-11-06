//go:build !windows

// factory_lease_unix_test.go — SPEC-FACTORY-ATOMIC-LEASE-001 (card t1458) tests
// that need a POSIX shell `git` shim on PATH or an flock on the drift log's lock
// file: AC-FAL-006 (concurrent worktree steps never overlap), AC-FAL-009 (iii)
// (the allowed set of the section) and AC-FAL-015 at the lease function and at
// the verb. The Windows build compiles test files too, so these stay
// unix-tagged.
package cli

import (
	"context"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// flEventLog is one append-only event file shared by the creator wrapper (Go)
// and the git shim (shell): its line order is a total order of the events.
type flEventLog struct {
	path string
	mu   sync.Mutex
}

func (l *flEventLog) add(line string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	_, _ = f.WriteString(line + "\n")
	_ = f.Close()
}

func (l *flEventLog) lines(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile(l.path)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return nmNonEmptyLines(string(raw))
}

// flInstallGitShim puts a `git` shim first on PATH that appends `git-start
// <args>` and `git-end <args>` around the real git to the file named by the
// FL_GIT_EVENT_LOG environment variable. It returns the real git's path.
func flInstallGitShim(t *testing.T) string {
	t.Helper()
	real, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("find git: %v", err)
	}
	dir := t.TempDir()
	script := "#!/bin/sh\n" +
		"echo \"git-start $*\" >> \"$FL_GIT_EVENT_LOG\"\n" +
		"\"" + real + "\" \"$@\"\n" +
		"rc=$?\n" +
		"echo \"git-end $*\" >> \"$FL_GIT_EVENT_LOG\"\n" +
		"exit $rc\n"
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return real
}

// flRenameEnd returns the index of the `git-end … branch -m …` line whose -C
// directory's leaf is cardID, or -1.
func flRenameEnd(lines []string, cardID string) int {
	for i, l := range lines {
		if !strings.HasPrefix(l, "git-end ") || !strings.Contains(l, " branch -m ") {
			continue
		}
		f := strings.Fields(l)
		for k := 0; k+1 < len(f); k++ {
			if f[k] == "-C" && filepath.Base(f[k+1]) == cardID {
				return i
			}
		}
	}
	return -1
}

// flStepOverlap reports whether, in the event log, a creator-enter of the other
// lane falls between one lane's creator-enter and the end of that lane's
// `git branch -m` (the step is not serialized).
func flStepOverlap(lines []string, cards []string) bool {
	for _, a := range cards {
		enter := -1
		for i, l := range lines {
			if l == "creator-enter "+a {
				enter = i
				break
			}
		}
		if enter < 0 {
			continue
		}
		end := flRenameEnd(lines, a)
		if end < 0 {
			end = len(lines)
		}
		for i := enter + 1; i < end; i++ {
			for _, b := range cards {
				if b != a && lines[i] == "creator-enter "+b {
					return true
				}
			}
		}
	}
	return false
}

// TestFactoryEnsureCardWorktreeConcurrentRealMaterializer — AC-FAL-006: two
// leased cards with no recorded worktree in one repository, the REAL worktree
// materializer wrapped so that, once a lane's worktree is created, the lane is
// held until the other lane has also entered the creator or 1.5 s pass (the
// forced overlap of ledger L13), and a git shim logging every call to the same
// event file. Both lanes run the step at the same instant, 12 iterations. Every
// step returns nil, each card records its own worktree and its branch is
// WT-<slug>, and in every iteration the event log shows no creator-enter of one
// lane between the other lane's creator-enter and the end of that lane's
// `git branch -m`.
func TestFactoryEnsureCardWorktreeConcurrentRealMaterializer(t *testing.T) {
	const iters = 12
	const grace = 1500 * time.Millisecond
	realCreator := worktree.WorktreeCreator
	if realCreator == nil {
		t.Fatal("the real worktree creator is not wired in this test binary")
	}
	t.Cleanup(func() { worktree.WorktreeCreator = realCreator })
	realGit := flInstallGitShim(t)

	overlaps, failures, firstFailure := 0, 0, ""
	for i := 0; i < iters; i++ {
		root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
		nmSetText(t, store, "t1", "alpha rename probe")
		nmSetText(t, store, "t2", "beta rename probe")
		far := "2099-01-01T00:00:00Z"
		fcPlace(t, root,
			homestate.Card{CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: far, Stage: homestate.CardRun},
			homestate.Card{CardID: "t2", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2", LeaseExpiresAt: far, Stage: homestate.CardRun})
		cards := []homestate.Card{fcCard(t, root, "t1"), fcCard(t, root, "t2")}
		lanes := []string{"lane-1", "lane-2"}

		elog := &flEventLog{path: filepath.Join(t.TempDir(), "events.log")}
		t.Setenv("FL_GIT_EVENT_LOG", elog.path)
		var entered int32
		barrier := make(chan struct{})
		worktree.WorktreeCreator = func(name string, out io.Writer) (string, error) {
			elog.add("creator-enter " + name)
			if atomic.AddInt32(&entered, 1) == 2 {
				close(barrier)
			}
			path, err := realCreator(name, out)
			if err == nil {
				select {
				case <-barrier:
				case <-time.After(grace):
				}
			}
			elog.add("creator-exit " + name)
			return path, err
		}

		errs := make([]error, 2)
		paths := make([]string, 2)
		var wg sync.WaitGroup
		start := make(chan struct{})
		for k := range cards {
			wg.Add(1)
			go func(k int) {
				defer wg.Done()
				<-start
				paths[k], _, errs[k] = factoryEnsureCardWorktree(context.Background(), root, fcRun, cards[k], lanes[k], io.Discard)
			}(k)
		}
		close(start)
		wg.Wait()
		worktree.WorktreeCreator = realCreator

		lines := elog.lines(t)
		if flStepOverlap(lines, []string{"t1", "t2"}) {
			overlaps++
		}
		bad := false
		for k := range errs {
			if errs[k] != nil {
				bad = true
				if firstFailure == "" {
					firstFailure = errs[k].Error()
				}
				continue
			}
			id := cards[k].CardID
			if got := fcCard(t, root, id).WorktreePath; got != paths[k] {
				t.Errorf("iteration %d: %s records worktree %q, want %q", i, id, got, paths[k])
			}
			branch, err := exec.Command(realGit, "-C", paths[k], "branch", "--show-current").Output()
			if err != nil || !strings.HasPrefix(strings.TrimSpace(string(branch)), "WT-") {
				t.Errorf("iteration %d: %s's branch = %q (%v), want WT-<slug>", i, id, strings.TrimSpace(string(branch)), err)
			}
		}
		if bad {
			failures++
		}
	}
	t.Logf("iterations=%d overlap-iterations=%d failed-iterations=%d", iters, overlaps, failures)
	if overlaps > 0 {
		t.Errorf("in %d of %d iterations one lane entered the creator between the other lane's creator-enter and the end of its `git branch -m` (the worktree step is not serialized)", overlaps, iters)
	}
	if failures > 0 {
		t.Errorf("%d of %d iterations had a failed worktree step; first error: %.160s", failures, iters, firstFailure)
	}
}

// flFileStat is one regular file's size and modification time.
type flFileStat struct {
	size  int64
	mtime time.Time
}

// flSnapshotTree records every regular file under root (the .git directory of
// the project excluded: no git process runs in the section, which the shim's
// log checks separately).
func flSnapshotTree(t *testing.T, root string) map[string]flFileStat {
	t.Helper()
	out := map[string]flFileStat{}
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" && filepath.Dir(p) == root {
			return filepath.SkipDir
		}
		if !d.Type().IsRegular() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		out[p] = flFileStat{size: info.Size(), mtime: info.ModTime()}
		return nil
	})
	return out
}

// TestFactoryLeaseSectionAllowedSet — AC-FAL-009 (iii): between the section's
// entry and the creator stub's call the section runs no git subprocess (the
// shim's log holds no line between the entry marker and the stub's marker) and
// writes no path under the project root other than the queue store's and the
// factory record's own files. The section's entry is the pass-entry seam; the
// git processes the verb runs before it (the project's canonical-root lookups
// while opening the record and building the queue store) are no part of it.
func TestFactoryLeaseSectionAllowedSet(t *testing.T) {
	for _, form := range []string{"nominated", "bare"} {
		t.Run(form, func(t *testing.T) {
			root, store := nmBase(t, factory.BacklogStateQueued)
			nmLaneEnv(t, "lane-1", "")
			dir := t.TempDir()
			initGitRepo(t, dir)
			flInstallGitShim(t)
			elog := &flEventLog{path: filepath.Join(t.TempDir(), "events.log")}
			t.Setenv("FL_GIT_EVENT_LOG", elog.path)

			// The section starts at the pass-entry seam: the git processes the
			// verb runs BEFORE it (opening the record and building the queue
			// store resolve the project's canonical root through git) belong to
			// no section, and the files they create (the record database) are
			// the record's own, so both baselines are taken where the section
			// begins.
			var before, atStub map[string]flFileStat
			prevEntry := factoryLeaseAtPassEntry
			factoryLeaseAtPassEntry = func() {
				elog.add("section-enter")
				before = flSnapshotTree(t, root)
			}
			t.Cleanup(func() { factoryLeaseAtPassEntry = prevEntry })
			prevCreator := worktree.WorktreeCreator
			worktree.WorktreeCreator = func(string, io.Writer) (string, error) {
				elog.add("creator-stub")
				atStub = flSnapshotTree(t, root)
				return dir, nil
			}
			t.Cleanup(func() { worktree.WorktreeCreator = prevCreator })

			args := []string{"--run", fcRun}
			if form == "nominated" {
				args = append(args, "--card", "t1")
			}
			if _, stderr, err := qasRunNext(t, args...); err != nil {
				t.Fatalf("next: %v (stderr %q)", err, stderr)
			}
			if atStub == nil {
				t.Fatal("the creator stub was never called")
			}
			if before == nil {
				t.Fatal("the lease section's entry seam was never reached")
			}
			lines := elog.lines(t)
			enter, stub := -1, -1
			for i, l := range lines {
				switch l {
				case "section-enter":
					if enter < 0 {
						enter = i
					}
				case "creator-stub":
					stub = i
				}
			}
			if enter < 0 || stub < enter {
				t.Fatalf("the event log lacks the section-enter and creator-stub markers in order: %q", lines)
			}
			for _, l := range lines[enter+1 : stub] {
				t.Errorf("a git subprocess ran inside the section (between its entry and the creator stub): %q", l)
			}
			t.Logf("%s: %d git log lines before the section (the record open and the store's canonical-root lookups), 0 inside it expected", form, enter)

			// Paths are compared in their canonical spelling: the record's path is
			// built from the canonical root (/private/var on macOS) while the walk
			// reports the spelling of the temp directory (/var).
			canon := func(p string) string {
				if d, err := filepath.EvalSymlinks(filepath.Dir(p)); err == nil {
					return filepath.Join(d, filepath.Base(p))
				}
				return p
			}
			allowed := map[string]bool{}
			allow := func(p string) {
				p = canon(p)
				allowed[p], allowed[p+"-wal"], allowed[p+"-shm"], allowed[p+"-journal"] = true, true, true, true
			}
			allow(store.EnginePath())
			allow(store.LockPath())
			if dbPath, err := homestate.FactoryDBPath(root); err == nil {
				allow(dbPath)
			}
			// The creator stub runs inside the worktree step, which takes its own
			// lock after the section has ended (plan D3); that lock's artifact is
			// the step's file, not the section's, and the snapshot at the stub
			// reaches past the section into the step.
			allow(filepath.Join(root, ".moai", "state", "factory-worktree-step.lock"))
			var outside []string
			for p, st := range atStub {
				if b, ok := before[p]; ok && b == st {
					continue
				}
				if !allowed[canon(p)] {
					rel, _ := filepath.Rel(root, p)
					outside = append(outside, rel)
				}
			}
			sort.Strings(outside)
			t.Logf("%s: files created or modified under the root before the creator stub, outside the two stores: %v", form, outside)
			if len(outside) > 0 {
				t.Errorf("the section wrote paths outside the queue store and the factory record: %v", outside)
			}
		})
	}
}

// flHoldDriftLogLock takes an exclusive flock on the drift log's lock file
// through a separate open file description (flock conflicts across
// descriptions of one process — the tree's own
// TestFR_UnavailableLogAppendSurvivesRewrite relies on it) and returns the
// release function, safe to call more than once.
func flHoldDriftLogLock(t *testing.T, root string) func() {
	t.Helper()
	logPath, err := homestate.RecordUnavailablePath(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(logPath+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		t.Fatalf("flock the drift log's lock: %v", err)
	}
	var once sync.Once
	release := func() {
		once.Do(func() {
			_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
			_ = f.Close()
		})
	}
	t.Cleanup(release)
	return release
}

// flSeedDriftEntry appends one unreconciled drift-log entry for the run.
func flSeedDriftEntry(t *testing.T, root string) {
	t.Helper()
	if err := homestate.AppendRecordUnavailable(root, homestate.RecordUnavailableEntry{
		RunID: fcRun, CardID: "lost-1", Lane: "lane-2", Error: "probe",
	}); err != nil {
		t.Fatalf("seed the drift log: %v", err)
	}
}

// flUnreconciled counts the run's unreconciled drift-log entries.
func flUnreconciled(t *testing.T, root string) int {
	t.Helper()
	entries, _, err := homestate.ReadRecordUnavailable(root, fcRun)
	if err != nil {
		t.Fatal(err)
	}
	return len(entries)
}

// TestFactoryLeaseDriftLogStallBounded — AC-FAL-015 fixtures (a) and (b),
// observed at the return of the lease function: with one unreconciled drift-log
// entry for the run and the log's lock held (a backstop releases it after more
// than three times C), (i) the function returns within C + 500 ms with a nil
// error and the leased card, and the queue's lock is acquirable right after;
// (ii) the entry is still unreconciled and no record.drift event was appended;
// (iii) once the lock is released, one ordinary record write
// (RecordCardWorktree) reconciles the entry exactly once and a further write
// appends nothing more.
func TestFactoryLeaseDriftLogStallBounded(t *testing.T) {
	for _, form := range []string{"nominated", "bare"} {
		t.Run(form, func(t *testing.T) {
			root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
			nmLaneEnv(t, "lane-1", "")
			flSeedDriftEntry(t, root)
			release := flHoldDriftLogLock(t, root)
			time.AfterFunc(3*factoryLeaseClaimWaitCap+flMargin, release)

			ctx := context.Background()
			start := time.Now()
			var card homestate.Card
			var err error
			leased := false
			if form == "nominated" {
				card, err = factoryNextNominate(ctx, root, fcRun, "lane-1", "t1", "")
				leased = err == nil
			} else {
				card, leased, err = factoryNextLeaseOnceGated(ctx, root, fcRun, "lane-1", false)
			}
			elapsed := time.Since(start)
			unrec, drift := flUnreconciled(t, root), fcEventCount(t, root, "record.drift")
			queueFree, qerr := flQueueFreeWithin(store, flMargin)
			release()
			t.Logf("%s under a held drift-log lock: elapsed=%s err=%v leased=%v card=%s unreconciled-after=%d record.drift-events=%d", form, elapsed, err, leased, card.CardID, unrec, drift)
			if limit := factoryLeaseClaimWaitCap + flMargin; elapsed > limit {
				t.Errorf("clause (i): the lease function returned after %s with the drift-log lock held, want within %s (C + 500 ms)", elapsed, limit)
			}
			if err != nil || !leased {
				t.Fatalf("clause (i): the lease function returned err=%v leased=%v, want the leased card and a nil error", err, leased)
			}
			nmAssertLeased(t, root, card.CardID, "lane-1")
			if !queueFree || qerr != nil {
				t.Errorf("clause (i): the queue's lock was not acquirable right after (completed=%v err=%v)", queueFree, qerr)
			}
			if unrec != 1 || drift != 0 {
				t.Errorf("clause (ii): after the lease, unreconciled=%d (want 1) record.drift-events=%d (want 0): the skip must withhold both the events and the mark", unrec, drift)
			}

			db := fcOpen(t, root)
			if _, err := db.RecordCardWorktree(ctx, fcRun, card.CardID, t.TempDir(), "lane-1", factoryCardNow()); err != nil {
				t.Fatalf("clause (iii): the ordinary write: %v", err)
			}
			if u, d := flUnreconciled(t, root), fcEventCount(t, root, "record.drift"); u != 0 || d != 1 {
				t.Errorf("clause (iii): after one ordinary write, unreconciled=%d (want 0) record.drift-events=%d (want 1)", u, d)
			}
			if _, err := db.RenewLease(ctx, fcRun, card.CardID, "lane-1", factoryCardNow()); err != nil {
				t.Fatalf("clause (iii): the further write: %v", err)
			}
			if d := fcEventCount(t, root, "record.drift"); d != 1 {
				t.Errorf("clause (iii): a further write appended more drift events: %d, want 1", d)
			}
		})
	}
}

// flWriteStart holds the drift log's lock until the verb's card-worktree record
// write is waiting for it (SPEC-FACTORY-ATOMIC-LEASE-001 AC-FAL-015 (d)). The
// write-start signal is the first factoryCardNow read after the worktree creator
// returns: factoryEnsureCardWorktree reads the clock as the argument of
// RecordCardWorktree, right before that write waits for the drift log's lock.
// The release follows the signal by at least 100 ms. That hold is a test premise,
// an ordering lower bound that keeps the release after the write attempt starts; it
// is not a measurement. The measured interval starts at the release and ends at the
// drift mark (flHoldUntilWriteStart states why the hold is a time bound).
type flWriteStart struct {
	signalled  atomic.Value  // time.Time of the write-start signal
	released   atomic.Value  // time.Time of the release
	releasedCh chan struct{} // closed exactly once, at the release
}

// flWriteStartBackstop is a hang guard for flHoldUntilWriteStart, not a measurement.
// It releases the drift log's lock only if the write-start signal never comes, so a
// verb that fails before its card-worktree write cannot hang the test. AC-FAL-015 (d)
// requires only that the backstop exceed three times the claim wait cap; 30 s is a
// hang bound well above that. The earlier 3 x C + 1 s backstop fired before the write
// started on a loaded host, which is why flAssertWriteStartRelease requires the signal
// to precede the release.
const flWriteStartBackstop = 30 * time.Second

// flHoldUntilWriteStart takes the drift log's lock now and arms the write-start
// release. It runs after nmIsolatedWorktrees, so its creator wrapper stacks on
// the stub, and before the verb starts. The lock is held from before the verb
// starts; the backstop only guards against a hang, and flAssertWriteStartRelease
// then requires the signal to have preceded that release.
func flHoldUntilWriteStart(t *testing.T, root string) *flWriteStart {
	t.Helper()
	hold := flHoldDriftLogLock(t, root)
	w := &flWriteStart{releasedCh: make(chan struct{})}
	var releaseOnce sync.Once
	release := func() {
		releaseOnce.Do(func() {
			w.released.Store(time.Now())
			close(w.releasedCh)
			hold()
		})
	}
	var created atomic.Bool
	prevCreator := worktree.WorktreeCreator
	worktree.WorktreeCreator = func(name string, out io.Writer) (string, error) {
		dir, err := prevCreator(name, out)
		created.Store(true)
		return dir, err
	}
	t.Cleanup(func() { worktree.WorktreeCreator = prevCreator })
	var signalOnce sync.Once
	prevNow := factoryCardNow
	factoryCardNow = func() time.Time {
		if created.Load() {
			signalOnce.Do(func() {
				w.signalled.Store(time.Now())
				// The hold is a test premise, not a measurement: an ordering lower bound that keeps
				// the release after the write attempt starts. A signal from the write side that
				// the attempt has reached the drift-log lock would need a production seam, which
				// decision d-20261015T003848Z-8b69 condition 3 excludes, so the bound is a time.
				// Where a signal that the attempt has started can be received without a production
				// change, it is preferred to the time bound.
				// The window it must cover runs from this clock read through the record write
				// and its commit up to the settle step's lock attempt. A longer hold is safer and
				// does not change the measured interval, which starts at the release and ends at
				// the drift mark.
				time.AfterFunc(100*time.Millisecond, release)
			})
		}
		return prevNow()
	}
	t.Cleanup(func() { factoryCardNow = prevNow })
	// Hang guard only (see flWriteStartBackstop): the signal is what releases the lock.
	backstop := time.AfterFunc(flWriteStartBackstop, release)
	t.Cleanup(func() {
		backstop.Stop()
		release()
	})
	return w
}

// instants returns the signal and release instants; a zero value means that the
// event did not happen.
func (w *flWriteStart) instants() (signalled, released time.Time) {
	signalled, _ = w.signalled.Load().(time.Time)
	released, _ = w.released.Load().(time.Time)
	return signalled, released
}

// flFirstGhCall wraps todoRunCommand so that the first `gh` call is timestamped
// and answered by answer; every other command passes through to the runner it
// wraps. The returned getter gives that instant (zero when there was no call).
func flFirstGhCall(t *testing.T, answer func() (string, error)) func() time.Time {
	t.Helper()
	var (
		first atomic.Value
		once  sync.Once
	)
	prevRun := todoRunCommand
	prevContextRun := todoRunCommandContext
	todoRunCommand = func(name string, args ...string) (string, error) {
		if name != "gh" {
			return prevRun(name, args...)
		}
		once.Do(func() { first.Store(time.Now()) })
		return answer()
	}

	todoRunCommandContext = func(ctx context.Context, name string, args ...string) (string, error) {
		if name != "gh" {
			return prevContextRun(ctx, name, args...)
		}
		return todoRunCommand(name, args...)
	}
	t.Cleanup(func() { todoRunCommand = prevRun; todoRunCommandContext = prevContextRun })
	return func() time.Time {
		at, _ := first.Load().(time.Time)
		return at
	}
}

// flAssertWriteStartRelease asserts the premises that clause (vii)'s measurement
// rests on: the write-start signal fired, the signal preceded the release (so the
// release came from the write's start and not from the backstop, which would leave
// the write unmeasured), the open-PR query ran, and it ran before the release, so no
// network call lies inside the measured interval.
func flAssertWriteStartRelease(t *testing.T, signalled, released, gh time.Time) {
	t.Helper()
	if signalled.IsZero() {
		t.Fatal("clause (vii): the write-start signal never fired (no clock read after the worktree creator returned), so the backstop released the lock and the write's wait was not measured")
	}
	if released.IsZero() {
		t.Fatal("clause (vii): the lock was never released")
	}
	if !signalled.Before(released) {
		t.Fatalf("clause (vii): the lock was released %s before the write-start signal, so the backstop released it and the card-worktree write did not wait for the lock; the measurement is void", signalled.Sub(released))
	}
	if gh.IsZero() {
		t.Errorf("clause (vii): the open-PR query (gh) never ran in the verb, so the output path had no query prefetched")
		return
	}
	if !gh.Before(released) {
		t.Errorf("clause (vii): the open-PR query ran %s after the lock release, want before it: a network call inside the measured interval", gh.Sub(released))
	}
}

// flCommitPoll is the interval of the mark watch's reads and flCommitWatchLimit its
// hang bound. The bound is not a measurement: a write whose mark never appears fails
// the test through flAssertMarkAfterRelease. The poll interval is 2 ms, which is far
// below the 500 ms bound: the observation can lag the mark by at most one interval
// plus one read's latency, and the AC-FAL-015 (vii) watch requires an interval of
// 10 ms or less.
const (
	flCommitPoll       = 2 * time.Millisecond
	flCommitWatchLimit = 30 * time.Second
)

// flMarkWatch returns the getter of the instant the drift-log reconcile mark of the
// run's card-worktree record write is set: the first read that finds no unreconciled
// entry of the run (AC-FAL-015 (vii), as amended by decisions d-20261015T074309Z-fb83 and
// d-20261015T074756Z-e04f).
// The mark is set in the write's settle step, after the write's commit and after the
// drift-log lock wait. The record commit that makes the card's worktree path visible
// comes before that wait, so the path cannot mark the end of the interval. The watch
// polls ReadRecordUnavailable, which takes no lock and writes nothing. The getter
// waits for that observation or the hang bound, so a read made after the verb
// returns still sees the mark.
func flMarkWatch(t *testing.T, root string) func() time.Time {
	t.Helper()
	var (
		marked atomic.Value
		stop   = make(chan struct{})
		done   = make(chan struct{})
	)
	go func() {
		defer close(done)
		deadline := time.Now().Add(flCommitWatchLimit)
		for time.Now().Before(deadline) {
			if entries, _, err := homestate.ReadRecordUnavailable(root, fcRun); err == nil && len(entries) == 0 {
				marked.Store(time.Now())
				return
			}
			select {
			case <-stop:
				return
			case <-time.After(flCommitPoll):
			}
		}
	}()
	t.Cleanup(func() {
		close(stop)
		<-done
	})
	return func() time.Time {
		<-done
		at, _ := marked.Load().(time.Time)
		return at
	}
}

// flAssertMarkAfterRelease asserts the premise of the amended clause (vii)
// measurement: the reconcile mark was set only after the drift-log lock was released,
// so the write's settle step waited for the lock and the mark is a true end of the
// measured interval.
func flAssertMarkAfterRelease(t *testing.T, released, marked time.Time) {
	t.Helper()
	if marked.IsZero() {
		t.Fatal("clause (vii): the drift-log reconcile mark of the card-worktree write was never observed, so the measured interval has no end")
	}
	if marked.Before(released) {
		t.Fatalf("clause (vii): the reconcile mark was observed %s before the lock release, so the write's settle step did not wait for the lock; the measurement is void", released.Sub(marked))
	}
}

// TestFactoryLeaseDriftLogVerbWorktreeWriteWaits — AC-FAL-015 fixture (d),
// clause (vii), driven through the verb with the creator stubbed: the drift log's
// lock is held from before the verb starts and released at the write-start signal
// (flHoldUntilWriteStart). The contract is order, not elapsed wall-clock (card
// t1746, following card t1807's pattern): the verb runs on a goroutine and must
// return only AFTER the release signal fired — the card-worktree write's settle
// step waits for the lock, so the reconcile mark, and with it the verb's return,
// cannot precede the release; a verb that returns with the lock still held
// skipped the wait. The mark itself must be observed, which the mark watch
// bounds at flCommitWatchLimit as a hang guard, not a measurement; elapsed times
// are also bounded by B + 500 ms. For each form the verb exits 0, the card is
// leased with its worktree recorded, the open-PR query ran before the release,
// and the run's unreconciled count is then 0 with exactly one record.drift event
// for the entry.
func TestFactoryLeaseDriftLogVerbWorktreeWriteWaits(t *testing.T) {
	for _, form := range []string{"nominated", "bare"} {
		t.Run(form, func(t *testing.T) {
			root, _ := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
			nmLaneEnv(t, "lane-1", "")
			nmIsolatedWorktrees(t, "t1", "t2")
			flSeedDriftEntry(t, root)
			gh := flFirstGhCall(t, func() (string, error) { return "[]", nil })
			ws := flHoldUntilWriteStart(t, root)
			cw := flMarkWatch(t, root)

			args := []string{"--run", fcRun}
			if form == "nominated" {
				args = append(args, "--card", "t1")
			}
			res := <-flRunNextAsync(t, ws.releasedCh, args...)
			out, stderr, err := res.out, res.stderr, res.err
			if err != nil {
				t.Fatalf("clause (vii): the verb exited %d: %v (stderr %q)", nmExit(err), err, stderr)
			}
			// Order check: the verb must return only after the release signal
			// fired (the settle step waited for the lock). The fact was sampled
			// inside the verb goroutine at the instant of the return
			// (flRunNextAsync); here the compliant direction is the opposite of
			// the queue-lock test's — the return must FOLLOW the release, so
			// heldAtReturn must be false.
			if res.heldAtReturn {
				t.Fatalf("clause (vii): the verb returned while the drift-log lock was still held: the card-worktree write's settle step did not wait for the lock")
			}
			head := nmLeasedHead(out)
			if head == "" {
				t.Fatalf("clause (vii): the verb printed no leased card (stdout %q)", out)
			}
			id := strings.Fields(head)[0]
			if id != "t1" {
				t.Fatalf("clause (vii): the verb leased %s, want t1: the mark watch polls t1", id)
			}
			nmAssertLeased(t, root, id, "lane-1")
			if c := fcCard(t, root, id); strings.TrimSpace(c.WorktreePath) == "" {
				t.Errorf("clause (vii): %s has no recorded worktree", id)
			}
			signalled, rel := ws.instants()
			flAssertWriteStartRelease(t, signalled, rel, gh())
			marked := cw()
			flAssertMarkAfterRelease(t, rel, marked)
			after := marked.Sub(rel)
			unrec, drift := flUnreconciled(t, root), fcEventCount(t, root, "record.drift")
			t.Logf("%s: write-start signal to release %s; reconcile mark %s after the release; verb returned %s after the release; unreconciled-after=%d record.drift-events=%d", form, rel.Sub(signalled), after, res.returned.Sub(rel), unrec, drift)
			if unrec != 0 || drift != 1 {
				t.Errorf("clause (vii): unreconciled=%d (want 0) record.drift-events=%d (want 1)", unrec, drift)
			}
		})
	}
}

// TestFactoryLeaseDriftLogVerbTailHoldsWithOutputOverhead — AC-FAL-015 fixture (d),
// clause (vii), with output-path overhead. The open-PR query (`gh pr list`, through
// todoRunCommand in fetchOpenPRs) takes flOverheadCeil (700 ms) in this test; every
// other command passes through to the real runner. The verb runs that query before
// it takes the queue lock (AC-FAL-015 (vii)), so the latency falls before the lease.
// The drift-log lock is released at the write-start signal (flHoldUntilWriteStart),
// and the gap from that release to the reconcile mark of the card-worktree write must
// stay within flMargin (500 ms). The landed check runs after that mark on the output
// path and lies outside the measured interval.
func TestFactoryLeaseDriftLogVerbTailHoldsWithOutputOverhead(t *testing.T) {
	for _, form := range []string{"nominated", "bare"} {
		t.Run(form, func(t *testing.T) {
			root, _ := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
			nmLaneEnv(t, "lane-1", "")
			nmIsolatedWorktrees(t, "t1", "t2")
			flSeedDriftEntry(t, root)
			gh := flFirstGhCall(t, func() (string, error) {
				time.Sleep(flOverheadCeil)
				return "", os.ErrDeadlineExceeded
			})
			ws := flHoldUntilWriteStart(t, root)
			cw := flMarkWatch(t, root)

			args := []string{"--run", fcRun}
			if form == "nominated" {
				args = append(args, "--card", "t1")
			}
			out, stderr, err := qasRunNext(t, args...)
			returned := time.Now()
			if err != nil {
				t.Fatalf("clause (vii): the verb exited %d: %v (stderr %q)", nmExit(err), err, stderr)
			}
			head := nmLeasedHead(out)
			if head == "" {
				t.Fatalf("clause (vii): the verb printed no leased card (stdout %q)", out)
			}
			if fields := strings.Fields(head); len(fields) == 0 || fields[0] != "t1" {
				t.Fatalf("clause (vii): the verb leased %q, want t1: the mark watch polls t1", head)
			}
			signalled, rel := ws.instants()
			flAssertWriteStartRelease(t, signalled, rel, gh())
			marked := cw()
			flAssertMarkAfterRelease(t, rel, marked)
			after := marked.Sub(rel)
			t.Logf("%s: %s of open-PR query latency; reconcile mark %s after the release; verb returned %s after the release", form, flOverheadCeil, after, returned.Sub(rel))
			if after > flMargin {
				t.Errorf("clause (vii): the reconcile mark of the card-worktree write came %s after the lock release with %s of open-PR query latency, want within %s", after, flOverheadCeil, flMargin)
			}
		})
	}
}
