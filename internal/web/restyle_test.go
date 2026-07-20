package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// osReadFile reads a package-relative source file (used for source-level
// contract assertions). The web package tests run with the package dir as the
// working directory.
func osReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// SPEC-WEB-CONSOLE-004 — 모두의AI design-system application (visual restyle,
// zero server-contract change). The tests below verify the restyle's structural
// markers and, more importantly, the server-contract regression assertions that
// guard "zero server-contract change" (the MUST-PASS invariant, REQ-WC4-009).
//
// A CSS/template restyle cannot be pixel-asserted in go test, so the strategy is
// (1) structural assertions on the rendered HTML / embedded CSS and
// (2) regression assertions that the restyle did not break a server contract.

// renderIndexBody boots a test app with a seeded profile and returns the
// rendered GET / HTML body. It mirrors the handlers_test.go pattern.
func renderIndexBody(t *testing.T, prefs profile.ProfilePreferences) string {
	t.Helper()
	a := newTestApp(t)
	a.readPreferences = func(string) (profile.ProfilePreferences, error) { return prefs, nil }
	a.listProfiles = func() []profile.ProfileEntry {
		return []profile.ProfileEntry{{Name: "default", Current: true}}
	}
	rec := serveGet(t, a.routes(), "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

// readEmbeddedAsset returns the bytes of an asset served from the static FS.
func readEmbeddedAsset(t *testing.T, path string) string {
	t.Helper()
	data, err := fs.ReadFile(staticFS(), path)
	if err != nil {
		t.Fatalf("read embedded asset %q: %v", path, err)
	}
	return string(data)
}

// --- M1: token + font layer (offline-safe foundation) ---

// TestConsoleCSSEmbedded verifies AC-WC4-001: the 모두의AI token layer is
// embedded and carries the brand tokens, and that there is NO Google-Fonts
// @import / external font-style URL anywhere in the embedded CSS.
func TestConsoleCSSEmbedded(t *testing.T) {
	css := readEmbeddedAsset(t, "console.css")

	// Brand tokens present (AC-WC4-001). The former [data-theme="dark"] entry is
	// INVERTED by SPEC-DESIGN-MOAIWEBV2-002 (light-only) — see TestDarkThemeAbsence.
	for _, want := range []string{
		"--color-primary: #3d7d5f",
		// SPEC-DESIGN-MOAIWEBV2-001 M3: bg de-tinted to the v2 achromatic canon.
		"--color-bg: #f4f4f4",
		"--gradient-signature:",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("console.css missing brand token %q", want)
		}
	}

	// No external font/style fetch (offline invariant, AC-WC4-001).
	for _, forbidden := range []string{
		"fonts.googleapis.com",
		"@import url(\"http",
		"https://fonts",
		"unpkg.com",
	} {
		if strings.Contains(css, forbidden) {
			t.Errorf("console.css contains forbidden external reference %q (offline invariant broken)", forbidden)
		}
	}
}

// TestPretendardFontSubsetEmbedded verifies AC-WC4-002: the Pretendard woff2
// subset font(s) + the OFL-1.1 license are embedded and reachable via the
// static FS, and the @font-face src uses a relative path (no https://).
func TestPretendardFontSubsetEmbedded(t *testing.T) {
	// At least one woff2 subset is present in the embed.
	entries, err := fs.ReadDir(staticFS(), "fonts")
	if err != nil {
		t.Fatalf("read embedded fonts dir: %v", err)
	}
	var woff2Count, licenseCount int
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".woff2") {
			woff2Count++
		}
		if name == "OFL.txt" || name == "LICENSE" {
			licenseCount++
		}
	}
	if woff2Count == 0 {
		t.Error("no Pretendard woff2 subset embedded under assets/fonts/")
	}
	if licenseCount == 0 {
		t.Error("OFL-1.1 license (OFL.txt / LICENSE) not embedded alongside the font")
	}

	// @font-face src must be a relative /static path, never an external https URL.
	css := readEmbeddedAsset(t, "console.css")
	if !strings.Contains(css, "@font-face") {
		t.Error("console.css has no @font-face declaration for the self-hosted font")
	}
	if !strings.Contains(css, `url("/static/fonts/Pretendard-`) {
		t.Error("@font-face src does not reference the self-hosted /static/fonts/ subset path")
	}
	if strings.Contains(css, "https://") {
		t.Error("console.css @font-face / token layer contains an https:// URL (must be offline)")
	}
}

// TestFontServedFromStatic verifies a woff2 font is served over /static/fonts/
// with no network fetch (consistent with AC-WC4-002/012).
func TestFontServedFromStatic(t *testing.T) {
	a := newTestApp(t)
	h := a.routes()

	req := httptest.NewRequest(http.MethodGet, "/static/fonts/Pretendard-Regular.subset.woff2", nil)
	req.Host = "evil.example.com" // GET static asset is not Host-gated
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET woff2 status = %d, want 200", rec.Code)
	}
	if rec.Body.Len() == 0 {
		t.Error("woff2 font body is empty")
	}
}

// --- M2: layout + component port (server-contract preservation gate) ---

// TestComponentChromePresent verifies AC-WC4-003: the 모두의AI component chrome
// markers render, and the langSelect/optSelect helpers still produce their
// select chrome (structure preserved).
//
// SPEC-WEB-CONSOLE-006 Class C mechanism retarget (spec.md §2.1.1 #2 / §D.3): the
// prior version additionally called the retired pageTemplate parse entry's Lookup
// for the langSelect / optSelect define blocks to assert the html/template
// {{define}} blocks survived. That parse entry is retired and the helpers are now
// typed Templ components; the structure-preserved
// intent is retargeted to the RENDERED BODY — the langSelect helper still emits
// `select select--lang` and the optSelect helper still emits the plain `select`.
func TestComponentChromePresent(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{UserName: "jline"})

	for _, marker := range []string{
		`class="section"`,           // section card per fieldset
		`class="section__head"`,     // section header chrome
		`class="field__title"`,      // per-field title
		`<code class="field__key">`, // key chip
		`class="field__desc"`,       // field description
		`class="select-wrap"`,       // styled select chevron affordance
		`class="seg"`,               // segment checkbox card
		`class="btn btn--primary"`,  // signature-gradient primary button
	} {
		if !strings.Contains(body, marker) {
			t.Errorf("rendered page missing component chrome marker %q", marker)
		}
	}

	// The langSelect/optSelect helpers still produce the lang/opt select chrome in
	// the render (retargeted from the retired pageTemplate parse-entry Lookup check).
	if !strings.Contains(body, `class="select select--lang"`) {
		t.Error("langSelect helper did not render the language select chrome")
	}
	if !strings.Contains(body, `class="select"`) {
		t.Error("optSelect helper did not render the plain select chrome")
	}
}

// TestAppbarRendered verifies AC-WC4-004 + AC-WC5-012: the appbar renders brand
// badge SVG + 모두의AI + loopback indicator + theme toggle, and (S3 landed) the
// interface-language picker + data-i18n chrome.
//
// SPEC-WEB-CONSOLE-005 reconciliation (R4 / AC-WC5-012): 004 forbade the S3-scope
// elements class="langpick", id="langSelect", and data-i18n in the appbar/body.
// 005 INTENTIONALLY lands them, so the guard is INVERTED — the langpick class
// token, the data-i18n chrome, and the NEW id="uiLangSelect" move from FORBIDDEN
// to EXPECTED. The stale id="langSelect" forbidden entry is NOT inverted: it
// referenced the never-landed original id; the appbar picker uses uiLangSelect,
// and langSelect is the live content-language {{define "langSelect"}} helper
// (whose rendered selects carry id="conversation_lang" etc., never the literal
// id="langSelect"). So we keep asserting the literal id="langSelect" is ABSENT.
func TestAppbarRendered(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{})

	// The former id="themeToggle" expected marker is INVERTED by
	// SPEC-DESIGN-MOAIWEBV2-002 (light-only) — see TestDarkThemeAbsence.
	for _, marker := range []string{
		`class="appbar"`,       // appbar present
		`class="brand__badge"`, // signature-gradient brand badge
		`MoAI-ADK`,             // brand name (mascot green theme rebrand)
		`class="loopback"`,     // loopback indicator
		`id="uiLangSelect"`,    // S3 langpick (the non-colliding interface id)
		`data-i18n`,            // S3 chrome translation markers
	} {
		if !strings.Contains(body, marker) {
			t.Errorf("appbar/body missing expected marker %q", marker)
		}
	}

	// The langpick carries the langpick class token (class="select langpick").
	if !appbarLangpickRe.MatchString(body) {
		t.Error("appbar missing the langpick class token (S3 interface-language picker)")
	}

	// The stale id="langSelect" literal must remain ABSENT — the appbar picker is
	// uiLangSelect, and the content-language helper renders id="<field-name>",
	// never the literal id="langSelect" (R4: do NOT invert this forbidden entry).
	if strings.Contains(body, `id="langSelect"`) {
		t.Error(`rendered body contains the literal id="langSelect" — the appbar picker must use id="uiLangSelect"`)
	}
}

// appbarLangpickRe matches the langpick class token whether standalone or
// alongside the reused .select chrome (class="select langpick").
var appbarLangpickRe = regexp.MustCompile(`class="[^"]*\blangpick\b[^"]*"`)

// TestLoopbackIndicatorShowsRealBindAddr verifies AC-WC4-005: the loopback
// indicator shows the real bound 127.0.0.1:<port> from the BindAddr view-model
// field, not a hardcoded 3041; the template has no literal 127.0.0.1:3041.
func TestLoopbackIndicatorShowsRealBindAddr(t *testing.T) {
	a := newTestApp(t)
	a.readPreferences = func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, nil
	}
	// Inject a known non-default bind address.
	a.bindAddr = func() string { return "127.0.0.1:7777" }

	rec := serveGet(t, a.routes(), "/")
	body := rec.Body.String()

	if !strings.Contains(body, "127.0.0.1:7777") {
		t.Errorf("loopback indicator did not show the injected bind address 127.0.0.1:7777:\n%s", body)
	}

	// SPEC-WEB-CONSOLE-006 Class C mechanism retarget (spec.md §2.1.1 #3 / §D.3):
	// the prior version grepped the page.html.tmpl SOURCE for a hardcoded
	// 127.0.0.1:3041 and for the {{.BindAddr}} directive. The template source is
	// deleted; the regression intent (the address is view-model-sourced, never a
	// hardcoded default port) is retargeted to the RENDERED BODY — the rendered
	// page shows the injected 127.0.0.1:7777 (asserted above) and must NOT contain
	// the default 127.0.0.1:3041, which would only appear if the port were
	// hardcoded rather than sourced from view.BindAddr.
	if strings.Contains(body, "127.0.0.1:3041") {
		t.Errorf("rendered body contains the default 127.0.0.1:3041 — the bind address must come from view.BindAddr, not a hardcoded port:\n%s", body)
	}
}

// TestNoNonCanonicalOptions verifies AC-WC4-008 (with plan-audit D1 structural
// strengthening): the rendered form and the template contain NO non-canonical
// design options — no es/fr/de language, no haiku[1m] model, and NO kebab-cased
// statusline segment key (structural pattern, not a 3-key enumeration).
func TestNoNonCanonicalOptions(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{})

	// (1) No non-canonical language options.
	for _, forbidden := range []string{`value="es"`, `value="fr"`, `value="de"`} {
		if strings.Contains(body, forbidden) {
			t.Errorf("rendered form contains non-canonical language option %q", forbidden)
		}
	}
	// (2) No non-canonical model option.
	if strings.Contains(body, "haiku[1m]") {
		t.Error("rendered form contains non-canonical model option haiku[1m]")
	}

	// (3)-(5) statusline segment-key guards removed
	// (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001): the statusline panel (including
	// the segment checkbox grid) is gone from the web console, so segment keys
	// no longer render and the kebab/snake_case segment guards are moot.
}

// TestNameAttributesPreserved verifies AC-WC4-009a: every canonical form field
// retains its name= POST attribute, and the form contract markers survive.
func TestNameAttributesPreserved(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{})

	wantNames := []string{
		"user_name", "conversation_lang", "git_commit_lang", "code_comment_lang",
		"doc_lang", "permission_mode", "model_policy", "model", "effort_level",
		// statusline_preset / statusline_theme removed
		// (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001) — no statusline panel.
		// development_mode / git_convention removed with the orphan `project`
		// panel (SPEC-DESIGN-MOAIWEBV2-001 M1) — editable via yaml config / CLI.
		"__profile",
	}
	for _, name := range wantNames {
		if !strings.Contains(body, `name="`+name+`"`) {
			t.Errorf("form field name=%q missing (POST attribute dropped)", name)
		}
	}
	// segment_<key> name-attribute assertions removed
	// (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001): the segment checkbox grid is gone.

	// Form contract: method/action + hidden __profile + server-side field-error block.
	if !strings.Contains(body, `method="POST"`) || !strings.Contains(body, `action="/save`) {
		t.Error("form method/action contract broken")
	}
	if !strings.Contains(body, `<input type="hidden" name="__profile"`) {
		t.Error("hidden __profile input missing")
	}

	// SPEC-WEB-CONSOLE-006 Class C mechanism retarget (spec.md §2.1.1 #5 / §D.3):
	// the prior version grepped the page.html.tmpl SOURCE for the `{{with index
	// .FieldErrors}}` server-side render block. The template source is deleted; the
	// server-rendered-field-error intent is retargeted to a RENDERED ERRORED BODY —
	// an invalid submission re-renders the form with a per-field `field-error` span
	// (the Templ equivalent of the {{with index .FieldErrors}} block).
	errored := renderErroredBody(t)
	if !strings.Contains(errored, `class="field-error"`) {
		t.Errorf("errored render missing the server-rendered per-field error span:\n%s", errored)
	}
}

// TestProfileSwitchNameAttrPreserved verifies the profile switcher keeps its
// name="__profile_select" POST attribute under the restyle (AC-WC4-009a).
func TestProfileSwitchNameAttrPreserved(t *testing.T) {
	a := newTestApp(t)
	a.readPreferences = func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, nil
	}
	a.listProfiles = func() []profile.ProfileEntry {
		return []profile.ProfileEntry{{Name: "default", Current: true}, {Name: "work"}}
	}
	body := serveGet(t, a.routes(), "/").Body.String()
	if !strings.Contains(body, `name="__profile_select"`) {
		t.Error("profile switcher dropped name=\"__profile_select\" under restyle")
	}
}

// TestBannerKindMapping verifies AC-WC4-010: server-set .BannerKind ok/error maps
// to banner--success/banner--error chrome; the server kind values are unchanged.
func TestBannerKindMapping(t *testing.T) {
	a := newTestApp(t)

	t.Run("error → banner--error", func(t *testing.T) {
		var got profile.ProfilePreferences
		_ = got
		a.writePreferences = func(string, profile.ProfilePreferences) error { return nil }
		a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
		// invalid submit → server sets BannerKind="error" + Banner text.
		form := url.Values{"__profile": {"default"}, "permission_mode": {"bogus-mode"}}
		rec := servePost(t, a.routes(), "/save", form)
		body := rec.Body.String()
		if !strings.Contains(body, "banner--error") {
			t.Errorf("error banner not mapped to banner--error chrome:\n%s", body)
		}
		if strings.Contains(body, "banner--success") {
			t.Error("error banner incorrectly rendered banner--success chrome")
		}
	})

	t.Run("ok → banner--success", func(t *testing.T) {
		a.writePreferences = func(string, profile.ProfilePreferences) error { return nil }
		a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
		a.writeProjectConfig = func(string, string, string) error { return nil }
		form := url.Values{"__profile": {"default"}, "permission_mode": {"acceptEdits"}}
		rec := servePost(t, a.routes(), "/save", form)
		body := rec.Body.String()
		if rec.Code != http.StatusOK {
			t.Fatalf("valid save status = %d, want 200", rec.Code)
		}
		if !strings.Contains(body, "banner--success") {
			t.Errorf("success banner not mapped to banner--success chrome:\n%s", body)
		}
	})

	// The Go handler must still emit "ok"/"error" kind values (template-local mapping).
	for _, src := range []string{readGoSource(t, "handlers.go")} {
		if !strings.Contains(src, `BannerKind = "ok"`) || !strings.Contains(src, `BannerKind = "error"`) {
			t.Error("handlers.go no longer sets BannerKind ok/error — server contract changed")
		}
	}
}

// --- M3 (SPEC-WEB-CONSOLE-004) dark mode — RETIRED by SPEC-DESIGN-MOAIWEBV2-002 ---

// TestDarkThemeAbsence verifies AC-MWA-001/002/003 by INVERTING the retired
// AC-WC4-006 dark-mode assertions (REQ-MWA-004 partially supersedes
// SPEC-WEB-CONSOLE-004 REQ-WC4-006): the console is light-only. Zero data-theme
// references in the stylesheet, no theme-toggle control, no data-theme
// attribute or theme FOUC branch in the rendered page, no moai-console-theme
// persistence key in app.js, no theme.aria i18n key — while the FOUC snippet's
// language branch (moai-console-lang -> <html lang>, REQ-WC5-005 CJK font
// activation lineage) is preserved verbatim (AC-MWA-003b).
func TestDarkThemeAbsence(t *testing.T) {
	css := readEmbeddedAsset(t, "console.css")
	body := renderIndexBody(t, profile.ProfilePreferences{})
	js := readEmbeddedAsset(t, "app.js")

	// (1) Stylesheet: zero data-theme references (AC-MWA-001); the
	// prefers-reduced-motion accessibility guard survives the retirement.
	if strings.Contains(css, "data-theme") {
		t.Error("console.css still contains a data-theme reference (dark theme must be fully retired)")
	}
	if !strings.Contains(css, "prefers-reduced-motion") {
		t.Error("token CSS missing prefers-reduced-motion guard (must survive dark retirement)")
	}

	// (2) Rendered settings page: no toggle, no data-theme attr (including the
	// server-rendered <html> attribute), no theme FOUC branch, no sun/moon icons
	// (AC-MWA-002/003a).
	for _, forbidden := range []string{`id="themeToggle"`, "data-theme", "moai-console-theme", "icon-sun", "icon-moon"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("rendered page still contains %q (dark theme must be fully retired)", forbidden)
		}
	}

	// (3) FOUC language branch preserved (AC-MWA-003b): the <head> still applies
	// the persisted interface language before first paint.
	if !strings.Contains(body, "moai-console-lang") {
		t.Error("FOUC language branch (moai-console-lang) missing from rendered <head> — REQ-MWA-003 preserves it verbatim")
	}

	// (4) app.js: theme read/write/toggle logic fully removed; the shared
	// htmx-boost re-bind path for the langpick survives (plan §B.3).
	for _, forbidden := range []string{"moai-console-theme", "data-theme", "applyTheme", "themeToggle"} {
		if strings.Contains(js, forbidden) {
			t.Errorf("app.js still contains %q (theme logic must be fully removed)", forbidden)
		}
	}
	if !strings.Contains(js, "uiLangSelect") {
		t.Error("app.js langpick wiring (uiLangSelect) lost during theme removal — plan §B.3 forbids this")
	}

	// (5) i18n dictionary: theme.aria removed from all 4 locales (AC-MWA-002c).
	if strings.Contains(readEmbeddedAsset(t, "i18n.js"), "theme.aria") {
		t.Error("i18n.js still carries the theme.aria key (must be removed from all 4 locales)")
	}

	// (6) Board page source (board.templ:15 lineage): the server-rendered
	// data-theme attribute and the board theme toggle are gone; lang="en" stays.
	boardSrc := readGoSource(t, "board.templ")
	if strings.Contains(boardSrc, "data-theme") || strings.Contains(boardSrc, "themeToggle") {
		t.Error("board.templ still carries data-theme / themeToggle (dark theme must be fully retired)")
	}
	if !strings.Contains(boardSrc, `lang="en"`) {
		t.Error(`board.templ <html> lost its lang="en" attribute (must be preserved)`)
	}

	// (7) Settings page source: theme markers gone, language branch present.
	rootSrc := readGoSource(t, "root.templ")
	if strings.Contains(rootSrc, "data-theme") || strings.Contains(rootSrc, "moai-console-theme") || strings.Contains(rootSrc, "themeToggle") {
		t.Error("root.templ still carries theme markers (dark theme must be fully retired)")
	}
	if !strings.Contains(rootSrc, "moai-console-lang") {
		t.Error("root.templ FOUC language branch (moai-console-lang) missing — AC-MWA-003b")
	}

	// (8) No theme field on the server persistence path (negative guard kept
	// from the retired test — the retirement must not push theme server-side).
	if strings.Contains(readGoSource(t, "handlers.go"), `"theme"`) {
		t.Error("a theme field leaked into the server handler (no server-side theme successor)")
	}
}

// --- M4: inline-SVG icon subset ---

// TestInlineSVGIconsNoCDN verifies AC-WC4-007: icons render as inline <svg>, no
// lucide CDN <script>, no data-lucide runtime markup, no icon-library runtime JS.
func TestInlineSVGIconsNoCDN(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{})

	// Inline SVGs render (the appbar brand badge + the icon subset).
	if strings.Count(body, "<svg") < 5 {
		t.Errorf("expected several inline <svg> icons, found %d", strings.Count(body, "<svg"))
	}

	// SPEC-WEB-CONSOLE-006 Class C mechanism retarget (spec.md §2.1.1 #7 / §D.3):
	// the prior version also grepped the page.html.tmpl SOURCE for icon CDN markers.
	// The template source is deleted; the no-CDN icon invariant is asserted against
	// the RENDERED BODY (the icons are inline <svg> emitted by the Templ icon helper,
	// so any CDN reference would appear in the body).
	for _, forbidden := range []string{
		"unpkg.com", "lucide@", "data-lucide", "lucide.min.js", "cdn.jsdelivr",
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("icon CDN / runtime reference %q present (offline icon invariant broken)", forbidden)
		}
	}
}

// --- M5: accessibility ---

// TestAccessibilityCues verifies AC-WC4-011: non-color error cues (icon + border
// + text), focus-visible outline rule, prefers-reduced-motion guard, and ARIA
// (aria-label on theme toggle, aria-invalid on errored fields).
func TestAccessibilityCues(t *testing.T) {
	css := readEmbeddedAsset(t, "console.css")
	body := renderIndexBody(t, profile.ProfilePreferences{})

	if !strings.Contains(css, "focus-visible") {
		t.Error("CSS missing a :focus-visible outline rule")
	}
	if !strings.Contains(css, "prefers-reduced-motion") {
		t.Error("CSS missing prefers-reduced-motion guard")
	}
	// SPEC-WEB-CONSOLE-006 Class C mechanism retarget (spec.md §2.1.1 #8 / §D.3):
	// these render in the BODY. SPEC-DESIGN-MOAIWEBV2-002 retired the theme
	// toggle, so the icon-only-control aria-label cue is asserted on the
	// remaining #serverShutdown appbar icon button instead.
	if !strings.Contains(body, `id="serverShutdown"`) || !strings.Contains(body, "aria-label=") {
		t.Error("icon-only appbar control (#serverShutdown) missing aria-label")
	}
	// Error cue is non-color: an errored render carries the field-error icon+text
	// span and the has-error border class.
	errored := renderErroredBody(t)
	if !strings.Contains(errored, `class="field-error"`) {
		t.Error("non-color error cue (.field-error icon+text span) missing")
	}
	if !strings.Contains(errored, "has-error") {
		t.Error("error border cue (.has-error) missing")
	}
	// An errored field renders aria-invalid + aria-describedby association.
	if !strings.Contains(errored, `aria-invalid="true"`) {
		t.Error("errored field missing aria-invalid")
	}
	if !strings.Contains(errored, "aria-describedby=") {
		t.Error("errored field missing aria-describedby association")
	}
}

// --- SPEC-DESIGN-MOAIWEBV2-002 M2: status tokens = docs-site bytes + AA carve-outs ---

// TestStatusTokensDocsSiteParity verifies AC-MWA-005/006 (regression lock): the
// four semantic status tokens are byte-equal to the docs-site moai-brand.css
// baseline (L37-40), and the contrast-failing status-TEXT usages consume the
// usage-scoped color-mix darkening toward --color-ink — the token bytes
// themselves never darken (token-vs-usage separation, REQ-MWA-005/006).
func TestStatusTokensDocsSiteParity(t *testing.T) {
	css := readEmbeddedAsset(t, "console.css")

	// (1) Token bytes equal to docs-site moai-brand.css.
	for _, want := range []string{
		"--color-success: #5db872",
		"--color-warning: #d4a017",
		"--color-danger:  #c64545",
		"--color-info:    #5db8a6",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("console.css missing docs-site status token %q", want)
		}
	}

	// (2) Usage-scoped AA carve-out: derived text tokens darken via color-mix
	// toward the ink token, and the pre-measured failing text usages
	// (.banner--success / .banner--error / has-error danger text) consume them.
	for _, want := range []string{
		"--status-text-success: color-mix(in srgb, var(--color-success), var(--color-ink)",
		"--status-text-danger: color-mix(in srgb, var(--color-danger), var(--color-ink)",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("console.css missing usage-scoped AA carve-out token %q", want)
		}
	}
	for _, usage := range []string{
		".banner--success",
		".banner--error",
	} {
		if !regexp.MustCompile(regexp.QuoteMeta(usage) + `[^}]*color: var\(--status-text-`).MatchString(css) {
			t.Errorf("%s text color does not consume the --status-text-* AA carve-out", usage)
		}
	}
}

// --- SPEC-DESIGN-MOAIWEBV2-002 M3: Goorm Sans Code self-host subset ---

// TestGoormSansCodeSelfHosted verifies AC-MWA-008/009/010 (regression lock):
// the Goorm Sans Code woff2 subset + its OFL license file are embedded under
// assets/fonts/, console.css registers the @font-face with a relative
// /static/fonts/ src and leads --font-mono with "Goorm Sans Code", and the
// subset is served offline from /static/fonts/ (REQ-MWA-008/009/010).
// License provenance: goorm-sans.goorm.io states Goorm Sans / Goorm Sans Code
// follow the SIL Open Font License (verified before the artifact was committed,
// REQ-MWA-011).
func TestGoormSansCodeSelfHosted(t *testing.T) {
	// (1) Embedded artifact + license file.
	entries, err := fs.ReadDir(staticFS(), "fonts")
	if err != nil {
		t.Fatalf("read embedded fonts dir: %v", err)
	}
	var woff2Count int
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "GoormSansCode") && strings.HasSuffix(e.Name(), ".woff2") {
			woff2Count++
		}
	}
	if woff2Count == 0 {
		t.Error("no GoormSansCode*.woff2 subset embedded under assets/fonts/")
	}
	license := readEmbeddedAsset(t, "fonts/OFL-GoormSansCode.txt")
	if !strings.Contains(license, "SIL OPEN FONT LICENSE") {
		t.Error("OFL-GoormSansCode.txt missing the SIL OPEN FONT LICENSE text")
	}

	// (2) @font-face registration: relative /static/fonts/ src, no external URL.
	css := readEmbeddedAsset(t, "console.css")
	if !strings.Contains(css, `font-family: "Goorm Sans Code"`) {
		t.Error("console.css missing the Goorm Sans Code @font-face registration")
	}
	if !strings.Contains(css, `url("/static/fonts/GoormSansCode-`) {
		t.Error("Goorm Sans Code @font-face src is not the relative /static/fonts/ subset path")
	}

	// (3) --font-mono leads with Goorm Sans Code; the OS fallback stack follows.
	if !regexp.MustCompile(`--font-mono:\s*"Goorm Sans Code",\s*ui-monospace`).MatchString(css) {
		t.Error(`--font-mono does not lead with "Goorm Sans Code" followed by the OS fallback stack`)
	}

	// (4) Served offline from the embed (200, non-empty).
	a := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/static/fonts/GoormSansCode-Regular.subset.woff2", nil)
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET GoormSansCode subset status = %d, want 200", rec.Code)
	}
	if rec.Body.Len() == 0 {
		t.Error("GoormSansCode subset served body is empty")
	}
}

// renderErroredBody POSTs an invalid submission and returns the re-rendered body
// (which carries a per-field error).
func renderErroredBody(t *testing.T) string {
	t.Helper()
	a := newTestApp(t)
	a.writePreferences = func(string, profile.ProfilePreferences) error { return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	form := url.Values{"__profile": {"default"}, "permission_mode": {"definitely-not-a-mode"}}
	return servePost(t, a.routes(), "/save", form).Body.String()
}

// readGoSource reads a Go source file from the package directory for source-level
// contract assertions (e.g., the server still emits BannerKind ok/error).
func readGoSource(t *testing.T, name string) string {
	t.Helper()
	data, err := osReadFile(name)
	if err != nil {
		t.Fatalf("read go source %q: %v", name, err)
	}
	return string(data)
}
