package web

// SPEC-WEB-CONSOLE-011 M4 (REQ-WC11-030/031/034) — profile-name 검증 갭
// reproduction test. 웹 쓰기 경로(handleSave)가 `?profile=` / `__profile` 값을
// isValidProfileName 검증 없이 profile.WritePreferences 로 흘려보내
// (preferences.go GetPreferencesPath/WritePreferences 는 MkdirAll 로 암묵 생성),
// path traversal (`__profile=../../x`) 로 profile store 밖 디렉터리를 만들 수
// 있다는 가설을 기계적으로 검증한다.
//
// verification-claim-integrity §1.1 surface 3: 본 결함은 repro test 가 green
// 이 될 때까지 UNVERIFIED HYPOTHESIS 다. RED 상태(수정 전)에서 이 테스트는
// FAIL 해야 하며, GREEN(수정 후) PASS 로 전환된다.

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestProfileNameTraversal drives the REAL write seam (profile.WritePreferences)
// through the web POST /save handler with a traversal `__profile` value and
// asserts the request is rejected 4xx AND no directory is created outside the
// isolated profile store. The profile base is nested two levels deep under
// t.TempDir() so a `../../` escape lands INSIDE t.TempDir() (auto-cleaned) but
// OUTSIDE the profile base — never polluting the real filesystem.
func TestProfileNameTraversal(t *testing.T) {
	tmp := t.TempDir()
	// base = <tmp>/sub1/sub2/profiles → `../../escaped` resolves to <tmp>/sub1/escaped.
	base := filepath.Join(tmp, "sub1", "sub2", "profiles")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir base: %v", err)
	}

	orig := profile.BaseDirOverride
	profile.BaseDirOverride = base
	t.Cleanup(func() { profile.BaseDirOverride = orig })

	a := newApp(Config{ProjectRoot: t.TempDir(), ProfileName: "default"})
	h := a.routes()

	// The escape target for `../../<sentinel>` from base.
	const sentinel = "moai-repro-escaped"
	escaped := filepath.Join(tmp, "sub1", sentinel)

	cases := []struct {
		name    string
		profile string
	}{
		{"parent-traversal", "../../" + sentinel},
		{"slash-name", "a/" + sentinel},
		{"backslash-name", "a\\" + sentinel},
		{"dotdot", ".."},
		{"dot-prefix", "." + sentinel},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			form := url.Values{
				"__profile":       {tc.profile},
				"permission_mode": {""}, // valid (empty) — isolate the profile-name path
			}
			req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Host = "127.0.0.1" // loopback so hostCheckMiddleware does not 403 for the wrong reason
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code < 400 || rec.Code >= 500 {
				t.Errorf("POST /save with __profile=%q status = %d, want 4xx", tc.profile, rec.Code)
			}
		})
	}

	// No MkdirAll side effect: the escaped directory must NOT exist after all
	// rejected requests (REQ-WC11-034 — no implicit directory creation).
	if _, err := os.Stat(escaped); err == nil {
		t.Errorf("traversal created a directory outside the profile store: %s", escaped)
	} else if !os.IsNotExist(err) {
		t.Errorf("unexpected stat error for %s: %v", escaped, err)
	}
}
