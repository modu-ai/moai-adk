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
// markers render, and the langSelect/optSelect define blocks are still present
// and used (structure preserved).
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

	// langSelect + optSelect define blocks are still present in the template AND
	// still produce the lang/opt select chrome in the render.
	tmpl, err := pageTemplate()
	if err != nil {
		t.Fatalf("pageTemplate: %v", err)
	}
	if tmpl.Lookup("langSelect") == nil {
		t.Error("langSelect define block missing (helper structure must be preserved)")
	}
	if tmpl.Lookup("optSelect") == nil {
		t.Error("optSelect define block missing (helper structure must be preserved)")
	}
	if !strings.Contains(body, `class="select select--lang"`) {
		t.Error("langSelect helper did not render the language select chrome")
	}
}

// TestAppbarRendered verifies AC-WC4-004: the appbar renders brand badge SVG +
// 모두의AI + loopback indicator + theme toggle, and contains NO interface-language
// picker (S3 exclusion).
func TestAppbarRendered(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{})

	for _, marker := range []string{
		`class="appbar"`,       // appbar present
		`class="brand__badge"`, // signature-gradient brand badge
		`모두의AI`,                // brand name
		`class="loopback"`,     // loopback indicator
		`id="themeToggle"`,     // theme toggle button
	} {
		if !strings.Contains(body, marker) {
			t.Errorf("appbar missing marker %q", marker)
		}
	}

	// NO interface-language picker in the appbar (S3 scope, §4 E.1).
	for _, forbidden := range []string{
		`class="langpick"`,
		`id="langSelect"`,
		`data-i18n`,
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("appbar contains forbidden S3-scope element %q (interface i18n is S3)", forbidden)
		}
	}
}

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

	// The TEMPLATE source must not hardcode 127.0.0.1:3041 — the address is
	// always view-model-sourced (regression guard).
	tmplSrc := readEmbeddedAsset(t, "page.html.tmpl")
	if strings.Contains(tmplSrc, "127.0.0.1:3041") {
		t.Error("page.html.tmpl hardcodes 127.0.0.1:3041 — the bind address must come from {{.BindAddr}}")
	}
	if !strings.Contains(tmplSrc, "{{.BindAddr}}") {
		t.Error("page.html.tmpl does not render {{.BindAddr}} for the loopback indicator")
	}
}

// TestNoNonCanonicalOptions verifies AC-WC4-008 (with plan-audit D1 structural
// strengthening): the rendered form and the template contain NO non-canonical
// design options — no es/fr/de language, no haiku[1m] model, and NO kebab-cased
// statusline segment key (structural pattern, not a 3-key enumeration).
func TestNoNonCanonicalOptions(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{})
	tmplSrc := readEmbeddedAsset(t, "page.html.tmpl")

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
	if kebabSeg.MatchString(tmplSrc) {
		t.Errorf("page.html.tmpl contains a kebab-cased statusline segment key (matched %q)",
			kebabSeg.FindString(tmplSrc))
	}

	// (4) Positive: the 15 canonical snake_case segment keys still render via {{range}}.
	for _, key := range allSegments {
		if !strings.Contains(body, "segment_"+key) {
			t.Errorf("canonical segment key segment_%s missing from rendered form", key)
		}
	}
	// (5) The template still drives segments through {{range .AllSegments}} (not hardcoded).
	if !strings.Contains(tmplSrc, "{{range .AllSegments}}") {
		t.Error("statusline segments are no longer {{range .AllSegments}}-driven (server-render contract broken)")
	}
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

	// Form contract: method/action + hidden __profile + server-side .FieldErrors block.
	if !strings.Contains(body, `method="POST"`) || !strings.Contains(body, `action="/save`) {
		t.Error("form method/action contract broken")
	}
	if !strings.Contains(body, `<input type="hidden" name="__profile"`) {
		t.Error("hidden __profile input missing")
	}

	// .FieldErrors server-side render block present in the template source.
	tmplSrc := readEmbeddedAsset(t, "page.html.tmpl")
	if !strings.Contains(tmplSrc, `{{with index .FieldErrors`) {
		t.Error("server-side .FieldErrors render block missing from template")
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
	tmplSrc := readEmbeddedAsset(t, "page.html.tmpl")
	js := readEmbeddedAsset(t, "app.js")

	if !strings.Contains(css, `[data-theme="dark"]`) {
		t.Error("token CSS missing [data-theme=\"dark\"] override block")
	}
	if !strings.Contains(css, "prefers-reduced-motion") {
		t.Error("token CSS missing prefers-reduced-motion guard")
	}
	if !strings.Contains(tmplSrc, `id="themeToggle"`) {
		t.Error("theme-toggle element missing from template")
	}
	// FOUC-prevention inline <head> snippet applies persisted theme before paint.
	if !strings.Contains(tmplSrc, "moai-console-theme") || !strings.Contains(tmplSrc, "data-theme") {
		t.Error("FOUC-prevention inline theme init missing from <head>")
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
	tmplSrc := readEmbeddedAsset(t, "page.html.tmpl")

	// Inline SVGs render (the appbar brand badge + the icon subset).
	if strings.Count(body, "<svg") < 5 {
		t.Errorf("expected several inline <svg> icons, found %d", strings.Count(body, "<svg"))
	}

	// No CDN / runtime icon library.
	for _, forbidden := range []string{
		"unpkg.com", "lucide@", "data-lucide", "lucide.min.js", "cdn.jsdelivr",
	} {
		if strings.Contains(tmplSrc, forbidden) || strings.Contains(body, forbidden) {
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
	tmplSrc := readEmbeddedAsset(t, "page.html.tmpl")

	if !strings.Contains(css, "focus-visible") {
		t.Error("CSS missing a :focus-visible outline rule")
	}
	if !strings.Contains(css, "prefers-reduced-motion") {
		t.Error("CSS missing prefers-reduced-motion guard")
	}
	// Theme toggle + icon-only control carries aria-label.
	if !strings.Contains(tmplSrc, `id="themeToggle"`) || !strings.Contains(tmplSrc, "aria-label=") {
		t.Error("theme toggle / icon-only control missing aria-label")
	}
	// Error cue is non-color: the field-error span carries an icon (alert-circle)
	// plus the message text, and the field gets a has-error border class.
	if !strings.Contains(tmplSrc, `class="field-error"`) {
		t.Error("non-color error cue (.field-error icon+text span) missing")
	}
	if !strings.Contains(tmplSrc, "has-error") {
		t.Error("error border cue (.has-error) missing")
	}

	// An errored field renders aria-invalid + aria-describedby association.
	body := renderErroredBody(t)
	if !strings.Contains(body, `aria-invalid="true"`) {
		t.Error("errored field missing aria-invalid")
	}
	if !strings.Contains(body, "aria-describedby=") {
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
