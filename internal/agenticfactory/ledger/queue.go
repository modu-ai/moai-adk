package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

const actorEnqueue = "enqueue"

// latestResult is the result of the card's newest evidence of that kind about that
// commit, or "" when there is none.
func latestResult(ctx context.Context, q querier, cardID, kind, subject string) (string, error) {
	var res string
	switch err := q.QueryRowContext(ctx, `SELECT result FROM evidence WHERE card_id=? AND kind=? AND subject_sha=? ORDER BY id DESC LIMIT 1`, cardID, kind, subject).Scan(&res); {
	case err == sql.ErrNoRows:
		return "", nil
	case err != nil:
		return "", fmt.Errorf("ledger: %w", err)
	}
	return res, nil
}

// factsTx are the core facts the ledger's own tables show for the card at its current
// epoch: receipt from its newest green evidence, receipt-accepted when the newest gate
// evidence for the commit that green run measured passes, queued and (once landed) merged, closed and settled from its queue row
// of the current card epoch. A note is information and stands for nothing.
func factsTx(ctx context.Context, q querier, card Card) ([]string, error) {
	var have []string
	var subject sql.NullString
	var result string
	// Only a run of the card's current epoch is a receipt: after a FIX or a reclaim the
	// earlier green and gate pair, even for the same commit, is not this holder's.
	switch err := q.QueryRowContext(ctx, `SELECT subject_sha,result FROM evidence WHERE card_id=? AND kind=? AND epoch=? ORDER BY id DESC LIMIT 1`, card.ID, KindGreen, card.Epoch).Scan(&subject, &result); {
	case err == sql.ErrNoRows:
	case err != nil:
		return nil, fmt.Errorf("ledger: %w", err)
	case result == ResultPass:
		have = append(have, core.FactReceipt)
		// The newest gate row for that commit decides, as it does at Enqueue.
		gate, err := latestResult(ctx, q, card.ID, KindGate, subject.String)
		if err != nil {
			return nil, err
		}
		if gate == ResultPass {
			have = append(have, core.FactReceiptAccepted)
		}
	}
	var state, candidate string
	var qe int64
	switch err := q.QueryRowContext(ctx, `SELECT state,card_epoch,candidate_sha FROM merge_queue WHERE card_id=?`, card.ID).Scan(&state, &qe, &candidate); {
	case err == sql.ErrNoRows:
	case err != nil:
		return nil, fmt.Errorf("ledger: %w", err)
	case qe == card.Epoch:
		have = append(have, core.FactQueued)
		if state == string(QueueLanded) {
			// The row says its candidate reached the trunk. That completes the card only while
			// the candidate is still the card's newest work: if the newest green evidence
			// names another commit, the landed one is not what the card now offers,
			// and the trunk does not hold the newer one. The ledger holds no git history, so
			// it asks for the same commit rather than for one that contains the other.
			same, err := newestWorkIs(ctx, q, card.ID, candidate)
			if err != nil {
				return nil, err
			}
			if same {
				// A card taken back for rework since its candidate was offered needs work
				// recorded after that move; the landed row and the old receipt are older.
				stands, err := landingStands(ctx, q, card.ID, candidate)
				if err != nil {
					return nil, err
				}
				if stands {
					have = append(have, core.FactMerged, core.FactClosed, core.FactSettled)
				}
			}
		}
	}
	return have, nil
}

// newestWorkIs says whether the card's newest green evidence, the receipt of what it built, is
// about commit sha. A card with no green has no newer work to contradict the row. A gate row
// alone does not decide: a holder may gate a second candidate that was refused, and still finish
// the card on the one that landed.
func newestWorkIs(ctx context.Context, q querier, cardID, sha string) (bool, error) {
	var subject sql.NullString
	switch err := q.QueryRowContext(ctx, `SELECT subject_sha FROM evidence WHERE card_id=? AND kind=? ORDER BY id DESC LIMIT 1`, cardID, KindGreen).Scan(&subject); {
	case err == sql.ErrNoRows:
		return true, nil
	case err != nil:
		return false, fmt.Errorf("ledger: %w", err)
	}
	return subject.String == sha, nil
}

// landingStands says whether the landed candidate may still finish the card. It does unless the
// card went back into the build stage after that candidate was first offered to the queue (see
// offerOf) and has not measured the same commit again since its latest such move. Going back means the holder took the work back; the old receipt and the landed row
// are older than that, so nothing in the record says the rework was done. The same commit
// measured again does count (the trunk holds exactly what was measured, and Enqueue would
// refuse it a second time), but only while the newest run of every kind that was run again
// (green, gate) passed, and at least one did; a different commit is the newest-work check's
// business. The order of the move and the evidence is the move's
// evidence_mark, a count in the ledger itself, so the answer does not depend on the process
// that asks or on the clock.
func landingStands(ctx context.Context, q querier, cardID, candidate string) (bool, error) {
	mark, reworked, err := reworkMark(ctx, q, cardID, candidate)
	if err != nil || !reworked {
		return true, err
	}
	// For each kind that was run again, the newest run decides, as it does for the receipt and at
	// Enqueue: a pass followed by a failure leaves a failure, a failure followed by a pass is
	// repaired. At least one kind must have been run again, and none may end failing.
	rows, err := q.QueryContext(ctx, `SELECT e.result FROM evidence e WHERE e.card_id=? AND e.kind IN (?,?) AND e.subject_sha=? AND e.id>?
		AND e.id=(SELECT MAX(x.id) FROM evidence x WHERE x.card_id=e.card_id AND x.kind=e.kind AND x.subject_sha=e.subject_sha AND x.id>?)`,
		cardID, KindGreen, KindGate, candidate, mark, mark)
	if err != nil {
		return false, fmt.Errorf("ledger: %w", err)
	}
	defer func() { _ = rows.Close() }()
	passed := false
	for rows.Next() {
		var res string
		if err := rows.Scan(&res); err != nil {
			return false, fmt.Errorf("ledger: %w", err)
		}
		if res != ResultPass {
			return false, nil
		}
		passed = true
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("ledger: %w", err)
	}
	return passed, nil
}

// enqueueReason is the reason Enqueue records on the transition that offers a candidate.
func enqueueReason(sha string) string { return "enqueued " + sha }

// offerOf finds where the candidate was first offered to the queue: the seq of that transition and
// the evidence_mark it carries (the card's newest evidence when the offer was made). Only
// Enqueue moves a card into MERGING (Move refuses it), so a transition into MERGING that names
// the candidate is its offer; the actor column is the caller's free text and says nothing about
// which transition that is. The FIRST offer counts: offering the same commit again, after c2
// came and went, is not new work and does not move its mark. A caller can write the same reason
// on a move in place inside MERGING, but that can only be earlier than the real offer of a
// commit it names, never later, so it makes the answer stricter and never looser. When no
// transition names the candidate (a ledger that wrote another reason), the latest entry into
// MERGING is the offer.
func offerOf(ctx context.Context, q querier, cardID, candidate string) (seq, mark int64, found bool, err error) {
	const markSQL = `COALESCE(evidence_mark,(SELECT COALESCE(MAX(id),0) FROM evidence WHERE card_id=?))`
	switch err := q.QueryRowContext(ctx, `SELECT seq,`+markSQL+` FROM transitions WHERE card_id=? AND to_stage=? AND reason=? ORDER BY seq ASC LIMIT 1`,
		cardID, cardID, string(core.StageMerging), enqueueReason(candidate)).Scan(&seq, &mark); {
	case err == nil:
		return seq, mark, true, nil
	case err != sql.ErrNoRows:
		return 0, 0, false, fmt.Errorf("ledger: %w", err)
	}
	switch err := q.QueryRowContext(ctx, `SELECT seq,`+markSQL+` FROM transitions WHERE card_id=? AND to_stage=? AND COALESCE(from_stage,'')<>? ORDER BY seq DESC LIMIT 1`,
		cardID, cardID, string(core.StageMerging), string(core.StageMerging)).Scan(&seq, &mark); {
	case err == sql.ErrNoRows:
		return 0, 0, false, nil
	case err != nil:
		return 0, 0, false, fmt.Errorf("ledger: %w", err)
	}
	return seq, mark, true, nil
}

// reworkMark says whether the card went back into the build stage after the candidate was first
// offered to the queue, and if so how much evidence it had when it last did: the evidence_mark
// of its latest such transition. A move inside BUILDING (a sub-stage change, a park) is not going
// back; neither is a card that never offered the candidate.
func reworkMark(ctx context.Context, q querier, cardID, candidate string) (mark int64, reworked bool, err error) {
	offered, _, found, err := offerOf(ctx, q, cardID, candidate)
	if err != nil || !found {
		return 0, false, err
	}
	// A move without a mark cannot say what came after it, so it counts all the evidence as older.
	switch err := q.QueryRowContext(ctx, `SELECT COALESCE(evidence_mark,(SELECT COALESCE(MAX(id),0) FROM evidence WHERE card_id=?)) FROM transitions
		WHERE card_id=? AND seq>? AND to_stage=? AND COALESCE(from_stage,'')<>? ORDER BY seq DESC LIMIT 1`,
		cardID, cardID, offered, string(core.StageBuilding), string(core.StageBuilding)).Scan(&mark); {
	case err == sql.ErrNoRows:
		return 0, false, nil
	case err != nil:
		return 0, false, fmt.Errorf("ledger: %w", err)
	}
	return mark, true, nil
}

// queueSettledAt says whether the card's queue row was queued at that card epoch and has
// ended, landed or failed.
func queueSettledAt(ctx context.Context, q querier, cardID string, cardEpoch int64) (bool, error) {
	var n int
	if err := q.QueryRowContext(ctx, `SELECT count(*) FROM merge_queue WHERE card_id=? AND card_epoch=? AND state IN ('landed','failed')`, cardID, cardEpoch).Scan(&n); err != nil {
		return false, fmt.Errorf("ledger: %w", err)
	}
	return n > 0, nil
}

// Enqueue puts the card's candidate on the merge queue and moves the card into MERGING;
// see the API.
func (s *Store) Enqueue(ctx context.Context, cardID string, epoch int64, candidateSHA string) error {
	if strings.TrimSpace(candidateSHA) == "" {
		return fmt.Errorf("%w: a candidate needs its commit SHA", ErrInvalid)
	}
	return s.write(ctx, func(tx *sql.Tx) error {
		if err := leaseTx(ctx, tx, cardID, epoch); err != nil {
			return err
		}
		card, err := cardTx(ctx, tx, cardID)
		if err != nil {
			return err
		}
		if card.Position.Park != nil {
			return fmt.Errorf("%w: %s is parked", core.ErrTransitionInvalid, cardID)
		}
		if res, err := latestResult(ctx, tx, cardID, KindGate, candidateSHA); err != nil {
			return err
		} else if res != ResultPass {
			return fmt.Errorf("%w: no passing gate evidence for commit %s", ErrEvidenceStale, candidateSHA)
		}
		if len(card.Spec.Risk) > 0 {
			if card.Position.Stage != core.StageReviewing {
				return fmt.Errorf("%w: %s has risk %v and must be reviewed first", ErrEvidenceStale, cardID, card.Spec.Risk)
			}
			if res, err := latestResult(ctx, tx, cardID, KindReview, candidateSHA); err != nil {
				return err
			} else if res != ResultPass {
				return fmt.Errorf("%w: no passing review evidence for commit %s", ErrEvidenceStale, candidateSHA)
			}
		}
		var cur string
		var oldSHA, oldReason sql.NullString
		var oldEpoch int64
		rowErr := tx.QueryRowContext(ctx, `SELECT state,candidate_sha,fail_reason,card_epoch FROM merge_queue WHERE card_id=?`, cardID).Scan(&cur, &oldSHA, &oldReason, &oldEpoch)
		// Whether the card was taken back for rework since the landed candidate was offered.
		reworked := false
		if rowErr == nil && cur == string(QueueLanded) {
			if _, reworked, err = reworkMark(ctx, tx, cardID, oldSHA.String); err != nil {
				return err
			}
		}
		switch err := rowErr; {
		case err == sql.ErrNoRows:
		case err != nil:
			return fmt.Errorf("ledger: %w", err)
		case cur == string(QueueLanded) && (oldSHA.String == candidateSHA || oldEpoch != card.Epoch || !reworked):
			// A landed row is history: the same commit is never landed twice, a holder of a
			// later card epoch does not replace the row of an earlier one (N3 stays as it was),
			// and neither does a card that was not taken back for rework.
			return fmt.Errorf("%w: %s has landed", ErrInvalid, cardID)
		case cur == string(QueueLanded):
			// The holder that landed c1, at the epoch it still holds, offers a different commit
			// after the card went back into the build stage: new work. Refusing it would leave the
			// card no way to the trunk at all. It goes through the queue like any other and
			// takes the row, so the landing that is already in the trunk stays on record as
			// a note, and the card cannot finish on it any more, since this commit is now its
			// newest work.
			text := fmt.Sprintf("card %s: candidate %s had landed at card epoch %d; queued again as new work %s", cardID, oldSHA.String, oldEpoch, candidateSHA)
			if _, err := tx.ExecContext(ctx, `INSERT INTO notes(kind,card_id,text,at) VALUES('decision',?,?,`+nowSQL+`)`, cardID, text); err != nil {
				return fmt.Errorf("ledger: %w", err)
			}
		case cur == string(QueueLanding):
			// A lander may be working under the merge lock; its row is settled, not replaced.
			return fmt.Errorf("%w: %s is landing (card epoch %d); the merge lock holder settles it first", ErrInvalid, cardID, oldEpoch)
		case cur != string(QueueFailed) && oldEpoch == card.Epoch && (cur != string(QueueWaiting) || card.Position.Stage == core.StageMerging):
			// A waiting row of a card that is no longer in MERGING was withdrawn: no lander has
			// taken it (verifying and landing need the card in MERGING), so it is restarted below.
			return fmt.Errorf("%w: %s is already in the queue (%s)", ErrInvalid, cardID, cur)
		default:
			// The restart overwrites the row, so what it said is kept as a note: a failure, or
			// an open row of an older card epoch, which the epoch fence leaves nobody able to write.
			text := fmt.Sprintf("card %s: candidate %s failed in the merge queue (%s); queued again as %s", cardID, oldSHA.String, oldReason.String, candidateSHA)
			if cur != string(QueueFailed) && oldEpoch != card.Epoch {
				text = fmt.Sprintf("card %s: candidate %s was %s at card epoch %d when the card was claimed again at epoch %d; queued again as %s", cardID, oldSHA.String, cur, oldEpoch, card.Epoch, candidateSHA)
			} else if cur != string(QueueFailed) {
				text = fmt.Sprintf("card %s: candidate %s was left %s when the card was withdrawn from MERGING at epoch %d; queued again as %s", cardID, oldSHA.String, cur, oldEpoch, candidateSHA)
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO notes(kind,card_id,text,at) VALUES('decision',?,?,`+nowSQL+`)`, cardID, text); err != nil {
				return fmt.Errorf("ledger: %w", err)
			}
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO merge_queue(card_id,candidate_sha,state,card_epoch,enqueued_at,updated_at) VALUES(?,?,'waiting',?,`+nowSQL+`,`+nowSQL+`)
			ON CONFLICT(card_id) DO UPDATE SET candidate_sha=excluded.candidate_sha, trunk_before=NULL, state='waiting', fail_reason=NULL, card_epoch=excluded.card_epoch, enqueued_at=excluded.enqueued_at, updated_at=excluded.updated_at`,
			cardID, candidateSHA, card.Epoch); err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		have, err := factsTx(ctx, tx, card)
		if err != nil {
			return err
		}
		to := core.Position{Stage: core.StageMerging}
		if err := core.CheckMove(card.Position, to, have); err != nil {
			return err
		}
		if err := setPosition(ctx, tx, cardID, to); err != nil {
			return err
		}
		return insertTransition(ctx, tx, cardID, card.Position, to, epoch, actorEnqueue, enqueueReason(candidateSHA))
	})
}

// queueNext is the states a queue row may go to from each state. A verifying row may be
// entered again at the card epoch: a lander that could not take the landing lock leaves it
// verifying and the next Land comes back. This also means verifying is no longer a
// one-at-a-time marker: two processes landing the same card may both verify outside the
// lock; the landing lock and the ancestry reconcile still serialise the landing itself, so
// the cost is a duplicate verification and an overwritten trunk_before, not a double landing. A reconcile sends a verifying or landing row back
// to waiting; landed and failed are final (a failed card is
// queued again by Enqueue, which starts the row over).
var queueNext = map[QueueState][]QueueState{
	QueueWaiting:   {QueueVerifying, QueueFailed},
	QueueVerifying: {QueueVerifying, QueueLanding, QueueFailed, QueueWaiting},
	QueueLanding:   {QueueLanded, QueueFailed, QueueWaiting},
}

// QueueUpdate moves a queue item to state. The caller presents its epoch with its kind:
// verifying and failed need the card epoch the row was queued at, which must still be the
// card's; landing, landed and the reconcile of a verifying or landing row back to waiting
// need the epoch of the landing lock (a waiting target from any other state is no edge). A kind the table row
// does not want is ErrStaleEpoch, even when the numbers coincide. Every write also
// names the candidate its caller verified (ErrInvalid when it names none, ErrStaleEpoch when
// the row was restarted with another). Naming the candidate shows the caller read
// the row, not that the candidate reached trunk; a landed settlement makes the row final, so
// the card holder who then finishes the card is judging that the landed candidate satisfied
// it. Known cost: the live holder of the one merge lock may
// settle any card's older landing row; the candidate it names is the only check on which.
func (s *Store) QueueUpdate(ctx context.Context, cardID string, presented QueueEpoch, state QueueState, trunkBefore, failReason string) error {
	epoch := presented.N
	if presented.Kind != EpochCard && presented.Kind != EpochLock {
		return fmt.Errorf("%w: epoch kind %q", ErrInvalid, presented.Kind)
	}
	switch state {
	case QueueWaiting, QueueVerifying, QueueLanding, QueueLanded:
	case QueueFailed:
		if strings.TrimSpace(failReason) == "" {
			return fmt.Errorf("%w: a failure needs its reason", ErrInvalid)
		}
	default:
		return fmt.Errorf("%w: queue state %q", ErrInvalid, state)
	}
	return s.write(ctx, func(tx *sql.Tx) error {
		var cur, rowCandidate string
		var rowEpoch int64
		switch err := tx.QueryRowContext(ctx, `SELECT state,card_epoch,candidate_sha FROM merge_queue WHERE card_id=?`, cardID).Scan(&cur, &rowEpoch, &rowCandidate); {
		case err == sql.ErrNoRows:
			return fmt.Errorf("%w: %s is not in the queue", ErrNotFound, cardID)
		case err != nil:
			return fmt.Errorf("ledger: %w", err)
		}
		var cardEpoch int64
		if err := tx.QueryRowContext(ctx, `SELECT epoch FROM cards WHERE id=?`, cardID).Scan(&cardEpoch); err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		curState := QueueState(cur)
		// A row of an older card epoch belongs to a holder that was fenced: no one writes it,
		// with one exception. A landing row may have a lander still working under the merge
		// lock, so the live holder of that lock, and only it, settles the row to landed or
		// failed from the trunk ancestry result. The row keeps its epoch.
		settlesOlder := false
		if rowEpoch != cardEpoch {
			if curState == QueueLanding && presented.Kind == EpochLock && (state == QueueLanded || state == QueueFailed) {
				held, err := lockHeldAt(ctx, tx, LockMerge, epoch)
				if err != nil {
					return err
				}
				settlesOlder = held
			}
			if !settlesOlder {
				return fmt.Errorf("%w: %s was claimed again (queued at epoch %d, now %d)", ErrStaleEpoch, cardID, rowEpoch, cardEpoch)
			}
		}
		lockStates := state == QueueLanding || state == QueueLanded ||
			(state == QueueWaiting && (curState == QueueVerifying || curState == QueueLanding))
		// The caller says which epoch it holds; the table row says which one the write needs.
		// A number that happens to match the other table is no authority.
		want := EpochCard
		if lockStates {
			want = EpochLock
		}
		if settlesOlder {
			want = EpochLock // the lock was just checked above
		}
		if presented.Kind != want {
			return fmt.Errorf("%w: %s from %s needs a %s epoch, not a %s epoch", ErrStaleEpoch, state, curState, want, presented.Kind)
		}
		// A write names the candidate its caller verified; a row restarted with another
		// candidate is not that caller's to move (subject_sha is the authority).
		if presented.Candidate != rowCandidate {
			if strings.TrimSpace(presented.Candidate) == "" {
				return fmt.Errorf("%w: a write names the candidate it verified", ErrInvalid)
			}
			return fmt.Errorf("%w: %s is queued with candidate %s, not %s", ErrStaleEpoch, cardID, rowCandidate, presented.Candidate)
		}
		if settlesOlder {
			// nothing more to check: the live lock holder is the authority
		} else if lockStates {
			held, err := lockHeldAt(ctx, tx, LockMerge, epoch)
			if err != nil {
				return err
			}
			if !held {
				return fmt.Errorf("%w: %s needs the %s lock epoch", ErrStaleEpoch, state, LockMerge)
			}
		} else if epoch != rowEpoch {
			return fmt.Errorf("%w: %s is at card epoch %d, not %d", ErrStaleEpoch, cardID, rowEpoch, epoch)
		}
		// Taking a row into verifying, landing or landed offers the card's work to the trunk, so
		// the card must still be in MERGING at the row's epoch. Failing a row and the reconcile
		// back to waiting only end an attempt, and the settlement of an older landing row is the
		// exception above; none of those is guarded. Landed for a row that is ALREADY landing is
		// the one other exception: the trunk may have moved under the lock, so the lock holder
		// can still say so after the card backed out, and the mismatch is noted below.
		var stage string
		var parked int
		if err := tx.QueryRowContext(ctx, `SELECT stage,parked FROM cards WHERE id=?`, cardID).Scan(&stage, &parked); err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		// A hold that lands mid-gate also stops the row from being taken further: verifying and
		// landing offer the card's work to the trunk. Landed after landing, failing and the
		// reconcile back to waiting stay possible, because the trunk may already have moved.
		if parked == 1 && (state == QueueVerifying || state == QueueLanding) {
			return fmt.Errorf("%w: %s for %s at epoch %d refused: the card is parked", ErrStaleEpoch, state, cardID, rowEpoch)
		}
		landedAfterLanding := state == QueueLanded && curState == QueueLanding
		if !settlesOlder && !landedAfterLanding && (state == QueueVerifying || state == QueueLanding || state == QueueLanded) {
			if stage != string(core.StageMerging) {
				return fmt.Errorf("%w: %s for %s at epoch %d needs the card in MERGING, but it is at %s", ErrStaleEpoch, state, cardID, rowEpoch, stage)
			}
		}
		ok := false
		for _, n := range queueNext[curState] {
			ok = ok || n == state
		}
		if !ok {
			return fmt.Errorf("%w: a queue row goes from %s to %s only through %v", ErrInvalid, cur, state, queueNext[curState])
		}
		var fail any
		if state == QueueFailed {
			fail = failReason
		}
		// A landed settlement of an older row follows the card to its current epoch, so the
		// merged, closed and settled facts show for the card as it is now.
		newEpoch := rowEpoch
		if settlesOlder && state == QueueLanded {
			newEpoch = cardEpoch
		}
		_, err := tx.ExecContext(ctx, `UPDATE merge_queue SET state=?, trunk_before=COALESCE(NULLIF(?,''),trunk_before), fail_reason=?, card_epoch=?, updated_at=`+nowSQL+` WHERE card_id=?`,
			string(state), trunkBefore, fail, newEpoch, cardID)
		if err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		if settlesOlder {
			text := fmt.Sprintf("card %s: landing row of card epoch %d settled as %s by the merge lock holder (the card is at epoch %d now)", cardID, rowEpoch, state, cardEpoch)
			if state == QueueLanded {
				text += "; the row now carries the card's current epoch"
			}
			if state == QueueFailed {
				text += ": " + failReason
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO notes(kind,card_id,text,at) VALUES('decision',?,?,`+nowSQL+`)`, cardID, text); err != nil {
				return fmt.Errorf("ledger: %w", err)
			}
		}
		if landedAfterLanding && !settlesOlder && stage != string(core.StageMerging) {
			text := fmt.Sprintf("card %s: row of card epoch %d landed while the card was at %s, not MERGING; the trunk had moved under the lock, so the row says what happened and the card is the holder's to reconcile", cardID, rowEpoch, stage)
			if _, err := tx.ExecContext(ctx, `INSERT INTO notes(kind,card_id,text,at) VALUES('decision',?,?,`+nowSQL+`)`, cardID, text); err != nil {
				return fmt.Errorf("ledger: %w", err)
			}
		}
		return nil
	})
}

// Queue lists the merge queue, oldest first.
func (s *Store) Queue(ctx context.Context) ([]QueueItem, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT card_id,card_epoch,candidate_sha,trunk_before,state,fail_reason,enqueued_at,updated_at FROM merge_queue ORDER BY enqueued_at, card_id`)
	if err != nil {
		return nil, fmt.Errorf("ledger: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []QueueItem
	for rows.Next() {
		var (
			it                  QueueItem
			trunk, fail         sql.NullString
			state, enq, updated string
		)
		if err := rows.Scan(&it.CardID, &it.CardEpoch, &it.CandidateSHA, &trunk, &state, &fail, &enq, &updated); err != nil {
			return nil, err
		}
		it.TrunkBefore, it.FailReason, it.State = trunk.String, fail.String, QueueState(state)
		if it.EnqueuedAt, err = parseTime(enq); err != nil {
			return nil, fmt.Errorf("ledger: queue %s: %w", it.CardID, err)
		}
		if it.UpdatedAt, err = parseTime(updated); err != nil {
			return nil, fmt.Errorf("ledger: queue %s: %w", it.CardID, err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}
