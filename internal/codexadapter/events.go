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
// Measurement basis: codex-cli 0.147.0
// (.moai/reports/t83/precondition-measurement.md and -round3.md), re-measured
// on codex-cli 0.153.4 by the t496 firing campaign
// (.moai/reports/t496/codex-event-campaign.md).
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

// CodexEventInterrupt is the Codex-only interrupt event. It is defined in
// this package rather than internal/hook because it has no Claude-side hook
// counterpart — internal/hook is the Claude-side dispatcher vocabulary, and
// nothing under internal/hook is modified (REQ-7,
// SPEC-CODEX-HOOK-ADAPTER-001; SPEC-CODEX-EVENT-COVERAGE-001 REQ-CEV-002).
const CodexEventInterrupt hook.EventType = "Interrupt"

// EventTable is the complete Codex event set and its dispatcher counterparts.
//
// The table carries all twelve documented Codex hook events. Eleven of them
// have a MoAI dispatcher counterpart: MoAI's dispatcher registers a
// subcommand for each. Interrupt is the exception — it is Codex-only and has
// no MoAI dispatcher counterpart, so its DispatcherArg is the empty string,
// the marker for "no counterpart".
//
// Eight rows are adapted: the six with a payload capture and observed behavior
// on the 0.147.0 basis, plus SubagentStart/SubagentStop, which the 0.153.4
// campaign measured FIRING with payloads captured (agent_id / agent_type on
// both; SubagentStop additionally carries agent_transcript_path,
// stop_hook_active, last_assistant_message — it reverses the 0.147.0
// observation that SubagentStop never fires).
//
// PreCompact/PostCompact are held back: compaction could not be triggered in
// a non-interactive run (264,808 input tokens max under the 1,048,576-char
// input cap produced no compaction) — trigger-not-achieved, not not-fired.
// PermissionRequest is held back on the same distinction: three non-interactive
// configurations never raised an approval request (interactive TUI untested).
// Interrupt is held back because it has no MoAI dispatcher counterpart to map
// to, even though it was measured firing on SIGINT.
var EventTable = []EventRow{
	{hook.EventPreToolUse, "pre-tool", true},
	{hook.EventPostToolUse, "post-tool", true},
	{hook.EventSessionStart, "session-start", true},
	{hook.EventSessionEnd, "session-end", true},
	{hook.EventStop, "stop", true},
	{hook.EventUserPromptSubmit, "user-prompt-submit", true},
	{hook.EventSubagentStart, "subagent-start", true},
	{hook.EventSubagentStop, "subagent-stop", true},

	{hook.EventPreCompact, "compact", false},
	{hook.EventPostCompact, "post-compact", false},
	{hook.EventPermissionRequest, "permission-request", false},

	{CodexEventInterrupt, "", false},
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
			if row.DispatcherArg == "" {
				// No MoAI dispatcher counterpart (Interrupt): asserting a
				// dispatcher arg exists would be false.
				return "", fmt.Errorf("%w: %q (no MoAI dispatcher counterpart; this milestone does not adapt it)",
					ErrUnadapted, codexEvent)
			}
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
