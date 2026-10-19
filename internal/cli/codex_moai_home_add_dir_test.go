package cli

import (
	"os"
	"os/exec"
	"slices"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/spf13/cobra"
)

// codexAddDirValue returns the value following --add-dir in argv, or "" when
// the flag is absent.
func codexAddDirValue(argv []string) string {
	i := slices.Index(argv, "--add-dir")
	if i < 0 || i+1 >= len(argv) {
		return ""
	}
	return argv[i+1]
}

// withRealCodexMoaiHomeAddDir restores the real --add-dir derivation that
// withCodexLaunchCapture pins off, under an isolated MOAI_HOME.
func withRealCodexMoaiHomeAddDir(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	prev := codexMoaiHomeAddDirArgs
	codexMoaiHomeAddDirArgs = codexMoaiHomeAddDir
	t.Cleanup(func() { codexMoaiHomeAddDirArgs = prev })
	return home
}

// TestCodexLaunch_CliAddsMoaiHomeWritableDir — the interactive launch hands
// codex `--add-dir <moai home>` so a workspace-write sandbox can open the
// project databases under ~/.moai (SQLite needs to create -wal/-shm/journal
// files beside them). The flag precedes the operator tail, which keeps a
// literal `--` in the tail from turning it into a prompt.
func TestCodexLaunch_CliAddsMoaiHomeWritableDir(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	home := withRealCodexMoaiHomeAddDir(t)
	withCodexProjectRoot(t, t.TempDir())
	tail := []string{"--model", "o3"}
	if _, _, err := runCodexCmd(t, append([]string{"cli", "--"}, tail...)...); err != nil {
		t.Fatalf("cli: %v", err)
	}
	argv := codexOnlyRecord(t, cap).Argv[1:]
	if got := codexAddDirValue(argv); got != home {
		t.Fatalf("argv = %q, want --add-dir %q", argv, home)
	}
	if want := append([]string{"--add-dir", home}, tail...); !slices.Equal(argv, want) {
		t.Errorf("argv = %q, want %q (flag before the operator tail)", argv, want)
	}
}

// TestCodexLaunch_AppCarriesNoAddDir — the desktop app subcommand is not the
// sandboxed CLI session; its argv stays exactly [app].
func TestCodexLaunch_AppCarriesNoAddDir(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withRealCodexMoaiHomeAddDir(t)
	withCodexProjectRoot(t, t.TempDir())
	if _, _, err := runCodexCmd(t, "app"); err != nil {
		t.Fatalf("app: %v", err)
	}
	if argv := codexOnlyRecord(t, cap).Argv[1:]; !slices.Equal(argv, []string{"app"}) {
		t.Errorf("app argv = %q, want exactly [app]", argv)
	}
}

// TestCodexMoaiHomeAddDir_AbsentHomeAddsNothing — a moai home that does not
// exist yet is not handed to codex; the launch proceeds without the flag.
func TestCodexMoaiHomeAddDir_AbsentHomeAddsNothing(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir()+"/missing")
	if got := codexMoaiHomeAddDir(); got != nil {
		t.Errorf("codexMoaiHomeAddDir() = %q, want nil for an absent home", got)
	}
}

// TestCodexLaneLaunch_AddsMoaiHomeWritableDir — `moai codex -l` runs
// `moai todo --auto` inside the codex sandbox, which reads the backlog
// database under ~/.moai; the lane child must carry the same --add-dir.
func TestCodexLaneLaunch_AddsMoaiHomeWritableDir(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdRecordLeaderRun(t, root, fcRun, factory.BackendClaude)
	t.Chdir(root)
	netScrubLaneEnv(t)
	// The lane scrub clears MOAI_HOME, so the expectation is whatever home the
	// scrubbed environment resolves — the same resolution the launcher makes.
	home, err := paths.MoaiHome()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(home, 0o700); err != nil {
		t.Fatal(err)
	}
	prevAdd := codexMoaiHomeAddDirArgs
	codexMoaiHomeAddDirArgs = codexMoaiHomeAddDir
	var argv []string
	prevLook, prevDirect := codexLookPath, codexDirectLaunchFn
	codexLookPath = func(string) (string, error) { return "/sentinel/codex", nil }
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		if argv == nil {
			argv = c.Args
		}
		return nil
	}
	t.Cleanup(func() {
		codexLookPath, codexDirectLaunchFn, codexMoaiHomeAddDirArgs = prevLook, prevDirect, prevAdd
	})
	if _, _, err := runCodexCmd(t, "-l"); err != nil {
		t.Fatalf("codex -l: %v", err)
	}
	if got := codexAddDirValue(argv); got != home {
		t.Errorf("lane argv = %q, want --add-dir %q", argv, home)
	}
	if argv[len(argv)-1] != codexLaneAutoPrompt {
		t.Errorf("lane argv last = %q, want the auto directive", argv[len(argv)-1])
	}
}

// pinCodexMoaiHomeAddDir makes the --add-dir seam return a fixed pair so the
// tests below do not depend on the real MOAI_HOME existing.
func pinCodexMoaiHomeAddDir(t *testing.T, dir string) {
	t.Helper()
	prev := codexMoaiHomeAddDirArgs
	codexMoaiHomeAddDirArgs = func() []string { return []string{"--add-dir", dir} }
	t.Cleanup(func() { codexMoaiHomeAddDirArgs = prev })
}

// TestCodexLaunch_ManagedDivertCarriesNoAddDir — the managed owner's option
// parser is headless and refuses every flag but -c/--config and -m/--model, so
// the launcher-injected moai-home --add-dir must stay off the argv handed to
// it, while the operator tail still rides along (card t1823).
func TestCodexLaunch_ManagedDivertCarriesNoAddDir(t *testing.T) {
	_, managed := withManagedCodexWiring(t)
	factoryLaneEnv(t)
	managedOptIn(t)
	pinCodexMoaiHomeAddDir(t, t.TempDir())

	c := &cobra.Command{Use: "codex"}
	if err := runCodex(c, []string{"cli", "--", "--model", "m1"}); err != nil {
		t.Fatalf("factory codex launch: %v", err)
	}
	if managed.calls != 1 {
		t.Fatalf("managed codex owner calls = %d, want 1", managed.calls)
	}
	if slices.Contains(managed.args, "--add-dir") {
		t.Errorf("managed argv = %q, must not carry --add-dir", managed.args)
	}
	if !slices.Contains(managed.args, "--model") {
		t.Errorf("managed argv = %q, lost the operator tail", managed.args)
	}
}

// TestCodexLaunch_DirectDoorKeepsAddDirWhenNotDiverted — the same launch
// without the managed opt-in reaches the direct door with the --add-dir intact.
func TestCodexLaunch_DirectDoorKeepsAddDirWhenNotDiverted(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	dir := t.TempDir()
	pinCodexMoaiHomeAddDir(t, dir)

	if _, _, err := runCodexCmd(t, "cli", "--", "--model", "m1"); err != nil {
		t.Fatalf("cli: %v", err)
	}
	argv := codexOnlyRecord(t, cap).Argv[1:]
	if got := codexAddDirValue(argv); got != dir {
		t.Errorf("direct argv = %q, want --add-dir %q", argv, dir)
	}
}
