package homestate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
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
	leaseHolder, leaseExpiresAt   string
	worktreePath, specID          string
	version                       int64
}

// strandsProgressOver names the progress fields (stage, worktree path, spec
// id) the source row carries that the target row does not already hold
// identically — the info an abandon-with-pointer would strand on the abandoned
// row while the target's fresh row continues without it. Round 3 keyed this
// decision on the source row's STATE; that missed rows a lease expiry or a
// question receipt returned to assigned/needs-decision KEEPING their progress
// (card t1973 round 4), so the ACTUAL remaining info decides now. An empty
// return means the source row is a bare stub, or the target already carries
// everything it holds.
func (c consolidationCard) strandsProgressOver(tgtStage, tgtWorktree, tgtSpec string) []string {
	var stranded []string
	if c.stage != "" && c.stage != tgtStage {
		stranded = append(stranded, "stage")
	}
	if c.worktreePath != "" && c.worktreePath != tgtWorktree {
		stranded = append(stranded, "worktree")
	}
	if c.specID != "" && c.specID != tgtSpec {
		stranded = append(stranded, "spec_id")
	}
	return stranded
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
//	    non-terminal row for one of the source's cards (a dead-owner third run
//	    must also provably hold no live lease and no living worker); otherwise
//	    re-keys each non-terminal source card to the target, or — when the
//	    target already holds the card as a NON-TERMINAL row with the same
//	    evidence — moves the source row to abandoned with a card.rebound
//	    pointer event (a terminal or evidence-divergent target row refuses
//	    the whole operation, card t1973),
//	(5) repoints the card_dispatch binding of each card step 4 moved (re-keyed or
//	    abandoned with a pointer) — not those of terminal cards left in the source,
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
		switch o.RunID {
		case src.RunID:
			// Step 1, re-classification: a source proved dead through the lead
			// peer or the boot proof has no row-level stamp ((0,'')) for the
			// transaction to compare, so this fresh classification — taken just
			// before BEGIN — is the only check that its owner is still dead.
			// The window is narrowed, not closed: the peer identity lives in the
			// broker database, outside this transaction.
			if (src.Basis == BasisPeer || src.Basis == BasisBoot) && !retirable(o.Classification) {
				return RebindResult{}, fmt.Errorf("%w: %w: run %s owner classified %s on re-classification before the transaction (it was %s on basis %q)",
					ErrRebindRefused, ErrRunOwnerNotDead, src.RunID, o.Classification, src.Classification, src.Basis)
			}
		case tgt.RunID:
			// The target's fresh classification is applied, not discarded: a
			// dead owner leaves the row's stamp unchanged, so requireRunStamp
			// inside the transaction cannot see it (REQ-FCR-020, step 2).
			if o.Classification != OwnerLive {
				return RebindResult{}, rebindRefusal("target run %s owner classified %s on re-classification before the transaction, a consolidation target must be live", tgt.RunID, o.Classification)
			}
		default:
			others[o.RunID] = o
		}
	}
	// A target missing from the active set (retired or deleted since
	// classification) is refused by requireRunStamp inside the transaction.

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
		// The source leader's death does not kill a card lease, and the
		// worker guard is run-scoped: a holder whose registration names
		// another run (or none) is invisible to refuseLiveWorker, so the
		// row's own lease decides (card t1973 round 2, P1).
		if holdsValidLease(Card{State: c.state, LeaseHolder: c.leaseHolder, LeaseExpiresAt: c.leaseExpiresAt}, now) {
			return RebindResult{}, rebindRefusal("card %s of run %s holds a live lease (%s, expires %s); a leased card is not moved",
				c.id, src.RunID, c.leaseHolder, c.leaseExpiresAt)
		}
		if err := refuseThirdRunHolder(ctx, tx, c.id, src.RunID, tgt.RunID, others, now, opts.workerProbe()); err != nil {
			return RebindResult{}, err
		}
	}

	var res RebindResult
	nowText := now.Format(time.RFC3339Nano)
	for i, c := range cards {
		var tgtState, tgtEvidence, tgtStage, tgtWorktree, tgtSpec string
		tgtErr := tx.QueryRowContext(ctx, `SELECT state,evidence_sha,stage,worktree_path,spec_id FROM cards WHERE run_id=? AND card_id=?`, tgt.RunID, c.id).Scan(&tgtState, &tgtEvidence, &tgtStage, &tgtWorktree, &tgtSpec)
		if tgtErr != nil && !errors.Is(tgtErr, sql.ErrNoRows) {
			return RebindResult{}, tgtErr
		}
		if tgtErr == nil {
			// A target row is a wall the source's live work is never silently
			// discarded against (card t1973, P1): a TERMINAL row is a completed
			// past carrying its own evidence and approval, and a row whose
			// evidence diverges is different work. Either refuses the whole
			// operation — the source's row stays live and the conflict needs a
			// fresh decision, not the target's stale completion.
			//
			// Two EMPTY evidence values are a match only when the source row
			// carries no progress the target row lacks (card t1973 round 2 P2,
			// rounds 3-4): the t1963-designed dedup — a bare evidence-less
			// source stub beside the target's fresh row, or a pair whose
			// stage/worktree/spec_id the target already holds — keeps the
			// abandon-with-pointer behavior the operator-move and leader-start-
			// consolidation paths live on. A source row that carries ANY of
			// stage/worktree/spec_id the target does not already hold refuses
			// (round 4: a lease expiry or a question receipt returns a row to
			// assigned/needs-decision KEEPING its progress, so the row's STATE
			// cannot make this decision — its actual remaining info does).
			bothEmpty := tgtEvidence == "" && c.evidenceSHA == ""
			stranded := []string(nil)
			if bothEmpty {
				stranded = c.strandsProgressOver(tgtStage, tgtWorktree, tgtSpec)
			}
			if IsTerminalCardState(tgtState) || tgtEvidence != c.evidenceSHA || len(stranded) > 0 {
				return RebindResult{}, rebindRefusal("card %s: run %s already holds a %s row (evidence %q, source %q)%s; the live row of run %s stays and the conflict needs a fresh decision",
					c.id, tgt.RunID, tgtState, tgtEvidence, c.evidenceSHA, bothEmptyNote(stranded), src.RunID)
			}
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
	// Step 5 — the binding of every card step 4 moved, re-keyed or abandoned with
	// a pointer. A card that stays behind in the retired source (a terminal one)
	// keeps its binding: its row still lives in the source run, and a binding
	// sent to the target would name a run holding no row for it, which the
	// completion gate's row resolution reports as unresolvable (fails closed).
	// One parameterised UPDATE per card id, scoped to bindings that name the source.
	for _, id := range slices.Concat(res.Moved, res.Abandoned) {
		if _, err := tx.ExecContext(ctx, `UPDATE card_dispatch SET run_id=?, recorded_at=? WHERE card_id=? AND run_id=?`, tgt.RunID, nowText, id, src.RunID); err != nil {
			return RebindResult{}, err
		}
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

// bothEmptyNote names the both-empty-evidence shape inside a target-row
// conflict refusal: a source row carrying progress the target lacks beside an
// evidence-less target row is not proven to be the same work (card t1973
// round 2 P2, rounds 3-4 — the gate is the source row's ACTUAL remaining
// progress, not its state; a bare stub keeps the t1963 abandon).
func bothEmptyNote(stranded []string) string {
	if len(stranded) == 0 {
		return ""
	}
	return " — the source row carries progress the target lacks (" + strings.Join(stranded, ", ") + ") and neither row records evidence"
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

// workerRow is one registered worker of a run.
type workerRow struct {
	label string
	pid   int
}

// workerPIDs lists the run's registered workers in registration order. The
// cursor is drained and closed before it returns, so the caller can issue its
// next statement on a single-connection handle.
func workerPIDs(ctx context.Context, q rowsQueryer, runID string) ([]workerRow, error) {
	rows, err := q.QueryContext(ctx, `SELECT label, pid FROM workers WHERE run_id=? ORDER BY rowid`, runID)
	if err != nil {
		return nil, err
	}
	var workers []workerRow
	for rows.Next() {
		var w workerRow
		if err := rows.Scan(&w.label, &w.pid); err != nil {
			_ = rows.Close()
			return nil, err
		}
		workers = append(workers, w)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	return workers, rows.Close()
}

// refuseLiveWorker refuses when ANY worker row of the run is not positively
// dead (an undecided probe counts as live). Every row is probed — a check of
// only the first or only the last row would let a live lane slip through.
func refuseLiveWorker(ctx context.Context, tx *sql.Tx, runID string, probe func(int) ProcessIdentityState) error {
	workers, err := workerPIDs(ctx, tx, runID)
	if err != nil {
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
	rows, err := tx.QueryContext(ctx, `SELECT card_id,state,version,stage,evidence_sha,lease_holder,lease_expires_at,worktree_path,spec_id FROM cards WHERE run_id=? AND state NOT IN `+sqlTerminalStates+` ORDER BY card_id`, runID)
	if err != nil {
		return nil, err
	}
	var cards []consolidationCard
	for rows.Next() {
		var c consolidationCard
		if err := rows.Scan(&c.id, &c.state, &c.version, &c.stage, &c.evidenceSHA, &c.leaseHolder, &c.leaseExpiresAt, &c.worktreePath, &c.specID); err != nil {
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
// stamp it was classified on — AND must provably hold nothing live: a lane can
// outlive its leader (REQ-FCR-019), so the holder row's valid lease and the
// run's registered workers are checked before the dead-owner skip applies
// (card t1973, P2). A lease that has not expired, a worker the probe cannot
// call positively dead, or a worker listing that cannot be read, all refuse.
func refuseThirdRunHolder(ctx context.Context, tx *sql.Tx, cardID, srcID, tgtID string, others map[string]RunOwner, now time.Time, probe func(int) ProcessIdentityState) error {
	// SQL: sqlTerminalStates is a compile-time constant; every value goes through a placeholder.
	rows, err := tx.QueryContext(ctx, `SELECT c.run_id, c.state, c.lease_holder, c.lease_expires_at, r.lead_pid, r.lead_process_start FROM cards c JOIN runs r ON r.run_id=c.run_id
WHERE c.card_id=? AND c.run_id NOT IN (?,?) AND r.status='active' AND c.state NOT IN `+sqlTerminalStates+` ORDER BY c.run_id`, cardID, srcID, tgtID)
	if err != nil {
		return err
	}
	type holder struct {
		run, state, leaseHolder, leaseExpires string
		pid                                   int
		start                                 string
	}
	var holders []holder
	for rows.Next() {
		var h holder
		if err := rows.Scan(&h.run, &h.state, &h.leaseHolder, &h.leaseExpires, &h.pid, &h.start); err != nil {
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
			// The leader is dead under the classified stamp, but the run may
			// still hold live work: a valid lease on the row, or a registered
			// worker that is not positively dead, refuses instead of skipping.
			if holdsValidLease(Card{State: h.state, LeaseHolder: h.leaseHolder, LeaseExpiresAt: h.leaseExpires}, now) {
				return rebindRefusal("card %s is held by run %s under a live lease (%s, expires %s); a card a third run still works is not moved",
					cardID, h.run, h.leaseHolder, h.leaseExpires)
			}
			workers, err := workerPIDs(ctx, tx, h.run)
			if err != nil {
				return err
			}
			for _, w := range workers {
				if probe(w.pid) != ProcessIdentityDead {
					return rebindRefusal("card %s is held by run %s which still has live worker %q (pid %d); a card a third run still works is not moved",
						cardID, h.run, w.label, w.pid)
				}
			}
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
