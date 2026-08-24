package quality

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Card t218 — a step killed because the CALLER's deadline expired used to be
// reported as if the step had blown its own budget.
//
// context propagates a parent's cancellation to every child, so stepCtx.Err()
// reads DeadlineExceeded in both cases. The old reason string never looked at
// the parent, so a 30s dispatcher budget expiring mid-`go test` produced
// "go test exceeded 2m0s" — 30 seconds elapsed, a 2-minute budget blamed.
func TestRunStep_ParentDeadlineIsNotBlamedOnTheStep(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX shell to sleep past the parent deadline")
	}
	g := NewQualityGate(DefaultGateConfig())

	// Parent budget far shorter than the step's own: whatever kills the step,
	// it is not the step budget.
	parent, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ok, msg := g.runStep(parent, "go test", 2*time.Minute, "sh", "-c", "sleep 5")
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
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX shell to sleep past the step deadline")
	}
	g := NewQualityGate(DefaultGateConfig())

	parent, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	ok, msg := g.runStep(parent, "go test", 100*time.Millisecond, "sh", "-c", "sleep 5")
	if ok {
		t.Fatal("a step exceeding its own budget must fail")
	}
	if !strings.Contains(msg, "go test exceeded 100ms") {
		t.Errorf("reason must name the step budget, got: %q", msg)
	}
}
