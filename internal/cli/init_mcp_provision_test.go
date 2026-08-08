package cli

// init_mcp_provision_test.go — REQ-MCP-015 / AC-MCP-002 / AC-MCP-006.
//
// Covers the `moai init` opt-in wiring that turns the wizard's
// `mcp_tools_opt_in` answer into exactly one neutral `.mcp.json` entry. The
// provisioning seam itself (provisionMoaiMCPServerEntryAt) is covered by
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

// TestProvisionMCPEntryIfOptedIn_OptedOut verifies opt-in default-off (C6 /
// AC-MCP-002): when the user declined, no `.mcp.json` is created and nothing
// is announced.
func TestProvisionMCPEntryIfOptedIn_OptedOut(t *testing.T) {
	tmp := t.TempDir()
	var out, errOut bytes.Buffer

	provisionMCPEntryIfOptedIn(&out, &errOut, tmp, false)

	if _, err := os.Stat(filepath.Join(tmp, ".mcp.json")); !os.IsNotExist(err) {
		t.Errorf("declining the opt-in must leave .mcp.json absent, stat err = %v", err)
	}
	if out.Len() != 0 || errOut.Len() != 0 {
		t.Errorf("declining the opt-in must be silent, stdout=%q stderr=%q", out.String(), errOut.String())
	}
}

// TestProvisionMCPEntryIfOptedIn_OptedIn verifies the reachability link
// (AC-MCP-006): opting in writes exactly one neutral `moai` entry into the
// project's `.mcp.json` and reports it on stdout.
func TestProvisionMCPEntryIfOptedIn_OptedIn(t *testing.T) {
	tmp := t.TempDir()
	var out, errOut bytes.Buffer

	provisionMCPEntryIfOptedIn(&out, &errOut, tmp, true)

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

// TestProvisionMCPEntryIfOptedIn_FailureIsNonFatal verifies the best-effort
// contract: a provisioning failure warns on stderr and never panics, so a
// broken config can never fail `moai init`.
func TestProvisionMCPEntryIfOptedIn_FailureIsNonFatal(t *testing.T) {
	tmp := t.TempDir()
	// A directory where the file must be makes the atomic write fail.
	if err := os.Mkdir(filepath.Join(tmp, ".mcp.json"), 0o755); err != nil {
		t.Fatalf("seed unwritable .mcp.json: %v", err)
	}
	var out, errOut bytes.Buffer

	provisionMCPEntryIfOptedIn(&out, &errOut, tmp, true)

	if !strings.Contains(strings.ToLower(errOut.String()), "warning") {
		t.Errorf("a provisioning failure must warn on stderr, got %q", errOut.String())
	}
}

// TestRunInit_CallsMCPProvisioning is the reachability guard: it asserts by
// source inspection that runInit actually calls the provisioning helper gated
// on opts.MCPToolsOptIn. Without this call the wizard question is collected
// and then dropped, and the MCP server stays unreachable at runtime — the
// regression this test exists to prevent.
func TestRunInit_CallsMCPProvisioning(t *testing.T) {
	src, err := os.ReadFile("init.go")
	if err != nil {
		t.Fatalf("read init.go: %v", err)
	}
	body := string(src)
	if !strings.Contains(body, "provisionMCPEntryIfOptedIn(") {
		t.Error("runInit must call provisionMCPEntryIfOptedIn — the wizard's mcp_tools_opt_in answer is otherwise collected and dropped")
	}
	if !strings.Contains(body, "opts.MCPToolsOptIn") {
		t.Error("the provisioning call must be gated on opts.MCPToolsOptIn (opt-in default-off, C6)")
	}
}
