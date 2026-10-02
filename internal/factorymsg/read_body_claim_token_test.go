package factorymsg

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestReadBodyClaimToken is SPEC-FACTORY-MANAGED-SESSION-001 AC-MS-005
// (REQ-MS-004): a claimed message's body is readable only through its claim
// token. A read without the token — wrong token or empty token — is refused,
// and the token-gated read returns the exact body. Characterization of the
// store's standing claim-token gate: the capability predates the SPEC, so
// this test pins the behavior the managed session layer relies on rather
// than driving new store code.
func TestReadBodyClaimToken(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t)
	lane := f.register("lane-1", "lane", "lane-1-s1", "start-lane-1")

	if _, err := f.s.Send(ctx, SendRequest{
		From: f.lead, To: lane, Kind: KindStatusRequest,
		IdempotencyKey: "body-token-once", TaskRef: "t1375", CorrelationID: "c1375",
		TTL: time.Hour, Payload: []byte("CLAIM_TOKEN_BODY_PROBE"),
	}); err != nil {
		t.Fatal(err)
	}
	claims, err := f.s.Claim(ctx, lane, MaxBatch, time.Minute)
	if err != nil || len(claims) != 1 {
		t.Fatalf("lane claim=%+v err=%v, want exactly one claim", claims, err)
	}
	claim := claims[0]

	if _, err := f.s.ReadBody(ctx, lane, claim.ID, "wrong-token"); err == nil || !strings.Contains(err.Error(), "claim identity mismatch") {
		t.Fatalf("wrong-token read = %v, want claim identity mismatch", err)
	}
	if _, err := f.s.ReadBody(ctx, lane, claim.ID, ""); err == nil {
		t.Fatal("empty-token read was accepted")
	}
	body, err := f.s.ReadBody(ctx, lane, claim.ID, claim.ClaimToken)
	if err != nil || string(body) != "CLAIM_TOKEN_BODY_PROBE" {
		t.Fatalf("token-gated read = %q, %v; want the sent body", body, err)
	}

	// The read window is the claim itself: once the claim is settled with a
	// receipt, the same token no longer reads the body.
	if err := f.s.RecordDisposition(ctx, lane, claim.ID, claim.ClaimToken, DispositionAccepted); err != nil {
		t.Fatal(err)
	}
	if err := f.s.Receipt(ctx, lane, claim.ID, claim.ClaimToken); err != nil {
		t.Fatal(err)
	}
	if _, err := f.s.ReadBody(ctx, lane, claim.ID, claim.ClaimToken); err == nil {
		t.Fatal("token-gated read survived the claim settlement")
	}
}
