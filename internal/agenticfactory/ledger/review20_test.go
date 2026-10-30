//go:build afred

package ledger

import (
	"database/sql"
	"strings"
	"testing"
)

func stalledNotes(t *testing.T, s *Store, card string) []string {
	t.Helper()
	rows, err := s.db.QueryContext(bg, `SELECT text FROM notes WHERE kind='stalled' AND card_id=? ORDER BY seq`, card)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var x string
		if err := rows.Scan(&x); err != nil {
			t.Fatal(err)
		}
		out = append(out, x)
	}
	return out
}

// (a) One card's trouble costs only that card: it is skipped, noted once per card and
// epoch, and the rest of the transaction commits.
func TestACardThatCannotBeReadIsSkippedAndNotedOnceWhileOthersAreClaimed(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "bad", 9, "A", nil, "internal/b/x.go")
	addCard(t, s, "good", 1, "A", nil, "internal/g/x.go")
	if _, err := s.db.ExecContext(bg, `UPDATE cards SET spec_json='not json' WHERE id='bad'`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		c, err := s.Claim(bg, req("s"+string(rune('1'+i)), 10+i))
		if i == 0 && (err != nil || c.CardID != "good") {
			t.Fatalf("the readable card is claimed past the broken one: %+v %v", c, err)
		}
		if i == 1 && !strings.Contains(errString(err), "no card can be claimed") {
			t.Fatalf("only the broken card is left: %+v %v", c, err)
		}
	}
	if n := stalledNotes(t, s, "bad"); len(n) != 1 || !strings.Contains(n[0], "skipped") {
		t.Errorf("one note for the broken card, however many claims looked at it: %v", n)
	}
}

func TestOneBrokenLeaseDoesNotStopTheReclaimOfTheOthers(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "x", 2, "A", nil, "internal/x/x.go")
	addCard(t, s, "y", 1, "A", nil, "internal/y/x.go")
	for i, id := range []string{"x", "y"} {
		if c, err := s.Claim(bg, req(id, 20+i)); err != nil || c.CardID != id {
			t.Fatalf("%s: %+v %v", id, c, err)
		}
	}
	if _, err := s.db.ExecContext(bg, `UPDATE cards SET spec_json='not json' WHERE id='x'`); err != nil {
		t.Fatal(err)
	}
	got, err := s.Reap(bg, alwaysDead)
	if err != nil || len(got) != 1 || got[0].CardID != "y" {
		t.Fatalf("y is reclaimed although x could not be: %+v %v", got, err)
	}
	if n := stalledNotes(t, s, "x"); len(n) != 1 {
		t.Errorf("x is noted once: %v", n)
	}
	if _, err := s.Reap(bg, alwaysDead); err != nil {
		t.Fatal(err)
	}
	if n := stalledNotes(t, s, "x"); len(n) != 1 {
		t.Errorf("not noted again for the same epoch: %v", n)
	}
}

// (d) An expired lease whose holder is alive is stalled too: the holder lives but does
// not renew.
func TestAnExpiredLeaseOfAnAliveHolderIsNotedOnceAndKept(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 1, "A", nil, "internal/a/x.go")
	if _, err := s.Claim(bg, req("a", 5)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(bg, `UPDATE leases SET expires_at='2000-01-01T00:00:00.000Z'`); err != nil {
		t.Fatal(err)
	}
	alive := func(int, string) Liveness { return LivenessAlive }
	for i := 0; i < 2; i++ {
		if got, err := s.Reap(bg, alive); err != nil || len(got) != 0 {
			t.Fatalf("an alive holder is never reclaimed: %v %v", got, err)
		}
	}
	if n := stalledNotes(t, s, "a"); len(n) != 1 || !strings.Contains(n[0], "alive") {
		t.Errorf("one stalled note saying the holder is alive: %v", n)
	}
	if ls, _ := s.Leases(bg); len(ls) != 1 {
		t.Errorf("the lease stays: %+v", ls)
	}
}

// (e) A card waiting on a CANCELLED dependency can never run: skipped, noted once.
func TestACardWaitingOnACancelledDependencyIsSkippedAndNotedOnce(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "dep", 1, "A", nil, "internal/d/x.go")
	addCard(t, s, "waits", 9, "A", []string{"dep"}, "internal/w/x.go")
	addCard(t, s, "free", 1, "A", nil, "internal/f/x.go")
	if _, err := s.db.ExecContext(bg, `UPDATE cards SET stage='CANCELLED', sub_stage=NULL WHERE id='dep'`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		c, err := s.Claim(bg, req("s"+string(rune('1'+i)), 30+i))
		if i == 0 && (err != nil || c.CardID != "free") {
			t.Fatalf("the card with a cancelled dependency is skipped: %+v %v", c, err)
		}
	}
	n := stalledNotes(t, s, "waits")
	if len(n) != 1 || !strings.Contains(n[0], "cancelled") || !strings.Contains(n[0], "dep") {
		t.Errorf("one note naming the cancelled dependency: %v", n)
	}
}

// (f) A card's first transition row is its adding; rows whose from and to are the same
// position are not edges (claimed again, reclaimed, parked in place).
func TestAddCardRecordsTheFirstTransitionRow(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 1, "A", nil, "internal/a/x.go")
	var (
		from, fromSub sql.NullString
		to, actor     string
		reason        sql.NullString
		epoch         int64
	)
	err := s.db.QueryRowContext(bg, `SELECT from_stage,from_sub,to_stage,epoch,actor,reason FROM transitions WHERE card_id='a' ORDER BY seq LIMIT 1`).Scan(&from, &fromSub, &to, &epoch, &actor, &reason)
	if err != nil {
		t.Fatal(err)
	}
	if from.Valid || to != "READY" || epoch != 0 || actor != "add" || reason.String != "added" {
		t.Errorf("first row: from=%v to=%s epoch=%d actor=%s reason=%q", from, to, epoch, actor, reason.String)
	}
}

// (g) recursive_triggers makes INSERT OR REPLACE fire the delete triggers, so a row
// cannot be replaced away either.
func TestAReplaceCannotDeleteHistory(t *testing.T) {
	s := openTemp(t)
	if _, err := s.db.ExecContext(bg, `INSERT INTO notes(kind,text,at) VALUES('decision','kept','2026-10-30T00:00:00.000Z')`); err != nil {
		t.Fatal(err)
	}
	var on int
	if err := s.db.QueryRowContext(bg, `PRAGMA recursive_triggers`).Scan(&on); err != nil || on != 1 {
		t.Errorf("recursive_triggers = %d %v", on, err)
	}
	_, err := s.db.ExecContext(bg, `INSERT OR REPLACE INTO notes(seq,kind,text,at) VALUES(1,'decision','replaced','2026-10-30T00:00:01.000Z')`)
	if err == nil || !strings.Contains(err.Error(), "never deleted") {
		t.Errorf("a replace of a history row: %v", err)
	}
	var text string
	if err := s.db.QueryRowContext(bg, `SELECT text FROM notes WHERE seq=1`).Scan(&text); err != nil || text != "kept" {
		t.Errorf("the note was replaced: %q %v", text, err)
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
