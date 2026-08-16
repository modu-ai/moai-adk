package mcp

import (
	"sort"
	"testing"
)

// TestMoaiMCPTools_Count21 asserts the catalog declares exactly 21 tools,
// matching the registration count in registerMoaiMCPTools (AC-C-001 /
// AC-C-002). A change here without a matching server-side change is the drift
// the guard test (internal/cli TestMoaiMCPServer_RegistrationMatchesCatalog)
// catches.
func TestMoaiMCPTools_Count21(t *testing.T) {
	tools := MoaiMCPTools()
	if len(tools) != 21 {
		t.Fatalf("catalog declares %d tools, want 21", len(tools))
	}
}

// TestMoaiMCPTools_SixWriteCapable asserts exactly the six write-capable
// tools carry WriteCapable=true (REQ-C-3 / AC-C-003), and the other 15 are
// read-only.
func TestMoaiMCPTools_SixWriteCapable(t *testing.T) {
	want := map[string]bool{
		"goal_arm":         true,
		"verify_snapshot":  true,
		"codex_task":       true,
		"codex_job_cancel": true,
		"glm_task":         true,
		"glm_job_cancel":   true,
	}
	var got []string
	for _, tool := range MoaiMCPTools() {
		if tool.WriteCapable {
			got = append(got, tool.Name)
		}
	}
	sort.Strings(got)
	if len(got) != 6 {
		t.Fatalf("write-capable tool count = %d (%v), want 6", len(got), got)
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
	seen := make(map[string]bool, 21)
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
