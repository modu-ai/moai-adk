package web

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// TestConsoleTabsIncludesReport asserts the 7th 'report' tab is registered for
// the report config section and that the pre-existing six tabs remain intact
// and in order. report.format was relocated here off the launch tab.
func TestConsoleTabsIncludesReport(t *testing.T) {
	tabs := consoleTabs()
	wantOrder := []string{"identity", "language", "launch", "llm", "agentfm", "report"}
	if len(tabs) != len(wantOrder) {
		t.Fatalf("consoleTabs() returned %d tabs, want %d", len(tabs), len(wantOrder))
	}
	for i, want := range wantOrder {
		if tabs[i].ID != want {
			t.Errorf("consoleTabs()[%d].ID = %q, want %q", i, tabs[i].ID, want)
		}
	}
	var report *consoleTab
	for i := range tabs {
		if tabs[i].ID == "report" {
			report = &tabs[i]
			break
		}
	}
	if report == nil {
		t.Fatal("consoleTabs() missing report entry")
	}
	if report.LabelKey != "sec.report.title" {
		t.Errorf("report LabelKey = %q, want sec.report.title", report.LabelKey)
	}
	if report.Baseline != "Report" {
		t.Errorf("report Baseline = %q, want Report", report.Baseline)
	}
}

// TestSchemaSectionMetasIncludesReport asserts the report section renders via the
// generic fieldsetSchemaSection path (root.templ loops schemaSectionMetas()).
func TestSchemaSectionMetasIncludesReport(t *testing.T) {
	metas := schemaSectionMetas()
	var found bool
	for _, m := range metas {
		if m.ID == settings.SectionReport {
			found = true
			if m.Title == "" || m.Desc == "" {
				t.Errorf("report section meta has empty Title/Desc: %+v", m)
			}
			break
		}
	}
	if !found {
		t.Fatalf("schemaSectionMetas() missing SectionReport: %+v", metas)
	}
}

// TestConsoleRendersReportTab is a boundary check across the templ render layer:
// the rendered console HTML contains the report tab button + tabpanel + the
// report.format select, while the launch panel no longer duplicates
// report.format (de-duplication after the field moved to its own tab).
func TestConsoleRendersReportTab(t *testing.T) {
	html := renderConsolePage(t)
	if !strings.Contains(html, `data-tab="report"`) {
		t.Error(`rendered console missing report tab button (data-tab="report")`)
	}
	if !strings.Contains(html, `data-panel="report"`) {
		t.Error(`rendered console missing report tabpanel (data-panel="report")`)
	}
	if !strings.Contains(html, `name="report.format"`) {
		t.Error("rendered console missing report.format select input")
	}
	// De-duplication boundary: report.format must NOT render inside the launch panel.
	// SPEC-DESIGN-MOAIWEBV2-001 M1: the orphan `project` panel was removed, so the
	// panel rendered immediately after `launch` is now `llm` (first schemaSectionMeta).
	launchStart := strings.Index(html, `data-panel="launch"`)
	nextStart := strings.Index(html, `data-panel="llm"`)
	if launchStart < 0 || nextStart < 0 || nextStart <= launchStart {
		t.Fatal("could not locate launch/llm panel boundaries in rendered HTML")
	}
	launchPanel := html[launchStart:nextStart]
	if strings.Contains(launchPanel, `report.format`) {
		t.Error("launch panel still renders report.format (de-duplication failed)")
	}
	// SPEC-DESIGN-MOAIWEBV2-001 M1: no orphan `project` panel remains in the render.
	if strings.Contains(html, `data-panel="project"`) {
		t.Error(`orphan project panel still rendered (data-panel="project" should be removed)`)
	}
}
