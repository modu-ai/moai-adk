package quality

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// helperSleepEnv switches the test binary into "sleep and exit" mode so the
// timeout tests below have a long-running child process on every platform,
// without depending on a POSIX shell.
const helperSleepEnv = "MOAI_TEST_HELPER_SLEEP"

// TestHelperSleep is not a test: re-executed as a child by sleeperCommand, it
// blocks past any deadline the parent sets. Without the env marker it returns
// immediately, so a normal suite run does nothing.
func TestHelperSleep(t *testing.T) {
	if os.Getenv(helperSleepEnv) != "1" {
		t.Skip("helper process only")
	}
	time.Sleep(30 * time.Second)
}

// sleeperCommand returns the argv of a process that outlives any timeout used
// in this file: this same test binary, run with only the helper selected.
func sleeperCommand() (string, []string) {
	return os.Args[0], []string{"-test.run=^TestHelperSleep$", "-test.timeout=60s"}
}

func sleeperEnv(t *testing.T) {
	t.Helper()
	t.Setenv(helperSleepEnv, "1")
}

// Card t218 — a step killed because the CALLER's deadline expired used to be
// reported as if the step had blown its own budget.
//
// context propagates a parent's cancellation to every child, so stepCtx.Err()
// reads DeadlineExceeded in both cases. The old reason string never looked at
// the parent, so a 30s dispatcher budget expiring mid-`go test` produced
// "go test exceeded 2m0s" — 30 seconds elapsed, a 2-minute budget blamed.
func TestRunStep_ParentDeadlineIsNotBlamedOnTheStep(t *testing.T) {
	sleeperEnv(t)
	g := NewQualityGate(DefaultGateConfig())

	// Parent budget far shorter than the step's own: whatever kills the step,
	// it is not the step budget.
	parent, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	name, args := sleeperCommand()
	ok, msg := g.runStep(parent, "go test", "", 2*time.Minute, name, args...)
	if ok {
		t.Fatal("a step outliving the parent deadline must fail")
	}
	if strings.Contains(msg, "go test exceeded 2m0s") {
		t.Errorf("reason blames the step's own budget for a parent-deadline kill: %q", msg)
	}
	if !strings.Contains(msg, "overall") {
		t.Errorf("reason must name the overall (caller) budget as the cause, got: %q", msg)
	}
}

// The step's own budget must still be named when the step really does exceed it.
func TestRunStep_StepDeadlineStillBlamesTheStep(t *testing.T) {
	sleeperEnv(t)
	g := NewQualityGate(DefaultGateConfig())

	parent, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	name, args := sleeperCommand()
	ok, msg := g.runStep(parent, "go test", "", 100*time.Millisecond, name, args...)
	if ok {
		t.Fatal("a step exceeding its own budget must fail")
	}
	if !strings.Contains(msg, "go test exceeded 100ms") {
		t.Errorf("reason must name the step budget, got: %q", msg)
	}
}
