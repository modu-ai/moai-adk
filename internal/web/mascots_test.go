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

// TestHeaderBrandBadgeUsesThinkingPose verifies AC-MWV2-012: the settings page
// header brand badge references mascot-thinking.png (the confirmed v2 header pose)
// and no longer references the removed mascot-coding.png.
func TestHeaderBrandBadgeUsesThinkingPose(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{UserName: "test"})
	if !strings.Contains(body, "/static/mascots/mascot-thinking.png") {
		t.Error("header brand badge missing the mascot-thinking.png reference")
	}
	if strings.Contains(body, "/static/mascots/mascot-coding.png") {
		t.Error("header brand badge still references the removed mascot-coding.png")
	}
}
