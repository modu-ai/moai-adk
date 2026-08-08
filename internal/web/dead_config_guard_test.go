package web

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// dead_config_guard_test.go — M1 guards for SPEC-WEB-CONSOLE-REDESIGN-001:
// the 7 dead-config fields are absent from the edit schema (AC-WCR-001/002) and
// the git_strategy render surface is restored (AC-WCR-004).

// deadConfigFieldNames is the exact set of field names removed from the web edit
// surface in M1. The yaml keys, the internal/config struct members, and the
// accessors are PRESERVED — only the FieldDef entries are gone (REQ-WCR-003; the
// preservation half is pinned by internal/config.TestWorkflowBackwardCompat).
var deadConfigFieldNames = []string{
	"workflow.token_budget.plan",
	"workflow.token_budget.run",
	"workflow.token_budget.sync",
	"workflow.auto_clear.enabled",
	"workflow.auto_clear.after_plan",
	"workflow.auto_clear.after_run",
	"workflow.auto_clear.token_threshold",
}

// TestDeadConfigAbsentFromSchema verifies AC-WCR-001 + AC-WCR-002: none of the
// 7 removed names appears in settings.AllFields(), so the generic form parser
// (parseSchemaForm iterates AllFields) can never collect them as edit targets.
func TestDeadConfigAbsentFromSchema(t *testing.T) {
	present := map[string]bool{}
	for _, f := range settings.AllFields() {
		present[f.Name] = true
	}
	for _, name := range deadConfigFieldNames {
		if present[name] {
			t.Errorf("dead-config field %q is still registered in settings.AllFields() — the web edit surface must not expose it", name)
		}
	}
}

// TestDeadConfigAbsentFromRender is the render-layer half: no form control
// carrying a removed name reaches the HTML.
func TestDeadConfigAbsentFromRender(t *testing.T) {
	html := renderConsolePage(t)
	for _, name := range deadConfigFieldNames {
		if strings.Contains(html, `name="`+name+`"`) {
			t.Errorf("rendered console still carries a form control for the removed field %q", name)
		}
	}
}

// TestDeadConfigLiveNeighborsPreserved is the AP-3 guard: the fields ADJACENT to
// the removed ones in seamSectionFields() have prose consumers (F3) and must
// survive the deletion. A sweep that removed `auto_clear` and took
// `agentic_loop` / `loop_prevention` with it is the most plausible failure of
// this milestone.
func TestDeadConfigLiveNeighborsPreserved(t *testing.T) {
	present := map[string]bool{}
	for _, f := range settings.AllFields() {
		present[f.Name] = true
	}
	for _, name := range []string{
		"workflow.default_mode",
		"workflow.execution_mode",
		"workflow.agentic_loop.max_iterations",
		"workflow.loop_prevention.failure_pattern_detection",
		"workflow.loop_prevention.max_iterations",
		"workflow.loop_prevention.max_retries_per_operation",
		"workflow.worktree.auto_create",
		"workflow.worktree.auto_merge",
		"workflow.worktree.auto_cleanup",
		"workflow.worktree.tmux_preferred",
		"workflow.branch_guard.enabled",
	} {
		if !present[name] {
			t.Errorf("prose-consumed field %q was removed alongside the dead config (AP-3 violation)", name)
		}
	}
}

// TestGitStrategyRendered verifies AC-WCR-004: git_strategy.mode and the three
// per-profile merge_method controls render in the console HTML. Before M1 the
// FieldDefs existed but no render meta referenced them, so the section had no
// UI at all (F1).
func TestGitStrategyRendered(t *testing.T) {
	html := renderConsolePage(t)
	want := []string{
		`name="git_strategy.mode"`,
		`name="git_strategy.manual.merge_method"`,
		`name="git_strategy.personal.merge_method"`,
		`name="git_strategy.team.merge_method"`,
	}
	for _, marker := range want {
		if !strings.Contains(html, marker) {
			t.Errorf("rendered console missing git_strategy control %q (render surface not restored)", marker)
		}
	}
}

// TestGitWorktreePanelRegistered verifies the git-worktree panel is reachable:
// a tab button and a matching tabpanel both carry the git-worktree id. A panel
// without a tab is invisible (CSS hides every non-active panel), which is the
// exact F1 shape this SPEC repairs.
func TestGitWorktreePanelRegistered(t *testing.T) {
	html := renderConsolePage(t)
	for _, marker := range []string{
		`data-tab="git-worktree"`,
		`data-panel="git-worktree"`,
	} {
		if !strings.Contains(html, marker) {
			t.Errorf("rendered console missing %q", marker)
		}
	}
}
