package web

import (
	"net/http"
	"strings"

	"github.com/modu-ai/moai-adk/internal/jevcred"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// SPEC-JEV-OPTIN-MEASURE-001 — the Jev credential field rendered in the Jev
// settings section. This file owns the credential's hand-built parse /
// validate / view-model path, following the GLM precedent in glmkey.go.
//
// It is deliberately OUT of the schema FieldDef set: the credential never
// enters settings.AllFields(), so no generic schema-walking loop (bulk value
// read, form-state dump, diagnostics view) can pick it up and write or render
// it. The structural guarantee is only as good as the thing that notices when
// it stops holding, which is TestJevCredential_AbsentFromSchema.
//
// TWO deliberate differences from the GLM precedent:
//
//   - No reveal route. GLM has one (glmKeyRevealPath) because a user needs to
//     read that key back out; the disclosure requirement here is satisfied by
//     the configured flag plus the trailing four characters, so the one path
//     that crosses the never-echo contract is simply not built.
//   - The credential is read and written through internal/jevcred, the reader
//     the core capability already uses. There is no second reader.

// jevAPIKeyFormField is the name attribute of the Jev credential input. It is
// the ONLY surface this field uses to round-trip through the form.
const jevAPIKeyFormField = "jev_api_key"

// jevPrivacyNoteKey is the i18n key of the panel-header privacy statement
// (REQ-JEVO-003): the sentence stating, at the point of choice, that enabling
// sends card text or request text to a third-party server.
const jevPrivacyNoteKey = "sec.jev.note"

// jevPrivacyNoteBaseline is the inline English fallback rendered alongside the
// data-i18n key. It is what a reader sees before the catalogue loads, so the
// privacy statement does not depend on client-side i18n having run.
const jevPrivacyNoteBaseline = "Jev answers a typed question about supplied state and returns a probability; it decides nothing. While it is enabled, card text or request text is sent to a third-party server."

// jevKeyViewHint carries the redacted disclosure the Jev section may surface
// about a stored credential. The full value NEVER crosses into the view model
// — only a "configured" boolean and, for a value longer than four characters,
// its final four characters.
type jevKeyViewHint struct {
	Configured bool
	Hint       string
}

// computeJevKeyHint reads the stored credential through the shared reader and
// returns the bounded disclosure.
//
// For a value of four characters or fewer, Hint is empty and only Configured
// is true: a naive "last four, or the whole value if shorter" fallback would
// disclose a short credential entirely, which is the exact inverse of the
// requirement.
func computeJevKeyHint() jevKeyViewHint {
	v := jevcred.View()
	return jevKeyViewHint{Configured: v.Configured, Hint: v.Hint}
}

// parseJevKeyForm extracts the submitted credential from the POST form. It
// returns the raw submitted value (no trimming) so the validator can
// distinguish empty/whitespace-only (preserve) from surrounding whitespace
// around a body (trim then persist) from a value containing a line break
// (reject).
func parseJevKeyForm(r *http.Request) string {
	return r.PostFormValue(jevAPIKeyFormField)
}

// validateJevKey applies the submitted-value rules. It returns a map of
// field-name -> error-message; an empty map means the value is acceptable
// (including the empty/preserve case).
//
// Trimming before the line-break check means a copy-paste with a trailing
// newline is treated as surrounding whitespace and accepted, while a value
// that genuinely spans lines is rejected rather than written into the dotenv
// file in a shape the reader cannot round-trip.
func validateJevKey(submitted string) map[string]string {
	trimmed := strings.TrimSpace(submitted)
	if trimmed == "" {
		return map[string]string{}
	}
	if strings.ContainsAny(trimmed, "\r\n") {
		return map[string]string{
			jevAPIKeyFormField: "Jev credential must not contain line breaks",
		}
	}
	return map[string]string{}
}

// normalizeJevKey returns the value to persist for a valid submission, or the
// empty string when no write should happen (empty / whitespace-only means
// preserve). Callers must run validateJevKey first and abort on any error.
func normalizeJevKey(submitted string) string {
	return strings.TrimSpace(submitted)
}

// jevFieldBelongsToPanel reports whether a schema field renders inside the Jev
// sub-section rather than loose among the general workflow scalars. The field
// keeps SectionWorkflow and the workflow.yaml seam — this is a render
// placement, not a section reclassification, exactly as the audit tab's fields
// are.
func jevFieldBelongsToPanel(name string) bool {
	return name == settings.JevEnabledField
}

// jevSectionFields returns the schema fields the Jev sub-section renders.
//
// The sub-section lives INSIDE the workflow panel rather than on a tab of its
// own. That is a deliberate scope choice, not an oversight: this repository
// couples the console's tab list to eight documentation surfaces (four README
// locales and four docs-site locales), and a run-phase change does not own
// those files. A sub-section satisfies "one Jev section carrying an enable
// toggle and a credential field" without reaching into documentation this SPEC
// has no mandate over.
func jevSectionFields() []settings.FieldDef {
	_, _, _, jev := partitionWorkflowFields()
	return jev
}

// jevSectionMarker is the attribute value that identifies the Jev sub-section
// in the rendered markup, so a test can slice it the way panelHTML slices a
// panel.
const jevSectionMarker = "jev"
