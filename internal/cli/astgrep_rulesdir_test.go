package cli

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// t50: gate.yaml ast_grep_gate.rules_dir is the single source of truth for
// where a project keeps its ast-grep rules. The CLI commands (moai ast-grep,
// moai ast-edit) resolve their rules directory as
//
//	explicit --rules-dir flag  >  gate.yaml ast_grep_gate.rules_dir  >  empty
//
// with NO hardcoded path fallback: a code-level default cannot serve both the
// template deployment path (.moai/config/astgrep-rules, where distributed
// users' rules live) and per-project overrides (this repository tracks its
// ruleset at .moai/astgrep-rules), so the code stops guessing and the config
// answers. The writeGateYAML helper is shared with gate_config_test.go.

// gateYAMLWithRulesDir builds an ast_grep_gate body carrying rules_dir.
func gateYAMLWithRulesDir(rulesDir string) string {
	return "    rules_dir: \"" + rulesDir + "\"\n"
}

func TestResolveRulesDir(t *testing.T) {
	t.Parallel()

	t.Run("flag wins over config", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeGateYAML(t, dir, gateYAMLWithRulesDir(".moai/config/from-gate-yaml"))
		if got := resolveRulesDir("/explicit/flag/path", dir); got != "/explicit/flag/path" {
			t.Errorf("resolveRulesDir(flag, dir) = %q, want /explicit/flag/path", got)
		}
	})

	t.Run("empty flag resolves gate.yaml value (project-root relative)", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeGateYAML(t, dir, gateYAMLWithRulesDir(".moai/astgrep-rules"))
		// The quality gate resolves gate.yaml rules_dir against the project
		// root (RunAstGrepGateV2 joins projectDir); the CLI must agree so the
		// same config means the same directory when cwd differs from the
		// project root.
		want := filepath.Join(dir, ".moai", "astgrep-rules")
		if got := resolveRulesDir("", dir); got != want {
			t.Errorf("resolveRulesDir(\"\", dir) = %q, want %q", got, want)
		}
	})

	t.Run("absolute gate.yaml value passes through unjoined", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeGateYAML(t, dir, gateYAMLWithRulesDir("/opt/shared/astgrep-rules"))
		if got := resolveRulesDir("", dir); got != "/opt/shared/astgrep-rules" {
			t.Errorf("resolveRulesDir(\"\", dir) = %q, want /opt/shared/astgrep-rules (absolute must not be joined)", got)
		}
	})

	t.Run("no config yields empty, not a guessed path", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if got := resolveRulesDir("", dir); got != "" {
			t.Errorf("resolveRulesDir(\"\", empty dir) = %q, want empty", got)
		}
	})

	t.Run("empty gate.yaml value maps through as empty", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeGateYAML(t, dir, gateYAMLWithRulesDir(""))
		if got := resolveRulesDir("", dir); got != "" {
			t.Errorf("resolveRulesDir(\"\", dir-with-empty-value) = %q, want empty", got)
		}
	})
}

// TestAstGrepCmdRulesDirFlagNoDefault pins the flag registration: --rules-dir
// must default to empty (resolved from gate.yaml at run time), not to a
// hardcoded directory path.
func TestAstGrepCmdRulesDirFlagNoDefault(t *testing.T) {
	t.Parallel()

	for name, cmd := range map[string]*cobra.Command{
		"ast-grep": NewAstGrepCmd(),
		"ast-edit": NewAstEditCmd(),
	} {
		name, cmd := name, cmd
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fl := cmd.Flags().Lookup("rules-dir")
			if fl == nil {
				t.Fatalf("--rules-dir is not registered on %s", name)
			}
			if def := fl.DefValue; def != "" {
				t.Errorf("%s --rules-dir default = %q, want empty (resolved from gate.yaml)", name, def)
			}
		})
	}
}
