// card_wake_events.go — card t2013: a wake is three distinct recorded events,
// so the chain end -> result pickup -> next action can tell what actually
// happened. `delivered` is written by SettleLandedWaits when the wake is handed
// out, `received` when the waker acknowledges it, `started` when the waker
// actually begins the awaited work. An acknowledgement is never a start:
// WakeState.Started is true only after a `started` event. The events live in
// the existing events table (no schema change); each is recorded at most once
// per wait, so duplicates and a restart write nothing.
package homestate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// The three wake phases; the event kind is "wake." + phase.
const (
	WakeDelivered = "delivered"
	WakeReceived  = "received"
	WakeStarted   = "started"
)

// WakeState reports which of the three wake events a wait has recorded.
type WakeState struct{ Delivered, Received, Started bool }

func wakeKind(phase string) (string, bool) {
	switch phase {
	case WakeDelivered, WakeReceived, WakeStarted:
		return "wake." + phase, true
	}
	return "", false
}

// wakeRecorded reports whether the wait already carries the event kind.
func wakeRecorded(ctx context.Context, tx *sql.Tx, waitSeq int64, kind string) (bool, error) {
	var one int
	// SQL: values go through ? placeholders.
	err := tx.QueryRowContext(ctx, `SELECT 1 FROM events WHERE kind=? AND json_extract(payload_json,'$.wait_seq')=? LIMIT 1`, kind, waitSeq).Scan(&one)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, sql.ErrNoRows):
		return false, nil
	}
	return false, err
}

// RecordWakeEvent records one wake phase for a wait. It reports true when the
// event was written and false (no error, no write) when the wait already
// carries it. An unknown wait or phase is ErrInvalidCardInput.
func (f *FactoryDB) RecordWakeEvent(ctx context.Context, waitSeq int64, phase, actor string, now time.Time) (bool, error) {
	kind, ok := wakeKind(phase)
	if !ok {
		return false, fmt.Errorf("%w: wake phase %q is not %s, %s or %s", ErrInvalidCardInput, phase, WakeDelivered, WakeReceived, WakeStarted)
	}
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	recorded := false
	err := retryFactoryBusy(ctx, func() error {
		recorded = false
		tx, err := f.DB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		var cardID string
		err = tx.QueryRowContext(ctx, `SELECT card_id FROM card_waits WHERE seq=?`, waitSeq).Scan(&cardID)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: no wait %d is recorded", ErrInvalidCardInput, waitSeq)
		}
		if err != nil {
			return err
		}
		seen, err := wakeRecorded(ctx, tx, waitSeq, kind)
		if err != nil {
			return err
		}
		if seen {
			return nil
		}
		if err := appendEvent(ctx, tx, "", kind, map[string]any{"wait_seq": waitSeq, "card_id": cardID, "actor": actor}, now); err != nil {
			return err
		}
		recorded = true
		return tx.Commit()
	})
	if err != nil {
		return false, err
	}
	return recorded, nil
}

// LoadWakeState reads which wake events the wait has recorded.
func (f *FactoryDB) LoadWakeState(ctx context.Context, waitSeq int64) (WakeState, error) {
	var st WakeState
	for phase, dst := range map[string]*bool{WakeDelivered: &st.Delivered, WakeReceived: &st.Received, WakeStarted: &st.Started} {
		kind, _ := wakeKind(phase)
		var one int
		// SQL: values go through ? placeholders.
		err := f.DB.QueryRowContext(ctx, `SELECT 1 FROM events WHERE kind=? AND json_extract(payload_json,'$.wait_seq')=? LIMIT 1`, kind, waitSeq).Scan(&one)
		switch {
		case err == nil:
			*dst = true
		case errors.Is(err, sql.ErrNoRows):
		default:
			return WakeState{}, err
		}
	}
	return st, nil
}
