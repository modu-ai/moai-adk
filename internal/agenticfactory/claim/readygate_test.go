package claim

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// setRequest walks a request for the Issue through the real stages up to the preparer
// having stopped (WT_PREPARED); with ack it is then acknowledged by the agent.
func setRequest(t *testing.T, s *Store, key string, issue, agentN int, ack bool) {
	t.Helper()
	if _, _, err := s.PutRequest(key, "h-"+key, 1, pol); err != nil {
		t.Fatal(err)
	}
	step := func(fn func(*Request)) {
		t.Helper()
		if _, err := s.UpdateRequest(key, func(r *Request) error { fn(r); return nil }); err != nil {
			t.Fatal(err)
		}
	}
	step(func(r *Request) { r.Issue = issue })
	step(func(r *Request) { r.Stage = StageIssueConfirmed; r.Branch = &Branch{Ref: "WT-x", BaseOID: sha40} })
	step(func(r *Request) { r.Stage = StageBranchRegistered })
	a, err := s.RegisterAgent(sid(agentN))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReserveRequest(key, a, slots); err != nil {
		t.Fatal(err)
	}
	step(func(r *Request) {
		r.Stage, r.Ready = StageWTPrepared, &Ready{Host: "h", Slot: r.Slot, Path: "/wt", Head: sha40, WriterStopped: true}
	})
	if ack {
		if _, err := s.AckRequestReady(key, a); err != nil {
			t.Fatal(err)
		}
	}
}

// F21: several requests can name one Issue. The newest is the current attempt, so it
// is the one the gate reads; an older request is history and decides nothing, in
// either direction.
func TestReadyGateReadsTheNewestRequestOfTheIssue(t *testing.T) {
	t.Run("an old request that never became ready does not block a newer acknowledged one", func(t *testing.T) {
		s := open(t)
		setRequest(t, s, "old", 42, 41, false)
		setRequest(t, s, "new", 42, 42, true)
		if err := s.ReadyGate(42, sid(42)); err != nil {
			t.Errorf("the newest request is acknowledged: %v", err)
		}
	})
	t.Run("an old acknowledged request does not pass a newer one that is not ready", func(t *testing.T) {
		s := open(t)
		setRequest(t, s, "old", 42, 42, true)
		setRequest(t, s, "new", 42, 43, false)
		if err := s.ReadyGate(42, sid(42)); !errors.Is(err, core.ErrTransitionInvalid) {
			t.Errorf("the newest request is not ready: %v", err)
		}
	})
	t.Run("the newest request, not the first key in order, is read", func(t *testing.T) {
		s := open(t)
		setRequest(t, s, "zzz-old", 42, 42, true) // sorts last by name, oldest by Seq
		setRequest(t, s, "aaa-new", 42, 43, false)
		if err := s.ReadyGate(42, sid(42)); !errors.Is(err, core.ErrTransitionInvalid) {
			t.Errorf("names must not decide: %v", err)
		}
	})
	t.Run("another Issue's requests are not read", func(t *testing.T) {
		s := open(t)
		setRequest(t, s, "a", 42, 42, true)
		setRequest(t, s, "b", 43, 44, false)
		if err := s.ReadyGate(42, sid(42)); err != nil {
			t.Errorf("issue 42: %v", err)
		}
		if err := s.ReadyGate(43, sid(42)); !errors.Is(err, core.ErrTransitionInvalid) {
			t.Errorf("issue 43: %v", err)
		}
	})
}
