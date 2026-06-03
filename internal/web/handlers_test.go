package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// newTestApp builds an app with the given config and the canonical option lists,
// allowing tests to override the profile seams before serving.
func newTestApp(t *testing.T) *app {
	t.Helper()
	return newApp(Config{
		ProjectRoot: t.TempDir(),
		ProfileName: "default",
	})
}

// --- Phase 3: embedded assets (AC-WC-005) ---

// TestStaticAssetsServedFromEmbed verifies AC-WC-005 / REQ-WC-005: CSS and JS are
// served from go:embed with the correct Content-Type, no network fetch.
func TestStaticAssetsServedFromEmbed(t *testing.T) {
	a := newTestApp(t)
	h := a.routes()

	cases := []struct {
		path        string
		wantSubstr  string
		wantCTParts []string
	}{
		// SPEC-WEB-CONSOLE-004: style.css → console.css (모두의AI token + component layer).
		{"/static/console.css", "--color-primary", []string{"text/css"}},
		{"/static/app.js", "MoAI Web Console", []string{"javascript"}},
	}
	for _, c := range cases {
		// GET on a static asset must NOT be Host-gated; use a foreign host to
		// prove read assets are reachable regardless.
		req := httptest.NewRequest(http.MethodGet, c.path, nil)
		req.Host = "evil.example.com"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("%s: status = %d, want 200", c.path, rec.Code)
			continue
		}
		body := rec.Body.String()
		if !strings.Contains(body, c.wantSubstr) {
			t.Errorf("%s: body missing %q", c.path, c.wantSubstr)
		}
		ct := rec.Header().Get("Content-Type")
		matched := false
		for _, part := range c.wantCTParts {
			if strings.Contains(ct, part) {
				matched = true
			}
		}
		if !matched {
			t.Errorf("%s: Content-Type = %q, want one of %v", c.path, ct, c.wantCTParts)
		}
	}
}

// --- Phase 4: READ handler (AC-WC-006, AC-WC-010, AC-WC-011) ---

// TestIndexRendersPopulatedForm verifies AC-WC-006 / REQ-WC-006: GET / renders the
// current ProfilePreferences values as a pre-populated editable form.
func TestIndexRendersPopulatedForm(t *testing.T) {
	a := newTestApp(t)
	a.readPreferences = func(name string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{
			UserName:         "Goos",
			ConversationLang: "ko",
			PermissionMode:   "acceptEdits",
			StatuslineTheme:  "catppuccin-latte",
		}, nil
	}
	a.listProfiles = func() []profile.ProfileEntry {
		return []profile.ProfileEntry{{Name: "default", Current: true}}
	}
	h := a.routes()

	rec := serveGet(t, h, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`value="Goos"`,     // UserName populated
		`method="POST"`,    // editable POST form present (restyle adds class="form")
		`action="/save`,    // form posts to the save handler
		"catppuccin-latte", // theme option
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered form missing %q", want)
		}
	}
	// Selected language option should be marked.
	if !strings.Contains(body, `<option value="ko" selected`) {
		t.Errorf("conversation_lang=ko not marked selected:\n%s", body)
	}
	// permission_mode=acceptEdits should be selected.
	if !strings.Contains(body, `<option value="acceptEdits" selected`) {
		t.Error("permission_mode=acceptEdits not marked selected")
	}
}

// TestIndexNeutralDefaultsForZeroValueProfile verifies AC-WC-010 / REQ-WC-010:
// a zero-value profile renders neutral defaults rather than erroring.
func TestIndexNeutralDefaultsForZeroValueProfile(t *testing.T) {
	a := newTestApp(t)
	a.readPreferences = func(name string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, nil // zero-value, no error (mirrors ReadPreferences)
	}
	h := a.routes()

	rec := serveGet(t, h, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / on zero-value profile status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `method="POST"`) {
		t.Error("zero-value profile did not render a POST form (page may be blank)")
	}
	// The unset language option must be the selected default.
	if !strings.Contains(body, `<option value="" selected>(unset)</option>`) {
		t.Error("zero-value profile did not render neutral (unset) language default")
	}
}

// TestIndexReadErrorRendersInlineError verifies AC-WC-010 / REQ-WC-010: a read
// error surfaces a readable inline error (never blank, never panic).
func TestIndexReadErrorRendersInlineError(t *testing.T) {
	a := newTestApp(t)
	a.readPreferences = func(name string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, assertErr("disk on fire")
	}
	h := a.routes()

	rec := serveGet(t, h, "/")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("read error status = %d, want 500", rec.Code)
	}
	body := rec.Body.String()
	if strings.TrimSpace(body) == "" {
		t.Fatal("read error produced a blank page")
	}
	if !strings.Contains(body, "disk on fire") {
		t.Errorf("inline error does not surface the cause:\n%s", body)
	}
}

// TestIndexProfileSelectionShownForMultipleProfiles verifies AC-WC-011 /
// REQ-WC-011: with >1 profile, a selector listing profiles + current marker is
// rendered; with only default, the selector is omitted.
func TestIndexProfileSelectionShownForMultipleProfiles(t *testing.T) {
	a := newTestApp(t)
	a.readPreferences = func(name string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, nil
	}

	t.Run("multiple profiles → selector shown", func(t *testing.T) {
		a.listProfiles = func() []profile.ProfileEntry {
			return []profile.ProfileEntry{
				{Name: "default", Current: true},
				{Name: "work", Current: false},
			}
		}
		rec := serveGet(t, a.routes(), "/")
		body := rec.Body.String()
		if !strings.Contains(body, `name="__profile_select"`) {
			t.Error("profile selector not shown for multiple profiles")
		}
		if !strings.Contains(body, "work") {
			t.Error("profile list missing the 'work' profile")
		}
		if !strings.Contains(body, "(current)") {
			t.Error("current profile not marked")
		}
	})

	t.Run("single profile → selector omitted", func(t *testing.T) {
		a.listProfiles = func() []profile.ProfileEntry {
			return []profile.ProfileEntry{{Name: "default", Current: true}}
		}
		rec := serveGet(t, a.routes(), "/")
		if strings.Contains(rec.Body.String(), `name="__profile_select"`) {
			t.Error("profile selector should be omitted when only default exists")
		}
	})
}

// TestIndexProfileQueryParamSelectsProfile verifies the ?profile= query param
// selects which profile the form binds to (REQ-WC-011).
func TestIndexProfileQueryParamSelectsProfile(t *testing.T) {
	a := newTestApp(t)
	var readName string
	a.readPreferences = func(name string) (profile.ProfilePreferences, error) {
		readName = name
		return profile.ProfilePreferences{}, nil
	}
	a.listProfiles = func() []profile.ProfileEntry {
		return []profile.ProfileEntry{{Name: "default", Current: true}, {Name: "work"}}
	}
	serveGet(t, a.routes(), "/?profile=work")
	if readName != "work" {
		t.Errorf("ReadPreferences called with %q, want \"work\"", readName)
	}
}

// --- Phase 5: WRITE handler + validation (AC-WC-007, AC-WC-008, AC-WC-012) ---

// TestSaveValidRoundTrip verifies AC-WC-007 / REQ-WC-007: a valid submit calls
// WritePreferences + SyncToProjectConfig with the bound values; no direct YAML
// write occurs (the seams are the only persistence path).
func TestSaveValidRoundTrip(t *testing.T) {
	a := newTestApp(t)
	var (
		wroteName   string
		wrotePrefs  profile.ProfilePreferences
		syncedRoot  string
		syncedPrefs profile.ProfilePreferences
	)
	a.writePreferences = func(name string, prefs profile.ProfilePreferences) error {
		wroteName, wrotePrefs = name, prefs
		return nil
	}
	a.syncToProject = func(root string, prefs profile.ProfilePreferences) error {
		syncedRoot, syncedPrefs = root, prefs
		return nil
	}
	h := a.routes()

	form := url.Values{
		"__profile":         {"default"},
		"user_name":         {"Goos"},
		"conversation_lang": {"ko"},
		"permission_mode":   {"acceptEdits"},
		"statusline_theme":  {"catppuccin-latte"},
	}
	rec := servePost(t, h, "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	if wroteName != "default" {
		t.Errorf("WritePreferences profile = %q, want default", wroteName)
	}
	if wrotePrefs.ConversationLang != "ko" || wrotePrefs.UserName != "Goos" {
		t.Errorf("WritePreferences prefs = %+v, want ConversationLang=ko UserName=Goos", wrotePrefs)
	}
	if syncedRoot != a.cfg.ProjectRoot {
		t.Errorf("SyncToProjectConfig root = %q, want %q", syncedRoot, a.cfg.ProjectRoot)
	}
	if syncedPrefs.StatuslineTheme != "catppuccin-latte" {
		t.Errorf("SyncToProjectConfig theme = %q, want catppuccin-latte", syncedPrefs.StatuslineTheme)
	}
	if !strings.Contains(rec.Body.String(), "Settings saved") {
		t.Error("success banner not rendered")
	}
}

// TestSaveInvalidPermissionModeRejected verifies AC-WC-008 / REQ-WC-008: an
// invalid PermissionMode is rejected, no persistence occurs, and the form
// re-renders with a per-field error.
func TestSaveInvalidPermissionModeRejected(t *testing.T) {
	a := newTestApp(t)
	var wrote, synced bool
	a.writePreferences = func(string, profile.ProfilePreferences) error { wrote = true; return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { synced = true; return nil }
	h := a.routes()

	form := url.Values{
		"__profile":       {"default"},
		"permission_mode": {"definitely-not-a-mode"},
	}
	rec := servePost(t, h, "/save", form)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("invalid submit status = %d, want 400", rec.Code)
	}
	if wrote || synced {
		t.Error("persistence functions called despite validation failure (state must be unchanged)")
	}
	body := rec.Body.String()
	if !strings.Contains(body, "unrecognized permission mode") {
		t.Errorf("per-field permission_mode error not rendered:\n%s", body)
	}
	if !strings.Contains(body, "no changes were saved") {
		t.Error("rejection banner not rendered")
	}
}

// TestSaveInvalidThemeRejected verifies AC-WC-008: an unrecognized statusline
// theme is rejected via the canonical list, state unchanged.
func TestSaveInvalidThemeRejected(t *testing.T) {
	a := newTestApp(t)
	var wrote bool
	a.writePreferences = func(string, profile.ProfilePreferences) error { wrote = true; return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	form := url.Values{
		"__profile":        {"default"},
		"permission_mode":  {"acceptEdits"},
		"statusline_theme": {"hot-pink-9000"},
	}
	rec := servePost(t, h, "/save", form)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("invalid theme status = %d, want 400", rec.Code)
	}
	if wrote {
		t.Error("WritePreferences called despite invalid theme")
	}
	if !strings.Contains(rec.Body.String(), "unrecognized statusline theme") {
		t.Error("per-field theme error not rendered")
	}
}

// TestSaveSyncFailureSurfacesReadableError verifies plan-auditor advisory D1: a
// SyncToProjectConfig failure after a successful WritePreferences surfaces a
// readable error mentioning the partial state, rather than silent success.
func TestSaveSyncFailureSurfacesReadableError(t *testing.T) {
	a := newTestApp(t)
	wrote := false
	a.writePreferences = func(string, profile.ProfilePreferences) error { wrote = true; return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error {
		return assertErr("config dir read-only")
	}
	h := a.routes()

	form := url.Values{"__profile": {"default"}, "permission_mode": {"acceptEdits"}}
	rec := servePost(t, h, "/save", form)

	if !wrote {
		t.Fatal("WritePreferences should have been called before the sync failure")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("sync failure status = %d, want 500", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "sync failed") || !strings.Contains(body, "config dir read-only") {
		t.Errorf("sync failure did not surface a readable partial-state error:\n%s", body)
	}
}

// TestSaveScopeBoundary verifies AC-WC-012 / REQ-WC-012: persistence touches only
// the profile/sync seams (profile preferences + user/language/statusline). The
// handler has no path to any other config section — proven by asserting the only
// persistence calls are WritePreferences + SyncToProjectConfig.
func TestSaveScopeBoundary(t *testing.T) {
	a := newTestApp(t)
	var calls []string
	a.writePreferences = func(string, profile.ProfilePreferences) error {
		calls = append(calls, "WritePreferences")
		return nil
	}
	a.syncToProject = func(string, profile.ProfilePreferences) error {
		calls = append(calls, "SyncToProjectConfig")
		return nil
	}
	h := a.routes()

	form := url.Values{"__profile": {"default"}, "permission_mode": {"acceptEdits"}, "conversation_lang": {"en"}}
	servePost(t, h, "/save", form)

	if len(calls) != 2 || calls[0] != "WritePreferences" || calls[1] != "SyncToProjectConfig" {
		t.Errorf("persistence calls = %v, want exactly [WritePreferences SyncToProjectConfig]", calls)
	}
}

// TestSaveCustomSegmentsRoundTrip verifies EC-4: custom-preset segment toggles
// (map[string]bool) round-trip through the form without dropping keys.
func TestSaveCustomSegmentsRoundTrip(t *testing.T) {
	a := newTestApp(t)
	var got profile.ProfilePreferences
	a.writePreferences = func(_ string, prefs profile.ProfilePreferences) error { got = prefs; return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	form := url.Values{
		"__profile":          {"default"},
		"permission_mode":    {"acceptEdits"},
		"statusline_preset":  {"custom"},
		"segment_model":      {"1"},
		"segment_git_branch": {"1"},
		// other segments unchecked → recorded as false
	}
	rec := servePost(t, h, "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("custom segment save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	if got.StatuslineSegments == nil {
		t.Fatal("StatuslineSegments not bound for custom preset")
	}
	if !got.StatuslineSegments["model"] || !got.StatuslineSegments["git_branch"] {
		t.Errorf("checked segments not true: %+v", got.StatuslineSegments)
	}
	if got.StatuslineSegments["pr"] {
		t.Error("unchecked segment 'pr' should be false, not dropped or true")
	}
	if len(got.StatuslineSegments) != len(allSegments) {
		t.Errorf("segment map has %d keys, want all %d", len(got.StatuslineSegments), len(allSegments))
	}
}

// TestSaveNonCustomPresetLeavesSegmentsNil verifies EC-5: a non-custom preset
// leaves StatuslineSegments nil so syncStatusline applies its defaults rather
// than writing an empty map.
func TestSaveNonCustomPresetLeavesSegmentsNil(t *testing.T) {
	a := newTestApp(t)
	var got profile.ProfilePreferences
	a.writePreferences = func(_ string, prefs profile.ProfilePreferences) error { got = prefs; return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	form := url.Values{"__profile": {"default"}, "permission_mode": {"acceptEdits"}, "statusline_preset": {"full"}}
	servePost(t, h, "/save", form)
	if got.StatuslineSegments != nil {
		t.Errorf("non-custom preset bound segments = %+v, want nil", got.StatuslineSegments)
	}
}

// --- Phase 6: Host-check middleware (AC-WC-009) ---

// TestHostCheckRejectsForeignHostOnPost verifies AC-WC-009 / REQ-WC-009: a POST
// with a foreign Host header is rejected 403 and no persistence occurs.
func TestHostCheckRejectsForeignHostOnPost(t *testing.T) {
	a := newTestApp(t)
	var wrote bool
	a.writePreferences = func(string, profile.ProfilePreferences) error { wrote = true; return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	form := url.Values{"__profile": {"default"}, "permission_mode": {"acceptEdits"}}
	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "attacker.example.com"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("foreign-Host POST status = %d, want 403", rec.Code)
	}
	if wrote {
		t.Error("persistence occurred on a foreign-Host POST (state must be unchanged)")
	}
}

// TestHostCheckAllowsLoopbackHostsOnPost verifies AC-WC-009: loopback Hosts
// (127.0.0.1:port, localhost:port, [::1]:port) pass the Host check.
func TestHostCheckAllowsLoopbackHostsOnPost(t *testing.T) {
	a := newTestApp(t)
	a.writePreferences = func(string, profile.ProfilePreferences) error { return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	for _, host := range []string{"127.0.0.1:8080", "localhost:8080", "[::1]:8080", "127.0.0.1"} {
		form := url.Values{"__profile": {"default"}, "permission_mode": {"acceptEdits"}}
		req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Host = host
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code == http.StatusForbidden {
			t.Errorf("loopback host %q rejected with 403, want allowed", host)
		}
	}
}

// TestHostCheckDoesNotGateGet verifies AC-WC-009: GET is not Host-gated — read
// remains accessible even with a foreign Host header.
func TestHostCheckDoesNotGateGet(t *testing.T) {
	a := newTestApp(t)
	a.readPreferences = func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, nil
	}
	h := a.routes()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "attacker.example.com"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code == http.StatusForbidden {
		t.Error("GET / was Host-gated (403); read must remain accessible")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("GET / with foreign Host status = %d, want 200", rec.Code)
	}
}

// --- helpers ---

func serveGet(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func servePost(t *testing.T, h http.Handler, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// assertErr is a tiny error type for injecting failures in seams.
type assertErr string

func (e assertErr) Error() string { return string(e) }
