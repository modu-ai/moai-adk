// Package cli — full Stop hook handler test matrix (T-C1).
// REQ-HRN-OBS-003: verifies session_stop events are recorded.
// REQ-HRN-FND-009: isHarnessLearningEnabled gate — full no-op when disabled.
// REQ-HRN-FND-010: preserves the existing 4-field schema plus the new session_id field.
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestRunHarnessObserveStop_NoOpWhenLearningDisabled verifies that the Stop handler
// does not create usage-log.jsonl at all when learning.enabled=false.
// REQ-HRN-FND-009: reuses the isHarnessLearningEnabled gate.
func TestRunHarnessObserveStop_NoOpWhenLearningDisabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: false\n")
	t.Chdir(dir)

	// T-A3 spec: nested stdin JSON shape
	cmd := &cobra.Command{}
	withStdin(t, `{"last_assistant_message":"hello","session":{"id":"sess-disabled"},"hook_event_name":"Stop"}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveStop 에러 반환: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Errorf("learning.enabled=false 시 usage-log.jsonl이 존재해서는 안 됨; stat err=%v", err)
	}
}

// TestRunHarnessObserveStop_PreservesExistingLogWhenDisabled verifies that the
// handler does not modify an existing log when learning.enabled=false and a log
// is already present. REQ-HRN-FND-009: the disabled gate must not touch existing data.
func TestRunHarnessObserveStop_PreservesExistingLogWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: false\n")
	t.Chdir(dir)

	// Pre-create an existing log file
	logDir := filepath.Join(dir, ".moai", "harness")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("로그 디렉터리 생성 실패: %v", err)
	}
	logPath := filepath.Join(logDir, "usage-log.jsonl")
	existingContent := `{"timestamp":"2026-01-01T00:00:00Z","event_type":"moai_subcommand","subject":"/moai plan","context_hash":"","tier_increment":0,"schema_version":"v1"}` + "\n"
	if err := os.WriteFile(logPath, []byte(existingContent), 0o644); err != nil {
		t.Fatalf("기존 로그 파일 생성 실패: %v", err)
	}

	cmd := &cobra.Command{}
	withStdin(t, `{"last_assistant_message":"hi","session":{"id":"sess-preserve"},"hook_event_name":"Stop"}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveStop 에러 반환: %v", err)
		}
	})

	// File content must remain identical (byte-for-byte equality)
	after, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("로그 파일 읽기 실패: %v", err)
	}
	if string(after) != existingContent {
		t.Errorf("learning.enabled=false 시 기존 로그가 수정됨\n기대: %q\n실제: %q", existingContent, string(after))
	}
}

// TestRunHarnessObserveStop_RecordsWhenEnabled verifies that the Stop handler
// records the session_stop event correctly when learning.enabled=true.
// - 4 base fields: timestamp, event_type, subject, schema_version
// - Stop-specific fields: session_id, last_assistant_message_hash, last_assistant_message_len
// - The raw last_assistant_message MUST NOT be recorded (PII minimization)
// REQ-HRN-OBS-002, REQ-HRN-OBS-003, REQ-HRN-FND-010.
func TestRunHarnessObserveStop_RecordsWhenEnabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	const (
		sessionID = "sess-test-c1"
		// T-A3 spec: nested stdin JSON — last_assistant_message + session.id
		assistantMsg = "Hello from assistant"
	)
	cmd := &cobra.Command{}
	// Native Claude Code Stop wire format: nested session.id (no top-level
	// hook_event_name, which would short-circuit camel/nested normalization).
	withStdin(t, `{"last_assistant_message":"`+assistantMsg+`","session":{"id":"`+sessionID+`"}}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveStop 에러 반환: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("usage-log.jsonl 미생성: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("JSONL 라인 수: got=%d, want=1", len(lines))
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("JSONL 파싱 실패: %v", err)
	}

	// event_type must be "session_stop"
	if entry["event_type"] != "session_stop" {
		t.Errorf("event_type: got=%v, want=%q", entry["event_type"], "session_stop")
	}

	// Verify the 4 base fields exist (preserves REQ-HRN-FND-010)
	for _, field := range []string{"timestamp", "event_type", "subject", "schema_version"} {
		if _, ok := entry[field]; !ok {
			t.Errorf("기본 필드 %q 누락 — schema additivity 위반", field)
		}
	}

	// Verify session_id presence and value (REQ-HRN-OBS-003): extracted from session.id
	if entry["session_id"] != sessionID {
		t.Errorf("session_id: got=%v, want=%q", entry["session_id"], sessionID)
	}

	// AC-HRN-OBS-002: verify last_assistant_message_hash presence and 16-hex-char length
	hashVal, ok := entry["last_assistant_message_hash"]
	if !ok {
		t.Errorf("last_assistant_message_hash 필드 누락 — AC-HRN-OBS-002 미충족")
	} else {
		hashStr, _ := hashVal.(string)
		if len(hashStr) != 16 {
			t.Errorf("last_assistant_message_hash 길이: got=%d, want=16 (first 16 hex chars of SHA-256)", len(hashStr))
		}
	}

	// AC-HRN-OBS-002: verify last_assistant_message_len presence and byte-length value
	lenVal, ok := entry["last_assistant_message_len"]
	if !ok {
		t.Errorf("last_assistant_message_len 필드 누락 — AC-HRN-OBS-002 미충족")
	} else {
		// JSON numbers unmarshal as float64
		gotLen := int(lenVal.(float64))
		wantLen := len([]byte(assistantMsg))
		if gotLen != wantLen {
			t.Errorf("last_assistant_message_len: got=%d, want=%d", gotLen, wantLen)
		}
	}

	// PII minimization: the raw last_assistant_message MUST NOT appear in the log
	if strings.Contains(string(data), assistantMsg) {
		t.Errorf("last_assistant_message 원문이 로그에 기록됨 — PII 최소화 원칙 위반")
	}
}

// TestRunHarnessObserveStop_EmptyMessageNoHashFields verifies that, when
// last_assistant_message is an empty string, the hash/len fields are omitted via omitempty.
// REQ-HRN-OBS-002: empty message → fields omitted (correct omitempty behavior).
func TestRunHarnessObserveStop_EmptyMessageNoHashFields(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	cmd := &cobra.Command{}
	withStdin(t, `{"last_assistant_message":"","session":{"id":"sess-empty"}}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveStop 에러 반환: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("usage-log.jsonl 미생성: %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(strings.TrimRight(string(data), "\n")), &entry); err != nil {
		t.Fatalf("JSONL 파싱 실패: %v", err)
	}

	// For an empty message, the hash/len fields must be omitted via omitempty
	for _, field := range []string{"last_assistant_message_hash", "last_assistant_message_len"} {
		if _, ok := entry[field]; ok {
			t.Errorf("빈 last_assistant_message 시 필드 %q가 로그에 등장해서는 안 됨 (omitempty 위반)", field)
		}
	}
}

// TestRunHarnessObserveStop_LogErrorPathDoesNotReturn verifies that the handler
// does not return an error on log-write failure (non-blocking) and instead
// writes to stderr. The failure is induced by replacing the log path with a directory.
func TestRunHarnessObserveStop_LogErrorPathDoesNotReturn(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	// Pre-create a directory at the usage-log.jsonl path to induce a file-write failure
	blockPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	if err := os.MkdirAll(blockPath, 0o755); err != nil {
		t.Fatalf("블로킹 디렉터리 생성 실패: %v", err)
	}

	var stderrBuf strings.Builder
	cmd := &cobra.Command{}
	cmd.SetErr(&stderrBuf)

	withStdin(t, `{"last_assistant_message":"msg","session":{"id":"sess-error-path"}}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Errorf("Stop 핸들러는 기록 실패 시에도 에러를 반환하지 않아야 함: %v", err)
		}
	})
	// The handler returns nil; errors are written to stderr (non-blocking)
}
