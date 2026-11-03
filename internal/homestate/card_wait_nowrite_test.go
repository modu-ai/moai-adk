package homestate

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

// waitWriteCounter installs row-level triggers on card_waits in a TEST database
// only (the production DDL is untouched) that log every INSERT and UPDATE
// statement's effect, so a test can count write operations rather than infer
// them. SQLite fires an UPDATE trigger even when the new values equal the old
// ones, so an UPDATE that rewrites identical values is counted.
type waitWriteCounter struct {
	t  *testing.T
	db *FactoryDB
}

func newWaitWriteCounter(t *testing.T, db *FactoryDB) *waitWriteCounter {
	t.Helper()
	for _, stmt := range []string{
		`CREATE TABLE wait_write_log (id INTEGER PRIMARY KEY AUTOINCREMENT, op TEXT NOT NULL)`,
		`CREATE TRIGGER wait_log_ins AFTER INSERT ON card_waits BEGIN INSERT INTO wait_write_log(op) VALUES('INSERT'); END`,
		`CREATE TRIGGER wait_log_upd AFTER UPDATE ON card_waits BEGIN INSERT INTO wait_write_log(op) VALUES('UPDATE'); END`,
	} {
		if _, err := db.DB.Exec(stmt); err != nil {
			t.Fatalf("install counter %q: %v", stmt, err)
		}
	}
	return &waitWriteCounter{t: t, db: db}
}

func (c *waitWriteCounter) count(op string) int {
	c.t.Helper()
	var n int
	if err := c.db.DB.QueryRow(`SELECT count(*) FROM wait_write_log WHERE op=?`, op).Scan(&n); err != nil {
		c.t.Fatal(err)
	}
	return n
}

func (c *waitWriteCounter) reset() {
	c.t.Helper()
	if _, err := c.db.DB.Exec(`DELETE FROM wait_write_log`); err != nil {
		c.t.Fatal(err)
	}
}

func openWaitDB(t *testing.T) *FactoryDB {
	t.Helper()
	db, err := OpenFactoryPath(filepath.Join(t.TempDir(), "factory.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// card t2013: re-recording an identical wait performs ZERO write operations and
// leaves the row (every column, including recorded_at) as it was.
func TestRecordCardWaitIdenticalReRecordPerformsNoWrite(t *testing.T) {
	ctx := context.Background()
	db := openWaitDB(t)
	counter := newWaitWriteCounter(t, db)
	first := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	later := first.Add(time.Hour)
	w := CardWait{CardID: "t2020", CondKind: WaitCondPredecessorLanded, CondRef: "t2010", Waker: "lane-3", Reason: "needs t2010", RecordedBy: "lane-3"}

	if err := db.RecordCardWait(ctx, w, first); err != nil {
		t.Fatal(err)
	}
	if got := counter.count("INSERT"); got != 1 {
		t.Fatalf("first record: want 1 INSERT, got %d", got)
	}
	snapBefore, err := db.LoadCardWaits(ctx, "t2020")
	if err != nil || len(snapBefore) != 1 {
		t.Fatalf("snapshot before: %+v err=%v", snapBefore, err)
	}
	counter.reset()

	if err := db.RecordCardWait(ctx, w, later); err != nil {
		t.Fatal(err)
	}
	t.Logf("identical re-record: UPDATE statements=%d INSERT statements=%d", counter.count("UPDATE"), counter.count("INSERT"))
	if got := counter.count("UPDATE") + counter.count("INSERT"); got != 0 {
		t.Errorf("identical re-record performed %d write operation(s), want 0", got)
	}
	snapAfter, err := db.LoadCardWaits(ctx, "t2020")
	if err != nil || len(snapAfter) != 1 {
		t.Fatalf("snapshot after: %+v err=%v", snapAfter, err)
	}
	if snapBefore[0] != snapAfter[0] {
		t.Errorf("row changed on an identical re-record:\n before %+v\n after  %+v", snapBefore[0], snapAfter[0])
	}
}

// A re-record that changes waker, reason, or recorded_by still updates in place
// (one UPDATE, same seq, recorded_at unchanged).
func TestRecordCardWaitChangedReRecordStillWrites(t *testing.T) {
	ctx := context.Background()
	db := openWaitDB(t)
	counter := newWaitWriteCounter(t, db)
	first := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	w := CardWait{CardID: "t2020", CondKind: WaitCondPredecessorLanded, CondRef: "t2010", Waker: "lane-3", Reason: "first", RecordedBy: "a"}
	if err := db.RecordCardWait(ctx, w, first); err != nil {
		t.Fatal(err)
	}
	before, _ := db.LoadCardWaits(ctx, "t2020")

	for name, mutate := range map[string]func(*CardWait){
		"reason":      func(w *CardWait) { w.Reason = "second" },
		"waker":       func(w *CardWait) { w.Waker = "lane-4" },
		"recorded_by": func(w *CardWait) { w.RecordedBy = "b" },
	} {
		counter.reset()
		changed := w
		mutate(&changed)
		if err := db.RecordCardWait(ctx, changed, first.Add(time.Hour)); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got := counter.count("UPDATE"); got != 1 {
			t.Errorf("%s changed: want 1 UPDATE, got %d", name, got)
		}
		after, _ := db.LoadCardWaits(ctx, "t2020")
		if len(after) != 1 || after[0].Seq != before[0].Seq || after[0].RecordedAt != before[0].RecordedAt {
			t.Errorf("%s changed: row identity/recorded_at moved: before %+v after %+v", name, before, after)
		}
		w = changed // next case builds on the stored values
		before = after
	}
}
