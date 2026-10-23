package homestate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
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

// cardWorktreeTargetProbe is the registered-target check of the rebind
// operation (SPEC-WORKTREE-HOME-ROOT-001 REQ-WHR-011(b)): it verifies at
// call time that the target path is a registered worktree. The default
// probe is mechanical — a directory carrying a .git entry is registered —
// and the migration engine's journal carries the authoritative same-HEAD-
// and-branch identity (design.md D-12).
var cardWorktreeTargetProbe = defaultCardWorktreeTargetProbe

// defaultCardWorktreeTargetProbe reports whether path names a registered
// worktree: a directory holding a .git entry (file for a linked tree, dir
// for a checkout).
func defaultCardWorktreeTargetProbe(path string) (bool, error) {
	info, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil || !info.IsDir() && !info.Mode().IsRegular() {
		return false, nil
	}
	return true, nil
}

// RebindCardWorktree moves every run row of cardID that records oldPath onto
// newPath in one transaction (SPEC-WORKTREE-HOME-ROOT-001 REQ-WHR-011) — the
// recorded-path rebind the migration performs after `git worktree move`. It
// is version-checked like every record write, appends one card.fields event
// per rebound row recording both paths, and returns the number of rows
// rebound.
//
// Refusals: when no run row of the card records oldPath (a stale old path),
// or when the registered-target probe refuses newPath — the check runs at
// call time, so the migration's reverse rebind is exercised after the
// reverse move has re-registered the pre-move path. RecordCardWorktree's
// never-moves-a-card refusal is unchanged and stands beside this operation.
//
// @MX:ANCHOR: [AUTO] rebind — the recorded-path move the migration rebinds cards through
// @MX:REASON: M6's engine consumes this operation for forward and reverse moves; a rebind that dropped a run row would strand that run's card pointer on the pre-move path (REQ-WHR-011)
func (f *FactoryDB) RebindCardWorktree(ctx context.Context, cardID, oldPath, newPath, actor string, now time.Time) (int, error) {
	if !ValidCardID(cardID) {
		return 0, fmt.Errorf("%w: a valid card id is required", ErrInvalidCardInput)
	}
	if oldPath == "" || !filepath.IsAbs(oldPath) || newPath == "" || !filepath.IsAbs(newPath) {
		return 0, fmt.Errorf("%w: rebind paths must be absolute (old %q, new %q)", ErrInvalidCardInput, oldPath, newPath)
	}
	if oldPath == newPath {
		return 0, fmt.Errorf("%w: a rebind with equal old and new paths is a no-op", ErrInvalidCardInput)
	}
	registered, err := cardWorktreeTargetProbe(newPath)
	if err != nil {
		return 0, fmt.Errorf("rebind: registered-target check for %s: %w", newPath, err)
	}
	if !registered {
		return 0, fmt.Errorf("%w: rebind target %s is not a registered worktree", ErrIllegalTransition, newPath)
	}
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	rebound := 0
	err = f.withCardTx(ctx, cardID, func(tx *sql.Tx) (func(), error) {
		rows, err := tx.QueryContext(ctx, `SELECT run_id, version FROM cards WHERE card_id = ? AND worktree_path = ?`, cardID, oldPath)
		if err != nil {
			return nil, err
		}
		var targets []struct {
			runID   string
			version int
		}
		for rows.Next() {
			var runID string
			var version int
			if err := rows.Scan(&runID, &version); err != nil {
				_ = rows.Close()
				return nil, err
			}
			targets = append(targets, struct {
				runID   string
				version int
			}{runID, version})
		}
		_ = rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
		if len(targets) == 0 {
			return nil, fmt.Errorf("%w: no run row of card %s records the worktree %s; the rebind's old path is stale", ErrIllegalTransition, cardID, oldPath)
		}
		for _, tgt := range targets {
			cur, err := loadCard(ctx, tx, tgt.runID, cardID)
			if err != nil {
				return nil, err
			}
			if cur.WorktreePath != oldPath {
				// Raced with another writer between the query and the
				// update: this row no longer carries the old path.
				continue
			}
			next := cur
			next.WorktreePath = newPath
			next.Version = cur.Version + 1
			next.UpdatedAt = now.Format(time.RFC3339Nano)
			if err := updateCardRow(ctx, tx, next, cur.Version); err != nil {
				return nil, err
			}
			if err := appendEvent(ctx, tx, tgt.runID, "card.fields", map[string]any{
				"card_id":           cardID,
				"version":           next.Version,
				"actor":             actor,
				"worktree":          newPath,
				"previous_worktree": oldPath,
			}, now); err != nil {
				return nil, err
			}
			rebound++
		}
		return nil, nil
	})
	if err != nil {
		return 0, err
	}
	return rebound, nil
}

// CardRecordingPath returns the card id of the run row that records the
// given worktree path (SPEC-WORKTREE-HOME-ROOT-001 M6: the migration's
// rebind finds the card bound to the moved tree). ok is false when no run
// row records the path. It never writes.
func (f *FactoryDB) CardRecordingPath(ctx context.Context, path string) (string, bool, error) {
	const query = `SELECT card_id FROM cards WHERE worktree_path = ? ORDER BY updated_at DESC LIMIT 1`
	var cardID string
	err := f.DB.QueryRowContext(ctx, query, path).Scan(&cardID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return cardID, true, nil
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
