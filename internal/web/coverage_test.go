package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestServer_HandlerAccessor verifies Handler() returns the wired handler so
// httptest-based integration callers can drive it directly.
func TestServer_HandlerAccessor(t *testing.T) {
	srv, err := NewServer(Config{ProjectRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if srv.Handler() == nil {
		t.Fatal("Handler() returned nil")
	}
	// The handler routes GET / without error.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Handler GET / status = %d, want 200", rec.Code)
	}
}

// TestOpenDefaultBrowser exercises the cross-platform browser opener. The
// command may not exist in the sandbox; in that case a LookPath error is
// expected and acceptable (REQ-WC-004 treats it as non-fatal). We only assert
// it does not panic and returns within the contract.
func TestOpenDefaultBrowser(t *testing.T) {
	// A clearly non-routable URL — we never want a real browser to launch in CI.
	err := openDefaultBrowser("http://127.0.0.1:1/")
	// Either the opener command was missing (LookPath error) or the command
	// started successfully. Both are valid outcomes; the contract is "do not
	// panic, return an error or nil". We assert no panic by reaching here.
	_ = err
}

// TestHandleIndex_NotFoundPath verifies a non-root path under "/" returns 404
// (the handleIndex guard).
func TestHandleIndex_NotFoundPath(t *testing.T) {
	a := newTestApp(t)
	a.readPreferences = func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, nil
	}
	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("unknown path status = %d, want 404", rec.Code)
	}
}

// TestHandleSave_MethodNotAllowed verifies a GET to /save is rejected 405.
func TestHandleSave_MethodNotAllowed(t *testing.T) {
	a := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/save", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /save status = %d, want 405", rec.Code)
	}
}

// TestRender_NilTemplateSurfacesError verifies the render path surfaces a
// readable error when the template failed to parse (REQ-WC-010).
func TestRender_NilTemplateSurfacesError(t *testing.T) {
	a := newTestApp(t)
	a.tmpl = nil // simulate a template parse failure
	rec := httptest.NewRecorder()
	a.render(rec, http.StatusOK, a.newPageView(profile.ProfilePreferences{}, "default"))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("nil-template render status = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "template unavailable") {
		t.Errorf("nil-template render did not surface a readable error:\n%s", rec.Body.String())
	}
}

// TestValidatePrefs_AllInvalidFields exercises every validation branch:
// invalid permission mode, all four languages, and all three statusline fields.
func TestValidatePrefs_AllInvalidFields(t *testing.T) {
	errs := validatePrefs(profile.ProfilePreferences{
		PermissionMode:   "nope",
		ConversationLang: "xx",
		GitCommitLang:    "yy",
		CodeCommentLang:  "zz",
		DocLang:          "qq",
		StatuslineMode:   "weird",
		StatuslinePreset: "weird",
		StatuslineTheme:  "weird",
	})
	for _, field := range []string{
		"permission_mode", "conversation_lang", "git_commit_lang",
		"code_comment_lang", "doc_lang", "statusline_mode",
		"statusline_preset", "statusline_theme",
	} {
		if _, ok := errs[field]; !ok {
			t.Errorf("expected validation error for %q, got none", field)
		}
	}
}

// TestValidatePrefs_AllValidEmpty verifies empty values pass (they mean "unset",
// mirroring SyncToProjectConfig's non-empty-only overwrite semantics).
func TestValidatePrefs_AllValidEmpty(t *testing.T) {
	errs := validatePrefs(profile.ProfilePreferences{PermissionMode: "acceptEdits"})
	if len(errs) != 0 {
		t.Errorf("empty-value prefs produced errors: %v", errs)
	}
}

// TestPageTemplateParses verifies the embedded page template parses cleanly,
// exercising the funcmap construction (including the dict helper registration).
func TestPageTemplateParses(t *testing.T) {
	tmpl, err := pageTemplate()
	if err != nil {
		t.Fatalf("pageTemplate: %v", err)
	}
	if tmpl == nil {
		t.Fatal("pageTemplate returned nil template")
	}
	if tmpl.Lookup("langSelect") == nil {
		t.Error("nested langSelect template not defined")
	}
}

// TestIsLoopbackHost covers the host-classification branches directly.
func TestIsLoopbackHost(t *testing.T) {
	cases := map[string]bool{
		"127.0.0.1:8080":  true,
		"127.0.0.1":       true,
		"localhost:8080":  true,
		"localhost":       true,
		"[::1]:8080":      true,
		"::1":             true,
		"0.0.0.0:8080":    false,
		"10.0.0.5:8080":   false,
		"example.com:80":  false,
		"":                false,
		"not a host:port": false,
	}
	for host, want := range cases {
		if got := isLoopbackHost(host); got != want {
			t.Errorf("isLoopbackHost(%q) = %v, want %v", host, got, want)
		}
	}
}
