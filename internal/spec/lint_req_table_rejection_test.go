package spec

import (
	"path/filepath"
	"strings"
	"testing"
)

// SPEC-SPEC-LINT-BLIND-AXES-001 axis 1 (card t518, milestone M-A2b).
//
// AC-SLB-005 — rejection observability. M-A1 landed discriminator C-d, which
// rejects a table row carrying no L1 modality marker. The rejection was
// SILENT: a disposition table and a document with no requirements at all
// produced byte-identical output, which is the same "empty looks like a pass"
// shape this whole card exists to break.
//
// THE FOLD IS PER TABLE, AND THE LINE CARRIES N. One advisory finding per
// rejected TABLE, not per rejected row (787 per-row lines would push default
// output past the point where warnings get read). The count N of rejected rows
// rides on the line, which makes the OUTPUT a second instrument for the census
// figure — the sum of every N is cross-checkable against a direct row scan.
// This card has watched a single instrument go quiet three times; one
// instrument is exactly as many as it takes to not notice.
//
// The §A disciplines bind here as everywhere: control pairs (a rejected table
// AND an accepted one), no exit-code assertions, and mutant results recorded in
// progress.md §E.2 including the mutants that were NOT caught.

// rejectionFindings runs the rule over a body and returns its findings.
func rejectionFindings(t *testing.T, body string) []Finding {
	t.Helper()
	doc := &SPECDoc{Path: "fixture/spec.md", Body: body}
	return (&REQTableRejectionRule{}).Check(doc, nil)
}

// countRejectedRows counts, by direct line scan, the rows discriminator C-d
// rejects. It is the INDEPENDENT half of every count assertion below — the rule
// must not be measured against itself.
func countRejectedRows(body string) int {
	n := 0
	for _, line := range strings.Split(body, "\n") {
		if reqTableRowPattern.MatchString(line) && !isTableDefinitionRow(line) {
			n++
		}
	}
	return n
}

// multiRowRejectedTable is ONE table carrying three rejected rows. It measures
// the fold: three rows, one line.
const multiRowRejectedTable = `## 추적

| 요구 | 수락 |
|---|---|
| REQ-FXR-001 | AC-FXR-001 |
| REQ-FXR-002 | AC-FXR-002 |
| REQ-FXR-003 | AC-FXR-003 |
`

// mixedTableBody carries ONE table holding one definition row and two
// non-definition rows. It separates "N counts rejected rows" from "N counts
// rows".
const mixedTableBody = `## 요구

| 요구 | 내용 |
|---|---|
| REQ-FXM-001 | 도구는 기각을 기록해야 한다 |
| REQ-FXM-002 | 폐기 — t400에서 흡수 |
| REQ-FXM-003 | AC-FXM-003 |
`

// --- AC-SLB-005 (rejection is observable) — REQ-SLB-005 -------------------

func TestTableRejection_OneLinePerRejectedTable(t *testing.T) {
	body := dispositionFixtureBody

	// The fixture must actually produce what the assertions below are about.
	// A control that matches nothing is not a control — this card has already
	// written four of those.
	rejected := countRejectedRows(body)
	if rejected != 2 {
		t.Fatalf("fixture precondition: %d rejected rows, want 2 (one disposition, one tracking) — the fixture is not exercising the rejection path", rejected)
	}
	if collected := parseREQsTable(body); len(collected) != 0 {
		t.Fatalf("fixture precondition: %d rows were COLLECTED, want 0 — this fixture must be pure rejection", len(collected))
	}

	got := rejectionFindings(t, body)
	if len(got) != 2 {
		t.Fatalf("emitted %d findings, want 2 (one per rejected table): %#v", len(got), got)
	}

	sum := 0
	lines := strings.Split(body, "\n")
	for i, f := range got {
		if f.Code != "REQTableRowsRejected" {
			t.Errorf("finding %d: code = %q, want REQTableRowsRejected", i, f.Code)
		}
		n := rejectedRowCountOf(t, f)
		if n != 1 {
			t.Errorf("finding %d: N = %d, want 1 (each fixture table rejects exactly one row)", i, n)
		}
		sum += n
		// The line must point INTO the table it is reporting, not at a
		// document-level position.
		if f.Line < 1 || f.Line > len(lines) {
			t.Fatalf("finding %d: line %d out of range", i, f.Line)
		}
		if !reqTableRowPattern.MatchString(lines[f.Line-1]) {
			t.Errorf("finding %d: line %d = %q, want a REQ-ID table row", i, f.Line, lines[f.Line-1])
		}
	}
	// [HARD] The control: the folded counts must add back up to the rows.
	if sum != rejected {
		t.Errorf("sum of N = %d, direct row scan = %d — folding lost or invented rows", sum, rejected)
	}
}

// --- The control pair: a rejected table AND an accepted one ---------------

func TestTableRejection_ControlPairDiverges(t *testing.T) {
	accepted := rejectionFindings(t, tableFixtureBody)
	rejectedF := rejectionFindings(t, dispositionFixtureBody)

	// Both halves must be non-vacuous in their OWN terms before the divergence
	// means anything: the accepted half must actually collect, the rejected
	// half must actually reject.
	if n := len(parseREQsTable(tableFixtureBody)); n == 0 {
		t.Fatalf("accepted half collected 0 rows — the pair is vacuous")
	}
	if n := countRejectedRows(dispositionFixtureBody); n == 0 {
		t.Fatalf("rejected half rejected 0 rows — the pair is vacuous")
	}

	if len(accepted) != 0 {
		t.Errorf("accepted half: %d rejection findings, want 0 — a definition table must not be reported as rejected: %#v", len(accepted), accepted)
	}
	if len(rejectedF) == 0 {
		t.Error("rejected half: 0 rejection findings, want ≥1 — this is the silence the milestone removes")
	}
	if len(accepted) == len(rejectedF) {
		t.Errorf("the two halves did not diverge (both %d findings) — a pair that does not split proves nothing", len(accepted))
	}
}

// --- The fold itself: three rows, one line -------------------------------

func TestTableRejection_MultipleRowsFoldIntoOneFinding(t *testing.T) {
	rejected := countRejectedRows(multiRowRejectedTable)
	if rejected != 3 {
		t.Fatalf("fixture precondition: %d rejected rows, want 3", rejected)
	}

	got := rejectionFindings(t, multiRowRejectedTable)
	if len(got) != 1 {
		t.Fatalf("emitted %d findings, want exactly 1 — the fold is per TABLE, not per row: %#v", len(got), got)
	}
	if n := rejectedRowCountOf(t, got[0]); n != rejected {
		t.Errorf("N = %d, want %d — the line must carry the table's rejected-row count", n, rejected)
	}
}

// N counts REJECTED rows, not rows. A mixed table separates the two.
func TestTableRejection_MixedTableCountsOnlyRejectedRows(t *testing.T) {
	collected := parseREQsTable(mixedTableBody)
	rejected := countRejectedRows(mixedTableBody)
	if len(collected) != 1 || rejected != 2 {
		t.Fatalf("fixture precondition: collected=%d rejected=%d, want 1 and 2 — the fixture must be genuinely mixed", len(collected), rejected)
	}

	got := rejectionFindings(t, mixedTableBody)
	if len(got) != 1 {
		t.Fatalf("emitted %d findings, want 1: %#v", len(got), got)
	}
	if n := rejectedRowCountOf(t, got[0]); n != rejected {
		t.Errorf("N = %d, want %d — N must count rejected rows only, not every row in the table", n, rejected)
	}
}

// A document with no table at all must stay silent. Without this, "reports
// every rejection" and "reports unconditionally" are indistinguishable.
func TestTableRejection_SilentWhenNothingIsRejected(t *testing.T) {
	if got := rejectionFindings(t, listFixtureBody); len(got) != 0 {
		t.Errorf("list-only fixture produced %d rejection findings, want 0: %#v", len(got), got)
	}
}

// A body that ENDS on a rejected row, with no trailing newline, must still
// report it.
//
// [HARD] This guard was added AFTER mutant M4 (trailing flush removed) escaped
// the first-pass suite, and the first-pass result is kept: "caught because a
// guard was added afterwards" is a different fact from "a guard existed". Every
// other fixture here — and every spec.md in the corpus — ends with a newline,
// so strings.Split yields a trailing empty line that flushes the run as a side
// effect. The trailing flush was therefore live code no test could reach, which
// is exactly the shape that rots unnoticed.
func TestTableRejection_TableRunningToEndOfBody(t *testing.T) {
	body := strings.TrimRight(multiRowRejectedTable, "\n")
	if strings.HasSuffix(body, "\n") {
		t.Fatal("fixture precondition: body must NOT end with a newline, or it does not exercise the boundary")
	}

	got := rejectionFindings(t, body)
	if len(got) != 1 {
		t.Fatalf("emitted %d findings, want 1 — a table that runs to the end of the body is still a table: %#v", len(got), got)
	}
	if n := rejectedRowCountOf(t, got[0]); n != countRejectedRows(body) {
		t.Errorf("N = %d, want %d", n, countRejectedRows(body))
	}
}

// --- Severity: this signal reports, it never gates -----------------------

func TestTableRejection_IsAdvisoryWarning(t *testing.T) {
	got := rejectionFindings(t, dispositionFixtureBody)
	if len(got) == 0 {
		t.Fatal("no findings — cannot judge severity of nothing")
	}
	for i, f := range got {
		if f.Severity != SeverityWarning {
			t.Errorf("finding %d: severity = %v, want warning", i, f.Severity)
		}
		if !f.Advisory {
			t.Errorf("finding %d: Advisory = false — a signal that only records non-collection must never gate a build", i)
		}
	}
}

// The rule must be wired into the linter. Check() passing in isolation while
// the rule is unregistered is exactly the "the instrument was never run" shape.
func TestTableRejection_RuleIsRegistered(t *testing.T) {
	l := NewLinter(LinterOptions{})
	for _, r := range l.rules {
		if r.Code() == "REQTableRowsRejected" {
			return
		}
	}
	t.Fatal("REQTableRowsRejected is not in the linter's rule set — the finding can never reach a user")
}

// --- The corpus control: sum of N == the census -------------------------
//
// [HARD] This is the third instrument. The census test in
// lint_req_table_test.go counts rejected rows directly; this one adds up what
// the OUTPUT says. If the two ever disagree, one of them is lying and the
// disagreement is visible rather than silent.
func TestTableRejection_CorpusSumEqualsRowCensus(t *testing.T) {
	root := findRepoRoot(t)
	paths, err := filepath.Glob(filepath.Join(root, ".moai/specs/SPEC-*/spec.md"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}

	rule := &REQTableRejectionRule{}
	censusRows, emittedSum, lines := 0, 0, 0
	for _, p := range paths {
		doc := parseSPECDoc(p)
		if doc.ParseError != nil {
			continue
		}
		censusRows += countRejectedRows(doc.Body)
		for _, f := range rule.Check(doc, nil) {
			lines++
			emittedSum += rejectedRowCountOf(t, f)
		}
	}

	if censusRows == 0 || emittedSum == 0 {
		t.Fatalf("census is vacuous: rows=%d emitted=%d — an empty operand makes this comparison assert nothing", censusRows, emittedSum)
	}
	if emittedSum != censusRows {
		t.Errorf("sum of N across %d emitted lines = %d, direct row census = %d", lines, emittedSum, censusRows)
	}
	if lines >= censusRows {
		t.Errorf("emitted %d lines for %d rejected rows — the fold is not folding", lines, censusRows)
	}
	t.Logf("corpus rejection: rows=%d emitted lines=%d sum(N)=%d", censusRows, lines, emittedSum)
}

// rejectedRowCountOf reads N back out of the finding message. Reading the
// count from the rendered MESSAGE (rather than from a struct field the user
// never sees) is deliberate: the message is what an operator reads, and it is
// the surface the sum control is supposed to be checking.
func rejectedRowCountOf(t *testing.T, f Finding) int {
	t.Helper()
	n, ok := parseRejectedRowCount(f.Message)
	if !ok {
		t.Fatalf("message does not carry a rejected-row count: %q", f.Message)
	}
	return n
}
