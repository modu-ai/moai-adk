package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

// ReplayScope binds private conversation receipts to the account, not a rotating
// access token. Ownership and generation are checked under the credential lock.
func (s *Store) ReplayScope(ctx context.Context, ref CredentialRef) (string, error) {
	r, ok := ref.(*storeRef)
	if !ok || r == nil || r.s != s {
		return "", ErrCredentialAbsent
	}
	unlock, err := s.lock(ctx, "state.lock")
	if err != nil {
		return "", err
	}
	defer unlock()
	state, err := s.read()
	if err != nil {
		return "", err
	}
	if state.Tombstone {
		return "", ErrCredentialAbsent
	}
	if state.Generation != r.expected {
		return "", ErrCredentialChanged
	}
	data, _, err := parseAuth(state.Auth)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte("moai-openai-replay-account-v1\x00" + data.Tokens.Account))
	return hex.EncodeToString(digest[:]), nil
}
