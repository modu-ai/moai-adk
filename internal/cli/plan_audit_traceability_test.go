// plan_audit_traceability_test.go: Fixture-based validation of the Group 4
// traceability verb documented in the plan-auditor agent definition
// (`internal/template/templates/.claude/agents/moai/plan-auditor.md`, mirrored
// at `.claude/agents/moai/plan-auditor.md`).
//
// Like plan_audit_d7_d8_test.go, these tests do NOT invoke the plan-auditor
// agent. They EXTRACT the first ```bash block under the Group 4 heading and run
// it against inline SPEC fixtures, so a weakened verb turns a named cell red.
//
// Contract pinned here (card t524):
//   - REQ definitions are collected in three forms: list item, heading, and
//     table first column; the count is printed on a line-head COLLECTED line.
//   - A bare numeric tail expands to a full REQ ID only inside the same table
//     cell or comma-separated list as a preceding complete REQ ID.
//   - Findings are line-head markers (UNCOVERED / ORPHAN / GAP); the script
//     never prints PASS. Zero definitions and an unreadable spec are GAP.
//   - Both copies carry the same verb and the citation-discipline clause.
//
// Findings are judged with a line-head matcher, never a substring match: the
// markers appear inside the prose of the agent definition and could appear
// inside a GAP message.
//
// Sentinel on failure: PLAN_AUDIT_TRACEABILITY_DRIFT — the Group 4 verb or
// clause in the agent definition no longer matches the documented behavior.
package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const traceDrift = "PLAN_AUDIT_TRACEABILITY_DRIFT"

// citationDisciplineKey is one stable sentence of the citation-discipline
// clause that both copies of the agent definition must carry in Group 4.
const citationDisciplineKey = "may be cited as corroboration only after showing that tool collected N > 0"

// traceVerb returns the Group 4 verb from doc with both placeholders resolved.
func traceVerb(t *testing.T, doc string) string {
	t.Helper()
	return strings.ReplaceAll(auditVerb(t, doc, "### Group 4:"), "<acceptance.md>", "acceptance.md")
}

// runTraceVerb writes new-spec.md and acceptance.md (a nil body leaves that
// file absent) into a fresh directory and runs verb there.
func runTraceVerb(t *testing.T, verb string, spec, acceptance *string) string {
	t.Helper()

	tmp := t.TempDir()
	for name, body := range map[string]*string{"new-spec.md": spec, "acceptance.md": acceptance} {
		if body == nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(tmp, name), []byte(*body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	cmd := exec.Command("bash", "-c", verb)
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: verb execution failed: %v\noutput: %s", traceDrift, err, out)
	}
	if hasFinding(string(out), "PASS") {
		t.Errorf("%s: verb printed a PASS line; the auditor decides; got: %s", traceDrift, out)
	}
	return string(out)
}

// findingLines returns every output line that starts with marker.
func findingLines(out, marker string) []string {
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, marker) {
			lines = append(lines, line)
		}
	}
	return lines
}

const fixASpec = `---
id: SPEC-FIXA-001
status: draft
---

# SPEC-FIXA-001

## Requirements

- REQ-FIXA-001: The system shall do one thing.
- REQ-FIXA-002: The system shall do a second thing.
- REQ-FIXA-003: The system shall do a third thing.
`

// TestPlanAuditTraceability_ShorthandMappingIsCovered (fixture A): a shorthand
// comma list maps the second REQ, so nothing is reported uncovered.
func TestPlanAuditTraceability_ShorthandMappingIsCovered(t *testing.T) {
	t.Parallel()

	out := runTraceVerb(t, traceVerb(t, planAuditorDocs["template"]), strPtr(fixASpec), strPtr(`# Acceptance — SPEC-FIXA-001

- AC-FIXA-001 (maps REQ-FIXA-001, 002): Given one, When two, Then three.
- AC-FIXA-002 (maps REQ-FIXA-003): Given one, When two, Then three.
`))
	t.Logf("Group 4 output:\n%s", out)

	if lines := findingLines(out, "UNCOVERED:"); len(lines) != 0 {
		t.Errorf("%s: shorthand REQ-FIXA-001, 002 was not expanded; got UNCOVERED %v", traceDrift, lines)
	}
	if lines := findingLines(out, "ORPHAN:"); len(lines) != 0 {
		t.Errorf("%s: shorthand expansion produced an undefined REQ; got ORPHAN %v", traceDrift, lines)
	}
	if !hasFinding(out, "COLLECTED: 3 REQ definitions") {
		t.Errorf("%s: expected a COLLECTED: 3 REQ definitions line; got: %s", traceDrift, out)
	}
}

// TestPlanAuditTraceability_FullIDsHaveNoFalsePositive (fixture B, control):
// full IDs in a maps list produce neither UNCOVERED nor ORPHAN.
func TestPlanAuditTraceability_FullIDsHaveNoFalsePositive(t *testing.T) {
	t.Parallel()

	out := runTraceVerb(t, traceVerb(t, planAuditorDocs["template"]), strPtr(`---
id: SPEC-FIXB-001
status: draft
---

# SPEC-FIXB-001

## Requirements

- REQ-FIXB-001: The system shall do one thing.
- REQ-FIXB-002: The system shall do a second thing.
- REQ-FIXB-003: The system shall do a third thing.
`), strPtr(`# Acceptance — SPEC-FIXB-001

- AC-FIXB-001 (maps REQ-FIXB-001, REQ-FIXB-002): Given one, When two, Then three.
- AC-FIXB-002 (maps REQ-FIXB-003): Given one, When two, Then three.
`))

	if lines := append(findingLines(out, "UNCOVERED:"), findingLines(out, "ORPHAN:")...); len(lines) != 0 {
		t.Errorf("%s: full-ID mapping produced findings; got %v", traceDrift, lines)
	}
}

// TestPlanAuditTraceability_TableMappingIsCovered (fixture C): a table row
// without a maps keyword still maps every REQ it names.
func TestPlanAuditTraceability_TableMappingIsCovered(t *testing.T) {
	t.Parallel()

	out := runTraceVerb(t, traceVerb(t, planAuditorDocs["template"]), strPtr(`---
id: SPEC-FIXC-001
status: draft
---

# SPEC-FIXC-001

## Requirements

- REQ-FIXC-001: The system shall do one thing.
- REQ-FIXC-002: The system shall do a second thing.
- REQ-FIXC-003: The system shall do a third thing.
`), strPtr(`# Acceptance — SPEC-FIXC-001

| AC | REQ | Then |
|----|-----|------|
| AC-FIXC-001 | REQ-FIXC-001, REQ-FIXC-002 | three |
| AC-FIXC-002 | REQ-FIXC-003 | three |
`))

	if lines := findingLines(out, "UNCOVERED:"); len(lines) != 0 {
		t.Errorf("%s: table mapping was not read; got UNCOVERED %v", traceDrift, lines)
	}
}

// TestPlanAuditTraceability_HeadingDefinitionsExposeRealGap (fixture D): with
// heading-form definitions and a shorthand mapping, exactly the real gap is
// reported. This is the executable proxy for the circular-citation failure: a
// collector blind to heading definitions reports nothing here.
func TestPlanAuditTraceability_HeadingDefinitionsExposeRealGap(t *testing.T) {
	t.Parallel()

	out := runTraceVerb(t, traceVerb(t, planAuditorDocs["template"]), strPtr(`---
id: SPEC-FIXD-001
status: draft
---

# SPEC-FIXD-001

## Requirements

### REQ-FIXD-001

The system shall do one thing.

### REQ-FIXD-002

The system shall do a second thing.

### REQ-FIXD-003

The system shall do a third thing.
`), strPtr(`# Acceptance — SPEC-FIXD-001

REQ-FIXD-003 is deliberately left without any acceptance criterion: this fixture
carries a real coverage gap, so a silent lint run is a missed defect, not a pass.

- AC-FIXD-001 (maps REQ-FIXD-001, 002): Given one, When two, Then three.
`))
	t.Logf("Group 4 output:\n%s", out)

	lines := findingLines(out, "UNCOVERED:")
	if len(lines) != 1 || lines[0] != "UNCOVERED: REQ-FIXD-003" {
		t.Errorf("%s: expected exactly one line UNCOVERED: REQ-FIXD-003; got %v", traceDrift, lines)
	}
	if !hasFinding(out, "COLLECTED: 3 REQ definitions") {
		t.Errorf("%s: heading-form definitions were not collected; got: %s", traceDrift, out)
	}
}

// TestPlanAuditTraceability_BareNumbersOutsideListDoNotExpand: a number in
// another table cell, in prose, or after a comma but followed by a word stays
// a number, so the REQs it resembles remain uncovered.
func TestPlanAuditTraceability_BareNumbersOutsideListDoNotExpand(t *testing.T) {
	t.Parallel()

	out := runTraceVerb(t, traceVerb(t, planAuditorDocs["template"]), strPtr(`---
id: SPEC-FIXE-001
status: draft
---

# SPEC-FIXE-001

## Requirements

- REQ-FIXE-001: The system shall do one thing.
- REQ-FIXE-002: The system shall do a second thing.
- REQ-FIXE-003: The system shall do a third thing.
`), strPtr(`# Acceptance — SPEC-FIXE-001

| AC | REQ | Extra |
|----|-----|-------|
| AC-FIXE-001 | REQ-FIXE-001 | 002 |

- AC-FIXE-002: exercises REQ-FIXE-001 over 003 iterations.
- AC-FIXE-003: exercises REQ-FIXE-001, 003 times over.
`))
	t.Logf("Group 4 output:\n%s", out)

	for _, want := range []string{"UNCOVERED: REQ-FIXE-002", "UNCOVERED: REQ-FIXE-003"} {
		found := false
		for _, line := range findingLines(out, "UNCOVERED:") {
			if line == want {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: a bare number outside a comma list was expanded; missing %q in: %s", traceDrift, want, out)
		}
	}
	if lines := findingLines(out, "ORPHAN:"); len(lines) != 0 {
		t.Errorf("%s: a bare number produced an undefined REQ; got ORPHAN %v", traceDrift, lines)
	}
}

// TestPlanAuditTraceability_ZeroDefinitionsIsGap: requirements named only in
// prose yield zero definitions, reported as a gap rather than a silent pass.
func TestPlanAuditTraceability_ZeroDefinitionsIsGap(t *testing.T) {
	t.Parallel()

	out := runTraceVerb(t, traceVerb(t, planAuditorDocs["template"]), strPtr(`# SPEC-FIXZ-001

## Requirements

The system covers REQ-FIXZ-001 and REQ-FIXZ-002, described here in running prose.
`), strPtr(`- AC-FIXZ-001 (maps REQ-FIXZ-001, 002): Given one, When two, Then three.
`))

	if !hasFinding(out, "GAP:") || !hasFinding(out, "COLLECTED: 0 REQ definitions") {
		t.Errorf("%s: zero definitions did not produce COLLECTED: 0 and a GAP line; got: %s", traceDrift, out)
	}
	if lines := append(findingLines(out, "UNCOVERED:"), findingLines(out, "ORPHAN:")...); len(lines) != 0 {
		t.Errorf("%s: zero definitions still produced findings; got %v", traceDrift, lines)
	}
}

// TestPlanAuditTraceability_UnreadableSpecIsGap: a missing spec is a gap.
func TestPlanAuditTraceability_UnreadableSpecIsGap(t *testing.T) {
	t.Parallel()

	out := runTraceVerb(t, traceVerb(t, planAuditorDocs["template"]), nil, nil)

	if !hasFinding(out, "GAP:") {
		t.Errorf("%s: an unreadable spec was not reported as GAP; got: %s", traceDrift, out)
	}
	if hasFinding(out, "COLLECTED:") || hasFinding(out, "UNCOVERED:") {
		t.Errorf("%s: an unreadable spec produced a measurement; got: %s", traceDrift, out)
	}
}

// TestPlanAuditTraceability_InlineACsWithoutAcceptance: a small-tier SPEC with
// inline ACs and no acceptance.md is still traced, and the absent input is
// named on the COLLECTED line.
func TestPlanAuditTraceability_InlineACsWithoutAcceptance(t *testing.T) {
	t.Parallel()

	out := runTraceVerb(t, traceVerb(t, planAuditorDocs["template"]), strPtr(`# SPEC-FIXS-001

## Requirements

- **REQ-FIXS-001**: The system shall do one thing.
- **REQ-FIXS-002**: The system shall do a second thing.
- **REQ-FIXS-003**: The system shall do a third thing.

## Acceptance Criteria

- AC-FIXS-001 (maps REQ-FIXS-001, 002): Given one, When two, Then three.
`), nil)

	if !hasFinding(out, "COLLECTED: 3 REQ definitions (acceptance input: absent)") {
		t.Errorf("%s: expected COLLECTED: 3 with an absent acceptance input; got: %s", traceDrift, out)
	}
	lines := findingLines(out, "UNCOVERED:")
	if len(lines) != 1 || lines[0] != "UNCOVERED: REQ-FIXS-003" {
		t.Errorf("%s: expected exactly one line UNCOVERED: REQ-FIXS-003; got %v", traceDrift, lines)
	}
}

// TestPlanAuditTraceability_VerbIdenticalAcrossCopies: the local mirror carries
// the same Group 4 verb as the distributed template copy.
func TestPlanAuditTraceability_VerbIdenticalAcrossCopies(t *testing.T) {
	t.Parallel()

	if traceVerb(t, planAuditorDocs["template"]) != traceVerb(t, planAuditorDocs["local"]) {
		t.Errorf("%s: the Group 4 verb differs between the template and local copies", traceDrift)
	}
}

// TestPlanAuditTraceability_CitationDisciplineClausePresent: both copies carry
// the citation-discipline clause inside the Group 4 section.
func TestPlanAuditTraceability_CitationDisciplineClausePresent(t *testing.T) {
	t.Parallel()

	for name, doc := range planAuditorDocs {
		raw, err := os.ReadFile(doc)
		if err != nil {
			t.Fatalf("%s: read %s: %v", traceDrift, doc, err)
		}
		body := string(raw)
		start := strings.Index(body, "\n### Group 4:")
		if start < 0 {
			t.Fatalf("%s: no Group 4 heading in the %s copy", traceDrift, name)
		}
		section := body[start+1:]
		if end := strings.Index(section, "\n### "); end >= 0 {
			section = section[:end]
		}
		if !strings.Contains(section, citationDisciplineKey) {
			t.Errorf("%s: the %s copy's Group 4 section lacks the citation-discipline clause", traceDrift, name)
		}
	}
}
