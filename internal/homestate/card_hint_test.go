package homestate

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// RecordAssignedHint (card t1711) sets the order hint and appends a dispatch
// note on an assigned card, without moving its state, owner, or lease.

func strPtr(s string) *string { return &s }

func TestRecordAssignedHintSetsOrderAndNote(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	frPlace(t, db, Card{RunID: frRun, CardID: "ah1", State: CardAssigned, OwnerLabel: "lane-1", Version: 2})
	card, err := db.RecordAssignedHint(ctx, frRun, "ah1", AssignedHint{After: strPtr("ah0"), Note: strPtr(" after ah0 per leader ")}, "assign", frNow)
	if err != nil {
		t.Fatalf("record hint: %v", err)
	}
	if card.HintAfter != "ah0" || card.State != CardAssigned || card.OwnerLabel != "lane-1" || card.Version != 3 {
		t.Fatalf("card = %+v, want assigned / lane-1 / after ah0 / v3", card)
	}
	notes, err := db.CardNotes(ctx, frRun, "ah1")
	if err != nil || len(notes) != 1 || notes[0] != "after ah0 per leader" {
		t.Fatalf("notes = %q err=%v, want the trimmed note", notes, err)
	}
	// A note alone leaves the row (and its version) alone.
	same, err := db.RecordAssignedHint(ctx, frRun, "ah1", AssignedHint{Note: strPtr("second")}, "assign", frNow)
	if err != nil || same.Version != 3 {
		t.Fatalf("note-only write: version %d err=%v, want v3 unchanged", same.Version, err)
	}
	if notes, _ = db.CardNotes(ctx, frRun, "ah1"); len(notes) != 2 || notes[1] != "second" {
		t.Fatalf("notes = %q, want two in order", notes)
	}
	// An empty order clears.
	cleared, err := db.RecordAssignedHint(ctx, frRun, "ah1", AssignedHint{After: strPtr("")}, "assign", frNow)
	if err != nil || cleared.HintAfter != "" || cleared.Version != 4 {
		t.Fatalf("clear: %+v err=%v, want empty after at v4", cleared, err)
	}
	// Another card's notes stay apart.
	if notes, _ = db.CardNotes(ctx, frRun, "ah-other"); len(notes) != 0 {
		t.Fatalf("a different card read notes %q", notes)
	}
}

func TestRecordAssignedHintRefusals(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	for _, c := range []Card{
		{RunID: frRun, CardID: "ah1", State: CardAssigned, OwnerLabel: "lane-1", Version: 2},
		{RunID: frRun, CardID: "ah2", State: CardAssigned, OwnerLabel: "lane-1", Version: 2, HintAfter: "ah1"},
		{RunID: frRun, CardID: "ah3", State: CardLeased, OwnerLabel: "lane-1", LeaseHolder: "lane-1", Version: 2},
		{RunID: frRun, CardID: "ah4", State: CardPicked, Version: 1},
	} {
		frPlace(t, db, c)
	}
	long := strings.Repeat("x", maxCardNoteBytes+1)
	cases := []struct {
		name string
		card string
		hint AssignedHint
		want error
	}{
		{"self", "ah1", AssignedHint{After: strPtr("ah1")}, ErrInvalidCardInput},
		{"cycle", "ah1", AssignedHint{After: strPtr("ah2")}, ErrInvalidCardInput},
		{"not a card id", "ah1", AssignedHint{After: strPtr("no way")}, ErrInvalidCardInput},
		{"leased", "ah3", AssignedHint{After: strPtr("ah1")}, ErrIllegalTransition},
		{"picked", "ah4", AssignedHint{Note: strPtr("n")}, ErrIllegalTransition},
		{"missing", "ah9", AssignedHint{After: strPtr("ah1")}, ErrCardNotFound},
		{"empty request", "ah1", AssignedHint{}, ErrInvalidCardInput},
		{"blank note", "ah1", AssignedHint{Note: strPtr("  ")}, ErrInvalidCardInput},
		{"multi-line note", "ah1", AssignedHint{Note: strPtr("a\nb")}, ErrInvalidCardInput},
		{"long note", "ah1", AssignedHint{Note: &long}, ErrInvalidCardInput},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := db.RecordAssignedHint(ctx, frRun, c.card, c.hint, "assign", frNow); !errors.Is(err, c.want) {
				t.Fatalf("err = %v, want %v", err, c.want)
			}
		})
	}
	if got, _ := db.LoadCard(ctx, frRun, "ah1"); got.HintAfter != "" || got.Version != 2 {
		t.Fatalf("ah1 moved on refused writes: %+v", got)
	}
}
