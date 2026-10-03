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

// TestFactoryLeaseDriftLogVerbWorktreeWriteWaits — AC-FAL-015 fixture (d),
// clause (vii), driven through the verb with the creator stubbed: the drift
// log's lock is held for 2 s from before the verb starts by a timer that
// records the instant it releases. For each form the verb exits 0, the card is
// leased with its worktree recorded, the verb returned not before the release
// and within 500 ms after it (the verb's own card-worktree write waited for the
// lock), and the run's unreconciled count is then 0 with exactly one
// record.drift event for the entry.
func TestFactoryLeaseDriftLogVerbWorktreeWriteWaits(t *testing.T) {
	for _, form := range []string{"nominated", "bare"} {
		t.Run(form, func(t *testing.T) {
			root, _ := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
			nmLaneEnv(t, "lane-1", "")
			nmIsolatedWorktrees(t, "t1", "t2")
			flSeedDriftEntry(t, root)
			release := flHoldDriftLogLock(t, root)
			var releasedAt atomic.Value
			time.AfterFunc(2*time.Second, func() {
				releasedAt.Store(time.Now())
				release()
			})

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
			id := strings.Fields(head)[0]
			nmAssertLeased(t, root, id, "lane-1")
			if c := fcCard(t, root, id); strings.TrimSpace(c.WorktreePath) == "" {
				t.Errorf("clause (vii): %s has no recorded worktree", id)
			}
			rel, _ := releasedAt.Load().(time.Time)
			if rel.IsZero() {
				t.Fatalf("clause (vii): the verb returned before the timer released the lock (it did not wait for the held lock)")
			}
			after := returned.Sub(rel)
			unrec, drift := flUnreconciled(t, root), fcEventCount(t, root, "record.drift")
			t.Logf("%s: returned %s after the release; unreconciled-after=%d record.drift-events=%d", form, after, unrec, drift)
			if after < 0 || after > flMargin {
				t.Errorf("clause (vii): the verb returned %s relative to the release, want not before it and within %s after it", after, flMargin)
			}
			if unrec != 0 || drift != 1 {
				t.Errorf("clause (vii): unreconciled=%d (want 0) record.drift-events=%d (want 1)", unrec, drift)
			}
		})
	}
}
