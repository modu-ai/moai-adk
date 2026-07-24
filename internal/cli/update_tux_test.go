package cli

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/merge"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// M2 presentation-wiring tests for moai update (SPEC-CLI-TUX-INIT-UPDATE-001).
// These lock the observable output of the new render helpers in update_tux.go.
// All colour is sourced from tui.Theme tokens; NO_COLOR is simulated with the
// zero-value MonochromeTheme (Theme{} — every token is ""), which drives the
// tui.Pill/Box degrade paths.

// sgrColor matches any ANSI SGR sequence (colour, bold, reset). Under NO_COLOR
// the render helpers MUST emit zero of these (AC-TUXIU-013a / REQ-TUXIU-041).
var sgrColor = regexp.MustCompile("\x1b\\[[0-9;]*m")

// --- AC-TUXIU-001: card-style classification summary (all three pills) ---

func TestRenderClassificationSummary_AllThreePills(t *testing.T) {
	th := tui.LightTheme()
	got := renderClassificationSummary(1, 23, 2, th)

	// Accent-bordered card: a rounded-border glyph must appear.
	if !strings.ContainsAny(got, "╭╮╯╰─│") {
		t.Errorf("expected an accent box border in the summary card, got:\n%q", got)
	}
	for _, label := range []string{"+ 1 add", "~ 23 update", "! 2 conflict"} {
		if !strings.Contains(stripSGR(got), label) {
			t.Errorf("expected pill label %q in summary card, got:\n%q", label, stripSGR(got))
		}
	}
}

// --- AC-TUXIU-002a: zero-count conflict pill omitted ---

func TestRenderClassificationSummary_ZeroConflictOmitted(t *testing.T) {
	th := tui.LightTheme()
	got := stripSGR(renderClassificationSummary(1, 23, 0, th))
	if strings.Contains(got, "conflict") {
		t.Errorf("zero-count conflict pill must be omitted, got:\n%q", got)
	}
	if !strings.Contains(got, "1 add") || !strings.Contains(got, "23 update") {
		t.Errorf("non-zero add/update pills must still render, got:\n%q", got)
	}
}

// --- AC-TUXIU-002b: only the non-zero update pill renders ---

func TestRenderClassificationSummary_OnlyUpdatePill(t *testing.T) {
	th := tui.LightTheme()
	got := stripSGR(renderClassificationSummary(0, 5, 0, th))
	if !strings.Contains(got, "5 update") {
		t.Errorf("expected the update pill, got:\n%q", got)
	}
	if strings.Contains(got, "add") || strings.Contains(got, "conflict") {
		t.Errorf("both zero-count pills (add, conflict) must be omitted, got:\n%q", got)
	}
}

// A fully clean run (all counts zero) renders no card at all.
func TestRenderClassificationSummary_AllZeroEmpty(t *testing.T) {
	if got := renderClassificationSummary(0, 0, 0, tui.LightTheme()); got != "" {
		t.Errorf("all-zero summary must render empty, got:\n%q", got)
	}
}

// --- AC-TUXIU-013a: NO_COLOR → pills degrade to [label], zero SGR ---

func TestRenderClassificationSummary_NoColorBracketPills(t *testing.T) {
	mono := tui.MonochromeTheme() // Theme{} — NO_COLOR analogue
	got := renderClassificationSummary(1, 23, 2, mono)
	if n := len(sgrColor.FindAllString(got, -1)); n != 0 {
		t.Errorf("NO_COLOR summary must emit zero SGR sequences, got %d in:\n%q", n, got)
	}
	for _, label := range []string{"[+ 1 add]", "[~ 23 update]", "[! 2 conflict]"} {
		if !strings.Contains(got, label) {
			t.Errorf("NO_COLOR pill must degrade to %q, got:\n%q", label, got)
		}
	}
}

// --- classifyUpdateCounts derivation (feeds the card) ---

func TestClassifyUpdateCounts(t *testing.T) {
	files := []merge.FileAnalysis{
		{Path: "a.md", Changes: "new file", RiskLevel: "low"},
		{Path: "b.md", Changes: "update existing", RiskLevel: "medium"},
		{Path: "c.md", Changes: "update existing", RiskLevel: "low"},
		{Path: "settings.json", Changes: "update existing", RiskLevel: "high"},
		{Path: "CLAUDE.md", Changes: "new file", RiskLevel: "high"},
	}
	add, upd, conflict := classifyUpdateCounts(files)
	// high-risk files count as conflict (mutually exclusive, precedence).
	if conflict != 2 {
		t.Errorf("conflict count = %d, want 2 (the two high-risk files)", conflict)
	}
	if add != 1 {
		t.Errorf("add count = %d, want 1 (new+non-high)", add)
	}
	if upd != 2 {
		t.Errorf("update count = %d, want 2 (existing+non-high)", upd)
	}
}

// --- AC-TUXIU-008: identity header band with solid PillPrimary version pill ---

func TestRenderIdentityBand(t *testing.T) {
	th := tui.LightTheme()
	got := renderIdentityBand("v3.0.1", th)
	plain := stripSGR(got)
	if !strings.Contains(plain, "◆ MoAI-ADK") {
		t.Errorf("identity band must contain the ◆ MoAI-ADK marker, got:\n%q", plain)
	}
	if !strings.Contains(plain, "v3.0.1") {
		t.Errorf("identity band must contain the version, got:\n%q", plain)
	}
	if !strings.Contains(plain, "· claude") {
		t.Errorf("identity band must contain the '· claude' suffix, got:\n%q", plain)
	}
	// go-runtime token (e.g. "go1.23") — assert the "go" runtime prefix appears.
	if !strings.Contains(plain, "go") {
		t.Errorf("identity band must contain the go-runtime version, got:\n%q", plain)
	}
}

func TestRenderIdentityBand_NoColorBracketVersion(t *testing.T) {
	got := renderIdentityBand("v3.0.1", tui.MonochromeTheme())
	if n := len(sgrColor.FindAllString(got, -1)); n != 0 {
		t.Errorf("NO_COLOR identity band must emit zero SGR, got %d in:\n%q", n, got)
	}
	if !strings.Contains(got, "[v3.0.1]") {
		t.Errorf("NO_COLOR version pill must degrade to [v3.0.1], got:\n%q", got)
	}
}

// --- AC-TUXIU-003: unified deploy-step glyphs resolved from tui.StatusIcon ---

func TestDeployStepStateIcon(t *testing.T) {
	th := tui.LightTheme()
	tests := []struct {
		state deployStepState
		glyph rune
		name  string
	}{
		{stepPending, tui.GlyphSkip, "pending → ○"},
		{stepRunning, tui.GlyphRun, "running → ●"},
		{stepDone, tui.GlyphDone, "done → ✓"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripSGR(deployStepStateIcon(tt.state, th))
			if !strings.ContainsRune(got, tt.glyph) {
				t.Errorf("state icon = %q, want glyph %q (from tui.StatusIcon)", got, string(tt.glyph))
			}
			// It must be exactly the StatusIcon-resolved glyph, never a redefinition.
			wantKind := map[deployStepState]string{stepPending: "skip", stepRunning: "run", stepDone: "ok"}[tt.state]
			if stripSGR(got) != tui.StatusIcon(wantKind) {
				t.Errorf("state icon %q must equal tui.StatusIcon(%q)=%q", stripSGR(got), wantKind, tui.StatusIcon(wantKind))
			}
		})
	}
}

// --- AC-TUXIU-005: block progress bar reflecting done/total ---

func TestRenderDeployProgress_Bar(t *testing.T) {
	th := tui.LightTheme()
	got := stripSGR(renderDeployProgress(3, 5, th))
	if !strings.Contains(got, "█") || !strings.Contains(got, "░") {
		t.Errorf("progress bar must contain filled (█) and empty (░) cells for 3/5, got:\n%q", got)
	}
	if !strings.Contains(got, "3/5") {
		t.Errorf("progress line must show 3/5 step count, got:\n%q", got)
	}
	// Running (3<5) → leading ● from StatusIcon("run").
	if !strings.ContainsRune(got, tui.GlyphRun) {
		t.Errorf("in-progress deploy line must lead with ● (running), got:\n%q", got)
	}
}

func TestRenderDeployProgress_AllDone(t *testing.T) {
	got := stripSGR(renderDeployProgress(5, 5, tui.LightTheme()))
	if strings.Contains(got, "░") {
		t.Errorf("a fully complete bar (5/5) must have no empty cells, got:\n%q", got)
	}
	if !strings.ContainsRune(got, tui.GlyphDone) {
		t.Errorf("completed deploy line must lead with ✓ (done), got:\n%q", got)
	}
}

// --- AC-TUXIU-009: outcome banner — solid PillOk + dim detail note ---

func TestRenderUpdateOutcome(t *testing.T) {
	th := tui.LightTheme()
	var buf bytes.Buffer
	renderUpdateOutcome(&buf, 24, ".moai-backups/20260725_020747", th)
	out := stripSGR(buf.String())
	if !strings.Contains(out, "Updated 24 files") {
		t.Errorf("outcome must state 'Updated 24 files', got:\n%q", out)
	}
	if !strings.Contains(out, "Backup: .moai-backups/20260725_020747") {
		t.Errorf("outcome dim note must carry the backup path, got:\n%q", out)
	}
	if !strings.Contains(out, "Recover: moai update --restore-config .moai-backups/20260725_020747") {
		t.Errorf("outcome dim note must carry the recover command, got:\n%q", out)
	}
}

func TestRenderUpdateOutcome_SingularFile(t *testing.T) {
	var buf bytes.Buffer
	renderUpdateOutcome(&buf, 1, "", tui.LightTheme())
	out := stripSGR(buf.String())
	if !strings.Contains(out, "Updated 1 file") {
		t.Errorf("singular outcome must state 'Updated 1 file', got:\n%q", out)
	}
	// No backup path → no dim note lines.
	if strings.Contains(out, "Backup:") || strings.Contains(out, "Recover:") {
		t.Errorf("with no backup path, the dim note must be omitted, got:\n%q", out)
	}
}

func TestRenderUpdateOutcome_NoColor(t *testing.T) {
	var buf bytes.Buffer
	renderUpdateOutcome(&buf, 24, ".moai-backups/x", tui.MonochromeTheme())
	out := buf.String()
	if n := len(sgrColor.FindAllString(out, -1)); n != 0 {
		t.Errorf("NO_COLOR outcome must emit zero SGR, got %d in:\n%q", n, out)
	}
	// Solid pill degrades to bracketed label.
	if !strings.Contains(out, "[✓ Updated 24 files]") {
		t.Errorf("NO_COLOR outcome pill must degrade to [✓ Updated 24 files], got:\n%q", out)
	}
}

// stripSGR removes ANSI SGR sequences so text assertions are colour-agnostic.
func stripSGR(s string) string {
	return sgrColor.ReplaceAllString(s, "")
}
