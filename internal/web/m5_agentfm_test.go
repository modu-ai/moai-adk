package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/v4manifest"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// Tests for SPEC-WEBCONF-SIMPLIFY-001 M5: agentfm tier-badge UI (REQ-WC-006/007/008,
// design.md §E). The badge is display-only (Option A — name-keyed table, NOT the
// agent's effort file). M1 supplied TierForAgent/TierColor/TierSuggestedModelEffort;
// M5 wires them into the agentFMRow render.

// TestM5AgentTierBadgeAll20Agents verifies AC-WC-005: every catalog agent gets a
// non-custom tier badge from the name-keyed lookup table, with the distribution
// 🔴×4 / 🟠×4 / 🔵×5 / 🩵×7 (design.md §C).
func TestM5AgentTierBadgeAll20Agents(t *testing.T) {
	counts := map[string]int{}
	total := 0
	for name := range v4manifest.AllAgentTiers() {
		total++
		b := agentTierBadge(name, "", "") // empty model/effort = no override sentinel
		if !b.HasBadge {
			t.Errorf("agent %q: expected a tier badge, got none (EC-6 unmapped?)", name)
			continue
		}
		if b.IsCustom {
			t.Errorf("agent %q: unexpected custom badge (no max/inherit override set)", name)
		}
		counts[b.Glyph]++
	}
	if total != 20 {
		t.Errorf("catalog agent count = %d, want 20", total)
	}
	for glyph, want := range map[string]int{"🔴": 4, "🟠": 4, "🔵": 5, "🩵": 7} {
		if got := counts[glyph]; got != want {
			t.Errorf("tier glyph %q: %d agents, want %d", glyph, got, want)
		}
	}
}

// TestM5AgentTierBadgeCustomOverride verifies AC-WC-018 / EC-2: when the agent's
// current effort is `max`, the badge renders a neutral "custom" marker (NOT a
// tier color).
//
// `model: inherit` was REMOVED from the custom predicate by the console UX fix
// batch (G2-3): SPEC-MODEL-PROFILE-MATRIX-001 stopped mutating agent frontmatter,
// making `inherit` the SHIPPED default rather than a manual override — see
// TestAgentTierBadgeInheritIsDefault. `effort: max` is now the sole signal.
func TestM5AgentTierBadgeCustomOverride(t *testing.T) {
	cases := []struct {
		name, model, effort string
	}{
		{"manager-spec", v4manifest.ModelOpus, v4manifest.EffortMax},
		{"hns-github-specialist", v4manifest.ModelInherit, v4manifest.EffortMax},
	}
	for _, tc := range cases {
		b := agentTierBadge(tc.name, tc.model, tc.effort)
		if !b.IsCustom {
			t.Errorf("%s (model=%s effort=%s): expected custom badge, got glyph=%q isCustom=%v", tc.name, tc.model, tc.effort, b.Glyph, b.IsCustom)
		}
		if b.Glyph != "custom" {
			t.Errorf("%s: custom badge glyph = %q, want \"custom\"", tc.name, b.Glyph)
		}
		if b.TooltipKey != "fieldDesc.agentfm.custom" {
			t.Errorf("%s: custom tooltip key = %q, want fieldDesc.agentfm.custom", tc.name, b.TooltipKey)
		}
	}
}

// NOTE: TestM5AgentTierSuggestedMarkers was removed by the G3 profile-matrix
// repoint. The badge-tier "suggested model/effort" helpers (agentIsSuggestedModel /
// agentIsSuggestedEffort) that it exercised are retired — the agentfm read path now
// derives each agent's selected value from the runtime profile matrix
// (template.ResolveAgentModelEffort), not the badge-tier table. Profile-matrix
// resolution is covered by TestG3ReadPathDerivesFromProfileMatrix.

// seedAgentFMFile writes a minimal agent .md frontmatter file for a render test.
func seedAgentFMFile(t *testing.T, root, dir, name, model, effort string) {
	t.Helper()
	agentsDir := filepath.Join(root, ".claude", "agents", dir)
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fm := "---\nmodel: " + model + "\n"
	if effort != "" {
		fm += "effort: " + effort + "\n"
	}
	fm += "---\n# " + name + "\n"
	if err := os.WriteFile(filepath.Join(agentsDir, name+".md"), []byte(fm), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestM5AgentFMRenderBadgeAndSelects verifies the rendered agentfm row carries the
// tier badge + model/effort selects (AC-WC-005/007/008). Seeds a 🔴-tier agent
// (manager-spec) and asserts the badge glyph + the two selects render.
func TestM5AgentFMRenderBadgeAndSelects(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
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
	body := rec.Body.String()

	// Tier badge renders (🔴 for manager-spec — display-only, from the name table).
	if !strings.Contains(body, `class="agentfm-badge"`) {
		t.Error(`rendered page missing .agentfm-badge element (REQ-WC-006 tier badge)`)
	}
	if !strings.Contains(body, "🔴") {
		t.Error(`rendered page missing the 🔴 tier glyph for manager-spec`)
	}
	// Model + effort selects present + individually overrideable (REQ-WC-008).
	if !strings.Contains(body, `name="agentfm.manager-spec.model"`) {
		t.Error(`rendered page missing the model select for manager-spec (REQ-WC-008)`)
	}
	if !strings.Contains(body, `name="agentfm.manager-spec.effort"`) {
		t.Error(`rendered page missing the effort select for manager-spec (REQ-WC-008)`)
	}
	// Per-option data-i18n-title on the model options (M4 description mechanism, REQ-WC-015).
	if !strings.Contains(body, `data-i18n-title="fieldDesc.agentfm.model.opus"`) {
		t.Error(`rendered page missing per-option data-i18n-title for the opus model option`)
	}
}
