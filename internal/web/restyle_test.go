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

	// Brand tokens present (AC-WC4-001).
	for _, want := range []string{
		"--color-primary: #144a46",
		"--color-bg: #f3f3f3",
		"--gradient-signature:",
		`[data-theme="dark"]`,
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

	for _, marker := range []string{
		`class="appbar"`,       // appbar present
		`class="brand__badge"`, // signature-gradient brand badge
		`모두의AI`,                // brand name
		`class="loopback"`,     // loopback indicator
		`id="themeToggle"`,     // theme toggle button
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

	// (3) Structural kebab-segment-key guard (plan-audit D1): a statusline
	// segment key of the kebab form segment_<a-z>+-<a-z> must NOT appear. This
	// covers all 15 design segment keys structurally, not just 3 literals.
	kebabSeg := regexp.MustCompile(`segment_[a-z]+-[a-z]`)
	if kebabSeg.MatchString(body) {
		t.Errorf("rendered form contains a kebab-cased statusline segment key (matched %q): design keys must be dropped",
			kebabSeg.FindString(body))
	}

	// (4) Positive: the 15 canonical snake_case segment keys still render.
	for _, key := range allSegments {
		if !strings.Contains(body, "segment_"+key) {
			t.Errorf("canonical segment key segment_%s missing from rendered form", key)
		}
	}
	// (5) SPEC-WEB-CONSOLE-006 Class C mechanism retarget (spec.md §2.1.1 #4 / §D.3):
	// the prior version grepped the page.html.tmpl SOURCE for the kebab pattern and
	// for the literal `{{range .AllSegments}}` directive (proof the segments are
	// server-rendered, not hardcoded). The template source is deleted; the
	// server-render intent is retargeted to the RENDERED BODY — all 15 canonical
	// keys render (asserted in (4), which a hardcoded-3-key block would fail), and
	// the kebab guard runs against the rendered body (asserted in (3)). The
	// segments are still driven by `for _, s := range view.AllSegments` in the Templ
	// fieldsetStatusline component.
}

// TestNameAttributesPreserved verifies AC-WC4-009a: every canonical form field
// retains its name= POST attribute, and the form contract markers survive.
func TestNameAttributesPreserved(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{})

	wantNames := []string{
		"user_name", "conversation_lang", "git_commit_lang", "code_comment_lang",
		"doc_lang", "permission_mode", "model_policy", "model", "effort_level",
		"statusline_mode", "statusline_preset", "statusline_theme",
		"development_mode", "git_convention", "__profile",
	}
	for _, name := range wantNames {
		if !strings.Contains(body, `name="`+name+`"`) {
			t.Errorf("form field name=%q missing (POST attribute dropped)", name)
		}
	}
	// All 15 segment_<key> name attributes present.
	for _, key := range allSegments {
		if !strings.Contains(body, `name="segment_`+key+`"`) {
			t.Errorf("segment field name=\"segment_%s\" missing", key)
		}
	}

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

// --- M3: dark mode + theme toggle ---

// TestDarkModeAndThemeToggle verifies AC-WC4-006: the [data-theme] override block
// is present in the token CSS, the theme-toggle element is present, the FOUC
// inline <head> snippet applies the persisted theme before first paint, the
// prefers-reduced-motion guard is present, and theme persistence is client-side
// only (no server theme field).
func TestDarkModeAndThemeToggle(t *testing.T) {
	css := readEmbeddedAsset(t, "console.css")
	body := renderIndexBody(t, profile.ProfilePreferences{})
	js := readEmbeddedAsset(t, "app.js")

	if !strings.Contains(css, `[data-theme="dark"]`) {
		t.Error("token CSS missing [data-theme=\"dark\"] override block")
	}
	if !strings.Contains(css, "prefers-reduced-motion") {
		t.Error("token CSS missing prefers-reduced-motion guard")
	}
	// SPEC-WEB-CONSOLE-006 Class C mechanism retarget (spec.md §2.1.1 #6 / §D.3):
	// the prior version grepped the page.html.tmpl SOURCE for id="themeToggle" + the
	// FOUC `moai-console-theme` / `data-theme` <head> snippet. The template source is
	// deleted; both render in the BODY (the appbar emits the themeToggle button and
	// the <html data-theme> + <head> FOUC <script> render), so the assertions are
	// retargeted to the rendered body.
	if !strings.Contains(body, `id="themeToggle"`) {
		t.Error("theme-toggle element missing from rendered page")
	}
	if !strings.Contains(body, "moai-console-theme") || !strings.Contains(body, "data-theme") {
		t.Error("FOUC-prevention inline theme init missing from rendered <head>")
	}
	// Client-side persistence in app.js, no server round-trip.
	if !strings.Contains(js, "localStorage") || !strings.Contains(js, "moai-console-theme") {
		t.Error("app.js theme toggle does not persist client-side via localStorage")
	}
	// No theme field added to the server persistence path (negative guard).
	if strings.Contains(readGoSource(t, "handlers.go"), `"theme"`) {
		t.Error("a theme field leaked into the server handler (theme must be client-side only)")
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
	// the prior version grepped the page.html.tmpl SOURCE for id="themeToggle" +
	// aria-label + class="field-error" + has-error. The template source is deleted;
	// these render in the BODY — the appbar emits the themeToggle button with its
	// aria-label, and an errored render carries the field-error / has-error cues.
	if !strings.Contains(body, `id="themeToggle"`) || !strings.Contains(body, "aria-label=") {
		t.Error("theme toggle / icon-only control missing aria-label")
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
