package cli

// factory_capacity_test.go — SPEC-CODEX-LANE-SLOTS-001 AC-006 (REQ-004/005):
// the run record carries the declared lane capacity from leader start, and
// the codex lane join reads exactly what the record holds. An explicit count
// (`-k N`) records N; a count-less leader start records the derived-capacity
// marker.

import (
	"context"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestRecordFactoryRunStartCapacity(t *testing.T) {
	root := t.TempDir()

	// An explicit-count leader start records N (REQ-004).
	if err := recordFactoryRunStart(root, "run-explicit01", "codex", "", 3); err != nil {
		t.Fatalf("record explicit-capacity run: %v", err)
	}
	// A count-less leader start records the derived-capacity marker (REQ-004):
	// the codex twin carries no count form, so its runs are capacity-open.
	if err := recordFactoryRunStart(root, "run-derived01", "codex", "", homestate.LaneCapacityDerived); err != nil {
		t.Fatalf("record derived-capacity run: %v", err)
	}

	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()

	capacity, found, err := db.RunLaneCapacity(context.Background(), "run-explicit01")
	if err != nil || !found || capacity != 3 {
		t.Fatalf("explicit run capacity = (%d, %v, %v), want (3, true, nil)", capacity, found, err)
	}
	capacity, found, err = db.RunLaneCapacity(context.Background(), "run-derived01")
	if err != nil || !found || capacity != homestate.LaneCapacityDerived {
		t.Fatalf("derived run capacity = (%d, %v, %v), want (LaneCapacityDerived, true, nil)", capacity, found, err)
	}
	if _, found, err := db.RunLaneCapacity(context.Background(), "run-absent01"); err != nil || found {
		t.Fatalf("absent run read = (found=%v, err=%v), want (false, nil)", found, err)
	}

	// The codex lane join reads exactly what the record holds (AC-006): the
	// explicit count becomes the join bound; a capacity-open record and an
	// absent record fall back to the leader fan-out default. The claim engine
	// re-reads the record inside its own transaction for the automatic scan.
	if got := factoryJoinLaneBound(root, "run-explicit01"); got != 3 {
		t.Fatalf("join bound for explicit-3 run = %d, want 3", got)
	}
	if got := factoryJoinLaneBound(root, "run-derived01"); got != config.DefaultFactoryLeaderLanes {
		t.Fatalf("join bound for derived run = %d, want the leader default %d", got, config.DefaultFactoryLeaderLanes)
	}
	if got := factoryJoinLaneBound(root, "run-absent01"); got != config.DefaultFactoryLeaderLanes {
		t.Fatalf("join bound for unrecorded run = %d, want the leader default %d", got, config.DefaultFactoryLeaderLanes)
	}
}
