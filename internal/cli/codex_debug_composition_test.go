package cli

// codex_debug_composition_test.go — SPEC-CODEX-DEBUG-MODE-001 M4: the
// composition with t1378 REQ-012 (SPEC-CODEX-LANE-SLOTS-001). Under debug
// mode a codex lane launch prints the per-step pre-exec timing lines
// UNCONDITIONALLY — the slow-launch threshold does not gate them (REQ-014,
// AC-012) — and the lines are the shared collector's recorded steps (the
// single recording site). With debug OFF, the t1378 threshold behavior is
// frozen (REQ-015, AC-013): a fast launch prints nothing, a slow launch
// prints exactly the threshold report, and none of the debug-vocabulary
// steps leak into it.
//
// Both cells drive the REAL lane join through runCodexLaunch (staged
// discovery leader, git-inited fixture root) with the launch seams stubbed
// — no real codex process.

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/discovery"
	"github.com/spf13/cobra"
)

// stagedSingleLeader stages one verified leader whose run the lane join
// discovers and resumes (the record-absence fallback path).
func stagedSingleLeader(t *testing.T) []discovery.VerifiedLeader {
	t.Helper()
	return []discovery.VerifiedLeader{verifiedTestLeader("run-debug-lane")}
}

// stagedLaneEntry is the -f lane factory entry runCodexLaunch's lane wiring
// arms the collector for.
func stagedLaneEntry() factoryFlagParse {
	return factoryFlagParse{Enabled: true, LaneRole: true}
}

// driveCodexLaneLaunch runs one codex lane launch through runCodexLaunch with
// every seam stubbed, and returns the stderr the launcher wrote. The entry is a
// LaneRole factory entry handed straight to runCodexLaunch: runCodex routes -l
// to the relaunch loop and never reaches that branch. The join lands on the
// staged leader's run (the discovery fallback the shared join gate owns).
func driveCodexLaneLaunch(t *testing.T, debug bool, slowLaunchMS string) string {
	t.Helper()
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	// Set the threshold AFTER clearFactoryTestEnv: M2 (SPEC-TEST-ENV-HERMETIC-001)
	// added EnvMoaiFactorySlowLaunchMS to factoryAmbientEnvKeys, so the clear now
	// wipes a value the caller set before this helper ran. The threshold here is
	// the test's own deliberate value, not ambient lane state.
	t.Setenv(config.EnvMoaiFactorySlowLaunchMS, slowLaunchMS)
	stageDiscoveredLeaders(t, stagedSingleLeader(t))
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	c := &cobra.Command{Use: "codex"}
	c.SetContext(context.Background())
	errB := new(strings.Builder)
	c.SetOut(new(strings.Builder))
	c.SetErr(errB)
	err := runCodexLaunch(c, codexVerbLaunchCli, nil, false, codexWorktreeArg{},
		stagedLaneEntry(), debug)
	if err != nil {
		t.Fatalf("runCodexLaunch(lane, debug=%v): %v", debug, err)
	}
	if cap.count() != 1 {
		t.Fatalf("launches = %d, want 1", cap.count())
	}
	return errB.String()
}

// factoryLaunchLineCount counts the t1378 threshold-report lines.
func factoryLaunchLineCount(stderr string) int {
	n := 0
	for _, line := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(line, "factory launch:") {
			n++
		}
	}
	return n
}

// AC-012: under debug mode the per-step timing lines print regardless of the
// threshold — here the threshold is a full hour, so reportSlow is silent by
// construction, and the debug dump must still carry the collector's recorded
// lane steps.
func TestCodexDebugSupersedesLaunchThreshold(t *testing.T) {
	// A threshold no real pre-exec phase exceeds: the threshold path must
	// stay silent while the debug dump prints.
	t.Setenv(config.EnvMoaiFactorySlowLaunchMS, "3600000")
	stderr := driveCodexLaneLaunch(t, true, "3600000")
	if got := factoryLaunchLineCount(stderr); got != 0 {
		t.Errorf("threshold report printed under a one-hour threshold (%d lines) — REQ-015 boundary broken:\n%s", got, stderr)
	}
	for _, step := range []string{"lane claim", "exec handoff"} {
		if got := debugTraceStepLines(stderr, step); got != 1 {
			t.Errorf("debug dump lacks the %q timing line (%d found) — REQ-014:\n%s", step, got, stderr)
		}
	}
	if !strings.Contains(stderr, "took") {
		t.Errorf("debug timing lines carry no wall-time:\n%s", stderr)
	}
	if !strings.Contains(stderr, "label=") {
		t.Errorf("lane-claim trace lacks the claimed label detail:\n%s", stderr)
	}
}

// AC-013 part 1: with debug OFF the t1378 REQ-012 behavior is frozen.
//   - (a) a fast lane launch prints no timing lines and no trace lines;
//   - (b) a lane launch exceeding the threshold (threshold 0) prints exactly
//     the threshold report — every line under "factory launch:", none of the
//     debug-vocabulary steps in it (beginDebug measured nothing).
//     (Part 2 — the existing timing tests pass unmodified — runs as its own
//     command; see acceptance.md AC-013.)
func TestCodexDebugOffKeepsTimingReportFrozen(t *testing.T) {
	t.Run("fast lane launch is silent (debug off)", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactorySlowLaunchMS, "3600000")
		stderr := driveCodexLaneLaunch(t, false, "3600000")
		if got := factoryLaunchLineCount(stderr); got != 0 {
			t.Errorf("fast debug-off lane launch printed %d threshold lines:\n%s", got, stderr)
		}
		if got := launcherTraceLineCount(stderr); got != 0 {
			t.Errorf("fast debug-off lane launch printed %d trace lines:\n%s", got, stderr)
		}
	})
	t.Run("slow lane launch prints exactly the threshold report (debug off)", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactorySlowLaunchMS, "0")
		stderr := driveCodexLaneLaunch(t, false, "0")
		lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
		if len(lines) < 2 {
			t.Fatalf("slow debug-off lane launch printed %d lines, want the summary + step lines:\n%s", len(lines), stderr)
		}
		for i, line := range lines {
			if !strings.HasPrefix(line, "factory launch:") {
				t.Errorf("line %d is not a threshold-report line: %q\n%s", i, line, stderr)
			}
		}
		if !strings.Contains(lines[0], "pre-exec phase took") {
			t.Errorf("report does not open with the phase summary: %q", lines[0])
		}
		// The freeze: none of the debug-vocabulary steps may appear — the
		// debug-off collector records exactly its t1378 step set.
		for _, step := range []string{
			"binary resolution", "project-root resolution", "init-offer gate",
			"local-instruction load", "child-env assembly",
		} {
			if strings.Contains(stderr, launcherDebugTracePrefix+" "+step) {
				t.Errorf("debug-vocabulary step %q leaked into the debug-off report — REQ-015 freeze broken", step)
			}
			if strings.Contains(stderr, " "+step+" took ") {
				t.Errorf("debug-vocabulary step %q appears in the debug-off threshold report — REQ-015 freeze broken", step)
			}
		}
		if !strings.Contains(stderr, "lane claim took") {
			t.Errorf("threshold report lacks the t1378 lane-claim step:\n%s", stderr)
		}
	})
}
