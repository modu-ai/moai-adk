package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestEmitAcceptEditsConfirmationAnchor covers AC-CCI-006 and AC-TRI-007: when
// the wizard saves permissionMode "acceptEdits", an explicit confirmation line
// MUST be emitted to the output writer carrying a deterministic, grep-stable
// anchor — in EVERY locale (SPEC-CLI-TUX-RENDER-I18N-001 REQ-TRI-006: the
// notice was the M1 sweep's one English-fixed surface; localization must
// preserve the anchor tokens verbatim).
//
// The anchor states the two facts REQ-CCI-006 requires (reversed by card
// t1414 — the template settings.json stopped shipping a defaultMode default
// in 20b4ff0f6, so an absent override lets CC 2.1.283+'s built-in default
// win):
//
//	(1) "acceptEdits" WILL be written to settings.local.json as defaultMode;
//	(2) the persisted mode survives Claude Code's built-in default.
//
// Anchor tokens per locale: the config tokens "acceptEdits",
// "settings.local.json" and "defaultMode" survive verbatim in all four
// locales (REQ-TRI-006's anchor-token contract, owned by
// SPEC-V3R6-CLI-CONFIG-INTEGRITY-001 REQ-CCI-006); the prose token "will be
// written" is asserted on the English column only — the localized columns
// assert their own native sentence and the ABSENCE of the English original
// (AC-TRI-006 residue-zero contract).
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
			native:  "will be written",
			anchors: []string{"acceptEdits", "settings.local.json", "defaultMode"},
		},
		{
			locale:     "ko",
			native:     "기록합니다",
			anchors:    []string{"acceptEdits", "settings.local.json", "defaultMode"},
			enResidues: []string{"will be written", "built-in default"},
		},
		{
			locale:     "ja",
			native:     "書き込まれます",
			anchors:    []string{"acceptEdits", "settings.local.json", "defaultMode"},
			enResidues: []string{"will be written", "built-in default"},
		},
		{
			locale:     "zh",
			native:     "写入",
			anchors:    []string{"acceptEdits", "settings.local.json", "defaultMode"},
			enResidues: []string{"will be written", "built-in default"},
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

	// The en line MUST state the write fact so the user does not perceive the
	// acceptEdits selection as a silent no-op (card t1414).
	var buf bytes.Buffer
	emitAcceptEditsConfirmation(&buf, "en")
	if !strings.Contains(strings.ToLower(buf.String()), "will be written") {
		t.Errorf("acceptEdits confirmation line must state that the override WILL be written; got:\n%s", buf.String())
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
