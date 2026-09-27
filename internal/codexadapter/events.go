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
// The table carries all twelve documented Codex hook events, and every row is
// adapted. Eleven dispatcher args are the `moai hook <arg>` subcommands MoAI
// also registers for Claude Code; Interrupt's `interrupt` arg is Codex-only —
// a subcommand handled entirely in internal/cli that records a user
// cancellation, with no Claude-side counterpart and no constant in
// internal/hook (SPEC-DUAL-HARNESS-HOOK-PARITY-001 design.md §D7; REQ-CEV-002
// kept).
//
// Adaptation history. Six rows were adapted on the codex-cli 0.147.0 basis.
// SubagentStart/SubagentStop joined after the 0.153.4 campaign measured them
// FIRING with payloads captured. SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2e adapted
// the last four: PreCompact/PostCompact (compaction could not be triggered in
// a non-interactive run — trigger-not-achieved, not not-fired),
// PermissionRequest (an approval request was never raised non-interactively;
// the interactive TUI is untested), and Interrupt (measured firing on SIGINT).
// Adapting a row makes MoAI's handler run when Codex fires the event; whether
// Codex fires the three trigger-not-achieved events at all is a live
// measurement this adaptation does not make.
var EventTable = []EventRow{
	{hook.EventPreToolUse, "pre-tool", true},
	{hook.EventPostToolUse, "post-tool", true},
	{hook.EventSessionStart, "session-start", true},
	{hook.EventSessionEnd, "session-end", true},
	{hook.EventStop, "stop", true},
	{hook.EventUserPromptSubmit, "user-prompt-submit", true},
	{hook.EventSubagentStart, "subagent-start", true},
	{hook.EventSubagentStop, "subagent-stop", true},

	{hook.EventPreCompact, "compact", true},
	{hook.EventPostCompact, "post-compact", true},
	{hook.EventPermissionRequest, "permission-request", true},

	{CodexEventInterrupt, CodexInterruptDispatcherArg, true},
}

// CodexInterruptDispatcherArg is the Codex-only `moai hook` subcommand that
// handles the Interrupt event (design.md §D7).
const CodexInterruptDispatcherArg = "interrupt"

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
	return resolveIn(EventTable, codexEvent)
}

// resolveIn is Resolve over an explicit table, so the refusal paths stay
// testable after every shipped row is adapted.
func resolveIn(rows []EventRow, codexEvent string) (string, error) {
	for _, row := range rows {
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
