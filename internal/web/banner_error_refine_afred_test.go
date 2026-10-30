//go:build afred

// banner_error_refine_afred_test.go: B14 RED, committed before the repair. Error
// banners and error pages are built from raw error text: renderPage's "render
// failed", the form-parse failures, the settings-save failure banner
// (handleSave → renderErrorPage) and the profile CRUD failure banners
// (renderProfileResult). renderError only HTML-escapes, which is not
// refining. Excluded from the default run by the afred build tag; run it with:
//
//	go test -tags afred -run '^TestErrorBannersAreRefined$' ./internal/web/
//
// One value of each of five categories — e-mail, home path, token, telephone
// number, panic stack — goes into the error of each path; none may appear in the
// response. Sentinel on failure: AF_BANNER_REFINE.
package web

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/a-h/templ"

	"github.com/modu-ai/moai-adk/internal/profile"
)

var bannerPayloads = map[string]string{
	"email":     "alice.private@example.com",
	"home-path": "/Users/alice/secret-project/notes",
	"token":     "ghp_" + strings.Repeat("Z", 36),
	"phone":     "010-1234-5678",
	"stack":     "goroutine 1 [running]:\nmain.main()\n\t/Users/alice/work/main.go:10 +0x1d",
}

func TestErrorBannersAreRefined(t *testing.T) {
	cases := map[string]func(t *testing.T, payload string) string{
		"renderPage render failure": func(t *testing.T, p string) string {
			a := newApp(Config{ProjectRoot: t.TempDir()})
			rec := httptest.NewRecorder()
			a.renderPage(rec, templ.ComponentFunc(func(context.Context, io.Writer) error { return errors.New("draw " + p) }))
			return rec.Body.String()
		},
		"renderError direct": func(t *testing.T, p string) string {
			a := newApp(Config{ProjectRoot: t.TempDir()})
			rec := httptest.NewRecorder()
			a.renderError(rec, http.StatusInternalServerError, "failed: "+p)
			return rec.Body.String()
		},
		"settings save failure banner": func(t *testing.T, p string) string {
			a := newTestApp(t)
			var calls []saveStep
			recordingSeams(a, &calls, "")
			called := false
			a.writePreferences = func(string, profile.ProfilePreferences) error { called = true; return errors.New("write " + p) }
			body := servePost(t, a.routes(), "/save", reproForm()).Body.String()
			if !called {
				t.Fatal("AF_BANNER_REFINE: the write seam was never reached, so this case proves nothing")
			}
			return body
		},
		"profile create failure banner": func(t *testing.T, p string) string {
			a := newTestApp(t)
			called := false
			a.createProfile = func(string) error { called = true; return errors.New("mkdir " + p) }
			body := servePost(t, a.routes(), "/profile/create", url.Values{"profile_name": {"fresh"}}).Body.String()
			if !called {
				t.Fatal("AF_BANNER_REFINE: the create seam was never reached, so this case proves nothing")
			}
			return body
		},
	}
	for name, run := range cases {
		for cat, payload := range bannerPayloads {
			t.Run(name+"/"+cat, func(t *testing.T) {
				body := run(t, payload)
				if body == "" {
					t.Fatalf("AF_BANNER_REFINE: %s produced no body", name)
				}
				if strings.Contains(body, payload) || strings.Contains(body, "alice") && cat != "phone" {
					t.Errorf("AF_BANNER_REFINE: the %s reached the response of %s", cat, name)
				}
			})
		}
	}
}
