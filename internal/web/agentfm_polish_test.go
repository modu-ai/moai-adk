package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// Tests for the post-close agentfm polish rounds (SPEC-WEBCONF-SIMPLIFY-001):
// Round 1: namespace grouping → name font → actual-value selection.
// Round 2: sub-tabs (replacing group headers) → agent description per row.

// TestAgentFMSubTabs verifies the sub-tab nav renders with the two labels, both
// group panels exist in the DOM, and the default-active panel is "subagents".
func TestAgentFMSubTabs(t *testing.T) {
	root := t.TempDir()
	// Seed one moai-core + one harness agent.
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
	seedAgentFMFile(t, root, "harness", "hook-ci-specialist", "", "")
	body := renderAgentFMBody(t, root)

	// Sub-tab nav buttons with i18n keys.
	if !strings.Contains(body, `data-agentfm-tab="subagents"`) {
		t.Error(`missing Sub-agents sub-tab button (data-agentfm-tab="subagents")`)
	}
	if !strings.Contains(body, `data-agentfm-tab="harness"`) {
		t.Error(`missing Harness agents sub-tab button (data-agentfm-tab="harness")`)
	}
	if !strings.Contains(body, `data-i18n="agentfm.subtab.subagents"`) {
		t.Error(`missing i18n key agentfm.subtab.subagents on the sub-tab label`)
	}
	if !strings.Contains(body, `data-i18n="agentfm.subtab.harness"`) {
		t.Error(`missing i18n key agentfm.subtab.harness on the sub-tab label`)
	}

	// Both panels render (DOM-resident for atomic Save).
	if !strings.Contains(body, `data-agentfm-panel="subagents"`) {
		t.Error(`missing subagents panel container`)
	}
	if !strings.Contains(body, `data-agentfm-panel="harness"`) {
		t.Error(`missing harness panel container`)
	}

	// Default active: subagents panel is active, harness panel is NOT active.
	if !strings.Contains(body, `class="agentfm-panel is-active" data-agentfm-panel="subagents"`) {
		t.Error(`subagents panel should be the default-active panel (is-active class)`)
	}
	if strings.Contains(body, `data-agentfm-panel="harness"`) && strings.Contains(body, `agentfm-panel is-active" data-agentfm-panel="harness"`) {
		t.Error(`harness panel should NOT be active by default`)
	}

	// Agents are partitioned: moai core agent in the DOM, harness agent in the DOM.
	if !strings.Contains(body, `agentfm.manager-spec.model`) {
		t.Error(`moai core agent (manager-spec) row not rendered`)
	}
	if !strings.Contains(body, `agentfm.hook-ci-specialist.model`) {
		t.Error(`harness agent (hook-ci-specialist) row not rendered`)
	}

	// Old group headers should NOT render (replaced by sub-tabs).
	if strings.Contains(body, `agentfm.group.moai`) {
		t.Error(`old agentfm.group.moai header should be removed (replaced by sub-tabs)`)
	}
	if strings.Contains(body, `agentfm.group.harness`) {
		t.Error(`old agentfm.group.harness header should be removed (replaced by sub-tabs)`)
	}
}

// TestAgentFMActualValueSelection verifies the select shows the agent's ACTUAL
// current value as the selected option (not the "(keep current)" sentinel).
func TestAgentFMActualValueSelection(t *testing.T) {
	root := t.TempDir()
	// manager-spec: explicit effort=xhigh → xhigh should be the selected effort.
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
	body := renderAgentFMBody(t, root)

	if !strings.Contains(body, `<option value="xhigh" selected`) {
		t.Error(`explicit effort xhigh not shown as the selected effort option`)
	}
}

// TestAgentFMAbsentEffortShowsTierDefault verifies an absent-effort agent shows
// the tier-suggested default as the selected option + "(default)" annotation.
// hook-ci-specialist is 🩵 tier → suggested effort = low.
func TestAgentFMAbsentEffortShowsTierDefault(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "harness", "hook-ci-specialist", "", "")
	body := renderAgentFMBody(t, root)

	// 🩵 tier suggested effort = low → should be selected.
	if !strings.Contains(body, `<option value="low" selected`) {
		t.Error(`absent-effort agent (hook-ci-specialist, 🩵 tier) should show tier-suggested "low" as selected`)
	}
	// "(default)" annotation for the derived default.
	if !strings.Contains(body, `data-i18n="agentfm.default"`) {
		t.Error(`missing "(default)" annotation for the tier-suggested default`)
	}
}

// TestAgentFMDescriptionShown verifies the frontmatter description renders as a
// compact one-line summary in the agent row (post-close polish round 2).
func TestAgentFMDescriptionShown(t *testing.T) {
	root := t.TempDir()
	agentsDir := filepath.Join(root, ".claude", "agents", "moai")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fm := "---\n" +
		"model: opus\n" +
		"effort: xhigh\n" +
		"description: SPEC creation specialist for plan-phase artifact authoring.\n" +
		"---\n# manager-spec\n"
	if err := os.WriteFile(filepath.Join(agentsDir, "manager-spec.md"), []byte(fm), 0o644); err != nil {
		t.Fatal(err)
	}
	body := renderAgentFMBody(t, root)

	if !strings.Contains(body, `class="agentfm-desc"`) {
		t.Error(`missing agentfm-desc span for the description`)
	}
	if !strings.Contains(body, "SPEC creation specialist") {
		t.Error(`agent description text not rendered in the row`)
	}
}

// TestAgentFMDescriptionAbsentGraceful verifies an agent WITHOUT a description
// field renders gracefully (no empty desc span, row still functional — E7).
func TestAgentFMDescriptionAbsentGraceful(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
	body := renderAgentFMBody(t, root)

	// The row should still render (model select present).
	if !strings.Contains(body, `agentfm.manager-spec.model`) {
		t.Error(`agent row should render even without a description`)
	}
}

// renderAgentFMBody renders GET / with the given project root + returns the body.
func renderAgentFMBody(t *testing.T, root string) string {
	t.Helper()
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.readPreferences = func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, nil
	}
	a.writePreferences = func(string, profile.ProfilePreferences) error { return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	return rec.Body.String()
}
