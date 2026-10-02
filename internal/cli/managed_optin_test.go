package cli

// managed_optin_test.go — SPEC-FACTORY-MANAGED-SESSION-001: the managed
// layer is an explicit opt-in (MOAI_FACTORY_MANAGED, default off). The
// divert needs BOTH the switch and the factory stamps; either alone leaves
// the launch on its ordinary door.
//
// Routing facts the Codex tests rest on (measured on this tree):
//   - `moai codex -l` routes to runCodexFactoryLane and returns before
//     runCodexLaunch; its per-card children launch through
//     launchCodexCardSession → codexDirectLaunchFn, never the divert.
//   - every other `-f` form is refused before any launch, so runCodexLaunch
//     never sees factoryEntry.Enabled from the real flag entry.
//   - the only reachable Codex divert shape is a plain `moai codex` run
//     whose PROCESS env already carries the stamps.

import (
	"os"
	"os/exec"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/spf13/cobra"
)

// managedOptIn turns the managed switch on for the test.
func managedOptIn(t *testing.T) {
	t.Helper()
	t.Setenv(config.EnvMoaiFactoryManaged, "1")
}

func TestFactoryManagedRequested(t *testing.T) {
	for _, tc := range []struct {
		env  []string
		want bool
	}{
		{nil, false},
		{[]string{config.EnvMoaiFactoryManaged + "="}, false},
		{[]string{config.EnvMoaiFactoryManaged + "=0"}, false},
		{[]string{config.EnvMoaiFactoryManaged + "=yes"}, false},
		{[]string{config.EnvMoaiFactoryManaged + "=1"}, true},
		{[]string{config.EnvMoaiFactoryManaged + "=true"}, true},
		{[]string{config.EnvMoaiFactoryManaged + "= TRUE "}, true},
	} {
		if got := factoryManagedRequested(tc.env); got != tc.want {
			t.Errorf("factoryManagedRequested(%v) = %v, want %v", tc.env, got, tc.want)
		}
	}
}

// captureClaudeExecDoor routes the exec handoff to a counter instead of the
// fixture's fail-if-reached stub.
func captureClaudeExecDoor(t *testing.T) *int {
	t.Helper()
	calls := 0
	orig := execOrSpawnClaudeFunc
	execOrSpawnClaudeFunc = func(string, []string, []string) error { calls++; return nil }
	t.Cleanup(func() { execOrSpawnClaudeFunc = orig })
	return &calls
}

func TestManagedLaunchRequiresOptIn(t *testing.T) {
	t.Run("stamps without the switch reach the exec door", func(t *testing.T) {
		_, managed := withManagedWiringFixture(t)
		exec := captureClaudeExecDoor(t)
		factoryLaneEnv(t)
		if err := runLaunchClaude("", nil); err != nil {
			t.Fatal(err)
		}
		if managed.calls != 0 || *exec != 1 {
			t.Fatalf("managed calls=%d exec calls=%d, want 0/1", managed.calls, *exec)
		}
	})
	t.Run("the switch without stamps reaches the exec door", func(t *testing.T) {
		_, managed := withManagedWiringFixture(t)
		exec := captureClaudeExecDoor(t)
		managedOptIn(t)
		if err := runLaunchClaude("", nil); err != nil {
			t.Fatal(err)
		}
		if managed.calls != 0 || *exec != 1 {
			t.Fatalf("managed calls=%d exec calls=%d, want 0/1", managed.calls, *exec)
		}
	})
	t.Run("switch and stamps divert", func(t *testing.T) {
		_, managed := withManagedWiringFixture(t)
		factoryLaneEnv(t)
		managedOptIn(t)
		if err := runLaunchClaude("", nil); err != nil {
			t.Fatal(err)
		}
		if managed.calls != 1 {
			t.Fatalf("managed calls=%d, want 1", managed.calls)
		}
	})
}

// captureCodexDirectDoor routes the direct door to a counter instead of the
// fixture's fail-if-reached stub.
func captureCodexDirectDoor(t *testing.T) *int {
	t.Helper()
	calls := 0
	orig := codexDirectLaunchFn
	codexDirectLaunchFn = func(*exec.Cmd) error { calls++; return nil }
	t.Cleanup(func() { codexDirectLaunchFn = orig })
	return &calls
}

func TestManagedCodexLaunchRequiresOptIn(t *testing.T) {
	t.Run("stamps in the process env without the switch reach the direct door", func(t *testing.T) {
		_, managed := withManagedCodexWiring(t)
		direct := captureCodexDirectDoor(t)
		factoryLaneEnv(t)
		if err := runCodex(&cobra.Command{Use: "codex"}, []string{"cli"}); err != nil {
			t.Fatal(err)
		}
		if managed.calls != 0 || *direct != 1 {
			t.Fatalf("managed calls=%d direct calls=%d, want 0/1", managed.calls, *direct)
		}
	})
	t.Run("the switch without stamps reaches the direct door", func(t *testing.T) {
		_, managed := withManagedCodexWiring(t)
		direct := captureCodexDirectDoor(t)
		managedOptIn(t)
		if err := runCodex(&cobra.Command{Use: "codex"}, []string{"cli"}); err != nil {
			t.Fatal(err)
		}
		if managed.calls != 0 || *direct != 1 {
			t.Fatalf("managed calls=%d direct calls=%d, want 0/1", managed.calls, *direct)
		}
	})
	t.Run("switch and stamps divert with the launch dir", func(t *testing.T) {
		root, managed := withManagedCodexWiring(t)
		factoryLaneEnv(t)
		managedOptIn(t)
		if err := runCodex(&cobra.Command{Use: "codex"}, []string{"cli"}); err != nil {
			t.Fatal(err)
		}
		if managed.calls != 1 || managed.dir != root {
			t.Fatalf("managed calls=%d dir=%q, want 1 call in %q", managed.calls, managed.dir, root)
		}
	})
}

// A managed Codex launch (switch + stamps, plain `moai codex`) leaves before
// the launcher's child-env assembly and exec-handoff steps, so under -d the
// trace lacks those lines and no RUST_LOG injection reaches the owner's env.
// The control run (no switch) proves the same launch traces both once.
func TestManagedCodexLaunchSkipsLaterDebugSteps(t *testing.T) {
	t.Run("control: ordinary launch traces both steps once", func(t *testing.T) {
		withManagedCodexWiring(t)
		captureCodexDirectDoor(t)
		factoryLaneEnv(t)
		_, stderr, err := runCodexCmd(t, "-d")
		if err != nil {
			t.Fatal(err)
		}
		for _, step := range []string{"child-env assembly", "exec handoff"} {
			if got := debugTraceStepLines(stderr, step); got != 1 {
				t.Errorf("control step %q traced %d times, want 1:\n%s", step, got, stderr)
			}
		}
	})
	t.Run("managed launch skips them and carries no RUST_LOG injection", func(t *testing.T) {
		_, managed := withManagedCodexWiring(t)
		factoryLaneEnv(t)
		managedOptIn(t)
		_, stderr, err := runCodexCmd(t, "-d")
		if err != nil {
			t.Fatal(err)
		}
		if managed.calls != 1 {
			t.Fatalf("managed calls=%d, want 1", managed.calls)
		}
		for _, step := range []string{"child-env assembly", "exec handoff"} {
			if got := debugTraceStepLines(stderr, step); got != 0 {
				t.Errorf("managed launch traced %q %d times; the divert is expected to leave first:\n%s", step, got, stderr)
			}
		}
		if _, set := os.LookupEnv(config.EnvRustLog); !set && managed.envValue(config.EnvRustLog) != "" {
			t.Errorf("managed env carries an injected %s", config.EnvRustLog)
		}
	})
}

// `moai codex -l` is a supervising loop: even with the switch and the
// stamps set, each card child launches through the direct-door seam and the
// managed owner is never consulted (the loop bypasses runCodexLaunch).
func TestManagedSwitchDoesNotReachCodexLaneLoop(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked, kanban.BacklogStatePicked)
	sdRecordLeaderRun(t, root, fcRun, kanban.BackendClaude)
	t.Chdir(root)
	sdScrubLauncherEnv(t)
	managedOptIn(t)
	t.Setenv(config.EnvMoaiKanbanID, fcRun)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")

	managedCalls, directCalls := 0, 0
	prevLook, prevDirect, prevManaged := codexLookPath, codexDirectLaunchFn, managedFactoryCodexLaunchFunc
	codexLookPath = func(string) (string, error) { return "/sentinel/codex", nil }
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		directCalls++
		env := sdEnvOf(t, c.Env)
		sdCodexSessionWork(t, root, env[config.EnvMoaiKanbanCard])
		return nil
	}
	managedFactoryCodexLaunchFunc = func(string, []string, []string, string) error { managedCalls++; return nil }
	t.Cleanup(func() {
		codexLookPath, codexDirectLaunchFn, managedFactoryCodexLaunchFunc = prevLook, prevDirect, prevManaged
	})

	if _, _, err := runCodexCmd(t, "-l"); err != nil {
		t.Fatalf("codex lane: %v", err)
	}
	if managedCalls != 0 {
		t.Fatalf("managed owner consulted by the lane loop: %d calls", managedCalls)
	}
	if directCalls != 2 {
		t.Fatalf("direct-door launches = %d, want 2 (one per picked card)", directCalls)
	}
}
