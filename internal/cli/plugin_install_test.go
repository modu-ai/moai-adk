package cli

// plugin_install_test.go — SPEC-PLUGIN-MARKETPLACE-001 M3a unit criteria
// AC-010 to AC-015 and AC-017. Every test injects a recording runner through
// the pluginRunner seam; none starts a real claude or codex, and every
// home-reading value is pinned to t.TempDir().

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// pluginCall is one recorded runner invocation.
type pluginCall struct {
	bin  string
	args []string
	env  []string
}

// vector renders a call as "<bin base> <args...>" for compact assertions.
func (c pluginCall) vector() string {
	return filepath.Base(c.bin) + " " + strings.Join(c.args, " ")
}

// fakePluginRunner records every call and answers through fn (nil = success).
type fakePluginRunner struct {
	calls []pluginCall
	fn    func(ctx context.Context, call pluginCall, n int) ([]byte, error)
}

func (f *fakePluginRunner) Run(ctx context.Context, bin string, args []string, env []string) ([]byte, error) {
	call := pluginCall{bin: bin, args: append([]string(nil), args...), env: append([]string(nil), env...)}
	f.calls = append(f.calls, call)
	if f.fn != nil {
		return f.fn(ctx, call, len(f.calls))
	}
	return nil, nil
}

func (f *fakePluginRunner) vectors() []string {
	out := make([]string, 0, len(f.calls))
	for _, c := range f.calls {
		out = append(out, c.vector())
	}
	return out
}

// withPluginRunner installs r as the package runner for the test.
func withPluginRunner(t *testing.T, r pluginCommandRunner) {
	t.Helper()
	orig := pluginRunner
	pluginRunner = r
	t.Cleanup(func() { pluginRunner = orig })
}

// pluginToolEnv is the scratch world one plugin test runs in.
type pluginToolEnv struct {
	claudeBin  string // the pinned Claude binary (MOAI_CLAUDE_BIN)
	codexBin   string // what the Codex PATH seam resolves
	claudeHome string
	codexHome  string
}

// setupPluginTools pins Claude through MOAI_CLAUDE_BIN to an inert file,
// resolves Codex through the codexWiringLookPath seam, points both config
// homes at scratch directories and clears the opt-out.
func setupPluginTools(t *testing.T) pluginToolEnv {
	t.Helper()
	e := pluginToolEnv{
		claudeBin:  writeExecutable(t, filepath.Join(t.TempDir(), "claude-pinned")),
		codexBin:   filepath.Join(t.TempDir(), "codex-fake"),
		claudeHome: t.TempDir(),
		codexHome:  t.TempDir(),
	}
	t.Setenv(config.EnvClaudeBin, e.claudeBin)
	t.Setenv(config.EnvSkipPluginInstall, "")
	t.Setenv(config.EnvClaudeConfigDir, e.claudeHome)
	t.Setenv(codexHomeEnvVar, e.codexHome)

	orig := codexWiringLookPath
	codexWiringLookPath = func(string) (string, error) { return e.codexBin, nil }
	t.Cleanup(func() { codexWiringLookPath = orig })
	return e
}

func bothTools() []pluginTool { return []pluginTool{pluginToolClaude, pluginToolCodex} }

// stepOptions builds options with a short bound and a project-less root.
func stepOptions(t *testing.T, tools ...pluginTool) pluginInstallOptions {
	t.Helper()
	opts := newPluginInstallOptions(tools, t.TempDir(), false)
	opts.Timeout = 5 * time.Second
	return opts
}

// countLines counts the output lines containing sub.
func pluginCountLines(out, sub string) int {
	n := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, sub) {
			n++
		}
	}
	return n
}

const guidanceMarker = "Install it yourself"

func TestPluginInstallStep_Sequence(t *testing.T) {
	t.Run("both-present", func(t *testing.T) {
		e := setupPluginTools(t)
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		var out bytes.Buffer
		if err := runPluginInstallStep(&out, stepOptions(t, pluginToolClaude, pluginToolCodex)); err != nil {
			t.Fatalf("step returned %v, want nil", err)
		}
		want := []string{
			filepath.Base(e.claudeBin) + " plugin marketplace add modu-ai/moai-adk",
			filepath.Base(e.claudeBin) + " plugin install moai@moai-adk",
			filepath.Base(e.codexBin) + " plugin marketplace add modu-ai/moai-adk",
			filepath.Base(e.codexBin) + " plugin add moai@moai-adk",
		}
		if got := r.vectors(); strings.Join(got, "|") != strings.Join(want, "|") {
			t.Fatalf("vectors:\n got  %q\n want %q", got, want)
		}
	})
	t.Run("claude-only", func(t *testing.T) {
		e := setupPluginTools(t)
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		_ = runPluginInstallStep(&bytes.Buffer{}, stepOptions(t, pluginToolClaude))
		if len(r.calls) != 2 || r.calls[0].bin != e.claudeBin || r.calls[1].bin != e.claudeBin {
			t.Fatalf("want exactly two Claude calls, got %q", r.vectors())
		}
	})
	t.Run("codex-only", func(t *testing.T) {
		e := setupPluginTools(t)
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		_ = runPluginInstallStep(&bytes.Buffer{}, stepOptions(t, pluginToolCodex))
		if len(r.calls) != 2 || r.calls[0].bin != e.codexBin || r.calls[1].bin != e.codexBin {
			t.Fatalf("want exactly two Codex calls, got %q", r.vectors())
		}
	})
	t.Run("install-skipped-after-add-fails", func(t *testing.T) {
		e := setupPluginTools(t)
		r := &fakePluginRunner{fn: func(_ context.Context, c pluginCall, _ int) ([]byte, error) {
			if c.bin == e.claudeBin && len(c.args) > 1 && c.args[1] == "marketplace" {
				return []byte("boom"), errors.New("exit status 1")
			}
			return nil, nil
		}}
		withPluginRunner(t, r)
		_ = runPluginInstallStep(&bytes.Buffer{}, stepOptions(t, pluginToolClaude, pluginToolCodex))
		for _, v := range r.vectors() {
			if strings.Contains(v, "claude") && strings.Contains(v, "plugin install") {
				t.Fatalf("Claude install ran after its marketplace add failed: %q", r.vectors())
			}
		}
		if len(r.calls) != 3 {
			t.Fatalf("want claude add (failed) + both Codex commands = 3 calls, got %q", r.vectors())
		}
	})
	t.Run("already-present-is-success", func(t *testing.T) {
		setupPluginTools(t)
		r := &fakePluginRunner{fn: func(context.Context, pluginCall, int) ([]byte, error) {
			return []byte("moai-adk is already installed"), nil // exit 0 whatever the message
		}}
		withPluginRunner(t, r)
		var out bytes.Buffer
		_ = runPluginInstallStep(&out, stepOptions(t, pluginToolClaude, pluginToolCodex))
		if len(r.calls) != 4 {
			t.Fatalf("exit 0 with an 'already installed' message must not stop the sequence: %q", r.vectors())
		}
		if strings.Contains(out.String(), guidanceMarker) {
			t.Fatalf("guidance printed for an exit-0 repeat run:\n%s", out.String())
		}
	})
	t.Run("no-scope-argument", func(t *testing.T) {
		setupPluginTools(t)
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		_ = runPluginInstallStep(&bytes.Buffer{}, stepOptions(t, pluginToolClaude, pluginToolCodex))
		for _, c := range r.calls {
			for _, a := range c.args {
				if strings.HasPrefix(a, "--scope") || a == "--yes" || a == "-y" {
					t.Fatalf("unexpected argument %q in %q", a, c.vector())
				}
			}
		}
	})
	t.Run("pinned-binary-honored", func(t *testing.T) {
		e := setupPluginTools(t)
		// A claude on PATH must lose to the pin.
		pathDir, pathClaude := fakePATHDir(t)
		t.Setenv("PATH", pathDir)
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		_ = runPluginInstallStep(&bytes.Buffer{}, stepOptions(t, pluginToolClaude))
		if len(r.calls) == 0 {
			t.Fatal("no Claude call recorded")
		}
		for _, c := range r.calls {
			if c.bin != e.claudeBin || c.bin == pathClaude {
				t.Fatalf("call used %q, want the pinned %q (PATH copy %q must not win)", c.bin, e.claudeBin, pathClaude)
			}
		}
	})
}

func TestPluginInstallStep_Environment(t *testing.T) {
	t.Run("env-unchanged", func(t *testing.T) {
		e := setupPluginTools(t)
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		_ = runPluginInstallStep(&bytes.Buffer{}, stepOptions(t, pluginToolClaude, pluginToolCodex))
		if len(r.calls) == 0 {
			t.Fatal("no call recorded")
		}
		want := append([]string(nil), os.Environ()...)
		sort.Strings(want)
		for _, c := range r.calls {
			got := append([]string(nil), c.env...)
			sort.Strings(got)
			if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
				t.Fatalf("child environment differs from the parent's for %q (len got %d, want %d)", c.vector(), len(got), len(want))
			}
			if !containsEnv(c.env, config.EnvClaudeConfigDir+"="+e.claudeHome) || !containsEnv(c.env, codexHomeEnvVar+"="+e.codexHome) {
				t.Fatalf("child lost a config-home variable: %q", c.vector())
			}
		}
	})
	t.Run("prints-config-home-claude", func(t *testing.T) {
		e := setupPluginTools(t)
		withPluginRunner(t, &fakePluginRunner{})
		var out bytes.Buffer
		_ = runPluginInstallStep(&out, stepOptions(t, pluginToolClaude))
		if !strings.Contains(out.String(), "Claude Code config home: "+e.claudeHome) {
			t.Fatalf("Claude config home not printed:\n%s", out.String())
		}
	})
	t.Run("prints-config-home-codex", func(t *testing.T) {
		e := setupPluginTools(t)
		withPluginRunner(t, &fakePluginRunner{})
		var out bytes.Buffer
		_ = runPluginInstallStep(&out, stepOptions(t, pluginToolCodex))
		if !strings.Contains(out.String(), "Codex config home: "+e.codexHome) {
			t.Fatalf("Codex config home not printed:\n%s", out.String())
		}
	})
	t.Run("default-home-when-unset", func(t *testing.T) {
		setupPluginTools(t)
		_ = os.Unsetenv(config.EnvClaudeConfigDir)
		_ = os.Unsetenv(codexHomeEnvVar)
		withPluginRunner(t, &fakePluginRunner{})
		var out bytes.Buffer
		_ = runPluginInstallStep(&out, stepOptions(t, pluginToolClaude, pluginToolCodex))
		for _, want := range []string{"Claude Code config home: ~/.claude", "Codex config home: ~/.codex"} {
			if !strings.Contains(out.String(), want) {
				t.Fatalf("missing %q in:\n%s", want, out.String())
			}
		}
	})
}

func containsEnv(env []string, kv string) bool {
	for _, e := range env {
		if e == kv {
			return true
		}
	}
	return false
}

func TestPluginInstallStep_FailOpen(t *testing.T) {
	failFirst := func(context.Context, pluginCall, int) ([]byte, error) {
		return []byte("tool said no"), errors.New("exit status 1")
	}
	t.Run("exit-nonzero", func(t *testing.T) {
		setupPluginTools(t)
		withPluginRunner(t, &fakePluginRunner{fn: failFirst})
		var out bytes.Buffer
		if err := runPluginInstallStep(&out, stepOptions(t, pluginToolClaude)); err != nil {
			t.Fatalf("step returned %v, want nil", err)
		}
		if got := strings.Count(out.String(), guidanceMarker); got != 1 {
			t.Fatalf("want exactly one guidance block, got %d:\n%s", got, out.String())
		}
	})
	t.Run("timeout", func(t *testing.T) {
		setupPluginTools(t)
		blocked := &fakePluginRunner{fn: func(ctx context.Context, _ pluginCall, _ int) ([]byte, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		}}
		withPluginRunner(t, blocked)
		opts := stepOptions(t, pluginToolClaude)
		opts.Timeout = 30 * time.Millisecond
		var out bytes.Buffer
		start := time.Now()
		if err := runPluginInstallStep(&out, opts); err != nil {
			t.Fatalf("step returned %v, want nil", err)
		}
		if elapsed := time.Since(start); elapsed > 3*time.Second {
			t.Fatalf("step waited %v for a 30ms bound", elapsed)
		}
		if len(blocked.calls) != 1 {
			t.Fatalf("a timed-out add must not be followed by install: %q", blocked.vectors())
		}
		if !strings.Contains(out.String(), guidanceMarker) {
			t.Fatalf("no guidance after a timeout:\n%s", out.String())
		}
	})
	t.Run("bound-equals-constant", func(t *testing.T) {
		opts := newPluginInstallOptions(bothTools(), "", false)
		if opts.Timeout != config.DefaultPluginInstallCommandTimeout {
			t.Fatalf("production wiring bound = %v, want config.DefaultPluginInstallCommandTimeout (%v)",
				opts.Timeout, config.DefaultPluginInstallCommandTimeout)
		}
		if config.DefaultPluginInstallCommandTimeout != 60*time.Second {
			t.Fatalf("DefaultPluginInstallCommandTimeout = %v, want 60s", config.DefaultPluginInstallCommandTimeout)
		}
	})
	t.Run("invalid-pin", func(t *testing.T) {
		setupPluginTools(t)
		t.Setenv(config.EnvClaudeBin, filepath.Join(t.TempDir(), "no-such-claude"))
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		var out bytes.Buffer
		if err := runPluginInstallStep(&out, stepOptions(t, pluginToolClaude)); err != nil {
			t.Fatalf("step returned %v, want nil", err)
		}
		if len(r.calls) != 0 {
			t.Fatalf("a command ran for an invalid pin: %q", r.vectors())
		}
		if got := strings.Count(out.String(), guidanceMarker); got != 1 {
			t.Fatalf("want one guidance block for an invalid pin, got %d:\n%s", got, out.String())
		}
		if strings.Contains(out.String(), "not found on PATH") {
			t.Fatalf("an invalid pin must print guidance, not the absent-tool skip line:\n%s", out.String())
		}
	})
	t.Run("add-fails-no-install", func(t *testing.T) {
		setupPluginTools(t)
		r := &fakePluginRunner{fn: failFirst}
		withPluginRunner(t, r)
		_ = runPluginInstallStep(&bytes.Buffer{}, stepOptions(t, pluginToolClaude, pluginToolCodex))
		// Each tool's add fails; neither install may follow.
		if len(r.calls) != 2 {
			t.Fatalf("want one failed add per tool and no install, got %q", r.vectors())
		}
	})
	t.Run("guidance-names-both-commands", func(t *testing.T) {
		setupPluginTools(t)
		withPluginRunner(t, &fakePluginRunner{fn: failFirst})
		var out bytes.Buffer
		_ = runPluginInstallStep(&out, stepOptions(t, pluginToolClaude, pluginToolCodex))
		for _, want := range []string{
			"claude plugin marketplace add modu-ai/moai-adk",
			"claude plugin install moai@moai-adk",
			"codex plugin marketplace add modu-ai/moai-adk",
			"codex plugin add moai@moai-adk",
		} {
			if !strings.Contains(out.String(), want) {
				t.Fatalf("guidance missing the manual command %q:\n%s", want, out.String())
			}
		}
		if got := strings.Count(out.String(), guidanceMarker); got != 2 {
			t.Fatalf("want one block per failed tool (2), got %d", got)
		}
	})
	t.Run("returns-nil", func(t *testing.T) {
		setupPluginTools(t)
		withPluginRunner(t, &fakePluginRunner{fn: failFirst})
		if err := runPluginInstallStep(&bytes.Buffer{}, stepOptions(t, pluginToolClaude, pluginToolCodex)); err != nil {
			t.Fatalf("step returned %v after every command failed, want nil", err)
		}
	})
}

func TestPluginInstallStep_ToolAbsent(t *testing.T) {
	skipLine := "not found on PATH"
	t.Run("claude-absent", func(t *testing.T) {
		setupPluginTools(t)
		t.Setenv(config.EnvClaudeBin, "")
		t.Setenv("PATH", t.TempDir())
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		var out bytes.Buffer
		if err := runPluginInstallStep(&out, stepOptions(t, pluginToolClaude)); err != nil {
			t.Fatalf("step returned %v, want nil", err)
		}
		if got := pluginCountLines(out.String(), skipLine); got != 1 {
			t.Fatalf("want exactly one skip line, got %d:\n%s", got, out.String())
		}
		if strings.Contains(out.String(), guidanceMarker) || len(r.calls) != 0 {
			t.Fatalf("absent tool printed guidance or ran a command:\n%s\n%q", out.String(), r.vectors())
		}
	})
	t.Run("codex-absent", func(t *testing.T) {
		setupPluginTools(t)
		codexWiringLookPath = func(string) (string, error) { return "", os.ErrNotExist }
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		var out bytes.Buffer
		if err := runPluginInstallStep(&out, stepOptions(t, pluginToolCodex)); err != nil {
			t.Fatalf("step returned %v, want nil", err)
		}
		if got := pluginCountLines(out.String(), skipLine); got != 1 {
			t.Fatalf("want exactly one skip line, got %d:\n%s", got, out.String())
		}
		if strings.Contains(out.String(), guidanceMarker) || len(r.calls) != 0 {
			t.Fatalf("absent tool printed guidance or ran a command:\n%s\n%q", out.String(), r.vectors())
		}
	})
	t.Run("both-absent", func(t *testing.T) {
		setupPluginTools(t)
		t.Setenv(config.EnvClaudeBin, "")
		t.Setenv("PATH", t.TempDir())
		codexWiringLookPath = func(string) (string, error) { return "", os.ErrNotExist }
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		var out bytes.Buffer
		if err := runPluginInstallStep(&out, stepOptions(t, pluginToolClaude, pluginToolCodex)); err != nil {
			t.Fatalf("step returned %v, want nil", err)
		}
		if got := pluginCountLines(out.String(), skipLine); got != 2 {
			t.Fatalf("want one skip line per missing tool (2), got %d:\n%s", got, out.String())
		}
		if strings.Contains(out.String(), guidanceMarker) || len(r.calls) != 0 {
			t.Fatalf("absent tools printed guidance or ran a command:\n%s\n%q", out.String(), r.vectors())
		}
	})
}

func TestPluginInstallStep_OptOut(t *testing.T) {
	run := func(t *testing.T, noPlugin bool, env string) *fakePluginRunner {
		t.Helper()
		setupPluginTools(t)
		t.Setenv(config.EnvSkipPluginInstall, env)
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		opts := stepOptions(t, pluginToolClaude, pluginToolCodex)
		opts.NoPlugin = noPlugin
		var out bytes.Buffer
		if err := runPluginInstallStep(&out, opts); err != nil {
			t.Fatalf("step returned %v, want nil", err)
		}
		return r
	}
	for _, tc := range []struct {
		name     string
		noPlugin bool
		env      string
		wantRuns bool
	}{
		{"flag", true, "", false},
		{"env-1", false, "1", false},
		{"env-true", false, "true", false},
		{"env-empty-is-not-optout", false, "", true},
		{"env-0-is-not-optout", false, "0", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := run(t, tc.noPlugin, tc.env)
			if tc.wantRuns && len(r.calls) == 0 {
				t.Fatal("step opted out although it must not")
			}
			if !tc.wantRuns && len(r.calls) != 0 {
				t.Fatalf("opt-out still ran commands: %q", r.vectors())
			}
		})
	}
}

// recordingScript writes an executable that appends its arguments to a log and
// returns (script, log). It is POSIX-only; callers skip on Windows.
func recordingScript(t *testing.T, dir, name, log string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	body := "#!/bin/sh\necho \"" + name + " $*\" >> '" + log + "'\nexit 0\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write recording script: %v", err)
	}
	return path
}

func readLog(t *testing.T, log string) string {
	t.Helper()
	b, err := os.ReadFile(log)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("read %s: %v", log, err)
	}
	return string(b)
}

func TestPluginInstallStep_NoRealRunnerUnderTest(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("recording shell scripts are POSIX-only")
	}
	t.Run("default-runner-refuses", func(t *testing.T) {
		log := filepath.Join(t.TempDir(), "record.log")
		script := recordingScript(t, t.TempDir(), "claude", log)
		_, err := execPluginRunner{}.Run(context.Background(), script, []string{"plugin", "install", "moai@moai-adk"}, os.Environ())
		if !errors.Is(err, errPluginRunnerRefused) {
			t.Fatalf("default runner returned %v, want errPluginRunnerRefused", err)
		}
		if got := readLog(t, log); got != "" {
			t.Fatalf("default runner started a process under go test: %q", got)
		}
	})
	t.Run("step-silent-under-default-runner", func(t *testing.T) {
		// Under the default runner the step prints nothing, so the many tests
		// that call runInit see no new stderr output.
		setupPluginTools(t)
		withPluginRunner(t, execPluginRunner{})
		var out bytes.Buffer
		if err := runPluginInstallStep(&out, stepOptions(t, pluginToolClaude, pluginToolCodex)); err != nil {
			t.Fatalf("step returned %v, want nil", err)
		}
		if out.Len() != 0 {
			t.Fatalf("step printed under the default runner in a test binary:\n%s", out.String())
		}
	})
	t.Run("step-is-reached-positive-control", func(t *testing.T) {
		// The empty-record assertion below is only meaningful if runInit
		// actually reaches the step: with an injected runner it must record.
		setupPluginTools(t)
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		runInitForAutonomy(t, nil, map[string]string{"llm": "both"})
		if len(r.calls) == 0 {
			t.Fatal("runInit never reached the plugin step with an injected runner")
		}
	})
	t.Run("pin-and-path-shims-untouched", func(t *testing.T) {
		log := filepath.Join(t.TempDir(), "record.log")
		shimDir := t.TempDir()
		recordingScript(t, shimDir, "claude", log)
		recordingScript(t, shimDir, "codex", log)
		pin := recordingScript(t, t.TempDir(), "claude-pinned", log)
		t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))
		t.Setenv(config.EnvClaudeBin, pin)
		t.Setenv(config.EnvSkipPluginInstall, "")
		t.Setenv(config.EnvClaudeConfigDir, t.TempDir())
		t.Setenv(codexHomeEnvVar, t.TempDir())
		withPluginRunner(t, execPluginRunner{}) // the default, explicitly
		runInitForAutonomy(t, nil, map[string]string{"llm": "both"})
		if got := readLog(t, log); got != "" {
			t.Fatalf("runInit started a recorded process through the default runner: %q", got)
		}
	})
	t.Run("rune-callers-covered", func(t *testing.T) {
		log := filepath.Join(t.TempDir(), "record.log")
		shimDir := t.TempDir()
		recordingScript(t, shimDir, "claude", log)
		recordingScript(t, shimDir, "codex", log)
		pin := recordingScript(t, t.TempDir(), "claude-pinned", log)
		t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))
		t.Setenv(config.EnvClaudeBin, pin)
		t.Setenv(config.EnvSkipPluginInstall, "")
		t.Setenv(config.EnvClaudeConfigDir, t.TempDir())
		t.Setenv(codexHomeEnvVar, t.TempDir())
		t.Setenv("HOME", t.TempDir())
		withPluginRunner(t, execPluginRunner{})

		root := t.TempDir()
		initCmd.SetOut(&bytes.Buffer{})
		initCmd.SetErr(&bytes.Buffer{})
		for k, v := range map[string]string{"root": root, "non-interactive": "true", "name": "rune-proj", "language": "Go", "mode": "ddd"} {
			if err := initCmd.Flags().Set(k, v); err != nil {
				t.Fatalf("set --%s: %v", k, err)
			}
		}
		if err := initCmd.RunE(initCmd, []string{}); err != nil {
			t.Fatalf("initCmd.RunE: %v", err)
		}
		if got := readLog(t, log); got != "" {
			t.Fatalf("initCmd.RunE started a recorded process: %q", got)
		}
	})
}

func TestPluginTestProgramDetector(t *testing.T) {
	for _, tc := range []struct {
		argv0 string
		want  bool
	}{
		{"/tmp/go-build123/b001/cli.test", true},
		{`C:\Temp\cli.test.exe`, true},
		{"/usr/local/bin/moai", false},
		{"moai", false},
		{"/home/u/app.test/bin/moai", false}, // a directory named .test is not the program
	} {
		if got := isPluginTestProgram(tc.argv0); got != tc.want {
			t.Errorf("isPluginTestProgram(%q) = %v, want %v", tc.argv0, got, tc.want)
		}
	}
	if !isPluginTestBinary() {
		t.Error("isPluginTestBinary() = false inside go test")
	}
}

// TestInitPluginStep_AfterDeployment: the step runs after the deployed file
// set is complete, never before (AC-010 a).
func TestInitPluginStep_AfterDeployment(t *testing.T) {
	setupPluginTools(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")
	projectDir := filepath.Join(t.TempDir(), "plugin-proj")

	var earlyCalls int
	deployed := func() bool {
		for _, rel := range []string{"CLAUDE.md", ".claude/settings.json", ".moai/config/sections/quality.yaml", ".moai/manifest.json"} {
			if _, err := os.Stat(filepath.Join(projectDir, rel)); err != nil {
				return false
			}
		}
		return true
	}
	r := &fakePluginRunner{fn: func(context.Context, pluginCall, int) ([]byte, error) {
		if !deployed() {
			earlyCalls++
		}
		return nil, nil
	}}
	withPluginRunner(t, r)

	cmd := newInitTestCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	if err := runInit(cmd, []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v (stderr: %s)", err, errBuf.String())
	}
	if len(r.calls) == 0 {
		t.Fatal("init made no plugin call at all")
	}
	if earlyCalls != 0 {
		t.Fatalf("%d plugin call(s) ran before the deployed file set was complete", earlyCalls)
	}
}

// TestInitPluginStep_OptOut: the opt-out reaches the step through `moai init`
// itself, by flag and by environment (AC-015 a through the init entry point).
func TestInitPluginStep_OptOut(t *testing.T) {
	if initCmd.Flags().Lookup("no-plugin") == nil {
		t.Fatal("moai init has no --no-plugin flag")
	}
	runWith := func(t *testing.T, env string, noPlugin bool) *fakePluginRunner {
		t.Helper()
		setupPluginTools(t)
		t.Setenv(config.EnvSkipPluginInstall, env)
		t.Setenv("HOME", t.TempDir())
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		cmd := newInitTestCmd()
		cmd.Flags().Bool("no-plugin", false, "")
		if noPlugin {
			if err := cmd.Flags().Set("no-plugin", "true"); err != nil {
				t.Fatal(err)
			}
		}
		var out, errBuf bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&errBuf)
		if err := runInit(cmd, []string{filepath.Join(t.TempDir(), "optout-proj")}); err != nil {
			t.Fatalf("runInit: %v (stderr: %s)", err, errBuf.String())
		}
		return r
	}
	t.Run("control-runs-without-optout", func(t *testing.T) {
		if r := runWith(t, "", false); len(r.calls) == 0 {
			t.Fatal("control: init made no plugin call, so the opt-out cases prove nothing")
		}
	})
	t.Run("flag", func(t *testing.T) {
		if r := runWith(t, "", true); len(r.calls) != 0 {
			t.Fatalf("--no-plugin still ran commands: %q", r.vectors())
		}
	})
	t.Run("env", func(t *testing.T) {
		if r := runWith(t, "1", false); len(r.calls) != 0 {
			t.Fatalf("%s=1 still ran commands: %q", config.EnvSkipPluginInstall, r.vectors())
		}
	})
}

func TestInitPluginStep_HarnessGating(t *testing.T) {
	for _, tc := range []struct {
		name       string
		flags      map[string]string
		wantClaude bool
		wantCodex  bool
	}{
		{"claude-default", nil, true, false},
		{"claude-explicit", map[string]string{"llm": "claude"}, true, false},
		{"gpt", map[string]string{"llm": "gpt"}, false, true},
		{"both", map[string]string{"llm": "both"}, true, true},
		{"unrecognized-falls-back-to-claude", map[string]string{"llm": "gemini"}, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := setupPluginTools(t)
			r := &fakePluginRunner{}
			withPluginRunner(t, r)
			runInitForAutonomy(t, nil, tc.flags)
			var sawClaude, sawCodex bool
			firstClaude, firstCodex := -1, -1
			for i, c := range r.calls {
				switch c.bin {
				case e.claudeBin:
					sawClaude = true
					if firstClaude < 0 {
						firstClaude = i
					}
				case e.codexBin:
					sawCodex = true
					if firstCodex < 0 {
						firstCodex = i
					}
				default:
					t.Fatalf("call to an unexpected binary %q", c.bin)
				}
			}
			if sawClaude != tc.wantClaude || sawCodex != tc.wantCodex {
				t.Fatalf("harness %v: claude=%v codex=%v, want claude=%v codex=%v (%q)",
					tc.flags, sawClaude, sawCodex, tc.wantClaude, tc.wantCodex, r.vectors())
			}
			if sawClaude && sawCodex && firstClaude > firstCodex {
				t.Fatalf("Claude must come first: %q", r.vectors())
			}
		})
	}
}
