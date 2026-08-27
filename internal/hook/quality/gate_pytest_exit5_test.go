package quality

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// Issue #1265 (secondary finding) — pytest exits 5 for "no tests collected".
//
// The gate treated any non-zero exit as a failure, so a Python project with no
// tests yet could never commit: the test step reported "quality gate failed:
// pytest" for a run that found nothing wrong. Exit 5 is pytest's documented
// EXIT_NOTESTSCOLLECTED and carries no diagnostic, so it must pass the gate.
//
// The three cases below pin the fix narrowly: exit 5 passes ONLY for the pytest
// step, and pytest's real failure code (1) still fails.
func TestRunStep_PytestNoTestsCollectedPasses(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX shell to synthesise the exit code")
	}
	g := NewQualityGate(DefaultGateConfig())
	ok, msg := g.runStep(context.Background(), "pytest", "", 10*time.Second, "sh", "-c", "exit 5")
	if !ok {
		t.Fatalf("pytest exit 5 (no tests collected) must pass the gate, got failure: %s", msg)
	}
}

func TestRunStep_PytestRealFailureStillFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX shell to synthesise the exit code")
	}
	g := NewQualityGate(DefaultGateConfig())
	ok, msg := g.runStep(context.Background(), "pytest", "", 10*time.Second, "sh", "-c", "echo boom >&2; exit 1")
	if ok {
		t.Fatal("pytest exit 1 is a real test failure and must still fail the gate")
	}
	if msg == "" {
		t.Fatal("a failing step must carry a diagnostic message")
	}
}

func TestRunStep_NonPytestExit5StillFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX shell to synthesise the exit code")
	}
	g := NewQualityGate(DefaultGateConfig())
	ok, _ := g.runStep(context.Background(), "go test", "", 10*time.Second, "sh", "-c", "exit 5")
	if ok {
		t.Fatal("exit 5 is pytest-specific; other steps must still fail on it")
	}
}
