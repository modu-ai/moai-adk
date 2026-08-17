// Package harness — observer and JSONL schema unit tests.
package harness

import (
	"bytes"
	"encoding/json"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/timing"
)

// ─────────────────────────────────────────────
// T-P1-01: Event marshal/unmarshal round-trip tests
// ─────────────────────────────────────────────

// TestEventMarshalUnmarshal verifies that the Event struct can be JSON-serialized and deserialized.
// REQ-HL-001: each JSONL line must contain timestamp, event_type, subject, context_hash,
// and tier_increment fields.
func TestEventMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	original := Event{
		Timestamp:     time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
		EventType:     EventTypeMoaiSubcommand,
		Subject:       "/moai plan",
		ContextHash:   "abc123",
		TierIncrement: 1,
		SchemaVersion: LogSchemaVersion,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal 실패: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal 실패: %v", err)
	}

	if decoded.EventType != original.EventType {
		t.Errorf("EventType: want=%q got=%q", original.EventType, decoded.EventType)
	}
	if decoded.Subject != original.Subject {
		t.Errorf("Subject: want=%q got=%q", original.Subject, decoded.Subject)
	}
	if decoded.ContextHash != original.ContextHash {
		t.Errorf("ContextHash: want=%q got=%q", original.ContextHash, decoded.ContextHash)
	}
	if decoded.TierIncrement != original.TierIncrement {
		t.Errorf("TierIncrement: want=%d got=%d", original.TierIncrement, decoded.TierIncrement)
	}
	if decoded.SchemaVersion != LogSchemaVersion {
		t.Errorf("SchemaVersion: want=%q got=%q", LogSchemaVersion, decoded.SchemaVersion)
	}
}

// TestEventTypeValues verifies that the EventType enum values are defined as expected.
func TestEventTypeValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		et   EventType
		want string
	}{
		{EventTypeMoaiSubcommand, "moai_subcommand"},
		{EventTypeAgentInvocation, "agent_invocation"},
		{EventTypeSpecReference, "spec_reference"},
		{EventTypeFeedback, "feedback"},
	}

	for _, tc := range tests {
		if string(tc.et) != tc.want {
			t.Errorf("EventType %q: want string value %q", tc.et, tc.want)
		}
	}
}

// TestLogSchemaVersion verifies the constant value is "v2.1".
// Bumped "v1" → "v2" by SPEC-HARNESS-OUTCOME-CAPTURE-001 (REQ-OC-010) to mark the
// additive apply_outcome event type + its omitempty Event fields.
// Bumped "v2" → "v2.1" by SPEC-V3R6-CONTEXT-GOV-AXIS-001 (REQ-CGA-002) to mark the
// additive eager-vs-on-demand context-weight fields.
func TestLogSchemaVersion(t *testing.T) {
	t.Parallel()

	if LogSchemaVersion != "v2.1" {
		t.Errorf("LogSchemaVersion: want %q got %q", "v2.1", LogSchemaVersion)
	}
}

// ─────────────────────────────────────────────
// T-P1-02: RecordEvent tests
// ─────────────────────────────────────────────

// TestRecordEventWritesJSONL verifies that RecordEvent writes valid JSONL to the file.
// REQ-HL-001: observer runs as a PostToolUse hook handler and must complete in <100ms per event.
func TestRecordEventWritesJSONL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := dir + "/usage-log.jsonl"

	obs := NewObserver(logPath)

	if err := obs.RecordEvent(EventTypeMoaiSubcommand, "/moai plan", "hash001"); err != nil {
		t.Fatalf("RecordEvent 실패: %v", err)
	}
	if err := obs.RecordEvent(EventTypeAgentInvocation, "expert-backend", "hash002"); err != nil {
		t.Fatalf("RecordEvent 두 번째 호출 실패: %v", err)
	}

	// Verify the file was created
	data, err := readFileBytes(logPath)
	if err != nil {
		t.Fatalf("로그 파일 읽기 실패: %v", err)
	}

	// Parse JSONL: verify each line is valid JSON
	lines := splitNonEmptyLines(string(data))
	if len(lines) != 2 {
		t.Errorf("기록된 줄 수: want=2 got=%d", len(lines))
	}

	for i, line := range lines {
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Errorf("줄 %d json 파싱 실패: %v", i, err)
		}
		if evt.SchemaVersion != LogSchemaVersion {
			t.Errorf("줄 %d: SchemaVersion want=%q got=%q", i, LogSchemaVersion, evt.SchemaVersion)
		}
	}
}

// TestRecordEvent100Sequential verifies that 100 consecutive RecordEvent calls
// stay well inside the 5s hook budget. REQ-HL-001: observer must not block the
// parent tool call. Internal for-loop, NOT a Benchmark — the test asserts
// latency bounds over the steady-state samples, not a throughput average. It
// logs the median / p95 / worst / avg and the calibrated ratio so the bounds
// are mechanically observable in output.
//
// Three bounds, each answering a different question (enforced by
// internal/timing.Assert; same rationale as TestBranchGuard_Latency):
//
//   - p95 <= 20% of the 5s hook budget (1s POSIX). The distribution
//     detector. A real regression — an fsync per write, a whole-file rewrite
//     per append, a network call — makes EVERY write slow, so it moves the
//     whole distribution and trips p95. Scheduler jitter moves ONE sample
//     and does not.
//   - worst < 100% of budget (5s). The contract itself: no single write may
//     consume the whole hook budget, because one that does stalls the user's
//     tool call.
//   - median <= 2.0x ONE reference append cycle (open O_APPEND + write one
//     JSONL-sized line + close), measured in this same run on the same
//     filesystem. The CALIBRATED arm — it measures code, not machine.
//     RecordEvent is one such cycle plus sub-millisecond marshal/mkdir, so
//     the healthy ratio sits at ~1.0x on any machine under any load; an
//     fsync per write lands at 5-50x, a whole-file rewrite at 2-3x, and both
//     trip the bound even while absolute figures stay inside the generous
//     budget fractions. This closes the gap the earlier forms left: the
//     100ms-per-sample absolute ceiling measured machine load (a loaded box
//     breached it on healthy code; see the load incidents of 2026-08-15),
//     and the budget fractions alone cannot see a 2x slowdown under 1s. A
//     persisted baseline cannot close it either — the verify snapshot store
//     keys records to exact tree state, and a baseline recorded at another
//     time measures another load state; the in-run reference is the only
//     honest baseline (full rationale: internal/timing package doc).
//
// Windows keeps the existing skip: on GitHub-hosted Windows runners + race
// detector + antivirus, file-write latency reliably exceeded even generous
// ceilings in past observations (CIAUT Wave 6); the budget fractions were
// not measured there.
func TestRecordEvent100Sequential(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows runner file-write latency is unmeasured against budget fractions — Linux/macOS only")
	}
	t.Parallel()

	dir := t.TempDir()
	obs := NewObserver(dir + "/usage-log.jsonl")

	// Reference unit: ONE append cycle of the same cost class RecordEvent
	// performs — open(O_APPEND|O_CREATE|O_WRONLY) + write one JSONL-sized
	// line + close — on a sibling file in the same t.TempDir filesystem,
	// measured immediately before the measured loop.
	refPath := dir + "/reference-append.jsonl"
	refLine := append(bytes.Repeat([]byte("x"), 159), '\n')
	refUnit := timing.Median(func() {
		f, err := os.OpenFile(refPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatalf("reference append open failed: %v", err)
		}
		if _, err := f.Write(refLine); err != nil {
			_ = f.Close()
			t.Fatalf("reference append write failed: %v", err)
		}
		if err := f.Close(); err != nil {
			t.Fatalf("reference append close failed: %v", err)
		}
	}, 3, 20)

	const hookBudget = 5 * time.Second
	steadyCeiling := hookBudget / 5 // 20% — 1s on POSIX

	// Healthy ratio is ~1.0x (one append cycle + sub-ms marshal). MaxUnits
	// 2.0 keeps headroom over healthy code while tripping on fsync-per-write
	// (5-50x) and whole-file rewrite (2-3x) — regression classes invisible
	// to the budget fractions when absolute figures stay generous.
	timing.Assert(t, timing.Bound{
		Name:          "RecordEvent",
		Budget:        hookBudget,
		SteadyCeiling: steadyCeiling,
		MaxUnits:      2.0,
		Iterations:    100,
		Warmup:        5,
	}, refUnit, func() {
		if err := obs.RecordEvent(EventTypeMoaiSubcommand, "test-subject", "hash"); err != nil {
			t.Fatalf("RecordEvent 실패: %v", err)
		}
	})
}

// TestRecordEventAppends verifies that new events are appended to an existing file.
func TestRecordEventAppends(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := dir + "/usage-log.jsonl"
	obs := NewObserver(logPath)

	for range 5 {
		if err := obs.RecordEvent(EventTypeSpecReference, "SPEC-001", "h"); err != nil {
			t.Fatalf("RecordEvent 실패: %v", err)
		}
	}

	data, _ := readFileBytes(logPath)
	lines := splitNonEmptyLines(string(data))
	if len(lines) != 5 {
		t.Errorf("append 후 줄 수: want=5 got=%d", len(lines))
	}
}

// ─────────────────────────────────────────────
// T-P1-04: PruneStaleEntries tests
// ─────────────────────────────────────────────

// TestPruneStaleEntriesRemovesOldEvents verifies that events older than retentionDays
// are removed and added to the archive. REQ-HL-011.
func TestPruneStaleEntriesRemovesOldEvents(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := dir + "/usage-log.jsonl"

	// Test events: 2 stale, 1 fresh
	now := time.Now().UTC()
	old1 := Event{
		Timestamp:     now.AddDate(0, 0, -10), // 10 days ago
		EventType:     EventTypeMoaiSubcommand,
		Subject:       "old1",
		ContextHash:   "h1",
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
	}
	old2 := Event{
		Timestamp:     now.AddDate(0, 0, -8), // 8 days ago
		EventType:     EventTypeAgentInvocation,
		Subject:       "old2",
		ContextHash:   "h2",
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
	}
	fresh := Event{
		Timestamp:     now.AddDate(0, 0, -1), // 1 day ago (fresh)
		EventType:     EventTypeFeedback,
		Subject:       "fresh",
		ContextHash:   "h3",
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
	}

	// Write events directly to the file
	if err := writeEventsToFile(logPath, []Event{old1, old2, fresh}); err != nil {
		t.Fatalf("테스트 데이터 기록 실패: %v", err)
	}

	archiveDir := dir + "/archive"
	retention := NewRetention(logPath, archiveDir, func() time.Time { return now })

	if err := retention.PruneStaleEntries(7); err != nil { // 7-day retention
		t.Fatalf("PruneStaleEntries 실패: %v", err)
	}

	// Only the fresh event must remain in the log file
	data, err := readFileBytes(logPath)
	if err != nil {
		t.Fatalf("로그 파일 읽기 실패: %v", err)
	}
	lines := splitNonEmptyLines(string(data))
	if len(lines) != 1 {
		t.Errorf("prune 후 줄 수: want=1 got=%d", len(lines))
	}
	if len(lines) == 1 {
		var evt Event
		if err := json.Unmarshal([]byte(lines[0]), &evt); err != nil {
			t.Errorf("남은 이벤트 파싱 실패: %v", err)
		}
		if evt.Subject != "fresh" {
			t.Errorf("남은 이벤트 subject: want=fresh got=%q", evt.Subject)
		}
	}

	// Verify an archive file was created (gzip)
	archiveFiles := listFilesInDir(archiveDir)
	if len(archiveFiles) == 0 {
		t.Error("아카이브 파일이 생성되지 않았습니다")
	}
}

// TestPruneSkipsIfRecentlyPruned verifies pruning is skipped if it ran less than 1 hour ago.
func TestPruneSkipsIfRecentlyPruned(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := dir + "/usage-log.jsonl"

	now := time.Now().UTC()
	// Record one stale event
	old := Event{
		Timestamp:     now.AddDate(0, 0, -10),
		EventType:     EventTypeMoaiSubcommand,
		Subject:       "old",
		ContextHash:   "h",
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
	}
	if err := writeEventsToFile(logPath, []Event{old}); err != nil {
		t.Fatalf("테스트 데이터 기록 실패: %v", err)
	}

	archiveDir := dir + "/archive"

	// First prune: must succeed
	retention := NewRetention(logPath, archiveDir, func() time.Time { return now })
	if err := retention.PruneStaleEntries(7); err != nil {
		t.Fatalf("첫 번째 PruneStaleEntries 실패: %v", err)
	}

	// Add the event again (stale)
	if err := writeEventsToFile(logPath, []Event{old}); err != nil {
		t.Fatalf("재기록 실패: %v", err)
	}

	// Second prune: must be skipped because it is within 1 hour
	// (reusing the same retention instance — lastPruneAt is set)
	if err := retention.PruneStaleEntries(7); err != nil {
		t.Fatalf("두 번째 PruneStaleEntries 실패: %v", err)
	}

	// The event must still be in the file (pruning was skipped)
	data, _ := readFileBytes(logPath)
	lines := splitNonEmptyLines(string(data))
	if len(lines) != 1 {
		t.Errorf("skip 후 줄 수: want=1 got=%d (pruning이 skip되지 않음)", len(lines))
	}
}

// ─────────────────────────────────────────────
// Helper functions
// ─────────────────────────────────────────────

// readFileBytes reads the file contents as bytes.
func readFileBytes(path string) ([]byte, error) {
	// os is already imported in the harness package (from observer.go)
	return readFile(path)
}

// splitNonEmptyLines splits a string by newlines and drops empty lines.
func splitNonEmptyLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if line != "" {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	if start < len(s) {
		if line := s[start:]; line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// writeEventsToFile writes a slice of events to the file in JSONL format.
func writeEventsToFile(path string, events []Event) error {
	return appendEventsJSONL(path, events)
}

// listFilesInDir returns the list of files in the directory.
func listFilesInDir(dir string) []string {
	return listDir(dir)
}
