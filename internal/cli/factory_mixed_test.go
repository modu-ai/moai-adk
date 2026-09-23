package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestFactoryLauncherRegistersLaunchPendingPeers(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	run := "run-launch-pending"
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{RunID: run, Backend: "codex", ManifestJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	start, state := homestate.ProbeProcessIdentity(os.Getpid())
	if state != homestate.ProcessIdentityLive || start == "" {
		t.Fatal("test process identity unavailable")
	}
	env := []string{
		config.EnvMoaiKanbanID + "=" + run,
		config.EnvMoaiKanbanBackend + "=codex",
		config.EnvMoaiFactoryWorkers + "=1",
	}
	peer, err := registerFactoryLaunchPending(context.Background(), root, env, os.Getpid(), start)
	if err != nil {
		t.Fatal(err)
	}
	if peer.Slot != "lead" || peer.PID != os.Getpid() || peer.ProcessStart != start {
		t.Fatalf("pending peer=%+v", peer)
	}
	s, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	closeOnCleanup(t, "factory message broker", s)
	status, err := s.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Lanes) != 1 || status.Lanes[0].BindingState != factorymsg.BindingLaunchPending || status.Lanes[0].SessionUUID != "" {
		t.Fatalf("pending roster leaked a session identity: %+v", status.Lanes)
	}
	if _, err := s.ResolveLane(context.Background(), "lead"); !errors.Is(err, factorymsg.ErrEndpointLaunchPending) {
		t.Fatalf("pending lane resolved for delivery: %v", err)
	}

	if peer, err := registerFactoryLaunchPending(context.Background(), root, os.Environ(), os.Getpid(), start); err != nil || peer != (factorymsg.Peer{}) {
		t.Fatalf("non-factory launch mutated broker: peer=%+v err=%v", peer, err)
	}
}

func TestFactoryRunSelectionAtomicSlotsAndArgv(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MOAI_HOME", t.TempDir())
	if _, err := factorymsg.ResolveActiveRun(context.Background(), root, ""); err == nil || !strings.Contains(err.Error(), "NO_ACTIVE_FACTORY") {
		t.Fatalf("zero=%v", err)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{RunID: "run-a", Backend: "codex", ManifestJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	if got, err := factorymsg.ResolveActiveRun(context.Background(), root, ""); err != nil || got != "run-a" {
		t.Fatalf("one=%q %v", got, err)
	}
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{RunID: "run-b", Backend: "claude", ManifestJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	if _, err := factorymsg.ResolveActiveRun(context.Background(), root, ""); err == nil || !strings.Contains(err.Error(), "AMBIGUOUS_FACTORY") {
		t.Fatalf("many=%v", err)
	}
	if got, err := factorymsg.ResolveActiveRun(context.Background(), root, "run-b"); err != nil || got != "run-b" {
		t.Fatalf("explicit=%q %v", got, err)
	}
	p, err := parseFactoryFlag([]string{"-f", "agent", "--factory-run", "run-b", "--", "--factory-run", "child", "x"})
	if err != nil {
		t.Fatal(err)
	}
	if p.RunID != "run-b" || !p.AgentRole || len(p.Rest) != 4 || p.Rest[0] != "--" || p.Rest[1] != "--factory-run" || p.Rest[2] != "child" || p.Rest[3] != "x" {
		t.Fatalf("parse=%+v", p)
	}
	s, err := factorymsg.Open(root, "run-b")
	if err != nil {
		t.Fatal(err)
	}
	closeOnCleanup(t, "factory message broker", s)
	start := homestate.CurrentProcessFingerprint()
	if start == "" {
		start = "test-start"
	}
	var wg sync.WaitGroup
	slots := make(chan string, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			peer := factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: "run-b", Backend: "codex", Role: "worker", Slot: "agent", SessionUUID: fmt.Sprintf("session-%d", i), Generation: 1, PID: os.Getpid(), ProcessStart: start}
			got, e := s.RegisterPeer(context.Background(), peer)
			if e != nil {
				t.Errorf("register: %v", e)
				return
			}
			slots <- got.Slot
		}(i)
	}
	wg.Wait()
	close(slots)
	seen := map[string]bool{}
	for slot := range slots {
		if seen[slot] {
			t.Fatalf("duplicate atomic slot %s", slot)
		}
		seen[slot] = true
	}
	if len(seen) != 8 {
		t.Fatalf("atomic slots=%d", len(seen))
	}
}

// closeOnCleanup closes c when the test finishes and reports a close failure
// as a test error instead of discarding it.
func closeOnCleanup(t *testing.T, what string, c io.Closer) {
	t.Helper()
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Errorf("close %s: %v", what, err)
		}
	})
}
