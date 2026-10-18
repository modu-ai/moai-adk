//go:build !windows

package cli

// migrate_ownership_nonregular_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 (card
// t1865, second card-review finding): the verdict function must decide what kind
// of file it was handed BEFORE it reads anything. The common-asset walk hands it
// every non-directory, and a FIFO (or a socket, a device, a symlink to one)
// makes os.ReadFile block until a writer shows up, so `moai update` hung even
// with the removal switch off. The base commit preserved an unregistered path
// without reading it.
//
// Symlinks (chunk 13, leader disposition B): an unregistered path with a symlink
// among its components is never followed, so a symlink to a regular file, and a
// symlink to a pipe, are kept unread with `symlink-escape` (the first used to be
// judged by its target's bytes — a baseline that CHANGED by that decision). A
// dangling symlink is still a failed read (`read-failed`). Only a path whose own
// final element is not a regular file takes the reason `non-regular-file`.
//
// The test must FAIL, not hang, while the bug is present: the migration runs in a
// goroutine and the test waits a bounded time ONLY to detect that it never
// returned. That wait is not an assertion about speed. On a hang the FIFOs are
// poked (opened and closed read-write) until the blocked read ends and the goroutine can finish before
// the temporary tree goes away.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

const nonRegularReason = keepReason("non-regular-file")

const (
	nrDir        = ".agents/skills/moai-nonregular"
	nrFifo       = nrDir + "/pipe"            // unregistered FIFO
	nrPlain      = nrDir + "/plain.md"        // unregistered regular file
	nrLinkReg    = nrDir + "/link-to-plain"   // symlink to the regular file
	nrLinkFifo   = nrDir + "/link-to-pipe"    // symlink to the FIFO
	nrDangling   = nrDir + "/link-to-nothing" // symlink to nothing
	nrRegFifo    = ".claude/skills/moai-nonregular-reg/pipe"
	nrPlainBytes = "# a plain unregistered file\n"
)

func mkfifo(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(path, 0o644); err != nil {
		// Not a skip: a skipped test sweeps nothing and the hang it guards could come
		// back unseen on a file system that refuses mkfifo. The build tag above is the
		// only exclusion.
		t.Fatalf("mkfifo is not available on this file system (%s): %v", filepath.Dir(path), err)
	}
}

// pokeFifos ends reads that are blocked on the FIFOs: opening a FIFO read-write
// never blocks, and the open wakes a reader waiting for a writer; closing it again
// leaves that reader with no writer, so it reads end-of-file. A reader that wakes
// late blocks again, so the caller repeats it until the goroutine is done.
func pokeFifos(paths []string) {
	for _, p := range paths {
		if f, err := os.OpenFile(p, os.O_RDWR, 0); err == nil {
			_ = f.Close()
		}
	}
}

type nonRegularWorld struct {
	root, home string
	fifos      []string // absolute paths of every FIFO planted
	registered []byte   // the bytes the registered path had before it became a FIFO
}

func nonRegularFixture(t *testing.T) nonRegularWorld {
	t.Helper()
	root, home := ownershipFixtureBase(t) // switch on inside this fixture only
	w := nonRegularWorld{root: root, home: home}
	abs := func(rel string) string { return filepath.Join(root, filepath.FromSlash(rel)) }

	ownAddUnregistered(t, root, nrPlain, []byte(nrPlainBytes))
	mkfifo(t, abs(nrFifo))
	if err := os.Symlink("plain.md", abs(nrLinkReg)); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("pipe", abs(nrLinkFifo)); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("does-not-exist", abs(nrDangling)); err != nil {
		t.Fatal(err)
	}
	// A registered FIFO: register a regular file (Track reads it), then swap it for
	// a FIFO the way a user could.
	w.registered = []byte("# registered, then replaced by a pipe\n")
	ownAddRegisteredManaged(t, root, nrRegFifo, w.registered)
	if err := os.Remove(abs(nrRegFifo)); err != nil {
		t.Fatal(err)
	}
	mkfifo(t, abs(nrRegFifo))
	w.fifos = []string{abs(nrFifo), abs(nrRegFifo)}
	return w
}

type nonRegularOutcome struct {
	removed []string
	rc      *migrationRunContext
	report  []string
	err     error
	hung    bool
}

// runNonRegularMigration runs the migration in a goroutine and waits a bounded
// time only to learn whether it ever returns.
func runNonRegularMigration(t *testing.T, w nonRegularWorld) nonRegularOutcome {
	t.Helper()
	rc := &migrationRunContext{userAssetsInstalled: true}
	var report []string
	type result struct {
		removed []string
		err     error
	}
	done := make(chan result, 1)
	go func() {
		removed, err := migrateProjectCommonAssets(w.root, w.home, true, nil, func(format string, args ...interface{}) {
			report = append(report, fmt.Sprintf(format, args...))
		}, rc)
		done <- result{removed, err}
	}()
	select {
	case r := <-done:
		return nonRegularOutcome{removed: r.removed, rc: rc, report: report, err: r.err}
	case <-time.After(30 * time.Second):
		// The call is blocked on a read of a FIFO. End the read, and wait for the
		// goroutine, so nothing outlives the temporary tree.
		deadline := time.After(30 * time.Second)
	poke:
		for {
			pokeFifos(w.fifos)
			select {
			case <-done:
				break poke
			case <-deadline:
				break poke
			case <-time.After(20 * time.Millisecond):
			}
		}
		return nonRegularOutcome{hung: true}
	}
}

// nonRegularProblems is the whole criterion as a pure function of an outcome and
// the planted tree.
func nonRegularProblems(w nonRegularWorld, o nonRegularOutcome) []string {
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }
	if o.hung {
		return []string{"the migration did not return: it is blocked reading a FIFO"}
	}
	if o.err != nil {
		add("the migration returned an error: %v", o.err)
	}
	row := func(rel string) (remainingRow, bool) { return findRemaining(o.rc, rel) }
	want := func(rel string, class ownershipClass, reason keepReason) {
		r, ok := row(rel)
		switch {
		case !ok:
			add("%s has no row in the remaining list", rel)
		case r.Reason != reason:
			add("%s: reason %q, want %q", rel, r.Reason, reason)
		case r.Class != class:
			add("%s: class %q, want %q", rel, r.Class, class)
		}
	}
	// The non-regular paths: kept, with the explicit reason (a link to a pipe is a link: never followed).
	want(nrFifo, ownershipUnregistered, nonRegularReason)
	want(nrLinkFifo, ownershipUnregistered, keepSymlinkEscape) // a link is never followed (chunk 13), so the pipe behind it is never touched
	want(nrRegFifo, ownershipUserModified, nonRegularReason)
	// The rest of the walk: a regular file is judged by content, a symlink to a
	// regular file is kept unread (changed baseline) and a dangling symlink is a
	// failed read.
	want(nrPlain, ownershipUnregistered, keepNoProvenanceRecord)
	want(nrLinkReg, ownershipUnregistered, keepSymlinkEscape) // changed baseline (chunk 13): was judged by the target's bytes
	want(nrDangling, ownershipUnregistered, keepReadFailed)
	// Reach: the registered sibling with every removal precondition went in the same
	// run, so the keeps above come from a walk that could remove.
	if !slices.Equal(o.removed, []string{migIdenticalSkill}) {
		add("removed %v, want exactly [%s]", o.removed, migIdenticalSkill)
	}
	// Nothing was removed or changed.
	for _, f := range w.fifos {
		if info, err := os.Lstat(f); err != nil || info.Mode()&os.ModeNamedPipe == 0 {
			add("%s is no longer a FIFO (%v)", f, err)
		}
	}
	abs := func(rel string) string { return filepath.Join(w.root, filepath.FromSlash(rel)) }
	for _, rel := range []string{nrLinkReg, nrLinkFifo, nrDangling} {
		if info, err := os.Lstat(abs(rel)); err != nil || info.Mode()&os.ModeSymlink == 0 {
			add("%s is no longer a symlink (%v)", rel, err)
		}
	}
	if got, err := os.ReadFile(abs(nrPlain)); err != nil || !bytes.Equal(got, []byte(nrPlainBytes)) {
		add("%s changed (%v)", nrPlain, err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(w.root); err != nil {
		add("load manifest: %v", err)
	} else if _, ok := mgr.GetEntry(nrRegFifo); !ok {
		add("the manifest lost its entry for the registered FIFO %s", nrRegFifo)
	}
	for _, line := range o.report {
		if strings.Contains(line, "removed:") && strings.Contains(line, "nonregular") {
			add("a non-regular path was reported removed: %s", line)
		}
	}
	return problems
}

func TestNonRegularFilesAreKeptWithoutBeingRead(t *testing.T) {
	w := nonRegularFixture(t)
	o := runNonRegularMigration(t, w)
	for _, p := range nonRegularProblems(w, o) {
		t.Error(p)
	}

	// The pure assertion goes red in front of wrong outcomes.
	t.Run("MutantHungIsCaught", func(t *testing.T) {
		if len(nonRegularProblems(w, nonRegularOutcome{hung: true})) == 0 {
			t.Errorf("a migration that never returned passed")
		}
	})
	if !o.hung {
		t.Run("MutantReadFailedReasonIsCaught", func(t *testing.T) {
			bad := o
			bad.rc = &migrationRunContext{remaining: append([]remainingRow(nil), o.rc.remaining...)}
			for i := range bad.rc.remaining {
				if bad.rc.remaining[i].Rel == nrFifo {
					bad.rc.remaining[i].Reason = keepReadFailed
				}
			}
			if len(nonRegularProblems(w, bad)) == 0 {
				t.Errorf("a FIFO kept under another reason passed")
			}
		})
		t.Run("MutantFifoNotListedIsCaught", func(t *testing.T) {
			bad := o
			bad.rc = &migrationRunContext{}
			for _, r := range o.rc.remaining {
				if r.Rel != nrRegFifo {
					bad.rc.remaining = append(bad.rc.remaining, r)
				}
			}
			if len(nonRegularProblems(w, bad)) == 0 {
				t.Errorf("a registered FIFO without a row passed")
			}
		})
		t.Run("MutantSymlinkToRegularChangesBehaviourIsCaught", func(t *testing.T) {
			bad := o
			bad.rc = &migrationRunContext{remaining: append([]remainingRow(nil), o.rc.remaining...)}
			for i := range bad.rc.remaining {
				if bad.rc.remaining[i].Rel == nrLinkReg {
					bad.rc.remaining[i].Reason = nonRegularReason
				}
			}
			if len(nonRegularProblems(w, bad)) == 0 {
				t.Errorf("a symlink to a regular file treated as non-regular passed: its behaviour must not change")
			}
		})
	}
}
