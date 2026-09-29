package hook

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// AC-012's hook half (SPEC-FACTORY-LANE-JOIN-SOCKET-001): a lane joined
// through discovery receives the resumed run's identity in its environment,
// and the EXISTING SessionStart bind chain — unchanged code (PRESERVE) —
// binds the lane peer (generation 1) into that run's broker. The launcher
// half of AC-012 (the env carries MOAI_KANBAN_ID + MOAI_KANBAN_LEAD_NAME) is
// asserted at the cli seam; this test proves the bind lands against a run
// restored by homestate.ResumeRun exactly as it would for one the lead
// recorded itself.
func TestFactoryHookBindsIntoResumedRun(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MOAI_HOME", filepath.Join(home, ".moai"))
	root := filepath.Join(t.TempDir(), "hook-project")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}

	// The resumed run: absent record, restored by the resume writer with the
	// verified leader's identity.
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err := db.ResumeRun(context.Background(), "runbind01", 424242, "1700000000.424242", `{"liveness":"pid+fingerprint"}`); err != nil {
		t.Fatalf("resume fixture run: %v", err)
	}

	// The child session's environment, exactly what the launcher exports on
	// the discovery path (REQ-009).
	t.Setenv(config.EnvMoaiKanbanID, "runbind01")
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-2")
	t.Setenv(config.EnvMoaiFactoryWorkers, "0")
	t.Setenv(config.EnvMoaiKanbanBackend, "glm")

	// The launcher half of the ordinary lane flow: exportFactoryLaunchFacts
	// stages the launch-pending peer entry the SessionStart bind consumes.
	// The hook code is unchanged (PRESERVE) — the fixture stages what it
	// reads, in the order production runs it.
	owner, start := factoryHookOwnerIdentity(t)
	s, err := factorymsg.Open(root, "runbind01")
	if err != nil {
		t.Fatalf("open resumed run's broker: %v", err)
	}
	defer closeFactoryHookStore(s)
	pending, err := s.RegisterLaunchPending(context.Background(), factorymsg.Peer{
		ProjectKey: homestate.ProjectKey(root), RunID: "runbind01", Backend: "glm",
		Role: "lane", Slot: "lane-2", PID: owner, ProcessStart: start,
	})
	if err != nil {
		t.Fatalf("stage launch-pending lane entry: %v", err)
	}

	input := &HookInput{SessionID: "sess-bind-01", ProjectDir: root}
	notice := registerFactorySessionStartPeer(context.Background(), input)
	if !strings.Contains(notice, "factory messaging bound") {
		t.Fatalf("bind notice = %q, want the bound confirmation", notice)
	}

	// The broker of the RESUMED run now carries the lane peer, bound exactly
	// as an ordinary join binds: generation = the staged launch-pending's + 1
	// (measured ordinary chain: pending 1 -> bind 2). AC-012's substance is
	// "the run identity and broker bind an ordinary join produces" — this
	// asserts that equality, not a hardcoded number.
	peer, err := s.Peer(context.Background(), "sess-bind-01")
	if err != nil {
		t.Fatalf("read bound peer: %v", err)
	}
	if peer.Role != "lane" || peer.Slot != "lane-2" {
		t.Errorf("bound peer = (%s,%s), want (lane,lane-2)", peer.Role, peer.Slot)
	}
	if peer.Generation != pending.Generation+1 {
		t.Errorf("bound peer generation = %d, want pending+1 (%d) — the ordinary join's bind shape", peer.Generation, pending.Generation+1)
	}
	if peer.RunID != "runbind01" {
		t.Errorf("bound peer run = %q, want the resumed run", peer.RunID)
	}
}
