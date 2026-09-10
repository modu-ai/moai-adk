package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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

func sha256ToolFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func TestToolEnableCodex_PreservesMCPAndUserConfig(t *testing.T) {
	root := t.TempDir()
	mcp := filepath.Join(root, ".mcp.json")
	if err := os.WriteFile(mcp, []byte(`{"mcpServers":{"user-own":{"command":"user-cmd"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	beforeMCP := sha256ToolFile(t, mcp)
	codexDir := filepath.Join(root, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte("[user_own]\nkey = \"user-value\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, warn bytes.Buffer
	if err := runToolEnableCodexAt(root, &out, &warn, false); err != nil {
		t.Fatal(err)
	}
	if got := sha256ToolFile(t, mcp); got != beforeMCP {
		t.Errorf(".mcp.json changed: before %s, after %s", beforeMCP, got)
	}
	cfg, err := os.ReadFile(filepath.Join(codexDir, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"user-value", "mcp_servers.moai"} {
		if !strings.Contains(string(cfg), want) {
			t.Errorf("config.toml missing %q:\n%s", want, cfg)
		}
	}
}

func TestToolEnableCodex_RerunIsIdempotent(t *testing.T) {
	root := t.TempDir()
	var out1, warn1 bytes.Buffer
	if err := runToolEnableCodexAt(root, &out1, &warn1, false); err != nil {
		t.Fatal(err)
	}
	before := sha256ToolFile(t, filepath.Join(root, codexwiring.SidecarPath))
	var out2, warn2 bytes.Buffer
	if err := runToolEnableCodexAt(root, &out2, &warn2, false); err != nil {
		t.Fatal(err)
	}
	if got := sha256ToolFile(t, filepath.Join(root, codexwiring.SidecarPath)); got != before {
		t.Errorf("sidecar changed across an unchanged rerun: before %s, after %s", before, got)
	}
	if strings.Contains(out2.String(), codexwiring.ReTrustGuidance) || strings.Contains(out2.String(), codexwiring.FirstTrustGuidance) {
		t.Errorf("unchanged rerun printed trust guidance:\n%s", out2.String())
	}
}

func TestToolEnableCodex_ValidationRefusalFailsLoud(t *testing.T) {
	root := t.TempDir()
	codexDir := filepath.Join(root, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	bad := []byte("{\n  \"version\": 1,\n  \"hooks\": {}\n}\n")
	if err := os.WriteFile(filepath.Join(codexDir, "hooks.json"), bad, 0o644); err != nil {
		t.Fatal(err)
	}
	var out, warn bytes.Buffer
	err := runToolEnableCodexAt(root, &out, &warn, false)
	if !errors.Is(err, codexwiring.ErrValidationRefused) {
		t.Fatalf("error = %v, want validation refusal", err)
	}
	after, readErr := os.ReadFile(filepath.Join(codexDir, "hooks.json"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !bytes.Equal(after, bad) {
		t.Errorf("refused write changed hooks.json")
	}
}

func TestInitAddCodexGuidancePrefersToolCommand(t *testing.T) {
	if !strings.Contains(addCodexReinitGuidance, "moai tool enable codex") {
		t.Fatalf("init guidance does not prefer the tool command: %s", addCodexReinitGuidance)
	}
	if strings.Contains(addCodexReinitGuidance, "add-codex") {
		t.Fatalf("init guidance names removed --add-codex flag: %s", addCodexReinitGuidance)
	}
}
