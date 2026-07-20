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
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// REQ-HRR-006: append a structured lesson stub to the repo-local inbox so
	// the orchestrator's Lessons Protocol can drain it into auto-memory.
	eventKey := string(harness.EventTypeToolFailure) + ":" + tool + ":" + string(category)
	appendLessonsInboxStub(root, eventKey, truncateSummary(input.Error, string(category)), "tool:"+tool)
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

	// REQ-HRR-006: append a structured lesson stub for the test-failure path.
	eventKey := string(harness.EventTypeTestFail) + ":" + pkg + ":"
	appendLessonsInboxStub(root, eventKey, "test failure in package "+pkg, "test:"+pkg)
}

// lessonsInboxStub is the JSONL schema for .moai/lessons-inbox.jsonl entries
// (REQ-HRR-006 / D3: append-only JSONL, minimum fields). The orchestrator's
// Lessons Protocol drains these stubs into auto-memory lesson entries.
type lessonsInboxStub struct {
	Timestamp string `json:"timestamp"`
	EventKey  string `json:"event_key"`
	Summary   string `json:"summary"`
	Source    string `json:"source"`
}

// appendLessonsInboxStub appends one structured stub to
// .moai/lessons-inbox.jsonl (REQ-HRR-006). The file is append-only JSONL
// (EC-4: concurrent Stop hooks tolerate interleaving; no read-modify-write).
// Permissions 0o600 are consistent with sibling state files. Fail-open: errors
// are logged and swallowed (a learning-loop write must never block the session).
func appendLessonsInboxStub(root, eventKey, summary, source string) {
	inboxPath := filepath.Join(root, ".moai", "lessons-inbox.jsonl")
	stub := lessonsInboxStub{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		EventKey:  eventKey,
		Summary:   summary,
		Source:    source,
	}
	data, err := json.Marshal(stub)
	if err != nil {
		slog.Warn("failure observer: marshal lessons-inbox stub", "error", err)
		return
	}
	data = append(data, '\n')
	// Auto-create parent directory (.moai/) if absent.
	if dir := filepath.Dir(inboxPath); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	f, err := os.OpenFile(inboxPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		slog.Warn("failure observer: open lessons-inbox", "error", err)
		return
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(data); err != nil {
		slog.Warn("failure observer: write lessons-inbox", "error", err)
	}
}

// truncateSummary produces a bounded human-readable failure summary for the
// lessons-inbox stub. It prefers the error text (truncated to 200 chars on a
// rune boundary) and falls back to the category token when the error is empty.
func truncateSummary(errorText, fallback string) string {
	errorText = strings.TrimSpace(errorText)
	if errorText == "" {
		return fallback
	}
	runes := []rune(errorText)
	if len(runes) > 200 {
		return string(runes[:200]) + "…"
	}
	return errorText
}
