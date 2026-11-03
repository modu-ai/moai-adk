package homestate

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

// Card t2013: the wake events are ordered. `received` needs `delivered`, and
// `started` needs `received`; an event recorded out of order is refused with
// ErrWakeOutOfOrder and writes nothing.

func TestWakeEventsRefuseOutOfOrder(t *testing.T) {
	ctx := context.Background()
	db, err := OpenFactoryPath(filepath.Join(t.TempDir(), "factory.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	now := time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)
	w := CardWait{CardID: "t2020", CondKind: WaitCondPredecessorLanded, CondRef: "t2010", Waker: "lane-3"}
	if err := db.RecordCardWait(ctx, w, now); err != nil {
		t.Fatal(err)
	}
	waits, err := db.LoadCardWaits(ctx, "t2020")
	if err != nil || len(waits) != 1 {
		t.Fatalf("waits = %+v, %v", waits, err)
	}
	seq := waits[0].Seq

	// Not delivered yet: neither received nor started may be recorded.
	for _, phase := range []string{WakeReceived, WakeStarted} {
		if _, err := db.RecordWakeEvent(ctx, seq, phase, "lane-3", now); !errors.Is(err, ErrWakeOutOfOrder) {
			t.Fatalf("%s before delivered = %v, want ErrWakeOutOfOrder", phase, err)
		}
	}
	if n := wakeEventCount(t, db); n != 0 {
		t.Fatalf("a refused event wrote: %d wake events, want 0", n)
	}

	if settled, err := db.SettleLandedWaits(ctx, "t2010", now); err != nil || len(settled) != 1 {
		t.Fatalf("settle = %+v, %v", settled, err)
	}
	// Delivered, not received: started is still refused.
	if _, err := db.RecordWakeEvent(ctx, seq, WakeStarted, "lane-3", now); !errors.Is(err, ErrWakeOutOfOrder) {
		t.Fatalf("started before received = %v, want ErrWakeOutOfOrder", err)
	}
	if n := wakeEventCount(t, db); n != 1 {
		t.Fatalf("wake events = %d, want 1 (delivered only)", n)
	}
	for _, phase := range []string{WakeReceived, WakeStarted} {
		if rec, err := db.RecordWakeEvent(ctx, seq, phase, "lane-3", now); err != nil || !rec {
			t.Fatalf("in-order %s = (%v, %v), want recorded", phase, rec, err)
		}
	}
}
