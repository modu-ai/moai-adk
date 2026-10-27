package factory

// factory_slots_consolidated_test.go — CARRY-6 (card t1963, ledger L15): a lane
// worker registration must not attach to a run a consolidation retired. The
// refusal sits inside the claim transaction, before the workers INSERT, and
// wraps homestate.ErrRunRetiredByConsolidation. A claim on any other run state
// (active, ordinarily retired, absent, no run id) behaves exactly as it did
// before the guard: the controls pin that.

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// seedConsolidatedSource records run-old and run-new, retires run-old and marks
// the retirement the way a consolidation does: status 'retired' plus a
// run.retired event carrying rebound_to (factory_run_consolidated.go).
func seedConsolidatedSource(t *testing.T, root string) {
	t.Helper()
	seedCapacityRun(t, root, "run-old", homestate.LaneCapacityDerived)
	seedCapacityRun(t, root, "run-new", homestate.LaneCapacityDerived)
	retireSeededRun(t, root, "run-old", `{"rebound_to":"run-new"}`)
}

// retireSeededRun retires runID with the given run.retired event payload.
func retireSeededRun(t *testing.T, root, runID, payload string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`UPDATE runs SET status='retired' WHERE run_id=?`, runID); err != nil {
		t.Fatalf("retire %s: %v", runID, err)
	}
	if _, err := db.DB.Exec(`INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,'run.retired',?,?)`,
		runID, payload, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatalf("record retirement of %s: %v", runID, err)
	}
}

// workerRowCount counts the workers rows registered for runID.
func workerRowCount(t *testing.T, root, runID string) int {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM workers WHERE run_id=?`, runID).Scan(&n); err != nil {
		t.Fatalf("count workers for %s: %v", runID, err)
	}
	return n
}

func TestClaimFactoryLaneRefusesAConsolidatedRun(t *testing.T) {
	alive := func(int) bool { return true }
	pid := os.Getpid()

	claims := []struct {
		name  string
		claim func(root string) (FactoryClaim, error)
	}{
		{"automatic unbounded", func(root string) (FactoryClaim, error) {
			return ClaimFactoryLane(root, "", true, pid, "run-old", alive)
		}},
		{"explicit label", func(root string) (FactoryClaim, error) {
			return ClaimFactoryLane(root, "lane-1", false, pid, "run-old", alive)
		}},
		{"bounded automatic", func(root string) (FactoryClaim, error) {
			return ClaimFactoryLaneWithin(root, "", true, pid, "run-old", 2, alive)
		}},
		{"with backend", func(root string) (FactoryClaim, error) {
			return ClaimFactoryLaneWithinWithBackend(root, "", true, pid, "run-old", 2, "codex", alive)
		}},
	}
	for _, c := range claims {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			seedConsolidatedSource(t, root)
			claim, err := c.claim(root)
			if !errors.Is(err, homestate.ErrRunRetiredByConsolidation) {
				t.Fatalf("claim on the consolidated run: label=%q err=%v, want homestate.ErrRunRetiredByConsolidation", claim.Label, err)
			}
			if n := workerRowCount(t, root, "run-old"); n != 0 {
				t.Fatalf("a refused claim left %d workers row(s) for the consolidated run", n)
			}
		})
	}
}

// The behaviour of every other run state is unchanged: the guard refuses only a
// consolidated source.
func TestClaimFactoryLaneConsolidationGuardLeavesOtherRunsAlone(t *testing.T) {
	alive := func(int) bool { return true }
	pid := os.Getpid()

	t.Run("active run", func(t *testing.T) {
		root := t.TempDir()
		seedCapacityRun(t, root, "run-active", homestate.LaneCapacityDerived)
		claim, err := ClaimFactoryLane(root, "", true, pid, "run-active", alive)
		if err != nil || claim.Label != "lane-1" {
			t.Fatalf("claim on an active run = (%q, %v), want lane-1", claim.Label, err)
		}
		if n := workerRowCount(t, root, "run-active"); n != 1 {
			t.Fatalf("workers rows for the active run = %d, want 1", n)
		}
	})

	t.Run("ordinarily retired run", func(t *testing.T) {
		root := t.TempDir()
		seedCapacityRun(t, root, "run-plain", homestate.LaneCapacityDerived)
		retireSeededRun(t, root, "run-plain", `{}`) // an ordinary retirement carries no rebound_to
		claim, err := ClaimFactoryLane(root, "", true, pid, "run-plain", alive)
		if err != nil || claim.Label != "lane-1" {
			t.Fatalf("claim on an ordinarily retired run = (%q, %v), want lane-1 (behaviour before the guard)", claim.Label, err)
		}
		if n := workerRowCount(t, root, "run-plain"); n != 1 {
			t.Fatalf("workers rows for the ordinarily retired run = %d, want 1", n)
		}
	})

	t.Run("run absent from the store", func(t *testing.T) {
		root := t.TempDir()
		claim, err := ClaimFactoryLane(root, "", true, pid, "run-missing", alive)
		if err != nil || claim.Label != "lane-1" {
			t.Fatalf("claim on an absent run = (%q, %v), want lane-1", claim.Label, err)
		}
	})

	t.Run("no run id", func(t *testing.T) {
		root := t.TempDir()
		claim, err := ClaimFactoryLane(root, "", true, pid, "", alive)
		if err != nil || claim.Label != "lane-1" {
			t.Fatalf("claim with no run id = (%q, %v), want lane-1", claim.Label, err)
		}
	})
}
