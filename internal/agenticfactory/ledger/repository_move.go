package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// RepositoryMove is one recorded move of the repository this ledger belongs to: the identity the
// caller said was current, the identity it moved to, why, and when the database clock wrote it.
type RepositoryMove struct {
	Seq    int64
	From   string
	To     string
	Reason string
	At     time.Time
}

// MoveRepository records that the repository moved on disk: the ledger's current identity goes
// from `from` to `to`, for `reason`. The record is append-only (the table refuses an update and a
// delete) and the first binding is never touched. The ledger runs no git (REQ-AFL-008): both identities
// are the caller's strings, compared as written, and the record is only as right as the caller's
// computation of them.
//
// The move must start from the current identity (the newest move's `to`, or the first binding): a
// caller passes the identity it believes is current, and a different one is refused with
// ErrWrongRepository naming both, so a stale or racing caller cannot fork the history. A ledger
// that is not bound has no current identity and refuses every move the same way. A move to the
// identity the ledger already has is refused with ErrInvalid (nothing moved, and the history should
// not say it did), as is a blank `from`, `to` or `reason`. A move back to an earlier identity is an
// ordinary move.
func (s *Store) MoveRepository(ctx context.Context, from, to, reason string) error {
	if strings.TrimSpace(from) == "" || strings.TrimSpace(to) == "" || strings.TrimSpace(reason) == "" {
		return fmt.Errorf("%w: a repository move needs the identity it starts from, the identity it goes to, and a reason", ErrInvalid)
	}
	if from == to {
		return fmt.Errorf("%w: the repository is already at %q", ErrInvalid, to)
	}
	return s.write(ctx, func(tx *sql.Tx) error {
		current, err := currentRepository(ctx, tx)
		if err != nil {
			return err
		}
		switch {
		case current == "":
			return fmt.Errorf("%w: it is not bound to a repository, so there is nothing to move from %q", ErrWrongRepository, from)
		case current != from:
			return fmt.Errorf("%w: it is bound to %q, not %q", ErrWrongRepository, current, from)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO repository_move(old_identity,new_identity,reason,moved_at) VALUES(?,?,?,`+nowSQL+`)`, from, to, reason); err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		return nil
	})
}

// RepositoryMoves returns the recorded moves, oldest first. A ledger whose repository never moved
// has none.
func (s *Store) RepositoryMoves(ctx context.Context) ([]RepositoryMove, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT seq,old_identity,new_identity,reason,moved_at FROM repository_move ORDER BY seq`)
	if err != nil {
		return nil, fmt.Errorf("ledger: %w", err)
	}
	defer rows.Close() //nolint:errcheck // a read-only cursor
	var moves []RepositoryMove
	for rows.Next() {
		var m RepositoryMove
		var at string
		if err := rows.Scan(&m.Seq, &m.From, &m.To, &m.Reason, &at); err != nil {
			return nil, fmt.Errorf("ledger: %w", err)
		}
		if m.At, err = parseTime(at); err != nil {
			return nil, fmt.Errorf("ledger: repository move %d: %w", m.Seq, err)
		}
		moves = append(moves, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ledger: %w", err)
	}
	return moves, nil
}

// currentRepository is the identity the ledger answers to now: the newest move's new identity, or
// the first binding while the repository has not moved, or "" while the ledger is not bound.
func currentRepository(ctx context.Context, q querier) (string, error) {
	var identity sql.NullString
	if err := q.QueryRowContext(ctx, `SELECT COALESCE((SELECT new_identity FROM repository_move ORDER BY seq DESC LIMIT 1),(SELECT identity FROM repository WHERE id=1))`).Scan(&identity); err != nil {
		return "", fmt.Errorf("ledger: %w", err)
	}
	return identity.String, nil
}
