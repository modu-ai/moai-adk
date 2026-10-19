package cli

import (
	"errors"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
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
