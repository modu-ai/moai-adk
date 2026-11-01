package ledger

import (
	"context"
	"fmt"
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

// MoveRepository is the parent tree's behavior: the binding is final, so no move can be recorded.
func (s *Store) MoveRepository(ctx context.Context, from, to, reason string) error {
	_, _, _, _ = ctx, from, to, reason
	return fmt.Errorf("ledger: the repository identity is bound once")
}

// RepositoryMoves is the parent tree's behavior: nothing is ever recorded.
func (s *Store) RepositoryMoves(ctx context.Context) ([]RepositoryMove, error) {
	_ = ctx
	return nil, nil
}
