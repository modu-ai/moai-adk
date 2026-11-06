package homestate

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

// IsReacquirableExpiredLease reports whether card is the one shape the owner's
// own re-lease acts on: `leased`, its lease expired at now, and label is both
// the recorded owner and the lease holder. Any other card, label, or an
// unexpired lease is not reacquirable.
func IsReacquirableExpiredLease(card Card, label string, now time.Time) bool {
	label = strings.TrimSpace(label)
	return label != "" && card.State == CardLeased && card.LeaseExpired(now) &&
		card.LeaseHolder == label && card.OwnerLabel == label
}

// ReacquireExpiredLease lets the registered owner of a `leased` card whose lease
// expired take it again in ONE transaction: the existing expiry return (T27,
// applyLeaseExpiry — the card goes to `assigned`, stage, worktree and evidence
// kept, `lease.expired` appended), then the existing lease edge (T3, which also
// enforces that label is a registered worker and the card's owner), then a
// `lease.reacquired` event. Either all of it commits or none of it does. It
// reads and writes the record only — never the card's worktree.
//
// A card that is not reacquirable for label (see IsReacquirableExpiredLease)
// is left exactly as it was and the call reports false with no error, so the
// caller keeps its existing refusal. beforeCommit, when set, runs with the
// re-leased card inside the transaction, as TransitionRequest.BeforeCommit does.
func (f *FactoryDB) ReacquireExpiredLease(ctx context.Context, runID, cardID, label string, now time.Time, beforeCommit func(Card) error) (Card, bool, error) {
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	label = strings.TrimSpace(label)
	var (
		result Card
		done   bool
	)
	err := f.withCardTx(ctx, runID, func(tx *sql.Tx) (func(), error) {
		cur, err := loadCard(ctx, tx, runID, cardID)
		if err != nil {
			return nil, err
		}
		if !IsReacquirableExpiredLease(cur, label, now) {
			return nil, nil
		}
		if err := applyLeaseExpiry(ctx, tx, cur, now); err != nil {
			return nil, err
		}
		leased, err := transitionTx(ctx, f, tx, TransitionRequest{
			RunID: runID, CardID: cardID, To: CardLeased,
			ExpectedVersion: cur.Version + 1, Actor: label, Now: now,
		}, now)
		if err != nil {
			return nil, err
		}
		if err := appendEvent(ctx, tx, runID, "lease.reacquired", map[string]any{
			"card_id": cardID, "holder": label, "expired_at": cur.LeaseExpiresAt,
			"expires_at": leased.LeaseExpiresAt, "from_version": cur.Version, "version": leased.Version,
		}, now); err != nil {
			return nil, err
		}
		if beforeCommit != nil {
			if err := beforeCommit(leased); err != nil {
				return nil, err
			}
		}
		result, done = leased, true
		return nil, nil
	})
	if err != nil {
		return Card{}, false, err
	}
	return result, done, nil
}
