package project

import "fmt"

// ProgressReporter reports progress during project initialization.
// Implemented by UI components to show real-time status updates.
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

func (r *NoOpReporter) StepStart(name, message string)  {}
func (r *NoOpReporter) StepUpdate(message string)         {}
func (r *NoOpReporter) StepComplete(message string)      {}
func (r *NoOpReporter) StepError(err error)             {}

// ConsoleReporter is a ProgressReporter that outputs to console.
type ConsoleReporter struct{}

// NewConsoleReporter creates a new ConsoleReporter.
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{}
}

func (r *ConsoleReporter) StepStart(name, message string) {
	if message != "" {
		fmt.Printf("  ○ %s: %s...\n", name, message)
	} else {
		fmt.Printf("  ○ %s...\n", name)
	}
}

func (r *ConsoleReporter) StepUpdate(message string) {
	fmt.Printf("    %s\n", message)
}

func (r *ConsoleReporter) StepComplete(message string) {
	if message != "" {
		fmt.Printf("\r  ✓ %s: %s\n", message, "completed")
	} else {
		fmt.Println("\r  ✓ Completed")
	}
}

func (r *ConsoleReporter) StepError(err error) {
	fmt.Printf("\r  ✗ Error: %v\n", err)
}
