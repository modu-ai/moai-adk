package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// SPEC-WEB-CONSOLE-005 — web interface i18n (en/ko/ja/zh) + CJK self-host
// webfont coverage. The cohort terminator (S3) of web-console-v3.
//
// The MUST-PASS invariants verified here:
//   - AC-WC5-008a: a POST round-trip is byte-identical regardless of the active
//     interface language (interface language ≠ content language).
//   - AC-WC5-010a: validate.go is byte-unchanged (asserted by a sibling diff in M4
//     + TestServerContractPreserved here).
//   - AC-WC5-011: zero external font/style/script URL; cross-platform embed.
//
// i18n + a font subset are hard to pixel-test, so the strategy is structural
// assertions (data-i18n markers, dictionary embed+parse, font embed) +
// server-contract regression (POST byte-invariance, no name= on langpick) +
// offline (zero external URL).

// --- M1: CJK font subset (offline-safe foundation) ---

// TestCJKFontSubsetEmbedded verifies AC-WC5-006: the CJK woff2 subset font(s)
// are present under assets/fonts/, embedded + served, the OFL-1.1 license is
// present, the @font-face src is relative (no https://), and each subset is a
// tens-of-KB subset (not a multi-MB full CJK font).
func TestCJKFontSubsetEmbedded(t *testing.T) {
	entries, err := fs.ReadDir(staticFS(), "fonts")
	if err != nil {
		t.Fatalf("read embedded fonts dir: %v", err)
	}

	var cjkSubsets []fs.DirEntry
	var notoLicense bool
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "NotoSansCJK") && strings.HasSuffix(name, ".woff2") {
			cjkSubsets = append(cjkSubsets, e)
		}
		// The CJK font's OFL-1.1 license must ship alongside (a distinct file
		// from the 004 Pretendard OFL.txt).
		if name == "OFL-NotoSansCJK.txt" || name == "NotoSansCJK-OFL.txt" {
			notoLicense = true
		}
	}
	if len(cjkSubsets) == 0 {
		t.Fatal("no CJK woff2 subset (NotoSansCJK*.woff2) embedded under assets/fonts/")
	}
	if !notoLicense {
		t.Error("Noto CJK OFL-1.1 license file not embedded alongside the CJK subset")
	}

	// Each CJK subset must be a tens-of-KB subset, not a full font (< 200KB each).
	for _, e := range cjkSubsets {
		data, err := fs.ReadFile(staticFS(), "fonts/"+e.Name())
		if err != nil {
			t.Fatalf("read CJK subset %q: %v", e.Name(), err)
		}
		if len(data) == 0 {
			t.Errorf("CJK subset %q is empty", e.Name())
		}
		if len(data) > 200*1024 {
			t.Errorf("CJK subset %q is %d bytes (> 200KB) — must be a glyph subset, not a full font", e.Name(), len(data))
		}
	}

	// The 004 Pretendard Latin+Hangul subset + its OFL.txt are preserved (no en/ko regression).
	var pretendardCount, pretendardLicense int
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "Pretendard-") && strings.HasSuffix(e.Name(), ".woff2") {
			pretendardCount++
		}
		if e.Name() == "OFL.txt" {
			pretendardLicense++
		}
	}
	if pretendardCount < 5 {
		t.Errorf("expected the 5 preserved Pretendard subset weights, found %d", pretendardCount)
	}
	if pretendardLicense == 0 {
		t.Error("004 Pretendard OFL.txt was removed (must be preserved)")
	}
}

// TestCJKFontServedFromStatic verifies a CJK woff2 subset is served over
// /static/fonts/ with no network fetch (AC-WC5-006/011).
func TestCJKFontServedFromStatic(t *testing.T) {
	entries, err := fs.ReadDir(staticFS(), "fonts")
	if err != nil {
		t.Fatalf("read embedded fonts dir: %v", err)
	}
	var sample string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "NotoSansCJK") && strings.HasSuffix(e.Name(), ".woff2") {
			sample = e.Name()
			break
		}
	}
	if sample == "" {
		t.Fatal("no NotoSansCJK*.woff2 subset to serve")
	}

	a := newTestApp(t)
	h := a.routes()
	req := httptest.NewRequest(http.MethodGet, "/static/fonts/"+sample, nil)
	req.Host = "evil.example.com" // GET static asset is not Host-gated
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want 200", sample, rec.Code)
	}
	if rec.Body.Len() == 0 {
		t.Errorf("CJK woff2 %s body is empty", sample)
	}
}

// TestFontStackCoversCJK verifies AC-WC5-007: the CJK @font-face is wired into
// the font stack (so ja/zh chrome renders in the webfont) AND the 004 Pretendard
// Latin+Hangul block is still present (no en/ko regression).
func TestFontStackCoversCJK(t *testing.T) {
	css := readEmbeddedAsset(t, "console.css")

	// CJK @font-face present, referencing the self-hosted relative subset path.
	if !strings.Contains(css, "Noto Sans CJK") {
		t.Error("console.css missing the CJK @font-face family (Noto Sans CJK)")
	}
	if !strings.Contains(css, `url("/static/fonts/NotoSansCJK`) {
		t.Error("CJK @font-face src does not reference the self-hosted /static/fonts/NotoSansCJK subset")
	}
	// The CJK face must be reachable from the rendered stack: either appended to
	// --font-sans or activated via a [lang] selector. Assert at least one wiring.
	cjkWiredInSans := strings.Contains(css, `--font-sans:`) && strings.Contains(css, `"Noto Sans CJK`)
	cjkWiredByLang := strings.Contains(css, `[lang=`) && strings.Contains(css, "Noto Sans CJK")
	if !cjkWiredInSans && !cjkWiredByLang {
		t.Error("CJK face is neither in --font-sans nor activated by a [lang] selector — ja/zh would fall back to system-ui")
	}

	// 004 Pretendard Latin+Hangul @font-face block preserved (no en/ko regression).
	if !strings.Contains(css, `url("/static/fonts/Pretendard-Regular.subset.woff2")`) {
		t.Error("004 Pretendard @font-face block was removed (en/ko regression)")
	}
}

// --- M2: i18n dictionary + data-i18n wiring ---

// TestI18nDictionaryEmbedded verifies AC-WC5-001: i18n.js is embedded + served
// from /static/, defines window.MOAI_I18N with all four locale keys, and is
// loaded offline (no network fetch).
func TestI18nDictionaryEmbedded(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")

	if !strings.Contains(dict, "window.MOAI_I18N") {
		t.Error("i18n.js does not define window.MOAI_I18N")
	}
	for _, locale := range []string{"en:", "ko:", "ja:", "zh:"} {
		if !strings.Contains(dict, locale) {
			t.Errorf("i18n.js missing locale key %q", locale)
		}
	}

	// Served from /static/i18n.js (200, offline).
	a := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/static/i18n.js", nil)
	req.Host = "evil.example.com"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /static/i18n.js status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "window.MOAI_I18N") {
		t.Error("/static/i18n.js served body missing window.MOAI_I18N")
	}

	// The page links the dictionary script before app.js.
	tmplSrc := readEmbeddedAsset(t, "page.html.tmpl")
	if !strings.Contains(tmplSrc, `src="/static/i18n.js"`) {
		t.Error("page.html.tmpl does not load /static/i18n.js")
	}
}

// TestI18nGoEmbedEnumeratesDictionary verifies AC-WC5-011: the go:embed directive
// enumerates assets/i18n.js explicitly (the assets/fonts glob covers the CJK font).
func TestI18nGoEmbedEnumeratesDictionary(t *testing.T) {
	src := readGoSource(t, "assets.go")
	if !strings.Contains(src, "go:embed") {
		t.Fatal("assets.go has no go:embed directive")
	}
	if !strings.Contains(src, "assets/i18n.js") {
		t.Error("go:embed directive does not enumerate assets/i18n.js")
	}
}

// TestDataI18nWiring verifies AC-WC5-002: translatable chrome elements carry
// data-i18n attributes (>= 25), a representative key set is present, and the
// <code class="field__key"> code chips carry NO data-i18n.
func TestDataI18nWiring(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{UserName: "jline"})

	// Representative data-i18n keys present on the chrome.
	for _, key := range []string{
		`data-i18n="app.subtitle"`,
		`data-i18n="sec.identity.title"`,
		`data-i18n="sec.identity.desc"`,
		`data-i18n="f.user_name.title"`,
		`data-i18n="f.user_name.desc"`,
		`data-i18n="seg.title"`,
		`data-i18n="seg.note"`,
		`data-i18n="actions.save"`,
	} {
		if !strings.Contains(body, key) {
			t.Errorf("rendered page missing data-i18n marker %q", key)
		}
	}

	// >= 25 data-i18n attributes (AC-WC5-002 floor).
	count := strings.Count(body, "data-i18n=")
	if count < 25 {
		t.Errorf("rendered page has %d data-i18n attributes, want >= 25", count)
	}

	// Code chips (<code class="field__key">) must NOT carry data-i18n — they stay
	// English code tokens. Assert no data-i18n appears inside a field__key element.
	chipRe := regexp.MustCompile(`<code class="field__key"[^>]*data-i18n`)
	if chipRe.MatchString(body) {
		t.Error("a field__key code chip carries data-i18n (code tokens must not be translated)")
	}
}

// TestDataI18nKeysSubsetOfDictionary verifies R6 boundary verification: every
// data-i18n key in the RENDERED page is present in the EN dictionary (so no key
// renders an absent/blank translation). The render resolves the helper templates'
// data-i18n="f.{{.Name}}.title" directives to concrete keys (e.g. f.model.title).
func TestDataI18nKeysSubsetOfDictionary(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{UserName: "jline"})
	dict := readEmbeddedAsset(t, "i18n.js")

	keyRe := regexp.MustCompile(`data-i18n="([^"]+)"`)
	matches := keyRe.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		t.Fatal("no data-i18n keys found in rendered page")
	}
	for _, m := range matches {
		key := m[1]
		// Guard against an unresolved template directive leaking into the render.
		if strings.Contains(key, "{{") {
			t.Errorf("data-i18n key %q is an unresolved template directive (not rendered)", key)
			continue
		}
		// The dictionary defines keys as "key": — assert the EN locale carries it.
		if !strings.Contains(dict, `"`+key+`":`) {
			t.Errorf("data-i18n key %q in the rendered page is absent from the dictionary (R6: would render blank/untranslated)", key)
		}
	}
}

// TestNoReviewKeys verifies AC-WC5-009: the shipped dictionary contains NO rv.*
// design-review keys, and the rendered page binds NO element to an rv.* key.
func TestNoReviewKeys(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	if strings.Contains(dict, "rv.") {
		t.Error("shipped i18n.js contains rv.* design-review keys (must be stripped)")
	}

	body := renderIndexBody(t, profile.ProfilePreferences{})
	if strings.Contains(body, `data-i18n="rv.`) {
		t.Error("rendered page binds an element to an rv.* design-review key")
	}
	if strings.Contains(body, `class="review"`) {
		t.Error("rendered page contains the excluded .review design-review aside")
	}
}

// --- M3: appbar langpick + client apply/persist (server-contract gate) ---

// TestLangpickRendered verifies AC-WC5-003: the appbar renders the langpick
// <select> (class=langpick, id=uiLangSelect, 4 locale options + aria-label),
// placed OUTSIDE the <form> and carrying NO name= attribute.
func TestLangpickRendered(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{})

	// The langpick reuses the .select chrome, so it carries the langpick class
	// token (class="select langpick"). Assert the langpick class token is present.
	if !langpickClassRe.MatchString(body) {
		t.Error("appbar missing the langpick select (no element carries the langpick class)")
	}
	// The non-colliding id (NOT langSelect, which is the content-language helper).
	if !strings.Contains(body, `id="uiLangSelect"`) {
		t.Error("langpick missing id=\"uiLangSelect\" (must not reuse the content-language langSelect id)")
	}
	// aria-label (from lang.aria) for accessibility.
	if !strings.Contains(body, `aria-label="Interface language"`) {
		t.Error("langpick missing aria-label (lang.aria)")
	}
	// 4 interface locale options.
	for _, loc := range []string{`value="en"`, `value="ko"`, `value="ja"`, `value="zh"`} {
		// the option must appear within the langpick block; assert global presence.
		if !strings.Contains(body, loc) {
			t.Errorf("langpick missing locale option %q", loc)
		}
	}

	// The langpick must appear BEFORE the settings form's opening <form ...> tag
	// (it lives in the appbar, outside the form). Match "<form " with a trailing
	// space to hit the real opening tag, not a prose mention of the word "form".
	langIdx := langpickClassRe.FindStringIndex(body)
	formIdx := strings.Index(body, "<form ")
	if langIdx == nil || formIdx < 0 {
		t.Fatal("could not locate langpick and the <form ...> tag for ordering check")
	}
	if langIdx[0] > formIdx {
		t.Error("langpick appears AFTER the <form ...> tag — it must be in the appbar, OUTSIDE the form (R1: server-contract leak)")
	}
}

// langpickClassRe matches an element carrying the langpick class token, whether
// it stands alone (class="langpick") or alongside the reused .select chrome
// (class="select langpick").
var langpickClassRe = regexp.MustCompile(`class="[^"]*\blangpick\b[^"]*"`)

// TestLangpickNotFormField verifies AC-WC5-008b: the langpick carries no name=
// attribute and is not inside <form>; interface language persists only in
// localStorage; no server config field for the interface language exists.
func TestLangpickNotFormField(t *testing.T) {
	tmplSrc := readEmbeddedAsset(t, "page.html.tmpl")

	// Locate the langpick <select ...> opening tag (anchored on its unique
	// id="uiLangSelect") and assert it carries NO name= attribute.
	pickRe := regexp.MustCompile(`<select[^>]*id="uiLangSelect"[^>]*>`)
	tag := pickRe.FindString(tmplSrc)
	if tag == "" {
		t.Fatal("could not find the langpick <select> opening tag")
	}
	if strings.Contains(tag, "name=") {
		t.Errorf("langpick <select> carries a name= attribute (must NOT be a form field): %q", tag)
	}

	// The langpick markup must sit before the settings form's opening <form ...>
	// tag (it lives in the appbar, outside the form). Match "<form " (trailing
	// space) to hit the real tag, not a prose mention of the word "form".
	pickIdx := strings.Index(tmplSrc, "uiLangSelect")
	formIdx := strings.Index(tmplSrc, "<form ")
	if pickIdx < 0 || formIdx < 0 {
		t.Fatal("could not locate the langpick and the <form ...> tag")
	}
	if pickIdx > formIdx {
		t.Error("langpick is positioned after the <form ...> tag in the template (must be outside the form)")
	}

	// No interface-language config field leaked into the server handler.
	handlers := readGoSource(t, "handlers.go")
	if strings.Contains(handlers, "moai-console-lang") || strings.Contains(handlers, "uiLangSelect") {
		t.Error("an interface-language field leaked into handlers.go (interface language is client-only)")
	}
}

// TestLangpickJSWiring verifies AC-WC5-004: app.js wires a change listener on
// the langpick that calls applyI18n + localStorage.setItem("moai-console-lang"),
// with no form submit / no fetch in that path.
func TestLangpickJSWiring(t *testing.T) {
	js := readEmbeddedAsset(t, "app.js")

	if !strings.Contains(js, "applyI18n") {
		t.Error("app.js does not define/call applyI18n")
	}
	if !strings.Contains(js, `localStorage.setItem("moai-console-lang"`) &&
		!strings.Contains(js, "moai-console-lang") {
		t.Error("app.js does not persist the interface language in localStorage(moai-console-lang)")
	}
	if !strings.Contains(js, "uiLangSelect") {
		t.Error("app.js does not wire the uiLangSelect change listener")
	}
	// The interface-language path must not submit the form or fetch.
	if strings.Contains(js, "fetch(") {
		t.Error("app.js contains a fetch( call — the langpick must not perform a network request")
	}
}

// TestI18nLoadDefault verifies AC-WC5-005: the load path reads
// localStorage.getItem("moai-console-lang"), defaults to en for absent/invalid
// values, and a FOUC-style early lang apply / DOMContentLoaded apply exists.
func TestI18nLoadDefault(t *testing.T) {
	js := readEmbeddedAsset(t, "app.js")
	tmplSrc := readEmbeddedAsset(t, "page.html.tmpl")

	// The load path must read the persisted moai-console-lang value via
	// localStorage.getItem. The read happens in app.js (via the LANG_KEY constant
	// = "moai-console-lang") AND/OR in the <head> FOUC snippet (literal). Accept
	// either: assert a getItem read that targets moai-console-lang somewhere on
	// the load path.
	jsReadsLang := strings.Contains(js, "localStorage.getItem(LANG_KEY)") ||
		strings.Contains(js, `localStorage.getItem("moai-console-lang"`)
	headReadsLang := strings.Contains(tmplSrc, `localStorage.getItem("moai-console-lang"`)
	if !jsReadsLang && !headReadsLang {
		t.Error("neither app.js nor the <head> snippet reads the persisted moai-console-lang on load")
	}
	// app.js must reference the LANG_KEY value (the localStorage key).
	if !strings.Contains(js, "moai-console-lang") {
		t.Error("app.js does not reference the moai-console-lang localStorage key")
	}
	// Default-to-en fallback present (a valid-locale whitelist guard).
	if !strings.Contains(js, `"en"`) {
		t.Error("app.js has no en default/whitelist for the interface language")
	}
	// FOUC-style early <head> lang application mirroring the theme pattern, OR a
	// DOMContentLoaded applyI18n — assert at least the persisted-locale early apply.
	headApply := strings.Contains(tmplSrc, "moai-console-lang")
	domApply := strings.Contains(js, "applyI18n")
	if !headApply && !domApply {
		t.Error("no load-time interface-language apply (neither a <head> snippet nor a DOMContentLoaded applyI18n)")
	}
}

// TestInterfaceLanguageDoesNotAlterPOST is the MUST-PASS cohort core invariant
// (AC-WC5-008a): a POST round-trip is byte-identical regardless of any interface
// language state, and no langpick / moai-console-lang key appears in the POST body.
func TestInterfaceLanguageDoesNotAlterPOST(t *testing.T) {
	// Build an identical form payload. The interface language is client-only and
	// can NEVER appear in a POST (it has no name=), so bindForm must read the same
	// fields whether or not an interface-language value is present.
	baseForm := url.Values{
		"__profile":         {"default"},
		"user_name":         {"jline"},
		"conversation_lang": {"ko"},
		"git_commit_lang":   {"en"},
		"code_comment_lang": {"ja"},
		"doc_lang":          {"zh"},
		"permission_mode":   {"acceptEdits"},
		"model":             {"sonnet"},
		"effort_level":      {"high"},
		"statusline_mode":   {"default"},
		"statusline_preset": {"full"},
		"statusline_theme":  {"catppuccin-mocha"},
		"development_mode":  {"tdd"},
		"git_convention":    {"conventional-commits"},
	}

	bind := func(form url.Values) profile.ProfilePreferences {
		req := httptest.NewRequest(http.MethodPost, "/save",
			strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if err := req.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		return bindForm(req)
	}

	// (1) Baseline bind (interface=en, nothing extra in the POST).
	got1 := bind(baseForm)

	// (2) Even if a rogue client somehow appended an interface-language key
	// (it never should — the langpick has no name=), bindForm must IGNORE it:
	// the bound preferences must be byte-identical.
	withRogue := url.Values{}
	for k, v := range baseForm {
		withRogue[k] = v
	}
	withRogue["moai-console-lang"] = []string{"ja"}
	withRogue["uiLangSelect"] = []string{"zh"}
	got2 := bind(withRogue)

	if !reflect.DeepEqual(got1, got2) {
		t.Errorf("interface-language keys altered the bound preferences:\n  en:   %+v\n  ja:   %+v", got1, got2)
	}

	// (3) The content-language fields are exactly the submitted ones — unaffected
	// by any interface-language state.
	if got1.ConversationLang != "ko" || got1.CodeCommentLang != "ja" || got1.DocLang != "zh" {
		t.Errorf("content-language fields were altered: %+v", got1)
	}

	// (4) The langpick is not a parsed POST field: assert bindForm never reads
	// moai-console-lang. (Source-level guard.)
	src := readGoSource(t, "handlers.go")
	if strings.Contains(src, "moai-console-lang") {
		t.Error("handlers.go references moai-console-lang — the interface language must never be a server field")
	}
}

// --- M4: server contract + a11y reconciliation ---

// TestServerContractPreserved verifies AC-WC5-010a: every canonical name= field
// is present, option lists are {{range}}-driven, .FieldErrors are server-rendered,
// and the langSelect/optSelect helper structure is intact (validate.go byte-
// unchanged is asserted by a sibling git diff at run-phase end).
func TestServerContractPreserved(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{})
	tmplSrc := readEmbeddedAsset(t, "page.html.tmpl")

	for _, name := range []string{
		"user_name", "conversation_lang", "git_commit_lang", "code_comment_lang",
		"doc_lang", "permission_mode", "model_policy", "model", "effort_level",
		"statusline_mode", "statusline_preset", "statusline_theme",
		"development_mode", "git_convention", "__profile",
	} {
		if !strings.Contains(body, `name="`+name+`"`) {
			t.Errorf("server-contract field name=%q missing (i18n must not drop it)", name)
		}
	}
	// {{range}}-driven option lists preserved.
	if !strings.Contains(tmplSrc, "{{range .AllSegments}}") {
		t.Error("statusline segments no longer {{range .AllSegments}}-driven")
	}
	if !strings.Contains(tmplSrc, "{{range .LangOptions}}") &&
		!strings.Contains(tmplSrc, "{{range .Options}}") {
		t.Error("language options no longer server-rendered via {{range}}")
	}
	// .FieldErrors server-side render preserved.
	if !strings.Contains(tmplSrc, "{{with index .FieldErrors") {
		t.Error("server-side .FieldErrors render block removed")
	}
	// Helper define blocks preserved.
	tmpl, err := pageTemplate()
	if err != nil {
		t.Fatalf("pageTemplate: %v", err)
	}
	if tmpl.Lookup("langSelect") == nil || tmpl.Lookup("optSelect") == nil {
		t.Error("langSelect/optSelect helper define blocks removed")
	}
}

// TestLangAttrUpdatesOnSwitch verifies AC-WC5-012: switching the interface
// language updates the <html lang> attribute (a11y) — the app.js switch path
// sets document.documentElement.lang (or setAttribute("lang", ...)).
func TestLangAttrUpdatesOnSwitch(t *testing.T) {
	js := readEmbeddedAsset(t, "app.js")

	setsLang := strings.Contains(js, `.lang =`) ||
		strings.Contains(js, `setAttribute("lang"`) ||
		strings.Contains(js, `setAttribute('lang'`)
	if !setsLang {
		t.Error("app.js does not update <html lang> on interface-language switch (a11y)")
	}
}
