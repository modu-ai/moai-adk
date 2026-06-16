package web

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// SPEC-WEB-CONSOLE-006 후속 (버그 B 수정): statusline preset <select> 에
// "(unchanged)" 빈 옵션을 추가한다. 사용자가 다른 필드만 바꿔 저장할 때
// statusline_preset 이 빈 값으로 제출되면, sync.go:71 의 빈-skip 경로가
// 기존 statusline.yaml 을 그대로 보존한다(SLR-3 HARD-7). 이 파일은 그 동작의
// RED→GREEN 테스트를 담는다.

// statuslinePresetSelectRe 는 statusline_preset <select> 블록 전체를 잡는 정규식이다.
// 파일 전체 Contains 로 잡으면 statusline_theme select 와 옵션 값이 섞이므로,
// 반드시 id="statusline_preset" <select> 영역으로 scope 한다.
var statuslinePresetSelectRe = regexp.MustCompile(`(?s)<select[^>]*\bid="statusline_preset"[^>]*>.*?</select>`)

// statuslineThemeSelectRe 는 statusline_theme <select> 블록을 잡는다(빈 옵션이
// 없어야 함을 검증하기 위해 theme 쪽도 scope 한다).
var statuslineThemeSelectRe = regexp.MustCompile(`(?s)<select[^>]*\bid="statusline_theme"[^>]*>.*?</select>`)

// TestStatuslinePresetRendersUnchangedEmptyOption verifies Bug B fix part 1:
// the statusline preset <select> renders a leading empty "(unchanged)" option
// (value=""), localizable via data-i18n="opt.unchanged". The user can leave
// statusline untouched by submitting this empty value.
func TestStatuslinePresetRendersUnchangedEmptyOption(t *testing.T) {
	t.Parallel()
	body := renderIndexBody(t, profile.ProfilePreferences{StatuslinePreset: ""})

	presetSel := statuslinePresetSelectRe.FindString(body)
	if presetSel == "" {
		t.Fatal("statusline_preset <select> not found in rendered page")
	}

	// 빈 옵션이 맨 앞에 존재해야 한다. value="" 토큰 + selected/data-i18n 속성이
	// 올 수 있으므로 정규식으로 빈-value <option> 을 잡는다.
	emptyOptRe := regexp.MustCompile(`<option value=""[^>]*>`)
	if !emptyOptRe.MatchString(presetSel) {
		t.Errorf("statusline_preset select missing the leading empty (unchanged) option\nselect:\n%s", presetSel)
	}

	// 빈 옵션 라벨이 i18n 키로 현지화되어야 한다(applyI18n 이 opt.unchanged 를
	// 치환). data-i18n 속성 부여로 다른 옵션들과 일관되게 처리된다.
	if !strings.Contains(presetSel, `data-i18n="opt.unchanged"`) {
		t.Errorf("statusline_preset empty option must carry data-i18n=\"opt.unchanged\" for localization\nselect:\n%s", presetSel)
	}

	// 정규 옵션(full/compact/minimal/custom)은 그대로 존재해야 한다(기존 동작 보존).
	for _, want := range []string{`value="full"`, `value="compact"`, `value="minimal"`, `value="custom"`} {
		if !strings.Contains(presetSel, want) {
			t.Errorf("statusline_preset select missing canonical option %s", want)
		}
	}
}

// TestStatuslineThemeSelectHasNoEmptyOption verifies Bug B scope discipline: only
// the preset select gets the empty "(unchanged)" option — the theme select keeps
// its existing no-empty-option behavior (theme 의 빈 값은 이미 preset/theme/segments
// 게이트에서 무시되므로 별도 보존 로직 불필요).
func TestStatuslineThemeSelectHasNoEmptyOption(t *testing.T) {
	t.Parallel()
	body := renderIndexBody(t, profile.ProfilePreferences{})

	themeSel := statuslineThemeSelectRe.FindString(body)
	if themeSel == "" {
		t.Fatal("statusline_theme <select> not found in rendered page")
	}
	if strings.Contains(themeSel, `<option value="">`) {
		t.Errorf("statusline_theme select must NOT get the empty option (only preset does)\nselect:\n%s", themeSel)
	}
}

// TestSaveEmptyStatuslinePresetSkipsSync verifies Bug B fix part 2: submitting an
// empty statusline_preset leaves StatuslinePreset == "" in the bound
// ProfilePreferences, so sync.go:71 (Preset != "" || Theme != "" || Segments != nil)
// skips statusline sync entirely and the existing statusline.yaml is preserved
// (SLR-3 HARD-7). We assert the empty value reaches the sync seam unchanged.
func TestSaveEmptyStatuslinePresetSkipsSync(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)
	// syncToProject seam 을 spy 로 교체 — 실제 파일시스템을 건드리지 않고
	// 전달된 prefs 의 StatuslinePreset 값을 관찰한다(sync.go:71 게이트의 입력).
	var got profile.ProfilePreferences
	a.writePreferences = func(_ string, prefs profile.ProfilePreferences) error { got = prefs; return nil }
	a.syncToProject = func(_ string, prefs profile.ProfilePreferences) error {
		got = prefs
		return nil
	}

	// statusline_preset 을 빈 채로 제출(= "(unchanged)" 옵션). 다른 필드는
	// validatePrefs 를 통과하는 최소 세트만 채운다.
	form := url.Values{
		"__profile":         {"default"},
		"statusline_preset": {""},
	}
	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("empty-preset save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	if got.StatuslinePreset != "" {
		t.Errorf("empty statusline_preset reached sync seam as %q, want \"\" (sync.go:71 must then skip statusline sync and preserve statusline.yaml)", got.StatuslinePreset)
	}
	if got.StatuslineSegments != nil {
		t.Errorf("empty statusline_preset bound segments = %+v, want nil (no segment materialization on empty preset)", got.StatuslineSegments)
	}
}
