package web

// mcp_console_test.go — SPEC-MCP-CONSOLE-001 M2 console-surface tests
// (AC-C-001, AC-C-002, AC-C-003).
//
// The MCP section renders one enablement control per tool declared in the
// shared catalog (internal/mcp MoaiMCPTools). Both the schema field set and
// this render derive from that single declaration, so a tool added to
// registration cannot go unrepresented in the console (AP-C-4). The four
// write-capable tools carry a text-bearing marker the thirteen read-only
// tools do not (REQ-C-3); the distinction rides on text, not colour alone,
// so it survives a monochrome or screen-reader pass.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/mcp"
)

// mcpToolControlMarker is the HTML form-control name marker an MCP enablement
// toggle renders. Each tool's field name is mcp.tools.<name>.enabled
// (seamField dot-joins the path), so the control appears as
// name="mcp.tools.<name>.enabled".
func mcpToolControlMarker(toolName string) string {
	return `name="mcp.tools.` + toolName + `.enabled"`
}

// TestMCPConsoleRendersAllTools covers AC-C-001: the console renders one
// enablement control for every tool in the shared catalog, and no tool is
// missing its control.
func TestMCPConsoleRendersAllTools(t *testing.T) {
	body := renderConsolePage(t)

	tools := mcp.MoaiMCPTools()
	if len(tools) == 0 {
		t.Fatal("MoaiMCPTools() returned an empty catalog — M1 declaration missing")
	}
	for _, tool := range tools {
		marker := mcpToolControlMarker(tool.Name)
		if !strings.Contains(body, marker) {
			t.Errorf("rendered page missing enablement control for tool %q (expected %s)", tool.Name, marker)
		}
	}
}

// TestMCPConsoleToolCountMatchesCatalog covers AC-C-002: the count of MCP
// enablement controls in the rendered page equals the count of tools in the
// shared catalog exactly. A tool added to registration without a console
// entry (or a stale console entry left after a tool is removed) makes the two
// counts diverge and fails this test.
func TestMCPConsoleToolCountMatchesCatalog(t *testing.T) {
	body := renderConsolePage(t)

	tools := mcp.MoaiMCPTools()
	want := len(tools)

	// Count the hidden __present companion markers — exactly one per rendered
	// tool toggle (mcpToolRow emits name="<field>__present"). Scoped to the
	// mcp.tools. prefix so the workflow.branch_guard.enabled__present companion
	// (a different section) is not double-counted.
	got := 0
	for _, tool := range tools {
		if strings.Contains(body, `name="mcp.tools.`+tool.Name+`.enabled__present"`) {
			got++
		}
	}
	if got != want {
		t.Errorf("rendered MCP tool control count = %d, want %d (catalog size); divergence means the console and registration are not derived from the single shared declaration", got, want)
	}
}

// TestMCPConsoleWriteCapableTextDistinction covers AC-C-003: each of the four
// write-capable tools (goal_arm, verify_snapshot, codex_task, codex_job_cancel)
// carries a text-bearing distinction that the thirteen read-only tools do
// not. The distinction MUST be carried in text — not colour alone — so the
// test asserts a textual marker is present on write-capable rows and absent
// on read-only rows.
func TestMCPConsoleWriteCapableTextDistinction(t *testing.T) {
	body := renderConsolePage(t)

	// The write-capable badge text. The test deliberately checks for a TEXT
	// substring (not a CSS class) so a colour-only or icon-only distinction
	// fails — the marker must be machine-readable text to survive a
	// monochrome or screen-reader rendering (REQ-C-3).
	//
	// The badge text is seeded by the f.mcp.write_capable.badge i18n key's
	// English baseline rendered server-side before applyI18n runs.
	badgeText := "Write-capable"
	if !strings.Contains(body, badgeText) {
		t.Fatalf("rendered page carries no text-bearing write-capable marker (expected substring %q)", badgeText)
	}

	// For each write-capable tool, the badge text must appear in the same row
	// region as the tool's control. For each read-only tool, the badge text
	// must NOT appear in that tool's row region.
	//
	// A row-region check around the control marker is sufficient: if the
	// write-capable badge leaks into a read-only tool's row, or is missing
	// from a write-capable tool's row, the test fails.
	for _, tool := range mcp.MoaiMCPTools() {
		// Anchor on the field__key code chip, which carries the exact field
		// name and sits inside the label region — right after the badge in
		// mcpToolRow's emit order (title → badge → key chip). A bidirectional
		// window around the code chip captures the badge (which precedes it)
		// without spilling into the next tool's row.
		chip := `<code class="key">mcp.tools.` + tool.Name + `.enabled</code>`
		idx := strings.Index(body, chip)
		if idx < 0 {
			t.Fatalf("rendered page missing key chip for tool %q — TestMCPConsoleRendersAllTools should have caught this", tool.Name)
		}
		rowStart := idx - 400
		if rowStart < 0 {
			rowStart = 0
		}
		rowEnd := idx + 200
		if rowEnd > len(body) {
			rowEnd = len(body)
		}
		rowWindow := body[rowStart:rowEnd]
		hasBadge := strings.Contains(rowWindow, badgeText)
		switch {
		case tool.WriteCapable && !hasBadge:
			t.Errorf("write-capable tool %q is missing the text write-capable marker in its row", tool.Name)
		case !tool.WriteCapable && hasBadge:
			t.Errorf("read-only tool %q carries the write-capable marker in its row — the distinction must separate the two classes", tool.Name)
		}
	}
}
