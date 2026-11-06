// factorymsg_claim_stability_test.go — SPEC-MCP-NEUTRALITY-001 M5 (card
// t1632, REQ-MN-016/017): generation advance migrates pending messages to
// the new generation (the 9-orphan defect family), and a same-endpoint
// re-claim inside a live lease keeps the claim token and extends the lease
// instead of minting a new one (the "claim identity mismatch" defect family).
package cli

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// csFixture builds a broker at the fixture root with a leader and a lane
// registered at generation 1, plus one pending message for the lane.
func csFixture(t *testing.T) (string, *factorymsg.Store, factorymsg.Peer, factorymsg.Peer) {
	t.Helper()
	root, _ := sdMoaiFixture(t)
	ctx := context.Background()
	store, err := factorymsg.Open(root, "run-cli")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	key := homestate.ProjectKey(root)
	owner := os.Getpid()
	start := homestate.CurrentProcessFingerprint()
	// The lane's initial registration carries a DEAD process identity, so the
	// same-session relaunch below passes the live-owner check and advances
	// the generation (the production /clear + restart shape).
	lane, err := store.RegisterPeer(ctx, factorymsg.Peer{ProjectKey: key, RunID: "run-cli", Backend: "claude", Role: "lane", Slot: "lane-3", SessionUUID: "s1", Generation: 1, PID: 99999, ProcessStart: "dead-fixture-start"})
	if err != nil {
		t.Fatal(err)
	}
	lead, err := store.RegisterPeer(ctx, factorymsg.Peer{ProjectKey: key, RunID: "run-cli", Backend: "claude", Role: "leader", Slot: "leader", SessionUUID: "lead", Generation: 1, PID: owner, ProcessStart: start})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Send(ctx, factorymsg.SendRequest{From: lead, To: lane, Kind: factorymsg.KindStatusRequest, IdempotencyKey: "k1", TaskRef: "t1632", CorrelationID: "c1", TTL: time.Hour, Payload: []byte("m1")}); err != nil {
		t.Fatal(err)
	}
	return root, store, lane, lead
}

// TestStoreGenerationAdvanceMigratesPending pins REQ-MN-016: a generation
// advance on re-registration MIGRATES the pending messages addressed to the
// old generation to the new one — no pending message is left unreachable.
func TestStoreGenerationAdvanceMigratesPending(t *testing.T) {
	root, store, lane, _ := csFixture(t)
	ctx := context.Background()

	// The SAME session relaunches under a new process (a /clear + restart):
	// the generation advances and the pending message migrates.
	owner := os.Getpid() + 1
	start := homestate.CurrentProcessFingerprint() + "-relaunched"
	advanced, err := store.RegisterPeer(ctx, factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: "run-cli", Backend: "claude", Role: "lane", Slot: lane.Slot, SessionUUID: lane.SessionUUID, Generation: 0, PID: owner, ProcessStart: start})
	if err != nil {
		t.Fatal(err)
	}
	if advanced.Generation <= lane.Generation {
		t.Fatalf("generation did not advance: %d → %d", lane.Generation, advanced.Generation)
	}

	// The pending message is now addressed to generation 2 — visible to the
	// lane's Claim at the new generation.
	claims, err := store.Claim(ctx, advanced, 16, 30*time.Second)
	if err != nil {
		t.Fatalf("claim at generation %d: %v", advanced.Generation, err)
	}
	if len(claims) != 1 {
		t.Fatalf("claims at generation %d = %d, want 1 (the pending message must migrate, not orphan)", advanced.Generation, len(claims))
	}
}

// TestClaimTokenStableOnRelist pins REQ-MN-017: a same-endpoint re-claim
// inside a live lease keeps the claim token and the messages — the second
// factory_msg_list returns the already-claimed entries with the SAME token
// (extended lease), never fresh tokens for live claims.
func TestClaimTokenStableOnRelist(t *testing.T) {
	root, store, lane, _ := csFixture(t)
	ctx := context.Background()

	first, err := store.Claim(ctx, lane, 16, 30*time.Second)
	if err != nil || len(first) != 1 {
		t.Fatalf("first claim = %v (len %d), want 1 claim", err, len(first))
	}

	second, err := store.Claim(ctx, lane, 16, 30*time.Second)
	if err != nil {
		t.Fatalf("second claim: %v", err)
	}
	if len(second) != 1 {
		t.Fatalf("second claim returned %d entries, want the already-claimed entry", len(second))
	}
	if second[0].ID != first[0].ID {
		t.Errorf("second claim id = %q, want the same message %q", second[0].ID, first[0].ID)
	}
	if second[0].ClaimToken != first[0].ClaimToken {
		t.Errorf("claim token rotated on relist: %q → %q (the identity-mismatch defect)", first[0].ClaimToken, second[0].ClaimToken)
	}

	// The body is still readable with the ORIGINAL token — the identity the
	// first claim handed out stayed valid.
	if _, err := store.ReadBody(ctx, lane, first[0].ID, first[0].ClaimToken); err != nil {
		t.Errorf("ReadBody with the first claim's token after relist: %v", err)
	}
	_ = root
	_ = os.Getpid
}
