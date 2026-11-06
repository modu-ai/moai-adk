package factorymsg

import (
	"context"
	"testing"
	"time"
)

// TestClaimFreshSkipsAHeldClaim pins the split between the two claim forms
// (card t1632). Claim keeps REQ-MN-017: a re-claim inside the lease returns the
// same token. ClaimFresh never delivers a row the endpoint still holds, and
// treats a row whose lease has expired as fresh again (new token) — the
// redelivery policy a one-turn-per-delivery consumer relies on.
func TestClaimFreshSkipsAHeldClaim(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t)
	lane := f.register("lane-1", "lane", "lane-1-s1", "start-lane-1")

	if _, err := f.s.Send(ctx, SendRequest{
		From: f.lead, To: lane, Kind: KindStatusRequest,
		IdempotencyKey: "claim-fresh-once", TaskRef: "t1632", CorrelationID: "c1632",
		TTL: time.Hour, Payload: []byte("held"),
	}); err != nil {
		t.Fatal(err)
	}

	base := time.Now().UTC()
	clock := base
	f.s.now = func() time.Time { return clock }

	first, err := f.s.ClaimFresh(ctx, lane, MaxBatch, time.Minute)
	if err != nil || len(first) != 1 {
		t.Fatalf("first ClaimFresh = %+v, %v; want exactly the pending row", first, err)
	}

	again, err := f.s.ClaimFresh(ctx, lane, MaxBatch, time.Minute)
	if err != nil || len(again) != 0 {
		t.Fatalf("ClaimFresh over a held claim = %+v, %v; want nothing (no redelivery inside the lease)", again, err)
	}

	// REQ-MN-017 is untouched: Claim still re-delivers the live claim, same token.
	same, err := f.s.Claim(ctx, lane, MaxBatch, time.Minute)
	if err != nil || len(same) != 1 || same[0].ClaimToken != first[0].ClaimToken {
		t.Fatalf("Claim over a live claim = %+v, %v; want the same token %q", same, err, first[0].ClaimToken)
	}

	// Past the lease the row is fresh again and the old identity dies with it.
	clock = base.Add(10 * time.Minute)
	renewed, err := f.s.ClaimFresh(ctx, lane, MaxBatch, time.Minute)
	if err != nil || len(renewed) != 1 {
		t.Fatalf("ClaimFresh after the lease = %+v, %v; want the expired row", renewed, err)
	}
	if renewed[0].ClaimToken == first[0].ClaimToken {
		t.Errorf("an expired claim kept its token %q; the fencing contract needs a new one", renewed[0].ClaimToken)
	}
}
