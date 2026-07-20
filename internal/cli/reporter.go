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

// printerReporter forwards ProgressReporter events to a printer StepHandle
// (or, when useSpinner is set, a SpinnerHandle — the animated live-progress
// surface used by moai init template deployment, SPEC-CLI-TUX-V3-002
// REQ-TUX2-010). PhaseExecutor drives well-paired StepStart -> StepUpdate*
// -> StepComplete/StepError sequences; unpaired calls degrade gracefully to
// direct printer status lines instead of panicking.
type printerReporter struct {
	p          printer.Printer
	useSpinner bool
	step       printer.StepHandle
	spin       printer.SpinnerHandle
}

var _ project.ProgressReporter = (*printerReporter)(nil)

// newPrinterReporter creates a project.ProgressReporter backed by p.
func newPrinterReporter(p printer.Printer) *printerReporter {
	return &printerReporter{p: p}
}

// newSpinnerReporter creates a ProgressReporter whose long-running steps
// render through the animated Spinner handle (REQ-TUX2-010). Non-TTY,
// NO_COLOR and MOAI_REDUCED_MOTION surfaces degrade inside the handle to
// plain single-frame lines (REQ-TUX2-011) — the reporter needs no fallback
// logic of its own.
func newSpinnerReporter(p printer.Printer) *printerReporter {
	return &printerReporter{p: p, useSpinner: true}
}

func (r *printerReporter) StepStart(name, message string) {
	label := name
	if message != "" {
		label = name + ": " + message
	}
	if r.useSpinner {
		r.spin = r.p.Spinner(label + "...")
		return
	}
	r.step = r.p.Step(label + "...")
}

func (r *printerReporter) StepUpdate(message string) {
	if r.useSpinner {
		if r.spin == nil {
			r.p.Info("%s", message)
			return
		}
		r.spin.Update(message)
		return
	}
	if r.step == nil {
		r.p.Info("%s", message)
		return
	}
	r.step.Update("%s", message)
}

func (r *printerReporter) StepComplete(message string) {
	if r.useSpinner {
		if r.spin == nil {
			if message != "" {
				r.p.Success("%s", message)
			}
			return
		}
		if message == "" {
			message = "Completed"
		}
		r.spin.Done("%s", message)
		r.spin = nil
		return
	}
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
	if r.useSpinner {
		if r.spin == nil {
			r.p.Error("%v", err)
			return
		}
		r.spin.Fail("%v", err)
		r.spin = nil
		return
	}
	if r.step == nil {
		r.p.Error("%v", err)
		return
	}
	r.step.Fail(err)
	r.step = nil
}
