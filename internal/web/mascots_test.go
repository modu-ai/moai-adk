package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// SPEC-DESIGN-MOAIWEBV2-001 M2 — mascot 6-pose library.
//
// The console mascot library adopts the v2 canonical 6-pose set
// (Coffee/Explaining/Pointing/Searching/Teaching/Thinking) under lowercase-kebab
// filenames, embedded via the assets/mascots glob (assets.go) and served under
// /static/mascots/. The header brand badge is repointed to mascot-thinking.png;
// the earlier mascot-coding.png / mascot-talking.png assets are removed.

// mascotV2Poses is the v2 canonical 6-pose set (REQ-MWV2-010/014).
var mascotV2Poses = []string{"coffee", "explaining", "pointing", "searching", "teaching", "thinking"}

// TestMascotPosesEmbeddedAndServed verifies AC-MWV2-010/011: the 6 lowercase-kebab
// pose files are embedded (assets/mascots glob) and served at
// /static/mascots/mascot-<pose>.png (200, non-empty body, offline).
func TestMascotPosesEmbeddedAndServed(t *testing.T) {
	a := newTestApp(t)
	for _, pose := range mascotV2Poses {
		name := "mascots/mascot-" + pose + ".png"
		// Embedded under assets/ and reachable via the static FS.
		data := readEmbeddedAsset(t, name)
		if len(data) == 0 {
			t.Errorf("embedded mascot %q is empty", name)
		}
		// Served from /static/mascots/mascot-<pose>.png (200, offline — a static GET
		// is not Host-gated, so a foreign host still reaches it).
		path := "/static/" + name
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Host = "evil.example.com"
		rec := httptest.NewRecorder()
		a.routes().ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", path, rec.Code)
		}
		if rec.Body.Len() == 0 {
			t.Errorf("%s served body is empty", path)
		}
	}
}

// TestMascotRetiredAssetsRemoved verifies AC-MWV2-012/013: the retired
// mascot-coding.png (replaced by mascot-thinking.png) and mascot-talking.png
// (0 references) are gone from the embed — a GET returns 404.
func TestMascotRetiredAssetsRemoved(t *testing.T) {
	a := newTestApp(t)
	for _, retired := range []string{"mascot-coding.png", "mascot-talking.png"} {
		path := "/static/mascots/" + retired
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		a.routes().ServeHTTP(rec, req)
		if rec.Code == http.StatusOK {
			t.Errorf("retired asset %s still served (status 200) — must be removed", path)
		}
	}
}

// TestRailBrandIsInlineMark: 재설계 셸의 브랜드 자리는 마스코트 PNG 가 아니라
// 인라인 SVG 로고 마크다. 마스코트는 화면 장식으로 남지 않는다 — 무채색 단일
// 체계에서 컬러 PNG 는 유일한 유채색 규칙(danger 하나)을 깨기 때문이다.
//
// AC-MWV2-012 의 원래 의도(브랜드 자리가 폐기된 mascot-coding.png 를 참조하지
// 않는다)는 그대로 지킨다.
func TestRailBrandIsInlineMark(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{UserName: "test"})
	if !strings.Contains(body, `class="rail__logo"`) {
		t.Error("the rail brand mark is missing")
	}
	if strings.Contains(body, "/static/mascots/") {
		t.Error("the settings screen still references a mascot image — the redesigned shell uses the inline brand mark")
	}
}
