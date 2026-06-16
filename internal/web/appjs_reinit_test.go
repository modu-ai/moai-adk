package web

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// 버그 A 수정(app.js htmx:afterSettle 재초기화)의 소스-수준 계약 테스트.
// app.js 는 vanilla JS 라 브라우저 없이 단위 테스트하기 어렵다 — 대신 임베드된
// app.js 소스가 (1) initConsole 명명 함수로 초기화를 추출했고, (2) DOMContentLoaded
// 와 htmx:afterSettle 양쪽에 바인딩했으며, (3) persisted 테마를 재적용하는
// readPersistedTheme helper 를 가지는지 검증한다. 실제 swap-후 재초기화 동작은
// 수동 E2E(§E Gaps)에서 검증한다.

// TestAppJsInitConsoleBoundToBothEvents verifies Bug A fix: the embedded app.js
// binds the named initConsole initializer to BOTH DOMContentLoaded and
// htmx:afterSettle, so a boost body swap re-runs i18n/theme re-application and
// re-wires the toggle/langpick listeners that the new (listener-less) DOM lost.
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

// TestAppJsReadPersistedThemeHelper verifies Bug A fix: app.js declares a
// readPersistedTheme helper (mirror of readPersistedLang) and initConsole calls
// applyTheme(readPersistedTheme()) so the persisted theme is re-applied after a
// swap without depending on the <head> FOUC script surviving the body swap.
func TestAppJsReadPersistedThemeHelper(t *testing.T) {
	t.Parallel()
	src := readEmbeddedAsset(t, "app.js")

	if !strings.Contains(src, "function readPersistedTheme(") {
		t.Error("app.js missing the `function readPersistedTheme(` helper (mirror of readPersistedLang)")
	}
	// initConsole 이 테마를 재적용해야 한다(applyTheme(readPersistedTheme())).
	if !strings.Contains(src, "applyTheme(readPersistedTheme())") {
		t.Error("app.js initConsole must call applyTheme(readPersistedTheme()) to re-apply the persisted theme after a boost swap")
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
