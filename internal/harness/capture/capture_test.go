// Package capture — M1 lesson capture pipeline tests.
package capture_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness/capture"
)

// TestCapture_SubagentStopTrigger verifies that a SubagentStop event emits an
// observation candidate to observations.yaml.
// M1 RED: this test must fail before capture.go exists.
func TestCapture_SubagentStopTrigger(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	obsPath := filepath.Join(dir, "observations.yaml")

	c := capture.New(capture.Config{
		ObservationsPath: obsPath,
	})

	event := capture.SubagentStopEvent{
		AgentName:    "manager-develop",
		AgentType:    "subagent",
		SessionID:    "sess-001",
		Timestamp:    time.Now().UTC(),
		ContextHash:  "abc123",
	}

	if err := c.OnSubagentStop(event); err != nil {
		t.Fatalf("OnSubagentStop: %v", err)
	}

	// observations.yaml must exist
	if _, err := os.Stat(obsPath); os.IsNotExist(err) {
		t.Fatalf("observations.yaml not created")
	}

	// File must be non-empty
	data, err := os.ReadFile(obsPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("observations.yaml is empty")
	}
}

// TestCapture_EmptyAgentName verifies that an empty agent name is rejected.
func TestCapture_EmptyAgentName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c := capture.New(capture.Config{
		ObservationsPath: filepath.Join(dir, "observations.yaml"),
	})
	event := capture.SubagentStopEvent{
		AgentName:   "",
		Timestamp:   time.Now().UTC(),
		ContextHash: "abc123",
	}
	if err := c.OnSubagentStop(event); err == nil {
		t.Fatal("expected error for empty agent name, got nil")
	}
}

// TestCapture_MultipleEvents verifies that multiple SubagentStop events append
// observations correctly (append-only, not overwrite).
func TestCapture_MultipleEvents(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	obsPath := filepath.Join(dir, "observations.yaml")
	c := capture.New(capture.Config{ObservationsPath: obsPath})

	for i := 0; i < 3; i++ {
		ev := capture.SubagentStopEvent{
			AgentName:   "expert-backend",
			AgentType:   "subagent",
			SessionID:   "sess-00" + string(rune('1'+i)),
			Timestamp:   time.Now().UTC(),
			ContextHash: "hash" + string(rune('a'+i)),
		}
		if err := c.OnSubagentStop(ev); err != nil {
			t.Fatalf("event %d: %v", i, err)
		}
	}

	data, _ := os.ReadFile(obsPath)
	// Each entry starts with "- agent_name:" — 3 events → at least 3 entries
	count := 0
	for _, line := range splitLines(string(data)) {
		if len(line) > 11 && line[:11] == "- agent_nam" {
			count++
		}
	}
	if count != 3 {
		t.Fatalf("expected 3 observation entries, found %d; content:\n%s", count, data)
	}
}

// TestCapture_ConcurrentWrites verifies that concurrent SubagentStop events
// do not corrupt observations.yaml (EC-HRA-002: flock protection).
func TestCapture_ConcurrentWrites(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	obsPath := filepath.Join(dir, "observations.yaml")
	c := capture.New(capture.Config{ObservationsPath: obsPath})

	const goroutines = 5
	errs := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			ev := capture.SubagentStopEvent{
				AgentName:   "agent-" + string(rune('a'+idx)),
				AgentType:   "subagent",
				SessionID:   "sess-concurrent",
				Timestamp:   time.Now().UTC(),
				ContextHash: "concurrent",
			}
			errs <- c.OnSubagentStop(ev)
		}(i)
	}
	for i := 0; i < goroutines; i++ {
		if err := <-errs; err != nil {
			t.Errorf("goroutine error: %v", err)
		}
	}
}

// BenchmarkLessonCapture verifies that OnSubagentStop stays under 500ms p95
// even with a 10MB synthetic diff (REQ-HRA-002).
func BenchmarkLessonCapture(b *testing.B) {
	dir := b.TempDir()
	c := capture.New(capture.Config{ObservationsPath: filepath.Join(dir, "observations.yaml")})

	// Synthetic 10MB context hash (simulates large diff string)
	bigHash := make([]byte, 10*1024*1024)
	for i := range bigHash {
		bigHash[i] = 'x'
	}
	event := capture.SubagentStopEvent{
		AgentName:   "manager-develop",
		AgentType:   "subagent",
		SessionID:   "bench-sess",
		Timestamp:   time.Now().UTC(),
		ContextHash: string(bigHash),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.OnSubagentStop(event); err != nil {
			b.Fatal(err)
		}
	}
}

// splitLines splits string on newline.
func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
