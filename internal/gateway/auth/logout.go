package auth

import (
	"context"
	"os"
	"path/filepath"
	"time"
)

// LogoutBroker must honor cancellation and return only after its subprocess has
// terminated. Its RPC completion does not establish remote token revocation.
type LogoutBroker interface {
	Logout(context.Context, string) error
}

type LogoutResult struct {
	Generation     uint64
	LocalCompleted bool
	BrokerOutcome  string
	RemoteOutcome  string
}

// LogoutWithBroker publishes local revocation before creating any broker scratch.
// It never takes operation.lock and never republishes the captured credential.
func (s *Store) LogoutWithBroker(ctx context.Context, broker LogoutBroker) (out LogoutResult, err error) {
	generation, snapshot, err := s.logoutSnapshot(ctx)
	if err != nil {
		return out, err
	}
	defer clear(snapshot)
	out = LogoutResult{Generation: generation, LocalCompleted: true, BrokerOutcome: "not_needed", RemoteOutcome: "not_attempted"}
	if len(snapshot) == 0 {
		return out, nil
	}
	if broker == nil {
		out.BrokerOutcome = "unavailable"
		return out, nil
	}
	if ctx.Err() != nil {
		out.BrokerOutcome = "failed"
		return out, ctx.Err()
	}
	home, err := privateScratch(s.dir, "logout-")
	if err != nil {
		out.BrokerOutcome = "failed"
		return out, ErrAuthState
	}
	defer func() {
		if os.RemoveAll(home) != nil {
			out.BrokerOutcome = "cleanup_failed"
			err = ErrAuthState
		}
	}()
	releaseHome, holdErr := holdPrivateDirectory(home)
	if holdErr != nil {
		out.BrokerOutcome = "failed"
		return out, ErrAuthState
	}
	defer releaseHome()
	if seedPrivateFile(filepath.Join(home, "auth.json"), snapshot) != nil {
		out.BrokerOutcome = "failed"
		return out, ErrAuthState
	}
	bounded, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	out.RemoteOutcome = "unknown"
	if broker.Logout(bounded, home) != nil || bounded.Err() != nil {
		out.BrokerOutcome = "failed"
		return out, nil
	}
	out.BrokerOutcome = "completed"
	return out, nil
}
