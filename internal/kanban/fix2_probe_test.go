// fix2_probe_test.go — RED probe for re-review ITEM 1: WriteBoardState is
// the exported mutation anchor (its @MX:ANCHOR names fan_in >= 3 including
// any future board mutator), so the SpecID guard must hold at the anchor,
// not only at one caller. A mutation that injects a traversal-shaped SpecID
// into the state must be refused BEFORE the atomic write — the board must
// never persist an identity it can never subsequently read.
package kanban

import (
	"os"
	"testing"
)

// TestWriteBoardState_RejectsTraversalSpecIDFromMutate — ITEM 1, reproducing
// the reviewer's measurement: a lead-role caller whose mutate closure
// appends a card with SpecID "../../../escape" must be REFUSED and the board
// left without the poisoned card.
func TestWriteBoardState_RejectsTraversalSpecIDFromMutate(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "s1")

	err := WriteBoardState(root, "s1", func(st *BoardState) error {
		st.Cards = append(st.Cards, Card{SpecID: "../../../escape", Column: ColumnRun})
		return nil
	})
	if err == nil {
		t.Fatal("WriteBoardState(mutate injects traversal SpecID) err = nil, want validation refusal")
	}

	// The poisoned identity must not persist.
	if _, statErr := os.Stat(BoardPath(root)); statErr == nil {
		st, lerr := LoadBoard(root)
		if lerr != nil {
			t.Fatalf("LoadBoard after refusal: %v", lerr)
		}
		for _, c := range st.Cards {
			if c.SpecID == "../../../escape" {
				t.Fatal("traversal SpecID persisted to board.json — the board accepted an identity it can never read")
			}
		}
	}

	// The sweep is post-mutate and shape-conditional: a canonical id from the
	// same closure path still lands (positive control).
	err = WriteBoardState(root, "s1", func(st *BoardState) error {
		st.Cards = append(st.Cards, Card{SpecID: "SPEC-OK-001", Column: ColumnPlan})
		return nil
	})
	if err != nil {
		t.Fatalf("canonical id refused: %v — the guard is conditional on shape", err)
	}
	st, err := LoadBoard(root)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	if len(st.Cards) != 1 || st.Cards[0].SpecID != "SPEC-OK-001" {
		t.Fatalf("canonical card did not land: %+v", st.Cards)
	}
}
