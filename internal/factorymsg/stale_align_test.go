package factorymsg

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestStaleEndpointErrorJoinsErrStalePeerClass proves a StaleEndpointError
// produced by a real stale path is in the ErrStalePeer class while still
// carrying the redirect metadata AC-FLH-007 requires.
func TestStaleEndpointErrorJoinsErrStalePeerClass(t *testing.T) {
	s := openTestStore(t)
	_, b := registerPair(t, s)
	stale := b
	stale.Generation++

	err := s.verifyPeer(context.Background(), stale)
	se, ok := StaleEndpoint(err)
	if !ok {
		t.Fatalf("stale generation: err=%v, want *StaleEndpointError", err)
	}
	if se.Code != NackStaleGeneration || se.Current.SessionUUID != b.SessionUUID || se.Current.Generation != b.Generation {
		t.Fatalf("redirect=%+v, want %s at %s gen %d", *se, NackStaleGeneration, b.SessionUUID, b.Generation)
	}
	if !errors.Is(err, ErrStalePeer) {
		t.Fatalf("errors.Is(%v, ErrStalePeer)=false", err)
	}
	if errors.Is(err, ErrEndpointLaunchPending) {
		t.Fatalf("stale endpoint matched ErrEndpointLaunchPending: %v", err)
	}
}

// TestVerifyPeerOnStaleInsideOpenTxKeepsRedirect proves the stale
// classification runs on the caller's handle: inside an open transaction on the
// one-connection pool it answers promptly with the redirect, instead of waiting
// on s.db for the connection the transaction holds.
func TestVerifyPeerOnStaleInsideOpenTxKeepsRedirect(t *testing.T) {
	s := openTestStore(t)
	if got := s.db.Stats().MaxOpenConnections; got != 1 {
		t.Fatalf("precondition: MaxOpenConnections=%d, want 1", got)
	}
	_, b := registerPair(t, s)
	stale := b
	stale.Generation++
	missing := b
	missing.Slot, missing.SessionUUID = "agent-9", "never-registered"

	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	check := func(name string, q queryer, p Peer, want func(error) bool) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), seamDeadline)
		defer cancel()
		start := time.Now()
		err := s.verifyPeerOn(ctx, q, p)
		elapsed := time.Since(start)
		if errors.Is(err, context.DeadlineExceeded) || elapsed >= seamDeadline/2 {
			t.Fatalf("%s: verifyPeerOn(tx) err=%v after %v — waited on the pool", name, err, elapsed)
		}
		if !want(err) {
			t.Fatalf("%s: unexpected err=%v", name, err)
		}
	}
	check("stale generation", tx, stale, func(err error) bool {
		se, ok := StaleEndpoint(err)
		return ok && errors.Is(err, ErrStalePeer) && se.Current.Generation == b.Generation
	})
	check("unregistered", tx, missing, func(err error) bool { return err == ErrStalePeer })
}
