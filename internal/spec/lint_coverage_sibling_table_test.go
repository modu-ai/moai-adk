package spec_test

// lint_coverage_sibling_table_test.go — CoverageRule reading table-form REQ
// mappings from the sibling acceptance.md (card t561, GH #1696).
//
// The sibling reader already collects the `maps REQ-…` form. The corpus mostly
// writes a table instead (`| AC-… | REQ-…, 002 | … |`), which that form cannot
// see. The rule pinned here, as decided for the card:
//
//   - a column is read only when its header cell names a requirement
//     (REQ / 요구 / requirement, case-insensitive); a table with no such column
//     collects nothing, which is today's behavior;
//   - a bare numeric tail expands to the prefix of a full REQ id only when that
//     full id comes before it, comma-separated, in the SAME cell.
//
// Fixtures follow lint_coverage_sibling_test.go: one case per directory under
// testdata/coveragesibling/, linted one at a time, acceptance.md without
// frontmatter. Judgment is by the exact set of uncovered REQ ids.

import (
	"regexp"
	"sort"
	"strings"
	"testing"
)

var uncoveredREQPattern = regexp.MustCompile(`REQ-[A-Z0-9-]+`)

// uncoveredREQs returns the sorted REQ ids named by CoverageIncomplete findings.
func uncoveredREQs(t *testing.T, c string) []string {
	t.Helper()
	var ids []string
	for _, f := range coverageSiblingFindings(t, c) {
		ids = append(ids, uncoveredREQPattern.FindString(f.Message))
	}
	sort.Strings(ids)
	return ids
}

// TestCoverageSiblingTable_HeaderColumnWithSameCellExpansion: a Korean
// `요구사항` header identifies the REQ column, `REQ-CST-001-001, 002` covers
// both 001 and 002 by same-cell expansion, and a second row covers 003.
//
// Before the repair all three REQs are reported, because the sibling reader
// only knows the `maps REQ-…` form.
func TestCoverageSiblingTable_HeaderColumnWithSameCellExpansion(t *testing.T) {
	if got := uncoveredREQs(t, "sibling-table-expand"); len(got) != 0 {
		t.Fatalf("uncovered REQs = %v, want none", got)
	}
}

// TestCoverageSiblingTable_OnlyHeaderColumnsAndOnlySameCell pins both scopings
// in one run. Exactly REQ-CST-002-002 and REQ-CST-002-003 must stay uncovered:
//
//   - 002 is a bare tail in the `Related REQ` cell with no full id before it in
//     that cell — expanding it from the neighbouring cell would hide it;
//   - 003 appears only in the `Note` column, whose header names no requirement —
//     reading every column would hide it.
//
// MUTATIONS: dropping header identification leaves only {002}; dropping the
// cell boundary leaves only {003}. Either way the set differs from the one
// asserted here.
func TestCoverageSiblingTable_OnlyHeaderColumnsAndOnlySameCell(t *testing.T) {
	got := uncoveredREQs(t, "sibling-table-scoped")
	want := []string{"REQ-CST-002-002", "REQ-CST-002-003"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("uncovered REQs = %v, want %v", got, want)
	}
}

// TestCoverageSiblingTable_RowWithoutACIsNotCoverage: in a REQ-first
// traceability table the REQ column is header-identified, but a row whose AC
// cell names no AC (`| REQ-… | (Optional) | AC 면제 |`) is not a mapping and
// its REQ stays uncovered. The row that names an AC is covered.
//
// MUTATION: dropping the row condition makes REQ-CST-004-002 covered, leaving
// no uncovered REQ.
func TestCoverageSiblingTable_RowWithoutACIsNotCoverage(t *testing.T) {
	got := uncoveredREQs(t, "sibling-table-req-first")
	want := []string{"REQ-CST-004-002"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("uncovered REQs = %v, want %v", got, want)
	}
}

// TestCoverageSiblingTable_UnidentifiedHeaderCollectsNothing: a table whose
// header names no requirement contributes nothing, even though its rows
// mention a REQ id — the same result the linter produces today.
func TestCoverageSiblingTable_UnidentifiedHeaderCollectsNothing(t *testing.T) {
	got := uncoveredREQs(t, "sibling-table-no-header")
	want := []string{"REQ-CST-003-001"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("uncovered REQs = %v, want %v", got, want)
	}
}
