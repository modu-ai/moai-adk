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
// workers (card t1707). Live records are what separate "the leader died
// mid-run" from "this run is genuinely over".
//
// q is the queryer seam (card_record.go queryRower) so the check can run on a
// caller's transaction: the pool holds a single connection, and a read issued
// through the pool handle while a transaction holds it would deadlock.
func runHasLiveRecords(ctx context.Context, q queryRower, runID string) (bool, error) {
	var live int
	if err := q.QueryRowContext(ctx, `SELECT EXISTS(
SELECT 1 FROM cards WHERE run_id=?1 AND state NOT IN `+sqlTerminalStates+`)
OR EXISTS(SELECT 1 FROM workers WHERE run_id=?1)`, runID).Scan(&live); err != nil {
		return false, err
	}
	return live == 1, nil
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
		live, err := runHasLiveRecords(ctx, f.DB, o.RunID)
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

// OrphanRun is one orphan run — an ACTIVE run a fresh leader start would
// strand — together with the count of its non-terminal cards, the number the
// leader-start guidance renders (SPEC-FACTORY-LEADER-ADOPT-CORE-001).
type OrphanRun struct {
	RunOwner
	// NonTerminalCards counts the run's cards outside the shared terminal
	// set (sqlTerminalStates). The count is what makes the run an orphan:
	// a dead-owner run with only terminal cards and no workers is finished,
	// and the lane-join reconciliation is free to retire it.
	NonTerminalCards int
}

// FindOrphanRuns returns every orphan run — an ACTIVE run whose owner is
// classified dead ON THE RUN ROW'S OWN STAMP (BasisStamp, the same admit
// rule as FindResumableRun) and which holds at least one non-terminal card —
// with each run's non-terminal card count. A peer-classified run (BasisPeer)
// or a boot-proof run (BasisBoot) carries no row-level identity to
// compare-and-set against, so neither is ever an orphan; live and
// indeterminate owners are not; a workers-only run holds no card to orphan.
//
// The finder is SELECT-only: it classifies the rows and counts cards and
// writes nothing (AC-FLAC-001's content-snapshot guarantee).
//
// @MX:NOTE: [AUTO] the orphan admit rule is Classification==OwnerDead && Basis==BasisStamp && NonTerminalCards>=1 — the stamp basis is the identity the adoption's compare-and-set needs
// @MX:REASON: a boot-proof or peer-classified run has no row-level owner identity, so an adopt keyed on it could not refuse a competing writer (SPEC-FACTORY-LEADER-ADOPT-CORE-001 REQ-FLAC-001)
func (f *FactoryDB) FindOrphanRuns(ctx context.Context, opts ReconcileOptions) ([]OrphanRun, error) {
	owners, err := f.classifyRuns(ctx, opts, "active")
	if err != nil {
		return nil, err
	}
	var orphans []OrphanRun
	for _, o := range owners {
		if o.Classification != OwnerDead || o.Basis != BasisStamp {
			continue
		}
		// The count is a read on the pool handle after classifyRuns drained
		// its row cursor, so the single-connection pool cannot deadlock.
		var nonTerminal int
		if err := f.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM cards WHERE run_id=?1 AND state NOT IN `+sqlTerminalStates, o.RunID).Scan(&nonTerminal); err != nil {
			return nil, err
		}
		if nonTerminal < 1 {
			continue
		}
		orphans = append(orphans, OrphanRun{RunOwner: o, NonTerminalCards: nonTerminal})
	}
	return orphans, nil
}

// FindRunOwner classifies the ONE named run, whatever its classification —
// the deliberate --factory-run adopt classifies the run it is pointed at; it
// does not require an orphan (SPEC-FACTORY-LEADER-ADOPT-CORE-001 M2). A run
// id the database does not hold is an error.
func (f *FactoryDB) FindRunOwner(ctx context.Context, opts ReconcileOptions, runID string) (RunOwner, error) {
	owners, err := f.classifyRuns(ctx, opts, "")
	if err != nil {
		return RunOwner{}, err
	}
	for _, o := range owners {
		if o.RunID == runID {
			return o, nil
		}
	}
	return RunOwner{}, fmt.Errorf("factory run %s not found", runID)
}

// AdoptRun re-adopts one ACTIVE run for the resuming leader process: it
// restamps the owner identity to the supplied pid and process-start
// fingerprint, writes the adopting leader's backend on the row, and appends
// one run.resumed event carrying the adoption basis, so the audit trail
// records why the run id stayed the same across the leader process change
// (card t1707; the backend write is SPEC-FACTORY-LEADER-ADOPT-CORE-001's
// plan-audit D1 — a takeover that skipped it would leave the prior leader's
// backend on the row).
//
// adopt is the finder's classified candidate, and its recorded owner stamp
// AND backend are part of the write's WHERE clause: the restamp is a
// compare-and-set on the identity that was classified, so two resume
// launches racing on the same dead-owner run cannot both adopt — the second
// finds the stamp already replaced and is refused instead of overwriting the
// first adopter's live identity (sync-audit finding F1) — and a backend flip
// landing after the classified read (the codex-led refusal's own axis,
// invisible to a stamp-only CAS) fails the write the same way (sync-audit
// round 3 F1'). The row must still measure active under that stamp;
// anything else is refused with an error, because re-activating a retired
// row is the ResumeRun discovery path's decision, not this one.
//
// peerGuard, when non-nil, is invoked INSIDE the adoption transaction —
// after the write and before the commit — so a refusal it returns rolls the
// write back. It carries the legacy-peer broker check for callers whose
// guard cannot join this transaction (the broker lives in a separate
// factorymsg database and the factorymsg import direction is fixed): the
// window it closes is everything up to the commit, the narrowest a
// cross-database check allows (sync-audit round 3 F1').
func (f *FactoryDB) AdoptRun(ctx context.Context, adopt RunOwner, pid int, processStart, backend, basis string, peerGuard func(context.Context) error) error {
	if strings.TrimSpace(adopt.RunID) == "" {
		return errors.New("factory run id is empty")
	}
	if pid < 1 || strings.TrimSpace(processStart) == "" {
		return errors.New("factory run adoption requires the resuming owner identity")
	}
	if strings.TrimSpace(backend) == "" {
		return errors.New("factory run adoption requires the adopting leader's backend")
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
	res, err := tx.ExecContext(ctx, `UPDATE runs SET lead_pid=?, lead_process_start=?, lead_backend=?, updated_at=? WHERE run_id=? AND status='active' AND lead_pid=? AND lead_process_start=? AND lead_backend=?`,
		pid, processStart, backend, now, adopt.RunID, adopt.LeadPID, adopt.LeadProcessStart, adopt.Backend)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("factory run %s no longer carries the classified owner stamp and backend; adoption refused", adopt.RunID)
	}
	payload := fmt.Sprintf(`{"basis":%q,"outcome":"adopted"}`, basis)
	if _, err := tx.ExecContext(ctx, `INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,'run.resumed',?,?)`, adopt.RunID, payload, now); err != nil {
		return err
	}
	if peerGuard != nil {
		if err := peerGuard(ctx); err != nil {
			return err
		}
	}
	return tx.Commit()
}
