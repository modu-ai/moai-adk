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
// it named the wrong repository the wrong one stays bound. "Bound" means the current identity:
// once the repository has moved on disk and MoveRepository recorded it, the new identity is the one
// that binds as a no-op and the earlier ones are refused.
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

// Repository returns the identity the ledger is bound to now, or "" while it is not bound: the
// newest recorded move's new identity (MoveRepository), or the first binding while the repository
// has not moved.
func (s *Store) Repository(ctx context.Context) (string, error) {
	return currentRepository(ctx, s.db)
}
