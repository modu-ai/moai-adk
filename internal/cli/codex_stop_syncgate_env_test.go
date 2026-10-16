package cli

import (
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

// The sync-gate child gets a from-scratch environment; the pinned
// GIT_CONFIG_GLOBAL must survive it so the child never reads the operator's
// real global git config (card t1801). Neither test calls t.Parallel.

func TestSyncGateEnv_PassesThroughGitConfigGlobal(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", "/pinned/gitconfig")
	got := syncGateEnv("/bin", "/proj", "EXTRA=1")
	if !slices.Contains(got, "GIT_CONFIG_GLOBAL=/pinned/gitconfig") {
		t.Fatalf("child env lacks the pinned GIT_CONFIG_GLOBAL: %v", got)
	}
	for _, want := range []string{"EXTRA=1", "PATH=/bin", "CLAUDE_PROJECT_DIR=/proj"} {
		if !slices.Contains(got, want) {
			t.Fatalf("child env lost %q: %v", want, got)
		}
	}
}

// The wiring: the command runClaudeSyncGate actually builds carries the pin, so
// reverting its Env to an inline form fails here (audit iter6 N3).
func TestRunClaudeSyncGate_ChildEnvCarriesThePin(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", "/pinned/gitconfig")
	var seen []string
	orig := syncGateOutput
	syncGateOutput = func(cmd *exec.Cmd) ([]byte, error) {
		seen = cmd.Env
		return nil, nil
	}
	t.Cleanup(func() { syncGateOutput = orig })

	f := &stopFixture{root: t.TempDir(), path: "/bin"}
	f.runClaudeSyncGate(t, "{}", "EXTRA=1")

	for _, want := range []string{"GIT_CONFIG_GLOBAL=/pinned/gitconfig", "EXTRA=1"} {
		if !slices.Contains(seen, want) {
			t.Fatalf("the sync-gate child command lacks %q: %v", want, seen)
		}
	}
}

func TestSyncGateEnv_DoesNotInventGitConfigGlobal(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", "") // registers the restore
	if err := os.Unsetenv("GIT_CONFIG_GLOBAL"); err != nil {
		t.Fatal(err)
	}
	for _, kv := range syncGateEnv("/bin", "/proj") {
		if strings.HasPrefix(kv, "GIT_CONFIG_GLOBAL=") {
			t.Fatalf("child env carries %q although the process has none", kv)
		}
	}
}
