package cli

// card t1711 — the dispatch order of an ALREADY-ASSIGNED card.
//
// A lane held t1702 (serial) and an assigned t1642 that was meant to start
// after it. The intended order lived only in a peer message: `assign --after`
// edited picked rows only, so the factory record could not carry the order and
// the lane's bare `factory next` leased t1642 first. These tests pin the two
// halves of the cure: `assign` may set the order hint (and a dispatch note) on
// an assigned row, and the unnominated and nominated leases honour that hint.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// fahAssignedPair places t1 (the predecessor, queue-held so no arm can take
// it) and t2 (the successor, assigned to lane-1) in the record.
func fahAssignedPair(t *testing.T) string {
	t.Helper()
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStateHold, factory.BacklogStatePicked)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Version: 2},
		homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-1", Version: 2},
	)
	sdRegisterLane(t, root, "lane-1")
	sdClearLaneEnv(t)
	t.Chdir(root)
	return root
}

// fahNotes reads the dispatch notes recorded for cardID straight from the
// record's own event log, oldest first.
func fahNotes(t *testing.T, root, cardID string) []string {
	t.Helper()
	db := fcOpen(t, root)
	rows, err := db.DB.Query(`SELECT payload_json FROM events WHERE kind='card.note' ORDER BY seq`)
	if err != nil {
		t.Fatalf("read note events: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var notes []string
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var p struct {
			CardID string `json:"card_id"`
			Note   string `json:"note"`
		}
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			t.Fatal(err)
		}
		if p.CardID == cardID {
			notes = append(notes, p.Note)
		}
	}
	return notes
}

// assign --after on an assigned card records the order and changes nothing
// else: state, owner and lease stay, the version moves once.
func TestFahAssignAfterOnAssignedCard(t *testing.T) {
	root := fahAssignedPair(t)
	before := fcCard(t, root, "t2")

	if _, _, err := runFactory(t, "assign", "t2", "--after", "t1", "--run", fcRun); err != nil {
		t.Fatalf("assign t2 --after t1 on an assigned card: %v", err)
	}
	got := fcCard(t, root, "t2")
	if got.HintAfter != "t1" || got.State != homestate.CardAssigned || got.OwnerLabel != "lane-1" {
		t.Fatalf("t2 = %s owner=%q after=%q, want assigned / lane-1 / t1", got.State, got.OwnerLabel, got.HintAfter)
	}
	if got.Version != before.Version+1 {
		t.Fatalf("t2 version = %d, want %d (one record write)", got.Version, before.Version+1)
	}

	if _, _, err := runFactory(t, "assign", "t2", "--after", "", "--run", fcRun); err != nil {
		t.Fatalf("assign t2 --after \"\" (clear): %v", err)
	}
	if c := fcCard(t, root, "t2"); c.HintAfter != "" {
		t.Fatalf("t2 after = %q after the clear, want empty", c.HintAfter)
	}
}

// The dispatch note is persisted in the record, alone or beside the order.
func TestFahAssignNoteOnAssignedCard(t *testing.T) {
	root := fahAssignedPair(t)
	if _, _, err := runFactory(t, "assign", "t2", "--note", "start only after t1 lands", "--run", fcRun); err != nil {
		t.Fatalf("assign t2 --note: %v", err)
	}
	if _, _, err := runFactory(t, "assign", "t2", "--after", "t1", "--note", "order per leader 10-19", "--run", fcRun); err != nil {
		t.Fatalf("assign t2 --after --note: %v", err)
	}
	notes := fahNotes(t, root, "t2")
	if len(notes) != 2 || notes[0] != "start only after t1 lands" || notes[1] != "order per leader 10-19" {
		t.Fatalf("t2 notes = %q, want both notes in order", notes)
	}
	if c := fcCard(t, root, "t2"); c.HintAfter != "t1" {
		t.Fatalf("t2 after = %q, want t1", c.HintAfter)
	}
}

// The hint edit stays on the leader path, stays inside the assigned state, and
// refuses an order that would wait on itself.
func TestFahAssignHintOnAssignedCardGuards(t *testing.T) {
	root := fahAssignedPair(t)

	if _, _, err := runFactory(t, "assign", "t2", "--after", "t2", "--run", fcRun); err == nil {
		t.Fatal("assign t2 --after t2 succeeded — a card cannot wait on itself")
	}
	if _, _, err := runFactory(t, "assign", "t2", "--after", "t1", "--run", fcRun); err != nil {
		t.Fatalf("seed t2 after t1: %v", err)
	}
	if _, _, err := runFactory(t, "assign", "t1", "--after", "t2", "--run", fcRun); err == nil {
		t.Fatal("assign t1 --after t2 succeeded — it closes the cycle t1 -> t2 -> t1")
	}
	if c := fcCard(t, root, "t1"); c.HintAfter != "" {
		t.Fatalf("t1 after = %q after a refused cycle, want empty", c.HintAfter)
	}

	// A lane session never edits the order (leader path, like the re-bind).
	sdLaneEnv(t, "lane-1", "")
	if _, _, err := runFactory(t, "assign", "t2", "--after", "", "--run", fcRun); err == nil || !strings.Contains(err.Error(), "refused") {
		t.Fatalf("lane-session assign --after: err = %v, want a refusal", err)
	}
	sdClearLaneEnv(t)

	// Past assigned the row keeps its order: the lease took the slot.
	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`UPDATE cards SET state=?,lease_holder='lane-1' WHERE card_id='t2'`, homestate.CardLeased); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runFactory(t, "assign", "t2", "--after", "", "--run", fcRun); err == nil {
		t.Fatal("assign --after on a leased card succeeded — the order only edits an assigned row")
	}
}

// The incident shape end to end: the order is recorded on the assigned row,
// and the lane's bare `next` waits instead of leasing it; once the
// predecessor lands the same call leases it.
func TestFahNextWaitsForRecordedPredecessorOfAssignedCard(t *testing.T) {
	root := fahAssignedPair(t)
	if _, _, err := runFactory(t, "assign", "t2", "--after", "t1", "--run", fcRun); err != nil {
		t.Fatalf("record the order: %v", err)
	}
	if got := fbLeasedCard(t, root, "lane-1"); got != "" {
		t.Fatalf("bare next leased %q while its recorded predecessor t1 has not landed", got)
	}
	if c := fcCard(t, root, "t2"); c.State != homestate.CardAssigned {
		t.Fatalf("t2 = %s after the waiting next, want still assigned", c.State)
	}

	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`UPDATE cards SET state=? WHERE card_id='t1'`, homestate.CardMergedLocal); err != nil {
		t.Fatal(err)
	}
	if got := fbLeasedCard(t, root, "lane-1"); got != "t2" {
		t.Fatalf("bare next leased %q after t1 reached merged-local, want t2", got)
	}
}

// A github-flow predecessor (merged-pr) releases the successor the same way.
func TestFahNextReleasesAssignedCardOnMergedPRPredecessor(t *testing.T) {
	root := fahAssignedPair(t)
	if _, _, err := runFactory(t, "assign", "t2", "--after", "t1", "--run", fcRun); err != nil {
		t.Fatalf("record the order: %v", err)
	}
	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`UPDATE cards SET state=? WHERE card_id='t1'`, homestate.CardMergedPR); err != nil {
		t.Fatal(err)
	}
	if got := fbLeasedCard(t, root, "lane-1"); got != "t2" {
		t.Fatalf("bare next leased %q after t1 reached merged-pr, want t2", got)
	}
}

// The nominated door reads the same order: `next --card` refuses the assigned
// successor while the predecessor is unlanded.
func TestFahNextCardNominationRefusesUnlandedPredecessor(t *testing.T) {
	root := fahAssignedPair(t)
	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`UPDATE cards SET hint_after='t1' WHERE card_id='t2'`); err != nil {
		t.Fatal(err)
	}
	sdLaneEnv(t, "lane-1", "")
	if _, _, err := runFactory(t, "next", "--card", "t2", "--run", fcRun); err == nil {
		t.Fatal("next --card t2 leased an assigned card whose recorded predecessor t1 has not landed")
	}
	if c := fcCard(t, root, "t2"); c.State != homestate.CardAssigned {
		t.Fatalf("t2 = %s after the refused nomination, want assigned", c.State)
	}
}
