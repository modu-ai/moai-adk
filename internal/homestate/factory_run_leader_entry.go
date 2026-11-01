package homestate

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"
)

// factory_run_leader_entry.go — the classification half of the leader-entry
// adoption (SPEC-FACTORY-COMPLETION-RECOVERY-001 REQ-FCR-017/018, card t1963):
// which active runs forbid an adoption, which are adoption candidates, and in
// what order the dead-stamp ones are chosen. The launcher decides what to do
// with the answer; this reads the database and writes nothing.

// ErrAdoptStampChanged marks an adoption that lost the owner-stamp
// compare-and-set: another leader restamped or retired the run between the
// classification and the write. The launcher tells it apart from a database
// failure — the loser mints a fresh run id instead of failing the launch.
var ErrAdoptStampChanged = errors.New("factory run adoption lost the owner-stamp compare-and-set")

// adoptStampChangedError keeps AdoptRun's original message and matches
// ErrAdoptStampChanged through errors.Is.
type adoptStampChangedError struct{ runID string }

func (e *adoptStampChangedError) Error() string {
	return fmt.Sprintf("factory run %s no longer carries the classified owner stamp; adoption refused", e.runID)
}

func (e *adoptStampChangedError) Is(target error) bool { return target == ErrAdoptStampChanged }

// LeaderEntryPlan is one classification of the active runs for a leader start.
type LeaderEntryPlan struct {
	// Blocked is true when any active run classifies OwnerLive or
	// OwnerIndeterminate: it may be a live leader (REQ-003b reads an unknown
	// owner as live), and adopting a dead run beside it would split the
	// factory. Blockers names those runs.
	Blocked  bool
	Blockers []RunOwner
	// Candidates are the active runs whose owner is dead or indeterminate and
	// which still carry live records — the set FindResumableRun lists. It is
	// what the launcher names in its warnings when it adopts nothing.
	Candidates []RunOwner
	// Stamped is the candidates classified dead on their OWN owner stamp,
	// newest first (created_at descending, then run_id descending). A run proved
	// dead through the lead peer or the boot proof is a candidate, never a
	// source: it carries no row-level stamp to compare-and-set against.
	Stamped []RunOwner
}

// PlanLeaderEntry classifies the active runs once and reports the leader-entry
// plan. It reads only.
func (f *FactoryDB) PlanLeaderEntry(ctx context.Context, opts ReconcileOptions) (LeaderEntryPlan, error) {
	owners, err := f.classifyRuns(ctx, opts, "active")
	if err != nil {
		return LeaderEntryPlan{}, err
	}
	probe := opts.workerProbe()
	var plan LeaderEntryPlan
	for _, o := range owners {
		switch o.Classification {
		case OwnerLive:
			plan.Blocked = true
			plan.Blockers = append(plan.Blockers, o)
			continue
		case OwnerIndeterminate:
			plan.Blocked = true
			plan.Blockers = append(plan.Blockers, o)
		case OwnerDead:
		default:
			continue
		}
		live, err := runHasLiveRecords(ctx, f.DB, o.RunID, probe)
		if err != nil {
			return LeaderEntryPlan{}, err
		}
		if !live {
			continue
		}
		plan.Candidates = append(plan.Candidates, o)
		if o.Classification == OwnerDead && o.Basis == BasisStamp {
			plan.Stamped = append(plan.Stamped, o)
		}
	}
	sort.SliceStable(plan.Stamped, func(i, j int) bool {
		ti, tj := runCreatedAt(plan.Stamped[i]), runCreatedAt(plan.Stamped[j])
		if !ti.Equal(tj) {
			return ti.After(tj)
		}
		return plan.Stamped[i].RunID > plan.Stamped[j].RunID
	})
	return plan, nil
}

// runCreatedAt parses a run's creation time. Stored times are RFC 3339 with a
// variable-length fraction, so text order is not time order; an unparseable
// value reads as the zero time (the oldest).
func runCreatedAt(o RunOwner) time.Time {
	at, err := time.Parse(time.RFC3339Nano, o.CreatedAt)
	if err != nil {
		return time.Time{}
	}
	return at
}
