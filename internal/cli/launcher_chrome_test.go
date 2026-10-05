package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// captureLaunchArgs drives launchClaudeDefault through the exec seam and
// returns the argv handed to it. The binary is pinned so no real `claude` on
// PATH is needed, and settingsLocal (when non-empty) is written as the
// project's .claude/settings.local.json before the launch.
func captureLaunchArgs(t *testing.T, settingsLocal string, extraArgs []string) []string {
	t.Helper()
	root := fakeMoaiProject(t)
	if settingsLocal != "" {
		if err := os.MkdirAll(filepath.Join(root, ".claude"), 0o755); err != nil {
			t.Fatalf("create .claude: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, ".claude", "settings.local.json"), []byte(settingsLocal), 0o644); err != nil {
			t.Fatalf("write settings.local.json: %v", err)
		}
	}
	t.Setenv("PATH", t.TempDir())
	t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude-pinned")))

	origExec := execOrSpawnClaudeFunc
	t.Cleanup(func() { execOrSpawnClaudeFunc = origExec })
	var captured []string
	execOrSpawnClaudeFunc = func(_ string, args, _ []string) error {
		captured = args
		return nil
	}

	if err := launchClaudeDefault("", extraArgs); err != nil {
		t.Fatalf("launch: %v", err)
	}
	if len(captured) == 0 {
		t.Fatal("exec seam was never reached")
	}
	return captured
}

func countArg(args []string, want string) int {
	n := 0
	for _, a := range args {
		if a == want {
			n++
		}
	}
	return n
}

// TestLaunchDoesNotInjectNoChromeByDefault pins card t1110: the launcher must
// not add --no-chrome on its own. It used to, whenever DO_CLAUDE_CHROME was
// not "true", which left every cc/glm/factory session unable to attach Chrome
// via /chrome.
func TestLaunchDoesNotInjectNoChromeByDefault(t *testing.T) {
	args := captureLaunchArgs(t, "", nil)
	if n := countArg(args, "--no-chrome"); n != 0 {
		t.Errorf("argv = %v carries --no-chrome %d time(s); the launcher must not disable Chrome by default", args, n)
	}
}

// TestLaunchIgnoresLegacyChromeSetting proves an old DO_CLAUDE_CHROME value
// neither errors nor brings the injection back.
func TestLaunchIgnoresLegacyChromeSetting(t *testing.T) {
	args := captureLaunchArgs(t, `{"env":{"DO_CLAUDE_CHROME":"false"}}`, nil)
	if n := countArg(args, "--no-chrome"); n != 0 {
		t.Errorf("argv = %v carries --no-chrome; a legacy DO_CLAUDE_CHROME value must be ignored", args)
	}
}

// TestLaunchForwardsExplicitChromeFlags proves the user's own choice reaches
// Claude Code unchanged, exactly once.
func TestLaunchForwardsExplicitChromeFlags(t *testing.T) {
	for _, tc := range []struct {
		name  string
		extra []string
		flag  string
	}{
		{"--no-chrome", []string{"--no-chrome"}, "--no-chrome"},
		{"--chrome", []string{"--chrome"}, "--chrome"},
		{"after separator", []string{"--", "--no-chrome"}, "--no-chrome"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := captureLaunchArgs(t, "", tc.extra)
			if n := countArg(args, tc.flag); n != 1 {
				t.Errorf("argv = %v carries %s %d time(s), want exactly 1 (user choice passed through)", args, tc.flag, n)
			}
		})
	}
}
