package telemetry

import (
	"testing"
)

// TestDetermineOutcome_DefaultIsUnknown verifies that the conservative default
// outcome is OutcomeUnknown when no signal events are present.
func TestDetermineOutcome_DefaultIsUnknown(t *testing.T) {
	t.Parallel()

	got := DetermineOutcome(nil)
	if got != OutcomeUnknown {
		t.Errorf("DetermineOutcome(nil) = %q, want %q", got, OutcomeUnknown)
	}

	got = DetermineOutcome([]Event{})
	if got != OutcomeUnknown {
		t.Errorf("DetermineOutcome([]) = %q, want %q", got, OutcomeUnknown)
	}
}

// TestDetermineOutcome_ErrorToolsAfterSkill verifies that error-related tool
// usage after skill invocation results in OutcomeError.
func TestDetermineOutcome_ErrorToolsAfterSkill(t *testing.T) {
	t.Parallel()

	events := []Event{
		{ToolName: "Bash", IsError: true},
	}

	got := DetermineOutcome(events)
	if got != OutcomeError {
		t.Errorf("DetermineOutcome(error events) = %q, want %q", got, OutcomeError)
	}
}

// TestDetermineOutcome_AllTestsPassingMeansSuccess verifies that a session
// where all tests pass (and no errors) results in OutcomeSuccess.
func TestDetermineOutcome_AllTestsPassingMeansSuccess(t *testing.T) {
	t.Parallel()

	events := []Event{
		{ToolName: "Bash", IsTestPass: true},
	}

	got := DetermineOutcome(events)
	if got != OutcomeSuccess {
		t.Errorf("DetermineOutcome(all tests pass) = %q, want %q", got, OutcomeSuccess)
	}
}

// TestDetermineOutcome_MixedSignalsMeansPartial verifies that mixed signals
// (some tests pass, some fail) result in OutcomePartial.
func TestDetermineOutcome_MixedSignalsMeansPartial(t *testing.T) {
	t.Parallel()

	events := []Event{
		{ToolName: "Bash", IsTestPass: true},
		{ToolName: "Bash", IsTestFail: true},
	}

	got := DetermineOutcome(events)
	if got != OutcomePartial {
		t.Errorf("DetermineOutcome(mixed signals) = %q, want %q", got, OutcomePartial)
	}
}

// TestDetermineOutcome_ErrorDominatesSuccess verifies that when both error
// and success signals are present, error takes precedence.
func TestDetermineOutcome_ErrorDominatesSuccess(t *testing.T) {
	t.Parallel()

	events := []Event{
		{ToolName: "Bash", IsTestPass: true},
		{ToolName: "Bash", IsError: true},
	}

	got := DetermineOutcome(events)
	// Error signal with test pass -> partial (not pure error)
	if got == OutcomeUnknown {
		t.Errorf("DetermineOutcome(error+pass) = %q, should not be unknown", got)
	}
}

// TestDetermineOutcome_NoSignalMeansUnknown verifies that events without
// any signal flag produce OutcomeUnknown.
func TestDetermineOutcome_NoSignalMeansUnknown(t *testing.T) {
	t.Parallel()

	events := []Event{
		{ToolName: "Read"},
		{ToolName: "Grep"},
		{ToolName: "Glob"},
	}

	got := DetermineOutcome(events)
	if got != OutcomeUnknown {
		t.Errorf("DetermineOutcome(no signal events) = %q, want %q", got, OutcomeUnknown)
	}
}
