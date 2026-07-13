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

// buildInitSuccessCard renders the init completion card (REQ-TUX2-016):
// created-artifact counts, the next-action sequence, and — when warnings
// were collected — a one-line pointer to the stderr warning summary.
func buildInitSuccessCard(projectName string, dirs, files, warnCount int) string {
	details := []string{
		uikit.RenderKeyValueLines([]uikit.KVPair{
			{Key: "Directories", Value: fmt.Sprintf("%d created", dirs)},
			{Key: "Files", Value: fmt.Sprintf("%d created", files)},
		}),
		"Next steps:\n" +
			"  1. cd " + projectName + "\n" +
			"  2. moai cc\n" +
			"  3. /moai plan \"describe your first feature\"",
	}
	if warnCount > 0 {
		details = append(details, uikit.WarnStyle.Render(
			fmt.Sprintf("%d warning(s) collected — see the warning summary on stderr below", warnCount)))
	}
	return uikit.RenderSuccessCard("MoAI project initialized", details...)
}
