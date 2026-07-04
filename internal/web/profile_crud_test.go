package web

// SPEC-WEB-CONSOLE-011 M4 — profile CRUD (create / switch / delete) tests.
// AC-WC11-032 (전 흐름), AC-WC11-033 (default/active delete 거부), AC-WC11-034
// (불법 이름 4xx + no MkdirAll), AC-WC11-061 (신규 i18n 키 4-locale parity).

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

// crudApp builds an app whose profile store is isolated under a temp dir and
// whose CLAUDE_CONFIG_DIR is neutralized so profile.GetCurrentName() == "default"
// deterministically. Not parallel-safe (mutates process globals).
func crudApp(t *testing.T, profileName string) (*app, string) {
	t.Helper()
	profileBase := t.TempDir()
	orig := profile.BaseDirOverride
	profile.BaseDirOverride = profileBase
	t.Cleanup(func() { profile.BaseDirOverride = orig })
	t.Setenv(envClaudeConfigDir, "") // GetCurrentName → "default"

	a := newApp(Config{ProjectRoot: t.TempDir(), ProfileName: profileName})
	return a, profileBase
}

// envClaudeConfigDir mirrors config.EnvClaudeConfigDir for the test without
// importing internal/config here.
const envClaudeConfigDir = "CLAUDE_CONFIG_DIR"

func postProfile(t *testing.T, h http.Handler, path, name string) *httptest.ResponseRecorder {
	t.Helper()
	form := url.Values{"profile_name": {name}}
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "127.0.0.1"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// TestProfileCRUDFlow covers AC-WC11-032: create → list → switch → delete.
func TestProfileCRUDFlow(t *testing.T) {
	a, base := crudApp(t, "default")
	h := a.routes()

	// --- CREATE ---
	rec := postProfile(t, h, "/profile/create", "work")
	if rec.Code != http.StatusOK {
		t.Fatalf("create status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(base, "work")); err != nil {
		t.Fatalf("create did not make the profile dir: %v", err)
	}

	// --- LIST (GET /) reflects the new profile ---
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "work") {
		t.Fatalf("GET / did not list 'work' (status %d)", rec.Code)
	}

	// --- SWITCH (GET /?profile=work reuses the existing load path) ---
	req = httptest.NewRequest(http.MethodGet, "/?profile=work", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("switch GET /?profile=work status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "selected: <strong>work</strong>") &&
		!strings.Contains(rec.Body.String(), "profilemgr__name--active") {
		t.Errorf("switch did not select 'work' in the view")
	}

	// --- DELETE (work is not active → allowed) ---
	rec = postProfile(t, h, "/profile/delete", "work")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(base, "work")); !os.IsNotExist(err) {
		t.Errorf("delete did not remove the profile dir (err=%v)", err)
	}
}

// TestProfileDeleteGuards covers AC-WC11-033: default and the currently active
// profile are refused 4xx and survive.
func TestProfileDeleteGuards(t *testing.T) {
	a, base := crudApp(t, "keepme")
	h := a.routes()

	// Seed an active profile "keepme" (cfg.ProfileName = keepme).
	if err := os.MkdirAll(filepath.Join(base, "keepme"), 0o755); err != nil {
		t.Fatal(err)
	}

	// default → 4xx, still conceptually present (it is the base store, never removed).
	rec := postProfile(t, h, "/profile/delete", "default")
	if rec.Code < 400 || rec.Code >= 500 {
		t.Errorf("delete default status = %d, want 4xx", rec.Code)
	}

	// active profile (keepme == cfg.ProfileName) → 4xx + dir survives.
	rec = postProfile(t, h, "/profile/delete", "keepme")
	if rec.Code < 400 || rec.Code >= 500 {
		t.Errorf("delete active 'keepme' status = %d, want 4xx", rec.Code)
	}
	if _, err := os.Stat(filepath.Join(base, "keepme")); err != nil {
		t.Errorf("active profile 'keepme' was removed despite guard: %v", err)
	}
}

// TestProfileCreateInvalidName covers AC-WC11-034 (CRUD side): an illegal/empty/
// reserved name is rejected 4xx with NO directory side effect.
func TestProfileCreateInvalidName(t *testing.T) {
	a, base := crudApp(t, "default")
	h := a.routes()

	// The escape target for a `../../` traversal from <base>.
	// base is a t.TempDir() leaf, so `../<sentinel>` lands in its parent (still
	// under the test's managed temp tree — auto-cleaned).
	cases := []string{"", "   ", "default", "../evil", "a/b", "a\\b", "..", ".hidden"}
	for _, name := range cases {
		rec := postProfile(t, h, "/profile/create", name)
		if rec.Code < 400 || rec.Code >= 500 {
			t.Errorf("create %q status = %d, want 4xx", name, rec.Code)
		}
	}

	// No stray directory: only the base itself exists, no traversal escape.
	escaped := filepath.Join(filepath.Dir(base), "evil")
	if _, err := os.Stat(escaped); err == nil {
		t.Errorf("create traversal made a dir outside the base: %s", escaped)
	}
}

// TestProfileCRUDI18nKeys covers AC-WC11-061 for the M4 CRUD strings: every new
// data-i18n key exists in all 4 locales (en/ko/ja/zh).
func TestProfileCRUDI18nKeys(t *testing.T) {
	keys := []string{
		"profile.manage.title",
		"profile.manage.desc",
		"profile.current",
		"profile.create.label",
		"profile.create.button",
		"profile.delete.button",
	}
	for _, k := range keys {
		if !i18nKeyInAllLocales(t, k) {
			t.Errorf("i18n.js missing CRUD key %q in all 4 locales", k)
		}
	}
}

// TestProfileCRUDMethodNotAllowed: non-POST on the CRUD routes → 405.
func TestProfileCRUDMethodNotAllowed(t *testing.T) {
	a, _ := crudApp(t, "default")
	h := a.routes()
	for _, path := range []string{"/profile/create", "/profile/delete"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("GET %s status = %d, want 405", path, rec.Code)
		}
	}
}

// TestProfileDeleteNonexistent: deleting a valid, non-active, non-existent
// profile surfaces profile.Delete's error as a 4xx (readable, never blank).
func TestProfileDeleteNonexistent(t *testing.T) {
	a, _ := crudApp(t, "default")
	h := a.routes()
	rec := postProfile(t, h, "/profile/delete", "ghost")
	if rec.Code < 400 || rec.Code >= 500 {
		t.Errorf("delete nonexistent 'ghost' status = %d, want 4xx", rec.Code)
	}
}

// TestProfileDeleteInvalidName: a traversal name that clears the default/active
// guards is still rejected 4xx by the isValidProfileName check (REQ-WC11-034 —
// delete route).
func TestProfileDeleteInvalidName(t *testing.T) {
	a, _ := crudApp(t, "default")
	h := a.routes()
	for _, name := range []string{"a/b", "a\\b", ".hidden"} {
		rec := postProfile(t, h, "/profile/delete", name)
		if rec.Code < 400 || rec.Code >= 500 {
			t.Errorf("delete %q status = %d, want 4xx", name, rec.Code)
		}
	}
}

// TestProfileCRUDSeamErrors: an injected create/delete seam failure surfaces a
// 5xx / 4xx error page (covers the seam-error branches).
func TestProfileCRUDSeamErrors(t *testing.T) {
	a, base := crudApp(t, "default")
	a.createProfile = func(string) error { return os.ErrPermission }
	rec := postProfile(t, a.routes(), "/profile/create", "work")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("create seam error status = %d, want 500", rec.Code)
	}

	// A delete-seam failure on an existing, non-active profile → 4xx.
	if err := os.MkdirAll(filepath.Join(base, "gone"), 0o755); err != nil {
		t.Fatal(err)
	}
	a.deleteProfile = func(string) error { return os.ErrPermission }
	rec = postProfile(t, a.routes(), "/profile/delete", "gone")
	if rec.Code < 400 || rec.Code >= 500 {
		t.Errorf("delete seam error status = %d, want 4xx", rec.Code)
	}
}

// TestActiveProfileNameFallback: an empty cfg.ProfileName resolves to "default".
func TestActiveProfileNameFallback(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir(), ProfileName: ""})
	if got := a.activeProfileName(); got != "default" {
		t.Errorf("activeProfileName() = %q, want default", got)
	}
	b := newApp(Config{ProjectRoot: t.TempDir(), ProfileName: "work"})
	if got := b.activeProfileName(); got != "work" {
		t.Errorf("activeProfileName() = %q, want work", got)
	}
}

// TestCreateProfileDirRejectsReserved: the default create seam rejects reserved
// and traversal names without creating a directory.
func TestCreateProfileDirRejectsReserved(t *testing.T) {
	profileBase := t.TempDir()
	orig := profile.BaseDirOverride
	profile.BaseDirOverride = profileBase
	t.Cleanup(func() { profile.BaseDirOverride = orig })

	for _, name := range []string{"", "default", "../x", "a/b", ".hidden"} {
		if err := createProfileDir(name); err == nil {
			t.Errorf("createProfileDir(%q) = nil, want error", name)
		}
	}
	// A valid name creates the directory.
	if err := createProfileDir("ok"); err != nil {
		t.Errorf("createProfileDir(ok) = %v, want nil", err)
	}
	if _, err := os.Stat(filepath.Join(profileBase, "ok")); err != nil {
		t.Errorf("createProfileDir(ok) did not create the dir: %v", err)
	}
}

// TestProfileResultDegradesOnBadSelection: when the request's own ?profile
// selection is itself a traversal name, the error re-render degrades to the
// minimal error page (never feeds a traversal name into the read seam).
func TestProfileResultDegradesOnBadSelection(t *testing.T) {
	a, _ := crudApp(t, "default")
	// profile_name empty → 4xx error path → renderProfileResult; ?profile=../bad
	// makes selectedProfile return a traversal name → renderError degrade.
	form := url.Values{"profile_name": {""}}
	req := httptest.NewRequest(http.MethodPost, "/profile/create?profile=../bad", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "127.0.0.1"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code < 400 || rec.Code >= 500 {
		t.Errorf("bad-selection error path status = %d, want 4xx", rec.Code)
	}
}

// TestProfileResultDegradesOnReadError: a read-seam failure while re-rendering a
// CRUD result degrades gracefully (error path → 4xx minimal page; success path →
// 500), never panicking.
func TestProfileResultDegradesOnReadError(t *testing.T) {
	a, _ := crudApp(t, "default")
	a.readProjectConfig = func(string) (string, string, error) {
		return "", "", os.ErrPermission
	}
	// Error path: invalid name → renderProfileResult → buildIndexView errors → 4xx.
	rec := postProfile(t, a.routes(), "/profile/create", "")
	if rec.Code < 400 || rec.Code >= 500 {
		t.Errorf("read-error error path status = %d, want 4xx", rec.Code)
	}
	// Success path: valid create → renderProfileSuccess → buildIndexView errors → 500.
	rec = postProfile(t, a.routes(), "/profile/create", "work")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("read-error success path status = %d, want 500", rec.Code)
	}
}

// TestProfileManagerRendered verifies the CRUD manager renders on GET / with its
// create form + i18n-keyed controls (AC-WC11-032 render half).
func TestProfileManagerRendered(t *testing.T) {
	body := renderConsolePage(t)
	for _, marker := range []string{
		`id="profile-manager"`,
		`action="/profile/create"`,
		`name="profile_name"`,
		`data-i18n="profile.create.button"`,
		`data-i18n="profile.manage.title"`,
	} {
		if !strings.Contains(body, marker) {
			t.Errorf("rendered page missing profile-manager marker %q", marker)
		}
	}
}
