// admission_test.go — the card record, WIP admission, the unheld steady
// state, and column dispatchability (SPEC-KANBAN-BOARD-001 REQ-KB-004/009/
// 010/011/012, M2; AC-KB-008/011/012/013/014).
//
// The WIP limit and the deployed coder-session count are two knobs
// (REQ-KB-010): the forward direction varies the WIP knob and watches the
// admitted count track it; the reverse direction is an ABSENCE claim over
// the board's own inputs — the admission path reads no session-count value
// from anywhere. Admission is never gated on a session being free: with WIP
// 2 and one session, the second card enters run UNHELD and that is a legal
// steady state (§A.5).
package kanban

import (
	"testing"
)

// TestCardRecord_RoundTrip — AC-KB-008: the SPEC identifier, column,
// holding session id, and last-transition instant round-trip unchanged; an
// unheld card round-trips with an EMPTY holder rather than a synthesized one.
func TestCardRecord_RoundTrip(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "lead-sess")

	want := []Card{
		{SpecID: "SPEC-RT-001", Column: ColumnRun, Holder: "run-sess-1", LastMovedAt: "2026-08-14T01:00:00Z"},
		{SpecID: "SPEC-RT-002", Column: ColumnReview, Holder: "", LastMovedAt: "2026-08-14T02:00:00Z"},
		{SpecID: "SPEC-RT-003", Column: ColumnBacklog, Holder: "", LastMovedAt: "2026-08-14T03:00:00Z"},
	}
	err := WriteBoardState(root, "lead-sess", func(st *BoardState) error {
		st.Cards = append(st.Cards, want...)
		return nil
	})
	if err != nil {
		t.Fatalf("seed cards: %v", err)
	}

	st, err := LoadBoard(root)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	if len(st.Cards) != len(want) {
		t.Fatalf("card count = %d, want %d", len(st.Cards), len(want))
	}
	for i := range want {
		if st.Cards[i] != want[i] {
			t.Errorf("card[%d] = %+v, want %+v", i, st.Cards[i], want[i])
		}
	}
	for _, i := range []int{1, 2} {
		if st.Cards[i].Holder != "" {
			t.Errorf("unheld card[%d] holder = %q, want empty — never a synthesized holder", i, st.Cards[i].Holder)
		}
	}
}

// TestAdmission_WipKnobVaries — AC-KB-012 forward direction: with every
// other input held fixed, the number of cards admitted to run tracks the WIP
// bound. Table-driven over 1, 2, and 3.
func TestAdmission_WipKnobVaries(t *testing.T) {
	t.Parallel()
	for _, limit := range []int{1, 2, 3} {
		limit := limit
		t.Run(map[int]string{1: "wip1", 2: "wip2", 3: "wip3"}[limit], func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			seedLead(t, root, "lead-sess")
			opts := BoardOptions{RunWIPLimit: limit}

			admitted := 0
			for i := 0; i < limit+2; i++ {
				spec := "SPEC-WIP-" + string(rune('0'+limit)) + "-" + string(rune('A'+i))
				err := TransitionIntoRunOpts(root, "lead-sess", spec, opts)
				if err == nil {
					admitted++
					continue
				}
				if !IsWipLimitExceeded(err) {
					t.Fatalf("transition %d: unexpected error %v", i, err)
				}
			}
			if admitted != limit {
				t.Fatalf("admitted = %d, want %d — the admitted count must track the WIP bound", admitted, limit)
			}
		})
	}
}

// TestAdmission_WipDefaultsToTwo — the production default is 2: with
// default options exactly two cards enter run.
func TestAdmission_WipDefaultsToTwo(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "lead-sess")

	if err := TransitionIntoRun(root, "lead-sess", "SPEC-DEF-1"); err != nil {
		t.Fatalf("first default admission: %v", err)
	}
	if err := TransitionIntoRun(root, "lead-sess", "SPEC-DEF-2"); err != nil {
		t.Fatalf("second default admission: %v", err)
	}
	if err := TransitionIntoRun(root, "lead-sess", "SPEC-DEF-3"); !IsWipLimitExceeded(err) {
		t.Fatalf("third default admission: err = %v, want WIP refusal (default bound is 2)", err)
	}
}

// TestAdmission_UnheldInRunIsLegalSteadyState — AC-KB-013: with one card
// already held in run and NO session free, a second card is admitted and
// recorded UNHELD; later reads report a valid steady state, not an error or
// a stall. Positive control: a THIRD admission is refused by the WIP limit —
// bounded by WIP, never by session availability.
func TestAdmission_UnheldInRunIsLegalSteadyState(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "lead-sess")

	// One card already held in run.
	if err := TransitionIntoRun(root, "lead-sess", "SPEC-UNH-1"); err != nil {
		t.Fatalf("first admission: %v", err)
	}
	// Assign its holder explicitly (the admission path itself never consults
	// session availability — REQ-KB-011).
	err := WriteBoardState(root, "lead-sess", func(st *BoardState) error {
		for i := range st.Cards {
			if st.Cards[i].SpecID == "SPEC-UNH-1" {
				st.Cards[i].Holder = "run-sess-only"
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("assign holder: %v", err)
	}

	// No session is free; the second admission still succeeds and records
	// an EMPTY holder — unheld, not refused, not synthesized.
	if err := TransitionIntoRun(root, "lead-sess", "SPEC-UNH-2"); err != nil {
		t.Fatalf("second admission with no session free refused: %v — admission is gated on WIP, not on session availability", err)
	}
	st, err := LoadBoard(root)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	var second *Card
	for i := range st.Cards {
		if st.Cards[i].SpecID == "SPEC-UNH-2" {
			second = &st.Cards[i]
		}
	}
	if second == nil {
		t.Fatal("second card not on the board")
	}
	if second.Holder != "" {
		t.Fatalf("unheld card holder = %q, want empty", second.Holder)
	}
	if second.Column != ColumnRun {
		t.Fatalf("second card column = %q, want run", second.Column)
	}

	// Positive control: the THIRD admission is refused — by WIP, proving the
	// bound rather than session availability does the limiting.
	if err := TransitionIntoRun(root, "lead-sess", "SPEC-UNH-3"); !IsWipLimitExceeded(err) {
		t.Fatalf("third admission: err = %v, want WIP refusal", err)
	}
}

// TestAdmission_BacklogDoneNotDispatchable — AC-KB-014: cards in backlog and
// done are not dispatchable (neither column has an owning session); a card
// in plan IS dispatchable — the refusal is column-conditional.
func TestAdmission_BacklogDoneNotDispatchable(t *testing.T) {
	t.Parallel()
	cases := []struct {
		col      Column
		status   string
		wantDisp bool
	}{
		{ColumnBacklog, StatusDraft, false},
		{ColumnBacklog, "", false}, // no spec.md yet — still not dispatchable
		{ColumnDone, StatusCompleted, false},
		{ColumnPlan, StatusDraft, true},
		{ColumnPlan, "", true}, // planned artifacts appear mid-column
	}
	for _, tc := range cases {
		card := Card{SpecID: "SPEC-DISP-001", Column: tc.col, LastMovedAt: "t0"}
		specPresent := tc.status != ""
		cs := &CardStatus{Status: tc.status, Source: StatusSourcePrimary, SpecFilePresent: specPresent}
		view := ReconcileCard(card, cs)
		if view.Dispatchable != tc.wantDisp {
			t.Errorf("(%s, %q): Dispatchable = %v, want %v", tc.col, tc.status, view.Dispatchable, tc.wantDisp)
		}
	}
}
