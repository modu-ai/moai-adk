package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// leaseTx checks that the card's lease is at epoch, inside a write transaction. A lease
// that is gone, or at another epoch, is ErrStaleEpoch: the holder was fenced.
func leaseTx(ctx context.Context, q querier, cardID string, epoch int64) error {
	var held int64
	switch err := q.QueryRowContext(ctx, `SELECT epoch FROM leases WHERE card_id=?`, cardID).Scan(&held); {
	case err == sql.ErrNoRows:
		return fmt.Errorf("%w: %s holds no lease", ErrStaleEpoch, cardID)
	case err != nil:
		return fmt.Errorf("ledger: %w", err)
	case held != epoch:
		return fmt.Errorf("%w: %s is at epoch %d, not %d", ErrStaleEpoch, cardID, held, epoch)
	}
	return nil
}

// write runs fn in one write transaction.
func (s *Store) write(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // nothing to undo after a commit
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	return nil
}

// Heartbeat renews the lease from the database clock and keeps its length.
func (s *Store) Heartbeat(ctx context.Context, cardID string, epoch int64, nextWake *time.Time) error {
	var wake any
	if nextWake != nil {
		wake = nextWake.UTC().Format(timeLayout)
	}
	return s.write(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `UPDATE leases SET
			expires_at=strftime('%Y-%m-%dT%H:%M:%fZ','now', printf('+%.3f seconds',(julianday(expires_at)-julianday(heartbeat_at))*86400)),
			heartbeat_at=`+nowSQL+`, next_wake_at=? WHERE card_id=? AND epoch=?`, wake, cardID, epoch)
		if err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return leaseTx(ctx, tx, cardID, epoch)
		}
		return nil
	})
}

// Release drops the card's lease and keys; the card row stays.
func (s *Store) Release(ctx context.Context, cardID string, epoch int64) error {
	return s.write(ctx, func(tx *sql.Tx) error {
		if err := leaseTx(ctx, tx, cardID, epoch); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM lease_keys WHERE card_id=?`, cardID); err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM leases WHERE card_id=? AND epoch=?`, cardID, epoch); err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		return nil
	})
}

// ExpandKeys adds conflict keys to the lease: all of them, or none when another card
// holds one of them.
func (s *Store) ExpandKeys(ctx context.Context, cardID string, epoch int64, keys []string) error {
	for _, k := range keys {
		if strings.TrimSpace(k) == "" || strings.ContainsAny(k, " \t\r\n") {
			return fmt.Errorf("%w: key %q", ErrInvalid, k)
		}
	}
	return s.write(ctx, func(tx *sql.Tx) error {
		if err := leaseTx(ctx, tx, cardID, epoch); err != nil {
			return err
		}
		for _, k := range keys {
			var holder string
			switch err := tx.QueryRowContext(ctx, `SELECT card_id FROM lease_keys WHERE key=?`, k).Scan(&holder); {
			case err == sql.ErrNoRows:
			case err != nil:
				return fmt.Errorf("ledger: %w", err)
			case holder != cardID:
				return fmt.Errorf("%w: %s is held by %s", ErrKeyConflict, k, holder)
			}
		}
		for _, k := range keys {
			if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO lease_keys(key,card_id) VALUES(?,?)`, k, cardID); err != nil {
				return fmt.Errorf("ledger: %w", err)
			}
		}
		return nil
	})
}
