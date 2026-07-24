package cli

// Warning collector + completion card for moai init
// (SPEC-CLI-TUX-V3-002 M2d, REQ-TUX2-013/014/016).
//
// Warnings during an init run previously scattered across the phases with a
// single aggregation point (the executor result warnings rendered inside the
// success card). The collector wraps the Printer so every Warn is recorded,
// then re-emits ALL warnings exactly once as a consolidated stderr summary
// panel when the run terminates (success or failure). stdout never carries
// warning text (channel discipline inherited from REQ-CTX-012/016).

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// warnCollector wraps a printer.Printer, recording every Warn message for
// the exit summary while still streaming the individual line to stderr as
// it occurs.
type warnCollector struct {
	printer.Printer
	mu       sync.Mutex
	warnings []string
	emitted  bool
}

// newWarnCollector wraps p with warning collection (REQ-TUX2-013).
func newWarnCollector(p printer.Printer) *warnCollector {
	return &warnCollector{Printer: p}
}

// Warn records the message and forwards it to the wrapped printer (the
// individual stderr line still appears at the moment it occurs).
func (w *warnCollector) Warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	w.mu.Lock()
	w.warnings = append(w.warnings, msg)
	w.mu.Unlock()
	w.Printer.Warn("%s", msg)
}

// Collect records a warning WITHOUT an individual re-emission — used for
// warnings that surface only at completion time (executor result warnings),
// so the summary panel remains their single emission point.
func (w *warnCollector) Collect(msg string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.warnings = append(w.warnings, msg)
}

// Count reports the number of collected warnings.
func (w *warnCollector) Count() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.warnings)
}

// emitSummary writes the consolidated warning summary panel exactly once
// (REQ-TUX2-013). Zero warnings emit nothing (acceptance.md §C edge case).
// The caller passes the stderr writer — warning text never reaches stdout
// (REQ-TUX2-014).
func (w *warnCollector) emitSummary(errOut io.Writer) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.emitted || len(w.warnings) == 0 {
		return
	}
	w.emitted = true

	var b strings.Builder
	fmt.Fprintf(&b, "%d warning(s) during init:", len(w.warnings))
	for i, msg := range w.warnings {
		fmt.Fprintf(&b, "\n  %d. %s", i+1, msg)
	}
	_, _ = fmt.Fprintln(errOut)
	_, _ = fmt.Fprintln(errOut, uikit.WarnStyle.Render(b.String()))
}

// buildInitSuccessCard renders the init completion card (REQ-TUX2-016 /
// REQ-TUXIU-031): created-artifact counts, the next-action sequence, and — when
// warnings were collected — a one-line pointer to the stderr warning summary.
//
// SPEC-CLI-TUX-INIT-UPDATE-001 M3 (AC-TUXIU-011) re-dresses the card in the
// shared tui.Box + tui.Pill visual language, mirroring the update
// classification summary (renderClassificationSummary in update_tux.go): the
// counts ride semantic tui.Pill pills (dirs = PillOk, files = PillInfo) inside
// an accent-bordered tui.Box. This is presentation-only — the emitted data
// (counts, next-action steps, warning pointer) is byte-identical to the legacy
// form. Colour is sourced from tui.Theme tokens (paintToken); under NO_COLOR
// every helper degrades to plain text (tui.Pill → "[label]", tui.Box border
// runes intact) — REQ-TUXIU-040/041.
func buildInitSuccessCard(projectName string, dirs, files, warnCount int) string {
	th := resolveTheme()

	countRow := tui.Pill(tui.PillOpts{Kind: tui.PillOk, Label: fmt.Sprintf("%d dirs", dirs), Theme: &th}) +
		"  " + tui.Pill(tui.PillOpts{Kind: tui.PillInfo, Label: fmt.Sprintf("%d files", files), Theme: &th})

	nextSteps := "Next steps:\n" +
		"  1. cd " + projectName + "\n" +
		"  2. moai cc\n" +
		"  3. /moai plan \"describe your first feature\""

	body := countRow + "\n\n" + nextSteps
	if warnCount > 0 {
		body += "\n\n" + paintToken(
			fmt.Sprintf("%d warning(s) collected — see the warning summary on stderr below", warnCount),
			th.Warning, false)
	}

	return tui.Box(tui.BoxOpts{
		Title:  paintToken(tui.StatusIcon("ok")+" MoAI project initialized", th.Success, true),
		Accent: true,
		Body:   body,
		Theme:  &th,
	})
}
