package cli

// init_mcp_provision_test.go — REQ-MCP-015 / AC-MCP-002 / AC-MCP-006
// (amended 2026-08-12 per SPEC-MCP-DEFAULT-ON-001).
//
// Covers the `moai init` default-on MCP provisioning wiring. After the
// SPEC-MCP-DEFAULT-ON-001 inversion, provisioning is DEFAULT-ON: a fresh
// project that accepts the wizard defaults gets the single neutral `moai`
// entry written into `.mcp.json`; an explicit decline is honored silently.
//
// The provisioning seam itself (provisionMoaiMCPServerEntryAt) is covered by
// mcp_server_test.go; these tests cover the CALL SITE — the link that makes
// the seam reachable at runtime instead of test-only code.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestProvisionMCPEntryUnlessDeclined_Default verifies the default-on path
// (SPEC-MCP-DEFAULT-ON-001 AC-A-008): when the user did not decline
// (declined=false), exactly one neutral `moai` entry is written into the
// project's `.mcp.json` and the provisioning is reported on stdout.
func TestProvisionMCPEntryUnlessDeclined_Default(t *testing.T) {
	tmp := t.TempDir()
	var out, errOut bytes.Buffer

	provisionMCPEntryUnlessDeclined(&out, &errOut, tmp, false)

	data, err := os.ReadFile(filepath.Join(tmp, ".mcp.json"))
	if err != nil {
		t.Fatalf("read provisioned .mcp.json: %v", err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("parse provisioned .mcp.json: %v", err)
	}
	servers, ok := root["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers missing/not-object: %v", root)
	}
	entry, ok := servers[moaiMCPServerKey].(map[string]any)
	if !ok {
		t.Fatalf("moai entry missing under mcpServers (key=%q): %v", moaiMCPServerKey, servers)
	}
	if entry["command"] != moaiMCPServerCommand {
		t.Errorf("command = %v, want %q", entry["command"], moaiMCPServerCommand)
	}
	if _, hasEnv := entry["env"]; hasEnv {
		t.Error("provisioned entry MUST NOT carry an env block (secret hygiene C3)")
	}
	if !strings.Contains(out.String(), ".mcp.json") {
		t.Errorf("stdout must report the provisioning, got %q", out.String())
	}
	if errOut.Len() != 0 {
		t.Errorf("successful provisioning must not warn, stderr=%q", errOut.String())
	}
}

// TestProvisionMCPEntryUnlessDeclined_Declined verifies the explicit-decline
// path (SPEC-MCP-DEFAULT-ON-001 AC-A-009): when the user explicitly declined
// (declined=true), no `.mcp.json` is created and nothing is announced on
// stdout or stderr.
func TestProvisionMCPEntryUnlessDeclined_Declined(t *testing.T) {
	tmp := t.TempDir()
	var out, errOut bytes.Buffer

	provisionMCPEntryUnlessDeclined(&out, &errOut, tmp, true)

	if _, err := os.Stat(filepath.Join(tmp, ".mcp.json")); !os.IsNotExist(err) {
		t.Errorf("an explicit decline must leave .mcp.json absent, stat err = %v", err)
	}
	if out.Len() != 0 || errOut.Len() != 0 {
		t.Errorf("an explicit decline must be silent, stdout=%q stderr=%q", out.String(), errOut.String())
	}
}

// TestProvisionMCPEntryUnlessDeclined_FailureIsNonFatal verifies the
// best-effort contract: a provisioning failure warns on stderr and never
// panics, so a broken config can never fail `moai init`.
func TestProvisionMCPEntryUnlessDeclined_FailureIsNonFatal(t *testing.T) {
	tmp := t.TempDir()
	// A directory where the file must be makes the atomic write fail.
	if err := os.Mkdir(filepath.Join(tmp, ".mcp.json"), 0o755); err != nil {
		t.Fatalf("seed unwritable .mcp.json: %v", err)
	}
	var out, errOut bytes.Buffer

	provisionMCPEntryUnlessDeclined(&out, &errOut, tmp, false)

	if !strings.Contains(strings.ToLower(errOut.String()), "warning") {
		t.Errorf("a provisioning failure must warn on stderr, got %q", errOut.String())
	}
}

// TestRunInit_CallsMCPProvisioning is the reachability guard: it asserts by
// source inspection that runInit actually calls the provisioning helper.
// Without this call the wizard answer is collected and then dropped, and the
// MCP server stays unreachable at runtime — the regression this test exists
// to prevent.
func TestRunInit_CallsMCPProvisioning(t *testing.T) {
	src, err := os.ReadFile("init.go")
	if err != nil {
		t.Fatalf("read init.go: %v", err)
	}
	body := string(src)
	if !strings.Contains(body, "provisionMCPEntryUnlessDeclined(") {
		t.Error("runInit must call provisionMCPEntryUnlessDeclined — without it the wizard answer is collected and dropped")
	}
	if !strings.Contains(body, "opts.MCPProvision") {
		t.Error("the provisioning call must be gated on opts.MCPProvision (default-on, SPEC-MCP-DEFAULT-ON-001)")
	}
}
