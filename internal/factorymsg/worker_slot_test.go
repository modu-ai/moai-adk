package factorymsg

import (
	"context"
	"testing"
)

// TestRegisterPeerAllocatesWorkerSlots pins the auto-slot allocator on the
// worker vocabulary: the bare `worker` sentinel (and the legacy bare `agent`
// sentinel) take the next free `worker-<n>` slot.
func TestRegisterPeerAllocatesWorkerSlots(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	for i, sentinel := range []string{"worker", "agent"} {
		p := testPeer("run", "alloc-"+sentinel, 0)
		p.Slot = sentinel
		got, err := store.RegisterPeer(ctx, p)
		if err != nil {
			t.Fatalf("register %s: %v", sentinel, err)
		}
		want := []string{"worker-1", "worker-2"}[i]
		if got.Slot != want {
			t.Errorf("sentinel %q allocated %q, want %q", sentinel, got.Slot, want)
		}
	}
}

// TestRegisterPeerLegacySlotRowsStayReadableAndBlockTheirNumber is the
// persisted-format compat read: a broker.db row written under the legacy
// `agent-<n>` slot (an older launcher in the same run) keeps its number and
// stays addressable by its session.
func TestRegisterPeerLegacySlotRowsStayReadableAndBlockTheirNumber(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	legacy := testPeer("run", "legacy-session", 1) // Slot "agent-1"
	if _, err := store.RegisterPeer(ctx, legacy); err != nil {
		t.Fatalf("seed legacy row: %v", err)
	}
	fresh := testPeer("run", "fresh-session", 0)
	fresh.Slot = "worker"
	got, err := store.RegisterPeer(ctx, fresh)
	if err != nil {
		t.Fatalf("register worker: %v", err)
	}
	if got.Slot != "worker-2" {
		t.Errorf("worker allocated %q beside a legacy agent-1 row, want worker-2", got.Slot)
	}
	read, err := store.Peer(ctx, "legacy-session")
	if err != nil || read.Slot != "agent-1" {
		t.Errorf("legacy row read = (%+v, %v), want slot agent-1 unchanged", read, err)
	}
}
