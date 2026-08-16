package web

import (
	"strings"
	"testing"
)

// i18n_chrome_wiring_test.go — kanban card t45 regression net: the console
// chrome strings that compose a count into the sentence ("4 in-progress",
// "registry 5", "Showing the 10 most recently updated of 40") and the idle-role
// attention row carry data-i18n wiring whose baseline is English but whose
// displayed language is decided client-side. The i18n governance suite covers
// the dictionary; these tests pin the EMITTED attributes so a templ refactor
// cannot silently drop the wiring back to untranslated English.

// TestOverviewCountedChromeWiring asserts the overview screen emits param-carrying
// i18n attributes for the counted stat notes, the Sessions registry meta, the
// registry banner, and the idle-role attention row (which must share the
// chain.stopped key with the chain band — one message, one key).
func TestOverviewCountedChromeWiring(t *testing.T) {
	t.Parallel()

	vm := OverviewVM{
		Stats: []StatVM{
			{Label: "SPEC", Value: "29", Note: "4 in-progress", NoteKey: "statNote.in-progress", NoteParams: "4"},
			{Label: "drift", Value: "3", Note: "MUST-FIX", NoteKey: "statNote.must-fix"},
			{Label: "session", Value: "2/5", Note: "PID confirmed / registry", NoteKey: "statNote.pid-confirmed-registry"},
			{Label: "verify", Value: "pass", Note: "12 keys", NoteKey: "statNote.keys", NoteParams: "12"},
		},
		Chain: ChainVM{Present: true, IdleRole: "lead"},
		Attention: []AttentionVM{{
			Icon:      "alert",
			Source:    "kanban",
			Text:      "lead session not started — the chain stops here",
			Role:      "lead",
			Badge:     "idle",
			BadgeKind: "danger",
			Href:      "/kanban",
		}},
	}
	html := renderTempl(t, Overview(ShellVM{Area: "overview", Title: "Overview"}, vm))

	for _, want := range []string{
		`data-i18n="statNote.in-progress" data-i18n-params="4"`,
		`data-i18n="statNote.must-fix"`,
		`data-i18n="statNote.pid-confirmed-registry"`,
		`data-i18n="statNote.keys" data-i18n-params="12"`,
		`data-i18n="panelMeta.registry" data-i18n-params="0"`,
		`data-i18n="banner.registry"`,
		// The attention row renders through the SAME keys the chain band uses,
		// so the two paths cannot drift apart in translation coverage again.
		`data-i18n="chain.sessionNotStarted"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("overview markup is missing %s", want)
		}
	}
	// chain.stopped appears exactly twice: once in the chain band, once in the
	// attention row — the dual render path unified onto one key.
	if n := strings.Count(html, `data-i18n="chain.stopped"`); n != 2 {
		t.Errorf("chain.stopped wiring count = %d, want 2 (chain band + attention row)", n)
	}
}

// TestSpecsFilterChipWiring asserts every status filter chip carries its
// specs.filter.* key with the count in data-i18n-params, keeping the English
// baseline ("all 29") intact for the no-JS / en path.
func TestSpecsFilterChipWiring(t *testing.T) {
	t.Parallel()

	l := SpecListVM{
		Filters: []FilterVM{
			{Label: "all", Count: 29, Href: "/specs", Active: true},
			{Label: "draft", Count: 3, Href: "/specs?status=draft"},
			{Label: "in-progress", Count: 4, Href: "/specs?status=in-progress"},
			{Label: "implemented", Count: 20, Href: "/specs?status=implemented"},
			{Label: "completed", Count: 2, Href: "/specs?status=completed"},
		},
	}
	html := renderTempl(t, Specs(ShellVM{Area: "specs", Title: "Specs"}, l))

	for _, want := range []string{
		`data-i18n="specs.filter.all" data-i18n-params="29"`,
		`data-i18n="specs.filter.draft" data-i18n-params="3"`,
		`data-i18n="specs.filter.in-progress" data-i18n-params="4"`,
		`data-i18n="specs.filter.implemented" data-i18n-params="20"`,
		`data-i18n="specs.filter.completed" data-i18n-params="2"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("specs chip markup is missing %s", want)
		}
	}
	if !strings.Contains(html, ">all 29</a>") {
		t.Error("chip English baseline 'all 29' no longer emitted")
	}
}

// TestCloseDebtShowingWiring asserts the truncation disclosure sentence keeps
// its baseline while carrying the shown/total pair through params.
func TestCloseDebtShowingWiring(t *testing.T) {
	t.Parallel()

	rows := make([]SpecRowVM, 40)
	for i := range rows {
		rows[i] = SpecRowVM{ID: "SPEC-X", Title: "t", Status: "implemented", Updated: "now"}
	}
	html := renderTempl(t, closeDebtPanel(rows))

	if !strings.Contains(html, `data-i18n="board.closedebt.showing" data-i18n-params="10,40"`) {
		t.Error("close-debt showing line is missing its key/params wiring")
	}
	if !strings.Contains(html, "Showing the 10 most recently updated of 40.") {
		t.Error("close-debt showing English baseline changed")
	}
}

// TestAppJsParamsSubstitution asserts the client actually performs the {0}
// substitution — the wiring above is inert without it.
func TestAppJsParamsSubstitution(t *testing.T) {
	t.Parallel()

	appjs := readEmbeddedAsset(t, "app.js")
	for _, want := range []string{
		"i18nSubstParams",
		`getAttribute("data-i18n-params")`,
		"textContent = i18nSubstParams(",
	} {
		if !strings.Contains(appjs, want) {
			t.Errorf("app.js params substitution is missing %q", want)
		}
	}
}
