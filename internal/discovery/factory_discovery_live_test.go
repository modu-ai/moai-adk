package discovery

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// The live-process leg of AC-001's Given: a REAL process the REAL platform
// readers must verify — leader label in its argv, this project's canonical
// cwd, its own MOAI_KANBAN_ID env. Only the candidate LIST is injected
// (design.md §D: tests never depend on the host's socket population); the
// liveness probe, argv/env/cwd readers, and fingerprint are the production
// implementations.
//
// The helper child re-executes this test binary (the standard helper-process
// pattern): its argv carries `--name leader` after a non-flag argument, which
// stops the test framework's flag parsing — so the process is live, same-uid,
// and shaped exactly like the leader argv discovery reads.
func TestDiscoverLeaderVerifiesLiveProcessWithPlatformReaders(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skipf("live-process smoke runs where the platform readers exist; GOOS=%s degrades to a refusal", runtime.GOOS)
	}

	dir := t.TempDir() // REQ-012: outside this repository's worktree set

	cmd := exec.Command(os.Args[0], "-test.run=TestHelperLeaderChild$", "-test.v=true", "helper", "--name", "leader")
	cmd.Env = append(os.Environ(),
		"T1330_HELPER_LEADER=1",
		config.EnvMoaiKanbanID+"=runlive01",
	)
	cmd.Dir = dir
	if err := cmd.Start(); err != nil {
		t.Fatalf("start helper leader: %v", err)
	}
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_, _ = cmd.Process.Wait()
		}
	})

	// Wait for the child to be fully up (its fingerprint probe must answer
	// live) before discovery runs, bounded.
	deadline := time.Now().Add(10 * time.Second)
	for {
		if fp, state := probeIdentity(cmd.Process.Pid); state == "live" && fp != "" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("helper leader pid %d never became probe-live", cmd.Process.Pid)
		}
		time.Sleep(50 * time.Millisecond)
	}

	origCandidates := candidatePIDs
	candidatePIDs = func(context.Context, string) ([]int, error) { return []int{cmd.Process.Pid}, nil }
	t.Cleanup(func() { candidatePIDs = origCandidates })

	verified, err := DiscoverLeader(context.Background(), dir, "leader")
	if err != nil {
		t.Fatalf("DiscoverLeader: %v", err)
	}
	if len(verified) != 1 {
		t.Fatalf("verified = %d (%+v), want the live helper leader", len(verified), verified)
	}
	v := verified[0]
	if v.RunID != "runlive01" {
		t.Errorf("run id = %q, want runlive01 read from the helper's env", v.RunID)
	}
	if v.Name != "leader" {
		t.Errorf("name = %q, want leader read from the helper's argv", v.Name)
	}
	if v.PID != cmd.Process.Pid || v.ProcessStart == "" {
		t.Errorf("owner stamp = (%d,%q), want the helper's live identity", v.PID, v.ProcessStart)
	}
	if v.Basis == "" {
		t.Errorf("basis empty on a platform-verified leader")
	}
}

// TestHelperLeaderChild is the helper the live-process test re-executes. It
// blocks until killed; the parent's cleanup is the only exit.
func TestHelperLeaderChild(t *testing.T) {
	if os.Getenv("T1330_HELPER_LEADER") != "1" {
		t.Skip("helper only")
	}
	time.Sleep(5 * time.Minute)
}
