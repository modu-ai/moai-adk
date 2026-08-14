// column_test.go — the six-column closed enumeration
// (SPEC-KANBAN-BOARD-001 REQ-KB-003, M2; decided by `go test ./internal/kanban/ -run Column`).
//
// The six are fixed and ordered; no seventh value exists and no constructor
// accepts a value outside the set. A `test` column and an operator-extensible
// set are both rejected decisions (spec.md §C), and no held or blocked column
// exists — the unheld state is a card field, not a column (§A.5).
package kanban

import (
	"testing"
)

// TestColumn_EnumerationExactlySixOrdered — AC-KB-007: six values, in the
// fixed order backlog, plan, run, review, sync, done.
func TestColumn_EnumerationExactlySixOrdered(t *testing.T) {
	t.Parallel()
	want := []Column{
		ColumnBacklog, ColumnPlan, ColumnRun, ColumnReview, ColumnSync, ColumnDone,
	}
	got := Columns()
	if len(got) != len(want) {
		t.Fatalf("column count = %d, want exactly 6 (no seventh value)", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Columns()[%d] = %q, want %q — the order is fixed", i, got[i], want[i])
		}
	}
}

// TestColumn_ParseRejectsOutsideSet — AC-KB-007: no constructor accepts a
// value outside the set — including the rejected `test` column, a
// held/blocked pseudo-column, and arbitrary strings.
func TestColumn_ParseRejectsOutsideSet(t *testing.T) {
	t.Parallel()
	for _, bad := range []string{"", "test", "blocked", "held", "waiting", "wip", "RUN", "backlogs", "archived", "reviewed"} {
		if _, err := ParseColumn(bad); err == nil {
			t.Errorf("ParseColumn(%q) err = nil, want rejection — the set is closed", bad)
		}
	}
}

// TestColumn_ParseAcceptsEveryDeclaredValue — the constructor's positive
// control: every declared value parses.
func TestColumn_ParseAcceptsEveryDeclaredValue(t *testing.T) {
	t.Parallel()
	for _, c := range Columns() {
		got, err := ParseColumn(string(c))
		if err != nil {
			t.Fatalf("ParseColumn(%q) error = %v", c, err)
		}
		if got != c {
			t.Fatalf("ParseColumn(%q) = %q, round-trip broken", c, got)
		}
	}
}

// TestColumn_DispatchableColumnConditional — REQ-KB-012's column predicate
// used by AC-KB-014: backlog and done have no owning session; every working
// column does.
func TestColumn_DispatchableColumnConditional(t *testing.T) {
	t.Parallel()
	if ColumnBacklog.HasOwningSession() {
		t.Error("backlog has an owning session; it is a queue")
	}
	if ColumnDone.HasOwningSession() {
		t.Error("done has an owning session; it is terminal")
	}
	for _, c := range []Column{ColumnPlan, ColumnRun, ColumnReview, ColumnSync} {
		if !c.HasOwningSession() {
			t.Errorf("column %q has no owning session; only backlog and done are ownerless", c)
		}
	}
}
