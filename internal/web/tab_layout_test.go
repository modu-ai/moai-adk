package web

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// tab_layout_test.go — M2 guards for SPEC-WEB-CONSOLE-REDESIGN-001: the 9-tab
// restructure (AC-WCR-010..013). Tab placement is a RENDER concern; the
// persistence section of every moved field is unchanged (AP-4).

// wantTabOrder is the canonical tab order. It is asserted as a SEQUENCE, not
// a set — the tab nav and the tabpanel sequence must both follow it.
// SPEC-MCP-CONSOLE-001 M2 appends the mcp panel (10th tab).
var wantTabOrder = []string{
	"identity", "language", "launch", "llm", "workflow",
	"git-worktree", "audit", "agentfm", "report", "mcp",
}

// TestConsoleTabsOrder verifies AC-WCR-010.
func TestConsoleTabsOrder(t *testing.T) {
	tabs := consoleTabs()
	if len(tabs) != len(wantTabOrder) {
		got := make([]string, 0, len(tabs))
		for _, tb := range tabs {
			got = append(got, tb.ID)
		}
		t.Fatalf("consoleTabs() returned %d tabs %v, want %d %v", len(tabs), got, len(wantTabOrder), wantTabOrder)
	}
	for i, want := range wantTabOrder {
		if tabs[i].ID != want {
			t.Errorf("consoleTabs()[%d].ID = %q, want %q", i, tabs[i].ID, want)
		}
	}
}

// TestTabPanelRenderOrderMatchesTabs verifies the second half of AC-WCR-010:
// the rendered tabpanel sequence follows the tab sequence. A panel rendered out
// of order is not a display bug (JS drives visibility) but it breaks the
// contract that the nav and the DOM tell the same story.
func TestTabPanelRenderOrderMatchesTabs(t *testing.T) {
	html := renderConsolePage(t)
	var got []string
	rest := html
	for {
		i := strings.Index(rest, `data-panel="`)
		if i < 0 {
			break
		}
		rest = rest[i+len(`data-panel="`):]
		end := strings.Index(rest, `"`)
		if end < 0 {
			break
		}
		got = append(got, rest[:end])
		rest = rest[end:]
	}
	if len(got) != len(wantTabOrder) {
		t.Fatalf("rendered %d tabpanels %v, want %d %v", len(got), got, len(wantTabOrder), wantTabOrder)
	}
	for i, want := range wantTabOrder {
		if got[i] != want {
			t.Errorf("tabpanel[%d] = %q, want %q (panel order must match tab order)", i, got[i], want)
		}
	}
}

// panelHTML slices the rendered markup for one tabpanel: from its data-panel
// marker to the next one (or end of document for the last panel).
func panelHTML(t *testing.T, html, panel string) string {
	t.Helper()
	start := strings.Index(html, `data-panel="`+panel+`"`)
	if start < 0 {
		t.Fatalf("panel %q not found in rendered console", panel)
	}
	rest := html[start+len(`data-panel="`+panel+`"`):]
	if next := strings.Index(rest, `data-panel="`); next >= 0 {
		rest = rest[:next]
	}
	return rest
}

// assertPanelFields asserts every name renders inside panel and nowhere else.
func assertPanelFields(t *testing.T, html, panel string, names []string) {
	t.Helper()
	body := panelHTML(t, html, panel)
	for _, name := range names {
		marker := `name="` + name + `"`
		if !strings.Contains(body, marker) {
			t.Errorf("panel %q missing control %q", panel, name)
			continue
		}
		if strings.Count(html, marker) != strings.Count(body, marker) {
			t.Errorf("control %q also renders outside panel %q (a field must live on exactly one tab)", name, panel)
		}
	}
}

// TestWorkflowTabRetainsProseConsumedFields verifies AC-WCR-011: the workflow
// tab keeps the four prose-consumed key groups (F3). These have no Go caller
// but ARE read by the moai.md / loop.md skills, so they are NOT dead config.
func TestWorkflowTabRetainsProseConsumedFields(t *testing.T) {
	assertPanelFields(t, renderConsolePage(t), "workflow", []string{
		"workflow.execution_mode",
		"workflow.default_mode",
		"workflow.agentic_loop.max_iterations",
		"workflow.loop_prevention.failure_pattern_detection",
		"workflow.loop_prevention.max_iterations",
		"workflow.loop_prevention.max_retries_per_operation",
	})
}

// TestGitWorktreeTabFields verifies AC-WCR-012.
func TestGitWorktreeTabFields(t *testing.T) {
	assertPanelFields(t, renderConsolePage(t), "git-worktree", []string{
		"git_strategy.mode",
		"git_strategy.manual.merge_method",
		"git_strategy.personal.merge_method",
		"git_strategy.team.merge_method",
		"workflow.worktree.auto_create",
		"workflow.worktree.auto_merge",
		"workflow.worktree.auto_cleanup",
		"workflow.worktree.tmux_preferred",
		"workflow.branch_guard.enabled",
	})
}

// TestAuditTabFields verifies AC-WCR-013 — including the AP-4 half: moving the
// audit fields to their own TAB must not reclassify their persistence SECTION.
// A section change would silently reroute the write from workflow.yaml.
func TestAuditTabFields(t *testing.T) {
	auditNames := []string{
		"workflow.audit.model",
		"workflow.audit.gates.claude",
		"workflow.audit.gates.codex",
		"workflow.audit.gates.glm",
	}
	assertPanelFields(t, renderConsolePage(t), "audit", auditNames)

	byName := map[string]settings.FieldDef{}
	for _, f := range settings.AllFields() {
		byName[f.Name] = f
	}
	for _, name := range auditNames {
		f, ok := byName[name]
		if !ok {
			t.Errorf("audit field %q missing from the schema", name)
			continue
		}
		if f.Section != settings.SectionWorkflow {
			t.Errorf("%q Section = %q, want %q — a tab move must not reclassify the persistence section (AP-4)", name, f.Section, settings.SectionWorkflow)
		}
		if f.Persist.Kind != settings.PersistSeam || f.Persist.Section != "workflow" {
			t.Errorf("%q persist target = %+v, want seam/workflow (unchanged by the tab move)", name, f.Persist)
		}
	}
}

// TestEveryTabHasAPanel is the reachability guard that F1 (a section with
// FieldDefs but no render meta) failed: every tab id must have a matching
// tabpanel, and vice versa.
func TestEveryTabHasAPanel(t *testing.T) {
	html := renderConsolePage(t)
	for _, tb := range consoleTabs() {
		if !strings.Contains(html, `data-tab="`+tb.ID+`"`) {
			t.Errorf("tab %q has no nav button", tb.ID)
		}
		if !strings.Contains(html, `data-panel="`+tb.ID+`"`) {
			t.Errorf("tab %q has no matching tabpanel (an unreachable panel is invisible)", tb.ID)
		}
	}
}
