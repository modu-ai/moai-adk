package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/discovery"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// t1628 — the codex -l launch past the factory join, with one verified live
// leader. The child session must start in the parent checkout with the local
// instructions injected, and carry the lane stamps the factory reads. The
// child launch is captured at its seam (codexDirectLaunchFn); nothing starts.
func TestCodexLaneJoinsDiscoveredLeader(t *testing.T) {
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	t1628Enter(t, root)
	t1628StubCodex(t)
	asked := stageDiscoveredLeaders(t, []discovery.VerifiedLeader{verifiedTestLeader("runlead01")})
	if err := os.WriteFile(filepath.Join(root, "AGENTS.local.md"), []byte("t1628-marker local instruction\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var launched *exec.Cmd
	origLaunch := codexDirectLaunchFn
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		launched = c
		return nil
	}
	t.Cleanup(func() { codexDirectLaunchFn = origLaunch })

	if err := t1628RunLane(t); err != nil {
		t.Fatalf("codex -l with one verified leader: %v", err)
	}
	if launched == nil {
		t.Fatal("the lane launch never reached the child launch seam")
	}
	if len(*asked) != 1 || (*asked)[0] != "leader" {
		t.Errorf("discovery asked = %v, want exactly [leader]", *asked)
	}

	// -C names the parent checkout (the factory root), compared through symlinks.
	if got := t1628ArgAfter(launched.Args, "-C"); !t1628SameRealPath(got, root) {
		t.Errorf("codex -C = %q, want the parent checkout %q", got, root)
	}
	// The local instruction file is injected as the developer instructions.
	if !strings.Contains(strings.Join(launched.Args, "\n"), "t1628-marker") {
		t.Errorf("the local instruction file was not injected into the child args: %q", launched.Args)
	}
	// The lane stamps the factory reads: lane role, the verified leader's run,
	// the claimed lane label, and the backend.
	env := t1628EnvMap(launched.Env)
	if env[config.EnvFactoryRole] != config.FactoryRoleLane {
		t.Errorf("%s = %q, want the lane role", config.EnvFactoryRole, env[config.EnvFactoryRole])
	}
	if env[config.EnvFactoryRunID] != "runlead01" {
		t.Errorf("%s = %q, want the verified leader's run", config.EnvFactoryRunID, env[config.EnvFactoryRunID])
	}
	if env[config.EnvMoaiFactoryWorker] == "" {
		t.Errorf("%s is empty: the lane label was not claimed", config.EnvMoaiFactoryWorker)
	}
	if env[config.EnvFactoryBackend] != factory.BackendGPT {
		t.Errorf("%s = %q, want %q", config.EnvFactoryBackend, env[config.EnvFactoryBackend], factory.BackendGPT)
	}
}

// t1628ArgAfter returns the argument that follows flag, or "".
func t1628ArgAfter(args []string, flag string) string {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag {
			return args[i+1]
		}
	}
	return ""
}

// t1628SameRealPath reports whether a and b name the same directory after
// symlink resolution (macOS temporary directories live under a symlink).
func t1628SameRealPath(a, b string) bool {
	ra, errA := filepath.EvalSymlinks(a)
	rb, errB := filepath.EvalSymlinks(b)
	return errA == nil && errB == nil && ra == rb
}

// t1628EnvMap turns a KEY=VALUE slice into a map; later entries win.
func t1628EnvMap(env []string) map[string]string {
	out := map[string]string{}
	for _, kv := range env {
		if k, v, ok := strings.Cut(kv, "="); ok {
			out[k] = v
		}
	}
	return out
}
