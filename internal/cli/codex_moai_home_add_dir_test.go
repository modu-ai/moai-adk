package cli

import (
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/internal/testguard"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
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

// TestCodexLaunch_CliAddsMoaiHomeWritableDir pins the narrowed per-project
// database/run/search/worktree grants. SQLite's WAL/SHM files stay writable, and the
// grant block still precedes the operator's literal argument tail.
func TestCodexLaunch_CliAddsMoaiHomeWritableDir(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withRealCodexMoaiHomeAddDir(t)
	root := t.TempDir()
	withCodexProjectRoot(t, root)
	tail := []string{"--model", "o3"}
	if _, _, err := runCodexCmd(t, append([]string{"cli", "--"}, tail...)...); err != nil {
		t.Fatalf("cli: %v", err)
	}
	argv := codexOnlyRecord(t, cap).Argv[1:]
	wantGrants := expectedCodexProjectStateGrants(t, root)
	if got := codexAddDirValue(argv); got != wantGrants[1] {
		t.Fatalf("argv = %q, want project database grant %q", argv, wantGrants[1])
	}
	if want := append(wantGrants, tail...); !slices.Equal(argv, want) {
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
	if got := codexMoaiHomeAddDir(t.TempDir()); got != nil {
		t.Errorf("codexMoaiHomeAddDir() = %q, want nil for an absent home", got)
	}
}

// TestCodexLaneLaunch_AddsMoaiHomeWritableDir — `moai codex -l` runs
// `moai todo --auto` inside the codex sandbox, which reads the backlog
// database under ~/.moai; the lane child must carry the same --add-dir.
func TestCodexLaneLaunch_AddsMoaiHomeWritableDir(t *testing.T) {
	testguard.RedirectHome(t)
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdRecordLeaderRun(t, root, fcRun, factory.BackendClaude)
	t.Chdir(root)
	netScrubLaneEnv(t)
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
	wantGrants := expectedCodexProjectStateGrants(t, root)
	if got := codexAddDirValue(argv); got != wantGrants[1] {
		t.Errorf("lane argv = %q, want project database grant %q", argv, wantGrants[1])
	}
	grantAt := slices.Index(argv, "--add-dir")
	for i, want := range wantGrants {
		if grantAt < 0 || grantAt+i >= len(argv) || argv[grantAt+i] != want {
			t.Fatalf("lane argv = %q, want grant block %q", argv, wantGrants)
		}
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
	codexMoaiHomeAddDirArgs = func(string) []string { return []string{"--add-dir", dir} }
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

// Exact positive contract: project database, run, search and worktree namespaces are
// writable; the shared home, credentials and sibling namespaces are excluded.
func expectedCodexProjectStateGrants(t *testing.T, root string) []string {
	t.Helper()
	db, err := homestate.ProjectDir(root)
	if err != nil {
		t.Fatal(err)
	}
	run, err := homestate.RunProjectDir(root)
	if err != nil {
		t.Fatal(err)
	}
	search, err := homestate.SearchDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	var args []string
	for _, dir := range []string{db, run, filepath.Dir(search)} {
		args = append(args, "--add-dir", m2Real(t, dir))
	}
	if key := worktreeroot.RecordedKey(root); key != "" {
		home, err := paths.MoaiHome()
		if err != nil {
			t.Fatal(err)
		}
		args = append(args, "--add-dir", m2Real(t, filepath.Join(home, "worktrees", "projects", key)))
	}
	return args
}
