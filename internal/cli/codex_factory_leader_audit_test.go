package cli

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestCodexFactoryLeaderNativeCWDRefusedBeforeEffects(t *testing.T) {
	for _, form := range []string{"-C", "--cd", "-C=", "--cd=", "-Cattached"} {
		t.Run(form, func(t *testing.T) {
			root, _ := fcFixture(t)
			t.Chdir(root)
			netScrubLaneEnv(t)
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, root)
			args := []string{"-f", "--"}
			switch form {
			case "-C", "--cd":
				args = append(args, form, t.TempDir())
			case "-Cattached":
				args = append(args, "-C"+t.TempDir())
			default:
				args = append(args, form+t.TempDir())
			}
			before := factoryMarkerEnv()
			_, _, err := runCodexCmd(t, args...)
			if err == nil || !strings.Contains(err.Error(), "-w") {
				t.Fatalf("native cwd should refuse with -w guidance: %v", err)
			}
			if cap.count() != 0 || launcherRunCount(t, root) != 0 || !reflect.DeepEqual(before, factoryMarkerEnv()) {
				t.Fatal("native cwd refusal mutated launcher state")
			}
		})
	}
}

func TestCodexFactoryLeaderModelOptionShapes(t *testing.T) {
	for _, tail := range [][]string{{"-mcustom"}, {"-m=custom"}, {"--model=custom"}, {"-m", "custom"}, {"--model", "custom"}} {
		t.Run(strings.Join(tail, "_"), func(t *testing.T) {
			root, _ := fcFixture(t)
			t.Chdir(root)
			netScrubLaneEnv(t)
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, root)
			codexSessionDefaultArgsFn = codexSessionDefaultArgs
			if _, _, err := runCodexCmd(t, append([]string{"-f", "--"}, tail...)...); err != nil {
				t.Fatal(err)
			}
			rec := codexOnlyRecord(t, cap)
			if slices.Contains(rec.Argv, config.CodexModelSol) {
				t.Fatalf("explicit model got default model too: %q", rec.Argv)
			}
		})
	}
}

func TestCodexFactoryLeaderNativeOptionBoundaries(t *testing.T) {
	for _, tail := range [][]string{{"--", "--model"}, {"--", `--config=model_reasoning_effort="low"`}, {"--", "-C/tmp"}, {"--config", "-C/tmp"}, {"--config", "-mcustom"}} {
		t.Run(strings.Join(tail, "_"), func(t *testing.T) {
			root, _ := fcFixture(t)
			t.Chdir(root)
			netScrubLaneEnv(t)
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, root)
			codexSessionDefaultArgsFn = codexSessionDefaultArgs
			if _, _, err := runCodexCmd(t, append([]string{"-f", "--"}, tail...)...); err != nil {
				t.Fatalf("native option value/prompt mistaken for launcher option: %v", err)
			}
			rec := codexOnlyRecord(t, cap)
			if !slices.Contains(rec.Argv, config.CodexModelSol) || !slices.Contains(rec.Argv, `model_reasoning_effort="high"`) {
				t.Fatalf("prompt/config value suppressed defaults: %q", rec.Argv)
			}
		})
	}
}

func TestCodexFactoryLeaderSpawnPrimitiveFailureClearsOwner(t *testing.T) {
	root, _ := fcFixture(t)
	t.Chdir(root)
	netScrubLaneEnv(t)
	withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	oldTmux, oldLook, oldSpawn := inTmuxFn, spawnLookPath, tmuxSpawnFn
	t.Cleanup(func() { inTmuxFn, spawnLookPath, tmuxSpawnFn = oldTmux, oldLook, oldSpawn })
	inTmuxFn = func() bool { return true }
	spawnLookPath = func(s string) (string, error) { return "/sentinel/" + s, nil }
	injected := errors.New("tmux primitive failed")
	tmuxSpawnFn = func(string, string) (string, error) { return "", injected }
	var id string
	codexSpawnLaunchFn = func(dir, program string, args, env []string) error {
		id = launchEnvValue(env, config.EnvFactoryRunID)
		return defaultCodexSpawnLaunch(dir, program, args, env)
	}
	_, _, err := runCodexCmd(t, "-f", "--spawn")
	if !errors.Is(err, injected) {
		t.Fatalf("spawn error=%v", err)
	}
	if pid, start := runOwnerStamp(t, root, id); pid != 0 || start != "" {
		t.Fatalf("failed spawn retained owner=%d %q (launcher=%d)", pid, start, os.Getpid())
	}
	if n := peerCount(t, root, id, factory.RoleLeader); n != 0 {
		t.Fatalf("failed spawn left pending peer count=%d", n)
	}
}

func TestCodexFactoryLeaderHelpNamesEntry(t *testing.T) {
	if !strings.Contains(codexCmd.Long, "-f, --factory") {
		t.Fatal("Codex help omits factory leader entry")
	}
}

func TestCodexFactoryLeaderPreHandoffFailureClearsOwner(t *testing.T) {
	for _, failure := range []string{"direct-instruction-size", "spawn-instruction-size", "anchor-lock-race"} {
		t.Run(failure, func(t *testing.T) {
			root, _ := fcFixture(t)
			t.Chdir(root)
			netScrubLaneEnv(t)
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, root)
			if err := recordFactoryRunStart(root, "existing-run", factory.BackendClaude, "", 0); err != nil {
				t.Fatal(err)
			}
			t.Setenv(config.EnvFactoryRunID, "existing-run")
			before := factoryMarkerEnv()
			args := []string{"-f"}
			wantError := "ceiling"
			if failure == "anchor-lock-race" {
				tree := filepath.Join(root, ".claude", "worktrees", "plain")
				if err := os.MkdirAll(tree, 0o755); err != nil {
					t.Fatal(err)
				}
				args = append(args, "-w", tree)
				wantError = "anchor claimed after preflight"
				codexWorktreeAnchorLock = func(string, int, string) error { return errors.New(wantError) }
			} else {
				if err := os.WriteFile(filepath.Join(root, codexLocalInstructionName), []byte(strings.Repeat("x", config.DefaultCodexInstructionArgBytes)), 0o600); err != nil {
					t.Fatal(err)
				}
				if failure == "spawn-instruction-size" {
					args = append(args, "--spawn")
					oldTmux, oldLook := inTmuxFn, spawnLookPath
					t.Cleanup(func() { inTmuxFn, spawnLookPath = oldTmux, oldLook })
					inTmuxFn = func() bool { return true }
					spawnLookPath = func(s string) (string, error) { return "/sentinel/" + s, nil }
				}
			}
			_, stderr, err := runCodexCmd(t, args...)
			if err == nil || !strings.Contains(err.Error()+stderr, wantError) || cap.count() != 0 {
				t.Fatalf("pre-handoff error=%v stderr=%s launches=%d", err, stderr, cap.count())
			}
			db := fcOpen(t, root)
			var id string
			if err := db.DB.QueryRow("SELECT run_id FROM runs WHERE lead_backend=?", BackendCodex).Scan(&id); err != nil {
				t.Fatalf("fresh run was not recorded before failure: %v", err)
			}
			if pid, start := runOwnerStamp(t, root, id); pid != 0 || start != "" {
				t.Errorf("pre-handoff failure retained owner=%d %q", pid, start)
			}
			if pid, _ := runOwnerStamp(t, root, "existing-run"); pid != os.Getpid() {
				t.Errorf("existing owner changed: %d", pid)
			}
			if !reflect.DeepEqual(before, factoryMarkerEnv()) {
				t.Error("pre-handoff failure did not restore outer identity")
			}
		})
	}
}

func TestCodexFactoryLeaderSuccessfulHandoffKeepsOwner(t *testing.T) {
	root, _ := fcFixture(t)
	t.Chdir(root)
	netScrubLaneEnv(t)
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	if _, _, err := runCodexCmd(t, "-f"); err != nil {
		t.Fatal(err)
	}
	id := launchEnvValue(codexOnlyRecord(t, cap).Env, config.EnvFactoryRunID)
	if pid, start := runOwnerStamp(t, root, id); pid != os.Getpid() || start != homestate.CurrentProcessFingerprint() {
		t.Fatalf("successful handoff lost owner=%d %q", pid, start)
	}
}
