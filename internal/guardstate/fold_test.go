package guardstate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stateTableRowCells returns every numbered row of the normative table as its
// six cells, keyed by row number. The table is the artifact (REQ-GSM-006); this
// reads it rather than restating it, so a row re-decided there is re-read here
// instead of drifting from a paraphrase.
func stateTableRowCells(t *testing.T) map[string][]string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), specTablePath))
	if err != nil {
		t.Fatalf("read %s: %v", specTablePath, err)
	}
	rows := map[string][]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "| ") {
			continue
		}
		cells := strings.Split(strings.Trim(trimmed, "|"), "|")
		if len(cells) != 6 {
			continue
		}
		num := strings.TrimSpace(cells[0])
		if len(num) != 1 || num[0] < '1' || num[0] > '9' {
			continue
		}
		rows[num] = cells
	}
	return rows
}

// tableCell strips the markdown emphasis and code fencing a cell may carry, so
// a comparison is against the VALUE rather than against its formatting.
func tableCell(s string) string {
	return strings.TrimSpace(strings.NewReplacer("`", "", "*", "").Replace(s))
}

// ---------------------------------------------------------------------------
// AC-GSM-016 — every classification folds, the fold EQUALS the table's
// `Surface fold` column, none folds to `fail`, and the meaningless/incomplete
// boundary holds.
//
// Mutant (a) kills: a presence check ("some fold exists") is satisfied by a
// fold that is wrong for every value. The assertion is an EQUALITY against the
// column, value-against-column.
//
// Mutant (b) kills: an implementation folding UNDECLARED or STALE to `fail`
// satisfies every other criterion while promoting a routine sweep to a failing
// exit status. Card t326 reached the same conclusion for the same reason.
//
// Mutant (c) kills, and it is this SPEC's own subject: an implementation that
// adopts t326's leniency WHOLESALE folds all uncertainty to `ok` — UNREADABLE,
// UNKNOWN and UNRESOLVED alike. It passes (a) and (b), reads as principled
// reuse, and reproduces the defect this card exists to catch: a mechanism
// reporting green about a set it never learned.
// ---------------------------------------------------------------------------

func TestFold_EveryValueFoldsAsTheTableSays(t *testing.T) {
	rows := stateTableRowCells(t)
	if len(rows) != stateTableRowCount {
		t.Fatalf("read %d rows from %s, want %d — a partial read would check a partial table and pass", len(rows), specTablePath, stateTableRowCount)
	}

	// Build value → declared fold from the table itself. Rows 3 and 5 both
	// classify OK, so the map is 7 values over 8 rows; a row disagreeing with
	// an earlier row on the same value is itself a defect and is reported.
	declared := map[Classification]Surface{}
	for num, cells := range rows {
		value := Classification(tableCell(cells[2]))
		fold := Surface(tableCell(cells[3]))
		if value == "" || fold == "" {
			t.Fatalf("row %s has an empty classification (%q) or fold (%q); the table is not being read", num, value, fold)
		}
		if prior, seen := declared[value]; seen && prior != fold {
			t.Errorf("row %s folds %q to %q while an earlier row folds it to %q; one value cannot have two folds", num, value, fold, prior)
		}
		declared[value] = fold
	}
	if len(declared) != 7 {
		t.Fatalf("the table declares %d distinct classifications, want 7 (8 rows, rows 3 and 5 both `OK`)", len(declared))
	}

	// (a) — equality over all seven, value against column.
	for value, want := range declared {
		if got := value.Fold(); got != want {
			t.Errorf("(a) %s folds to %q, and the table's `Surface fold` column says %q", value, got, want)
		}
	}
	// Anti-vacuity: the vocabulary in code and the vocabulary in the table are
	// the same set. A value in code that no row declares would be skipped by
	// the loop above and never compared.
	for _, v := range Classifications() {
		if _, ok := declared[v]; !ok {
			t.Errorf("(a) value %q exists in code and is declared by no row; its fold is unchecked", v)
		}
	}

	// (b) — no classification folds to `fail`.
	for _, v := range Classifications() {
		if v.Fold() == SurfaceFail {
			t.Errorf("(b) %s folds to `fail`, which promotes a routine non-blocking sweep to a failing exit status", v)
		}
	}

	// (c) — the boundary between MEANINGLESS and INCOMPLETE.
	if got := ClassUnreadable.Fold(); got != SurfaceOK {
		t.Errorf("(c) UNREADABLE folds to %q, want `ok`: no reader exists, so the comparison never applied — the meaningless case, where t326's leniency transfers", got)
	}
	for _, incomplete := range []Classification{ClassUnknown, ClassUnresolved, ClassOrphaned} {
		if got := incomplete.Fold(); got != SurfaceWarn {
			t.Errorf("(c) %s folds to %q, want `warn`: the comparison applied and could not be completed — folding it to `ok` reproduces this card's own subject inside its solution", incomplete, got)
		}
	}
}

// TestFold_OnlyOKIsClean pins the axis on which UNREADABLE is distinguishable
// from OK at all: both fold to `ok` and both imply no action, so CLEANLINESS is
// the only observable difference between them (REQ-GSM-007, §D.7). A consumer
// reading the FOLD instead of the designation would treat UNREADABLE as
// nothing-to-report, which is the seam violation the sibling package's doc
// warns about.
func TestFold_OnlyOKIsClean(t *testing.T) {
	clean := 0
	for _, v := range Classifications() {
		if v.IsClean() {
			clean++
			if v != ClassOK {
				t.Errorf("%s reads clean; exactly one value — OK — denotes nothing to report", v)
			}
		}
	}
	if clean != 1 {
		t.Errorf("%d values read clean, want exactly 1", clean)
	}
	if ClassUnreadable.IsClean() {
		t.Error("UNREADABLE reads clean while folding to `ok`; the two axes are what separate it from OK, and collapsing them makes the pair indistinguishable")
	}
	if ClassUnreadable.Fold() != ClassOK.Fold() {
		t.Skip("UNREADABLE and OK no longer share a fold; the gap this test guards has moved")
	}
}
