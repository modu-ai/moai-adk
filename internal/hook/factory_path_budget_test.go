package hook

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The copied test executable records actual Git subprocess boundaries and then
// delegates to the real Git. It never synthesizes a successful path resolution.
func factoryBudgetGitHelper() {
	real := os.Getenv("MOAI_TEST_FACTORY_BUDGET_REAL_GIT")
	if real == "" || !strings.EqualFold(strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe"), "git") {
		return
	}
	f, err := os.OpenFile(os.Getenv("MOAI_TEST_FACTORY_BUDGET_GIT_LOG"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		os.Exit(2)
	}
	err = json.NewEncoder(f).Encode(os.Args[1:])
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		os.Exit(2)
	}
	cmd := exec.Command(real, os.Args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			os.Exit(exit.ExitCode())
		}
		os.Exit(1)
	}
	os.Exit(0)
}

func traceFactoryBudgetGit(t *testing.T) func() [][]string {
	t.Helper()
	real, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	name := "git"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if err := os.WriteFile(filepath.Join(dir, name), raw, 0o700); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(t.TempDir(), "git.jsonl")
	t.Setenv("MOAI_TEST_FACTORY_BUDGET_REAL_GIT", real)
	t.Setenv("MOAI_TEST_FACTORY_BUDGET_GIT_LOG", log)
	t.Setenv("GORACE", os.Getenv("GORACE")+" atexit_sleep_ms=0")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return func() [][]string {
		t.Helper()
		raw, err := os.ReadFile(log)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			t.Fatal(err)
		}
		var calls [][]string
		for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
			var args []string
			if err := json.Unmarshal([]byte(line), &args); err != nil {
				t.Fatal(err)
			}
			calls = append(calls, args)
		}
		return calls
	}
}

func TestFactoryRebindResolvesProjectIdentityOnce(t *testing.T) {
	rebindEnv(t, "runX", "lane-3", "claude")
	root := rebindRoot(t, "runX=retired", "runY=active")
	wantKey := homestate.ProjectKey(root)
	calls := traceFactoryBudgetGit(t)
	got := registerFactoryHookPeer(context.Background(), &HookInput{SessionID: "s1", ProjectDir: root}, factoryPeerBindUserPrompt)
	if !strings.Contains(got, "factory lane rebound:") {
		t.Fatalf("rebind did not complete: %q", got)
	}
	traced := calls()
	t.Logf("rebind Git subprocesses (%d): %q", len(traced), traced)
	// A non-Git root needs the primary and older-Git fallback probes, once.
	if len(traced) != 2 {
		t.Errorf("rebind resolved the project repeatedly: %d Git subprocesses, want 2", len(traced))
	}
	for _, args := range traced {
		if len(args) < 4 || args[0] != "-C" || args[1] != root || args[2] != "rev-parse" {
			t.Errorf("unexpected identity probe: %q", args)
		}
	}
	s, err := factorymsg.OpenExistingWithDeadline(root, "runY", factoryHookInspectionDeadline)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	p, err := s.Peer(context.Background(), "s1")
	if err != nil || p.ProjectKey != wantKey || p.RunID != "runY" || p.Slot != "lane-3" || p.PID != os.Getpid() || p.ProcessStart == "" {
		t.Fatalf("rebound peer lost its identity: %+v err=%v", p, err)
	}
}

func TestFactoryInboxSpentContextDoesNotResolveProject(t *testing.T) {
	_, _, _, _, input := factoryHookFixture(t)
	calls := traceFactoryBudgetGit(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, continuation, state := factoryHookBatch(ctx, input, EventStop)
	if continuation || !strings.HasPrefix(state, "degraded:") || !strings.Contains(state, context.Canceled.Error()) {
		t.Errorf("spent inbox context = continuation:%v state:%q", continuation, state)
	}
	if traced := calls(); len(traced) != 0 {
		t.Fatalf("spent inbox context still started Git: %q", traced)
	}
}
