package factorymsg

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// t1772 reproduction. The run's owner moves to a live leader, but the broker's
// "leader" slot stays bound to the dead session that registered it first. A
// lane that addresses "leader" must not be resolved onto that dead endpoint:
// the message would be queued for a session that can no longer read it.
func TestResolveLeaderSlotRebindsDeadOwnerToRunOwner(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	dead := testPeer(s.runID, "sess-dead-leader", 1)
	dead.Role, dead.Slot = "leader", "leader"
	dead.PID, dead.ProcessStart = 15140, "dead-leader-start"
	if _, err := s.RegisterPeer(ctx, dead); err != nil {
		t.Fatal(err)
	}

	// openTestStore reads only this process as live, so the run owner below is
	// the live leader that replaced the dead session.
	recordRunOwner(t, s.root, s.runID, "sess-live-leader", os.Getpid(), testPeer(s.runID, "x", 1).ProcessStart)

	got, err := s.ResolveLane(ctx, "leader")
	if err != nil {
		t.Fatalf("ResolveLane(leader) err=%v, want the live run owner", err)
	}
	if got.SessionUUID == dead.SessionUUID {
		t.Fatalf("ResolveLane(leader) returned the dead owner %s; a lane message would be queued for a session that cannot read it", got.SessionUUID)
	}
	if got.SessionUUID != "sess-live-leader" || got.Generation <= dead.Generation {
		t.Fatalf("ResolveLane(leader) = %s gen %d, want sess-live-leader at a generation above %d", got.SessionUUID, got.Generation, dead.Generation)
	}

	// The rebind is stored, not only returned: the slot now names the live lead.
	stored, err := s.Peer(ctx, "sess-live-leader")
	if err != nil || stored.Slot != "leader" || stored.Generation != got.Generation {
		t.Fatalf("slot row not rebound in the store: %+v err=%v", stored, err)
	}
	if _, err := s.Peer(ctx, dead.SessionUUID); err == nil {
		t.Fatal("dead session still addresses a slot after the rebind")
	}
}

// When the run's lead is dead as well, there is no live endpoint to rebind to,
// so the lookup is refused and never resolves onto the dead owner.
func TestResolveLeaderSlotRefusesWhenRunLeadIsDead(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	dead := testPeer(s.runID, "sess-dead-leader", 1)
	dead.Role, dead.Slot = "leader", "leader"
	dead.PID, dead.ProcessStart = 15140, "dead-leader-start"
	if _, err := s.RegisterPeer(ctx, dead); err != nil {
		t.Fatal(err)
	}
	recordRunOwner(t, s.root, s.runID, "sess-also-dead", 15141, "also-dead-start")

	if _, err := s.ResolveLane(ctx, "leader"); !errors.Is(err, ErrLeaderOwnerDead) {
		t.Fatalf("ResolveLane(leader) err=%v, want ErrLeaderOwnerDead", err)
	}
}

// A leader slot whose owner is live is returned unchanged: no rebind happens.
func TestResolveLeaderSlotKeepsLiveOwner(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	live := testPeer(s.runID, "sess-live-leader", 1)
	live.Role, live.Slot = "leader", "leader"
	if _, err := s.RegisterPeer(ctx, live); err != nil {
		t.Fatal(err)
	}

	got, err := s.ResolveLane(ctx, "leader")
	if err != nil || got.SessionUUID != live.SessionUUID || got.Generation != live.Generation {
		t.Fatalf("ResolveLane(leader) = %+v err=%v, want the live row unchanged", got, err)
	}
}

// A run row with no session cannot prove which process the owner pid belongs
// to, so the slot is refused rather than bound to an empty session (t1772 review P1).
func TestResolveLeaderSlotRefusesWhenRunSessionUnknown(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	dead := testPeer(s.runID, "sess-dead-leader", 1)
	dead.Role, dead.Slot = "leader", "leader"
	dead.PID, dead.ProcessStart = 15140, "dead-leader-start"
	if _, err := s.RegisterPeer(ctx, dead); err != nil {
		t.Fatal(err)
	}
	recordRunOwner(t, s.root, s.runID, "", os.Getpid(), testPeer(s.runID, "x", 1).ProcessStart)

	if _, err := s.ResolveLane(ctx, "leader"); !errors.Is(err, ErrLeaderOwnerDead) {
		t.Fatalf("ResolveLane(leader) with an unknown run session err=%v, want ErrLeaderOwnerDead", err)
	}
}

// A run row that still names the dead session is not a rebind target, even when
// its pid is live: a resume can keep the old session on the row (t1772 review P2).
func TestResolveLeaderSlotRefusesStaleRunSession(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	dead := testPeer(s.runID, "sess-dead-leader", 1)
	dead.Role, dead.Slot = "leader", "leader"
	dead.PID, dead.ProcessStart = 15140, "dead-leader-start"
	if _, err := s.RegisterPeer(ctx, dead); err != nil {
		t.Fatal(err)
	}
	recordRunOwner(t, s.root, s.runID, "sess-dead-leader", os.Getpid(), testPeer(s.runID, "x", 1).ProcessStart)

	if _, err := s.ResolveLane(ctx, "leader"); !errors.Is(err, ErrLeaderOwnerDead) {
		t.Fatalf("ResolveLane(leader) with the dead session still on the run err=%v, want ErrLeaderOwnerDead", err)
	}
}

// recordRunOwner writes the run row the way the leader's start does, so the
// run's current owner is the given session.
func recordRunOwner(t *testing.T, root, runID, sessionUUID string, pid int, processStart string) {
	t.Helper()
	registry, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = registry.Close() }()
	if err := registry.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: runID, LeadSessionID: sessionUUID, LeadPID: pid, LeadProcessStart: processStart,
		Backend: "claude", ManifestJSON: "{}",
	}); err != nil {
		t.Fatal(err)
	}
}
