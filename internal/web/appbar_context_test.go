package web

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestShellContextChips: 셸이 지금 무엇을 편집 중인지 — 프로젝트 · 프로필 ·
// 대화 언어 · 모델 · effort · 개발 모드 — 를 화면 위쪽에 보여 준다.
//
// 재설계 이전에는 이 정보가 가로 appbar 의 배지 두 개(📁project / 👤profile)와
// 한 줄짜리 문맥 텍스트였다. 그 문맥 줄은 **하드코딩된 리터럴**이라
// (`lang: ko · model: claude-opus-5 · effort: high · dev: ddd`) 프로필 값이
// 무엇이든 같은 문자열을 냈다. 지금은 뷰모델에서 실제 값을 읽어 상단바의
// chip-kv 쌍으로 낸다. 그래서 이 테스트는 예전 리터럴과 **다른** 값을 넣어,
// 하드코딩이면 통과할 수 없게 만든다.
func TestShellContextChips(t *testing.T) {
	a := newTestApp(t)

	testProjectRoot := "/Users/test/moai-adk-go"
	a.cfg.ProjectRoot = testProjectRoot

	a.readPreferences = func(name string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{
			ConversationLang: "ja",
			Model:            "sonnet",
			EffortLevel:      "low",
		}, nil
	}
	a.listProfiles = func() []profile.ProfileEntry {
		return []profile.ProfileEntry{
			{Name: "default", Current: true},
			{Name: "work", Current: false},
		}
	}
	a.readProjectConfig = func(projectRoot string) (devMode, convention string, err error) {
		return "tdd", "conventional-commits", nil
	}

	rec := serveGet(t, a.routes(), "/settings")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /settings status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()

	// (1) 프로젝트 이름은 레일 하단의 문맥 버튼에, 전체 경로는 그 툴팁에 있다.
	projectName := filepath.Base(testProjectRoot)
	if !strings.Contains(body, `<span class="ctx-btn__v" title="`+testProjectRoot+`">`+projectName+`</span>`) {
		t.Errorf("project name/path not rendered in the rail context button (want name %q, tooltip %q)", projectName, testProjectRoot)
	}

	// (2) 편집 중인 프로필도 레일에 있고, 세션 시작 시 활성인 프로필은 팝오버에서
	// "(current)" 로 구분된다.
	if !strings.Contains(body, `<span class="ctx-btn__v">default</span>`) {
		t.Error("the profile being edited is not rendered in the rail context button")
	}
	if !strings.Contains(body, `default (current)`) {
		t.Error("the launch-active profile is not marked in the profile popover")
	}

	// (3) 문맥 칩 4종이 실제 값으로 렌더된다. 예전 하드코딩 리터럴과 다른 값을
	// 넣었으므로, 하나라도 리터럴로 되돌아가면 여기서 잡힌다.
	for _, want := range []struct{ key, value string }{
		{"lang", "ja"},
		{"model", "sonnet"},
		{"effort", "low"},
		{"dev", "tdd"},
	} {
		chip := `<span class="chip-kv__k">` + want.key + `</span> <span class="chip-kv__v">` + want.value + `</span>`
		if !strings.Contains(body, chip) {
			t.Errorf("context chip %s=%s not rendered from the view model", want.key, want.value)
		}
	}

	// (4) 폐기된 하드코딩 문맥 줄이 되살아나지 않았는지 확인한다.
	if strings.Contains(body, "lang: ko · model: claude-opus-5") {
		t.Error("the hardcoded appbar context line is back — the chips must come from the view model")
	}
}

// TestShellContextChipsOmitEmptyValues: 값이 없는 항목은 칩 자체를 만들지
// 않는다. 빈 칩은 "설정되지 않음"이 아니라 "값이 빈 문자열"로 읽히고, 둘은
// 사용자가 해야 할 일이 다르다.
func TestShellContextChipsOmitEmptyValues(t *testing.T) {
	a := newTestApp(t)
	a.readPreferences = func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{ConversationLang: "ko"}, nil
	}
	a.readProjectConfig = func(string) (string, string, error) { return "", "", nil }

	body := serveGet(t, a.routes(), "/settings").Body.String()

	if !strings.Contains(body, `<span class="chip-kv__k">lang</span>`) {
		t.Error("a chip with a value was omitted")
	}
	for _, absent := range []string{"model", "effort", "dev"} {
		if strings.Contains(body, `<span class="chip-kv__k">`+absent+`</span>`) {
			t.Errorf("chip %q rendered with no value — an empty chip reads as a blank setting", absent)
		}
	}
}
