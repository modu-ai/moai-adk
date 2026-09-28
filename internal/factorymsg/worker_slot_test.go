package factorymsg

// worker_slot_test.go — the broker's auto-slot allocator, updated to the
// SPEC-ROLE-NAMING-CODE-001 vocabulary: the bare `lane` role input takes the
// next free `lane-<n>` slot probing ONLY the canonical shape; bare legacy
// role inputs (`worker`, `agent`) are refused, and legacy slots are not
// addressable (REQ-RNC-009, REQ-RNC-013).

import (
	"context"
	"strings"
	"testing"
)

// TestRegisterPeerAllocatesLaneSlots: the bare `lane` sentinel takes the next
// free `lane-<n>` slot, probing only the canonical shape — a legacy row in
// the run holds no number.
func TestRegisterPeerAllocatesLaneSlots(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	for i, want := range []string{"lane-1", "lane-2"} {
		p := testPeer("run", "alloc-"+want, 0)
		p.Slot = "lane"
		got, err := store.RegisterPeer(ctx, p)
		if err != nil {
			t.Fatalf("register lane %d: %v", i+1, err)
		}
		if got.Slot != want {
			t.Errorf("bare lane allocated %q, want %q", got.Slot, want)
		}
	}
}

// TestRegisterPeerRefusesBareLegacyRoleInputs: `worker` / `agent` as a slot
// input are refused, never numbered (writers write the new vocabulary).
func TestRegisterPeerRefusesBareLegacyRoleInputs(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	for _, sentinel := range []string{"worker", "agent"} {
		p := testPeer("run", "refused-"+sentinel, 0)
		p.Slot = sentinel
		if _, err := store.RegisterPeer(ctx, p); err == nil || !strings.Contains(err.Error(), "lane") {
			t.Errorf("bare %q = %v, want refusal naming lane", sentinel, err)
		}
	}
}

// TestRegisterPeerLegacySlotRowsIgnoredByLaneNumbering: a broker row written
// under the legacy `agent-<n>` slot (an older launcher in the run) is never
// rewritten, and the canonical lane numbering does not step around it — the
// same-run refusal replaced the shared number space (design §4).
func TestRegisterPeerLegacySlotRowsIgnoredByLaneNumbering(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	legacy := testPeer("run", "legacy-session", 1) // Slot "lane-1" via the fixture
	legacy.Role = "worker"
	legacy.Slot = "agent-1"
	if _, err := store.RegisterPeer(ctx, legacy); err != nil {
		t.Fatalf("seed legacy row: %v", err)
	}
	fresh := testPeer("run", "fresh-session", 0)
	fresh.Slot = "lane"
	got, err := store.RegisterPeer(ctx, fresh)
	if err != nil {
		t.Fatalf("register lane: %v", err)
	}
	if got.Slot != "lane-1" {
		t.Errorf("lane allocated %q beside a legacy agent-1 row, want lane-1 (legacy holds no number)", got.Slot)
	}
	read, err := store.Peer(ctx, "legacy-session")
	if err != nil || read.Slot != "agent-1" {
		t.Errorf("legacy row read = (%+v, %v), want slot agent-1 unchanged", read, err)
	}
}
