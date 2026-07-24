package web

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/v4manifest"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// Tests for the console UX fix batch (web console feedback round):
//
//	G1-1 brand badge renders the mascot without a fill/rounded box
//	G1-2 enum <option> labels stay English in every locale
//	G1-3 hint.effort.go_unbound wording + English server-side baseline
//	G1-4 "LLM" tab/section renamed to "3rd Party LLM"
//	G2-2 harness sub-tab removed (single-panel agentfm) + honest section count
//	G2-3 CUSTOM badge predicate no longer fires on the shipped `model: inherit`

// cssRuleBlock returns the declaration block of the FIRST rule whose selector
// text matches sel exactly (the text between "sel {" and the closing "}").
func cssRuleBlock(t *testing.T, css, sel string) string {
	t.Helper()
	start := strings.Index(css, sel+" {")
	if start < 0 {
		t.Fatalf("css rule %q not found", sel)
	}
	body := css[start+len(sel)+2:]
	end := strings.Index(body, "}")
	if end < 0 {
		t.Fatalf("css rule %q is unclosed", sel)
	}
	return body[:end]
}

// TestBrandBadgeHasNoFill verifies G1-1: the top-left nav brand badge renders the
// (alpha-transparent) mascot bare — no gradient fill, no shadow, no rounded box —
// while the .brand__badge element itself stays in the DOM as the flex anchor.
func TestBrandBadgeHasNoFill(t *testing.T) {
	css := readEmbeddedAsset(t, "console.css")

	badge := cssRuleBlock(t, css, ".brand__badge")
	for _, banned := range []string{"background", "box-shadow", "border-radius"} {
		if strings.Contains(badge, banned+":") {
			t.Errorf(".brand__badge still declares %q (the badge must render the mascot with no fill/rounded box):\n%s", banned, badge)
		}
	}
	// Layout anchoring is preserved (sizing + flex centering).
	for _, kept := range []string{"width", "height", "align-items", "justify-content"} {
		if !strings.Contains(badge, kept+":") {
			t.Errorf(".brand__badge lost layout declaration %q — the appbar layout must be unchanged:\n%s", kept, badge)
		}
	}

	mascot := cssRuleBlock(t, css, ".brand__badge img.brand__mascot")
	if strings.Contains(mascot, "border-radius:") {
		t.Errorf(".brand__badge img.brand__mascot still declares border-radius (the PNG is alpha-transparent):\n%s", mascot)
	}

	// The element itself must survive — this is a CSS-only change.
	body := renderIndexBody(t, profile.ProfilePreferences{})
	if !strings.Contains(body, `class="brand__badge"`) {
		t.Error("brand__badge element disappeared from the appbar (G1-1 is CSS-only)")
	}
}

// TestOptionLabelsStayEnglish verifies G1-2: applyI18n resolves keys containing
// ".opt." against the ENGLISH dictionary regardless of the active locale, so enum
// option tokens stay untranslated while titles/descriptions/tooltips localize.
func TestOptionLabelsStayEnglish(t *testing.T) {
	js := readEmbeddedAsset(t, "app.js")

	start := strings.Index(js, "function applyI18n")
	if start < 0 {
		t.Fatal("app.js does not define applyI18n")
	}
	searchFrom := start + len("function applyI18n")
	end := len(js)
	if next := strings.Index(js[searchFrom:], "\n  function "); next >= 0 {
		end = searchFrom + next
	}
	fn := js[start:end]

	if !strings.Contains(fn, `".opt."`) {
		t.Error(`applyI18n has no ".opt." guard — enum option labels would follow the active locale (G1-2)`)
	}
	if !strings.Contains(fn, "MOAI_I18N.en") {
		t.Error("applyI18n never resolves against the English dictionary (window.MOAI_I18N.en) for enum option keys")
	}
	// The guard must be key-scoped, NOT tag-scoped: placeholder options
	// (opt.project_default / opt.unset / opt.runtime_default) still translate.
	if strings.Contains(fn, `tagName === "OPTION"`) || strings.Contains(fn, "tagName == 'OPTION'") {
		t.Error(`applyI18n guards on tagName === "OPTION" — that would also freeze the translated placeholder options; use the ".opt." key substring`)
	}
	// The tooltip (data-i18n-title) pass stays locale-driven.
	if !strings.Contains(fn, "data-i18n-title") {
		t.Error("applyI18n lost the data-i18n-title pass")
	}
}

// TestEffortGoUnboundWording verifies G1-3: the mistranslated Korean "(Go 미독)"
// ("unread") is replaced in all 4 locales, and the templ server-side baseline
// renders the ENGLISH string (a no-JS render must not show Korean to everyone).
func TestEffortGoUnboundWording(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")

	for _, want := range []string{
		`"hint.effort.go_unbound": "(declarative — not read by the runtime)"`,
		`"hint.effort.go_unbound": "(런타임 미반영)"`,
		`"hint.effort.go_unbound": "（宣言用 — ランタイム未使用）"`,
		`"hint.effort.go_unbound": "（声明式 — 运行时不读取）"`,
	} {
		if !strings.Contains(dict, want) {
			t.Errorf("i18n.js missing corrected hint.effort.go_unbound entry: %s", want)
		}
	}
	for _, banned := range []string{"(Go 미독)", "(not Go-bound)", "（Go 未バインド）", "（Go 未绑定）"} {
		if strings.Contains(dict, banned) {
			t.Errorf("i18n.js still carries the old hint.effort.go_unbound value %q", banned)
		}
	}

	// The hint renders on the agentfm effort row, so the render needs an agent.
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
	body := renderAgentFMBody(t, root)
	if !strings.Contains(body, `data-i18n="hint.effort.go_unbound"`) {
		t.Fatal("the effort hint badge did not render — the baseline assertion below would be vacuous")
	}
	if strings.Contains(body, "(Go 미독)") {
		t.Error("the templ baseline still hardcodes the Korean (Go 미독) string — a no-JS render shows Korean to every locale")
	}
	if !strings.Contains(body, "(declarative — not read by the runtime)") {
		t.Error("the templ baseline does not render the English hint.effort.go_unbound text")
	}

	// Both templ call sites carry the English baseline (the schemaSelectRow branch
	// has no live `.effort` FieldDef today, so it is asserted at the source level).
	src := readGoSource(t, "fieldsets.templ")
	if strings.Contains(src, "(Go 미독)") {
		t.Error("fieldsets.templ still hardcodes the Korean (Go 미독) baseline")
	}
	if n := strings.Count(src, `data-i18n="hint.effort.go_unbound">(declarative — not read by the runtime)`); n != 2 {
		t.Errorf("fieldsets.templ has %d English hint.effort.go_unbound baselines, want 2 (schemaSelectRow + agentFMRow)", n)
	}
}

// TestLLMSectionRenamedThirdParty verifies G1-4: the tab strip label and the
// section title read "3rd Party LLM" (key sec.llm.title and tab id llm unchanged),
// and sec.llm.desc describes only the GLM backend tiers.
func TestLLMSectionRenamedThirdParty(t *testing.T) {
	for _, tab := range consoleTabs() {
		if tab.ID == "llm" {
			if tab.LabelKey != "sec.llm.title" {
				t.Errorf("llm tab LabelKey = %q, want sec.llm.title (key must not change)", tab.LabelKey)
			}
			if tab.Baseline != "3rd Party LLM" {
				t.Errorf("llm tab Baseline = %q, want %q", tab.Baseline, "3rd Party LLM")
			}
		}
	}
	for _, meta := range schemaSectionMetas() {
		if string(meta.ID) == "llm" && meta.Title != "3rd Party LLM" {
			t.Errorf("llm section Title = %q, want %q", meta.Title, "3rd Party LLM")
		}
	}

	dict := readEmbeddedAsset(t, "i18n.js")
	for _, want := range []string{
		`"sec.llm.title": "3rd Party LLM"`,
		`"sec.llm.title": "서드파티 LLM"`,
		`"sec.llm.title": "サードパーティ LLM"`,
		`"sec.llm.title": "第三方 LLM"`,
	} {
		if !strings.Contains(dict, want) {
			t.Errorf("i18n.js missing renamed section title entry: %s", want)
		}
	}
	// sec.llm.desc must no longer advertise the removed Claude tier mappings.
	descRe := regexp.MustCompile(`"sec\.llm\.desc": "([^"]*)"`)
	descs := descRe.FindAllStringSubmatch(dict, -1)
	if len(descs) != 4 {
		t.Fatalf("sec.llm.desc appears %d times in i18n.js, want 4 (one per locale)", len(descs))
	}
	for _, d := range descs {
		if strings.Contains(d[1], "Claude") {
			t.Errorf("sec.llm.desc still mentions Claude (the Claude tier mappings were removed): %q", d[1])
		}
	}
}

// TestAgentFMSinglePanel verifies G2-2: the "Harness agents" sub-tab is gone. Only
// the .claude/agents/moai/ rows render, the sub-tab/panel chrome is absent, and
// the section count reports the RENDERED agent count (not the full catalog).
func TestAgentFMSinglePanel(t *testing.T) {
	root := t.TempDir()
	// Two moai-core agents + one harness agent.
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "xhigh")
	seedAgentFMFile(t, root, "moai", "manager-docs", "sonnet", "medium")
	seedAgentFMFile(t, root, "harness", "hook-ci-specialist", "", "")
	body := renderAgentFMBody(t, root)

	// Sub-tab chrome removed entirely.
	for _, banned := range []string{
		`data-agentfm-tab=`,
		`data-agentfm-panel=`,
		`class="agentfm-subtabs"`,
		`data-i18n="agentfm.subtab.subagents"`,
		`data-i18n="agentfm.subtab.harness"`,
	} {
		if strings.Contains(body, banned) {
			t.Errorf("agentfm section still renders removed sub-tab markup %q", banned)
		}
	}

	// moai-core rows render; the harness agent does not.
	if !strings.Contains(body, `agentfm.manager-spec.model`) {
		t.Error("moai-core agent (manager-spec) row not rendered")
	}
	if strings.Contains(body, `agentfm.hook-ci-specialist.model`) {
		t.Error("harness agent (hook-ci-specialist) row still renders after the harness sub-tab removal")
	}

	// The count matches what actually renders (2 moai-core agents, not 3).
	if got := agentFMSectionCount(t, body); got != 2 {
		t.Errorf("agentfm section count = %d, want 2 (rendered moai-core agents only)", got)
	}
}

// agentFMSectionCount extracts the numeric section__count rendered next to the
// agentfm legend. sec.agentfm.title appears twice — first in the tab strip, then
// in the fieldset legend — so the LAST occurrence is the legend.
func agentFMSectionCount(t *testing.T, body string) int {
	t.Helper()
	legend := strings.LastIndex(body, `data-i18n="sec.agentfm.title"`)
	if legend < 0 {
		t.Fatal("agentfm legend not found in render")
	}
	if !strings.HasPrefix(body[legend:], `data-i18n="sec.agentfm.title">Sub-agent Frontmatter</span></legend>`) {
		t.Fatalf("expected the agentfm legend at the last sec.agentfm.title occurrence, got: %.80q", body[legend:])
	}
	rest := body[legend:]
	marker := `class="section__count">`
	at := strings.Index(rest, marker)
	if at < 0 {
		t.Fatal("agentfm section__count not found in render")
	}
	rest = rest[at+len(marker):]
	digits := regexp.MustCompile(`^\s*(\d+)`).FindStringSubmatch(rest)
	if digits == nil {
		t.Fatalf("agentfm section__count is not numeric: %.40q", rest)
	}
	n, err := strconv.Atoi(digits[1])
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// TestAgentTierBadgeInheritIsDefault verifies G2-3: `model: inherit` is the SHIPPED
// default for the core agents (SPEC-MODEL-PROFILE-MATRIX-001 stopped mutating agent
// frontmatter), so it must NOT be treated as a manual override. Only `effort: max`
// still marks a row CUSTOM.
func TestAgentTierBadgeInheritIsDefault(t *testing.T) {
	// inherit alone → the agent's tier glyph, not the CUSTOM pill.
	for _, name := range []string{"manager-spec", "manager-develop", "manager-docs"} {
		b := agentTierBadge(name, v4manifest.ModelInherit, v4manifest.EffortXhigh)
		if b.IsCustom {
			t.Errorf("%s (model=inherit): CUSTOM badge rendered for the SHIPPED default (G2-3)", name)
		}
		if !b.HasBadge {
			t.Errorf("%s (model=inherit): expected a tier badge", name)
		}
		tier, _ := v4manifest.AgentTier(name)
		if want := v4manifest.TierColor(tier); b.Glyph != want {
			t.Errorf("%s (model=inherit): glyph = %q, want the tier glyph %q", name, b.Glyph, want)
		}
	}

	// effort=max remains the sole manual-override signal.
	if b := agentTierBadge("manager-spec", v4manifest.ModelInherit, v4manifest.EffortMax); !b.IsCustom {
		t.Error("effort=max must still render the CUSTOM badge")
	}
	if b := agentTierBadge("manager-spec", v4manifest.ModelOpus, v4manifest.EffortMax); !b.IsCustom {
		t.Error("effort=max must still render the CUSTOM badge (with a concrete model pin)")
	}
}

// TestAgentTierBadgeGlyphConsistency verifies the G2-3 user-visible symptom is
// gone: with the shipped `model: inherit` frontmatter, EVERY tier-mapped core
// agent renders a tier glyph — no mixed CUSTOM/glyph rows.
func TestAgentTierBadgeGlyphConsistency(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"manager-spec", "manager-develop", "manager-docs", "manager-git"} {
		seedAgentFMFile(t, root, "moai", name, v4manifest.ModelInherit, "")
	}
	body := renderAgentFMBody(t, root)

	if strings.Contains(body, "agentfm-badge--custom") {
		t.Error("a CUSTOM badge renders for shipped `model: inherit` agents (G2-3 — inconsistent glyph/label)")
	}
	for _, glyph := range []string{"🔴", "🟠", "🔵"} {
		if !strings.Contains(body, glyph) {
			t.Errorf("tier glyph %q missing from the render — inherit agents must show their tier color", glyph)
		}
	}
}
