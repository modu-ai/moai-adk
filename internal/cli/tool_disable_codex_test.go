package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// TestToolDisableCodexCommand covers the `moai tool disable codex` surface
// (REQ-DHR-005): the command is registered beside `tool enable codex` with
// the same flags, its report names what it removed and what it kept and why,
// a part MoAI cannot prove it wrote carries manual-removal guidance, and
// --dry-run writes nothing.
func TestToolDisableCodexCommand(t *testing.T) {
	t.Run("registered", func(t *testing.T) {
		cmd := newToolCmd()
		found, _, err := cmd.Find([]string{"disable", "codex"})
		if err != nil || found.Name() != "codex" || found.Parent().Name() != "disable" {
			t.Fatalf("tool disable codex not registered: %v", err)
		}
		if found.Flags().Lookup("dry-run") == nil || found.Flags().Lookup("project-root") == nil {
			t.Fatal("disable codex lacks the flags enable codex carries")
		}
	})

	t.Run("report", func(t *testing.T) {
		root := t.TempDir()
		// An older binary's trace: a MoAI-looking table and the sidecar, but
		// no part record — the table is recorded unknown on the next wiring.
		for rel, body := range map[string]string{
			codexwiring.ConfigRelPath: "[mcp_servers.moai]\ncommand = \"moai\"\n",
			codexwiring.SidecarPath:   "{}\n",
		} {
			path := filepath.Join(root, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		var out, warn bytes.Buffer
		if err := runToolEnableCodexAt(root, &out, &warn, false); err != nil {
			t.Fatal(err)
		}
		out.Reset()
		if err := runToolDisableCodexAt(root, &out, &warn, false); err != nil {
			t.Fatalf("disable: %v (%s)", err, warn.String())
		}
		report := out.String()
		for _, want := range []string{
			"removed " + codexwiring.HooksRelPath,
			"removed " + codexwiring.ConfigRelPath + " [tui]",
			"kept " + codexwiring.ConfigRelPath + " [mcp_servers.moai] (unknown-origin)",
			"remove it by hand",
			"moai tool enable codex",
		} {
			if !strings.Contains(report, want) {
				t.Errorf("report lacks %q:\n%s", want, report)
			}
		}
	})

	t.Run("dry_run_writes_nothing", func(t *testing.T) {
		root := wiredProject(t)
		before := treeHashes(t, root)
		var out, warn bytes.Buffer
		if err := runToolDisableCodexAt(root, &out, &warn, true); err != nil {
			t.Fatal(err)
		}
		after := treeHashes(t, root)
		if len(after) != len(before) {
			t.Fatal("dry-run changed the tree")
		}
		for k, v := range before {
			if after[k] != v {
				t.Fatalf("dry-run changed %s", k)
			}
		}
		if !strings.Contains(out.String(), "nothing written") {
			t.Fatalf("dry-run output: %s", out.String())
		}
	})
}
