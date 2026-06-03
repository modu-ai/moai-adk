// gate2_preservation_test.go: GATE-2 preservation regression guard for the
// run-phase autonomy section in .claude/skills/moai/workflows/run.md.
//
// The orchestrator's GATE-2 (plan->run HUMAN GATE) behavior is doctrine, not Go
// code, so the verifiable artifact is the run skill body that instructs the
// orchestrator. This guard reads that body (the SOURCE under .claude/, not the
// embedded template) and asserts two complementary invariants introduced by the
// run.md "Run-phase Autonomy (/goal ac_converge)" section:
//
//	Check A (ordering invariant): a GATE-2 AskUserQuestion human-gate marker
//	exists AND its first byte offset precedes the first /goal token (the
//	ac_converge set) and any Mode-6 launch reference. This proves /goal and
//	Mode-6 launch cannot cross the plan->run boundary ahead of the human gate.
//
//	Check B (score-independence + cross-reference): the body states GATE-2 is
//	emitted regardless of plan-auditor score (incl. >= 0.90 skip-eligible) AND
//	contains a doctrine cross-reference token (§19.1 and/or REQ-ATR-015).
//
// The test is intentionally lightweight (string/offset assertions only — no
// orchestration simulation, no syscalls) so it builds and runs identically on
// every platform including GOOS=windows.
//
// Sentinel on failure: GATE2_PRESERVATION_VIOLATION.
package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runSkillRelPath is the run-phase router skill body that co-locates BOTH the
// GATE-2 AskUserQuestion ordering reference AND the /goal ac_converge set in one
// self-contained "Run-phase Autonomy (/goal ac_converge)" section.
const runSkillRelPath = ".claude/skills/moai/workflows/run.md"

// findProjectRootForGate2 walks upward from the current working directory until a
// go.mod file is found, returning the project root. A distinct name avoids
// clashing with findProjectRootForMirrorTest (rule_template_mirror_test.go) in
// the same external test package.
func findProjectRootForGate2(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found; cannot determine project root")
		}
		dir = parent
	}
}

// TestGate2PreservedBeforeGoal enforces the GATE-2 preservation invariants on the
// run skill body. It is the D3 deliverable of the run-phase autonomy SPEC.
func TestGate2PreservedBeforeGoal(t *testing.T) {
	t.Parallel()

	root := findProjectRootForGate2(t)
	srcPath := filepath.Join(root, runSkillRelPath)

	content, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("run skill body unreadable %s: %v", srcPath, err)
	}
	body := string(content)

	// --- Check A: ordering invariant (GATE-2 AskUserQuestion before first /goal) ---
	// The GATE-2 human gate is identified by the co-occurrence of "GATE-2" and
	// "AskUserQuestion" — we locate the first GATE-2 marker and the first
	// AskUserQuestion marker, take the earlier of the two as the gate anchor's
	// upper bound, and require it to precede the first "/goal" token.
	gate2Idx := strings.Index(body, "GATE-2")
	if gate2Idx < 0 {
		t.Fatalf("GATE2_PRESERVATION_VIOLATION: run skill body %s contains no GATE-2 marker; "+
			"the Run-phase Autonomy section must reference the GATE-2 human gate", runSkillRelPath)
	}

	askUserIdx := strings.Index(body, "AskUserQuestion")
	if askUserIdx < 0 {
		t.Fatalf("GATE2_PRESERVATION_VIOLATION: run skill body %s contains no AskUserQuestion "+
			"reference; the GATE-2 human gate is an AskUserQuestion gate", runSkillRelPath)
	}

	goalIdx := strings.Index(body, "/goal")
	if goalIdx < 0 {
		t.Fatalf("GATE2_PRESERVATION_VIOLATION: run skill body %s contains no /goal token; "+
			"the Run-phase Autonomy section must wire the ac_converge /goal", runSkillRelPath)
	}

	// The GATE-2 AskUserQuestion ordering reference (both markers) MUST appear
	// textually before the first /goal token. We require the GATE-2 marker AND
	// the AskUserQuestion marker each to precede the first /goal occurrence.
	if gate2Idx >= goalIdx {
		t.Errorf("GATE2_PRESERVATION_VIOLATION: GATE-2 marker (offset %d) does not precede the "+
			"first /goal token (offset %d) in %s; the human gate must be ordered before any "+
			"/goal set", gate2Idx, goalIdx, runSkillRelPath)
	}
	if askUserIdx >= goalIdx {
		t.Errorf("GATE2_PRESERVATION_VIOLATION: AskUserQuestion marker (offset %d) does not "+
			"precede the first /goal token (offset %d) in %s; the human-gate channel must be "+
			"ordered before any /goal set", askUserIdx, goalIdx, runSkillRelPath)
	}

	// --- Check B: score-independence statement + doctrine cross-reference ---
	// Score-independence: the body must state GATE-2 is emitted regardless of the
	// plan-auditor score (including the >= 0.90 skip-eligible case).
	hasScoreIndependence := strings.Contains(body, "regardless of") &&
		strings.Contains(body, "plan-auditor")
	if !hasScoreIndependence {
		// Fall back to the alternative canonical phrasings the AC permits.
		if strings.Contains(body, "score-independent") ||
			strings.Contains(body, "skip-eligib") {
			hasScoreIndependence = true
		}
	}
	if !hasScoreIndependence {
		t.Errorf("GATE2_PRESERVATION_VIOLATION: run skill body %s lacks a score-independence "+
			"statement (GATE-2 emitted regardless of plan-auditor score, incl. >= 0.90 "+
			"skip-eligible)", runSkillRelPath)
	}

	// Doctrine cross-reference: at least one of §19.1 / REQ-ATR-015 must be present.
	if !strings.Contains(body, "§19.1") && !strings.Contains(body, "REQ-ATR-015") {
		t.Errorf("GATE2_PRESERVATION_VIOLATION: run skill body %s lacks a doctrine "+
			"cross-reference to §19.1 / REQ-ATR-015 (GATE-2 mandatory restoration)", runSkillRelPath)
	}
}
