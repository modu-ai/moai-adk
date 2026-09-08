package spec

import (
	"fmt"
	"strconv"
	"strings"
)

// Rejection observability — SPEC-SPEC-LINT-BLIND-AXES-001 axis 1, REQ-SLB-005
// (card t518, milestone M-A2b).
//
// THE DEFECT THIS REMOVES. M-A1 landed discriminator C-d: a table row with a
// leading REQ ID is collected as a definition only when the row carries an L1
// modality marker. Everything else is rejected — 787 rows corpus-wide — and the
// rejection was entirely silent. A disposition table and a document with no
// requirements at all produced byte-identical output, which is the same
// "silence looks like a pass" shape the whole card exists to break. C-d's
// measured miss (SPEC-INIT-001's `| REQ-N-001 | 시스템은 … 덮어쓰지 않아야 한다 |`,
// a real definition C-d drops) was accepted as a CHOSEN COST on the explicit
// premise that the miss stays observable. This file is what makes that premise
// true; without it the cost was accepted and not paid.
//
// THE FOLD IS PER TABLE, NOT PER ROW — AND THE COST OF THE FOLD IS NAMED. One
// finding per rejected TABLE. 787 per-row lines would land on top of the 460
// findings axis 1 already added (+5.1% on default output) and push warnings
// into the range where operators stop reading them, which would exceed the cost
// the operator actually approved. The fold moves the unit of observation from
// row to table, so what is now visible is precisely: WHICH TABLE was rejected,
// with WHICH ROW inside it visible only as the count N. A --verbose per-row
// expansion was considered and rejected — a second emission path widens this
// card's radius and needs its own counts verified separately.
//
// WHY N RIDES ON THE LINE. Without N, the corpus census test is the SOLE
// evidence for the 787 figure, and this card has watched a sole instrument go
// quiet three times (M7's unmeasured bold-ID tolerance, M8's `"THE "` prefix,
// and acceptance.md's own REQ-FX-011 fixture that could not make its guard).
// With N, the output itself is a second instrument: the sum of every emitted N
// is cross-checkable against a direct row scan, and a disagreement between the
// two is visible rather than silent. That control is
// TestTableRejection_CorpusSumEqualsRowCensus.
//
// SEVERITY. Advisory warning, set here for the whole code and NOT through
// reqFindingSeverity — exactly as ModalityUnjudged does, and for the same
// reason: this is a per-CODE property (a signal that only records
// non-collection must never gate a build), not a second entry-keyed severity
// axis. There is no REQEntry behind these findings at all: a rejected row
// produced no entry, which is the fact being reported.

// rejectedTableRunPrefix marks a markdown table line. A table is a maximal run
// of consecutive lines matching it — header, separator and body rows alike.
// Grouping by adjacency is what turns rows into tables, and it is deliberately
// the cheapest rule that does so: any grouping produces the same SUM (every
// rejected row lies inside exactly one run), so the census control holds
// whatever the grouping, and only the LINE COUNT depends on this choice.
func isTableLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "|")
}

// rejectedTable is one contiguous markdown table that discriminator C-d
// declined to read as a definition table, together with how many of its rows it
// declined.
type rejectedTable struct {
	// Line is the 1-based line of the FIRST rejected row in the table — a
	// position that points at real content the reader can go and look at,
	// rather than at a table header or a document-level position.
	Line int
	// Count is how many rows in this table were rejected. Rows the
	// discriminator ADMITTED are not counted: a mixed table reports only what
	// was left behind.
	Count int
}

// collectRejectedTables groups the rows discriminator C-d rejected into the
// tables that carry them, in document order.
func collectRejectedTables(body string) []rejectedTable {
	var out []rejectedTable
	lines := strings.Split(body, "\n")

	cur := rejectedTable{}
	flush := func() {
		if cur.Count > 0 {
			out = append(out, cur)
		}
		cur = rejectedTable{}
	}

	for i, line := range lines {
		if !isTableLine(line) {
			// A non-table line ends the run, and therefore the table.
			flush()
			continue
		}
		if !reqTableRowPattern.MatchString(line) {
			continue // header, separator, or a row with no leading REQ ID
		}
		if isTableDefinitionRow(line) {
			continue // admitted — collected as a definition, nothing to report
		}
		if cur.Count == 0 {
			cur.Line = i + 1
		}
		cur.Count++
	}
	flush()

	return out
}

// rejectedRowCountToken is the machine-readable carrier of N inside the
// message. It exists so the count can be read back off the RENDERED OUTPUT —
// the surface an operator actually reads — rather than off a struct field no
// user ever sees. The sum control reads this token; a message that stops
// carrying it fails that control loudly instead of degrading quietly.
const rejectedRowCountToken = "rejected_rows="

// parseRejectedRowCount reads N back out of a rendered finding message.
func parseRejectedRowCount(message string) (int, bool) {
	idx := strings.Index(message, rejectedRowCountToken)
	if idx < 0 {
		return 0, false
	}
	rest := message[idx+len(rejectedRowCountToken):]
	end := 0
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, false
	}
	n, err := strconv.Atoi(rest[:end])
	if err != nil {
		return 0, false
	}
	return n, true
}

// REQTableRejectionRule reports every table whose REQ-ID rows discriminator C-d
// declined to read as definitions.
type REQTableRejectionRule struct{}

func (r *REQTableRejectionRule) Code() string { return "REQTableRowsRejected" }

func (r *REQTableRejectionRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	var findings []Finding
	for _, tbl := range collectRejectedTables(doc.Body) {
		findings = append(findings, Finding{
			File:     doc.Path,
			Line:     tbl.Line,
			Severity: SeverityWarning,
			Advisory: true, // records non-collection; never gates — see the header
			Code:     "REQTableRowsRejected",
			Message: fmt.Sprintf(
				"table NOT read as a requirement-definition table (%s%d): rows here carry a leading REQ ID but none of the modality markers the discriminator keys on (SHALL / 해야 한다 / 해서는 안 된다), so they were not collected. This is NOT a claim that those rows are malformed — it is what keeps \"no requirement here\" and \"not collected here\" distinguishable in this output. Counted per table, not per row.",
				rejectedRowCountToken, tbl.Count),
		})
	}
	return findings
}
