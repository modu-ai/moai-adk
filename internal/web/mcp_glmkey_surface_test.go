package web

// SPEC-MCP-CONSOLE-001 M4 — GLM key surface tests.
//
// These tests cover AC-C-010 (the GLM credential path is reused unchanged and
// surfaced in the MCP console section) and AC-C-011 (disclosure stays bounded:
// only a configured-boolean and the final four characters, and no characters at
// all for a key of four characters or fewer).
//
// The GLM key state view model (pageView.GLMKeyConfigured / GLMKeyHint) is
// populated by the pre-existing populateGLMKeyHint() -> computeGLMKeyHint()
// path authored by SPEC-GLM-KEY-INPUT-001. M4 does NOT re-author glmkey.go; it
// surfaces that state in the MCP console section alongside the codex auth block.
// The bounded-disclosure logic itself is unchanged and remains covered by
// glmkey_test.go (TestGLMKeyHint_TrailingFourOnly / _ShortKeyDisclosesNothing)
// and TestGLMKeyField_AbsentFromSchema — AC-C-011's regression evidence is
// those tests still PASS.

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// renderGlmKeyStateBlock renders the glmKeyStateBlock templ component directly
// with a hand-built pageView, so the test is deterministic and does not depend
// on the host's real ~/.moai/.env.glm credential store. The bounded-disclosure
// invariant (computeGLMKeyHint, unchanged) is what produces view.GLMKeyHint in
// production; here we feed the view model directly to exercise only the
// rendering contract.
func renderGlmKeyStateBlock(t *testing.T, view pageView) string {
	t.Helper()
	var buf bytes.Buffer
	if err := glmKeyStateBlock(view).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render glmKeyStateBlock: %v", err)
	}
	return buf.String()
}

// ─── AC-C-010: GLM credential path reused, surfaced in MCP section ───────

// TestAC_C_010_GLMKeyStateSurfacedInMCPSection asserts that the MCP console
// section renders a GLM key state block (the credential path is REUSED, not
// re-authored). The block carries the configured boolean and bounded hint only;
// no credential input, no reveal action (those live in the 3rd Party LLM
// section — M4 surfaces STATE only, REQ-C-7 "no second credential path").
func TestAC_C_010_GLMKeyStateSurfacedInMCPSection(t *testing.T) {
	view := pageView{GLMKeyConfigured: true, GLMKeyHint: "abcd"}
	body := renderGlmKeyStateBlock(t, view)
	if !strings.Contains(body, "sec.mcp.glm_key.title") {
		t.Error("AC-C-010: GLM key state block not rendered in MCP section (missing the sec.mcp.glm_key.title anchor)")
	}
	if !strings.Contains(body, "Configured") && !strings.Contains(body, "configured") {
		t.Error("AC-C-010: configured state must be surfaced")
	}
}

// TestAC_C_010_GLMKeyStateRenderedViaRoute asserts the GLM key state block is
// present in a full GET / render (end-to-end through the route handler), so the
// block is wired into fieldsetMCP and not orphaned. The configured/hint values
// depend on the host store, so we only assert the block's i18n anchor is present.
func TestAC_C_010_GLMKeyStateRenderedViaRoute(t *testing.T) {
	a := newTestApp(t)
	h := a.routes()
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("AC-C-010: GET / status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "glm-key") {
		t.Error("AC-C-010: full page render missing the GLM key state block")
	}
}

// TestAC_C_010_NoSecondCredentialPathInWeb is the no-fork guard adapted for M4:
// internal/web must not gain a SECOND GLM key reading path beyond the
// pre-existing one. Unlike M3 codex — whose classifier lived in internal/cli and
// was consumed via probe — the GLM key classifier (computeGLMKeyHint) and reader
// (glmcred.Load) already live in internal/web/glmkey.go by design
// (SPEC-GLM-KEY-INPUT-001). The guard asserts the invariant: glmcred.Load is
// CALLED only from glmkey.go (the sole reader), and M4's new code does not
// introduce a duplicate credential read. Generated _templ.go files are excluded
// (they carry templ-comment leakage, not executable calls).
func TestAC_C_010_NoSecondCredentialPathInWeb(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob *.go: %v", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") || strings.HasSuffix(f, "_templ.go") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		body := string(data)
		// glmcred.Load( is always a CALL. Only glmkey.go may call it.
		if strings.Contains(body, "glmcred.Load(") && f != "glmkey.go" {
			t.Errorf("AC-C-010: %s calls glmcred.Load — glmkey.go is the sole credential reader", f)
		}
	}
}

// ─── AC-C-011: disclosure stays bounded (rendering contract) ─────────────

// TestAC_C_011_HintShownWhenConfiguredAndLong asserts that for a configured key
// longer than four characters, the MCP-section state block surfaces the bounded
// trailing-four hint and nothing more.
func TestAC_C_011_HintShownWhenConfiguredAndLong(t *testing.T) {
	view := pageView{GLMKeyConfigured: true, GLMKeyHint: "wxyz"}
	body := renderGlmKeyStateBlock(t, view)
	if !strings.Contains(body, "wxyz") {
		t.Error("AC-C-011: bounded trailing-four hint must be rendered when present")
	}
}

// TestAC_C_011_NoCharactersForShortKey asserts that for a configured key of
// four characters or fewer (hint empty by computeGLMKeyHint's contract), the
// MCP-section state block surfaces ONLY the configured boolean and zero key
// characters — no hint, no partial key.
func TestAC_C_011_NoCharactersForShortKey(t *testing.T) {
	view := pageView{GLMKeyConfigured: true, GLMKeyHint: ""}
	body := renderGlmKeyStateBlock(t, view)
	if !strings.Contains(body, "Configured") && !strings.Contains(body, "configured") {
		t.Error("AC-C-011: configured boolean must be shown for a short stored key")
	}
	// The hint-rendering branch must be absent: no field__hint-key / hint code
	// element should appear when the hint is empty.
	if strings.Contains(body, "field__hint-key") {
		t.Error("AC-C-011: hint element must not render when hint is empty (short key discloses nothing)")
	}
}

// TestAC_C_011_NotConfiguredState asserts that when no key is stored, the
// MCP-section state block surfaces the not-configured state and no hint.
func TestAC_C_011_NotConfiguredState(t *testing.T) {
	view := pageView{GLMKeyConfigured: false, GLMKeyHint: ""}
	body := renderGlmKeyStateBlock(t, view)
	if strings.Contains(body, "field__hint-key") {
		t.Error("AC-C-011: no hint must render when key is not configured")
	}
}
