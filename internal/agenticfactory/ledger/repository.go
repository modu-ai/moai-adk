package ledger

import (
	"context"
	"errors"
)

// ErrWrongRepository is returned when a ledger already bound to one repository is asked to bind
// another.
var ErrWrongRepository = errors.New("ledger: the ledger is bound to another repository")

// BindRepository is the RED stub of the repository identity (N38): not implemented yet.
func (s *Store) BindRepository(ctx context.Context, identity string) error {
	return ErrNotImplemented
}

// Repository is the RED stub of the repository identity read (N38): not implemented yet.
func (s *Store) Repository(ctx context.Context) (string, error) {
	return "", ErrNotImplemented
}
