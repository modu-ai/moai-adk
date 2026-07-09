// failure_observer.go wires FAILURE signals (tool failures, test failures) into
// the harness learning loop by recording them as usage-log.jsonl events through
// the existing harness.Observer (SPEC-HARNESS-RATCHET-REWIRE-001 REQ-HRR-001 /
// REQ-HRR-002).
//
// Prior to this file the PostToolUseFailure handler and the evidence-writer
// test-fail detection existed but fed NO classifier/proposal path — failures,
// the highest-information learning signal, were invisible to the ratchet. This
// file is the write-side seam that makes them observable. The read-side
// (AggregatePatterns) already consumes any event_type recorded here because the
// new EventTypeToolFailure / EventTypeTestFail constants were added to
// PatternBearingEventTypes.
//
// Fail-open discipline: all errors are logged via slog.Warn and never returned.
// A hook handler MUST NOT block the user's session on a learning-loop write
// failure (mirrors the evidence_writer.go logEvidence contract).
package hook

import (
	"log/slog"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// unknownToolSubject is the non-empty placeholder substituted when a
// PostToolUseFailure event arrives with an empty tool name (EC-1). It prevents
// a degenerate tool_failure::<sig> pattern key — the empty-subject class that
// the M2 eligibility predicate (REQ-HRR-003) would otherwise exclude. By
// keeping the subject non-empty the event remains a valid (if coarse) failure
// signal rather than degenerate lifecycle noise.
const unknownToolSubject = "unknown-tool"

// recordToolFailureEvent records a tool_failure:<tool>:<signature> event to
// usage-log.jsonl via the harness observer. The signature (ContextHash slot)
// is the low-cardinality error-class token (D1 resolution) — NOT a hash of the
// full error text, which would explode key cardinality and defeat pattern
// aggregation.
//
// Fail-open: errors are logged and swallowed; the PostToolUseFailure handler
// must never block the user's session end.
func recordToolFailureEvent(input *HookInput, category ErrorCategory) {
	root := resolveProjectRoot(input)
	if root == "" {
		return // no project root — silent skip (REQ-SEW-013 graceful degradation).
	}

	tool := input.ToolName
	if tool == "" {
		tool = unknownToolSubject // EC-1: never emit a degenerate tool_failure::<sig> key.
	}

	logPath := filepath.Join(root, ".moai", "harness", "usage-log.jsonl")
	obs := harness.NewObserver(logPath)

	evt := harness.Event{
		EventType:  harness.EventTypeToolFailure,
		Subject:    tool,
		ContextHash: string(category), // low-cardinality error-class token (D1)
	}
	if err := obs.RecordExtendedEvent(evt); err != nil {
		slog.Warn("failure observer: failed to record tool_failure event",
			"tool", tool,
			"category", category,
			"session_id", input.SessionID,
			"error", err,
		)
	}
}

// recordTestFailEvent records a test_fail:<package>: event to usage-log.jsonl
// via the harness observer. The ContextHash slot is empty by design — the
// package alone identifies the failure pattern (REQ-HRR-002 key form
// test_fail:<package>:). The package is a low-cardinality identifier derived
// from the test command by extractTestPackage.
//
// Fail-open: errors are logged and swallowed.
func recordTestFailEvent(input *HookInput, pkg string) {
	root := resolveProjectRoot(input)
	if root == "" {
		return
	}

	logPath := filepath.Join(root, ".moai", "harness", "usage-log.jsonl")
	obs := harness.NewObserver(logPath)

	evt := harness.Event{
		EventType: harness.EventTypeTestFail,
		Subject:   pkg,
		// ContextHash intentionally empty (REQ-HRR-002 key form test_fail:<pkg>:)
	}
	if err := obs.RecordExtendedEvent(evt); err != nil {
		slog.Warn("failure observer: failed to record test_fail event",
			"package", pkg,
			"session_id", input.SessionID,
			"error", err,
		)
	}
}
