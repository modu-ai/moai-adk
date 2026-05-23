// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M2 handoff guidance tests.
//
// Covers `printHandoff` (P4 default path, no --team) and
// `printHandoffWithError` (P1/P2 pane-spawn fallback notice path, used by M3).
//
// REQ-WTL-004 (P4 dispatch: --team absent → stdout handoff guidance).

package worktree

import (
	"bytes"
	"strings"
	"testing"
)

// TestPrintHandoff_CC_Mode verifies the paste-ready handoff line for the
// default CC launcher. AC-WTL-004 requires both "cd " and " && moai" literals
// to appear in stdout to support copy-paste from the terminal.
func TestPrintHandoff_CC_Mode(t *testing.T) {
	cfg := TeamLaunchConfig{
		Pattern:      PatternP4Handoff,
		WorktreePath: "/tmp/test-cc",
		SpecID:       "SPEC-WTL-DEMO-CC",
		LLM:          "cc",
	}
	var buf bytes.Buffer
	printHandoff(&buf, cfg)
	out := buf.String()

	if !strings.Contains(out, "cd /tmp/test-cc") {
		t.Errorf("handoff missing `cd /tmp/test-cc` line; got: %q", out)
	}
	if !strings.Contains(out, "moai cc") {
		t.Errorf("handoff missing `moai cc` invocation; got: %q", out)
	}
	if strings.Contains(out, "moai glm") {
		t.Errorf("CC-mode handoff must NOT include `moai glm`; got: %q", out)
	}
	// AC-WTL-004 paste-ready literals
	if !strings.Contains(out, "&& moai") {
		t.Errorf("handoff must contain `&& moai` (paste-ready chain); got: %q", out)
	}
}

// TestPrintHandoff_GLM_Mode verifies CG-mode handoff routes to `moai glm`.
func TestPrintHandoff_GLM_Mode(t *testing.T) {
	cfg := TeamLaunchConfig{
		Pattern:      PatternP4Handoff,
		WorktreePath: "/tmp/test-glm",
		SpecID:       "SPEC-WTL-DEMO-GLM",
		LLM:          "glm",
	}
	var buf bytes.Buffer
	printHandoff(&buf, cfg)
	out := buf.String()

	if !strings.Contains(out, "cd /tmp/test-glm") {
		t.Errorf("handoff missing `cd /tmp/test-glm` line; got: %q", out)
	}
	if !strings.Contains(out, "moai glm") {
		t.Errorf("GLM-mode handoff missing `moai glm`; got: %q", out)
	}
	if strings.Contains(out, "moai cc") {
		t.Errorf("GLM-mode handoff must NOT include `moai cc`; got: %q", out)
	}
}

// TestPrintHandoffWithError verifies the failure-fallback variant emits a
// stderr-side notice plus the same handoff body on stdout. Used by M3 when
// tmux new-window fails (AC-WTL-007 P4 fallback).
func TestPrintHandoffWithError(t *testing.T) {
	cfg := TeamLaunchConfig{
		Pattern:      PatternP2TmuxCC,
		WorktreePath: "/tmp/test-fallback",
		SpecID:       "SPEC-WTL-DEMO-FALLBACK",
		LLM:          "cc",
	}
	var stdout, stderr bytes.Buffer
	printHandoffWithError(&stdout, &stderr, cfg, "tmux pane spawn failed: exit status 1")

	// stderr must contain a warning prefix + the supplied reason
	if !strings.Contains(stderr.String(), "warning:") {
		t.Errorf("stderr must contain `warning:` prefix; got: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "tmux pane spawn failed") {
		t.Errorf("stderr must echo the supplied reason; got: %q", stderr.String())
	}
	// stdout must still contain the paste-ready handoff
	if !strings.Contains(stdout.String(), "cd /tmp/test-fallback") {
		t.Errorf("stdout missing handoff path line; got: %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "moai cc") {
		t.Errorf("stdout missing `moai cc`; got: %q", stdout.String())
	}
}

// TestPrintHandoff_PathWithSpaces verifies worktree paths containing spaces
// (edge case per acceptance.md §2) are printed verbatim in the cd line.
func TestPrintHandoff_PathWithSpaces(t *testing.T) {
	cfg := TeamLaunchConfig{
		Pattern:      PatternP4Handoff,
		WorktreePath: "/tmp/path with spaces",
		LLM:          "cc",
	}
	var buf bytes.Buffer
	printHandoff(&buf, cfg)
	if !strings.Contains(buf.String(), "/tmp/path with spaces") {
		t.Errorf("handoff must preserve worktree path verbatim; got: %q", buf.String())
	}
}
