package homestate

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

// Card t2013: a wake is three distinct recorded events — delivered, received,
// started — so an acknowledgement is never read as a start. Recording the same
// event twice, or again after a restart, writes nothing.

func wakeFixture(t *testing.T) (*FactoryDB, string, CardWait, time.Time) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "factory.db")
	db, err := OpenFactoryPath(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	now := time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)
	w := CardWait{CardID: "t2020", CondKind: WaitCondPredecessorLanded, CondRef: "t2010", Waker: "lane-3", RecordedBy: "lane-3"}
	if err := db.RecordCardWait(context.Background(), w, now); err != nil {
		t.Fatal(err)
	}
	settled, err := db.SettleLandedWaits(context.Background(), "t2010", now)
	if err != nil || len(settled) != 1 {
		t.Fatalf("settle = %+v, %v, want one wait", settled, err)
	}
	return db, path, settled[0], now
}

func wakeEventCount(t *testing.T, db *FactoryDB) int {
	t.Helper()
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM events WHERE kind LIKE 'wake.%'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestWakeSettleRecordsDeliveredOnly(t *testing.T) {
	db, _, w, _ := wakeFixture(t)
	st, err := db.LoadWakeState(context.Background(), w.Seq)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Delivered || st.Received || st.Started {
		t.Fatalf("state after settle = %+v, want delivered only", st)
	}
	if n := wakeEventCount(t, db); n != 1 {
		t.Fatalf("wake events = %d, want 1 (delivered)", n)
	}
}

func TestWakeReceivedIsNotStarted(t *testing.T) {
	db, _, w, now := wakeFixture(t)
	ctx := context.Background()
	if rec, err := db.RecordWakeEvent(ctx, w.Seq, WakeReceived, "lane-3", now); err != nil || !rec {
		t.Fatalf("record received = (%v, %v), want recorded", rec, err)
	}
	st, _ := db.LoadWakeState(ctx, w.Seq)
	if !st.Delivered || !st.Received || st.Started {
		t.Fatalf("state after the acknowledgement = %+v, want delivered+received, NOT started", st)
	}
	if rec, err := db.RecordWakeEvent(ctx, w.Seq, WakeStarted, "lane-3", now); err != nil || !rec {
		t.Fatalf("record started = (%v, %v), want recorded", rec, err)
	}
	st, _ = db.LoadWakeState(ctx, w.Seq)
	if !st.Started {
		t.Fatalf("state after start = %+v, want started", st)
	}
}

func TestWakeEventsAreIdempotentAcrossRestart(t *testing.T) {
	db, path, w, now := wakeFixture(t)
	ctx := context.Background()
	for _, phase := range []string{WakeReceived, WakeStarted} {
		if rec, err := db.RecordWakeEvent(ctx, w.Seq, phase, "lane-3", now); err != nil || !rec {
			t.Fatalf("first %s = (%v, %v)", phase, rec, err)
		}
		if rec, err := db.RecordWakeEvent(ctx, w.Seq, phase, "lane-3", now); err != nil || rec {
			t.Fatalf("duplicate %s = (%v, %v), want (false, nil)", phase, rec, err)
		}
	}
	if n := wakeEventCount(t, db); n != 3 {
		t.Fatalf("wake events = %d, want 3 (delivered, received, started once each)", n)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := OpenFactoryPath(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reopened.Close() }()
	for _, phase := range []string{WakeDelivered, WakeReceived, WakeStarted} {
		if rec, err := reopened.RecordWakeEvent(ctx, w.Seq, phase, "lane-3", now); err != nil || rec {
			t.Fatalf("after restart %s = (%v, %v), want (false, nil)", phase, rec, err)
		}
	}
	if n := wakeEventCount(t, reopened); n != 3 {
		t.Fatalf("wake events after restart = %d, want 3", n)
	}
	if again, err := reopened.SettleLandedWaits(ctx, "t2010", now); err != nil || len(again) != 0 {
		t.Fatalf("settle after restart = %+v, %v, want nothing", again, err)
	}
	if n := wakeEventCount(t, reopened); n != 3 {
		t.Fatalf("a repeated settle added wake events: %d, want 3", n)
	}
}

func TestWakeEventRefusesUnknownWaitAndPhase(t *testing.T) {
	db, _, w, now := wakeFixture(t)
	ctx := context.Background()
	if _, err := db.RecordWakeEvent(ctx, w.Seq+99, WakeStarted, "lane-3", now); !errors.Is(err, ErrInvalidCardInput) {
		t.Fatalf("unknown wait = %v, want ErrInvalidCardInput", err)
	}
	if _, err := db.RecordWakeEvent(ctx, w.Seq, "woke-up", "lane-3", now); !errors.Is(err, ErrInvalidCardInput) {
		t.Fatalf("unknown phase = %v, want ErrInvalidCardInput", err)
	}
	if n := wakeEventCount(t, db); n != 1 {
		t.Fatalf("a refused record wrote: wake events = %d, want 1", n)
	}
}
