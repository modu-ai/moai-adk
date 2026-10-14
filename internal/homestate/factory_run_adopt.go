package homestate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// factory_run_adopt.go — the run re-adoption half of the leader-resume fix
// (card t1707): the finder that resolves the one run a resume-shaped leader
// start may re-adopt, the adoption write that keeps the prior run id, and the
// shared live-records predicate both this path and the retirement guard read.

// runHasLiveRecords reports whether the run still carries records a lifecycle
// decision must respect: cards in a non-terminal state (anything but done and
// abandoned) or registered workers (card t1707). Live records are what
// separate "the leader died mid-run" from "this run is genuinely over".
func (f *FactoryDB) runHasLiveRecords(ctx context.Context, runID string) (bool, error) {
	var live int
	if err := f.DB.QueryRowContext(ctx, `SELECT EXISTS(
SELECT 1 FROM cards WHERE run_id=?1 AND state NOT IN ('done','abandoned'))
OR EXISTS(SELECT 1 FROM workers WHERE run_id=?1)`, runID).Scan(&live); err != nil {
		return false, err
	}
	return live == 1, nil
}

// FindResumableRun resolves the single run a resume-shaped leader start may
// re-adopt instead of minting a fresh run id (card t1707).
//
// A candidate is an ACTIVE run whose owner classifies dead or indeterminate —
// the incident shape: the prior leader process died with the machine state
// intact, and the resuming leader is a NEW process (new pid, new fingerprint),
// so the recorded stamp classifies dead even though the run's records are
// alive — and which still carries live records per runHasLiveRecords.
//
// Exactly one candidate → it is returned as adopt with the candidate list.
// Zero candidates (nothing to adopt — a cold start mints as before) and more
// than one (ambiguous — the caller warns and offers the deliberate
// --factory-run path) return a zero adopt with the full candidate list, which
// the caller distinguishes by length.
//
// @MX:NOTE: [AUTO] candidate admission is dead-or-indeterminate plus live records; a live owner's run is never adoptable
// @MX:REASON: adopting a live leader's run would hijack it, and adoption is the alternative to minting over a run's records (card t1707)
func (f *FactoryDB) FindResumableRun(ctx context.Context, opts ReconcileOptions) (RunOwner, []RunOwner, error) {
	owners, err := f.classifyRuns(ctx, opts, "active")
	if err != nil {
		return RunOwner{}, nil, err
	}
	var candidates []RunOwner
	for _, o := range owners {
		switch o.Classification {
		case OwnerDead, OwnerIndeterminate:
		default:
			continue
		}
		live, err := f.runHasLiveRecords(ctx, o.RunID)
		if err != nil {
			return RunOwner{}, nil, err
		}
		if live {
			candidates = append(candidates, o)
		}
	}
	if len(candidates) == 1 {
		return candidates[0], candidates, nil
	}
	return RunOwner{}, candidates, nil
}

// AdoptRun re-adopts one ACTIVE run for the resuming leader process: it
// restamps the owner identity to the supplied pid and process-start
// fingerprint and appends one run.resumed event carrying the adoption basis,
// so the audit trail records why the run id stayed the same across the leader
// process change (card t1707). The row must measure active: a retired run is
// refused with an error, because re-activating a retired row is the ResumeRun
// discovery path's decision, not this one.
func (f *FactoryDB) AdoptRun(ctx context.Context, runID string, pid int, processStart, basis string) error {
	if strings.TrimSpace(runID) == "" {
		return errors.New("factory run id is empty")
	}
	if pid < 1 || strings.TrimSpace(processStart) == "" {
		return errors.New("factory run adoption requires the resuming owner identity")
	}
	if strings.TrimSpace(basis) == "" {
		return errors.New("factory run adoption requires a basis")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `UPDATE runs SET lead_pid=?, lead_process_start=?, updated_at=? WHERE run_id=? AND status='active'`, pid, processStart, now, runID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("factory run %s is not active; adoption refused", runID)
	}
	payload := fmt.Sprintf(`{"basis":%q,"outcome":"adopted"}`, basis)
	if _, err := tx.ExecContext(ctx, `INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,'run.resumed',?,?)`, runID, payload, now); err != nil {
		return err
	}
	return tx.Commit()
}
