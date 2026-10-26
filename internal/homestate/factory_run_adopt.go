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

// sqlTerminalStates renders the shared terminalStates set (card_record.go:
// done, failed, abandoned) as the SQL literal the live-records predicate
// uses. SQLite cannot bind a list, so the set is spelled here; the alignment
// test in factory_run_adopt_test.go pins this literal to terminalStates so
// the two cannot drift apart.
const sqlTerminalStates = `('done','failed','abandoned')`

// runHasLiveRecords reports whether the run still carries records a lifecycle
// decision must respect: cards outside the shared terminal set or registered
// workers that are not positively dead (card t1707, REQ-FCR-022). Live records
// are what separate "the leader died mid-run" from "this run is genuinely
// over".
//
// A worker row counts as live unless its pid probes positively dead; a probe
// that cannot decide counts as live — the safe direction, since a wrongly kept
// run has a working operator escape and a wrongly released one does not.
// `workers` carries no process-start fingerprint, so the probe is the pid
// alone; pid reuse can only make a dead worker look alive.
//
// q is the queryer seam so the check can run on a caller's transaction: the
// pool holds a single connection, and a read issued through the pool handle
// while a transaction holds it would deadlock. Every statement is drained
// before the next starts and before any pid is probed.
func runHasLiveRecords(ctx context.Context, q rowsRowQueryer, runID string, probe func(int) ProcessIdentityState) (bool, error) {
	var nonTerminal int
	if err := q.QueryRowContext(ctx, `SELECT EXISTS(
SELECT 1 FROM cards WHERE run_id=?1 AND state NOT IN `+sqlTerminalStates+`)`, runID).Scan(&nonTerminal); err != nil {
		return false, err
	}
	if nonTerminal == 1 {
		return true, nil
	}
	workers, err := workerPIDs(ctx, q, runID)
	if err != nil {
		return false, err
	}
	for _, w := range workers {
		if probe(w.pid) != ProcessIdentityDead {
			return true, nil
		}
	}
	return false, nil
}

// rowsRowQueryer is the read surface runHasLiveRecords needs: single-row and
// multi-row queries on the pool handle or on a caller's transaction.
type rowsRowQueryer interface {
	queryRower
	rowsQueryer
}

// FindResumableRun resolves the single run a resume-shaped leader start may
// re-adopt instead of minting a fresh run id (card t1707).
//
// A candidate is an ACTIVE run whose owner classifies dead or indeterminate
// and which still carries live records per runHasLiveRecords — the incident
// shape: the prior leader process died with the machine state intact, and the
// resuming leader is a NEW process (new pid, new fingerprint), so the
// recorded stamp classifies dead even though the run's records are alive.
//
// Adoption is admitted ONLY for a single candidate classified dead ON THE
// ROW'S OWN STAMP (sync-audit finding F2, card-review round 2): live records
// prove unfinished work, not the owner's death, and the batch-probe miss path
// classifies a probe-verified-live owner indeterminate — auto-adopting on
// indeterminate could overwrite a live leader's stamp. An unstamped legacy
// row classified through the leader-peer fallback carries no row-level
// identity to compare-and-set against, so it routes to the caller's
// warn-and-offer path too — the peer's identity is the broker's, not the
// row's. Absent-or-unproven owners therefore never auto-adopt, the same
// direction ClassifyOwnerWith gives every other absence of knowledge
// (REQ-003b).
//
// Exactly one candidate classified dead on its stamp → it is returned as
// adopt with the candidate list. Anything else — zero candidates (a cold
// start mints as before), a single indeterminate or peer-classified
// candidate, or several candidates (the caller warns and offers the
// deliberate --factory-run path) — returns a zero adopt with the full
// candidate list, which the caller distinguishes by length and
// classification.
//
// @MX:NOTE: [AUTO] only a single DEAD-owner candidate classified on its own stamp is returned as adopt; indeterminate and peer-classified candidates surface through the warn list
// @MX:REASON: auto-adopting an unproven or identity-less owner can overwrite a live leader's stamp; the asymmetric risk mirrors REQ-003b (sync-audit t1707 F2, card-review round 2)
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
		live, err := runHasLiveRecords(ctx, f.DB, o.RunID, opts.workerProbe())
		if err != nil {
			return RunOwner{}, nil, err
		}
		if live {
			candidates = append(candidates, o)
		}
	}
	if len(candidates) == 1 && candidates[0].Classification == OwnerDead && candidates[0].Basis == BasisStamp {
		return candidates[0], candidates, nil
	}
	return RunOwner{}, candidates, nil
}

// AdoptRun re-adopts one ACTIVE run for the resuming leader process: it
// restamps the owner identity to the supplied pid and process-start
// fingerprint and appends one run.resumed event carrying the adoption basis,
// so the audit trail records why the run id stayed the same across the leader
// process change (card t1707).
//
// adopt is the finder's classified candidate, and its recorded owner stamp is
// part of the write's WHERE clause: the restamp is a compare-and-set on the
// identity that was classified, so two resume launches racing on the same
// dead-owner run cannot both adopt — the second finds the stamp already
// replaced and is refused instead of overwriting the first adopter's live
// identity (sync-audit finding F1). The row must still measure active under
// that stamp; anything else is refused with an error, because re-activating a
// retired row is the ResumeRun discovery path's decision, not this one.
func (f *FactoryDB) AdoptRun(ctx context.Context, adopt RunOwner, pid int, processStart, basis string) error {
	if strings.TrimSpace(adopt.RunID) == "" {
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
	res, err := tx.ExecContext(ctx, `UPDATE runs SET lead_pid=?, lead_process_start=?, updated_at=? WHERE run_id=? AND status='active' AND lead_pid=? AND lead_process_start=?`,
		pid, processStart, now, adopt.RunID, adopt.LeadPID, adopt.LeadProcessStart)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("factory run %s no longer carries the classified owner stamp; adoption refused", adopt.RunID)
	}
	payload := fmt.Sprintf(`{"basis":%q,"outcome":"adopted"}`, basis)
	if _, err := tx.ExecContext(ctx, `INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,'run.resumed',?,?)`, adopt.RunID, payload, now); err != nil {
		return err
	}
	return tx.Commit()
}
