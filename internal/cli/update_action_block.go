package cli

// update_action_block.go — the end-of-run ACTION REQUIRED / Reference terminal
// block for `moai update` (card t1527 D5).
//
// Before this block the update tail was ~15 emitters in raw call-site order:
// failures surfaced mid-run (a failed hook install, a failed profile sync)
// scrolled off between ✓ progress lines and never came back, and advisories
// mixed with actions. This file owns the registry those failures record into
// and the terminal surface that renders it AFTER everything else:
//
//	Post-update actions ----------------------------------------
//	  Action required
//	    ✗ pre-push hook install failed: ...
//	    ! Hooks: running Claude Code sessions need a /hooks review ...
//	  Reference
//	    · reconfigure project settings: moai update -c
//	    · this checkout is shared across concurrent sessions; ...
//
// Skeleton boundary (approval condition 2, card t1527): the observed defect-4
// emitter ("Install it yourself", plugin-install failure) is NOT pinned in this
// tree — 0 grep hits in internal/ pkg/ cmd/ and absent from the sandbox
// captures. The registry below IS the escalation skeleton: when the real
// producer is pinned, its site calls updateLedger.requiref(...) and the
// terminal block carries it. No speculative producer ships here.
//
// The ledger is run-scoped package state following the updateVerboseMode
// precedent (update.go): `moai update` is single-process sequential, the run
// resets it on entry, and it is not read outside the update command.

import (
	"fmt"
	"io"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/update/report"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// updateActionRow is one recorded terminal-block row: its severity owns the
// glyph, the message carries the rest.
type updateActionRow struct {
	sev updateSeverity
	msg string
}

// updateActionLedger records failures (required) and advisories (reference)
// for the terminal block. Zero-value usable; not safe for concurrent use —
// the update command is sequential.
type updateActionLedger struct {
	required  []updateActionRow
	reference []updateActionRow
}

// updateLedger is the run-scoped registry the update flow records into.
var updateLedger = newUpdateActionLedger()

func newUpdateActionLedger() *updateActionLedger {
	return &updateActionLedger{}
}

// reset clears both sections (runUpdate entry).
func (l *updateActionLedger) reset() {
	l.required = nil
	l.reference = nil
}

// empty reports whether nothing was recorded (the block renders nothing).
func (l *updateActionLedger) empty() bool {
	return len(l.required) == 0 && len(l.reference) == 0
}

// requiref records an ACTION REQUIRED row — a failure or a conflict the
// operator must act on before the next run is trustworthy.
func (l *updateActionLedger) requiref(sev updateSeverity, format string, args ...any) {
	l.required = append(l.required, updateActionRow{sev: sev, msg: fmt.Sprintf(format, args...)})
}

// referencef records a Reference row — an advisory collapsed to one line.
func (l *updateActionLedger) referencef(format string, args ...any) {
	l.reference = append(l.reference, updateActionRow{sev: sevNote, msg: fmt.Sprintf(format, args...)})
}

// renderUpdateTerminalBlock writes the terminal block to w. An empty ledger
// renders nothing (a clean run ends at the outcome pill, not at an empty
// box). The hooks-restart guidance rides the ACTION REQUIRED section via the
// shared report text so the wording cannot drift between surfaces.
func renderUpdateTerminalBlock(w io.Writer, th tui.Theme) {
	if updateLedger.empty() {
		return
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, tui.Section("Post-update actions", tui.SectionOpts{Theme: &th}))

	if len(updateLedger.required) > 0 {
		_, _ = fmt.Fprintln(w, paintToken("Action required", th.Body, true))
		for _, row := range updateLedger.required {
			emitSeverityLine(w, row.sev, th, "%s", row.msg)
		}
	}
	if len(updateLedger.reference) > 0 {
		_, _ = fmt.Fprintln(w, paintToken("Reference", th.Body, true))
		for _, row := range updateLedger.reference {
			emitSeverityLine(w, row.sev, th, "%s", row.msg)
		}
	}
}

// hooksReviewGuidanceMsg returns the hook-restart guidance sentence from the
// shared report text (one wording for every surface).
func hooksReviewGuidanceMsg() string {
	var buf strings.Builder
	report.EmitHooksReviewGuidance(&buf)
	return strings.TrimRight(buf.String(), "\n")
}
