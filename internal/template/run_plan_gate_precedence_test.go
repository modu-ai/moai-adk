package template

import (
	"io/fs"
	"strings"
	"testing"
)

// Card t2008 / RUN-01. The run workflow body teaches a plan-audit grace
// fail-through (FAIL_WARNED), an override on FAIL, an acknowledgement on
// INCONCLUSIVE, and an env/flag bypass, while the run router and
// auto-semantics §9 make FAIL, INCONCLUSIVE, BYPASSED and an absent verdict
// hard blocks and admit PASS and PASS-WITH-DEBT. The body must say which
// reading wins, and it must keep the PASS-WITH-DEBT path intact.

const phaseExecutionPath = ".claude/skills/moai/workflows/run/phase-execution.md"

// section returns the text from the first occurrence of start up to the next
// occurrence of end after it (or the end of the text). ok is false when start
// is absent.
func section(text, start, end string) (string, bool) {
	_, rest, found := strings.Cut(text, start)
	if !found {
		return "", false
	}
	if before, _, ok := strings.Cut(rest, end); ok {
		rest = before
	}
	return start + rest, true
}

// planGatePrecedenceViolations lists the precedence statements missing from a
// phase-execution.md body. An empty result means every statement is present.
func planGatePrecedenceViolations(text string) []string {
	var v []string
	need := func(label, body string, ok bool, all ...string) {
		if !ok {
			v = append(v, label+": section not found")
			return
		}
		for _, w := range all {
			if !strings.Contains(body, w) {
				v = append(v, label+": missing "+w)
			}
		}
	}

	routing, ok := section(text, "### Step 4: Verdict Routing", "### Step 5:")
	need("Step 4 precedence", routing, ok,
		"Precedence", "auto-semantics.md", "§9.1", "BYPASSED", "hard block")

	pass, ok := section(text, "**4a. PASS**", "**4b.")
	need("Step 4a PASS-WITH-DEBT", pass, ok,
		"PASS-WITH-DEBT", "Binding run conditions")

	skip, ok := section(text, "### When --skip-audit Flag Is Provided", "### When Plan-Auditor Fails")
	need("--skip-audit legacy marker", skip, ok,
		"Legacy", "retired", "§9.1")

	gate, ok := section(text, "## HUMAN GATE: Plan Approval", "## Phase 6:")
	need("Plan Approval operator form", gate, ok,
		"keep-set", "§9.1")

	return v
}

// TestRunPlanGatePrecedence_Deployed checks the deployed template body.
func TestRunPlanGatePrecedence_Deployed(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	data, err := fs.ReadFile(fsys, phaseExecutionPath)
	if err != nil {
		t.Fatalf("read %s: %v", phaseExecutionPath, err)
	}
	if v := planGatePrecedenceViolations(string(data)); len(v) > 0 {
		t.Errorf("%s does not state plan-gate precedence:\n  %s", phaseExecutionPath, strings.Join(v, "\n  "))
	}
}

// TestRunPlanGatePrecedence_Mutants proves the checker can fail: a body with
// the grace/override/bypass teaching and no precedence statement is rejected,
// and each required statement is individually load-bearing.
func TestRunPlanGatePrecedence_Mutants(t *testing.T) {
	t.Parallel()

	const bare = "### Step 4: Verdict Routing\n**4a. PASS**\n- proceed\n**4b. FAIL**\n### Step 5: Persist\n" +
		"### When --skip-audit Flag Is Provided\nSkip Steps 1-4.\n### When Plan-Auditor Fails or Times Out\n" +
		"## HUMAN GATE: Plan Approval\nAsk.\n## Phase 6: Task Decomposition\n"
	if v := planGatePrecedenceViolations(bare); len(v) < 4 {
		t.Fatalf("bare body must fail every check group, got %d violations: %v", len(v), v)
	}

	const full = "### Step 4: Verdict Routing\nPrecedence: auto-semantics.md §9.1 - FAIL, INCONCLUSIVE and BYPASSED are a hard block.\n" +
		"**4a. PASS**\n- PASS-WITH-DEBT copies debt into Binding run conditions\n**4b. FAIL**\n### Step 5: Persist\n" +
		"### When --skip-audit Flag Is Provided\nLegacy: retired by §9.1.\n### When Plan-Auditor Fails or Times Out\n" +
		"## HUMAN GATE: Plan Approval\nkeep-set cases only, §9.1.\n## Phase 6: Task Decomposition\n"
	if v := planGatePrecedenceViolations(full); len(v) != 0 {
		t.Fatalf("positive control must pass, got: %v", v)
	}

	for _, drop := range []string{"PASS-WITH-DEBT", "Binding run conditions", "BYPASSED", "keep-set", "Legacy"} {
		mutant := strings.ReplaceAll(full, drop, "x")
		if v := planGatePrecedenceViolations(mutant); len(v) == 0 {
			t.Errorf("dropping %q must be detected", drop)
		}
	}
}
