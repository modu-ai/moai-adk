package cli

// gate_lock_cli_test.go — SPEC-GATE-THREE-AXES-001 M3, `moai gate` level
// (card t235).
//
// What the unit file cannot see: the lock wrapping an actual gate run. Two
// concurrent runGate invocations against one project directory are two
// separate runs presenting the same surface to the kernel the lock uses, so
// the serialization, the degradation, and the never-fails-the-run property
// are verified here end to end — through runGate, reading real stderr, with
// the M1 execution summary's timestamps as the observation instrument.

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// helperGateLockSleepEnv is the channel between this test and its re-executed
// child: the child sleeps the configured milliseconds so a gate step's
// wall-clock span is known in advance.
const helperGateLockSleepEnv = "MOAI_TEST_GATE_LOCK_SLEEP_MS"

// TestGateLockHelperSleep is not a test. Re-executed as a child by
// gateSleeperCommand, it sleeps for the configured span. The child bounds
// itself: it exits on its own, and -test.timeout caps it from outside.
func TestGateLockHelperSleep(t *testing.T) {
	raw := os.Getenv(helperGateLockSleepEnv)
	if raw == "" {
		t.Skip("helper process only")
	}
	ms, err := strconv.Atoi(raw)
	if err != nil {
		t.Fatalf("%s = %q: %v", helperGateLockSleepEnv, raw, err)
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// gateSleeperCommand returns the whitespace-separated command line the
// gate.typecheck.command axis is parsed from — the shape that re-executes
// this test binary as a bounded sleeper.
func gateSleeperCommand() string {
	return strings.Join([]string{
		os.Args[0],
		"-test.run=^TestGateLockHelperSleep$",
		"-test.timeout=60s",
	}, " ")
}

// gateLockFixture lays down a passing Go project with a gate.yaml whose
// lock_wait is lockWaitSeconds, and — when sleepMS > 0 — whose typecheck axis
// runs the bounded sleeper for sleepMS milliseconds. The fixture becomes the
// active CLAUDE_PROJECT_DIR for the test (t.Setenv forces serial, which these
// timing-sensitive tests want anyway).
func gateLockFixture(t *testing.T, lockWaitSeconds, sleepMS int) string {
	t.Helper()

	dir := t.TempDir()
	// No `go` directive: naming a version above the installed toolchain sends
	// the vet and test steps looking for another toolchain to download.
	writeGateFixtureFile(t, dir, "go.mod", "module t235gatelock\n")
	writeGateFixtureFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")

	var b strings.Builder
	fmt.Fprintf(&b, "gate:\n  enabled: true\n  timeouts:\n    lock_wait: %d\n", lockWaitSeconds)
	if sleepMS > 0 {
		fmt.Fprintf(&b, "  typecheck:\n    enabled: true\n    command: %q\n", gateSleeperCommand())
	}
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sections, "gate.yaml"), []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write gate.yaml: %v", err)
	}

	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	if sleepMS > 0 {
		t.Setenv(helperGateLockSleepEnv, strconv.Itoa(sleepMS))
	}
	return dir
}

// newGateRunCmd builds a runGate target with its own stderr buffer, so two
// concurrent runs never share the package-level gateCmd's writer.
func newGateRunCmd() (*cobra.Command, *bytes.Buffer) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetErr(&buf)
	return cmd, &buf
}

// waitUntilHolderRecorded polls until the lock artifact names a holder, so a
// second run started after it genuinely contends rather than racing the
// first's acquisition.
func waitUntilHolderRecorded(t *testing.T, dir string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := ReadGateLockHolder(dir); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("first run never recorded a holder in the lock artifact under %s", dir)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// executedStep is one executed summary row's measured facts.
type executedStep struct {
	label string
	start time.Time
	dur   time.Duration
}

// executedRowRe matches the summary rows the M1 renderer writes for executed
// steps: "  - <label>: executed in <dur> at <instant> — <command>". The
// instant is what orders two runs' windows against each other.
var executedRowRe = regexp.MustCompile(`^- ([^:]+): executed in (\S+) at (\S+)(?: — .*)?$`)

// parseExecutedSteps extracts every executed row's start instant and measured
// duration from a run's stderr.
func parseExecutedSteps(t *testing.T, stderr string) []executedStep {
	t.Helper()
	var steps []executedStep
	for _, line := range strings.Split(stderr, "\n") {
		m := executedRowRe.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		dur, err := time.ParseDuration(m[2])
		if err != nil {
			t.Fatalf("parse duration %q from summary row %q: %v", m[2], line, err)
		}
		start, err := time.Parse(time.RFC3339, m[3])
		if err != nil {
			t.Fatalf("parse start instant %q from summary row %q: %v", m[3], line, err)
		}
		steps = append(steps, executedStep{label: m[1], start: start, dur: dur})
	}
	return steps
}

// --- AC-GTA-011 --------------------------------------------------------------

// TestGateCmd_SecondRunWaitsForFirst — AC-GTA-011.
//
// Two gate runs against one project directory, the first executing a step
// that runs for a known duration; the second starts within it. The second's
// first executed step, as timestamped in its own summary, must begin after
// the first run released the lock — which is after the first's LAST executed
// step ended. The assertion binds to the summary's execution timestamps, not
// to the presence of a "waiting" message: emitting a waiting line while
// running anyway is the criterion's named mutant, and it fails here.
func TestGateCmd_SecondRunWaitsForFirst(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("go is not on PATH, so the Go toolchain's steps cannot run: %v", err)
	}
	dir := gateLockFixture(t, 30, 2500)

	cmdA, bufA := newGateRunCmd()
	cmdB, bufB := newGateRunCmd()

	var errA error
	doneA := make(chan struct{})
	go func() {
		defer close(doneA)
		errA = runGate(cmdA, nil)
	}()
	waitUntilHolderRecorded(t, dir)

	var errB error
	doneB := make(chan struct{})
	go func() {
		defer close(doneB)
		errB = runGate(cmdB, nil)
	}()

	<-doneA
	if errA != nil {
		t.Fatalf("first run failed: %v\nstderr:\n%s", errA, bufA.String())
	}
	<-doneB
	if errB != nil {
		t.Fatalf("second run failed: %v\nstderr:\n%s", errB, bufB.String())
	}

	// The second run waited, and its notice named the holder the artifact
	// records — the first run, which is this same process in this fixture.
	if want := fmt.Sprintf("held by pid %d", os.Getpid()); !strings.Contains(bufB.String(), want) {
		t.Errorf("second run's stderr does not carry the waiting notice naming the holder; want %q in:\n%s", want, bufB.String())
	}

	aSteps := parseExecutedSteps(t, bufA.String())
	bSteps := parseExecutedSteps(t, bufB.String())
	if len(aSteps) == 0 {
		t.Fatalf("first run's summary carries no executed rows:\n%s", bufA.String())
	}
	if len(bSteps) == 0 {
		t.Fatalf("second run's summary carries no executed rows:\n%s", bufB.String())
	}

	aLastEnd := aSteps[0].start.Add(aSteps[0].dur)
	for _, s := range aSteps[1:] {
		if end := s.start.Add(s.dur); end.After(aLastEnd) {
			aLastEnd = end
		}
	}
	bFirstStart := bSteps[0].start
	for _, s := range bSteps[1:] {
		if s.start.Before(bFirstStart) {
			bFirstStart = s.start
		}
	}
	if !bFirstStart.After(aLastEnd) {
		t.Fatalf("the runs' execution windows overlap: second run's first executed step began %s, first run's last executed step ended %s\nfirst stderr:\n%s\nsecond stderr:\n%s",
			bFirstStart.Format(time.RFC3339Nano), aLastEnd.Format(time.RFC3339Nano), bufA.String(), bufB.String())
	}
}

// --- AC-GTA-013 / AC-GTA-015 (verdict half) -----------------------------------

// TestGateCmd_WaitExpiryRunsUnserializedWithOwnVerdict — AC-GTA-013 (CLI
// half) and the verdict half of AC-GTA-015.
//
// A live holder outlasting the wait budget: the run begins executing only
// after the budget elapsed, its output states that it ran unserialized, and
// the exit is the GATE's own verdict — nil on a passing project, the quality
// error on a failing one. Failing the run when the wait expires is AC-GTA-013's
// named Mutant B, and it fails the second subtest.
func TestGateCmd_WaitExpiryRunsUnserializedWithOwnVerdict(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("go is not on PATH, so the Go toolchain's steps cannot run: %v", err)
	}

	t.Run("passing project returns the gate's pass verdict", func(t *testing.T) {
		dir := gateLockFixture(t, 1, 0)
		holder, err := AcquireGateLock(dir)
		if err != nil {
			t.Fatalf("holder acquire failed: %v", err)
		}
		defer func() {
			if err := holder.Release(); err != nil {
				t.Errorf("release: %v", err)
			}
		}()

		cmd, buf := newGateRunCmd()
		started := time.Now()
		err = runGate(cmd, nil)
		if err != nil {
			t.Fatalf("the lock's wait expiry failed a passing run: %v\nstderr:\n%s", err, buf.String())
		}
		if elapsed := time.Since(started); elapsed < time.Second {
			t.Errorf("run against a held lock finished in %s, before the 1s wait budget could have elapsed", elapsed)
		}
		if !strings.Contains(buf.String(), "unserialized") {
			t.Errorf("run's output does not state that it ran unserialized:\n%s", buf.String())
		}
	})

	t.Run("failing project returns the gate's fail verdict", func(t *testing.T) {
		dir := gateLockFixture(t, 1, 0)
		// A syntax error: go vet fails, and the gate's own verdict is FAIL.
		writeGateFixtureFile(t, dir, "main.go", "package main\n\nfunc main( {}\n")
		holder, err := AcquireGateLock(dir)
		if err != nil {
			t.Fatalf("holder acquire failed: %v", err)
		}
		defer func() {
			if err := holder.Release(); err != nil {
				t.Errorf("release: %v", err)
			}
		}()

		cmd, buf := newGateRunCmd()
		err = runGate(cmd, nil)
		if err == nil {
			t.Fatalf("the lock masked the gate's own failure — a failing project passed while the lock was held\nstderr:\n%s", buf.String())
		}
		if !strings.Contains(buf.String(), "unserialized") {
			t.Errorf("run's output does not state that it ran unserialized:\n%s", buf.String())
		}
	})
}

// --- AC-GTA-015 (read-only directory) -----------------------------------------

// TestGateCmd_ReadOnlyLockDirReturnsGateVerdict — AC-GTA-015.
//
// The lock artifact's directory cannot be created: the gate still executes,
// returns its own verdict, and reports that the lock was unavailable.
// Propagating the lock error as the command's error is the criterion's named
// mutant.
func TestGateCmd_ReadOnlyLockDirReturnsGateVerdict(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("posix permission bits are advisory on windows — os.Chmod cannot make the state directory read-only, so the fixture premise is unbuildable there")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("go is not on PATH, so the Go toolchain's steps cannot run: %v", err)
	}
	dir := gateLockFixture(t, 30, 0)

	moaiDir := filepath.Join(dir, ".moai")
	// Reading .moai stays possible — the config loader below still works, so
	// the gate runs under the fixture's own config; creating .moai/state is
	// what must fail.
	if err := os.Chmod(moaiDir, 0o500); err != nil {
		t.Skipf("cannot make .moai read-only on this filesystem: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(moaiDir, 0o755); err != nil {
			t.Logf("restore .moai permissions: %v", err)
		}
	})

	cmd, buf := newGateRunCmd()
	if err := runGate(cmd, nil); err != nil {
		t.Fatalf("a read-only lock directory failed the run instead of degrading: %v\nstderr:\n%s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "unavailable") {
		t.Errorf("run's output does not report the lock as unavailable:\n%s", out)
	}
	if !strings.Contains(out, "unserialized") {
		t.Errorf("run's output does not state that it ran unserialized:\n%s", out)
	}
}

// --- AC-GTA-016 --------------------------------------------------------------

// TestGateCmd_DegradedRunNeverReacquires — AC-GTA-016.
//
// A run degrades at its wait budget; the original holder releases while the
// degraded run is still executing; a third run started at that moment must
// acquire immediately. Opportunistically re-acquiring after degrading is the
// criterion's named mutant, and it makes the third run wait — which this
// test catches.
func TestGateCmd_DegradedRunNeverReacquires(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("go is not on PATH, so the Go toolchain's steps cannot run: %v", err)
	}
	// Wait budget 1s; the degraded run's typecheck sleeps 2s, so the original
	// holder can release while the degraded run is still executing.
	dir := gateLockFixture(t, 1, 2000)

	holder, err := AcquireGateLock(dir)
	if err != nil {
		t.Fatalf("original holder acquire failed: %v", err)
	}

	cmdB, bufB := newGateRunCmd()
	var errB error
	doneB := make(chan struct{})
	go func() {
		defer close(doneB)
		errB = runGate(cmdB, nil)
	}()

	// The degraded run's own budget expires at ~1s and its typecheck runs
	// until ~3s. Releasing at 1.6s lands inside that window.
	time.Sleep(1600 * time.Millisecond)
	if err := holder.Release(); err != nil {
		t.Fatalf("original holder release failed: %v", err)
	}
	// A gap between the release and the third run's start, so a run that
	// opportunistically re-acquires after degrading has had every chance to
	// take the freed lock: the third run then waits, and the criterion fails
	// it. A run that left the lock free acquires on its first attempt
	// whenever it starts — the gap costs the compliant shape nothing.
	time.Sleep(300 * time.Millisecond)

	var bufC bytes.Buffer
	res := waitForGateLock(dir, 30*time.Second, &bufC)
	if res.lock == nil {
		t.Fatalf("third run could not acquire after the original holder released: %s\nnotices:\n%s", res.line, bufC.String())
	}
	if err := res.lock.Release(); err != nil {
		t.Errorf("third run release failed: %v", err)
	}
	if res.waited > time.Second {
		t.Errorf("third run waited %s to acquire — the degraded run must have left the lock free for it", res.waited)
	}

	<-doneB
	if errB != nil {
		t.Fatalf("degraded run failed its gate: %v\nstderr:\n%s", errB, bufB.String())
	}
	if !strings.Contains(bufB.String(), "unserialized") {
		t.Errorf("degraded run's output does not state that it ran unserialized:\n%s", bufB.String())
	}
}
