package factorylane

import (
	"testing"
	"time"
)

// base is the fixed instant observation tests schedule requests against.
var base = time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)

func newTestStore(t *testing.T) (*Store, *FakeClock) {
	t.Helper()
	clock := &FakeClock{Current: base}
	return NewStore(t.TempDir(), clock), clock
}

// REQ-FLA-002: a directed request is recorded as a readable observation with
// its request timestamp.
func TestRecordRequestAppendsReadableObservation(t *testing.T) {
	store, _ := newTestStore(t)
	if _, err := store.RecordRequest("lane-1"); err != nil {
		t.Fatalf("RecordRequest: %v", err)
	}
	obs, err := store.Observations("lane-1")
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(obs) != 1 {
		t.Fatalf("observation count = %d, want 1", len(obs))
	}
	if !obs[0].RequestedAt.Equal(base) {
		t.Fatalf("RequestedAt = %s, want %s", obs[0].RequestedAt, base)
	}
	if obs[0].AckedAt != nil || obs[0].NoResponseAt != nil {
		t.Fatalf("fresh observation must be unacked and unexpired, got acked=%v noResponse=%v", obs[0].AckedAt, obs[0].NoResponseAt)
	}
	if obs[0].Timer == "" || obs[0].Bound == "" {
		t.Fatalf("timer/bound snapshots empty (%q/%q) — the bound must travel with the observation", obs[0].Timer, obs[0].Bound)
	}
}

// REQ-FLA-002: an ack before the timer expiry marks the observation acked; a
// second ack finds nothing pending.
func TestAckPendingMarksLatestUnackedObservation(t *testing.T) {
	store, clock := newTestStore(t)
	if _, err := store.RecordRequest("lane-1"); err != nil {
		t.Fatalf("RecordRequest: %v", err)
	}
	clock.Current = base.Add(2 * time.Minute)
	acked, err := store.AckPending("lane-1")
	if err != nil || !acked {
		t.Fatalf("AckPending = (%v, %v), want (true, nil)", acked, err)
	}
	obs, err := store.Observations("lane-1")
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if obs[0].AckedAt == nil || !obs[0].AckedAt.Equal(clock.Current) {
		t.Fatalf("AckedAt = %v, want %s", obs[0].AckedAt, clock.Current)
	}
	if obs[0].NoResponseAt != nil {
		t.Fatalf("acked observation must never carry NoResponseAt, got %v", obs[0].NoResponseAt)
	}
	again, err := store.AckPending("lane-1")
	if err != nil {
		t.Fatalf("second AckPending: %v", err)
	}
	if again {
		t.Fatal("second AckPending = true, want false (nothing pending)")
	}
}

// REQ-FLA-002: when the bounded no-response timer expires with no ack, the
// sweep records the no-response observation with its timestamp.
func TestSweepRecordsNoResponseWhenTimerExpiresUnacked(t *testing.T) {
	store, clock := newTestStore(t)
	if _, err := store.RecordRequest("lane-1"); err != nil {
		t.Fatalf("RecordRequest: %v", err)
	}
	timer, err := time.ParseDuration(timerSnapshot)
	if err != nil {
		t.Fatalf("parse timer snapshot %q: %v", timerSnapshot, err)
	}
	clock.Current = base.Add(timer).Add(time.Second)
	obs, err := store.SweepNoResponse("lane-1")
	if err != nil {
		t.Fatalf("SweepNoResponse: %v", err)
	}
	if len(obs) != 1 || obs[0].NoResponseAt == nil {
		t.Fatalf("sweep result %+v, want one observation with NoResponseAt set", obs)
	}
	if !obs[0].NoResponseAt.Equal(clock.Current) {
		t.Fatalf("NoResponseAt = %s, want the sweep instant %s", obs[0].NoResponseAt, clock.Current)
	}
	// The record must be persisted, not just returned.
	reread, err := store.Observations("lane-1")
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
	if reread[0].NoResponseAt == nil {
		t.Fatal("NoResponseAt not persisted — the observation was only mutated in memory")
	}
}

// REQ-FLA-002: a pending (timer not yet expired) and an acked observation are
// both left untouched by the sweep.
func TestSweepLeavesPendingAndAckedObservationsUntouched(t *testing.T) {
	store, clock := newTestStore(t)
	if _, err := store.RecordRequest("lane-1"); err != nil {
		t.Fatalf("RecordRequest: %v", err)
	}
	clock.Current = base.Add(time.Minute)
	if _, err := store.RecordRequest("lane-1"); err != nil {
		t.Fatalf("RecordRequest 2: %v", err)
	}
	if _, err := store.AckPending("lane-1"); err != nil {
		t.Fatalf("AckPending: %v", err)
	}
	clock.Current = base.Add(11 * time.Minute)
	obs, err := store.SweepNoResponse("lane-1")
	if err != nil {
		t.Fatalf("SweepNoResponse: %v", err)
	}
	// The un-acked first request expires; the acked second must not.
	byAcked := 0
	for _, o := range obs {
		if o.NoResponseAt != nil && o.AckedAt != nil {
			t.Fatalf("acked observation recorded no-response: %+v", o)
		}
		if o.AckedAt != nil {
			byAcked++
		}
	}
	if byAcked != 1 {
		t.Fatalf("acked observations = %d, want 1", byAcked)
	}
	var expired int
	for _, o := range obs {
		if o.NoResponseAt != nil {
			expired++
		}
	}
	if expired != 1 {
		t.Fatalf("no-response observations = %d, want 1 (only the unacked, expired one)", expired)
	}
}
