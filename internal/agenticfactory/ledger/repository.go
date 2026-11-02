package ledger

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ErrWrongRepository is returned when a ledger already bound to one repository is asked to bind
// another.
var ErrWrongRepository = errors.New("ledger: the ledger is bound to another repository")

// BindRepository records the identity of the repository this ledger belongs to: the git common
// directory, as a plain string the caller computed (the ledger runs no git, REQ-AFL-008, and does not
// normalise the value; two spellings of one path are two identities). The first call writes it.
// The same identity again changes nothing. A different one is refused with ErrWrongRepository and
// the ledger stays as it was; the binding is never overwritten. The first caller decides: if
// it named the wrong repository the wrong one stays bound.
func (s *Store) BindRepository(ctx context.Context, identity string) error {
	if strings.TrimSpace(identity) == "" {
		return fmt.Errorf("%w: a repository identity cannot be blank", ErrInvalid)
	}
	return s.write(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO repository(id,identity,bound_at) VALUES(1,?,`+nowSQL+`) ON CONFLICT(id) DO NOTHING`, identity); err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		bound, err := currentRepository(ctx, tx)
		if err != nil {
			return err
		}
		if bound != identity {
			return fmt.Errorf("%w: it is bound to %q, not %q", ErrWrongRepository, bound, identity)
		}
		return nil
	})
}

// RecordRepositoryMarker records the marker of the repository this ledger belongs to: a token the caller keeps inside the
// repository's git directory, which is what tells that repository from a clone, a mirror or a new repository standing at
// the same path (the identity is a path, and a path is not a repository). The first call writes it; any later call
// changes nothing. The marker on record is returned either way, so callers that race converge on one value and a caller
// that brought another one can see it. The ledger stores the string as written and runs no git (REQ-AFL-008).
func (s *Store) RecordRepositoryMarker(ctx context.Context, marker string) (string, error) {
	if strings.TrimSpace(marker) == "" {
		return "", fmt.Errorf("%w: a repository marker cannot be blank", ErrInvalid)
	}
	var recorded string
	err := s.write(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO repository_marker(id,marker,recorded_at) VALUES(1,?,`+nowSQL+`) ON CONFLICT(id) DO NOTHING`, marker); err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		var err error
		recorded, err = currentMarker(ctx, tx)
		return err
	})
	return recorded, err
}

// RepositoryMarker returns the marker recorded for the repository, or "" while none is (a ledger bound before markers
// existed, or one no caller has marked yet).
func (s *Store) RepositoryMarker(ctx context.Context) (string, error) {
	return currentMarker(ctx, s.db)
}

func currentMarker(ctx context.Context, q querier) (string, error) {
	var marker sql.NullString
	if err := q.QueryRowContext(ctx, `SELECT (SELECT marker FROM repository_marker WHERE id=1)`).Scan(&marker); err != nil {
		return "", fmt.Errorf("ledger: %w", err)
	}
	return marker.String, nil
}

// Repository returns the identity the ledger is bound to, or "" while it is not bound.
func (s *Store) Repository(ctx context.Context) (string, error) {
	return currentRepository(ctx, s.db)
}

// currentRepository is the identity the ledger answers to: the first binding, or "" while the ledger is not bound.
func currentRepository(ctx context.Context, q querier) (string, error) {
	var identity sql.NullString
	if err := q.QueryRowContext(ctx, `SELECT identity FROM repository WHERE id=1`).Scan(&identity); err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("ledger: %w", err)
	}
	return identity.String, nil
}
