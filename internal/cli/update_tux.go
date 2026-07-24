package cli

// update_tux.go — M2 presentation render helpers for `moai update`
// (SPEC-CLI-TUX-INIT-UPDATE-001). These wire the already-existing internal/tui
// design-system primitives (Box, Pill, Progress, StatusIcon) into the update
// screen. This is a presentation-only layer: it re-dresses output the update
// flow already emits and introduces no data, no dependency, and no channel
// change. All colour is sourced from tui.Theme tokens — no hex literal appears
// in this file (REQ-TUXIU-040). Under NO_COLOR the effective theme is the
// zero-value MonochromeTheme (every token ""), so every helper degrades to
// plain text with zero SGR sequences (REQ-TUXIU-041).

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/modu-ai/moai-adk/internal/cli/update/report"
	"github.com/modu-ai/moai-adk/internal/merge"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// paintToken colours s with the given tui.Theme token. An empty token
// (MonochromeTheme / NO_COLOR) returns s verbatim so no SGR sequence — not even
// a bold attribute — is emitted. This is the single NO_COLOR-safe styling
// gateway for this file; it keeps colour decisions flowing through Theme tokens
// rather than raw hex (REQ-TUXIU-040/041).
func paintToken(s, token string, bold bool) string {
	if token == "" {
		return s
	}
	st := lipgloss.NewStyle().Foreground(lipgloss.Color(token))
	if bold {
		st = st.Bold(true)
	}
	return st.Render(s)
}

// deployStepState is the lifecycle state of one update deploy step.
type deployStepState int

const (
	// stepPending — the step has not started (○, Faint).
	stepPending deployStepState = iota
	// stepRunning — the step is in progress (●, Accent bold).
	stepRunning
	// stepDone — the step completed (✓, Success).
	stepDone
)

// deployStepStateIcon resolves the status glyph for a deploy step state from
// the single tui.StatusIcon source — ○ (Faint) / ● (Accent, bold) / ✓ (Success)
// — never a per-step redefined glyph (REQ-TUXIU-012/013, AC-TUXIU-003).
func deployStepStateIcon(state deployStepState, th tui.Theme) string {
	switch state {
	case stepRunning:
		return paintToken(tui.StatusIcon("run"), th.Accent, true)
	case stepDone:
		return paintToken(tui.StatusIcon("ok"), th.Success, false)
	default: // stepPending
		return paintToken(tui.StatusIcon("skip"), th.Faint, false)
	}
}

// renderIdentityBand renders the update identity header band
// "◆ MoAI-ADK <version> <go-runtime> · claude" with the version rendered as a
// solid brand pill (tui.Pill, PillPrimary, Solid) — REQ-TUXIU-015 /
// AC-TUXIU-008. The ◆ MoAI-ADK marker is in the AC-CLI-TUI-017 whitelist; the
// surrounding text is Theme-token painted (accent marker, dim runtime suffix)
// and degrades to plain under NO_COLOR.
func renderIdentityBand(version string, th tui.Theme) string {
	marker := paintToken("◆ MoAI-ADK", th.Accent, true)
	versionPill := tui.Pill(tui.PillOpts{Kind: tui.PillPrimary, Solid: true, Label: version, Theme: &th})
	suffix := paintToken(runtime.Version()+" · claude", th.Dim, false)
	return marker + " " + versionPill + " " + suffix
}

// classifyUpdateCounts derives the (add, update, conflict) summary counts from
// a merge analysis's per-file classification. The three buckets are mutually
// exclusive with conflict (high-risk) precedence, so they sum to the total file
// count. This reads the SAME FileAnalysis fields the deploy stage already
// populated (Changes via plan.DetermineChangeType, RiskLevel via
// plan.ClassifyFileRisk) — it introduces no parallel classification heuristic
// (presentation-only, REQ-TUXIU-044).
func classifyUpdateCounts(files []merge.FileAnalysis) (add, update, conflict int) {
	for _, f := range files {
		switch {
		case f.RiskLevel == "high":
			conflict++
		case strings.Contains(f.Changes, "new"):
			add++
		default:
			update++
		}
	}
	return add, update, conflict
}

// renderClassificationSummary renders the merge classification as an
// accent-bordered card (tui.Box, Accent) carrying up to three semantic count
// pills — PillOk "+ N add", PillInfo "~ N update", PillErr "! N conflict"
// (REQ-TUXIU-010, AC-TUXIU-001). A zero-count pill is omitted entirely
// (REQ-TUXIU-011, AC-TUXIU-002a/b); when every count is zero the card is
// suppressed (empty string) so a clean run shows no empty box.
func renderClassificationSummary(add, update, conflict int, th tui.Theme) string {
	var pills []string
	if add > 0 {
		pills = append(pills, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Label: fmt.Sprintf("+ %d add", add), Theme: &th}))
	}
	if update > 0 {
		pills = append(pills, tui.Pill(tui.PillOpts{Kind: tui.PillInfo, Label: fmt.Sprintf("~ %d update", update), Theme: &th}))
	}
	if conflict > 0 {
		pills = append(pills, tui.Pill(tui.PillOpts{Kind: tui.PillErr, Label: fmt.Sprintf("! %d conflict", conflict), Theme: &th}))
	}
	if len(pills) == 0 {
		return ""
	}
	return tui.Box(tui.BoxOpts{Accent: true, Body: strings.Join(pills, "  "), Theme: &th})
}

// renderDeployProgress renders the deploy-step progress as a leading status
// glyph (● while running, ✓ once all steps complete — resolved from
// tui.StatusIcon) followed by a block progress bar (tui.Progress, ██████░░░░)
// reflecting done/total and the "N/M steps" count. It replaces the legacy
// "N/M steps complete" plain text with the block bar (REQ-TUXIU-014,
// AC-TUXIU-005).
func renderDeployProgress(done, total int, th tui.Theme) string {
	lead := deployStepStateIcon(stepRunning, th)
	if done >= total {
		lead = deployStepStateIcon(stepDone, th)
	}
	bar := tui.Progress(done, total, tui.ProgressOpts{Theme: &th, Width: 10})
	return fmt.Sprintf("  %s %s %d/%d steps", lead, bar, done, total)
}

// renderUpdateOutcome writes the completion outcome as a solid success pill
// ("✓ Updated N files") followed by a dim detail note carrying the backup /
// recover commands (REQ-TUXIU-016, AC-TUXIU-009). The pill label and the note
// text are byte-identical to the data the legacy single-pill form emitted — the
// backup path and recover command survive verbatim (presentation-only,
// REQ-TUXIU-044); only the styling (solid pill + dim note vs one outline pill)
// changes.
func renderUpdateOutcome(w io.Writer, fileCount int, backupPath string, th tui.Theme) {
	header := report.RenderOutcome(report.OutcomeUpdatedFiles, fileCount, "")
	_, _ = fmt.Fprintln(w, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: true, Label: header, Theme: &th}))
	if backupPath != "" {
		note := "Backup: " + backupPath + "\nRecover: moai update --restore-config " + backupPath
		_, _ = fmt.Fprintln(w, paintToken(note, th.Dim, false))
	}
}
