package web

import (
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/settings"
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

	// The templ baseline carries the reworded English text. The agent-settings
	// row that also rendered it is gone; schemaSelectRow is the remaining call
	// site (asserted at the source level — no live `.effort` FieldDef today).
	src := readGoSource(t, "fieldsets.templ")
	if n := strings.Count(src, `data-i18n="hint.effort.go_unbound">Resolved from the performance tier above — per-agent edits save as overrides.`); n != 1 {
		t.Errorf("fieldsets.templ has %d reworded hint.effort.go_unbound baselines, want 1 (schemaSelectRow)", n)
	}
}

// TestLLMSectionRenamedGLMSettings verifies G1-4 (renamed again from
// "3rd Party LLM" to "GLM Settings", 2026-08-27): the tab strip label and the
// section title read "GLM Settings" (key sec.llm.title and tab id llm unchanged),
// and sec.llm.desc describes only the GLM backend tiers.
func TestLLMSectionRenamedGLMSettings(t *testing.T) {
	for _, tab := range consoleTabs() {
		if tab.ID == "llm" {
			if tab.LabelKey != "sec.llm.title" {
				t.Errorf("llm tab LabelKey = %q, want sec.llm.title (key must not change)", tab.LabelKey)
			}
			if tab.Baseline != "GLM Settings" {
				t.Errorf("llm tab Baseline = %q, want %q", tab.Baseline, "GLM Settings")
			}
		}
	}
	for _, meta := range schemaSectionMetas() {
		if string(meta.ID) == "llm" && meta.Title != "GLM Settings" {
			t.Errorf("llm section Title = %q, want %q", meta.Title, "GLM Settings")
		}
	}

	dict := readEmbeddedAsset(t, "i18n.js")
	for _, want := range []string{
		`"sec.llm.title": "GLM Settings"`,
		`"sec.llm.title": "GLM 설정"`,
		`"sec.llm.title": "GLM設定"`,
		`"sec.llm.title": "GLM设置"`,
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
// Fable 5 / Opus 5.5 (Recommended) / Sonnet 5 / Haiku 4.5. opus/sonnet/fable are
// exposed ONLY as their [1m] variants (1M always on), so the picker surface is
// exactly 4 options and each English label appears exactly 4 times in i18n.js
// (one per locale). The opus option carries the recommendation marker.
func TestModelOptLabelsEnglishUnified(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	wants := map[string]string{
		"f.model.opt.fable[1m]":  "Fable 5",
		"f.model.opt.opus[1m]":   "Opus 5.5 (Recommended)",
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
		`"f.model.opt.opus": "Opus 5"`,
		`"f.model.opt.opus[1m]": "Opus 5"`,
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

// TestEffortOptRecommendationLabels verifies the effort select marks medium as
// the recommended level in each locale (and no other level), and that the
// empty option states the launch fallback order honestly: the model-policy
// effort first, otherwise Claude Code's own model default
// (settings.RuntimeDefaultEffortModel). Entries are pinned per locale so one
// locale cannot satisfy another's check.
func TestEffortOptRecommendationLabels(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	for _, entry := range []string{
		`"f.effort_level.opt.medium": "Medium (Recommended)"`,
		`"f.effort_level.opt.medium": "중간 (권장)"`,
		`"f.effort_level.opt.medium": "中 (推奨)"`,
		`"f.effort_level.opt.medium": "中 (推荐)"`,
	} {
		if n := strings.Count(dict, entry); n != 1 {
			t.Errorf("i18n.js has %d occurrences of %s, want 1", n, entry)
		}
	}
	otherLevel := regexp.MustCompile(`"f\.effort_level\.opt\.(low|high|xhigh|max)": "[^"]*(Recommended|권장|推奨|推荐)[^"]*"`)
	if m := otherLevel.FindString(dict); m != "" {
		t.Errorf("a non-medium effort level carries a recommendation marker: %s", m)
	}
	for _, policy := range []string{"model policy", "모델 정책", "モデルポリシー", "模型策略"} {
		re := regexp.MustCompile(`"opt\.runtime_default": "[^"]*` + regexp.QuoteMeta(policy) + `[^"]*` + regexp.QuoteMeta(settings.RuntimeDefaultEffortModel) + `[^"]*"`)
		if n := len(re.FindAllString(dict, -1)); n != 1 {
			t.Errorf("opt.runtime_default naming %q then %s appears %d times, want 1", policy, settings.RuntimeDefaultEffortModel, n)
		}
	}
}
