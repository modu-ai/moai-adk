package cli

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// Card t1527 D4/D5 tests — the severity-line vocabulary and the terminal
// ACTION REQUIRED / Reference block. Behavior-level: glyph presence, section
// order, and the empty-ledger suppression, never byte-snapshots.

var t1527SGR = regexp.MustCompile("\x1b\\[[0-9;]*m")

func strip1527SGR(s string) string { return t1527SGR.ReplaceAllString(s, "") }

// --- D4: severity glyphs resolve from the tui whitelist SSOT ---

func TestSeverityGlyph_MapsWhitelistGlyphs(t *testing.T) {
	th := tui.LightTheme()
	want := map[updateSeverity]rune{
		sevOK:   tui.GlyphDone,
		sevWarn: tui.GlyphWarn,
		sevErr:  tui.GlyphErr,
		sevNote: tui.GlyphInfo,
		sevRun:  tui.GlyphRun,
	}
	for sev, glyph := range want {
		got := strip1527SGR(severityGlyph(sev, th))
		if !strings.ContainsRune(got, glyph) {
			t.Errorf("severity %d glyph = %q, want %q (tui whitelist)", sev, got, string(glyph))
		}
	}
}

func TestEmitSeverityLine_Shape(t *testing.T) {
	th := tui.LightTheme()
	var buf bytes.Buffer
	emitSeverityLine(&buf, sevWarn, th, "hook install failed: %v", "perm")
	got := strip1527SGR(buf.String())
	if !strings.HasPrefix(got, string(tui.GlyphWarn)+" ") {
		t.Errorf("warn line must lead with the ! glyph, got %q", got)
	}
	if !strings.HasSuffix(got, "hook install failed: perm\n") {
		t.Errorf("warn line must carry the formatted message, got %q", got)
	}
}

// --- D5: the terminal block ---

func TestUpdateTerminalBlock_RendersRequiredAndReference(t *testing.T) {
	th := tui.LightTheme()
	updateLedger.reset()
	defer updateLedger.reset()

	updateLedger.requiref(sevErr, "pre-commit hook installation failed: perm denied")
	updateLedger.requiref(sevWarn, "%s", "Hooks: running sessions need a /hooks review")
	updateLedger.referencef("Reconfigure project settings: moai update -c")
	updateLedger.referencef("%s", "this checkout is shared across concurrent sessions")

	var buf bytes.Buffer
	renderUpdateTerminalBlock(&buf, th)
	out := strip1527SGR(buf.String())

	for _, want := range []string{
		"Post-update actions",
		"Action required",
		"pre-commit hook installation failed: perm denied",
		"Hooks: running sessions need a /hooks review",
		"Reference",
		"Reconfigure project settings: moai update -c",
		"this checkout is shared across concurrent sessions",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("terminal block must contain %q, got:\n%s", want, out)
		}
	}
	// Section order: every Action-required row precedes the Reference header.
	if strings.Index(out, "Reference") < strings.Index(out, "pre-commit hook installation failed") {
		t.Errorf("Reference section must follow the Action-required rows, got:\n%s", out)
	}
}

func TestUpdateTerminalBlock_EmptyLedgerRendersNothing(t *testing.T) {
	th := tui.LightTheme()
	updateLedger.reset()
	defer updateLedger.reset()

	var buf bytes.Buffer
	renderUpdateTerminalBlock(&buf, th)
	if buf.Len() != 0 {
		t.Errorf("an empty ledger must render no block, got %q", buf.String())
	}
}

func TestUpdateTerminalBlock_NoColorZeroSGR(t *testing.T) {
	updateLedger.reset()
	defer updateLedger.reset()
	updateLedger.requiref(sevErr, "boom")

	var buf bytes.Buffer
	renderUpdateTerminalBlock(&buf, tui.MonochromeTheme())
	if n := len(t1527SGR.FindAllString(buf.String(), -1)); n != 0 {
		t.Errorf("NO_COLOR terminal block must emit zero SGR, got %d in %q", n, buf.String())
	}
}

// The registry boundary (approval condition 2): the recorded-failure registry
// is the skeleton the unpinned plugin-install producer will call into — a
// requiref row lands in the terminal block without any dedicated producer.
func TestUpdateTerminalBlock_SkeletonRegistryCarriesRecordedRows(t *testing.T) {
	updateLedger.reset()
	defer updateLedger.reset()

	// Simulate any mid-run producer recording a failure.
	updateLedger.requiref(sevErr, "plugin install failed (producer unpinned — skeleton row)")

	var buf bytes.Buffer
	renderUpdateTerminalBlock(&buf, tui.LightTheme())
	out := strip1527SGR(buf.String())
	if !strings.Contains(out, "plugin install failed (producer unpinned — skeleton row)") {
		t.Errorf("skeleton registry row must render in the terminal block, got:\n%s", out)
	}
}
