//go:build afred

package ledger

import (
	"errors"
	"testing"
	"time"
)

func TestNotesAreAppendedInOrderAndReadFromASeq(t *testing.T) {
	s := openTemp(t)
	for i, n := range []Note{
		{Kind: NoteDecision, CardID: "a", Text: "first"},
		{Kind: NoteStalled, Text: "no card"},
		{Kind: NoteLanded, CardID: "b", Text: "third"},
		{Kind: NoteGateWait, CardID: "a", Text: "fourth"},
	} {
		if err := s.AddNote(bg, n); err != nil {
			t.Fatalf("note %d: %v", i, err)
		}
	}
	all, err := s.NotesSince(bg, 0)
	if err != nil || len(all) != 4 {
		t.Fatalf("%+v %v", all, err)
	}
	for i := 1; i < len(all); i++ {
		if all[i].Seq <= all[i-1].Seq {
			t.Errorf("seq order: %+v", all)
		}
	}
	if all[0].Kind != NoteDecision || all[0].CardID != "a" || all[0].Text != "first" || time.Since(all[0].At) > time.Minute {
		t.Errorf("first: %+v", all[0])
	}
	if all[1].CardID != "" {
		t.Errorf("a note without a card: %+v", all[1])
	}
	tail, _ := s.NotesSince(bg, all[1].Seq)
	if len(tail) != 2 || tail[0].Text != "third" {
		t.Errorf("since %d: %+v", all[1].Seq, tail)
	}
	if none, _ := s.NotesSince(bg, all[3].Seq); len(none) != 0 {
		t.Errorf("past the end: %+v", none)
	}
	r, err := s.Notes(bg, all[0].Seq)
	if err != nil || len(r) != 3 {
		t.Errorf("Reader.Notes: %+v %v", r, err)
	}
}

func TestANoteNeedsAKnownKindAndSomeText(t *testing.T) {
	s := openTemp(t)
	for name, n := range map[string]Note{
		"kind":  {Kind: "gossip", Text: "x"},
		"empty": {Kind: NoteDecision, Text: ""},
		"blank": {Kind: NoteDecision, Text: "  \n"},
	} {
		if err := s.AddNote(bg, n); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if got, _ := s.NotesSince(bg, 0); len(got) != 0 {
		t.Errorf("a refused note was stored: %+v", got)
	}
}
