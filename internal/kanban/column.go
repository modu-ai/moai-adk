// column.go — the five-column closed enumeration
// (SPEC-KANBAN-BOARD-001 REQ-KB-003, M2; review removed per t113 — the
// three working columns are plan → run → sync, matching the D1 structure
// where the sync gate absorbs the review verdict).
//
// The columns are fixed and ordered: backlog → plan → run → sync →
// done. No sixth value exists and the set is not operator-extensible. A
// `test` column was considered and rejected (unit tests belong to the cycle
// inside `run`), and a held or blocked column is unnecessary — the unheld
// state is a card field serving both of its causes (spec.md §A.5).
package kanban

import "fmt"

// Column is one of the five board columns. The zero value is not a member of
// the set; every construction goes through ParseColumn or the constants.
type Column string

// The five columns, in their fixed order (REQ-KB-003).
const (
	ColumnBacklog Column = "backlog"
	ColumnPlan    Column = "plan"
	ColumnRun     Column = "run"
	ColumnSync    Column = "sync"
	ColumnDone    Column = "done"
)

// allColumns is the ordered, closed set. Appending here would create a
// sixth column, which REQ-KB-003 forbids; the list is deliberately
// exhaustive and unexported so no operator can extend it at runtime.
var allColumns = []Column{
	ColumnBacklog, ColumnPlan, ColumnRun, ColumnSync, ColumnDone,
}

// Columns returns the five columns in their fixed order.
func Columns() []Column {
	out := make([]Column, len(allColumns))
	copy(out, allColumns)
	return out
}

// ParseColumn is the constructor: it accepts exactly the five declared values
// and rejects everything else, so no value outside the set can enter the
// board through any path that constructs a Column.
func ParseColumn(s string) (Column, error) {
	for _, c := range allColumns {
		if string(c) == s {
			return c, nil
		}
	}
	return "", fmt.Errorf("parse column: %q is not one of the five columns", s)
}

// HasOwningSession reports whether the column maps onto a session role.
// backlog is a queue with no owning session and done is terminal with none;
// every working column has exactly one (REQ-KB-012 — the roles themselves
// are the bootstrap sibling's).
func (c Column) HasOwningSession() bool {
	switch c {
	case ColumnBacklog, ColumnDone:
		return false
	default:
		return true
	}
}

// Valid reports whether c is a member of the closed set.
func (c Column) Valid() bool {
	for _, want := range allColumns {
		if c == want {
			return true
		}
	}
	return false
}
