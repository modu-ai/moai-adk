package web

// mcp_secret_hygiene_test.go — SPEC-MCP-CONSOLE-001 M5 (REQ-C-8 / AC-C-012)
// secret-hygiene sweep.
//
// Consolidates the mechanical proof that NO console surface — view model,
// template, settings.AllFields() — ever serializes or renders a credential
// VALUE. Covers BOTH the codex auth surface (M3) and the GLM key surface (M4).
//
// This is a characterization test (DDD PRESERVE) over an invariant that already
// holds: M3/M4 were designed around the no-credential-in-view-model rule, and
// constraint C-C-2 keeps the GLM credential out of AllFields(). The sweep
// exists so a future change that silently adds a credential-bearing field to a
// view model, re-introduces the credential into the schema, or interpolates a
// credential reader directly inside a template fails here rather than leaking
// at render time. The per-surface tests (glmkey_test.go, mcp_codex_surface,
// mcp_glmkey_surface) cover each auth surface behaviorally; this file is the
// cross-surface structural sweep REQ-C-8 requires for the M5 close.

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// credentialNameFragments are field-name fragments a credential-bearing field
// would carry. A state/enum/hint view-model field (Installed, AuthProvider,
// GLMKeyConfigured, GLMKeyHint, ...) matches NONE of these; a field that did
// match would be carrying a raw credential and fail the audit. The fragments
// are deliberately lowercase-matched so the denylist is case-insensitive.
var credentialNameFragments = []string{
	"token",
	"secret",
	"credential",
	"password",
	"rawkey",
	"fullkey",
	"apikey",
}

// matchesCredentialFragment reports whether name contains any credential
// fragment (case-insensitive). Used by the view-model reflect audits below.
func matchesCredentialFragment(name string) bool {
	low := strings.ToLower(name)
	for _, frag := range credentialNameFragments {
		if strings.Contains(low, frag) {
			return true
		}
	}
	return false
}

// TestAC_C_012_NoCredentialFieldInCodexStateView asserts the M3 codex auth view
// model (CodexStateView) carries STATE/ENUM/HINT only — never a credential. It
// reflect-walks the struct and fails if any field matches a credential-name
// fragment. This is the structural proof that M3 did not smuggle a credential
// field (e.g. an OAuth token) into the view model (REQ-C-4 / AC-C-012).
func TestAC_C_012_NoCredentialFieldInCodexStateView(t *testing.T) {
	typ := reflect.TypeFor[CodexStateView]()
	if typ.Kind() != reflect.Struct {
		t.Fatalf("CodexStateView is not a struct (kind=%s) — reflect audit cannot proceed", typ.Kind())
	}
	for f := range typ.Fields() {
		if matchesCredentialFragment(f.Name) {
			t.Errorf("AC-C-012: CodexStateView.%s %s matches a credential-name fragment — the codex view model must carry state/enum/hint only, never a credential value", f.Name, f.Type)
		}
	}
}

// TestAC_C_012_GLMViewModelIsBoundedPairOnly asserts the M4 GLM key view-model
// surface is EXACTLY the bounded pair {GLMKeyConfigured bool, GLMKeyHint string}
// — never a field carrying the full key. It reflect-walks pageView's GLM-prefixed
// fields and fails if any matches a credential-name fragment OR is not one of
// the two known bounded fields. The allowlist is the stronger guard: a smuggled
// third GLM field fails here even if its name dodges the fragment denylist
// (REQ-C-7 / AC-C-011 / AC-C-012).
func TestAC_C_012_GLMViewModelIsBoundedPairOnly(t *testing.T) {
	typ := reflect.TypeFor[pageView]()
	if typ.Kind() != reflect.Struct {
		t.Fatalf("pageView is not a struct (kind=%s) — reflect audit cannot proceed", typ.Kind())
	}
	// allowed is the bounded-pair allowlist. Adding a legitimate third GLM
	// state field is a reviewed act: extend this allowlist deliberately.
	allowed := map[string]bool{
		"GLMKeyConfigured": true, // bool — configured/not-configured only
		"GLMKeyHint":       true, // string — bounded to <=4 trailing chars by computeGLMKeyHint
	}
	for f := range typ.Fields() {
		if !strings.HasPrefix(f.Name, "GLM") {
			continue
		}
		if matchesCredentialFragment(f.Name) {
			t.Errorf("AC-C-012: pageView.%s %s matches a credential-name fragment — the full GLM key must never be a view-model field (only the bounded hint pair is permitted)", f.Name, f.Type)
		}
		if !allowed[f.Name] {
			t.Errorf("AC-C-012: pageView.%s %s is not part of the bounded GLM hint pair — only GLMKeyConfigured (bool) and GLMKeyHint (<=4 chars) are permitted GLM view-model fields", f.Name, f.Type)
		}
	}
}

// TestAC_C_012_GLMCredentialAbsentFromAllFields asserts the C-C-2 structural
// anti-leak guarantee at the consolidated M5 level: the GLM credential form
// field (glm_api_key) and any apikey-named field are absent from
// settings.AllFields(), so no generic schema-walking loop (bulk value read,
// form-state dump, diagnostics view) can pick the credential up and render it.
// This consolidates TestGLMKeyField_AbsentFromSchema's assertion into the M5
// cross-surface sweep; the original test remains as the SPEC-GLM-KEY-INPUT-001
// anchor (constraint C-C-2 / REQ-C-8 / AC-C-012).
func TestAC_C_012_GLMCredentialAbsentFromAllFields(t *testing.T) {
	for _, f := range settings.AllFields() {
		low := strings.ToLower(f.Name)
		if strings.Contains(low, "glm_api_key") || strings.Contains(low, "apikey") {
			t.Errorf("AC-C-012: credential field %q leaked into settings.AllFields() — C-C-2 violated (a generic schema loop could render it)", f.Name)
		}
	}
}

// TestAC_C_012_NoTemplateInterpolatesCredentialReader asserts no .templ source
// bypasses the bounded view model by calling the credential reader directly.
// Templates MUST consume the pre-bounded view-model state (view.GLMKeyConfigured
// / view.GLMKeyHint), never reach into the glmcred package themselves — doing
// so would fork the credential path inside a template (REQ-C-7 / AC-C-010
// no-second-path, extended to the template surface for the AC-C-012 sweep).
// The view-model reflect audits above prove no credential field exists to
// interpolate; this guard closes the remaining bypass (a template calling the
// reader directly).
func TestAC_C_012_NoTemplateInterpolatesCredentialReader(t *testing.T) {
	files, err := filepath.Glob("*.templ")
	if err != nil {
		t.Fatalf("glob *.templ: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no .templ files found — the template sweep cannot run; verify the test cwd is internal/web")
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		body := string(data)
		// Direct credential-package access from a template bypasses the bounded
		// view model — it is the template-surface fork this guard forbids.
		if strings.Contains(body, "glmcred.") {
			t.Errorf("AC-C-012: %s references the glmcred package — templates must consume the bounded view model (view.GLMKeyConfigured / view.GLMKeyHint), never read the credential directly", f)
		}
	}
}
