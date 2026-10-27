package cli

// checklist_init.go — the glue between `moai init` and the group checklist
// (SPEC-TUI-CHECKLIST-001, design.md §3.3.1, §3.7). It keeps the call sites in
// init.go and in the preserved-line files one line long.

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// openOrSkipItem opens group/item, or records it as skipped when skip is true
// (a conditional step the run did not take, REQ-TCL-004).
func (s *checklistSession) openOrSkipItem(group, item string, skip bool) {
	if skip {
		s.skipItem(group, item)
		return
	}
	s.openItem(group, item)
}

// preserveAll marks every line either session writer receives until the
// returned function is called as a preserved line of the given kind. It is a
// no-op on an inactive session.
func (s *checklistSession) preserveAll(kind preservedKind) (end func()) {
	if !s.isActive() {
		return func() {}
	}
	endOut := beginPreserved(s.outW, kind)
	endErr := beginPreserved(s.errW, kind)
	return func() {
		endErr()
		endOut()
	}
}

// preservePrinterLines is preserveAll for code that only holds a printer: when
// the printer is the session's warning collector, its lines are marked.
func preservePrinterLines(p printer.Printer, kind preservedKind) (end func()) {
	if cs, ok := p.(interface{ checklistSession() *checklistSession }); ok {
		return cs.checklistSession().preserveAll(kind)
	}
	return func() {}
}

// writeWarningLine writes one warning line to w. While a checklist item is
// open and w is a session writer, the line becomes that item's warning instead;
// otherwise it is written exactly as before.
func writeWarningLine(w io.Writer, line string) {
	if r, ok := w.(warningReceiver); ok && r.ReceiveWarning(strings.TrimLeft(line, " ")) {
		return
	}
	_, _ = fmt.Fprintln(w, line)
}

// renderInitCompletion ends a successful init: with the checklist active it
// draws the summary block, otherwise it prints the legacy completion card.
// The counts are the ones the card always reported (decision D-9).
func renderInitCompletion(cmd interface{ ErrOrStderr() io.Writer }, sess *checklistSession, p *warnCollector, projectName, projectRoot string, dirs, files int, deployMode string) {
	name := projectName
	if name == "" {
		name = filepath.Base(projectRoot)
	}
	if sess.isActive() {
		sess.summary(tui.ChecklistSummary{
			CreatedFiles: &files,
			CreatedDirs:  &dirs,
			Warnings:     p.Count(),
			DeployMode:   deployMode,
			NextSteps:    "cd " + name + " → moai cc → /moai plan \"…\"",
		})
		return
	}
	_, _ = fmt.Fprintln(cmd.ErrOrStderr())
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), buildInitSuccessCard(name, dirs, files, p.Count(), deployMode))
}
