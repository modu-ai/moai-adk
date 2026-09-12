package auth

import (
	"context"
	"time"
)

// ResolveFresh resolves only an existing explicit login. An expired credential
// uses the existing serialized Refresh transaction; a missing login never runs
// the broker. Returned refs still require SendAuthorized's generation barrier.
func (s *Store) ResolveFresh(ctx context.Context, broker Broker, verify func(context.Context, CredentialRef) error) (CredentialRef, error) {
	before, err := s.Status(ctx)
	if err != nil {
		return nil, err
	}
	if !before.LoggedIn {
		return nil, ErrCredentialAbsent
	}
	if before.ExpiresAt.After(time.Now()) {
		return &storeRef{s: s, expected: before.Generation}, nil
	}
	_, refreshErr := s.Refresh(ctx, before.Generation, broker, verify)
	// A concurrent login/refresh may have published a newer valid generation.
	// Re-read under the caller's context, including after a failed stale attempt.
	after, err := s.Status(ctx)
	if err != nil {
		return nil, err
	}
	if !after.LoggedIn {
		return nil, ErrCredentialAbsent
	}
	if after.ExpiresAt.After(time.Now()) && after.Generation != before.Generation {
		return &storeRef{s: s, expected: after.Generation}, nil
	}
	if refreshErr != nil {
		return nil, refreshErr
	}
	return nil, ErrRefresh
}
