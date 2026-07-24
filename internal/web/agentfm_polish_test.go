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
//
// The web console UX fix batch (G2-2) then RETIRED the sub-tabs: only the
// .claude/agents/moai/ rows render, in a single flat panel.

// TestAgentFMNoSubTabs verifies the sub-tab chrome (and the harness rows behind
// it) no longer render — neither the buttons, the panels, nor the old group
// headers they had replaced.
func TestAgentFMNoSubTabs(t *testing.T) {
	root := t.TempDir()
	// Seed one moai-core + one harness agent.
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
	seedAgentFMFile(t, root, "harness", "hook-ci-specialist", "", "")
	body := renderAgentFMBody(t, root)

	// Sub-tab nav buttons + panels are gone.
	for _, banned := range []string{
		`data-agentfm-tab="subagents"`,
		`data-agentfm-tab="harness"`,
		`data-i18n="agentfm.subtab.subagents"`,
		`data-i18n="agentfm.subtab.harness"`,
		`data-agentfm-panel="subagents"`,
		`data-agentfm-panel="harness"`,
		`class="agentfm-subtabs"`,
	} {
		if strings.Contains(body, banned) {
			t.Errorf(`sub-tab markup %q must be removed (G2-2 single-panel agentfm)`, banned)
		}
	}

	// The moai core agent renders directly; the harness agent does not render.
	if !strings.Contains(body, `agentfm.manager-spec.model`) {
		t.Error(`moai core agent (manager-spec) row not rendered`)
	}
	if strings.Contains(body, `agentfm.hook-ci-specialist.model`) {
		t.Error(`harness agent (hook-ci-specialist) row must NOT render (G2-2)`)
	}

	// Old group headers stay removed.
	if strings.Contains(body, `agentfm.group.moai`) {
		t.Error(`old agentfm.group.moai header should be removed`)
	}
	if strings.Contains(body, `agentfm.group.harness`) {
		t.Error(`old agentfm.group.harness header should be removed`)
	}
}

// TestAgentFMActualValueSelection verifies the select shows the agent's ACTUAL
// current value as the selected option — post-G3 the "actual" value is the
// profile-matrix resolution, so an llm.agent_overrides pin is the selected option.
func TestAgentFMActualValueSelection(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh") // frontmatter ignored
	// An override pins manager-spec to sonnet/xhigh → those must be selected.
	writeLLMProfileYAML(t, root, "medium", map[string][2]string{"manager-spec": {"sonnet", "xhigh"}})
	body := renderAgentFMBody(t, root)

	if !strings.Contains(body, `<option value="sonnet" selected`) {
		t.Error(`override model (sonnet) not shown as the selected model option`)
	}
	if !strings.Contains(body, `<option value="xhigh" selected`) {
		t.Error(`override effort (xhigh) not shown as the selected effort option`)
	}
}

// TestAgentFMDefaultShowsProfileMatrixValue verifies an agent with no override
// shows the PROFILE-MATRIX default as the selected option + "(default)" annotation.
// manager-develop under the medium profile (develop group) resolves to opus/xhigh
// (defaultProfileMatrix). (A moai-core agent is used because the harness rows no
// longer render — G2-2.)
func TestAgentFMDefaultShowsProfileMatrixValue(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-develop", "", "")
	writeLLMProfileYAML(t, root, "medium", nil)
	body := renderAgentFMBody(t, root)

	// medium/develop → opus/xhigh (profile matrix), NOT the badge-tier "high".
	if !strings.Contains(body, `<option value="xhigh" selected`) {
		t.Error(`no-override agent (manager-develop) should show the medium-profile develop cell (xhigh) as selected`)
	}
	if strings.Contains(body, `<option value="high" selected`) {
		t.Error(`manager-develop shows the badge-tier "high" — read path must derive from the profile matrix (xhigh under medium)`)
	}
	// "(default)" annotation for the profile-derived (non-override) value.
	if !strings.Contains(body, `data-i18n="agentfm.default"`) {
		t.Error(`missing "(default)" annotation for the profile-derived default`)
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

// TestAgentFMTierSortOrder verifies agents render in tier order (most expensive
// first: 🔴 → 🟠 → 🔵 → 🩵), with alphabetical secondary within the same tier
// (post-close polish round 3).
func TestAgentFMTierSortOrder(t *testing.T) {
	root := t.TempDir()
	// Seed one agent from each tier present in .claude/agents/moai/.
	// 🔴 manager-spec, 🔴 plan-auditor, 🟠 manager-develop, 🔵 manager-docs.
	seedAgentFMFile(t, root, "moai", "manager-spec", "", "")
	seedAgentFMFile(t, root, "moai", "plan-auditor", "", "")
	seedAgentFMFile(t, root, "moai", "manager-develop", "", "")
	seedAgentFMFile(t, root, "moai", "manager-docs", "", "")
	body := renderAgentFMBody(t, root)

	// Red agents before orange before blue.
	// Use the form field name anchor (agentfm.<name>.model) to find document positions.
	specPos := strings.Index(body, `agentfm.manager-spec.model`)
	auditorPos := strings.Index(body, `agentfm.plan-auditor.model`)
	devPos := strings.Index(body, `agentfm.manager-develop.model`)
	docsPos := strings.Index(body, `agentfm.manager-docs.model`)
	if specPos < 0 || auditorPos < 0 || devPos < 0 || docsPos < 0 {
		t.Fatalf(`agent row anchors not found in render (spec=%d auditor=%d dev=%d docs=%d)`, specPos, auditorPos, devPos, docsPos)
	}
	// Red (manager-spec, plan-auditor) must come before orange (manager-develop).
	if devPos < specPos || devPos < auditorPos {
		t.Errorf(`orange agent (manager-develop pos=%d) must come AFTER red agents (manager-spec pos=%d, plan-auditor pos=%d)`, devPos, specPos, auditorPos)
	}
	// Orange (manager-develop) must come before blue (manager-docs).
	if docsPos < devPos {
		t.Errorf(`blue agent (manager-docs pos=%d) must come AFTER orange agent (manager-develop pos=%d)`, docsPos, devPos)
	}
	// Alphabetical within same tier (red): manager-spec before plan-auditor.
	if auditorPos < specPos {
		t.Errorf(`within red tier, manager-spec (pos=%d) should come before plan-auditor (pos=%d) alphabetically`, specPos, auditorPos)
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
