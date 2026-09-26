package cli

// hook_codex_failclosed.go — SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2c
// (REQ-HPR-009, AC-HPR-008 unit leg). Under `--harness codex`, a handler that
// times out, errors, or yields output the adapter cannot parse on a
// decision-bearing event is answered with a fail-closed deny rather than with
// a non-zero exit and an empty stdout, which Codex may resolve as allow. The
// seam stays in the CLI layer; nothing under internal/hook is modified
// (SPEC-CODEX-HOOK-ADAPTER-001 REQ-7).

import (
	"fmt"
	"os"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// hookFaultDiscardKey marks a fault record in the adapter's diagnostic sink.
const hookFaultDiscardKey = "hook-fault"

// isCodexDecisionBearing reports whether a fault on event must be answered
// fail-closed under --harness codex.
func isCodexDecisionBearing(event hook.EventType) bool {
	return codexadapter.IsDecisionBearing(event)
}

// @MX:WARN: [AUTO] fail-closed fault path — a fault on a decision-bearing event becomes a deny on exit 0; on Stop that is a block, so a handler that fails on every turn keeps the Codex turn going
// @MX:REASON: [AUTO] REQ-HPR-009 forbids resolving a fault as allow; the Stop-side loop bound (stop_hook_active, the §D3.8 cap) belongs to the M2d Stop chain, not to this writer
// writeCodexFailClosed writes the Codex fatal_error translation for event —
// a deny carrying the cause — to stdout, records the fault in the adapter's
// sink so the deny is visible rather than silent, and mirrors a line to
// stderr. It returns nil so the process exits 0 and Codex reads the deny.
func writeCodexFailClosed(event hook.EventType, cause error) error {
	causeText := ""
	if cause != nil {
		causeText = cause.Error()
	}
	out, _, err := codexadapter.TranslateCodex(event, codexadapter.DecisionFatalError, causeText)
	if err != nil {
		return fmt.Errorf("codex fail-closed for %s (cause: %v): %w", event, cause, err)
	}

	record := []codexadapter.Discard{{
		Event:         event,
		Key:           hookFaultDiscardKey,
		ContentLength: len(causeText),
		Reason:        "hook fault on a decision-bearing event answered with a fail-closed deny",
	}}
	if rerr := codexadapter.RecordDiscards(resolveHookProjectRoot(), record, false, os.Stderr); rerr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "codex harness: record hook fault: %v\n", rerr)
	}

	if _, werr := os.Stdout.Write(append(out, '\n')); werr != nil {
		return fmt.Errorf("write fail-closed hook output: %w", werr)
	}
	return nil
}
