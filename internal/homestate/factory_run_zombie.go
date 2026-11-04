package homestate

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// factory_run_zombie.go — the two writes that keep a dead leader's run from
// outliving it (Issue #1948 §8): its dead worker rows are removed with it, and
// its unfinished cards move to the one live run instead of pinning it active.

// reapDeadWorkers removes the run's worker rows whose pid probes positively
// dead and appends one worker.reaped event per row, inside the caller's
// transaction. A row the probe cannot decide stays (the safe direction).
func reapDeadWorkers(ctx context.Context, tx *sql.Tx, runID string, probe func(int) ProcessIdentityState, now time.Time) error {
	workers, err := workerPIDs(ctx, tx, runID)
	if err != nil {
		return err
	}
	for _, w := range workers {
		if probe(w.pid) != ProcessIdentityDead {
			continue
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM workers WHERE label=? AND run_id=? AND pid=?`, w.label, runID, w.pid); err != nil {
			return err
		}
		if err := appendEvent(ctx, tx, runID, "worker.reaped", map[string]any{"label": w.label, "pid": w.pid}, now); err != nil {
			return err
		}
	}
	return nil
}

// consolidateIntoLive retires a dead-owner run that still holds unfinished
// cards by moving them to the live run. A refusal (a live lease, a diverging
// row in the target, a changed stamp) leaves the run active and is not an
// error; only a failing store is.
func (f *FactoryDB) consolidateIntoLive(ctx context.Context, src, target RunOwner, opts ReconcileOptions) (bool, error) {
	if _, err := f.RebindRun(ctx, RebindRequest{Source: src, Target: target, RequireStampBasis: true}, opts); err != nil {
		if errors.Is(err, ErrRebindRefused) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
