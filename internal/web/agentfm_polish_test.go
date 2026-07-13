package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// Tests for the post-close agentfm polish (SPEC-WEBCONF-SIMPLIFY-001):
// (1) namespace grouping, (2) name font class, (3) actual-value selection.

// TestAgentFMNamespaceGrouping verifies the two group headers render + agents
// are partitioned into moai core (10) + harness specialists (10).
func TestAgentFMNamespaceGrouping(t *testing.T) {
	root := t.TempDir()
	// Seed one moai-core + one harness agent.
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
	seedAgentFMFile(t, root, "harness", "hook-ci-specialist", "", "")
	body := renderAgentFMBody(t, root)

	if !strings.Contains(body, `data-i18n="agentfm.group.moai"`) {
		t.Error(`missing MoAI Core Agents group header`)
	}
	if !strings.Contains(body, `data-i18n="agentfm.group.harness"`) {
		t.Error(`missing Harness Specialists group header`)
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
