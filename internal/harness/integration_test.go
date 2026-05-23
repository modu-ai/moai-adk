// Package harness — sample session replay integration tests.
// REQ-HL-001: replay an event sequence resembling a real session to verify the full pipeline.
package harness

import (
	"encoding/json"
	"runtime"
	"testing"
	"time"
)

// TestIntegration_SampleSessionReplay replays an event sequence that occurs in a
// real /moai session to verify the observer + retention pipeline.
// T-P1-05: integration_test.go sample session replay.
func TestIntegration_SampleSessionReplay(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := dir + "/.moai/harness/usage-log.jsonl"
	archiveDir := dir + "/.moai/harness/learning-history/archive"

	// Use a fixed time (test determinism)
	baseTime := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	nowIdx := 0
	times := []time.Time{
		baseTime,
		baseTime.Add(1 * time.Second),
		baseTime.Add(2 * time.Second),
		baseTime.Add(3 * time.Second),
		baseTime.Add(4 * time.Second),
	}
	nowFn := func() time.Time {
		t := times[nowIdx%len(times)]
		nowIdx++
		return t
	}

	retention := NewRetention(logPath, archiveDir, func() time.Time { return baseTime })
	obs := &Observer{
		logPath:   logPath,
		retention: retention,
		nowFn:     nowFn,
	}

	// Sample session event sequence
	events := []struct {
		et      EventType
		subject string
		hash    string
	}{
		{EventTypeMoaiSubcommand, "/moai plan", "ctx-abc"},
		{EventTypeSpecReference, "SPEC-V3R3-HARNESS-LEARNING-001", "ctx-abc"},
		{EventTypeAgentInvocation, "manager-spec", "ctx-abc"},
		{EventTypeAgentInvocation, "plan-auditor", "ctx-abc"},
		{EventTypeFeedback, "/moai feedback", "ctx-abc"},
	}

	for _, e := range events {
		if err := obs.RecordEvent(e.et, e.subject, e.hash); err != nil {
			t.Fatalf("RecordEvent(%s) failed: %v", e.et, err)
		}
	}

	// Verify the log file
	data, err := readFile(logPath)
	if err != nil {
		t.Fatalf("log file read failed: %v", err)
	}

	lines := splitNonEmptyLines(string(data))
	if len(lines) != len(events) {
		t.Errorf("recorded event count: want=%d got=%d", len(events), len(lines))
	}

	// Verify each line is valid JSON and contains the required fields
	for i, line := range lines {
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Errorf("line %d parse failed: %v", i, err)
			continue
		}
		if evt.SchemaVersion != LogSchemaVersion {
			t.Errorf("line %d SchemaVersion: want=%q got=%q", i, LogSchemaVersion, evt.SchemaVersion)
		}
		if evt.EventType == "" {
			t.Errorf("line %d EventType is empty", i)
		}
		if evt.Subject == "" {
			t.Errorf("line %d Subject is empty", i)
		}
		if evt.Timestamp.IsZero() {
			t.Errorf("line %d Timestamp is zero", i)
		}
	}
}

// TestIntegration_RetentionWithObserver verifies the full flow where Observer integrates
// with retention to prune stale events.
// T-P1-05: integration test.
func TestIntegration_RetentionWithObserver(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipped because pruning is not reflected immediately on Windows due to file locking")
	}

	t.Parallel()

	dir := t.TempDir()
	logPath := dir + "/usage-log.jsonl"
	archiveDir := dir + "/archive"

	// Pre-record an old event
	past := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	oldEvent := Event{
		Timestamp:     past,
		EventType:     EventTypeMoaiSubcommand,
		Subject:       "old-subcommand",
		ContextHash:   "old-hash",
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
	}
	if err := appendEventsJSONL(logPath, []Event{oldEvent}); err != nil {
		t.Fatalf("recording the old event failed: %v", err)
	}

	// Observer + Retention integration: now is 2026-04-27
	now := time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC)
	retention := NewRetention(logPath, archiveDir, func() time.Time { return now })

	// With defaultRetentionDays (30 days), 2026-01-01 is a pruning target
	obs := NewObserverWithRetention(logPath, retention)

	// Record a new event — lazy pruning runs at this point
	if err := obs.RecordEvent(EventTypeAgentInvocation, "expert-backend", "new-hash"); err != nil {
		t.Fatalf("RecordEvent failed: %v", err)
	}

	// Verify the old event was removed
	data, err := readFile(logPath)
	if err != nil {
		t.Fatalf("log file read failed: %v", err)
	}

	lines := splitNonEmptyLines(string(data))
	// Only the 1 new event must remain
	if len(lines) != 1 {
		t.Errorf("post-pruning line count: want=1 got=%d", len(lines))
	}

	// Verify an archive file was created
	archiveFiles := listDir(archiveDir)
	if len(archiveFiles) == 0 {
		t.Error("archive file was not created")
	}
}

// TestIntegration_JSONL_AllEventTypes verifies the round-trip of recording and parsing every EventType.
func TestIntegration_JSONL_AllEventTypes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	obs := NewObserver(dir + "/usage-log.jsonl")

	allTypes := []EventType{
		EventTypeMoaiSubcommand,
		EventTypeAgentInvocation,
		EventTypeSpecReference,
		EventTypeFeedback,
	}

	for _, et := range allTypes {
		if err := obs.RecordEvent(et, "test-subject", "test-hash"); err != nil {
			t.Errorf("RecordEvent(%s) failed: %v", et, err)
		}
	}

	data, err := readFile(dir + "/usage-log.jsonl")
	if err != nil {
		t.Fatalf("log file read failed: %v", err)
	}

	lines := splitNonEmptyLines(string(data))
	if len(lines) != len(allTypes) {
		t.Errorf("event count: want=%d got=%d", len(allTypes), len(lines))
	}

	for i, line := range lines {
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Errorf("line %d parse failed: %v", i, err)
			continue
		}
		if evt.EventType != allTypes[i] {
			t.Errorf("line %d EventType: want=%q got=%q", i, allTypes[i], evt.EventType)
		}
	}
}
