// Package codexadapter translates between the Codex hook surface and MoAI's
// own hook dispatcher.
//
// It exists because two things differ between Claude Code and codex-cli: the
// event name the harness passes, and three output keys Codex declares but does
// not act on. Everything else measured identical — payload field names are
// snake_case in both, and the observed key sets match the captured goldens — so
// this package sits in FRONT of the dispatcher rather than inside it, and
// nothing under internal/hook is modified (SPEC-CODEX-HOOK-ADAPTER-001 REQ-7).
//
// Measurement basis: codex-cli 0.147.0. See
// .moai/reports/t83/precondition-measurement.md and -round3.md.
package codexadapter

import (
	"errors"
	"fmt"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// EventRow is one row of the Codex-to-MoAI event mapping.
type EventRow struct {
	// CodexEvent is the event name Codex passes, matching the value it also
	// writes as hook_event_name in the payload.
	CodexEvent hook.EventType

	// DispatcherArg is the `moai hook <arg>` subcommand registered in
	// internal/cli/hook.go.
	DispatcherArg string

	// Adapted reports whether this milestone adapts the event. Rows with
	// Adapted false are recognized and refused rather than omitted — see
	// Resolve.
	Adapted bool
}

// EventTable is the complete Codex event set and its dispatcher counterparts.
//
// All eleven Codex events have a counterpart: MoAI's dispatcher registers a
// subcommand for each. Excluding an event from adaptation is therefore a
// scoping decision about measurement coverage, never an absence of a
// counterpart — an earlier draft of the SPEC asserted the absence and was
// wrong.
//
// Six rows are adapted: the events with both a payload capture and observed
// behavior. Four are held back for lack of any measurement, and SubagentStop is
// held back because it was measured NOT to fire — delegation surfaces as
// PostToolUse with a tool_name beginning "collaboration", so mapping it would
// wire a dead path.
var EventTable = []EventRow{
	{hook.EventPreToolUse, "pre-tool", true},
	{hook.EventPostToolUse, "post-tool", true},
	{hook.EventSessionStart, "session-start", true},
	{hook.EventSessionEnd, "session-end", true},
	{hook.EventStop, "stop", true},
	{hook.EventUserPromptSubmit, "user-prompt-submit", true},

	{hook.EventPreCompact, "compact", false},
	{hook.EventPostCompact, "post-compact", false},
	{hook.EventPermissionRequest, "permission-request", false},
	{hook.EventSubagentStart, "subagent-start", false},
	{hook.EventSubagentStop, "subagent-stop", false},
}

// ErrUnknownEvent marks a name absent from EventTable.
var ErrUnknownEvent = errors.New("unknown codex hook event")

// ErrUnadapted marks a name present in EventTable but not adapted by this
// milestone.
var ErrUnadapted = errors.New("codex hook event recognized but not adapted")

// Resolve maps a Codex event name to its dispatcher argument.
//
// The two refusal paths are deliberately distinct. Codex silently ignores
// unknown event names in its own config, so an adapter that also defaulted
// quietly would leave a hook that appears installed and never fires; and an
// unadapted-but-recognized event must be distinguishable from a typo, or the
// operator cannot tell a scoping decision from a mistake.
func Resolve(codexEvent string) (string, error) {
	for _, row := range EventTable {
		if string(row.CodexEvent) != codexEvent {
			continue
		}
		if !row.Adapted {
			return "", fmt.Errorf("%w: %q (dispatcher arg %q exists; this milestone does not adapt it)",
				ErrUnadapted, codexEvent, row.DispatcherArg)
		}
		return row.DispatcherArg, nil
	}
	return "", fmt.Errorf("%w: %q", ErrUnknownEvent, codexEvent)
}

// IsUnknownEvent reports whether err came from a name absent from EventTable.
func IsUnknownEvent(err error) bool { return errors.Is(err, ErrUnknownEvent) }

// IsUnadapted reports whether err came from a recognized but unadapted event.
func IsUnadapted(err error) bool { return errors.Is(err, ErrUnadapted) }
