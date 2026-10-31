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
	var state string
	var qe int64
	switch err := q.QueryRowContext(ctx, `SELECT state,card_epoch FROM merge_queue WHERE card_id=?`, card.ID).Scan(&state, &qe); {
	case err == sql.ErrNoRows:
	case err != nil:
		return nil, fmt.Errorf("ledger: %w", err)
	case qe == card.Epoch:
		have = append(have, core.FactQueued)
		if state == string(QueueLanded) {
			have = append(have, core.FactMerged, core.FactClosed, core.FactSettled)
		}
	}
	return have, nil
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
		switch err := tx.QueryRowContext(ctx, `SELECT state,candidate_sha,fail_reason,card_epoch FROM merge_queue WHERE card_id=?`, cardID).Scan(&cur, &oldSHA, &oldReason, &oldEpoch); {
		case err == sql.ErrNoRows:
		case err != nil:
			return fmt.Errorf("ledger: %w", err)
		case cur == string(QueueLanded):
			return fmt.Errorf("%w: %s has landed", ErrInvalid, cardID)
		case cur == string(QueueLanding):
			// A lander may be working under the merge lock; its row is settled, not replaced.
			return fmt.Errorf("%w: %s is landing (card epoch %d); the merge lock holder settles it first", ErrInvalid, cardID, oldEpoch)
		case cur != string(QueueFailed) && oldEpoch == card.Epoch:
			return fmt.Errorf("%w: %s is already in the queue (%s)", ErrInvalid, cardID, cur)
		default:
			// The restart overwrites the row, so what it said is kept as a note: a failure, or
			// an open row of an older card epoch, which the epoch fence leaves nobody able to write.
			text := fmt.Sprintf("card %s: candidate %s failed in the merge queue (%s); queued again as %s", cardID, oldSHA.String, oldReason.String, candidateSHA)
			if cur != string(QueueFailed) {
				text = fmt.Sprintf("card %s: candidate %s was %s at card epoch %d when the card was claimed again at epoch %d; queued again as %s", cardID, oldSHA.String, cur, oldEpoch, card.Epoch, candidateSHA)
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
		return insertTransition(ctx, tx, cardID, card.Position, to, epoch, actorEnqueue, "enqueued "+candidateSHA)
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
		// exception above; none of those is guarded.
		if !settlesOlder && (state == QueueVerifying || state == QueueLanding || state == QueueLanded) {
			var stage string
			if err := tx.QueryRowContext(ctx, `SELECT stage FROM cards WHERE id=?`, cardID).Scan(&stage); err != nil {
				return fmt.Errorf("ledger: %w", err)
			}
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
