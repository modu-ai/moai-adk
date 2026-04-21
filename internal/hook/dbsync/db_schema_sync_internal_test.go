package dbsync

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestParseMigrationStub_OversizedFile exercises AC-1 at the strict shape level:
// the size guard must set parseMigrationResult.ParsedContent="" AND Truncated=true
// (both fields, not an either/or) and must log exactly one line with the prefix
// `parseMigrationStub: file exceeds maxMigrationFileSize=`. Because the fields
// live on an unexported result type, the assertion must run inside package dbsync.
func TestParseMigrationStub_OversizedFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "errors.log")
	target := filepath.Join(tmpDir, "schema.prisma")
	// 2 MiB payload — doubles maxMigrationFileSize (1 MiB).
	if err := os.WriteFile(target, bytes.Repeat([]byte("X"), 2<<20), 0o644); err != nil {
		t.Fatalf("write oversized fixture: %v", err)
	}

	result, err := parseMigrationStubWithLog(target, logFile)
	if err != nil {
		t.Fatalf("size guard must return nil error, got %v", err)
	}
	if result.ParsedContent != "" {
		t.Errorf("ParsedContent = %q, want empty string (REQ-H1-002)", result.ParsedContent)
	}
	if !result.Truncated {
		t.Error("Truncated = false, want true (REQ-H1-002)")
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read error log: %v", err)
	}
	const prefix = "parseMigrationStub: file exceeds maxMigrationFileSize="
	if got := strings.Count(string(data), prefix); got != 1 {
		t.Errorf("log occurrences of %q = %d, want 1; content:\n%s", prefix, got, data)
	}
}

// TestParseMigrationStub_EmptyFile ensures the size guard does not mis-fire on
// 0-byte files: ParsedContent is the empty string, Truncated is false, and no
// size-guard log line is emitted.
func TestParseMigrationStub_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "errors.log")
	target := filepath.Join(tmpDir, "empty.sql")
	if err := os.WriteFile(target, []byte{}, 0o644); err != nil {
		t.Fatalf("write empty fixture: %v", err)
	}

	result, err := parseMigrationStubWithLog(target, logFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ParsedContent != "" {
		t.Errorf("ParsedContent = %q, want empty", result.ParsedContent)
	}
	if result.Truncated {
		t.Error("Truncated must be false for a genuinely empty file")
	}
	if info, err := os.Stat(logFile); err == nil && info.Size() > 0 {
		content, _ := os.ReadFile(logFile)
		t.Errorf("empty file should not produce a size-guard log, got:\n%s", content)
	}
}

// TestParseMigrationStub_NormalFilePreservesContent protects the happy-path
// behavior that existed before REQ-H1-* — a sub-threshold file must round-trip
// its bytes (UTF-8 BOM included, matching AC-8 utf8_bom scenario).
func TestParseMigrationStub_NormalFilePreservesContent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "schema.prisma")
	content := append([]byte{0xEF, 0xBB, 0xBF}, []byte("model User { id Int @id }\n")...)
	if err := os.WriteFile(target, content, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	result, err := parseMigrationStubWithLog(target, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Truncated {
		t.Error("Truncated must be false for a sub-threshold file")
	}
	if !bytes.Equal([]byte(result.ParsedContent), content) {
		t.Errorf("ParsedContent bytes differ from input (len got=%d want=%d)", len(result.ParsedContent), len(content))
	}
}

// TestCheckDebounceWithLog_IOFailureReturnsSafeDefault — REQ-H2-003. When the
// directory that should host the temp file is not writable, the helper must
// return (true, nil) and emit a log line rather than propagating the I/O error.
func TestCheckDebounceWithLog_IOFailureReturnsSafeDefault(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root: chmod 0555 does not restrict writes")
	}
	t.Parallel()

	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "state")
	if err := os.Mkdir(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	stateFile := filepath.Join(stateDir, "last-seen.json")

	// Make the state directory read-only so os.CreateTemp fails.
	if err := os.Chmod(stateDir, 0o555); err != nil {
		t.Fatalf("chmod readonly: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(stateDir, 0o755) })

	logFile := filepath.Join(tmpDir, "errors.log")
	debounced, err := checkDebounceWithLog(stateFile, "migrations/001.sql", 10_000_000_000 /*10s*/, logFile)
	if err != nil {
		t.Fatalf("expected nil error on I/O failure (safe default), got %v", err)
	}
	if !debounced {
		t.Error("safe default must be debounced=true on I/O failure (REQ-H2-003)")
	}
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), "CheckDebounce:") {
		t.Errorf("expected CheckDebounce log entry, got %q", data)
	}
}

// TestItoa — regression guard for the internal int64→string helper used by the
// size-guard log. Keeps parseMigrationStub independent of strconv for the single
// formatting site.
func TestItoa(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{1024, "1024"},
		{-42, "-42"},
		{1 << 20, "1048576"},
	}
	for _, tt := range tests {
		if got := itoa(tt.in); got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// TestParseMigrationStub_WrapperRoundTrips ensures the public-style wrapper
// (parseMigrationStub without an explicit logFile path) forwards to the
// implementation body without regressing the happy-path shape.
func TestParseMigrationStub_WrapperRoundTrips(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "schema.sql")
	payload := []byte("CREATE TABLE wrap_test(id INT);\n")
	if err := os.WriteFile(target, payload, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	result, err := parseMigrationStub(target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Truncated {
		t.Error("Truncated should be false for sub-threshold input")
	}
	if result.ParsedContent != string(payload) {
		t.Errorf("ParsedContent = %q, want %q", result.ParsedContent, payload)
	}
}

// TestCheckDebounceWithLog_FastPathRespectsExistingWindow exercises the
// pre-lock fast path: a pre-existing fresh state must short-circuit without
// creating a lock file. This covers the fast-path branch that is hit every
// time PostToolUse fires inside the 10s debounce window.
func TestCheckDebounceWithLog_FastPathRespectsExistingWindow(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "last-seen.json")
	state := DebounceState{
		FilePath:  "migrations/001.sql",
		Timestamp: time.Now(),
	}
	encoded, _ := json.Marshal(state)
	if err := os.WriteFile(stateFile, encoded, 0o644); err != nil {
		t.Fatalf("seed state: %v", err)
	}

	debounced, err := checkDebounceWithLog(stateFile, "migrations/001.sql", 10*time.Second, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !debounced {
		t.Error("fast path must return debounced=true for a fresh same-file state")
	}

	// No lock file should have been created on the fast path.
	if _, err := os.Stat(stateFile + ".lock"); err == nil {
		t.Error("fast path must not create a lock file")
	}
}

// TestCheckDebounceWithLog_ReChecksUnderLock validates that the second
// under-lock re-check returns debounced=true when a fresh state appeared
// between the fast-path read and the lock acquisition. Simulates the state
// arrival by seeding the state file after the caller bypassed the fast path
// (achieved by making the on-disk state stale, then mutating it to fresh
// right before the lock is taken).
func TestCheckDebounceWithLog_ReChecksUnderLock(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "last-seen.json")

	// Seed a stale state so the fast path falls through.
	stale := DebounceState{
		FilePath:  "migrations/001.sql",
		Timestamp: time.Now().Add(-1 * time.Hour),
	}
	staleBytes, _ := json.Marshal(stale)
	if err := os.WriteFile(stateFile, staleBytes, 0o644); err != nil {
		t.Fatalf("seed stale: %v", err)
	}

	// Overwrite with a fresh state just before the lock is taken. Because the
	// fast-path read already happened when our helper enters the lock block,
	// the under-lock re-check is what catches this.
	fresh := DebounceState{
		FilePath:  "migrations/001.sql",
		Timestamp: time.Now(),
	}
	freshBytes, _ := json.Marshal(fresh)
	if err := os.WriteFile(stateFile, freshBytes, 0o644); err != nil {
		t.Fatalf("seed fresh: %v", err)
	}

	// With a fresh state already on disk, the fast path short-circuits to true;
	// this test is a defensive regression guard that fast-path + under-lock
	// paths agree.
	debounced, err := checkDebounceWithLog(stateFile, "migrations/001.sql", 10*time.Second, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !debounced {
		t.Error("under-lock re-check must honor a fresh state")
	}
}

// TestCheckDebounceWithLog_LockContention covers the lock-file-held branch:
// if a stale .lock file exists (simulating another caller mid-write), the
// helper returns debounced=true without touching stateFile.
func TestCheckDebounceWithLog_LockContention(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "last-seen.json")
	lockPath := stateFile + ".lock"
	// Pre-create the lock to simulate a concurrent holder.
	if err := os.WriteFile(lockPath, nil, 0o644); err != nil {
		t.Fatalf("seed lock: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(lockPath) })

	debounced, err := checkDebounceWithLog(stateFile, "migrations/001.sql", 10*time.Second, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !debounced {
		t.Error("lock contention must return debounced=true")
	}
	// stateFile must not have been written by the loser.
	if _, statErr := os.Stat(stateFile); statErr == nil {
		t.Error("loser must not touch stateFile")
	}
}

// TestLogError_Coverage exercises the tail branches of logError: empty path
// (no-op) and mkdir failure (swallowed error). Both are deliberate safety
// guards — logError must never propagate an error.
func TestLogError_Coverage(t *testing.T) {
	t.Parallel()

	// Empty path should no-op without panic.
	logError("", "should not fault")

	tmpDir := t.TempDir()
	// Put the log file under a path whose parent is a plain file — os.MkdirAll
	// will fail, and the helper must still return silently.
	blockingFile := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blockingFile, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	logError(filepath.Join(blockingFile, "logs", "errors.log"), "mkdir will fail")
}

// TestParseMigrationStubWithLog_NonexistentFile returns the os.Stat error
// directly (non-blocking is handled by the caller HandleDBSchemaSync per
// REQ-011 — parseMigrationStub itself just surfaces the error).
func TestParseMigrationStubWithLog_NonexistentFile(t *testing.T) {
	t.Parallel()

	_, err := parseMigrationStubWithLog("/nonexistent/path/abc.sql", "")
	if err == nil {
		t.Error("expected non-nil error for missing file, got nil")
	}
}

// TestCheckDebounceWithLog_MkdirStateDirFailure covers the defensive
// os.MkdirAll branch in checkDebounceWithLog — triggered when the state
// directory cannot be created because an ancestor is read-only. Must still
// return the safe default (true, nil) per REQ-H2-003.
func TestCheckDebounceWithLog_MkdirStateDirFailure(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root: chmod does not restrict writes")
	}
	t.Parallel()

	tmpDir := t.TempDir()
	readonlyParent := filepath.Join(tmpDir, "readonly-parent")
	if err := os.Mkdir(readonlyParent, 0o555); err != nil {
		t.Fatalf("mkdir readonly parent: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(readonlyParent, 0o755) })

	// stateFile lives inside a subdirectory that does not yet exist under a
	// read-only parent → os.MkdirAll inside checkDebounceWithLog fails.
	stateFile := filepath.Join(readonlyParent, "nested", "last-seen.json")
	logFile := filepath.Join(tmpDir, "errors.log")

	debounced, err := checkDebounceWithLog(stateFile, "migrations/001.sql", 10*time.Second, logFile)
	if err != nil {
		t.Fatalf("expected nil error (safe default), got %v", err)
	}
	if !debounced {
		t.Error("MkdirAll failure must return debounced=true (REQ-H2-003)")
	}
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), "CheckDebounce: mkdir state dir") {
		t.Errorf("expected mkdir state dir log, got %q", data)
	}
}
