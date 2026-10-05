// plan_audit_order_conflict_test.go: Fixture-based validation of the Group 6
// CN-4 cross-artifact ordering verb documented in the plan-auditor agent
// definition (`internal/template/templates/.claude/agents/moai/plan-auditor.md`,
// mirrored at `.claude/agents/moai/plan-auditor.md`).
//
// Like plan_audit_traceability_test.go, these tests do NOT invoke the
// plan-auditor agent. They EXTRACT the first ```bash block under the Group 6
// heading and run it against inline plan and acceptance fixtures, so a weakened
// verb turns a named cell red.
//
// Contract pinned here (card t541):
//   - Milestone order is read from plan headings `M<n>` or `Milestone M<n>` at
//     any level; an `Exit:` line inside a milestone binds the criteria it names
//     (comma shorthand included). A sub-heading keeps the binding; a heading at
//     the milestone's level or above ends it.
//   - Every acceptance record carrying an ordering word is a CANDIDATE line.
//   - A CONFLICT line appears only when the record names a bound criterion and a
//     milestone after the ordering word, and the plan schedules the pair on the
//     forbidden side — in both the "before" and the "after" direction.
//   - No milestones is GAP, an unreadable input is GAP, no ordering record is
//     NONE; the script never prints PASS.
//   - Both copies carry the same verb and the MP-9 / CN-4 / retry clauses,
//     including the rule that a CANDIDATE never forces FAIL by itself.
//
// Sentinel on failure: PLAN_AUDIT_ORDER_CONFLICT_DRIFT — the Group 6 verb or its
// clauses in the agent definition no longer match the documented behavior.
package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const orderDrift = "PLAN_AUDIT_ORDER_CONFLICT_DRIFT"

// orderVerb returns the Group 6 verb from doc with both placeholders resolved.
func orderVerb(t *testing.T, doc string) string {
	t.Helper()
	verb := auditVerb(t, doc, "### Group 6:")
	verb = strings.ReplaceAll(verb, "<plan.md>", "plan.md")
	return strings.ReplaceAll(verb, "<acceptance.md>", "acceptance.md")
}

// runOrderVerb writes plan.md and acceptance.md (a nil body leaves that file
// absent) into a fresh directory and runs verb there.
func runOrderVerb(t *testing.T, plan, acceptance *string) string {
	t.Helper()

	tmp := t.TempDir()
	for name, body := range map[string]*string{"plan.md": plan, "acceptance.md": acceptance} {
		if body == nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(tmp, name), []byte(*body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	cmd := exec.Command("bash", "-c", orderVerb(t, planAuditorDocs["template"]))
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: verb execution failed: %v\noutput: %s", orderDrift, err, out)
	}
	if hasFinding(string(out), "PASS") {
		t.Errorf("%s: verb printed a PASS line; the auditor decides; got: %s", orderDrift, out)
	}
	return string(out)
}

const orderPlan = `# Plan

## Milestones

### M1 — the render change

Change the render.

Exit: AC-ORD-002 green.

### M2 — the guard

Add the guard test.

Exit: AC-ORD-001 green, with its RED recorded.

### M3 — closure

Exit: AC-ORD-003 confirmed.
`

// orderAcceptance returns an acceptance file whose Definition of Done carries
// one ordering clause on line 10, continued on line 11.
func orderAcceptance(clause string) string {
	return `# Acceptance

## Criteria

### AC-ORD-001 — guard
**Then** the guard is green.

## Definition of Done

- ` + clause + `
  with the command and its exit code.
`
}

// TestPlanAuditOrder_ConflictIsReported: the plan schedules AC-ORD-001's
// milestone after M1 while the Definition of Done orders it before M1.
func TestPlanAuditOrder_ConflictIsReported(t *testing.T) {
	t.Parallel()

	out := runOrderVerb(t, strPtr(orderPlan), strPtr(orderAcceptance("AC-ORD-001's RED output recorded verbatim, captured BEFORE the M1 render change,")))
	t.Logf("Group 6 output:\n%s", out)

	want := "CONFLICT: acceptance.md:10 orders AC-ORD-001 before M1, but plan.md binds AC-ORD-001 to the exit of M2, which the plan places after M1"
	lines := findingLines(out, "CONFLICT:")
	if len(lines) != 1 || lines[0] != want {
		t.Errorf("%s: expected exactly %q; got %v", orderDrift, want, lines)
	}
	if lines := findingLines(out, "CANDIDATE: acceptance.md:10:"); len(lines) != 1 {
		t.Errorf("%s: the conflicting record was not surfaced as a candidate; got: %s", orderDrift, out)
	}
	if !hasFinding(out, "COLLECTED: 3 milestones in plan order (M1 M2 M3), 3 exit bindings, 1 ordering candidates") {
		t.Errorf("%s: unexpected COLLECTED line; got: %s", orderDrift, out)
	}
}

// TestPlanAuditOrder_SatisfiableAfterIsClean (control): the same clause turned
// to "after" agrees with the plan, so it stays a candidate only.
func TestPlanAuditOrder_SatisfiableAfterIsClean(t *testing.T) {
	t.Parallel()

	out := runOrderVerb(t, strPtr(orderPlan), strPtr(orderAcceptance("AC-ORD-001's RED output recorded verbatim, captured AFTER the M1 render change,")))

	if lines := findingLines(out, "CONFLICT:"); len(lines) != 0 {
		t.Errorf("%s: a satisfiable after-clause was reported as a conflict; got %v", orderDrift, lines)
	}
	if lines := findingLines(out, "CANDIDATE:"); len(lines) != 1 {
		t.Errorf("%s: the ordering record was not surfaced as a candidate; got: %s", orderDrift, out)
	}
}

// TestPlanAuditOrder_SatisfiableBeforeIsClean (control): a criterion bound to
// M1 ordered before M2 agrees with the plan.
func TestPlanAuditOrder_SatisfiableBeforeIsClean(t *testing.T) {
	t.Parallel()

	out := runOrderVerb(t, strPtr(orderPlan), strPtr(orderAcceptance("AC-ORD-002's output captured before the M2 guard lands,")))

	if lines := findingLines(out, "CONFLICT:"); len(lines) != 0 {
		t.Errorf("%s: a satisfiable before-clause was reported as a conflict; got %v", orderDrift, lines)
	}
}

// TestPlanAuditOrder_AfterConflictIsReported: a criterion bound to M1 ordered
// after M2 cannot be followed.
func TestPlanAuditOrder_AfterConflictIsReported(t *testing.T) {
	t.Parallel()

	out := runOrderVerb(t, strPtr(orderPlan), strPtr(orderAcceptance("AC-ORD-002's output recorded after the M2 guard lands,")))

	want := "CONFLICT: acceptance.md:10 orders AC-ORD-002 after M2, but plan.md binds AC-ORD-002 to the exit of M1, which the plan places before M2"
	lines := findingLines(out, "CONFLICT:")
	if len(lines) != 1 || lines[0] != want {
		t.Errorf("%s: expected exactly %q; got %v", orderDrift, want, lines)
	}
}

// TestPlanAuditOrder_MilestoneHeadingFormIsRead: `Milestone M<n>` headings
// carry the same order as `M<n>` headings.
func TestPlanAuditOrder_MilestoneHeadingFormIsRead(t *testing.T) {
	t.Parallel()

	plan := strings.ReplaceAll(orderPlan, "### M", "### Milestone M")
	out := runOrderVerb(t, strPtr(plan), strPtr(orderAcceptance("AC-ORD-001's RED output recorded verbatim, captured BEFORE the M1 render change,")))
	t.Logf("Group 6 output:\n%s", out)

	if lines := findingLines(out, "CONFLICT:"); len(lines) != 1 {
		t.Errorf("%s: Milestone-form headings did not yield the conflict; got: %s", orderDrift, out)
	}
	if !hasFinding(out, "COLLECTED: 3 milestones in plan order (M1 M2 M3)") {
		t.Errorf("%s: Milestone-form headings were not collected; got: %s", orderDrift, out)
	}
}

// TestPlanAuditOrder_SectionCriterionIsTheSubject: a Then clause under a
// criterion heading binds that criterion even when the record does not name it.
func TestPlanAuditOrder_SectionCriterionIsTheSubject(t *testing.T) {
	t.Parallel()

	out := runOrderVerb(t, strPtr(orderPlan), strPtr(`# Acceptance

### AC-ORD-001 — guard

**Then** the RED output is captured before the M1 render change.
`))

	want := "CONFLICT: acceptance.md:5 orders AC-ORD-001 before M1, but plan.md binds AC-ORD-001 to the exit of M2, which the plan places after M1"
	lines := findingLines(out, "CONFLICT:")
	if len(lines) != 1 || lines[0] != want {
		t.Errorf("%s: expected exactly %q; got %v", orderDrift, want, lines)
	}
}

// TestPlanAuditOrder_ShorthandExitBinding: `Exit: AC-ORD-001, 002` binds both.
func TestPlanAuditOrder_ShorthandExitBinding(t *testing.T) {
	t.Parallel()

	plan := `# Plan

### M1 — render

Exit: AC-ORD-004 green.

### M2 — guard

Exit: AC-ORD-001, 002 green.
`
	out := runOrderVerb(t, strPtr(plan), strPtr(orderAcceptance("AC-ORD-002's RED output captured before the M1 render change,")))

	if lines := findingLines(out, "CONFLICT: acceptance.md:10 orders AC-ORD-002 before M1"); len(lines) != 1 {
		t.Errorf("%s: the shorthand-bound criterion did not yield a conflict; got: %s", orderDrift, out)
	}
	if !hasFinding(out, "COLLECTED: 2 milestones in plan order (M1 M2), 3 exit bindings") {
		t.Errorf("%s: shorthand exit binding was not counted; got: %s", orderDrift, out)
	}
}

// TestPlanAuditOrder_SubheadingKeepsBinding: an Exit line under a deeper
// sub-heading still belongs to its milestone.
func TestPlanAuditOrder_SubheadingKeepsBinding(t *testing.T) {
	t.Parallel()

	plan := `# Plan

## M1 — render

#### Steps

Exit: AC-ORD-002 green.

## M2 — guard

#### Steps

Exit: AC-ORD-001 green.
`
	out := runOrderVerb(t, strPtr(plan), strPtr(orderAcceptance("AC-ORD-001's RED output captured BEFORE the M1 render change,")))

	if lines := findingLines(out, "CONFLICT:"); len(lines) != 1 {
		t.Errorf("%s: a sub-heading ended the milestone binding; got: %s", orderDrift, out)
	}
}

// TestPlanAuditOrder_ExitOutsideMilestoneDoesNotBind: an Exit line under a
// sibling heading after the milestones binds nothing.
func TestPlanAuditOrder_ExitOutsideMilestoneDoesNotBind(t *testing.T) {
	t.Parallel()

	plan := orderPlan + `
## Notes

Exit: AC-ORD-009 green.
`
	out := runOrderVerb(t, strPtr(plan), strPtr(orderAcceptance("AC-ORD-009's output captured before the M1 render change,")))

	if lines := findingLines(out, "CONFLICT:"); len(lines) != 0 {
		t.Errorf("%s: an Exit line outside any milestone produced a binding; got %v", orderDrift, lines)
	}
	if !hasFinding(out, "COLLECTED: 3 milestones in plan order (M1 M2 M3), 3 exit bindings") {
		t.Errorf("%s: an Exit line outside any milestone was counted; got: %s", orderDrift, out)
	}
}

// TestPlanAuditOrder_UnboundPlanYieldsCandidateOnly: milestones without Exit
// lines give the auditor candidates, never a mechanical conflict.
func TestPlanAuditOrder_UnboundPlanYieldsCandidateOnly(t *testing.T) {
	t.Parallel()

	plan := "# Plan\n\n### M1 — render\n\nChange it.\n\n### M2 — guard\n\nAdd it.\n"
	out := runOrderVerb(t, strPtr(plan), strPtr(orderAcceptance("AC-ORD-001's RED output captured BEFORE the M1 render change,")))

	if lines := findingLines(out, "CONFLICT:"); len(lines) != 0 {
		t.Errorf("%s: an unbound plan produced a mechanical conflict; got %v", orderDrift, lines)
	}
	if lines := findingLines(out, "CANDIDATE:"); len(lines) != 1 {
		t.Errorf("%s: the ordering record was not surfaced; got: %s", orderDrift, out)
	}
	if !hasFinding(out, "COLLECTED: 2 milestones in plan order (M1 M2), 0 exit bindings, 1 ordering candidates") {
		t.Errorf("%s: unexpected COLLECTED line; got: %s", orderDrift, out)
	}
}

// TestPlanAuditOrder_NoOrderingClauseIsNone: an acceptance file with no
// ordering word is an observed absence, reported on a NONE line.
func TestPlanAuditOrder_NoOrderingClauseIsNone(t *testing.T) {
	t.Parallel()

	out := runOrderVerb(t, strPtr(orderPlan), strPtr("# Acceptance\n\n- AC-ORD-001: Given one, When two, Then three.\n"))

	if !hasFinding(out, "NONE: 0 records in acceptance.md") {
		t.Errorf("%s: no ordering record did not produce a NONE line; got: %s", orderDrift, out)
	}
	if lines := append(findingLines(out, "CANDIDATE:"), findingLines(out, "CONFLICT:")...); len(lines) != 0 {
		t.Errorf("%s: an ordering-free file produced findings; got %v", orderDrift, lines)
	}
}

// TestPlanAuditOrder_NoMilestonesIsGap: a plan without milestone headings is a
// gap, and the candidate still reaches the auditor.
func TestPlanAuditOrder_NoMilestonesIsGap(t *testing.T) {
	t.Parallel()

	out := runOrderVerb(t, strPtr("# Plan\n\nDo the work.\n\nExit: AC-ORD-001 green.\n"), strPtr(orderAcceptance("AC-ORD-001's RED output captured BEFORE the M1 render change,")))

	if !hasFinding(out, "GAP: 0 milestone headings") {
		t.Errorf("%s: a plan without milestones was not reported as GAP; got: %s", orderDrift, out)
	}
	if lines := findingLines(out, "CONFLICT:"); len(lines) != 0 {
		t.Errorf("%s: a plan without milestones produced a conflict; got %v", orderDrift, lines)
	}
	if lines := findingLines(out, "CANDIDATE:"); len(lines) != 1 {
		t.Errorf("%s: the candidate was dropped when milestones were absent; got: %s", orderDrift, out)
	}
}

// TestPlanAuditOrder_UnreadableInputIsGap: a missing plan is a gap, not a
// measurement.
func TestPlanAuditOrder_UnreadableInputIsGap(t *testing.T) {
	t.Parallel()

	out := runOrderVerb(t, nil, strPtr(orderAcceptance("AC-ORD-001 captured BEFORE the M1 change,")))

	if !hasFinding(out, "GAP: plan.md is not readable") {
		t.Errorf("%s: an unreadable plan was not reported as GAP; got: %s", orderDrift, out)
	}
	if hasFinding(out, "COLLECTED:") || hasFinding(out, "CANDIDATE:") {
		t.Errorf("%s: an unreadable plan produced a measurement; got: %s", orderDrift, out)
	}
}

// TestPlanAuditOrder_VerbIdenticalAcrossCopies: the local mirror carries the
// same Group 6 verb as the distributed template copy.
func TestPlanAuditOrder_VerbIdenticalAcrossCopies(t *testing.T) {
	t.Parallel()

	if orderVerb(t, planAuditorDocs["template"]) != orderVerb(t, planAuditorDocs["local"]) {
		t.Errorf("%s: the Group 6 verb differs between the template and local copies", orderDrift)
	}
}

// orderSection returns the part of body from the line starting with heading up
// to the next heading of the same or a higher level.
func orderSection(t *testing.T, name, body, heading string) string {
	t.Helper()
	start := strings.Index(body, "\n"+heading)
	if start < 0 {
		t.Fatalf("%s: no %q heading in the %s copy", orderDrift, heading, name)
	}
	section := body[start+1:]
	level := strings.Index(section, " ")
	if end := strings.Index(section[1:], "\n"+strings.Repeat("#", level)+" "); end >= 0 {
		section = section[:end+1]
	}
	return section
}

// TestPlanAuditOrder_ClausesPresent: both copies carry the CN-4 item, the MP-9
// firewall clause with its CANDIDATE/CONFLICT split, the report row, and the
// retry-loop exemption, each inside the section it governs.
func TestPlanAuditOrder_ClausesPresent(t *testing.T) {
	t.Parallel()

	cells := []struct{ heading, want string }{
		{"### M5: Must-Pass Firewall", "Nine criteria cannot be compensated"},
		{"### M5: Must-Pass Firewall", "**(MP-9) No unresolved cross-artifact ordering conflict**"},
		{"### M5: Must-Pass Firewall", "**forces FAIL on its own**"},
		{"### M5: Must-Pass Firewall", "**never forces FAIL by itself**"},
		{"### Group 6: Consistency", "- CN-4: "},
		{"## Must-Pass Results", "- [PASS/FAIL/N/A] MP-9 cross-artifact ordering consistency:"},
		{"## Retry Loop Contract", "the Group 6 CN-4 ordering verb runs in full on every iteration"},
	}
	for name, doc := range planAuditorDocs {
		raw, err := os.ReadFile(doc)
		if err != nil {
			t.Fatalf("%s: read %s: %v", orderDrift, doc, err)
		}
		for _, c := range cells {
			if !strings.Contains(orderSection(t, name, string(raw), c.heading), c.want) {
				t.Errorf("%s: the %s copy's %q section lacks %q", orderDrift, name, c.heading, c.want)
			}
		}
	}
}
