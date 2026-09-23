package factorymsg

import (
	"context"
	"database/sql"
)

// queryer is the read surface shared by *sql.DB and *sql.Tx, so checks that
// run inside a transaction never reach for a second connection (the store
// holds exactly one).
type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// ensureSchema brings an existing broker database up to the current schema.
// It is read-only when nothing is missing, so the hook hot path pays one
// catalogue lookup rather than a write lock.
func ensureSchema(ctx context.Context, db *sql.DB) error {
	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='dispatches'`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		if _, err := db.ExecContext(ctx, dispatchSchema); err != nil {
			return err
		}
	}
	return nil
}
