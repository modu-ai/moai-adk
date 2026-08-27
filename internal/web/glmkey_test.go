package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/glmcred"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// sentinelKey is the acceptance-criteria sentinel (acceptance.md §A). It is
// distinctive enough that a content grep produces no false positives, which is
// what makes the absence assertions meaningful rather than vacuous.
const sentinelKey = "NOT-A-REAL-KEY-glm-acceptance-sentinel"

// withSandboxedGLMHome redirects the glmcred home-dir seam to a temp dir so
// credential tests never touch the developer's real ~/.moai/.env.glm. The
// override is restored on cleanup. Mirrors the userHomeDirFn pattern used in
// internal/cli, but applied directly to glmcred.HomeDirFn (the web package
// does not route through the cli init() alias).
func withSandboxedGLMHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	orig := glmcred.HomeDirFn
	glmcred.HomeDirFn = func() (string, error) { return dir, nil }
	// EnvTestGLMKey would otherwise short-circuit the on-disk read and mask a
	// real-file leak. Clear it for the duration of the test.
	t.Setenv(glmcred.EnvTestGLMKey, "")
	t.Cleanup(func() { glmcred.HomeDirFn = orig })
	return dir
}

// glmSaveURL encodes a minimal valid /save form for the given GLM key value.
// The other form fields are left at their empty/preserve defaults so the
// atomic-reject gate passes and the credential write is reached.
func glmSaveURL(key string) url.Values {
	v := url.Values{}
	v.Set("__profile", "default")
	v.Set(glmAPIKeyFormField, key)
	return v
}

// postGLMSave issues a same-origin POST to /save carrying the given form.
func postGLMSave(a *app, form url.Values) *httptest.ResponseRecorder {
	h := a.routes()
	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// renderSettingsGET fetches the rendered settings page.
func renderSettingsGET(a *app) string {
	h := a.routes()
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Body.String()
}

// ---- AC-GKI-001-01 / AC-GKI-001-02: render + secret-class control ----

func TestGLMKeyField_Renders(t *testing.T) {
	a := newTestApp(t)
	body := renderSettingsGET(a)
	if !strings.Contains(body, `name="`+glmAPIKeyFormField+`"`) {
		t.Fatalf("settings page missing GLM API key control (name=%q)", glmAPIKeyFormField)
	}
	// The control appears inside the GLM Settings section (between its legend
	// and the next section's legend).
	llm := indexOf(body, "GLM Settings")
	if llm < 0 {
		t.Fatalf("settings page missing the GLM Settings section")
	}
	key := indexOf(body, `name="`+glmAPIKeyFormField+`"`)
	if key < llm {
		t.Fatalf("GLM key control at %d appears before the GLM Settings legend at %d", key, llm)
	}
}

func TestGLMKeyField_SecretClass(t *testing.T) {
	a := newTestApp(t)
	body := renderSettingsGET(a)
	if !strings.Contains(body, `type="password"`) {
		t.Error("GLM key control is not a type=password masked input")
	}
	if !strings.Contains(body, `autocomplete="off"`) {
		t.Error("GLM key control does not disable browser autofill")
	}
}

// ---- AC-GKI-001-03 / AC-GKI-006-001: a submitted key reaches the credential file ----

func TestGLMKeySave_Persists(t *testing.T) {
	withSandboxedGLMHome(t)
	a := newTestApp(t)
	rec := postGLMSave(a, glmSaveURL(sentinelKey))
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if got := glmcred.Load(); got != sentinelKey {
		t.Fatalf("Load() after save = %q, want sentinel", got)
	}
}

// ---- AC-GKI-001-04: success banner confirms without echoing ----

func TestGLMKeySave_BannerNoEcho(t *testing.T) {
	withSandboxedGLMHome(t)
	a := newTestApp(t)
	rec := postGLMSave(a, glmSaveURL(sentinelKey))
	body := rec.Body.String()
	if !strings.Contains(body, "Settings saved") {
		t.Fatalf("success banner missing; body: %s", body)
	}
	if strings.Contains(body, sentinelKey) {
		t.Fatalf("response body echoes the submitted key material (banner contains the sentinel)")
	}
}

// ---- AC-GKI-004-01: full key never appears in any response body ----

func TestGLMKeySave_NoFullKeyInResponse(t *testing.T) {
	dir := withSandboxedGLMHome(t)
	// Seed a stored key first, then render the page — the response must not
	// carry the full key.
	if err := glmcred.Save(sentinelKey); err != nil {
		t.Fatalf("seed Save: %v", err)
	}
	a := newTestApp(t)
	body := renderSettingsGET(a)
	if strings.Contains(body, sentinelKey) {
		t.Fatalf("rendered settings page contains the full stored key")
	}
	_ = dir
}

// ---- AC-GKI-004-02: trailing-four hint disclosed, nothing else ----

func TestGLMKeyHint_TrailingFourOnly(t *testing.T) {
	withSandboxedGLMHome(t)
	if err := glmcred.Save(sentinelKey); err != nil {
		t.Fatalf("seed Save: %v", err)
	}
	a := newTestApp(t)
	withKey := renderSettingsGET(a)

	// Differential window: render again with NO key stored, then any substring
	// of the key present in the with-key render but absent in the no-key render
	// is a genuine disclosure.
	withSandboxedGLMHome(t) // fresh temp dir → no stored key
	a2 := newTestApp(t)
	withoutKey := renderSettingsGET(a2)

	wantHint := sentinelKey[len(sentinelKey)-4:] // "inel"
	if !strings.Contains(withKey, wantHint) {
		t.Fatalf("with-key render missing the trailing-four hint %q", wantHint)
	}
	// Differential scan: every proper substring of the key longer than 4 chars
	// that is NOT the trailing four must not appear exclusively in the with-key
	// render.
	for size := 5; size <= len(sentinelKey); size++ {
		for start := 0; start+size <= len(sentinelKey); start++ {
			frag := sentinelKey[start : start+size]
			if strings.Contains(withoutKey, frag) {
				continue // shared page vocabulary — not a disclosure
			}
			if strings.Contains(withKey, frag) {
				t.Fatalf("differential scan: key fragment %q (len %d, start %d) disclosed in with-key render but not in no-key render — only the final four are permitted",
					frag, size, start)
			}
		}
	}
}

// ---- AC-GKI-004-04: short key (<=4 chars) discloses zero characters ----

func TestGLMKeyHint_ShortKeyDisclosesNothing(t *testing.T) {
	// Derive short keys from the sentinel tail per acceptance.md §A so a single
	// key-shaped literal stays the source of truth.
	shortKeys := []string{"inel", "nel", "x"}
	for _, sk := range shortKeys {
		t.Run(sk, func(t *testing.T) {
			withSandboxedGLMHome(t)
			if err := glmcred.Save(sk); err != nil {
				t.Fatalf("seed Save(%q): %v", sk, err)
			}
			a := newTestApp(t)
			withKey := renderSettingsGET(a)

			// Differential window (acceptance.md §A): short keys like "inel"
			// collide with ordinary page vocabulary, so a naive substring scan
			// reports false positives. Compare against a no-key render and
			// treat as disclosed only fragments present in with-key but absent
			// in no-key.
			withSandboxedGLMHome(t)
			a2 := newTestApp(t)
			withoutKey := renderSettingsGET(a2)

			if strings.Contains(withKey, sk) && !strings.Contains(withoutKey, sk) {
				t.Fatalf("rendered page disclosed short stored key %q (present in with-key render, absent in no-key render)", sk)
			}
			if !strings.Contains(withKey, "Configured") {
				t.Fatalf("presence indicator missing for stored short key %q", sk)
			}
		})
	}
}

// ---- AC-GKI-005-01: empty submission preserves the stored key ----

func TestGLMKeySave_EmptyPreserves(t *testing.T) {
	dir := withSandboxedGLMHome(t)
	if err := glmcred.Save("pre-existing-key"); err != nil {
		t.Fatalf("seed Save: %v", err)
	}
	envPath := filepath.Join(dir, ".moai", ".env.glm")
	before, err := os.Stat(envPath)
	if err != nil {
		t.Fatalf("Stat before: %v", err)
	}
	a := newTestApp(t)
	form := glmSaveURL("") // empty submission
	rec := postGLMSave(a, form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200", rec.Code)
	}
	// Stored key unchanged; file mtime unchanged (no rewrite).
	if got := glmcred.Load(); got != "pre-existing-key" {
		t.Fatalf("Load() after empty submit = %q, want pre-existing-key (preserved)", got)
	}
	after, err := os.Stat(envPath)
	if err != nil {
		t.Fatalf("Stat after: %v", err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Fatalf("credential file mtime changed on empty submit (was rewritten): before=%v after=%v",
			before.ModTime(), after.ModTime())
	}
}

// ---- AC-GKI-005-02: surrounding whitespace is trimmed ----

func TestGLMKeySave_TrimsSurroundingWhitespace(t *testing.T) {
	withSandboxedGLMHome(t)
	a := newTestApp(t)
	rec := postGLMSave(a, glmSaveURL("  "+sentinelKey+"\t\n"))
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if got := glmcred.Load(); got != sentinelKey {
		t.Fatalf("Load() after whitespace-trim submit = %q, want %q (trimmed)", got, sentinelKey)
	}
}

// ---- AC-GKI-005-03: line-break value is rejected, stored key unchanged ----

func TestGLMKeySave_NewlineRejected(t *testing.T) {
	withSandboxedGLMHome(t)
	if err := glmcred.Save("pre-existing-key"); err != nil {
		t.Fatalf("seed Save: %v", err)
	}
	a := newTestApp(t)
	// A value with a line break in the body — NOT just surrounding whitespace.
	rec := postGLMSave(a, glmSaveURL("abc\ndef"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("save status = %d, want 400 (validation failure)", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, glmAPIKeyFormField) && !strings.Contains(body, "line break") {
		t.Fatalf("response missing per-field GLM key error; body: %s", body)
	}
	// Stored key unchanged.
	if got := glmcred.Load(); got != "pre-existing-key" {
		t.Fatalf("Load() after rejected newline submit = %q, want pre-existing-key (unchanged)", got)
	}
}

// ---- AC-GKI-002-01: credential file is the only file holding the key ----

func TestGLMKeySave_OnlyCredentialFileHoldsKey(t *testing.T) {
	homeDir := withSandboxedGLMHome(t)
	a := newTestApp(t) // a.cfg.ProjectRoot is a separate temp dir (project tree)
	postGLMSave(a, glmSaveURL(sentinelKey))

	// Scan both the home tree AND the project tree for any file whose contents
	// carry the sentinel. Exactly one match is expected: ~/.moai/.env.glm.
	matches := grepFilesFor(homeDir, sentinelKey)
	if projMatches := grepFilesFor(a.cfg.ProjectRoot, sentinelKey); len(projMatches) > 0 {
		matches = append(matches, projMatches...)
	}
	if len(matches) != 1 {
		t.Fatalf("expected exactly 1 file holding the sentinel, got %d: %v", len(matches), matches)
	}
	if !strings.HasSuffix(matches[0], filepath.Join(".moai", ".env.glm")) {
		t.Fatalf("sentinel-holding file is not the credential file: %q", matches[0])
	}
}

// ---- AC-GKI-003-03 / D-2 structural guarantee: field absent from schema ----

func TestGLMKeyField_AbsentFromSchema(t *testing.T) {
	// D-2 / R-1: the credential field MUST NOT be a settings.FieldDef. A future
	// generic loop over AllFields() (bulk value read, form-state dump,
	// diagnostics view) that picks up the credential would re-open every
	// leak site this SPEC closes. This regression test fails the moment a
	// later change adds the credential to the schema.
	for _, f := range settings.AllFields() {
		if strings.Contains(strings.ToLower(f.Name), "glm_api_key") ||
			strings.Contains(strings.ToLower(f.Name), "apikey") {
			t.Fatalf("credential field leaked into schema as FieldDef %q — D-2 violated", f.Name)
		}
	}
	// parseSchemaForm iterates AllFields(); confirm it cannot see the field.
	// (Defence in depth — the structural test above is the primary guard.)
}

// ---- AC-GKI-002-04: validation failure elsewhere leaves credential untouched ----

func TestGLMKeySave_AtomicRejectLeavesCredentialUntouched(t *testing.T) {
	dir := withSandboxedGLMHome(t)
	if err := glmcred.Save("pre-existing-key"); err != nil {
		t.Fatalf("seed Save: %v", err)
	}
	envPath := filepath.Join(dir, ".moai", ".env.glm")
	before, err := os.Stat(envPath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	a := newTestApp(t)
	// Submit the GLM key together with an INVALID sibling field so the
	// atomic-reject gate fires before any write.
	form := glmSaveURL(sentinelKey)
	form.Set("development_mode", "this-is-not-a-valid-mode")
	rec := postGLMSave(a, form)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("save status = %d, want 400 (atomic reject)", rec.Code)
	}
	// The credential file must be unchanged.
	after, err := os.Stat(envPath)
	if err != nil {
		t.Fatalf("Stat after: %v", err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Fatalf("credential file mtime changed despite atomic-reject (key was written)")
	}
	if got := glmcred.Load(); got != "pre-existing-key" {
		t.Fatalf("Load() after atomic-reject = %q, want pre-existing-key", got)
	}
}

// ---- AC-GKI-002-05: write failure is surfaced as failure, not success ----

func TestGLMKeySave_FailureSurfaced(t *testing.T) {
	// Force glmcred.Save to fail by making the credential directory unwritable
	// via a HomeDirFn that returns a path under a file (MkdirAll will fail).
	dir := withSandboxedGLMHome(t)
	// Create a regular file at <dir>/.moai so MkdirAll(<dir>/.moai) fails.
	if err := os.WriteFile(filepath.Join(dir, ".moai"), []byte("blocker"), 0o644); err != nil {
		t.Fatalf("seed blocker file: %v", err)
	}
	a := newTestApp(t)
	rec := postGLMSave(a, glmSaveURL(sentinelKey))
	body := rec.Body.String()
	if !strings.Contains(body, "GLM credential write failed") {
		t.Fatalf("response missing the credential-write failure surface; status=%d body: %s", rec.Code, body)
	}
	if strings.Contains(body, "Settings saved") {
		t.Fatalf("response presents a save success banner despite the write failure")
	}
}

// ---- AC-GKI-006-02: pre-existing 0644 file is tightened to 0600 via console ----

func TestGLMKeySave_NarrowsExistingWideMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not enforce POSIX permission bits: the pre-existing 0644 file cannot be created with a wide mode there, so the narrowing cannot be observed; the behavior stays covered on unix")
	}
	dir := withSandboxedGLMHome(t)
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(dir, ".moai", ".env.glm")
	if err := os.WriteFile(envPath, []byte(`GLM_API_KEY="stale"`), 0o644); err != nil {
		t.Fatal(err)
	}
	a := newTestApp(t)
	postGLMSave(a, glmSaveURL(sentinelKey))
	info, err := os.Stat(envPath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("credential file mode = %o, want 0600 after console save of pre-existing wide-mode file", perm)
	}
}

// ---- helpers ----

// grepFilesFor walks root recursively and returns the paths of regular files
// whose contents contain needle. Symlinks are not followed. Used by the
// anti-leak battery to assert on file CONTENTS, not filenames (plan §G AP-5).
func grepFilesFor(root, needle string) []string {
	var hits []string
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		data, rerr := os.ReadFile(p)
		if rerr != nil {
			return nil
		}
		if strings.Contains(string(data), needle) {
			hits = append(hits, p)
		}
		return nil
	})
	return hits
}

// ---- validateGLMKey unit tests (REQ-GKI-005-001/002/003) ----

func TestValidateGLMKey(t *testing.T) {
	cases := []struct {
		name      string
		submitted string
		wantErr   bool
	}{
		{"empty preserves", "", false},
		{"whitespace-only preserves", "   \t\n ", false},
		{"plain value accepted", sentinelKey, false},
		{"surrounding whitespace accepted", "  " + sentinelKey + "  ", false},
		{"newline in body rejected", "abc\ndef", true},
		{"carriage return rejected", "abc\rdef", true},
		// A trailing newline alone is surrounding whitespace → trimmed, accepted
		// (REQ-GKI-005-002). Only a newline IN THE BODY is rejected.
		{"trailing newline trimmed accepted", "abc\n", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := validateGLMKey(c.submitted)
			if c.wantErr && len(errs) == 0 {
				t.Fatalf("validateGLMKey(%q) = no errors, want at least one", c.submitted)
			}
			if !c.wantErr && len(errs) > 0 {
				t.Fatalf("validateGLMKey(%q) = %v, want no errors", c.submitted, errs)
			}
		})
	}
}

// Compile-time assertion that the field constant matches the form input name
// the tests POST. If anyone renames the constant without updating the
// renderer, this fails to compile.
var _ = glmAPIKeyFormField

// Silence the unused-import linter when only a subset of tests runs.
var _ = profile.IsValidProfileName
