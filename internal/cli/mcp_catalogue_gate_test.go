// mcp_catalogue_gate_test.go — SPEC-MCP-NEUTRALITY-001 M2 (card t1632,
// REQ-MN-006): the catalogue is the hard source of the served surface.
// todo_claim is registered (internal/cli/mcp_todo.go) and MUST serve; and a
// registration naming a catalogue-absent tool fails the construction loudly —
// the enablement map is catalogue-seeded, so silence is exactly how
// todo_claim died (plan B4).
package cli

import (
	"testing"

	mcpcat "github.com/modu-ai/moai-adk/internal/mcp"
)

// TestMCPCatalogue_ServesTodoClaim pins the catalogue row: the registered
// todo_claim appears in tools/list (the wire surface the catalogue gates).
func TestMCPCatalogue_ServesTodoClaim(t *testing.T) {
	names := listToolNames(t)
	found := false
	for _, n := range names {
		if n == "todo_claim" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("todo_claim registered but absent from tools/list (the catalogue row is missing): %v", names)
	}
}

// TestMCPCatalogue_HasTool pins the lookup the registration gate runs.
func TestMCPCatalogue_HasTool(t *testing.T) {
	if !mcpcat.HasTool("todo_claim") {
		t.Error("HasTool(todo_claim) = false, want true (the catalogue row exists)")
	}
	if mcpcat.HasTool("totally_absent_tool") {
		t.Error("HasTool(totally_absent_tool) = true, want false")
	}
}

// TestMCPCatalogueGate_FailsClosedOnAbsentName pins the fail-closed
// registration gate: a name the catalogue does not carry must fail the
// construction loudly (panic), never serve silence.
func TestMCPCatalogueGate_FailsClosedOnAbsentName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("mcpCatalogueGate on a catalogue-absent name did not fail the construction")
		}
	}()
	mcpCatalogueGate("totally_absent_tool")
}

// TestMCPCatalogueGate_AdmitCataloguedName — the gate is transparent for a
// catalogue row.
func TestMCPCatalogueGate_AdmitCataloguedName(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("mcpCatalogueGate panicked on a catalogue row: %v", r)
		}
	}()
	mcpCatalogueGate("todo_claim")
}
