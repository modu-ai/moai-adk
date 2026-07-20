package web

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// 버그 A 수정(app.js htmx:afterSettle 재초기화)의 소스-수준 계약 테스트.
// app.js 는 vanilla JS 라 브라우저 없이 단위 테스트하기 어렵다 — 대신 임베드된
// app.js 소스가 (1) initConsole 명명 함수로 초기화를 추출했고, (2) DOMContentLoaded
// 와 htmx:afterSettle 양쪽에 바인딩했는지 검증한다. 실제 swap-후 재초기화 동작은
// 수동 E2E(§E Gaps)에서 검증한다. (구 readPersistedTheme helper 계약은 다크 테마
// 폐지와 함께 제거 — SPEC-DESIGN-MOAIWEBV2-002 REQ-MWA-003; 테마 부재 자체는
// TestDarkThemeAbsence 가 검증한다.)

// TestAppJsInitConsoleBoundToBothEvents verifies Bug A fix: the embedded app.js
// binds the named initConsole initializer to BOTH DOMContentLoaded and
// htmx:afterSettle, so a boost body swap re-runs i18n re-application and
// re-wires the langpick listener that the new (listener-less) DOM lost.
func TestAppJsInitConsoleBoundToBothEvents(t *testing.T) {
	t.Parallel()
	src := readEmbeddedAsset(t, "app.js")

	// initConsole 명명 함수로 추출되어야 한다(익명 DOMContentLoaded 핸들러가 아님).
	if !strings.Contains(src, "function initConsole(") {
		t.Error("app.js missing the named `function initConsole(` initializer (DOMContentLoaded handler body must be extracted into it)")
	}

	// 두 이벤트 모두 initConsole 에 바인딩되어야 한다.
	if !strings.Contains(src, `addEventListener("DOMContentLoaded", initConsole)`) {
		t.Error("app.js must bind initConsole to DOMContentLoaded")
	}
	if !strings.Contains(src, `addEventListener("htmx:afterSettle", initConsole)`) {
		t.Error("app.js must bind initConsole to htmx:afterSettle so a boost body swap re-initializes the console")
	}
}

// TestAppJsHxBoostPreserved verifies Bug A constraint: hx-boost="true" stays on
// the rendered form (the fix is app.js re-initialization, NOT disabling boost).
// This guards against a regression that removes boost to "fix" the swap bug.
func TestAppJsHxBoostPreserved(t *testing.T) {
	t.Parallel()
	body := renderIndexBody(t, profile.ProfilePreferences{})
	if !strings.Contains(body, `hx-boost="true"`) {
		t.Error("form must keep hx-boost=\"true\" — Bug A fix is app.js re-init on htmx:afterSettle, NOT removing boost (SPEC-WEB-CONSOLE-006)")
	}
}
