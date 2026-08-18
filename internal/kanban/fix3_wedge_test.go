// fix3_wedge_test.go — RED probe for rereview2's single finding: the
// empty-spec-id refusal added beyond AC-KB-022's shape-conditional conjunct
// WEDGES a board that loads fine. A board.json carrying
// {"spec_id":"","column":"backlog"} is readable (LoadBoard ok — the file
// parses), but the sweep refused EVERY mutation, including one that would
// repair the bad card, and RecoverBoard correctly reports "recovered" for a
// readable board — so the operator sees a successful recovery beside a board
// refusing every write, with no exit. The fix removes the empty-id refusal
// (operator decision, minimal): a board carrying an empty spec id must not
// block an UNRELATED mutation; the traversal sweep stays exactly as is.
package kanban

import (
	"os"
	"strings"
	"testing"
)

// seedBoardRaw writes a raw board.json body beneath root.
func seedBoardRaw(t *testing.T, root, body string) {
	t.Helper()
	if err := os.MkdirAll(BoardDir(root), 0o755); err != nil {
		t.Fatalf("mkdir board dir: %v", err)
	}
	if err := os.WriteFile(BoardPath(root), []byte(body), 0o600); err != nil {
		t.Fatalf("seed board: %v", err)
	}
}

// TestWriteBoardState_EmptySpecIDDoesNotWedgeBoard — the reviewer's
// measurement, inverted into the wanted behavior: a readable board carrying
// an empty spec_id permits an UNRELATED mutation (and one that repairs the
// bad card). The empty id remains useless — ReadCardStatus still refuses one
// — but it no longer freezes the board.
func TestWriteBoardState_EmptySpecIDDoesNotWedgeBoard(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "leader-sess")
	seedBoardRaw(t, root, `{"cards":[{"spec_id":"","column":"backlog"}]}`)

	// The board loads fine — it is readable, not unknown.
	st, err := LoadBoard(root)
	if err != nil || len(st.Cards) != 1 {
		t.Fatalf("LoadBoard = (%d cards, %v); the seeded board must be readable", len(st.Cards), err)
	}

	// An UNRELATED mutation must succeed.
	err = WriteBoardState(root, "leader-sess", func(st *BoardState) error {
		st.Cards = append(st.Cards, Card{SpecID: "SPEC-UNREL-001", Column: ColumnPlan})
		return nil
	})
	if err != nil {
		t.Fatalf("unrelated mutation refused on a readable board (%v) — one empty id must not wedge the whole board", err)
	}

	// And a mutation that REPAIRS the bad card must succeed too.
	err = WriteBoardState(root, "leader-sess", func(st *BoardState) error {
		for i := range st.Cards {
			if st.Cards[i].SpecID == "" {
				st.Cards[i].SpecID = "SPEC-REPAIRED-001"
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("repairing mutation refused: %v", err)
	}
	st, err = LoadBoard(root)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	if len(st.Cards) != 2 {
		t.Fatalf("cards = %d, want 2", len(st.Cards))
	}
	for _, c := range st.Cards {
		if c.SpecID == "" {
			t.Fatal("repaired card still carries an empty id")
		}
	}
}

// TestWriteBoardState_TraversalRenameInPlaceStillRefused — the traversal
// sweep is untouched by the removal: the reviewer's rename-in-place shape
// (mutating an EXISTING card's SpecID to a traversal value) still refuses
// with the traversal error.
func TestWriteBoardState_TraversalRenameInPlaceStillRefused(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "leader-sess")
	writeBoardRaw(t, BoardPath(root), &BoardState{Cards: []Card{
		{SpecID: "SPEC-OK-100", Column: ColumnBacklog, LastMovedAt: "t0"},
	}})

	err := WriteBoardState(root, "leader-sess", func(st *BoardState) error {
		st.Cards[0].SpecID = "../../../escape"
		return nil
	})
	if err == nil {
		t.Fatal("rename-in-place to a traversal SpecID accepted — the traversal sweep must be intact")
	}
	if !strings.Contains(err.Error(), "..") {
		t.Fatalf("err = %v, want the traversal ('..') validation error", err)
	}

	// The poisoned rename did not persist.
	st, lerr := LoadBoard(root)
	if lerr != nil {
		t.Fatalf("LoadBoard: %v", lerr)
	}
	if st.Cards[0].SpecID != "SPEC-OK-100" {
		t.Fatalf("rename persisted: %q", st.Cards[0].SpecID)
	}
}
