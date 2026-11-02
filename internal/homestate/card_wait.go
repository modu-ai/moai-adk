// card_wait.go — card t2013: a lane's wait recorded mechanically in the
// factory record. The wait names who must wake, and what condition releases
// it, so the chain end -> result pickup -> next action no longer depends on
// leader memory. The table is ADDITIVE (CREATE TABLE IF NOT EXISTS in
// factoryDDL, the precedent card_metadata_events set): no existing row is
// rewritten and the schema version is unchanged.
package homestate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// The wait condition kinds and the two non-lane wakers.
const (
	WaitCondPredecessorLanded = "predecessor-landed"
	WaitCondOperatorRuling    = "operator-ruling"
	WaitWakerLeader           = "leader"
	WaitWakerOperator         = "operator"
)

// CardWait is one recorded wait: Waker must be woken when the condition
// (CondKind, CondRef) is met. WokenAt is empty until the wake was handed out.
type CardWait struct {
	Seq        int64
	CardID     string
	CondKind   string
	CondRef    string
	Waker      string
	Reason     string
	RecordedBy string
	RecordedAt string
	WokenAt    string
}

const cardWaitColumns = `seq,card_id,cond_kind,cond_ref,waker,reason,recorded_by,recorded_at,woken_at`

func scanCardWait(row rowScanner) (CardWait, error) {
	var w CardWait
	err := row.Scan(&w.Seq, &w.CardID, &w.CondKind, &w.CondRef, &w.Waker, &w.Reason, &w.RecordedBy, &w.RecordedAt, &w.WokenAt)
	return w, err
}

func validateCardWait(w CardWait) error {
	if !ValidCardID(w.CardID) {
		return fmt.Errorf("%w: wait card %q is not a card id", ErrInvalidCardInput, w.CardID)
	}
	switch w.CondKind {
	case WaitCondPredecessorLanded:
		if !ValidCardID(w.CondRef) {
			return fmt.Errorf("%w: predecessor %q is not a card id", ErrInvalidCardInput, w.CondRef)
		}
	case WaitCondOperatorRuling:
	default:
		return fmt.Errorf("%w: wait condition %q is not %s or %s", ErrInvalidCardInput, w.CondKind, WaitCondPredecessorLanded, WaitCondOperatorRuling)
	}
	if strings.TrimSpace(w.Waker) == "" || strings.ContainsAny(w.Waker, " \t\r\n") {
		return fmt.Errorf("%w: waker %q must be one token (a lane name, leader, or operator)", ErrInvalidCardInput, w.Waker)
	}
	return nil
}

// RecordCardWait records a wait. A second record for the same card and
// condition that is still unwoken updates the waker and reason in place
// (idempotent); when waker, reason and recorder are all unchanged it performs
// no write at all. A woken wait is history and a new record opens a fresh one.
func (f *FactoryDB) RecordCardWait(ctx context.Context, w CardWait, now time.Time) error {
	if err := validateCardWait(w); err != nil {
		return err
	}
	if now.IsZero() {
		now = time.Now()
	}
	stamp := now.UTC().Format(time.RFC3339Nano)
	return retryFactoryBusy(ctx, func() error {
		tx, err := f.DB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		// SQL: every value goes through a ? placeholder.
		// Read the open (unwoken) wait first: an identical re-record is zero
		// writes - no UPDATE, no recorded_at or seq movement (card t2013).
		var curWaker, curReason, curBy string
		err = tx.QueryRowContext(ctx, `SELECT waker,reason,recorded_by FROM card_waits WHERE card_id=? AND cond_kind=? AND cond_ref=? AND woken_at='' ORDER BY seq LIMIT 1`,
			w.CardID, w.CondKind, w.CondRef).Scan(&curWaker, &curReason, &curBy)
		switch {
		case err == nil:
			if curWaker != w.Waker || curReason != w.Reason || curBy != w.RecordedBy {
				if _, err := tx.ExecContext(ctx, `UPDATE card_waits SET waker=?,reason=?,recorded_by=? WHERE card_id=? AND cond_kind=? AND cond_ref=? AND woken_at=''`,
					w.Waker, w.Reason, w.RecordedBy, w.CardID, w.CondKind, w.CondRef); err != nil {
					return err
				}
			}
		case errors.Is(err, sql.ErrNoRows):
			if _, err := tx.ExecContext(ctx, `INSERT INTO card_waits(card_id,cond_kind,cond_ref,waker,reason,recorded_by,recorded_at) VALUES(?,?,?,?,?,?,?)`,
				w.CardID, w.CondKind, w.CondRef, w.Waker, w.Reason, w.RecordedBy, stamp); err != nil {
				return err
			}
		default:
			return err
		}
		return tx.Commit()
	})
}

// LoadCardWaits returns every wait recorded for the card, oldest first.
func (f *FactoryDB) LoadCardWaits(ctx context.Context, cardID string) ([]CardWait, error) {
	// SQL: the column list is a constant; the card id goes through a ? placeholder.
	rows, err := f.DB.QueryContext(ctx, `SELECT `+cardWaitColumns+` FROM card_waits WHERE card_id=? ORDER BY seq`, cardID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []CardWait
	for rows.Next() {
		w, err := scanCardWait(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// SettleLandedWaits marks woken, in ONE transaction, every unwoken
// predecessor-landed wait on landedCardID and returns exactly those. A wait
// is handed out once: the woken_at stamp is written in the same update that
// selects it, so a re-run returns nothing. Operator-ruling waits never match.
func (f *FactoryDB) SettleLandedWaits(ctx context.Context, landedCardID string, now time.Time) ([]CardWait, error) {
	if now.IsZero() {
		now = time.Now()
	}
	stamp := now.UTC().Format(time.RFC3339Nano)
	var settled []CardWait
	err := retryFactoryBusy(ctx, func() error {
		settled = nil
		tx, err := f.DB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		// SQL: the column list is a constant; values go through ? placeholders.
		rows, err := tx.QueryContext(ctx, `SELECT `+cardWaitColumns+` FROM card_waits WHERE cond_kind=? AND cond_ref=? AND woken_at='' ORDER BY seq`,
			WaitCondPredecessorLanded, landedCardID)
		if err != nil {
			return err
		}
		var pending []CardWait
		for rows.Next() {
			w, err := scanCardWait(rows)
			if err != nil {
				_ = rows.Close()
				return err
			}
			pending = append(pending, w)
		}
		if err := rows.Close(); err != nil {
			return err
		}
		for _, w := range pending {
			res, err := tx.ExecContext(ctx, `UPDATE card_waits SET woken_at=? WHERE seq=? AND woken_at=''`, stamp, w.Seq)
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n == 1 {
				w.WokenAt = stamp
				settled = append(settled, w)
			}
		}
		return tx.Commit()
	})
	if err != nil {
		return nil, err
	}
	return settled, nil
}
