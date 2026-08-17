// reconcile_test.go — the (column, status) compatibility table and the
// reconciled card view (SPEC-KANBAN-BOARD-001 REQ-KB-008, M2;
// AC-KB-009/010/021).
//
// The table is exercised over EVERY pairing — the legal rows and the illegal
// ones alike — because an implementation that accepts everything passes a
// legal-rows-only test perfectly (AP-10). Four pairings are called out
// because the v0.1.0 table decided each wrongly or not at all:
// (sync, completed) legal; (backlog, planned) and (plan, planned) legal;
// (run, planned) illegal. The terminals (archived / superseded / rejected)
// produce NO CARD AT ALL — card absence, never an inconsistency report.
package kanban

import (
	"os"
	"testing"
)

// TestCompatibilityTable_EveryPairing — AC-KB-009: table-driven over all six
// columns × eight statuses. Legal pairs are accepted; illegal pairs mark the
// card inconsistent, not dispatchable, and repair NEITHER value.
func TestCompatibilityTable_EveryPairing(t *testing.T) {
	t.Parallel()

	statuses := []string{
		StatusDraft, StatusPlanned, StatusInProgress, StatusImplemented,
		StatusCompleted, StatusSuperseded, StatusArchived, StatusRejected,
	}
	// The table of spec.md §A.4 as revised at v0.2.0. Terminals appear in no
	// row: no card is created for them, which the eligibility test covers —
	// here they pair with NO column legally.
	legal := map[string]bool{
		"backlog|" + StatusDraft:     true,
		"backlog|" + StatusPlanned:   true,
		"plan|" + StatusDraft:        true,
		"plan|" + StatusPlanned:      true,
		"run|" + StatusInProgress:    true,
		"review|" + StatusInProgress: true,
		"sync|" + StatusInProgress:   true,
		"sync|" + StatusImplemented:  true,
		"sync|" + StatusCompleted:    true,
		"done|" + StatusCompleted:    true,
	}

	for _, col := range Columns() {
		for _, status := range statuses {
			col, status := col, status
			t.Run(string(col)+"/"+status, func(t *testing.T) {
				t.Parallel()
				card := Card{SpecID: "SPEC-TBL-001", Column: col, Holder: "sess-1", LastMovedAt: "t0"}
				cs := &CardStatus{Status: status, Source: StatusSourcePrimary, SpecFilePresent: true}

				view := ReconcileCard(card, cs)
				wantLegal := legal[string(col)+"|"+status]
				if view.Inconsistent == wantLegal {
					t.Fatalf("pairing (%s, %s): Inconsistent=%v, want %v — every pairing is decided, in both directions",
						col, status, view.Inconsistent, !wantLegal)
				}
				// Legal-and-dispatchable are distinct dimensions: backlog and
				// done hold legal pairings yet have no owning session
				// (REQ-KB-012), so the dispatch verdict is checked only for
				// the working columns.
				if wantLegal && col.HasOwningSession() && !view.Dispatchable {
					t.Errorf("legal pairing (%s, %s) not dispatchable", col, status)
				}
				if !wantLegal {
					if view.Dispatchable {
						t.Errorf("illegal pairing (%s, %s) dispatchable", col, status)
					}
					if view.Details == "" {
						t.Errorf("illegal pairing (%s, %s) surfaced nothing — both values must be surfaced", col, status)
					}
				}
			})
		}
	}
}

// TestCompatibilityTable_IllegalPairRepairsNothing — AP-4: an illegal
// pairing repairs neither the column nor the status. Reconciliation is a
// read-only operation over the recorded card and the read status, so the
// board file and the source document are byte-unchanged after it fires.
func TestCompatibilityTable_IllegalPairRepairsNothing(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "lead-sess")
	writeBoardRaw(t, BoardPath(root), &BoardState{Cards: []Card{
		{SpecID: "SPEC-ILL-001", Column: ColumnDone, Holder: "s1", LastMovedAt: "t0"},
	}})
	beforeBoard := readBoardBytes(t, root)

	specPath := mockSpecPath(t, "SPEC-ILL-001", StatusDraft)
	beforeSpec := readFileBytes(t, specPath)

	card := Card{SpecID: "SPEC-ILL-001", Column: ColumnDone, Holder: "s1", LastMovedAt: "t0"}
	cs := &CardStatus{Status: StatusDraft, Source: StatusSourcePrimary, SpecFilePresent: true}
	view := ReconcileCard(card, cs)

	if !view.Inconsistent {
		t.Fatal("(done, draft) not marked inconsistent — an outside-the-table pairing")
	}
	if view.Dispatchable {
		t.Fatal("inconsistent card dispatchable")
	}
	if string(beforeBoard) != string(readBoardBytes(t, root)) {
		t.Fatal("board record changed on reconciliation — repair neither")
	}
	if string(beforeSpec) != string(readFileBytes(t, specPath)) {
		t.Fatal("status source changed on reconciliation — repair neither")
	}
}

// TestCompatibilityTable_NoSpecMDRows — the no-spec.md rows: legal in
// backlog and plan (a card admitted before planning has no frontmatter),
// illegal everywhere else.
func TestCompatibilityTable_NoSpecMDRows(t *testing.T) {
	t.Parallel()
	for _, col := range Columns() {
		card := Card{SpecID: "SPEC-NOSPEC-1", Column: col}
		cs := &CardStatus{Status: "", Source: StatusSourcePrimary, SpecFilePresent: false}
		view := ReconcileCard(card, cs)
		switch col {
		case ColumnBacklog, ColumnPlan:
			if view.Inconsistent {
				t.Errorf("(%s, no spec.md) marked inconsistent; a card admitted before planning has no frontmatter", col)
			}
		default:
			if !view.Inconsistent {
				t.Errorf("(%s, no spec.md) accepted; only backlog and plan admit a card with no spec.md", col)
			}
		}
	}
}

// TestReconcileCard_CollisionsAreAmbiguousNotResolved — AC-KB-010: `draft`
// does not by itself decide between backlog and plan; `in-progress` does not
// decide between run and review. The reconciliation REPORTS the ambiguity of
// its status input rather than resolving it, and the recorded column stands.
func TestReconcileCard_CollisionsAreAmbiguousNotResolved(t *testing.T) {
	t.Parallel()
	cases := []struct {
		status string
		cols   []Column
	}{
		{StatusDraft, []Column{ColumnBacklog, ColumnPlan}},
		{StatusPlanned, []Column{ColumnBacklog, ColumnPlan}},
		{StatusInProgress, []Column{ColumnRun, ColumnSync}},
	}
	for _, tc := range cases {
		for _, col := range tc.cols {
			card := Card{SpecID: "SPEC-AMB-001", Column: col, Holder: "s1", LastMovedAt: "t0"}
			cs := &CardStatus{Status: tc.status, Source: StatusSourcePrimary, SpecFilePresent: true}
			view := ReconcileCard(card, cs)
			if view.Inconsistent {
				t.Fatalf("(%s, %s) must be legal — both members of the collision pair accept it", col, tc.status)
			}
			// The dispatch verdict additionally requires an owning session;
			// backlog holds this pairing legally but dispatches nothing
			// (REQ-KB-012), so the verdict is asserted where it applies.
			if col.HasOwningSession() && !view.Dispatchable {
				t.Fatalf("(%s, %s) legal yet not dispatchable", col, tc.status)
			}
			if !view.ColumnAmbiguous {
				t.Fatalf("(%s, %s) not reported ambiguous — %q does not by itself decide the column pair %v",
					col, tc.status, tc.status, tc.cols)
			}
		}
	}
	// The non-colliding statuses are not reported ambiguous.
	for _, status := range []string{StatusImplemented, StatusCompleted} {
		card := Card{SpecID: "SPEC-AMB-002", Column: ColumnSync, LastMovedAt: "t0"}
		cs := &CardStatus{Status: status, Source: StatusSourcePrimary, SpecFilePresent: true}
		if view := ReconcileCard(card, cs); view.ColumnAmbiguous {
			t.Fatalf("(%s, %s) reported ambiguous; only draft/planned and in-progress collide", ColumnSync, status)
		}
	}
}

// TestShouldCreateCard_TerminalsProduceNoCard — AC-KB-021's second half: a
// SPEC whose status is archived, superseded, or rejected produces NO card in
// any column. The test observes card ABSENCE rather than an inconsistency
// report — a different outcome, not a different message.
func TestShouldCreateCard_TerminalsProduceNoCard(t *testing.T) {
	t.Parallel()
	for _, status := range []string{StatusArchived, StatusSuperseded, StatusRejected} {
		cs := &CardStatus{Status: status, Source: StatusSourcePrimary, SpecFilePresent: true}
		if ShouldCreateCard(cs) {
			t.Errorf("status %q created a card — out-of-lifecycle terminals are not work in flight", status)
		}
	}
	for _, status := range []string{StatusDraft, StatusPlanned, StatusInProgress, StatusImplemented, StatusCompleted} {
		cs := &CardStatus{Status: status, Source: StatusSourcePrimary, SpecFilePresent: true}
		if !ShouldCreateCard(cs) {
			t.Errorf("status %q produced no card", status)
		}
	}
	// No spec.md at all is a backlog admission, not a terminal.
	if !ShouldCreateCard(&CardStatus{Status: "", Source: StatusSourcePrimary, SpecFilePresent: false}) {
		t.Error("a card with no spec.md produced no card — the backlog admission case")
	}
}

// TestReconcileCard_PlannedAdmittedOnlyInBacklogAndPlan — AC-KB-021's first
// half: `planned` admitted in backlog and plan and rejected everywhere
// else — notably (run, planned) is illegal, since planned asserts work has
// not started.
func TestReconcileCard_PlannedAdmittedOnlyInBacklogAndPlan(t *testing.T) {
	t.Parallel()
	for _, col := range Columns() {
		card := Card{SpecID: "SPEC-PLN-001", Column: col}
		cs := &CardStatus{Status: StatusPlanned, Source: StatusSourcePrimary, SpecFilePresent: true}
		view := ReconcileCard(card, cs)
		switch col {
		case ColumnBacklog, ColumnPlan:
			if view.Inconsistent {
				t.Errorf("(planned, %s) refused — planned collides here exactly as draft does", col)
			}
		default:
			if !view.Inconsistent {
				t.Errorf("(planned, %s) accepted — planned is admitted nowhere else: it asserts work has not started", col)
			}
		}
	}
}

// TestReconcileCard_InvalidRecordedColumn — a state document carrying a
// column outside the closed set still loads (the file is readable), and the
// reconciliation marks that card inconsistent rather than silently treating
// it as any column.
func TestReconcileCard_InvalidRecordedColumn(t *testing.T) {
	t.Parallel()
	card := Card{SpecID: "SPEC-BADCOL-1", Column: Column("escalated")}
	cs := &CardStatus{Status: StatusDraft, Source: StatusSourcePrimary, SpecFilePresent: true}
	view := ReconcileCard(card, cs)
	if !view.Inconsistent {
		t.Fatal("card with a column outside the closed set reconciled as consistent")
	}
	if view.Dispatchable {
		t.Fatal("card with an invalid column dispatchable")
	}
}

// mockSpecPath writes a minimal spec.md under a temp root and returns its path.
func mockSpecPath(t *testing.T, specID, status string) string {
	t.Helper()
	dir := t.TempDir()
	path := dir + string(os.PathSeparator) + specID + "-spec.md"
	body := "---\nstatus: " + status + "\n---\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write mock spec: %v", err)
	}
	return path
}

// readFileBytes reads a file for byte-equality checks.
func readFileBytes(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return raw
}
