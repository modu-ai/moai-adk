package homestate

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

// CanResumePROpenRepair reports whether card is the one shape the owner's
// repair resume acts on: `pr-open`, with label both non-empty and the recorded
// owner. Any other card or label is not resumable.
func CanResumePROpenRepair(card Card, label string) bool {
	label = strings.TrimSpace(label)
	return label != "" && card.State == CardPROpen && card.OwnerLabel == label
}

// ResumePROpenRepair lets the registered owner of a `pr-open` card take it
// back to repair it, in ONE transaction: the card returns to `assigned` at the
// run stage (worktree and evidence kept, `pr.repair_resumed` appended), then
// the existing lease edge (T3, which also enforces that label is a registered
// worker and the card's owner) leases it again. Either all of it commits or
// none of it does. It reads and writes the record only — never the card's
// worktree or its pull request. The edge is deliberately not a row of the
// requested-transition table: it is reachable only through this owner path.
//
// A card that is not resumable for label (see CanResumePROpenRepair) is left
// exactly as it was and the call reports false with no error, so the caller
// keeps its existing refusal. beforeCommit, when set, runs with the re-leased
// card inside the transaction, as TransitionRequest.BeforeCommit does.
func (f *FactoryDB) ResumePROpenRepair(ctx context.Context, runID, cardID, label string, now time.Time, beforeCommit func(Card) error) (Card, bool, error) {
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
		if !CanResumePROpenRepair(cur, label) {
			return nil, nil
		}
		next := cur
		next.State, next.Stage = CardAssigned, CardRun
		next.LeaseHolder, next.LeaseExpiresAt = "", ""
		next.Version = cur.Version + 1
		next.UpdatedAt = now.Format(time.RFC3339Nano)
		if err := updateCardRow(ctx, tx, next, cur.Version); err != nil {
			return nil, err
		}
		if err := appendEvent(ctx, tx, runID, "pr.repair_resumed", map[string]any{
			"card_id": cardID, "owner": label, "from": cur.State, "to": next.State, "version": next.Version,
		}, now); err != nil {
			return nil, err
		}
		leased, err := transitionTx(ctx, f, tx, TransitionRequest{
			RunID: runID, CardID: cardID, To: CardLeased,
			ExpectedVersion: next.Version, Actor: label, Now: now,
		}, now)
		if err != nil {
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
