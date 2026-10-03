package cli

// doctor_plugin_version_test.go — SPEC-PLUGIN-MARKETPLACE-001 M4, AC-020 to
// AC-023. Every test injects a runner and synthetic homes under t.TempDir();
// none reads a real Claude or Codex home and none starts a real tool.

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/version"
)

const (
	testBinaryVersion = "v3.4.5"
	testPluginVersion = "3.4.5"
)

func pvWriteClaudeRegistry(t *testing.T, home, body string) {
	t.Helper()
	dir := filepath.Join(home, "plugins")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "installed_plugins.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// pvClaudeRegistry renders the registry shape observed at claude 2.1.287 (P-08)
// with one entry per (scope, version) pair.
func pvClaudeRegistry(pairs ...[2]string) string {
	var entries []string
	for _, p := range pairs {
		entries = append(entries, fmt.Sprintf(`{"scope":%q,"installPath":"/x","version":%q}`, p[0], p[1]))
	}
	return `{"version":2,"plugins":{"moai@moai-adk":[` + strings.Join(entries, ",") + `]}}`
}

// pvCodexList renders the `codex plugin list --json` shape of P-34.
func pvCodexList(ver string) []byte {
	return []byte(fmt.Sprintf(`{"installed":[{"pluginId":"moai@moai-adk","name":"moai","marketplaceName":"moai-adk","version":%q,"installed":true,"enabled":true}],"available":[]}`, ver))
}

func pvCodexFound(string) (string, error)  { return "/fake/codex", nil }
func pvCodexAbsent(string) (string, error) { return "", os.ErrNotExist }

func pvListRunner(out []byte, err error) *fakePluginRunner {
	return &fakePluginRunner{fn: func(context.Context, pluginCall, int) ([]byte, error) { return out, err }}
}

func pvSetBinaryVersion(t *testing.T, v string) {
	t.Helper()
	orig := version.Version
	version.Version = v
	t.Cleanup(func() { version.Version = orig })
}

func pvWithCodexLookPath(t *testing.T, fn func(string) (string, error)) {
	t.Helper()
	orig := codexWiringLookPath
	codexWiringLookPath = fn
	t.Cleanup(func() { codexWiringLookPath = orig })
}

func pvWithUserHomeDirFn(t *testing.T, home string) {
	t.Helper()
	orig := userHomeDirFn
	userHomeDirFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDirFn = orig })
}

func pvWithCodexUserHomeDir(t *testing.T, fn func() (string, error)) {
	t.Helper()
	orig := codexUserHomeDir
	codexUserHomeDir = fn
	t.Cleanup(func() { codexUserHomeDir = orig })
}

func pvRequireInconclusive(t *testing.T, c DiagnosticCheck) {
	t.Helper()
	if c.Status != uikit.CheckOK && c.Status != uikit.CheckInfo {
		t.Errorf("status %q (%s), want ok or info", c.Status, c.Message)
	}
}

func TestCheckPluginVersion_ClaudeRead(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		home := t.TempDir()
		pvWriteClaudeRegistry(t, home, pvClaudeRegistry([2]string{"user", testPluginVersion}))
		c := checkPluginVersionAt(home, "", nil, pvCodexAbsent, testBinaryVersion, false)
		if c.Status != uikit.CheckOK || !strings.Contains(c.Message, testPluginVersion) {
			t.Errorf("got %q %q, want ok naming %s", c.Status, c.Message, testPluginVersion)
		}
	})
	t.Run("v-prefix-normalized", func(t *testing.T) {
		home := t.TempDir()
		pvWriteClaudeRegistry(t, home, pvClaudeRegistry([2]string{"user", "v" + testPluginVersion}))
		c := checkPluginVersionAt(home, "", nil, pvCodexAbsent, testBinaryVersion, false)
		if c.Status != uikit.CheckOK {
			t.Errorf("v-prefixed registry version vs v-prefixed binary: %q %q, want ok", c.Status, c.Message)
		}
		c = checkPluginVersionAt(home, "", nil, pvCodexAbsent, testPluginVersion, false)
		if c.Status != uikit.CheckOK {
			t.Errorf("v-prefixed registry version vs bare binary: %q %q, want ok", c.Status, c.Message)
		}
	})
	t.Run("prerelease-equal", func(t *testing.T) {
		home := t.TempDir()
		pvWriteClaudeRegistry(t, home, pvClaudeRegistry([2]string{"user", "3.2.0-rc.26"}))
		c := checkPluginVersionAt(home, "", nil, pvCodexAbsent, "v3.2.0-rc.26", false)
		if c.Status != uikit.CheckOK {
			t.Errorf("%q %q, want ok", c.Status, c.Message)
		}
	})
	t.Run("claude-config-dir-honored", func(t *testing.T) {
		pvSetBinaryVersion(t, testBinaryVersion)
		honored, decoy := t.TempDir(), t.TempDir()
		pvWriteClaudeRegistry(t, honored, pvClaudeRegistry([2]string{"user", testPluginVersion}))
		pvWriteClaudeRegistry(t, filepath.Join(decoy, ".claude"), pvClaudeRegistry([2]string{"user", "0.0.1"}))
		t.Setenv(config.EnvClaudeConfigDir, honored)
		t.Setenv(codexHomeEnvVar, t.TempDir())
		pvWithUserHomeDirFn(t, decoy)
		pvWithCodexLookPath(t, pvCodexAbsent)
		withPluginRunner(t, &fakePluginRunner{})
		c := checkPluginVersion(false)
		if c.Status != uikit.CheckOK {
			t.Errorf("%q %q: a registry under CLAUDE_CONFIG_DIR must win over ~/.claude", c.Status, c.Message)
		}
	})
	t.Run("user-scope-first", func(t *testing.T) {
		home := t.TempDir()
		pvWriteClaudeRegistry(t, home, pvClaudeRegistry([2]string{"project", "9.9.9"}, [2]string{"user", testPluginVersion}))
		c := checkPluginVersionAt(home, "", nil, pvCodexAbsent, testBinaryVersion, false)
		if c.Status != uikit.CheckOK || !strings.Contains(c.Message, "user") {
			t.Errorf("%q %q, want ok naming the user scope", c.Status, c.Message)
		}
	})
	t.Run("no-subprocess", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("shell shims")
		}
		shimDir, record := t.TempDir(), filepath.Join(t.TempDir(), "record")
		pvWriteRecordingShims(t, shimDir, record)
		t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))
		home := t.TempDir()
		pvWriteClaudeRegistry(t, home, pvClaudeRegistry([2]string{"user", testPluginVersion}))
		r := &fakePluginRunner{}
		c := checkPluginVersionAt(home, "", r, pvCodexAbsent, testBinaryVersion, false)
		if c.Status != uikit.CheckOK {
			t.Fatalf("%q %q, want ok", c.Status, c.Message)
		}
		if len(r.calls) != 0 {
			t.Errorf("runner started %d commands, want 0", len(r.calls))
		}
		if b, _ := os.ReadFile(record); len(b) != 0 {
			t.Errorf("a PATH shim ran: %q", b)
		}
	})
}

func TestCheckPluginVersion_CodexRead(t *testing.T) {
	t.Run("registered-version", func(t *testing.T) {
		r := pvListRunner(pvCodexList(testPluginVersion), nil)
		c := checkPluginVersionAt("", t.TempDir(), r, pvCodexFound, testBinaryVersion, false)
		if c.Status != uikit.CheckOK || !strings.Contains(c.Message, testPluginVersion) {
			t.Errorf("%q %q, want ok naming %s", c.Status, c.Message, testPluginVersion)
		}
		if len(r.calls) != 1 || strings.Join(r.calls[0].args, " ") != "plugin list --json" {
			t.Errorf("calls %q, want exactly one `plugin list --json`", r.vectors())
		}
	})
	t.Run("list-empty-is-not-installed", func(t *testing.T) {
		r := pvListRunner([]byte(`{"installed":[],"available":[]}`), nil)
		c := checkPluginVersionAt("", t.TempDir(), r, pvCodexFound, testBinaryVersion, false)
		pvRequireInconclusive(t, c)
		if strings.Contains(c.Message, testPluginVersion) {
			t.Errorf("empty list reported a version: %q", c.Message)
		}
	})
	t.Run("cache-without-registration", func(t *testing.T) {
		home := t.TempDir()
		if err := os.MkdirAll(filepath.Join(home, "plugins", "cache", "moai-adk", "moai", testPluginVersion), 0o755); err != nil {
			t.Fatal(err)
		}
		r := pvListRunner([]byte(`{"installed":[],"available":[]}`), nil)
		c := checkPluginVersionAt("", home, r, pvCodexFound, testBinaryVersion, false)
		pvRequireInconclusive(t, c)
		if strings.Contains(c.Message, testPluginVersion) {
			t.Errorf("a cache directory without a registration was reported as installed: %q", c.Message)
		}
	})
	t.Run("multiple-cache-versions", func(t *testing.T) {
		home := t.TempDir()
		for _, v := range []string{testPluginVersion, "9.9.9"} {
			if err := os.MkdirAll(filepath.Join(home, "plugins", "cache", "moai-adk", "moai", v), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		r := pvListRunner(pvCodexList(testPluginVersion), nil)
		c := checkPluginVersionAt("", home, r, pvCodexFound, testBinaryVersion, false)
		if c.Status != uikit.CheckOK || strings.Contains(c.Message, "9.9.9") {
			t.Errorf("%q %q, want ok on the listed version only", c.Status, c.Message)
		}
	})
	t.Run("codex-absent-no-spawn", func(t *testing.T) {
		r := &fakePluginRunner{}
		c := checkPluginVersionAt("", t.TempDir(), r, pvCodexAbsent, testBinaryVersion, false)
		pvRequireInconclusive(t, c)
		if len(r.calls) != 0 {
			t.Errorf("runner started %d commands with codex absent, want 0", len(r.calls))
		}
	})
	t.Run("probe-timeout-is-info", func(t *testing.T) {
		r := pvListRunner(nil, context.DeadlineExceeded)
		c := checkPluginVersionAt("", t.TempDir(), r, pvCodexFound, testBinaryVersion, false)
		pvRequireInconclusive(t, c)
	})
	t.Run("probe-malformed-json-is-info", func(t *testing.T) {
		for _, body := range []string{"{not json", `{"installed":"x"}`, `[1,2]`, ""} {
			r := pvListRunner([]byte(body), nil)
			c := checkPluginVersionAt("", t.TempDir(), r, pvCodexFound, testBinaryVersion, false)
			pvRequireInconclusive(t, c)
		}
	})
	t.Run("probe-bound-equals-constant", func(t *testing.T) {
		var remaining time.Duration
		var hasDeadline bool
		r := &fakePluginRunner{fn: func(ctx context.Context, _ pluginCall, _ int) ([]byte, error) {
			var dl time.Time
			dl, hasDeadline = ctx.Deadline()
			remaining = time.Until(dl)
			return pvCodexList(testPluginVersion), nil
		}}
		checkPluginVersionAt("", t.TempDir(), r, pvCodexFound, testBinaryVersion, false)
		if !hasDeadline {
			t.Fatal("the probe ran with no deadline")
		}
		if remaining > config.DefaultPluginVersionProbeTimeout || remaining < config.DefaultPluginVersionProbeTimeout-time.Second {
			t.Errorf("deadline in %s, want about %s", remaining, config.DefaultPluginVersionProbeTimeout)
		}
		if config.DefaultPluginVersionProbeTimeout != 3*time.Second {
			t.Errorf("DefaultPluginVersionProbeTimeout = %s, want 3s", config.DefaultPluginVersionProbeTimeout)
		}
	})
}

func TestCheckPluginVersion_Outcomes(t *testing.T) {
	t.Run("mismatch-warn", func(t *testing.T) {
		home := t.TempDir()
		pvWriteClaudeRegistry(t, home, pvClaudeRegistry([2]string{"user", "3.4.4"}))
		c := checkPluginVersionAt(home, "", nil, pvCodexAbsent, testBinaryVersion, false)
		if c.Status != uikit.CheckWarn {
			t.Errorf("%q %q, want warn", c.Status, c.Message)
		}
	})
	t.Run("mismatch-names-both-and-remedy", func(t *testing.T) {
		home := t.TempDir()
		pvWriteClaudeRegistry(t, home, pvClaudeRegistry([2]string{"user", "3.4.4"}))
		c := checkPluginVersionAt(home, "", nil, pvCodexAbsent, testBinaryVersion, false)
		for _, want := range []string{"3.4.4", testPluginVersion, "claude plugin update moai@moai-adk"} {
			if !strings.Contains(c.Message, want) {
				t.Errorf("claude message %q lacks %q", c.Message, want)
			}
		}
		r := pvListRunner(pvCodexList("3.4.4"), nil)
		c = checkPluginVersionAt("", t.TempDir(), r, pvCodexFound, testBinaryVersion, false)
		if c.Status != uikit.CheckWarn {
			t.Errorf("codex mismatch: %q %q, want warn", c.Status, c.Message)
		}
		for _, want := range []string{"3.4.4", testPluginVersion, "codex plugin marketplace upgrade moai-adk", "codex plugin add moai@moai-adk"} {
			if !strings.Contains(c.Message, want) {
				t.Errorf("codex message %q lacks %q", c.Message, want)
			}
		}
	})
	t.Run("not-installed", func(t *testing.T) {
		c := checkPluginVersionAt(t.TempDir(), t.TempDir(), nil, pvCodexAbsent, testBinaryVersion, false)
		pvRequireInconclusive(t, c)
	})
	t.Run("home-absent", func(t *testing.T) {
		gone := filepath.Join(t.TempDir(), "no", "such", "home")
		r := pvListRunner(nil, errors.New("exit status 1"))
		c := checkPluginVersionAt(gone, gone, r, pvCodexFound, testBinaryVersion, false)
		pvRequireInconclusive(t, c)
	})
	t.Run("registry-malformed", func(t *testing.T) {
		home := t.TempDir()
		pvWriteClaudeRegistry(t, home, "{not json")
		pvRequireInconclusive(t, checkPluginVersionAt(home, "", nil, pvCodexAbsent, testBinaryVersion, false))
	})
	t.Run("registry-unknown-shape", func(t *testing.T) {
		for _, body := range []string{
			`{"version":9,"plugins":["x"]}`,
			`{"plugins":{"moai@moai-adk":"oops"}}`,
			`{"plugins":{"moai@moai-adk":[{"scope":"user"}]}}`,
			`{"plugins":{"moai@moai-adk":[42]}}`,
			`[]`,
		} {
			home := t.TempDir()
			pvWriteClaudeRegistry(t, home, body)
			pvRequireInconclusive(t, checkPluginVersionAt(home, "", nil, pvCodexAbsent, testBinaryVersion, false))
		}
	})
	t.Run("dev-build", func(t *testing.T) {
		home := t.TempDir()
		pvWriteClaudeRegistry(t, home, pvClaudeRegistry([2]string{"user", "3.4.4"}))
		for _, bin := range []string{"dev", "v0.0.0-dirty", "", "none"} {
			pvRequireInconclusive(t, checkPluginVersionAt(home, "", nil, pvCodexAbsent, bin, false))
		}
	})
	t.Run("codex-timeout", func(t *testing.T) {
		r := pvListRunner(nil, fmt.Errorf("probe: %w", context.DeadlineExceeded))
		pvRequireInconclusive(t, checkPluginVersionAt("", t.TempDir(), r, pvCodexFound, testBinaryVersion, false))
	})
}

func TestCheckPluginVersion_OutputBounded(t *testing.T) {
	cases := map[string]func(t *testing.T) DiagnosticCheck{
		"equal": func(t *testing.T) DiagnosticCheck {
			home := t.TempDir()
			pvWriteClaudeRegistry(t, home, pvClaudeRegistry([2]string{"user", testPluginVersion}))
			return checkPluginVersionAt(home, "", nil, pvCodexAbsent, testBinaryVersion, false)
		},
		"mismatch": func(t *testing.T) DiagnosticCheck {
			home := t.TempDir()
			pvWriteClaudeRegistry(t, home, pvClaudeRegistry([2]string{"user", "3.4.4"}))
			return checkPluginVersionAt(home, "", nil, pvCodexAbsent, testBinaryVersion, false)
		},
		"not-installed": func(t *testing.T) DiagnosticCheck {
			return checkPluginVersionAt(t.TempDir(), "", nil, pvCodexAbsent, testBinaryVersion, false)
		},
	}
	for name, run := range cases {
		t.Run(name, func(t *testing.T) {
			c := run(t)
			if c.Detail != "" {
				t.Errorf("default run carries detail %q, want none", c.Detail)
			}
			if c.Message == "" || strings.Contains(c.Message, "\n") {
				t.Errorf("default run message %q, want one non-empty line", c.Message)
			}
		})
	}
	t.Run("verbose-adds-detail", func(t *testing.T) {
		home := t.TempDir()
		pvWriteClaudeRegistry(t, home, pvClaudeRegistry([2]string{"user", testPluginVersion}))
		c := checkPluginVersionAt(home, "", nil, pvCodexAbsent, testBinaryVersion, true)
		if c.Detail == "" {
			t.Error("--verbose adds no detail")
		}
	})
}

// pvWriteRecordingShims installs `claude` and `codex` scripts that append their
// argv to record.
func pvWriteRecordingShims(t *testing.T, dir, record string) {
	t.Helper()
	for _, name := range []string{"claude", "codex"} {
		script := "#!/bin/sh\necho \"" + name + " $*\" >> \"" + record + "\"\n"
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

// pvHashTree digests directory entries (path and type) as well as file content.
func pvHashTree(t *testing.T, root string) string {
	t.Helper()
	h := sha256.New()
	n := 0
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		n++
		rel, _ := filepath.Rel(root, p)
		fmt.Fprintf(h, "%s|%v\n", rel, d.IsDir())
		if !d.IsDir() {
			b, rerr := os.ReadFile(p)
			if rerr != nil {
				return rerr
			}
			h.Write(b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n < 2 {
		t.Fatalf("pvHashTree swept %d entries under %s: an empty sweep asserts nothing", n, root)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func pvEnvMinusCodexHome(env []string) []string {
	var out []string
	for _, e := range env {
		if !strings.HasPrefix(e, codexHomeEnvVar+"=") {
			out = append(out, e)
		}
	}
	sort.Strings(out)
	return out
}

func pvEnvValue(env []string, key string) (string, bool) {
	for _, e := range env {
		if v, ok := strings.CutPrefix(e, key+"="); ok {
			return v, true
		}
	}
	return "", false
}

// TestCheckPluginVersion_HomeIsolation pins AC-021 (c): the doctor probe
// resolves the Codex home through the existing seam, starts its command only
// through the refusing runner, and no test reaches a real home.
func TestCheckPluginVersion_HomeIsolation(t *testing.T) {
	t.Run("child-env-carries-resolved-home", func(t *testing.T) {
		pvSetBinaryVersion(t, testBinaryVersion)
		pvWithCodexLookPath(t, pvCodexFound)
		t.Setenv(config.EnvClaudeConfigDir, t.TempDir())

		seamHome := t.TempDir()
		pvWithCodexUserHomeDir(t, func() (string, error) { return seamHome, nil })
		for _, tc := range []struct {
			name, env, want string
		}{
			{"variable-set", t.TempDir(), ""},
			{"variable-unset", "", filepath.Join(seamHome, ".codex")},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Setenv(codexHomeEnvVar, tc.env)
				want := tc.want
				if want == "" {
					want = tc.env
				}
				r := pvListRunner(pvCodexList(testPluginVersion), nil)
				withPluginRunner(t, r)
				checkPluginVersion(false)
				if len(r.calls) != 1 {
					t.Fatalf("runner calls %q, want exactly one probe", r.vectors())
				}
				got, ok := pvEnvValue(r.calls[0].env, codexHomeEnvVar)
				if !ok || got != want {
					t.Errorf("child CODEX_HOME = %q (set=%v), want %q", got, ok, want)
				}
				if a, b := strings.Join(pvEnvMinusCodexHome(r.calls[0].env), "\x00"), strings.Join(pvEnvMinusCodexHome(os.Environ()), "\x00"); a != b {
					t.Errorf("the child environment differs from the parent's beyond CODEX_HOME")
				}
			})
		}
	})

	t.Run("canary-home-not-touched", func(t *testing.T) {
		canary := t.TempDir()
		if err := os.MkdirAll(filepath.Join(canary, ".codex", "plugins", "cache"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(canary, ".codex", "config.toml"), []byte("[plugins.\"moai@moai-adk\"]\nenabled = true\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		calls := 0
		pvWithCodexUserHomeDir(t, func() (string, error) { calls++; return canary, nil })
		t.Setenv(codexHomeEnvVar, "")
		t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
		t.Setenv(config.EnvClaudeConfigDir, t.TempDir())

		before := pvHashTree(t, canary)
		runDiagnosticChecks(false, pluginVersionCheckName)
		if calls < 1 {
			t.Errorf("the check never resolved the Codex home through codexUserHomeDir (calls=%d)", calls)
		}
		runDiagnosticChecks(false, "")
		if after := pvHashTree(t, canary); after != before {
			t.Error("a registry-wide run changed the canary home")
		}
	})

	t.Run("registry-wide-starts-nothing", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("shell shims")
		}
		shimDir, record := t.TempDir(), filepath.Join(t.TempDir(), "record")
		pvWriteRecordingShims(t, shimDir, record)

		// Positive control: the shims record when invoked, so an empty record
		// below means nothing started, not that the shims are inert.
		for _, name := range []string{"claude", "codex"} {
			if err := exec.Command(filepath.Join(shimDir, name), "control").Run(); err != nil {
				t.Fatalf("control run of the %s shim: %v", name, err)
			}
		}
		b, _ := os.ReadFile(record)
		if !strings.Contains(string(b), "claude control") || !strings.Contains(string(b), "codex control") {
			t.Fatalf("shims did not record when invoked: %q", b)
		}
		if err := os.Remove(record); err != nil {
			t.Fatal(err)
		}

		t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))
		t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
		t.Setenv(config.EnvClaudeConfigDir, t.TempDir())
		t.Setenv(codexHomeEnvVar, t.TempDir())
		withPluginRunner(t, execPluginRunner{})

		runDiagnosticChecks(false, "")
		if got, err := os.ReadFile(record); err == nil && len(got) != 0 {
			t.Errorf("a registry-wide run started a tool: %q", got)
		}
	})

	t.Run("testmain-sandbox-redirects-codex-home", func(t *testing.T) {
		if capturedRealHome == "" {
			t.Fatal("TestMain captured no real home; the comparison would be vacuous")
		}
		got, err := codexUserHomeDir()
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Clean(got) == filepath.Clean(capturedRealHome) {
			t.Fatalf("codexUserHomeDir() returned the real home %q", got)
		}
		if filepath.Clean(got) != filepath.Clean(homeSandboxDir) {
			t.Errorf("codexUserHomeDir() = %q, want the TestMain sandbox %q", got, homeSandboxDir)
		}
		if v := os.Getenv(codexHomeEnvVar); v != "" {
			t.Errorf("CODEX_HOME = %q at the start of the test run, want unset", v)
		}
	})
}
