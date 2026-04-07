package trace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeTraceFile writes entries as JSONL to dir/trace-{sessionID}.jsonl.
func writeTraceFile(t *testing.T, dir, sessionID string, entries []TraceEntry) {
	t.Helper()
	path := filepath.Join(dir, fmt.Sprintf("trace-%s.jsonl", sessionID))
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create trace file: %v", err)
	}
	defer f.Close()

	for _, e := range entries {
		data, err := json.Marshal(e)
		if err != nil {
			t.Fatalf("marshal entry: %v", err)
		}
		fmt.Fprintf(f, "%s\n", data)
	}
}

func TestGenerateSummary_Basic(t *testing.T) {
	dir := t.TempDir()
	sessionID := "summary-basic"
	base := time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC)

	entries := []TraceEntry{
		{Timestamp: base, Event: "PreToolUse", Handler: "h1", Tool: "Bash", DurationMs: 10, Decision: "allow", SessionID: sessionID},
		{Timestamp: base.Add(time.Second), Event: "PostToolUse", Handler: "h2", Tool: "Bash", DurationMs: 5, Decision: "allow", SessionID: sessionID},
		{Timestamp: base.Add(2 * time.Second), Event: "PreToolUse", Handler: "h1", Tool: "Read", DurationMs: 200, Decision: "deny", Reason: "blocked", SessionID: sessionID},
		{Timestamp: base.Add(3 * time.Second), Event: "SessionStart", Handler: "h3", DurationMs: 1, Decision: "allow", SessionID: sessionID},
		{Timestamp: base.Add(4 * time.Second), Event: "PreToolUse", Handler: "h1", Tool: "Write", DurationMs: 50, Decision: "allow", Error: "timeout", SessionID: sessionID},
	}
	writeTraceFile(t, dir, sessionID, entries)

	summary, err := GenerateSummary(dir, sessionID)
	if err != nil {
		t.Fatalf("GenerateSummary: %v", err)
	}

	if summary.TotalHooks != 5 {
		t.Errorf("TotalHooks: want 5, got %d", summary.TotalHooks)
	}
	if summary.EventBreakdown["PreToolUse"] != 3 {
		t.Errorf("EventBreakdown[PreToolUse]: want 3, got %d", summary.EventBreakdown["PreToolUse"])
	}
	if summary.DecisionBreakdown["allow"] != 4 {
		t.Errorf("DecisionBreakdown[allow]: want 4, got %d", summary.DecisionBreakdown["allow"])
	}
	if summary.DecisionBreakdown["deny"] != 1 {
		t.Errorf("DecisionBreakdown[deny]: want 1, got %d", summary.DecisionBreakdown["deny"])
	}
	if summary.ErrorCount != 1 {
		t.Errorf("ErrorCount: want 1, got %d", summary.ErrorCount)
	}
	if len(summary.Errors) != 1 || summary.Errors[0] != "timeout" {
		t.Errorf("Errors: want [timeout], got %v", summary.Errors)
	}
	wantDuration := 4 * time.Second
	if summary.Duration != wantDuration {
		t.Errorf("Duration: want %v, got %v", wantDuration, summary.Duration)
	}
}

func TestGenerateSummary_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	sessionID := "empty"
	// Create empty file.
	path := filepath.Join(dir, fmt.Sprintf("trace-%s.jsonl", sessionID))
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatalf("create: %v", err)
	}

	summary, err := GenerateSummary(dir, sessionID)
	if err != nil {
		t.Fatalf("GenerateSummary: %v", err)
	}
	if summary.TotalHooks != 0 {
		t.Errorf("TotalHooks: want 0, got %d", summary.TotalHooks)
	}
}

func TestGenerateSummary_MissingFile(t *testing.T) {
	dir := t.TempDir()
	summary, err := GenerateSummary(dir, "nonexistent")
	if err != nil {
		t.Fatalf("want no error for missing file, got: %v", err)
	}
	if summary == nil {
		t.Fatal("want non-nil summary")
	}
	if summary.TotalHooks != 0 {
		t.Errorf("TotalHooks: want 0, got %d", summary.TotalHooks)
	}
}

func TestGenerateSummary_MalformedLines(t *testing.T) {
	dir := t.TempDir()
	sessionID := "malformed"
	path := filepath.Join(dir, fmt.Sprintf("trace-%s.jsonl", sessionID))
	// Write one valid and one malformed line.
	content := `{"ts":"2026-04-07T10:00:00Z","event":"PreToolUse","handler":"h","duration_ms":5,"session_id":"malformed"}` + "\n" +
		`{invalid json` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	summary, err := GenerateSummary(dir, sessionID)
	if err != nil {
		t.Fatalf("GenerateSummary: %v", err)
	}
	// Only the valid line should be counted.
	if summary.TotalHooks != 1 {
		t.Errorf("TotalHooks: want 1, got %d", summary.TotalHooks)
	}
}

func TestSessionSummary_Top5Slowest(t *testing.T) {
	dir := t.TempDir()
	sessionID := "top5"
	base := time.Now()

	// Create 10 entries with increasing duration.
	var entries []TraceEntry
	for i := range 10 {
		entries = append(entries, TraceEntry{
			Timestamp:  base.Add(time.Duration(i) * time.Millisecond),
			Event:      "PreToolUse",
			Handler:    "h",
			DurationMs: int64(i + 1),
			Decision:   "allow",
			SessionID:  sessionID,
		})
	}
	writeTraceFile(t, dir, sessionID, entries)

	summary, err := GenerateSummary(dir, sessionID)
	if err != nil {
		t.Fatalf("GenerateSummary: %v", err)
	}

	if len(summary.Top5Slowest) != 5 {
		t.Fatalf("Top5Slowest: want 5 entries, got %d", len(summary.Top5Slowest))
	}
	// First entry should be slowest (duration 10).
	if summary.Top5Slowest[0].DurationMs != 10 {
		t.Errorf("Top5Slowest[0].DurationMs: want 10, got %d", summary.Top5Slowest[0].DurationMs)
	}
	// Should be descending.
	for i := 1; i < len(summary.Top5Slowest); i++ {
		if summary.Top5Slowest[i].DurationMs > summary.Top5Slowest[i-1].DurationMs {
			t.Errorf("Top5Slowest not descending at index %d", i)
		}
	}
}

func TestSessionSummary_FormatMarkdown(t *testing.T) {
	s := &SessionSummary{
		SessionID:  "md-test",
		TotalHooks: 3,
		Duration:   5 * time.Second,
		EventBreakdown: map[string]int{
			"PreToolUse": 2,
			"SessionEnd": 1,
		},
		DecisionBreakdown: map[string]int{
			"allow": 3,
		},
		Top5Slowest: []TraceEntry{
			{Event: "PreToolUse", Handler: "h", Tool: "Bash", DurationMs: 100},
		},
		ErrorCount: 0,
	}

	md := s.FormatMarkdown()

	checks := []string{
		"# Session Summary: md-test",
		"**Total Hook Invocations:** 3",
		"**Session Duration:**",
		"## Event Breakdown",
		"PreToolUse",
		"SessionEnd",
		"## Decision Breakdown",
		"allow",
		"## Top 5 Slowest Hook Executions",
		"Bash",
		"## Errors (0)",
	}
	for _, check := range checks {
		if !strings.Contains(md, check) {
			t.Errorf("markdown missing %q", check)
		}
	}
}

func TestSessionSummary_FormatMarkdown_Empty(t *testing.T) {
	s := &SessionSummary{
		SessionID:         "empty-md",
		EventBreakdown:    map[string]int{},
		DecisionBreakdown: map[string]int{},
	}
	md := s.FormatMarkdown()
	if !strings.Contains(md, "# Session Summary: empty-md") {
		t.Errorf("markdown missing header")
	}
	if !strings.Contains(md, "No events recorded") {
		t.Errorf("markdown should indicate no events")
	}
}

func TestSessionSummary_FormatMarkdown_WithErrors(t *testing.T) {
	s := &SessionSummary{
		SessionID:         "err-md",
		TotalHooks:        2,
		EventBreakdown:    map[string]int{"SessionStart": 2},
		DecisionBreakdown: map[string]int{},
		ErrorCount:        1,
		Errors:            []string{"connection refused"},
	}
	md := s.FormatMarkdown()
	if !strings.Contains(md, "connection refused") {
		t.Errorf("markdown missing error message")
	}
	if !strings.Contains(md, "## Errors (1)") {
		t.Errorf("markdown missing error count")
	}
}
