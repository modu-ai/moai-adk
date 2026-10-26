package homestate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// factory_run_rebind.go — the dead-owner run consolidation operation
// (SPEC-FACTORY-COMPLETION-RECOVERY-001 M2-B, REQ-FCR-020/023, card t1963).
//
// A leader restart gives the new process a new owner stamp, so the previous
// run is structurally dead while its cards are still in flight. The operation
// moves those cards into a live run and retires the dead one, in ONE write
// transaction: a partial move would leave a half-moved run behind.

// ErrRebindRefused marks a refused run consolidation: one of the operation's
// guards declined, and nothing was written.
var ErrRebindRefused = errors.New("factory run consolidation refused")

// RebindRequest names the two runs of a consolidation, each as the caller
// classified it. RequireStampBasis is the leader-entry path's extra condition:
// the source must have been classified dead on its own stamp, never through
// the peer or boot proof.
type RebindRequest struct {
	Source            RunOwner
	Target            RunOwner
	RequireStampBasis bool
}

// RebindResult lists the card ids the consolidation re-keyed and the ones it
// abandoned with a pointer (the target already held the card).
type RebindResult struct {
	Moved     []string
	Abandoned []string
}

// The injection points of the consolidation transaction. rebindFault is the
// test seam: nil in production, and a test makes it return an error at a named
// point to prove the whole operation rolls back from there.
const (
	rebindPointSecondCard   = "second-card"   // right after the second card was processed
	rebindPointAfterCards   = "after-cards"   // every card moved, binding not yet repointed
	rebindPointBeforeCommit = "before-commit" // every write done, commit not yet issued
)

var rebindFault func(point string) error

func rebindInject(point string) error {
	if rebindFault == nil {
		return nil
	}
	return rebindFault(point)
}

func rebindRefusal(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrRebindRefused}, args...)...)
}

// consolidationCard is the slice of a source card row the operation needs.
type consolidationCard struct {
	id, state, stage, evidenceSHA string
	version                       int64
}

// RebindRun consolidates the source run into the target run (REQ-FCR-020).
// Source and target are the callers' classifications; the operation never
// trusts them beyond what the transaction re-reads — it compares the owner
// stamp each was classified on.
//
// Inside one BEGIN IMMEDIATE transaction it
//
//	(0) refuses unless the source owner classified OwnerDead (on the leader-entry
//	    path additionally on its own stamp),
//	(1) re-reads the source: still active under the classified stamp,
//	(2) re-reads the target: active, stamp unchanged, classified OwnerLive,
//	(3) refuses a source that holds a worker not positively dead,
//	(4) refuses when any third active run whose owner is not dead holds a
//	    non-terminal row for one of the source's cards; otherwise re-keys each
//	    non-terminal source card to the target, or — when the target already
//	    holds the card — moves the source row to abandoned with a
//	    card.rebound pointer event,
//	(5) repoints every card_dispatch binding that names the source,
//	(6) flips the source to retired and appends one run.retired event.
//
// Any refusal and any failure rolls everything back. No row is deleted, and
// leader_approvals is never touched (REQ-FCR-023): a re-keyed card needs a
// fresh approval under the target run.
//
// @MX:ANCHOR: [AUTO] the single consolidation operation behind leader entry, `runs --retire --rebind-to` and the join prompt's m<n>
// @MX:REASON: three operator-facing surfaces call it (fan_in >= 3); a guard dropped here is dropped for all of them, and a wrong move strands or steals a live run's cards
// @MX:WARN: [AUTO] moves rows between runs inside one write transaction; every guard must precede its first write
// @MX:REASON: a partial move leaves a half-consolidated run that neither the source's owner nor the target can finish (REQ-FCR-020)
func (f *FactoryDB) RebindRun(ctx context.Context, req RebindRequest, opts ReconcileOptions) (RebindResult, error) {
	src, tgt := req.Source, req.Target
	if strings.TrimSpace(src.RunID) == "" || strings.TrimSpace(tgt.RunID) == "" || src.RunID == tgt.RunID {
		return RebindResult{}, rebindRefusal("source %q and target %q must be two distinct runs", src.RunID, tgt.RunID)
	}
	// Step 0 — the source owner's death is the premise of the whole operation.
	if !retirable(src.Classification) {
		return RebindResult{}, fmt.Errorf("%w: %w: run %s owner classified %s", ErrRebindRefused, ErrRunOwnerNotDead, src.RunID, src.Classification)
	}
	if req.RequireStampBasis && src.Basis != BasisStamp {
		return RebindResult{}, fmt.Errorf("%w: %w: run %s owner classified %s on basis %q, not on its own stamp",
			ErrRebindRefused, ErrRunOwnerNotDead, src.RunID, src.Classification, src.Basis)
	}
	if tgt.Classification != OwnerLive {
		return RebindResult{}, rebindRefusal("target run %s owner classified %s, a consolidation target must be live", tgt.RunID, tgt.Classification)
	}
	// The other active runs are classified before the transaction opens: the
	// pool holds one connection, and classification issues queries of its own.
	active, err := f.classifyRuns(ctx, opts, "active")
	if err != nil {
		return RebindResult{}, err
	}
	others := make(map[string]RunOwner, len(active))
	for _, o := range active {
		if o.RunID != src.RunID && o.RunID != tgt.RunID {
			others[o.RunID] = o
		}
	}

	now := time.Now().UTC()
	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return RebindResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := requireRunStamp(ctx, tx, src, "source"); err != nil {
		return RebindResult{}, err
	}
	if err := requireRunStamp(ctx, tx, tgt, "target"); err != nil {
		return RebindResult{}, err
	}
	if err := refuseLiveWorker(ctx, tx, src.RunID, opts.workerProbe()); err != nil {
		return RebindResult{}, err
	}
	cards, err := sourceCards(ctx, tx, src.RunID)
	if err != nil {
		return RebindResult{}, err
	}
	for _, c := range cards {
		if err := refuseThirdRunHolder(ctx, tx, c.id, src.RunID, tgt.RunID, others); err != nil {
			return RebindResult{}, err
		}
	}

	var res RebindResult
	nowText := now.Format(time.RFC3339Nano)
	for i, c := range cards {
		var held int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM cards WHERE run_id=? AND card_id=?`, tgt.RunID, c.id).Scan(&held); err != nil {
			return RebindResult{}, err
		}
		if held > 0 {
			if err := abandonWithPointer(ctx, tx, src.RunID, tgt.RunID, c, "consolidation", now); err != nil {
				return RebindResult{}, err
			}
			res.Abandoned = append(res.Abandoned, c.id)
		} else {
			// Re-key: only run_id changes; state, version, stage, lease, evidence
			// and every other column ride along untouched.
			r, err := tx.ExecContext(ctx, `UPDATE cards SET run_id=? WHERE run_id=? AND card_id=?`, tgt.RunID, src.RunID, c.id)
			if err != nil {
				return RebindResult{}, err
			}
			if n, err := r.RowsAffected(); err != nil || n != 1 {
				return RebindResult{}, rebindRefusal("card %s of run %s changed during the consolidation", c.id, src.RunID)
			}
			res.Moved = append(res.Moved, c.id)
		}
		if i == 1 {
			if err := rebindInject(rebindPointSecondCard); err != nil {
				return RebindResult{}, err
			}
		}
	}
	if err := rebindInject(rebindPointAfterCards); err != nil {
		return RebindResult{}, err
	}
	// Step 5 — every binding that names the source, the abandoned cards' too.
	if _, err := tx.ExecContext(ctx, `UPDATE card_dispatch SET run_id=?, recorded_at=? WHERE run_id=?`, tgt.RunID, nowText, src.RunID); err != nil {
		return RebindResult{}, err
	}
	// Step 6 — flip the source under the classified stamp, then record why.
	flip, err := tx.ExecContext(ctx, `UPDATE runs SET status='retired', updated_at=? WHERE run_id=? AND status='active' AND lead_pid=? AND lead_process_start=?`,
		nowText, src.RunID, src.LeadPID, src.LeadProcessStart)
	if err != nil {
		return RebindResult{}, err
	}
	if n, err := flip.RowsAffected(); err != nil || n != 1 {
		return RebindResult{}, rebindRefusal("run %s changed owner or status since classification", src.RunID)
	}
	if err := appendEvent(ctx, tx, src.RunID, "run.retired", map[string]any{
		"classification":  string(src.Classification),
		"basis":           string(src.Basis),
		"rebound_to":      tgt.RunID,
		"moved":           len(res.Moved),
		"abandoned":       len(res.Abandoned),
		"moved_cards":     nonNilStrings(res.Moved),
		"abandoned_cards": nonNilStrings(res.Abandoned),
	}, now); err != nil {
		return RebindResult{}, err
	}
	if err := rebindInject(rebindPointBeforeCommit); err != nil {
		return RebindResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return RebindResult{}, err
	}
	return res, nil
}

func nonNilStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// requireRunStamp re-reads one run inside the transaction and refuses unless
// it is still active under the owner stamp the caller classified it on.
func requireRunStamp(ctx context.Context, tx *sql.Tx, o RunOwner, role string) error {
	var status, start string
	var pid int
	err := tx.QueryRowContext(ctx, `SELECT status, lead_pid, lead_process_start FROM runs WHERE run_id=?`, o.RunID).Scan(&status, &pid, &start)
	if errors.Is(err, sql.ErrNoRows) {
		return rebindRefusal("%s run %s does not exist", role, o.RunID)
	}
	if err != nil {
		return err
	}
	if status != "active" || pid != o.LeadPID || start != o.LeadProcessStart {
		return rebindRefusal("%s run %s changed owner or status since classification (status %s)", role, o.RunID, status)
	}
	return nil
}

// refuseLiveWorker refuses when ANY worker row of the run is not positively
// dead (an undecided probe counts as live). Every row is probed — a check of
// only the first or only the last row would let a live lane slip through.
func refuseLiveWorker(ctx context.Context, tx *sql.Tx, runID string, probe func(int) ProcessIdentityState) error {
	rows, err := tx.QueryContext(ctx, `SELECT label, pid FROM workers WHERE run_id=? ORDER BY rowid`, runID)
	if err != nil {
		return err
	}
	type worker struct {
		label string
		pid   int
	}
	var workers []worker
	for rows.Next() {
		var w worker
		if err := rows.Scan(&w.label, &w.pid); err != nil {
			_ = rows.Close()
			return err
		}
		workers = append(workers, w)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, w := range workers {
		if probe(w.pid) != ProcessIdentityDead {
			return rebindRefusal("run %s still has live worker %q (pid %d)", runID, w.label, w.pid)
		}
	}
	return nil
}

// sourceCards reads the source's non-terminal cards in card-id order.
func sourceCards(ctx context.Context, tx *sql.Tx, runID string) ([]consolidationCard, error) {
	// SQL: sqlTerminalStates is a compile-time constant; the run id goes through a placeholder.
	rows, err := tx.QueryContext(ctx, `SELECT card_id,state,version,stage,evidence_sha FROM cards WHERE run_id=? AND state NOT IN `+sqlTerminalStates+` ORDER BY card_id`, runID)
	if err != nil {
		return nil, err
	}
	var cards []consolidationCard
	for rows.Next() {
		var c consolidationCard
		if err := rows.Scan(&c.id, &c.state, &c.version, &c.stage, &c.evidenceSHA); err != nil {
			_ = rows.Close()
			return nil, err
		}
		cards = append(cards, c)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	return cards, rows.Close()
}

// refuseThirdRunHolder refuses when a third ACTIVE run, whose owner is not
// dead, holds a non-terminal row for cardID: that run may be working the card
// right now, and moving it from under the live holder would split the record.
// A third run absent from the pre-transaction classification appeared after
// it and counts as not dead; one that classified dead must still carry the
// stamp it was classified on.
func refuseThirdRunHolder(ctx context.Context, tx *sql.Tx, cardID, srcID, tgtID string, others map[string]RunOwner) error {
	// SQL: sqlTerminalStates is a compile-time constant; every value goes through a placeholder.
	rows, err := tx.QueryContext(ctx, `SELECT c.run_id, r.lead_pid, r.lead_process_start FROM cards c JOIN runs r ON r.run_id=c.run_id
WHERE c.card_id=? AND c.run_id NOT IN (?,?) AND r.status='active' AND c.state NOT IN `+sqlTerminalStates+` ORDER BY c.run_id`, cardID, srcID, tgtID)
	if err != nil {
		return err
	}
	type holder struct {
		run   string
		pid   int
		start string
	}
	var holders []holder
	for rows.Next() {
		var h holder
		if err := rows.Scan(&h.run, &h.pid, &h.start); err != nil {
			_ = rows.Close()
			return err
		}
		holders = append(holders, h)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, h := range holders {
		o, known := others[h.run]
		if known && retirable(o.Classification) && o.LeadPID == h.pid && o.LeadProcessStart == h.start {
			continue
		}
		class := "unclassified"
		if known {
			class = string(o.Classification)
		}
		return rebindRefusal("card %s is also held by run %s (owner %s); a card another live run holds is not moved", cardID, h.run, class)
	}
	return nil
}

// abandonWithPointer moves one non-terminal row of fromRun to abandoned and
// appends the card.rebound event that records where the card went and what the
// row looked like — nothing is deleted, and the previous state is recoverable
// from the event. The update is version-checked like every other transition.
func abandonWithPointer(ctx context.Context, tx *sql.Tx, fromRun, toRun string, c consolidationCard, via string, now time.Time) error {
	r, err := tx.ExecContext(ctx, `UPDATE cards SET state=?, version=version+1, updated_at=? WHERE run_id=? AND card_id=? AND version=?`,
		CardAbandoned, now.Format(time.RFC3339Nano), fromRun, c.id, c.version)
	if err != nil {
		return err
	}
	if n, err := r.RowsAffected(); err != nil || n != 1 {
		return rebindRefusal("card %s of run %s changed under the abandon", c.id, fromRun)
	}
	return appendEvent(ctx, tx, fromRun, "card.rebound", map[string]any{
		"card_id":      c.id,
		"from_run":     fromRun,
		"to_run":       toRun,
		"state":        c.state,
		"version":      c.version,
		"stage":        c.stage,
		"evidence_sha": c.evidenceSHA,
		"via":          via,
	}, now)
}
