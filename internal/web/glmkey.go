package web

import (
	"net/http"
	"strings"

	"github.com/modu-ai/moai-adk/internal/glmcred"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// SPEC-GLM-KEY-INPUT-001 — GLM API key field rendered in the 3rd Party LLM
// settings section. This file owns the credential field's hand-built parse /
// validate / view-model path. It is deliberately OUT of the schema FieldDef set
// (D-2): the credential never enters settings.AllFields() so no generic
// schema-walking loop (bulk value read, form-state dump, diagnostics view)
// can pick it up and write or render it. The structural guarantee is the
// anti-leak guard (M6 regression test: TestGLMKeyField_AbsentFromSchema).

// glmAPIKeyFormField is the name attribute of the GLM API key input control.
// It is the ONLY surface this field uses to round-trip through the form.
const glmAPIKeyFormField = "glm_api_key"

// glmKeyViewHint carries the redacted disclosure the 3rd Party LLM section may
// surface about a stored GLM API key. The full key NEVER crosses into the view
// model — only a "configured" boolean and, for a key longer than four
// characters, its final four characters (REQ-GKI-004-002 / -004-004).
type glmKeyViewHint struct {
	Configured bool
	Hint       string
}

// computeGLMKeyHint reads the stored credential through the shared credential
// reader and returns the bounded disclosure.
//
// For a key longer than four characters, Hint is exactly the final four
// characters and nothing else. For a key of four characters or fewer, Hint is
// empty and only Configured is true — a naive "last four or the whole key"
// fallback would disclose a short key entirely, which is the exact inverse of
// the requirement (REQ-GKI-004-004 / plan §G AP-2c).
func computeGLMKeyHint() glmKeyViewHint {
	key := glmcred.Load()
	if key == "" {
		return glmKeyViewHint{Configured: false}
	}
	if len(key) <= 4 {
		return glmKeyViewHint{Configured: true}
	}
	return glmKeyViewHint{Configured: true, Hint: key[len(key)-4:]}
}

// parseGLMKeyForm extracts the submitted GLM API key value from the POST form.
// It returns the raw submitted value (no trimming) so the validator can
// distinguish empty/whitespace-only (preserve, REQ-GKI-005-001) from
// surrounding-whitespace-around-a-body (trim then persist, REQ-GKI-005-002)
// from a value containing a line break (reject, REQ-GKI-005-003).
func parseGLMKeyForm(r *http.Request) string {
	return r.PostFormValue(glmAPIKeyFormField)
}

// validateGLMKey applies the submitted-value validation rules. It returns a
// map of field-name → error-message; an empty map means the submitted value
// is acceptable (including the empty/preserve case).
//
// Rules:
//   - empty or whitespace-only → preserve (no error, no write).
//   - body with surrounding whitespace → trimmed by the persist step.
//   - a line-break character IN THE BODY (after trimming) → reject. Trimming
//     first means a copy-paste with a trailing newline is treated as
//     surrounding whitespace and accepted (REQ-GKI-005-002), while a key that
//     genuinely spans lines — `"abc\ndef"` — is rejected (REQ-GKI-005-003).
//     This closes the §A.3 latent defect without extending the dotenv escaper
//     to faithfully round-trip a malformed value.
func validateGLMKey(submitted string) map[string]string {
	trimmed := strings.TrimSpace(submitted)
	if trimmed == "" {
		// Empty/whitespace-only: preserve the stored key. Not an error.
		return map[string]string{}
	}
	if strings.ContainsAny(trimmed, "\r\n") {
		return map[string]string{
			glmAPIKeyFormField: "GLM API key must not contain line breaks",
		}
	}
	return map[string]string{}
}

// normalizeGLMKey returns the value that should be persisted for a valid
// submitted key, or the empty string when no write should happen (empty /
// whitespace-only → preserve). The caller is responsible for calling
// validateGLMKey first and aborting on any error.
func normalizeGLMKey(submitted string) string {
	return strings.TrimSpace(submitted)
}

// glmKeyRevealPath is the endpoint that returns the stored GLM key in plaintext
// on explicit user action (SPEC-WEB-CONSOLE-REDESIGN-001 REQ-WCR-034).
//
// It is a POST for a read, deliberately. The never-echo contract above is what
// keeps the key out of the default page render; reveal is the one path that
// crosses it, so it must be an explicit act rather than something a page load,
// a prefetch, or an <img src> can trigger. POST also routes the request through
// hostCheckMiddleware's loopback-Host + same-origin gates, which safe methods
// are not subject to. The handler re-checks loopback itself so the guarantee
// does not rest solely on the middleware being wired.
const glmKeyRevealPath = "/glm-key/reveal"

// handleGLMKeyReveal serves POST /glm-key/reveal: the stored credential in
// plaintext, once, for a same-origin loopback caller.
//
// The console binds to loopback only, but that is an argument for why reveal is
// tolerable here — not a reason to skip the check. A non-loopback Host is
// refused, a non-POST method is refused, and an absent credential is a 404
// rather than an empty 200 (an empty body would be indistinguishable from a
// stored empty key).
func (a *app) handleGLMKeyReveal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isLoopbackHost(r.Host) {
		http.Error(w, "forbidden: non-loopback Host header", http.StatusForbidden)
		return
	}
	key := glmcred.Load()
	if key == "" {
		http.Error(w, "no GLM API key is configured", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// The plaintext must not be cached by anything between here and the tab.
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(key))
}

// glmKeyFieldBelongsToLLMSection reports whether the LLM schema section is the
// one that hosts the GLM API key control. Used by fieldsetSchemaSection to
// decide whether to render the hand-built key control alongside the schema
// tier-mapping fields.
func glmKeyFieldBelongsToLLMSection(id settings.SectionID) bool {
	return id == settings.SectionLLM
}
