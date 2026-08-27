package quality

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The fixture for §D.2 of acceptance.md: a step command that spawns a
// descendant which outlives it and inherits the step's stdout and stderr. That
// inherited pipe is what blocks cmd.Wait past the deadline meant to bound it,
// so the descendant is what this milestone has to survive.
//
// Every process started here is bounded twice over: the grandchild carries its
// own -test.timeout, and the parent test registers a t.Cleanup kill. A trailing
// kill line would be skipped by every early return, so there is none.
const (
	// helperOrphanEnv switches the test binary into "spawn a grandchild and
	// exit" mode.
	helperOrphanEnv = "MOAI_TEST_HELPER_ORPHAN"
	// helperOrphanChildEnv switches it into the grandchild's own mode: hold the
	// inherited streams open and sleep.
	helperOrphanChildEnv = "MOAI_TEST_HELPER_ORPHAN_CHILD"
	// helperOrphanPIDPathEnv names the file the grandchild's PID is written to,
	// so the test can probe that PID afterwards (AC-GTA-009).
	helperOrphanPIDPathEnv = "MOAI_TEST_HELPER_ORPHAN_PIDFILE"
	// helperNoisyEnv switches it into "write to both streams and exit non-zero"
	// mode, the within-deadline step AC-GTA-010 preserves.
	helperNoisyEnv = "MOAI_TEST_HELPER_NOISY"
)

// orphanChildSleep far exceeds any step timeout plus grace used below, so no
// choice of T lets the descendant finish on its own before the bound.
const orphanChildSleep = 30 * time.Second

// TestHelperOrphan is not a test. Re-executed as the step's direct child, it
// starts a grandchild that inherits this process's stdout and stderr, records
// the grandchild's PID, and returns immediately — leaving the step's output
// pipes held by a process the step no longer owns.
func TestHelperOrphan(t *testing.T) {
	if os.Getenv(helperOrphanEnv) != "1" {
		t.Skip("helper process only")
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperOrphanChild$", "-test.timeout=60s")
	cmd.Env = append(os.Environ(), helperOrphanChildEnv+"=1", helperOrphanEnv+"=0")
	// The point of the fixture: the grandchild holds the step's own streams.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "helper could not start grandchild: %v\n", err)
		os.Exit(1)
	}
	if path := os.Getenv(helperOrphanPIDPathEnv); path != "" {
		_ = os.WriteFile(path, []byte(strconv.Itoa(cmd.Process.Pid)), 0o600)
	}
	os.Exit(0)
}

// TestHelperOrphanChild is not a test: it is the grandchild. Its own
// -test.timeout bounds it from outside, so a failure to kill it cannot leak a
// process beyond that bound.
func TestHelperOrphanChild(t *testing.T) {
	if os.Getenv(helperOrphanChildEnv) != "1" {
		t.Skip("helper process only")
	}
	time.Sleep(orphanChildSleep)
}

// TestHelperNoisy is not a test: it writes to both streams, reports the working
// directory it was launched in, and exits non-zero well inside any deadline.
// os.Exit keeps the test framework's own trailing output off the streams, so
// what the gate captures is exactly what is written here.
func TestHelperNoisy(t *testing.T) {
	if os.Getenv(helperNoisyEnv) != "1" {
		t.Skip("helper process only")
	}
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "unknown"
	}
	fmt.Fprint(os.Stderr, "stderr-line\n")
	_, _ = fmt.Fprintf(os.Stdout, "stdout-line cwd=%s\n", cwd)
	os.Exit(3)
}

// orphanCommand returns the argv of the step command that leaves an orphan
// behind. The -test.run narrowing is load-bearing: re-executing the test binary
// without it would re-run the whole suite from inside the suite.
func orphanCommand() (string, []string) {
	return os.Args[0], []string{"-test.run=^TestHelperOrphan$", "-test.timeout=60s"}
}

func noisyCommand() (string, []string) {
	return os.Args[0], []string{"-test.run=^TestHelperNoisy$", "-test.timeout=60s"}
}

// orphanFixture arms the helper env, returns the PID file path, and registers
// the cleanup that bounds the grandchild if the code under test failed to.
func orphanFixture(t *testing.T) string {
	t.Helper()
	pidPath := filepath.Join(t.TempDir(), "orphan.pid")
	t.Setenv(helperOrphanEnv, "1")
	t.Setenv(helperOrphanPIDPathEnv, pidPath)
	t.Cleanup(func() {
		pid, err := readOrphanPID(pidPath)
		if err != nil {
			return
		}
		if p, err := os.FindProcess(pid); err == nil {
			_ = p.Kill()
		}
	})
	return pidPath
}

func readOrphanPID(path string) (int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(raw)))
}

// AC-GTA-008 — the gate returns within a bounded grace period even while a
// descendant holds the step's inherited output stream.
func TestRunStep_ReturnsWithinGraceWhenADescendantHoldsTheStream(t *testing.T) {
	orphanFixture(t)

	const stepTimeout = 2 * time.Second
	bound := stepTimeout + stepWaitGrace

	g := NewQualityGate(DefaultGateConfig())
	parent, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	name, args := orphanCommand()
	started := time.Now()
	ok, msg := g.runStep(parent, "orphan", stepTimeout, name, args...)
	elapsed := time.Since(started)

	if elapsed >= bound {
		t.Fatalf("runStep returned after %s; it must return within T+G = %s (ok=%v, msg=%q)",
			elapsed.Round(time.Millisecond), bound, ok, msg)
	}
	if !strings.Contains(msg, descendantTerminationNote) {
		t.Errorf("the reason must state what happened to the step's descendants; got %q", msg)
	}
}

// AC-GTA-010, second half — a step that completes inside its deadline keeps its
// exit status, its captured output, and its working directory. The expected
// strings below were measured on the pre-change tree (ba22f41cf).
func TestRunStep_WithinDeadlineOutputAndCwdAreUnchanged(t *testing.T) {
	t.Setenv(helperNoisyEnv, "1")

	g := NewQualityGate(DefaultGateConfig())
	name, args := noisyCommand()
	ok, msg := g.runStep(context.Background(), "noisy", time.Minute, name, args...)

	if ok {
		t.Fatal("a step exiting non-zero inside its deadline must still fail")
	}
	dir := resolveQualityProjectDir(*g.config, "test")
	want := fmt.Sprintf("quality gate failed: noisy\n\nstderr-line\nstdout-line cwd=%s", dir)
	if msg != want {
		t.Errorf("within-deadline failure message changed:\n got %q\nwant %q", msg, want)
	}
}
