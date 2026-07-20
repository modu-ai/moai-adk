package migration

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestLog_AppendsJSONLine verifies JSONL-format append.
// REQ-V3R2-RT-007-014: every applied migration records a structured log entry.
func TestLog_AppendsJSONLine(t *testing.T) {
	root := t.TempDir()
	entry := LogEntry{Version: 1, Name: "m001", Result: "success", Details: "applied"}
	if err := Append(root, entry); err != nil {
		t.Fatalf("Append: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".moai", "logs", logFileName))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, `"version":1`) {
		t.Errorf("log missing version field: %s", s)
	}
	if !strings.Contains(s, `"result":"success"`) {
		t.Errorf("log missing result field: %s", s)
	}
	if !strings.HasSuffix(s, "\n") {
		t.Errorf("log line must end with newline (JSONL): %q", s)
	}
}

// TestLog_PreservesPriorEntries verifies preservation of existing entries.
// REQ-V3R2-RT-007-014: the log is append-only and preserves existing entries.
func TestLog_PreservesPriorEntries(t *testing.T) {
	root := t.TempDir()
	if err := Append(root, LogEntry{Version: 1, Name: "m001", Result: "success"}); err != nil {
		t.Fatal(err)
	}
	if err := Append(root, LogEntry{Version: 2, Name: "m002", Result: "success"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".moai", "logs", logFileName))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(lines))
	}
	// LastApplied must return the most recent success.
	last, err := LastApplied(root)
	if err != nil {
		t.Fatalf("LastApplied: %v", err)
	}
	if last == nil || last.Version != 2 {
		t.Errorf("LastApplied: got %+v, want version 2", last)
	}
}

// TestLog_HandlesConcurrentWrites verifies behavior under concurrent writes.
// REQ-V3R2-RT-007-014: log writes must be thread-safe.
func TestLog_HandlesConcurrentWrites(t *testing.T) {
	root := t.TempDir()
	const goroutines = 10
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = Append(root, LogEntry{Version: n, Name: "m", Result: "success"})
		}(i)
	}
	wg.Wait()
	data, err := os.ReadFile(filepath.Join(root, ".moai", "logs", logFileName))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	// os.O_APPEND with small writes (≤ PIPE_BUF) is atomic on POSIX; each
	// goroutine writes exactly one JSONL line.
	if len(lines) != goroutines {
		t.Errorf("expected %d lines, got %d (concurrent append interleaving?)", goroutines, len(lines))
	}
}

// TestLog_LastApplied_Absent verifies LastApplied on a missing log file.
func TestLog_LastApplied_Absent(t *testing.T) {
	root := t.TempDir()
	last, err := LastApplied(root)
	if err != nil {
		t.Fatalf("LastApplied on absent log: %v", err)
	}
	if last != nil {
		t.Errorf("LastApplied on absent log: got %+v, want nil", last)
	}
}

// TestLog_LastApplied_SkipsFailed verifies LastApplied ignores failed entries.
func TestLog_LastApplied_SkipsFailed(t *testing.T) {
	root := t.TempDir()
	_ = Append(root, LogEntry{Version: 1, Name: "m001", Result: "failed"})
	_ = Append(root, LogEntry{Version: 2, Name: "m002", Result: "success"})
	last, err := LastApplied(root)
	if err != nil {
		t.Fatalf("LastApplied: %v", err)
	}
	if last == nil || last.Version != 2 {
		t.Errorf("LastApplied should skip failed and return version 2; got %+v", last)
	}
}

// TestLog_LastApplied_MalformedLine verifies LastApplied skips unparseable JSONL lines.
func TestLog_LastApplied_MalformedLine(t *testing.T) {
	root := t.TempDir()
	logsDir := filepath.Join(root, ".moai", "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A malformed line followed by a valid one.
	content := "garbage-not-json\n{\"version\":2,\"name\":\"m\",\"result\":\"success\",\"details\":\"\"}\n"
	if err := os.WriteFile(filepath.Join(logsDir, logFileName), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	last, err := LastApplied(root)
	if err != nil {
		t.Fatalf("LastApplied: %v", err)
	}
	if last == nil || last.Version != 2 {
		t.Errorf("LastApplied should skip malformed line and return version 2; got %+v", last)
	}
}

// TestLog_parseJSONL_TrailingData exercises the trailing-data branch (content
// without a trailing newline).
func TestLog_parseJSONL_TrailingData(t *testing.T) {
	lines := parseJSONL([]byte("a\nb"))
	if len(lines) != 2 {
		t.Errorf("trailing data: got %d lines, want 2", len(lines))
	}
}
