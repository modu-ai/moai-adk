package cli

// managed_optin_test.go — SPEC-FACTORY-MANAGED-SESSION-001: the managed
// layer is an explicit opt-in (MOAI_FACTORY_MANAGED, default off). The
// divert needs BOTH the switch and the factory stamps; either alone leaves
// the launch on its ordinary door.
//
// Routing facts the Codex tests rest on (measured on this tree):
//   - `moai codex -l` routes to runCodexFactoryLane and returns before
//     runCodexLaunch; its single parent session uses codexDirectLaunchFn.
//   - every other `-f` form is refused before any launch, so runCodexLaunch
//     never sees factoryEntry.Enabled from the real flag entry.
//   - the only reachable Codex divert shape is a plain `moai codex` run
//     whose PROCESS env already carries the stamps.

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
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
