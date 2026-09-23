package factorymsg

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

// seamDeadline bounds every call that could wait for the store's single
// connection. A call that returns well inside it did not wait for the pool.
const seamDeadline = 2 * time.Second

func TestStalePeerOutcomeMatchesErrStalePeer(t *testing.T) {
	s := openTestStore(t)
	_, b := registerPair(t, s)
	ctx := context.Background()

	stale := b
	stale.Generation++
	err := s.verifyPeer(ctx, stale)
	if !errors.Is(err, ErrStalePeer) {
		t.Fatalf("stale peer: errors.Is(%v, ErrStalePeer)=false", err)
	}
	if err.Error() != staleText {
		t.Fatalf("stale peer text=%q, want %q", err.Error(), staleText)
	}

	missing := b
	missing.Slot, missing.SessionUUID = "agent-9", "never-registered"
	if err := s.verifyPeer(ctx, missing); !errors.Is(err, ErrStalePeer) {
		t.Fatalf("unregistered peer: errors.Is(%v, ErrStalePeer)=false", err)
	}

	pending := b
	pending.SessionUUID = launchPendingSessionPrefix + "x"
	err = s.verifyPeer(ctx, pending)
	if !errors.Is(err, ErrEndpointLaunchPending) || errors.Is(err, ErrStalePeer) {
		t.Fatalf("launch-pending err=%v: want ErrEndpointLaunchPending and not ErrStalePeer", err)
	}

	invalid := b
	invalid.Backend = ""
	if err := s.verifyPeer(ctx, invalid); err == nil || errors.Is(err, ErrStalePeer) {
		t.Fatalf("invalid peer err=%v: want a validate error that is not ErrStalePeer", err)
	}
}

// Compile-time proof that every handle a caller may hold satisfies queryer.
var (
	_ queryer = (*sql.DB)(nil)
	_ queryer = (*sql.Tx)(nil)
	_ queryer = (*sql.Conn)(nil)
)

// richStale stands in for a stale error type defined by another package: the
// documented hook is an Is method answering true for ErrStalePeer.
type richStale struct{ slot string }

func (e richStale) Error() string        { return "stale endpoint for " + e.slot }
func (e richStale) Is(target error) bool { return target == ErrStalePeer }
func TestRicherStaleTypeMatchesViaIsHook(t *testing.T) {
	if !errors.Is(richStale{slot: "worker-1"}, ErrStalePeer) {
		t.Fatal("a type implementing Is(ErrStalePeer) must match ErrStalePeer")
	}
}

// TestVerifyPeerOnRunsInsideOpenTx proves the seam: with one pooled
// connection held by an open transaction, verifyPeerOn(tx) answers promptly,
// while the s.db path waits for that same connection until the deadline.
func TestVerifyPeerOnRunsInsideOpenTx(t *testing.T) {
	s := openTestStore(t)
	if got := s.db.Stats().MaxOpenConnections; got != 1 {
		t.Fatalf("precondition: MaxOpenConnections=%d, want 1", got)
	}
	_, b := registerPair(t, s)
	stale := b
	stale.Generation++

	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	for name, tc := range map[string]struct {
		p    Peer
		want error
	}{"registered": {b, nil}, "stale": {stale, ErrStalePeer}} {
		ctx, cancel := context.WithTimeout(context.Background(), seamDeadline)
		start := time.Now()
		err := s.verifyPeerOn(ctx, tx, tc.p)
		elapsed := time.Since(start)
		cancel()
		if elapsed >= seamDeadline/2 {
			t.Fatalf("%s: verifyPeerOn(tx) took %v, want well under %v", name, elapsed, seamDeadline)
		}
		if tc.want == nil && err != nil {
			t.Fatalf("%s: err=%v, want nil", name, err)
		}
		if tc.want != nil && !errors.Is(err, tc.want) {
			t.Fatalf("%s: err=%v, want %v", name, err, tc.want)
		}
	}

	// Negative control: the s.db wrapper needs the connection the open tx
	// holds, so it must run into the deadline. Without this the test above
	// could pass for a reason unrelated to the seam.
	ctx, cancel := context.WithTimeout(context.Background(), seamDeadline)
	defer cancel()
	start := time.Now()
	err = s.verifyPeer(ctx, b)
	elapsed := time.Since(start)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("control: s.db path inside open tx err=%v after %v, want context.DeadlineExceeded", err, elapsed)
	}
	if elapsed < seamDeadline*9/10 {
		t.Fatalf("control: s.db path returned after %v, want to wait ~%v", elapsed, seamDeadline)
	}
	t.Logf("control: s.db path blocked %v then %v", elapsed.Round(time.Millisecond), err)
}
