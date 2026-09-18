package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestEmitAcceptEditsConfirmationAnchor covers AC-CCI-006 and AC-TRI-007: when
// the wizard normalizes permissionMode "acceptEdits" to empty string, an
// explicit confirmation line MUST be emitted to the output writer carrying a
// deterministic, grep-stable anchor — in EVERY locale (SPEC-CLI-TUX-RENDER-
// I18N-001 REQ-TRI-006: the notice was the M1 sweep's one English-fixed
// surface; localization must preserve the anchor tokens verbatim).
//
// The anchor states the two facts REQ-CCI-006 requires:
//
//	(1) "acceptEdits" is the project default;
//	(2) settings.local.json will NOT receive a defaultMode override.
//
// Anchor tokens per locale: the config tokens "acceptEdits" and
// "settings.local.json" survive verbatim in all four locales (REQ-TRI-006's
// anchor-token contract, owned by SPEC-V3R6-CLI-CONFIG-INTEGRITY-001
// REQ-CCI-006); the prose token "project default" is asserted on the English
// column only — the localized columns assert their own native sentence and
// the ABSENCE of the English original (AC-TRI-006 residue-zero contract).
func TestEmitAcceptEditsConfirmationAnchor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		locale     string
		native     string // a substring of the locale's own translation
		anchors    []string
		enResidues []string // English-original fragments that must NOT appear
	}{
		{
			locale:  "en",
			native:  "project default",
			anchors: []string{"acceptEdits", "project default", "settings.local.json"},
		},
		{
			locale:     "ko",
			native:     "프로젝트 기본값이므로",
			anchors:    []string{"acceptEdits", "settings.local.json"},
			enResidues: []string{"project default", "will be written"},
		},
		{
			locale:     "ja",
			native:     "プロジェクトのデフォルト",
			anchors:    []string{"acceptEdits", "settings.local.json"},
			enResidues: []string{"project default", "will be written"},
		},
		{
			locale:     "zh",
			native:     "项目默认值",
			anchors:    []string{"acceptEdits", "settings.local.json"},
			enResidues: []string{"project default", "will be written"},
		},
	} {
		var buf bytes.Buffer
		emitAcceptEditsConfirmation(&buf, tc.locale)
		out := buf.String()

		for _, anchor := range tc.anchors {
			if !strings.Contains(out, anchor) {
				t.Errorf("[%s] acceptEdits confirmation line missing anchor %q; got:\n%s", tc.locale, anchor, out)
			}
		}
		if !strings.Contains(out, tc.native) {
			t.Errorf("[%s] acceptEdits confirmation line is not the localized sentence (missing %q); got:\n%s", tc.locale, tc.native, out)
		}
		for _, residue := range tc.enResidues {
			if strings.Contains(out, residue) {
				t.Errorf("[%s] acceptEdits confirmation line carries the English-original residue %q (REQ-TRI-006); got:\n%s", tc.locale, residue, out)
			}
		}
	}

	// The en line MUST still state the "no override will be written" fact so
	// the user does not perceive the acceptEdits selection as a silent no-op.
	var buf bytes.Buffer
	emitAcceptEditsConfirmation(&buf, "en")
	if !strings.Contains(strings.ToLower(buf.String()), "no") {
		t.Errorf("acceptEdits confirmation line must state that NO override will be written; got:\n%s", buf.String())
	}
}

// TestAcceptEditsConfirmationEmittedOnce ensures the helper writes exactly one
// line (no duplicate prints, no trailing blank-line inflation) in every locale.
func TestAcceptEditsConfirmationEmittedOnce(t *testing.T) {
	t.Parallel()

	for _, locale := range []string{"en", "ko", "ja", "zh", "unknown"} {
		var buf bytes.Buffer
		emitAcceptEditsConfirmation(&buf, locale)
		out := buf.String()
		if out == "" {
			t.Fatalf("[%s] emitAcceptEditsConfirmation wrote nothing", locale)
		}
		// Exactly one trailing newline (single line emitted).
		if got := strings.Count(out, "\n"); got != 1 {
			t.Errorf("[%s] emitAcceptEditsConfirmation must emit exactly one line (1 trailing newline); got %d newlines:\n%s", locale, got, out)
		}
	}
}
