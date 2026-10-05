package homestate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// RecordCardWorktree sets the worktree path of a card the caller holds — the
// record step of `factory next` leasing a card with no recorded worktree
// (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-011). It is version-checked like
// every other record write, refuses to move a card onto a different tree
// than the one already recorded (a card never uses another card's tree), and
// appends one `card.fields` event so the change is on the record's own log.
// Setting the recorded path again is idempotent.
func (f *FactoryDB) RecordCardWorktree(ctx context.Context, runID, cardID, path, actor string, now time.Time) (Card, error) {
	if strings.TrimSpace(runID) == "" || !ValidCardID(cardID) {
		return Card{}, fmt.Errorf("%w: run id and a valid card id are required", ErrInvalidCardInput)
	}
	path = strings.TrimSpace(path)
	if path == "" || !filepath.IsAbs(path) {
		return Card{}, fmt.Errorf("%w: worktree path %q is not absolute", ErrInvalidCardInput, path)
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
		if cur.WorktreePath == path {
			result = cur
			return nil, nil
		}
		if cur.WorktreePath != "" {
			return nil, fmt.Errorf("%w: card %s already records the worktree %s; a card never moves to another tree", ErrIllegalTransition, cur.CardID, cur.WorktreePath)
		}
		next := cur
		next.WorktreePath = path
		next.Version = cur.Version + 1
		next.UpdatedAt = now.Format(time.RFC3339Nano)
		if err := updateCardRow(ctx, tx, next, cur.Version); err != nil {
			return nil, err
		}
		if err := appendEvent(ctx, tx, runID, "card.fields", map[string]any{"card_id": cardID, "version": next.Version, "actor": actor, "worktree": path}, now); err != nil {
			return nil, err
		}
		result = next
		return nil, nil
	})
	if err != nil {
		return Card{}, err
	}
	return result, nil
}

// PreviousCardWorktree reads the newest worktree binding recorded for cardID
// in a run other than excludeRunID — the binding a run replacement may carry
// into the new run's row (card t1521). Card rows are keyed per run, so a
// replacement run's row is born without the binding its predecessor held;
// this read is what names that binding again. ok is false when no other run
// recorded a binding for the card. It never writes.
//
// @MX:NOTE: [AUTO] the cross-run binding read the lease-boundary carry-over is built on (card t1521)
// @MX:REASON: an ordering or exclusion change here changes which tree a replaced run re-enters — the one fact the REQ-SD-011 refusal is calibrated against
func (f *FactoryDB) PreviousCardWorktree(ctx context.Context, cardID, excludeRunID string) (string, bool, error) {
	// SQL: one constant statement; cardID and excludeRunID ride ? placeholders,
	// never the statement text.
	const query = `SELECT c.worktree_path
FROM cards c LEFT JOIN runs r ON r.run_id = c.run_id
WHERE c.card_id = ? AND c.run_id != ? AND c.worktree_path != ''
ORDER BY r.created_at DESC, c.updated_at DESC
LIMIT 1`
	var path string
	err := f.DB.QueryRowContext(ctx, query, cardID, excludeRunID).Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return path, true, nil
}
