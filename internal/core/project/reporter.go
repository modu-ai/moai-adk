package project

// ProgressReporter reports progress during project initialization.
// Implemented by UI components to show real-time status updates.
//
// The console implementation lives on the CLI side (internal/cli
// printerReporter, backed by internal/cli/printer) so that core stays free
// of terminal styling concerns: colours come from internal/tui tokens via
// the printer, and status output is routed to stderr per the CLI output
// streams convention (SPEC-CLI-TUX-V3-001 REQ-CTX-015, REQ-CTX-008).
//
// @MX:ANCHOR: [AUTO] ProgressReporter is the core progress contract; fan_in >= 3
// @MX:REASON: PhaseExecutor drives it and both init/update CLI flows inject
// implementations; changing a method signature breaks all of them at once.
type ProgressReporter interface {
	// StepStart indicates the beginning of a step.
	StepStart(name, message string)

	// StepUpdate provides a status update for the current step.
	StepUpdate(message string)

	// StepComplete marks the current step as successfully completed.
	StepComplete(message string)

	// StepError marks the current step as failed with an error.
	StepError(err error)
}

// NoOpReporter is a ProgressReporter that does nothing (used when no UI is needed).
type NoOpReporter struct{}

func (r *NoOpReporter) StepStart(name, message string) {}
func (r *NoOpReporter) StepUpdate(message string)      {}
func (r *NoOpReporter) StepComplete(message string)    {}
func (r *NoOpReporter) StepError(err error)            {}
