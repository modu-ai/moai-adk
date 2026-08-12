package cli

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	mcpcat "github.com/modu-ai/moai-adk/internal/mcp"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// writeMCPConfig writes a `.moai/config/sections/mcp.yaml` with the given tool
// enablement map under the nested `mcp.tools.<name>.enabled` path. Tools not
// in the map are omitted (default-enabled).
func writeMCPConfig(t *testing.T, dir string, disabled map[string]bool) {
	t.Helper()
	cfgDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	var b []byte
	b = append(b, "mcp:\n  tools:\n"...)
	for _, name := range mcpcat.MoaiMCPToolNames() {
		if !disabled[name] {
			continue
		}
		b = append(b, "    "...)
		b = append(b, name...)
		b = append(b, ":\n      enabled: false\n"...)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "mcp.yaml"), b, 0o644); err != nil {
		t.Fatalf("write mcp.yaml: %v", err)
	}
}

// listToolNames constructs an in-process MCP server, runs initialize +
// tools/list, and returns the registered tool names.
func listToolNames(t *testing.T) []string {
	t.Helper()
	srv := newMoaiMCPServer()
	if srv == nil {
		t.Fatal("newMoaiMCPServer returned nil")
	}
	ctx := context.Background()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	res, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}
	names := make([]string, 0, len(res.Tools))
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
	}
	sort.Strings(names)
	return names
}

// TestAC_C_004_DisabledToolAbsentFromToolsList is the REQ-C-2 falsifiability
// test (acceptance.md AC-C-004). Given a project configuration disabling one
// tool, when the server starts and a host requests tools/list, the disabled
// tool is absent AND the remaining tools are present. A test asserting only
// that the config value round-trips does NOT satisfy this AC — the effect
// (absent from tools/list) is what is asserted here.
func TestAC_C_004_DisabledToolAbsentFromToolsList(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)
	// Disable exactly one write-capable tool.
	writeMCPConfig(t, tmp, map[string]bool{"goal_arm": true})

	names := listToolNames(t)

	for _, n := range names {
		if n == "goal_arm" {
			t.Errorf("goal_arm is present in tools/list but was disabled in mcp.yaml (AC-C-004)")
		}
	}

	// Every OTHER catalog tool must still be present.
	present := make(map[string]bool, len(names))
	for _, n := range names {
		present[n] = true
	}
	for _, def := range mcpcat.MoaiMCPTools() {
		if def.Name == "goal_arm" {
			continue
		}
		if !present[def.Name] {
			t.Errorf("tool %q missing from tools/list after disabling only goal_arm (AC-C-004: remaining tools must be present)", def.Name)
		}
	}
}

// TestAC_C_004_DisabledToolAbsent_ReadonlyTool also covers disabling a
// read-only tool, so the gating seam is not accidentally write-only.
func TestAC_C_004_DisabledToolAbsent_ReadonlyTool(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)
	writeMCPConfig(t, tmp, map[string]bool{"spec_audit": true})

	names := listToolNames(t)
	present := make(map[string]bool, len(names))
	for _, n := range names {
		present[n] = true
	}
	if present["spec_audit"] {
		t.Errorf("spec_audit is present in tools/list but was disabled (AC-C-004)")
	}
	for _, def := range mcpcat.MoaiMCPTools() {
		if def.Name == "spec_audit" {
			continue
		}
		if !present[def.Name] {
			t.Errorf("tool %q missing after disabling only spec_audit (AC-C-004)", def.Name)
		}
	}
}

// TestAC_C_004_NoConfig_AllEnabled verifies the owner-decided default: with no
// mcp.yaml at all, all 17 tools are registered (all-enabled default).
func TestAC_C_004_NoConfig_AllEnabled(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)
	// Intentionally write NO mcp.yaml.

	names := listToolNames(t)
	present := make(map[string]bool, len(names))
	for _, n := range names {
		present[n] = true
	}
	for _, def := range mcpcat.MoaiMCPTools() {
		if !present[def.Name] {
			t.Errorf("tool %q missing from tools/list with no mcp.yaml (all-enabled default, AC-C-004)", def.Name)
		}
	}
}

// TestAC_C_005_PerToolFieldsInSchema verifies the M1-adjacent half of AC-C-005:
// each of the 17 tools has a corresponding enablement field in
// settings.AllFields() under the MCP section, so the setting round-trips
// through the same schema-driven form + yamlpatch seam the audit-selection and
// branch_guard fields use. Mirrors the assertion shape of
// internal/web/mcp_audit_surface_test.go:22-40.
func TestAC_C_005_PerToolFieldsInSchema(t *testing.T) {
	fields := settings.AllFields()
	got := make(map[string]bool, len(fields))
	for _, f := range fields {
		got[f.Name] = true
	}
	for _, def := range mcpcat.MoaiMCPTools() {
		want := "mcp.tools." + def.Name + ".enabled"
		if !got[want] {
			t.Errorf("per-tool enablement field %q missing from settings.AllFields() (AC-C-005)", want)
		}
	}
}

// TestMoaiMCPServer_RegistrationMatchesCatalog is the single-declaration guard
// (REQ-C-1 / C-C-5 / AP-C-4). The tools registered by the server (with no
// disabling config) MUST equal the catalog exactly — no more, no less — so a
// tool added to registerMoaiMCPTools without a catalog entry (or vice versa)
// fails the build.
func TestMoaiMCPServer_RegistrationMatchesCatalog(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	names := listToolNames(t)
	present := make(map[string]bool, len(names))
	for _, n := range names {
		present[n] = true
	}
	for _, def := range mcpcat.MoaiMCPTools() {
		if !present[def.Name] {
			t.Errorf("catalog tool %q not registered (AP-C-4 single-declaration drift)", def.Name)
		}
	}
	if len(names) != len(mcpcat.MoaiMCPTools()) {
		t.Errorf("registered tool count = %d, catalog count = %d (AP-C-4: a registered tool has no catalog entry)", len(names), len(mcpcat.MoaiMCPTools()))
	}
}

// TestReadMCPToolEnablement_MalformedYAML verifies the fail-OPEN default: a
// malformed mcp.yaml (unparseable) yields all-enabled rather than erroring or
// disabling everything. This is the inverse posture of the codex gates.
func TestReadMCPToolEnablement_MalformedYAML(t *testing.T) {
	tmp := t.TempDir()
	cfgDir := filepath.Join(tmp, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Deliberately invalid YAML (unterminated flow mapping).
	if err := os.WriteFile(filepath.Join(cfgDir, "mcp.yaml"),
		[]byte("mcp: [unclosed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	enabled := readMCPToolEnablement(tmp)
	for _, def := range mcpcat.MoaiMCPTools() {
		if !enabled[def.Name] {
			t.Errorf("tool %q disabled by malformed yaml (fail-OPEN default requires all-enabled)", def.Name)
		}
	}
}
