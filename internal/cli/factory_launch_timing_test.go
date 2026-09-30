package cli

// factory_launch_timing_test.go — SPEC-CODEX-LANE-SLOTS-001 AC-010 (REQ-012):
// the codex lane launch's slow-launch report carries one line per pre-exec
// step, each naming the step (join gate, active-run resolution, codex
// pre-exec init, exec handoff at minimum) and its wall-time.

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestCodexLaneLaunchTimingNamesSteps(t *testing.T) {
	// Threshold 0ms makes the phase slow by construction, so the report
	// emits without depending on machine speed.
	t.Setenv(config.EnvMoaiFactorySlowLaunchMS, "0")

	timing := &factoryLaunchTiming{}
	endInit := timing.begin(factoryStepCodexPreInit)
	endInit()
	endJoin := timing.begin(factoryStepJoinGate)
	endResolve := timing.begin(factoryStepRunResolve)
	endResolve()
	endJoin()
	endClaim := timing.begin(factoryStepLaneClaim)
	endClaim()
	endHandoff := timing.begin(factoryStepExecHandoff)
	endHandoff()

	var out bytes.Buffer
	timing.reportSlow(&out)
	report := out.String()
	if report == "" {
		t.Fatal("slow-launch report is empty; want one line per pre-exec step")
	}
	for _, step := range []string{
		factoryStepCodexPreInit,
		factoryStepJoinGate,
		factoryStepRunResolve,
		factoryStepLaneClaim,
		factoryStepExecHandoff,
	} {
		if !strings.Contains(report, step) {
			t.Errorf("report does not name the %q step:\n%s", step, report)
		}
	}
	// One line per recorded step (plus the phase summary line), each carrying
	// a wall-time.
	lines := strings.Split(strings.TrimRight(report, "\n"), "\n")
	if len(lines) != 6 {
		t.Errorf("report carries %d lines, want 6 (phase summary + one per recorded step):\n%s", reportLines(report), report)
	}
	if !strings.Contains(lines[0], "pre-exec phase took") {
		t.Errorf("report does not open with the phase summary: %q", lines[0])
	}
	for _, line := range lines[1:] {
		if !strings.Contains(line, "took") {
			t.Errorf("report line lacks the step timing shape: %q", line)
		}
	}
}

// TestCodexLaneLaunchTimingQuietUnderThreshold pins the other half of the
// REQ-012 trigger: a phase under the threshold prints nothing.
func TestCodexLaneLaunchTimingQuietUnderThreshold(t *testing.T) {
	// A threshold no real phase exceeds keeps the report quiet.
	t.Setenv(config.EnvMoaiFactorySlowLaunchMS, "3600000")

	timing := &factoryLaunchTiming{}
	end := timing.begin(factoryStepCodexPreInit)
	end()

	var out bytes.Buffer
	timing.reportSlow(&out)
	if out.Len() != 0 {
		t.Errorf("report printed under a huge threshold: %q", out.String())
	}
}

// TestFactoryLaunchTimingNilSafe pins the nil-receiver contract the shared
// join gate relies on: a nil timing measures nothing and prints nothing, so
// the cc/glm twins pass nil unchanged through the instrumented path.
func TestFactoryLaunchTimingNilSafe(t *testing.T) {
	var timing *factoryLaunchTiming
	end := timing.begin(factoryStepJoinGate) // must not panic
	end()
	var out bytes.Buffer
	timing.reportSlow(&out) // must not panic
	if out.Len() != 0 {
		t.Errorf("nil timing printed a report: %q", out.String())
	}
	if got := timing.phaseElapsed(); got != 0 {
		t.Errorf("nil timing phase elapsed = %v, want 0", got)
	}
}

func reportLines(report string) int {
	n := 0
	for _, line := range strings.Split(report, "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}

// The step durations must be real wall-times: a begin/end pair around a
// measurable sleep records a positive elapsed.
func TestFactoryLaunchTimingMeasuresElapsed(t *testing.T) {
	timing := &factoryLaunchTiming{}
	end := timing.begin(factoryStepCodexPreInit)
	time.Sleep(2 * time.Millisecond)
	end()
	if len(timing.steps) != 1 {
		t.Fatalf("recorded %d steps, want 1", len(timing.steps))
	}
	if timing.steps[0].elapsed <= 0 {
		t.Errorf("step elapsed = %v, want positive", timing.steps[0].elapsed)
	}
	if got := timing.phaseElapsed(); got <= 0 {
		t.Errorf("phase elapsed = %v, want positive", got)
	}
}
