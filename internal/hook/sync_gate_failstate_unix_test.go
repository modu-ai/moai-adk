//go:build !windows

package hook

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Interrupted-run harness for the sync-phase gate (plan.md §F M1). Unix-only:
// it needs process groups and signals.

// sgfHarnessGuard bounds a run whose marker never appears and whose hook never
// exits. The sleeping stub bounds itself to about 30s, so a healthy run ends
// long before this guard; the guard exists only so a broken hook cannot hang
// the test binary.
const sgfHarnessGuard = 90 * time.Second

// runInterrupted executes one interrupted run and reports whether this run's
// own marker was observed.
//  1. The marker file is deleted before the run starts.
//  2. The hook starts in its own process group.
//  3. The harness waits for whichever comes first: the marker appears, or the
//     hook exits on its own. It never waits on a state-file token.
//  4. On the marker it kills the whole process group (a kill is also
//     registered with t.Cleanup). The caller then ages the record.
func (f *sgfFixture) runInterrupted(label, stdin string) bool {
	f.t.Helper()
	if err := os.Remove(f.marker); err != nil && !os.IsNotExist(err) {
		f.t.Errorf("%s: delete marker: %v", label, err)
	}
	cmd := exec.Command("bash", f.script)
	cmd.Dir = f.repo
	cmd.Env = f.env()
	cmd.Stdin = strings.NewReader(stdin)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = 5 * time.Second
	if err := cmd.Start(); err != nil {
		f.t.Errorf("%s: start hook: %v", label, err)
		return false
	}
	pid := cmd.Process.Pid
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	f.t.Cleanup(func() {
		select {
		case <-done:
		default:
			_ = syscall.Kill(-pid, syscall.SIGKILL)
			<-done
		}
	})
	markerSeen := func() bool {
		_, err := os.Stat(f.marker)
		return err == nil
	}
	guard := time.NewTimer(sgfHarnessGuard)
	defer guard.Stop()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		if markerSeen() {
			_ = syscall.Kill(-pid, syscall.SIGKILL)
			<-done
			f.t.Logf("%s: marker observed for this run; process group killed", label)
			return true
		}
		select {
		case <-done:
			if markerSeen() {
				f.t.Logf("%s: marker observed for this run; hook had already exited", label)
				return true
			}
			f.t.Logf("%s: marker not observed; hook exited on its own; output=%q", label, out.String())
			return false
		case <-guard.C:
			_ = syscall.Kill(-pid, syscall.SIGKILL)
			<-done
			f.t.Errorf("%s: harness guard %s elapsed with neither marker nor exit", label, sgfHarnessGuard)
			return false
		case <-tick.C:
		}
	}
}

// sgfBuildExhaustedRetry drives runs 1 and 2 of AC-006c: each run is
// interrupted by the harness and the record is then aged to 120 s. It returns
// whether each run observed its own marker.
func sgfBuildExhaustedRetry(f *sgfFixture, prefix string) (bool, bool) {
	f.t.Helper()
	f.setStub(sgfStubSpec{exits: map[string]int{"vet": 1}, writeMarker: true, sleep: true})
	m1 := f.runInterrupted(prefix+" run 1", "{}")
	f.ageRecord(prefix+" run 1 aging", sgfStaleAge)
	m2 := f.runInterrupted(prefix+" run 2", "{}")
	f.ageRecord(prefix+" run 2 aging", sgfStaleAge)
	return m1, m2
}

// TestSyncGateFailState_AC006c_RetryBoundAndNotice — AC-006c (release-blocking).
func TestSyncGateFailState_AC006c_RetryBoundAndNotice(t *testing.T) {
	tag := "AC-006c [" + rbClass + "]"
	f := newSGFFixture(t, sgfOpts{vetExit: 1})
	m1, m2 := sgfBuildExhaustedRetry(f, "AC-006c")
	if !m1 {
		t.Errorf("%s assertion 1: run 1 did not observe its marker (setup sanity)", tag)
	}
	if !m2 {
		t.Errorf("%s assertion 2: run 2 did not observe its own marker — the one allowed stale re-run did not start", tag)
	}
	f.setStub(sgfStubSpec{defaultExit: 1})
	if err := os.Remove(f.marker); err != nil && !os.IsNotExist(err) {
		t.Errorf("%s: delete marker before run 3: %v", tag, err)
	}
	before := f.count()
	out, code := f.run("{}")
	delta := f.count() - before
	t.Logf("%s run 3 stdout (%d bytes)=%q stub delta=%d exit=%d", tag, len(out), out, delta, code)
	if delta != 0 {
		t.Errorf("%s assertion 3: run 3 invoked the stub %d time(s); want 0", tag, delta)
	}
	if sgfHasDecision(out) {
		t.Errorf("%s assertion 4: run 3 stdout carries \"decision\"; stdout=%q", tag, out)
	}
	if !sgfHasSystemMessage(out) || !sgfNamesExhaustedRetry(out) {
		t.Errorf("%s assertion 5: run 3 stdout is not a systemMessage saying the gate run has not completed and naming deletion of %s; stdout=%q", tag, sgfRecordName, out)
	}
}

// TestSyncGateFailState_AC015_NoticesNeverBlockOrConsumeCap — AC-015 N1, N2a, N2b (release-blocking).
func TestSyncGateFailState_AC015_NoticesNeverBlockOrConsumeCap(t *testing.T) {
	const calls = 9
	swept := 0
	stdinFor := func(i int) string {
		if i%2 == 0 {
			return "{}"
		}
		return `{"stop_hook_active":true}`
	}
	type sweepResult struct {
		nonZeroExit, missingSys, decisions, missingText int
	}
	sweep := func(f *sgfFixture, before func(i int)) sweepResult {
		var r sweepResult
		for i := 0; i < calls; i++ {
			if before != nil {
				before(i)
			}
			out, code := f.run(stdinFor(i))
			swept++
			if code != 0 {
				r.nonZeroExit++
			}
			if !sgfHasSystemMessage(out) {
				r.missingSys++
			}
			if sgfHasDecision(out) {
				r.decisions++
			}
			if !sgfNamesExhaustedRetry(out) {
				r.missingText++
			}
			f.t.Logf("call %d stdin=%s exit=%d stdout=%q", i+1, stdinFor(i), code, out)
		}
		return r
	}
	report := func(t *testing.T, tag string, r sweepResult, stubDelta int) {
		t.Helper()
		t.Logf("%s summary: nonzero-exit=%d missing-systemMessage=%d decision-count=%d stub-delta=%d", tag, r.nonZeroExit, r.missingSys, r.decisions, stubDelta)
		if r.nonZeroExit != 0 {
			t.Errorf("%s: %d of %d runs exited non-zero", tag, r.nonZeroExit, calls)
		}
		if r.missingSys != 0 {
			t.Errorf("%s: %d of %d stdouts lack \"systemMessage\"", tag, r.missingSys, calls)
		}
		if r.decisions != 0 {
			t.Errorf("%s: %d of %d stdouts carry \"decision\"; want exactly 0", tag, r.decisions, calls)
		}
		if stubDelta != 0 {
			t.Errorf("%s: stub count grew by %d across the %d runs; want 0", tag, stubDelta, calls)
		}
	}

	t.Run("N1", func(t *testing.T) {
		tag := "AC-015 N1 [" + rbClass + "]"
		f := newSGFFixture(t, sgfOpts{vetExit: 1})
		m1, m2 := sgfBuildExhaustedRetry(f, "AC-015 N1")
		if !m1 {
			t.Errorf("%s setup assertion 1: run 1 did not observe its marker", tag)
		}
		if !m2 {
			t.Errorf("%s setup assertion 2: run 2 did not observe its own marker — the one allowed stale re-run did not start", tag)
		}
		f.setStub(sgfStubSpec{defaultExit: 1, writeMarker: true})
		before := f.count()
		r := sweep(f, func(int) {
			if err := os.Remove(f.marker); err != nil && !os.IsNotExist(err) {
				f.t.Errorf("%s: delete marker: %v", tag, err)
			}
		})
		report(t, tag, r, f.count()-before)
		if r.missingText != 0 {
			t.Errorf("%s: %d of %d notices do not say the gate run has not completed and name deleting %s", tag, r.missingText, calls, sgfRecordName)
		}
	})

	t.Run("N2a", func(t *testing.T) {
		tag := "AC-015 N2a [" + rbClass + "]"
		f := newSGFFixture(t, sgfOpts{vetExit: 1})
		head := f.rev("HEAD")
		f.writeRecord(head + " running\n")
		before := f.count()
		r := sweep(f, func(i int) { f.ageRecord(tag+" mtime refresh", 0) })
		report(t, tag, r, f.count()-before)
		if rec, ok := f.readRecord(tag); ok && rec != head+" running" {
			t.Errorf("%s: record after run %d = %q; want %q", tag, calls, rec, head+" running")
		}
	})

	t.Run("N2b", func(t *testing.T) {
		tag := "AC-015 N2b [" + rbClass + "]"
		f := newSGFFixture(t, sgfOpts{vetExit: 1})
		head := f.rev("HEAD")
		before := f.count()
		r := sweep(f, func(i int) {
			f.writeRecord(head + " running\n")
			f.ageRecord(tag+" rewrite", 0)
		})
		report(t, tag, r, f.count()-before)
	})

	t.Logf("AC-015 swept invocations: %d (want %d)", swept, 3*calls)
	if swept != 3*calls {
		t.Errorf("AC-015: swept %d invocations; want %d (partial sweep)", swept, 3*calls)
	}
}
