package kanban

// factory_slots_capacity_test.go — SPEC-CODEX-LANE-SLOTS-001 AC-001/002/003/009
// and the §D.1 edge cases: the bounded automatic scan reads the run's
// recorded lane capacity — grows to one past the highest LIVE claim on a
// capacity-open run (the derived-capacity marker, REQ-005), keeps the hard
// bound of the operator-declared count (REQ-006, the t1294 contract), never
// grows for explicit requests (REQ-003), and draws from the one run-scoped
// pool shared across backends (REQ-007/008).

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// seedCapacityRun records a run row carrying the given lane capacity — the
// leader-start record the bounded claim reads (REQ-004).
func seedCapacityRun(t *testing.T, root, runID string, capacity int) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: runID, Backend: "codex", ManifestJSON: "{}",
		LeadPID: os.Getpid(), LeadProcessStart: "capacity-test-lead",
		LaneCapacity: capacity,
	}); err != nil {
		t.Fatalf("record run %s: %v", runID, err)
	}
}

func TestClaimFactoryLaneWithinGrowsCapacityOpen(t *testing.T) {
	alive := func(int) bool { return true }

	t.Run("automatic claim grows past the default bound", func(t *testing.T) {
		root := t.TempDir()
		seedCapacityRun(t, root, "tm3yoq", homestate.LaneCapacityDerived)
		// The claude-shape (unbounded) join holds lane-1 — the repro's live claim.
		first, err := ClaimFactoryLane(root, "", true, 999999, "tm3yoq", alive)
		if err != nil || first.Label != "lane-1" {
			t.Fatalf("claude-shape claim = (%q, %v), want lane-1", first.Label, err)
		}
		// The codex-shape bounded join (the launcher's 1..1 guess) still
		// attaches: the recorded derived marker makes the run capacity-open.
		claim, err := ClaimFactoryLaneWithin(root, "", true, os.Getpid(), "tm3yoq", 1, alive)
		if err != nil {
			t.Fatalf("bounded claim on capacity-open run: %v", err)
		}
		if claim.Label != "lane-2" {
			t.Fatalf("bounded claim label = %q, want lane-2 (grown past lane-1)", claim.Label)
		}
	})

	t.Run("explicit request never grows", func(t *testing.T) {
		root := t.TempDir()
		seedCapacityRun(t, root, "run-open-explicit", homestate.LaneCapacityDerived)
		if _, err := ClaimFactoryLane(root, "", true, 999999, "run-open-explicit", alive); err != nil {
			t.Fatal(err)
		}
		// REQ-003: the explicit range check rides the launcher-side bound and
		// never grows — lane-9 against 1..1 refuses even on a capacity-open run.
		_, err := ClaimFactoryLaneWithin(root, "lane-9", false, os.Getpid(), "run-open-explicit", 1, alive)
		if err == nil || !strings.Contains(err.Error(), "outside the allowed slots 1..1") {
			t.Fatalf("explicit lane-9 on capacity-open run: err=%v, want the out-of-range refusal naming 1..1", err)
		}
	})

	t.Run("growth targets one past the highest live claim", func(t *testing.T) {
		root := t.TempDir()
		seedCapacityRun(t, root, "run-deadgap", homestate.LaneCapacityDerived)
		// A live lane-1 and a DEAD lane-2: the dead claim is stale, so it is
		// neither counted nor refused — growth targets live claims only and
		// the freed lane-2 number is reused (§D.1 edge). The probe marks only
		// the seeded dead pid dead; every other claim stays live.
		dead := 999998
		if _, err := ClaimFactoryLane(root, "", true, 999999, "run-deadgap", alive); err != nil {
			t.Fatal(err)
		}
		if err := seedWorkerRowRun(t, root, "lane-2", dead, "run-deadgap"); err != nil {
			t.Fatal(err)
		}
		probe := func(pid int) bool { return pid != dead }
		claim, err := ClaimFactoryLaneWithin(root, "", true, os.Getpid(), "run-deadgap", 1, probe)
		if err != nil {
			t.Fatalf("bounded claim with dead lane-2: %v", err)
		}
		if claim.Label != "lane-2" {
			t.Fatalf("bounded claim label = %q, want lane-2 (dead claim freed, live lane-1 grown past)", claim.Label)
		}
	})
}

func TestClaimFactoryLaneWithinExplicitCapacityFull(t *testing.T) {
	alive := func(int) bool { return true }

	t.Run("full declared run refuses", func(t *testing.T) {
		root := t.TempDir()
		seedCapacityRun(t, root, "run-full01", 1)
		if _, err := ClaimFactoryLane(root, "", true, 999999, "run-full01", alive); err != nil {
			t.Fatal(err)
		}
		// REQ-006: the explicit operator-declared count keeps the hard bound —
		// the full run refuses instead of growing (the t1294 contract).
		_, err := ClaimFactoryLaneWithin(root, "", true, os.Getpid(), "run-full01", 1, alive)
		if err == nil || !strings.Contains(err.Error(), "has no free lane slots in 1..1") {
			t.Fatalf("auto claim on full explicit run: err=%v, want the no-free-slot refusal naming 1..1", err)
		}
	})

	t.Run("the recorded count overrides the launcher-side bound", func(t *testing.T) {
		root := t.TempDir()
		seedCapacityRun(t, root, "run-declared1b", 1)
		if _, err := ClaimFactoryLane(root, "", true, 999999, "run-declared1b", alive); err != nil {
			t.Fatal(err)
		}
		// The record is the authority (AC-006): a launcher-side bound of 3
		// cannot widen a run the operator declared as 1 slot.
		_, err := ClaimFactoryLaneWithin(root, "", true, os.Getpid(), "run-declared1b", 3, alive)
		if err == nil || !strings.Contains(err.Error(), "has no free lane slots in 1..1") {
			t.Fatalf("auto claim bound-3 on declared-1 run: err=%v, want the refusal naming the recorded 1..1", err)
		}
	})

	t.Run("explicit request within the recorded bound claims", func(t *testing.T) {
		root := t.TempDir()
		seedCapacityRun(t, root, "run-declared3b", 3)
		if _, err := ClaimFactoryLane(root, "", true, 999999, "run-declared3b", alive); err != nil {
			t.Fatal(err)
		}
		// lane-2 sits inside the recorded 1..3 bound even though the
		// launcher-side guess is narrower.
		claim, err := ClaimFactoryLaneWithin(root, "lane-2", false, os.Getpid(), "run-declared3b", 3, alive)
		if err != nil || claim.Label != "lane-2" {
			t.Fatalf("explicit lane-2 on declared-3 run = (%q, %v), want lane-2", claim.Label, err)
		}
	})
}

func TestClaimFactoryLaneWithinSharedPoolDistinctNumbers(t *testing.T) {
	alive := func(int) bool { return true }
	root := t.TempDir()
	seedCapacityRun(t, root, "run-pool01", homestate.LaneCapacityDerived)

	// The claude-shape join holds lane-1; the codex-shape bounded join draws
	// the next distinct number from the SAME run-scoped pool (REQ-007) under
	// the identical selection, growth, and refusal rules (REQ-008) — the
	// claim engine records no backend and applies no backend-specific rule.
	first, err := ClaimFactoryLane(root, "", true, 999999, "run-pool01", alive)
	if err != nil || first.Label != "lane-1" {
		t.Fatalf("claude-shape claim = (%q, %v), want lane-1", first.Label, err)
	}
	second, err := ClaimFactoryLaneWithin(root, "", true, os.Getpid(), "run-pool01", 1, alive)
	if err != nil {
		t.Fatalf("codex-shape bounded claim: %v", err)
	}
	if second.Label != "lane-2" {
		t.Fatalf("codex-shape claim label = %q, want lane-2 (distinct number, one pool)", second.Label)
	}
	reg := LoadFactoryRegistry(FactoryRegistryPath(root))
	if len(reg) != 2 {
		t.Fatalf("registry holds %d claims, want the two live lanes", len(reg))
	}
	if reg["lane-1"].PID != 999999 || reg["lane-2"].PID != os.Getpid() {
		t.Fatalf("registry = %v, want lane-1 and lane-2 both live on their pids", reg)
	}
}

// TestClaimFactoryLaneWithinConcurrentGrownRange extends the one-slot
// concurrency pin to the grown range (§D.1): concurrent bounded claims into
// a capacity-open run each take exactly one distinct number — none refuse,
// none duplicate.
func TestClaimFactoryLaneWithinConcurrentGrownRange(t *testing.T) {
	alive := func(int) bool { return true }
	root := t.TempDir()
	seedCapacityRun(t, root, "run-concurrent01", homestate.LaneCapacityDerived)
	if _, err := ClaimFactoryLane(root, "", true, 999999, "run-concurrent01", alive); err != nil {
		t.Fatal(err)
	}
	const joiners = 3
	labels := make(chan string, joiners)
	errs := make(chan error, joiners)
	var wg sync.WaitGroup
	for i := 0; i < joiners; i++ {
		wg.Add(1)
		go func(pid int) {
			defer wg.Done()
			claim, err := ClaimFactoryLaneWithin(root, "", true, pid, "run-concurrent01", 1, alive)
			if err == nil {
				labels <- claim.Label
			}
			errs <- err
		}(20000 + i)
	}
	wg.Wait()
	close(labels)
	close(errs)
	seen := map[string]bool{}
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent claim on capacity-open run refused: %v", err)
		}
	}
	for label := range labels {
		if seen[label] {
			t.Fatalf("duplicate concurrent claim %q", label)
		}
		seen[label] = true
	}
	if len(seen) != joiners {
		t.Fatalf("distinct labels=%d, want %d: %v", len(seen), joiners, seen)
	}
}
