package factorymsg

import (
	"context"
	"errors"
	"testing"
)

// staleText is the message verifyPeer has always returned for a peer whose
// registered row does not match. Characterized here so a change to the
// outcome's shape cannot silently change its text.
const staleText = "stale or unregistered peer"

// TestCharacterizeVerifyPeerOutcomes pins the pre-existing verifyPeer outcomes
// (registered, each stale column, missing row, launch-pending, invalid shape).
func TestCharacterizeVerifyPeerOutcomes(t *testing.T) {
	s := openTestStore(t)
	_, b := registerPair(t, s)
	ctx := context.Background()

	if err := s.verifyPeer(ctx, b); err != nil {
		t.Fatalf("registered peer: err=%v, want nil", err)
	}

	stale := map[string]func(p *Peer){
		"wrong generation":    func(p *Peer) { p.Generation++ },
		"wrong session":       func(p *Peer) { p.SessionUUID = "other-session" },
		"wrong pid":           func(p *Peer) { p.PID++ },
		"wrong process_start": func(p *Peer) { p.ProcessStart = "other-start" },
		"missing row":         func(p *Peer) { p.Slot = "agent-9"; p.SessionUUID = "never-registered" },
	}
	for name, mutate := range stale {
		t.Run(name, func(t *testing.T) {
			p := b
			mutate(&p)
			err := s.verifyPeer(ctx, p)
			if err == nil || err.Error() != staleText {
				t.Fatalf("err=%v, want %q", err, staleText)
			}
			if errors.Is(err, ErrEndpointLaunchPending) {
				t.Fatalf("stale outcome matched ErrEndpointLaunchPending: %v", err)
			}
		})
	}

	t.Run("launch-pending session", func(t *testing.T) {
		p := b
		p.SessionUUID = launchPendingSessionPrefix + "x"
		if err := s.verifyPeer(ctx, p); !errors.Is(err, ErrEndpointLaunchPending) {
			t.Fatalf("err=%v, want ErrEndpointLaunchPending", err)
		}
	})

	invalid := map[string]struct {
		mutate func(p *Peer)
		want   string
	}{
		"invalid backend":   {func(p *Peer) { p.Backend = "" }, "invalid backend"},
		"identity mismatch": {func(p *Peer) { p.RunID = "other-run" }, "factory identity mismatch"},
		"invalid pid":       {func(p *Peer) { p.PID = 0 }, "invalid generation or pid"},
	}
	for name, tc := range invalid {
		t.Run(name, func(t *testing.T) {
			p := b
			tc.mutate(&p)
			err := s.verifyPeer(ctx, p)
			if err == nil || err.Error() != tc.want {
				t.Fatalf("err=%v, want %q", err, tc.want)
			}
		})
	}
}
