package web

import (
	"strings"
	"testing"
)

// i18n_chrome_wiring_test.go — the primary overview and route-specific screens
// share the same i18n contract. These tests pin the emitted attributes so a
// templ refactor cannot silently drop the overview copy back to English or
// reintroduce retired operational links to the primary surface.

// TestOverviewPrimarySurfaceWiring asserts that Overview is a focused entry
// point for health, Todo, Settings, and safe quick actions. Kanban, Specs, and
// Monitor remain separately testable routes but are not primary navigation.
func TestOverviewPrimarySurfaceWiring(t *testing.T) {
	t.Parallel()

	o := OverviewVM{
		Stats: []StatVM{
			{Label: "work tracked", Value: "29", Note: "4 in-progress", NoteKey: "statNote.in-progress", NoteParams: "4"},
			{Label: "review", Value: "3", Note: "needs review", NoteKey: "statNote.needs-review"},
			{Label: "session", Value: "2/5", Note: "PID confirmed / registry", NoteKey: "statNote.pid-confirmed-registry"},
			{Label: "verify", Value: "pass", Note: "12 keys", NoteKey: "statNote.keys", NoteParams: "12"},
		},
	}
	q := TodoVM{Root: "/tmp/backlog.db", Items: []TodoItemVM{{ID: "t1", Text: "review settings", State: "queued"}}}
	html := renderTempl(t, Overview(ShellVM{Area: "overview", Title: "Overview", Profile: "default", Project: "moai-adk-go", Host: "127.0.0.1"}, o, q))

	for _, want := range []string{
		`data-i18n="overview.eyebrow"`,
		`data-i18n="overview.title"`,
		`data-i18n="overview.health.title"`,
		`data-i18n="overview.todo.title"`,
		`data-i18n="overview.actions.title"`,
		`data-i18n="overview.scope.title"`,
		`data-i18n="stat.work-tracked"`,
		`data-i18n="stat.review"`,
		`data-i18n="statNote.in-progress" data-i18n-params="4"`,
		`data-i18n="statNote.needs-review"`,
		`data-i18n="statNote.pid-confirmed-registry"`,
		`data-i18n="statNote.keys" data-i18n-params="12"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("overview markup is missing %s", want)
		}
	}
	for _, retired := range []string{`href="/kanban"`, `href="/specs"`, `href="/monitor"`} {
		if strings.Contains(html, retired) {
			t.Errorf("overview primary surface exposes retired navigation link %s", retired)
		}
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
