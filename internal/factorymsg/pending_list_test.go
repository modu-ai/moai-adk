package factorymsg

import (
	"context"
	"slices"
	"testing"
	"time"
)

// TestPendingListIncludesExpiredClaim pins parity between the two hook-side
// surfaces (t1632): PendingCount already counts a row whose claim lease has
// expired as pending (REQ-MN-011), so PendingList must render it too — a
// session that claimed a message and died leaves the row `claimed` until the
// lease runs out, and from that moment the row is redeliverable. A LIVE claim
// stays excluded: only the expired lease re-enters the pending surface.
func TestPendingListIncludesExpiredClaim(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t)
	lane := f.register("lane-1", "lane", "lane-1-s1", "start-lane-1")

	if _, err := f.s.Send(ctx, SendRequest{
		From: f.lead, To: lane, Kind: KindStatusRequest,
		IdempotencyKey: "pending-list-expired", TaskRef: "t1632", CorrelationID: "c1632",
		TTL: time.Hour, Payload: []byte("redeliver me"),
	}); err != nil {
		t.Fatal(err)
	}

	base := time.Now().UTC()
	clock := base
	f.s.now = func() time.Time { return clock }

	if _, err := f.s.Claim(ctx, lane, MaxBatch, time.Minute); err != nil {
		t.Fatal(err)
	}

	// Inside the lease the claim is live: not pending for any surface.
	live, err := f.s.PendingList(ctx, lane, MaxBatch)
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 0 {
		t.Fatalf("PendingList over a live claim = %+v; want nothing (a live claim is not pending)", live)
	}

	// Past the lease the row is redeliverable — the counting path already says
	// so, and the rendered list must agree.
	clock = base.Add(10 * time.Minute)
	got, err := f.s.PendingList(ctx, lane, MaxBatch)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(got, func(e PendingListEntry) bool { return e.Kind == KindStatusRequest }) {
		t.Fatalf("PendingList after the claim lease = %+v; want the expired-claim row rendered as pending (the hook's empty-inbox path hid a redeliverable message)", got)
	}
}
