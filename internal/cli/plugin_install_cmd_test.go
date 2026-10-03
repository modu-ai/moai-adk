package cli

// plugin_install_cmd_test.go — SPEC-PLUGIN-MARKETPLACE-001 M3a, AC-019: the
// `moai plugin install` verb.

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// executeVerb runs a fresh verb command with args and returns its error and
// the combined stderr/stdout capture.
func executeVerb(t *testing.T, args ...string) (error, string) {
	t.Helper()
	cmd := newPluginCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(append([]string{"install"}, args...))
	err := cmd.Execute()
	return err, out.String()
}

func TestPluginInstallCmd(t *testing.T) {
	t.Run("registered-in-root-help", func(t *testing.T) {
		found, _, err := rootCmd.Find([]string{"plugin", "install"})
		if err != nil || found == nil || found.Name() != "install" {
			t.Fatalf("rootCmd.Find(plugin install) = %v, %v; want the install leaf", found, err)
		}
		plugin, _, _ := rootCmd.Find([]string{"plugin"})
		if plugin == nil || plugin.Name() != "plugin" || plugin.GroupID != "tools" {
			t.Fatalf("plugin noun missing or not in the tools help group: %+v", plugin)
		}
		if help := runRootHelpCapture(t); !strings.Contains(help, "plugin") {
			t.Fatalf("root --help does not list the plugin command:\n%s", help)
		}
	})
	t.Run("help-names-opt-out", func(t *testing.T) {
		found, _, _ := newPluginCmd().Find([]string{"install"})
		if found == nil || !strings.Contains(found.Long, config.EnvSkipPluginInstall) {
			t.Fatalf("install help must name %s", config.EnvSkipPluginInstall)
		}
	})
	t.Run("exit-0-on-fail-open-outcomes", func(t *testing.T) {
		setupPluginTools(t)
		failing := &fakePluginRunner{fn: func(context.Context, pluginCall, int) ([]byte, error) {
			return []byte("no"), errors.New("exit status 1")
		}}
		withPluginRunner(t, failing)
		err, out := executeVerb(t)
		if err != nil {
			t.Fatalf("verb returned %v after tool failures, want nil", err)
		}
		if len(failing.calls) == 0 || !strings.Contains(out, guidanceMarker) {
			t.Fatalf("verb did not run the commands / print guidance (calls=%d):\n%s", len(failing.calls), out)
		}

		// No tool on PATH at all: skip lines, still exit 0.
		t.Setenv(config.EnvClaudeBin, "")
		t.Setenv("PATH", t.TempDir())
		codexWiringLookPath = func(string) (string, error) { return "", os.ErrNotExist }
		none := &fakePluginRunner{}
		withPluginRunner(t, none)
		err, out = executeVerb(t)
		if err != nil || len(none.calls) != 0 || pluginCountLines(out, "not found on PATH") != 2 {
			t.Fatalf("no-tool run: err=%v calls=%d out:\n%s", err, len(none.calls), out)
		}

		// Opt-out: zero commands, exit 0.
		setupPluginTools(t)
		t.Setenv(config.EnvSkipPluginInstall, "1")
		optout := &fakePluginRunner{}
		withPluginRunner(t, optout)
		if err, _ = executeVerb(t); err != nil || len(optout.calls) != 0 {
			t.Fatalf("opt-out run: err=%v calls=%d", err, len(optout.calls))
		}
	})
	t.Run("exit-nonzero-on-unknown-flag", func(t *testing.T) {
		setupPluginTools(t)
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		err, _ := executeVerb(t, "--no-such-flag")
		if err == nil {
			t.Fatal("unknown flag accepted")
		}
		if len(r.calls) != 0 {
			t.Fatalf("a usage error still ran commands: %q", r.vectors())
		}
	})
	t.Run("no-harness-filter", func(t *testing.T) {
		e := setupPluginTools(t)
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		if err, _ := executeVerb(t); err != nil {
			t.Fatalf("verb: %v", err)
		}
		if len(r.calls) != 4 || r.calls[0].bin != e.claudeBin || r.calls[2].bin != e.codexBin {
			t.Fatalf("verb must act on every tool found (claude pair then codex pair), got %q", r.vectors())
		}
		found, _, _ := newPluginCmd().Find([]string{"install"})
		if found == nil || found.Flags().Lookup("llm") != nil {
			t.Fatal("the verb has no harness; it must not carry an --llm flag")
		}
	})
	t.Run("project-pin-read-from-working-directory", func(t *testing.T) {
		// The verb has no init target: Claude's llm.claude_bin pin is read from
		// the project the working directory sits in, like the launcher does.
		root := fakeMoaiProject(t)
		pin := writeExecutable(t, filepath.Join(t.TempDir(), "claude-from-config"))
		sections := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(sections, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte("llm:\n  claude_bin: "+pin+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		setupPluginTools(t)
		t.Setenv(config.EnvClaudeBin, "") // let the config pin decide
		r := &fakePluginRunner{}
		withPluginRunner(t, r)
		if err, _ := executeVerb(t); err != nil {
			t.Fatalf("verb: %v", err)
		}
		if len(r.calls) == 0 || r.calls[0].bin != pin {
			t.Fatalf("first call = %q, want the config-pinned %q", r.vectors(), pin)
		}
	})
}
