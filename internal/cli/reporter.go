package cli

// printerReporter adapts internal/cli/printer to the core
// project.ProgressReporter contract (SPEC-CLI-TUX-V3-001 REQ-CTX-015,
// design.md §C D-2). It replaces the former project.ConsoleReporter, whose
// direct printf-to-stdout calls and hex AdaptiveColor literals violated
// the channel discipline (REQ-CTX-012) and the token SSOT (REQ-CTX-008).
//
// The adapter lives on the CLI side so the dependency direction stays
// core <- cli (internal/core/project never imports internal/cli/printer).

import (
	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/core/project"
)

// printerReporter forwards ProgressReporter events to a printer StepHandle.
// PhaseExecutor drives well-paired StepStart -> StepUpdate* ->
// StepComplete/StepError sequences; unpaired calls degrade gracefully to
// direct printer status lines instead of panicking.
type printerReporter struct {
	p    printer.Printer
	step printer.StepHandle
}

var _ project.ProgressReporter = (*printerReporter)(nil)

// newPrinterReporter creates a project.ProgressReporter backed by p.
func newPrinterReporter(p printer.Printer) *printerReporter {
	return &printerReporter{p: p}
}

func (r *printerReporter) StepStart(name, message string) {
	label := name
	if message != "" {
		label = name + ": " + message
	}
	r.step = r.p.Step(label + "...")
}

func (r *printerReporter) StepUpdate(message string) {
	if r.step == nil {
		r.p.Info("%s", message)
		return
	}
	r.step.Update("%s", message)
}

func (r *printerReporter) StepComplete(message string) {
	if r.step == nil {
		if message != "" {
			r.p.Success("%s", message)
		}
		return
	}
	if message == "" {
		message = "Completed"
	}
	r.step.Complete("%s", message)
	r.step = nil
}

func (r *printerReporter) StepError(err error) {
	if r.step == nil {
		r.p.Error("%v", err)
		return
	}
	r.step.Fail(err)
	r.step = nil
}
