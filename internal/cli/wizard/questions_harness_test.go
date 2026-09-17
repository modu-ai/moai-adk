package wizard

// SPEC-INIT-HARNESS-001 M3 — the agent_wiring question describes DEPLOYMENT
// consequences (REQ-IH-013, AC-IH-012), synced across ko/ja/zh. Option
// values (claude/codex/both) are FROZEN — only Label/Desc/Title change.

import (
	"strings"
	"testing"
)

// deploymentKeywords maps each frozen option value to the deploy-path tokens
// its description must carry in every locale. Paths and identifiers stay
// verbatim across translations, so they are the stable assertion surface.
var deploymentKeywords = map[string][]string{
	"claude": {".claude", "AGENTS.md"},
	"codex":  {"AGENTS.md", ".claude"},
	"both":   {".claude", ".codex"},
}

// TestAgentWiringOptions verifies the question shape: exactly three options,
// the frozen value set in order, a deployment-describing default, and every
// description naming the deployment consequences of its value (AC-IH-012).
func TestAgentWiringOptions(t *testing.T) {
	qs := Page3Questions(t.TempDir())
	var q *Question
	for i := range qs {
		if qs[i].ID == "agent_wiring" {
			q = &qs[i]
			break
		}
	}
	if q == nil {
		t.Fatal("agent_wiring question not found in Page3Questions")
	}
	if len(q.Options) != 3 {
		t.Fatalf("agent_wiring option count = %d, want 3 (frozen 3-option set)", len(q.Options))
	}
	wantValues := []string{"claude", "codex", "both"}
	for i, want := range wantValues {
		if q.Options[i].Value != want {
			t.Errorf("option[%d].Value = %q, want %q (frozen order)", i, q.Options[i].Value, want)
		}
	}

	// The default option description must describe DEPLOYMENT, not just
	// wiring: "deployment" consequence wording per design.md D1.
	for _, opt := range q.Options {
		for _, kw := range deploymentKeywords[opt.Value] {
			if !strings.Contains(opt.Desc, kw) {
				t.Errorf("option %s description %q does not describe deployment: missing %q", opt.Value, opt.Desc, kw)
			}
		}
	}

	// codex description must state the claude-only surfaces are NOT deployed.
	if !strings.Contains(q.Options[1].Desc, "no .claude") && !strings.Contains(q.Options[1].Desc, "No .claude") {
		t.Errorf("codex option description %q must state no .claude/ tree is deployed", q.Options[1].Desc)
	}
}

// TestAgentWiringTranslationsSync verifies the ko/ja/zh blocks carry the same
// three frozen values and the same deployment-consequence path tokens
// (REQ-IH-013 cross-locale sync).
func TestAgentWiringTranslationsSync(t *testing.T) {
	for _, lang := range []string{"ko", "ja", "zh"} {
		trs, ok := translations[lang]
		if !ok {
			t.Fatalf("%s translation block missing", lang)
		}
		tr, ok := trs["agent_wiring"]
		if !ok {
			t.Fatalf("%s: agent_wiring translation missing", lang)
		}
		if len(tr.Options) != 3 {
			t.Errorf("%s: option count = %d, want 3", lang, len(tr.Options))
			continue
		}
		for i, want := range []string{"claude", "codex", "both"} {
			got := tr.Options[i]
			// Label order mirrors the frozen value order; labels name the
			// harness either by the English token or a localized form, so the
			// value-order assertion rides the INDEX, and the deployment
			// keywords ride the Desc.
			for _, kw := range deploymentKeywords[want] {
				if !strings.Contains(got.Desc, kw) {
					t.Errorf("%s: option[%d] (%s) description %q missing deployment token %q", lang, i, want, got.Desc, kw)
				}
			}
		}
		// codex (index 1) must state the absence of the .claude tree.
		if !strings.Contains(tr.Options[1].Desc, ".claude") {
			t.Errorf("%s: codex description %q must reference the withheld .claude/ tree", lang, tr.Options[1].Desc)
		}
	}
}
