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
		switch err := tx.QueryRowContext(ctx, `SELECT state,candidate_sha,fail_reason FROM merge_queue WHERE card_id=?`, cardID).Scan(&cur, &oldSHA, &oldReason); {
		case err == sql.ErrNoRows:
		case err != nil:
			return fmt.Errorf("ledger: %w", err)
		case cur != string(QueueFailed):
			return fmt.Errorf("%w: %s is already in the queue (%s)", ErrInvalid, cardID, cur)
		default:
			// The restart overwrites the row, so the earlier failure is kept as a note.
			text := fmt.Sprintf("card %s: candidate %s failed in the merge queue (%s); queued again as %s", cardID, oldSHA.String, oldReason.String, candidateSHA)
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
// verifying and the next Land comes back. A reconcile sends a verifying or landing row back
// to waiting; landed and failed are final (a failed card is
// queued again by Enqueue, which starts the row over).
var queueNext = map[QueueState][]QueueState{
	QueueWaiting:   {QueueVerifying, QueueFailed},
	QueueVerifying: {QueueVerifying, QueueLanding, QueueFailed, QueueWaiting},
	QueueLanding:   {QueueLanded, QueueFailed, QueueWaiting},
}

// QueueUpdate moves a queue item to state; see the API for which epoch each state checks.
func (s *Store) QueueUpdate(ctx context.Context, cardID string, epoch int64, state QueueState, trunkBefore, failReason string) error {
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
		var cur string
		var rowEpoch int64
		switch err := tx.QueryRowContext(ctx, `SELECT state,card_epoch FROM merge_queue WHERE card_id=?`, cardID).Scan(&cur, &rowEpoch); {
		case err == sql.ErrNoRows:
			return fmt.Errorf("%w: %s is not in the queue", ErrNotFound, cardID)
		case err != nil:
			return fmt.Errorf("ledger: %w", err)
		}
		var cardEpoch int64
		if err := tx.QueryRowContext(ctx, `SELECT epoch FROM cards WHERE id=?`, cardID).Scan(&cardEpoch); err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		// A row of an older card epoch belongs to a holder that was fenced: no one writes it.
		if rowEpoch != cardEpoch {
			return fmt.Errorf("%w: %s was claimed again (queued at epoch %d, now %d)", ErrStaleEpoch, cardID, rowEpoch, cardEpoch)
		}
		curState := QueueState(cur)
		lockStates := state == QueueLanding || state == QueueLanded ||
			(state == QueueWaiting && (curState == QueueVerifying || curState == QueueLanding))
		if lockStates {
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
		_, err := tx.ExecContext(ctx, `UPDATE merge_queue SET state=?, trunk_before=COALESCE(NULLIF(?,''),trunk_before), fail_reason=?, updated_at=`+nowSQL+` WHERE card_id=?`,
			string(state), trunkBefore, fail, cardID)
		if err != nil {
			return fmt.Errorf("ledger: %w", err)
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
