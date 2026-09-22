package factorymsg

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestBindLaunchPendingCannotOverwriteConcurrentAuthoritativePeer(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	for i := 0; i < 8; i++ {
		root := filepath.Join(t.TempDir(), "project")
		run := "run-bind-race"
		sessionStore, err := Open(root, run)
		if err != nil {
			t.Fatal(err)
		}
		promptStore, err := Open(root, run)
		if err != nil {
			_ = sessionStore.Close()
			t.Fatal(err)
		}
		start, state := homestate.ProbeProcessIdentity(os.Getpid())
		if state != homestate.ProcessIdentityLive || start == "" {
			t.Fatal("test process identity unavailable")
		}
		pending, err := sessionStore.RegisterLaunchPending(context.Background(), Peer{
			ProjectKey: "project", RunID: run, Backend: "codex", Role: "lead", Slot: "lead",
			PID: os.Getpid(), ProcessStart: start,
		})
		if err != nil {
			t.Fatal(err)
		}
		alias := pending
		alias.SessionUUID = "session-start-alias"
		actual := pending
		actual.SessionUUID = "actual-user-prompt-session"

		startRace := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		var bindErr, promptErr error
		go func() {
			defer wg.Done()
			<-startRace
			_, _, bindErr = sessionStore.BindLaunchPending(context.Background(), alias)
		}()
		go func() {
			defer wg.Done()
			<-startRace
			_, promptErr = promptStore.RegisterPeer(context.Background(), actual)
		}()
		close(startRace)
		wg.Wait()
		if bindErr != nil || promptErr != nil {
			t.Fatalf("iteration %d bind=%v prompt=%v", i, bindErr, promptErr)
		}
		got, err := sessionStore.ResolveLane(context.Background(), "lead")
		if err != nil {
			t.Fatal(err)
		}
		if got.SessionUUID != actual.SessionUUID {
			t.Fatalf("iteration %d SessionStart overwrote authoritative peer: %+v", i, got)
		}
		_ = promptStore.Close()
		_ = sessionStore.Close()
	}
}

func TestBindLaunchPendingRejectsOwnerMismatch(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	run := "run-bind-owner-mismatch"
	s, err := Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	start, state := homestate.ProbeProcessIdentity(os.Getpid())
	if state != homestate.ProcessIdentityLive || start == "" {
		t.Fatal("test process identity unavailable")
	}
	pending, err := s.RegisterLaunchPending(context.Background(), Peer{
		ProjectKey: "project", RunID: run, Backend: "codex", Role: "lead", Slot: "lead",
		PID: os.Getpid(), ProcessStart: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	mismatch := pending
	mismatch.SessionUUID = "session-start-alias"
	mismatch.ProcessStart = "different-owner"
	if _, _, err := s.BindLaunchPending(context.Background(), mismatch); err == nil {
		t.Fatal("pending owner mismatch accepted")
	}
	status, err := s.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Lanes) != 1 || status.Lanes[0].BindingState != BindingLaunchPending || status.Lanes[0].Generation != pending.Generation {
		t.Fatalf("owner mismatch rewrote pending row: %+v", status.Lanes)
	}
}

func TestRollbackLaunchPendingDeletesOnlyExactProvisionalOwner(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	s, err := Open(root, "run-rollback")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	start := homestate.CurrentProcessFingerprint()
	pending, err := s.RegisterLaunchPending(context.Background(), Peer{
		ProjectKey: homestate.ProjectKey(root), RunID: "run-rollback", Backend: "codex",
		Role: "lead", Slot: "lead", PID: os.Getpid(), ProcessStart: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	mismatch := pending
	mismatch.ProcessStart += "-mismatch"
	if deleted, err := s.RollbackLaunchPending(context.Background(), mismatch); err != nil || deleted {
		t.Fatalf("mismatched owner rollback=(%v,%v)", deleted, err)
	}
	status, err := s.Status(context.Background())
	if err != nil || len(status.Lanes) != 1 || status.Lanes[0].BindingState != BindingLaunchPending {
		t.Fatalf("pending removed by mismatch: status=%+v err=%v", status, err)
	}
	bound := pending
	bound.SessionUUID = "actual-session"
	if _, err := s.RegisterPeer(context.Background(), bound); err != nil {
		t.Fatal(err)
	}
	if deleted, err := s.RollbackLaunchPending(context.Background(), pending); err != nil || deleted {
		t.Fatalf("rebound row rollback=(%v,%v)", deleted, err)
	}
	resolved, err := s.ResolveLane(context.Background(), "lead")
	if err != nil || resolved.SessionUUID != "actual-session" {
		t.Fatalf("bound row removed: peer=%+v err=%v", resolved, err)
	}

	worker, err := s.RegisterLaunchPending(context.Background(), Peer{
		ProjectKey: homestate.ProjectKey(root), RunID: "run-rollback", Backend: "codex",
		Role: "worker", Slot: "agent-1", PID: os.Getpid(), ProcessStart: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if deleted, err := s.RollbackLaunchPending(context.Background(), worker); err != nil || !deleted {
		t.Fatalf("exact pending rollback=(%v,%v)", deleted, err)
	}
	status, err = s.Status(context.Background())
	if err != nil || len(status.Lanes) != 1 || status.Lanes[0].Slot != "lead" {
		t.Fatalf("exact rollback affected other row: status=%+v err=%v", status, err)
	}
}
