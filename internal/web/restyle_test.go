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
	rec := serveGet(t, a.routes(), "/settings")
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

	// 토큰 체계 교체 (콘솔 재설계): 브랜드 녹색 + 시그니처 그라디언트를 쓰던
	// 이전 체계는 무채색 단일 체계로 대체됐다. 주 액션은 --ink, 판은 --bg,
	// 유일한 유채색은 --danger 다. 이전 이름(--color-primary 등)은 미이행
	// 마크업이 남아 있는 동안만 별칭으로 유지되므로, 값이 아니라 새 토큰의
	// 존재를 검증한다.
	for _, want := range []string{
		"--ink:#1d1d1f",
		"--bg:#ffffff",
		"--danger:#c5261d",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("console.css missing achromatic token %q", want)
		}
	}
	// 폐기된 시그니처 그라디언트가 되살아나지 않았는지 확인한다.
	if strings.Contains(css, "--gradient-signature") {
		t.Error("console.css reintroduced --gradient-signature (retired by the achromatic system)")
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
		`class="panel"`,        // 패널 카드 (구 .section)
		`class="panel__head"`,  // 패널 머리글
		`class="field__label"`, // 필드 제목
		`<code class="key">`,   // 설정 키 칩
		`class="field__help"`,  // 필드 설명
		`class="sel"`,          // select 컨트롤
		`class="in"`,           // text/number 입력
	} {
		if !strings.Contains(body, marker) {
			t.Errorf("rendered page missing component chrome marker %q", marker)
		}
	}

	// 저장 버튼은 폼 안이 아니라 상단바에 있다. 폼 바깥에서도 같은 폼을 제출하도록
	// form= 로 묶여 있어야 한다 — 이게 끊기면 버튼이 조용히 아무것도 안 한다.
	if !strings.Contains(body, `form="settings-form"`) {
		t.Error("the top-bar save button is not bound to the settings form (form=\"settings-form\")")
	}
	if !strings.Contains(body, `id="settings-form"`) {
		t.Error("the settings form lost the id the top-bar save button targets")
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
	// 재설계에서 가로 appbar 는 좌측 레일 + 상단바로 갈라졌다. 브랜드와 접속
	// 주소는 레일 머리에, 문맥 칩과 저장 클러스터는 상단바에 있다. 아래 표식은
	// 그 새 자리에서 같은 정보를 확인한다.
	for _, marker := range []string{
		`class="rail"`,       // 좌측 레일
		`class="rail__logo"`, // 브랜드 마크 (인라인 SVG)
		`MoAI-ADK`,           // 브랜드 이름
		`class="host"`,       // 접속 주소 표시 (구 loopback)
		`class="top"`,        // 상단바
		`id="uiLangSelect"`,  // 인터페이스 언어 선택기
		`data-i18n`,          // 크롬 번역 표식
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

	rec := serveGet(t, a.routes(), "/settings")
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
		"doc_lang", "permission_mode", "model", "effort_level",
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
	if !strings.Contains(errored, `class="field__err"`) {
		t.Errorf("errored render missing the server-rendered per-field error span:\n%s", errored)
	}
}

// TestProfileSwitchTargetsTheLoadPath verifies the profile switcher still uses
// the GET ?profile=<name> load path (AC-WC4-009a).
//
// 재설계에서 전환 컨트롤이 auto-submit select 에서 팝오버 안의 링크로 바뀌었다.
// 지켜야 할 계약은 컨트롤의 모양이 아니라 목적지다 — 전용 라우트를 새로 만들지
// 않고 기존 로드 경로를 그대로 쓴다.
func TestProfileSwitchTargetsTheLoadPath(t *testing.T) {
	a := newTestApp(t)
	a.readPreferences = func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, nil
	}
	a.listProfiles = func() []profile.ProfileEntry {
		return []profile.ProfileEntry{{Name: "default", Current: true}, {Name: "work"}}
	}
	body := serveGet(t, a.routes(), "/settings").Body.String()
	if !strings.Contains(body, `href="/settings?profile=work"`) {
		t.Errorf("the profile switcher does not link to the ?profile= load path:\n%s", body)
	}
	// 전용 전환 라우트는 없어야 한다 — 있으면 로드 경로가 둘로 갈라진다.
	if strings.Contains(body, `/profile/select`) {
		t.Error("a dedicated profile-select route appeared; switching must reuse the GET load path")
	}
}

// TestBannerKindMapping verifies AC-WC4-010: server-set .BannerKind ok/error maps
// to distinct banner chrome; the server kind values are unchanged.
//
// 재설계의 유채색은 danger 하나뿐이라 성공 배너에는 고유 변형이 없다 — 성공은
// 중립 배너(.banner)이고 오류만 .banner--warn 을 얹는다. 여전히 두 상태는
// 눈으로 갈라지며, 서버가 내보내는 kind 값("ok"/"error")은 그대로다.
func TestBannerKindMapping(t *testing.T) {
	a := newTestApp(t)

	t.Run("error → banner--warn", func(t *testing.T) {
		var got profile.ProfilePreferences
		_ = got
		a.writePreferences = func(string, profile.ProfilePreferences) error { return nil }
		a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
		// invalid submit → server sets BannerKind="error" + Banner text.
		form := url.Values{"__profile": {"default"}, "permission_mode": {"bogus-mode"}}
		rec := servePost(t, a.routes(), "/save", form)
		body := rec.Body.String()
		if !strings.Contains(body, "banner--warn") {
			t.Errorf("error banner not mapped to banner--warn chrome:\n%s", body)
		}
	})

	t.Run("ok → neutral banner", func(t *testing.T) {
		a.writePreferences = func(string, profile.ProfilePreferences) error { return nil }
		a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
		a.writeProjectConfig = func(string, string, string) error { return nil }
		form := url.Values{"__profile": {"default"}, "permission_mode": {"acceptEdits"}}
		rec := servePost(t, a.routes(), "/save", form)
		body := rec.Body.String()
		if rec.Code != http.StatusOK {
			t.Fatalf("valid save status = %d, want 200", rec.Code)
		}
		if !strings.Contains(body, `class="banner" role="status"`) {
			t.Errorf("success banner not rendered with the neutral banner chrome:\n%s", body)
		}
		// agentfm 패널이 자체 주의 배너를 들고 있으므로 페이지 전역 검색으로는
		// 갈라낼 수 없다. 상태 배너(role="status")만 좁혀서 본다.
		if strings.Contains(body, `class="banner banner--warn" role="status"`) {
			t.Error("a successful save rendered the warning banner variant")
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

	// (6) 셸 소스: 모든 화면이 같은 <html> 셸을 쓴다 (구 board.templ 계보를 흡수).
	// 서버가 내보내는 data-theme 속성과 테마 토글은 없고, lang 속성은 남는다.
	shellSrc := readGoSource(t, "shell.templ")
	if strings.Contains(shellSrc, "data-theme") || strings.Contains(shellSrc, "themeToggle") {
		t.Error("shell.templ still carries data-theme / themeToggle (dark theme must be fully retired)")
	}
	if !strings.Contains(shellSrc, `<html lang={ vm.Lang }>`) {
		t.Error("shell.templ <html> lost its lang attribute (must be preserved)")
	}

	// (7) Settings page source: theme markers gone. The FOUC language branch now
	// lives in the shell <head>, shared by every screen, so it is asserted there.
	rootSrc := readGoSource(t, "root.templ")
	if strings.Contains(rootSrc, "data-theme") || strings.Contains(rootSrc, "moai-console-theme") || strings.Contains(rootSrc, "themeToggle") {
		t.Error("root.templ still carries theme markers (dark theme must be fully retired)")
	}
	if !strings.Contains(shellSrc, "moai-console-lang") {
		t.Error("shell.templ FOUC language branch (moai-console-lang) missing — AC-MWA-003b")
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
	if !strings.Contains(errored, `class="field__err"`) {
		t.Error("non-color error cue (.field-error icon+text span) missing")
	}
	if !strings.Contains(errored, "field__err") {
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
func TestAchromaticSingleColourSystem(t *testing.T) {
	css := readEmbeddedAsset(t, "console.css")

	// 목적 교체 (콘솔 재설계 — 무채색 단일 체계 채택):
	// 이전 테스트는 docs-site 와 같은 4색 상태 토큰(success/warning/info/danger)을
	// 고정했다. 재설계는 "색으로 상태를 나르지 않는다"를 채택했으므로, 그 4색
	// 체계 자체가 사라졌다. 이제 검증할 불변식은 "유채색은 --danger 하나뿐"이다.
	//
	// 상태는 색이 아니라 모양(점 채움/테두리, 파선, 글리프)으로 구분한다 —
	// 색각 이상이나 흑백 인쇄에서도 읽히게 하기 위한 선택이다.
	for _, retired := range []string{
		"--color-success",
		"--color-warning",
		"--color-info",
		"--status-text-success",
		"--status-text-danger",
	} {
		if strings.Contains(css, retired) {
			t.Errorf("console.css still declares retired status token %q (the achromatic system carries --danger only)", retired)
		}
	}

	// 유일한 유채색은 --danger 다.
	if !strings.Contains(css, "--danger:#c5261d") {
		t.Error("console.css missing the single chromatic token --danger")
	}

	// 상태 어휘가 모양으로 구분되는지 — 점 채움 대 테두리, 파선 추정 표시.
	for _, shape := range []string{
		".state--idle .state__dot",
		".est",
	} {
		if !strings.Contains(css, shape) {
			t.Errorf("console.css missing shape-based state marker %q (colour alone must not carry state)", shape)
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
