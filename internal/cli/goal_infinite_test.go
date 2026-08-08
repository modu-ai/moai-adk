package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/goal"
)

// driveGoalArmExec runs `moai goal arm <args>` through the registered command
// tree (same RunE path the real rootCmd registers) and reports the exit error
// + stdout/stderr buffer. CLAUDE_PROJECT_DIR is pointed at a fresh temp dir so
// the arm writes a per-session goal file under t.TempDir() (never the real
// project state dir).
func driveGoalArmExec(t *testing.T, args []string) (out string, exitErr error) {
	t.Helper()
	root, buf := newGoalTestRoot()
	root.SetArgs(append([]string{"goal", "arm"}, args...))
	exitErr = root.Execute()
	return buf.String(), exitErr
}

// loadArmedGoal reads back the per-session goal file the arm verb wrote, keyed
// by sessionID. Returns nil when no file exists.
func loadArmedGoal(t *testing.T, projectRoot, sessionID string) *goal.Goal {
	t.Helper()
	g, err := goal.LoadGoal(projectRoot, sessionID)
	if err != nil {
		t.Fatalf("load armed goal: %v", err)
	}
	return g
}

// TestAC001_MaxTurnsZeroDisablesCeiling asserts that arming with --max-turns 0
// produces Ceiling.MaxTurns == 0 (the infinite entry point) AND that the
// evaluator's `> 0` guard skips the ceiling check across simulated turns > 30.
// The arm-verb half lives here; the evaluator half lives in internal/goal
// (TestAC001_EvaluatorSkipsCeilingAtMaxTurnsZero).
func TestAC001_MaxTurnsZeroDisablesCeiling(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	out, exitErr := driveGoalArmExec(t, []string{
		"--session", "ac001",
		"--max-turns", "0",
		"--max-duration", "3600", // a real bound so the arm-time reject does not fire
		"echo hi exits 0",
	})
	if exitErr != nil {
		t.Fatalf("goal arm --max-turns 0: unexpected error %v (out=%q)", exitErr, out)
	}

	g := loadArmedGoal(t, tmp, "ac001")
	if g == nil {
		t.Fatalf("goal arm --max-turns 0: no goal file written (out=%q)", out)
	}
	if g.Ceiling.MaxTurns != 0 {
		t.Errorf("AC-001: Ceiling.MaxTurns = %d, want 0 (infinite entry point)", g.Ceiling.MaxTurns)
	}
}

// TestAC002_MaxTurnsOmittedDefaultsTo30 asserts backward compatibility: when
// --max-turns is omitted, Ceiling.MaxTurns == 30 (DefaultMaxTurns) — zero
// behavior delta.
func TestAC002_MaxTurnsOmittedDefaultsTo30(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	out, exitErr := driveGoalArmExec(t, []string{
		"--session", "ac002",
		"echo hi exits 0",
	})
	if exitErr != nil {
		t.Fatalf("goal arm (no --max-turns): unexpected error %v (out=%q)", exitErr, out)
	}

	g := loadArmedGoal(t, tmp, "ac002")
	if g == nil {
		t.Fatalf("goal arm: no goal file written (out=%q)", out)
	}
	if g.Ceiling.MaxTurns != goal.DefaultMaxTurns {
		t.Errorf("AC-002: Ceiling.MaxTurns = %d, want default %d", g.Ceiling.MaxTurns, goal.DefaultMaxTurns)
	}
}

// TestAC011_ArmRejectsUnboundedInfinite asserts the fail-closed arm-time
// enforcement: `--max-turns 0` with NEITHER --max-duration NOR --cost-cap is
// REJECTED (non-zero exit + stderr naming the missing bound + no goal file).
func TestAC011_ArmRejectsUnboundedInfinite(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	out, exitErr := driveGoalArmExec(t, []string{
		"--session", "ac011",
		"--max-turns", "0",
		"echo hi exits 0",
	})
	if exitErr == nil {
		t.Fatalf("AC-011: goal arm --max-turns 0 without a bound: expected non-zero exit, got nil (out=%q)", out)
	}
	if !strings.Contains(strings.ToLower(out), "max-duration") && !strings.Contains(strings.ToLower(out), "cost-cap") {
		t.Errorf("AC-011: stderr must name the missing bound (--max-duration / --cost-cap); got %q", out)
	}

	// No goal state file may be written on a rejected arm.
	if g := loadArmedGoal(t, tmp, "ac011"); g != nil {
		t.Errorf("AC-011: a goal file was written despite the reject (goal=%+v)", g)
	}
}

// TestAC011_ArmAcceptsBoundedInfinite asserts the positive path: --max-turns 0
// WITH --max-duration (or --cost-cap) arms successfully (exit 0 + goal written).
// This is the AC-005 positive companion.
func TestAC011_ArmAcceptsBoundedInfinite(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	// wall-clock bound
	out, exitErr := driveGoalArmExec(t, []string{
		"--session", "ac011-dur",
		"--max-turns", "0",
		"--max-duration", "3600",
		"echo hi exits 0",
	})
	if exitErr != nil {
		t.Fatalf("AC-011 positive (--max-duration): unexpected error %v (out=%q)", exitErr, out)
	}
	if g := loadArmedGoal(t, tmp, "ac011-dur"); g == nil {
		t.Fatalf("AC-011 positive (--max-duration): no goal file written (out=%q)", out)
	}

	// cost-cap is RECORDED-ONLY (SPEC-INFINITE-GOAL-001 REQ-004 / D1):
	// enforcement is a follow-up (§D.5), so cost-cap-alone does NOT satisfy the
	// real-bound requirement and MUST be rejected at arm time. The rejection
	// names --max-duration as the required real wall-clock bound.
	out2, exitErr2 := driveGoalArmExec(t, []string{
		"--session", "ac011-cost",
		"--max-turns", "0",
		"--cost-cap", "1000",
		"echo hi exits 0",
	})
	if exitErr2 == nil {
		t.Fatalf("AC-011 (--cost-cap alone): expected reject (cost-cap is recorded-only), got success (out=%q)", out2)
	}
	if g := loadArmedGoal(t, tmp, "ac011-cost"); g != nil {
		t.Fatalf("AC-011 (--cost-cap alone): no goal file should be written on reject")
	}
}
