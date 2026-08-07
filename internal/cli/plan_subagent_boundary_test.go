package cli

import (
	"os"
	"strings"
	"testing"
)

// assertNoAskUserQuestionInSource reads the named Go source file and fails if
// any NON-comment line references AskUserQuestion or mcp__askuser. Comments are
// out of scope per AC-WIRE-010 ("the assertion is on production source") — the
// filter mirrors internal/goal/subagent_boundary_test.go so docstring mentions
// of the boundary ("this CLI MUST NOT invoke AskUserQuestion") remain valid.
func assertNoAskUserQuestionInSource(t *testing.T, filename string) {
	t.Helper()
	src, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	for i, line := range strings.Split(string(src), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		for _, forbidden := range []string{"AskUserQuestion", "mcp__askuser"} {
			if strings.Contains(line, forbidden) {
				t.Errorf("%s:%d references %q — subagent-boundary violation",
					filename, i+1, forbidden)
			}
		}
	}
}

// TestPlan_NoAskUserQuestion is the C-HRA-008 / REQ-WIRE-011 static guard for
// the NEW `moai plan` CLI source: the subagent-boundary discipline (no
// AskUserQuestion / mcp__askuser in CLI code) MUST hold. Mirrors
// internal/cli/web_test.go::TestWeb_NoAskUserQuestion in shape with the
// canonical comment-filter from internal/goal/subagent_boundary_test.go.
func TestPlan_NoAskUserQuestion(t *testing.T) {
	assertNoAskUserQuestionInSource(t, "plan.go")
}

// TestGoal_NoAskUserQuestion re-asserts the boundary on goal.go after the M1/M2
// modifications (AC-WIRE-010 re-assertion for the modified CLI source).
func TestGoal_NoAskUserQuestion(t *testing.T) {
	assertNoAskUserQuestionInSource(t, "goal.go")
}

// TestSpecAssembly_RewrittenToCLIPath verifies AC-WIRE-006: the rewritten
// spec-assembly.md template source (Step 2.3.3a) no longer carries the dead
// `RenderPlanHTML(specDir=` LLM-instruction and DOES carry the executable
// `moai plan render-html` CLI path. The [HARD] Implementation Kickoff Approval
// paragraph and the fail-open clause are preserved.
func TestSpecAssembly_RewrittenToCLIPath(t *testing.T) {
	path := "../template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md"
	src, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("spec-assembly.md template not reachable from this test dir: %v", err)
	}
	body := string(src)

	// Dead instruction absent.
	if strings.Contains(body, "RenderPlanHTML(specDir=") {
		t.Errorf("spec-assembly.md still references the dead RenderPlanHTML(specDir= instruction")
	}
	// New executable path present.
	if !strings.Contains(body, "moai plan render-html") {
		t.Errorf("spec-assembly.md missing the executable `moai plan render-html` CLI path")
	}
	// [HARD] Implementation Kickoff Approval paragraph preserved.
	if !strings.Contains(body, "[HARD] The Implementation Kickoff Approval") {
		t.Errorf("spec-assembly.md lost the [HARD] Implementation Kickoff Approval paragraph")
	}
	// Fail-open clause preserved.
	if !strings.Contains(body, "Fail-open") {
		t.Errorf("spec-assembly.md lost the fail-open clause")
	}
}

// TestSpecAssembly_NoNewInternalTokens verifies AC-WIRE-011 §25 neutrality: the
// rewrite adds NO new internal SPEC IDs (SPEC-GOAL-HTML-WIRING-001 / -FLOW-001),
// NO new REQ tokens (REQ-WIRE-* / REQ-GHF-*), NO new AC tokens (AC-WIRE-* /
// AC-GHF-*). `RenderPlanHTML` already appeared pre-rewrite as the permitted
// renderer cross-reference; it is preserved and is the sole such reference.
func TestSpecAssembly_NoNewInternalTokens(t *testing.T) {
	path := "../template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md"
	src, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("spec-assembly.md template not reachable from this test dir: %v", err)
	}
	body := string(src)

	for _, forbidden := range []string{
		"SPEC-GOAL-HTML-WIRING-001",
		"SPEC-GOAL-HTML-FLOW-001",
		"REQ-WIRE-",
		"AC-WIRE-",
		"REQ-GHF-",
		"AC-GHF-",
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("spec-assembly.md carries forbidden internal token %q (§25 neutrality violation)", forbidden)
		}
	}
}
