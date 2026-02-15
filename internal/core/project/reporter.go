package project

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

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

func (r *NoOpReporter) StepStart(name, message string) {}
func (r *NoOpReporter) StepUpdate(message string)      {}
func (r *NoOpReporter) StepComplete(message string)    {}
func (r *NoOpReporter) StepError(err error)            {}

// ConsoleReporter is a ProgressReporter that outputs to console.
type ConsoleReporter struct{}

// NewConsoleReporter creates a new ConsoleReporter.
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{}
}

// Reporter styles for colored status icons.
var (
	reporterSuccess = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#10B981"})
	reporterError   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#EF4444"})
	reporterMuted   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"})
)

func (r *ConsoleReporter) StepStart(name, message string) {
	icon := reporterMuted.Render("\u25CB")
	if message != "" {
		fmt.Printf("  %s %s: %s...\n", icon, name, message)
	} else {
		fmt.Printf("  %s %s...\n", icon, name)
	}
}

func (r *ConsoleReporter) StepUpdate(message string) {
	fmt.Printf("    %s\n", message)
}

func (r *ConsoleReporter) StepComplete(message string) {
	icon := reporterSuccess.Render("\u2713")
	if message != "" {
		fmt.Printf("\r  %s %s: %s\n", icon, message, "completed")
	} else {
		fmt.Printf("\r  %s Completed\n", icon)
	}
}

func (r *ConsoleReporter) StepError(err error) {
	icon := reporterError.Render("\u2717")
	fmt.Printf("\r  %s Error: %v\n", icon, err)
}
