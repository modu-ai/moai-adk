package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

func TestToolEnableCodex_CommandSurface(t *testing.T) {
	cmd := newToolCmd()
	if cmd.Use != "tool" {
		t.Fatalf("tool command Use = %q, want tool", cmd.Use)
	}

	enable, _, err := cmd.Find([]string{"enable"})
	if err != nil {
		t.Fatalf("find tool enable: %v", err)
	}
	codex, _, err := enable.Find([]string{"codex"})
	if err != nil {
		t.Fatalf("find tool enable codex: %v", err)
	}
	if codex.Use != "codex" {
		t.Fatalf("codex command Use = %q, want codex", codex.Use)
	}
	if codex.Flags().Lookup("dry-run") == nil {
		t.Fatal("tool enable codex must expose --dry-run")
	}
	if codex.Flags().Lookup("project-root") == nil {
		t.Fatal("tool enable codex must expose --project-root")
	}
}

func TestToolEnableCodex_WiresExistingProject(t *testing.T) {
	root := t.TempDir()
	var out, warn bytes.Buffer

	if err := runToolEnableCodexAt(root, &out, &warn, false); err != nil {
		t.Fatalf("runToolEnableCodexAt: %v", err)
	}
	for _, rel := range []string{
		codexwiring.HooksRelPath,
		codexwiring.ConfigRelPath,
		codexwiring.SidecarPath,
	} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Errorf("artifact %s not created: %v", rel, err)
		}
	}
}

func TestToolEnableCodex_DryRunDoesNotWrite(t *testing.T) {
	root := t.TempDir()
	var out, warn bytes.Buffer

	if err := runToolEnableCodexAt(root, &out, &warn, true); err != nil {
		t.Fatalf("runToolEnableCodexAt dry-run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".codex")); !os.IsNotExist(err) {
		t.Fatalf("dry-run created .codex: %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"moai tool enable codex",
		codexwiring.HooksRelPath,
		codexwiring.ConfigRelPath,
		filepath.ToSlash(codexwiring.SidecarPath),
	} {
		if !strings.Contains(got, want) {
			t.Errorf("dry-run output missing %q:\n%s", want, got)
		}
	}
}

func TestUpdateAddCodex_IsDeprecatedAlias(t *testing.T) {
	f := updateCmd.Flags().Lookup("add-codex")
	if f == nil {
		t.Fatal("--add-codex compatibility alias is missing")
	}
	if !strings.Contains(f.Deprecated, "moai tool enable codex") {
		t.Fatalf("--add-codex deprecation = %q, want replacement command", f.Deprecated)
	}
}

func TestInitAddCodexGuidancePrefersToolCommand(t *testing.T) {
	if !strings.Contains(addCodexReinitGuidance, "moai tool enable codex") {
		t.Fatalf("init guidance does not prefer the tool command: %s", addCodexReinitGuidance)
	}
}
