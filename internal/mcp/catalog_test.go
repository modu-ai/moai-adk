package mcp

import (
	"sort"
	"testing"
)

// wantCatalogSize is the catalog-size invariant (AC-C-001 / AC-C-002): the
// moai MCP server's tool surface. Update it ONLY together with a matching
// registration change in registerMoaiMCPTools — the registration/catalog
// equality guard (internal/cli TestMoaiMCPServer_RegistrationMatchesCatalog)
// catches drift in either direction.
const wantCatalogSize = 29

// TestMoaiMCPTools_Count29 asserts the catalog declares exactly
// wantCatalogSize tools, matching the registration count in
// registerMoaiMCPTools.
func TestMoaiMCPTools_Count29(t *testing.T) {
	tools := MoaiMCPTools()
	if len(tools) != wantCatalogSize {
		t.Fatalf("catalog declares %d tools, want %d", len(tools), wantCatalogSize)
	}
}

// TestMoaiMCPTools_NineWriteCapable asserts exactly the nine write-capable
// tools carry WriteCapable=true (REQ-C-3 / AC-C-003), and the other 16 are
// read-only. session_msg_list is read-only: it enumerates registered peers
// without touching the store, unlike register/send/poll which write an agent
// record, append a message, and claim an inbox respectively.
func TestMoaiMCPTools_NineWriteCapable(t *testing.T) {
	want := map[string]bool{
		"goal_arm":             true,
		"verify_snapshot":      true,
		"codex_task":           true,
		"codex_job_cancel":     true,
		"glm_task":             true,
		"glm_job_cancel":       true,
		"session_msg_register": true,
		"session_msg_send":     true,
		"session_msg_poll":     true,
	}
	var got []string
	for _, tool := range MoaiMCPTools() {
		if tool.WriteCapable {
			got = append(got, tool.Name)
		}
	}
	sort.Strings(got)
	if len(got) != 9 {
		t.Fatalf("write-capable tool count = %d (%v), want 9", len(got), got)
	}
	for _, name := range got {
		if !want[name] {
			t.Errorf("tool %q marked WriteCapable but not in the expected set %v", name, want)
		}
	}
}

// TestMoaiMCPTools_NoDuplicateNames asserts no two entries share a name (the
// single-declaration invariant — AP-C-4).
func TestMoaiMCPTools_NoDuplicateNames(t *testing.T) {
	seen := make(map[string]bool, 25)
	for _, tool := range MoaiMCPTools() {
		if seen[tool.Name] {
			t.Errorf("duplicate tool name %q in catalog (AP-C-4)", tool.Name)
		}
		seen[tool.Name] = true
	}
}

// TestMoaiMCPToolNames_MatchesCatalog asserts the convenience accessor returns
// the same identifiers in the same order.
func TestMoaiMCPToolNames_MatchesCatalog(t *testing.T) {
	tools := MoaiMCPTools()
	names := MoaiMCPToolNames()
	if len(names) != len(tools) {
		t.Fatalf("MoaiMCPToolNames len = %d, catalog len = %d", len(names), len(tools))
	}
	for i, n := range names {
		if n != tools[i].Name {
			t.Errorf("index %d: name %q != catalog %q", i, n, tools[i].Name)
		}
	}
}
