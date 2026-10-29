package homestate

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ReacquireExpiredLease returns the same owner's expired implementation lease
// and takes it again in one transaction. A changed row never gets collected.
func (f *FactoryDB) ReacquireExpiredLease(ctx context.Context, runID, cardID, lane string, version int64, now time.Time, beforeCommit func(Card) error) (Card, error) {
	if strings.TrimSpace(runID) == "" || !ValidCardID(cardID) || strings.TrimSpace(lane) == "" || now.IsZero() {
		return Card{}, fmt.Errorf("%w: run, card, lane and explicit clock are required", ErrInvalidCardInput)
	}
	var result Card
	now = now.UTC()
	err := f.withCardTx(ctx, runID, func(tx *sql.Tx) (func(), error) {
		cur, err := loadCard(ctx, tx, runID, cardID)
		if err != nil {
			return nil, err
		}
		if cur.Version != version {
			return nil, staleErr(cur, version)
		}
		if lane == "" || cur.OwnerLabel != lane || cur.LeaseHolder != lane || !cur.LeaseExpired(now) || cur.State == CardMerging {
			return nil, fmt.Errorf("%w: card %s is not an expired lease held by its owner", ErrLeaseHolder, cardID)
		}
		args := []any{cardID, runID}
		for _, s := range leaseHoldingStates {
			args = append(args, s)
		}
		rows, err := tx.QueryContext(ctx, liveLeaseCrossRunQuery, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var other, holder, expires string
			if err := rows.Scan(&other, &holder, &expires); err != nil {
				_ = rows.Close()
				return nil, err
			}
			at, err := time.Parse(time.RFC3339Nano, expires)
			if err != nil || at.After(now) {
				_ = rows.Close()
				return nil, fmt.Errorf("%w: card %s has a live lease in run %s", ErrLeaseHolder, cardID, other)
			}
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return nil, err
		}
		_ = rows.Close()
		bound, err := workerRunBinding(ctx, tx, lane)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(bound) != "" && bound != runID {
			return nil, fmt.Errorf("%w: lane %s is bound to run %s", ErrRosterBinding, lane, bound)
		}
		var registered string
		if err := tx.QueryRowContext(ctx, `SELECT registered_at FROM workers WHERE label=?`, lane).Scan(&registered); err != nil {
			return nil, err
		}
		reg, regErr := time.Parse(time.RFC3339Nano, registered)
		beat, beatErr := time.Parse(time.RFC3339Nano, cur.HeartbeatAt)
		if regErr != nil || beatErr != nil || reg.After(beat) {
			return nil, fmt.Errorf("%w: lane %s registration does not match the lease generation", ErrLeaseHolder, lane)
		}
		if err := applyLeaseExpiry(ctx, tx, cur, now); err != nil {
			return nil, err
		}
		req := TransitionRequest{RunID: runID, CardID: cardID, To: CardLeased, ExpectedVersion: cur.Version + 1, Actor: lane, Now: now}
		next, err := transitionTx(ctx, f, tx, req, now)
		if err != nil {
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
