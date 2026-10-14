package homestate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ReassignCard moves a card to another registered lane (card t1696): lanes
// that relaunch come back under fresh labels and never reuse the old one, so
// a card owned by a dead label had no path back into a lease — the holder
// guard refused every actor, and `assign --to` only moves a picked card. It
// is the leader's recovery path and admits two shapes:
//
//   - an `assigned` card (defect D2): the owner swaps, the state stays;
//   - a lease-holding card whose lease expired (defect D4, lane-17's
//     state=run shape): the expiry lands like T27/T28 — an expired card
//     returns to `assigned`, and one that expired MID-MERGE lands `blocked`
//     with the interruption recorded (applyLeaseExpiry's merging branch),
//     because unblocking a merge is the operator's decision — under the new
//     owner, whose next lease edge is then the ordinary arm (a).
//
// A live lease refuses. The new owner must be registered in the workers
// table AND bound to this run (an empty binding admits — RosterSoleRun; a
// binding to another run refuses, or the lane's first renewal wedges on the
// roster rule, sync-audit F2 of card t1696). The `card.reassign` event names
// the previous owner.
//
// beforeCommit (nil = none) runs inside the transaction, once the new row and
// the event are written, right before the commit — the queue's
// current-dispatch mirror follows the same pattern as TransitionRequest's
// BeforeCommit, so a mirror that cannot be written rolls the reassignment
// back.
func (f *FactoryDB) ReassignCard(ctx context.Context, runID, cardID, newOwner, actor string, now time.Time, beforeCommit func(Card) error) (Card, error) {
	newOwner = strings.TrimSpace(newOwner)
	if strings.TrimSpace(runID) == "" || !ValidCardID(cardID) {
		return Card{}, fmt.Errorf("%w: run id and a valid card id are required", ErrInvalidCardInput)
	}
	if newOwner == "" {
		return Card{}, fmt.Errorf("%w: reassignment requires an owner label", ErrInvalidCardInput)
	}
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	var result Card
	err := f.withCardTx(ctx, runID, func(tx *sql.Tx) (func(), error) {
		cur, err := loadCard(ctx, tx, runID, cardID)
		if err != nil {
			return nil, err
		}
		holding := IsLeaseHoldingState(cur.State)
		expired := holding && cur.LeaseExpired(now)
		switch {
		case cur.Legacy():
			return nil, fmt.Errorf("%w: card %s holds pre-F1 state %q; only an operator abandon is accepted", ErrLegacyState, cur.CardID, cur.State)
		case holding && !expired:
			return nil, fmt.Errorf("%w: card %s holds a live lease (holder %q, expires %s); reassignment waits out the lease or follows its expiry", ErrLeaseHolder, cur.CardID, cur.LeaseHolder, cur.LeaseExpiresAt)
		case !holding && cur.State != CardAssigned:
			return nil, fmt.Errorf("%w: card %s is %s; reassignment moves an assigned card, or a card whose lease expired", ErrIllegalTransition, cur.CardID, cur.State)
		case newOwner == cur.OwnerLabel:
			return nil, fmt.Errorf("%w: card %s is already assigned to %q", ErrInvalidCardInput, cur.CardID, newOwner)
		}
		var bound string
		if err := tx.QueryRowContext(ctx, `SELECT run_id FROM workers WHERE label=?`, newOwner).Scan(&bound); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("%w: %q is not a registered lane; the lane joins (and registers) before a card can be reassigned to it", ErrInvalidCardInput, newOwner)
			}
			return nil, err
		}
		if bound != "" && bound != runID {
			// RosterSoleRun admission: a lane bound to another run takes the
			// card and then wedges on its first roster-checked renewal.
			return nil, fmt.Errorf("%w: %q is bound to run %s, not to %s; reassign to a lane of this run", ErrInvalidCardInput, newOwner, bound, runID)
		}
		next := cur
		next.OwnerLabel = newOwner
		next.LeaseHolder, next.LeaseExpiresAt = "", ""
		if holding {
			// The expired-lease landing (T27/T28), its merging branch
			// included: an expired card returns to `assigned`; one that
			// expired mid-merge lands `blocked` — the unblock is the
			// operator's decision.
			next.State = CardAssigned
			if cur.State == CardMerging {
				next.State = CardBlocked
			}
		}
		next.Version = cur.Version + 1
		next.UpdatedAt = now.Format(time.RFC3339Nano)
		payload := map[string]any{
			"card_id": cur.CardID, "from": cur.State, "to": next.State, "version": next.Version,
			"actor": actor, "prev_owner": cur.OwnerLabel, "new_owner": newOwner,
		}
		if holding && cur.State == CardMerging {
			payload["interrupted"] = "merge"
			payload["merge_sha"] = cur.MergeSHA
		}
		if err := updateCardRow(ctx, tx, next, cur.Version); err != nil {
			return nil, err
		}
		if err := appendEvent(ctx, tx, cur.RunID, "card.reassign", payload, now); err != nil {
			return nil, err
		}
		if beforeCommit != nil {
			if err := beforeCommit(next); err != nil {
				return nil, err
			}
		}
		result = next
		return nil, nil
	})
	if err != nil {
		return Card{}, err
	}
	return result, nil
}
