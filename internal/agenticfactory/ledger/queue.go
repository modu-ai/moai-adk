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

// newestWorkIs says whether the landed commit sha is still the card's newest work: the card's
// newest green evidence, the receipt of what it built, is about sha, or was recorded no later than
// the moment sha was first offered to the queue. A green about another commit that is OLDER than
// that offer is work the holder chose to leave behind (it withdrew c1 and offered a gate-only c2,
// which REQ-AFL-008 accepts); one NEWER than the offer is work the landing does not hold. A card
// with no green has no newer work to contradict the row. A gate row alone does not decide: a
// holder may gate a second candidate that was refused, and still finish the card on the one that
// landed. The order is the evidence ids against the offer's evidence_mark, counts in the ledger.
func newestWorkIs(ctx context.Context, q querier, cardID, sha string) (bool, error) {
	var id int64
	var subject sql.NullString
	switch err := q.QueryRowContext(ctx, `SELECT id,subject_sha FROM evidence WHERE card_id=? AND kind=? ORDER BY id DESC LIMIT 1`, cardID, KindGreen).Scan(&id, &subject); {
	case err == sql.ErrNoRows:
		return true, nil
	case err != nil:
		return false, fmt.Errorf("ledger: %w", err)
	}
	if subject.String == sha {
		return true, nil
	}
	_, mark, found, err := offerOf(ctx, q, cardID, sha)
	if err != nil {
		return false, err
	}
	return found && id <= mark, nil
}

// landingStands says whether the landed candidate may still finish the card. Three things must hold.
//
//  1. The newest record of each kind (green, gate, review) about it is not a failure, whenever
//     that record was made and whether or not the card was ever taken back, as it is for the
//     receipt and at Enqueue: a pass followed by a failure leaves a failure, a failure followed
//     by a pass is repaired, a kind never recorded for the commit is not a failure.
//  2. If the card went back into the build stage after the candidate was first offered to the
//     queue (see offerOf), at least one passing run (green or gate) of it follows the latest such
//     move. Going back means the holder took the work back; the old receipt and the landed row
//     are older than that, so nothing in the record says the rework was done. The same commit
//     measured again counts (the trunk holds exactly what was measured; Enqueue refuses to take a
//     commit that fails this check, so a landed row has normally passed it already).
//  3. If a green, gate or review about another commit FAILED after the candidate was offered, or
//     the card left MERGING after the offer and has recorded evidence about another commit since
//     its latest such move, the candidate was measured again after that evidence, with a passing
//     green and a passing gate. The card did other work in between, the ledger cannot tell
//     whether that work was given up, and a gate of the landed commit alone would be a way to
//     finish on it regardless. A passing record about another commit while the card never left
//     MERGING obliges nothing (D5). A different commit as the newest GREEN is the newest-work
//     check's business.
//
// The order of the moves and the evidence is the moves' evidence_mark, a count in the ledger
// itself, so the answer does not depend on the process that asks or on the clock.
func landingStands(ctx context.Context, q querier, cardID, candidate string) (bool, error) {
	if ok, err := noFailingNewestRun(ctx, q, cardID, candidate); err != nil || !ok {
		return false, err
	}
	mark, reworked, err := reworkMark(ctx, q, cardID, candidate)
	if err != nil {
		return false, err
	}
	if reworked {
		var passes int
		if err := q.QueryRowContext(ctx, `SELECT count(*) FROM evidence WHERE card_id=? AND kind IN (?,?) AND subject_sha=? AND result=? AND id>?`,
			cardID, KindGreen, KindGate, candidate, ResultPass, mark).Scan(&passes); err != nil {
			return false, fmt.Errorf("ledger: %w", err)
		}
		if passes == 0 {
			return false, nil
		}
	}
	return measuredAfterOtherWork(ctx, q, cardID, candidate)
}

// measuredAfterOtherWork is the third condition of landingStands. It finds the newest evidence that
// obliges the card to measure the candidate again: a FAILURE of any green, gate or review about
// another commit recorded after the candidate was offered (the card found something wrong beside
// the landed work, whether it was still in MERGING or already out of it), or, once the card left
// MERGING after the offer, ANY green, gate or review about another commit recorded since. A passing
// record about another commit while the card never left MERGING obliges nothing (D5). With such
// a record, the candidate must have a passing green and a passing gate after it.
func measuredAfterOtherWork(ctx context.Context, q querier, cardID, candidate string) (bool, error) {
	offered, offerMark, found, err := offerOf(ctx, q, cardID, candidate)
	if err != nil || !found {
		return true, err
	}
	var failed int64
	if err := q.QueryRowContext(ctx, `SELECT COALESCE(MAX(id),0) FROM evidence WHERE card_id=? AND kind IN (?,?,?) AND subject_sha<>? AND result<>? AND id>?`,
		cardID, KindGreen, KindGate, KindReview, candidate, ResultPass, offerMark).Scan(&failed); err != nil {
		return false, fmt.Errorf("ledger: %w", err)
	}
	// The latest move out of MERGING after the offer, to anything but a terminal stage. A move
	// without a mark counts all the evidence as older.
	var other int64
	var left int64
	switch err := q.QueryRowContext(ctx, `SELECT `+offerMarkSQL+` FROM transitions
		WHERE card_id=? AND seq>? AND from_stage=? AND to_stage NOT IN (?,?,?) ORDER BY seq DESC LIMIT 1`,
		cardID, cardID, offered, string(core.StageMerging), string(core.StageMerging), string(core.StageDone), string(core.StageCancelled)).Scan(&left); {
	case err == sql.ErrNoRows:
	case err != nil:
		return false, fmt.Errorf("ledger: %w", err)
	default:
		if err := q.QueryRowContext(ctx, `SELECT COALESCE(MAX(id),0) FROM evidence WHERE card_id=? AND kind IN (?,?,?) AND subject_sha<>? AND id>?`,
			cardID, KindGreen, KindGate, KindReview, candidate, left).Scan(&other); err != nil {
			return false, fmt.Errorf("ledger: %w", err)
		}
	}
	bar := max(failed, other)
	if bar == 0 {
		return true, nil
	}
	for _, kind := range []string{KindGreen, KindGate} {
		var passes int
		if err := q.QueryRowContext(ctx, `SELECT count(*) FROM evidence WHERE card_id=? AND kind=? AND subject_sha=? AND result=? AND id>?`,
			cardID, kind, candidate, ResultPass, bar).Scan(&passes); err != nil {
			return false, fmt.Errorf("ledger: %w", err)
		}
		if passes == 0 {
			return false, nil
		}
	}
	return true, nil
}

// noFailingNewestRun says whether the newest green run, the newest gate run and the newest review
// about the commit (whenever they were recorded) are not failures. A kind never recorded for it is
// not a failure. Merge and red evidence are about other commits by design and stay out.
func noFailingNewestRun(ctx context.Context, q querier, cardID, sha string) (bool, error) {
	for _, kind := range []string{KindGreen, KindGate, KindReview} {
		res, err := latestResult(ctx, q, cardID, kind, sha)
		if err != nil {
			return false, err
		}
		if res != "" && res != ResultPass {
			return false, nil
		}
	}
	return true, nil
}

// measuredInEpoch says whether the commit was measured in that card epoch: at least one passing
// green or gate run recorded under it, and no failing newest run of either kind. A run of an
// earlier epoch is the earlier holder's.
func measuredInEpoch(ctx context.Context, q querier, cardID, sha string, epoch int64) (bool, error) {
	var n int
	if err := q.QueryRowContext(ctx, `SELECT count(*) FROM evidence WHERE card_id=? AND kind IN (?,?) AND subject_sha=? AND result=? AND epoch=?`,
		cardID, KindGreen, KindGate, sha, ResultPass, epoch).Scan(&n); err != nil {
		return false, fmt.Errorf("ledger: %w", err)
	}
	if n == 0 {
		return false, nil
	}
	return noFailingNewestRun(ctx, q, cardID, sha)
}

// enqueueReason is the reason Enqueue records on the transition that offers a candidate.
func enqueueReason(sha string) string { return "enqueued " + sha }

// offerMarkSQL is the evidence_mark of a transition row; a row without one counts all the card's
// evidence as older (the one placeholder is the card id).
const offerMarkSQL = `COALESCE(evidence_mark,(SELECT COALESCE(MAX(id),0) FROM evidence WHERE card_id=?))`

// namedOfferOf finds the first transition into MERGING whose reason is Enqueue's reason for that
// candidate: the proof that this commit was itself offered to the queue.
func namedOfferOf(ctx context.Context, q querier, cardID, candidate string) (seq, mark int64, found bool, err error) {
	err = q.QueryRowContext(ctx, `SELECT seq,`+offerMarkSQL+` FROM transitions WHERE card_id=? AND to_stage=? AND reason=? ORDER BY seq ASC LIMIT 1`,
		cardID, cardID, string(core.StageMerging), enqueueReason(candidate)).Scan(&seq, &mark)
	if err == sql.ErrNoRows {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, fmt.Errorf("ledger: %w", err)
	}
	return seq, mark, true, nil
}

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
	if seq, mark, found, err = namedOfferOf(ctx, q, cardID, candidate); err != nil || found {
		return seq, mark, found, err
	}
	switch err := q.QueryRowContext(ctx, `SELECT seq,`+offerMarkSQL+` FROM transitions WHERE card_id=? AND to_stage=? AND COALESCE(from_stage,'')<>? ORDER BY seq DESC LIMIT 1`,
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
		takeOver := false // the landed row of an earlier epoch is kept and follows the card to this one
		if rowErr == nil && cur == string(QueueLanded) {
			if _, reworked, err = reworkMark(ctx, tx, cardID, oldSHA.String); err != nil {
				return err
			}
		}
		switch err := rowErr; {
		case err == sql.ErrNoRows:
		case err != nil:
			return fmt.Errorf("ledger: %w", err)
		case cur == string(QueueLanded) && oldEpoch != card.Epoch:
			// The row landed under an earlier holder (a lower card epoch). The landing is a fact
			// about the trunk, so the holder of this epoch may take it over, if the offer is that
			// same commit, or replace it with a new commit; either way only after it has measured
			// the commit itself, in its own epoch. Without that the card stays as it is.
			measured, err := measuredInEpoch(ctx, tx, cardID, candidateSHA, card.Epoch)
			if err != nil {
				return err
			}
			if !measured {
				return fmt.Errorf("%w: %s has landed at card epoch %d; commit %s needs a passing run in epoch %d, with no failing newest run, before it takes the card on", ErrInvalid, cardID, oldEpoch, candidateSHA, card.Epoch)
			}
			text := fmt.Sprintf("card %s: candidate %s had landed at card epoch %d; queued again as new work %s at card epoch %d", cardID, oldSHA.String, oldEpoch, candidateSHA, card.Epoch)
			if oldSHA.String == candidateSHA {
				takeOver = true
				text = fmt.Sprintf("card %s: candidate %s had landed at card epoch %d; taken over at card epoch %d after it was measured again", cardID, candidateSHA, oldEpoch, card.Epoch)
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO notes(kind,card_id,text,at) VALUES('decision',?,?,`+nowSQL+`)`, cardID, text); err != nil {
				return fmt.Errorf("ledger: %w", err)
			}
		case cur == string(QueueLanded) && (oldSHA.String == candidateSHA || !reworked):
			// A landed row of this epoch is history: the same commit is never landed twice, and a
			// card that was not taken back for rework does not replace it.
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
		// The candidate is the commit the card measured last. A gate-only candidate is that commit
		// the moment it is gated (REQ-AFL-008), but an older commit offered over a newer one's
		// green is not: the trunk would hold the old work while the card's newest work, which
		// may have failed, is not in it. The ledger cannot tell whether one commit contains the
		// other, so it asks for the same commit; measuring the older one again makes it the
		// newest. Red and merge evidence are about other commits by design and do not count.
		// Only a passing green or gate of the candidate measures it again: a review of it is a
		// reading, not a run, and a failing record is no measurement at all, so neither is the
		// newest evidence (a review of ANOTHER commit still is, as before).
		var lastSubject sql.NullString
		switch err := tx.QueryRowContext(ctx, `SELECT subject_sha FROM evidence WHERE card_id=? AND kind IN (?,?,?)
			AND NOT (COALESCE(subject_sha,'')=? AND (kind=? OR result<>?)) ORDER BY id DESC LIMIT 1`,
			cardID, KindGreen, KindGate, KindReview, candidateSHA, KindReview, ResultPass).Scan(&lastSubject); {
		case err == sql.ErrNoRows:
		case err != nil:
			return fmt.Errorf("ledger: %w", err)
		case lastSubject.Valid && lastSubject.String != candidateSHA:
			return fmt.Errorf("%w: the newest evidence of %s is about commit %s, not %s; measure %s again first", ErrEvidenceStale, cardID, lastSubject.String, candidateSHA, candidateSHA)
		}
		// A commit whose newest green run failed cannot finish the card (landingStands), so it is
		// not taken either; its gate is already known to pass.
		if ok, err := noFailingNewestRun(ctx, tx, cardID, candidateSHA); err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("%w: the newest run of commit %s failed; measure it again first", ErrEvidenceStale, candidateSHA)
		}
		// A commit this card already offered, and was then taken back for rework, is taken again
		// only with the run that finishing needs already recorded (see landingStands). Refusing it
		// here, while the holder is still there to measure it, is the only place a refusal
		// leaves a way out: after the landing, the same demand would strand the card in MERGING
		// with the work in the trunk and no one left to answer it. A commit never offered before
		// carries no mark, and a gate-only one is still taken (REQ-AFL-008).
		if _, _, offered, err := namedOfferOf(ctx, tx, cardID, candidateSHA); err != nil {
			return err
		} else if offered {
			stands, err := landingStands(ctx, tx, cardID, candidateSHA)
			if err != nil {
				return err
			}
			if !stands {
				return fmt.Errorf("%w: commit %s was offered before %s was taken back for rework and needs a passing run recorded since, with no failing newest run", ErrEvidenceStale, candidateSHA, cardID)
			}
		}
		if takeOver {
			// The commit already is in the trunk: the row stays landed and only follows the card.
			if _, err := tx.ExecContext(ctx, `UPDATE merge_queue SET card_epoch=?, updated_at=`+nowSQL+` WHERE card_id=?`, card.Epoch, cardID); err != nil {
				return fmt.Errorf("ledger: %w", err)
			}
		} else if _, err := tx.ExecContext(ctx, `INSERT INTO merge_queue(card_id,candidate_sha,state,card_epoch,enqueued_at,updated_at) VALUES(?,?,'waiting',?,`+nowSQL+`,`+nowSQL+`)
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
