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
// 재설계본 CSS 는 선택자와 여는 중괄호 사이에 공백을 두지 않는다. 두 형태를
// 모두 받아들여, 포맷 차이 때문에 규칙을 "없다"고 잘못 말하지 않게 한다.
func cssRuleBlock(t *testing.T, css, sel string) string {
	t.Helper()
	start := strings.Index(css, sel+"{")
	open := len(sel) + 1
	if start < 0 {
		start = strings.Index(css, sel+" {")
		open = len(sel) + 2
	}
	if start < 0 {
		t.Fatalf("css rule %q not found", sel)
	}
	body := css[start+open:]
	end := strings.Index(body, "}")
	if end < 0 {
		t.Fatalf("css rule %q is unclosed", sel)
	}
	return body[:end]
}

// TestBrandBadgeHasNoFill verifies G1-1: the top-left nav brand badge renders the
// (alpha-transparent) mascot bare — no gradient fill, no shadow, no rounded box —
// while the .brand__badge element itself stays in the DOM as the flex anchor.
// 재설계에서 브랜드 자리는 마스코트 PNG 배지에서 인라인 SVG 로고 마크로 바뀌었다.
// 지켜야 할 성질은 그대로다: 채움도 그림자도 둥근 상자도 없이 마크만 놓인다.
func TestBrandMarkHasNoFill(t *testing.T) {
	css := readEmbeddedAsset(t, "console.css")

	logo := cssRuleBlock(t, css, ".rail__logo")
	for _, banned := range []string{"background", "box-shadow", "border-radius"} {
		if strings.Contains(logo, banned+":") {
			t.Errorf(".rail__logo declares %q — the brand mark carries no fill/rounded box:\n%s", banned, logo)
		}
	}
	// 마크는 무채색 체계를 따른다 — 원본 브랜드 녹색이 화면에 나오지 않는다.
	if !strings.Contains(logo, "grayscale") {
		t.Errorf(".rail__logo lost its grayscale filter — the brand green would break the achromatic system:\n%s", logo)
	}

	body := renderIndexBody(t, profile.ProfilePreferences{})
	if !strings.Contains(body, `class="rail__logo"`) {
		t.Error("the brand mark disappeared from the rail")
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

// TestEffortGoUnboundWording verifies G3-6: the stale "(declarative — not read by
// the runtime)" caption is reworded in all 4 locales (post-G3-1 the per-agent
// model/effort IS runtime-bound via the profile matrix, so the old caption is
// misleading), and the templ server-side baseline renders the new ENGLISH string.
func TestEffortGoUnboundWording(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")

	for _, want := range []string{
		`"hint.effort.go_unbound": "Resolved from the performance tier above — per-agent edits save as overrides."`,
		`"hint.effort.go_unbound": "위의 성능 티어에서 결정됩니다. 개별 편집은 override로 저장됩니다."`,
		`"hint.effort.go_unbound": "上のパフォーマンスティアで決まります。個別の編集はオーバーライドとして保存されます。"`,
		`"hint.effort.go_unbound": "由上方的性能层级决定；单独修改会保存为覆盖项。"`,
	} {
		if !strings.Contains(dict, want) {
			t.Errorf("i18n.js missing reworded hint.effort.go_unbound entry: %s", want)
		}
	}
	// The stale/mistranslated values must all be gone.
	for _, banned := range []string{"(declarative — not read by the runtime)", "(런타임 미반영)", "(Go 미독)", "（Go 未バインド）"} {
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
	if !strings.Contains(body, "Resolved from the performance tier above — per-agent edits save as overrides.") {
		t.Error("the templ baseline does not render the reworded English hint.effort.go_unbound text")
	}

	// Both templ call sites carry the reworded English baseline (the schemaSelectRow
	// branch has no live `.effort` FieldDef today, so it is asserted at the source level).
	src := readGoSource(t, "fieldsets.templ")
	if n := strings.Count(src, `data-i18n="hint.effort.go_unbound">Resolved from the performance tier above — per-agent edits save as overrides.`); n != 2 {
		t.Errorf("fieldsets.templ has %d reworded hint.effort.go_unbound baselines, want 2 (schemaSelectRow + agentFMRow)", n)
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

// TestModelOptLabelsEnglishUnified verifies the model <option> labels render as
// unified English names across all 4 locales (no context-window annotation):
// Fable 5 / Opus 5 / Sonnet 5 / Haiku 4.5. opus/sonnet/fable are exposed ONLY as
// their [1m] variants (1M always on), so the picker surface is exactly 4 options
// and each English label appears exactly 4 times in i18n.js (one per locale).
func TestModelOptLabelsEnglishUnified(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	wants := map[string]string{
		"f.model.opt.fable[1m]":  "Fable 5",
		"f.model.opt.opus[1m]":   "Opus 5",
		"f.model.opt.sonnet[1m]": "Sonnet 5",
		"f.model.opt.haiku":      "Haiku 4.5",
	}
	for key, val := range wants {
		entry := `"` + key + `": "` + val + `"`
		if n := strings.Count(dict, entry); n != 4 {
			t.Errorf("i18n.js has %d occurrences of %s, want 4 (one per locale, English-unified)", n, entry)
		}
	}
	// The old context-window-annotated labels must be gone from the model opt keys.
	for _, banned := range []string{
		`"f.model.opt.opus": "Opus 4.8 (200K)"`,
		`"f.model.opt.opus[1m]": "Opus 4.8 (1M)"`,
		`"f.model.opt.sonnet": "Sonnet 5 (200K)"`,
		`"f.model.opt.sonnet[1m]": "Sonnet 5 (1M)"`,
		`"f.model.opt.fable": "Fable 5 (200K)"`,
		`"f.model.opt.fable[1m]": "Fable 5 (1M)"`,
		`"f.model.opt.haiku": "Haiku 4.5 (200K)"`,
	} {
		if strings.Contains(dict, banned) {
			t.Errorf("i18n.js still carries the old context-window label %q", banned)
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

// agentFMSectionCount extracts the numeric field count rendered next to the
// agentfm panel title. sec.agentfm.title appears twice — first in the rail's tab
// list, then in the panel heading — so the LAST occurrence is the heading.
func agentFMSectionCount(t *testing.T, body string) int {
	t.Helper()
	heading := strings.LastIndex(body, `data-i18n="sec.agentfm.title"`)
	if heading < 0 {
		t.Fatal("agentfm panel heading not found in render")
	}
	if !strings.HasPrefix(body[heading:], `data-i18n="sec.agentfm.title">Agents</span></h2>`) {
		t.Fatalf("expected the agentfm panel heading at the last sec.agentfm.title occurrence, got: %.80q", body[heading:])
	}
	rest := body[heading:]
	marker := `class="panel__meta">`
	at := strings.Index(rest, marker)
	if at < 0 {
		t.Fatal("agentfm panel field count not found in render")
	}
	rest = rest[at+len(marker):]
	digits := regexp.MustCompile(`^\s*(\d+)`).FindStringSubmatch(rest)
	if digits == nil {
		t.Fatalf("agentfm panel field count is not numeric: %.40q", rest)
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
	// inherit → the inherit glyph (🩵), the default model badge — NOT the CUSTOM pill.
	for _, name := range []string{"manager-spec", "manager-develop", "manager-docs"} {
		b := agentTierBadge(name, v4manifest.ModelInherit, v4manifest.EffortXhigh)
		if b.IsCustom {
			t.Errorf("%s (model=inherit): CUSTOM badge rendered (effort=xhigh is not max)", name)
		}
		if !b.HasBadge {
			t.Errorf("%s (model=inherit): expected a badge", name)
		}
		if want := v4manifest.ModelColor(v4manifest.ModelInherit); b.Glyph != want {
			t.Errorf("%s (model=inherit): glyph = %q, want the inherit glyph %q", name, b.Glyph, want)
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

// TestAgentTierBadgeGlyphConsistency verifies the badge is model-derived and
// consistent: with the shipped `model: inherit` frontmatter, EVERY seeded core
// agent renders the inherit glyph (🩵) — no mixed CUSTOM/glyph rows.
func TestAgentTierBadgeGlyphConsistency(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"manager-spec", "manager-develop", "manager-docs", "manager-git"} {
		seedAgentFMFile(t, root, "moai", name, v4manifest.ModelInherit, "")
	}
	body := renderAgentFMBody(t, root)

	if strings.Contains(body, "agentfm-badge--custom") {
		t.Error("a CUSTOM badge renders for shipped `model: inherit` agents")
	}
	want := v4manifest.ModelColor(v4manifest.ModelInherit)
	if !strings.Contains(body, want) {
		t.Errorf("inherit glyph %q missing from the render — all seeded agents share model=inherit", want)
	}
}
