package cli

// managed_launcher_wiring_test.go — SPEC-FACTORY-MANAGED-SESSION-001 M3
// launcher wiring (design.md D-7): a factory leader/lane launch — the env
// carries the factory stamps — enters the managed session owner instead of
// the exec/spawn handoff, and a general (non-factory) launch reaches the
// existing doors byte-for-byte unchanged.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/spf13/cobra"
)

// withManagedWiringFixture pins the launch seams for the wiring tests: a
// temp project root (cwd), a fake claude binary, and both divert seams —
// the managed owner records its call, the exec/spawn handoff FAILS the test
// if reached (its callers below override it per test).
func withManagedWiringFixture(t *testing.T) (binPath string, managed *managedLaunchCapture) {
	t.Helper()
	tmpDir := t.TempDir()
	for _, dir := range []string{".moai", ".claude"} {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	origDir, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	binPath = filepath.Join(tmpDir, "fake-claude")
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvClaudeBin, binPath)

	origLaunch, origExec := managedFactoryLaunchFunc, execOrSpawnClaudeFunc
	t.Cleanup(func() { managedFactoryLaunchFunc, execOrSpawnClaudeFunc = origLaunch, origExec })
	managed = &managedLaunchCapture{}
	managedFactoryLaunchFunc = func(glmBackend bool, bin string, args, env []string) error {
		managed.record(glmBackend, bin, args, env)
		return nil
	}
	execOrSpawnClaudeFunc = func(claudeBin string, args, env []string) error {
		t.Errorf("exec/spawn handoff reached on the managed wiring path: %s %v", claudeBin, args)
		return nil
	}
	return binPath, managed
}

// managedLaunchCapture records one managed-owner divert.
type managedLaunchCapture struct {
	calls      int
	glmBackend bool
	bin        string
	args       []string
	env        []string
}

func (c *managedLaunchCapture) record(glmBackend bool, bin string, args, env []string) {
	c.calls++
	c.glmBackend, c.bin, c.args, c.env = glmBackend, bin, args, env
}

func (c *managedLaunchCapture) envValue(key string) string {
	return launchEnvValue(c.env, key)
}

// factoryLeaderEnv stamps the leader shape of a factory session.
func factoryLeaderEnv(t *testing.T) {
	t.Helper()
	t.Setenv(config.EnvMoaiKanbanID, "run-wire0001")
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
}

// factoryLaneEnv stamps the lane shape of a factory session.
func factoryLaneEnv(t *testing.T) {
	t.Helper()
	t.Setenv(config.EnvMoaiKanbanID, "run-wire0001")
	t.Setenv(config.EnvMoaiFactoryWorkers, "8")
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-2")
}

// TestManagedLaunchDivertsFactoryLeaderLaunch — a factory leader launch (the
// env carries MOAI_FACTORY_WORKERS) enters the managed owner instead of the
// exec handoff; the operator's `--`-tailed claude arguments flow verbatim
// (the marker itself consumed, R2's tail discipline) and the run id reaches
// the owner's env.
func TestManagedLaunchDivertsFactoryLeaderLaunch(t *testing.T) {
	bin, managed := withManagedWiringFixture(t)
	factoryLeaderEnv(t)
	managedOptIn(t)

	if err := runLaunchClaude("", []string{"--model", "m1", "--", "--progress"}); err != nil {
		t.Fatalf("factory leader launch: %v", err)
	}
	if managed.calls != 1 {
		t.Fatalf("managed owner calls = %d, want 1", managed.calls)
	}
	if managed.glmBackend {
		t.Error("claude launch diverted with the GLM backend label")
	}
	if managed.bin != bin {
		t.Errorf("bin = %q, want %q", managed.bin, bin)
	}
	if managed.envValue(config.EnvMoaiKanbanID) != "run-wire0001" {
		t.Errorf("run id missing from the managed env: %v", managed.env)
	}
	// The `--` tail is claude's own: the marker is consumed by the MoAI parse
	// exactly as on the exec path, and the tail tokens reach the owner intact.
	if !containsToken(managed.args, "--progress") || containsToken(managed.args, "--") {
		t.Errorf("tail discipline broken: args = %v", managed.args)
	}
	if !containsToken(managed.args, "--model") {
		t.Errorf("launcher-interpreted args missing: args = %v", managed.args)
	}
}

// TestManagedLaunchDivertsFactoryLaneLaunch — a factory lane launch (the env
// carries MOAI_FACTORY_WORKER) enters the managed owner the same way.
func TestManagedLaunchDivertsFactoryLaneLaunch(t *testing.T) {
	_, managed := withManagedWiringFixture(t)
	factoryLaneEnv(t)
	managedOptIn(t)

	if err := runLaunchClaude("", nil); err != nil {
		t.Fatalf("factory lane launch: %v", err)
	}
	if managed.calls != 1 {
		t.Fatalf("managed owner calls = %d, want 1", managed.calls)
	}
	if managed.envValue(config.EnvMoaiFactoryWorker) != "lane-2" {
		t.Errorf("lane label missing from the managed env: %v", managed.env)
	}
}

// TestManagedLaunchDivertsFactoryGLMLaunch — a GLM-backend factory launch
// reaches the managed owner under the GLM backend label.
func TestManagedLaunchDivertsFactoryGLMLaunch(t *testing.T) {
	_, managed := withManagedWiringFixture(t)
	factoryLeaderEnv(t)
	managedOptIn(t)
	if err := persistTeamMode(".", "glm"); err != nil {
		t.Fatalf("persist glm team mode: %v", err)
	}

	if err := runLaunchClaude("", nil); err != nil {
		t.Fatalf("factory GLM launch: %v", err)
	}
	if managed.calls != 1 {
		t.Fatalf("managed owner calls = %d, want 1", managed.calls)
	}
	if !managed.glmBackend {
		t.Error("GLM launch diverted without the GLM backend label")
	}
}

// TestLaunchWithoutFactoryEnvReachesExecDoor — the general (non-factory)
// launch path is unchanged: the exec/spawn handoff runs and the managed
// owner is never consulted.
func TestLaunchWithoutFactoryEnvReachesExecDoor(t *testing.T) {
	bin, managed := withManagedWiringFixture(t)

	var execArgs, execEnv []string
	var execBins []string
	origExec := execOrSpawnClaudeFunc
	execOrSpawnClaudeFunc = func(claudeBin string, args, env []string) error {
		execBins = append(execBins, claudeBin)
		execArgs, execEnv = args, env
		return nil
	}
	t.Cleanup(func() { execOrSpawnClaudeFunc = origExec })

	if err := runLaunchClaude("", []string{"--model", "m1", "--", "--progress"}); err != nil {
		t.Fatalf("general launch: %v", err)
	}
	if managed.calls != 0 {
		t.Fatalf("managed owner consulted on the general path: %d calls", managed.calls)
	}
	if len(execBins) != 1 || execBins[0] != bin {
		t.Fatalf("exec handoff bins = %v, want [%s]", execBins, bin)
	}
	if containsToken(execArgs, "--") || !containsToken(execArgs, "--progress") {
		t.Errorf("general-path tail discipline broken: args = %v", execArgs)
	}
	if launchEnvValue(execEnv, config.EnvMoaiKanbanID) != "" {
		t.Errorf("general launch carried a factory run id: %v", execEnv)
	}
}

// TestManagedLaunchRefusesContinue — `--continue` is a plain-launch resume;
// the managed path refuses it instead of silently dropping the flag.
func TestManagedLaunchRefusesContinue(t *testing.T) {
	_, managed := withManagedWiringFixture(t)
	factoryLaneEnv(t)
	managedOptIn(t)

	err := runLaunchClaude("", []string{"-c"})
	if err == nil {
		t.Fatal("managed --continue launch succeeded, want refusal")
	}
	if !strings.Contains(err.Error(), "managed") {
		t.Errorf("refusal does not name the managed session: %v", err)
	}
	if managed.calls != 0 {
		t.Fatalf("managed owner consulted on the refused path: %d calls", managed.calls)
	}
}

// TestManagedLaunchDivertsOnEitherStamp — the gate reads the factory stamps
// through the launchEnv the exec door would have carried, so a stamp present
// ONLY in the launch env (not the process env) still diverts.
func TestManagedLaunchDivertsOnEitherStamp(t *testing.T) {
	_, managed := withManagedWiringFixture(t)
	factoryLeaderEnv(t)
	managedOptIn(t)

	if err := runLaunchClaude("", nil); err != nil {
		t.Fatalf("factory launch: %v", err)
	}
	if managed.calls != 1 {
		t.Fatalf("managed owner calls = %d, want 1", managed.calls)
	}
}

// withManagedCodexWiring pins the codex launch seams for the wiring tests:
// the codex binary lookup returns a sentinel, the init-offer gate stays
// open, and both divert seams are stubbed (the managed owner records; the
// direct door FAILS the test if reached).
func withManagedCodexWiring(t *testing.T) (root string, managed *managedCodexLaunchCapture) {
	t.Helper()
	root = t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	withCodexGateOpen(t)
	withCodexProjectRoot(t, root)

	origLook, origDirect := codexLookPath, codexDirectLaunchFn
	codexLookPath = func(string) (string, error) { return sentinelCodexBinaryPath, nil }
	codexDirectLaunchFn = func(cmd *exec.Cmd) error {
		t.Errorf("direct door reached on the managed codex wiring path: %s %v", cmd.Path, cmd.Args)
		return nil
	}
	origManaged := managedFactoryCodexLaunchFunc
	managed = &managedCodexLaunchCapture{}
	managedFactoryCodexLaunchFunc = func(bin string, args, env []string, dir string) error {
		managed.record(bin, args, env)
		managed.dir = dir
		return nil
	}
	t.Cleanup(func() {
		codexLookPath, codexDirectLaunchFn = origLook, origDirect
		managedFactoryCodexLaunchFunc = origManaged
	})
	return root, managed
}

// managedCodexLaunchCapture records one managed codex divert.
type managedCodexLaunchCapture struct {
	calls int
	bin   string
	args  []string
	env   []string
	dir   string
}

func (c *managedCodexLaunchCapture) record(bin string, args, env []string) {
	c.calls++
	c.bin, c.args, c.env = bin, args, env
}

func (c *managedCodexLaunchCapture) envValue(key string) string {
	return launchEnvValue(c.env, key)
}

// TestManagedCodexLaunchDivertsFactorySession — a `codex cli` launch inside
// a factory-stamped process (leader or lane) enters the managed Codex owner
// instead of the exec handoff; the operator's tail args ride along.
func TestManagedCodexLaunchDivertsFactorySession(t *testing.T) {
	_, managed := withManagedCodexWiring(t)
	factoryLaneEnv(t)
	managedOptIn(t)

	c := &cobra.Command{Use: "codex"}
	if err := runCodex(c, []string{"cli", "--", "--model", "m1"}); err != nil {
		t.Fatalf("factory codex launch: %v", err)
	}
	if managed.calls != 1 {
		t.Fatalf("managed codex owner calls = %d, want 1", managed.calls)
	}
	if managed.envValue(config.EnvMoaiKanbanID) != "run-wire0001" {
		t.Errorf("run id missing from the managed env: %v", managed.env)
	}
	if !containsToken(managed.args, "--model") {
		t.Errorf("tail args missing from the managed argv: %v", managed.args)
	}
	if len(managed.args) == 0 || managed.args[0] != managed.bin {
		t.Errorf("argv[0] convention broken: bin = %q args = %v", managed.bin, managed.args)
	}
}

// TestManagedCodexLaunchCarriesLaunchDir — the divert hands the owner the
// directory the ordinary door would launch in (F1 of the t1375 sync audit):
// the project root even when the process cwd is a subdirectory, and the -w
// worktree (the same directory the anchor lock names).
func TestManagedCodexLaunchCarriesLaunchDir(t *testing.T) {
	t.Run("subdirectory cwd resolves to the project root", func(t *testing.T) {
		root, managed := withManagedCodexWiring(t)
		factoryLaneEnv(t)
		managedOptIn(t)
		sub := filepath.Join(root, "pkg", "deep")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		origDir, _ := os.Getwd()
		t.Cleanup(func() { _ = os.Chdir(origDir) })
		if err := os.Chdir(sub); err != nil {
			t.Fatal(err)
		}
		if err := runCodex(&cobra.Command{Use: "codex"}, []string{"cli"}); err != nil {
			t.Fatalf("factory codex launch: %v", err)
		}
		if managed.calls != 1 || managed.dir != root {
			t.Fatalf("managed owner got dir %q (calls %d), want the project root %q", managed.dir, managed.calls, root)
		}
	})
	t.Run("-w worktree is the owner's directory and the anchor's", func(t *testing.T) {
		root, managed := withManagedCodexWiring(t)
		factoryLaneEnv(t)
		managedOptIn(t)
		tree := filepath.Join(root, ".moai", "worktrees", "card")
		if err := os.MkdirAll(tree, 0o755); err != nil {
			t.Fatal(err)
		}
		prevCheck, prevLock := codexWorktreeWriterCheck, codexWorktreeAnchorLock
		var anchored string
		codexWorktreeWriterCheck = func(string) error { return nil }
		codexWorktreeAnchorLock = func(dir string, _ int, _ string) error { anchored = dir; return nil }
		t.Cleanup(func() { codexWorktreeWriterCheck, codexWorktreeAnchorLock = prevCheck, prevLock })
		if err := runCodex(&cobra.Command{Use: "codex"}, []string{"cli", "-w", "card"}); err != nil {
			t.Fatalf("factory codex worktree launch: %v", err)
		}
		if managed.calls != 1 || managed.dir != tree || anchored != tree {
			t.Fatalf("owner dir %q, anchor dir %q (calls %d), want both %q", managed.dir, anchored, managed.calls, tree)
		}
	})
}

// TestCodexLaunchWithoutFactoryEnvReachesDirectDoor — a general codex launch
// (no factory stamps) reaches the direct door and the managed owner is never
// consulted.
func TestCodexLaunchWithoutFactoryEnvReachesDirectDoor(t *testing.T) {
	_, managed := withManagedCodexWiring(t)

	directCalls := 0
	origDirect := codexDirectLaunchFn
	codexDirectLaunchFn = func(cmd *exec.Cmd) error {
		directCalls++
		return nil
	}
	t.Cleanup(func() { codexDirectLaunchFn = origDirect })

	c := &cobra.Command{Use: "codex"}
	if err := runCodex(c, []string{"cli"}); err != nil {
		t.Fatalf("general codex launch: %v", err)
	}
	if managed.calls != 0 {
		t.Fatalf("managed codex owner consulted on the general path: %d calls", managed.calls)
	}
	if directCalls != 1 {
		t.Fatalf("direct door calls = %d, want 1", directCalls)
	}
}

// TestManagedCodexLaunchKeepsSpawnDoor — the tmux `--spawn` door is a
// general-launch surface and stays untouched by the managed wiring even
// under factory stamps.
func TestManagedCodexLaunchKeepsSpawnDoor(t *testing.T) {
	_, managed := withManagedCodexWiring(t)
	factoryLaneEnv(t)
	managedOptIn(t)

	// The spawn door's prereq gate is pinned open the way the launcher-SPEC
	// spawn tests do — this cell measures the divert, not the prereqs.
	origInTmux, origLook := inTmuxFn, spawnLookPath
	inTmuxFn = func() bool { return true }
	spawnLookPath = func(string) (string, error) { return "/usr/bin/fake", nil }
	t.Cleanup(func() { inTmuxFn, spawnLookPath = origInTmux, origLook })

	spawnCalls := 0
	origSpawn := codexSpawnLaunchFn
	codexSpawnLaunchFn = func(dir, program string, args, factoryEnv []string) error {
		spawnCalls++
		return nil
	}
	t.Cleanup(func() { codexSpawnLaunchFn = origSpawn })

	c := &cobra.Command{Use: "codex"}
	if err := runCodex(c, []string{"cli", "--spawn"}); err != nil {
		t.Fatalf("spawned codex launch: %v", err)
	}
	if managed.calls != 0 {
		t.Fatalf("managed codex owner consulted on the spawn path: %d calls", managed.calls)
	}
	if spawnCalls != 1 {
		t.Fatalf("spawn door calls = %d, want 1", spawnCalls)
	}
}

// containsToken reports whether the exact token appears in args.
func containsToken(args []string, token string) bool {
	for _, a := range args {
		if a == token {
			return true
		}
	}
	return false
}
