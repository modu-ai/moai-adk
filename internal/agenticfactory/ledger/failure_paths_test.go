package ledger

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// issue-1948, audit db0a repair, coverage of the failure paths: a ledger whose table is gone, whose
// column is renamed (an older or damaged file), or whose write is refused by the engine must come
// back as a "ledger: ..." error from the code added in the repair, and must leave the rows as they
// were. Each case damages the file after the card is set up and before the call under test.

// damage runs statements straight on the ledger's connection pool.
func damage(t *testing.T, s *Store, stmts ...string) {
	t.Helper()
	for _, q := range stmts {
		if _, err := s.db.Exec(q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
}

func wantLedgerError(t *testing.T, err error, detail string) {
	t.Helper()
	if err == nil {
		t.Fatalf("no error; wanted one naming %q", detail)
	}
	if !strings.HasPrefix(err.Error(), "ledger: ") || !strings.Contains(err.Error(), detail) {
		t.Fatalf("error %q is not a ledger error naming %q", err, detail)
	}
}

func count(t *testing.T, s *Store, q string, args ...any) int {
	t.Helper()
	var n int
	if err := s.db.QueryRow(q, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return n
}

func TestQueueHelpers_ReturnTheEngineErrorWhenTheEvidenceTableIsGone(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	card := cardOf(t, s, "a")
	damage(t, s, `DROP TABLE evidence`)
	calls := map[string]func() error{
		"latestResult":     func() error { _, err := latestResult(bg, s.db, "a", KindGate, "c1"); return err },
		"newestWorkIs":     func() error { _, err := newestWorkIs(bg, s.db, "a", "c1"); return err },
		"noFailingNewest":  func() error { _, err := noFailingNewestRun(bg, s.db, "a", "c1"); return err },
		"measuredInEpoch":  func() error { _, err := measuredInEpoch(bg, s.db, "a", "c1", card.Epoch); return err },
		"landingStands":    func() error { _, err := landingStands(bg, s.db, "a", "c1"); return err },
		"factsTx":          func() error { _, err := factsTx(bg, s.db, card); return err },
		"measuredAfterOth": func() error { _, err := measuredAfterOtherWork(bg, s.db, "a", "c1"); return err },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			wantLedgerError(t, call(), "no such table: evidence")
		})
	}
}

func TestQueueHelpers_ReturnTheEngineErrorWhenTheTransitionsAreUnreadable(t *testing.T) {
	// "dropped": the table is gone. "column": the table is there but from_stage is not, so the
	// lookup of a named offer still answers (it never reads from_stage) and the next query fails.
	for _, damageKind := range []string{"dropped", "column"} {
		t.Run(damageKind, func(t *testing.T) {
			s := openTemp(t)
			c := gated(t, s, "a", "c1")
			if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
				t.Fatal(err)
			}
			backOut(t, s, c.Epoch)
			greenOn(t, s, c.Epoch, "c2", ResultPass)
			detail := "no such table: transitions"
			if damageKind == "dropped" {
				damage(t, s, `DROP TABLE transitions`)
			} else {
				damage(t, s, `ALTER TABLE transitions RENAME COLUMN from_stage TO from_stage_gone`)
				detail = "no such column"
			}
			if damageKind == "dropped" {
				_, _, _, err := namedOfferOf(bg, s.db, "a", "c1")
				wantLedgerError(t, err, detail)
				_, _, _, err = offerOf(bg, s.db, "a", "c1")
				wantLedgerError(t, err, detail)
			}
			// c9 was never offered by name: the fallback lookup is the one that reads from_stage.
			_, _, _, err := offerOf(bg, s.db, "a", "c9")
			wantLedgerError(t, err, detail)
			_, _, err = reworkMark(bg, s.db, "a", "c1")
			wantLedgerError(t, err, detail)
			_, err = measuredAfterOtherWork(bg, s.db, "a", "c1")
			wantLedgerError(t, err, detail)
			_, err = landingStands(bg, s.db, "a", "c1")
			wantLedgerError(t, err, detail)
			if damageKind == "dropped" {
				// The card's newest green is about c2, so the newest-work check has to ask for c1's offer.
				_, err = newestWorkIs(bg, s.db, "a", "c1")
				wantLedgerError(t, err, detail)
			}
		})
	}
}

func TestMeasuredAfterOtherWork_ReturnsTheEngineErrorOfItsEvidenceQueries(t *testing.T) {
	// c1 offered, taken back out of MERGING, then other work (c2) recorded: both evidence queries
	// of the third landing condition run. Renaming evidence.result breaks only the second.
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	backOut(t, s, c.Epoch)
	greenOn(t, s, c.Epoch, "c2", ResultPass)
	damage(t, s, `ALTER TABLE evidence RENAME COLUMN result TO result_gone`)
	_, err := measuredAfterOtherWork(bg, s.db, "a", "c1")
	wantLedgerError(t, err, "no such column")

	// Breaking the commit column breaks the query that looks for the other work.
	s2 := openTemp(t)
	c2 := gated(t, s2, "a", "c1")
	if err := s2.Enqueue(bg, "a", c2.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	backOut(t, s2, c2.Epoch)
	damage(t, s2, `ALTER TABLE evidence RENAME COLUMN subject_sha TO subject_gone`)
	_, err = measuredAfterOtherWork(bg, s2.db, "a", "c1")
	wantLedgerError(t, err, "no such column")
}

func TestFactsTx_ReturnsTheErrorOfAnUnreadableQueueOrEpochColumn(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	card := cardOf(t, s, "a")
	if card.Epoch != c.Epoch {
		t.Fatalf("epoch %d vs %d", card.Epoch, c.Epoch)
	}
	damage(t, s, `DROP TABLE merge_queue`)
	_, err := factsTx(bg, s.db, card)
	wantLedgerError(t, err, "no such table: merge_queue")

	s2 := openTemp(t)
	c2 := gated(t, s2, "a", "c1")
	damage(t, s2, `ALTER TABLE evidence RENAME COLUMN epoch TO epoch_gone`)
	_, err = factsTx(bg, s2.db, cardOf(t, s2, "a"))
	wantLedgerError(t, err, "no such column")
	_ = c2
}

func TestFactsTx_ALandedRowWhoseHistoryIsUnreadableIsAnErrorNotAFinishedCard(t *testing.T) {
	for _, newerGreen := range []bool{false, true} {
		name := "the landed commit is the newest work"
		if newerGreen {
			name = "the newest green is about another commit"
		}
		t.Run(name, func(t *testing.T) {
			s := openTemp(t)
			c := gated(t, s, "a", "c1")
			if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
				t.Fatal(err)
			}
			landOne(t, s, "a", c.Epoch)
			if newerGreen {
				greenOn(t, s, c.Epoch, "c2", ResultPass)
			}
			card := cardOf(t, s, "a")
			damage(t, s, `DROP TABLE transitions`)
			have, err := factsTx(bg, s.db, card)
			wantLedgerError(t, err, "no such table: transitions")
			if have != nil {
				t.Errorf("facts %v were returned next to the error", have)
			}
		})
	}
}

func TestEnqueue_ADamagedLedgerRefusesTheOfferAndKeepsTheCard(t *testing.T) {
	cases := []struct {
		name   string
		damage []string
		detail string
	}{
		{"no lease table", []string{`DROP TABLE leases`}, "no such table: leases"},
		{"no card table", []string{`DROP TABLE cards`}, "no such table: cards"},
		{"no evidence table", []string{`DROP TABLE evidence`}, "no such table: evidence"},
		{"no queue table", []string{`DROP TABLE merge_queue`}, "no such table: merge_queue"},
		{"an evidence epoch the facts cannot read", []string{`ALTER TABLE evidence RENAME COLUMN epoch TO epoch_gone`}, "no such column"},
		{"a queue write the engine refuses", []string{
			`CREATE TRIGGER refuse_queue BEFORE INSERT ON merge_queue BEGIN SELECT RAISE(ABORT,'queue refused'); END`}, "queue refused"},
		{"a position write the engine refuses", []string{
			`CREATE TRIGGER refuse_move BEFORE UPDATE OF stage ON cards BEGIN SELECT RAISE(ABORT,'move refused'); END`}, "move refused"},
		{"a transition write the engine refuses", []string{
			`CREATE TRIGGER refuse_transition BEFORE INSERT ON transitions BEGIN SELECT RAISE(ABORT,'transition refused'); END`}, "transition refused"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := openTemp(t)
			c := gated(t, s, "a", "c1")
			damage(t, s, tc.damage...)
			wantLedgerError(t, s.Enqueue(bg, "a", c.Epoch, "c1"), tc.detail)
			// Whatever the damage, the offer did not half-happen.
			if !strings.Contains(tc.name, "queue table") {
				if n := count(t, s, `SELECT count(*) FROM merge_queue`); n != 0 {
					t.Errorf("the refused offer left %d queue rows", n)
				}
			}
			if !strings.Contains(tc.name, "card table") {
				if n := count(t, s, `SELECT count(*) FROM cards WHERE id='a' AND stage='MERGING'`); n != 0 {
					t.Errorf("the refused offer moved the card into MERGING")
				}
			}
			if !strings.Contains(tc.name, "evidence table") && !strings.Contains(tc.name, "transition") {
				if n := count(t, s, `SELECT count(*) FROM transitions WHERE card_id='a' AND to_stage='MERGING'`); n != 0 {
					t.Errorf("the refused offer wrote %d transitions into MERGING", n)
				}
			}
		})
	}
}

func TestEnqueue_ARiskCardWithUnreadableReviewEvidenceIsAnError(t *testing.T) {
	// A risk card is checked for review evidence after the gate; with the whole evidence table
	// gone the first lookup (the gate) already fails, so the review lookup is reached only by
	// breaking the one column the gate lookup does not use, which no lookup avoids. The refusal
	// of an unreviewed risk card is therefore asserted on the engine error, not on a sentinel.
	s := openTemp(t)
	c := gated(t, s, "a", "c1", "security")
	damage(t, s, `DROP TABLE evidence`)
	wantLedgerError(t, s.Enqueue(bg, "a", c.Epoch, "c1"), "no such table: evidence")
	if n := count(t, s, `SELECT count(*) FROM merge_queue`); n != 0 {
		t.Errorf("%d queue rows", n)
	}
}

func TestEnqueue_ARefusedNoteKeepsTheLandedRowAsItWas(t *testing.T) {
	for _, tc := range []struct {
		name string
		sha  string
	}{
		{"taking the landing over", "c1"},
		{"replacing the landing with a new commit", "c2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, e2 := landedThenReclaimed(t, false, nil)
			evidenceOn(t, s, e2, tc.sha)
			damage(t, s, `CREATE TRIGGER refuse_note BEFORE INSERT ON notes BEGIN SELECT RAISE(ABORT,'note refused'); END`)
			before, _ := queueRow(t, s, "a")
			wantLedgerError(t, s.Enqueue(bg, "a", e2, tc.sha), "note refused")
			if after, _ := queueRow(t, s, "a"); after != before {
				t.Errorf("the refused note changed the row: %+v -> %+v", before, after)
			}
			if got := cardOf(t, s, "a").Position.Stage; got == "MERGING" {
				t.Errorf("the card moved to %s", got)
			}
		})
	}
}

func TestEnqueue_ARefusedNoteKeepsTheReworkedLandedRowAsItWas(t *testing.T) {
	// The holder that landed c1 took the card back through BUILDING and offers c2: the landing is
	// kept as a note, and when the note cannot be written nothing else changes either.
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	e := c.Epoch
	if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
		t.Fatal(err)
	}
	landOne(t, s, "a", e)
	backOut(t, s, e)
	mustMove(t, s, e, posImplement, "lane")
	evidenceOn(t, s, e, "c2")
	mustMove(t, s, e, posReviewing, "lane")
	damage(t, s, `CREATE TRIGGER refuse_note BEFORE INSERT ON notes BEGIN SELECT RAISE(ABORT,'note refused'); END`)
	before, _ := queueRow(t, s, "a")
	wantLedgerError(t, s.Enqueue(bg, "a", e, "c2"), "note refused")
	if after, _ := queueRow(t, s, "a"); after != before {
		t.Errorf("the refused note changed the row: %+v -> %+v", before, after)
	}
}

func TestEnqueue_ATakeOverThatCannotWriteTheRowChangesNothing(t *testing.T) {
	s, e2 := landedThenReclaimed(t, false, nil)
	evidenceOn(t, s, e2, "c1")
	damage(t, s, `CREATE TRIGGER refuse_row BEFORE UPDATE ON merge_queue BEGIN SELECT RAISE(ABORT,'row refused'); END`)
	before, _ := queueRow(t, s, "a")
	wantLedgerError(t, s.Enqueue(bg, "a", e2, "c1"), "row refused")
	if after, _ := queueRow(t, s, "a"); after != before {
		t.Errorf("the refused take-over changed the row: %+v -> %+v", before, after)
	}
	// The decision note written before the row update was rolled back with it.
	if n := count(t, s, `SELECT count(*) FROM notes WHERE card_id='a' AND text LIKE '%taken over%'`); n != 0 {
		t.Errorf("%d take-over notes survived the refused take-over", n)
	}
}

func TestEnqueue_ARestartedFailedRowNeedsItsNote(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	e := c.Epoch
	if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(e, "c1"), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(e, "c1"), QueueFailed, "", "rebase conflict"); err != nil {
		t.Fatal(err)
	}
	backOut(t, s, e)
	evidenceOn(t, s, e, "c2")
	damage(t, s, `CREATE TRIGGER refuse_note BEFORE INSERT ON notes BEGIN SELECT RAISE(ABORT,'note refused'); END`)
	wantLedgerError(t, s.Enqueue(bg, "a", e, "c2"), "note refused")
	if it, _ := queueRow(t, s, "a"); it.State != QueueFailed || it.CandidateSHA != "c1" {
		t.Errorf("the failed row was restarted without its note: %+v", it)
	}
}

func TestEnqueue_AMeasuredInEpochLookupThatFailsRefusesTheTakeOver(t *testing.T) {
	s, e2 := landedThenReclaimed(t, false, nil)
	evidenceOn(t, s, e2, "c1")
	damage(t, s, `ALTER TABLE evidence RENAME COLUMN epoch TO epoch_gone`)
	before, _ := queueRow(t, s, "a")
	err := s.Enqueue(bg, "a", e2, "c1")
	wantLedgerError(t, err, "no such column")
	if after, _ := queueRow(t, s, "a"); after != before {
		t.Errorf("the failed lookup changed the row: %+v -> %+v", before, after)
	}
}

func TestOpen_AFileThatIsNotALedgerIsAnErrorAndNoStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "af", "ledger.db")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Repeat("this is not a database file\n", 200)), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := Open(bg, path)
	if err == nil {
		_ = s.Close()
		t.Fatal("a text file opened as a ledger")
	}
	if s != nil {
		t.Errorf("a store came back with the error: %+v", s)
	}
	if !strings.HasPrefix(err.Error(), "ledger: ") {
		t.Errorf("not a ledger error: %v", err)
	}
}

func TestOpen_ADirectoryInPlaceOfTheDatabaseIsAnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "af", "ledger.db")
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if s, err := Open(bg, path); err == nil {
		_ = s.Close()
		t.Fatal("a directory opened as a ledger")
	} else if !strings.HasPrefix(err.Error(), "ledger: ") {
		t.Errorf("not a ledger error: %v", err)
	}
}

// v1 ledger with part of its history missing: the migration runs in one transaction, so a failure
// leaves the file at schema 1 with nothing changed.
func TestMigrate_ABrokenVersion1LedgerFailsAndStaysVersion1(t *testing.T) {
	cases := []struct {
		name   string
		damage string
		detail string
	}{
		// The schema statements only create what is missing, so the damaged table is replaced by one
		// that exists but lacks the column the migration ranks by.
		{"evidence without its timestamp", `DROP TABLE evidence; CREATE TABLE evidence(id INTEGER PRIMARY KEY, card_id TEXT)`, "migrate"},
		{"transitions without its timestamp", `DROP TABLE transitions; CREATE TABLE transitions(seq INTEGER PRIMARY KEY, card_id TEXT, to_stage TEXT)`, "migrate"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := v1Millisecond(t, false, nil)
			db, err := sql.Open("sqlite", "file:"+path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(tc.damage); err != nil {
				t.Fatal(err)
			}
			_ = db.Close()
			s, err := Open(context.Background(), path)
			if err == nil {
				_ = s.Close()
				t.Fatal("a broken version 1 ledger was migrated")
			}
			if !strings.Contains(err.Error(), tc.detail) {
				t.Errorf("error %q does not say the migration failed", err)
			}
			raw, err := sql.Open("sqlite", "file:"+path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = raw.Close() }()
			var v int
			if err := raw.QueryRow(`PRAGMA user_version`).Scan(&v); err != nil {
				t.Fatal(err)
			}
			if v != 1 {
				t.Errorf("user_version %d after a failed migration; it must stay 1", v)
			}
		})
	}
}
