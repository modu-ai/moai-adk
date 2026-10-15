package cli

// codex_lane_shell_env_test.go — card t1819.
//
// A Codex session hands its shell commands the environment its
// shell_environment_policy allows. Under `inherit = "core"` (or "none") the
// lane keys the launcher puts on the Codex process never reach a shell, so
// `moai todo --auto-lane` inside the lane judged "not a lane session". The
// launcher therefore mirrors the lane keys into
// `-c shell_environment_policy.set.<KEY>="<value>"` overrides, which reach the
// shell regardless of the inherit setting.

import (
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

const codexShellEnvSetPrefix = "shell_environment_policy.set."

// codexShellEnvSetValues collects the `-c shell_environment_policy.set.K=V`
// overrides in argv as K -> raw V (the TOML-encoded value text).
func codexShellEnvSetValues(argv []string) map[string]string {
	out := map[string]string{}
	for i := 0; i+1 < len(argv); i++ {
		if argv[i] != "-c" || !strings.HasPrefix(argv[i+1], codexShellEnvSetPrefix) {
			continue
		}
		key, value, _ := strings.Cut(strings.TrimPrefix(argv[i+1], codexShellEnvSetPrefix), "=")
		out[key] = value
	}
	return out
}

func codexEnvLookup(env []string, key string) string {
	for _, entry := range env {
		if k, v, ok := strings.Cut(entry, "="); ok && k == key {
			return v
		}
	}
	return ""
}

// TestCodexLaneLaunch_MirrorsLaneKeysIntoShellEnvPolicy — `moai codex -l`
// carries the lane role marker, the lane label, the backend and the run id as
// shell_environment_policy.set overrides, each value equal to the one on the
// child process env, all before the auto directive (the positional prompt).
func TestCodexLaneLaunch_MirrorsLaneKeysIntoShellEnvPolicy(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdRecordLeaderRun(t, root, fcRun, factory.BackendClaude)
	t.Chdir(root)
	netScrubLaneEnv(t)
	var argv, env []string
	prevLook, prevDirect := codexLookPath, codexDirectLaunchFn
	codexLookPath = func(string) (string, error) { return "/sentinel/codex", nil }
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		if argv == nil {
			argv, env = c.Args, c.Env
		}
		return nil
	}
	t.Cleanup(func() { codexLookPath, codexDirectLaunchFn = prevLook, prevDirect })
	if _, _, err := runCodexCmd(t, "-l"); err != nil {
		t.Fatalf("codex -l: %v", err)
	}
	got := codexShellEnvSetValues(argv)
	for _, key := range []string{config.EnvFactoryRole, config.EnvMoaiFactoryWorker, config.EnvFactoryBackend, config.EnvFactoryRunID} {
		want := `"` + codexEnvLookup(env, key) + `"`
		if want == `""` {
			t.Fatalf("child env lacks %s: %q", key, env)
		}
		if got[key] != want {
			t.Errorf("lane argv override %s = %q, want %q (argv %q)", key, got[key], want, argv)
		}
	}
	if got[config.EnvFactoryRole] != `"`+config.FactoryRoleLane+`"` {
		t.Errorf("role override = %q, want %q", got[config.EnvFactoryRole], `"`+config.FactoryRoleLane+`"`)
	}
	if argv[len(argv)-1] != codexLaneAutoPrompt {
		t.Errorf("lane argv last = %q, want the auto directive", argv[len(argv)-1])
	}
}

// TestCodexLaunch_PlainCarriesNoShellEnvPolicy — a plain `moai codex`
// session is no lane and gains no shell_environment_policy override.
func TestCodexLaunch_PlainCarriesNoShellEnvPolicy(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	if _, _, err := runCodexCmd(t, "cli"); err != nil {
		t.Fatalf("cli: %v", err)
	}
	argv := codexOnlyRecord(t, cap).Argv
	if slices.ContainsFunc(argv, func(a string) bool { return strings.Contains(a, "shell_environment_policy") }) {
		t.Errorf("plain argv = %q, want no shell_environment_policy override", argv)
	}
}

// TestCodexShellEnvSetArgs_EscapesValues — each entry becomes one
// `-c shell_environment_policy.set.KEY=<TOML basic string>` pair; quotes,
// backslashes and control characters are escaped so the value parses as the
// same string, and an entry without "=" is skipped.
func TestCodexShellEnvSetArgs_EscapesValues(t *testing.T) {
	got := codexShellEnvSetArgs([]string{
		"A=lane",
		`B=a"b\c`,
		"C=line1\nline2\t",
		"D=",
		"malformed",
	})
	want := []string{
		"-c", `shell_environment_policy.set.A="lane"`,
		"-c", `shell_environment_policy.set.B="a\"b\\c"`,
		"-c", `shell_environment_policy.set.C="line1\nline2\t"`,
		"-c", `shell_environment_policy.set.D=""`,
	}
	if !slices.Equal(got, want) {
		t.Errorf("codexShellEnvSetArgs = %q, want %q", got, want)
	}
	if got := codexShellEnvSetArgs(nil); len(got) != 0 {
		t.Errorf("codexShellEnvSetArgs(nil) = %q, want empty", got)
	}
}
