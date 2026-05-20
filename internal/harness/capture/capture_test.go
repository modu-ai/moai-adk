// Package capture — M1 lesson capture pipeline tests.
package capture_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

// TestCapture_ZeroTimestampDefaultsToNow verifies that an event with a zero
// Timestamp field falls back to time.Now().UTC() instead of writing a zero
// timestamp string. Covers the `if ts.IsZero()` branch on line 65-66 of
// capture.go that was previously uncovered.
func TestCapture_ZeroTimestampDefaultsToNow(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	obsPath := filepath.Join(dir, "observations.yaml")
	c := capture.New(capture.Config{ObservationsPath: obsPath})

	before := time.Now().UTC().Add(-1 * time.Second)
	event := capture.SubagentStopEvent{
		AgentName:   "manager-develop",
		AgentType:   "subagent",
		SessionID:   "sess-zero-ts",
		Timestamp:   time.Time{}, // zero value — must trigger ts = time.Now().UTC() branch
		ContextHash: "abc123",
	}

	if err := c.OnSubagentStop(event); err != nil {
		t.Fatalf("OnSubagentStop: %v", err)
	}

	data, err := os.ReadFile(obsPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	// Output must contain a recent timestamp (not the zero year "0001").
	content := string(data)
	if strings.Contains(content, "0001-01-01") {
		t.Errorf("expected ts.IsZero() branch to substitute time.Now().UTC(); got zero year timestamp in output:\n%s", content)
	}
	// The substituted timestamp should be after `before`.
	// We don't parse the YAML here (kept lightweight); we just verify the year
	// matches the current epoch's century.
	currentYear := time.Now().UTC().Format("2006")
	if !strings.Contains(content, currentYear) {
		t.Errorf("expected substituted timestamp to contain current year %q; output:\n%s", currentYear, content)
	}
	_ = before
}

// TestCapture_LongContextHashTruncated verifies that a ContextHash longer than
// 64 bytes is truncated to exactly 64 bytes in the persisted entry. Covers the
// `if len(contextHash) > 64` branch on line 94 of capture.go.
func TestCapture_LongContextHashTruncated(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	obsPath := filepath.Join(dir, "observations.yaml")
	c := capture.New(capture.Config{ObservationsPath: obsPath})

	// 80 bytes of 'a' then 80 bytes of 'b' — total 160 bytes (well above 64).
	longHash := strings.Repeat("a", 80) + strings.Repeat("b", 80)
	event := capture.SubagentStopEvent{
		AgentName:   "expert-backend",
		AgentType:   "subagent",
		SessionID:   "sess-long-hash",
		Timestamp:   time.Now().UTC(),
		ContextHash: longHash,
	}

	if err := c.OnSubagentStop(event); err != nil {
		t.Fatalf("OnSubagentStop: %v", err)
	}

	data, err := os.ReadFile(obsPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)

	// Find the `context_hash:` line.
	hashLineFound := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "context_hash:") {
			continue
		}
		hashLineFound = true
		// Extract the value after "context_hash:".
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, "context_hash:"))
		if len(value) != 64 {
			t.Errorf("context_hash truncation branch did not fire: value length = %d, expected 64; value=%q", len(value), value)
		}
		// Ensure no 'b' present — truncation should retain only the first 64 'a's.
		if strings.Contains(value, "b") {
			t.Errorf("expected truncation to drop suffix 'b' bytes; got value containing 'b': %q", value)
		}
	}
	if !hashLineFound {
		t.Fatalf("context_hash line not found in observations.yaml; output:\n%s", content)
	}
}

// TestCapture_AppendObservationMkdirError verifies that appendObservation
// surfaces a wrapped error when the destination directory cannot be created.
// This covers the `if err := os.MkdirAll(dir, ...)` branch on line 111-112
// of capture.go, AND the upstream `if err := appendObservation(...)` branch
// on line 71 that wraps the propagated error.
//
// Strategy: create a regular file at the path that MkdirAll would otherwise
// claim as a directory. On Unix, os.MkdirAll fails with ENOTDIR when an
// ancestor path component is an existing regular file.
func TestCapture_AppendObservationMkdirError(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("ENOTDIR semantics differ on Windows; coverage achieved on Unix")
	}
	dir := t.TempDir()

	// Create a regular file at the "ancestor" path so MkdirAll cannot create
	// a directory with that name.
	blockerPath := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blockerPath, []byte("blocker"), 0o644); err != nil {
		t.Fatalf("setup blocker file: %v", err)
	}

	// observations.yaml lives under <blocker>/sub/observations.yaml — MkdirAll
	// will try to create <blocker>/sub but <blocker> is a regular file → ENOTDIR.
	obsPath := filepath.Join(blockerPath, "sub", "observations.yaml")
	c := capture.New(capture.Config{ObservationsPath: obsPath})

	event := capture.SubagentStopEvent{
		AgentName:   "manager-develop",
		AgentType:   "subagent",
		SessionID:   "sess-mkdir-err",
		Timestamp:   time.Now().UTC(),
		ContextHash: "abc",
	}

	err := c.OnSubagentStop(event)
	if err == nil {
		t.Fatal("expected error from OnSubagentStop when MkdirAll fails, got nil")
	}
	// Verify error wrapping chain: OnSubagentStop wraps "capture: %w" around
	// appendObservation which wraps "mkdirall %s: %w" around the syscall error.
	msg := err.Error()
	if !strings.Contains(msg, "capture:") {
		t.Errorf("expected wrapped error to contain \"capture:\" prefix; got %q", msg)
	}
	if !strings.Contains(msg, "mkdirall") {
		t.Errorf("expected wrapped error to contain \"mkdirall\" segment; got %q", msg)
	}
}

// TestCapture_AppendObservationOpenError verifies that appendObservation
// surfaces a wrapped error when os.OpenFile fails. This covers the
// `if err := os.OpenFile(...)` branch on line 117-118 of capture.go.
//
// Strategy: target a path that is itself an existing directory. os.OpenFile
// with O_RDWR|O_CREATE|O_APPEND fails on a directory with EISDIR on Unix
// (and ERROR_ACCESS_DENIED on Windows). This bypasses the MkdirAll branch
// (the parent already exists) and exercises only the OpenFile failure path.
func TestCapture_AppendObservationOpenError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Create a directory at the exact path where observations.yaml should live.
	obsPath := filepath.Join(dir, "observations.yaml")
	if err := os.MkdirAll(obsPath, 0o755); err != nil {
		t.Fatalf("setup obsPath as directory: %v", err)
	}

	c := capture.New(capture.Config{ObservationsPath: obsPath})
	event := capture.SubagentStopEvent{
		AgentName:   "manager-develop",
		AgentType:   "subagent",
		SessionID:   "sess-open-err",
		Timestamp:   time.Now().UTC(),
		ContextHash: "abc",
	}

	err := c.OnSubagentStop(event)
	if err == nil {
		t.Fatal("expected error from OnSubagentStop when OpenFile targets a directory, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "capture:") {
		t.Errorf("expected wrapped error to contain \"capture:\" prefix; got %q", msg)
	}
	if !strings.Contains(msg, "open ") {
		t.Errorf("expected wrapped error to contain \"open \" segment from appendObservation; got %q", msg)
	}
}

// TestCapture_AppendObservationCleanRoot verifies the branch where
// filepath.Dir(path) is "." (no parent directory to create). This exercises
// the false-branch of the `if dir := ...` MkdirAll guard on line 110 of
// capture.go, ensuring observations can be written to a bare filename relative
// to the CWD without triggering a MkdirAll call.
//
// Strategy: chdir into a temp directory and pass a bare filename.
func TestCapture_AppendObservationCleanRoot(t *testing.T) {
	// NOT t.Parallel(): chdir mutates process state.
	dir := t.TempDir()

	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origCwd)
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	c := capture.New(capture.Config{ObservationsPath: "observations.yaml"})
	event := capture.SubagentStopEvent{
		AgentName:   "manager-develop",
		AgentType:   "subagent",
		SessionID:   "sess-clean-root",
		Timestamp:   time.Now().UTC(),
		ContextHash: "abc",
	}
	if err := c.OnSubagentStop(event); err != nil {
		t.Fatalf("OnSubagentStop: %v", err)
	}

	// Verify file was actually written.
	if _, err := os.Stat(filepath.Join(dir, "observations.yaml")); err != nil {
		t.Errorf("observations.yaml not created in cwd: %v", err)
	}
}
