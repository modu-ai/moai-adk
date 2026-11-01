package closure

import "testing"

// TestParseProgressRowsIgnoresHeadingsInsideCodeFences: a "# comment" line in a
// fenced code block under §E.2 is not a heading and must not close the section
// (second card review finding on the S13 repair).
func TestParseProgressRowsIgnoresHeadingsInsideCodeFences(t *testing.T) {
	t.Parallel()

	const doc = "# Progress\n\n" +
		"## §E.2 Run-phase Evidence\n\n" +
		"| AC-1 | PASS | before the fence |\n\n" +
		"```text\n# command output, not a heading\n## also not a heading\n```\n\n" +
		"| AC-2 | PASS | after the fence, still inside E.2 |\n\n" +
		"## §E.4 Sync-phase Signal\n\n" +
		"| AC-9 | PASS | outside the matrix |\n"

	rows, found := ParseProgressRows([]byte(doc))
	if !found {
		t.Fatal("the §E.2 heading exists and must be found")
	}
	var ids []string
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	want := []string{"AC-1", "AC-2"}
	if len(ids) != len(want) || ids[0] != want[0] || ids[1] != want[1] {
		t.Fatalf("S13_FENCE_HEADING_CLOSES_E2: rows = %v, want %v", ids, want)
	}
}

// TestParseProgressRowsAcceptsIndentedHeadings: Markdown allows up to three
// leading spaces before an ATX heading, so " ## §E.2 ..." opens the section and
// an indented following heading closes it (fourth card review finding on the
// S13 repair); four spaces is an indented code block, not a heading.
func TestParseProgressRowsAcceptsIndentedHeadings(t *testing.T) {
	t.Parallel()

	const doc = " ## §E.2 Run-phase Evidence\n\n" +
		"| AC-1 | PASS | under an indented heading |\n\n" +
		"   ## §E.4 Sync-phase Signal\n\n" +
		"| AC-9 | PASS | outside the matrix |\n"
	rows, found := ParseProgressRows([]byte(doc))
	if !found || len(rows) != 1 || rows[0].ID != "AC-1" {
		t.Fatalf("S13_INDENTED_HEADING: found=%v rows=%v, want the one AC-1 row", found, rows)
	}

	const codeBlock = "    ## §E.2 Run-phase Evidence\n\n| AC-1 | PASS | a four-space code block is not a heading |\n"
	if rows, found := ParseProgressRows([]byte(codeBlock)); found || len(rows) != 0 {
		t.Fatalf("S13_INDENTED_CODE_BLOCK: found=%v rows=%v, want no §E.2 section", found, rows)
	}
}

// TestParseProgressRowsFenceClosesOnlyOnAMatchingFence: a four-backtick fence
// that quotes a three-backtick example is closed by a four-backtick line, not by
// the quoted three-backtick lines (third card review finding on the S13 repair).
func TestParseProgressRowsFenceClosesOnlyOnAMatchingFence(t *testing.T) {
	t.Parallel()

	const doc = "## §E.2 Run-phase Evidence\n\n" +
		"| AC-1 | PASS | before the fence |\n\n" +
		"````markdown\n" +
		"```text\n" +
		"| AC-7 | PASS | an example row quoted inside the outer fence |\n" +
		"```\n" +
		"# quoted, not a heading\n" +
		"````\n\n" +
		"| AC-2 | PASS | after the outer fence, still inside E.2 |\n\n" +
		"## §E.4 Sync-phase Signal\n\n" +
		"| AC-9 | PASS | outside the matrix |\n"

	rows, found := ParseProgressRows([]byte(doc))
	if !found {
		t.Fatal("the §E.2 heading exists and must be found")
	}
	var ids []string
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	want := []string{"AC-1", "AC-2"}
	if len(ids) != len(want) || ids[0] != want[0] || ids[1] != want[1] {
		t.Fatalf("S13_FENCE_KIND_LENGTH: rows = %v, want %v", ids, want)
	}
}

// TestParseProgressRowsKeepsE2DepthAcrossSubheadings: a sub-heading under the
// §E.2 heading that itself names §E.2 (for example "### §E.2.1") must not move
// the section's reference depth, or the next sibling sub-heading closes the
// section early and AC rows that still belong to §E.2 go unreported (card
// review finding on the S13 repair).
func TestParseProgressRowsKeepsE2DepthAcrossSubheadings(t *testing.T) {
	t.Parallel()

	const doc = `# Progress

## §E.2 Run-phase Evidence

| AC-1 | PASS | top of the matrix |

### §E.2.1 Details

| AC-2 | PASS | under the E.2.1 sub-heading |

### Further detail

| AC-3 | PASS | sibling sub-heading, still inside E.2 |

## §E.4 Sync-phase Signal

| AC-9 | PASS | outside the matrix |
`
	rows, found := ParseProgressRows([]byte(doc))
	if !found {
		t.Fatal("the §E.2 heading exists and must be found")
	}
	var ids []string
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	want := []string{"AC-1", "AC-2", "AC-3"}
	if len(ids) != len(want) {
		t.Fatalf("S13_SUBHEADING_CLOSES_E2: rows = %v, want %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("S13_SUBHEADING_CLOSES_E2: rows = %v, want %v", ids, want)
		}
	}
}

// TestParseProgressRowsStaysInsideE2 (card t2009, S13): the AC matrix is the
// table under the §E.2 heading. Tables elsewhere in progress.md (a later
// section's reference table, an appendix) and a body sentence that merely
// mentions §E.2 are not matrix rows.
func TestParseProgressRowsStaysInsideE2(t *testing.T) {
	t.Parallel()

	const withLaterTables = `# Progress

## §E.2 Run-phase Evidence

| AC | Status | Evidence |
|----|--------|----------|
| AC-1 | PASS | go test ok |
| AC-2 | FAIL | go test failed |

### E.2 notes

| AC-3 | PASS | nested under E.2 |

## §E.4 Sync-phase Signal

| AC-1 | NOT_RUN | reference table outside the matrix |
| AC-9 | PASS | appendix-only row |

## Appendix

| AC-2 | PASS | stale copy |
`
	rows, found := ParseProgressRows([]byte(withLaterTables))
	if !found {
		t.Fatal("the §E.2 heading exists and must be found")
	}
	var ids []string
	for _, r := range rows {
		ids = append(ids, r.ID+"="+r.Status)
	}
	want := []string{"AC-1=PASS", "AC-2=FAIL", "AC-3=PASS"}
	if len(ids) != len(want) {
		t.Fatalf("S13_ROWS_OUTSIDE_E2: rows = %v, want exactly the matrix under §E.2 %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("S13_ROWS_OUTSIDE_E2: rows = %v, want %v", ids, want)
		}
	}

	const mentionOnly = `# Progress

The matrix lives in §E.2, which this run has not written yet.

## §E.4 Sync-phase Signal

| AC-1 | PASS | table in another section |
`
	rows, found = ParseProgressRows([]byte(mentionOnly))
	if found || len(rows) != 0 {
		t.Fatalf("S13_MENTION_IS_NOT_A_HEADING: found=%v rows=%v, want no §E.2 section and no rows", found, rows)
	}
}
