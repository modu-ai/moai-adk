package web

// page_chrome_states_test.go — TG-5 (SPEC-WEB-CONSOLE-018 REQ-009).
//
// Defect answer: these tests catch a profile action rendering for the wrong
// permission state, an unknown icon name silently rendering a wrong or broken
// glyph, a status banner losing its error variant, or the settings form losing
// its atomic-save chrome (hidden profile/tab inputs).

import (
	"strings"
	"testing"
)

// tg5ShellVM is a settings-area shell with one sibling profile.
func tg5ShellVM() ShellVM {
	return ShellVM{
		Area: "settings", Tab: "identity", Title: "Settings", Crumb: "c",
		Host: "127.0.0.1:3041", Profile: "default", Project: "proj",
		ProjectPath: "/tmp/proj", Lang: "en", Live: "on", RenderedAt: "12:00:00",
		Tabs: []TabVM{{ID: "identity", LabelKey: "sec.identity.title", Label: "Identity", Fields: 1}},
	}
}

// TestProfileModifiableGating pins the mutability gate exactly: the default
// profile, the session-current profile, and the profile being edited are all
// off limits; anything else is modifiable. The server enforces the same block —
// the UI offering a control the server will refuse is the defect.
func TestProfileModifiableGating(t *testing.T) {
	view := pageView{SelectedProfile: "work"}
	cases := []struct {
		name    string
		profile string
		current bool
		want    bool
	}{
		{"default profile is protected", "default", false, false},
		{"session-current profile is protected", "work", true, false},
		{"profile under edit is protected", "work", false, false},
		{"an unrelated stored profile is modifiable", "lab", false, true},
	}
	for _, tc := range cases {
		if got := profileModifiable(view, tc.profile, tc.current); got != tc.want {
			t.Errorf("%s: profileModifiable = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestProfileActionRoutes pins the three action targets: they are the POST
// routes the server registers — a renamed route here would submit into a 404.
func TestProfileActionRoutes(t *testing.T) {
	if got := profileCreateAction(); got != "/profile/create" {
		t.Errorf("create action = %q", got)
	}
	if got := profileRenameAction(); got != "/profile/rename" {
		t.Errorf("rename action = %q", got)
	}
	if got := profileDeleteAction(); got != "/profile/delete" {
		t.Errorf("delete action = %q", got)
	}
}

// TestIconSVGKnownAndUnknown pins the 1-arg icon family: a known name renders
// its inline SVG; an unknown name renders NOTHING — never a stale glyph from
// another name, and never a CDN fallback.
func TestIconSVGKnownAndUnknown(t *testing.T) {
	if svg := iconSVG("save"); !strings.Contains(svg, `class="icon-save"`) {
		t.Errorf("iconSVG(save) lost its glyph: %q", svg)
	}
	if svg := iconSVG("not-an-icon"); svg != "" {
		t.Errorf("iconSVG(unknown) rendered %q, want empty", svg)
	}
}

// TestIconAtKnownUnknownAndSize pins the sized icon family: a known name renders
// an svg at the requested pixel size, an unknown name renders nothing, and the
// path helper carries its geometry.
func TestIconAtKnownUnknownAndSize(t *testing.T) {
	known := renderTempl(t, iconAt("monitor", 16))
	for _, want := range []string{`<svg viewBox="0 0 18 18"`, `width="16"`, `height="16"`, `d="M2 10h3l2-5 2.6 10L12 10h4"`} {
		if !strings.Contains(known, want) {
			t.Errorf("iconAt(monitor,16) missing %q:\n%s", want, known)
		}
	}
	if out := renderTempl(t, iconAt("not-an-icon", 16)); strings.Contains(out, "<svg") {
		t.Errorf("iconAt(unknown) rendered a glyph box:\n%s", out)
	}
	path := renderTempl(t, iconPath("M1 1L2 2", 14))
	if !strings.Contains(path, `d="M1 1L2 2"`) || !strings.Contains(path, `width="14"`) {
		t.Errorf("iconPath lost its geometry or size:\n%s", path)
	}
	if out := renderTempl(t, brandMark()); !strings.Contains(out, `class="rail__logo"`) {
		t.Errorf("brandMark lost its logo markup:\n%s", out)
	}
}

// TestBannerVariants pins the status banner: an error kind takes the danger
// chrome and the alert icon, a success kind stays the neutral note — the design
// system carries exactly one chromatic accent, so a success banner must not
// borrow the error variant.
func TestBannerVariants(t *testing.T) {
	errored := renderTempl(t, banner(pageView{Banner: "save rejected", BannerKind: "error"}))
	for _, want := range []string{`banner banner--warn`, `role="status"`, "save rejected"} {
		if !strings.Contains(errored, want) {
			t.Errorf("error banner missing %q:\n%s", want, errored)
		}
	}
	ok := renderTempl(t, banner(pageView{Banner: "saved", BannerKind: "ok"}))
	if strings.Contains(ok, "banner--warn") {
		t.Errorf("a success banner borrowed the error variant:\n%s", ok)
	}
}

// TestBannerClassMapping pins the class helper directly: only "error" maps to
// the warn chrome; every other kind is neutral.
func TestBannerClassMapping(t *testing.T) {
	if got := bannerClass("error"); got != "banner banner--warn" {
		t.Errorf("bannerClass(error) = %q", got)
	}
	for _, kind := range []string{"ok", "info", ""} {
		if got := bannerClass(kind); got != "banner" {
			t.Errorf("bannerClass(%q) = %q, want neutral", kind, got)
		}
	}
}

// TestSaveActionProfileTarget pins the form action: a selected profile targets
// the per-profile save route; no selection targets plain /save.
func TestSaveActionProfileTarget(t *testing.T) {
	if got := saveAction("work"); got != "/save?profile=work" {
		t.Errorf("saveAction(selected) = %q", got)
	}
	if got := saveAction(""); got != "/save" {
		t.Errorf("saveAction(empty) = %q", got)
	}
}

// TestSettingsPageHeaderNamesActiveTab pins the header: it names the active
// tab, counts the sibling pages from the viewmodel, and keeps the direct-link
// route pointing at the active tab.
func TestSettingsPageHeaderNamesActiveTab(t *testing.T) {
	html := renderTempl(t, settingsPageHeader(tg5ShellVM(), pageView{SelectedProfile: "default"}))
	for _, want := range []string{
		`data-i18n="sec.identity.title"`, // the active tab's label key
		`Identity`,
		`settings.detail.pages`, // the page-count line
		`/settings?tab=identity`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("settings page header missing %q:\n%s", want, html)
		}
	}
}

// TestSettingsPageFormContract pins the atomic-save chrome of the full page:
// the form carries the hidden profile and tab inputs, the active tabpanel is
// marked is-active while every panel stays in the DOM (an absent panel's fields
// are never submitted — the atomic contract depends on their presence), and an
// error banner renders above the header.
func TestSettingsPageFormContract(t *testing.T) {
	view := pageView{
		SelectedProfile: "work",
		FieldErrors:     map[string]string{"user_name": "required"},
		Banner:          "save rejected", BannerKind: "error",
		SchemaValues: map[string]string{},
	}
	html := renderTempl(t, page(tg5ShellVM(), view))

	for _, want := range []string{
		`name="__profile" value="work"`, // hidden profile input
		`name="__tab" value="identity"`, // hidden tab input
		`action="/save?profile=work"`,   // per-profile save target
		`is-active`,                     // the active panel is marked
		`banner--warn`, "save rejected", // the error banner rendered
		`aria-invalid="true"`, "required", // the field error surfaced
	} {
		if !strings.Contains(html, want) {
			t.Errorf("settings page missing %q:\n%s", want, html)
		}
	}
	// Every console tab contributes its panel element regardless of active state.
	for _, id := range []string{"identity", "language", "launch", "agentfm", "mcp", "codex"} {
		if !strings.Contains(html, `data-panel="`+id+`"`) {
			t.Errorf("panel %q is missing from the DOM (atomic-save contract): %s", id, id)
		}
	}

	// No banner kind → no banner element at all.
	cleanView := pageView{SelectedProfile: "work", SchemaValues: map[string]string{}}
	clean := renderTempl(t, page(tg5ShellVM(), cleanView))
	if strings.Contains(clean, `role="status" aria-live="polite"`) {
		t.Errorf("a banner-less view rendered a banner:\n%s", clean)
	}
}
