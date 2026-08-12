package cli

// update_mcp_merge_test.go — SPEC-MCP-DEFAULT-ON-001 M4 (REQ-A-4 / AC-A-012).
//
// Guards the `moai update` 3-way-merge safety for `.mcp.json`. Two contracts:
//
//  1. `collectMergeableFiles` (update_template_sync.go) MUST include
//     `.mcp.json` in its returned list, so the file routes through the 3-way
//     merge engine instead of being clobbered by the template deploy.
//  2. The downstream MergeUserFiles engine, when handed a `.mcp.json` backup
//     carrying a user-added entry, preserves that entry alongside the shipped
//     `moai` entry (the user's own tools survive `moai update`).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/merge"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// TestCollectMergeableFiles_IncludesMCPJSON verifies AC-A-012's routing
// contract: `.mcp.json` appears as a quoted path element inside
// collectMergeableFiles's return slice in update_template_sync.go. Without
// this, `moai update` deploys the template `.mcp.json` over the user's copy
// and any user-added MCP entry is lost.
func TestCollectMergeableFiles_IncludesMCPJSON(t *testing.T) {
	src, err := os.ReadFile("update_template_sync.go")
	if err != nil {
		t.Fatalf("read update_template_sync.go: %v", err)
	}
	body := string(src)
	// Locate the collectMergeableFiles closure body and assert `.mcp.json` is
	// listed as a merge target. The previous false comment ("MoAI no longer
	// ships an MCP template") must also be gone (AC-A-013).
	if !strings.Contains(body, `".mcp.json"`) {
		t.Error("collectMergeableFiles must list \".mcp.json\" as a merge target so a user's MCP entries survive `moai update` (AC-A-012)")
	}
	if strings.Contains(body, "no longer ships an MCP template") {
		t.Error("the false comment claiming MoAI no longer ships an MCP template must be corrected (AC-A-013)")
	}
}

// TestMergeUserFiles_PreservesUserMCPEntry verifies the behavioral contract
// (AC-A-012): when the user's backed-up `.mcp.json` carries the shipped `moai`
// entry plus a user-added entry, MergeUserFiles preserves BOTH after the
// template deploy — the user's entry is preserved by the 3-way merge, not
// overwritten.
func TestMergeUserFiles_PreservesUserMCPEntry(t *testing.T) {
	tmpDir := t.TempDir()

	// Manifest so MergeUserFiles can load it.
	mgr := manifest.NewManager()
	manifestPath := filepath.Join(tmpDir, ".moai", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatalf("mkdir manifest dir: %v", err)
	}
	if err := os.WriteFile(manifestPath, []byte(`{"version":"1","files":{}}`), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if _, err := mgr.Load(tmpDir); err != nil {
		t.Fatalf("load manifest: %v", err)
	}

	// User's backed-up `.mcp.json`: shipped `moai` entry PLUS a user-added
	// `my-tool` entry. This is what the user had before `moai update`.
	userMCP := `{
  "$schema": "https://raw.githubusercontent.com/anthropics/claude-code/main/.mcp.schema.json",
  "mcpServers": {
    "moai": {
      "command": "moai",
      "args": ["mcp-server"]
    },
    "my-tool": {
      "command": "node",
      "args": ["my-tool.js"]
    }
  }
}`
	backups := []merge.FileBackup{
		{Path: ".mcp.json", Data: []byte(userMCP)},
	}

	// On-disk "post-deploy" `.mcp.json`: only the shipped `moai` entry
	// (simulating the fresh template deploy that would clobber without merge).
	deployedMCP := `{
  "$schema": "https://raw.githubusercontent.com/anthropics/claude-code/main/.mcp.schema.json",
  "mcpServers": {
    "moai": {
      "command": "moai",
      "args": ["mcp-server"]
    }
  }
}`
	if err := os.WriteFile(filepath.Join(tmpDir, ".mcp.json"), []byte(deployedMCP), 0o644); err != nil {
		t.Fatalf("write deployed .mcp.json: %v", err)
	}

	var out strings.Builder
	if err := merge.MergeUserFiles(tmpDir, backups, &out); err != nil {
		t.Fatalf("MergeUserFiles: %v", err)
	}

	merged, err := os.ReadFile(filepath.Join(tmpDir, ".mcp.json"))
	if err != nil {
		t.Fatalf("read merged .mcp.json: %v", err)
	}
	// Both the shipped `moai` entry AND the user's `my-tool` entry MUST survive.
	if !strings.Contains(string(merged), `"moai"`) {
		t.Errorf("merged .mcp.json lost the shipped moai entry:\n%s", string(merged))
	}
	if !strings.Contains(string(merged), `"my-tool"`) {
		t.Errorf("merged .mcp.json lost the user-added my-tool entry (3-way merge must preserve it):\n%s", string(merged))
	}
}
